package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/repository"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
)

var ErrUnknownMetricType = errors.New("unknown metric type")

type Service struct {
	r repository.Storager
}

func New(s *storage.Storage) *Service {
	return &Service{
		r: s,
	}
}

func (s *Service) Save(metricType, metricName, metricValue string) error {
	metric := models.Metric{Name: metricName, Type: metricType}
	switch metricType {
	case models.GaugeType:
		metric.ValueFloat64, _ = strconv.ParseFloat(metricValue, 64)
	case models.CounterType:
		metric.ValueInt64 = s.calculateCounterValue(metricName, metricValue)
	}
	err := s.r.Create(metric)
	return err
}

func (s *Service) Validate(metricType, value string) error {
	var err error
	switch metricType {
	case models.CounterType:
		_, err = strconv.ParseInt(value, 10, 64)
	case models.GaugeType:
		_, err = strconv.ParseFloat(value, 64)
	default:
		err = ErrUnknownMetricType
	}
	return err
}

func (s *Service) calculateCounterValue(name, value string) int64 {
	metric, err := s.r.Get(models.CounterType, name)
	currValue, _ := strconv.ParseInt(value, 10, 64)
	if err == nil {
		currValue += metric.ValueInt64
	}
	return currValue
}

func (s *Service) GetMetricValue(mType, mName string) (string, error) {
	m, err := s.r.Get(mType, mName)
	if err != nil {
		return "", err
	}
	switch mType {
	case models.CounterType:
		return fmt.Sprintf("%d", m.ValueInt64), nil
	default:
		return strconv.FormatFloat(m.ValueFloat64, 'g', -1, 64), nil
	}
}

func (s *Service) GetAll() ([]models.Metric, error) {
	return s.r.GetAll()
}
