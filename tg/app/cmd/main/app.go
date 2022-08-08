package main

import (
	"app/internal"
	"app/internal/config"
	"app/pkg/logging"
	"log"
)

type app struct {
	cfg *config.Config
	log *logging.Logger
}

func main() {
	log.Println("application star")
	cfg := config.GetConfig()
	log.Println("logger init")
	logger := logging.GetLogger(cfg.Level)
	logger.Println("Creating Application")
	app, err := internal.NewApp(logger, cfg)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("Running Application")
	app.Run()
}
