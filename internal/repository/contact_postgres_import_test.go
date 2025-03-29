package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

func TestBatchImportContacts(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
	repo := NewContactRepository(workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create some test contacts
	contact1 := &domain.Contact{
		Email:           "contact1@example.com",
		ExternalID:      &domain.NullableString{String: "ext1", IsNull: false},
		Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
		Language:        &domain.NullableString{String: "en-US", IsNull: false},
		FirstName:       &domain.NullableString{String: "John", IsNull: false},
		LastName:        &domain.NullableString{String: "Doe", IsNull: false},
		Phone:           &domain.NullableString{String: "", IsNull: true},
		AddressLine1:    &domain.NullableString{String: "", IsNull: true},
		AddressLine2:    &domain.NullableString{String: "", IsNull: true},
		Country:         &domain.NullableString{String: "", IsNull: true},
		Postcode:        &domain.NullableString{String: "", IsNull: true},
		State:           &domain.NullableString{String: "", IsNull: true},
		JobTitle:        &domain.NullableString{String: "", IsNull: true},
		LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
		LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomString1:   &domain.NullableString{String: "", IsNull: true},
		CustomString2:   &domain.NullableString{String: "", IsNull: true},
		CustomString3:   &domain.NullableString{String: "", IsNull: true},
		CustomString4:   &domain.NullableString{String: "", IsNull: true},
		CustomString5:   &domain.NullableString{String: "", IsNull: true},
		CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	contact2 := &domain.Contact{
		Email:           "contact2@example.com",
		ExternalID:      &domain.NullableString{String: "ext2", IsNull: false},
		Timezone:        &domain.NullableString{String: "America/New_York", IsNull: false},
		Language:        &domain.NullableString{String: "en-US", IsNull: false},
		FirstName:       &domain.NullableString{String: "Jane", IsNull: false},
		LastName:        &domain.NullableString{String: "Smith", IsNull: false},
		Phone:           &domain.NullableString{String: "", IsNull: true},
		AddressLine1:    &domain.NullableString{String: "", IsNull: true},
		AddressLine2:    &domain.NullableString{String: "", IsNull: true},
		Country:         &domain.NullableString{String: "", IsNull: true},
		Postcode:        &domain.NullableString{String: "", IsNull: true},
		State:           &domain.NullableString{String: "", IsNull: true},
		JobTitle:        &domain.NullableString{String: "", IsNull: true},
		LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
		LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomString1:   &domain.NullableString{String: "", IsNull: true},
		CustomString2:   &domain.NullableString{String: "", IsNull: true},
		CustomString3:   &domain.NullableString{String: "", IsNull: true},
		CustomString4:   &domain.NullableString{String: "", IsNull: true},
		CustomString5:   &domain.NullableString{String: "", IsNull: true},
		CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
		CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
		CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	testContacts := []*domain.Contact{contact1, contact2}

	// Test case 1: Successful batch import
	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO contacts`))

	anyArgs38 := []driver.Value{}
	for i := 0; i < 38; i++ {
		anyArgs38 = append(anyArgs38, sqlmock.AnyArg())
	}

	for range testContacts {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO contacts`)).
			WithArgs(
				anyArgs38...,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	err := repo.BatchImportContacts(context.Background(), "workspace123", testContacts)
	require.NoError(t, err)

	// Test case 2: Transaction begin error
	mock.ExpectBegin().WillReturnError(errors.New("transaction error"))

	err = repo.BatchImportContacts(context.Background(), "workspace123", testContacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")

	// Test case 3: Statement preparation error
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`).WillReturnError(errors.New("prepare error"))
	mock.ExpectRollback()

	err = repo.BatchImportContacts(context.Background(), "workspace123", testContacts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare statement")

	// Test case 4: Execution error
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`)
	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(anyArgs38...).
		WillReturnError(errors.New("execution error"))
	mock.ExpectRollback()

	err = repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{contact1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute statement")

	// Test case 5: Commit error
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO contacts`)
	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			anyArgs38...,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err = repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{contact1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")

	t.Run("should handle empty contact list", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO contacts`))
		mock.ExpectCommit()

		err := repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{})
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle JSON marshaling errors", func(t *testing.T) {
		// Create a contact with invalid JSON data
		contact := &domain.Contact{
			Email:           "test@example.com",
			ExternalID:      &domain.NullableString{String: "ext1", IsNull: false},
			Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
			Language:        &domain.NullableString{String: "en-US", IsNull: false},
			FirstName:       &domain.NullableString{String: "John", IsNull: false},
			LastName:        &domain.NullableString{String: "Doe", IsNull: false},
			Phone:           &domain.NullableString{String: "", IsNull: true},
			AddressLine1:    &domain.NullableString{String: "", IsNull: true},
			AddressLine2:    &domain.NullableString{String: "", IsNull: true},
			Country:         &domain.NullableString{String: "", IsNull: true},
			Postcode:        &domain.NullableString{String: "", IsNull: true},
			State:           &domain.NullableString{String: "", IsNull: true},
			JobTitle:        &domain.NullableString{String: "", IsNull: true},
			LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
			LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomString1:   &domain.NullableString{String: "", IsNull: true},
			CustomString2:   &domain.NullableString{String: "", IsNull: true},
			CustomString3:   &domain.NullableString{String: "", IsNull: true},
			CustomString4:   &domain.NullableString{String: "", IsNull: true},
			CustomString5:   &domain.NullableString{String: "", IsNull: true},
			CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomJSON1:     &domain.NullableJSON{Data: func() {}, IsNull: false}, // Invalid JSON data (function value)
			CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO contacts`))
		mock.ExpectRollback()

		err := repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{contact})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal CustomJSON1")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle nullable field edge cases", func(t *testing.T) {
		contact := &domain.Contact{
			Email:           "test@example.com",
			ExternalID:      &domain.NullableString{String: "", IsNull: true},
			Timezone:        &domain.NullableString{String: "", IsNull: true},
			Language:        &domain.NullableString{String: "", IsNull: true},
			FirstName:       &domain.NullableString{String: "", IsNull: true},
			LastName:        &domain.NullableString{String: "", IsNull: true},
			Phone:           &domain.NullableString{String: "", IsNull: true},
			AddressLine1:    &domain.NullableString{String: "", IsNull: true},
			AddressLine2:    &domain.NullableString{String: "", IsNull: true},
			Country:         &domain.NullableString{String: "", IsNull: true},
			Postcode:        &domain.NullableString{String: "", IsNull: true},
			State:           &domain.NullableString{String: "", IsNull: true},
			JobTitle:        &domain.NullableString{String: "", IsNull: true},
			LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
			LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomString1:   &domain.NullableString{String: "", IsNull: true},
			CustomString2:   &domain.NullableString{String: "", IsNull: true},
			CustomString3:   &domain.NullableString{String: "", IsNull: true},
			CustomString4:   &domain.NullableString{String: "", IsNull: true},
			CustomString5:   &domain.NullableString{String: "", IsNull: true},
			CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
			WithArgs(contact.Email).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectExec(`INSERT INTO contacts`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		isNew, err := repo.UpsertContact(context.Background(), "workspace123", contact)
		require.NoError(t, err)
		assert.True(t, isNew)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle database error during upsert", func(t *testing.T) {
		// Arrange
		contact := &domain.Contact{
			Email:           "test@example.com",
			ExternalID:      &domain.NullableString{String: "ext1", IsNull: false},
			Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
			Language:        &domain.NullableString{String: "en-US", IsNull: false},
			FirstName:       &domain.NullableString{String: "John", IsNull: false},
			LastName:        &domain.NullableString{String: "Doe", IsNull: false},
			Phone:           &domain.NullableString{String: "", IsNull: true},
			AddressLine1:    &domain.NullableString{String: "", IsNull: true},
			AddressLine2:    &domain.NullableString{String: "", IsNull: true},
			Country:         &domain.NullableString{String: "", IsNull: true},
			Postcode:        &domain.NullableString{String: "", IsNull: true},
			State:           &domain.NullableString{String: "", IsNull: true},
			JobTitle:        &domain.NullableString{String: "", IsNull: true},
			LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
			LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomString1:   &domain.NullableString{String: "", IsNull: true},
			CustomString2:   &domain.NullableString{String: "", IsNull: true},
			CustomString3:   &domain.NullableString{String: "", IsNull: true},
			CustomString4:   &domain.NullableString{String: "", IsNull: true},
			CustomString5:   &domain.NullableString{String: "", IsNull: true},
			CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
			WithArgs(contact.Email).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectExec(`INSERT INTO contacts`).
			WillReturnError(errors.New("database error"))

		// Act
		isNew, err := repo.UpsertContact(context.Background(), "workspace123", contact)

		// Assert
		assert.Error(t, err)
		assert.False(t, isNew)
		assert.Contains(t, err.Error(), "failed to upsert contact")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle transaction rollback", func(t *testing.T) {
		// Create a contact that will cause a database error
		contact := &domain.Contact{
			Email:      "test@example.com",
			ExternalID: &domain.NullableString{String: "ext1", IsNull: false},
			Timezone:   &domain.NullableString{String: "Europe/Paris", IsNull: false},
			Language:   &domain.NullableString{String: "en-US", IsNull: false},
			FirstName:  &domain.NullableString{String: "John", IsNull: false},
			LastName:   &domain.NullableString{String: "Doe", IsNull: false},
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		// Set up mock to fail during statement execution
		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO contacts`).WillReturnError(errors.New("prepare error"))
		mock.ExpectRollback()

		err := repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{contact})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare statement")
	})

	t.Run("should handle batch processing errors", func(t *testing.T) {
		// Create a contact that will cause a database error
		contact := &domain.Contact{
			Email:           "test@example.com",
			ExternalID:      &domain.NullableString{String: "ext1", IsNull: false},
			Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
			Language:        &domain.NullableString{String: "en-US", IsNull: false},
			FirstName:       &domain.NullableString{String: "John", IsNull: false},
			LastName:        &domain.NullableString{String: "Doe", IsNull: false},
			Phone:           &domain.NullableString{String: "", IsNull: true},
			AddressLine1:    &domain.NullableString{String: "", IsNull: true},
			AddressLine2:    &domain.NullableString{String: "", IsNull: true},
			Country:         &domain.NullableString{String: "", IsNull: true},
			Postcode:        &domain.NullableString{String: "", IsNull: true},
			State:           &domain.NullableString{String: "", IsNull: true},
			JobTitle:        &domain.NullableString{String: "", IsNull: true},
			LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
			LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomString1:   &domain.NullableString{String: "", IsNull: true},
			CustomString2:   &domain.NullableString{String: "", IsNull: true},
			CustomString3:   &domain.NullableString{String: "", IsNull: true},
			CustomString4:   &domain.NullableString{String: "", IsNull: true},
			CustomString5:   &domain.NullableString{String: "", IsNull: true},
			CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		// Set up mock to fail during statement execution
		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO contacts`))

		anyArgs38 := []driver.Value{}
		for i := 0; i < 38; i++ {
			anyArgs38 = append(anyArgs38, sqlmock.AnyArg())
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO contacts`)).
			WithArgs(anyArgs38...).
			WillReturnError(errors.New("execution error"))
		mock.ExpectRollback()

		err := repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{contact})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute statement")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle commit errors", func(t *testing.T) {
		// Create a contact
		contact := &domain.Contact{
			Email:           "test@example.com",
			ExternalID:      &domain.NullableString{String: "ext1", IsNull: false},
			Timezone:        &domain.NullableString{String: "Europe/Paris", IsNull: false},
			Language:        &domain.NullableString{String: "en-US", IsNull: false},
			FirstName:       &domain.NullableString{String: "John", IsNull: false},
			LastName:        &domain.NullableString{String: "Doe", IsNull: false},
			Phone:           &domain.NullableString{String: "", IsNull: true},
			AddressLine1:    &domain.NullableString{String: "", IsNull: true},
			AddressLine2:    &domain.NullableString{String: "", IsNull: true},
			Country:         &domain.NullableString{String: "", IsNull: true},
			Postcode:        &domain.NullableString{String: "", IsNull: true},
			State:           &domain.NullableString{String: "", IsNull: true},
			JobTitle:        &domain.NullableString{String: "", IsNull: true},
			LifetimeValue:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			OrdersCount:     &domain.NullableFloat64{Float64: 0, IsNull: true},
			LastOrderAt:     &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomString1:   &domain.NullableString{String: "", IsNull: true},
			CustomString2:   &domain.NullableString{String: "", IsNull: true},
			CustomString3:   &domain.NullableString{String: "", IsNull: true},
			CustomString4:   &domain.NullableString{String: "", IsNull: true},
			CustomString5:   &domain.NullableString{String: "", IsNull: true},
			CustomNumber1:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber2:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber3:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber4:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomNumber5:   &domain.NullableFloat64{Float64: 0, IsNull: true},
			CustomDatetime1: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime2: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime3: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime4: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomDatetime5: &domain.NullableTime{Time: time.Time{}, IsNull: true},
			CustomJSON1:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
			CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		// Set up mock to fail during commit
		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO contacts`))

		anyArgs38 := []driver.Value{}
		for i := 0; i < 38; i++ {
			anyArgs38 = append(anyArgs38, sqlmock.AnyArg())
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO contacts`)).
			WithArgs(anyArgs38...).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err := repo.BatchImportContacts(context.Background(), "workspace123", []*domain.Contact{contact})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
