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

	// Create custom enum type for ServiceClassCode
	db.Exec(`
		DO $$ BEGIN
			CREATE TYPE service_class_code AS ENUM ('ECONOMY', 'PREMIUM_ECONOMY', 'BUSINESS', 'FIRST', 'OTHER');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`)

	// Migrate the Schema
	err = db.AutoMigrate(
		&schemas.User{},
		&schemas.Subscription{},
		&schemas.WpNumber{},
		&schemas.Webhook{},
		&schemas.Airport{},
		&schemas.Airline{},
		&schemas.LoyaltyProgram{},
		&schemas.FlightAward{},
		&schemas.FlightAwardAvailableDate{},
		&schemas.FlightAwardProgramCost{},
		&schemas.FlightConnection{},
	)
	if err != nil {
		logger.Errorf("postgres automigration error: %v", err)
		return nil, err
	}
	// Return the DB
	return db, nil
}
