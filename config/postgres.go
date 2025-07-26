package config

import (
	schemas "github.com/Unicorn-s-Club/whats-unicorn/schemas"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializePostgres() (*gorm.DB, error) {
	// Create DB and connect
	dsn := "host=dbgo user=docker password=docker dbname=mydb port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Errorf("postgres opening error: %v", err)
		return nil, err
	}

	// add UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	// Migrate the Schema
	err = db.AutoMigrate(&schemas.User{}, &schemas.Subscription{}, &schemas.WpNumber{}, &schemas.Webhook{})
	if err != nil {
		logger.Errorf("postgres automigration error: %v", err)
		return nil, err
	}
	// Return the DB
	return db, nil
}
