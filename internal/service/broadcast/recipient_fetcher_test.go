package broadcast

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	bmocks "github.com/Notifuse/notifuse/internal/service/broadcast/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRecipientFetcher_GetTotalRecipientCount_Success(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	audienceSettings := domain.AudienceSettings{
		Lists: []string{"list1", "list2"},
	}

	// Create mock broadcast
	broadcast := &domain.Broadcast{
		ID:       broadcastID,
		Audience: audienceSettings,
	}

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockContactRepo.EXPECT().
		CountContactsForBroadcast(ctx, workspaceID, audienceSettings).
		Return(100, nil)

	// Create recipient fetcher with mocks
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	count, err := fetcher.GetTotalRecipientCount(ctx, workspaceID, broadcastID)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 100, count)
}

func TestRecipientFetcher_GetTotalRecipientCount_BroadcastNotFound(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"

	// Create expected error
	expectedError := errors.New("broadcast not found")

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(nil, expectedError)

	// Create recipient fetcher with mocks
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	count, err := fetcher.GetTotalRecipientCount(ctx, workspaceID, broadcastID)

	// Verify results
	assert.Error(t, err)
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok, "Error should be of type BroadcastError")
	assert.Equal(t, ErrCodeBroadcastNotFound, broadcastErr.Code)
	assert.Equal(t, 0, count)
}

func TestRecipientFetcher_GetTotalRecipientCount_CountError(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	audienceSettings := domain.AudienceSettings{
		Lists: []string{"list1", "list2"},
	}

	// Create mock broadcast
	broadcast := &domain.Broadcast{
		ID:       broadcastID,
		Audience: audienceSettings,
	}

	// Create expected error
	expectedError := errors.New("failed to count contacts")

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockContactRepo.EXPECT().
		CountContactsForBroadcast(ctx, workspaceID, audienceSettings).
		Return(0, expectedError)

	// Create recipient fetcher with mocks
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	count, err := fetcher.GetTotalRecipientCount(ctx, workspaceID, broadcastID)

	// Verify results
	assert.Error(t, err)
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok, "Error should be of type BroadcastError")
	assert.Equal(t, ErrCodeRecipientFetch, broadcastErr.Code)
	assert.Equal(t, 0, count)
}

func TestRecipientFetcher_FetchBatch_Success(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	offset := 0
	limit := 10
	audienceSettings := domain.AudienceSettings{
		Lists: []string{"list1", "list2"},
	}

	// Create mock broadcast
	broadcast := &domain.Broadcast{
		ID:       broadcastID,
		Audience: audienceSettings,
	}

	// Create mock contacts
	contacts := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "contact1@example.com",
			},
			ListID:   "list1",
			ListName: "List 1",
		},
		{
			Contact: &domain.Contact{
				Email: "contact2@example.com",
			},
			ListID:   "list2",
			ListName: "List 2",
		},
	}

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockContactRepo.EXPECT().
		GetContactsForBroadcast(ctx, workspaceID, audienceSettings, limit, offset).
		Return(contacts, nil)

	// Create recipient fetcher with mocks
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	result, err := fetcher.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, contacts, result)
	assert.Len(t, result, 2)
}

func TestRecipientFetcher_FetchBatch_BroadcastNotFound(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	offset := 0
	limit := 10

	// Create expected error
	expectedError := errors.New("broadcast not found")

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(nil, expectedError)

	// Create recipient fetcher with mocks
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	result, err := fetcher.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)

	// Verify results
	assert.Error(t, err)
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok, "Error should be of type BroadcastError")
	assert.Equal(t, ErrCodeBroadcastNotFound, broadcastErr.Code)
	assert.Nil(t, result)
}

func TestRecipientFetcher_FetchBatch_FetchError(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	offset := 0
	limit := 10
	audienceSettings := domain.AudienceSettings{
		Lists: []string{"list1", "list2"},
	}

	// Create mock broadcast
	broadcast := &domain.Broadcast{
		ID:       broadcastID,
		Audience: audienceSettings,
	}

	// Create expected error
	expectedError := errors.New("failed to fetch contacts")

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockContactRepo.EXPECT().
		GetContactsForBroadcast(ctx, workspaceID, audienceSettings, limit, offset).
		Return(nil, expectedError)

	// Create recipient fetcher with mocks
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		nil, // Use default config
	)

	// Call the method being tested
	result, err := fetcher.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)

	// Verify results
	assert.Error(t, err)
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok, "Error should be of type BroadcastError")
	assert.Equal(t, ErrCodeRecipientFetch, broadcastErr.Code)
	assert.Nil(t, result)
}

func TestRecipientFetcher_FetchBatch_DefaultLimit(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger mock to return itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	offset := 0
	limit := 0          // Zero limit should use default from config
	defaultLimit := 100 // Default from config
	audienceSettings := domain.AudienceSettings{
		Lists: []string{"list1", "list2"},
	}

	// Create mock broadcast
	broadcast := &domain.Broadcast{
		ID:       broadcastID,
		Audience: audienceSettings,
	}

	// Create mock config with custom fetch batch size
	config := DefaultConfig()
	config.FetchBatchSize = defaultLimit

	// Create mock contacts
	contacts := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "contact1@example.com",
			},
			ListID:   "list1",
			ListName: "List 1",
		},
	}

	// Set up expectations
	mockBroadcastService.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockContactRepo.EXPECT().
		GetContactsForBroadcast(ctx, workspaceID, audienceSettings, defaultLimit, offset).
		Return(contacts, nil)

	// Create recipient fetcher with mocks and custom config
	fetcher := NewRecipientFetcher(
		mockBroadcastService,
		mockContactRepo,
		mockLogger,
		config,
	)

	// Call the method being tested
	result, err := fetcher.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, contacts, result)
}

func TestRecipientFetcher_WithMock(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock recipient fetcher
	mockFetcher := bmocks.NewMockRecipientFetcher(ctrl)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	offset := 0
	limit := 10

	// Create mock contacts
	contacts := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "contact1@example.com",
			},
			ListID: "list1",
		},
	}

	// Set up expectations
	mockFetcher.EXPECT().
		GetTotalRecipientCount(ctx, workspaceID, broadcastID).
		Return(100, nil)

	mockFetcher.EXPECT().
		FetchBatch(ctx, workspaceID, broadcastID, offset, limit).
		Return(contacts, nil)

	// Use the mock
	count, err := mockFetcher.GetTotalRecipientCount(ctx, workspaceID, broadcastID)
	assert.NoError(t, err)
	assert.Equal(t, 100, count)

	result, err := mockFetcher.FetchBatch(ctx, workspaceID, broadcastID, offset, limit)
	assert.NoError(t, err)
	assert.Equal(t, contacts, result)
}
