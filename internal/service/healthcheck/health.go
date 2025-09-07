package healthcheck

import (
	"errors"
	"orderstreamrest/internal/config"

	"github.com/gin-gonic/gin"
)

// Health - Healthcheck endpoint
func Health(cfg *config.App) gin.HandlerFunc {

	return func(c *gin.Context) {

		cfg.Logger.Info("Healthcheck endpoint hit")
		cfg.Logger.Error("Vish, deu erro aqui", errors.New("Ola"))

		c.JSON(200, gin.H{
			"status": "OK",
		})
	}
}
