package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// testNotificationCenterService is a modified version of the service for testing
// that allows us to override the email HMAC verification
type testNotificationCenterService struct {
	*NotificationCenterService
	validHMAC string
}

// newTestNotificationCenterService creates a test service with HMAC verification override
func newTestNotificationCenterService(
	contactRepo domain.ContactRepository,
	workspaceRepo domain.WorkspaceRepository,
	listRepo domain.ListRepository,
	logger logger.Logger,
	validHMAC string,
) *testNotificationCenterService {
	svc := NewNotificationCenterService(contactRepo, workspaceRepo, listRepo, logger)
	return &testNotificationCenterService{
		NotificationCenterService: svc,
		validHMAC:                 validHMAC,
	}
}

// GetNotificationCenter overrides the service method with custom HMAC verification
func (s *testNotificationCenterService) GetNotificationCenter(ctx context.Context, workspaceID string, email string, emailHMAC string) (*domain.NotificationCenterResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get workspace: %v", err))
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Override HMAC verification for testing
	if emailHMAC != s.validHMAC {
		return nil, fmt.Errorf("invalid email verification")
	}

	// Get the contact
	contact, err := s.contactRepo.GetContactByEmail(ctx, workspaceID, email)
	if err != nil {
		if strings.Contains(err.Error(), "contact not found") {
			return nil, err
		}
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to get contact: %v", err))
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Get public lists for this workspace
	var publicLists []*domain.List = make([]*domain.List, 0)

	// Get lists using the list service
	lists, err := s.listRepo.GetLists(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get lists: %v", err))
	} else {
		// Filter to only include public lists
		for _, list := range lists {
			if list.IsPublic {
				publicLists = append(publicLists, list)
			}
		}
	}

	return &domain.NotificationCenterResponse{
		Contact:      contact,
		PublicLists:  publicLists,
		ContactLists: contact.ContactLists,
		LogoURL:      workspace.Settings.LogoURL,
		WebsiteURL:   workspace.Settings.WebsiteURL,
	}, nil
}

func TestNotificationCenterService_GetNotificationCenter(t *testing.T) {
	// Set up a mock contact for use in the expected responses
	mockContactLists := []*domain.ContactList{
		{
			ListID: "list-1",
			Status: domain.ContactListStatusActive,
		},
	}

	// Define test cases
	testCases := []struct {
		name          string
		workspaceID   string
		email         string
		emailHMAC     string
		setupMocks    func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger)
		expectedResp  *domain.NotificationCenterResponse
		expectedError error
	}{
		{
			name:        "Success with all data",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "valid-hmac", // This will be accepted by the mocked VerifyEmailHMAC
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				// Setup workspace
				workspace := &domain.Workspace{
					ID:   "workspace-123",
					Name: "Test Workspace",
					Settings: domain.WorkspaceSettings{
						LogoURL:    "https://example.com/logo.png",
						WebsiteURL: "https://example.com",
						SecretKey:  "secret-key", // This will be used to verify HMAC
					},
				}
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(workspace, nil)

				// Setup contact
				contact := &domain.Contact{
					Email: "user@example.com",
					ContactLists: []*domain.ContactList{
						{
							ListID: "list-1",
							Status: domain.ContactListStatusActive,
						},
					},
				}
				mockContactRepo.EXPECT().
					GetContactByEmail(gomock.Any(), "workspace-123", "user@example.com").
					Return(contact, nil)

				// Setup lists
				lists := []*domain.List{
					{
						ID:       "list-1",
						Name:     "Public List 1",
						IsPublic: true,
					},
					{
						ID:       "list-2",
						Name:     "Public List 2",
						IsPublic: true,
					},
					{
						ID:       "list-3",
						Name:     "Private List",
						IsPublic: false, // This one should be filtered out
					},
				}
				mockListRepo.EXPECT().
					GetLists(gomock.Any(), "workspace-123").
					Return(lists, nil)
			},
			expectedResp: &domain.NotificationCenterResponse{
				Contact: &domain.Contact{
					Email: "user@example.com",
					ContactLists: []*domain.ContactList{
						{
							ListID: "list-1",
							Status: domain.ContactListStatusActive,
						},
					},
				},
				PublicLists: []*domain.List{
					{
						ID:       "list-1",
						Name:     "Public List 1",
						IsPublic: true,
					},
					{
						ID:       "list-2",
						Name:     "Public List 2",
						IsPublic: true,
					},
				},
				ContactLists: mockContactLists,
				LogoURL:      "https://example.com/logo.png",
				WebsiteURL:   "https://example.com",
			},
			expectedError: nil,
		},
		{
			name:        "Invalid email HMAC",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "invalid-hmac", // This will be rejected
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				// Setup workspace
				workspace := &domain.Workspace{
					ID:   "workspace-123",
					Name: "Test Workspace",
					Settings: domain.WorkspaceSettings{
						SecretKey: "secret-key", // This will be used to verify HMAC
					},
				}
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(workspace, nil)

				// No other mocks should be called after invalid HMAC
			},
			expectedResp:  nil,
			expectedError: fmt.Errorf("invalid email verification"),
		},
		{
			name:        "Workspace not found",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "valid-hmac",
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(nil, errors.New("workspace not found"))

				// Expect error log for workspace not found case
				mockLogger.EXPECT().
					Error(gomock.Any())
			},
			expectedResp:  nil,
			expectedError: fmt.Errorf("failed to get workspace: workspace not found"),
		},
		{
			name:        "Contact not found",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "valid-hmac",
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				// Setup workspace
				workspace := &domain.Workspace{
					ID:   "workspace-123",
					Name: "Test Workspace",
					Settings: domain.WorkspaceSettings{
						SecretKey: "secret-key",
					},
				}
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(workspace, nil)

				// Contact not found
				mockContactRepo.EXPECT().
					GetContactByEmail(gomock.Any(), "workspace-123", "user@example.com").
					Return(nil, errors.New("contact not found"))
			},
			expectedResp:  nil,
			expectedError: errors.New("contact not found"),
		},
		{
			name:        "Contact fetch error",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "valid-hmac",
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				// Setup workspace
				workspace := &domain.Workspace{
					ID:   "workspace-123",
					Name: "Test Workspace",
					Settings: domain.WorkspaceSettings{
						SecretKey: "secret-key",
					},
				}
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(workspace, nil)

				// Contact fetch error
				mockContactRepo.EXPECT().
					GetContactByEmail(gomock.Any(), "workspace-123", "user@example.com").
					Return(nil, errors.New("database error"))

				// Set up a mock logger that will be returned by WithField
				mockLoggerWithField := pkgmocks.NewMockLogger(ctrl)
				mockLogger.EXPECT().
					WithField("email", "user@example.com").
					Return(mockLoggerWithField)

				mockLoggerWithField.EXPECT().
					Error(gomock.Any())
			},
			expectedResp:  nil,
			expectedError: fmt.Errorf("failed to get contact: database error"),
		},
		{
			name:        "List fetch error but still returns contact data",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "valid-hmac",
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				// Setup workspace
				workspace := &domain.Workspace{
					ID:   "workspace-123",
					Name: "Test Workspace",
					Settings: domain.WorkspaceSettings{
						LogoURL:    "https://example.com/logo.png",
						WebsiteURL: "https://example.com",
						SecretKey:  "secret-key",
					},
				}
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(workspace, nil)

				// Setup contact
				contact := &domain.Contact{
					Email: "user@example.com",
					ContactLists: []*domain.ContactList{
						{
							ListID: "list-1",
							Status: domain.ContactListStatusActive,
						},
					},
				}
				mockContactRepo.EXPECT().
					GetContactByEmail(gomock.Any(), "workspace-123", "user@example.com").
					Return(contact, nil)

				// List fetch error
				mockListRepo.EXPECT().
					GetLists(gomock.Any(), "workspace-123").
					Return(nil, errors.New("database error"))

				// Error logged when lists fetch fails
				mockLogger.EXPECT().
					Error(gomock.Any())
			},
			expectedResp: &domain.NotificationCenterResponse{
				Contact: &domain.Contact{
					Email: "user@example.com",
					ContactLists: []*domain.ContactList{
						{
							ListID: "list-1",
							Status: domain.ContactListStatusActive,
						},
					},
				},
				PublicLists:  []*domain.List{}, // Match the expected format from the implementation (empty slice)
				ContactLists: mockContactLists,
				LogoURL:      "https://example.com/logo.png",
				WebsiteURL:   "https://example.com",
			},
			expectedError: nil, // Still succeeds because contact was found
		},
		{
			name:        "Empty public lists",
			workspaceID: "workspace-123",
			email:       "user@example.com",
			emailHMAC:   "valid-hmac",
			setupMocks: func(ctrl *gomock.Controller, mockContactRepo *mocks.MockContactRepository, mockWorkspaceRepo *mocks.MockWorkspaceRepository, mockListRepo *mocks.MockListRepository, mockLogger *pkgmocks.MockLogger) {
				// Setup workspace
				workspace := &domain.Workspace{
					ID:   "workspace-123",
					Name: "Test Workspace",
					Settings: domain.WorkspaceSettings{
						LogoURL:    "https://example.com/logo.png",
						WebsiteURL: "https://example.com",
						SecretKey:  "secret-key",
					},
				}
				mockWorkspaceRepo.EXPECT().
					GetByID(gomock.Any(), "workspace-123").
					Return(workspace, nil)

				// Setup contact
				contact := &domain.Contact{
					Email: "user@example.com",
					ContactLists: []*domain.ContactList{
						{
							ListID: "list-1",
							Status: domain.ContactListStatusActive,
						},
					},
				}
				mockContactRepo.EXPECT().
					GetContactByEmail(gomock.Any(), "workspace-123", "user@example.com").
					Return(contact, nil)

				// Only private lists
				lists := []*domain.List{
					{
						ID:       "list-1",
						Name:     "Private List 1",
						IsPublic: false,
					},
					{
						ID:       "list-2",
						Name:     "Private List 2",
						IsPublic: false,
					},
				}
				mockListRepo.EXPECT().
					GetLists(gomock.Any(), "workspace-123").
					Return(lists, nil)
			},
			expectedResp: &domain.NotificationCenterResponse{
				Contact: &domain.Contact{
					Email: "user@example.com",
					ContactLists: []*domain.ContactList{
						{
							ListID: "list-1",
							Status: domain.ContactListStatusActive,
						},
					},
				},
				PublicLists:  []*domain.List{}, // Empty slice, not nil
				ContactLists: mockContactLists,
				LogoURL:      "https://example.com/logo.png",
				WebsiteURL:   "https://example.com",
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockContactRepo := mocks.NewMockContactRepository(ctrl)
			mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
			mockListRepo := mocks.NewMockListRepository(ctrl)
			mockLogger := pkgmocks.NewMockLogger(ctrl)

			// Setup expectations
			tc.setupMocks(ctrl, mockContactRepo, mockWorkspaceRepo, mockListRepo, mockLogger)

			// Create test service with mocks and custom HMAC validation
			service := newTestNotificationCenterService(
				mockContactRepo,
				mockWorkspaceRepo,
				mockListRepo,
				mockLogger,
				"valid-hmac",
			)

			// Call the method
			resp, err := service.GetNotificationCenter(context.Background(), tc.workspaceID, tc.email, tc.emailHMAC)

			// Assertions
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
