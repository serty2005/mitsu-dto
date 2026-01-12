package gui

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lxn/walk"
)

func runNetworkScan() {
	mw.Synchronize(func() {
		actionBtn.SetEnabled(false)
		actionBtn.SetText("Сканирование...")
		logMsg("--- Запуск поиска по ARP (MAC 00-22-00...) ---")
	})

	// 1. "Прогрев" ARP
	if err := triggerArpDiscovery(); err != nil {
		logMsg("Ошибка инициации ARP: %v", err)
	}

	// Даем ОС время на обновление таблицы
	time.Sleep(1 * time.Second)

	// 2. Читаем таблицу
	arpTable, err := getArpTable()
	if err != nil {
		logMsg("Ошибка чтения ARP таблицы: %v", err)
		restoreBtnState()
		return
	}

	// 3. Фильтруем по MAC Mitsu
	var candidates []string
	mitsuPrefix := "00-22-00"

	for ip, mac := range arpTable {
		// Нормализуем MAC
		cleanMac := strings.ReplaceAll(mac, "-", "")
		cleanMac = strings.ReplaceAll(cleanMac, ":", "")
		cleanMac = strings.ToUpper(cleanMac)
		cleanPrefix := strings.ReplaceAll(mitsuPrefix, "-", "")

		if strings.HasPrefix(cleanMac, cleanPrefix) {
			logMsg("Найден кандидат в ARP: %s [%s]", ip, mac)
			candidates = append(candidates, ip)
		}
	}

	if len(candidates) == 0 {
		mw.Synchronize(func() {
			logMsg("Устройства Mitsu не найдены в ARP.")
			walk.MsgBox(mw, "Результат", "Устройства не найдены.\nПопробуйте пропинговать устройство вручную.", walk.MsgBoxIconInformation)
			restoreBtnState()
		})
		return
	}

	// 4. Проверяем открытый порт 8200
	foundChan := make(chan string, len(candidates))
	var wg sync.WaitGroup

	logMsg("Проверка порта 8200 у %d кандидатов...", len(candidates))

	for _, ip := range candidates {
		wg.Add(1)
		go func(targetIP string) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:8200", targetIP), 500*time.Millisecond)
			if err == nil {
				conn.Close()
				foundChan <- targetIP
			} else {
				logMsg("IP %s: MAC совпал, но порт 8200 закрыт.", targetIP)
			}
		}(ip)
	}

	wg.Wait()
	close(foundChan)

	var foundList []string
	for ip := range foundChan {
		logMsg("ПОДТВЕРЖДЕНО: %s", ip)
		foundList = append(foundList, fmt.Sprintf("%s:8200", ip))
	}

	mw.Synchronize(func() {
		if len(foundList) > 0 {
			// ИСПРАВЛЕНИЕ: Передаем mainApp в getInitialDeviceList
			newList := getInitialDeviceList(mainApp)
			searchItem := newList[len(newList)-1]
			newList = newList[:len(newList)-1]
			newList = append(newList, foundList...)
			newList = append(newList, searchItem)

			addrCombo.SetModel(newList)
			addrCombo.SetText(foundList[0])
			logMsg("Найдено %d устр.", len(foundList))
		} else {
			logMsg("Порт 8200 недоступен у найденных MAC.")
			walk.MsgBox(mw, "Результат", "Устройства найдены по MAC, но порт 8200 закрыт.", walk.MsgBoxIconWarning)
		}
		updateUIState()
		actionBtn.SetEnabled(true)
	})
}

func restoreBtnState() {
	mw.Synchronize(func() {
		updateUIState()
	})
}

// triggerArpDiscovery пингует подсеть UDP пакетами
func triggerArpDiscovery() error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 100)

	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
			continue
		}

		ip := ipnet.IP.To4()
		mask := ipnet.Mask
		ones, _ := mask.Size()
		if ones < 24 {
			continue // Пропускаем большие сети
		}

		baseIP := ip.Mask(mask)
		for i := 1; i < 255; i++ {
			targetIP := net.IPv4(baseIP[0], baseIP[1], baseIP[2], byte(i))
			if targetIP.Equal(ip) {
				continue
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(ipStr string) {
				defer wg.Done()
				defer func() { <-sem }()
				// Шлем пакет на порт 8200 (или любой другой), чтобы инициировать ARP запрос
				conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:8200", ipStr), 100*time.Millisecond)
				if err == nil {
					conn.Write([]byte{0x00})
					conn.Close()
				}
			}(targetIP.String())
		}
	}
	wg.Wait()
	return nil
}

// getArpTable парсит 'arp -a'
func getArpTable() (map[string]string, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	// Регулярка для IP и MAC (поддержка Windows "-" и Unix ":")
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+([0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2})`)

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			result[matches[1]] = matches[2]
		}
	}
	return result, nil
}
