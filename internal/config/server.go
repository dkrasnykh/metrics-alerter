package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func NewServerConfig() (*ServerConfig, error) {
	var c ServerConfig
	flag.StringVar(&c.Address, "a", ":8080", "address and port to run server")
	flag.IntVar(&c.StoreInterval, "i", 300, "time interval (sec) to backup server data")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "path to the file to backup data")
	flag.BoolVar(&c.Restore, "r", true, "flag to recover data from file")
	flag.StringVar(&c.DatabaseDSN, "d", "", "url for database connection")
	flag.StringVar(&c.Key, "k", "", "hashing key")
	flag.Parse()

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
