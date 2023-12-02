package service

import (
	"errors"
	"fmt"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"strconv"
	"strings"
)

const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

type Service struct {
	r *storage.MemStorage
}

func NewService(s *storage.MemStorage) *Service {
	return &Service{s}
}

func (s *Service) ValidateAndSave(metricType, metricName, metricValue string) error {
	if strings.TrimSpace(metricName) == `` {
		return errors.New(``)
	}
	switch metricType {
	case Gauge:
		err := s.validateGaudeValue(metricValue)
		if err != nil {
			return errors.New("incorrect value for gauge type metric")
		}
		s.saveGaudeValue(metricName, metricValue)
	case Counter:
		err := s.validateCounterValue(metricValue)
		if err != nil {
			return errors.New("incorrect value for counter type metric")
		}
		s.saveCounterValue(metricName, metricValue)
	default:
		return errors.New("unknown metric type")
	}
	return nil
}

func (s *Service) saveGaudeValue(metricName, metricValue string) {
	s.r.Update(Gauge, metricName, metricValue)
}

func (s *Service) validateGaudeValue(metricValue string) error {
	_, err := strconv.ParseFloat(metricValue, 64)
	return err
}

func (s *Service) validateCounterValue(metricValue string) error {
	_, err := strconv.ParseInt(metricValue, 10, 64)
	return err
}

func (s *Service) saveCounterValue(metricName, metricValue string) {
	var newValue = metricValue
	if value, ok := s.r.Get(Counter, metricName); ok {
		currentValue, _ := strconv.ParseInt(value, 10, 64)
		v, _ := strconv.ParseInt(metricValue, 10, 64)
		newValue = fmt.Sprintf("%d", currentValue+v)
	}
	s.r.Update(Counter, metricName, newValue)
}
