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
)

func TestAddContactToList(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactListRepository(db)

	// Test case 1: Successful add contact to list
	contactList := &domain.ContactList{
		ContactID: "contact-123",
		ListID:    "list-123",
		Status:    domain.ContactListStatusActive,
	}

	mock.ExpectExec(`INSERT INTO contact_lists`).
		WithArgs(
			contactList.ContactID, contactList.ListID, contactList.Status,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddContactToList(context.Background(), contactList)
	require.NoError(t, err)

	// Test case 2: Error during adding contact to list
	contactListWithError := &domain.ContactList{
		ContactID: "contact-error",
		ListID:    "list-error",
		Status:    domain.ContactListStatusActive,
	}

	mock.ExpectExec(`INSERT INTO contact_lists`).
		WithArgs(
			contactListWithError.ContactID, contactListWithError.ListID, contactListWithError.Status,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(errors.New("database error"))

	err = repo.AddContactToList(context.Background(), contactListWithError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add contact to list")
}

func TestGetContactListByIDs(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactListRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	contactID := "contact-123"
	listID := "list-123"

	// Test case 1: Contact list found
	rows := sqlmock.NewRows([]string{
		"contact_id", "list_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			contactID, listID, domain.ContactListStatusActive, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE contact_id = \$1 AND list_id = \$2`).
		WithArgs(contactID, listID).
		WillReturnRows(rows)

	contactList, err := repo.GetContactListByIDs(context.Background(), contactID, listID)
	require.NoError(t, err)
	assert.Equal(t, contactID, contactList.ContactID)
	assert.Equal(t, listID, contactList.ListID)
	assert.Equal(t, domain.ContactListStatusActive, contactList.Status)

	// Test case 2: Contact list not found
	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE contact_id = \$1 AND list_id = \$2`).
		WithArgs("non-existent-contact", "non-existent-list").
		WillReturnError(sql.ErrNoRows)

	contactList, err = repo.GetContactListByIDs(context.Background(), "non-existent-contact", "non-existent-list")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactListNotFound{}, err)
	assert.Nil(t, contactList)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE contact_id = \$1 AND list_id = \$2`).
		WithArgs("error-contact", "error-list").
		WillReturnError(errors.New("database error"))

	contactList, err = repo.GetContactListByIDs(context.Background(), "error-contact", "error-list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get contact list")
	assert.Nil(t, contactList)
}

func TestGetContactsByListID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactListRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	listID := "list-123"

	// Test case 1: Multiple contact lists found
	rows := sqlmock.NewRows([]string{
		"contact_id", "list_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			"contact-1", listID, domain.ContactListStatusActive, now, now,
		).
		AddRow(
			"contact-2", listID, domain.ContactListStatusUnsubscribed, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE list_id = \$1 ORDER BY created_at DESC`).
		WithArgs(listID).
		WillReturnRows(rows)

	contactLists, err := repo.GetContactsByListID(context.Background(), listID)
	require.NoError(t, err)
	assert.Len(t, contactLists, 2)
	assert.Equal(t, "contact-1", contactLists[0].ContactID)
	assert.Equal(t, "contact-2", contactLists[1].ContactID)
	assert.Equal(t, domain.ContactListStatusActive, contactLists[0].Status)
	assert.Equal(t, domain.ContactListStatusUnsubscribed, contactLists[1].Status)

	// Test case 2: No contact lists found (empty result)
	emptyRows := sqlmock.NewRows([]string{
		"contact_id", "list_id", "status", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE list_id = \$1 ORDER BY created_at DESC`).
		WithArgs("empty-list").
		WillReturnRows(emptyRows)

	contactLists, err = repo.GetContactsByListID(context.Background(), "empty-list")
	require.NoError(t, err)
	assert.Empty(t, contactLists)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE list_id = \$1 ORDER BY created_at DESC`).
		WithArgs("error-list").
		WillReturnError(errors.New("database error"))

	contactLists, err = repo.GetContactsByListID(context.Background(), "error-list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get contacts for list")
	assert.Nil(t, contactLists)
}

func TestGetListsByContactID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactListRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	contactID := "contact-123"

	// Test case 1: Multiple contact lists found
	rows := sqlmock.NewRows([]string{
		"contact_id", "list_id", "status", "created_at", "updated_at",
	}).
		AddRow(
			contactID, "list-1", domain.ContactListStatusActive, now, now,
		).
		AddRow(
			contactID, "list-2", domain.ContactListStatusUnsubscribed, now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE contact_id = \$1 ORDER BY created_at DESC`).
		WithArgs(contactID).
		WillReturnRows(rows)

	contactLists, err := repo.GetListsByContactID(context.Background(), contactID)
	require.NoError(t, err)
	assert.Len(t, contactLists, 2)
	assert.Equal(t, "list-1", contactLists[0].ListID)
	assert.Equal(t, "list-2", contactLists[1].ListID)
	assert.Equal(t, domain.ContactListStatusActive, contactLists[0].Status)
	assert.Equal(t, domain.ContactListStatusUnsubscribed, contactLists[1].Status)

	// Test case 2: No contact lists found (empty result)
	emptyRows := sqlmock.NewRows([]string{
		"contact_id", "list_id", "status", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE contact_id = \$1 ORDER BY created_at DESC`).
		WithArgs("empty-contact").
		WillReturnRows(emptyRows)

	contactLists, err = repo.GetListsByContactID(context.Background(), "empty-contact")
	require.NoError(t, err)
	assert.Empty(t, contactLists)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM contact_lists WHERE contact_id = \$1 ORDER BY created_at DESC`).
		WithArgs("error-contact").
		WillReturnError(errors.New("database error"))

	contactLists, err = repo.GetListsByContactID(context.Background(), "error-contact")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get lists for contact")
	assert.Nil(t, contactLists)
}

func TestUpdateContactListStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactListRepository(db)
	contactID := "contact-123"
	listID := "list-123"
	newStatus := domain.ContactListStatusUnsubscribed

	// Test case 1: Successful update
	mock.ExpectExec(`UPDATE contact_lists SET status = \$1, updated_at = \$2 WHERE contact_id = \$3 AND list_id = \$4`).
		WithArgs(newStatus, sqlmock.AnyArg(), contactID, listID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateContactListStatus(context.Background(), contactID, listID, newStatus)
	require.NoError(t, err)

	// Test case 2: Contact list not found
	mock.ExpectExec(`UPDATE contact_lists SET status = \$1, updated_at = \$2 WHERE contact_id = \$3 AND list_id = \$4`).
		WithArgs(newStatus, sqlmock.AnyArg(), "non-existent-contact", "non-existent-list").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateContactListStatus(context.Background(), "non-existent-contact", "non-existent-list", newStatus)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactListNotFound{}, err)

	// Test case 3: Database error
	mock.ExpectExec(`UPDATE contact_lists SET status = \$1, updated_at = \$2 WHERE contact_id = \$3 AND list_id = \$4`).
		WithArgs(newStatus, sqlmock.AnyArg(), "error-contact", "error-list").
		WillReturnError(errors.New("database error"))

	err = repo.UpdateContactListStatus(context.Background(), "error-contact", "error-list", newStatus)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update contact list status")
}

func TestRemoveContactFromList(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewContactListRepository(db)
	contactID := "contact-123"
	listID := "list-123"

	// Test case 1: Successful removal
	mock.ExpectExec(`DELETE FROM contact_lists WHERE contact_id = \$1 AND list_id = \$2`).
		WithArgs(contactID, listID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveContactFromList(context.Background(), contactID, listID)
	require.NoError(t, err)

	// Test case 2: Contact list not found
	mock.ExpectExec(`DELETE FROM contact_lists WHERE contact_id = \$1 AND list_id = \$2`).
		WithArgs("non-existent-contact", "non-existent-list").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.RemoveContactFromList(context.Background(), "non-existent-contact", "non-existent-list")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrContactListNotFound{}, err)

	// Test case 3: Database error
	mock.ExpectExec(`DELETE FROM contact_lists WHERE contact_id = \$1 AND list_id = \$2`).
		WithArgs("error-contact", "error-list").
		WillReturnError(errors.New("database error"))

	err = repo.RemoveContactFromList(context.Background(), "error-contact", "error-list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove contact from list")
}
