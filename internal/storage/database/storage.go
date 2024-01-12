package database

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type Storage struct {
	db *sqlx.DB
}

func New(url string) (*Storage, error) {
	URL = url
	db, err := NewPostrgresDB()
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, InitDB(db)
}

func InitDB(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS metrics
				(
 				    id            serial       not null unique,
				    name          varchar(255) not null,
 					type          varchar(255) not null,
				    delta         bigint,
					value         double precision
				);`,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Storage) Create(metric models.Metrics) (models.Metrics, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Metrics{}, err
	}
	_, err = tx.ExecContext(context.Background(), `INSERT INTO metrics (name, type, delta, value) VALUES ($1, $2, $3, $4);`,
		metric.ID, metric.MType, deltaOrDefault(metric.Delta), valueOrDefault(metric.Value))
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return metric, tx.Commit()
}

func (s *Storage) Get(mType, name string) (models.Metrics, error) {
	row := s.db.QueryRowContext(context.Background(), `select delta, value from metrics where name=$1 and type=$2;`, name, mType)
	var delta int64
	var value float64

	err := row.Scan(&delta, &value)

	if err != nil {
		return models.Metrics{}, err
	}
	return models.Metrics{MType: mType, ID: name, Delta: &delta, Value: &value}, nil
}

func (s *Storage) GetAll() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)
	rows, err := s.db.QueryContext(context.Background(), "SELECT name, type, delta, value from metrics;")
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.Metrics
		var delta int64
		var value float64
		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		m.Delta = &delta
		m.Value = &value
		metrics = append(metrics, m)
	}

	err = rows.Close()
	if err != nil {
		logger.Error(err.Error())
	}

	return metrics, nil
}

func (s *Storage) Update(metric models.Metrics) (models.Metrics, error) {
	_, err := s.Get(metric.MType, metric.ID)
	if err != nil {
		return models.Metrics{}, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return models.Metrics{}, err
	}
	_, err = tx.ExecContext(context.Background(), `UPDATE metrics SET delta=$1, value=$2 WHERE name=$3 and type=$4;`,
		deltaOrDefault(metric.Delta), valueOrDefault(metric.Value), metric.ID, metric.MType)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return metric, tx.Commit()
}

func (s *Storage) Delete(mType, name string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(context.Background(), `DELETE FROM metrics WHERE name=$1 and type=$2;`, name, mType)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return tx.Commit()

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
