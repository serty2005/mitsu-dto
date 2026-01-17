package ui

import (
	"mitsuscanner/internal/ui/controller"
	"mitsuscanner/internal/ui/view"
)

// Run запускает графическое приложение с переданными контроллерами.
func Run(mainCtrl *controller.MainController, serviceCtrl *controller.ServiceController, registrationCtrl *controller.RegistrationController) error {
	return view.Run(mainCtrl, serviceCtrl, registrationCtrl)
}
