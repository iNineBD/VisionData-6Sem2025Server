package routes

import (
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/service/healthcheck"
	"orderstreamrest/internal/service/metrics"
	"orderstreamrest/internal/service/tickets"
	"orderstreamrest/internal/service/users"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitiateRoutes is a function that initializes the routes for the application
func InitiateRoutes(engine *gin.Engine, cfg *config.App) {

	users.InitOAuthConfig()

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	healthGroup := engine.Group("/healthcheck")
	{
		healthGroup.GET("/", healthcheck.Health(cfg))
	}

	metricsGroup := engine.Group("/metrics", middleware.Auth(1))
	{
		metricsGroup.GET("/tickets", metrics.GetTicketsMetrics(cfg))
		metricsGroup.GET("/tickets/mean-time-resolution-by-priority", metrics.MeanTimeByPriority(cfg))
		metricsGroup.GET("/tickets/qtd-tickets-by-status-year-month", metrics.QtdTicketsByStatusYearMonth(cfg))
		metricsGroup.GET("/tickets/qtd-tickets-by-month", metrics.TicketsByMonth(cfg))
		metricsGroup.GET("/tickets/qtd-tickets-by-priority-year-month", metrics.TicketsByPriorityAndMonth(cfg))
	}

	ticketsGroup := engine.Group("/tickets", middleware.Auth(1))
	{
		ticketsGroup.GET("/:id", tickets.SearchTicketByID(cfg))
		ticketsGroup.GET("/query", tickets.GetByWord(cfg))
	}

	userRoutes := engine.Group("/users", middleware.Auth(2))
	{
		userRoutes.POST("", users.CreateUser(cfg))
		userRoutes.GET("", users.GetAllUsers(cfg))
		userRoutes.GET("/:id", users.GetUser(cfg))
		userRoutes.PUT("/:id", users.UpdateUser(cfg))
		userRoutes.DELETE("/:id", users.DeleteUser(cfg))

		userRoutes.POST("/change-password", users.ChangePassword(cfg))
	}

	authRoutes := engine.Group("/auth")
	{
		authRoutes.POST("/login", users.LoginHandler(cfg))

		authRoutes.GET("/microsoft/login", users.MicrosoftLoginHandler())

		authRoutes.GET("/microsoft/callback", users.MicrosoftCallbackHandler(cfg))
	}

}
