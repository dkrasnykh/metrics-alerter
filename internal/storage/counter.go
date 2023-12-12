package storage

import (
	"fmt"
	"sync"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type CounterStorage struct {
	storage map[string]int64
	mx      sync.RWMutex
}

func NewCounterStorage() *CounterStorage {
	return &CounterStorage{storage: make(map[string]int64),
		mx: sync.RWMutex{},
	}
}

func (s *CounterStorage) Create(name string, value any) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.storage[name] = value.(int64)

	return nil
}

func (s *CounterStorage) Get(name string) (any, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	value, ok := s.storage[name]
	if !ok {
		return value, fmt.Errorf("value by counter type and %s name not found", name)
	}
	return value, nil
}

func (s *CounterStorage) GetAll() (any, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	values := make([]models.Counter, 0, len(s.storage))
	for k, v := range s.storage {
		values = append(values, models.Counter{Name: k, Value: v})
	}
	return values, nil
}

func (s *CounterStorage) Update(name string, value any) error {
	return s.Create(name, value)
}

func (s *CounterStorage) Delete(name string) error {
	_, err := s.Get(name)
	if err != nil {
		return err
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	delete(s.storage, name)
	return nil
}
