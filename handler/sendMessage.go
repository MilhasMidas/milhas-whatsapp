package handler

import (
	"strings"

	"github.com/Unicorn-s-Club/whats-unicorn/meow"
	"github.com/Unicorn-s-Club/whats-unicorn/schemas"
	"github.com/gin-gonic/gin"
	"go.mau.fi/whatsmeow/types"
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

func formatMessage(message string) string {
	message = strings.ReplaceAll(message, "/l", "\n")
	message = strings.ReplaceAll(message, "/k", "\n - ")
	return message
}
func SendMessageHandler(ctx *gin.Context) {
	request := SendMessageRequest{}

	ctx.BindJSON(&request)

	// Convert string to types.JID
	recipient, err := types.ParseJID(request.Sender)
	if err != nil {
		Logger.Errorf("error parsing recipient: %v", err)
		return
	}

	Logger.Debugf("request: %v", request)
	Logger.Debugf("recipient: %v", recipient)
	client := meow.GetClient()

	meow.SendMessageHandler(client, recipient, formatMessage(request.Message))

	sendMessage := schemas.SendMessageResponse{
		Message: request.Message,
		Sender:  recipient,
	}

	SendSuccess(ctx, "send-message", sendMessage)
}
