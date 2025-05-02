// Tests for message_history_repository.go
// Run with: go test -v ./internal/repository -run="^TestMessageHistory"

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMessageHistoryMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *MessageHistoryRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	repo := NewMessageHistoryRepository(db)
	return db, mock, repo
}

func createTestMessageHistory() *domain.MessageHistory {
	now := time.Now().UTC().Truncate(time.Second)
	broadcastID := "broadcast123"

	return &domain.MessageHistory{
		ID:              "msg123",
		ContactID:       "contact456",
		BroadcastID:     &broadcastID,
		TemplateID:      "template789",
		TemplateVersion: 1,
		Channel:         "email",
		Status:          domain.MessageStatusSent,
		MessageData: domain.MessageData{
			Data: map[string]interface{}{
				"subject": "Welcome!",
				"name":    "John",
			},
			Metadata: map[string]interface{}{
				"campaign": "onboarding",
			},
		},
		SentAt:    now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMessageHistoryRepository_Create(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	message := createTestMessageHistory()

	// Expect the INSERT query
	mock.ExpectExec(`INSERT INTO message_history \(.*\) VALUES \(.*\)`).
		WithArgs(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			sqlmock.AnyArg(), // MessageData is serialized to JSON, so we can't match exactly
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			message.CreatedAt,
			message.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method
	err := repo.Create(ctx, workspace, message)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test database error
	mock.ExpectExec(`INSERT INTO message_history \(.*\) VALUES \(.*\)`).
		WithArgs(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			sqlmock.AnyArg(), // MessageData
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			message.CreatedAt,
			message.UpdatedAt,
		).
		WillReturnError(errors.New("database error"))

	err = repo.Create(ctx, workspace, message)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create message history")
	assert.Contains(t, err.Error(), "database error")
}

func TestMessageHistoryRepository_Update(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	message := createTestMessageHistory()

	// Expect the UPDATE query
	mock.ExpectExec(`UPDATE message_history SET.*WHERE id = \$1`).
		WithArgs(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			sqlmock.AnyArg(), // MessageData
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			sqlmock.AnyArg(), // updated_at is set to time.Now() in the method
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method
	err := repo.Update(ctx, workspace, message)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test database error
	mock.ExpectExec(`UPDATE message_history SET.*WHERE id = \$1`).
		WithArgs(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			sqlmock.AnyArg(), // MessageData
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnError(errors.New("database error"))

	err = repo.Update(ctx, workspace, message)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update message history")
	assert.Contains(t, err.Error(), "database error")
}

func TestMessageHistoryRepository_Get(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	messageID := "msg123"
	message := createTestMessageHistory()

	columns := []string{
		"id", "contact_id", "broadcast_id", "template_id", "template_version",
		"channel", "status", "message_data", "sent_at", "delivered_at",
		"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
		"unsubscribed_at", "created_at", "updated_at",
	}

	// Serialize message data to JSON for the mock
	messageDataJSON, err := json.Marshal(message.MessageData)
	require.NoError(t, err)

	// Expect the SELECT query
	rows := sqlmock.NewRows(columns).
		AddRow(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			messageDataJSON,
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			message.CreatedAt,
			message.UpdatedAt,
		)

	mock.ExpectQuery(`SELECT.*FROM message_history.*WHERE id = \$1`).
		WithArgs(messageID).
		WillReturnRows(rows)

	// Call the method
	result, err := repo.Get(ctx, workspace, messageID)
	require.NoError(t, err)
	assert.Equal(t, message.ID, result.ID)
	assert.Equal(t, message.ContactID, result.ContactID)
	assert.Equal(t, *message.BroadcastID, *result.BroadcastID)
	assert.Equal(t, message.TemplateID, result.TemplateID)
	assert.Equal(t, message.TemplateVersion, result.TemplateVersion)
	assert.Equal(t, message.Channel, result.Channel)
	assert.Equal(t, message.Status, result.Status)
	assert.Equal(t, message.SentAt.Unix(), result.SentAt.Unix())
	assert.Equal(t, message.CreatedAt.Unix(), result.CreatedAt.Unix())
	assert.Equal(t, message.UpdatedAt.Unix(), result.UpdatedAt.Unix())
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test not found
	mock.ExpectQuery(`SELECT.*FROM message_history.*WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	result, err = repo.Get(ctx, workspace, "nonexistent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "message history with id nonexistent not found")

	// Test database error
	mock.ExpectQuery(`SELECT.*FROM message_history.*WHERE id = \$1`).
		WithArgs("error-id").
		WillReturnError(errors.New("database error"))

	result, err = repo.Get(ctx, workspace, "error-id")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get message history")
}

func TestMessageHistoryRepository_GetByContact(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	contactID := "contact456"
	limit := 10
	offset := 0
	totalCount := 2

	// Create test data
	message1 := createTestMessageHistory()
	message2 := createTestMessageHistory()
	message2.ID = "msg456"
	message2.SentAt = message1.SentAt.Add(-time.Hour) // Older message

	// Serialize message data to JSON for the mock
	messageData1JSON, err := json.Marshal(message1.MessageData)
	require.NoError(t, err)
	messageData2JSON, err := json.Marshal(message2.MessageData)
	require.NoError(t, err)

	t.Run("successful case", func(t *testing.T) {
		// 1. Expect the COUNT query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		// 2. Expect the main SELECT query
		columns := []string{
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}

		rows := sqlmock.NewRows(columns).
			AddRow(
				message1.ID,
				message1.ContactID,
				message1.BroadcastID,
				message1.TemplateID,
				message1.TemplateVersion,
				message1.Channel,
				message1.Status,
				messageData1JSON,
				message1.SentAt,
				message1.DeliveredAt,
				message1.FailedAt,
				message1.OpenedAt,
				message1.ClickedAt,
				message1.BouncedAt,
				message1.ComplainedAt,
				message1.UnsubscribedAt,
				message1.CreatedAt,
				message1.UpdatedAt,
			).
			AddRow(
				message2.ID,
				message2.ContactID,
				message2.BroadcastID,
				message2.TemplateID,
				message2.TemplateVersion,
				message2.Channel,
				message2.Status,
				messageData2JSON,
				message2.SentAt,
				message2.DeliveredAt,
				message2.FailedAt,
				message2.OpenedAt,
				message2.ClickedAt,
				message2.BouncedAt,
				message2.ComplainedAt,
				message2.UnsubscribedAt,
				message2.CreatedAt,
				message2.UpdatedAt,
			)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, limit, offset).
			WillReturnRows(rows)

		// Call the method
		messages, count, err := repo.GetByContact(ctx, workspace, contactID, limit, offset)
		require.NoError(t, err)
		assert.Equal(t, totalCount, count)
		assert.Len(t, messages, 2)
		assert.Equal(t, message1.ID, messages[0].ID)
		assert.Equal(t, message2.ID, messages[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("default limit and offset", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test with default limit and offset
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		columns := []string{
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}

		rows := sqlmock.NewRows(columns).
			AddRow(
				message1.ID,
				message1.ContactID,
				message1.BroadcastID,
				message1.TemplateID,
				message1.TemplateVersion,
				message1.Channel,
				message1.Status,
				messageData1JSON,
				message1.SentAt,
				message1.DeliveredAt,
				message1.FailedAt,
				message1.OpenedAt,
				message1.ClickedAt,
				message1.BouncedAt,
				message1.ComplainedAt,
				message1.UnsubscribedAt,
				message1.CreatedAt,
				message1.UpdatedAt,
			).
			AddRow(
				message2.ID,
				message2.ContactID,
				message2.BroadcastID,
				message2.TemplateID,
				message2.TemplateVersion,
				message2.Channel,
				message2.Status,
				messageData2JSON,
				message2.SentAt,
				message2.DeliveredAt,
				message2.FailedAt,
				message2.OpenedAt,
				message2.ClickedAt,
				message2.BouncedAt,
				message2.ComplainedAt,
				message2.UnsubscribedAt,
				message2.CreatedAt,
				message2.UpdatedAt,
			)

		newMock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, 50, 0). // Default values
			WillReturnRows(rows)

		messages, count, err := newRepo.GetByContact(ctx, workspace, contactID, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, totalCount, count)
		assert.Len(t, messages, 2)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("count error", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test count error
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnError(fmt.Errorf("count error"))

		messages, count, err := newRepo.GetByContact(ctx, workspace, contactID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Zero(t, count)
		assert.Contains(t, err.Error(), "failed to count message history")
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test query error
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		newMock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, limit, offset).
			WillReturnError(fmt.Errorf("query error"))

		messages, count, err := newRepo.GetByContact(ctx, workspace, contactID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Zero(t, count)
		assert.Contains(t, err.Error(), "failed to query message history")
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test scan error
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		invalidRows := sqlmock.NewRows([]string{"id"}).AddRow("invalid") // Missing columns
		newMock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, limit, offset).
			WillReturnRows(invalidRows)

		messages, count, err := newRepo.GetByContact(ctx, workspace, contactID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Zero(t, count)
		assert.Contains(t, err.Error(), "failed to scan message history")
		assert.NoError(t, newMock.ExpectationsWereMet())
	})
}

func TestMessageHistoryRepository_GetByBroadcast(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	broadcastID := "broadcast123"
	limit := 10
	offset := 0
	totalCount := 2

	// Create test data
	message1 := createTestMessageHistory()
	message2 := createTestMessageHistory()
	message2.ID = "msg456"
	message2.SentAt = message1.SentAt.Add(-time.Hour) // Older message

	// Serialize message data to JSON for the mock
	messageData1JSON, err := json.Marshal(message1.MessageData)
	require.NoError(t, err)
	messageData2JSON, err := json.Marshal(message2.MessageData)
	require.NoError(t, err)

	t.Run("successful case", func(t *testing.T) {
		// 1. Expect the COUNT query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(countRows)

		// 2. Expect the main SELECT query
		columns := []string{
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}

		rows := sqlmock.NewRows(columns).
			AddRow(
				message1.ID,
				message1.ContactID,
				message1.BroadcastID,
				message1.TemplateID,
				message1.TemplateVersion,
				message1.Channel,
				message1.Status,
				messageData1JSON,
				message1.SentAt,
				message1.DeliveredAt,
				message1.FailedAt,
				message1.OpenedAt,
				message1.ClickedAt,
				message1.BouncedAt,
				message1.ComplainedAt,
				message1.UnsubscribedAt,
				message1.CreatedAt,
				message1.UpdatedAt,
			).
			AddRow(
				message2.ID,
				message2.ContactID,
				message2.BroadcastID,
				message2.TemplateID,
				message2.TemplateVersion,
				message2.Channel,
				message2.Status,
				messageData2JSON,
				message2.SentAt,
				message2.DeliveredAt,
				message2.FailedAt,
				message2.OpenedAt,
				message2.ClickedAt,
				message2.BouncedAt,
				message2.ComplainedAt,
				message2.UnsubscribedAt,
				message2.CreatedAt,
				message2.UpdatedAt,
			)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(broadcastID, limit, offset).
			WillReturnRows(rows)

		// Call the method
		messages, count, err := repo.GetByBroadcast(ctx, workspace, broadcastID, limit, offset)
		require.NoError(t, err)
		assert.Equal(t, totalCount, count)
		assert.Len(t, messages, 2)
		assert.Equal(t, message1.ID, messages[0].ID)
		assert.Equal(t, message2.ID, messages[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("default limit and offset", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test with default limit and offset
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(countRows)

		columns := []string{
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}

		rows := sqlmock.NewRows(columns).
			AddRow(
				message1.ID,
				message1.ContactID,
				message1.BroadcastID,
				message1.TemplateID,
				message1.TemplateVersion,
				message1.Channel,
				message1.Status,
				messageData1JSON,
				message1.SentAt,
				message1.DeliveredAt,
				message1.FailedAt,
				message1.OpenedAt,
				message1.ClickedAt,
				message1.BouncedAt,
				message1.ComplainedAt,
				message1.UnsubscribedAt,
				message1.CreatedAt,
				message1.UpdatedAt,
			).
			AddRow(
				message2.ID,
				message2.ContactID,
				message2.BroadcastID,
				message2.TemplateID,
				message2.TemplateVersion,
				message2.Channel,
				message2.Status,
				messageData2JSON,
				message2.SentAt,
				message2.DeliveredAt,
				message2.FailedAt,
				message2.OpenedAt,
				message2.ClickedAt,
				message2.BouncedAt,
				message2.ComplainedAt,
				message2.UnsubscribedAt,
				message2.CreatedAt,
				message2.UpdatedAt,
			)

		newMock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(broadcastID, 50, 0). // Default values
			WillReturnRows(rows)

		messages, count, err := newRepo.GetByBroadcast(ctx, workspace, broadcastID, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, totalCount, count)
		assert.Len(t, messages, 2)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("count error", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test count error
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnError(fmt.Errorf("count error"))

		messages, count, err := newRepo.GetByBroadcast(ctx, workspace, broadcastID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Zero(t, count)
		assert.Contains(t, err.Error(), "failed to count message history")
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test query error
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(countRows)

		newMock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(broadcastID, limit, offset).
			WillReturnError(fmt.Errorf("query error"))

		messages, count, err := newRepo.GetByBroadcast(ctx, workspace, broadcastID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Zero(t, count)
		assert.Contains(t, err.Error(), "failed to query message history")
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Setup new mock
		newDb, newMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer newDb.Close()
		newRepo := NewMessageHistoryRepository(newDb)

		// Test scan error
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
		newMock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(countRows)

		invalidRows := sqlmock.NewRows([]string{"id"}).AddRow("invalid") // Missing columns
		newMock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(broadcastID, limit, offset).
			WillReturnRows(invalidRows)

		messages, count, err := newRepo.GetByBroadcast(ctx, workspace, broadcastID, limit, offset)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Zero(t, count)
		assert.Contains(t, err.Error(), "failed to scan message history")
		assert.NoError(t, newMock.ExpectationsWereMet())
	})
}

func TestMessageHistoryRepository_UpdateStatus(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	messageID := "msg123"
	timestamp := time.Now().UTC().Truncate(time.Second)

	testCases := []struct {
		name      string
		status    domain.MessageStatus
		field     string
		expectErr bool
	}{
		{"delivered", domain.MessageStatusDelivered, "delivered_at", false},
		{"failed", domain.MessageStatusFailed, "failed_at", false},
		{"opened", domain.MessageStatusOpened, "opened_at", false},
		{"clicked", domain.MessageStatusClicked, "clicked_at", false},
		{"bounced", domain.MessageStatusBounced, "bounced_at", false},
		{"complained", domain.MessageStatusComplained, "complained_at", false},
		{"unsubscribed", domain.MessageStatusUnsubscribed, "unsubscribed_at", false},
		{"invalid status", "invalid", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectErr {
				// No need to setup mock expectations for invalid status
				err := repo.UpdateStatus(ctx, workspace, messageID, tc.status, timestamp)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid status")
			} else {
				// Setup mock expectations for valid status
				mock.ExpectExec(`UPDATE message_history SET status = \$1, `+tc.field+` = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs(tc.status, timestamp, sqlmock.AnyArg(), messageID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				err := repo.UpdateStatus(ctx, workspace, messageID, tc.status, timestamp)
				require.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())

				// Test database error
				mock.ExpectExec(`UPDATE message_history SET status = \$1, `+tc.field+` = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs(tc.status, timestamp, sqlmock.AnyArg(), messageID).
					WillReturnError(errors.New("database error"))

				err = repo.UpdateStatus(ctx, workspace, messageID, tc.status, timestamp)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to update message status")
				assert.Contains(t, err.Error(), "database error")
			}
		})
	}
}

func TestMessageHistoryRepository_GetByContact_Simple(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	contactID := "contact456"
	limit := 10
	offset := 0
	totalCount := 1

	// Create test data
	message := createTestMessageHistory()

	// Serialize message data to JSON for the mock
	messageDataJSON, err := json.Marshal(message.MessageData)
	require.NoError(t, err)

	// 1. Expect the COUNT query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
		WithArgs(contactID).
		WillReturnRows(countRows)

	// 2. Expect the main SELECT query
	columns := []string{
		"id", "contact_id", "broadcast_id", "template_id", "template_version",
		"channel", "status", "message_data", "sent_at", "delivered_at",
		"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
		"unsubscribed_at", "created_at", "updated_at",
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			messageDataJSON,
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			message.CreatedAt,
			message.UpdatedAt,
		)

	// The expected query should include ORDER BY, LIMIT, and OFFSET
	mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(contactID, limit, offset).
		WillReturnRows(rows)

	// Call the method
	messages, count, err := repo.GetByContact(ctx, workspace, contactID, limit, offset)
	require.NoError(t, err)
	assert.Equal(t, totalCount, count)
	assert.Len(t, messages, 1)
	assert.Equal(t, message.ID, messages[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHistoryRepository_GetByBroadcast_Simple(t *testing.T) {
	db, mock, repo := setupMessageHistoryMock(t)
	defer db.Close()

	ctx := context.Background()
	workspace := "testworkspace"
	broadcastID := "broadcast123"
	limit := 10
	offset := 0
	totalCount := 1

	// Create test data
	message := createTestMessageHistory()

	// Serialize message data to JSON for the mock
	messageDataJSON, err := json.Marshal(message.MessageData)
	require.NoError(t, err)

	// 1. Expect the COUNT query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
		WithArgs(broadcastID).
		WillReturnRows(countRows)

	// 2. Expect the main SELECT query
	columns := []string{
		"id", "contact_id", "broadcast_id", "template_id", "template_version",
		"channel", "status", "message_data", "sent_at", "delivered_at",
		"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
		"unsubscribed_at", "created_at", "updated_at",
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
			messageDataJSON,
			message.SentAt,
			message.DeliveredAt,
			message.FailedAt,
			message.OpenedAt,
			message.ClickedAt,
			message.BouncedAt,
			message.ComplainedAt,
			message.UnsubscribedAt,
			message.CreatedAt,
			message.UpdatedAt,
		)

	// The expected query should include ORDER BY, LIMIT, and OFFSET
	mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(broadcastID, limit, offset).
		WillReturnRows(rows)

	// Call the method
	messages, count, err := repo.GetByBroadcast(ctx, workspace, broadcastID, limit, offset)
	require.NoError(t, err)
	assert.Equal(t, totalCount, count)
	assert.Len(t, messages, 1)
	assert.Equal(t, message.ID, messages[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
