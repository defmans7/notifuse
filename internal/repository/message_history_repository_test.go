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
		ContactID:       "contact-123",
		BroadcastID:     &broadcastID,
		TemplateID:      "template-123",
		TemplateVersion: 1,
		Channel:         "email",
		Status:          domain.MessageStatusSent,
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
				message.ContactID,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				message.Status,
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
				message.ContactID,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				message.Status,
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
				message.ContactID,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				message.Status,
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
				message.ContactID,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				message.Status,
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
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}).AddRow(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
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
		assert.Equal(t, message.ContactID, result.ContactID)
		assert.Equal(t, *message.BroadcastID, *result.BroadcastID)
		assert.Equal(t, message.Status, result.Status)
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
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
		}).AddRow(
			message.ID,
			message.ContactID,
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
	contactID := "contact-123"
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
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		// Set up data query
		dataRows := sqlmock.NewRows([]string{
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}).AddRow(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
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

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, limit, offset).
			WillReturnRows(dataRows)

		results, count, err := repo.GetByContact(ctx, workspaceID, contactID, limit, offset)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 1, count)
		require.Len(t, results, 1)
		assert.Equal(t, message.ID, results[0].ID)
		assert.Equal(t, message.ContactID, results[0].ContactID)
	})

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		results, count, err := repo.GetByContact(ctx, workspaceID, contactID, limit, offset)
		require.Error(t, err)
		require.Nil(t, results)
		require.Zero(t, count)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("count query error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnError(errors.New("count error"))

		results, count, err := repo.GetByContact(ctx, workspaceID, contactID, limit, offset)
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
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		// But data query fails
		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, limit, offset).
			WillReturnError(errors.New("query error"))

		results, count, err := repo.GetByContact(ctx, workspaceID, contactID, limit, offset)
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
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		// Return incomplete row to cause scan error
		dataRows := sqlmock.NewRows([]string{"id", "contact_id"}).
			AddRow("msg-123", "contact-123")

		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, limit, offset).
			WillReturnRows(dataRows)

		results, count, err := repo.GetByContact(ctx, workspaceID, contactID, limit, offset)
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
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM message_history WHERE contact_id = \$1`).
			WithArgs(contactID).
			WillReturnRows(countRows)

		// Should use default limit of 50 and offset of 0
		mock.ExpectQuery(`SELECT .* FROM message_history WHERE contact_id = \$1 ORDER BY sent_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(contactID, 50, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "contact_id", "broadcast_id", "template_id", "template_version",
				"channel", "status", "message_data", "sent_at", "delivered_at",
				"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
				"unsubscribed_at", "created_at", "updated_at",
			}).AddRow(
				message.ID,
				message.ContactID,
				message.BroadcastID,
				message.TemplateID,
				message.TemplateVersion,
				message.Channel,
				message.Status,
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
		results, count, err := repo.GetByContact(ctx, workspaceID, contactID, -5, -10)
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
			"id", "contact_id", "broadcast_id", "template_id", "template_version",
			"channel", "status", "message_data", "sent_at", "delivered_at",
			"failed_at", "opened_at", "clicked_at", "bounced_at", "complained_at",
			"unsubscribed_at", "created_at", "updated_at",
		}).AddRow(
			message.ID,
			message.ContactID,
			message.BroadcastID,
			message.TemplateID,
			message.TemplateVersion,
			message.Channel,
			message.Status,
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

func TestMessageHistoryRepository_UpdateStatus(t *testing.T) {
	mockWorkspaceRepo, repo, mock, db, cleanup := setupMessageHistoryTest(t)
	defer cleanup()

	ctx := context.Background()
	workspaceID := "workspace-123"
	messageID := "msg-123"
	timestamp := time.Now()

	testCases := []struct {
		name          string
		status        domain.MessageStatus
		expectedField string
	}{
		{
			name:          "delivered status",
			status:        domain.MessageStatusDelivered,
			expectedField: "delivered_at",
		},
		{
			name:          "failed status",
			status:        domain.MessageStatusFailed,
			expectedField: "failed_at",
		},
		{
			name:          "opened status",
			status:        domain.MessageStatusOpened,
			expectedField: "opened_at",
		},
		{
			name:          "clicked status",
			status:        domain.MessageStatusClicked,
			expectedField: "clicked_at",
		},
		{
			name:          "bounced status",
			status:        domain.MessageStatusBounced,
			expectedField: "bounced_at",
		},
		{
			name:          "complained status",
			status:        domain.MessageStatusComplained,
			expectedField: "complained_at",
		},
		{
			name:          "unsubscribed status",
			status:        domain.MessageStatusUnsubscribed,
			expectedField: "unsubscribed_at",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWorkspaceRepo.EXPECT().
				GetConnection(ctx, workspaceID).
				Return(db, nil)

			mock.ExpectExec(`UPDATE message_history SET status = \$1, `+tc.expectedField+` = \$2, updated_at = \$3 WHERE id = \$4`).
				WithArgs(tc.status, timestamp, sqlmock.AnyArg(), messageID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.UpdateStatus(ctx, workspaceID, messageID, tc.status, timestamp)
			require.NoError(t, err)
		})
	}

	t.Run("workspace connection error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(nil, errors.New("connection error"))

		err := repo.UpdateStatus(ctx, workspaceID, messageID, domain.MessageStatusDelivered, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("invalid status", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		invalidStatus := domain.MessageStatus("invalid")
		err := repo.UpdateStatus(ctx, workspaceID, messageID, invalidStatus, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid status")
	})

	t.Run("execution error", func(t *testing.T) {
		mockWorkspaceRepo.EXPECT().
			GetConnection(ctx, workspaceID).
			Return(db, nil)

		mock.ExpectExec(`UPDATE message_history SET status = \$1, delivered_at = \$2, updated_at = \$3 WHERE id = \$4`).
			WithArgs(domain.MessageStatusDelivered, timestamp, sqlmock.AnyArg(), messageID).
			WillReturnError(errors.New("execution error"))

		err := repo.UpdateStatus(ctx, workspaceID, messageID, domain.MessageStatusDelivered, timestamp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to update message status")
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
		mock.ExpectExec(`UPDATE message_history SET clicked_at = \$1, status = 'clicked', updated_at = NOW\(\) WHERE id = \$2 AND clicked_at IS NULL`).
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
		mock.ExpectExec(`UPDATE message_history SET clicked_at = \$1, status = 'clicked', updated_at = NOW\(\) WHERE id = \$2 AND clicked_at IS NULL`).
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
		mock.ExpectExec(`UPDATE message_history SET clicked_at = \$1, status = 'clicked', updated_at = NOW\(\) WHERE id = \$2 AND clicked_at IS NULL`).
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
