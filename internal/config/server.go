package config

import (
	"flag"
	"strconv"
	"time"

	"github.com/caarlos0/env/v10"
)

type EnvConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   string `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         string `env:"RESTORE"`
}

type ServerConfig struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func NewServerConfig() (*ServerConfig, error) {
	var runAddr string
	var runStoreInterval int
	var runFilePath string
	var runRestore bool
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&runStoreInterval, "i", 0, "time interval (sec) to backup server data")
	flag.StringVar(&runFilePath, "f", "/tmp/metrics-db.json", "path to the file to backup data")
	flag.BoolVar(&runRestore, "r", true, "flag to recover data from file")
	flag.Parse()
	var ec EnvConfig
	var c ServerConfig
	err := env.Parse(&ec)
	if err != nil {
		return nil, err
	}
	if ec.Address == "" {
		c.Address = runAddr
	} else {
		c.Address = ec.Address
	}
	if ec.StoreInterval == "" {
		c.StoreInterval = time.Second * time.Duration(runStoreInterval)
	} else {
		v, err := strconv.ParseInt(ec.StoreInterval, 10, 64)
		if err != nil {
			return nil, err
		}
		c.StoreInterval = time.Second * time.Duration(v)
	}
	if ec.FileStoragePath == "" {
		c.FileStoragePath = runFilePath
	} else {
		c.FileStoragePath = ec.FileStoragePath
	}
	if ec.Restore == "" {
		c.Restore = runRestore
	} else {
		v, err := strconv.ParseBool(ec.Restore)
		if err != nil {
			return nil, err
		}
		c.Restore = v
	}
	return &c, nil
}
