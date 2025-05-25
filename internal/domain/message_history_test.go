package domain

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		ContactEmail:    "contact456",
		BroadcastID:     &broadcastID,
		TemplateID:      "template789",
		TemplateVersion: 1,
		Channel:         "email",
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
	assert.Equal(t, "contact456", message.ContactEmail)
	assert.Equal(t, "broadcast123", *message.BroadcastID)
	assert.Equal(t, "template789", message.TemplateID)
	assert.Equal(t, int64(1), message.TemplateVersion)
	assert.Equal(t, "email", message.Channel)
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

func TestMessageListParams_FromQuery(t *testing.T) {
	tests := []struct {
		name      string
		queryData map[string][]string
		want      MessageListParams
		wantErr   bool
	}{
		{
			name:      "empty query",
			queryData: map[string][]string{},
			want: MessageListParams{
				Limit: 20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "basic string filters",
			queryData: map[string][]string{
				"cursor":        {"next_page"},
				"channel":       {"email"},
				"status":        {"delivered"},
				"contact_email": {"contact@example.com"},
				"broadcast_id":  {"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"},
				"template_id":   {"7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d"},
			},
			want: MessageListParams{
				Cursor:       "next_page",
				Channel:      "email",
				ContactEmail: "contact@example.com",
				BroadcastID:  "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
				TemplateID:   "7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d",
				Limit:        20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "with custom limit",
			queryData: map[string][]string{
				"limit": {"50"},
			},
			want: MessageListParams{
				Limit: 50,
			},
			wantErr: false,
		},
		{
			name: "with invalid limit",
			queryData: map[string][]string{
				"limit": {"not_a_number"},
			},
			wantErr: true,
		},
		{
			name: "with has_error true",
			queryData: map[string][]string{
				"has_error": {"true"},
			},
			want: MessageListParams{
				HasError: boolPtr(true),
				Limit:    20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "with has_error false",
			queryData: map[string][]string{
				"has_error": {"false"},
			},
			want: MessageListParams{
				HasError: boolPtr(false),
				Limit:    20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "with invalid has_error",
			queryData: map[string][]string{
				"has_error": {"not_a_boolean"},
			},
			wantErr: true,
		},
		{
			name: "with time parameters",
			queryData: map[string][]string{
				"sent_after":     {"2023-01-01T00:00:00Z"},
				"sent_before":    {"2023-12-31T23:59:59Z"},
				"updated_after":  {"2023-02-01T00:00:00Z"},
				"updated_before": {"2023-11-30T23:59:59Z"},
			},
			want: MessageListParams{
				SentAfter:     timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				SentBefore:    timePtr(time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)),
				UpdatedAfter:  timePtr(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedBefore: timePtr(time.Date(2023, 11, 30, 23, 59, 59, 0, time.UTC)),
				Limit:         20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "with invalid time format",
			queryData: map[string][]string{
				"sent_after": {"2023/01/01"}, // Invalid RFC3339 format
			},
			wantErr: true,
		},
		{
			name: "with invalid channel",
			queryData: map[string][]string{
				"channel": {"invalid_channel"}, // Not one of the allowed values
			},
			wantErr: true,
		},
		{
			name: "with invalid contact_email",
			queryData: map[string][]string{
				"contact_email": {"not-a-email"}, // Not a UUID format
			},
			wantErr: true,
		},
		{
			name: "with invalid broadcast_id",
			queryData: map[string][]string{
				"broadcast_id": {"not-a-uuid"}, // Not a UUID format
			},
			wantErr: true,
		},
		{
			name: "with invalid template_id",
			queryData: map[string][]string{
				"template_id": {"not-a-uuid"}, // Not a UUID format
			},
			wantErr: true,
		},
		{
			name: "with invalid time range (sent)",
			queryData: map[string][]string{
				"sent_after":  {"2023-12-31T23:59:59Z"},
				"sent_before": {"2023-01-01T00:00:00Z"}, // Before the after date
			},
			wantErr: true,
		},
		{
			name: "with invalid time range (updated)",
			queryData: map[string][]string{
				"updated_after":  {"2023-12-31T23:59:59Z"},
				"updated_before": {"2023-01-01T00:00:00Z"}, // Before the after date
			},
			wantErr: true,
		},
		{
			name: "with too large limit",
			queryData: map[string][]string{
				"limit": {"200"}, // Above the cap of 100
			},
			want: MessageListParams{
				Limit: 100, // Should be capped to 100
			},
			wantErr: false,
		},
		{
			name: "with negative limit",
			queryData: map[string][]string{
				"limit": {"-10"}, // Negative
			},
			wantErr: true,
		},
		{
			name: "with all parameters",
			queryData: map[string][]string{
				"cursor":         {"next_page"},
				"channel":        {"email"},
				"status":         {"delivered"},
				"contact_email":  {"contact@example.com"},
				"broadcast_id":   {"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"},
				"template_id":    {"7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d"},
				"has_error":      {"true"},
				"limit":          {"50"},
				"sent_after":     {"2023-01-01T00:00:00Z"},
				"sent_before":    {"2023-12-31T23:59:59Z"},
				"updated_after":  {"2023-02-01T00:00:00Z"},
				"updated_before": {"2023-11-30T23:59:59Z"},
			},
			want: MessageListParams{
				Cursor:        "next_page",
				Channel:       "email",
				ContactEmail:  "contact@example.com",
				BroadcastID:   "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
				TemplateID:    "7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d",
				HasError:      boolPtr(true),
				Limit:         50,
				SentAfter:     timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				SentBefore:    timePtr(time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)),
				UpdatedAfter:  timePtr(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedBefore: timePtr(time.Date(2023, 11, 30, 23, 59, 59, 0, time.UTC)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryValues := url.Values(tt.queryData)
			var params MessageListParams
			err := params.FromQuery(queryValues)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Cursor, params.Cursor)
			assert.Equal(t, tt.want.Channel, params.Channel)
			assert.Equal(t, tt.want.ContactEmail, params.ContactEmail)
			assert.Equal(t, tt.want.BroadcastID, params.BroadcastID)
			assert.Equal(t, tt.want.TemplateID, params.TemplateID)
			assert.Equal(t, tt.want.Limit, params.Limit)

			if tt.want.HasError != nil {
				assert.NotNil(t, params.HasError)
				assert.Equal(t, *tt.want.HasError, *params.HasError)
			} else {
				assert.Nil(t, params.HasError)
			}

			if tt.want.SentAfter != nil {
				assert.NotNil(t, params.SentAfter)
				assert.True(t, tt.want.SentAfter.Equal(*params.SentAfter))
			} else {
				assert.Nil(t, params.SentAfter)
			}

			if tt.want.SentBefore != nil {
				assert.NotNil(t, params.SentBefore)
				assert.True(t, tt.want.SentBefore.Equal(*params.SentBefore))
			} else {
				assert.Nil(t, params.SentBefore)
			}

			if tt.want.UpdatedAfter != nil {
				assert.NotNil(t, params.UpdatedAfter)
				assert.True(t, tt.want.UpdatedAfter.Equal(*params.UpdatedAfter))
			} else {
				assert.Nil(t, params.UpdatedAfter)
			}

			if tt.want.UpdatedBefore != nil {
				assert.NotNil(t, params.UpdatedBefore)
				assert.True(t, tt.want.UpdatedBefore.Equal(*params.UpdatedBefore))
			} else {
				assert.Nil(t, params.UpdatedBefore)
			}
		})
	}
}

func TestMessageListParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  MessageListParams
		want    MessageListParams
		wantErr bool
	}{
		{
			name:   "default values",
			params: MessageListParams{},
			want: MessageListParams{
				Limit: 20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "negative limit",
			params: MessageListParams{
				Limit: -10,
			},
			wantErr: true,
		},
		{
			name: "zero limit becomes default",
			params: MessageListParams{
				Limit: 0,
			},
			want: MessageListParams{
				Limit: 20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "limit too high gets capped",
			params: MessageListParams{
				Limit: 500,
			},
			want: MessageListParams{
				Limit: 100, // Capped to max
			},
			wantErr: false,
		},
		{
			name: "valid channel",
			params: MessageListParams{
				Channel: "email",
			},
			want: MessageListParams{
				Channel: "email",
				Limit:   20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "invalid channel",
			params: MessageListParams{
				Channel: "invalid_channel",
			},
			wantErr: true,
		},
		{
			name: "valid IDs",
			params: MessageListParams{
				ContactEmail: "contact@example.com",
				BroadcastID:  "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
				TemplateID:   "7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d",
			},
			want: MessageListParams{
				ContactEmail: "contact@example.com",
				BroadcastID:  "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
				TemplateID:   "7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d",
				Limit:        20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "invalid contact ID",
			params: MessageListParams{
				ContactEmail: "not-a-uuid",
			},
			wantErr: true,
		},
		{
			name: "invalid broadcast ID",
			params: MessageListParams{
				BroadcastID: "not-a-uuid",
			},
			wantErr: true,
		},
		{
			name: "invalid template ID",
			params: MessageListParams{
				TemplateID: "not-a-uuid",
			},
			wantErr: true,
		},
		{
			name: "valid time ranges",
			params: MessageListParams{
				SentAfter:     timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				SentBefore:    timePtr(time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)),
				UpdatedAfter:  timePtr(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedBefore: timePtr(time.Date(2023, 11, 30, 23, 59, 59, 0, time.UTC)),
			},
			want: MessageListParams{
				SentAfter:     timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				SentBefore:    timePtr(time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)),
				UpdatedAfter:  timePtr(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedBefore: timePtr(time.Date(2023, 11, 30, 23, 59, 59, 0, time.UTC)),
				Limit:         20, // Default limit
			},
			wantErr: false,
		},
		{
			name: "invalid sent time range",
			params: MessageListParams{
				SentAfter:  timePtr(time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)),
				SentBefore: timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)), // Before the after date
			},
			wantErr: true,
		},
		{
			name: "invalid updated time range",
			params: MessageListParams{
				UpdatedAfter:  timePtr(time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)),
				UpdatedBefore: timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)), // Before the after date
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the params to validate
			params := tt.params

			err := params.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check if params were modified as expected
			if tt.want.Limit != 0 {
				assert.Equal(t, tt.want.Limit, params.Limit)
			}
			assert.Equal(t, tt.want.Channel, params.Channel)
			assert.Equal(t, tt.want.ContactEmail, params.ContactEmail)
			assert.Equal(t, tt.want.BroadcastID, params.BroadcastID)
			assert.Equal(t, tt.want.TemplateID, params.TemplateID)

			if tt.want.SentAfter != nil {
				assert.NotNil(t, params.SentAfter)
				assert.True(t, tt.want.SentAfter.Equal(*params.SentAfter))
			}
			if tt.want.SentBefore != nil {
				assert.NotNil(t, params.SentBefore)
				assert.True(t, tt.want.SentBefore.Equal(*params.SentBefore))
			}
			if tt.want.UpdatedAfter != nil {
				assert.NotNil(t, params.UpdatedAfter)
				assert.True(t, tt.want.UpdatedAfter.Equal(*params.UpdatedAfter))
			}
			if tt.want.UpdatedBefore != nil {
				assert.NotNil(t, params.UpdatedBefore)
				assert.True(t, tt.want.UpdatedBefore.Equal(*params.UpdatedBefore))
			}
		})
	}
}

// Helper functions to create pointers
func boolPtr(b bool) *bool {
	return &b
}

func timePtr(t time.Time) *time.Time {
	return &t
}
