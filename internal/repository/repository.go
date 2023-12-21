package repository

import "github.com/dkrasnykh/metrics-alerter/internal/models"

type Storager interface {
	Create(metric models.Metric) error
	Get(mType, name string) (models.Metric, error)
	GetAll() ([]models.Metric, error)
	Update(metric models.Metric) error
	Delete(mType, name string) error
}
