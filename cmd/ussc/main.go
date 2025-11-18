package main

import (
	"log"
	"os"

	app "github.com/boskuv/udp-receiver/internal/ussc"

	"github.com/boskuv/udp-receiver/internal/config"
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
