package main

import (
	"log"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/gui"
)

func main() {
	// Для отладки интерфейса с реальным подключением к ККТ используем MitsuDriver
	log.Printf("[DEBUG] Запуск в режиме реального подключения (MitsuDriver)...")

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
	driver.SetActive(realDriver)

	// Создаем структуру окна
	mw := new(walk.MainWindow)

	// ВАЖНО: Инициализируем глобальную переменную в пакете GUI,
	// чтобы loadServiceInitial не падал с nil pointer.
	gui.SetMainWindow(mw)

	// Получаем вкладку сервиса
	// Внутри GetServiceTab запустятся горутины, которые будут обращаться к gui.mw
	tab := gui.GetServiceTab()

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

	mw.Run()
}
