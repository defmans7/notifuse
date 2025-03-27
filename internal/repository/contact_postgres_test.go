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

	repo := NewContactRepository(db)
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

	contact, err := repo.GetContactByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.Equal(t, email, contact.Email)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByEmail(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
}

func TestGetContactByExternalID(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
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

	_, err := repo.GetContactByExternalID(context.Background(), externalID)
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE external_id = \$1`).
		WithArgs("nonexistent-ext-id").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetContactByExternalID(context.Background(), "nonexistent-ext-id")
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
		contact, err := repo.GetContactByExternalID(context.Background(), externalID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, contact)
		assert.Equal(t, "test@example.com", contact.Email)
		assert.Equal(t, "e-123", contact.ExternalID.String)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetContacts(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)

	t.Run("should get contacts with pagination", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
		}

		now := time.Now().UTC()
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
			"contact1@example.com", "ext-1", "UTC", "en-US", "Contact", "One", nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now, now,
		).AddRow(
			"contact2@example.com", "ext-2", "UTC", "en-US", "Contact", "Two", nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now.Add(-1*time.Hour), now,
		)

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
			WithArgs(req.WorkspaceID, req.Limit+1).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Contacts, 2)
		assert.Empty(t, response.NextCursor) // No more results
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle cursor-based pagination", func(t *testing.T) {
		// Arrange
		cursorTime := time.Now().UTC().Truncate(time.Second)
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
			Cursor:      cursorTime.Format(time.RFC3339),
		}

		now := cursorTime.Add(-1 * time.Hour)
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
			"contact1@example.com", "ext-1", "UTC", "en-US", "Contact", "One", nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now, now,
		).AddRow(
			"contact2@example.com", "ext-2", "UTC", "en-US", "Contact", "Two", nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now.Add(-1*time.Hour), now,
		)

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 AND created_at < \\$2 ORDER BY created_at DESC LIMIT \\$3").
			WithArgs(req.WorkspaceID, cursorTime, req.Limit+1).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Contacts, 2)
		assert.Empty(t, response.NextCursor) // No more results
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return empty response when no contacts", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "empty-workspace",
			Limit:       2,
		}

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
		})

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
			WithArgs(req.WorkspaceID, req.Limit+1).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Contacts)
		assert.Empty(t, response.NextCursor)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle database error", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "error-workspace",
			Limit:       2,
		}

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
			WithArgs(req.WorkspaceID, req.Limit+1).
			WillReturnError(errors.New("database error"))

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get contacts")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle filtering by various fields", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
			Email:       "test@",
			ExternalID:  "ext-",
			FirstName:   "John",
			LastName:    "Doe",
			Phone:       "+123",
			Country:     "USA",
		}

		now := time.Now().UTC()
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
			"test@example.com", "ext-1", "UTC", "en-US", "John", "Doe", "+1234567890",
			nil, nil, "USA", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now, now,
		)

		expectedQuery := "SELECT (.+) FROM contacts WHERE workspace_id = \\$1 AND email ILIKE \\$2 AND external_id ILIKE \\$3 AND first_name ILIKE \\$4 AND last_name ILIKE \\$5 AND phone ILIKE \\$6 AND country ILIKE \\$7 ORDER BY created_at DESC LIMIT \\$8"
		mock.ExpectQuery(expectedQuery).
			WithArgs(
				req.WorkspaceID,
				"%"+req.Email+"%",
				"%"+req.ExternalID+"%",
				"%"+req.FirstName+"%",
				"%"+req.LastName+"%",
				"%"+req.Phone+"%",
				"%"+req.Country+"%",
				req.Limit+1,
			).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Contacts, 1)
		assert.Equal(t, "test@example.com", response.Contacts[0].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid cursor format", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
			Cursor:      "invalid-timestamp",
		}

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid cursor format")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle next cursor when there are more results", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
		}

		now := time.Now().UTC()
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
			"contact1@example.com", "ext-1", "UTC", "en-US", "Contact", "One", nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now, now,
		).AddRow(
			"contact2@example.com", "ext-2", "UTC", "en-US", "Contact", "Two", nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now.Add(-1*time.Hour), now,
		).AddRow(
			"contact3@example.com", "ext-3", "UTC", "en-US", "Contact", "Three", nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			now.Add(-2*time.Hour), now,
		)

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
			WithArgs(req.WorkspaceID, req.Limit+1).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Contacts, 2)
		assert.NotEmpty(t, response.NextCursor)
		assert.Equal(t, "contact2@example.com", response.Contacts[1].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle database error during row iteration", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
		}

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
			"contact1@example.com", "ext-1", "UTC", "en-US", "Contact", "One", nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			time.Now(), time.Now(),
		).RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
			WithArgs(req.WorkspaceID, req.Limit+1).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "row iteration error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle scan error", func(t *testing.T) {
		// Arrange
		req := &domain.GetContactsRequest{
			WorkspaceID: "workspace123",
			Limit:       2,
		}

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
			"contact1@example.com", "ext-1", "UTC", "en-US", "Contact", "One", nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			[]byte("{invalid json}"), []byte("{}"), []byte("{}"), []byte("{}"), []byte("{}"),
			time.Now(), time.Now(),
		).RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery("SELECT (.+) FROM contacts WHERE workspace_id = \\$1 ORDER BY created_at DESC LIMIT \\$2").
			WithArgs(req.WorkspaceID, req.Limit+1).
			WillReturnRows(rows)

		// Act
		response, err := repo.GetContacts(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "error iterating contacts rows")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteContact(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	email := "test@example.com"

	// Test case 1: Contact successfully deleted
	mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteContact(context.Background(), email)
	require.NoError(t, err)

	// Test case 2: Contact not found
	mock.ExpectExec(`DELETE FROM contacts WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteContact(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactNotFound{}, err)
}
