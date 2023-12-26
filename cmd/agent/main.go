package main

import (
	"log"

	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"github.com/dkrasnykh/metrics-alerter/internal/config"
)

func main() {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	a := agent.New(cfg.Address, cfg.PollInterval, cfg.ReportInterval)
	a.Run()
}
