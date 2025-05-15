package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMessageHistoryService_ListMessages(t *testing.T) {
	// Create fixed timestamps for testing
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)

	// Define test cases
	testCases := []struct {
		name           string
		workspaceID    string
		params         domain.MessageListParams
		setupMocks     func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger)
		expectedResult *domain.MessageListResult
		expectedError  error
	}{
		{
			name:        "Success with messages and next cursor",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit: 10,
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-1",
						ContactEmail: "user@example.com",
						TemplateID:   "template-1",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
					},
					{
						ID:           "msg-2",
						ContactEmail: "user2@example.com",
						TemplateID:   "template-1",
						Channel:      "email",
						Status:       domain.MessageStatusDelivered,
					},
				}
				nextCursor := "cursor-value"
				mockRepo.EXPECT().
					ListMessages(gomock.Any(), "workspace-123", gomock.Any()).
					Return(messages, nextCursor, nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-1",
						ContactEmail: "user@example.com",
						TemplateID:   "template-1",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
					},
					{
						ID:           "msg-2",
						ContactEmail: "user2@example.com",
						TemplateID:   "template-1",
						Channel:      "email",
						Status:       domain.MessageStatusDelivered,
					},
				},
				NextCursor: "cursor-value",
				HasMore:    true,
			},
			expectedError: nil,
		},
		{
			name:        "Success with empty result",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit: 10,
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				mockRepo.EXPECT().
					ListMessages(gomock.Any(), "workspace-123", gomock.Any()).
					Return([]*domain.MessageHistory{}, "", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages:   []*domain.MessageHistory{},
				NextCursor: "",
				HasMore:    false,
			},
			expectedError: nil,
		},
		{
			name:        "Repository error",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit: 10,
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				mockRepo.EXPECT().
					ListMessages(gomock.Any(), "workspace-123", gomock.Any()).
					Return(nil, "", errors.New("database error"))

				// We expect the error to be logged
				mockLogger.EXPECT().
					Error(gomock.Any())
			},
			expectedResult: nil,
			expectedError:  errors.New("database error"),
		},
		{
			name:        "With filters",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit:        10,
				Channel:      "email",
				Status:       domain.MessageStatusSent,
				ContactEmail: "user@example.com",
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-1",
						ContactEmail: "user@example.com",
						TemplateID:   "template-1",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
					},
				}
				mockRepo.EXPECT().
					ListMessages(
						gomock.Any(),
						"workspace-123",
						gomock.Any(),
					).
					Do(func(_ context.Context, _ string, params domain.MessageListParams) {
						assert.Equal(t, "email", params.Channel)
						assert.Equal(t, domain.MessageStatusSent, params.Status)
						assert.Equal(t, "user@example.com", params.ContactEmail)
						assert.Equal(t, 10, params.Limit)
					}).
					Return(messages, "", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-1",
						ContactEmail: "user@example.com",
						TemplateID:   "template-1",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
					},
				},
				NextCursor: "",
				HasMore:    false,
			},
			expectedError: nil,
		},
		{
			name:        "With cursor-based pagination",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Cursor: "previous-cursor",
				Limit:  10,
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-3",
						ContactEmail: "user3@example.com",
						TemplateID:   "template-2",
						Channel:      "email",
						Status:       domain.MessageStatusDelivered,
					},
					{
						ID:           "msg-4",
						ContactEmail: "user4@example.com",
						TemplateID:   "template-2",
						Channel:      "email",
						Status:       domain.MessageStatusOpened,
					},
				}

				mockRepo.EXPECT().
					ListMessages(
						gomock.Any(),
						"workspace-123",
						gomock.Any(),
					).
					Do(func(_ context.Context, _ string, params domain.MessageListParams) {
						assert.Equal(t, "previous-cursor", params.Cursor)
						assert.Equal(t, 10, params.Limit)
					}).
					Return(messages, "next-cursor", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-3",
						ContactEmail: "user3@example.com",
						TemplateID:   "template-2",
						Channel:      "email",
						Status:       domain.MessageStatusDelivered,
					},
					{
						ID:           "msg-4",
						ContactEmail: "user4@example.com",
						TemplateID:   "template-2",
						Channel:      "email",
						Status:       domain.MessageStatusOpened,
					},
				},
				NextCursor: "next-cursor",
				HasMore:    true,
			},
			expectedError: nil,
		},
		{
			name:        "With time filters",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit:      10,
				SentAfter:  &twoHoursAgo,
				SentBefore: &now,
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-5",
						ContactEmail: "user5@example.com",
						TemplateID:   "template-3",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
						SentAt:       twoHoursAgo.Add(time.Hour),
					},
				}

				mockRepo.EXPECT().
					ListMessages(
						gomock.Any(),
						"workspace-123",
						gomock.Any(),
					).
					Do(func(_ context.Context, _ string, params domain.MessageListParams) {
						assert.Equal(t, twoHoursAgo, *params.SentAfter)
						assert.Equal(t, now, *params.SentBefore)
					}).
					Return(messages, "", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-5",
						ContactEmail: "user5@example.com",
						TemplateID:   "template-3",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
						SentAt:       twoHoursAgo.Add(time.Hour),
					},
				},
				NextCursor: "",
				HasMore:    false,
			},
			expectedError: nil,
		},
		{
			name:        "With error filter",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit:    10,
				HasError: boolPtr(true),
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				errMsg := "failed to send email"
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-6",
						ContactEmail: "user6@example.com",
						TemplateID:   "template-4",
						Channel:      "email",
						Status:       domain.MessageStatusFailed,
						Error:        &errMsg,
					},
				}

				mockRepo.EXPECT().
					ListMessages(
						gomock.Any(),
						"workspace-123",
						gomock.Any(),
					).
					Do(func(_ context.Context, _ string, params domain.MessageListParams) {
						assert.True(t, *params.HasError)
					}).
					Return(messages, "", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-6",
						ContactEmail: "user6@example.com",
						TemplateID:   "template-4",
						Channel:      "email",
						Status:       domain.MessageStatusFailed,
						Error:        strPtr("failed to send email"),
					},
				},
				NextCursor: "",
				HasMore:    false,
			},
			expectedError: nil,
		},
		{
			name:        "Multiple filter combinations",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Limit:       10,
				Channel:     "email",
				Status:      domain.MessageStatusFailed,
				BroadcastID: "broadcast-123",
				TemplateID:  "template-5",
				HasError:    boolPtr(true),
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				broadcastID := "broadcast-123"
				errMsg := "template error"
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-7",
						ContactEmail: "user7@example.com",
						BroadcastID:  &broadcastID,
						TemplateID:   "template-5",
						Channel:      "email",
						Status:       domain.MessageStatusFailed,
						Error:        &errMsg,
					},
				}

				mockRepo.EXPECT().
					ListMessages(
						gomock.Any(),
						"workspace-123",
						gomock.Any(),
					).
					Do(func(_ context.Context, _ string, params domain.MessageListParams) {
						assert.Equal(t, "email", params.Channel)
						assert.Equal(t, domain.MessageStatusFailed, params.Status)
						assert.Equal(t, "broadcast-123", params.BroadcastID)
						assert.Equal(t, "template-5", params.TemplateID)
						assert.True(t, *params.HasError)
					}).
					Return(messages, "", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-7",
						ContactEmail: "user7@example.com",
						BroadcastID:  strPtr("broadcast-123"),
						TemplateID:   "template-5",
						Channel:      "email",
						Status:       domain.MessageStatusFailed,
						Error:        strPtr("template error"),
					},
				},
				NextCursor: "",
				HasMore:    false,
			},
			expectedError: nil,
		},
		{
			name:        "Last page of results",
			workspaceID: "workspace-123",
			params: domain.MessageListParams{
				Cursor: "last-page-cursor",
				Limit:  10,
			},
			setupMocks: func(mockRepo *mocks.MockMessageHistoryRepository, mockLogger *pkgmocks.MockLogger) {
				messages := []*domain.MessageHistory{
					{
						ID:           "msg-8",
						ContactEmail: "user8@example.com",
						TemplateID:   "template-6",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
					},
				}
				// Empty next cursor indicates no more pages
				mockRepo.EXPECT().
					ListMessages(gomock.Any(), "workspace-123", gomock.Any()).
					Return(messages, "", nil)
			},
			expectedResult: &domain.MessageListResult{
				Messages: []*domain.MessageHistory{
					{
						ID:           "msg-8",
						ContactEmail: "user8@example.com",
						TemplateID:   "template-6",
						Channel:      "email",
						Status:       domain.MessageStatusSent,
					},
				},
				NextCursor: "",
				HasMore:    false,
			},
			expectedError: nil,
		},
	}

	// Execute test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockMessageHistoryRepository(ctrl)
			mockLogger := pkgmocks.NewMockLogger(ctrl)
			tc.setupMocks(mockRepo, mockLogger)

			// Create service with mocks
			service := NewMessageHistoryService(mockRepo, mockLogger)

			// Call the method under test
			result, err := service.ListMessages(context.Background(), tc.workspaceID, tc.params)

			// Verify expectations
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
