// Package logging provides structured JSON logging for loko.
// All logs go to stderr to avoid interfering with stdout/MCP protocol.
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Ensure Logger implements usecases.Logger interface.
var _ usecases.Logger = (*Logger)(nil)

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
	level  Level
	fields map[string]any
	ctx    context.Context
}

// New creates a new logger with the given level.
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		fields: make(map[string]any),
		ctx:    context.Background(),
	}
}

// WithContext returns a logger that includes the given context.
// This can be used for request/operation tracking via context values.
func (l *Logger) WithContext(ctx context.Context) usecases.Logger {
	newLogger := &Logger{
		level:  l.level,
		fields: copyFields(l.fields),
		ctx:    ctx,
	}
	return newLogger
}

// WithFields returns a logger with additional structured fields.
func (l *Logger) WithFields(keysAndValues ...any) usecases.Logger {
	newLogger := &Logger{
		level:  l.level,
		fields: copyFields(l.fields),
		ctx:    l.ctx,
	}
	mergeKeysAndValues(newLogger.fields, keysAndValues)
	return newLogger
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	if l.level != LevelDebug {
		return
	}
	l.log(LevelDebug, msg, keysAndValues)
}

// Info logs an info message.
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.log(LevelInfo, msg, keysAndValues)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.log(LevelWarn, msg, keysAndValues)
}

// Error logs an error message.
func (l *Logger) Error(msg string, err error, keysAndValues ...any) {
	fields := parseKeysAndValues(keysAndValues)
	if err != nil {
		fields["error"] = err.Error()
	}
	l.logWithFields(LevelError, msg, fields)
}

// log writes a structured JSON log entry to stderr.
func (l *Logger) log(level Level, message string, keysAndValues []any) {
	fields := parseKeysAndValues(keysAndValues)
	l.logWithFields(level, message, fields)
}

// logWithFields writes a structured JSON log entry to stderr with pre-parsed fields.
func (l *Logger) logWithFields(level Level, message string, fields map[string]any) {
	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}

	// Merge logger's persistent fields
	maps.Copy(entry, l.fields)

	// Merge call-specific fields
	maps.Copy(entry, fields)

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error":"failed to marshal log entry: %v"}`+"\n", err)
		return
	}

	// Write to stderr (don't interfere with MCP stdio)
	fmt.Fprintf(os.Stderr, "%s\n", string(data))
}

// parseKeysAndValues converts variadic key-value pairs into a map.
// Keys must be strings; non-string keys are skipped with a warning.
func parseKeysAndValues(keysAndValues []any) map[string]any {
	fields := make(map[string]any)
	mergeKeysAndValues(fields, keysAndValues)
	return fields
}

// mergeKeysAndValues merges variadic key-value pairs into an existing map.
func mergeKeysAndValues(fields map[string]any, keysAndValues []any) {
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			// Skip non-string keys but continue processing
			continue
		}
		fields[key] = keysAndValues[i+1]
	}
}

// copyFields creates a shallow copy of the fields map.
func copyFields(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	maps.Copy(dst, src)
	return dst
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
