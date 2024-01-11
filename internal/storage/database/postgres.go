package database

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
)

var Url string

func NewPostrgresDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", Url)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// test connectivity
func Ping() error {
	db, err := NewPostrgresDB()
	if err != nil {
		logger.Error(fmt.Sprintf(`error database %s connection`, Url))
		return err
	}
	err = db.Ping()
	if err != nil {
		logger.Error(fmt.Sprintf(`error database %s connection`, Url))
		return err
	}
	err = db.Close()
	if err != nil {
		logger.Error(fmt.Sprintf(`error closing connection %s`, Url))
		return err
	}
	return nil
}
