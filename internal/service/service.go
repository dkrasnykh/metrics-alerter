package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

var ErrUnknownMetricType = errors.New("unknown metric type")
var ErrIDIsEmpty = errors.New("metric ID is empty")

type Service struct {
	r Storager
}

func New(s Storager) *Service {
	return &Service{r: s}
}

func (s *Service) Validate(m models.Metrics) error {
	if m.ID == `` {
		return ErrIDIsEmpty
	}
	switch m.MType {
	case models.GaugeType:
		if m.Value == nil {
			return fmt.Errorf(`value undefined for metric type %s`, m.MType)
		}
	case models.CounterType:
		if m.Delta == nil {
			return fmt.Errorf(`delta undefined for metric type %s`, m.MType)
		}
	default:
		return ErrUnknownMetricType
	}
	return nil
}

func (s *Service) Save(m models.Metrics) (models.Metrics, error) {
	if m.MType == models.CounterType {
		delta := s.calculateCounterValue(m.ID, *m.Delta)
		m.Delta = &delta
	}

	return s.r.Create(m)
}

func (s *Service) calculateCounterValue(name string, value int64) int64 {
	metric, err := s.r.Get(models.CounterType, name)
	if err != nil || metric.Delta == nil {
		return value
	}
	value += *metric.Delta
	return value
}

func (s *Service) GetMetricValue(mType, mName string) (string, error) {
	m, err := s.r.Get(mType, mName)
	if err != nil {
		return "", err
	}
	switch mType {
	case models.CounterType:
		return fmt.Sprintf("%d", *m.Delta), nil
	default:
		return strconv.FormatFloat(*m.Value, 'g', -1, 64), nil
	}
}

func (s *Service) GetAll() ([]models.Metrics, error) {
	return s.r.GetAll()
}

func (s *Service) Get(mType, mName string) (models.Metrics, error) {
	return s.r.Get(mType, mName)
}

func (s *Service) Load(metrics []models.Metrics) error {
	counters := map[string]int64{}
	gauges := map[string]float64{}
	for i := 0; i < len(metrics); i++ {
		m := metrics[i]
		switch m.MType {
		case models.CounterType:
			counters[m.ID] += *m.Delta
		case models.GaugeType:
			gauges[m.ID] = *m.Value
		}
	}
	validated := []models.Metrics{}
	for name, value := range counters {
		delta := s.calculateCounterValue(name, value)
		m := models.Metrics{MType: models.CounterType, ID: name, Delta: &delta}
		validated = append(validated, m)
	}
	for name, value := range gauges {
		m := models.Metrics{MType: models.GaugeType, ID: name, Value: &value}
		validated = append(validated, m)
	}
	return s.r.Load(validated)
}
