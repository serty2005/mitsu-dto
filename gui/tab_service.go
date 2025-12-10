package gui

import (
	"fmt"
	"mitsuscanner/driver"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
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
	drv := driver.Active
	if drv == nil {

		serviceModel.KktTime = "Нет подключения"
		serviceModel.PcTime = time.Now().Format("15:04:05")
		return
	}

	// Загружаем время
	go func() {
		t, err := drv.GetDateTime()
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
		ofd, err := drv.GetOfdSettings()
		if err == nil {
			mw.Synchronize(func() {
				if ofd != nil {
					serviceModel.OfdAddr = ofd.Addr
					serviceModel.OfdPort = ofd.Port
					serviceModel.OfdClient = ofd.Client
					serviceModel.TimerFN = ofd.TimerFN
					serviceModel.TimerOFD = ofd.TimerOFD
				}
			})
		}
	}()

	// Загружаем OISM
	go func() {
		oism, err := drv.GetOismSettings()
		if err == nil {
			mw.Synchronize(func() {
				if oism != nil {
					serviceModel.OismAddr = oism.Addr
					serviceModel.OismPort = oism.Port
				}
			})
		}
	}()

	// Загружаем LAN
	go func() {
		lan, err := drv.GetLanSettings()
		if err == nil {
			mw.Synchronize(func() {
				if lan != nil {
					serviceModel.LanAddr = lan.Addr
					serviceModel.LanPort = lan.Port
					serviceModel.LanMask = lan.Mask
					serviceModel.LanDns = lan.Dns
					serviceModel.LanGw = lan.Gw
				}
			})
		}
	}()

	// Загружаем клише
	go func() {
		h, _ := drv.GetHeader(1)
		f, _ := drv.GetHeader(3)

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

func GetServiceTab() d.TabPage {

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

	return d.TabPage{
		Title:  "Сервис",
		Layout: d.VBox{Margins: d.Margins{Left: 6, Top: 6, Right: 6, Bottom: 6}},

		Children: []d.Widget{

			// ---------------------------------------------------------------------
			// 1. ВЕРХНЯЯ ПАНЕЛЬ ВРЕМЕНИ И УПРАВЛЕНИЯ
			// ---------------------------------------------------------------------

			d.Composite{
				Layout: d.Grid{Columns: 2, Spacing: 8},
				Children: []d.Widget{
					d.Composite{
						Layout: d.HBox{Spacing: 8},
						Children: []d.Widget{
							d.PushButton{
								Text:      "Запросить время",
								OnClicked: onQueryTime,
								MinSize:   d.Size{Width: 140},
							},
							d.Label{Text: "ККТ:"},
							d.Label{Text: d.Bind("KktTime"), MinSize: d.Size{Width: 70}},
							d.Label{Text: "ПК:"},
							d.Label{Text: d.Bind("PcTime"), MinSize: d.Size{Width: 70}},
							d.CheckBox{Text: "Режим обмена", Checked: d.Bind("ExchangeMode")},
						},
					},
					d.Composite{
						Layout: d.HBox{Spacing: 8},
						Children: []d.Widget{
							d.PushButton{Text: "Перезагрузка", OnClicked: onRebootDevice},
							d.PushButton{Text: "Тех. обнуление", OnClicked: onTechReset},
							d.PushButton{Text: "Открыть ящик", OnClicked: onOpenDrawer},
							d.PushButton{Text: "Печать X-отчёта", OnClicked: onPrintXReport},
						},
					},
				},
			},

			// ---------------------------------------------------------------------
			// 2. ПАРАМЕТРЫ ОФД / ОИСМ / LAN В ОДНОМ РЯДУ
			// ---------------------------------------------------------------------

			d.Composite{
				Layout: d.HBox{Spacing: 10},
				Children: []d.Widget{

					// ОФД
					d.GroupBox{
						Title:  "ОФД",
						Layout: d.VBox{Spacing: 6},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "Адрес:"},
									d.LineEdit{Text: d.Bind("OfdAddr"), MinSize: d.Size{Width: 120}},
									d.Label{Text: "Порт:"},
									d.NumberEdit{Value: d.Bind("OfdPort"), MinSize: d.Size{Width: 60}},
								},
							},
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "Клиент:"},
									d.LineEdit{Text: d.Bind("OfdClient"), MinSize: d.Size{Width: 120}},
									d.Label{Text: "Тайм-аут ФН:"},
									d.NumberEdit{Value: d.Bind("TimerFN"), MinSize: d.Size{Width: 60}},
								},
							},
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "Тайм-аут ОФД:"},
									d.NumberEdit{Value: d.Bind("TimerOFD"), MinSize: d.Size{Width: 60}},
									d.HSpacer{},
									d.Composite{
										Layout: d.HBox{Spacing: 6},
										Children: []d.Widget{
											d.PushButton{
												Text:      "Прочитать",
												OnClicked: onReadOfdSettings,
												MinSize:   d.Size{Width: 80},
											},
											d.PushButton{
												Text:      "Записать",
												OnClicked: onWriteOfdSettings,
												MinSize:   d.Size{Width: 80},
											},
										},
									},
								},
							},
						},
					},

					// ОИСМ
					d.GroupBox{
						Title:  "ОИСМ",
						Layout: d.VBox{Spacing: 6},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "Адрес:"},
									d.LineEdit{Text: d.Bind("OismAddr"), MinSize: d.Size{Width: 120}},
									d.Label{Text: "Порт:"},
									d.NumberEdit{Value: d.Bind("OismPort"), MinSize: d.Size{Width: 60}},
									d.HSpacer{},
									d.Composite{
										Layout: d.HBox{Spacing: 6},
										Children: []d.Widget{
											d.PushButton{
												Text:      "Прочитать",
												OnClicked: onReadOismSettings,
												MinSize:   d.Size{Width: 80},
											},
											d.PushButton{
												Text:      "Записать",
												OnClicked: onWriteOismSettings,
												MinSize:   d.Size{Width: 80},
											},
										},
									},
								},
							},
						},
					},

					// LAN
					d.GroupBox{
						Title:  "LAN",
						Layout: d.VBox{Spacing: 6},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "IP:"},
									d.LineEdit{Text: d.Bind("LanAddr"), MinSize: d.Size{Width: 120}},
									d.Label{Text: "Маска:"},
									d.LineEdit{Text: d.Bind("LanMask"), MinSize: d.Size{Width: 120}},
								},
							},
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "Шлюз:"},
									d.LineEdit{Text: d.Bind("LanGw"), MinSize: d.Size{Width: 120}},
									d.Label{Text: "DNS:"},
									d.LineEdit{Text: d.Bind("LanDns"), MinSize: d.Size{Width: 120}},
								},
							},
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.Label{Text: "Порт:"},
									d.NumberEdit{Value: d.Bind("LanPort"), MinSize: d.Size{Width: 60}},
									d.HSpacer{},
									d.Composite{
										Layout: d.HBox{Spacing: 6},
										Children: []d.Widget{
											d.PushButton{
												Text:      "Прочитать",
												OnClicked: onReadLanSettings,
												MinSize:   d.Size{Width: 80},
											},
											d.PushButton{
												Text:      "Записать",
												OnClicked: onWriteLanSettings,
												MinSize:   d.Size{Width: 80},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			// ---------------------------------------------------------------------
			// 3. КЛИШЕ И ПОДВАЛ В ОДНОМ РЯДУ С КНОПКАМИ
			// ---------------------------------------------------------------------

			d.GroupBox{
				Title:  "Клише и Подвал",
				Layout: d.HBox{Spacing: 10},
				Children: []d.Widget{

					// Левая таблица - Заголовок
					d.Composite{
						Layout: d.VBox{},
						Children: []d.Widget{
							d.TableView{
								MinSize: d.Size{Width: 240, Height: 80},
								Columns: []d.TableViewColumn{{Title: "Заголовок"}},
								Model:   serviceModel.HeaderModel,
							},
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.PushButton{Text: "Прочитать", OnClicked: func() {
										onReadHeaderSingle(1)
									}},
									d.PushButton{Text: "Записать", OnClicked: func() {
										onWriteHeaderSingle(1)
									}},
								},
							},
						},
					},

					// Правая таблица - Подвал
					d.Composite{
						Layout: d.VBox{},
						Children: []d.Widget{
							d.TableView{
								MinSize: d.Size{Width: 240, Height: 80},
								Columns: []d.TableViewColumn{{Title: "Подвал"}},
								Model:   serviceModel.FooterModel,
							},
							d.Composite{
								Layout: d.HBox{Spacing: 6},
								Children: []d.Widget{
									d.PushButton{Text: "Прочитать", OnClicked: func() {
										onReadHeaderSingle(3)
									}},
									d.PushButton{Text: "Записать", OnClicked: func() {
										onWriteHeaderSingle(3)
									}},
								},
							},
						},
					},
				},
			},
		},

		DataBinder: d.DataBinder{
			AssignTo:       &serviceBinder,
			DataSource:     serviceModel,
			ErrorPresenter: d.ToolTipErrorPresenter{},
		},
	}
}

// -----------------------------
// ОБРАБОТЧИКИ КНОПОК
// -----------------------------

func onQueryTime() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		t, err := drv.GetDateTime()
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	now := time.Now()

	go func() {
		err := drv.SetDateTime(now)
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		ofd, err := drv.GetOfdSettings()
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := drv.SetOfdSettings(driver.OfdSettings{
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		oism, err := drv.GetOismSettings()
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := drv.SetOismSettings(driver.ServerSettings{
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		lan, err := drv.GetLanSettings()
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := drv.SetLanSettings(driver.LanSettings{
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := drv.RebootDevice()
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка перезагрузки: %v", err), walk.MsgBoxIconError)
				return
			}
		})
	}()
}

func onTechReset() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	if walk.MsgBox(mw, "Подтверждение", "Выполнить тех. обнуление?", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}

	go func() {
		err := drv.DeviceJob(1)
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := drv.DeviceJob(2)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка: %v", err), walk.MsgBoxIconError)
			}
		})
	}()
}

func onPrintXReport() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		err := drv.PrintXReport()
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
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		rows, err := drv.GetHeader(mode)
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
	drv := driver.Active
	if drv == nil {
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
			err := drv.SetHeaderLine(mode, i, text, "")
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
