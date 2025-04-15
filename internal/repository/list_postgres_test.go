package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
)

func TestListRepository(t *testing.T) {
	repo, _, mockWorkspaceRepo := setupListRepositoryTest(t)

	// Create a test list
	testList := &domain.List{
		ID:                  "list123",
		Name:                "Test List",
		IsDoubleOptin:       true,
		IsPublic:            true,
		Description:         "This is a test list",
		TotalActive:         0,
		TotalPending:        0,
		TotalUnsubscribed:   0,
		TotalBounced:        0,
		TotalComplained:     0,
		DoubleOptInTemplate: nil,
		WelcomeTemplate:     nil,
		UnsubscribeTemplate: nil,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	// Setup workspace connection mock
	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockWorkspaceRepo.EXPECT().
		GetConnection(gomock.Any(), "workspace123").
		Return(db, nil).
		AnyTimes()

	t.Run("CreateList", func(t *testing.T) {
		t.Run("successful creation", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`
				INSERT INTO lists (id, name, is_double_optin, is_public, description, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`)).WithArgs(
				testList.ID,
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.CreateList(context.Background(), "workspace123", testList)
			require.NoError(t, err)
		})

		t.Run("database error", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`
				INSERT INTO lists (id, name, is_double_optin, is_public, description, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`)).WithArgs(
				testList.ID,
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).WillReturnError(errors.New("database error"))

			err := repo.CreateList(context.Background(), "workspace123", testList)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create list")
		})
	})

	t.Run("GetListByID", func(t *testing.T) {
		t.Run("list found", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{
				"id", "name", "is_double_optin", "is_public", "description", "total_active", "total_pending",
				"total_unsubscribed", "total_bounced", "total_complained", "double_optin_template",
				"welcome_template", "unsubscribe_template", "created_at", "updated_at",
				"deleted_at",
			}).AddRow(
				testList.ID,
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				testList.TotalActive,
				testList.TotalPending,
				testList.TotalUnsubscribed,
				testList.TotalBounced,
				testList.TotalComplained,
				testList.DoubleOptInTemplate,
				testList.WelcomeTemplate,
				testList.UnsubscribeTemplate,
				testList.CreatedAt,
				testList.UpdatedAt,
				nil,
			)

			sqlMock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
				total_unsubscribed, total_bounced, total_complained, double_optin_template, 
				welcome_template, unsubscribe_template, created_at, updated_at
				FROM lists
				WHERE id = $1 AND deleted_at IS NULL
			`)).WithArgs(testList.ID).WillReturnRows(rows)

			list, err := repo.GetListByID(context.Background(), "workspace123", testList.ID)
			require.NoError(t, err)
			assert.Equal(t, testList.ID, list.ID)
			assert.Equal(t, testList.Name, list.Name)
			assert.Equal(t, testList.IsDoubleOptin, list.IsDoubleOptin)
			assert.Equal(t, testList.IsPublic, list.IsPublic)
			assert.Equal(t, testList.Description, list.Description)
		})

		t.Run("list not found", func(t *testing.T) {
			sqlMock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
				total_unsubscribed, total_bounced, total_complained, double_optin_template, 
				welcome_template, unsubscribe_template, created_at, updated_at
				FROM lists
				WHERE id = $1 AND deleted_at IS NULL
			`)).WithArgs(testList.ID).WillReturnError(sql.ErrNoRows)

			list, err := repo.GetListByID(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.Nil(t, list)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
		})

		t.Run("database error", func(t *testing.T) {
			sqlMock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
				total_unsubscribed, total_bounced, total_complained, double_optin_template, 
				welcome_template, unsubscribe_template, created_at, updated_at
				FROM lists
				WHERE id = $1 AND deleted_at IS NULL
			`)).WithArgs(testList.ID).WillReturnError(errors.New("database error"))

			list, err := repo.GetListByID(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.Nil(t, list)
			assert.Contains(t, err.Error(), "failed to get list")
		})
	})

	t.Run("GetLists", func(t *testing.T) {
		t.Run("successful retrieval", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{
				"id", "name", "is_double_optin", "is_public", "description", "total_active", "total_pending",
				"total_unsubscribed", "total_bounced", "total_complained", "double_optin_template",
				"welcome_template", "unsubscribe_template", "created_at", "updated_at",
				"deleted_at",
			}).AddRow(
				testList.ID,
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				testList.TotalActive,
				testList.TotalPending,
				testList.TotalUnsubscribed,
				testList.TotalBounced,
				testList.TotalComplained,
				testList.DoubleOptInTemplate,
				testList.WelcomeTemplate,
				testList.UnsubscribeTemplate,
				testList.CreatedAt,
				testList.UpdatedAt,
				nil,
			)

			sqlMock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
				total_unsubscribed, total_bounced, total_complained, double_optin_template, 
				welcome_template, unsubscribe_template, created_at, updated_at
				FROM lists
				WHERE deleted_at IS NULL
				ORDER BY created_at DESC
			`)).WillReturnRows(rows)

			lists, err := repo.GetLists(context.Background(), "workspace123")
			require.NoError(t, err)
			require.Len(t, lists, 1)
			assert.Equal(t, testList.ID, lists[0].ID)
			assert.Equal(t, testList.Name, lists[0].Name)
			assert.Equal(t, testList.IsDoubleOptin, lists[0].IsDoubleOptin)
			assert.Equal(t, testList.IsPublic, lists[0].IsPublic)
			assert.Equal(t, testList.Description, lists[0].Description)
		})

		t.Run("database error", func(t *testing.T) {
			sqlMock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, is_double_optin, is_public, description, total_active, total_pending, 
				total_unsubscribed, total_bounced, total_complained, double_optin_template, 
				welcome_template, unsubscribe_template, created_at, updated_at
				FROM lists
				WHERE deleted_at IS NULL
				ORDER BY created_at DESC
			`)).WillReturnError(errors.New("database error"))

			lists, err := repo.GetLists(context.Background(), "workspace123")
			require.Error(t, err)
			assert.Nil(t, lists)
			assert.Contains(t, err.Error(), "failed to get lists")
		})
	})

	t.Run("UpdateList", func(t *testing.T) {
		t.Run("successful update", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`
				UPDATE lists
				SET name = $1, is_double_optin = $2, is_public = $3, description = $4, updated_at = $5
				WHERE id = $6 AND deleted_at IS NULL
			`)).WithArgs(
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				testList.ID,
			).WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.UpdateList(context.Background(), "workspace123", testList)
			require.NoError(t, err)
		})

		t.Run("list not found", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`
				UPDATE lists
				SET name = $1, is_double_optin = $2, is_public = $3, description = $4, updated_at = $5
				WHERE id = $6 AND deleted_at IS NULL
			`)).WithArgs(
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				testList.ID,
			).WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.UpdateList(context.Background(), "workspace123", testList)
			require.Error(t, err)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
		})

		t.Run("database error", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`
				UPDATE lists
				SET name = $1, is_double_optin = $2, is_public = $3, description = $4, updated_at = $5
				WHERE id = $6 AND deleted_at IS NULL
			`)).WithArgs(
				testList.Name,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				testList.ID,
			).WillReturnError(errors.New("database error"))

			err := repo.UpdateList(context.Background(), "workspace123", testList)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to update list")
		})
	})

	t.Run("DeleteList", func(t *testing.T) {
		t.Run("successful deletion", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE lists SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
				WithArgs(sqlmock.AnyArg(), testList.ID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.NoError(t, err)
		})

		t.Run("list not found", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE lists SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
				WithArgs(sqlmock.AnyArg(), testList.ID).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
		})

		t.Run("database error", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE lists SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
				WithArgs(sqlmock.AnyArg(), testList.ID).
				WillReturnError(errors.New("database error"))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to soft delete list")
		})

		t.Run("list already deleted", func(t *testing.T) {
			sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE lists SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
				WithArgs(sqlmock.AnyArg(), testList.ID).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
			assert.Contains(t, err.Error(), "list not found or already deleted")
		})
	})
}

func setupListRepositoryTest(t *testing.T) (*listRepository, sqlmock.Sqlmock, *mocks.MockWorkspaceRepository) {
	ctrl := gomock.NewController(t)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	repo := NewListRepository(mockWorkspaceRepo).(*listRepository)
	return repo, nil, mockWorkspaceRepo
}

func TestListRepository_IncrementTotal_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewListRepository(mockWorkspaceRepo)

	err := repo.IncrementTotal(context.Background(), "workspace123", "list123", "invalid_type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid total type")
}

func TestListRepository_DecrementTotal_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	repo := NewListRepository(mockWorkspaceRepo)

	err := repo.DecrementTotal(context.Background(), "workspace123", "list123", "invalid_type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid total type")
}

func TestListRepository_IncrementTotal(t *testing.T) {
	tests := []struct {
		name      string
		totalType domain.ContactListTotalType
		column    string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "increment active total",
			totalType: domain.TotalTypeActive,
			column:    "total_active",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_active = total_active \\+ 1 WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "increment pending total",
			totalType: domain.TotalTypePending,
			column:    "total_pending",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_pending = total_pending \\+ 1 WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "increment unsubscribed total",
			totalType: domain.TotalTypeUnsubscribed,
			column:    "total_unsubscribed",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_unsubscribed = total_unsubscribed \\+ 1 WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "increment bounced total",
			totalType: domain.TotalTypeBounced,
			column:    "total_bounced",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_bounced = total_bounced \\+ 1 WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "increment complained total",
			totalType: domain.TotalTypeComplained,
			column:    "total_complained",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_complained = total_complained \\+ 1 WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "workspace connection error",
			totalType: domain.TotalTypeActive,
			mockSetup: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
			repo := NewListRepository(mockWorkspaceRepo)

			if tt.mockSetup != nil {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				tt.mockSetup(mock)

				mockWorkspaceRepo.EXPECT().
					GetConnection(gomock.Any(), "workspace123").
					Return(db, nil)
			} else {
				mockWorkspaceRepo.EXPECT().
					GetConnection(gomock.Any(), "workspace123").
					Return(nil, fmt.Errorf("workspace connection error"))
			}

			err := repo.IncrementTotal(context.Background(), "workspace123", "list123", tt.totalType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListRepository_DecrementTotal(t *testing.T) {
	tests := []struct {
		name      string
		totalType domain.ContactListTotalType
		column    string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "decrement active total",
			totalType: domain.TotalTypeActive,
			column:    "total_active",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_active = GREATEST\\(total_active - 1, 0\\) WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "decrement pending total",
			totalType: domain.TotalTypePending,
			column:    "total_pending",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_pending = GREATEST\\(total_pending - 1, 0\\) WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "decrement unsubscribed total",
			totalType: domain.TotalTypeUnsubscribed,
			column:    "total_unsubscribed",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_unsubscribed = GREATEST\\(total_unsubscribed - 1, 0\\) WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "decrement bounced total",
			totalType: domain.TotalTypeBounced,
			column:    "total_bounced",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_bounced = GREATEST\\(total_bounced - 1, 0\\) WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "decrement complained total",
			totalType: domain.TotalTypeComplained,
			column:    "total_complained",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE lists SET total_complained = GREATEST\\(total_complained - 1, 0\\) WHERE id = \\$1 AND deleted_at IS NULL").
					WithArgs("list123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "workspace connection error",
			totalType: domain.TotalTypeActive,
			mockSetup: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
			repo := NewListRepository(mockWorkspaceRepo)

			if tt.mockSetup != nil {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				tt.mockSetup(mock)

				mockWorkspaceRepo.EXPECT().
					GetConnection(gomock.Any(), "workspace123").
					Return(db, nil)
			} else {
				mockWorkspaceRepo.EXPECT().
					GetConnection(gomock.Any(), "workspace123").
					Return(nil, fmt.Errorf("workspace connection error"))
			}

			err := repo.DecrementTotal(context.Background(), "workspace123", "list123", tt.totalType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
