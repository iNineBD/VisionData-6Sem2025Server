package main

import (
	"fmt"
	"log"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/routes"
	"orderstreamrest/internal/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	envPath := "/app/.env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = "../../.env"
	}
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	fmt.Println((os.Getenv("ENVIRONMENT_APP")))

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}
	defer cfg.CloseAll()

	cfg.Logger.Info(fmt.Sprintf("Starting server with execution ID %s", cfg.Logger.ExecutionID))

	engine := middleware.SetupServer(cfg)

	routes.InitiateRoutes(engine, cfg)

	startServer(engine)
}

func startServer(engine *gin.Engine) {
	certFile, keyFile := utils.GetCertFiles()
	if certFile != "" && keyFile != "" {
		log.Println("Starting server with TLS...")
		if err := engine.RunTLS(":8080", certFile, keyFile); err != nil {
			log.Fatalf("Error starting TLS server: %v", err)
		}
	} else {
		log.Println("Starting server...")
		if err := engine.Run(":8080"); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}
	log.Println("Server started on port 8080")
}
