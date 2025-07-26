package router

import (
	"os"

	"github.com/gin-gonic/gin"
)

func Initialize() {
	// Initialize Router
	ginRouter := gin.Default()

	// Initialize Routes
	initializeRoutes(ginRouter)

	// Get the port from the environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Run the server
	ginRouter.Run("0.0.0.0:" + port)
}
