package models

import (
	"context"
)

type Storager interface {
	Create(ctx context.Context, metric Metrics) (Metrics, error)
	Get(ctx context.Context, mType, name string) (Metrics, error)
	GetAll(ctx context.Context) ([]Metrics, error)
	Load(ctx context.Context, metrics []Metrics) error
	Ping(ctx context.Context) error
}
