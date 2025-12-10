package main

import (
	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/gui"
)

func main() {
	driver.SetActive(driver.NewFakeDriver())

	mw := &walk.MainWindow{}
	tab := gui.GetServiceTab()

	err := d.MainWindow{
		AssignTo: &mw,
		Title:    "DEBUG SERVICE TAB",
		MinSize:  d.Size{900, 600},
		Layout:   d.VBox{},
		Children: []d.Widget{
			tab,
		},
	}.Create()

	if err != nil {
		panic(err)
	}

	mw.Run()
}
