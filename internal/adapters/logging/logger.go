// Package logging provides structured JSON logging for loko.
// All logs go to stderr to avoid interfering with stdout/MCP protocol.
package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Level represents a log level.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger provides structured JSON logging.
type Logger struct {
	level   Level
	context map[string]interface{}
}

// New creates a new logger with the given level.
func New(level Level) *Logger {
	return &Logger{
		level:   level,
		context: make(map[string]interface{}),
	}
}

// WithContext adds context fields to the logger.
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	l.context[key] = value
	return l
}

// Debug logs a debug message.
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	if l.level != LevelDebug {
		return
	}
	l.log(LevelDebug, message, fields)
}

// Info logs an info message.
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(LevelInfo, message, fields)
}

// Warn logs a warning message.
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log(LevelWarn, message, fields)
}

// Error logs an error message.
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(LevelError, message, fields)
}

// log writes a structured JSON log entry to stderr.
func (l *Logger) log(level Level, message string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}

	// Merge context
	for k, v := range l.context {
		entry[k] = v
	}

	// Merge fields
	if fields != nil {
		for k, v := range fields {
			entry[k] = v
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error":"failed to marshal log entry: %v"}`, err)
		return
	}

	// Write to stderr (don't interfere with MCP stdio)
	fmt.Fprintf(os.Stderr, "%s\n", string(data))
}

// Global logger instance
var global = New(LevelInfo)

// SetLevel sets the global log level.
func SetLevel(level Level) {
	global.level = level
}

// GetLogger returns the global logger.
func GetLogger() *Logger {
	return global
}
