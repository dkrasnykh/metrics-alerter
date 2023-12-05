package main

import (
	"github.com/dkrasnykh/metrics-alerter/internal/agent"
	"time"
)

func main() {
	pollInterval := 2
	reportInterval := 10
	serverAddress := "localhost"
	serverPort := "8080"
	a := agent.NewAgent(serverAddress, serverPort,
		time.NewTicker(time.Duration(pollInterval)*time.Second),
		time.NewTicker(time.Duration(reportInterval)*time.Second))
	a.Run()
}
