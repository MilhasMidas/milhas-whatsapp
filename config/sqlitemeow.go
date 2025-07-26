package config

import (
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func InitializeSQLiteMeow() (*sqlstore.Container, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:./db/whatsmeow.db?_foreign_keys=on", dbLog)
	logger.Errorf("Erro to create Database: %v", err)

	return container, err

}
