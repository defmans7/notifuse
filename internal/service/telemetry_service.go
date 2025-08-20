package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
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
}

const (
	// TelemetryEndpoint is the hardcoded endpoint for sending telemetry data
	TelemetryEndpoint = "https://telemetry.notifuse.com"
)

// TelemetryServiceConfig contains configuration for the telemetry service
type TelemetryServiceConfig struct {
	Enabled       bool
	APIEndpoint   string
	WorkspaceRepo domain.WorkspaceRepository
	Logger        logger.Logger
	HTTPClient    *http.Client
}

// TelemetryService handles sending telemetry metrics
type TelemetryService struct {
	enabled       bool
	apiEndpoint   string
	workspaceRepo domain.WorkspaceRepository
	logger        logger.Logger
	httpClient    *http.Client
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(config TelemetryServiceConfig) *TelemetryService {
	// Use a default HTTP client with 5 second timeout if none provided
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	return &TelemetryService{
		enabled:       config.Enabled,
		apiEndpoint:   config.APIEndpoint,
		workspaceRepo: config.WorkspaceRepo,
		logger:        config.Logger,
		httpClient:    httpClient,
	}
}

// SendMetricsForAllWorkspaces collects and sends telemetry metrics for all workspaces
func (t *TelemetryService) SendMetricsForAllWorkspaces(ctx context.Context) error {
	if !t.enabled {
		t.logger.Debug("Telemetry is disabled, skipping metrics collection")
		return nil
	}

	// Get all workspaces
	workspaces, err := t.workspaceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	t.logger.WithField("workspace_count", len(workspaces)).Info("Collecting telemetry metrics for workspaces")

	// Collect and send metrics for each workspace
	for _, workspace := range workspaces {
		if err := t.sendMetricsForWorkspace(ctx, workspace.ID); err != nil {
			// Log error but continue with other workspaces
			t.logger.WithField("workspace_id", workspace.ID).
				WithField("error", err).
				Error("Failed to send telemetry metrics for workspace")
		}
	}

	return nil
}

// sendMetricsForWorkspace collects and sends telemetry metrics for a specific workspace
func (t *TelemetryService) sendMetricsForWorkspace(ctx context.Context, workspaceID string) error {
	// Create SHA1 hash of workspace ID
	hasher := sha1.New()
	hasher.Write([]byte(workspaceID))
	workspaceIDSHA1 := hex.EncodeToString(hasher.Sum(nil))

	// Collect metrics
	metrics := TelemetryMetrics{
		WorkspaceIDSHA1: workspaceIDSHA1,
		APIEndpoint:     t.apiEndpoint,
	}

	// Get workspace database connection
	db, err := t.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		t.logger.WithField("workspace_id", workspaceID).
			WithField("error", err).
			Warn("Failed to get workspace database connection for telemetry")
		// Continue without database metrics
	} else {
		// Count contacts
		if contactsCount, err := t.countContacts(ctx, db); err != nil {
			t.logger.WithField("workspace_id", workspaceID).
				WithField("error", err).
				Warn("Failed to count contacts for telemetry")
		} else {
			metrics.ContactsCount = contactsCount
		}

		// Count broadcasts
		if broadcastsCount, err := t.countBroadcasts(ctx, db); err != nil {
			t.logger.WithField("workspace_id", workspaceID).
				WithField("error", err).
				Warn("Failed to count broadcasts for telemetry")
		} else {
			metrics.BroadcastsCount = broadcastsCount
		}

		// Count transactional notifications
		if transactionalCount, err := t.countTransactional(ctx, db); err != nil {
			t.logger.WithField("workspace_id", workspaceID).
				WithField("error", err).
				Warn("Failed to count transactional notifications for telemetry")
		} else {
			metrics.TransactionalCount = transactionalCount
		}

		// Count messages
		if messagesCount, err := t.countMessages(ctx, db); err != nil {
			t.logger.WithField("workspace_id", workspaceID).
				WithField("error", err).
				Warn("Failed to count messages for telemetry")
		} else {
			metrics.MessagesCount = messagesCount
		}

		// Count lists
		if listsCount, err := t.countLists(ctx, db); err != nil {
			t.logger.WithField("workspace_id", workspaceID).
				WithField("error", err).
				Warn("Failed to count lists for telemetry")
		} else {
			metrics.ListsCount = listsCount
		}
	}

	// Send metrics to telemetry endpoint
	return t.sendMetrics(ctx, metrics)
}

// countContacts counts the total number of contacts in a workspace
func (t *TelemetryService) countContacts(ctx context.Context, db *sql.DB) (int, error) {
	query := `SELECT COUNT(DISTINCT email) FROM contacts`
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count contacts: %w", err)
	}
	return count, nil
}

// countBroadcasts counts the total number of broadcasts in a workspace
func (t *TelemetryService) countBroadcasts(ctx context.Context, db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM broadcasts`
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count broadcasts: %w", err)
	}
	return count, nil
}

// countTransactional counts the total number of transactional notifications in a workspace
func (t *TelemetryService) countTransactional(ctx context.Context, db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM transactional_notifications WHERE deleted_at IS NULL`
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactional notifications: %w", err)
	}
	return count, nil
}

// countMessages counts the total number of messages in a workspace
func (t *TelemetryService) countMessages(ctx context.Context, db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM message_history`
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}
	return count, nil
}

// countLists counts the total number of lists in a workspace
func (t *TelemetryService) countLists(ctx context.Context, db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM lists WHERE deleted_at IS NULL`
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count lists: %w", err)
	}
	return count, nil
}

// sendMetrics sends the collected metrics to the telemetry endpoint
func (t *TelemetryService) sendMetrics(ctx context.Context, metrics TelemetryMetrics) error {
	// Marshal metrics to JSON
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry metrics: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", TelemetryEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create telemetry request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Notifuse-Telemetry/1.0")

	// Send request (will fail silently if endpoint is offline due to 5s timeout)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		// Log the error but don't return it to fail silently
		t.logger.WithField("workspace_id_sha1", metrics.WorkspaceIDSHA1).
			WithField("error", err).
			Debug("Failed to send telemetry metrics (endpoint may be offline)")
		return nil // Fail silently as requested
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		t.logger.WithField("workspace_id_sha1", metrics.WorkspaceIDSHA1).
			WithField("status_code", resp.StatusCode).
			Debug("Telemetry endpoint returned error status")
		return nil // Fail silently as requested
	}

	t.logger.WithField("workspace_id_sha1", metrics.WorkspaceIDSHA1).
		WithField("contacts_count", metrics.ContactsCount).
		WithField("broadcasts_count", metrics.BroadcastsCount).
		WithField("transactional_count", metrics.TransactionalCount).
		WithField("messages_count", metrics.MessagesCount).
		WithField("lists_count", metrics.ListsCount).
		Debug("Successfully sent telemetry metrics")

	return nil
}

// StartDailyScheduler starts a goroutine that sends telemetry metrics daily
func (t *TelemetryService) StartDailyScheduler(ctx context.Context) {
	if !t.enabled {
		t.logger.Debug("Telemetry is disabled, not starting daily scheduler")
		return
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		t.logger.Info("Started daily telemetry scheduler")

		for {
			select {
			case <-ctx.Done():
				t.logger.Info("Stopping daily telemetry scheduler")
				return
			case <-ticker.C:
				t.logger.Debug("Daily telemetry tick - collecting metrics")
				if err := t.SendMetricsForAllWorkspaces(ctx); err != nil {
					t.logger.WithField("error", err).Error("Failed to send daily telemetry metrics")
				}
			}
		}
	}()
}
