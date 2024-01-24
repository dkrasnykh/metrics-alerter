package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/repository"
	"github.com/dkrasnykh/metrics-alerter/internal/utils"
)

var FilePath string

type data struct {
	Metrics []models.Metrics `json:"metrics"`
}

func Load(path string) ([]models.Metrics, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var bytes []byte
	bytes, err = os.ReadFile(path)
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
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		utils.LogError(err)
	}(file)
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
	utils.LogError(err)
	FilePath = path + "/metrics.tmp"
	return FilePath
}

func Restore(r repository.Storager) error {
	if FilePath == "" {
		return errors.New("the path is undefined")
	}
	data, err := Load(FilePath)
	if err != nil {
		return err
	}
	for _, m := range data {
		_, err = r.Create(context.Background(), m)
		if err != nil {
			return err
		}
	}
	return nil
}
