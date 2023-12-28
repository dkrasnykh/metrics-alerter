package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func Convert(mtype, mname, value string) Metrics {
	m := Metrics{MType: mtype, ID: mname}
	switch mtype {
	case CounterType:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			m.Delta = &delta
		}
	case GaugeType:
		gvalue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			m.Value = &gvalue
		}
	}
	return m
}

func GetMetric(mtype, name string, value float64, delta int64) Metrics {
	m := Metrics{MType: mtype, ID: name}
	switch mtype {
	case CounterType:
		m.Delta = &delta
	case GaugeType:
		m.Value = &value
	}
	return m
}

func ExtractBody(req *http.Request) (*Metrics, error) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, errors.New(`request body is empty`)
	}
	var m Metrics
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type data struct {
	Metrics []Metrics `json:"metrics"`
}

func Load(path string) ([]Metrics, error) {
	_, err := os.Stat(path)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return nil, err
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return nil, fmt.Errorf("can't read backup file %s: %w", path, err)
	}
	if len(bytes) == 0 {
		return nil, fmt.Errorf(`file %s is empty`, path)
	}
	v := data{}
	err = json.Unmarshal(bytes, &v)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return nil, err
	}

	return v.Metrics, nil
}

func Save(path string, ms []Metrics) error {
	file, err := os.Create(path)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
	}(file)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return err
	}
	v := data{ms}
	bytes, err := json.Marshal(&v)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return fmt.Errorf("can't unmarshal data for backup %w", err)
	}
	_, err = file.Write(bytes)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return fmt.Errorf("can't write file backup %w", err)
	}
	return nil
}
