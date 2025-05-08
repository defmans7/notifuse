package tracing

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/config"
)

func TestInitTracing_Disabled(t *testing.T) {
	// Setup a tracing config with tracing disabled
	cfg := &config.TracingConfig{
		Enabled: false,
	}

	// Verify that initialization succeeds but does not set up any tracing
	err := InitTracing(cfg)
	if err != nil {
		t.Fatalf("Expected no error when tracing is disabled, got: %v", err)
	}
}

func TestInitTracing_WithInvalidExporter(t *testing.T) {
	// Setup a tracing config with an invalid exporter
	cfg := &config.TracingConfig{
		Enabled:       true,
		TraceExporter: "invalid",
	}

	// Expect an error
	err := InitTracing(cfg)
	if err == nil {
		t.Error("Expected error with invalid exporter, got nil")
	}
}

func TestInitMetricsExporters_WithInvalidExporter(t *testing.T) {
	// Setup a tracing config with an invalid metrics exporter
	cfg := &config.TracingConfig{
		Enabled:         true,
		MetricsExporter: "invalid",
	}

	// Expect an error
	err := initMetricsExporters(cfg)
	if err == nil {
		t.Error("Expected error with invalid metrics exporter, got nil")
	}
}

func TestInitMetricsExporters_Disabled(t *testing.T) {
	// Setup a tracing config with metrics disabled
	cfg := &config.TracingConfig{
		Enabled:         true,
		MetricsExporter: "none",
	}

	// Verify that initialization succeeds but does not set up any metrics
	err := initMetricsExporters(cfg)
	if err != nil {
		t.Fatalf("Expected no error when metrics are disabled, got: %v", err)
	}
}

func TestInitMetricsExporters_WithMultipleExportersSplitting(t *testing.T) {
	// This test simply checks if we correctly parse multiple exporter names
	// We don't actually initialize the exporters because that would require
	// external dependencies
	exporterStr := "prometheus, stackdriver,  datadog,, "
	exporters := strings.Split(exporterStr, ",")

	// Check that we get the expected number of non-empty exporters after trimming
	count := 0
	for _, exp := range exporters {
		if strings.TrimSpace(exp) != "" {
			count++
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 non-empty exporters, got %d", count)
	}

	// Now verify each one
	foundPrometheus := false
	foundStackdriver := false
	foundDatadog := false

	for _, exp := range exporters {
		exp = strings.TrimSpace(exp)
		switch exp {
		case "prometheus":
			foundPrometheus = true
		case "stackdriver":
			foundStackdriver = true
		case "datadog":
			foundDatadog = true
		case "":
			// Skip empty strings
		default:
			t.Errorf("Unexpected exporter name: %s", exp)
		}
	}

	if !foundPrometheus {
		t.Error("Expected to find 'prometheus' exporter")
	}
	if !foundStackdriver {
		t.Error("Expected to find 'stackdriver' exporter")
	}
	if !foundDatadog {
		t.Error("Expected to find 'datadog' exporter")
	}
}

func TestGetHTTPOptions(t *testing.T) {
	transport := GetHTTPOptions()

	// Create a test request to check span naming
	req := httptest.NewRequest("GET", "/test-path", nil)
	spanName := transport.FormatSpanName(req)

	expectedSpanName := "GET /test-path"
	if spanName != expectedSpanName {
		t.Errorf("Expected span name to be %s, got %s", expectedSpanName, spanName)
	}

	// Verify sampler is not nil
	if transport.StartOptions.Sampler == nil {
		t.Fatal("Expected StartOptions.Sampler to be set")
	}
}

func TestRegisterHTTPServerViews(t *testing.T) {
	// This test just verifies the function does not error
	err := RegisterHTTPServerViews()
	if err != nil {
		t.Fatalf("Expected no error when registering HTTP server views, got: %v", err)
	}
}

func TestRegisterCustomViews(t *testing.T) {
	// This test verifies the custom views registration does not error
	err := registerCustomViews()
	if err != nil {
		t.Fatalf("Expected no error when registering custom views, got: %v", err)
	}
}
