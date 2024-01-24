package database

import (
	"github.com/avast/retry-go"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
)

var URL string

func NewPostrgresDB() (*sqlx.DB, error) {
	var db *sqlx.DB
	err := retry.Do(
		func() error {
			var err error
			db, err = sqlx.Open("pgx", URL)
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}
