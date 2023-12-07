package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dkrasnykh/metrics-alerter/internal/server"
	"log"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func main() {
	var runAddr string
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	var c Config
	err := env.Parse(&c)
	if err != nil {
		log.Fatal("error retrieving environment variables")
	}
	
	if c.Address != "" {
		runAddr = c.Address
	}

	s := server.NewServer(runAddr)
	s.Run()
}
