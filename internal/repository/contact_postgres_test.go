package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

func TestGetContactByEmail(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
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

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	contact, err := repo.GetContactByEmail(context.Background(), email, "workspace123")
	require.NoError(t, err)
	assert.Equal(t, email, contact.Email)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByEmail(context.Background(), "nonexistent@example.com", "workspace123")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
}

func TestGetContactByExternalID(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
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

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE external_id = \$1`).
		WithArgs(externalID).
		WillReturnRows(rows)

	_, err := repo.GetContactByExternalID(context.Background(), externalID, "workspace123")
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE external_id = \$1`).
		WithArgs("nonexistent-ext-id").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByExternalID(context.Background(), "nonexistent-ext-id", "workspace123")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)

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

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE external_id = \\$1").
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
		mockDB, mock, cleanup := testutil.SetupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(mockDB)
		workspaceRepo.AddWorkspaceDB("workspace123", mockDB)
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

		mock.ExpectQuery(`SELECT (.+) FROM contacts ORDER BY created_at DESC LIMIT \$1`).
			WithArgs(11).
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       10,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should get contacts with multiple filters", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := testutil.SetupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(mockDB)
		workspaceRepo.AddWorkspaceDB("workspace123", mockDB)
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

		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email ILIKE \$1 AND first_name ILIKE \$2 AND country ILIKE \$3 ORDER BY created_at DESC LIMIT \$4`).
			WithArgs("%test@example.com%", "%John%", "%US%", 11).
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Email:       "test@example.com",
			FirstName:   "John",
			Country:     "US",
			Limit:       10,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle cursor pagination edge cases", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := testutil.SetupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(mockDB)
		workspaceRepo.AddWorkspaceDB("workspace123", mockDB)
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

		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE created_at < \$1 ORDER BY created_at DESC LIMIT \$2`).
			WithArgs(sqlmock.AnyArg(), 11).
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Cursor:      time.Now().Format(time.RFC3339),
			Limit:       10,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle workspace connection errors", func(t *testing.T) {
		// Create a new mock workspace repository without a DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(nil)
		repo := NewContactRepository(workspaceRepo)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       10,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
	})

	t.Run("should handle complex filter combinations", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := testutil.SetupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(mockDB)
		workspaceRepo.AddWorkspaceDB("workspace123", mockDB)
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

		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email ILIKE \$1 AND external_id ILIKE \$2 AND first_name ILIKE \$3 AND last_name ILIKE \$4 AND phone ILIKE \$5 AND country ILIKE \$6 ORDER BY created_at DESC LIMIT \$7`).
			WithArgs("%test@example.com%", "%ext123%", "%John%", "%Doe%", "%+1234567890%", "%US%", 11).
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Email:       "test@example.com",
			ExternalID:  "ext123",
			FirstName:   "John",
			LastName:    "Doe",
			Phone:       "+1234567890",
			Country:     "US",
			Limit:       10,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, resp.Contacts, 1)
		assert.Equal(t, "test@example.com", resp.Contacts[0].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid cursor format", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := testutil.SetupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(mockDB)
		workspaceRepo.AddWorkspaceDB("workspace123", mockDB)
		repo := NewContactRepository(workspaceRepo)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Cursor:      "invalid-timestamp",
			Limit:       10,
		}

		_, err := repo.GetContacts(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor format")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle empty result set", func(t *testing.T) {
		// Create a mock workspace database
		mockDB, mock, cleanup := testutil.SetupMockDB(t)
		defer cleanup()

		// Create a new repository with the mock DB
		workspaceRepo := testutil.NewMockWorkspaceRepository(mockDB)
		workspaceRepo.AddWorkspaceDB("workspace123", mockDB)
		repo := NewContactRepository(workspaceRepo)

		// Set up expectations for an empty result set
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

		mock.ExpectQuery(`SELECT (.+) FROM contacts ORDER BY created_at DESC LIMIT \$1`).
			WithArgs(11).
			WillReturnRows(rows)

		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       10,
		}

		resp, err := repo.GetContacts(context.Background(), req)
		require.NoError(t, err)
		assert.Empty(t, resp.Contacts)
		assert.Empty(t, resp.NextCursor)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteContact(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	// Add the workspace database to the workspace repository
	workspaceRepo.AddWorkspaceDB("workspace123", db)
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
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
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
		workspaceRepo := testutil.NewMockWorkspaceRepository(nil)
		repo := NewContactRepository(workspaceRepo)

		err := repo.DeleteContact(context.Background(), email, "workspace123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
	})
}
