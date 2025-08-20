package telemetry

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

// TelemetryMetrics represents the metrics data sent to the telemetry endpoint
type TelemetryMetrics struct {
	WorkspaceIDSHA1    string `json:"workspace_id_sha1"`
	ContactsCount      int    `json:"contacts_count"`
	BroadcastsCount    int    `json:"broadcasts_count"`
	TransactionalCount int    `json:"transactional_count"`
	MessagesCount      int    `json:"messages_count"`
	ListsCount         int    `json:"lists_count"`
	APIEndpoint        string `json:"api_endpoint"`

	// Integration flags - boolean for each email provider
	Mailgun   bool `json:"mailgun"`
	AmazonSES bool `json:"amazonses"`
	Mailjet   bool `json:"mailjet"`
	SendGrid  bool `json:"sendgrid"`
	Postmark  bool `json:"postmark"`
	SMTP      bool `json:"smtp"`
	S3        bool `json:"s3"`
}

// LogEntry represents the structured log entry for Google Cloud Logging
type LogEntry struct {
	Timestamp          time.Time `json:"timestamp"`
	WorkspaceIDSHA1    string    `json:"workspace_id_sha1"`
	ContactsCount      int       `json:"contacts_count"`
	BroadcastsCount    int       `json:"broadcasts_count"`
	TransactionalCount int       `json:"transactional_count"`
	MessagesCount      int       `json:"messages_count"`
	ListsCount         int       `json:"lists_count"`
	APIEndpoint        string    `json:"api_endpoint"`
	Source             string    `json:"source"`
	EventType          string    `json:"event_type"`

	// Integration flags - boolean for each email provider
	Mailgun   bool `json:"mailgun"`
	AmazonSES bool `json:"amazonses"`
	Mailjet   bool `json:"mailjet"`
	SendGrid  bool `json:"sendgrid"`
	Postmark  bool `json:"postmark"`
	SMTP      bool `json:"smtp"`
	S3        bool `json:"s3"`
}

var (
	loggingClient *logging.Client
	logger        *logging.Logger
)

func init() {
	// Register the HTTP function
	functions.HTTP("ReceiveTelemetry", receiveTelemetry)

	// Initialize Google Cloud Logging client
	ctx := context.Background()
	var err error
	loggingClient, err = logging.NewClient(ctx, "")
	if err != nil {
		log.Fatalf("Failed to create logging client: %v", err)
	}

	// Create a logger with the name "telemetry"
	logger = loggingClient.Logger("telemetry")
}

// receiveTelemetry is the main HTTP function handler
func receiveTelemetry(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow requests from any origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, User-Agent")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check if user agent contains "Notifuse-Telemetry"
	userAgent := r.Header.Get("User-Agent")
	if !strings.Contains(userAgent, "Notifuse-Telemetry") {
		// Fail silently - return success but don't process the request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":    "success",
			"message":   "Request received",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse JSON payload
	var metrics TelemetryMetrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		log.Printf("Failed to decode JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Create structured log entry
	logEntry := LogEntry{
		Timestamp:          time.Now(),
		WorkspaceIDSHA1:    metrics.WorkspaceIDSHA1,
		ContactsCount:      metrics.ContactsCount,
		BroadcastsCount:    metrics.BroadcastsCount,
		TransactionalCount: metrics.TransactionalCount,
		MessagesCount:      metrics.MessagesCount,
		ListsCount:         metrics.ListsCount,
		APIEndpoint:        metrics.APIEndpoint,
		Source:             "notifuse-platform",
		EventType:          "telemetry_metrics",
		Mailgun:            metrics.Mailgun,
		AmazonSES:          metrics.AmazonSES,
		Mailjet:            metrics.Mailjet,
		SendGrid:           metrics.SendGrid,
		Postmark:           metrics.Postmark,
		SMTP:               metrics.SMTP,
		S3:                 metrics.S3,
	}

	// Log to Google Cloud Logging with structured data
	logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  logEntry,
		Labels: map[string]string{
			"workspace_id_sha1": metrics.WorkspaceIDSHA1,
			"event_type":        "telemetry_metrics",
			"source":            "notifuse-platform",
		},
	})

	// Count active integrations for logging
	integrationCount := 0
	if metrics.Mailgun {
		integrationCount++
	}
	if metrics.AmazonSES {
		integrationCount++
	}
	if metrics.Mailjet {
		integrationCount++
	}
	if metrics.SendGrid {
		integrationCount++
	}
	if metrics.Postmark {
		integrationCount++
	}
	if metrics.SMTP {
		integrationCount++
	}
	if metrics.S3 {
		integrationCount++
	}

	// Log basic info to stdout for Cloud Functions logs
	log.Printf("Received telemetry for workspace %s: contacts=%d, broadcasts=%d, transactional=%d, messages=%d, lists=%d, integrations=%d",
		metrics.WorkspaceIDSHA1,
		metrics.ContactsCount,
		metrics.BroadcastsCount,
		metrics.TransactionalCount,
		metrics.MessagesCount,
		metrics.ListsCount,
		integrationCount,
	)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Telemetry data received and logged",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// Cleanup function to close the logging client (called automatically by Cloud Functions runtime)
func cleanup() {
	if loggingClient != nil {
		loggingClient.Close()
	}
}
