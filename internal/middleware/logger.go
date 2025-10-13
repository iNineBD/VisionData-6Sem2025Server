package middleware

import (
	"bytes"
	"io"
	"orderstreamrest/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// setupLogger -
func setupLogger(engine *gin.Engine, logger *logger.ElasticsearchLogger) {

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
	// Function to extract user context from Gin context
	UserExtractor func(*gin.Context) *logger.UserContext
	// Function to extract trace context from Gin context
	TraceExtractor func(*gin.Context) *logger.TraceContext
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
			"/swagger",
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
func LoggerMiddleware(esLogger *logger.ElasticsearchLogger, config ...MiddlewareConfig) gin.HandlerFunc {
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

		// Build query string
		// if raw != "" {
		// 	path = fmt.Sprintf("%s?%s", path, raw)
		// }

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

		// Build HTTP context
		httpContext := &logger.HTTPContext{
			Method:       c.Request.Method,
			URL:          c.Request.URL.String(),
			Path:         c.Request.URL.Path,
			Query:        c.Request.URL.RawQuery,
			UserAgent:    c.Request.UserAgent(),
			RemoteIP:     c.ClientIP(),
			Headers:      headers,
			StatusCode:   statusCode,
			ResponseSize: int64(c.Writer.Size()),
			ContentType:  c.Writer.Header().Get("Content-Type"),
			Referer:      c.Request.Referer(),
			RequestID:    requestID,
			RequestBody:  requestBody,
			ResponseBody: responseBody,
		}

		// Build performance context
		performanceContext := &logger.PerformanceContext{
			Duration:   duration,
			DurationMs: float64(duration.Nanoseconds()) / 1e6,
		}

		// Extract user context if extractor is provided
		var userContext *logger.UserContext
		if cfg.UserExtractor != nil {
			userContext = cfg.UserExtractor(c)
		}

		// Extract trace context if extractor is provided
		var traceContext *logger.TraceContext
		if cfg.TraceExtractor != nil {
			traceContext = cfg.TraceExtractor(c)
		}

		// Collect any errors from the context
		var errorContext *logger.ErrorContext
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()
			errorContext = &logger.ErrorContext{
				Message: lastError.Error(),
			}
		}

		// Determine log level based on status code
		var level logger.LogLevel
		var message string

		switch {
		case statusCode >= 500:
			level = logger.LevelError
			message = "HTTP Server Error"
		case statusCode >= 400:
			level = logger.LevelWarn
			message = "HTTP Client Error"
		case statusCode >= 300:
			level = logger.LevelInfo
			message = "HTTP Redirect"
		default:
			level = logger.LevelInfo
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

		// Log the request
		logContext := logger.LogContext{
			HTTP:        httpContext,
			Performance: performanceContext,
			User:        userContext,
			Trace:       traceContext,
			Error:       errorContext,
			Fields:      fields,
		}

		esLogger.WithContext(level, message, logContext)
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
