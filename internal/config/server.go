package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

func NewServerConfig() (*ServerConfig, error) {
	var runAddr string
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
	var c ServerConfig
	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}
	if c.Address == "" {
		c.Address = runAddr
	}
	return &c, nil
}
