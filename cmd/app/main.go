package main

import (
	infraDriver "mitsuscanner/internal/infrastructure/driver"
	"mitsuscanner/internal/infrastructure/logger"
	"mitsuscanner/internal/infrastructure/storage"
	"mitsuscanner/internal/service/connection"
	registrationService "mitsuscanner/internal/service/registration"
	"mitsuscanner/internal/service/settings"
	timeService "mitsuscanner/internal/service/time"
	"mitsuscanner/internal/ui"
	"mitsuscanner/internal/ui/controller"
	"mitsuscanner/internal/ui/viewmodel"
	legacyDriver "mitsuscanner/pkg/mitsudriver"
)

func main() {
	// 1. Initialize logger (infrastructure)
	log := logger.NewStdLogger("MitsuScanner: ")
	log.Info("Application starting")

	// 2. Initialize profile repository (infrastructure)
	repo, err := storage.NewFileProfileRepository("profiles.json")
	if err != nil {
		log.Fatal("Failed to initialize profile repository: %v", err)
	}

	// 3. Create the driver adapter (infrastructure with dummy config)
	dummyConfig := legacyDriver.Config{
		ConnectionType: 0, // COM port (default)
		BaudRate:       115200,
		Timeout:        3000,
		Logger: func(msg string) {
			log.Debug(msg)
		},
	}
	legacyDrv := legacyDriver.NewMitsuDriver(dummyConfig)
	driverAdapter := infraDriver.NewMitsuDriverAdapter(legacyDrv)

	// 4. Create services (connection, settings, time, registration)
	connService := connection.NewConnectionService(driverAdapter, repo)
	settingsService := settings.NewSettingsService(driverAdapter)
	timeService := timeService.NewTimeService(driverAdapter)
	regService := registrationService.NewRegistrationService(driverAdapter)

	// 5. Create view models and controllers
	mainVM := viewmodel.NewMainViewModel()
	mainCtrl := controller.NewMainController(mainVM, connService)

	serviceVM := viewmodel.NewServiceViewModel()
	serviceCtrl := controller.NewServiceController(serviceVM, settingsService, timeService)

	regVM := viewmodel.NewRegistrationViewModel()
	regCtrl := controller.NewRegistrationController(regVM, regService)

	// 6. Run the GUI application
	log.Info("Initialization complete, starting GUI")
	if err := ui.Run(mainCtrl, serviceCtrl, regCtrl); err != nil {
		log.Fatal("GUI error: %v", err)
	}
}
