package main

import (
	"log"
	"os"

	"udp-receiver/internal/config"
	app "udp-receiver/internal/ussc"
)

func main() {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := app.Run(cfg); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}
