package config

import (
	"fmt"

	"go.mau.fi/whatsmeow/store/sqlstore"
	"gorm.io/gorm"
)

var (
	db        *gorm.DB
	container *sqlstore.Container
	logger    *Logger
)

func Init() []error {
	var err []error = make([]error, 2)

	err[0] = InitSql()
	err[1] = InitDbMeow()

	logger.Debugf("Config initialized: %v", err[0])
	logger.Debugf("Config initialized: %v", err[1])

	return err
}
func InitSql() error {
	var err error

	// Initialize Postgres
	db, err = InitializePostgres()

	if err != nil {
		return fmt.Errorf("error initializing postgres: %v", err)
	}

	return nil
}

func InitDbMeow() error {
	var err error
	container, err = InitializeSQLiteMeow()

	if err != nil {
		logger.Errorf("error initializing sqlite: %v", err)
		return fmt.Errorf("error initializing sqlite: %v", err)
	}
	return nil
}

func GetPostgres() *gorm.DB {
	return db
}

func GetSQLiteMeow() *sqlstore.Container {
	return container
}

func GetLogger(p string) *Logger {
	// Initialize Logger
	logger = NewLogger(p)
	return logger
}
