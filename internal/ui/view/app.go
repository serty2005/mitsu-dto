package view

import (
	"mitsuscanner/internal/ui/controller"
)

// Run запускает графическое приложение
func Run(mainController *controller.MainController, serviceController *controller.ServiceController, registrationController *controller.RegistrationController) error {
	// Создание основного окна
	mw := NewMainWindowView(mainController, serviceController, registrationController)

	// Создание и инициализация окна
	if err := mw.Create(); err != nil {
		return err
	}

	// Запуск главного цикла сообщений
	mw.Run()
	return nil
}
