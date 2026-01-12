package main

import (
	"log"
	"mitsuscanner/gui"
	"mitsuscanner/internal/app"
)

func main() {
	storagePath := "profiles.json"
	appInstance, err := app.NewApp(storagePath)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	if err := gui.RunApp(appInstance); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
