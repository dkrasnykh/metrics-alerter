package service

import "github.com/dkrasnykh/metrics-alerter/internal/models"

type Storager interface {
	Create(metric models.Metrics) (models.Metrics, error)
	Get(mType, name string) (models.Metrics, error)
	GetAll() ([]models.Metrics, error)
	Update(metric models.Metrics) (models.Metrics, error)
	Delete(mType, name string) error
}
