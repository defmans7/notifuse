package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/repository"
)

func TestWebhookEventRepository_StoreEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create test webhook event
	event := domain.NewWebhookEvent(
		"event-123",
		domain.EmailEventDelivered,
		domain.EmailProviderKindPostmark,
		"integration-123",
		"test@example.com",
		"msg-123",
		time.Now(),
		`{"event":"delivered"}`,
	)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the correct insert query with parameters
		mock.ExpectExec("INSERT INTO webhook_events").
			WithArgs(
				event.ID,
				string(event.Type),
				string(event.EmailProviderKind),
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
				sqlmock.AnyArg(), // CreatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Call the repository method
		err = eventRepo.StoreEvent(ctx, event)
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("connection error")

		// Mock workspace repository GetConnection with error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(nil, expectedErr)

		// Call the repository method
		err = eventRepo.StoreEvent(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get system database connection")
	})

	t.Run("Database insert error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("insert error")

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the insert query but return an error
		mock.ExpectExec("INSERT INTO webhook_events").
			WithArgs(
				event.ID,
				string(event.Type),
				string(event.EmailProviderKind),
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
				sqlmock.AnyArg(), // CreatedAt
			).
			WillReturnError(dbErr)

		// Call the repository method
		err = eventRepo.StoreEvent(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store webhook event")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookEventRepository_GetEventByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	eventID := "event-123"
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define the expected rows
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		}).AddRow(
			eventID, "delivered", "postmark", "integration-123", "test@example.com",
			"msg-123", "", "", now, `{"event":"delivered"}`,
			"", "", "", "", now,
		)

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE id = \\$1").
			WithArgs(eventID).
			WillReturnRows(rows)

		// Call the repository method
		event, err := eventRepo.GetEventByID(ctx, eventID)
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, eventID, event.ID)
		assert.Equal(t, domain.EmailEventDelivered, event.Type)
		assert.Equal(t, domain.EmailProviderKindPostmark, event.EmailProviderKind)
		assert.Equal(t, "test@example.com", event.RecipientEmail)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not found", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the query with no rows
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE id = \\$1").
			WithArgs(eventID).
			WillReturnError(sql.ErrNoRows)

		// Call the repository method
		event, err := eventRepo.GetEventByID(ctx, eventID)
		assert.Error(t, err)
		assert.Nil(t, event)

		// Verify the error type
		var notFoundErr *domain.ErrWebhookEventNotFound
		assert.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, eventID, notFoundErr.ID)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("database error")

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the query with error
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE id = \\$1").
			WithArgs(eventID).
			WillReturnError(dbErr)

		// Call the repository method
		event, err := eventRepo.GetEventByID(ctx, eventID)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "failed to get webhook event")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookEventRepository_GetEventsByMessageID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	messageID := "msg-123"
	now := time.Now()
	limit := 10
	offset := 0

	t.Run("Success with results", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define the expected rows
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		}).
			AddRow(
				"event-1", "delivered", "postmark", "integration-123", "user1@example.com",
				messageID, "trans-1", "", now, `{"event":"delivered"}`,
				"", "", "", "", now,
			).
			AddRow(
				"event-2", "bounce", "postmark", "integration-123", "user2@example.com",
				messageID, "trans-1", "", now.Add(-1*time.Hour), `{"event":"bounce"}`,
				"hard", "5.0.0", "rejected", "", now,
			)

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE message_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(messageID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByMessageID(ctx, messageID, limit, offset)
		assert.NoError(t, err)
		assert.Len(t, events, 2)

		// Verify first event properties
		assert.Equal(t, "event-1", events[0].ID)
		assert.Equal(t, domain.EmailEventDelivered, events[0].Type)
		assert.Equal(t, "trans-1", events[0].TransactionalID)

		// Verify second event properties
		assert.Equal(t, "event-2", events[1].ID)
		assert.Equal(t, domain.EmailEventBounce, events[1].Type)
		assert.Equal(t, "hard", events[1].BounceType)
		assert.Equal(t, "5.0.0", events[1].BounceCategory)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with no results", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define empty result set
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		})

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE message_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(messageID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByMessageID(ctx, messageID, limit, offset)
		assert.NoError(t, err)
		assert.Empty(t, events)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("connection error")

		// Mock workspace repository GetConnection with error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(nil, expectedErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByMessageID(ctx, messageID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get system database connection")
	})

	t.Run("Database query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("query error")

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the query with error
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE message_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(messageID, limit, offset).
			WillReturnError(dbErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByMessageID(ctx, messageID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get webhook events by message ID")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Row scan error", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define rows with wrong column count (missing some columns)
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind",
			// Missing many required columns which will cause scan error
		}).AddRow("event-1", "delivered", "postmark")

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE message_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(messageID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByMessageID(ctx, messageID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to scan webhook event row")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookEventRepository_GetEventsByTransactionalID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	transactionalID := "trans-123"
	now := time.Now()
	limit := 10
	offset := 0

	t.Run("Success with results", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define the expected rows
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		}).
			AddRow(
				"event-1", "delivered", "postmark", "integration-123", "user1@example.com",
				"msg-1", transactionalID, "", now, `{"event":"delivered"}`,
				"", "", "", "", now,
			).
			AddRow(
				"event-2", "complaint", "postmark", "integration-123", "user2@example.com",
				"msg-2", transactionalID, "", now.Add(-1*time.Hour), `{"event":"complaint"}`,
				"", "", "", "abuse", now,
			)

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE transactional_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(transactionalID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByTransactionalID(ctx, transactionalID, limit, offset)
		assert.NoError(t, err)
		assert.Len(t, events, 2)

		// Verify first event properties
		assert.Equal(t, "event-1", events[0].ID)
		assert.Equal(t, domain.EmailEventDelivered, events[0].Type)
		assert.Equal(t, transactionalID, events[0].TransactionalID)

		// Verify second event properties
		assert.Equal(t, "event-2", events[1].ID)
		assert.Equal(t, domain.EmailEventComplaint, events[1].Type)
		assert.Equal(t, "abuse", events[1].ComplaintFeedbackType)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with no results", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define empty result set
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		})

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE transactional_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(transactionalID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByTransactionalID(ctx, transactionalID, limit, offset)
		assert.NoError(t, err)
		assert.Empty(t, events)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("connection error")

		// Mock workspace repository GetConnection with error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(nil, expectedErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByTransactionalID(ctx, transactionalID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get system database connection")
	})

	t.Run("Database query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("query error")

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the query with error
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE transactional_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(transactionalID, limit, offset).
			WillReturnError(dbErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByTransactionalID(ctx, transactionalID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get webhook events by transactional ID")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookEventRepository_GetEventsByBroadcastID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	broadcastID := "broadcast-123"
	now := time.Now()
	limit := 10
	offset := 0

	t.Run("Success with results", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define the expected rows
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		}).
			AddRow(
				"event-1", "delivered", "postmark", "integration-123", "user1@example.com",
				"msg-1", "", broadcastID, now, `{"event":"delivered"}`,
				"", "", "", "", now,
			).
			AddRow(
				"event-2", "bounce", "postmark", "integration-123", "user2@example.com",
				"msg-2", "", broadcastID, now.Add(-1*time.Hour), `{"event":"bounce"}`,
				"hard", "5.1.1", "rejected", "", now,
			)

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE broadcast_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(broadcastID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByBroadcastID(ctx, broadcastID, limit, offset)
		assert.NoError(t, err)
		assert.Len(t, events, 2)

		// Verify first event properties
		assert.Equal(t, "event-1", events[0].ID)
		assert.Equal(t, domain.EmailEventDelivered, events[0].Type)
		assert.Equal(t, broadcastID, events[0].BroadcastID)

		// Verify second event properties
		assert.Equal(t, "event-2", events[1].ID)
		assert.Equal(t, domain.EmailEventBounce, events[1].Type)
		assert.Equal(t, "hard", events[1].BounceType)
		assert.Equal(t, "5.1.1", events[1].BounceCategory)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with no results", func(t *testing.T) {
		ctx := context.Background()

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Define empty result set
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		})

		// Expect the query
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE broadcast_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(broadcastID, limit, offset).
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByBroadcastID(ctx, broadcastID, limit, offset)
		assert.NoError(t, err)
		assert.Empty(t, events)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("connection error")

		// Mock workspace repository GetConnection with error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(nil, expectedErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByBroadcastID(ctx, broadcastID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get system database connection")
	})

	t.Run("Database query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("query error")

		// Mock workspace repository GetConnection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Expect the query with error
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE broadcast_id = \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs(broadcastID, limit, offset).
			WillReturnError(dbErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByBroadcastID(ctx, broadcastID, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get webhook events by broadcast ID")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookEventRepository_GetEventsByType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	workspaceID := "workspace-123"
	eventType := domain.EmailEventDelivered
	now := time.Now()
	limit := 10
	offset := 0

	t.Run("Success with results", func(t *testing.T) {
		ctx := context.Background()

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query
		transRows := sqlmock.NewRows([]string{"id"}).
			AddRow("trans-1").
			AddRow("trans-2")

		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Setup mock for broadcast IDs query
		broadcastRows := sqlmock.NewRows([]string{"id"}).
			AddRow("broadcast-1").
			AddRow("broadcast-2")

		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(broadcastRows)

		// Define the expected rows for webhook events
		rows := sqlmock.NewRows([]string{
			"id", "type", "email_provider_kind", "integration_id", "recipient_email",
			"message_id", "transactional_id", "broadcast_id", "timestamp", "raw_payload",
			"bounce_type", "bounce_category", "bounce_diagnostic", "complaint_feedback_type", "created_at",
		}).
			AddRow(
				"event-1", string(eventType), "postmark", "integration-123", "user1@example.com",
				"msg-1", "trans-1", "", now, `{"event":"delivered"}`,
				"", "", "", "", now,
			).
			AddRow(
				"event-2", string(eventType), "postmark", "integration-123", "user2@example.com",
				"msg-2", "trans-2", "", now.Add(-1*time.Hour), `{"event":"delivered"}`,
				"", "", "", "", now,
			).
			AddRow(
				"event-3", string(eventType), "postmark", "integration-123", "user3@example.com",
				"msg-3", "", "broadcast-1", now.Add(-2*time.Hour), `{"event":"delivered"}`,
				"", "", "", "", now,
			)

		// This regex is a bit more complex as the actual query is built dynamically
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE (.+) ORDER BY timestamp DESC LIMIT \\$\\d+ OFFSET \\$\\d+").
			WillReturnRows(rows)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.NoError(t, err)
		assert.Len(t, events, 3)

		// Verify event properties
		assert.Equal(t, "event-1", events[0].ID)
		assert.Equal(t, eventType, events[0].Type)
		assert.Equal(t, "trans-1", events[0].TransactionalID)

		assert.Equal(t, "event-2", events[1].ID)
		assert.Equal(t, "trans-2", events[1].TransactionalID)

		assert.Equal(t, "event-3", events[2].ID)
		assert.Equal(t, "broadcast-1", events[2].BroadcastID)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No transactional or broadcast IDs found", func(t *testing.T) {
		ctx := context.Background()

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query - empty result
		transRows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Setup mock for broadcast IDs query - empty result
		broadcastRows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(broadcastRows)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.NoError(t, err)
		assert.Empty(t, events)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("System database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("system connection error")

		// Mock system database connection error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(nil, expectedErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get system database connection")
	})

	t.Run("Workspace database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("workspace connection error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, expectedErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("Transactional IDs query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("transactional query error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Expect the transactional IDs query with error
		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnError(dbErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get transactional notification IDs")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Broadcast IDs query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("broadcast query error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query
		transRows := sqlmock.NewRows([]string{"id"}).
			AddRow("trans-1")

		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Expect the broadcast IDs query with error
		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnError(dbErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get broadcast IDs")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Webhook events query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("events query error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query
		transRows := sqlmock.NewRows([]string{"id"}).
			AddRow("trans-1")

		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Setup mock for broadcast IDs query
		broadcastRows := sqlmock.NewRows([]string{"id"}).
			AddRow("broadcast-1")

		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(broadcastRows)

		// Expect the webhook events query with error
		mock.ExpectQuery("SELECT \\* FROM webhook_events WHERE (.+) ORDER BY timestamp DESC LIMIT \\$\\d+ OFFSET \\$\\d+").
			WillReturnError(dbErr)

		// Call the repository method
		events, err := eventRepo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Contains(t, err.Error(), "failed to get webhook events by type")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookEventRepository_GetEventCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	eventRepo := repository.NewWebhookEventRepository(mockWorkspaceRepo)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	workspaceID := "workspace-123"
	eventType := domain.EmailEventDelivered

	t.Run("Success with count", func(t *testing.T) {
		ctx := context.Background()
		expectedCount := 42

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query
		transRows := sqlmock.NewRows([]string{"id"}).
			AddRow("trans-1").
			AddRow("trans-2")

		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Setup mock for broadcast IDs query
		broadcastRows := sqlmock.NewRows([]string{"id"}).
			AddRow("broadcast-1").
			AddRow("broadcast-2")

		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(broadcastRows)

		// Setup mock for count query
		countRow := sqlmock.NewRows([]string{"count"}).
			AddRow(expectedCount)

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM webhook_events WHERE (.+)").
			WillReturnRows(countRow)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No transactional or broadcast IDs found", func(t *testing.T) {
		ctx := context.Background()

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query - empty result
		transRows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Setup mock for broadcast IDs query - empty result
		broadcastRows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(broadcastRows)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("System database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("system connection error")

		// Mock system database connection error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(nil, expectedErr)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get system database connection")
	})

	t.Run("Workspace database connection error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("workspace connection error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection error
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, expectedErr)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("Transactional IDs query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("transactional query error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Expect the transactional IDs query with error
		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnError(dbErr)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get transactional notification IDs")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Broadcast IDs query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("broadcast query error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query
		transRows := sqlmock.NewRows([]string{"id"}).
			AddRow("trans-1")

		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Expect the broadcast IDs query with error
		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnError(dbErr)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get broadcast IDs")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Count query error", func(t *testing.T) {
		ctx := context.Background()
		dbErr := errors.New("count query error")

		// Mock system database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, "system").
			Return(db, nil)

		// Mock workspace database connection
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Setup mock for transactional IDs query
		transRows := sqlmock.NewRows([]string{"id"}).
			AddRow("trans-1")

		mock.ExpectQuery("SELECT id FROM transactional_notifications WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(transRows)

		// Setup mock for broadcast IDs query
		broadcastRows := sqlmock.NewRows([]string{"id"}).
			AddRow("broadcast-1")

		mock.ExpectQuery("SELECT id FROM broadcasts WHERE workspace_id = \\$1").
			WithArgs(workspaceID).
			WillReturnRows(broadcastRows)

		// Expect the count query with error
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM webhook_events WHERE (.+)").
			WillReturnError(dbErr)

		// Call the repository method
		count, err := eventRepo.GetEventCount(ctx, workspaceID, eventType)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get webhook event count")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
