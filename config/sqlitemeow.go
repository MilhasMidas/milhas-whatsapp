package config

import (
	"log"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func InitializeSQLiteMeow() (*sqlstore.Container, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:./db/whatsmeow.db?_foreign_keys=on", dbLog)
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
