package config

import (
	"context"
	"log"
	"path/filepath"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func InitializeSQLiteMeow() (*sqlstore.Container, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)

	// Get the absolute path to the database file
	dbPath := filepath.Join("db", "whatsmeow.db")
	absolutePath, err := filepath.Abs(dbPath)
	if err != nil {
		if logger != nil {
			logger.Errorf("Error getting absolute path for database: %v", err)
		} else {
			log.Printf("Error getting absolute path for database: %v", err)
		}
		return nil, err
	}

	dsn := "file:" + absolutePath + "?_foreign_keys=on"
	container, err := sqlstore.New(context.Background(), "sqlite3", dsn, dbLog)
	if err != nil {
		// Use a temporary logger if the global logger is not initialized yet
		if logger != nil {
			logger.Errorf("Error creating Database: %v", err)
		} else {
			log.Printf("Error creating Database: %v", err)
		}
		return nil, err
	}

	return container, nil
}
