package main

import (
	"flag"
	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"time"
)

func main() {
	var flagRunAddr string
	var flagReportInterval int
	var flagPollInterval int
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of collecting metrics from runtime package")
	flag.Parse()
	a := agent.NewAgent(flagRunAddr,
		time.NewTicker(time.Duration(flagPollInterval)*time.Second),
		time.NewTicker(time.Duration(flagReportInterval)*time.Second))
	a.Run()
}
