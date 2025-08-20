package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetryRepository_GetLastMessageAt(t *testing.T) {
	t.Run("returns last message timestamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock workspace repository
		mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		repo := NewTelemetryRepository(mockWorkspaceRepo)

		// Expected query uses ORDER BY with LIMIT for performance
		expectedQuery := `SELECT created_at FROM message_history 
			  WHERE created_at IS NOT NULL 
			  ORDER BY created_at DESC, id DESC 
			  LIMIT 1`
		
		expectedTime := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
		mock.ExpectQuery(expectedQuery).
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).
				AddRow(expectedTime))

		result, err := repo.GetLastMessageAt(context.Background(), db)

		require.NoError(t, err)
		assert.Equal(t, expectedTime.Format(time.RFC3339), result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty string when no messages found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		repo := NewTelemetryRepository(mockWorkspaceRepo)

		expectedQuery := `SELECT created_at FROM message_history 
			  WHERE created_at IS NOT NULL 
			  ORDER BY created_at DESC, id DESC 
			  LIMIT 1`
		
		mock.ExpectQuery(expectedQuery).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetLastMessageAt(context.Background(), db)

		require.NoError(t, err)
		assert.Equal(t, "", result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles database errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
		repo := NewTelemetryRepository(mockWorkspaceRepo)

		expectedQuery := `SELECT created_at FROM message_history 
			  WHERE created_at IS NOT NULL 
			  ORDER BY created_at DESC, id DESC 
			  LIMIT 1`
		
		mock.ExpectQuery(expectedQuery).
			WillReturnError(assert.AnError)

		result, err := repo.GetLastMessageAt(context.Background(), db)

		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "failed to get last message timestamp")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
