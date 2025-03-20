package logger

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	// Save original output
	oldStdout := os.Stdout

	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function that generates output
	f()

	// Close the writer and restore original stdout
	w.Close()
	os.Stdout = oldStdout

	// Read all captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestNewLogger(t *testing.T) {
	// Test creating a new logger
	logger := NewLogger()
	assert.NotNil(t, logger)
	assert.IsType(t, &zerologLogger{}, logger)
}

func TestDebug(t *testing.T) {
	// Configure zerolog to output in a more consistent format for testing
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Test debug logging
	output := captureOutput(func() {
		logger := NewLogger()
		logger.Debug("debug message")
	})

	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, `"level":"debug"`)
}

func TestInfo(t *testing.T) {
	// Configure zerolog to output in a more consistent format for testing
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Test info logging
	output := captureOutput(func() {
		logger := NewLogger()
		logger.Info("info message")
	})

	assert.Contains(t, output, "info message")
	assert.Contains(t, output, `"level":"info"`)
}

func TestWarn(t *testing.T) {
	// Configure zerolog to output in a more consistent format for testing
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	// Test warn logging
	output := captureOutput(func() {
		logger := NewLogger()
		logger.Warn("warn message")
	})

	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, `"level":"warn"`)
}

func TestError(t *testing.T) {
	// Configure zerolog to output in a more consistent format for testing
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// Test error logging
	output := captureOutput(func() {
		logger := NewLogger()
		logger.Error("error message")
	})

	assert.Contains(t, output, "error message")
	assert.Contains(t, output, `"level":"error"`)
}

// We can't easily test Fatal without mocking os.Exit
// and that would require modifying the logger.go file

func TestWithField(t *testing.T) {
	// Configure zerolog to output in a more consistent format for testing
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Test adding fields to logger
	output := captureOutput(func() {
		logger := NewLogger()
		logger = logger.WithField("test_key", "test_value")
		logger.Info("message with field")
	})

	assert.Contains(t, output, "message with field")
	assert.Contains(t, output, `"test_key":"test_value"`)

	// Test multiple fields
	output = captureOutput(func() {
		logger := NewLogger()
		logger = logger.WithField("key1", "value1")
		logger = logger.WithField("key2", "value2")
		logger.Info("message with multiple fields")
	})

	assert.Contains(t, output, "message with multiple fields")
	assert.Contains(t, output, `"key1":"value1"`)
	assert.Contains(t, output, `"key2":"value2"`)

	// Test with different types of values
	output = captureOutput(func() {
		logger := NewLogger()
		logger = logger.WithField("int_field", 123)
		logger = logger.WithField("bool_field", true)
		logger.Info("message with typed fields")
	})

	assert.Contains(t, output, "message with typed fields")
	assert.Contains(t, output, `"int_field":123`)
	assert.Contains(t, output, `"bool_field":true`)
}

func TestWithFieldChaining(t *testing.T) {
	// Configure zerolog to output in a more consistent format for testing
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Test chaining of WithField calls
	output := captureOutput(func() {
		logger := NewLogger().
			WithField("field1", "value1").
			WithField("field2", "value2")
		logger.Info("chained fields")
	})

	assert.Contains(t, output, "chained fields")
	assert.Contains(t, output, `"field1":"value1"`)
	assert.Contains(t, output, `"field2":"value2"`)
}

func TestLogLevelFiltering(t *testing.T) {
	// Test that log levels are properly filtered

	// Set level to Info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Debug should not be logged
	output := captureOutput(func() {
		logger := NewLogger()
		logger.Debug("debug should be filtered")
	})

	assert.NotContains(t, output, "debug should be filtered")

	// Info should be logged
	output = captureOutput(func() {
		logger := NewLogger()
		logger.Info("info should be logged")
	})

	assert.Contains(t, output, "info should be logged")

	// Set level to Error
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// Info should not be logged
	output = captureOutput(func() {
		logger := NewLogger()
		logger.Info("info should be filtered when level is error")
	})

	assert.NotContains(t, output, "info should be filtered when level is error")

	// Error should be logged
	output = captureOutput(func() {
		logger := NewLogger()
		logger.Error("error should be logged")
	})

	assert.Contains(t, output, "error should be logged")
}
