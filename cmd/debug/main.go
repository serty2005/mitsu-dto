package main

import (
	"log"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/gui"
	"mitsuscanner/internal/app"
)

func main() {
	// Для отладки интерфейса с реальным подключением к ККТ
	log.Printf("[DEBUG] Запуск в режиме реального подключения (MitsuDriver)...")

	// Инициализируем приложение (без загрузки файла профилей)
	appInstance, _ := app.NewApp("profiles_debug.json")

	// Инициализируем драйвер с дефолтными настройками (COM порт)
	config := driver.Config{
		ConnectionType: 0, // COM
		ComName:        "COM9",
		BaudRate:       115200,
		Timeout:        3000,
		Logger: func(msg string) {
			log.Printf("[DEBUG] %s", msg)
		},
	}
	realDriver := driver.NewMitsuDriver(config)

	// Устанавливаем драйвер в приложение
	appInstance.SetDriver(realDriver)

	// Создаем структуру окна
	// ВАЖНО: Не присваиваем appInstance.MainWindow здесь, так как окно еще не создано
	var mw *walk.MainWindow

	// ВАЖНО: Инициализируем глобальную переменную в пакете GUI (для legacy функций)
	// Для отладочного режима это нужно сделать, но так как mw еще nil,
	// это будет работать только если SetMainWindow вызывается позже или если мы передадим &mw (что невозможно для SetMainWindow)
	// В данном случае мы просто создаем ServiceTab, который будет ждать инициализации окна

	// Получаем вкладку сервиса
	serviceTab := gui.NewServiceTab(appInstance)
	tab := serviceTab.Create()

	err := d.MainWindow{
		AssignTo: &mw,
		Title:    "DEBUG SERVICE TAB (REAL MODE)",
		MinSize:  d.Size{Width: 900, Height: 600},
		Layout:   d.VBox{},
		Children: []d.Widget{
			d.TabWidget{
				Pages: []d.TabPage{
					tab,
				},
			},
		},
	}.Create()

	if err != nil {
		panic(err)
	}

	// ВАЖНО: Присваиваем созданное окно в appInstance
	appInstance.MainWindow = mw
	gui.SetMainWindow(mw)

	mw.Run()
}
