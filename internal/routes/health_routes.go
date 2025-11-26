package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupHealthRoutes(router *gin.Engine) {
	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"status":  "ok",
				"message": "Technical test for Indico API is running successfully",
			})
		})
	}
}
