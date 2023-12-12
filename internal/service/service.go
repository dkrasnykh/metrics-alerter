package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/repository"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
)

type Service struct {
	r map[string]repository.Storage
}

func NewService(cs *storage.CounterStorage, gs *storage.GaugeStorage) *Service {
	return &Service{r: map[string]repository.Storage{
		models.CounterType: cs,
		models.GaugeType:   gs},
	}
}

func (s *Service) ValidateAndSave(metricType, metricName, metricValue string) error {
	switch metricType {
	case models.GaugeType:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.New("incorrect value for gauge type metric")
		}
		return s.r[models.GaugeType].Update(metricName, value)
	case models.CounterType:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return errors.New("incorrect value for counter type metric")
		}
		return s.saveCounterValue(metricName, value)
	default:
		return errors.New("unknown metric type")
	}
}

func (s *Service) saveCounterValue(metricName string, value int64) error {
	currValueAny, err := s.r[models.CounterType].Get(metricName)
	if err == nil {
		currValue, ok := currValueAny.(int64)
		if !ok {
			return fmt.Errorf("error retriving current value for counter metric %s", metricName)
		}
		value += currValue
	}
	return s.r[models.CounterType].Update(metricName, value)
}

func (s *Service) GetMetricValue(metricType, metricName string) (string, error) {
	switch metricType {
	case models.GaugeType:
		return s.getGaugeValue(metricName)
	case models.CounterType:
		return s.getCounterValue(metricName)
	default:
		return "", errors.New("unknown metric type")
	}
}

func (s *Service) getGaugeValue(metricName string) (string, error) {
	valueAny, err := s.r[models.GaugeType].Get(metricName)
	if err != nil {
		return "", err
	}
	value, ok := valueAny.(float64)
	if !ok {
		return "", fmt.Errorf("error retriving gauge value %s", metricName)
	}
	return strconv.FormatFloat(value, 'g', -1, 64), nil
}

func (s *Service) getCounterValue(metricName string) (string, error) {
	valueAny, err := s.r[models.CounterType].Get(metricName)
	if err != nil {
		return "", err
	}
	value, ok := valueAny.(int64)
	if !ok {
		return "", fmt.Errorf("error retriving counter value %s", metricName)
	}
	return fmt.Sprintf("%d", value), nil
}

func (s *Service) GetAllCounter() ([]models.Counter, error) {
	valueAny, err := s.r[models.CounterType].GetAll()
	if err != nil {
		return nil, err
	}
	value, ok := valueAny.([]models.Counter)
	if !ok {
		return nil, fmt.Errorf("error retriving all counter values")
	}
	return value, nil
}

func (s *Service) GetAllGauge() ([]models.Gauge, error) {
	valueAny, err := s.r[models.GaugeType].GetAll()
	if err != nil {
		return nil, err
	}
	value, ok := valueAny.([]models.Gauge)
	if !ok {
		return nil, fmt.Errorf("error retriving all gauge values")
	}
	return value, nil
}
