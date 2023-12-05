package service

import (
	"errors"
	"fmt"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"strconv"
)

type Service struct {
	r *storage.MemStorage
}

func NewService(s *storage.MemStorage) *Service {
	return &Service{s}
}

func (s *Service) ValidateAndSave(metricType, metricName, metricValue string) error {
	switch metricType {
	case storage.Gauge:
		err := s.validateGaudeValue(metricValue)
		if err != nil {
			return errors.New("incorrect value for gauge type metric")
		}
		return s.saveGaudeValue(metricName, metricValue)
	case storage.Counter:
		err := s.validateCounterValue(metricValue)
		if err != nil {
			return errors.New("incorrect value for counter type metric")
		}
		return s.saveCounterValue(metricName, metricValue)
	default:
		return errors.New("unknown metric type")
	}
}

func (s *Service) saveGaudeValue(metricName, metricValue string) error {
	return s.r.Update(storage.Gauge, metricName, metricValue)
}

func (s *Service) validateGaudeValue(metricValue string) error {
	_, err := strconv.ParseFloat(metricValue, 64)
	return err
}

func (s *Service) validateCounterValue(metricValue string) error {
	_, err := strconv.ParseInt(metricValue, 10, 64)
	return err
}

func (s *Service) saveCounterValue(metricName, metricValue string) error {
	var newValue = metricValue
	value, err := s.r.Get(storage.Counter, metricName)
	if err == nil {
		currentValue, _ := strconv.ParseInt(value, 10, 64)
		v, _ := strconv.ParseInt(metricValue, 10, 64)
		newValue = fmt.Sprintf("%d", currentValue+v)
	}
	return s.r.Update(storage.Counter, metricName, newValue)
}
