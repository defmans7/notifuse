package logger

import (
	"testing"
)

// TestLogger is a simple logger implementation for testing
type TestLogger struct {
	T *testing.T
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) Logger {
	return &TestLogger{T: t}
}

// Debug logs a debug message
func (l *TestLogger) Debug(msg string) {
	if l.T != nil {
		l.T.Logf("[DEBUG] %s", msg)
	}
}

// Info logs an info message
func (l *TestLogger) Info(msg string) {
	if l.T != nil {
		l.T.Logf("[INFO] %s", msg)
	}
}

// Warn logs a warning message
func (l *TestLogger) Warn(msg string) {
	if l.T != nil {
		l.T.Logf("[WARN] %s", msg)
	}
}

// Error logs an error message
func (l *TestLogger) Error(msg string) {
	if l.T != nil {
		l.T.Logf("[ERROR] %s", msg)
	}
}

// Fatal logs a fatal message
func (l *TestLogger) Fatal(msg string) {
	if l.T != nil {
		l.T.Logf("[FATAL] %s", msg)
	}
}

// WithField returns a logger with a field
func (l *TestLogger) WithField(key string, value interface{}) Logger {
	return l
}

// WithFields returns a logger with fields
func (l *TestLogger) WithFields(fields map[string]interface{}) Logger {
	return l
}

// NewMockLogger creates a simple logger for use in tests
// It can be called with or without a testing.T parameter
func NewMockLogger(t ...*testing.T) Logger {
	if len(t) > 0 {
		return NewTestLogger(t[0])
	}
	return NewTestLogger(nil)
}
