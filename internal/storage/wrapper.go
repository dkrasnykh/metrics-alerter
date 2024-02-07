package storage

import (
	"context"

	"github.com/avast/retry-go"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/repository"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/database"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

type StorageWrap struct {
	r repository.Storager
}

func New(c *config.ServerConfig) (repository.Storager, error) {
	var r repository.Storager
	var err error
	if c.DatabaseDSN == `` {
		r = memory.New(c.FileStoragePath, c.StoreInterval)
		if c.Restore {
			err := retry.Do(
				func() error {
					err := memory.Restore(r, c.FileStoragePath)
					return err
				},
				retry.Attempts(config.Attempts),
				retry.DelayType(config.DelayType),
				retry.OnRetry(config.OnRetry),
			)
			logger.LogErrorIfNotNil(err)
		}
	} else {
		err = retry.Do(
			func() error {
				var err error
				r, err = database.New(c.DatabaseDSN)
				return err
			},
			retry.Attempts(config.Attempts),
			retry.DelayType(config.DelayType),
			retry.OnRetry(config.OnRetry),
		)
	}
	if err != nil {
		return nil, err
	}
	return &StorageWrap{r: r}, nil
}

func (s *StorageWrap) Create(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	var m models.Metrics
	err := retry.Do(
		func() error {
			var err error
			m, err = s.r.Create(ctx, metric)
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
	return m, err
}

func (s *StorageWrap) Get(ctx context.Context, mType, name string) (models.Metrics, error) {
	var m models.Metrics
	err := retry.Do(
		func() error {
			var err error
			m, err = s.r.Get(ctx, mType, name)
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
	return m, err
}

func (s *StorageWrap) GetAll(ctx context.Context) ([]models.Metrics, error) {
	var m []models.Metrics
	err := retry.Do(
		func() error {
			var err error
			m, err = s.r.GetAll(ctx)
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
	return m, err
}

func (s *StorageWrap) Load(ctx context.Context, metrics []models.Metrics) error {
	return retry.Do(
		func() error {
			err := s.r.Load(ctx, metrics)
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
}

func (s *StorageWrap) Ping(ctx context.Context) error {
	return s.r.Ping(ctx)
}
