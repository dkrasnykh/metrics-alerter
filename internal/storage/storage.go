package storage

import (
	"fmt"
	"sync"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type Key struct {
	Type string
	Name string
}

type Value struct {
	ValueFloat64 float64
	ValueInt64   int64
}

type Storage struct {
	storage map[Key]Value
	mx      sync.RWMutex
}

func New() *Storage {
	return &Storage{storage: make(map[Key]Value),
		mx: sync.RWMutex{},
	}
}

func (s *Storage) Create(m models.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	k := Key{m.Type, m.Name}
	v := Value{m.ValueFloat64, m.ValueInt64}
	s.storage[k] = v
	return nil
}

func (s *Storage) Get(mType, mName string) (models.Metric, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	m := models.Metric{Type: mType, Name: mName}
	k := Key{mType, mName}
	v, ok := s.storage[k]
	if !ok {
		return m, fmt.Errorf("value by %s type and %s name not found", mType, mName)
	}
	m.ValueInt64, m.ValueFloat64 = v.ValueInt64, v.ValueFloat64
	return m, nil
}

func (s *Storage) GetAll() ([]models.Metric, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	ms := make([]models.Metric, 0, len(s.storage))
	for k, v := range s.storage {
		m := models.Metric{Type: k.Type, Name: k.Name,
			ValueFloat64: v.ValueFloat64, ValueInt64: v.ValueInt64}
		ms = append(ms, m)
	}
	return ms, nil
}

func (s *Storage) Update(m models.Metric) error {
	return s.Create(m)
}

func (s *Storage) Delete(mType, mName string) error {
	m, err := s.Get(mType, mName)
	if err != nil {
		return err
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	k := Key{m.Type, m.Name}
	delete(s.storage, k)
	return nil
}
