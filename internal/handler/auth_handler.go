package handler

import (
	"net/http"

	"github.com/alifdwt/techtest-indico-be/internal/dto"
	"github.com/alifdwt/techtest-indico-be/internal/service"
	"github.com/alifdwt/techtest-indico-be/internal/util"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return token
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login request"
// @Success 200 {object} util.Response{data=dto.LoginResponse}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /login [post]
func (ah *AuthHandler) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := dto.ValidateStruct(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	res, err := ah.authService.Login(&req)
	if err != nil {
		util.ErrorResponse(ctx, http.StatusInternalServerError, "Login failed: "+err.Error())
		return
	}

	util.SuccessResponse(ctx, http.StatusOK, "Login success", res)
}
