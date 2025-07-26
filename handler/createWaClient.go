package handler

import (
	"log"

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
func CreateWaClientHandler(ctx *gin.Context) {

	device := meow.CreateDevice()

	client := meow.CreateClient(device)
	meow.ShowQrCode(client)

	myClient := &meow.MyClient{WAClient: client}
	myClient.Register()

	err := client.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
}
