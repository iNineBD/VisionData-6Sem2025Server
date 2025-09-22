package routes

import (
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/service/metrics"

	"github.com/gin-gonic/gin"
)

// SetupMetricsRoutes sets up the routes for metrics
func SetupMetricsRoutes(engine *gin.Engine, cfg *config.App) {

	metricsGroup := engine.Group("/metrics", middleware.Auth())

	metricsGroup.GET("/tickets", metrics.GetTicketsMetrics(cfg))
}
