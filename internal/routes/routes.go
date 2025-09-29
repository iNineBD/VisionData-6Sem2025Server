package routes

import (
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/service/healthcheck"
	"orderstreamrest/internal/service/metrics"
	"orderstreamrest/internal/service/tickets"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitiateRoutes is a function that initializes the routes for the application
func InitiateRoutes(engine *gin.Engine, cfg *config.App) {

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	healthGroup := engine.Group("/healthcheck")
	healthGroup.GET("/", healthcheck.Health(cfg))

	metricsGroup := engine.Group("/metrics")
	metricsGroup.GET("/tickets", metrics.GetTicketsMetrics(cfg))

	ticketsGroup := engine.Group("/tickets")
	ticketsGroup.GET("/:id", tickets.SearchTicketByID(cfg))
	ticketsGroup.GET("/query", tickets.GetByWord(cfg))

}
