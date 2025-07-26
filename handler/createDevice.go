package handler

import (
	"github.com/Unicorn-s-Club/whats-unicorn/meow"
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
func CreateDeviceHandler(ctx *gin.Context) {

	device := meow.CreateDevice()

	Logger.Debugf("devices: %v", device.ID)
	Logger.Debugf("devices: %v", device.GetJID())
	SendSuccess(ctx, "send-message", device)
}
