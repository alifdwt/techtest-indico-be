package middleware

import (
	"net/http"
	"strings"

	"github.com/alifdwt/techtest-indico-be/internal/util"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			util.ErrorResponse(ctx, http.StatusUnauthorized, "Authorization header is missing")
			ctx.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString != "iniadalahtokenbohongan" {
			util.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
