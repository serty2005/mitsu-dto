package main

import (
	"fmt"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/gui"
)

func main() {
	// Для отладки интерфейса БЕЗ кассы используем FakeDriver
	fmt.Println("Запуск в режиме эмуляции (FakeDriver)...")

	// Инициализируем и сразу активируем фейковый драйвер
	fake := driver.NewFakeDriver()
	driver.SetActive(fake)

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
		Title:    "DEBUG SERVICE TAB (MOCK MODE)",
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
