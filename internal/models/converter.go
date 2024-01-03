package models

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
