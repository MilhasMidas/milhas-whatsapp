package user

import (
	"net/http"

	"github.com/Unicorn-s-Club/whats-unicorn/handler"
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
func CreateUserHandler(ctx *gin.Context) {
	request := schemas.CreateUserRequest{}

	ctx.BindJSON(&request)

	handler.Logger.Debugf("request: %v", request)

	user := schemas.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Active:   true,
	}

	if err := handler.Db.Create(&user).Error; err != nil {
		handler.Logger.Errorf("error creating user: %v", err.Error())
		handler.SendError(ctx, http.StatusInternalServerError, "error creating user on database")
		return
	}

	handler.SendSuccess(ctx, "create-user", user)
}
