package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/repository"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
)

var ErrUnknownMetricType = errors.New("unknown metric type")
var ErrIDIsEmpty = errors.New("metric ID is empty")

type Service struct {
	r          repository.Storager
	c          *config.ServerConfig
	backupPath string
}

func New(s *storage.Storage, conf *config.ServerConfig) *Service {
	return &Service{
		r: s,
		c: conf,
	}
}

func (s *Service) InitBackup() error {
	err := os.MkdirAll(s.c.FileStoragePath+"/", 0777)
	if err != nil {
		return err
	}
	s.backupPath = s.c.FileStoragePath + "/metrics.tmp"
	if s.c.Restore {
		err = s.Restore()
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
	}
	return nil
}

func (s *Service) Restore() error {
	data, err := models.Load(s.backupPath)
	if err != nil {
		return err
	}
	err = s.r.Load(data)
	if err != nil {
		return err
	}
	return nil
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
	if s.c != nil && s.c.FileStoragePath != "" {
		time.AfterFunc(s.c.StoreInterval, func() {
			checker := func(err error) {
				if err != nil {
					log.Printf("error: %s", err.Error())
				}
			}
			ms, err := s.r.GetAll()
			checker(err)
			err = models.Save(s.backupPath, ms)
			checker(err)
		})
	}
	return s.r.Update(m)
}

func (s *Service) calculateCounterValue(name string, value int64) int64 {
	metric, err := s.r.Get(models.CounterType, name)
	if err == nil {
		value += *metric.Delta
	}
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
