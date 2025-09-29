package middleware

import (
	"log"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/utils"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

// sets up a new gin engine with a semaphore and cors middleware
func SetupServer(rd *config.App) (engine *gin.Engine) {

	gin.SetMode(gin.ReleaseMode)
	engine = gin.New()

	setupSemaphore(engine)
	setupCors(engine)
	setupRedisDB(engine, rd)
	setupLogger(engine, rd.Logger)
	setupIds(engine)

	certFile, keyFile := utils.GetCertFiles()
	if certFile != "" && keyFile != "" {
		setupSSL(engine)
	}

	engine.Use(gin.Recovery())

	return engine
}

// setupSSL is a function that sets up the SSL configuration for the server
func setupSSL(engine *gin.Engine) {
	engine.Use(func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     ":8080",
		})
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			log.Println("Error traying make a secure https: " + err.Error())
			return
		}
		c.Next()
	})
}

func getEnvAsInt64(name string, defaultValue int64) int64 {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return int64(value)
}
