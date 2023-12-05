package storage

import (
	"errors"
	"fmt"
	"sync"
)

const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

type KeyStorage struct {
	MetricType string
	MetricName string
}

type MemStorage struct {
	storage map[KeyStorage]string
	mx      sync.RWMutex
}

func NewStorage() *MemStorage {
	return &MemStorage{storage: make(map[KeyStorage]string),
		mx: sync.RWMutex{},
	}
}

func (s *MemStorage) Create(metricType, metricName, value string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	key := KeyStorage{MetricType: metricType, MetricName: metricName}
	s.storage[key] = value

	return nil
}

func (s *MemStorage) Get(metricType, metricName string) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	key := KeyStorage{MetricType: metricType, MetricName: metricName}
	value, ok := s.storage[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("value by type: %s and value: %s not found", metricType, metricName))
	}
	return value, nil
}

func (s *MemStorage) GetAll() ([][]string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	values := make([][]string, 0, len(s.storage))
	for k, v := range s.storage {
		values = append(values, []string{k.MetricType, k.MetricName, v})
	}
	return values, nil
}

func (s *MemStorage) Update(metricType, metricName, value string) error {
	return s.Create(metricType, metricName, value)
}

func (s *MemStorage) Delete(metricType, metricName string) error {
	_, err := s.Get(metricType, metricName)
	if err != nil {
		return err
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	key := KeyStorage{MetricType: metricType, MetricName: metricName}
	delete(s.storage, key)
	return nil
}
