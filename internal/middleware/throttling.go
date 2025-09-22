package middleware

import (
	"context"
	"fmt"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	redisInternal "orderstreamrest/internal/repositories/redis"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/semaphore"
)

const (
	defaultMaxRequests = 5
	rateLimitWindow    = 60 * time.Second
)

// RateLimiter encapsula a lógica de rate limiting
type RateLimiter struct {
	redis       *redisInternal.RedisInternal
	maxRequests int
	window      time.Duration
}

// NewRateLimiter cria uma nova instância do rate limiter
func NewRateLimiter(redisClient *redisInternal.RedisInternal, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:       redisClient,
		maxRequests: maxRequests,
		window:      window,
	}
}

// setupRedisDB configura o middleware de rate limiting
func setupRedisDB(engine *gin.Engine, cfg *config.App) {
	// Limpa o Redis (opcional - considere remover em produção)
	cfg.Redis.FlushAll(context.Background())

	// Obtém a configuração do limite máximo
	maxRequests := int(getEnvAsInt64("MAX_REQUEST_COUNT_BY_IP", defaultMaxRequests))

	// Cria o rate limiter
	rateLimiter := NewRateLimiter(cfg.Redis, maxRequests, rateLimitWindow)

	// Adiciona o middleware
	engine.Use(rateLimiter.Middleware())
}

// Middleware retorna o middleware do Gin para rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Permite requisições para qualquer rota que contenha "swagger" sem rate limiting
		if strings.Contains(c.FullPath(), "swagger") {
			c.Next()
			return
		}

		ip := c.ClientIP()

		allowed, retryAfter, err := rl.checkRateLimit(c.Request.Context(), ip)
		if err != nil {
			rl.handleError(c, err)
			return
		}

		if !allowed {
			rl.handleRateLimitExceeded(c, retryAfter)
			return
		}

		c.Next()
	}
}

// checkRateLimit verifica se o IP pode fazer a requisição
func (rl *RateLimiter) checkRateLimit(ctx context.Context, ip string) (allowed bool, retryAfter time.Duration, err error) {
	// Tenta obter o contador atual
	val, err := rl.redis.Get(ctx, ip).Result()

	// Primeira requisição do IP
	if err == redis.Nil {
		err = rl.redis.Set(ctx, ip, 1, rl.window).Err()
		if err != nil {
			return false, 0, err
		}
		return true, 0, nil
	}

	// Erro ao acessar Redis
	if err != nil {
		return false, 0, err
	}

	// Converte o valor para int
	requestCount, err := strconv.Atoi(val)
	if err != nil {
		return false, 0, err
	}

	// Verifica se excedeu o limite
	if requestCount >= rl.maxRequests {
		ttl, err := rl.redis.TTL(ctx, ip).Result()
		if err != nil {
			return false, 0, err
		}
		return false, ttl, nil
	}

	// Incrementa o contador
	err = rl.redis.Incr(ctx, ip).Err()
	if err != nil {
		return false, 0, err
	}

	return true, 0, nil
}

// handleError trata erros internos
func (rl *RateLimiter) handleError(c *gin.Context, err error) {
	errorResponse := dto.NewErrorResponse(
		c,
		http.StatusInternalServerError,
		"internal_server_error",
		"Internal server error",
		map[string]interface{}{
			"original_error": err.Error(),
		},
	)

	c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
}

// handleRateLimitExceeded trata quando o limite é excedido
func (rl *RateLimiter) handleRateLimitExceeded(c *gin.Context, retryAfter time.Duration) {
	// Adicionar headers de rate limiting
	c.Writer.Header().Set("Retry-After", retryAfter.String())
	c.Writer.Header().Set("X-RateLimit-Limit", "100") // ajuste conforme sua configuração
	c.Writer.Header().Set("X-RateLimit-Remaining", "0")
	c.Writer.Header().Set("X-RateLimit-Reset", time.Now().Add(retryAfter).Format(time.RFC3339))

	errorResponse := dto.NewRateLimitErrorResponse(
		c,
		retryAfter.String(),
		100, // limite por minuto - ajuste conforme sua configuração
		0,   // requests restantes
		time.Now().Add(retryAfter),
	)

	c.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse)
}
func setupSemaphore(engine *gin.Engine) {
	max := getEnvAsInt64("MAX_REQUEST_COUNT_GLOBAL", int64(10))
	sema := semaphore.NewWeighted(max)

	engine.Use(func(c *gin.Context) {
		if err := sema.Acquire(c.Request.Context(), 1); err != nil {
			errorResponse := dto.NewRateLimitErrorResponse(
				c,
				"60s", // retry after 60 seconds
				int(max),
				0,
				time.Now().Add(time.Minute),
			)

			// Adicionar headers de rate limiting
			c.Writer.Header().Set("Retry-After", "60")
			c.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", max))
			c.Writer.Header().Set("X-RateLimit-Remaining", "0")
			c.Writer.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Minute).Format(time.RFC3339))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse)
			return
		}
		defer sema.Release(1)
		c.Next()
	})
}
