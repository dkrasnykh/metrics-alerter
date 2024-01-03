package memory

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type data struct {
	Metrics []models.Metrics `json:"metrics"`
}

func Load(path string) ([]models.Metrics, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading data from file %s: %w", path, err)
	}
	if len(bytes) == 0 {
		return nil, fmt.Errorf(`file %s is empty`, path)
	}

	v := data{}
	err = json.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	return v.Metrics, nil
}

func Save(path string, ms []models.Metrics) error {
	file, err := os.Create(path)

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}(file)

	if err != nil {
		return err
	}

	v := data{ms}
	bytes, err := json.Marshal(&v)
	if err != nil {
		return fmt.Errorf("error converting data to json %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing data into file %s; %w", path, err)
	}
	return nil
}

func InitDir(path string) string {
	err := os.MkdirAll(path+"/", 0777)
	if err != nil {
		logger.Error(err.Error())
	}
	return path + "/metrics.tmp"
}
