package service

import (
	"context"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Note: These tests focus on input validation rather than testing the AWS interactions directly.
// Fully testing the SES service would require complex mocking of AWS clients or integration tests
// against actual AWS services. The approach taken here is to test the validation layer to ensure
// the service rejects invalid configurations appropriately.

func TestNewSESService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewSESService(mockAuthService, mockLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockAuthService, service.authService)
	assert.Equal(t, mockLogger, service.logger)
}

func TestSESService_RegisterWebhooks_InvalidProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	service := NewSESService(mockAuthService, mockLogger)

	// Test with nil provider config
	result, err := service.RegisterWebhooks(
		context.Background(),
		"test-workspace",
		"test-integration",
		"https://example.com",
		[]domain.EmailEventType{domain.EmailEventDelivered},
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")

	// Test with empty SES config
	result, err = service.RegisterWebhooks(
		context.Background(),
		"test-workspace",
		"test-integration",
		"https://example.com",
		[]domain.EmailEventType{domain.EmailEventDelivered},
		&domain.EmailProvider{
			Kind: domain.EmailProviderKindSES,
			SES:  &domain.AmazonSESSettings{},
		},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")

	// Test with missing required fields
	result, err = service.RegisterWebhooks(
		context.Background(),
		"test-workspace",
		"test-integration",
		"https://example.com",
		[]domain.EmailEventType{domain.EmailEventDelivered},
		&domain.EmailProvider{
			Kind: domain.EmailProviderKindSES,
			SES: &domain.AmazonSESSettings{
				Region: "us-east-1", // missing AccessKey and SecretKey
			},
		},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")
}

func TestSESService_GetWebhookStatus_InvalidProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	service := NewSESService(mockAuthService, mockLogger)

	// Test with nil provider config
	result, err := service.GetWebhookStatus(
		context.Background(),
		"test-workspace",
		"test-integration",
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")

	// Test with empty SES config
	result, err = service.GetWebhookStatus(
		context.Background(),
		"test-workspace",
		"test-integration",
		&domain.EmailProvider{
			Kind: domain.EmailProviderKindSES,
			SES:  &domain.AmazonSESSettings{},
		},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")
}

func TestSESService_UnregisterWebhooks_InvalidProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Set up logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	service := NewSESService(mockAuthService, mockLogger)

	// Test with nil provider config
	err := service.UnregisterWebhooks(
		context.Background(),
		"test-workspace",
		"test-integration",
		nil,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")

	// Test with empty SES config
	err = service.UnregisterWebhooks(
		context.Background(),
		"test-workspace",
		"test-integration",
		&domain.EmailProvider{
			Kind: domain.EmailProviderKindSES,
			SES:  &domain.AmazonSESSettings{},
		},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SES configuration is missing or invalid")
}
