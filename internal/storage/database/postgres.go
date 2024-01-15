package database

import (
	"fmt"

	"github.com/avast/retry-go"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
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
		retry.Attempts(models.Attempts),
		retry.DelayType(models.DelayType),
		retry.OnRetry(models.OnRetry),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Ping() error {
	db, err := NewPostrgresDB()
	if err != nil {
		logger.Error(fmt.Sprintf(`error database %s connection`, URL))
		return err
	}
	err = db.Ping()
	if err != nil {
		logger.Error(fmt.Sprintf(`error ping database %s`, URL))
		return err
	}
	err = db.Close()
	if err != nil {
		logger.Error(fmt.Sprintf(`error closing connection %s`, URL))
		return err
	}
	return nil
}
