package storage

import (
	"fmt"
	"sync"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type GaugeStorage struct {
	storage map[string]float64
	mx      sync.RWMutex
}

func NewGaugeStorage() *GaugeStorage {
	return &GaugeStorage{storage: make(map[string]float64),
		mx: sync.RWMutex{},
	}
}

func (s *GaugeStorage) Create(name string, value any) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.storage[name] = value.(float64)

	return nil
}

func (s *GaugeStorage) Get(name string) (any, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	value, ok := s.storage[name]
	if !ok {
		return value, fmt.Errorf("value by gauge type and %s name not found", name)
	}
	return value, nil
}

func (s *GaugeStorage) GetAll() (any, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	values := make([]models.Gauge, 0, len(s.storage))
	for k, v := range s.storage {
		values = append(values, models.Gauge{Name: k, Value: v})
	}
	return values, nil
}

func (s *GaugeStorage) Update(name string, value any) error {
	return s.Create(name, value)
}

func (s *GaugeStorage) Delete(name string) error {
	_, err := s.Get(name)
	if err != nil {
		return err
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	delete(s.storage, name)
	return nil
}
