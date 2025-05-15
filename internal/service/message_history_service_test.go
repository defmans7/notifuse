package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMessageHistoryService_ListMessages(t *testing.T) {
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
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockRepo := mocks.NewMockMessageHistoryRepository(ctrl)
			mockLogger := pkgmocks.NewMockLogger(ctrl)

			// Setup expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create service with mocks
			service := NewMessageHistoryService(mockRepo, mockLogger)

			// Call the method
			result, err := service.ListMessages(context.Background(), tc.workspaceID, tc.params)

			// Assertions
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
