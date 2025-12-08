package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lxn/walk"

	d "github.com/lxn/walk/declarative"
	"go.bug.st/serial"

	"mitsuscanner/mitsu"
)

// Global state
var (
	mw      *walk.MainWindow
	logView *walk.TextEdit

	// Элементы управления одной строкой
	addrCombo  *walk.ComboBox   // Умный комбобокс (COMы + IP)
	paramInput *walk.ComboBox   // Скорость или Порт
	actionBtn  *walk.PushButton // Кнопка действия (Искать/Подкл/Откл)

	// Элементы вкладки "Информация"
	infoView *walk.TextEdit // Текстовое поле для инфо

	driver mitsu.Driver
)

const (
	itemSearchLAN  = "Поиск в сети / Ввести IP..."
	defaultTCPPort = "8200"
	defaultBaud    = "115200"
)

func main() {
	if _, err := (d.MainWindow{
		AssignTo: &mw,
		Title:    "Mitsu Driver Utility",
		Size:     d.Size{Width: 460, Height: 550},
		MinSize:  d.Size{Width: 460, Height: 500},
		MaxSize:  d.Size{Width: 460, Height: 600},
		Layout:   d.VBox{MarginsZero: true, Spacing: 5},
		Children: []d.Widget{
			// --- Единая строка подключения ---
			d.GroupBox{
				// Title:  "Подключение",
				Layout: d.Grid{Columns: 5, Margins: d.Margins{Left: 5, Top: 1, Right: 3, Bottom: 6}, Spacing: 4},
				Children: []d.Widget{
					d.Label{Text: "Устройство:"},
					d.ComboBox{
						AssignTo:              &addrCombo,
						Editable:              true,
						Model:                 getInitialDeviceList(),
						CurrentIndex:          0,
						OnCurrentIndexChanged: onDeviceSelectionChanged,
						OnTextChanged:         onDeviceTextChanged,
						MinSize:               d.Size{Width: 150, Height: 0},
					},

					d.Label{Text: "Настройки:"},
					d.ComboBox{
						AssignTo:      &paramInput,
						Editable:      true,
						Model:         []string{"9600", "115200", "8200"},
						Value:         defaultBaud,
						OnTextChanged: updateUIState, // Проверяем валидность порта при вводе
						MinSize:       d.Size{Width: 70, Height: 0},
					},

					d.PushButton{
						AssignTo:  &actionBtn,
						Text:      "Подключить",
						OnClicked: onActionBtnClicked,
						MinSize:   d.Size{Width: 80},
					},
				},
			},

			// --- Вкладки ---
			d.TabWidget{
				Pages: []d.TabPage{
					// 1. Информация
					{
						Title:  "Информация",
						Layout: d.VBox{Margins: d.Margins{Left: 6, Top: 6, Right: 6, Bottom: 6}, Spacing: 5},
						Children: []d.Widget{
							d.PushButton{Text: "Обновить данные", OnClicked: refreshInfo},
							// Текстовое поле вместо таблицы
							d.TextEdit{
								AssignTo: &infoView,
								ReadOnly: true,
								VScroll:  true,
								Font:     d.Font{Family: "Consolas", PointSize: 9}, // Моноширинный шрифт
								MinSize:  d.Size{Width: 100, Height: 150},
							},
							// Панель операционных кнопок
							d.Composite{
								Layout: d.HBox{Alignment: d.AlignHCenterVCenter},
								Children: []d.Widget{
									d.Composite{
										Layout: d.Grid{Columns: 2, Spacing: 10},
										Children: []d.Widget{
											d.PushButton{Text: "X-Отчет", OnClicked: onPrintX, MinSize: d.Size{Width: 160}},
											d.PushButton{Text: "Копия документа", OnClicked: onPrintCopy, MinSize: d.Size{Width: 160}},
											d.PushButton{Text: "Z-Отчет (Закрыть смену)", OnClicked: onPrintZ, MinSize: d.Size{Width: 160}},
											d.PushButton{Text: "Прогон и отрезка", OnClicked: onFeedAndCut, MinSize: d.Size{Width: 160}},
										},
									},
								},
							},
							d.VSpacer{}, // Прижимаем контент к верху
						},
					},
					// 2. Регистрация
					GetRegistrationTab(),
					// 3. Сервис
					// GetServiceTab(),
				},
			},

			d.VSpacer{},

			// --- Лог ---
			d.GroupBox{
				Title:   "Лог",
				Layout:  d.VBox{MarginsZero: true},
				MinSize: d.Size{Height: 200},
				MaxSize: d.Size{Height: 200},
				Children: []d.Widget{
					d.TextEdit{
						AssignTo: &logView,
						ReadOnly: true,
						VScroll:  true,
						HScroll:  true,
					},
				},
			},
		},
	}).Run(); err != nil {
		log.Fatal(err)
	}
}

// --- Логика UI ---

// getInitialDeviceList формирует список: COM порты + пункт поиска
func getInitialDeviceList() []string {
	ports, _ := serial.GetPortsList()
	sort.Strings(ports)
	// Всегда добавляем пункт для LAN в конец
	ports = append(ports, itemSearchLAN)
	return ports
}

// onDeviceSelectionChanged вызывается при выборе из выпадающего списка
func onDeviceSelectionChanged() {
	if driver != nil {
		return
	}

	idx := addrCombo.CurrentIndex()
	if idx < 0 {
		return
	}

	// Получаем реальный выбранный текст из модели по индексу,
	// так как addrCombo.Text() может быть еще не обновлен системой
	model, ok := addrCombo.Model().([]string)
	if !ok || idx >= len(model) {
		return
	}
	selected := model[idx]

	// Явно обновляем текст в поле, чтобы updateUIState увидел актуальное значение
	if addrCombo.Text() != selected {
		addrCombo.SetText(selected)
	}

	if selected == itemSearchLAN {
		// Режим поиска
		paramInput.SetText(defaultTCPPort)
	} else if strings.HasPrefix(selected, "COM") {
		// Режим COM
		paramInput.SetText(defaultBaud)
	} else {
		// Режим IP (из истории)
		paramInput.SetText(defaultTCPPort)
	}

	updateUIState()
}

// onDeviceTextChanged отслеживает ручной ввод
func onDeviceTextChanged() {
	updateUIState()
}

// updateUIState - центральная логика состояния кнопки и валидации
func updateUIState() {
	if driver != nil {
		actionBtn.SetText("Отключить")
		actionBtn.SetEnabled(true)
		return
	}

	text := strings.TrimSpace(addrCombo.Text())
	portText := strings.TrimSpace(paramInput.Text())

	// 1. Режим ПОИСКА
	// Срабатывает, если поле пустое ИЛИ если в нем текст пункта меню "Поиск..."
	if text == "" || text == itemSearchLAN {
		actionBtn.SetText("Искать")
		actionBtn.SetEnabled(true)

		// Подставляем порт поиска, если там сейчас скорость COM
		if portText == "9600" || portText == "115200" {
			paramInput.SetText(defaultTCPPort)
		}
		return
	}

	// 2. Режим COM-порта
	if strings.HasPrefix(strings.ToUpper(text), "COM") {
		actionBtn.SetText("Подключить")
		actionBtn.SetEnabled(len(text) > 3 && portText != "")
		return
	}

	// 3. Режим IP или Домена
	// Если текст не пустой, не "Поиск..." и не "COM..." -> значит это ввод адреса
	actionBtn.SetText("Подключить")

	// Проверка на доменное имя (наличие букв)
	isDomain := false
	if match, _ := regexp.MatchString(`[a-zA-Z]`, text); match {
		isDomain = true
	}

	// Если это домен и порт похож на скорость COM - очищаем, просим ввести порт
	if isDomain && (portText == defaultBaud || portText == "9600") {
		paramInput.SetText("")
	}

	// Валидация порта: должен быть числом
	portValid := false
	if _, err := strconv.Atoi(portText); err == nil {
		portValid = true
	}

	actionBtn.SetEnabled(portValid)
}

func onActionBtnClicked() {
	// 1. Сценарий отключения
	if driver != nil {
		if err := driver.Disconnect(); err != nil {
			logMsg("Ошибка отключения: %v", err)
		}
		driver = nil

		addrCombo.SetEnabled(true)
		paramInput.SetEnabled(true)
		updateUIState()
		logMsg("Отключено.")
		return
	}

	currentText := strings.TrimSpace(addrCombo.Text())

	// 2. Сценарий Поиска
	if actionBtn.Text() == "Искать" {
		go runNetworkScan()
		return
	}

	// 3. Сценарий Подключения
	cfg := mitsu.Config{
		Timeout: 3000,
		Logger:  func(s string) { logMsg(s) },
	}

	// Определяем тип подключения
	if strings.HasPrefix(strings.ToUpper(currentText), "COM") {
		cfg.ConnectionType = 0
		cfg.ComName = currentText
		fmt.Sscanf(paramInput.Text(), "%d", &cfg.BaudRate)
	} else {
		// IP или Домен
		cfg.ConnectionType = 6
		cfg.IPAddress = currentText
		fmt.Sscanf(paramInput.Text(), "%d", &cfg.TCPPort)
	}

	logMsg("Подключение к %s...", getConnString(&cfg))
	setControlsEnabled(false)

	go func() {
		drv := mitsu.New(cfg)
		if err := drv.Connect(); err != nil {
			mw.Synchronize(func() {
				logMsg("ОШИБКА: %v", err)
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Не удалось подключиться: %v", err), walk.MsgBoxIconError)
				setControlsEnabled(true)
				updateUIState()
			})
			return
		}

		// Успех
		mw.Synchronize(func() {
			driver = drv
			updateUIState()
			logMsg("Успешное подключение!")
			refreshInfo()
		})
	}()
}

func setControlsEnabled(enabled bool) {
	addrCombo.SetEnabled(enabled)
	paramInput.SetEnabled(enabled)
	actionBtn.SetEnabled(enabled)
}

func getConnString(c *mitsu.Config) string {
	if c.ConnectionType == 0 {
		return fmt.Sprintf("%s:%d", c.ComName, c.BaudRate)
	}
	return fmt.Sprintf("%s:%d", c.IPAddress, c.TCPPort)
}

// --- Сканер сети (ARP based) ---

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
		foundList = append(foundList, ip)
	}

	mw.Synchronize(func() {
		if len(foundList) > 0 {
			newList := getInitialDeviceList()
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

// --- Функции обновления данных и утилиты ---

func refreshInfo() {
	if driver == nil {
		return
	}
	// Очищаем поле перед загрузкой (опционально, можно и не очищать)
	mw.Synchronize(func() { infoView.SetText("Загрузка данных...") })

	go func() {
		info, err := driver.GetFiscalInfo()
		if err != nil {
			mw.Synchronize(func() {
				infoView.SetText(fmt.Sprintf("ОШИБКА ПОЛУЧЕНИЯ ДАННЫХ:\r\n%v", err))
			})
			return
		}

		// Сбор данных в мапу для последующего форматирования
		// Мы используем слайс структур для сохранения порядка
		type kv struct {
			k, v string
		}
		var lines []kv

		lines = append(lines, kv{"Модель ККТ", info.ModelName})
		lines = append(lines, kv{"Заводской номер", info.SerialNumber})
		lines = append(lines, kv{"Версия прошивки", info.SoftwareDate})
		lines = append(lines, kv{"РНМ", info.RNM})
		lines = append(lines, kv{"ИНН организации", info.Inn})
		lines = append(lines, kv{"Организация", info.OrganizationName})
		lines = append(lines, kv{"ОФД", info.OfdName})
		lines = append(lines, kv{"Дата регистрации", info.RegistrationDate})
		lines = append(lines, kv{"Версия ФФД", info.FfdVersion})
		lines = append(lines, kv{"Срок действия ФН", info.FnEndDate})
		lines = append(lines, kv{"Исполнение ФН", info.FnEdition})

		sh, err := driver.GetShiftStatus()
		if err == nil {
			st := "Закрыта"
			if sh.State == "1" {
				st = "Открыта"
			}
			lines = append(lines, kv{"Смена", fmt.Sprintf("№%d (%s)", sh.ShiftNum, st)})

			// Логика отображения неотправленных документов
			ofdInfo := fmt.Sprintf("%d", sh.Ofd.Count)
			if sh.Ofd.Count > 0 {
				ofdInfo += fmt.Sprintf(" (Первый: №%d от %s %s)", sh.Ofd.First, sh.Ofd.Date, sh.Ofd.Time)
			}
			lines = append(lines, kv{"Неотправленных ФД", ofdInfo})

		} else {
			lines = append(lines, kv{"Смена", "Ошибка получения статуса"})
		}

		// Формирование текста с выравниванием
		var sb strings.Builder
		maxKeyLen := 0
		for _, item := range lines {
			if len(item.k) > maxKeyLen {
				maxKeyLen = len(item.k)
			}
		}
		// Добавим немного отступа
		maxKeyLen += 2

		for _, item := range lines {
			// %-20s выравнивает по левому краю, добавляя пробелы справа
			format := fmt.Sprintf("%%-%ds : %%s\r\n", maxKeyLen)
			sb.WriteString(fmt.Sprintf(format, item.k, item.v))
		}

		finalText := sb.String()

		mw.Synchronize(func() {
			infoView.SetText(finalText)
		})
	}()
}

// --- Обработчики операционных кнопок ---

func onPrintX() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}
	go func() {
		if err := driver.PrintXReport(); err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка печати", err.Error(), walk.MsgBoxIconError) })
		} else {
			logMsg("X-отчет распечатан успешно.")
		}
	}()
}

func onPrintCopy() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}
	go func() {
		if err := driver.PrintLastDocument(); err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка печати", err.Error(), walk.MsgBoxIconError) })
		} else {
			logMsg("Копия документа распечатана.")
		}
	}()
}

func onPrintZ() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	if walk.MsgBox(mw, "Подтверждение", "Вы действительно хотите закрыть смену (Z-отчет)?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
		return
	}

	go func() {
		// Используем имя "Системный администратор" или пустую строку, если драйвер это позволяет
		if err := driver.CloseShift("Администратор"); err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка закрытия смены", err.Error(), walk.MsgBoxIconError) })
			return
		}
		logMsg("Смена закрыта успешно. Печать отчета...")

		// Автоматическая печать после успешного закрытия
		time.Sleep(500 * time.Millisecond) // Небольшая пауза для надежности
		if err := driver.PrintLastDocument(); err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка печати Z-отчета", err.Error(), walk.MsgBoxIconWarning)
			})
		}
		refreshInfo() // Обновляем статус смены на экране
	}()
}

func onFeedAndCut() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}
	go func() {
		if err := driver.Feed(24); err != nil {
			logMsg("Ошибка прогона бумаги: %v", err)
			return
		}
		if err := driver.Cut(); err != nil {
			logMsg("Ошибка отрезки: %v", err)
			return
		}
		logMsg("Прогон и отрезка выполнены.")
	}()
}

func logMsg(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fullMsg := fmt.Sprintf("[%s] %s\r\n", time.Now().Format("15:04:05.000"), msg)
	if mw != nil {
		mw.Synchronize(func() { logView.AppendText(fullMsg) })
	} else {
		log.Print(fullMsg)
	}
}
