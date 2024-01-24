package repository

import (
	"context"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type Storager interface {
	Create(ctx context.Context, metric models.Metrics) (models.Metrics, error)
	Get(ctx context.Context, mType, name string) (models.Metrics, error)
	GetAll(ctx context.Context) ([]models.Metrics, error)
	Load(ctx context.Context, metrics []models.Metrics) error
	Ping(ctx context.Context) error
}
