package main

import (
	"go.uber.org/zap"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/server"
)

func main() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	cfg, err := config.NewServerConfig()
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	s := server.New(cfg)
	err = s.Run()
	if err != nil {
		zap.L().Fatal(err.Error())
	}
}
