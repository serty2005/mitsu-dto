package gui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
	"go.bug.st/serial"

	"mitsuscanner/driver"
	"mitsuscanner/internal/app"
	"mitsuscanner/internal/models"
)

// Global state
var (
	// mw остается глобальным только для вспомогательных функций (logMsg, и т.д.),
	// но контроллеры получают доступ через app.App.
	mw      *walk.MainWindow
	logView *walk.TextEdit
	mainApp *app.App

	// Элементы управления
	addrCombo        *walk.ComboBox   // Строка подключения (Умный комбобокс)
	actionBtn        *walk.PushButton // Кнопка действия (Искать/Подкл/Откл)
	clearProfilesBtn *walk.PushButton // Кнопка очистки профилей

	// Панель информации о ККТ (появляется после подключения)
	kktInfoComposite *walk.Composite // Контейнер для инфо
	modelLabel       *walk.Label     // Модель
	serialLabel      *walk.Label     // Серийный номер
	unsentDocsLabel  *walk.Label     // Неотправленные документы
	rebootIndicator  *walk.Label     // Индикатор перезагрузки (Цветная точка)

	// Элементы вкладки "Информация"
	infoView *walk.TextEdit // Текстовое поле для инфо
)

// SetMainWindow позволяет установить главное окно извне (для debug режима).
func SetMainWindow(w *walk.MainWindow) {
	mw = w
}

const (
	itemSearchLAN = "Поиск в сети / Ввести IP..."
	defaultPort   = 8200
	defaultBaud   = 115200
)

func RunApp(appInstance *app.App) error {
	mainApp = appInstance

	// Загружаем профили подключений перед формированием UI
	if err := appInstance.Storage.LoadProfiles(); err != nil {
		log.Printf("[GUI] Ошибка загрузки профилей при старте: %v", err)
	}

	// Инициализация контроллеров вкладок
	// ВАЖНО: MainWindow в appInstance еще nil, но контроллеры это учитывают в своих горутинах
	serviceTab := NewServiceTab(appInstance)
	regTab := NewRegistrationTab(appInstance)

	err := d.MainWindow{
		AssignTo: &mw, // Walk запишет сюда указатель на СОЗДАННОЕ окно
		Title:    "Mitsu Driver Utility",
		Size:     d.Size{Width: 600, Height: 600},
		MinSize:  d.Size{Width: 600, Height: 500},
		Layout:   d.VBox{MarginsZero: true, Spacing: 5},
		Children: []d.Widget{
			// --- Верхняя панель (Подключение + Инфо) ---
			d.GroupBox{
				Layout: d.HBox{Margins: d.Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}, Spacing: 10},
				Children: []d.Widget{

					// ЛЕВАЯ ЧАСТЬ: Управление подключением
					d.Composite{
						Layout: d.HBox{MarginsZero: true, Spacing: 5},
						Children: []d.Widget{
							d.ComboBox{
								AssignTo:              &addrCombo,
								Editable:              true,
								Model:                 getInitialDeviceList(appInstance),
								CurrentIndex:          0,
								OnCurrentIndexChanged: onDeviceSelectionChanged,
								OnTextChanged:         onDeviceTextChanged,
								MinSize:               d.Size{Width: 220, Height: 0},
								ToolTipText:           "Введите COMx:Baud или IP:Port. Примеры: COM9:115200, 192.168.1.50:8200",
							},
							d.PushButton{
								AssignTo:  &actionBtn,
								Text:      "Подключить",
								OnClicked: func() { onActionBtnClicked(appInstance) },
								MinSize:   d.Size{Width: 90},
							},
							d.PushButton{
								AssignTo:    &clearProfilesBtn,
								Text:        "🗑️",
								MaxSize:     d.Size{Width: 30},
								ToolTipText: "Очистить сохранённые профили",
								OnClicked:   func() { onClearProfiles(appInstance) },
							},
						},
					},

					// РАЗДЕЛИТЕЛЬ
					d.VSeparator{},

					// ПРАВАЯ ЧАСТЬ: Инфо о ККТ (Model, SN, Reboot status)
					d.Composite{
						AssignTo: &kktInfoComposite,
						Visible:  false, // Скрыт до подключения
						Layout:   d.HBox{MarginsZero: true, Spacing: 8, Alignment: d.AlignHNearVCenter},
						Children: []d.Widget{
							d.Label{AssignTo: &modelLabel, Text: "Mitsu", Font: d.Font{Bold: true}},
							d.Label{AssignTo: &serialLabel, Text: "SN: ..."},
							d.Label{AssignTo: &unsentDocsLabel, Text: "ОФД: 0"},
							d.Label{Text: "|"},
							d.Label{Text: "Статус:"},
							d.Label{
								AssignTo:    &rebootIndicator,
								Text:        "⦿", // Кружок
								Font:        d.Font{PointSize: 14, Bold: true},
								TextColor:   walk.RGB(0, 200, 0), // Зеленый
								ToolTipText: "Зеленый: Норма (Флаг=1)\nКрасный: Был сбой питания (Флаг=0)",
							},
						},
					},
					// Растяжка, чтобы прижать всё влево
					d.HSpacer{},
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
							d.TextEdit{
								AssignTo: &infoView,
								ReadOnly: true,
								VScroll:  true,
								Font:     d.Font{Family: "Consolas", PointSize: 9},
							},
							// Панель операционных кнопок
							d.Composite{
								Layout: d.HBox{Alignment: d.AlignHCenterVCenter},
								Children: []d.Widget{
									d.Composite{
										Layout: d.Grid{Columns: 4, Spacing: 10},
										Children: []d.Widget{
											d.PushButton{Text: "X-Отчет", OnClicked: onPrintX, MinSize: d.Size{Width: 120}},
											d.PushButton{Text: "Копия док.", OnClicked: onPrintCopy, MinSize: d.Size{Width: 120}},
											d.PushButton{Text: "Z-Отчет", OnClicked: onPrintZ, MinSize: d.Size{Width: 120}},
											d.PushButton{Text: "Перезагрузка", OnClicked: onRebootDeviceMain, MinSize: d.Size{Width: 120}},
										},
									},
								},
							},
						},
					},
					// 2. Регистрация
					regTab.Create(),
					// 3. Сервис
					serviceTab.Create(),
				},
			},

			// --- Лог ---
			d.GroupBox{
				Title:   "Лог",
				Layout:  d.VBox{MarginsZero: true},
				MinSize: d.Size{Height: 150},
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
	}.Create()

	if err != nil {
		return err
	}

	// КРИТИЧНО: Присваиваем созданное окно в appInstance ТОЛЬКО СЕЙЧАС
	appInstance.MainWindow = mw

	// Автовыбор первого профиля при старте
	if addrCombo.Model() != nil {
		onDeviceSelectionChanged()
	}

	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		drv := mainApp.GetDriver()
		if drv != nil {
			_ = drv.Disconnect()
			mainApp.SetDriver(nil)
		}
	})

	mw.Run()
	return nil
}

// --- Логика UI ---

// getInitialDeviceList формирует список
func getInitialDeviceList(appInstance *app.App) []string {
	var items []string
	var profiles []*models.ConnectionProfile

	// 1. Профили
	if appInstance != nil && appInstance.Storage != nil {
		profiles = GetProfiles(appInstance.Storage)
		for _, p := range profiles {
			items = append(items, p.DisplayString())
		}
	}

	// 2. COM-порты (чистые)
	ports, _ := serial.GetPortsList()
	sort.Strings(ports)
	for _, port := range ports {
		if !isPortInProfiles(port, profiles) {
			items = append(items, port) // Просто COMx, без скорости
		}
	}

	// 3. Поиск
	items = append(items, itemSearchLAN)

	return items
}

func isPortInProfiles(port string, profiles []*models.ConnectionProfile) bool {
	for _, p := range profiles {
		if p.ConnectionType == 0 && p.ComName == port {
			return true
		}
	}
	return false
}

func refreshDeviceList(appInstance *app.App) {
	mw.Synchronize(func() {
		addrCombo.SetModel(getInitialDeviceList(appInstance))
		if addrCombo.CurrentIndex() < 0 && len(getInitialDeviceList(appInstance)) > 0 {
			addrCombo.SetCurrentIndex(0)
		}
	})
}

// onConnectSuccess - действия после успешного соединения
func onConnectSuccess(drv driver.Driver, cfg driver.Config, appInstance *app.App) {
	logMsg("[SYSTEM] Подключение установлено. Чтение информации...")

	// 1. Читаем статику (Модель, Версия, SN)
	model, _ := drv.GetModel()
	ver, serial, _, _ := drv.GetVersion()
	shiftStatus, _ := drv.GetShiftStatus()

	unsent := 0
	if shiftStatus != nil {
		unsent = shiftStatus.Ofd.Count
	}

	logMsg("[INFO] %s, SN: %s, FW: %s", model, serial, ver)

	// 2. Сохраняем профиль
	profile := &models.ConnectionProfile{
		SerialNumber:   serial,
		ConnectionType: int(cfg.ConnectionType),
		ComName:        cfg.ComName,
		BaudRate:       int(cfg.BaudRate),
		IPAddress:      cfg.IPAddress,
		TCPPort:        int(cfg.TCPPort),
		FirmwareVer:    ver,
		ModelName:      model,
		LastUsed:       time.Now(),
	}
	go func() {
		UpsertProfile(appInstance.Storage, profile)
		mw.Synchronize(func() { refreshDeviceList(appInstance) })
	}()

	// 3. УСТАНОВКА ФЛАГА ПИТАНИЯ
	if err := drv.SetPowerFlag(1); err != nil {
		logMsg("[WARN] Не удалось установить флаг питания: %v", err)
	}

	// 4. Запускаем мониторинг
	StartMonitor(drv, model, serial, unsent)
	SetUpdateCallback(updateKktInfoPanel)

	// 5. Показываем панель
	mw.Synchronize(func() {
		modelLabel.SetText(model)
		serialLabel.SetText("SN: " + serial)
		unsentDocsLabel.SetText(fmt.Sprintf("ОФД: %d", unsent))
		rebootIndicator.SetTextColor(walk.RGB(0, 200, 0))
		kktInfoComposite.SetVisible(true)
	})
}

func updateKktInfoPanel(status *models.KktStatus) {
	mw.Synchronize(func() {
		if status.PowerFlag {
			rebootIndicator.SetText("⦿")
			rebootIndicator.SetTextColor(walk.RGB(0, 200, 0))
			rebootIndicator.SetToolTipText("Питание в норме")
		} else {
			rebootIndicator.SetText("○")
			rebootIndicator.SetTextColor(walk.RGB(255, 0, 0))
			rebootIndicator.SetToolTipText("ВНИМАНИЕ: Произошла перезагрузка ККТ!")
		}
	})
}

func onDeviceSelectionChanged() {
	if mainApp != nil && mainApp.GetDriver() != nil {
		return
	}
	updateUIState()
}

func onDeviceTextChanged() {
	updateUIState()
}

func onClearProfiles(appInstance *app.App) {
	if walk.MsgBox(mw, "Подтверждение", "Очистить все профили?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
		return
	}

	actionBtn.SetEnabled(false)
	go func() {
		err := ClearProfiles(appInstance.Storage)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError)
			} else {
				logMsg("Профили очищены.")
				refreshDeviceList(appInstance)
			}
			updateUIState()
		})
	}()
}

func updateUIState() {
	if mainApp != nil && mainApp.GetDriver() != nil {
		actionBtn.SetText("Отключить")
		actionBtn.SetEnabled(true)
		addrCombo.SetEnabled(false)
		return
	}

	addrCombo.SetEnabled(true)
	text := strings.TrimSpace(addrCombo.Text())

	if text == "" || text == itemSearchLAN {
		actionBtn.SetText("Искать")
		actionBtn.SetEnabled(true)
		return
	}

	actionBtn.SetText("Подключить")
	actionBtn.SetEnabled(true)
}

func parseConnectionString(input string) (host string, port int, isCom bool) {
	input = strings.TrimSpace(input)
	isCom = strings.HasPrefix(strings.ToUpper(input), "COM")
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		host = parts[0]
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			}
		}
	} else {
		host = input
	}
	if port == 0 {
		if isCom {
			port = defaultBaud
		} else {
			port = defaultPort
		}
	}
	return host, port, isCom
}

func extractSNFromProfileString(s string) string {
	parts := strings.Split(s, " - ")
	if len(parts) > 0 {
		return strings.TrimPrefix(parts[0], "SN")
	}
	return ""
}

func onActionBtnClicked(appInstance *app.App) {
	if drv := appInstance.GetDriver(); drv != nil {
		_ = drv.Disconnect()
		appInstance.SetDriver(nil)
		StopMonitor()
		kktInfoComposite.SetVisible(false)
		updateUIState()
		logMsg("Отключено.")
		return
	}

	rawText := strings.TrimSpace(addrCombo.Text())

	if actionBtn.Text() == "Искать" {
		go runNetworkScan()
		return
	}

	cfg := driver.Config{
		Timeout: 3000,
		Logger:  func(s string) { logMsg(s) },
	}

	if strings.HasPrefix(rawText, "SN") {
		sn := extractSNFromProfileString(rawText)
		profile := FindProfile(appInstance.Storage, sn)
		if profile != nil {
			logMsg("Подключение по профилю: %s...", profile.SerialNumber)
			cfg.ConnectionType = int32(profile.ConnectionType)
			if cfg.ConnectionType == 0 {
				cfg.ComName = profile.ComName
				cfg.BaudRate = int32(profile.BaudRate)
			} else {
				cfg.IPAddress = profile.IPAddress
				cfg.TCPPort = int32(profile.TCPPort)
			}
		} else {
			logMsg("[WARN] Профиль не найден, пробуем парсить строку...")
			h, p, isCom := parseConnectionString(rawText)
			if isCom {
				cfg.ConnectionType = 0
				cfg.ComName = h
				cfg.BaudRate = int32(p)
			} else {
				cfg.ConnectionType = 6
				cfg.IPAddress = h
				cfg.TCPPort = int32(p)
			}
		}
	} else {
		h, p, isCom := parseConnectionString(rawText)
		if isCom {
			cfg.ConnectionType = 0
			cfg.ComName = h
			cfg.BaudRate = int32(p)
		} else {
			cfg.ConnectionType = 6
			cfg.IPAddress = h
			cfg.TCPPort = int32(p)
		}
	}

	logMsg("Соединение с %s...", getConnString(&cfg))
	setControlsEnabled(false)

	go func() {
		drv := driver.NewMitsuDriver(cfg)
		if err := drv.Connect(); err != nil {
			mw.Synchronize(func() {
				logMsg("ОШИБКА: %v", err)
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Не удалось подключиться: %v", err), walk.MsgBoxIconError)
				setControlsEnabled(true)
				updateUIState()
			})
			return
		}

		appInstance.SetDriver(drv)

		mw.Synchronize(func() {
			updateUIState()
		})

		onConnectSuccess(drv, cfg, appInstance)
		refreshInfo()
	}()
}

func setControlsEnabled(enabled bool) {
	addrCombo.SetEnabled(enabled)
	actionBtn.SetEnabled(enabled)
}

func getConnString(c *driver.Config) string {
	if c.ConnectionType == 0 {
		return fmt.Sprintf("%s:%d", c.ComName, c.BaudRate)
	}
	return fmt.Sprintf("%s:%d", c.IPAddress, c.TCPPort)
}

func refreshInfo() {
	if mainApp == nil {
		return
	}
	drv := mainApp.GetDriver()
	if drv == nil {
		return
	}
	mw.Synchronize(func() { infoView.SetText("Загрузка данных...") })

	go func() {
		info, err := drv.GetFiscalInfo()
		if err != nil {
			mw.Synchronize(func() {
				infoView.SetText(fmt.Sprintf("ОШИБКА ПОЛУЧЕНИЯ ДАННЫХ:\r\n%v", err))
			})
			return
		}

		type kv struct{ k, v string }
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

		sh, err := drv.GetShiftStatus()
		if err == nil {
			st := "Закрыта"
			if sh.State == "1" {
				st = "Открыта"
			}
			lines = append(lines, kv{"Смена", fmt.Sprintf("№%d (%s)", sh.ShiftNum, st)})
			ofdInfo := fmt.Sprintf("%d", sh.Ofd.Count)
			if sh.Ofd.Count > 0 {
				ofdInfo += fmt.Sprintf(" (Первый: №%d от %s %s)", sh.Ofd.First, sh.Ofd.Date, sh.Ofd.Time)
			}
			lines = append(lines, kv{"Неотправленных ФД", ofdInfo})
		} else {
			lines = append(lines, kv{"Смена", "Ошибка получения статуса"})
		}

		var sb strings.Builder
		maxKeyLen := 0
		for _, item := range lines {
			if len(item.k) > maxKeyLen {
				maxKeyLen = len(item.k)
			}
		}
		maxKeyLen += 2
		for _, item := range lines {
			format := fmt.Sprintf("%%-%ds : %%s\r\n", maxKeyLen)
			sb.WriteString(fmt.Sprintf(format, item.k, item.v))
		}
		mw.Synchronize(func() { infoView.SetText(sb.String()) })
	}()
}

func onPrintX() {
	if mainApp != nil {
		if drv := mainApp.GetDriver(); drv != nil {
			go func() {
				if err := drv.PrintXReport(); err != nil {
					logMsg("Error X: %v", err)
				}
			}()
		}
	}
}

func onPrintZ() {
	if mainApp == nil {
		return
	}
	drv := mainApp.GetDriver()
	if drv != nil {
		if walk.MsgBox(mw, "Подтверждение", "Закрыть смену?", walk.MsgBoxYesNo) == walk.DlgCmdYes {
			go func() {
				drv.CloseShift("Admin")
				time.Sleep(500 * time.Millisecond)
				drv.PrintLastDocument()
				refreshInfo()
			}()
		}
	}
}

func onPrintCopy() {
	if mainApp != nil {
		if drv := mainApp.GetDriver(); drv != nil {
			go drv.PrintLastDocument()
		}
	}
}

func onRebootDeviceMain() {
	if mainApp == nil {
		return
	}
	drv := mainApp.GetDriver()
	if drv == nil {
		return
	}
	go func() {
		drv.RebootDevice()
		mw.Synchronize(func() {
			walk.MsgBox(mw, "Инфо", "Команда перезагрузки отправлена", walk.MsgBoxIconInformation)
		})
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
