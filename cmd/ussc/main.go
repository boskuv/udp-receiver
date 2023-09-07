package main

import (
	"log"
	"udp-receiver/internal/config"
	app "udp-receiver/internal/ussc"
)

func main() {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(cfg)
}
