package routes

import (
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/service/healthcheck"
	"orderstreamrest/internal/service/tickets"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitiateRoutes is a function that initializes the routes for the application
func InitiateRoutes(engine *gin.Engine, cfg *config.App) {

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	healthGroup := engine.Group("/healthcheck", middleware.Auth())
	healthGroup.GET("/", healthcheck.Health(cfg))

	// Inicializar rotas de métricas
	SetupMetricsRoutes(engine, cfg)

	ticketsGroup := engine.Group("/tickets")
	ticketsGroup.GET("/:id", tickets.SearchTicketByID(cfg))
	ticketsGroup.GET("/query", tickets.GetByWord(cfg))

}
