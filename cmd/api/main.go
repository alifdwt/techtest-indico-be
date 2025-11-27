package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alifdwt/techtest-indico-be/internal/config"
	"github.com/alifdwt/techtest-indico-be/internal/handler"
	"github.com/alifdwt/techtest-indico-be/internal/repository"
	"github.com/alifdwt/techtest-indico-be/internal/routes"
	"github.com/alifdwt/techtest-indico-be/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the token.

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	cfg := config.LoadConfig()
	config.SetupLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(ctx, config.LoadConfig().Database.GetDSN())
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	repo := repository.New(connPool)

	authService := service.NewAuthService()
	voucherService := service.NewVoucherService(repo)

	authHandler := handler.NewAuthHandler(authService)
	voucherHandler := handler.NewVoucherHandler(voucherService)

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupVoucherRoutes(router, voucherHandler)
	routes.SetupHealthRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 1. Buat object http.Server secara eksplisit
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	stop()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
