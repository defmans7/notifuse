package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertContact(t *testing.T) {
	db, _, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "test@example.com"
	workspaceID := "workspace123"

	testContact := &domain.Contact{
		Email:      email,
		ExternalID: &domain.NullableString{String: "ext123", IsNull: false},
		Timezone:   &domain.NullableString{String: "Europe/Paris", IsNull: false},
		Language:   &domain.NullableString{String: "en-US", IsNull: false},
		FirstName:  &domain.NullableString{String: "John", IsNull: false},
		LastName:   &domain.NullableString{String: "Doe", IsNull: false},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	t.Run("insert new contact", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update that returns no rows
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Expect insert
		newMock.ExpectExec(`INSERT INTO contacts`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect transaction commit
		newMock.ExpectCommit()

		// Execute the function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, testContact)
		require.NoError(t, err)
		assert.True(t, isNew)

		// Verify all expectations were met
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("update existing contact", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Create an existing contact with some fields
		existingContact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "old-ext", IsNull: false},
			FirstName:  &domain.NullableString{String: "Old", IsNull: false},
			LastName:   &domain.NullableString{String: "Name", IsNull: false},
			CreatedAt:  now.Add(-24 * time.Hour), // Created yesterday
			UpdatedAt:  now.Add(-24 * time.Hour),
		}

		// Prepare mock for existing contact
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name", "phone",
			"address_line_1", "address_line_2", "country", "postcode", "state", "job_title",
			"lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
			"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
			"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
			"created_at", "updated_at",
		}).
			AddRow(
				existingContact.Email, "old-ext", nil, nil, "Old", "Name", nil,
				nil, nil, nil, nil, nil, nil,
				nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				existingContact.CreatedAt, existingContact.UpdatedAt,
			)

		// New contact data with updates
		updateContact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "updated-ext", IsNull: false},
			Timezone:   &domain.NullableString{String: "America/New_York", IsNull: false},
			Language:   &domain.NullableString{String: "en-US", IsNull: false},
		}

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update returning the existing contact
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnRows(rows)

		// Expect update with merged data
		newMock.ExpectExec(`UPDATE contacts SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect transaction commit
		newMock.ExpectCommit()

		// Execute the function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, updateContact)
		require.NoError(t, err)
		assert.False(t, isNew)

		// Verify all expectations were met
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails when workspace connection fails", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, _, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		// Don't add the invalid workspace
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Setup new mock with error
		invalidWorkspaceID := "invalid-workspace"

		// Execute function with invalid workspace
		isNew, err := newRepo.UpsertContact(context.Background(), invalidWorkspaceID, testContact)

		// Should fail with workspace connection error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace connection")
		assert.False(t, isNew)
	})

	t.Run("fails when transaction begin fails", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Expect error from BeginTx
		newMock.ExpectBegin().WillReturnError(errors.New("begin tx error"))

		// Execute function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, testContact)

		// Should fail with transaction error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails when select query fails", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select with error
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnError(errors.New("query error"))

		// Expect rollback
		newMock.ExpectRollback()

		// Execute function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, testContact)

		// Should fail with query error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check existing contact")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails when insert fails", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update that returns no rows
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Expect insert with error
		newMock.ExpectExec(`INSERT INTO contacts`).
			WillReturnError(errors.New("insert error"))

		// Expect rollback
		newMock.ExpectRollback()

		// Execute function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, testContact)

		// Should fail with insert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert contact")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails when update fails", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Create a mock for existing contact
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name", "phone",
			"address_line_1", "address_line_2", "country", "postcode", "state", "job_title",
			"lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
			"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
			"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
			"created_at", "updated_at",
		}).
			AddRow(
				email, "old-ext", nil, nil, "Old", "Name", nil,
				nil, nil, nil, nil, nil, nil,
				nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				now.Add(-24*time.Hour), now.Add(-24*time.Hour),
			)

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update returning the existing contact
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnRows(rows)

		// Expect update with error
		newMock.ExpectExec(`UPDATE contacts SET`).
			WillReturnError(errors.New("update error"))

		// Expect rollback
		newMock.ExpectRollback()

		// Execute function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, testContact)

		// Should fail with update error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update contact")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails when commit fails", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update that returns no rows
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Expect insert
		newMock.ExpectExec(`INSERT INTO contacts`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect commit to fail
		newMock.ExpectCommit().WillReturnError(errors.New("commit error"))

		// Execute function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, testContact)

		// Should fail with commit error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails with JSON marshal error on insert", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Create a contact with an unmarshalable JSON field
		contactWithBadJSON := &domain.Contact{
			Email: email,
			// Create a custom JSON field with a value that can't be marshaled
			CustomJSON1: &domain.NullableJSON{
				Data:   make(chan int), // channels can't be marshaled to JSON
				IsNull: false,
			},
		}

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update that returns no rows
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Expect rollback due to JSON marshal error
		newMock.ExpectRollback()

		// Execute the function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, contactWithBadJSON)

		// Should fail with JSON marshal error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal custom_json_1")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})

	t.Run("fails with JSON marshal error on update", func(t *testing.T) {
		// Setup new mock DB for this test
		newDb, newMock, newCleanup := testutil.SetupMockDB(t)
		defer newCleanup()

		newWorkspaceRepo := testutil.NewMockWorkspaceRepository(newDb)
		newWorkspaceRepo.AddWorkspaceDB("workspace123", newDb)
		newRepo := NewContactRepository(newWorkspaceRepo)

		// Create an existing contact
		rows := sqlmock.NewRows([]string{
			"email", "external_id", "timezone", "language", "first_name", "last_name", "phone",
			"address_line_1", "address_line_2", "country", "postcode", "state", "job_title",
			"lifetime_value", "orders_count", "last_order_at",
			"custom_string_1", "custom_string_2", "custom_string_3", "custom_string_4", "custom_string_5",
			"custom_number_1", "custom_number_2", "custom_number_3", "custom_number_4", "custom_number_5",
			"custom_datetime_1", "custom_datetime_2", "custom_datetime_3", "custom_datetime_4", "custom_datetime_5",
			"custom_json_1", "custom_json_2", "custom_json_3", "custom_json_4", "custom_json_5",
			"created_at", "updated_at",
		}).
			AddRow(
				email, "ext123", nil, nil, "John", "Doe", nil,
				nil, nil, nil, nil, nil, nil,
				nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				nil, nil, nil, nil, nil,
				time.Now(), time.Now(),
			)

		// Create an update with unmarshalable JSON
		contactWithBadJSON := &domain.Contact{
			Email: email,
			// Add a custom JSON field with a value that can't be marshaled
			CustomJSON2: &domain.NullableJSON{
				Data:   make(chan int), // channels can't be marshaled to JSON
				IsNull: false,
			},
		}

		// Expect transaction begin
		newMock.ExpectBegin()

		// Expect select for update that returns existing contact
		newMock.ExpectQuery(`SELECT c\.\* FROM contacts c WHERE c\.email = \$1 FOR UPDATE`).
			WithArgs(email).
			WillReturnRows(rows)

		// Expect rollback due to JSON marshal error
		newMock.ExpectRollback()

		// Execute the function
		isNew, err := newRepo.UpsertContact(context.Background(), workspaceID, contactWithBadJSON)

		// Should fail with JSON marshal error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal custom_json_2")
		assert.False(t, isNew)
		assert.NoError(t, newMock.ExpectationsWereMet())
	})
}
