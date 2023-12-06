package main

import "github.com/dkrasnykh/metrics-alerter/internal/server"

import (
	"flag"
)

func main() {
	var flagRunAddr string
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
	s := server.NewServer(flagRunAddr)
	s.Run()
}
