package routes

import (
	"github.com/alifdwt/techtest-indico-be/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
) {
	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/login", authHandler.Login)
	}
}
