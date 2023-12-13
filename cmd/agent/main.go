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
	a := agent.NewAgent(cfg.Address, cfg.PollInterval, cfg.ReportInterval)
	err = a.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
