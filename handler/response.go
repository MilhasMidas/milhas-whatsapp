package handler

import (
	"fmt"
	"net/http"

	schemas "github.com/Unicorn-s-Club/whats-unicorn/schemas"
	"github.com/gin-gonic/gin"
)

func SendError(ctx *gin.Context, code int, msg string) {
	ctx.Header("Content-type", "application/json")
	ctx.JSON(code, gin.H{
		"message":   msg,
		"errorCode": code,
	})
}

func SendSuccess(ctx *gin.Context, op string, data interface{}) {
	ctx.Header("Content-type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("operation from handler: %s successfull", op),
		"data":    data,
	})
}

type ErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

type CreateWebhookResponse struct {
	Message string                  `json:"message"`
	Data    schemas.WebhookResponse `json:"data"`
}

type DeleteWebhookResponse struct {
	Message string                  `json:"message"`
	Data    schemas.WebhookResponse `json:"data"`
}
type ShowWebhookResponse struct {
	Message string                  `json:"message"`
	Data    schemas.WebhookResponse `json:"data"`
}
type ListWebhooksResponse struct {
	Message string                    `json:"message"`
	Data    []schemas.WebhookResponse `json:"data"`
}
type UpdateWebhookResponse struct {
	Message string                  `json:"message"`
	Data    schemas.WebhookResponse `json:"data"`
}
