package main

import (
	"context"
	"go.uber.org/zap"

	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"github.com/dkrasnykh/metrics-alerter/internal/config"
)

func main() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	cfg, err := config.NewAgentConfig()
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	a := agent.New(cfg)
	a.Run(context.Background())
}
