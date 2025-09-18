// Package logger provides a simple asynchronous logging solution for Go applications
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
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

// LogEntry represents a simple log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Caller    string
	Message   string
}

// Config holds the logger configuration
type Config struct {
	LogDir        string        // Directory for log files
	LogLevel      LogLevel      // Minimum log level to process
	FlushInterval time.Duration // How often to flush logs to file
	BufferSize    int           // Channel buffer size
	EnableCaller  bool          // Whether to capture caller information
}

// Logger is the main logger instance
type Logger struct {
	config      Config
	logChannel  chan LogEntry
	wg          sync.WaitGroup
	done        chan struct{}
	file        *os.File
	mu          sync.Mutex
	currentDate string
}

// NewLogger creates a new Logger instance
func NewLogger(config Config) *Logger {
	// Set defaults
	if config.LogDir == "" {
		config.LogDir = "./logs"
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 1 * time.Second
	}
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.LogLevel == "" {
		config.LogLevel = LevelInfo
	}

	// Create log directory
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
	}

	logger := &Logger{
		config:     config,
		logChannel: make(chan LogEntry, config.BufferSize),
		done:       make(chan struct{}),
	}

	// Initialize log file
	if err := logger.initLogFile(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize log file: %v\n", err)
	}

	// Start background goroutine for processing logs
	logger.wg.Add(1)
	go logger.processLogs()

	return logger
}

// initLogFile creates or opens the log file for today
func (l *Logger) initLogFile() error {
	// Close existing file if open
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}

	// Create filename based on current date
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("app-%s.log", today)
	logPath := filepath.Join(l.config.LogDir, filename)

	// Open file for appending (create if doesn't exist)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	l.file = file
	l.currentDate = today
	return nil
}

// processLogs handles writing logs to file asynchronously
func (l *Logger) processLogs() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.config.FlushInterval)
	defer ticker.Stop()

	batch := make([]LogEntry, 0, 100)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		l.mu.Lock()
		defer l.mu.Unlock()

		// Check if we need to rotate file (new day)
		today := time.Now().Format("2006-01-02")
		if l.currentDate != today {
			if err := l.initLogFile(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
				return
			}
		}

		// Ensure we have a valid file
		if l.file == nil {
			if err := l.initLogFile(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize log file: %v\n", err)
				return
			}
		}

		// Write all entries in batch
		for _, entry := range batch {
			line := l.formatLogEntry(entry)
			if _, err := l.file.WriteString(line + "\n"); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
			}
		}

		// Sync to disk
		if err := l.file.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync log file: %v\n", err)
		}

		// Clear batch
		batch = batch[:0]
	}

	for {
		select {
		case entry := <-l.logChannel:
			batch = append(batch, entry)

			// Flush if batch is full or if it's a fatal error
			if len(batch) >= 100 || entry.Level == LevelFatal {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-l.done:
			// Final flush before shutting down
			flush()
			return
		}
	}
}

// formatLogEntry formats a log entry into a simple readable string
func (l *Logger) formatLogEntry(entry LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	if l.config.EnableCaller && entry.Caller != "" {
		return fmt.Sprintf("[%s] %s %s: %s",
			timestamp, entry.Level, entry.Caller, entry.Message)
	}

	return fmt.Sprintf("[%s] %s: %s",
		timestamp, entry.Level, entry.Message)
}

// shouldLog checks if the log level should be processed
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}

	return levels[level] >= levels[l.config.LogLevel]
}

// getCaller returns the caller function name and line
func (l *Logger) getCaller() string {
	if !l.config.EnableCaller {
		return ""
	}

	// Skip 3 frames: getCaller -> log method -> actual caller
	if pc, file, line, ok := runtime.Caller(3); ok {
		filename := filepath.Base(file)
		funcName := "unknown"
		if fn := runtime.FuncForPC(pc); fn != nil {
			funcName = filepath.Base(fn.Name())
		}
		return fmt.Sprintf("%s:%d:%s", filename, line, funcName)
	}
	return ""
}

// log sends a log entry to the processing channel
func (l *Logger) log(level LogLevel, message string) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Caller:    l.getCaller(),
		Message:   message,
	}

	select {
	case l.logChannel <- entry:
		// Successfully queued
	default:
		// Channel is full, log to stderr as fallback
		fmt.Fprintf(os.Stderr, "Logger channel full, dropping log: %s\n", message)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(LevelDebug, message)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log(LevelInfo, message)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(LevelWarn, message)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(LevelError, message)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message
func (l *Logger) Fatal(message string) {
	l.log(LevelFatal, message)
}

// Fatalf logs a formatted fatal message
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(LevelFatal, fmt.Sprintf(format, args...))
}

// Close gracefully shuts down the logger
func (l *Logger) Close() error {
	// Signal shutdown
	close(l.done)

	// Wait for goroutine to finish
	l.wg.Wait()

	// Close channel
	close(l.logChannel)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// Flush forces immediate flush of pending logs
func (l *Logger) Flush() {
	// Create a dummy entry to trigger immediate flush
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LevelDebug,
		Message:   "__FLUSH__",
	}

	select {
	case l.logChannel <- entry:
		// Wait a bit to ensure it gets processed
		time.Sleep(200 * time.Millisecond)
	default:
		// Channel full, just wait
		time.Sleep(200 * time.Millisecond)
	}
}
