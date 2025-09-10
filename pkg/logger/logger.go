// Package logger provides a professional asynchronous logging solution for Go applications
// with file-based storage and automatic rotation.
package logger

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

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

// LogEntry represents the complete structure of a log record
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
	LogDir          string        // Directory for log files
	FlushInterval   time.Duration // How often to flush logs to file
	BatchSize       int           // Maximum number of logs to batch
	BufferSize      int           // Channel buffer size
	LogLevel        LogLevel      // Minimum log level to process
	EnableCaller    bool          // Whether to capture caller information
	EnableBody      bool          // Whether to log request/response bodies
	MaxBodySize     int           // Maximum body size to log
	SensitiveFields []string      // Fields to redact in logs
	ExecutionID     string        // Unique ID for each request
	MaxFileSize     int64         // Maximum file size in bytes (default 10MB)
	BufferSize64KB  int           // Buffer size for file writer (default 64KB)
}

// fileWriter manages the current log file with buffering
type fileWriter struct {
	mu           sync.RWMutex
	file         *os.File
	writer       *bufio.Writer
	currentSize  int64
	currentDate  string
	currentIndex int
	maxSize      int64
	logDir       string
	bufferSize   int
}

// FileLogger is the main logger instance
type FileLogger struct {
	config      Config
	logChannel  chan LogEntry
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	hostname    string
	pid         int
	ExecutionID string
	fileWriter  *fileWriter
}

// NewLogger creates a new FileLogger instance
func NewLogger(config Config) *FileLogger {
	// Set defaults
	if config.LogDir == "" {
		config.LogDir = "./logs"
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 1 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100 // Aumentado para melhor performance
	}
	if config.BufferSize == 0 {
		config.BufferSize = 10000 // Aumentado buffer
	}
	if config.LogLevel == "" {
		config.LogLevel = LevelInfo
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 10MB
	}
	if config.BufferSize64KB == 0 {
		config.BufferSize64KB = 64 * 1024 // 64KB buffer
	}

	hostname, _ := os.Hostname()
	ctx, cancel := context.WithCancel(context.Background())

	// Create log directory
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
	}

	fw := &fileWriter{
		maxSize:    config.MaxFileSize,
		logDir:     config.LogDir,
		bufferSize: config.BufferSize64KB,
	}

	logger := &FileLogger{
		config:     config,
		logChannel: make(chan LogEntry, config.BufferSize),
		ctx:        ctx,
		cancel:     cancel,
		hostname:   hostname,
		pid:        os.Getpid(),
		fileWriter: fw,
	}

	// Start background goroutine for processing logs
	logger.wg.Add(1)
	go logger.processLogs()

	return logger
}

// ensureCurrentFile ensures we have a valid file for the current date
func (fw *fileWriter) ensureCurrentFile() error {
	currentDate := time.Now().Format("2006-01-02")

	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Check if we need to create a new file (new day or size limit)
	if fw.file == nil || fw.currentDate != currentDate || fw.currentSize >= fw.maxSize {
		if err := fw.rotateFile(currentDate); err != nil {
			return err
		}
	}

	return nil
}

// rotateFile creates a new log file
func (fw *fileWriter) rotateFile(date string) error {
	// Close current file if exists
	if fw.writer != nil {
		_ = fw.writer.Flush()
		fw.writer = nil
	}
	if fw.file != nil {
		_ = fw.file.Close()
		fw.file = nil
	}

	// If it's a new day, reset index
	if fw.currentDate != date {
		fw.currentIndex = 0
		fw.currentDate = date
	} else {
		fw.currentIndex++
	}

	// Create new file
	filename := fmt.Sprintf("app-%s-%03d.log", date, fw.currentIndex)
	logFilePath := filepath.Join(fw.logDir, date, filename)

	// Create directory for the date
	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create date directory: %w", err)
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	fw.file = file
	fw.writer = bufio.NewWriterSize(file, fw.bufferSize)
	fw.currentSize = stat.Size()

	return nil
}

// writeEntry writes a log entry to the current file
func (fw *fileWriter) writeEntry(entry LogEntry) error {
	if err := fw.ensureCurrentFile(); err != nil {
		return err
	}

	// Serialize to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Write to buffer
	n, err := fw.writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	// Add newline
	if _, err := fw.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	fw.currentSize += int64(n + 1)

	return nil
}

// flush forces a flush of the buffer
func (fw *fileWriter) flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.writer != nil {
		return fw.writer.Flush()
	}
	return nil
}

// close closes the file writer
func (fw *fileWriter) close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	var err error
	if fw.writer != nil {
		err = fw.writer.Flush()
		fw.writer = nil
	}
	if fw.file != nil {
		if e := fw.file.Close(); e != nil && err == nil {
			err = e
		}
		fw.file = nil
	}
	return err
}

// processLogs handles batching and writing logs to files
func (l *FileLogger) processLogs() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.config.FlushInterval)
	defer ticker.Stop()

	// Pre-allocate batch slice for better performance
	batch := make([]LogEntry, 0, l.config.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		// Write batch to file
		for _, entry := range batch {
			if err := l.fileWriter.writeEntry(entry); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
			}
		}

		// Flush file buffer
		if err := l.fileWriter.flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush log buffer: %v\n", err)
		}

		// Reset batch without reallocation
		batch = batch[:0]
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

// shouldLog checks if the log level should be processed
func (l *FileLogger) shouldLog(level LogLevel) bool {
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
func (l *FileLogger) createLogEntry(level LogLevel, message string) LogEntry {
	entry := LogEntry{
		ID:          uuid.New().String(),
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Message:     message,
		Logger:      "file-logger",
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
func (l *FileLogger) log(entry LogEntry) {
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
func (l *FileLogger) Debug(message string, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelDebug, message)
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Info logs an info message
func (l *FileLogger) Info(message string, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelInfo, message)
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Warn logs a warning message
func (l *FileLogger) Warn(message string, fields ...map[string]interface{}) {
	entry := l.createLogEntry(LevelWarn, message)
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}
	l.log(entry)
}

// Error logs an error message
func (l *FileLogger) Error(message string, err error, fields ...map[string]interface{}) {
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
func (l *FileLogger) Fatal(message string, err error, fields ...map[string]interface{}) {
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
func (l *FileLogger) WithContext(level LogLevel, message string, ctx LogContext) {
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
func (l *FileLogger) Close() error {
	l.cancel()
	l.wg.Wait()
	close(l.logChannel)
	return l.fileWriter.close()
}

// Flush forces immediate flush of pending logs
func (l *FileLogger) Flush() error {
	// Send empty entry to trigger flush
	select {
	case l.logChannel <- LogEntry{}:
	default:
	}
	time.Sleep(100 * time.Millisecond)
	return l.fileWriter.flush()
}
