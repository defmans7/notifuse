package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestContactHandler_HandleUpsert(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		expectedAction string
		checkResult    func(*testing.T, *MockContactService)
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
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = true // Indicate this is a new contact
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
				assert.NotNil(t, m.LastContactUpserted)
				if m.LastContactUpserted.Email != "new@example.com" {
					t.Errorf("Expected contact email %s, got %s", "new@example.com", m.LastContactUpserted.Email)
				}
				if m.LastContactUpserted.FirstName.String != "John" || m.LastContactUpserted.FirstName.IsNull {
					t.Errorf("Expected contact first name 'John', got '%+v'", m.LastContactUpserted.FirstName)
				}
			},
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
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = true // Indicate this is a new contact
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
				assert.NotNil(t, m.LastContactUpserted)
				if m.LastContactUpserted.Email != "new@example.com" {
					t.Errorf("Expected contact email %s, got %s", "new@example.com", m.LastContactUpserted.Email)
				}
				if m.LastContactUpserted.FirstName.String != "John" || m.LastContactUpserted.FirstName.IsNull {
					t.Errorf("Expected contact first name 'John', got '%+v'", m.LastContactUpserted.FirstName)
				}
			},
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
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = false // Indicate this is an existing contact

				// Add existing contact
				m.contacts["old@example.com"] = &domain.Contact{
					ExternalID: &domain.NullableString{String: "old-ext", IsNull: false},
					Email:      "old@example.com",
					Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
					FirstName: &domain.NullableString{
						String: "Old",
						IsNull: false,
					},
					LastName: &domain.NullableString{
						String: "Name",
						IsNull: false,
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedAction: "updated",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
				assert.NotNil(t, m.LastContactUpserted)

				// Create expected contact by merging the update
				expectedContact := &domain.Contact{
					ExternalID: &domain.NullableString{String: "old-ext", IsNull: false},
					Email:      "old@example.com",
					Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
					FirstName: &domain.NullableString{
						String: "Old",
						IsNull: false,
					},
					LastName: &domain.NullableString{
						String: "Name",
						IsNull: false,
					},
				}

				// Create update contact with only the fields being updated
				updateContact := &domain.Contact{
					ExternalID: &domain.NullableString{String: "updated-ext", IsNull: false},
					Email:      "old@example.com",
					Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
				}

				// Merge the update into the expected contact
				expectedContact.Merge(updateContact)

				// Compare the merged contact with the actual result
				assert.Equal(t, expectedContact.Email, m.LastContactUpserted.Email)
				assert.Equal(t, expectedContact.ExternalID.String, m.LastContactUpserted.ExternalID.String)
				assert.Equal(t, expectedContact.ExternalID.IsNull, m.LastContactUpserted.ExternalID.IsNull)
				assert.Equal(t, expectedContact.Timezone.String, m.LastContactUpserted.Timezone.String)
				assert.Equal(t, expectedContact.Timezone.IsNull, m.LastContactUpserted.Timezone.IsNull)
				assert.Equal(t, expectedContact.FirstName.String, m.LastContactUpserted.FirstName.String)
				assert.Equal(t, expectedContact.FirstName.IsNull, m.LastContactUpserted.FirstName.IsNull)
				assert.Equal(t, expectedContact.LastName.String, m.LastContactUpserted.LastName.String)
				assert.Equal(t, expectedContact.LastName.IsNull, m.LastContactUpserted.LastName.IsNull)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactService) {
				m.UpsertContactCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			expectedAction: "",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.False(t, m.UpsertContactCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.UpsertContactCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAction: "",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.False(t, m.UpsertContactCalled)
			},
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
			setupMock: func(m *MockContactService) {
				m.UpsertContactCalled = false
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedAction: "",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockContactService{
				contacts: make(map[string]*domain.Contact),
			}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactHandler(mockService, mockLogger)

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

			// Run specific checks
			tc.checkResult(t, mockService)
		})
	}
}

func TestContactHandler_HandleUpsertWithCustomJSON(t *testing.T) {
	mockService := &MockContactService{
		contacts: map[string]*domain.Contact{
			"test@example.com": {
				Email:      "test@example.com",
				ExternalID: &domain.NullableString{String: "old-ext", IsNull: false},
				Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
				Language:   &domain.NullableString{String: "en-US", IsNull: false},
			},
		},
		UpsertIsNewToReturn: false,
	}
	mockLogger := &MockLoggerForContact{}
	handler := NewContactHandler(mockService, mockLogger)

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

	req, err := http.NewRequest(http.MethodPost, "/api/contacts.upsert", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.handleUpsert(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
	}

	// Verify that the service was called with the correct contact
	if !mockService.UpsertContactCalled {
		t.Error("Expected UpsertContact to be called, but it wasn't")
	}

	if mockService.LastContactUpserted == nil {
		t.Error("Expected LastContactUpserted to be set, but it wasn't")
	} else {
		if mockService.LastContactUpserted.Email != "test@example.com" {
			t.Errorf("Expected contact email %s, got %s", "test@example.com", mockService.LastContactUpserted.Email)
		}

		// Check external_id regardless of how it's stored internally
		expectedExternalId := "ext123"
		actualExternalId := mockService.LastContactUpserted.ExternalID.String
		if actualExternalId != expectedExternalId || mockService.LastContactUpserted.ExternalID.IsNull {
			t.Errorf("Expected contact external_id %s, got %v", expectedExternalId, mockService.LastContactUpserted.ExternalID)
		}

		// Check timezone regardless of how it's stored internally
		expectedTimezone := "Europe/Paris"
		actualTimezone := mockService.LastContactUpserted.Timezone.String
		if actualTimezone != expectedTimezone || mockService.LastContactUpserted.Timezone.IsNull {
			t.Errorf("Expected contact timezone %s, got %v", expectedTimezone, mockService.LastContactUpserted.Timezone)
		}

		if mockService.LastContactUpserted.Language.String != "en-US" || mockService.LastContactUpserted.Language.IsNull {
			t.Errorf("Expected contact language %s, got %v", "en-US", mockService.LastContactUpserted.Language)
		}

		// Verify custom JSON fields
		if mockService.LastContactUpserted.CustomJSON1.IsNull {
			t.Error("Expected CustomJSON1 to not be null")
		}

		if !mockService.LastContactUpserted.CustomJSON2.IsNull {
			t.Error("Expected CustomJSON2 to be null")
		}

		if mockService.LastContactUpserted.CustomJSON3.IsNull {
			t.Error("Expected CustomJSON3 to not be null")
		}
	}

	// Verify response
	var response map[string]interface{}
	if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	contact, ok := response["contact"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'contact' field in response, but not found")
	}

	action, ok := response["action"].(string)
	if !ok {
		t.Error("Expected 'action' field in response, but not found")
	}

	if mockService.UpsertIsNewToReturn {
		if action != "created" {
			t.Errorf("Expected action 'created', got %s", action)
		}
	} else {
		if action != "updated" {
			t.Errorf("Expected action 'updated', got %s", action)
		}
	}

	// Verify contact fields in response
	if contact != nil {
		if email, ok := contact["email"].(string); !ok || email != "test@example.com" {
			t.Errorf("Expected contact email %s, got %v", "test@example.com", email)
		}

		// Check external_id field - could be a string or a map
		externalID, ok := contact["external_id"]
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
		timezone, ok := contact["timezone"]
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
	mockService := NewMockContactService()
	mockLogger := &MockLoggerForContact{}
	handler := NewContactHandler(mockService, mockLogger)

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
			req := httptest.NewRequest(http.MethodPost, "/api/contacts.upsert", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.handleUpsert(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
