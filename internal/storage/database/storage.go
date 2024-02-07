package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type Storage struct {
	db *sqlx.DB
}

func New(url string) (*Storage, error) {
	db, err := sqlx.Open("pgx", url)
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err = tx.ExecContext(ctx,
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
		err = tx.Rollback()
		logger.LogErrorIfNotNil(err)
	}
	return tx.Commit()
}

func (s *Storage) Create(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	var err error
	switch metric.MType {
	case models.GaugeType:
		_, err = s.db.ExecContext(ctx, `INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3);`,
			metric.ID, metric.MType, *metric.Value)
	case models.CounterType:
		_, err = s.db.ExecContext(ctx, `INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3);`,
			metric.ID, metric.MType, *metric.Delta)
	}
	if err != nil {
		return models.Metrics{}, err
	}
	return metric, nil
}

func (s *Storage) Get(ctx context.Context, mType, name string) (models.Metrics, error) {
	row := s.db.QueryRowContext(ctx, `select delta, value from metrics where name=$1 and type=$2 ORDER BY time DESC LIMIT 1;`, name, mType)
	if row.Err() != nil {
		return models.Metrics{}, row.Err()
	}
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := row.Scan(&delta, &value)

	if err != nil {
		return models.Metrics{}, err
	}

	return metric(models.Metrics{MType: mType, ID: name}, delta, value), nil
}

func (s *Storage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)
	rows, err := s.db.QueryContext(ctx,
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
		var delta sql.NullInt64
		var value sql.NullFloat64
		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		logger.LogErrorIfNotNil(err)
		metrics = append(metrics, metric(m, delta, value))
	}
	err = rows.Close()
	logger.LogErrorIfNotNil(err)
	return metrics, nil
}

func (s *Storage) Load(ctx context.Context, metrics []models.Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return nil
	}
	for _, m := range metrics {
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
		if err != nil {
			err = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
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
