package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/utils"
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
	storage           map[Key]Value
	filePath          string
	fileStoreInterval int
	mx                sync.RWMutex
}

func New(path string, interval int) *Storage {
	return &Storage{
		storage:           make(map[Key]Value),
		filePath:          InitDir(path),
		fileStoreInterval: interval,
		mx:                sync.RWMutex{},
	}
}

func (s *Storage) Create(ctx context.Context, m models.Metrics) (models.Metrics, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	k := Key{m.MType, m.ID}
	v := Value{valueOrDefault(m.Value), deltaOrDefault(m.Delta)}
	s.storage[k] = v

	s.StoreIntoFile()

	return m, nil
}

func (s *Storage) Get(ctx context.Context, mType, mName string) (models.Metrics, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	k := Key{mType, mName}
	v, ok := s.storage[k]
	if !ok {
		return models.Metrics{}, fmt.Errorf("value by %s type and %s name not found", mType, mName)
	}
	return utils.GetMetric(k.MType, k.ID, v.Value, v.Delta), nil
}

func (s *Storage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	ms := make([]models.Metrics, 0, len(s.storage))
	for k, v := range s.storage {
		ms = append(ms, utils.GetMetric(k.MType, k.ID, v.Value, v.Delta))
	}
	return ms, nil
}

func (s *Storage) Load(ctx context.Context, metrics []models.Metrics) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, m := range metrics {
		key := Key{MType: m.MType, ID: m.ID}
		value := Value{Value: valueOrDefault(m.Value), Delta: deltaOrDefault(m.Delta)}
		s.storage[key] = value
	}
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return errors.New(`database is not used`)
}

func (s *Storage) StoreIntoFile() {
	if s.filePath != "" {
		timeDuration := time.Duration(s.fileStoreInterval) * time.Second
		time.AfterFunc(timeDuration, func() {
			checker := func(err error) {
				if err != nil {
					logger.Error(err.Error())
				}
			}
			ms, err := s.GetAll(context.Background())
			checker(err)
			err = Save(s.filePath, ms)
			checker(err)
		})
	}
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
