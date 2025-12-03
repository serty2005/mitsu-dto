package main

import (
	"fmt"
	"mitsuscanner/mitsu"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// -----------------------------
// МОДЕЛИ ТАБЛИЦ
// -----------------------------

type HeaderModel struct {
	walk.TableModelBase
	items []map[string]interface{}
}

func (m *HeaderModel) RowCount() int {
	return len(m.items)
}

func (m *HeaderModel) Value(row, col int) interface{} {
	return m.items[row]["Text"]
}

func (m *HeaderModel) SetValue(row, col int, value interface{}) error {
	m.items[row]["Text"] = value.(string)
	return nil
}

type FooterModel struct {
	walk.TableModelBase
	items []map[string]interface{}
}

func (m *FooterModel) RowCount() int {
	return len(m.items)
}

func (m *FooterModel) Value(row, col int) interface{} {
	return m.items[row]["Text"]
}

func (m *FooterModel) SetValue(row, col int, value interface{}) error {
	m.items[row]["Text"] = value.(string)
	return nil
}

// -----------------------------
// МОДЕЛЬ ВЬЮ
// -----------------------------

type ServiceViewModel struct {
	KktTime      string
	PcTime       string
	OfdAddr      string
	OfdPort      int
	OfdClient    string
	TimerFN      int
	TimerOFD     int
	OismAddr     string
	OismPort     int
	LanAddr      string
	LanPort      int
	LanMask      string
	LanDns       string
	LanGw        string
	ExchangeMode bool

	HeaderLines []map[string]interface{}
	FooterLines []map[string]interface{}

	HeaderModel *HeaderModel
	FooterModel *FooterModel

	ExchangeLog string
}

var serviceModel *ServiceViewModel
var serviceBinder *walk.DataBinder

// Таймеры обновления времени
var pcTicker *time.Ticker
var kktTicker *time.Ticker

// -----------------------------
// ЗАГРУЗКА ДАННЫХ
// -----------------------------

func loadServiceInitial() {
	if driver == nil {
		serviceModel.KktTime = "Нет подключения"
		serviceModel.PcTime = time.Now().Format("15:04:05")
		return
	}

	// Загружаем время
	go func() {
		t, err := driver.GetDateTime()
		if err != nil {
			mw.Synchronize(func() {
				serviceModel.KktTime = "Ошибка"
			})
			return
		}

		mw.Synchronize(func() {
			serviceModel.KktTime = t.Format("15:04:05")
		})
	}()

	// Загружаем ОФД
	go func() {
		ofd, err := driver.GetOfdSettings()
		if err == nil {
			mw.Synchronize(func() {
				serviceModel.OfdAddr = ofd.Addr
				serviceModel.OfdPort = ofd.Port
				serviceModel.OfdClient = ofd.Client
				serviceModel.TimerFN = ofd.TimerFN
				serviceModel.TimerOFD = ofd.TimerOFD
			})
		}
	}()

	// Загружаем OISM
	go func() {
		oism, err := driver.GetOismSettings()
		if err == nil {
			mw.Synchronize(func() {
				serviceModel.OismAddr = oism.Addr
				serviceModel.OismPort = oism.Port
			})
		}
	}()

	// Загружаем LAN
	go func() {
		lan, err := driver.GetLanSettings()
		if err == nil {
			mw.Synchronize(func() {
				serviceModel.LanAddr = lan.Addr
				serviceModel.LanPort = lan.Port
				serviceModel.LanMask = lan.Mask
				serviceModel.LanDns = lan.Dns
				serviceModel.LanGw = lan.Gw
			})
		}
	}()

	// Загружаем клише
	go func() {
		h, _ := driver.GetHeader(1)
		f, _ := driver.GetHeader(3)

		mw.Synchronize(func() {
			for i := range serviceModel.HeaderLines {
				if i < len(h) {
					serviceModel.HeaderLines[i]["Text"] = h[i]
				}
			}
			serviceModel.HeaderModel.items = serviceModel.HeaderLines
			serviceModel.HeaderModel.PublishRowsReset()

			for i := range serviceModel.FooterLines {
				if i < len(f) {
					serviceModel.FooterLines[i]["Text"] = f[i]
				}
			}
			serviceModel.FooterModel.items = serviceModel.FooterLines
			serviceModel.FooterModel.PublishRowsReset()

			serviceBinder.Reset()
		})
	}()
}

// -----------------------------
// ЗАПУСК ТИКЕРОВ
// -----------------------------

func startClocks() {
	// ПК
	if pcTicker == nil {
		pcTicker = time.NewTicker(time.Second)
		go func() {
			for range pcTicker.C {
				mw.Synchronize(func() {
					serviceModel.PcTime = time.Now().Format("15:04:05")
					serviceBinder.Reset()
				})
			}
		}()
	}

	// ККТ: увеличиваем время на секунду
	if kktTicker == nil {
		kktTicker = time.NewTicker(time.Second)
		go func() {
			for range kktTicker.C {
				mw.Synchronize(func() {
					if serviceModel.KktTime == "" || len(serviceModel.KktTime) < 5 {
						return
					}
					t, err := time.Parse("15:04:05", serviceModel.KktTime)
					if err != nil {
						return
					}
					t = t.Add(time.Second)
					serviceModel.KktTime = t.Format("15:04:05")
					serviceBinder.Reset()
				})
			}
		}()
	}
}

// -----------------------------
// СОЗДАНИЕ ВКЛАДКИ
// -----------------------------

func GetServiceTab() TabPage {

	serviceModel = &ServiceViewModel{
		HeaderLines: make([]map[string]interface{}, 5),
		FooterLines: make([]map[string]interface{}, 5),
	}

	for i := range serviceModel.HeaderLines {
		serviceModel.HeaderLines[i] = map[string]interface{}{"Text": ""}
	}
	for i := range serviceModel.FooterLines {
		serviceModel.FooterLines[i] = map[string]interface{}{"Text": ""}
	}

	serviceModel.HeaderModel = &HeaderModel{items: serviceModel.HeaderLines}
	serviceModel.FooterModel = &FooterModel{items: serviceModel.FooterLines}

	loadServiceInitial()
	startClocks()

	return TabPage{
		Title:  "Сервис",
		Layout: VBox{Margins: Margins{Left: 6, Top: 6, Right: 6, Bottom: 6}},

		Children: []Widget{

			// ---------------------------------------------------------------------
			// 1. ВЕРХНЯЯ ПАНЕЛЬ ВРЕМЕНИ
			// ---------------------------------------------------------------------

			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 8},
				Children: []Widget{
					PushButton{
						Text:      "Запросить время",
						OnClicked: onQueryTime,
						MinSize:   Size{Width: 140},
					},
					Label{Text: "ККТ:"},
					Label{Text: Bind("KktTime"), MinSize: Size{Width: 70}},
					Label{Text: "ПК:"},
					Label{Text: Bind("PcTime"), MinSize: Size{Width: 70}},
					HSpacer{},
					PushButton{
						Text:      "Синхронизировать",
						OnClicked: onSyncTime,
						MinSize:   Size{Width: 150},
					},
				},
			},

			// ---------------------------------------------------------------------
			// 2. ПАРАМЕТРЫ ОФД / ОИСМ / LAN В ОДНОМ РЯДУ
			// ---------------------------------------------------------------------

			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{

					// ОФД
					GroupBox{
						Title:  "ОФД",
						Layout: Grid{Columns: 2, Spacing: 6},
						Children: []Widget{
							Label{Text: "Адрес:"},
							LineEdit{Text: Bind("OfdAddr"), MinSize: Size{Width: 120}},
							Label{Text: "Порт:"},
							NumberEdit{Value: Bind("OfdPort"), MinSize: Size{Width: 60}},
							Label{Text: "Клиент:"},
							LineEdit{Text: Bind("OfdClient"), MinSize: Size{Width: 120}},
							Label{Text: "Тайм-аут ФН:"},
							NumberEdit{Value: Bind("TimerFN"), MinSize: Size{Width: 60}},
							Label{Text: "Тайм-аут ОФД:"},
							NumberEdit{Value: Bind("TimerOFD"), MinSize: Size{Width: 60}},
							Composite{
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									PushButton{
										Text:      "Прочитать",
										OnClicked: onReadOfdSettings,
										MinSize:   Size{Width: 80},
									},
									PushButton{
										Text:      "Записать",
										OnClicked: onWriteOfdSettings,
										MinSize:   Size{Width: 80},
									},
								},
							},
							HSpacer{},
						},
					},

					// ОИСМ
					GroupBox{
						Title:  "ОИСМ",
						Layout: Grid{Columns: 2, Spacing: 6},
						Children: []Widget{
							Label{Text: "Адрес:"},
							LineEdit{Text: Bind("OismAddr"), MinSize: Size{Width: 120}},
							Label{Text: "Порт:"},
							NumberEdit{Value: Bind("OismPort"), MinSize: Size{Width: 60}},
							Composite{
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									PushButton{
										Text:      "Прочитать",
										OnClicked: onReadOismSettings,
										MinSize:   Size{Width: 80},
									},
									PushButton{
										Text:      "Записать",
										OnClicked: onWriteOismSettings,
										MinSize:   Size{Width: 80},
									},
								},
							},
							HSpacer{},
						},
					},

					// LAN
					GroupBox{
						Title:  "LAN",
						Layout: Grid{Columns: 2, Spacing: 6},
						Children: []Widget{
							Label{Text: "IP:"},
							LineEdit{Text: Bind("LanAddr"), MinSize: Size{Width: 120}},
							Label{Text: "Маска:"},
							LineEdit{Text: Bind("LanMask"), MinSize: Size{Width: 120}},
							Label{Text: "Шлюз:"},
							LineEdit{Text: Bind("LanGw"), MinSize: Size{Width: 120}},
							Label{Text: "DNS:"},
							LineEdit{Text: Bind("LanDns"), MinSize: Size{Width: 120}},
							Label{Text: "Порт:"},
							NumberEdit{Value: Bind("LanPort"), MinSize: Size{Width: 60}},
							Composite{
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									PushButton{
										Text:      "Прочитать",
										OnClicked: onReadLanSettings,
										MinSize:   Size{Width: 80},
									},
									PushButton{
										Text:      "Записать",
										OnClicked: onWriteLanSettings,
										MinSize:   Size{Width: 80},
									},
								},
							},
							HSpacer{},
						},
					},
				},
			},

			// ---------------------------------------------------------------------
			// 3. КЛИШЕ И ПОДВАЛ В ОДНОМ РЯДУ С КНОПКАМИ
			// ---------------------------------------------------------------------

			GroupBox{
				Title:  "Клише и Подвал",
				Layout: HBox{Spacing: 10},
				Children: []Widget{

					// Левая таблица - Заголовок
					Composite{
						Layout: VBox{},
						Children: []Widget{
							TableView{
								MinSize: Size{Width: 240, Height: 180},
								Columns: []TableViewColumn{{Title: "Заголовок"}},
								Model:   serviceModel.HeaderModel,
							},
							Composite{
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									PushButton{Text: "Прочитать", OnClicked: func() {
										onReadHeaderSingle(1)
									}},
									PushButton{Text: "Записать", OnClicked: func() {
										onWriteHeaderSingle(1)
									}},
								},
							},
						},
					},

					// Правая таблица - Подвал
					Composite{
						Layout: VBox{},
						Children: []Widget{
							TableView{
								MinSize: Size{Width: 240, Height: 180},
								Columns: []TableViewColumn{{Title: "Подвал"}},
								Model:   serviceModel.FooterModel,
							},
							Composite{
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									PushButton{Text: "Прочитать", OnClicked: func() {
										onReadHeaderSingle(3)
									}},
									PushButton{Text: "Записать", OnClicked: func() {
										onWriteHeaderSingle(3)
									}},
								},
							},
						},
					},
				},
			},

			// ---------------------------------------------------------------------
			// 4. УПРАВЛЕНИЕ УСТРОЙСТВОМ
			// ---------------------------------------------------------------------

			Composite{
				Layout: HBox{Spacing: 8},
				Children: []Widget{
					PushButton{Text: "Перезагрузка", OnClicked: onRebootDevice},
					PushButton{Text: "Тех. обнуление", OnClicked: onTechReset},
					PushButton{Text: "Открыть ящик", OnClicked: onOpenDrawer},
					PushButton{Text: "Печать X-отчёта", OnClicked: onPrintXReport},
				},
			},

			// ---------------------------------------------------------------------
			// 5. ЛОГ ОБМЕНА
			// ---------------------------------------------------------------------

			GroupBox{
				Title:  "Лог обмена",
				Layout: VBox{},
				Children: []Widget{
					TextEdit{
						Text:               Bind("ExchangeLog"),
						ReadOnly:           true,
						MinSize:            Size{Height: 150},
						VScroll:            true,
						AlwaysConsumeSpace: true,
					},
				},
			},
		},

		DataBinder: DataBinder{
			AssignTo:       &serviceBinder,
			DataSource:     serviceModel,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
	}
}

// -----------------------------
// ОБРАБОТЧИКИ КНОПОК
// -----------------------------

func onQueryTime() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		t, err := driver.GetDateTime()
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Не удалось получить время: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			serviceModel.KktTime = t.Format("15:04:05")
			serviceBinder.Reset()
		})
	}()
}

func onSyncTime() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	now := time.Now()

	go func() {
		err := driver.SetDateTime(now)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Не удалось синхронизировать: %v", err), walk.MsgBoxIconError)
				return
			}
			serviceModel.KktTime = now.Format("15:04:05")
			serviceBinder.Reset()
		})
	}()
}

// -----------------------------
// ОФД НАСТРОЙКИ
// -----------------------------

func onReadOfdSettings() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		ofd, err := driver.GetOfdSettings()
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка чтения ОФД: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			serviceModel.OfdAddr = ofd.Addr
			serviceModel.OfdPort = ofd.Port
			serviceModel.OfdClient = ofd.Client
			serviceModel.TimerFN = ofd.TimerFN
			serviceModel.TimerOFD = ofd.TimerOFD
			serviceBinder.Reset()
		})
	}()
}

func onWriteOfdSettings() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := driver.SetOfdSettings(mitsu.OfdSettings{
			Addr:     serviceModel.OfdAddr,
			Port:     serviceModel.OfdPort,
			Client:   serviceModel.OfdClient,
			TimerFN:  serviceModel.TimerFN,
			TimerOFD: serviceModel.TimerOFD,
		})

		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка записи ОФД: %v", err), walk.MsgBoxIconError)
				return
			}

			walk.MsgBox(mw, "Успех", "Настройки ОФД успешно записаны.", walk.MsgBoxIconInformation)
		})
	}()
}

// -----------------------------
// ОИСМ НАСТРОЙКИ
// -----------------------------

func onReadOismSettings() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		oism, err := driver.GetOismSettings()
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка чтения ОИСМ: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			serviceModel.OismAddr = oism.Addr
			serviceModel.OismPort = oism.Port
			serviceBinder.Reset()
		})
	}()
}

func onWriteOismSettings() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := driver.SetOismSettings(mitsu.ServerSettings{
			Addr: serviceModel.OismAddr,
			Port: serviceModel.OismPort,
		})

		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка записи ОИСМ: %v", err), walk.MsgBoxIconError)
				return
			}

			walk.MsgBox(mw, "Успех", "Настройки ОИСМ успешно записаны.", walk.MsgBoxIconInformation)
		})
	}()
}

// -----------------------------
// LAN НАСТРОЙКИ
// -----------------------------

func onReadLanSettings() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		lan, err := driver.GetLanSettings()
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка чтения LAN: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			serviceModel.LanAddr = lan.Addr
			serviceModel.LanPort = lan.Port
			serviceModel.LanMask = lan.Mask
			serviceModel.LanDns = lan.Dns
			serviceModel.LanGw = lan.Gw
			serviceBinder.Reset()
		})
	}()
}

func onWriteLanSettings() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := driver.SetLanSettings(mitsu.LanSettings{
			Addr: serviceModel.LanAddr,
			Port: serviceModel.LanPort,
			Mask: serviceModel.LanMask,
			Dns:  serviceModel.LanDns,
			Gw:   serviceModel.LanGw,
		})

		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка записи LAN: %v", err), walk.MsgBoxIconError)
				return
			}

			walk.MsgBox(mw, "Успех", "Настройки LAN успешно записаны.", walk.MsgBoxIconInformation)
		})
	}()
}

// -----------------------------
// УПРАВЛЕНИЕ УСТРОЙСТВОМ
// -----------------------------

func onRebootDevice() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := driver.RebootDevice()
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка перезагрузки: %v", err), walk.MsgBoxIconError)
				return
			}
		})
	}()
}

func onTechReset() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	if walk.MsgBox(mw, "Подтверждение", "Выполнить тех. обнуление?", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}

	go func() {
		err := driver.DeviceJob(1)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка: %v", err), walk.MsgBoxIconError)
				return
			}
			walk.MsgBox(mw, "Успех", "Тех. обнуление выполнено.", walk.MsgBoxIconInformation)
		})
	}()
}

func onOpenDrawer() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := driver.DeviceJob(2)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка: %v", err), walk.MsgBoxIconError)
			}
		})
	}()
}

func onPrintXReport() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := driver.PrintXReport()
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка печати: %v", err), walk.MsgBoxIconError)
			}
		})
	}()
}

// -----------------------------
// КЛИШЕ / ПОДВАЛ
// -----------------------------

func onReadHeaderSingle(mode int) {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		rows, err := driver.GetHeader(mode)
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка чтения: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			if mode == 1 {
				for i := range serviceModel.HeaderLines {
					if i < len(rows) {
						serviceModel.HeaderLines[i]["Text"] = rows[i]
					}
				}
				serviceModel.HeaderModel.PublishRowsReset()
			} else if mode == 3 {
				for i := range serviceModel.FooterLines {
					if i < len(rows) {
						serviceModel.FooterLines[i]["Text"] = rows[i]
					}
				}
				serviceModel.FooterModel.PublishRowsReset()
			}
			serviceBinder.Reset()
		})
	}()
}

func onWriteHeaderSingle(mode int) {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	var list []string

	if mode == 1 {
		for _, m := range serviceModel.HeaderLines {
			list = append(list, fmt.Sprintf("%v", m["Text"]))
		}
	} else {
		for _, m := range serviceModel.FooterLines {
			list = append(list, fmt.Sprintf("%v", m["Text"]))
		}
	}

	go func() {
		for i, text := range list {
			err := driver.SetHeaderLine(mode, i, text, "")
			if err != nil {
				mw.Synchronize(func() {
					walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка записи строки %d: %v", i, err), walk.MsgBoxIconError)
				})
				return
			}
		}
		mw.Synchronize(func() {
			walk.MsgBox(mw, "Успех", "Строки записаны.", walk.MsgBoxIconInformation)
		})
	}()
}
