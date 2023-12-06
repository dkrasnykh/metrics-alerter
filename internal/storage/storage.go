package storage

import (
	"fmt"
	"sync"
)

const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

type Key struct {
	Type string
	Name string
}

type MemStorage struct {
	storage map[Key]string
	mx      sync.RWMutex
}

func NewStorage() *MemStorage {
	return &MemStorage{storage: make(map[Key]string),
		mx: sync.RWMutex{},
	}
}

func (s *MemStorage) Create(metricType, metricName, value string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	key := Key{Type: metricType, Name: metricName}
	s.storage[key] = value

	return nil
}

func (s *MemStorage) Get(metricType, metricName string) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	key := Key{Type: metricType, Name: metricName}
	value, ok := s.storage[key]
	if !ok {
		return "", fmt.Errorf("value by type: %s and value: %s not found", metricType, metricName)
	}
	return value, nil
}

func (s *MemStorage) GetAll() (map[string][][2]string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	values := make(map[string][][2]string)
	for k, v := range s.storage {
		if _, ok := values[k.Type]; !ok {
			values[k.Type] = make([][2]string, 0)
		}
		values[k.Type] = append(values[k.Type], [...]string{k.Name, v})
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
	key := Key{Type: metricType, Name: metricName}
	delete(s.storage, key)
	return nil
}
