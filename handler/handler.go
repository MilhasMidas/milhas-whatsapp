package handler

import (
	"github.com/Unicorn-s-Club/whats-unicorn/config"
	"gorm.io/gorm"
)

var (
	Logger *config.Logger
	Db     *gorm.DB
)

func InitializeHandler() {
	Logger = config.GetLogger("handler")
	Db = config.GetPostgres()
}
