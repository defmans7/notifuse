package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

type webhookEventRepository struct {
	workspaceRepo domain.WorkspaceRepository
}

// NewWebhookEventRepository creates a new PostgreSQL repository for webhook events
func NewWebhookEventRepository(workspaceRepo domain.WorkspaceRepository) domain.WebhookEventRepository {
	return &webhookEventRepository{
		workspaceRepo: workspaceRepo,
	}
}

// webhookEventModel is the database model for webhook events
type webhookEventModel struct {
	ID                    string    `db:"id"`
	Type                  string    `db:"type"`
	EmailProviderKind     string    `db:"email_provider_kind"`
	IntegrationID         string    `db:"integration_id"`
	RecipientEmail        string    `db:"recipient_email"`
	MessageID             string    `db:"message_id"`
	TransactionalID       string    `db:"transactional_id"`
	BroadcastID           string    `db:"broadcast_id"`
	Timestamp             time.Time `db:"timestamp"`
	RawPayload            string    `db:"raw_payload"`
	BounceType            string    `db:"bounce_type"`
	BounceCategory        string    `db:"bounce_category"`
	BounceDiagnostic      string    `db:"bounce_diagnostic"`
	ComplaintFeedbackType string    `db:"complaint_feedback_type"`
	CreatedAt             time.Time `db:"created_at"`
}

// toDomain converts a database model to a domain model
func (m *webhookEventModel) toDomain() *domain.WebhookEvent {
	event := domain.NewWebhookEvent(
		m.ID,
		domain.EmailEventType(m.Type),
		domain.EmailProviderKind(m.EmailProviderKind),
		m.IntegrationID,
		m.RecipientEmail,
		m.MessageID,
		m.Timestamp,
		m.RawPayload,
	)

	if m.TransactionalID != "" {
		event.SetTransactionalID(m.TransactionalID)
	}

	if m.BroadcastID != "" {
		event.SetBroadcastID(m.BroadcastID)
	}

	if m.Type == string(domain.EmailEventBounce) && (m.BounceType != "" || m.BounceCategory != "" || m.BounceDiagnostic != "") {
		event.SetBounceInfo(m.BounceType, m.BounceCategory, m.BounceDiagnostic)
	}

	if m.Type == string(domain.EmailEventComplaint) && m.ComplaintFeedbackType != "" {
		event.SetComplaintInfo(m.ComplaintFeedbackType)
	}

	return event
}

// scanWebhookEventModel scans a database row into a webhookEventModel
func scanWebhookEventModel(scanner interface {
	Scan(dest ...interface{}) error
}) (*webhookEventModel, error) {
	var model webhookEventModel
	err := scanner.Scan(
		&model.ID,
		&model.Type,
		&model.EmailProviderKind,
		&model.IntegrationID,
		&model.RecipientEmail,
		&model.MessageID,
		&model.TransactionalID,
		&model.BroadcastID,
		&model.Timestamp,
		&model.RawPayload,
		&model.BounceType,
		&model.BounceCategory,
		&model.BounceDiagnostic,
		&model.ComplaintFeedbackType,
		&model.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// toModel converts a domain model to a database model
func eventToModel(e *domain.WebhookEvent) *webhookEventModel {
	return &webhookEventModel{
		ID:                    e.ID,
		Type:                  string(e.Type),
		EmailProviderKind:     string(e.EmailProviderKind),
		IntegrationID:         e.IntegrationID,
		RecipientEmail:        e.RecipientEmail,
		MessageID:             e.MessageID,
		TransactionalID:       e.TransactionalID,
		BroadcastID:           e.BroadcastID,
		Timestamp:             e.Timestamp,
		RawPayload:            e.RawPayload,
		BounceType:            e.BounceType,
		BounceCategory:        e.BounceCategory,
		BounceDiagnostic:      e.BounceDiagnostic,
		ComplaintFeedbackType: e.ComplaintFeedbackType,
		CreatedAt:             time.Now(),
	}
}

// StoreEvent stores a webhook event in the database
func (r *webhookEventRepository) StoreEvent(ctx context.Context, event *domain.WebhookEvent) error {
	// Since webhook events need to be stored in a central place, we'll use "system" as the workspace ID
	// This corresponds to the system database that contains webhook events for all workspaces
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return fmt.Errorf("failed to get system database connection: %w", err)
	}

	model := eventToModel(event)

	query := `
		INSERT INTO webhook_events (
			id, type, email_provider_kind, integration_id, recipient_email, message_id, 
			transactional_id, broadcast_id, timestamp, raw_payload, 
			bounce_type, bounce_category, bounce_diagnostic, 
			complaint_feedback_type, created_at
		) 
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err = systemDB.ExecContext(ctx, query,
		model.ID, model.Type, model.EmailProviderKind, model.IntegrationID, model.RecipientEmail, model.MessageID,
		model.TransactionalID, model.BroadcastID, model.Timestamp, model.RawPayload,
		model.BounceType, model.BounceCategory, model.BounceDiagnostic,
		model.ComplaintFeedbackType, model.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to store webhook event: %w", err)
	}

	return nil
}

// GetEventByID retrieves a webhook event by its ID
func (r *webhookEventRepository) GetEventByID(ctx context.Context, id string) (*domain.WebhookEvent, error) {
	// Use the system database for webhook events
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database connection: %w", err)
	}

	query := `
		SELECT * FROM webhook_events WHERE id = $1
	`

	row := systemDB.QueryRowContext(ctx, query, id)
	model, err := scanWebhookEventModel(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.ErrWebhookEventNotFound{ID: id}
		}
		return nil, fmt.Errorf("failed to get webhook event: %w", err)
	}

	return model.toDomain(), nil
}

// GetEventsByMessageID retrieves all webhook events associated with a message ID
func (r *webhookEventRepository) GetEventsByMessageID(ctx context.Context, messageID string, limit, offset int) ([]*domain.WebhookEvent, error) {
	// Use the system database for webhook events
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database connection: %w", err)
	}

	query := `
		SELECT * FROM webhook_events 
		WHERE message_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := systemDB.QueryContext(ctx, query, messageID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events by message ID: %w", err)
	}
	defer rows.Close()

	var events []*domain.WebhookEvent

	for rows.Next() {
		model, err := scanWebhookEventModel(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event row: %w", err)
		}

		events = append(events, model.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating through webhook events: %w", err)
	}

	return events, nil
}

// GetEventsByTransactionalID retrieves all webhook events associated with a transactional ID
func (r *webhookEventRepository) GetEventsByTransactionalID(ctx context.Context, transactionalID string, limit, offset int) ([]*domain.WebhookEvent, error) {
	// Use the system database for webhook events
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database connection: %w", err)
	}

	query := `
		SELECT * FROM webhook_events 
		WHERE transactional_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := systemDB.QueryContext(ctx, query, transactionalID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events by transactional ID: %w", err)
	}
	defer rows.Close()

	var events []*domain.WebhookEvent

	for rows.Next() {
		model, err := scanWebhookEventModel(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event row: %w", err)
		}

		events = append(events, model.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating through webhook events: %w", err)
	}

	return events, nil
}

// GetEventsByBroadcastID retrieves all webhook events associated with a broadcast ID
func (r *webhookEventRepository) GetEventsByBroadcastID(ctx context.Context, broadcastID string, limit, offset int) ([]*domain.WebhookEvent, error) {
	// Use the system database for webhook events
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database connection: %w", err)
	}

	query := `
		SELECT * FROM webhook_events 
		WHERE broadcast_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := systemDB.QueryContext(ctx, query, broadcastID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events by broadcast ID: %w", err)
	}
	defer rows.Close()

	var events []*domain.WebhookEvent

	for rows.Next() {
		model, err := scanWebhookEventModel(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event row: %w", err)
		}

		events = append(events, model.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating through webhook events: %w", err)
	}

	return events, nil
}

// GetEventsByType retrieves webhook events by type (delivered, bounce, complaint)
func (r *webhookEventRepository) GetEventsByType(ctx context.Context, workspaceID string, eventType domain.EmailEventType, limit, offset int) ([]*domain.WebhookEvent, error) {
	// Use the system database for webhook events
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database connection: %w", err)
	}

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First, get transactional notification IDs from this workspace
	transactionalQuery := `
		SELECT id FROM transactional_notifications 
		WHERE workspace_id = $1
	`

	transRows, err := workspaceDB.QueryContext(ctx, transactionalQuery, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactional notification IDs: %w", err)
	}
	defer transRows.Close()

	var transactionalIDs []string
	for transRows.Next() {
		var id string
		if err := transRows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan transactional ID: %w", err)
		}
		transactionalIDs = append(transactionalIDs, id)
	}

	// Next, get broadcast IDs from this workspace
	broadcastQuery := `
		SELECT id FROM broadcasts 
		WHERE workspace_id = $1
	`

	broadcastRows, err := workspaceDB.QueryContext(ctx, broadcastQuery, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast IDs: %w", err)
	}
	defer broadcastRows.Close()

	var broadcastIDs []string
	for broadcastRows.Next() {
		var id string
		if err := broadcastRows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan broadcast ID: %w", err)
		}
		broadcastIDs = append(broadcastIDs, id)
	}

	// Now, construct the query for webhook events
	var params []interface{}
	var conditions []string
	params = append(params, string(eventType))
	conditions = append(conditions, fmt.Sprintf("type = $%d", len(params)))

	// Add transactional IDs condition if any exist
	if len(transactionalIDs) > 0 {
		placeholders := make([]string, len(transactionalIDs))
		for i, id := range transactionalIDs {
			params = append(params, id)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("transactional_id IN (%s)", joinStrings(placeholders, ",")))
	}

	// Add broadcast IDs condition if any exist
	if len(broadcastIDs) > 0 {
		placeholders := make([]string, len(broadcastIDs))
		for i, id := range broadcastIDs {
			params = append(params, id)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("broadcast_id IN (%s)", joinStrings(placeholders, ",")))
	}

	// If no IDs were found, return empty result
	if len(transactionalIDs) == 0 && len(broadcastIDs) == 0 {
		return []*domain.WebhookEvent{}, nil
	}

	// Create the WHERE clause
	whereClause := joinStrings(conditions, " OR ")

	// Add LIMIT and OFFSET
	params = append(params, limit, offset)
	query := fmt.Sprintf(`
		SELECT * FROM webhook_events 
		WHERE %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(params)-1, len(params))

	// Execute the query
	rows, err := systemDB.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events by type: %w", err)
	}
	defer rows.Close()

	var events []*domain.WebhookEvent

	for rows.Next() {
		model, err := scanWebhookEventModel(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event row: %w", err)
		}

		events = append(events, model.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating through webhook events: %w", err)
	}

	return events, nil
}

// GetEventCount retrieves the count of events by type for a workspace
func (r *webhookEventRepository) GetEventCount(ctx context.Context, workspaceID string, eventType domain.EmailEventType) (int, error) {
	// Use the system database for webhook events
	systemDB, err := r.workspaceRepo.GetConnection(ctx, "system")
	if err != nil {
		return 0, fmt.Errorf("failed to get system database connection: %w", err)
	}

	// Get the workspace database connection
	workspaceDB, err := r.workspaceRepo.GetConnection(ctx, workspaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get workspace connection: %w", err)
	}

	// First, get transactional notification IDs from this workspace
	transactionalQuery := `
		SELECT id FROM transactional_notifications 
		WHERE workspace_id = $1
	`

	transRows, err := workspaceDB.QueryContext(ctx, transactionalQuery, workspaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get transactional notification IDs: %w", err)
	}
	defer transRows.Close()

	var transactionalIDs []string
	for transRows.Next() {
		var id string
		if err := transRows.Scan(&id); err != nil {
			return 0, fmt.Errorf("failed to scan transactional ID: %w", err)
		}
		transactionalIDs = append(transactionalIDs, id)
	}

	// Next, get broadcast IDs from this workspace
	broadcastQuery := `
		SELECT id FROM broadcasts 
		WHERE workspace_id = $1
	`

	broadcastRows, err := workspaceDB.QueryContext(ctx, broadcastQuery, workspaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get broadcast IDs: %w", err)
	}
	defer broadcastRows.Close()

	var broadcastIDs []string
	for broadcastRows.Next() {
		var id string
		if err := broadcastRows.Scan(&id); err != nil {
			return 0, fmt.Errorf("failed to scan broadcast ID: %w", err)
		}
		broadcastIDs = append(broadcastIDs, id)
	}

	// Now, construct the query for webhook events
	var params []interface{}
	var conditions []string
	params = append(params, string(eventType))
	conditions = append(conditions, fmt.Sprintf("type = $%d", len(params)))

	// Add transactional IDs condition if any exist
	if len(transactionalIDs) > 0 {
		placeholders := make([]string, len(transactionalIDs))
		for i, id := range transactionalIDs {
			params = append(params, id)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("transactional_id IN (%s)", joinStrings(placeholders, ",")))
	}

	// Add broadcast IDs condition if any exist
	if len(broadcastIDs) > 0 {
		placeholders := make([]string, len(broadcastIDs))
		for i, id := range broadcastIDs {
			params = append(params, id)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("broadcast_id IN (%s)", joinStrings(placeholders, ",")))
	}

	// If no IDs were found, return zero
	if len(transactionalIDs) == 0 && len(broadcastIDs) == 0 {
		return 0, nil
	}

	// Create the WHERE clause
	whereClause := joinStrings(conditions, " OR ")

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM webhook_events 
		WHERE %s
	`, whereClause)

	// Execute the query
	var count int
	err = systemDB.QueryRowContext(ctx, query, params...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get webhook event count: %w", err)
	}

	return count, nil
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}

	return result
}
