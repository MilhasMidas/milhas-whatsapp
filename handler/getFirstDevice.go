package handler

import (
	"github.com/Unicorn-s-Club/whats-unicorn/meow"
	"github.com/Unicorn-s-Club/whats-unicorn/schemas"
	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1

// @Summary Create opening
// @Description Create a new job opening
// @Tags Openings
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "Request body"
// @Success 200 {object} SendMessageRequest
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhook [post]
func GetFirstDeviceHandler(ctx *gin.Context) {

	device, error := meow.GetAllDevices()
	if error != nil {
		Logger.Errorf("error getting devices: %v", error)
		SendError(ctx, 500, "error getting devices")
		return
	}

	Logger.Debugf("devices: %v", device[0].ID)
	Logger.Debugf("devices: %v", device[1].ID)
	webhook := schemas.Webhook{
		URL:    "http://localhost:8080/api/v1/webhook",
		Active: true,
	}

	SendSuccess(ctx, "send-message", webhook)
}
