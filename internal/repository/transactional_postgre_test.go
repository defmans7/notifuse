package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/repository"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTransactionalRepositoryTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *mocks.MockWorkspaceRepository, *repository.TransactionalNotificationRepository, *pkgmocks.MockLogger) {
	// Create SQL mock
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Create mock controller for workspace repository
	ctrl := gomock.NewController(t)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	logger := pkgmocks.NewMockLogger(ctrl)

	// Setup common logger expectations
	logger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithFields(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()
	logger.EXPECT().Warn(gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create repository instance
	repo := repository.NewTransactionalNotificationRepository(db, workspaceRepo)

	return db, mock, workspaceRepo, repo, logger
}

func createTestNotification() *domain.TransactionalNotification {
	now := time.Now().UTC()
	return &domain.TransactionalNotification{
		ID:          "test-notification",
		Name:        "Test Notification",
		Description: "Test description",
		Channels: domain.ChannelTemplates{
			domain.TransactionalChannelEmail: domain.ChannelTemplate{
				TemplateID: "template-123",
				Version:    1,
				Settings: domain.MapOfAny{
					"subject": "Test Subject",
				},
			},
		},
		Status:    domain.TransactionalStatusActive,
		IsPublic:  false,
		Metadata:  domain.MapOfAny{"category": "test"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestTransactionalNotificationRepository_Create(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification := createTestNotification()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected prepared statement and query execution
	mockWorkspaceDB.ExpectExec("INSERT INTO transactional_notifications").
		WithArgs(
			notification.ID,
			notification.Name,
			notification.Description,
			notification.Channels,
			notification.Status,
			notification.IsPublic,
			notification.Metadata,
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the method under test
	err = repo.Create(context.Background(), workspace, notification)

	// Assertions
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Create_WorkspaceError(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	notification := createTestNotification()

	// Setup mock for workspace connection error
	expectedErr := errors.New("connection error")
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(nil, expectedErr)

	// Execute the method under test
	err := repo.Create(context.Background(), workspace, notification)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get workspace db")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Create_QueryError(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification := createTestNotification()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected prepared statement with execution error
	dbErr := errors.New("database error")
	mockWorkspaceDB.ExpectExec("INSERT INTO transactional_notifications").
		WithArgs(
			notification.ID,
			notification.Name,
			notification.Description,
			notification.Channels,
			notification.Status,
			notification.IsPublic,
			notification.Metadata,
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
		).
		WillReturnError(dbErr)

	// Execute the method under test
	err = repo.Create(context.Background(), workspace, notification)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create transactional notification")
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Update(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification := createTestNotification()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected prepared statement and query execution
	mockWorkspaceDB.ExpectExec("UPDATE transactional_notifications").
		WithArgs(
			notification.Name,
			notification.Description,
			notification.Channels,
			notification.Status,
			notification.IsPublic,
			notification.Metadata,
			sqlmock.AnyArg(), // UpdatedAt
			notification.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the method under test
	err = repo.Update(context.Background(), workspace, notification)

	// Assertions
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Update_NotFound(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification := createTestNotification()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected prepared statement with no rows affected
	mockWorkspaceDB.ExpectExec("UPDATE transactional_notifications").
		WithArgs(
			notification.Name,
			notification.Description,
			notification.Channels,
			notification.Status,
			notification.IsPublic,
			notification.Metadata,
			sqlmock.AnyArg(), // UpdatedAt
			notification.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute the method under test
	err = repo.Update(context.Background(), workspace, notification)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Get(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	id := "test-notification"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification := createTestNotification()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected query execution
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "channels", "status", "is_public", "metadata", "created_at", "updated_at", "deleted_at",
	}).
		AddRow(
			notification.ID,
			notification.Name,
			notification.Description,
			notification.Channels,
			notification.Status,
			notification.IsPublic,
			notification.Metadata,
			notification.CreatedAt,
			notification.UpdatedAt,
			notification.DeletedAt,
		)

	mockWorkspaceDB.ExpectQuery("SELECT (.+) FROM transactional_notifications").
		WithArgs(id).
		WillReturnRows(rows)

	// Execute the method under test
	result, err := repo.Get(context.Background(), workspace, id)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, notification.ID, result.ID)
	assert.Equal(t, notification.Name, result.Name)
	assert.Equal(t, notification.Description, result.Description)
	assert.Equal(t, notification.Status, result.Status)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Get_NotFound(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	id := "non-existent"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected query execution with no rows
	mockWorkspaceDB.ExpectQuery("SELECT (.+) FROM transactional_notifications").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	// Execute the method under test
	result, err := repo.Get(context.Background(), workspace, id)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_List(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	filter := map[string]interface{}{"status": domain.TransactionalStatusActive}
	limit := 10
	offset := 0

	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification1 := createTestNotification()
	notification2 := createTestNotification()
	notification2.ID = "test-notification-2"
	notification2.Name = "Test Notification 2"

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Mock the count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mockWorkspaceDB.ExpectQuery("SELECT COUNT\\(\\*\\) FROM transactional_notifications").
		WithArgs(domain.TransactionalStatusActive).
		WillReturnRows(countRows)

	// Mock the list query
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "channels", "status", "is_public", "metadata", "created_at", "updated_at", "deleted_at",
	}).
		AddRow(
			notification1.ID,
			notification1.Name,
			notification1.Description,
			notification1.Channels,
			notification1.Status,
			notification1.IsPublic,
			notification1.Metadata,
			notification1.CreatedAt,
			notification1.UpdatedAt,
			notification1.DeletedAt,
		).
		AddRow(
			notification2.ID,
			notification2.Name,
			notification2.Description,
			notification2.Channels,
			notification2.Status,
			notification2.IsPublic,
			notification2.Metadata,
			notification2.CreatedAt,
			notification2.UpdatedAt,
			notification2.DeletedAt,
		)

	mockWorkspaceDB.ExpectQuery("SELECT (.+) FROM transactional_notifications").
		WithArgs(domain.TransactionalStatusActive).
		WillReturnRows(rows)

	// Execute the method under test
	results, count, err := repo.List(context.Background(), workspace, filter, limit, offset)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, notification1.ID, results[0].ID)
	assert.Equal(t, notification2.ID, results[1].ID)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_List_WithSearch(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	searchTerm := "test"
	filter := map[string]interface{}{
		"search": searchTerm,
		"status": domain.TransactionalStatusActive,
	}
	limit := 10
	offset := 0

	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	notification := createTestNotification()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Mock the count query with search pattern
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mockWorkspaceDB.ExpectQuery("SELECT COUNT\\(\\*\\) FROM transactional_notifications").
		WithArgs(domain.TransactionalStatusActive, "%"+searchTerm+"%").
		WillReturnRows(countRows)

	// Mock the list query with search pattern
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "channels", "status", "is_public", "metadata", "created_at", "updated_at", "deleted_at",
	}).
		AddRow(
			notification.ID,
			notification.Name,
			notification.Description,
			notification.Channels,
			notification.Status,
			notification.IsPublic,
			notification.Metadata,
			notification.CreatedAt,
			notification.UpdatedAt,
			notification.DeletedAt,
		)

	mockWorkspaceDB.ExpectQuery("SELECT (.+) FROM transactional_notifications").
		WithArgs(domain.TransactionalStatusActive, "%"+searchTerm+"%").
		WillReturnRows(rows)

	// Execute the method under test
	results, count, err := repo.List(context.Background(), workspace, filter, limit, offset)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, notification.ID, results[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Delete(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	id := "test-notification"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected prepared statement and query execution
	mockWorkspaceDB.ExpectExec("UPDATE transactional_notifications").
		WithArgs(
			sqlmock.AnyArg(), // DeletedAt timestamp
			id,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the method under test
	err = repo.Delete(context.Background(), workspace, id)

	// Assertions
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}

func TestTransactionalNotificationRepository_Delete_NotFound(t *testing.T) {
	db, mock, workspaceRepo, repo, _ := setupTransactionalRepositoryTest(t)
	defer db.Close()

	workspace := "test-workspace"
	id := "non-existent"
	workspaceDB, mockWorkspaceDB, err := sqlmock.New()
	require.NoError(t, err)
	defer workspaceDB.Close()

	// Setup mock for workspace connection
	workspaceRepo.EXPECT().
		GetConnection(gomock.Any(), workspace).
		Return(workspaceDB, nil)

	// Expected prepared statement with no rows affected
	mockWorkspaceDB.ExpectExec("UPDATE transactional_notifications").
		WithArgs(
			sqlmock.AnyArg(), // DeletedAt timestamp
			id,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute the method under test
	err = repo.Delete(context.Background(), workspace, id)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, mockWorkspaceDB.ExpectationsWereMet())
}
