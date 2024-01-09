package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

var ErrUnknownMetricType = errors.New("unknown metric type")
var ErrIDIsEmpty = errors.New("metric ID is empty")

type Service struct {
	r Storager
}

func New(s *memory.Storage) *Service {
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

	return s.r.Update(m)
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
