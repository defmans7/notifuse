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

func TestUpsertContact(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
	repo := NewContactRepository(workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "test@example.com"
	workspaceID := "workspace123"

	testContact := &domain.Contact{
		Email:           email,
		ExternalID:      &domain.NullableString{String: "ext123", IsNull: false},
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

	// Test case 1: Insert new contact
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			testContact.Email,
			testContact.ExternalID.String,
			testContact.Timezone.String,
			testContact.Language.String,
			testContact.FirstName.String,
			testContact.LastName.String,
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			testContact.CreatedAt,
			testContact.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err := repo.UpsertContact(context.Background(), workspaceID, testContact)
	require.NoError(t, err)
	assert.True(t, created)

	// Test case 2: Update existing contact with only some fields
	partialContact := &domain.Contact{
		Email:     "partial@example.com",
		FirstName: &domain.NullableString{String: "Jane", IsNull: false},
		LastName:  &domain.NullableString{String: "Smith", IsNull: false},
		UpdatedAt: now,
	}

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
	}).AddRow(
		"partial@example.com", "old-ext-id", "UTC", "en-US",
		"Old", "Name", "", "", "",
		"", "", "", "",
		0.0, 0.0, time.Time{},
		"", "", "", "", "",
		0.0, 0.0, 0.0, 0.0, 0.0,
		time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{},
		"", "", "", "", "",
		now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("partial@example.com").
		WillReturnRows(rows)

	// Expect only the fields that are present in partialContact to be updated
	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			partialContact.Email,
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			partialContact.FirstName.String,
			partialContact.LastName.String,
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			partialContact.CreatedAt,
			partialContact.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err = repo.UpsertContact(context.Background(), workspaceID, partialContact)
	require.NoError(t, err)
	assert.False(t, created)

	// Test case 3: Error checking if contact exists
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(errors.New("check error"))

	created, err = repo.UpsertContact(context.Background(), workspaceID, testContact)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if contact exists")

	// Test case 4: Error upserting contact
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WillReturnError(errors.New("upsert error"))

	created, err = repo.UpsertContact(context.Background(), workspaceID, testContact)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert contact")

	// Test case 5: Contact with invalid JSON data
	contactWithInvalidJSON := &domain.Contact{
		Email:           "invalidjson@example.com",
		ExternalID:      &domain.NullableString{String: "ext789", IsNull: false},
		Timezone:        &domain.NullableString{String: "UTC", IsNull: false},
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
		CustomJSON1:     &domain.NullableJSON{Data: make(chan int), IsNull: false}, // Invalid JSON data
		CustomJSON2:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON3:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON4:     &domain.NullableJSON{Data: nil, IsNull: true},
		CustomJSON5:     &domain.NullableJSON{Data: nil, IsNull: true},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs("invalidjson@example.com").
		WillReturnError(sql.ErrNoRows)

	created, err = repo.UpsertContact(context.Background(), workspaceID, contactWithInvalidJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal CustomJSON1")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertContactWithOnlyEmail(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	workspaceRepo.AddWorkspaceDB("workspace123", db)
	repo := NewContactRepository(workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "minimal@example.com"
	workspaceID := "workspace123"

	minimalContact := &domain.Contact{
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test case: Insert new contact with only email
	mock.ExpectQuery(`SELECT (.+) FROM contacts WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO contacts`).
		WithArgs(
			minimalContact.Email,
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullFloat64{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullTime{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			sql.NullString{Valid: false},
			minimalContact.CreatedAt,
			minimalContact.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err := repo.UpsertContact(context.Background(), workspaceID, minimalContact)
	require.NoError(t, err)
	assert.True(t, created)

	assert.NoError(t, mock.ExpectationsWereMet())
}
