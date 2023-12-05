package main

import "github.com/dkrasnykh/metrics-alerter/internal/server"

func main() {
	address := "localhost"
	port := "8080"
	s := server.NewServer(address, port)
	s.Run()
}
