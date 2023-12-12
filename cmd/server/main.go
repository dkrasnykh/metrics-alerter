package main

import (
	"log"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/server"
)

func main() {
	cfg := config.NewServerConfig()
	err := cfg.Parse()
	if err != nil {
		log.Fatal(err.Error())
	}
	s := server.NewServer(cfg.Address)
	err = s.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
