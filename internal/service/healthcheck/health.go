package healthcheck

import (
	"fmt"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// Health godoc
// @Summary      Health Check
// @Description  Verifica a saúde do serviço. Endpoint protegido com autenticação e rate limiting.
// @Tags         health
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.HealthResponse           "Status do serviço"
// @Failure      400  {object}  dto.ErrorResponse            "Bad Request"
// @Failure      401  {object}  dto.AuthErrorResponse        "Unauthorized - Token inválido"
// @Failure      403  {object}  dto.ErrorResponse            "Forbidden - Sem permissão"
// @Failure      429  {object}  dto.RateLimitErrorResponse   "Rate limit excedido"
// @Failure      500  {object}  dto.ErrorResponse            "Internal Server Error"
// @Header       200  {string}  X-RateLimit-Limit            "Limite de requests por minuto"
// @Header       200  {string}  X-RateLimit-Remaining        "Requests restantes no período"
// @Header       200  {string}  X-RateLimit-Reset            "Timestamp do reset do rate limit"
// @Router       /healthcheck [get]
func Health(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg.Logger.Info(fmt.Sprintf("Healthcheck endpoint hit... IP %s", c.ClientIP()))

		// Removi o log de erro desnecessário
		// cfg.Logger.Error("Vish, deu erro aqui", errors.New("Ola"))

		// Verificações de saúde do sistema
		checks := make(map[string]string)

		// Verificar conexão com Redis (exemplo)
		if cfg.Redis != nil {
			checks["redis"] = "OK"
		} else {
			checks["redis"] = "UNAVAILABLE"
		}

		// Verificar outras dependências
		checks["database"] = "OK" // substitua pela verificação real
		checks["memory"] = "OK"   // você pode adicionar verificação de memória

		// Determinar status geral
		status := "OK"
		for _, checkStatus := range checks {
			if checkStatus != "OK" {
				status = "DEGRADED"
				break
			}
		}

		uptime := time.Since(startTime).String()

		healthResponse := dto.NewHealthResponse(
			c,
			status,
			"VisionData API",
			"1.0.0",
			uptime,
			checks,
		)

		cfg.Logger.Info(fmt.Sprintf("Healthcheck status: %s", status))

		// Status HTTP baseado no status das verificações
		httpStatus := http.StatusOK
		if status != "OK" {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, healthResponse)
	}
}
