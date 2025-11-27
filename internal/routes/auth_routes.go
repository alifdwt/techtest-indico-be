package routes

import (
	"github.com/alifdwt/techtest-indico-be/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
) {
	login := router.Group("/login")
	{
		login.POST("", authHandler.Login)
	}
}
