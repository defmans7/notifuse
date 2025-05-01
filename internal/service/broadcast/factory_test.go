package broadcast

import (
	"testing"

	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewFactory(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	config := DefaultConfig()

	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "With config provided",
			config: config,
		},
		{
			name:   "With nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(
				mockBroadcastService,
				mockTemplateService,
				mockEmailService,
				mockContactRepo,
				mockTaskRepo,
				mockLogger,
				tt.config,
			)

			assert.NotNil(t, factory)
			assert.Equal(t, mockBroadcastService, factory.broadcastService)
			assert.Equal(t, mockTemplateService, factory.templateService)
			assert.Equal(t, mockEmailService, factory.emailService)
			assert.Equal(t, mockContactRepo, factory.contactRepo)
			assert.Equal(t, mockTaskRepo, factory.taskRepo)
			assert.Equal(t, mockLogger, factory.logger)

			// Check if config was set to default when nil was provided
			if tt.config == nil {
				assert.NotNil(t, factory.config)
			} else {
				assert.Equal(t, tt.config, factory.config)
			}
		})
	}
}

func TestFactory_CreateMessageSender(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	config := DefaultConfig()

	factory := NewFactory(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		config,
	)

	// Test CreateMessageSender
	messageSender := factory.CreateMessageSender()

	// Assert messageSender is not nil and is of MessageSender type
	assert.NotNil(t, messageSender)
	_, ok := messageSender.(MessageSender)
	assert.True(t, ok, "Sender should implement MessageSender interface")
}

func TestFactory_CreateOrchestrator(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	config := DefaultConfig()

	factory := NewFactory(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		config,
	)

	// Test CreateOrchestrator
	orchestrator := factory.CreateOrchestrator()

	// Assert orchestrator is not nil
	assert.NotNil(t, orchestrator)
	_, ok := orchestrator.(BroadcastOrchestratorInterface)
	assert.True(t, ok, "Orchestrator should implement BroadcastOrchestratorInterface")
}

func TestFactory_RegisterWithTaskService(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockTaskService := mocks.NewMockTaskService(ctrl)
	config := DefaultConfig()

	// Setup expectations
	mockTaskService.EXPECT().RegisterProcessor(gomock.Any()).Return()
	mockLogger.EXPECT().Info(gomock.Any()).Return()

	factory := NewFactory(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockContactRepo,
		mockTaskRepo,
		mockLogger,
		config,
	)

	// Test RegisterWithTaskService
	factory.RegisterWithTaskService(mockTaskService)

	// With gomock, assertions are verified automatically when ctrl.Finish() is called
}
