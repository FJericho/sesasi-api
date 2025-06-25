package main

import (
	"fmt"

	"github.com/FJericho/sesasi-api/internal/config"
)

func main() {
	viperConfig := config.NewViperConfig()
	app := config.NewFiberConfig(viperConfig)
	log := config.NewLogrusConfig(viperConfig)
	validate := config.NewValidatorConfig(viperConfig)
	db := config.NewDatabaseConfig(viperConfig, log)

	config.StartServer(&config.AppConfig{
		DB:       db,
		App:      app,
		Validate: validate,
		Config:   viperConfig,
		Log:      log,
	})

	webPort := viperConfig.GetInt("web.port")

	err := app.Listen(fmt.Sprintf(":%d", webPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
