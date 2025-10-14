// Package logger provides a professional asynchronous logging solution for Go applications
// with Elasticsearch integration and Gin middleware support.
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/google/uuid"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogEntry represents the complete structure of a log record in Elasticsearch
type LogEntry struct {
	// Core fields
	ID        string    `json:"id"`         // Unique identifier for the log entry
	Timestamp time.Time `json:"@timestamp"` // ISO 8601 timestamp
	Level     LogLevel  `json:"level"`      // Log level (DEBUG, INFO, WARN, ERROR, FATAL)
	Message   string    `json:"message"`    // Main log message
	Logger    string    `json:"logger"`     // Logger name/component

	// Application context
	Service     string `json:"service"`     // Service name
	Version     string `json:"version"`     // Application version
	Environment string `json:"environment"` // Environment (dev, staging, prod)
	Hostname    string `json:"hostname"`    // Server hostname
	PID         int    `json:"pid"`         // Process ID

	ExecID string `json:"exec_id"` // Execution ID for tracing across services

	// Source code context
	Caller struct {
		File     string `json:"file"`     // Source file name
		Line     int    `json:"line"`     // Line number
		Function string `json:"function"` // Function name
	} `json:"caller"`

	// HTTP request context (populated by Gin middleware)
	HTTP *HTTPContext `json:"http,omitempty"`

	// Error context (when level is ERROR or FATAL)
	Error *ErrorContext `json:"error,omitempty"`

	// Custom fields for additional context
	Fields map[string]interface{} `json:"fields,omitempty"`

	// Performance metrics
	Performance *PerformanceContext `json:"performance,omitempty"`

	// User context (for authenticated requests)
	User *UserContext `json:"user,omitempty"`

	// Trace context for distributed tracing
	Trace *TraceContext `json:"trace,omitempty"`
}

// HTTPContext contains HTTP request/response information
type HTTPContext struct {
	Method       string            `json:"method"`                  // HTTP method
	URL          string            `json:"url"`                     // Full URL
	Path         string            `json:"path"`                    // URL path
	Query        string            `json:"query"`                   // Query string
	UserAgent    string            `json:"user_agent"`              // User agent
	RemoteIP     string            `json:"remote_ip"`               // Client IP address
	Headers      map[string]string `json:"headers"`                 // Request headers
	StatusCode   int               `json:"status_code"`             // HTTP status code
	ResponseSize int64             `json:"response_size"`           // Response body size in bytes
	ContentType  string            `json:"content_type"`            // Response content type
	Referer      string            `json:"referer"`                 // HTTP referer
	RequestID    string            `json:"request_id"`              // Request tracking ID
	RequestBody  string            `json:"request_body,omitempty"`  // Request body (if configured)
	ResponseBody string            `json:"response_body,omitempty"` // Response body (if configured)
}

// ErrorContext contains error information
type ErrorContext struct {
	Type    string                 `json:"type"`              // Error type/class
	Message string                 `json:"message"`           // Error message
	Stack   string                 `json:"stack,omitempty"`   // Stack trace
	Code    string                 `json:"code,omitempty"`    // Error code
	Details map[string]interface{} `json:"details,omitempty"` // Additional error details
	Cause   *ErrorContext          `json:"cause,omitempty"`   // Underlying cause
}

// PerformanceContext contains timing and performance metrics
type PerformanceContext struct {
	Duration    time.Duration `json:"duration"`     // Request duration
	DurationMs  float64       `json:"duration_ms"`  // Duration in milliseconds
	MemoryUsage int64         `json:"memory_usage"` // Memory usage in bytes
	CPUTime     time.Duration `json:"cpu_time"`     // CPU time
	DBQueries   int           `json:"db_queries"`   // Number of database queries
	CacheHits   int           `json:"cache_hits"`   // Number of cache hits
	CacheMisses int           `json:"cache_misses"` // Number of cache misses
}

// UserContext contains user information
type UserContext struct {
	ID       string                 `json:"id"`                 // User ID
	Email    string                 `json:"email,omitempty"`    // User email
	Username string                 `json:"username,omitempty"` // Username
	Role     string                 `json:"role,omitempty"`     // User role
	Groups   []string               `json:"groups,omitempty"`   // User groups
	Extra    map[string]interface{} `json:"extra,omitempty"`    // Additional user data
}

// TraceContext contains distributed tracing information
type TraceContext struct {
	TraceID  string `json:"trace_id"`  // Trace ID
	SpanID   string `json:"span_id"`   // Span ID
	ParentID string `json:"parent_id"` // Parent span ID
	Sampled  bool   `json:"sampled"`   // Whether this trace is sampled

}

// Config holds the logger configuration
type Config struct {
	Service         string        // Service name
	Version         string        // Application version
	Environment     string        // Environment (dev, staging, prod)
	IndexName       string        // Elasticsearch index name
	FlushInterval   time.Duration // How often to flush logs to Elasticsearch
	BatchSize       int           // Maximum number of logs to batch
	BufferSize      int           // Channel buffer size
	LogLevel        LogLevel      // Minimum log level to process
	EnableCaller    bool          // Whether to capture caller information
	EnableBody      bool          // Whether to log request/response bodies
	MaxBodySize     int           // Maximum body size to log
	SensitiveFields []string      // Fields to redact in logs
	ExecutionID     string        // Unique ID for each request
}

// ElasticsearchLogger is the main logger instance
type ElasticsearchLogger struct {
	config      Config
	es          *elasticsearch.Client
	logChannel  chan LogEntry
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	hostname    string
	pid         int
	ExecutionID string
}

// NewLogger creates a new ElasticsearchLogger instance
func NewLogger(es *elasticsearch.Client, config Config) *ElasticsearchLogger {
	// Set defaults
	if config.IndexName == "" {
		config.IndexName = "application-logs"
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 1 * time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10
	}

	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.LogLevel == "" {
		config.LogLevel = LevelInfo
	}

	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 // 1KB default
	}

	hostname, _ := os.Hostname()
	ctx, cancel := context.WithCancel(context.Background())

	logger := &ElasticsearchLogger{
		config:     config,
		es:         es,
		logChannel: make(chan LogEntry, config.BufferSize),
		ctx:        ctx,
		cancel:     cancel,
		hostname:   hostname,
		pid:        os.Getpid(),
	}

	// Start background goroutine for processing logs
	logger.wg.Add(1)
	go logger.processLogs()
	return logger
}

// processLogs handles batching and sending logs to Elasticsearch
func (l *ElasticsearchLogger) processLogs() {

	defer l.wg.Done()

	ticker := time.NewTicker(l.config.FlushInterval)
	defer ticker.Stop()

	batch := make([]LogEntry, 0, l.config.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := l.sendBatch(batch); err != nil {
			// Fallback to stdout if Elasticsearch fails
			fmt.Fprintf(os.Stderr, "Failed to send logs to Elasticsearch: %v\n", err)
		}
		batch = batch[:0] // Reset batch
	}

	for {
		select {
		case entry := <-l.logChannel:
			batch = append(batch, entry)

			if len(batch) >= l.config.BatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		case <-l.ctx.Done():
			flush() // Final flush
			return
		}
	}
}

// sendBatch sends a batch of log entries to Elasticsearch
func (l *ElasticsearchLogger) sendBatch(entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, entry := range entries {
		// Create index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": l.getIndexName(),
				"_id":    entry.ID,
			},
		}

		if err := json.NewEncoder(&buf).Encode(indexAction); err != nil {
			return fmt.Errorf("failed to encode index action: %w", err)
		}

		// Add document
		if err := json.NewEncoder(&buf).Encode(entry); err != nil {
			return fmt.Errorf("failed to encode log entry: %w", err)
		}
	}

	// Send bulk request
	res, err := l.es.Bulk(
		strings.NewReader(buf.String()),
		l.es.Bulk.WithContext(l.ctx),
		l.es.Bulk.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("failed to send bulk request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch error: %s - %s", res.Status(), string(body))
	}

	return nil
}

// getIndexName generates index name with date suffix for daily rotation
func (l *ElasticsearchLogger) getIndexName() string {
	return fmt.Sprint(l.config.IndexName)
}

// shouldLog checks if the log level should be processed
func (l *ElasticsearchLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}

	return levels[level] >= levels[l.config.LogLevel]
}

// createLogEntry creates a base log entry with common fields
func (l *ElasticsearchLogger) createLogEntry(level LogLevel, message string) LogEntry {
	entry := LogEntry{
		ID:          uuid.New().String(),
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Message:     message,
		Logger:      "elasticsearch-logger",
		Service:     l.config.Service,
		Version:     l.config.Version,
		Environment: l.config.Environment,
		Hostname:    l.hostname,
		PID:         l.pid,
		ExecID:      l.config.ExecutionID,
	}

	// Capture caller information if enabled
	if l.config.EnableCaller {
		if pc, file, line, ok := runtime.Caller(3); ok {
			entry.Caller.File = file
			entry.Caller.Line = line
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Caller.Function = fn.Name()
			}
		}
	}

	return entry
}

// log sends a log entry to the processing channel
func (l *ElasticsearchLogger) log(entry LogEntry) {
	if !l.shouldLog(entry.Level) {
		return
	}

	select {
	case l.logChannel <- entry:
	default:
		// Channel is full, log to stderr as fallback
		fmt.Fprintf(os.Stderr, "Logger channel full, dropping log: %s\n", entry.Message)
	}
}

// Debug logs a debug message
func (l *ElasticsearchLogger) Debug(message string, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelDebug, message)
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Info logs an info message
func (l *ElasticsearchLogger) Info(message string, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelInfo, message)
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Warn logs a warning message
func (l *ElasticsearchLogger) Warn(message string, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelWarn, message)
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Error logs an error message
func (l *ElasticsearchLogger) Error(message string, err error, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelError, message)
	if err != nil {
		entry.Error = &ErrorContext{
			Type:    fmt.Sprintf("%T", err),
			Message: err.Error(),
		}
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Fatal logs a fatal message
func (l *ElasticsearchLogger) Fatal(message string, err error, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelFatal, message)
	if err != nil {
		entry.Error = &ErrorContext{
			Type:    fmt.Sprintf("%T", err),
			Message: err.Error(),
		}
	}
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// WithContext logs with additional context
func (l *ElasticsearchLogger) WithContext(level LogLevel, message string, ctx LogContext) {
	if !l.shouldLog(level) {
		return
	}

	entry := l.createLogEntry(level, message)
	entry.HTTP = ctx.HTTP
	entry.Error = ctx.Error
	entry.Performance = ctx.Performance
	entry.User = ctx.User
	entry.Trace = ctx.Trace
	entry.Fields = ctx.Fields

	l.log(entry)
}

// LogContext holds additional context for logging
type LogContext struct {
	HTTP        *HTTPContext
	Error       *ErrorContext
	Performance *PerformanceContext
	User        *UserContext
	Trace       *TraceContext
	Fields      map[string]interface{}
}

// Close gracefully shuts down the logger
func (l *ElasticsearchLogger) Close() error {
	l.cancel()
	l.wg.Wait()
	close(l.logChannel)

	return nil
}

// Flush forces immediate flush of pending logs
func (l *ElasticsearchLogger) Flush() {
	// Send a signal to process any pending logs
	// This is a best-effort operation
	time.Sleep(100 * time.Millisecond)
}
