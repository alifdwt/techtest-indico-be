package main

import (
	"log"
	"net/http"

	"github.com/alifdwt/techtest-indico-be/internal/config"
	"github.com/alifdwt/techtest-indico-be/internal/handler"
	"github.com/alifdwt/techtest-indico-be/internal/routes"
	"github.com/alifdwt/techtest-indico-be/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/alifdwt/techtest-indico-be/docs"
)

// @title Technical Test Indico API
// @version 1.0
// @description API for managing vouchers
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the token.

func main() {
	cfg := config.LoadConfig()
	config.SetupLogger(cfg)

	authService := service.NewAuthService()
	authHandler := handler.NewAuthHandler(authService)

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	routes.SetupAuthRoutes(router, authHandler)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := ":" + cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
