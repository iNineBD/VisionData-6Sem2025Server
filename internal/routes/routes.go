package routes

import (
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/service/healthcheck"

	"github.com/gin-gonic/gin"
)

// InitiateRoutes is a function that initializes the routes for the application
func InitiateRoutes(engine *gin.Engine, cfg *config.App) {
	healthGroup := engine.Group("/healthcheck", middleware.Auth())

	healthGroup.GET("/", healthcheck.Health(cfg))

}
