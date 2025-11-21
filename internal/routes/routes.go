package routes

import (
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/service/healthcheck"
	"orderstreamrest/internal/service/metrics"
	"orderstreamrest/internal/service/terms"
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

	// Métricas: SUPPORT, MANAGER e ADMIN (Auth(3))
	metricsGroup := engine.Group("/metrics", middleware.Auth(3))
	{
		metricsGroup.GET("/tickets", metrics.GetTicketsMetrics(cfg))
		metricsGroup.GET("/tickets/mean-time-resolution-by-priority", metrics.MeanTimeByPriority(cfg))
		metricsGroup.GET("/tickets/qtd-tickets-by-status-year-month", metrics.QtdTicketsByStatusYearMonth(cfg))
		metricsGroup.GET("/tickets/qtd-tickets-by-month", metrics.TicketsByMonth(cfg))
		metricsGroup.GET("/tickets/qtd-tickets-by-priority-year-month", metrics.TicketsByPriorityAndMonth(cfg))
	}

	// Tickets: SUPPORT, MANAGER e ADMIN (Auth(3))
	ticketsGroup := engine.Group("/tickets", middleware.Auth(3))
	{
		ticketsGroup.GET("/:id", tickets.SearchTicketByID(cfg))
		ticketsGroup.GET("/query", tickets.GetByWord(cfg))
	}

	// Gerenciamento de usuários: MANAGER e ADMIN (Auth(2))
	userRoutes := engine.Group("/users", middleware.Auth(2))
	{
		userRoutes.GET("", users.GetAllUsers(cfg))
		userRoutes.GET("/:id", users.GetUser(cfg))
		userRoutes.PUT("/:id", users.UpdateUser(cfg))
	}

	// Endpoints públicos e autenticados (qualquer usuário logado)
	authRoutes := engine.Group("/auth")
	{
		// Públicos
		authRoutes.POST("/login", users.LoginHandler(cfg))
		authRoutes.GET("/microsoft/login", users.MicrosoftLoginHandler())
		authRoutes.GET("/microsoft/callback", users.MicrosoftCallbackHandler(cfg))
		authRoutes.GET("/terms/active", terms.GetActiveTerm(cfg))
		authRoutes.POST("/register", users.CreateUser(cfg))

		// Autenticados (qualquer role)
		authRoutes.POST("/change-password", middleware.Auth(3), users.ChangePassword(cfg))
		authRoutes.DELETE("/:id", middleware.Auth(3), users.DeleteUser(cfg))
	}

	// Gerenciamento de termos: apenas ADMIN (Auth(1))
	termsRoutes := engine.Group("/terms", middleware.Auth(1))
	{
		termsRoutes.GET("", terms.ListTerms(cfg))
		termsRoutes.POST("", terms.CreateTerm(cfg))
	}

	// Consentimentos
	consentsRoutes := engine.Group("/consents")
	{
		// Qualquer usuário autenticado vê seu próprio consentimento
		consentsRoutes.GET("/me", middleware.Auth(3), terms.GetMyConsentStatus(cfg))
		// Apenas ADMIN vê consentimento de outros usuários
		consentsRoutes.GET("/user/:userId", middleware.Auth(1), terms.GetUserConsent(cfg))
	}

}
