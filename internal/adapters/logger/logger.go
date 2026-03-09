// Package logger provides a ports.Logger implementation using the standard library log/slog.
package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/pablogore/devforge/internal/ports"
)

// Logger adapts log/slog to ports.Logger.
type Logger struct {
	inner *slog.Logger
}

// New returns a new Logger with the given level and format.
// Level is one of: debug, info, warn, error (case-insensitive).
// Format is "text" or "json".
func New(level, format string) ports.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info", "":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})
	}

	return &Logger{inner: slog.New(handler)}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.inner.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
}

// With returns a logger with additional key-value fields.
func (l *Logger) With(args ...any) ports.Logger {
	return &Logger{inner: l.inner.With(args...)}
}

// Sync flushes any buffered log entries. No-op for slog; returns nil.
func (l *Logger) Sync() error {
	return nil
}
