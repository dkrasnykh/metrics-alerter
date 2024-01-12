package database

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
)

var URL string

func NewPostrgresDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", URL)
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
