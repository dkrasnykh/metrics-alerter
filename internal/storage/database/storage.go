package database

import (
	"context"
	"database/sql"

	"github.com/avast/retry-go"
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

	err = retry.Do(
		func() error {
			var err error
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
			return err
		},
		retry.Attempts(models.Attempts),
		retry.DelayType(models.DelayType),
		retry.OnRetry(models.OnRetry),
	)

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Storage) Create(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Metrics{}, err
	}
	err = retry.Do(
		func() error {
			var err error

			switch metric.MType {
			case models.GaugeType:
				_, err = tx.ExecContext(ctx, `INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3);`,
					metric.ID, metric.MType, *metric.Value)
			case models.CounterType:
				_, err = tx.ExecContext(ctx, `INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3);`,
					metric.ID, metric.MType, *metric.Delta)
			}
			return err
		},
		retry.Attempts(models.Attempts),
		retry.DelayType(models.DelayType),
		retry.OnRetry(models.OnRetry),
	)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return metric, tx.Commit()
}

func (s *Storage) Get(ctx context.Context, mType, name string) (models.Metrics, error) {
	var row *sql.Row

	err := retry.Do(
		func() error {
			row = s.db.QueryRowContext(ctx, `select delta, value from metrics where name=$1 and type=$2 ORDER BY time DESC LIMIT 1;`, name, mType)
			return row.Err()
		},
		retry.Attempts(models.Attempts),
		retry.DelayType(models.DelayType),
		retry.OnRetry(models.OnRetry),
	)

	if err != nil {
		return models.Metrics{}, err
	}
	var delta sql.NullInt64
	var value sql.NullFloat64

	err = row.Scan(&delta, &value)

	if err != nil {
		return models.Metrics{}, err
	}

	return metric(models.Metrics{MType: mType, ID: name}, delta, value), nil
}

func (s *Storage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)
	var rows *sql.Rows

	err := retry.Do(
		func() error {
			var err error
			rows, err = s.db.QueryContext(ctx,
				`SELECT t1.name, t1.type, m.delta, m.value FROM 
				(select name, type, MAX(time) as time from metrics group by name, type) AS t1 
				LEFT JOIN metrics AS m ON t1.name = m.name AND t1.type=m.type AND t1.time = m.time;`)
			if rows.Err() != nil {
				logger.Error(rows.Err().Error())
			}
			return err
		},
		retry.Attempts(models.Attempts),
		retry.DelayType(models.DelayType),
		retry.OnRetry(models.OnRetry),
	)

	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64
		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			logger.Error(err.Error())
		}

		metrics = append(metrics, metric(m, delta, value))
	}

	err = rows.Close()
	if err != nil {
		logger.Error(err.Error())
	}

	return metrics, nil
}

func (s *Storage) Load(ctx context.Context, metrics []models.Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return nil
	}
	for _, m := range metrics {
		err = retry.Do(
			func() error {
				var err error

				switch m.MType {
				case models.CounterType:
					_, err = tx.ExecContext(ctx,
						"INSERT INTO metrics (name, type, delta) VALUES($1,$2,$3)",
						m.ID, m.MType, *m.Delta)
				case models.GaugeType:
					_, err = tx.ExecContext(ctx,
						"INSERT INTO metrics (name, type, value) VALUES($1,$2,$3)",
						m.ID, m.MType, *m.Value)
				}
				return err
			},
			retry.Attempts(models.Attempts),
			retry.DelayType(models.DelayType),
			retry.OnRetry(models.OnRetry),
		)
		if err != nil {
			err = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func metric(m models.Metrics, delta sql.NullInt64, value sql.NullFloat64) models.Metrics {
	switch m.MType {
	case models.CounterType:
		m.Delta = &delta.Int64
	case models.GaugeType:
		m.Value = &value.Float64
	}
	return m
}
