package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a handler with public key
func setupContactHandlerUpsertTest(t *testing.T) (*mocks.MockContactService, *MockLoggerForContact, *ContactHandler) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockContactService(ctrl)
	mockLogger := &MockLoggerForContact{}

	// Create key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	handler := NewContactHandler(mockService, publicKey, mockLogger)
	return mockService, mockLogger, handler
}

func TestContactHandler_HandleUpsert(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*mocks.MockContactService)
		expectedStatus int
		expectedAction string
	}{
		{
			name:   "Create Contact Without UUID",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contact": map[string]interface{}{
					"external_id": "new-ext",
					"email":       "new@example.com",
					"first_name":  "John",
					"last_name":   "Doe",
					"timezone":    "UTC",
				},
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					UpsertContact(gomock.Any(), "workspace123", gomock.Any()).
					Return(true, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
		},
		{
			name:   "Create Contact With Email",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contact": map[string]interface{}{
					"external_id": "new-ext",
					"email":       "new@example.com",
					"first_name":  "John",
					"last_name":   "Doe",
					"timezone":    "UTC",
				},
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					UpsertContact(gomock.Any(), "workspace123", gomock.Any()).
					Return(true, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
		},
		{
			name:   "Update Existing Contact",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contact": map[string]interface{}{
					"external_id": "updated-ext",
					"email":       "old@example.com",
					"timezone":    "UTC",
				},
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					UpsertContact(gomock.Any(), "workspace123", gomock.Any()).
					Return(false, nil)
			},
			expectedStatus: http.StatusOK,
			expectedAction: "updated",
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().UpsertContact(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			expectedAction: "",
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"timezone":    "UTC",
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().UpsertContact(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAction: "",
		},
		{
			name:   "Service Error on Upsert",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contact": map[string]interface{}{
					"external_id": "ext1",
					"email":       "test@example.com",
					"timezone":    "UTC",
				},
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					UpsertContact(gomock.Any(), "workspace123", gomock.Any()).
					Return(false, fmt.Errorf("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedAction: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerUpsertTest(t)
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				// If it's a string, just use it directly
				if str, ok := tc.reqBody.(string); ok {
					reqBody = *bytes.NewBufferString(str)
				} else {
					// Otherwise encode as JSON
					if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
						t.Fatalf("Failed to encode request body: %v", err)
					}
				}
			}

			req := httptest.NewRequest(tc.method, "/api/contacts.upsert", &reqBody)
			if err := req.ParseForm(); err != nil {
				t.Fatalf("Failed to parse form: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleUpsert(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check response body for success cases
			if tc.expectedStatus == http.StatusOK || tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check action field
				action, exists := response["action"]
				assert.True(t, exists)
				assert.Equal(t, tc.expectedAction, action)

				// Check contact exists
				_, exists = response["contact"]
				assert.True(t, exists)
			}
		})
	}
}

func TestContactHandler_HandleUpsertWithCustomJSON(t *testing.T) {
	mockService, _, handler := setupContactHandlerUpsertTest(t)

	// Test case 1: Successful upsert with custom JSON fields
	reqBody := `{
		"workspace_id": "workspace123",
		"contact": {
			"email": "test@example.com",
			"external_id": "ext123",
			"timezone": "Europe/Paris",
			"language": "en-US",
			"custom_json_1": {"key": "value1"},
			"custom_json_2": null,
			"custom_json_3": {"key": "value3"}
		}
	}`

	mockService.EXPECT().
		UpsertContact(gomock.Any(), "workspace123", gomock.Any()).
		Return(false, nil)

	req, err := http.NewRequest(http.MethodPost, "/api/contacts.upsert", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.handleUpsert(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
	}

	// Verify response
	var response map[string]interface{}
	if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	contactResponse, ok := response["contact"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'contact' field in response, but not found")
	}

	action, ok := response["action"].(string)
	if !ok {
		t.Error("Expected 'action' field in response, but not found")
	}

	if action != "updated" {
		t.Errorf("Expected action 'updated', got %s", action)
	}

	// Verify contact fields in response
	if contactResponse != nil {
		if email, ok := contactResponse["email"].(string); !ok || email != "test@example.com" {
			t.Errorf("Expected contact email %s, got %v", "test@example.com", email)
		}

		// Check external_id field - could be a string or a map
		externalID, ok := contactResponse["external_id"]
		if !ok {
			t.Errorf("Expected external_id in response, but not found")
		} else {
			// Get the value regardless of format
			var externalIDValue string
			switch v := externalID.(type) {
			case string:
				externalIDValue = v
			case map[string]interface{}:
				externalIDValue, _ = v["String"].(string)
			}

			expectedExternalId := "ext123"
			if externalIDValue != expectedExternalId {
				t.Errorf("Expected contact external_id %s, got %v", expectedExternalId, externalIDValue)
			}
		}

		// Check timezone field - could be a string or a map
		timezone, ok := contactResponse["timezone"]
		if !ok {
			t.Errorf("Expected timezone in response, but not found")
		} else {
			// Get the value regardless of format
			var timezoneValue string
			switch v := timezone.(type) {
			case string:
				timezoneValue = v
			case map[string]interface{}:
				timezoneValue, _ = v["String"].(string)
			}

			expectedTimezone := "Europe/Paris"
			if timezoneValue != expectedTimezone {
				t.Errorf("Expected contact timezone %s, got %v", expectedTimezone, timezoneValue)
			}
		}
	}
}

func TestContactHandler_HandleUpsertWithInvalidJSON(t *testing.T) {
	mockService, _, handler := setupContactHandlerUpsertTest(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "invalid JSON syntax",
			requestBody:    `{"email": "test@example.com", "language": "en-US", invalid_json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing required email field",
			requestBody:    `{"external_id": "ext123", "language": "en-US"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email", "language": "en-US"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request body",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.EXPECT().UpsertContact(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			req := httptest.NewRequest(http.MethodPost, "/api/contacts.upsert", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.handleUpsert(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
