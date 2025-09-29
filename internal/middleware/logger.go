package middleware

import (
	"bytes"
	"fmt"
	"io"
	"orderstreamrest/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// setupLogger -
func setupLogger(engine *gin.Engine, logger *logger.Logger) {

	middlewareConfig := MiddlewareConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     2048,
		ExcludedHeaders: []string{
			"authorization",
			"cookie",
			"x-api-key",
		},
		SkipPaths: []string{
			"/health",
			"/metrics",
		},
		ErrorsOnly:      false,
		RequestIDHeader: "X-Request-ID",
	}
	engine.Use(LoggerMiddleware(logger, middlewareConfig))
}

// MiddlewareConfig configures the logging middleware
type MiddlewareConfig struct {
	// Whether to log request bodies
	LogRequestBody bool
	// Whether to log response bodies
	LogResponseBody bool
	// Maximum size of bodies to log (in bytes)
	MaxBodySize int
	// Headers to exclude from logging (case-insensitive)
	ExcludedHeaders []string
	// Paths to skip logging (exact match)
	SkipPaths []string
	// Whether to log only errors (4xx, 5xx status codes)
	ErrorsOnly bool
	// Custom request ID header name
	RequestIDHeader string
}

// DefaultMiddlewareConfig returns a default configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     1024, // 1KB
		ExcludedHeaders: []string{
			"authorization",
			"cookie",
			"set-cookie",
			"x-api-key",
			"x-auth-token",
		},
		SkipPaths: []string{
			"/health",
		},
		ErrorsOnly:      false,
		RequestIDHeader: "X-Request-ID",
	}
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(data []byte) (int, error) {
	// Capture response body if configured
	if w.body != nil && w.body.Len()+len(data) <= cap(w.body.Bytes()) {
		w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

// LoggerMiddleware creates a Gin middleware that logs HTTP requests
func LoggerMiddleware(esLogger *logger.Logger, config ...MiddlewareConfig) gin.HandlerFunc {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Convert excluded headers to lowercase for case-insensitive comparison
	excludedHeaders := make(map[string]bool)
	for _, header := range cfg.ExcludedHeaders {
		excludedHeaders[strings.ToLower(header)] = true
	}

	// Convert skip paths to map for faster lookup
	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()

		// Generate or extract request ID
		requestID := c.GetHeader(cfg.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header(cfg.RequestIDHeader, requestID)
		}

		// Store request ID in context for use in handlers
		c.Set("request_id", requestID)

		// Read request body if configured
		var requestBody string
		if cfg.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore body for further processing
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Capture body if within size limit
				if len(bodyBytes) <= cfg.MaxBodySize {
					requestBody = string(bodyBytes)
				} else {
					requestBody = "[BODY TOO LARGE]"
				}
			}
		}

		// Prepare response body capture
		var responseBodyBuf *bytes.Buffer
		if cfg.LogResponseBody {
			responseBodyBuf = bytes.NewBuffer(make([]byte, 0, cfg.MaxBodySize))
			c.Writer = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           responseBodyBuf,
			}
		}

		// Process request
		c.Next()

		// Calculate duration and collect metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Skip logging if ErrorsOnly is enabled and status is not an error
		if cfg.ErrorsOnly && statusCode < 400 {
			return
		}

		// Collect headers (excluding sensitive ones)
		headers := make(map[string]string)
		for name, values := range c.Request.Header {
			lowerName := strings.ToLower(name)
			if !excludedHeaders[lowerName] && len(values) > 0 {
				headers[name] = values[0]
			}
		}

		// Get response body
		var responseBody string
		if cfg.LogResponseBody && responseBodyBuf != nil {
			if responseBodyBuf.Len() <= cfg.MaxBodySize {
				responseBody = responseBodyBuf.String()
			} else {
				responseBody = "[RESPONSE TOO LARGE]"
			}
		}

		// Determine log level based on status code
		var message string

		switch {
		case statusCode >= 500:
			message = "HTTP Server Error"
		case statusCode >= 400:
			message = "HTTP Client Error"
		case statusCode >= 300:
			message = "HTTP Redirect"
		default:
			message = "HTTP Request"
		}

		// Build additional fields
		fields := map[string]interface{}{
			"component": "http_middleware",
		}

		// Add custom fields from context if available
		if customFields, exists := c.Get("log_fields"); exists {
			if fieldMap, ok := customFields.(map[string]interface{}); ok {
				for k, v := range fieldMap {
					fields[k] = v
				}
			}
		}

		esLogger.Debug(fmt.Sprintf("%s - %s %s %d %s\n\nRequest Body: %s\n\nResponse Body: %s",
			message, c.Request.Method, c.Request.URL.Path, statusCode, duration.String(), requestBody, responseBody))
	}
}

// AddLogFields adds custom fields to be included in logs
func AddLogFields(c *gin.Context, fields map[string]interface{}) {
	existing, exists := c.Get("log_fields")
	if !exists {
		c.Set("log_fields", fields)
		return
	}

	if existingMap, ok := existing.(map[string]interface{}); ok {
		for k, v := range fields {
			existingMap[k] = v
		}
		c.Set("log_fields", existingMap)
	} else {
		c.Set("log_fields", fields)
	}
}
