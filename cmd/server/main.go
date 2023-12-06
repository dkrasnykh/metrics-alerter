package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dkrasnykh/metrics-alerter/internal/server"
	"log"
)

type config struct {
	address string `env:"ADDRESS"`
}

func main() {
	var runAddr string
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	var c config
	err := env.Parse(&c)
	if err != nil {
		log.Fatal("error retrieving environment variables")
	}
	if c.address != "" {
		runAddr = c.address
	}
	s := server.NewServer(runAddr)
	s.Run()
}
