package main

import (
	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/server"
)

func main() {
	err := logger.InitLogger()
	if err != nil {
		panic(err)
	}
	cfg, err := config.NewServerConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}
	s := server.New(cfg)
	err = s.Run()
	if err != nil {
		logger.Fatal(err.Error())
	}
}
