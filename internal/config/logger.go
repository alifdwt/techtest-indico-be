package config

import (
	"log"

	"github.com/gin-gonic/gin"
)

func SetupLogger(config *Config) {
	if config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Printf("Server running in %s mode", config.Server.Mode)
}
