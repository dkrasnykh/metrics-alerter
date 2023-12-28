package storage

import (
	"fmt"
	"sync"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type Key struct {
	MType string
	ID    string
}

type Value struct {
	Value float64
	Delta int64
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

func (s *Storage) Create(m models.Metrics) (models.Metrics, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	k := Key{m.MType, m.ID}
	v := Value{valueOrDefault(m.Value), deltaOrDefault(m.Delta)}
	s.storage[k] = v
	return m, nil
}

func (s *Storage) Get(mType, mName string) (models.Metrics, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	k := Key{mType, mName}
	v, ok := s.storage[k]
	if !ok {
		return models.Metrics{}, fmt.Errorf("value by %s type and %s name not found", mType, mName)
	}
	return models.GetMetric(k.MType, k.ID, v.Value, v.Delta), nil
}

func (s *Storage) GetAll() ([]models.Metrics, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	ms := make([]models.Metrics, 0, len(s.storage))
	for k, v := range s.storage {
		ms = append(ms, models.GetMetric(k.MType, k.ID, v.Value, v.Delta))
	}
	return ms, nil
}

func (s *Storage) Update(m models.Metrics) (models.Metrics, error) {
	return s.Create(m)
}

func (s *Storage) Delete(mType, mName string) error {
	m, err := s.Get(mType, mName)
	if err != nil {
		return err
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	k := Key{m.MType, m.ID}
	delete(s.storage, k)
	return nil
}

func (s *Storage) Load(ms []models.Metrics) error {
	for _, m := range ms {
		_, err := s.Update(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func deltaOrDefault(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func valueOrDefault(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
