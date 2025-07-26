package router

import (
	// docs "github.com/arthur404dev/gopportunities/docs"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Unicorn-s-Club/whats-unicorn/docs"
	"github.com/Unicorn-s-Club/whats-unicorn/handler"
	"github.com/Unicorn-s-Club/whats-unicorn/handler/user"
	swaggerfiles "github.com/swaggo/files"
)

func initializeRoutes(router *gin.Engine) {
	// Initialize Handler
	handler.InitializeHandler()
	basePath := "/api"
	docs.SwaggerInfo.BasePath = basePath
	api := router.Group(basePath)
	{
		api.GET("/firstdevice", handler.GetFirstDeviceHandler)
		api.POST("/webhook", handler.CreateWebhookHandler)
		api.POST("/device", handler.CreateDeviceHandler)

		api.POST("/user", user.CreateUserHandler)
		api.POST("waclient", handler.CreateWaClientHandler)
		// api.DELETE("/webhook", handler.)
		// api.PUT("/webhook", handler.)
		// api.GET("/webhooks", handler.)

		//whatsMeow
		api.POST("/send", handler.SendMessageHandler)
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Hello, world!"})
		})

	}
	// Initialize Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
