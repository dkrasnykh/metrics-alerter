package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

const HashHeader = "HashSHA256"

func Convert(mtype, mname, value string) models.Metrics {
	m := models.Metrics{MType: mtype, ID: mname}
	switch mtype {
	case models.CounterType:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			m.Delta = &delta
		}
	case models.GaugeType:
		gvalue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			m.Value = &gvalue
		}
	}
	return m
}

func GetMetric(mtype, name string, value float64, delta int64) models.Metrics {
	m := models.Metrics{MType: mtype, ID: name}
	switch mtype {
	case models.CounterType:
		m.Delta = &delta
	case models.GaugeType:
		m.Value = &value
	}
	return m
}

func ExtractBody(req *http.Request) (*models.Metrics, error) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, errors.New(`request body is empty`)
	}
	var m models.Metrics
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func LogError(err error) {
	if err != nil {
		logger.Error(err.Error())
	}
}

func Hash(bytes []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(bytes)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
