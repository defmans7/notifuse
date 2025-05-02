//go:build !test
// +build !test

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
)

// setupMockDB creates a mock database and sqlmock for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err, "Failed to create mock database")

	cleanup := func() {
		db.Close()
	}

	return db, mock, cleanup
}

func TestGetContactByEmail(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(db, nil).AnyTimes()

	repo := NewContactRepository(workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "test@example.com"

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone", "language",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
		"created_at", "updated_at",
	}).
		AddRow(
			email, "ext123", "Europe/Paris", "en-US",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			[]byte(`{"key": "value1"}`), []byte(`{"key": "value2"}`), []byte(`{"key": "value3"}`), []byte(`{"key": "value4"}`), []byte(`{"key": "value5"}`),
			now, now,
		)

	mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c.email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	contact, err := repo.GetContactByEmail(context.Background(), "workspace123", email)
	require.NoError(t, err)
	assert.Equal(t, email, contact.Email)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c.email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByEmail(context.Background(), "workspace123", "nonexistent@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contact not found")
}

func TestGetContactByExternalID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(db, nil).AnyTimes()

	repo := NewContactRepository(workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)
	externalID := "ext123"

	// Test case 1: Contact found
	rows := sqlmock.NewRows([]string{
		"email", "external_id", "timezone", "language",
		"first_name", "last_name", "phone", "address_line_1", "address_line_2",
		"country", "postcode", "state", "job_title",
		"lifetime_value", "orders_count", "last_order_at",
		"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
		"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
		"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
		"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
		"created_at", "updated_at",
	}).
		AddRow(
			"test@example.com", externalID, "Europe/Paris", "en-US",
			"John", "Doe", "+1234567890", "123 Main St", "Apt 4B",
			"USA", "12345", "CA", "Developer",
			100.50, 5, now,
			"Custom 1", "Custom 2", "Custom 3", "Custom 4", "Custom 5",
			42.0, 43.0, 44.0, 45.0, 46.0,
			now, now, now, now, now,
			[]byte(`{"key": "value1"}`), []byte(`{"key": "value2"}`), []byte(`{"key": "value3"}`), []byte(`{"key": "value4"}`), []byte(`{"key": "value5"}`),
			now, now,
		)

	mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c.external_id = \$1`).
		WithArgs(externalID).
		WillReturnRows(rows)

	_, err := repo.GetContactByExternalID(context.Background(), externalID, "workspace123")
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c.external_id = \$1`).
		WithArgs("nonexistent-ext-id").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByExternalID(context.Background(), "nonexistent-ext-id", "workspace123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contact not found")

	// Test: get contact by external ID successful case
	t.Run("successful_case", func(t *testing.T) {
		externalID := "e-123"

		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language",
			"first_name", "last_name", "phone",
			"address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
			"created_at", "updated_at",
		}).AddRow(
			"test@example.com", "e-123", "Europe/Paris", "en-US", "John", "Doe", "", "", "", "", "", "", "", 0, 0, time.Time{},
			"", "", "", "", "", 0, 0, 0, 0, 0, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{},
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			time.Now(), time.Now(),
		)

		mock.ExpectQuery("SELECT c\\.\\* FROM contacts c WHERE c.external_id = \\$1").
			WithArgs(externalID).
			WillReturnRows(rows)

		// Act
		contact, err := repo.GetContactByExternalID(context.Background(), externalID, "workspace123")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, contact)
		assert.Equal(t, "test@example.com", contact.Email)
		assert.Equal(t, "e-123", contact.ExternalID.String)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetContacts(t *testing.T) {
	t.Run("should get contacts with pagination", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Set up expectations for the workspace database query
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name",
			"phone", "address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4",
			"custom_json_5", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "ext123", "UTC", "en", "John", "Doe",
			"+1234567890", "123 Main St", "Apt 4B", "US", "12345", "CA",
			"Engineer", 100.0, 5, time.Now(),
			"custom1", "custom2", "custom3", "custom4", "custom5",
			1.0, 2.0, 3.0, 4.0, 5.0,
			time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT c\.\* FROM contacts c ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs().
			WillReturnRows(rows)

		// Set up expectations for the contact lists query
		listRows := sqlmock.NewRows([]string{
			"email", "list_id", "status", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "list1", "active", time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT email, list_id, status, created_at, updated_at FROM contact_lists WHERE email IN \(\$1\)`).
			WithArgs("test@example.com").
			WillReturnRows(listRows)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Limit:            10,
			WithContactLists: true,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.Len(t, resp.Contacts[0].ContactLists, 1)
		assert.Equal(t, "list1", resp.Contacts[0].ContactLists[0].ListID)
		assert.Equal(t, domain.ContactListStatusActive, resp.Contacts[0].ContactLists[0].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should get contacts with multiple filters", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Set up expectations for the workspace database query
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name",
			"phone", "address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4",
			"custom_json_5", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "ext123", "UTC", "en", "John", "Doe",
			"+1234567890", "123 Main St", "Apt 4B", "US", "12345", "CA",
			"Engineer", 100.0, 5, time.Now(),
			"custom1", "custom2", "custom3", "custom4", "custom5",
			1.0, 2.0, 3.0, 4.0, 5.0,
			time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email ILIKE \$1 AND c\.first_name ILIKE \$2 AND c\.country ILIKE \$3 ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs("%test@example.com%", "%John%", "%US%").
			WillReturnRows(rows)

		// Set up expectations for the contact lists query
		listRows := sqlmock.NewRows([]string{
			"email", "list_id", "status", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "list1", "active", time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT email, list_id, status, created_at, updated_at FROM contact_lists WHERE email IN \(\$1\)`).
			WithArgs("test@example.com").
			WillReturnRows(listRows)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Email:            "test@example.com",
			FirstName:        "John",
			Country:          "US",
			Limit:            10,
			WithContactLists: true,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.Len(t, resp.Contacts[0].ContactLists, 1)
		assert.Equal(t, "list1", resp.Contacts[0].ContactLists[0].ListID)
		assert.Equal(t, domain.ContactListStatusActive, resp.Contacts[0].ContactLists[0].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle cursor pagination with base64 encoding", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Set up expectations for the workspace database query
		// Create multiple rows to trigger pagination
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name",
			"phone", "address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4",
			"custom_json_5", "created_at", "updated_at",
		})

		// Add multiple contacts to ensure pagination works
		for i := 1; i <= 11; i++ { // 11 to trigger the limit+1 logic
			rows.AddRow(
				fmt.Sprintf("test%d@example.com", i), fmt.Sprintf("ext%d", i), "UTC", "en",
				fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i),
				fmt.Sprintf("+%d", i), "123 Main St", "Apt 4B", "US", "12345", "CA",
				"Engineer", 100.0, 5, now,
				"custom1", "custom2", "custom3", "custom4", "custom5",
				1.0, 2.0, 3.0, 4.0, 5.0,
				now, now, now, now, now,
				[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
				[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
				now.Add(time.Duration(-i)*time.Hour), now, // Use decreasing created_at times
			)
		}

		// Truncate time to seconds to match the expected format
		cursorTime := time.Now().Truncate(time.Second)
		cursorEmail := "previous@example.com"
		cursorStr := fmt.Sprintf("%s~%s", cursorTime.Format(time.RFC3339), cursorEmail)
		encodedCursor := base64.StdEncoding.EncodeToString([]byte(cursorStr))

		// Parse the time back from the string to ensure it matches exactly what the test expects
		parsedTime, _ := time.Parse(time.RFC3339, cursorTime.Format(time.RFC3339))

		// The query should have compound condition for cursor-based pagination
		// Use a simpler regex pattern that's more forgiving of whitespace variations
		mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE \(c\.created_at < \$1 OR \(c\.created_at = \$2 AND c\.email > \$3\)\) ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs(parsedTime, parsedTime, cursorEmail).
			WillReturnRows(rows)

		// Set up expectations for the contact lists query - should have multiple emails
		emails := make([]string, 10) // We only get 10 because the 11th is cut off for pagination
		for i := 1; i <= 10; i++ {
			emails[i-1] = fmt.Sprintf("test%d@example.com", i)
		}

		// Create the expected SQL pattern for the IN query with multiple params
		// Use this simpler pattern to match the actual SQL generated
		sqlPattern := `SELECT email, list_id, status, created_at, updated_at FROM contact_lists WHERE email IN \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\)`

		listRows := sqlmock.NewRows([]string{
			"email", "list_id", "status", "created_at", "updated_at",
		})

		// Add contact list records for each email
		for _, email := range emails {
			listRows.AddRow(
				email, "list1", "active", now, now,
			)
		}

		// Convert emails to proper args for the mock
		emailArgs := make([]driver.Value, len(emails))
		for i, email := range emails {
			emailArgs[i] = email
		}

		mock.ExpectQuery(sqlPattern).
			WithArgs(emailArgs...).
			WillReturnRows(listRows)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Cursor:           encodedCursor,
			Limit:            10,
			WithContactLists: true,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 10) // Should get 10 contacts

		// Verify first and last contact
		assert.Equal(t, "test1@example.com", resp.Contacts[0].Email)
		assert.Equal(t, "test10@example.com", resp.Contacts[9].Email)

		// Verify contact lists
		for _, contact := range resp.Contacts {
			assert.Len(t, contact.ContactLists, 1)
			assert.Equal(t, "list1", contact.ContactLists[0].ListID)
			assert.Equal(t, domain.ContactListStatusActive, contact.ContactLists[0].Status)
		}

		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify the next cursor is base64 encoded and contains expected data
		require.NotEmpty(t, resp.NextCursor, "NextCursor should not be empty")

		decodedBytes, err := base64.StdEncoding.DecodeString(resp.NextCursor)
		require.NoError(t, err)

		cursorParts := strings.Split(string(decodedBytes), "~")
		require.Len(t, cursorParts, 2)

		_, err = time.Parse(time.RFC3339, cursorParts[0])
		require.NoError(t, err)

		// The 11th contact email should be in the cursor
		assert.Equal(t, "test10@example.com", cursorParts[1])
	})

	t.Run("should handle invalid base64 encoded cursor", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Cursor:           "invalid-base64-data",
			Limit:            10,
			WithContactLists: true,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor encoding")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid cursor format after base64 decoding", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Create a cursor with invalid format (missing tilde separator)
		invalidCursor := base64.StdEncoding.EncodeToString([]byte("invalid-cursor-format"))

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Cursor:           invalidCursor,
			Limit:            10,
			WithContactLists: true,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor format: expected timestamp~email")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid timestamp in cursor", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Create a cursor with invalid timestamp
		invalidCursor := base64.StdEncoding.EncodeToString([]byte("invalid-time~email@example.com"))

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Cursor:           invalidCursor,
			Limit:            10,
			WithContactLists: true,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor timestamp format")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle workspace connection errors", func(t *testing.T) {
		// Create a new mock workspace repository without a DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(nil, errors.New("failed to get workspace connection")).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Limit:            10,
			WithContactLists: true,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("should handle complex filter combinations", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Set up expectations for the workspace database query
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name",
			"phone", "address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4",
			"custom_json_5", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "ext123", "UTC", "en", "John", "Doe",
			"+1234567890", "123 Main St", "Apt 4B", "US", "12345", "CA",
			"Engineer", 100.0, 5, time.Now(),
			"custom1", "custom2", "custom3", "custom4", "custom5",
			1.0, 2.0, 3.0, 4.0, 5.0,
			time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email ILIKE \$1 AND c\.external_id ILIKE \$2 AND c\.first_name ILIKE \$3 AND c\.last_name ILIKE \$4 AND c\.phone ILIKE \$5 AND c\.country ILIKE \$6 ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs("%test@example.com%", "%ext123%", "%John%", "%Doe%", "%+1234567890%", "%US%").
			WillReturnRows(rows)

		// Set up expectations for the contact lists query
		listRows := sqlmock.NewRows([]string{
			"email", "list_id", "status", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "list1", "active", time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT email, list_id, status, created_at, updated_at FROM contact_lists WHERE email IN \(\$1\)`).
			WithArgs("test@example.com").
			WillReturnRows(listRows)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Email:            "test@example.com",
			ExternalID:       "ext123",
			FirstName:        "John",
			LastName:         "Doe",
			Phone:            "+1234567890",
			Country:          "US",
			Limit:            10,
			WithContactLists: true,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.Len(t, resp.Contacts[0].ContactLists, 1)
		assert.Equal(t, "list1", resp.Contacts[0].ContactLists[0].ListID)
		assert.Equal(t, domain.ContactListStatusActive, resp.Contacts[0].ContactLists[0].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle database query errors", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Set up expectations for the query to fail
		mock.ExpectQuery(`SELECT c\.\* FROM contacts c ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs().
			WillReturnError(errors.New("database query error"))

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Limit:            10,
			WithContactLists: false,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute query")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle contact scan errors", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Create a row with invalid data that will cause scan to fail
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name",
			"phone", "address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4",
			"custom_json_5", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "ext123", "UTC", "en", "John", "Doe",
			"+1234567890", "123 Main St", "Apt 4B", "US", "12345", "CA",
			"Engineer", "not-a-number", 5, time.Now(), // lifetime_value as a string to cause scan error
			"custom1", "custom2", "custom3", "custom4", "custom5",
			1.0, 2.0, 3.0, 4.0, 5.0,
			time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT c\.\* FROM contacts c ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs().
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Limit:            10,
			WithContactLists: false,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan contact")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle error in rows iteration", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := setupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(mockDB, nil).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		// Create rows that will return an error during iteration
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name",
			"phone", "address_line_1", "address_line_2", "country", "postcode", "state",
			"job_title", "lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4",
			"custom_string_5", "custom_number_1", "custom_number_2", "custom_number_3",
			"custom_number_4", "custom_number_5", "custom_datetime_1", "custom_datetime_2",
			"custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4",
			"custom_json_5", "created_at", "updated_at",
		}).AddRow(
			"test@example.com", "ext123", "UTC", "en", "John", "Doe",
			"+1234567890", "123 Main St", "Apt 4B", "US", "12345", "CA",
			"Engineer", 100.0, 5, time.Now(),
			"custom1", "custom2", "custom3", "custom4", "custom5",
			1.0, 2.0, 3.0, 4.0, 5.0,
			time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			[]byte(`{"key": "value"}`), []byte(`{"key": "value"}`),
			time.Now(), time.Now(),
		).RowError(0, errors.New("row error"))

		mock.ExpectQuery(`SELECT c\.\* FROM contacts c ORDER BY c\.created_at DESC, c\.email ASC LIMIT 11`).
			WithArgs().
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID:      "workspace123",
			Limit:            10,
			WithContactLists: false,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error iterating over rows")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteContact(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(db, nil).AnyTimes()

	repo := NewContactRepository(workspaceRepo)
	email := "test@example.com"

	t.Run("should delete existing contact", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
			WithArgs(email).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteContact(context.Background(), email, "workspace123")
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle non-existent contact", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
			WithArgs("nonexistent@example.com").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteContact(context.Background(), "nonexistent@example.com", "workspace123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contact not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle database execution errors", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(errors.New("database error"))

		err := repo.DeleteContact(context.Background(), email, "workspace123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete contact")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle rows affected errors", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
			WithArgs(email).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.DeleteContact(context.Background(), email, "workspace123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get affected rows")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle workspace connection errors", func(t *testing.T) {
		// Create a new mock workspace repository without a DB
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		workspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace123").Return(nil, errors.New("failed to get workspace connection")).AnyTimes()

		repo := NewContactRepository(workspaceRepo)

		err := repo.DeleteContact(context.Background(), email, "workspace123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
	})
}
