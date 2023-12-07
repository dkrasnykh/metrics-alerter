package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"log"
	"time"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	var runAddr string
	var reportInterval int
	var pollInterval int
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&pollInterval, "p", 2, "frequency of collecting metrics from runtime package")
	flag.Parse()

	var c Config
	err := env.Parse(&c)
	if err != nil {
		log.Fatal("error retrieving environment variables")
	}

	if c.Address != "" {
		runAddr = c.Address
	}
	if c.ReportInterval != 0 {
		reportInterval = c.ReportInterval
	}
	if c.PollInterval != 0 {
		pollInterval = c.PollInterval
	}

	a := agent.NewAgent(runAddr,
		time.NewTicker(time.Duration(pollInterval)*time.Second),
		time.NewTicker(time.Duration(reportInterval)*time.Second))
	a.Run()
}
