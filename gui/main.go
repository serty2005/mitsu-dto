package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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

	// Модели данных
	infoModel *KeyValueModel
	driver    mitsu.Driver
)

const (
	itemSearchLAN  = "Поиск в сети / Ввести IP..."
	defaultTCPPort = "8200"
	defaultBaud    = "115200"
)

func main() {
	infoModel = NewKeyValueModel()

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "Mitsu Driver Utility",
		MinSize:  Size{Width: 900, Height: 700},
		Layout:   VBox{},
		Children: []Widget{
			// --- Единая строка подключения ---
			GroupBox{
				Title:  "Подключение",
				Layout: Grid{Columns: 5},
				Children: []Widget{
					Label{Text: "Устройство:"},
					ComboBox{
						AssignTo: &addrCombo,
						// ВАЖНО: Всегда Editable=true, чтобы можно было ввести IP.
						// Если выбран COM порт, текст просто подставится сам.
						Editable:              true,
						Model:                 getInitialDeviceList(),
						CurrentIndex:          0,
						OnCurrentIndexChanged: onDeviceSelectionChanged,
						OnTextChanged:         onDeviceTextChanged,
						MinSize:               Size{Width: 250, Height: 0},
					},

					Label{Text: "Настройки (Baud/Port):"},
					ComboBox{
						AssignTo: &paramInput,
						Editable: true,
						Model:    []string{"9600", "115200", "8200"},
						Value:    defaultBaud,
						MinSize:  Size{Width: 100, Height: 0},
					},

					PushButton{
						AssignTo:  &actionBtn,
						Text:      "Подключить",
						OnClicked: onActionBtnClicked,
					},
				},
			},

			// --- Вкладки ---
			TabWidget{
				Pages: []TabPage{
					// 1. Информация
					{
						Title:  "Информация",
						Layout: VBox{},
						Children: []Widget{
							PushButton{Text: "Обновить данные", OnClicked: refreshInfo},
							TableView{
								Model: infoModel,
								Columns: []TableViewColumn{
									{Title: "Параметр", Width: 250},
									{Title: "Значение", Width: 500},
								},
							},
						},
					},
					// 2. Регистрация
					GetRegistrationTab(),
					// 3. Сервис
					GetServiceTab(),
				},
			},

			// --- Лог ---
			GroupBox{
				Title:  "Лог обмена",
				Layout: VBox{},
				Children: []Widget{
					TextEdit{
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
	// Если портов нет, добавляем заглушку, чтобы список не был пустым
	if len(ports) == 0 {
		ports = append(ports, "COM-порты не найдены")
	}
	// Всегда добавляем пункт для LAN
	ports = append(ports, itemSearchLAN)
	return ports
}

// onDeviceSelectionChanged вызывается при выборе из выпадающего списка
func onDeviceSelectionChanged() {
	if driver != nil {
		return
	} // Если подключены, игнорируем

	selected := addrCombo.Text()

	// Проверяем, выбрал ли пользователь пункт "Поиск..."
	// Важно: ComboBox.Text() возвращает текст выбранного элемента
	if selected == itemSearchLAN {
		// Режим LAN/Поиск
		// Очищаем поле, чтобы пользователь мог ввести IP или нажать Искать
		// Делаем это через SetText(""), но аккуратно, чтобы не вызвать рекурсию событий
		// В данном случае, просто сменим кнопку.
		addrCombo.SetText("")
		paramInput.SetText(defaultTCPPort)
		actionBtn.SetText("Искать")
	} else if strings.HasPrefix(selected, "COM") {
		// Режим COM
		paramInput.SetText(defaultBaud)
		actionBtn.SetText("Подключить")
	} else {
		// IP адрес из списка (если ранее нашли)
		paramInput.SetText(defaultTCPPort)
		actionBtn.SetText("Подключить")
	}
}

// onDeviceTextChanged отслеживает ручной ввод
func onDeviceTextChanged() {
	if driver != nil {
		return
	}

	text := addrCombo.Text()

	// Если текст пустой -> предлагаем Искать
	if strings.TrimSpace(text) == "" {
		actionBtn.SetText("Искать")
	} else {
		// Если что-то введено (IP или COM) -> предлагаем Подключить
		actionBtn.SetText("Подключить")
	}
}

func onActionBtnClicked() {
	// 1. Сценарий отключения
	if driver != nil {
		if err := driver.Disconnect(); err != nil {
			logMsg("Ошибка отключения: %v", err)
		}
		driver = nil
		actionBtn.SetText("Подключить")
		addrCombo.SetEnabled(true)
		paramInput.SetEnabled(true)

		// Сбрасываем UI в исходное состояние
		addrCombo.SetModel(getInitialDeviceList())
		addrCombo.SetCurrentIndex(0)

		logMsg("Отключено.")
		return
	}

	currentText := strings.TrimSpace(addrCombo.Text())

	// 2. Сценарий Поиска (только если поле пустое)
	// Или если явно выбран пункт меню (хотя он очищается при выборе)
	if actionBtn.Text() == "Искать" || currentText == itemSearchLAN {
		go runNetworkScan()
		return
	}

	// 3. Сценарий Подключения (COM или IP)
	cfg := mitsu.Config{
		Timeout: 3000,
		Logger:  func(s string) { logMsg(s) },
	}

	// Определяем тип подключения по формату строки
	if strings.HasPrefix(strings.ToUpper(currentText), "COM") {
		cfg.ConnectionType = 0
		cfg.ComName = currentText
		fmt.Sscanf(paramInput.Text(), "%d", &cfg.BaudRate)
	} else {
		// Считаем что это IP
		cfg.ConnectionType = 6
		cfg.IPAddress = currentText
		fmt.Sscanf(paramInput.Text(), "%d", &cfg.TCPPort)

		if cfg.IPAddress == "" {
			walk.MsgBox(mw, "Ошибка", "Введите IP адрес или нажмите Искать", walk.MsgBoxIconWarning)
			return
		}
	}

	logMsg("Подключение к %s...", getConnString(&cfg))
	setControlsEnabled(false) // Блокируем на время попытки

	// Запускаем подключение в горутине
	go func() {
		drv := mitsu.New(cfg)
		if err := drv.Connect(); err != nil {
			mw.Synchronize(func() {
				logMsg("ОШИБКА: %v", err)
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Не удалось подключиться: %v", err), walk.MsgBoxIconError)
				setControlsEnabled(true)
			})
			return
		}

		// Успех
		mw.Synchronize(func() {
			driver = drv
			actionBtn.SetText("Отключить")
			actionBtn.SetEnabled(true)
			logMsg("Успешное подключение!")
			refreshInfo() // Автообновление инфо
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

// --- Сканер сети ---

func runNetworkScan() {
	mw.Synchronize(func() {
		actionBtn.SetEnabled(false)
		actionBtn.SetText("Сканирование...")
		logMsg("--- Запуск сканирования сети (порт 8200) ---")
	})

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logMsg("Ошибка получения интерфейсов: %v", err)
		restoreBtnState()
		return
	}

	var targetIPs []string
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4()
				baseIP := ip.Mask(ipnet.Mask)
				for i := 1; i < 255; i++ {
					targetIP := net.IPv4(baseIP[0], baseIP[1], baseIP[2], byte(i))
					if !targetIP.Equal(ip) {
						targetIPs = append(targetIPs, targetIP.String())
					}
				}
			}
		}
	}

	foundChan := make(chan string, 10)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50)

	for _, ip := range targetIPs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:8200", ip), 200*time.Millisecond)
			if err != nil {
				return
			}
			conn.Close()

			// Проверка протокола
			cfg := mitsu.Config{
				ConnectionType: 6, IPAddress: ip, TCPPort: 8200, Timeout: 1000,
			}
			drv := mitsu.New(cfg)
			if err := drv.Connect(); err == nil {
				if _, _, mac, err := drv.GetVersion(); err == nil {
					// Фильтр MAC Mitsu
					if strings.HasPrefix(strings.ToUpper(mac), "00-22-00") {
						foundChan <- ip
					}
				}
				drv.Disconnect()
			}
		}(ip)
	}

	go func() {
		wg.Wait()
		close(foundChan)
	}()

	var foundList []string
	for ip := range foundChan {
		logMsg("НАЙДЕНО: %s", ip)
		foundList = append(foundList, ip)
	}

	mw.Synchronize(func() {
		if len(foundList) > 0 {
			// Обновляем список в комбобоксе
			newList := getInitialDeviceList()
			// Вставляем найденные IP перед пунктом "Поиск"
			searchItem := newList[len(newList)-1]
			newList = newList[:len(newList)-1]
			newList = append(newList, foundList...)
			newList = append(newList, searchItem)

			addrCombo.SetModel(newList)

			// Выбираем первый найденный IP
			addrCombo.SetText(foundList[0])
			actionBtn.SetText("Подключить")

			logMsg("Сканирование завершено. Найдено %d устр.", len(foundList))
		} else {
			logMsg("Сканирование завершено. Устройства не найдены.")
			walk.MsgBox(mw, "Результат", "Устройства Mitsu не найдены.", walk.MsgBoxIconInformation)

			// Если ничего не нашли, сбрасываем в дефолт
			addrCombo.SetText("")
			actionBtn.SetText("Искать")
		}
		actionBtn.SetEnabled(true)
	})
}

func restoreBtnState() {
	mw.Synchronize(func() {
		actionBtn.SetEnabled(true)
		actionBtn.SetText("Искать")
	})
}

// --- Функции обновления данных и утилиты ---

func refreshInfo() {
	if driver == nil {
		return
	}
	go func() {
		infoModel.Clear()
		info, err := driver.GetFiscalInfo()
		if err != nil {
			addResult("ОШИБКА", err.Error())
			return
		}
		addResult("Модель ККТ", info.ModelName)
		addResult("Заводской номер", info.SerialNumber)
		addResult("Версия прошивки", info.SoftwareDate)
		addResult("РНМ", info.RNM)
		addResult("ИНН Владельца", info.Inn)
		addResult("ОФД", info.OfdName)
		addResult("Дата регистрации", info.RegistrationDate)
		addResult("Версия ФФД", info.FfdVersion)

		sh, err := driver.GetShiftStatus()
		if err == nil {
			st := "Закрыта"
			if sh.State == "1" {
				st = "Открыта"
			}
			addResult("Смена", fmt.Sprintf("№%d (%s)", sh.ShiftNum, st))
		}
	}()
}

func syncTime() {
	if driver == nil {
		return
	}
	now := time.Now().UTC()
	if err := driver.SetDateTime(now); err != nil {
		logMsg("Ошибка времени: %v", err)
	} else {
		logMsg("Время установлено (UTC): %s", now.Format("15:04:05"))
	}
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

func addResult(key, val string) {
	mw.Synchronize(func() { infoModel.Add(key, val) })
}

// KeyValueModel
type KeyValueItem struct{ Key, Value string }
type KeyValueModel struct {
	walk.TableModelBase
	items []*KeyValueItem
}

func NewKeyValueModel() *KeyValueModel { return &KeyValueModel{items: []*KeyValueItem{}} }
func (m *KeyValueModel) RowCount() int { return len(m.items) }
func (m *KeyValueModel) Value(row, col int) interface{} {
	if col == 0 {
		return m.items[row].Key
	}
	return m.items[row].Value
}
func (m *KeyValueModel) Add(key, value string) {
	m.items = append(m.items, &KeyValueItem{Key: key, Value: value})
	m.PublishRowsReset()
}
func (m *KeyValueModel) Clear() {
	m.items = []*KeyValueItem{}
	m.PublishRowsReset()
}
