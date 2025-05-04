package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWebhookEvent(t *testing.T) {
	now := time.Now()
	event := WebhookEvent{
		ID:                "webhook123",
		Type:              EmailEventDelivered,
		EmailProviderKind: EmailProviderKindSES,
		IntegrationID:     "integration123",
		RecipientEmail:    "test@example.com",
		MessageID:         "message123",
		Timestamp:         now,
		RawPayload:        `{"event": "delivery"}`,
	}

	assert.Equal(t, "webhook123", event.ID)
	assert.Equal(t, EmailEventDelivered, event.Type)
	assert.Equal(t, EmailProviderKindSES, event.EmailProviderKind)
	assert.Equal(t, "integration123", event.IntegrationID)
	assert.Equal(t, "test@example.com", event.RecipientEmail)
	assert.Equal(t, "message123", event.MessageID)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, `{"event": "delivery"}`, event.RawPayload)
}

func TestNewWebhookEvent(t *testing.T) {
	now := time.Now()
	event := NewWebhookEvent(
		"webhook123",
		EmailEventDelivered,
		EmailProviderKindSES,
		"integration123",
		"test@example.com",
		"message123",
		now,
		`{"event": "delivery"}`,
	)

	assert.Equal(t, "webhook123", event.ID)
	assert.Equal(t, EmailEventDelivered, event.Type)
	assert.Equal(t, EmailProviderKindSES, event.EmailProviderKind)
	assert.Equal(t, "integration123", event.IntegrationID)
	assert.Equal(t, "test@example.com", event.RecipientEmail)
	assert.Equal(t, "message123", event.MessageID)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, `{"event": "delivery"}`, event.RawPayload)
	assert.Empty(t, event.TransactionalID)
	assert.Empty(t, event.BroadcastID)
}

func TestSetBounceInfo(t *testing.T) {
	event := NewWebhookEvent(
		"webhook123",
		EmailEventBounce,
		EmailProviderKindSES,
		"integration123",
		"test@example.com",
		"message123",
		time.Now(),
		`{"event": "bounce"}`,
	)

	event.SetBounceInfo("permanent", "hard_bounce", "5.1.1 User unknown")

	assert.Equal(t, "permanent", event.BounceType)
	assert.Equal(t, "hard_bounce", event.BounceCategory)
	assert.Equal(t, "5.1.1 User unknown", event.BounceDiagnostic)
}

func TestSetComplaintInfo(t *testing.T) {
	event := NewWebhookEvent(
		"webhook123",
		EmailEventComplaint,
		EmailProviderKindSES,
		"integration123",
		"test@example.com",
		"message123",
		time.Now(),
		`{"event": "complaint"}`,
	)

	event.SetComplaintInfo("abuse")

	assert.Equal(t, "abuse", event.ComplaintFeedbackType)
}

func TestSetTransactionalID(t *testing.T) {
	event := NewWebhookEvent(
		"webhook123",
		EmailEventDelivered,
		EmailProviderKindSES,
		"integration123",
		"test@example.com",
		"message123",
		time.Now(),
		`{"event": "delivery"}`,
	)

	event.SetTransactionalID("trans123")

	assert.Equal(t, "trans123", event.TransactionalID)
}

func TestSetBroadcastID(t *testing.T) {
	event := NewWebhookEvent(
		"webhook123",
		EmailEventDelivered,
		EmailProviderKindSES,
		"integration123",
		"test@example.com",
		"message123",
		time.Now(),
		`{"event": "delivery"}`,
	)

	event.SetBroadcastID("broadcast123")

	assert.Equal(t, "broadcast123", event.BroadcastID)
}

func TestErrWebhookEventNotFound_Error(t *testing.T) {
	err := &ErrWebhookEventNotFound{ID: "webhook123"}
	expectedMsg := "webhook event with ID webhook123 not found"

	assert.Equal(t, expectedMsg, err.Error())
}

func TestGetEventsRequest_Validate(t *testing.T) {
	testCases := []struct {
		name           string
		request        GetEventsRequest
		expectedError  string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name: "Valid request with defaults",
			request: GetEventsRequest{
				WorkspaceID: "workspace123",
			},
			expectedError:  "",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name: "Valid request with custom values",
			request: GetEventsRequest{
				WorkspaceID: "workspace123",
				Type:        EmailEventDelivered,
				Limit:       50,
				Offset:      10,
			},
			expectedError:  "",
			expectedLimit:  50,
			expectedOffset: 10,
		},
		{
			name: "Missing workspace ID",
			request: GetEventsRequest{
				Type:  EmailEventDelivered,
				Limit: 50,
			},
			expectedError:  "workspace_id is required",
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name: "Limit too high",
			request: GetEventsRequest{
				WorkspaceID: "workspace123",
				Limit:       200,
			},
			expectedError:  "",
			expectedLimit:  100, // Should be capped at 100
			expectedOffset: 0,
		},
		{
			name: "Negative limit",
			request: GetEventsRequest{
				WorkspaceID: "workspace123",
				Limit:       -10,
			},
			expectedError:  "",
			expectedLimit:  20, // Should be set to default
			expectedOffset: 0,
		},
		{
			name: "Negative offset",
			request: GetEventsRequest{
				WorkspaceID: "workspace123",
				Offset:      -5,
			},
			expectedError:  "",
			expectedLimit:  20,
			expectedOffset: 0, // Should be set to 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedLimit, tc.request.Limit)
			assert.Equal(t, tc.expectedOffset, tc.request.Offset)
		})
	}
}

func TestGetEventByIDRequest_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		request       GetEventByIDRequest
		expectedError string
	}{
		{
			name: "Valid request",
			request: GetEventByIDRequest{
				ID: "webhook123",
			},
			expectedError: "",
		},
		{
			name: "Missing ID",
			request: GetEventByIDRequest{
				ID: "",
			},
			expectedError: "id is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEventsByMessageIDRequest_Validate(t *testing.T) {
	testCases := []struct {
		name           string
		request        GetEventsByMessageIDRequest
		expectedError  string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name: "Valid request with defaults",
			request: GetEventsByMessageIDRequest{
				MessageID: "message123",
			},
			expectedError:  "",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name: "Valid request with custom values",
			request: GetEventsByMessageIDRequest{
				MessageID: "message123",
				Limit:     50,
				Offset:    10,
			},
			expectedError:  "",
			expectedLimit:  50,
			expectedOffset: 10,
		},
		{
			name: "Missing message ID",
			request: GetEventsByMessageIDRequest{
				Limit:  50,
				Offset: 10,
			},
			expectedError:  "message_id is required",
			expectedLimit:  50,
			expectedOffset: 10,
		},
		{
			name: "Limit too high",
			request: GetEventsByMessageIDRequest{
				MessageID: "message123",
				Limit:     200,
			},
			expectedError:  "",
			expectedLimit:  100, // Should be capped at 100
			expectedOffset: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedLimit, tc.request.Limit)
			assert.Equal(t, tc.expectedOffset, tc.request.Offset)
		})
	}
}

func TestGetEventsByTransactionalIDRequest_Validate(t *testing.T) {
	testCases := []struct {
		name           string
		request        GetEventsByTransactionalIDRequest
		expectedError  string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name: "Valid request with defaults",
			request: GetEventsByTransactionalIDRequest{
				WorkspaceID:     "workspace123",
				TransactionalID: "trans123",
			},
			expectedError:  "",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name: "Valid request with custom values",
			request: GetEventsByTransactionalIDRequest{
				WorkspaceID:     "workspace123",
				TransactionalID: "trans123",
				Limit:           50,
				Offset:          10,
			},
			expectedError:  "",
			expectedLimit:  50,
			expectedOffset: 10,
		},
		{
			name: "Missing workspace ID",
			request: GetEventsByTransactionalIDRequest{
				TransactionalID: "trans123",
			},
			expectedError:  "workspace_id is required",
			expectedLimit:  0,
			expectedOffset: 0,
		},
		{
			name: "Missing transactional ID",
			request: GetEventsByTransactionalIDRequest{
				WorkspaceID: "workspace123",
			},
			expectedError:  "transactional_id is required",
			expectedLimit:  0,
			expectedOffset: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if err == nil {
				assert.Equal(t, tc.expectedLimit, tc.request.Limit)
				assert.Equal(t, tc.expectedOffset, tc.request.Offset)
			}
		})
	}
}

func TestGetEventsByBroadcastIDRequest_Validate(t *testing.T) {
	testCases := []struct {
		name           string
		request        GetEventsByBroadcastIDRequest
		expectedError  string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name: "Valid request with defaults",
			request: GetEventsByBroadcastIDRequest{
				WorkspaceID: "workspace123",
				BroadcastID: "broadcast123",
			},
			expectedError:  "",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name: "Valid request with custom values",
			request: GetEventsByBroadcastIDRequest{
				WorkspaceID: "workspace123",
				BroadcastID: "broadcast123",
				Limit:       50,
				Offset:      10,
			},
			expectedError:  "",
			expectedLimit:  50,
			expectedOffset: 10,
		},
		{
			name: "Missing workspace ID",
			request: GetEventsByBroadcastIDRequest{
				BroadcastID: "broadcast123",
			},
			expectedError:  "workspace_id is required",
			expectedLimit:  0,
			expectedOffset: 0,
		},
		{
			name: "Missing broadcast ID",
			request: GetEventsByBroadcastIDRequest{
				WorkspaceID: "workspace123",
			},
			expectedError:  "broadcast_id is required",
			expectedLimit:  0,
			expectedOffset: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if err == nil {
				assert.Equal(t, tc.expectedLimit, tc.request.Limit)
				assert.Equal(t, tc.expectedOffset, tc.request.Offset)
			}
		})
	}
}
