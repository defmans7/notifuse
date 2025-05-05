package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// ExampleService is a simple service that uses WebhookRegistrationService
type ExampleService struct {
	webhookRegistrationService domain.WebhookRegistrationService
}

func NewExampleService(webhookRegistrationService domain.WebhookRegistrationService) *ExampleService {
	return &ExampleService{
		webhookRegistrationService: webhookRegistrationService,
	}
}

func (s *ExampleService) RegisterWebhooks(ctx context.Context, req *domain.RegisterWebhookRequest) (*domain.WebhookRegistrationStatus, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	config := &domain.WebhookRegistrationConfig{
		IntegrationID: req.IntegrationID,
		EventTypes:    req.EventTypes,
	}

	return s.webhookRegistrationService.RegisterWebhooks(ctx, req.WorkspaceID, config)
}

func (s *ExampleService) GetWebhookStatus(ctx context.Context, req *domain.GetWebhookStatusRequest) (*domain.WebhookRegistrationStatus, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return s.webhookRegistrationService.GetWebhookStatus(ctx, req.WorkspaceID, req.IntegrationID)
}

func TestExampleService_RegisterWebhooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebhookService := mocks.NewMockWebhookRegistrationService(ctrl)
	service := NewExampleService(mockWebhookService)

	ctx := context.Background()
	req := &domain.RegisterWebhookRequest{
		WorkspaceID:   "ws-123",
		IntegrationID: "int-123",
		EventTypes: []domain.EmailEventType{
			domain.EmailEventDelivered,
			domain.EmailEventBounce,
		},
	}

	expectedConfig := &domain.WebhookRegistrationConfig{
		IntegrationID: req.IntegrationID,
		EventTypes:    req.EventTypes,
	}

	expectedStatus := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSES,
		IsRegistered:      true,
		RegisteredEvents:  req.EventTypes,
	}

	mockWebhookService.EXPECT().
		RegisterWebhooks(ctx, req.WorkspaceID, gomock.Any()).
		Do(func(_ context.Context, _ string, config *domain.WebhookRegistrationConfig) {
			assert.Equal(t, expectedConfig.IntegrationID, config.IntegrationID)
			assert.Equal(t, expectedConfig.EventTypes, config.EventTypes)
		}).
		Return(expectedStatus, nil)

	status, err := service.RegisterWebhooks(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func TestExampleService_RegisterWebhooks_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebhookService := mocks.NewMockWebhookRegistrationService(ctrl)
	service := NewExampleService(mockWebhookService)

	ctx := context.Background()
	req := &domain.RegisterWebhookRequest{
		// Missing WorkspaceID
		IntegrationID: "int-123",
	}

	// The mock should not be called because validation should fail
	status, err := service.RegisterWebhooks(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "workspace_id is required")
}

func TestExampleService_RegisterWebhooks_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebhookService := mocks.NewMockWebhookRegistrationService(ctrl)
	service := NewExampleService(mockWebhookService)

	ctx := context.Background()
	req := &domain.RegisterWebhookRequest{
		WorkspaceID:   "ws-123",
		IntegrationID: "int-123",
		EventTypes: []domain.EmailEventType{
			domain.EmailEventDelivered,
		},
	}

	expectedError := errors.New("service error")

	mockWebhookService.EXPECT().
		RegisterWebhooks(ctx, req.WorkspaceID, gomock.Any()).
		Return(nil, expectedError)

	status, err := service.RegisterWebhooks(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Equal(t, expectedError, err)
}

func TestExampleService_GetWebhookStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebhookService := mocks.NewMockWebhookRegistrationService(ctrl)
	service := NewExampleService(mockWebhookService)

	ctx := context.Background()
	req := &domain.GetWebhookStatusRequest{
		WorkspaceID:   "ws-123",
		IntegrationID: "int-123",
	}

	expectedStatus := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSES,
		IsRegistered:      true,
		RegisteredEvents: []domain.EmailEventType{
			domain.EmailEventDelivered,
			domain.EmailEventBounce,
		},
	}

	mockWebhookService.EXPECT().
		GetWebhookStatus(ctx, req.WorkspaceID, req.IntegrationID).
		Return(expectedStatus, nil)

	status, err := service.GetWebhookStatus(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func TestExampleService_GetWebhookStatus_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebhookService := mocks.NewMockWebhookRegistrationService(ctrl)
	service := NewExampleService(mockWebhookService)

	ctx := context.Background()
	req := &domain.GetWebhookStatusRequest{
		// Missing WorkspaceID
		IntegrationID: "int-123",
	}

	// The mock should not be called because validation should fail
	status, err := service.GetWebhookStatus(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "workspace_id is required")
}

func TestExampleService_GetWebhookStatus_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebhookService := mocks.NewMockWebhookRegistrationService(ctrl)
	service := NewExampleService(mockWebhookService)

	ctx := context.Background()
	req := &domain.GetWebhookStatusRequest{
		WorkspaceID:   "ws-123",
		IntegrationID: "int-123",
	}

	expectedError := errors.New("service error")

	mockWebhookService.EXPECT().
		GetWebhookStatus(ctx, req.WorkspaceID, req.IntegrationID).
		Return(nil, expectedError)

	status, err := service.GetWebhookStatus(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Equal(t, expectedError, err)
}
