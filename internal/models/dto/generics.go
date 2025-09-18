// Package dto contains Data Transfer Objects for API responses
package dto

import (
	"time"

	"github.com/gin-gonic/gin"
)

// BaseResponse contém campos comuns a todas as respostas
type BaseResponse struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// SuccessResponse representa uma resposta de sucesso
type SuccessResponse struct {
	BaseResponse
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	BaseResponse
	Error   string      `json:"error"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginatedResponse representa uma resposta paginada
type PaginatedResponse struct {
	BaseResponse
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Message    string      `json:"message,omitempty"`
}

// Pagination contém informações de paginação
type Pagination struct {
	CurrentPage  int   `json:"current_page" example:"1"`
	PerPage      int   `json:"per_page" example:"10"`
	TotalPages   int   `json:"total_pages" example:"5"`
	TotalRecords int64 `json:"total_records" example:"50"`
	HasNext      bool  `json:"has_next" example:"true"`
	HasPrev      bool  `json:"has_prev" example:"false"`
}

// HealthResponse representa a resposta do healthcheck
type HealthResponse struct {
	BaseResponse
	Status  string            `json:"status" example:"OK"`
	Service string            `json:"service" example:"VisionData API"`
	Version string            `json:"version" example:"1.0.0"`
	Uptime  string            `json:"uptime,omitempty" example:"1h30m45s"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// AuthErrorResponse representa erros específicos de autenticação
type AuthErrorResponse struct {
	BaseResponse
	Error    string `json:"error" example:"unauthorized"`
	Code     int    `json:"code" example:"401"`
	Message  string `json:"message" example:"Token de autorização inválido ou expirado"`
	LoginURL string `json:"login_url,omitempty" example:"/auth/login"`
}

// RateLimitErrorResponse representa erros de rate limit
type RateLimitErrorResponse struct {
	BaseResponse
	Error      string    `json:"error" example:"rate_limit_exceeded"`
	Code       int       `json:"code" example:"429"`
	Message    string    `json:"message" example:"Limite de requisições excedido"`
	RetryAfter string    `json:"retry_after" example:"60s"`
	Limit      int       `json:"limit" example:"100"`
	Remaining  int       `json:"remaining" example:"0"`
	ResetTime  time.Time `json:"reset_time" example:"2024-01-01T12:01:00Z"`
}

// Helper functions para criar responses padronizadas

// NewSuccessResponse cria uma nova resposta de sucesso
func NewSuccessResponse(c *gin.Context, data interface{}, message string) SuccessResponse {
	return SuccessResponse{
		BaseResponse: BaseResponse{
			Success:   true,
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Data:    data,
		Message: message,
	}
}

// NewErrorResponse cria uma nova resposta de erro
func NewErrorResponse(c *gin.Context, code int, error string, message string, details interface{}) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{
			Success:   false,
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Error:   error,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewPaginatedResponse cria uma nova resposta paginada
func NewPaginatedResponse(c *gin.Context, data interface{}, pagination Pagination, message string) PaginatedResponse {
	return PaginatedResponse{
		BaseResponse: BaseResponse{
			Success:   true,
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Data:       data,
		Pagination: pagination,
		Message:    message,
	}
}

// NewHealthResponse cria uma nova resposta de health
func NewHealthResponse(c *gin.Context, status, service, version, uptime string, checks map[string]string) HealthResponse {
	return HealthResponse{
		BaseResponse: BaseResponse{
			Success:   status == "OK",
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Status:  status,
		Service: service,
		Version: version,
		Uptime:  uptime,
		Checks:  checks,
	}
}

// NewAuthErrorResponse cria uma nova resposta de erro de autenticação
func NewAuthErrorResponse(c *gin.Context, message string) AuthErrorResponse {
	return AuthErrorResponse{
		BaseResponse: BaseResponse{
			Success:   false,
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Error:    "unauthorized",
		Code:     401,
		Message:  message,
		LoginURL: "/auth/login",
	}
}

// NewRateLimitErrorResponse cria uma nova resposta de rate limit
func NewRateLimitErrorResponse(c *gin.Context, retryAfter string, limit, remaining int, resetTime time.Time) RateLimitErrorResponse {
	return RateLimitErrorResponse{
		BaseResponse: BaseResponse{
			Success:   false,
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Error:      "rate_limit_exceeded",
		Code:       429,
		Message:    "Limite de requisições excedido",
		RetryAfter: retryAfter,
		Limit:      limit,
		Remaining:  remaining,
		ResetTime:  resetTime,
	}
}

// getRequestID extrai o request ID do contexto
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
