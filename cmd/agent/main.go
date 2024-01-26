package main

import (
	"context"

	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/logger"
)

func main() {
	err := logger.InitLogger()
	if err != nil {
		panic(err)
	}
	cfg, err := config.NewAgentConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}
	a := agent.New(cfg)
	a.Run(context.Background())
}
