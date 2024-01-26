package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func NewAgentConfig() (*AgentConfig, error) {
	var c AgentConfig
	flag.StringVar(&c.Address, "a", ":8080", "address and port for server connection")
	flag.IntVar(&c.ReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&c.PollInterval, "p", 2, "frequency of collecting metrics from runtime package")
	flag.StringVar(&c.Key, "k", "", "hashing key")
	flag.IntVar(&c.RateLimit, "l", 1, "rate limit")
	flag.Parse()

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
