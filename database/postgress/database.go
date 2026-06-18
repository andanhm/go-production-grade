// Package database holds the postgres db client connection
package database

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgres "gorm.io/driver/postgres"
)

type config interface {
	URIBuilder() string
	EnableLogging() bool
	Name() string
}

func New(config config) (*gorm.DB, error) {
	logging := logger.New(&Logger{}, logger.Config{
		LogLevel:                  logger.Error,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
		SlowThreshold:             time.Second * 30,
	})
	if config.EnableLogging() {
		logging = logger.Default.LogMode(logger.Info)
	}
	uri := config.URIBuilder()
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  uri,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger:      logging,
		PrepareStmt: false,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
