package domain

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageStatus(t *testing.T) {
	// Test all defined message status constants
	assert.Equal(t, MessageStatus("sent"), MessageStatusSent)
	assert.Equal(t, MessageStatus("delivered"), MessageStatusDelivered)
	assert.Equal(t, MessageStatus("failed"), MessageStatusFailed)
	assert.Equal(t, MessageStatus("opened"), MessageStatusOpened)
	assert.Equal(t, MessageStatus("clicked"), MessageStatusClicked)
	assert.Equal(t, MessageStatus("bounced"), MessageStatusBounced)
	assert.Equal(t, MessageStatus("complained"), MessageStatusComplained)
	assert.Equal(t, MessageStatus("unsubscribed"), MessageStatusUnsubscribed)
}

func TestMessageData_Value(t *testing.T) {
	tests := []struct {
		name     string
		data     MessageData
		expected string
		wantErr  bool
	}{
		{
			name: "empty data",
			data: MessageData{
				Data:     map[string]interface{}{},
				Metadata: nil,
			},
			expected: `{"data":{}}`,
			wantErr:  false,
		},
		{
			name: "with data only",
			data: MessageData{
				Data: map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				},
				Metadata: nil,
			},
			expected: `{"data":{"email":"john@example.com","name":"John Doe"}}`,
			wantErr:  false,
		},
		{
			name: "with data and metadata",
			data: MessageData{
				Data: map[string]interface{}{
					"name": "John Doe",
				},
				Metadata: map[string]interface{}{
					"source": "signup",
					"tags":   []string{"welcome", "new-user"},
				},
			},
			expected: `{"data":{"name":"John Doe"},"metadata":{"source":"signup","tags":["welcome","new-user"]}}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.data.Value()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Convert result to string for comparison
			gotBytes, ok := got.([]byte)
			require.True(t, ok)

			// Since JSON map keys can be in any order, we need to unmarshal both to compare
			var expected, actual map[string]interface{}
			err = json.Unmarshal([]byte(tt.expected), &expected)
			require.NoError(t, err)

			err = json.Unmarshal(gotBytes, &actual)
			require.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	}
}

func TestMessageData_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    MessageData
		wantErr bool
	}{
		{
			name:  "nil input",
			input: nil,
			want:  MessageData{},
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
		{
			name:  "valid json - empty",
			input: []byte(`{"data":{}}`),
			want: MessageData{
				Data: map[string]interface{}{},
			},
		},
		{
			name:  "valid json - with data",
			input: []byte(`{"data":{"name":"John Doe","age":30},"metadata":{"source":"api"}}`),
			want: MessageData{
				Data: map[string]interface{}{
					"name": "John Doe",
					"age":  float64(30),
				},
				Metadata: map[string]interface{}{
					"source": "api",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var md MessageData
			err := md.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.input == nil {
				assert.Empty(t, md.Data)
				assert.Empty(t, md.Metadata)
				return
			}

			assert.Equal(t, tt.want.Data, md.Data)
			assert.Equal(t, tt.want.Metadata, md.Metadata)
		})
	}
}

func TestMessageHistory(t *testing.T) {
	now := time.Now()

	// Create a test MessageHistory instance
	broadcastID := "broadcast123"
	message := MessageHistory{
		ID:              "msg123",
		ContactID:       "contact456",
		BroadcastID:     &broadcastID,
		TemplateID:      "template789",
		TemplateVersion: 1,
		Channel:         "email",
		Status:          MessageStatusSent,
		MessageData: MessageData{
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

	// Test basic field values
	assert.Equal(t, "msg123", message.ID)
	assert.Equal(t, "contact456", message.ContactID)
	assert.Equal(t, "broadcast123", *message.BroadcastID)
	assert.Equal(t, "template789", message.TemplateID)
	assert.Equal(t, 1, message.TemplateVersion)
	assert.Equal(t, "email", message.Channel)
	assert.Equal(t, MessageStatusSent, message.Status)
	assert.Equal(t, now, message.SentAt)
	assert.Equal(t, now, message.CreatedAt)
	assert.Equal(t, now, message.UpdatedAt)

	// Test message data
	assert.Equal(t, "Welcome!", message.MessageData.Data["subject"])
	assert.Equal(t, "John", message.MessageData.Data["name"])
	assert.Equal(t, "onboarding", message.MessageData.Metadata["campaign"])

	// Test optional timestamps are nil
	assert.Nil(t, message.DeliveredAt)
	assert.Nil(t, message.FailedAt)
	assert.Nil(t, message.OpenedAt)
	assert.Nil(t, message.ClickedAt)
	assert.Nil(t, message.BouncedAt)
	assert.Nil(t, message.ComplainedAt)
	assert.Nil(t, message.UnsubscribedAt)

	// Set a timestamp
	deliveredAt := now.Add(time.Hour)
	message.DeliveredAt = &deliveredAt
	assert.Equal(t, deliveredAt, *message.DeliveredAt)
}

// Mock implementation of MessageHistoryRepository for future tests
type mockMessageHistoryRepository struct {
	messages map[string]*MessageHistory
}

func newMockMessageHistoryRepository() *mockMessageHistoryRepository {
	return &mockMessageHistoryRepository{
		messages: make(map[string]*MessageHistory),
	}
}

func (m *mockMessageHistoryRepository) Create(ctx context.Context, workspace string, message *MessageHistory) error {
	key := workspace + "-" + message.ID
	m.messages[key] = message
	return nil
}

func (m *mockMessageHistoryRepository) Update(ctx context.Context, workspace string, message *MessageHistory) error {
	key := workspace + "-" + message.ID
	if _, exists := m.messages[key]; !exists {
		return &ErrNotFound{Entity: "message", ID: message.ID}
	}
	m.messages[key] = message
	return nil
}

func (m *mockMessageHistoryRepository) Get(ctx context.Context, workspace, id string) (*MessageHistory, error) {
	key := workspace + "-" + id
	message, exists := m.messages[key]
	if !exists {
		return nil, &ErrNotFound{Entity: "message", ID: id}
	}
	return message, nil
}

func (m *mockMessageHistoryRepository) GetByContact(ctx context.Context, workspace, contactID string, limit, offset int) ([]*MessageHistory, int, error) {
	var results []*MessageHistory
	for key, msg := range m.messages {
		if key[:len(workspace)] == workspace && msg.ContactID == contactID {
			results = append(results, msg)
		}
	}
	return results, len(results), nil
}

func (m *mockMessageHistoryRepository) GetByBroadcast(ctx context.Context, workspace, broadcastID string, limit, offset int) ([]*MessageHistory, int, error) {
	var results []*MessageHistory
	for key, msg := range m.messages {
		if key[:len(workspace)] == workspace && msg.BroadcastID != nil && *msg.BroadcastID == broadcastID {
			results = append(results, msg)
		}
	}
	return results, len(results), nil
}

func (m *mockMessageHistoryRepository) UpdateStatus(ctx context.Context, workspace, id string, status MessageStatus, timestamp time.Time) error {
	key := workspace + "-" + id
	message, exists := m.messages[key]
	if !exists {
		return &ErrNotFound{Entity: "message", ID: id}
	}

	message.Status = status

	switch status {
	case MessageStatusDelivered:
		message.DeliveredAt = &timestamp
	case MessageStatusFailed:
		message.FailedAt = &timestamp
	case MessageStatusOpened:
		message.OpenedAt = &timestamp
	case MessageStatusClicked:
		message.ClickedAt = &timestamp
	case MessageStatusBounced:
		message.BouncedAt = &timestamp
	case MessageStatusComplained:
		message.ComplainedAt = &timestamp
	case MessageStatusUnsubscribed:
		message.UnsubscribedAt = &timestamp
	}

	message.UpdatedAt = timestamp
	m.messages[key] = message

	return nil
}

func (m *mockMessageHistoryRepository) SetClicked(ctx context.Context, workspace, id string, timestamp time.Time) error {
	key := workspace + "-" + id
	message, exists := m.messages[key]
	if !exists {
		return &ErrNotFound{Entity: "message", ID: id}
	}

	// Only update if clicked_at is nil
	if message.ClickedAt == nil {
		message.ClickedAt = &timestamp
		message.Status = MessageStatusClicked
	}

	// Always ensure opened_at is set if not already
	if message.OpenedAt == nil {
		message.OpenedAt = &timestamp
	}

	message.UpdatedAt = timestamp
	m.messages[key] = message

	return nil
}

func (m *mockMessageHistoryRepository) SetOpened(ctx context.Context, workspace, id string, timestamp time.Time) error {
	key := workspace + "-" + id
	message, exists := m.messages[key]
	if !exists {
		return &ErrNotFound{Entity: "message", ID: id}
	}

	// Only update if opened_at is nil
	if message.OpenedAt == nil {
		message.OpenedAt = &timestamp

		// Only update status if it's not already clicked
		if message.Status != MessageStatusClicked {
			message.Status = MessageStatusOpened
		}
	}

	message.UpdatedAt = timestamp
	m.messages[key] = message

	return nil
}

// Test for MessageHistoryRepository interface compliance
func TestMockMessageHistoryRepository_ImplementsInterface(t *testing.T) {
	var _ MessageHistoryRepository = (*mockMessageHistoryRepository)(nil)
}

// Example tests using the mock repository
func TestMockMessageHistoryRepository(t *testing.T) {
	repo := newMockMessageHistoryRepository()
	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()

	// Test message
	msg := &MessageHistory{
		ID:              "msg1",
		ContactID:       "contact1",
		TemplateID:      "template1",
		TemplateVersion: 1,
		Channel:         "email",
		Status:          MessageStatusSent,
		SentAt:          now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Test Create
	t.Run("Create", func(t *testing.T) {
		err := repo.Create(ctx, workspace, msg)
		assert.NoError(t, err)

		// Verify it was stored
		stored, err := repo.Get(ctx, workspace, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, stored.ID)
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		msg.Status = MessageStatusDelivered
		deliveredAt := now.Add(time.Minute)
		msg.DeliveredAt = &deliveredAt

		err := repo.Update(ctx, workspace, msg)
		assert.NoError(t, err)

		// Verify it was updated
		stored, err := repo.Get(ctx, workspace, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, MessageStatusDelivered, stored.Status)
		assert.Equal(t, deliveredAt, *stored.DeliveredAt)
	})

	// Test Get non-existent
	t.Run("Get non-existent", func(t *testing.T) {
		_, err := repo.Get(ctx, workspace, "non-existent")
		assert.Error(t, err)
	})

	// Test UpdateStatus
	t.Run("UpdateStatus", func(t *testing.T) {
		openedAt := now.Add(time.Hour)
		err := repo.UpdateStatus(ctx, workspace, msg.ID, MessageStatusOpened, openedAt)
		assert.NoError(t, err)

		// Verify status and timestamp were updated
		stored, err := repo.Get(ctx, workspace, msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, MessageStatusOpened, stored.Status)
		assert.Equal(t, openedAt, *stored.OpenedAt)
		assert.Equal(t, openedAt, stored.UpdatedAt) // UpdatedAt should also be set
	})
}
