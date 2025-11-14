package milady

import (
	"github.com/miladystack/miladystack/pkg/log"
	"github.com/miladystack/miladystack/pkg/logger"
)

// miladyLogger provides an implementation of the logger.Logger interface.
type miladyLogger struct{}

// Ensure that miladyLogger implements the logger.Logger interface.
var _ logger.Logger = (*miladyLogger)(nil)

// NewLogger creates a new instance of miladyLogger.
func NewLogger() *miladyLogger {
	return &miladyLogger{}
}

// Debug logs a debug message with any additional key-value pairs.
func (l *miladyLogger) Debug(msg string, kvs ...any) {
	log.Debugw(msg, kvs...)
}

// Warn logs a warning message with any additional key-value pairs.
func (l *miladyLogger) Warn(msg string, kvs ...any) {
	log.Warnw(msg, kvs...)
}

// Info logs an informational message with any additional key-value pairs.
func (l *miladyLogger) Info(msg string, kvs ...any) {
	log.Infow(msg, kvs...)
}

// Error logs an error message with any additional key-value pairs.
func (l *miladyLogger) Error(msg string, kvs ...any) {
	log.Errorw(nil, msg, kvs...)
}
