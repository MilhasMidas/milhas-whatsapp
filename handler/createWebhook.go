package handler

import (
	"net/http"

	"github.com/Unicorn-s-Club/whats-unicorn/schemas"

	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1

// @Summary Create opening
// @Description Create a new job opening
// @Tags Openings
// @Accept json
// @Produce json
// @Param request body CreateWebhookRequest true "Request body"
// @Success 200 {object} CreateWebhookRequest
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhook [post]
func CreateWebhookHandler(ctx *gin.Context) {
	request := CreateWebhookRequest{}

	ctx.BindJSON(&request)

	Logger.Debugf("request: %v", request)

	if err := request.Validate(); err != nil {
		Logger.Errorf("validation error: %v", err.Error())
		SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	webhook := schemas.Webhook{
		URL:    request.URL,
		Active: true,
	}

	if err := Db.Create(&webhook).Error; err != nil {
		Logger.Errorf("error creating webhook: %v", err.Error())
		SendError(ctx, http.StatusInternalServerError, "error creating webhook on database")
		return
	}

	SendSuccess(ctx, "create-webhook", webhook)
}
