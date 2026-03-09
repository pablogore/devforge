// Package logger provides a ports.Logger implementation using kit-logger.
package logger

import (
	"github.com/getsyntegrity/kit-logger/pkg/logger"
	"github.com/pablogore/devforge/internal/ports"
)

// Logger adapts kit-logger to ports.Logger.
type Logger struct {
	logger logger.Logger
}

// New returns a new Logger with the given level and format.
func New(level, format string) ports.Logger {
	l := logger.New(logger.Config{
		Level:  level,
		Format: format,
	})
	return &Logger{logger: l}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With returns a logger with additional key-value fields.
func (l *Logger) With(args ...any) ports.Logger {
	return &Logger{logger: l.logger.With(args...)}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}
