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
					value         double precision,
					time          timestamp without time zone NOT NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC')
				);
				CREATE INDEX IF NOT EXISTS name_idx ON metrics (name);
				CREATE INDEX IF NOT EXISTS type_idx ON metrics (type);`,
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
	row := s.db.QueryRowContext(context.Background(), `select delta, value from metrics where name=$1 and type=$2 ORDER BY time DESC LIMIT 1;`, name, mType)
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
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT t1.name, t1.type, m.delta, m.value FROM 
				(select name, type, MAX(time) as time from metrics group by name, type) AS t1 
				LEFT JOIN metrics AS m ON t1.name = m.name AND t1.type=m.type AND t1.time = m.time;`)
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
		if err != nil {
			logger.Error(err.Error())
		}
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
	tx, err := s.db.Begin()
	if err != nil {
		return models.Metrics{}, err
	}
	row := tx.QueryRowContext(context.Background(), `select id from metrics where name=$1 and type=$2 ORDER BY time DESC LIMIT 1;`, metric.ID, metric.MType)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return models.Metrics{}, err
	}
	_, err = tx.ExecContext(context.Background(), `UPDATE metrics SET delta=$1, value=$2 WHERE id=$3;`,
		deltaOrDefault(metric.Delta), valueOrDefault(metric.Value), id)
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
	row := tx.QueryRowContext(context.Background(), `select id from metrics where name=$1 and type=$2 ORDER BY time DESC LIMIT 1;`, name, mType)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(context.Background(), `DELETE FROM metrics WHERE id=$1`, id)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return tx.Commit()

}

func (s *Storage) Load(metrics []models.Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return nil
	}
	for _, m := range metrics {
		_, err = tx.ExecContext(context.Background(),
			"INSERT INTO metrics (name, type, delta, value) VALUES($1,$2,$3,$4)",
			m.ID, m.MType, deltaOrDefault(m.Delta), valueOrDefault(m.Value))
		if err != nil {
			err = tx.Rollback()
			return err
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
