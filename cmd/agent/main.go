package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"log"
	"time"
)

type config struct {
	address        string `env:"ADDRESS"`
	reportInterval int    `env:"REPORT_INTERVAL"`
	pollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	var runAddr string
	var reportInterval int
	var pollInterval int
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&pollInterval, "p", 2, "frequency of collecting metrics from runtime package")
	flag.Parse()

	var c config
	err := env.Parse(&c)
	if err != nil {
		log.Fatal("error retrieving environment variables")
	}
	if c.address != "" {
		runAddr = c.address
	}
	if c.reportInterval != 0 {
		reportInterval = c.reportInterval
	}
	if c.pollInterval != 0 {
		pollInterval = c.pollInterval
	}

	a := agent.NewAgent(runAddr,
		time.NewTicker(time.Duration(pollInterval)*time.Second),
		time.NewTicker(time.Duration(reportInterval)*time.Second))
	a.Run()
}
