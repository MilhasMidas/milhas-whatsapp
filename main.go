package main

import (
	"github.com/Unicorn-s-Club/whats-unicorn/config"
	"github.com/Unicorn-s-Club/whats-unicorn/meow"
	"github.com/Unicorn-s-Club/whats-unicorn/router"
)

var (
	logger *config.Logger
)

func main() {
	logger = config.GetLogger("main")
	// Initialize Configs
	err := config.Init()
	if err[0] != nil && err[1] != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}

	meow.StartSession()

	router.Initialize()
}
