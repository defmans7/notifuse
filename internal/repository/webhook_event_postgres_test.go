package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookEventRepository_StoreEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	workspaceID := "ws-123"
	now := time.Now().UTC()

	event := &domain.WebhookEvent{
		ID:                "evt-123",
		Type:              domain.EmailEventDelivered,
		EmailProviderKind: domain.EmailProviderKindSES,
		IntegrationID:     "integration-123",
		RecipientEmail:    "test@example.com",
		MessageID:         "msg-123",
		TransactionalID:   "trans-123",
		BroadcastID:       "broadcast-123",
		Timestamp:         now,
		RawPayload:        `{"key": "value"}`,
		BounceType:        "hard",
		BounceCategory:    "unknown",
		BounceDiagnostic:  "550 user unknown",
		CreatedAt:         now,
	}

	// Set up the workspace connection expectation
	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(db, nil)

	// Expect the SQL query with parameters - use sqlmock.AnyArg() for the created_at timestamp
	mock.ExpectExec(`INSERT INTO webhook_events`).
		WithArgs(
			event.ID,
			event.Type,
			event.EmailProviderKind,
			event.IntegrationID,
			event.RecipientEmail,
			event.MessageID,
			event.TransactionalID,
			event.BroadcastID,
			event.Timestamp,
			event.RawPayload,
			event.BounceType,
			event.BounceCategory,
			event.BounceDiagnostic,
			event.ComplaintFeedbackType,
			sqlmock.AnyArg(), // created_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method
	err = repo.StoreEvent(ctx, workspaceID, event)
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookEventRepository_StoreEvent_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewWebhookEventRepository(mockWorkspaceRepo)

	ctx := context.Background()
	workspaceID := "ws-123"

	// Test case 1: Database connection error
	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(nil, errors.New("connection error"))

	err := repo.StoreEvent(ctx, workspaceID, &domain.WebhookEvent{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get workspace connection")

	// Test case 2: SQL execution error
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(db, nil)

	mock.ExpectExec(`INSERT INTO webhook_events`).
		WillReturnError(errors.New("database error"))

	err = repo.StoreEvent(ctx, workspaceID, &domain.WebhookEvent{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store webhook event")
}

func TestWebhookEventRepository_ListEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	workspaceID := "ws-123"
	now := time.Now().UTC()

	// Set up the workspace connection expectation
	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(db, nil)

	// Test with various filter parameters
	params := domain.WebhookEventListParams{
		Limit:          10,
		EventType:      domain.EmailEventBounce,
		RecipientEmail: "test@example.com",
		MessageID:      "msg-123",
	}

	// Set up rows for the SQL query result
	rows := sqlmock.NewRows([]string{
		"id", "type", "email_provider_kind", "integration_id", "recipient_email",
		"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
		"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type",
		"created_at",
	}).
		AddRow(
			"evt-1", domain.EmailEventBounce, domain.EmailProviderKindSES, "integration-1", "test@example.com",
			"msg-1", "trans-1", "broadcast-1", now, `{"key": "value1"}`,
			"hard", "unknown", "550 user unknown", "", now,
		).
		AddRow(
			"evt-2", domain.EmailEventBounce, domain.EmailProviderKindSES, "integration-2", "test@example.com",
			"msg-2", "trans-2", "broadcast-2", now, `{"key": "value2"}`,
			"soft", "mailbox_full", "452 mailbox full", "", now,
		)

	// Expect a SQL query with filters
	mock.ExpectQuery(`SELECT .+ FROM webhook_events WHERE`).
		WillReturnRows(rows)

	// Call the method
	result, err := repo.ListEvents(ctx, workspaceID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Events, 2)
	assert.Equal(t, "evt-1", result.Events[0].ID)
	assert.Equal(t, "evt-2", result.Events[1].ID)
	assert.Equal(t, "hard", result.Events[0].BounceType)
	assert.Equal(t, "soft", result.Events[1].BounceType)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookEventRepository_ListEvents_WithCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	workspaceID := "ws-123"
	now := time.Now().UTC()

	// Create a valid cursor (base64 encoded "timestamp~id")
	cursor := "MjAyMy0xMS0xMFQxMjozNDo1NiswMDowMH5ldnQtcHJldmlvdXM=" // Example base64 encoded cursor

	// Set up the workspace connection expectation
	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(db, nil)

	// Test with cursor parameter
	params := domain.WebhookEventListParams{
		Limit:  10,
		Cursor: cursor,
	}

	// Set up rows for the SQL query result
	rows := sqlmock.NewRows([]string{
		"id", "type", "email_provider_kind", "integration_id", "recipient_email",
		"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
		"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type",
		"created_at",
	}).
		AddRow(
			"evt-3", domain.EmailEventDelivered, domain.EmailProviderKindSES, "integration-3", "test3@example.com",
			"msg-3", "trans-3", "broadcast-3", now, `{"key": "value3"}`,
			"", "", "", "", now,
		).
		AddRow(
			"evt-4", domain.EmailEventDelivered, domain.EmailProviderKindSES, "integration-4", "test4@example.com",
			"msg-4", "trans-4", "broadcast-4", now, `{"key": "value4"}`,
			"", "", "", "", now,
		).
		AddRow(
			"evt-5", domain.EmailEventDelivered, domain.EmailProviderKindSES, "integration-5", "test5@example.com",
			"msg-5", "trans-5", "broadcast-5", now, `{"key": "value5"}`,
			"", "", "", "", now,
		)

	// Expect a SQL query with cursor condition
	mock.ExpectQuery(`SELECT .+ FROM webhook_events WHERE`).
		WillReturnRows(rows)

	// Call the method
	result, err := repo.ListEvents(ctx, workspaceID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Events, 3)
	assert.False(t, result.HasMore) // With 3 results and limit 10, HasMore should be false

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookEventRepository_ListEvents_InvalidCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewWebhookEventRepository(mockWorkspaceRepo)

	ctx := context.Background()
	workspaceID := "ws-123"

	// Test cases for invalid cursors
	testCases := []struct {
		name   string
		cursor string
	}{
		{
			name:   "Invalid base64",
			cursor: "not-base64!",
		},
		{
			name:   "Invalid format",
			cursor: "aW52YWxpZC1mb3JtYXQ=", // base64 of "invalid-format"
		},
		{
			name:   "Invalid timestamp",
			cursor: "bm90LWEtdGltZXN0YW1wfmV2dC0xMjM=", // base64 of "not-a-timestamp~evt-123"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := domain.WebhookEventListParams{
				Limit:  10,
				Cursor: tc.cursor,
			}

			// Set up the workspace connection expectation for each test case
			mockWorkspaceRepo.EXPECT().
				GetConnection(gomock.Any(), workspaceID).
				Return(nil, errors.New("connection error"))

			_, err := repo.ListEvents(ctx, workspaceID, params)
			assert.Error(t, err)
		})
	}
}

func TestWebhookEventRepository_ListEvents_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewWebhookEventRepository(mockWorkspaceRepo)

	ctx := context.Background()
	workspaceID := "ws-123"
	params := domain.WebhookEventListParams{Limit: 10}

	// Test case 1: Database connection error
	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(nil, errors.New("connection error"))

	result, err := repo.ListEvents(ctx, workspaceID, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get workspace connection")
	assert.Nil(t, result)

	// Test case 2: SQL execution error
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(db, nil)

	mock.ExpectQuery("SELECT").
		WillReturnError(errors.New("database error"))

	result, err = repo.ListEvents(ctx, workspaceID, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to query webhook events")
	assert.Nil(t, result)

	// Test case 3: Scan error
	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspaceID).
		Return(db, nil)

	rows := sqlmock.NewRows([]string{"id"}). // Deliberately wrong number of columns
							AddRow("evt-1")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err = repo.ListEvents(ctx, workspaceID, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan webhook event row")
	assert.Nil(t, result)
}
