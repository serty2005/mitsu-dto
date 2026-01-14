package gui

import (
	"fmt"
	"sort"

	"mitsuscanner/driver"
	"mitsuscanner/internal/service"

	"github.com/lxn/walk"
)

func ApplyChangesPipeline(changes []service.Change) {
	drv := driver.Active
	if drv == nil {
		return
	}

	// ЕСЛИ СПИСОК ИЗМЕНЕНИЙ ПУСТ (пользователь удалил всё в диалоге и нажал Применить)
	if len(changes) == 0 {
		mw.Synchronize(func() {
			// Восстанавливаем состояние из памяти
			restoreViewFromSnapshot(serviceModel, initialSnapshot)
			logMsg("[SYSTEM] Изменения отменены пользователем, возврат к исходным настройкам.")
		})
		return
	}

	// 1. Сортировка (Normal -> Cliche -> Network)
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Priority < changes[j].Priority
	})

	mw.Synchronize(func() {
		isLoadingData = true
		btnServiceAction.SetEnabled(false)
		btnServiceAction.SetText("Запись...")
	})

	go func() {
		var errCount int
		var needReboot bool

		for i, ch := range changes {
			mw.Synchronize(func() {
				btnServiceAction.SetText(fmt.Sprintf("Запись %d/%d...", i+1, len(changes)))
			})

			if err := ch.ApplyFunc(drv); err != nil {
				errCount++
				fmt.Printf("Error applying %s: %v\n", ch.ID, err)
			}

			if ch.Priority == service.PriorityNetwork {
				needReboot = true
			}
		}

		mw.Synchronize(func() {
			if errCount > 0 {
				walk.MsgBox(mw, "Завершено с ошибками", fmt.Sprintf("Не удалось применить %d настроек.\nСм. лог (консоль).", errCount), walk.MsgBoxIconWarning)
			} else {
				msg := "Настройки успешно сохранены."
				if needReboot {
					msg += "\n\nБыли изменены сетевые настройки. Устройство перезагружается..."
				}
				walk.MsgBox(mw, "Успех", msg, walk.MsgBoxIconInformation)
			}

			if needReboot {
				go drv.RebootDevice()
				onReadAllSettings()
			} else {
				onReadAllSettings()
			}
		})
	}()
}
