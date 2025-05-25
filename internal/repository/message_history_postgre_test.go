package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

func setupMessageHistoryTest(t *testing.T) (*mocks.MockWorkspaceRepository, domain.MessageHistoryRepository, sqlmock.Sqlmock, *sql.DB, func()) {
	ctrl := gomock.NewController(t)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Create a real DB connection with sqlmock
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewMessageHistoryRepository(mockWorkspaceRepo)

	// Set up cleanup function
	cleanup := func() {
		db.Close()
		ctrl.Finish()
	}

	return mockWorkspaceRepo, repo, mock, db, cleanup
}

func createSampleMessageHistory() *domain.MessageHistory {
	now := time.Now()
	broadcastID := "broadcast-123"
	messageData := domain.MessageData{
		Data: map[string]interface{}{
			"subject": "Test subject",
			"body":    "Test body",
		},
	}

	return &domain.MessageHistory{
		ID:              "msg-123",
		ContactEmail:    "contact-123",
		BroadcastID:     &broadcastID,
		TemplateID:      "template-123",
		TemplateVersion: 1,
		Channel:         "email",
		MessageData:     messageData,
		SentAt:          now,
		DeliveredAt:     nil,
		FailedAt:        nil,
		OpenedAt:        nil,
		ClickedAt:       nil,
		BouncedAt:       nil,
		ComplainedAt:    nil,
		UnsubscribedAt:  nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func TestMessageHistoryRepository_Create(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	message := createSampleMessageHistory()

	t.Run("successful creation", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectExec(`INSERT INTO message_history`).
			WithArgs(
				message.ID,
				message.ContactEmail,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				sqlmock.AnyArg(), // message_data - complex type
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

		err := repo.Create(ctx, workspaceID, message)
		require.NoError(t, err)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		err := repo.Create(ctx, workspaceID, message)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("execution error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectExec(`INSERT INTO message_history`).
			WithArgs(
				message.ID,
				message.ContactEmail,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				sqlmock.AnyArg(), // message_data
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
			WillReturnError(errors.New("execution error"))

		err := repo.Create(ctx, workspaceID, message)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create message history")
	})
}

func TestMessageHistoryRepository_Update(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	message := createSampleMessageHistory()

	t.Run("successful update", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectExec(`UPDATE message_history SET`).
			WithArgs(
				message.ID,
				message.ContactEmail,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				sqlmock.AnyArg(), // message_data
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
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Update(ctx, workspaceID, message)
		require.NoError(t, err)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		err := repo.Update(ctx, workspaceID, message)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("execution error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectExec(`UPDATE message_history SET`).
			WithArgs(
				message.ID,
				message.ContactEmail,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				sqlmock.AnyArg(), // message_data
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
			WillReturnError(errors.New("execution error"))

		err := repo.Update(ctx, workspaceID, message)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to update message history")
	})
}

func TestMessageHistoryRepository_Get(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "msg-123"
	message := createSampleMessageHistory()

	// Convert the MessageData to JSON for proper DB response mocking
	messageDataJSON, _ := json.Marshal(message.MessageData)

	t.Run("successful retrieval", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		rows := sqlmock.NewRows([]string{
			"id", "contact_email", "broadcast_id", "template_id", "template_version",
			"channel", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}).AddRow(
			message.ID,
			message.ContactEmail,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			messageDataJSON, // Use the actual JSON bytes
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

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE id = \$1`).
			WithArgs(messageID).
			WillReturnRows(rows)

		result, err := repo.Get(ctx, workspaceID, messageID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, message.ID, result.ID)
		assert.Equal(t, message.ContactEmail, result.ContactEmail)
		assert.Equal(t, *message.BroadcastID, *result.BroadcastID)
	})

	t.Run("not found", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE id = \$1`).
			WithArgs(messageID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.Get(ctx, workspaceID, messageID)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "message history with id msg-123 not found")
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		result, err := repo.Get(ctx, workspaceID, messageID)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("scan error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		rows := sqlmock.NewRows([]string{
			"id", "contact_email", "broadcast_id", "template_id", "template_version",
		}).AddRow(
			message.ID,
			message.ContactEmail,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
		) // Incomplete row to cause scan error

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE id = \$1`).
			WithArgs(messageID).
			WillReturnRows(rows)

		result, err := repo.Get(ctx, workspaceID, messageID)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to get message history")
	})
}

func TestMessageHistoryRepository_GetByContact(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	contactEmail := "contact@example.com"
	message := createSampleMessageHistory()
	limit := 10
	offset := 0

	// Convert the MessageData to JSON for proper DB response mocking
	messageDataJSON, _ := json.Marshal(message.MessageData)

	t.Run("successful retrieval", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Set up count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_email = \$1`).
			WithArgs(contactEmail).
			WillReturnRows(countRows)

		// Set up data query
		dataRows := sqlmock.NewRows([]string{
			"id", "contact_email", "broadcast_id", "template_id", "template_version",
			"channel", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}).AddRow(
			message.ID,
			message.ContactEmail,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			messageDataJSON, // Use the actual JSON bytes
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

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_email = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactEmail, limit, offset).
			WillReturnRows(dataRows)

		results, count, err := repo.GetByContact(ctx, workspaceID, contactEmail, limit, offset)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 1, count)
		require.Len(t, results, 1)
		assert.Equal(t, message.ID, results[0].ID)
		assert.Equal(t, message.ContactEmail, results[0].ContactEmail)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		results, count, err := repo.GetByContact(ctx, workspaceID, contactEmail, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("count query error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_email = \$1`).
			WithArgs(contactEmail).
			WillReturnError(errors.New("count error"))

		results, count, err := repo.GetByContact(ctx, workspaceID, contactEmail, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to count message history")
	})

	t.Run("data query error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Set up count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_email = \$1`).
			WithArgs(contactEmail).
			WillReturnRows(countRows)

		// But data query fails
		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_email = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactEmail, limit, offset).
			WillReturnError(errors.New("query error"))

		results, count, err := repo.GetByContact(ctx, workspaceID, contactEmail, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to query message history")
	})

	t.Run("scan error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Set up count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_email = \$1`).
			WithArgs(contactEmail).
			WillReturnRows(countRows)

		// Return incomplete row to cause scan error
		dataRows := sqlmock.NewRows([]string{"id", "contact_email"}).
			AddRow("msg-123", "contact-123")

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_email = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactEmail, limit, offset).
			WillReturnRows(dataRows)

		results, count, err := repo.GetByContact(ctx, workspaceID, contactEmail, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to scan message history")
	})

	t.Run("default limit and offset", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Set up count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_email = \$1`).
			WithArgs(contactEmail).
			WillReturnRows(countRows)

		// Should use default limit of 50 and offset of 0
		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_email = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactEmail, 50, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "contact_email", "broadcast_id", "template_id", "template_version",
				"channel", "message_data", "sent_at", "delivered_at",
				"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
				"unsubscribed_at", "created_at", "updated_at",
			}).AddRow(
				message.ID,
				message.ContactEmail,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				messageDataJSON, // Use the actual JSON bytes
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
			))

		// Call with negative limit and offset
		results, count, err := repo.GetByContact(ctx, workspaceID, contactEmail, -5, -10)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 1, count)
		require.Len(t, results, 1)
	})
}

func TestMessageHistoryRepository_GetByBroadcast(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	message := createSampleMessageHistory()
	limit := 10
	offset := 0

	// Convert the MessageData to JSON for proper DB response mocking
	messageDataJSON, _ := json.Marshal(message.MessageData)

	t.Run("successful retrieval", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Set up count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(countRows)

		// Set up data query
		dataRows := sqlmock.NewRows([]string{
			"id", "contact_email", "broadcast_id", "template_id", "template_version",
			"channel", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}).AddRow(
			message.ID,
			message.ContactEmail,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			messageDataJSON, // Use the actual JSON bytes
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

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(broadcastID, limit, offset).
			WillReturnRows(dataRows)

		results, count, err := repo.GetByBroadcast(ctx, workspaceID, broadcastID, limit, offset)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 1, count)
		require.Len(t, results, 1)
		assert.Equal(t, message.ID, results[0].ID)
		assert.Equal(t, *message.BroadcastID, *results[0].BroadcastID)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		results, count, err := repo.GetByBroadcast(ctx, workspaceID, broadcastID, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("count query error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnError(errors.New("count error"))

		results, count, err := repo.GetByBroadcast(ctx, workspaceID, broadcastID, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to count message history")
	})

	t.Run("data query error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Set up count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(countRows)

		// But data query fails
		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(broadcastID, limit, offset).
			WillReturnError(errors.New("query error"))

		results, count, err := repo.GetByBroadcast(ctx, workspaceID, broadcastID, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to query message history")
	})
}

func TestMessageHistoryRepository_SetClicked(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "msg-123"
	timestamp := time.Now()

	t.Run("successful click update", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Expect the clicked_at update query
		mock.ExpectExec(`UPDATE message_history SET clicked_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND clicked_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect the opened_at update query
		mock.ExpectExec(`UPDATE message_history SET opened_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND opened_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SetClicked(ctx, workspaceID, messageID, timestamp)
		require.NoError(t, err)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		err := repo.SetClicked(ctx, workspaceID, messageID, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("clicked update error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// First query fails
		mock.ExpectExec(`UPDATE message_history SET clicked_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND clicked_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnError(errors.New("execution error"))

		err := repo.SetClicked(ctx, workspaceID, messageID, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to set clicked")
	})

	t.Run("opened update error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// First query succeeds
		mock.ExpectExec(`UPDATE message_history SET clicked_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND clicked_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Second query fails
		mock.ExpectExec(`UPDATE message_history SET opened_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND opened_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnError(errors.New("execution error"))

		err := repo.SetClicked(ctx, workspaceID, messageID, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to set opened")
	})
}

func TestMessageHistoryRepository_SetOpened(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "msg-123"
	timestamp := time.Now()

	t.Run("successful open update", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Expect the opened_at update query
		mock.ExpectExec(`UPDATE message_history SET opened_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND opened_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SetOpened(ctx, workspaceID, messageID, timestamp)
		require.NoError(t, err)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		err := repo.SetOpened(ctx, workspaceID, messageID, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("opened update error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		// Query fails
		mock.ExpectExec(`UPDATE message_history SET opened_at = \$1, updated_at = NOW\(\) WHERE id = \$2 AND opened_at IS NULL`).
			WithArgs(timestamp, messageID).
			WillReturnError(errors.New("execution error"))

		err := repo.SetOpened(ctx, workspaceID, messageID, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to set opened")
	})
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper function to add a message history row to mock rows
func addRowFromMessage(rows *sqlmock.Rows, message *domain.MessageHistory) {
	// Convert nullable fields to sql.Null* types
	var broadcastID sql.NullString
	if message.BroadcastID != nil {
		broadcastID = sql.NullString{String: *message.BroadcastID, Valid: true}
	}

	var errorMsg sql.NullString
	if message.Error != nil {
		errorMsg = sql.NullString{String: *message.Error, Valid: true}
	}

	// Convert time fields to sql.NullTime
	var deliveredAt, failedAt, openedAt, clickedAt, bouncedAt, complainedAt, unsubscribedAt sql.NullTime

	if message.DeliveredAt != nil {
		deliveredAt = sql.NullTime{Time: *message.DeliveredAt, Valid: true}
	}
	if message.FailedAt != nil {
		failedAt = sql.NullTime{Time: *message.FailedAt, Valid: true}
	}
	if message.OpenedAt != nil {
		openedAt = sql.NullTime{Time: *message.OpenedAt, Valid: true}
	}
	if message.ClickedAt != nil {
		clickedAt = sql.NullTime{Time: *message.ClickedAt, Valid: true}
	}
	if message.BouncedAt != nil {
		bouncedAt = sql.NullTime{Time: *message.BouncedAt, Valid: true}
	}
	if message.ComplainedAt != nil {
		complainedAt = sql.NullTime{Time: *message.ComplainedAt, Valid: true}
	}
	if message.UnsubscribedAt != nil {
		unsubscribedAt = sql.NullTime{Time: *message.UnsubscribedAt, Valid: true}
	}

	rows.AddRow(
		message.ID,
		message.ContactEmail,
		broadcastID,
		message.TemplateID,
		message.TemplateVersion,
		message.Channel,
		errorMsg,
		message.MessageData,
		message.SentAt,
		deliveredAt,
		failedAt,
		openedAt,
		clickedAt,
		bouncedAt,
		complainedAt,
		unsubscribedAt,
		message.CreatedAt,
		message.UpdatedAt,
	)
}

func TestMessageHistoryRepository_GetBroadcastStats(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"

	t.Run("successful retrieval", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		rows := sqlmock.NewRows([]string{
			"total_sent", "total_delivered", "total_failed", "total_opened",
			"total_clicked", "total_bounced", "total_complained", "total_unsubscribed",
		}).AddRow(10, 8, 2, 5, 3, 1, 0, 1)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(rows)

		stats, err := repo.GetBroadcastStats(ctx, workspaceID, broadcastID)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 10, stats.TotalSent)
		assert.Equal(t, 8, stats.TotalDelivered)
		assert.Equal(t, 2, stats.TotalFailed)
		assert.Equal(t, 5, stats.TotalOpened)
		assert.Equal(t, 3, stats.TotalClicked)
		assert.Equal(t, 1, stats.TotalBounced)
		assert.Equal(t, 0, stats.TotalComplained)
		assert.Equal(t, 1, stats.TotalUnsubscribed)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(nil, errors.New("connection error"))

		stats, err := repo.GetBroadcastStats(ctx, workspaceID, broadcastID)
		require.Error(t, err)
		require.Nil(t, stats)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("sql error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnError(errors.New("sql error"))

		stats, err := repo.GetBroadcastStats(ctx, workspaceID, broadcastID)
		require.Error(t, err)
		require.Nil(t, stats)
		require.Contains(t, err.Error(), "failed to get broadcast stats")
	})

	t.Run("no rows", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnError(sql.ErrNoRows)

		stats, err := repo.GetBroadcastStats(ctx, workspaceID, broadcastID)
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 0, stats.TotalSent)
		assert.Equal(t, 0, stats.TotalDelivered)
		assert.Equal(t, 0, stats.TotalFailed)
		assert.Equal(t, 0, stats.TotalOpened)
		assert.Equal(t, 0, stats.TotalClicked)
		assert.Equal(t, 0, stats.TotalBounced)
		assert.Equal(t, 0, stats.TotalComplained)
		assert.Equal(t, 0, stats.TotalUnsubscribed)
	})

	t.Run("null values", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		// Create mock rows with some NULL values
		rows := sqlmock.NewRows([]string{
			"total_sent", "total_delivered", "total_failed", "total_opened",
			"total_clicked", "total_bounced", "total_complained", "total_unsubscribed",
		}).AddRow(10, nil, 2, nil, 3, nil, nil, 1)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1`).
			WithArgs(broadcastID).
			WillReturnRows(rows)

		stats, err := repo.GetBroadcastStats(ctx, workspaceID, broadcastID)
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 10, stats.TotalSent)
		assert.Equal(t, 0, stats.TotalDelivered) // Should be 0 for NULL
		assert.Equal(t, 2, stats.TotalFailed)
		assert.Equal(t, 0, stats.TotalOpened) // Should be 0 for NULL
		assert.Equal(t, 3, stats.TotalClicked)
		assert.Equal(t, 0, stats.TotalBounced)    // Should be 0 for NULL
		assert.Equal(t, 0, stats.TotalComplained) // Should be 0 for NULL
		assert.Equal(t, 1, stats.TotalUnsubscribed)
	})
}

func TestMessageHistoryRepository_GetBroadcastVariationStats(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	variationID := "variation-123"

	t.Run("successful retrieval", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		rows := sqlmock.NewRows([]string{
			"total_sent", "total_delivered", "total_failed", "total_opened",
			"total_clicked", "total_bounced", "total_complained", "total_unsubscribed",
		}).AddRow(10, 8, 2, 5, 3, 1, 0, 1)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 AND message_data->>'variation_id' = \$2`).
			WithArgs(broadcastID, variationID).
			WillReturnRows(rows)

		stats, err := repo.GetBroadcastVariationStats(ctx, workspaceID, broadcastID, variationID)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 10, stats.TotalSent)
		assert.Equal(t, 8, stats.TotalDelivered)
		assert.Equal(t, 2, stats.TotalFailed)
		assert.Equal(t, 5, stats.TotalOpened)
		assert.Equal(t, 3, stats.TotalClicked)
		assert.Equal(t, 1, stats.TotalBounced)
		assert.Equal(t, 0, stats.TotalComplained)
		assert.Equal(t, 1, stats.TotalUnsubscribed)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(nil, errors.New("connection error"))

		stats, err := repo.GetBroadcastVariationStats(ctx, workspaceID, broadcastID, variationID)
		require.Error(t, err)
		require.Nil(t, stats)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("sql error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 AND message_data->>'variation_id' = \$2`).
			WithArgs(broadcastID, variationID).
			WillReturnError(errors.New("sql error"))

		stats, err := repo.GetBroadcastVariationStats(ctx, workspaceID, broadcastID, variationID)
		require.Error(t, err)
		require.Nil(t, stats)
		require.Contains(t, err.Error(), "failed to get broadcast variation stats")
	})

	t.Run("no rows", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 AND message_data->>'variation_id' = \$2`).
			WithArgs(broadcastID, variationID).
			WillReturnError(sql.ErrNoRows)

		stats, err := repo.GetBroadcastVariationStats(ctx, workspaceID, broadcastID, variationID)
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 0, stats.TotalSent)
		assert.Equal(t, 0, stats.TotalDelivered)
		assert.Equal(t, 0, stats.TotalFailed)
		assert.Equal(t, 0, stats.TotalOpened)
		assert.Equal(t, 0, stats.TotalClicked)
		assert.Equal(t, 0, stats.TotalBounced)
		assert.Equal(t, 0, stats.TotalComplained)
		assert.Equal(t, 0, stats.TotalUnsubscribed)
	})

	t.Run("null values", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		// Create mock rows with some NULL values
		rows := sqlmock.NewRows([]string{
			"total_sent", "total_delivered", "total_failed", "total_opened",
			"total_clicked", "total_bounced", "total_complained", "total_unsubscribed",
		}).AddRow(10, nil, 2, nil, 3, nil, nil, 1)

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE broadcast_id = \$1 AND message_data->>'variation_id' = \$2`).
			WithArgs(broadcastID, variationID).
			WillReturnRows(rows)

		stats, err := repo.GetBroadcastVariationStats(ctx, workspaceID, broadcastID, variationID)
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 10, stats.TotalSent)
		assert.Equal(t, 0, stats.TotalDelivered) // Should be 0 for NULL
		assert.Equal(t, 2, stats.TotalFailed)
		assert.Equal(t, 0, stats.TotalOpened) // Should be 0 for NULL
		assert.Equal(t, 3, stats.TotalClicked)
		assert.Equal(t, 0, stats.TotalBounced)    // Should be 0 for NULL
		assert.Equal(t, 0, stats.TotalComplained) // Should be 0 for NULL
		assert.Equal(t, 1, stats.TotalUnsubscribed)
	})
}

func TestMessageHistoryRepository_SetStatusesIfNotSet(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	now := time.Now().UTC()

	// Create a batch of status updates
	updates := []domain.MessageEventUpdate{
		{
			ID:        "msg-123",
			Event:     domain.MessageEventDelivered,
			Timestamp: now,
		},
		{
			ID:        "msg-456",
			Event:     domain.MessageEventDelivered,
			Timestamp: now,
		},
		{
			ID:        "msg-789",
			Event:     domain.MessageEventBounced,
			Timestamp: now,
		},
	}

	t.Run("successful batch update - multiple statuses", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		// Expect batch query for delivered status updates (2 messages)
		mock.ExpectExec(`UPDATE message_history SET delivered_at = updates\.timestamp, updated_at = \$1::TIMESTAMP WITH TIME ZONE FROM \(VALUES \(\$2, \$3::TIMESTAMP WITH TIME ZONE\), \(\$4, \$5::TIMESTAMP WITH TIME ZONE\)\) AS updates\(id, timestamp\) WHERE message_history\.id = updates\.id AND delivered_at IS NULL`).
			WithArgs(
				sqlmock.AnyArg(), // updated_at timestamp
				"msg-123",
				now,
				"msg-456",
				now,
			).
			WillReturnResult(sqlmock.NewResult(0, 2))

		// Expect batch query for bounced status updates (1 message)
		mock.ExpectExec(`UPDATE message_history SET bounced_at = updates\.timestamp, updated_at = \$1::TIMESTAMP WITH TIME ZONE FROM \(VALUES \(\$2, \$3::TIMESTAMP WITH TIME ZONE\)\) AS updates\(id, timestamp\) WHERE message_history\.id = updates\.id AND bounced_at IS NULL`).
			WithArgs(
				sqlmock.AnyArg(), // updated_at timestamp
				"msg-789",
				now,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetStatusesIfNotSet(ctx, workspaceID, updates)
		require.NoError(t, err)
	})

	t.Run("successful batch update - single status", func(t *testing.T) {
		// Only one status type
		singleStatusUpdates := []domain.MessageEventUpdate{
			{
				ID:        "msg-123",
				Event:     domain.MessageEventOpened,
				Timestamp: now,
			},
			{
				ID:        "msg-456",
				Event:     domain.MessageEventOpened,
				Timestamp: now.Add(1 * time.Second),
			},
		}

		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		// Expect single batch query for opened status updates (2 messages)
		mock.ExpectExec(`UPDATE message_history SET opened_at = updates\.timestamp, updated_at = \$1::TIMESTAMP WITH TIME ZONE FROM \(VALUES \(\$2, \$3::TIMESTAMP WITH TIME ZONE\), \(\$4, \$5::TIMESTAMP WITH TIME ZONE\)\) AS updates\(id, timestamp\) WHERE message_history\.id = updates\.id AND opened_at IS NULL`).
			WithArgs(
				sqlmock.AnyArg(), // updated_at timestamp
				"msg-123",
				now,
				"msg-456",
				now.Add(1*time.Second),
			).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := repo.SetStatusesIfNotSet(ctx, workspaceID, singleStatusUpdates)
		require.NoError(t, err)
	})

	t.Run("empty updates", func(t *testing.T) {
		// No database calls should be made when the updates slice is empty
		err := repo.SetStatusesIfNotSet(ctx, workspaceID, []domain.MessageEventUpdate{})
		require.NoError(t, err)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(nil, errors.New("connection error"))

		err := repo.SetStatusesIfNotSet(ctx, workspaceID, updates)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("invalid status", func(t *testing.T) {
		invalidUpdates := []domain.MessageEventUpdate{
			{
				ID:        "msg-123",
				Event:     domain.MessageEvent("invalid"),
				Timestamp: now,
			},
		}

		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		err := repo.SetStatusesIfNotSet(ctx, workspaceID, invalidUpdates)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid status")
	})

	t.Run("database execution error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		mock.ExpectExec(`UPDATE message_history SET delivered_at = updates\.timestamp, updated_at = \$1::TIMESTAMP WITH TIME ZONE FROM \(VALUES \(\$2, \$3::TIMESTAMP WITH TIME ZONE\), \(\$4, \$5::TIMESTAMP WITH TIME ZONE\)\) AS updates\(id, timestamp\) WHERE message_history\.id = updates\.id AND delivered_at IS NULL`).
			WithArgs(
				sqlmock.AnyArg(),
				"msg-123",
				now,
				"msg-456",
				now,
			).
			WillReturnError(errors.New("database error"))

		// Only include the delivered status updates to trigger the first query error
		deliveredUpdates := []domain.MessageEventUpdate{
			{
				ID:        "msg-123",
				Event:     domain.MessageEventDelivered,
				Timestamp: now,
			},
			{
				ID:        "msg-456",
				Event:     domain.MessageEventDelivered,
				Timestamp: now,
			},
		}

		err := repo.SetStatusesIfNotSet(ctx, workspaceID, deliveredUpdates)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to batch update message statuses for status")
	})

	t.Run("integration with single status method", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(gomock.Any(), workspaceID).
			Return(db, nil)

		// Expect the batch version to be called with a single update
		mock.ExpectExec(`UPDATE message_history SET delivered_at = updates\.timestamp, updated_at = \$1::TIMESTAMP WITH TIME ZONE FROM \(VALUES \(\$2, \$3::TIMESTAMP WITH TIME ZONE\)\) AS updates\(id, timestamp\) WHERE message_history\.id = updates\.id AND delivered_at IS NULL`).
			WithArgs(
				sqlmock.AnyArg(),
				"msg-123",
				now,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Call the single status method
		err := repo.SetStatusesIfNotSet(ctx, workspaceID, []domain.MessageEventUpdate{
			{
				ID:        "msg-123",
				Event:     domain.MessageEventDelivered,
				Timestamp: now,
			},
		})
		require.NoError(t, err)
	})
}
