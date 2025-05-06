package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessWebhook_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().WithFields(gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	workspaceID := "workspace1"
	integrationID := "integration1"

	// Create test payload
	payload := domain.SESWebhookPayload{
		Message: `{"notificationType":"Bounce","bounce":{"bounceType":"Permanent","bounceSubType":"General","bouncedRecipients":[{"emailAddress":"test@example.com","diagnosticCode":"554"}],"timestamp":"2023-01-01T12:00:00Z"},"mail":{"messageId":"message1"}}`,
	}
	rawPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// Setup mock workspace
	workspace := &domain.Workspace{
		ID: workspaceID,
		Integrations: []domain.Integration{
			{
				ID: integrationID,
				EmailProvider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSES,
					SES: &domain.AmazonSESSettings{
						Region:    "us-east-1",
						AccessKey: "test-key",
						SecretKey: "test-secret",
					},
				},
			},
		},
	}

	// Setup mocks to handle expectations
	mockEvent := &domain.WebhookEvent{
		Type:              domain.EmailEventBounce,
		EmailProviderKind: domain.EmailProviderKindSES,
		IntegrationID:     integrationID,
		RecipientEmail:    "test@example.com",
		MessageID:         "message1",
	}

	// Setup expectations to match what the service will actually store
	workspaceRepo.EXPECT().GetByID(gomock.Any(), workspaceID).Return(workspace, nil)
	repo.EXPECT().StoreEvent(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, event *domain.WebhookEvent) error {
			assert.Equal(t, mockEvent.Type, event.Type)
			assert.Equal(t, mockEvent.EmailProviderKind, event.EmailProviderKind)
			assert.Equal(t, mockEvent.IntegrationID, event.IntegrationID)
			assert.Equal(t, mockEvent.RecipientEmail, event.RecipientEmail)
			assert.Equal(t, mockEvent.MessageID, event.MessageID)
			return nil
		})

	// Create service
	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	// Call method
	err = service.ProcessWebhook(context.Background(), workspaceID, integrationID, rawPayload)

	// Assert
	assert.NoError(t, err)
}

func TestProcessWebhook_WorkspaceNotFound(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().WithFields(gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	workspaceID := "workspace1"
	integrationID := "integration1"
	rawPayload := []byte(`{}`)

	// Setup expectations - simulate workspace not found
	workspaceError := errors.New("workspace not found")
	workspaceRepo.EXPECT().GetByID(gomock.Any(), workspaceID).Return(nil, workspaceError)

	// Create service
	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	// Call method
	err := service.ProcessWebhook(context.Background(), workspaceID, integrationID, rawPayload)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get workspace")
}
