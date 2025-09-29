package main

import (
	"fmt"
	"log"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/routes"
	"orderstreamrest/internal/utils"
	"os"

	_ "orderstreamrest/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title           VisionData API
// @version         1.0
// @description     API REST para aplicação VisionData com recursos de autenticação, rate limiting e monitoramento.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Inine
// @contact.email  https://github.com/iNineBD

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// Definições de responses globais reutilizáveis:
// @response Unauthorized {object} dto.AuthErrorResponse "Token inválido ou ausente"
// @response Forbidden {object} dto.ErrorResponse "Acesso negado"
// @response RateLimited {object} dto.RateLimitErrorResponse "Rate limit excedido"
// @response InternalServerError {object} dto.ErrorResponse "Erro interno do servidor"
// @response BadRequest {object} dto.ErrorResponse "Requisição inválida"

func main() {
	// Carregar variáveis de ambiente
	envPath := "/app/.env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = "../../.env"
	}
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	fmt.Printf("Environment: %s\n", os.Getenv("ENVIRONMENT_APP"))

	// Inicializar configuração
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}
	defer cfg.CloseAll()

	cfg.Logger.Info(fmt.Sprintf(
		"Starting VisionData API | execution_id=%s | version=1.0.0",
		os.Getenv("ENVIRONMENT_APP"),
	))

	// Setup do servidor
	engine := middleware.SetupServer(cfg)

	// Inicializar rotas
	routes.InitiateRoutes(engine, cfg)

	// Iniciar servidor
	startServer(engine, cfg)
}
func startServer(engine *gin.Engine, cfg *config.App) {
	certFile, keyFile := utils.GetCertFiles()

	if certFile != "" && keyFile != "" {
		cfg.Logger.Info(
			fmt.Sprintf("Starting server with TLS on port 8080, cert_file=%s, key_file=%s", certFile, keyFile),
		)

		if err := engine.RunTLS(":8080", certFile, keyFile); err != nil {
			cfg.Logger.Fatal(
				fmt.Sprintf("Error starting TLS server on port 8080: %v", err),
			)
		}
	} else {
		cfg.Logger.Info(
			"Starting server without TLS on port 8080",
		)

		if err := engine.Run(":8080"); err != nil {
			cfg.Logger.Fatal(
				fmt.Sprintf("Error starting server on port 8080: %v", err),
			)
		}
	}
}
