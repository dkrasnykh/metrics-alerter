package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}

func (c *ServerConfig) Parse() error {
	var runAddr string
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
	err := env.Parse(c)
	if err != nil {
		return err
	}
	if c.Address != "" {
		runAddr = c.Address
	}
	return nil
}
