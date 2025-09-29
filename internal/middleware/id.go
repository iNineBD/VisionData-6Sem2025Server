package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// setupIds -
func setupIds(engine *gin.Engine) {
	engine.Use(RequestIDMiddleware(""))
}

// GetRequestID retrieves the request ID from Gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RequestIDMiddleware adds request ID to context if not present
func RequestIDMiddleware(headerName string) gin.HandlerFunc {
	if headerName == "" {
		headerName = "X-Request-ID"
	}

	return func(c *gin.Context) {
		requestID := c.GetHeader(headerName)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header(headerName, requestID)
		}
		c.Set("request_id", requestID)
		c.Next()
	}
}
