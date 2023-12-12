package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func NewAgentConfig() (*AgentConfig, error) {
	var runAddr string
	var reportInterval int
	var pollInterval int
	flag.StringVar(&runAddr, "a", ":8080", "address and port for server connection")
	flag.IntVar(&reportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&pollInterval, "p", 2, "frequency of collecting metrics from runtime package")
	flag.Parse()
	var c AgentConfig
	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}
	if c.Address == "" {
		c.Address = runAddr
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = reportInterval
	}
	if c.PollInterval == 0 {
		c.PollInterval = pollInterval
	}
	return &c, nil
}
