package config

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializePostgres() (*gorm.DB, error) {
	// Create DB and connect
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		// dsn = "host=db-milhas-instance-1.cr6c324oad5p.us-east-1.rds.amazonaws.com user=postgres password=CM1YdUAuRg7ZM5DBuGth dbname=dbaward port=5432 sslmode=require TimeZone=America/Sao_Paulo"
		dsn = "host=db-postgresql-sfo3-66243-do-user-26605971-0.l.db.ondigitalocean.com user=doadmin password=AVNS_ohcF4yGvKSOmzUFxThD dbname=defaultdb port=25060 sslmode=require TimeZone=America/Sao_Paulo"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Errorf("postgres opening error: %v", err)
		return nil, err
	}

	// // add UUID extension
	// db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// // Create custom enum type for ServiceClassCode
	// db.Exec(`
	// 	DO $$ BEGIN
	// 		CREATE TYPE service_class_code AS ENUM ('ECONOMY', 'PREMIUM_ECONOMY', 'BUSINESS', 'FIRST', 'OTHER');
	// 	EXCEPTION
	// 		WHEN duplicate_object THEN null;
	// 	END $$;
	// `)

	// // Migrate the Schema
	// err = db.AutoMigrate(
	// 	&schemas.User{},
	// 	&schemas.Subscription{},
	// 	&schemas.WpNumber{},
	// 	&schemas.Webhook{},
	// 	&schemas.Airport{},
	// 	&schemas.Airline{},
	// 	&schemas.LoyaltyProgram{},
	// 	&schemas.FlightAward{},
	// 	&schemas.FlightAwardAvailableDate{},
	// 	&schemas.FlightAwardProgramCost{},
	// 	&schemas.FlightConnection{},
	// )
	// if err != nil {
	// 	logger.Errorf("postgres automigration error: %v", err)
	// 	return nil, err
	// }
	// Return the DB
	return db, nil
}
