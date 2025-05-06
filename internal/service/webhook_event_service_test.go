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

func TestNewWebhookEventService(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Call the function
	service := NewWebhookEventService(repo, authService, log, workspaceRepo)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
	assert.Equal(t, authService, service.authService)
	assert.Equal(t, log, service.logger)
	assert.Equal(t, workspaceRepo, service.workspaceRepo)
}

func TestGetEventByID(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	eventID := "event1"
	event := &domain.WebhookEvent{
		ID:                "event1",
		Type:              domain.EmailEventDelivered,
		EmailProviderKind: domain.EmailProviderKindSES,
		IntegrationID:     "integration1",
		RecipientEmail:    "test@example.com",
		MessageID:         "message1",
	}

	// Setup mock to return event
	repo.EXPECT().GetEventByID(gomock.Any(), eventID).Return(event, nil)

	// Create service
	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	// Call method
	result, err := service.GetEventByID(context.Background(), eventID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, event, result)
}

func TestGetEventsByType(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	workspaceID := "workspace1"
	eventType := domain.EmailEventBounce
	user := &domain.User{ID: "user1"}
	events := []*domain.WebhookEvent{
		{
			ID:                "event1",
			Type:              domain.EmailEventBounce,
			EmailProviderKind: domain.EmailProviderKindSES,
			IntegrationID:     "integration1",
			RecipientEmail:    "test@example.com",
			MessageID:         "message1",
		},
	}

	// Setup mocks for authentication and repository
	authService.EXPECT().AuthenticateUserForWorkspace(gomock.Any(), workspaceID).Return(
		context.Background(), user, nil)
	repo.EXPECT().GetEventsByType(gomock.Any(), workspaceID, eventType, 10, 0).Return(events, nil)

	// Create service
	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	// Call method
	result, err := service.GetEventsByType(context.Background(), workspaceID, eventType, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, events, result)
}

func TestGetEventsByMessageID(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	messageID := "message1"
	events := []*domain.WebhookEvent{
		{
			ID:                "event1",
			Type:              domain.EmailEventDelivered,
			EmailProviderKind: domain.EmailProviderKindSES,
			IntegrationID:     "integration1",
			RecipientEmail:    "test@example.com",
			MessageID:         messageID,
		},
	}

	// Setup mock to return events
	repo.EXPECT().GetEventsByMessageID(gomock.Any(), messageID, 10, 0).Return(events, nil)

	// Create service
	service := NewWebhookEventService(repo, authService, log, workspaceRepo)

	// Call method
	result, err := service.GetEventsByMessageID(context.Background(), messageID, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, events, result)
}

func TestGetEventsByTransactionalID(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	transactionalID := "transaction1"
	events := []*domain.WebhookEvent{
		{
			ID:                "event1",
			Type:              domain.EmailEventDelivered,
			EmailProviderKind: domain.EmailProviderKindSES,
			IntegrationID:     "integration1",
			RecipientEmail:    "test@example.com",
			MessageID:         "message1",
			TransactionalID:   transactionalID,
		},
	}

	// Setup mock to return events
	repo.EXPECT().GetEventsByTransactionalID(gomock.Any(), transactionalID, 10, 0).Return(events, nil)

	// Create service
	service := NewWebhookEventService(repo, authService, log, workspaceRepo)

	// Call method
	result, err := service.GetEventsByTransactionalID(context.Background(), transactionalID, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, events, result)
}

func TestGetEventsByBroadcastID(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	broadcastID := "broadcast1"
	events := []*domain.WebhookEvent{
		{
			ID:                "event1",
			Type:              domain.EmailEventDelivered,
			EmailProviderKind: domain.EmailProviderKindSES,
			IntegrationID:     "integration1",
			RecipientEmail:    "test@example.com",
			MessageID:         "message1",
			BroadcastID:       broadcastID,
		},
	}

	// Setup mock to return events
	repo.EXPECT().GetEventsByBroadcastID(gomock.Any(), broadcastID, 10, 0).Return(events, nil)

	// Create service
	service := NewWebhookEventService(repo, authService, log, workspaceRepo)

	// Call method
	result, err := service.GetEventsByBroadcastID(context.Background(), broadcastID, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, events, result)
}

func TestGetEventCount(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Setup logging expectations
	log.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
	log.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create test data
	workspaceID := "workspace1"
	eventType := domain.EmailEventBounce
	user := &domain.User{ID: "user1"}
	expectedCount := 5

	// Setup mocks for authentication and repository
	authService.EXPECT().AuthenticateUserForWorkspace(gomock.Any(), workspaceID).Return(
		context.Background(), user, nil)
	repo.EXPECT().GetEventCount(gomock.Any(), workspaceID, eventType).Return(expectedCount, nil)

	// Create service
	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	// Call method
	count, err := service.GetEventCount(context.Background(), workspaceID, eventType)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
}

func TestProcessSESWebhook(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	integrationID := "integration1"

	// Create test bounce payload
	payload := domain.SESWebhookPayload{
		Message: `{"notificationType":"Bounce","bounce":{"bounceType":"Permanent","bounceSubType":"General","bouncedRecipients":[{"emailAddress":"test@example.com","diagnosticCode":"554"}],"timestamp":"2023-01-01T12:00:00Z"},"mail":{"messageId":"message1"}}`,
	}
	rawPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// Call method
	event, err := service.processSESWebhook(integrationID, rawPayload)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, domain.EmailEventBounce, event.Type)
	assert.Equal(t, domain.EmailProviderKindSES, event.EmailProviderKind)
	assert.Equal(t, integrationID, event.IntegrationID)
	assert.Equal(t, "test@example.com", event.RecipientEmail)
	assert.Equal(t, "message1", event.MessageID)
	assert.Equal(t, "Permanent", event.BounceType)
	assert.Equal(t, "General", event.BounceCategory)
	assert.Equal(t, "554", event.BounceDiagnostic)
}

func TestProcessPostmarkWebhook(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	integrationID := "integration1"

	t.Run("Delivery Event", func(t *testing.T) {
		// Create test delivery payload using a map to ensure correct JSON structure
		rawPayload, err := json.Marshal(map[string]interface{}{
			"RecordType":  "Delivery",
			"MessageID":   "message1",
			"Recipient":   "test@example.com",
			"DeliveredAt": "2023-01-01T12:00:00Z",
			"Details":     "250 OK",
		})
		require.NoError(t, err)

		// Call method
		event, err := service.processPostmarkWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventDelivered, event.Type)
		assert.Equal(t, domain.EmailProviderKindPostmark, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
	})

	t.Run("Bounce Event", func(t *testing.T) {
		// Create test bounce payload using a map to ensure correct JSON structure
		rawPayload, err := json.Marshal(map[string]interface{}{
			"RecordType": "Bounce",
			"MessageID":  "message1",
			"Email":      "test@example.com",
			"Type":       "HardBounce",
			"TypeCode":   1,
			"Details":    "550 Address rejected",
			"BouncedAt":  "2023-01-01T12:00:00Z",
		})
		require.NoError(t, err)

		// Call method
		event, err := service.processPostmarkWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventBounce, event.Type)
		assert.Equal(t, domain.EmailProviderKindPostmark, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
		assert.Equal(t, "HardBounce", event.BounceType)
		assert.Equal(t, "HardBounce", event.BounceCategory)
		assert.Equal(t, "550 Address rejected", event.BounceDiagnostic)
	})

	t.Run("Complaint Event", func(t *testing.T) {
		// Create test complaint payload using a map to ensure correct JSON structure
		rawPayload, err := json.Marshal(map[string]interface{}{
			"RecordType":   "SpamComplaint",
			"MessageID":    "message1",
			"Email":        "test@example.com",
			"Type":         "SpamComplaint",
			"ComplainedAt": "2023-01-01T12:00:00Z",
		})
		require.NoError(t, err)

		// Call method
		event, err := service.processPostmarkWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventComplaint, event.Type)
		assert.Equal(t, domain.EmailProviderKindPostmark, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
		assert.Equal(t, "SpamComplaint", event.ComplaintFeedbackType)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create invalid payload
		rawPayload := []byte(`{invalid json`)

		// Call method
		event, err := service.processPostmarkWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("Unsupported Record Type", func(t *testing.T) {
		// Create unsupported record type
		rawPayload, err := json.Marshal(map[string]interface{}{
			"RecordType": "Unknown",
			"MessageID":  "message1",
		})
		require.NoError(t, err)

		// Call method
		event, err := service.processPostmarkWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})
}

func TestProcessSparkPostWebhook(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	integrationID := "integration1"

	t.Run("Delivery Event", func(t *testing.T) {
		// Create test delivery payload
		payload := domain.SparkPostWebhookPayload{
			MSys: domain.SparkPostMSys{
				DeliveryEvent: &domain.SparkPostDeliveryEvent{
					RecipientTo: "test@example.com",
					MessageID:   "message1",
					Timestamp:   "2023-01-01T12:00:00Z",
				},
			},
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSparkPostWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventDelivered, event.Type)
		assert.Equal(t, domain.EmailProviderKindSparkPost, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
	})

	t.Run("Bounce Event", func(t *testing.T) {
		// Create test bounce payload
		payload := domain.SparkPostWebhookPayload{
			MSys: domain.SparkPostMSys{
				BounceEvent: &domain.SparkPostBounceEvent{
					RecipientTo: "test@example.com",
					MessageID:   "message1",
					BounceClass: "21", // Hard bounce
					Reason:      "550 5.1.1 The email account does not exist",
					Timestamp:   "2023-01-01T12:00:00Z",
				},
			},
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSparkPostWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventBounce, event.Type)
		assert.Equal(t, domain.EmailProviderKindSparkPost, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
		assert.Equal(t, "Bounce", event.BounceType)
		assert.Equal(t, "21", event.BounceCategory)
		assert.Equal(t, "550 5.1.1 The email account does not exist", event.BounceDiagnostic)
	})

	t.Run("Complaint Event", func(t *testing.T) {
		// Create test complaint payload
		payload := domain.SparkPostWebhookPayload{
			MSys: domain.SparkPostMSys{
				SpamComplaint: &domain.SparkPostSpamComplaint{
					RecipientTo:  "test@example.com",
					MessageID:    "message1",
					FeedbackType: "abuse",
					Timestamp:    "2023-01-01T12:00:00Z",
				},
			},
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSparkPostWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventComplaint, event.Type)
		assert.Equal(t, domain.EmailProviderKindSparkPost, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
		assert.Equal(t, "abuse", event.ComplaintFeedbackType)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create invalid payload
		rawPayload := []byte(`{invalid json`)

		// Call method
		event, err := service.processSparkPostWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("No Supported Event", func(t *testing.T) {
		// Create payload with no supported event
		payload := domain.SparkPostWebhookPayload{
			MSys: domain.SparkPostMSys{},
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSparkPostWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})
}

func TestProcessMailjetWebhook(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	integrationID := "integration1"

	t.Run("Sent Event", func(t *testing.T) {
		// Create test sent payload
		payload := domain.MailjetWebhookPayload{
			Event:     "sent",
			Time:      1672574400, // 2023-01-01T12:00:00Z
			Email:     "test@example.com",
			MessageID: 12345,
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processMailjetWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventDelivered, event.Type)
		assert.Equal(t, domain.EmailProviderKindMailjet, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "12345", event.MessageID)
	})

	t.Run("Bounce Event", func(t *testing.T) {
		// Create test bounce payload
		payload := domain.MailjetWebhookPayload{
			Event:      "bounce",
			Time:       1672574400, // 2023-01-01T12:00:00Z
			Email:      "test@example.com",
			MessageID:  12345,
			HardBounce: true,
			Comment:    "Mailbox does not exist",
			ErrorCode:  "550",
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processMailjetWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventBounce, event.Type)
		assert.Equal(t, domain.EmailProviderKindMailjet, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "12345", event.MessageID)
		assert.Equal(t, "HardBounce", event.BounceType)
		assert.Equal(t, "Permanent", event.BounceCategory)
		assert.Equal(t, "Mailbox does not exist: 550", event.BounceDiagnostic)
	})

	t.Run("Spam Event", func(t *testing.T) {
		// Create test spam payload
		payload := domain.MailjetWebhookPayload{
			Event:     "spam",
			Time:      1672574400, // 2023-01-01T12:00:00Z
			Email:     "test@example.com",
			MessageID: 12345,
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processMailjetWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventComplaint, event.Type)
		assert.Equal(t, domain.EmailProviderKindMailjet, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "12345", event.MessageID)
		assert.Equal(t, "abuse", event.ComplaintFeedbackType)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create invalid payload
		rawPayload := []byte(`{invalid json`)

		// Call method
		event, err := service.processMailjetWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("Unsupported Event Type", func(t *testing.T) {
		// Create unsupported event type
		payload := domain.MailjetWebhookPayload{
			Event:     "unknown",
			Time:      1672574400,
			Email:     "test@example.com",
			MessageID: 12345,
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processMailjetWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})
}

func TestProcessSMTPWebhook(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockWebhookEventRepository(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	log := pkgmocks.NewMockLogger(ctrl)
	workspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	service := &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        log,
		workspaceRepo: workspaceRepo,
	}

	integrationID := "integration1"

	t.Run("Delivered Event", func(t *testing.T) {
		// Create test delivery payload
		payload := domain.SMTPWebhookPayload{
			Event:     "delivered",
			Timestamp: "2023-01-01T12:00:00Z",
			Recipient: "test@example.com",
			MessageID: "message1",
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSMTPWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventDelivered, event.Type)
		assert.Equal(t, domain.EmailProviderKindSMTP, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
	})

	t.Run("Bounce Event", func(t *testing.T) {
		// Create test bounce payload
		payload := domain.SMTPWebhookPayload{
			Event:          "bounce",
			Timestamp:      "2023-01-01T12:00:00Z",
			Recipient:      "test@example.com",
			MessageID:      "message1",
			BounceCategory: "Permanent",
			DiagnosticCode: "550 5.1.1 User unknown",
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSMTPWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventBounce, event.Type)
		assert.Equal(t, domain.EmailProviderKindSMTP, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
		assert.Equal(t, "Bounce", event.BounceType)
		assert.Equal(t, "Permanent", event.BounceCategory)
		assert.Equal(t, "550 5.1.1 User unknown", event.BounceDiagnostic)
	})

	t.Run("Complaint Event", func(t *testing.T) {
		// Create test complaint payload
		payload := domain.SMTPWebhookPayload{
			Event:         "complaint",
			Timestamp:     "2023-01-01T12:00:00Z",
			Recipient:     "test@example.com",
			MessageID:     "message1",
			ComplaintType: "abuse",
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSMTPWebhook(integrationID, rawPayload)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, domain.EmailEventComplaint, event.Type)
		assert.Equal(t, domain.EmailProviderKindSMTP, event.EmailProviderKind)
		assert.Equal(t, integrationID, event.IntegrationID)
		assert.Equal(t, "test@example.com", event.RecipientEmail)
		assert.Equal(t, "message1", event.MessageID)
		assert.Equal(t, "abuse", event.ComplaintFeedbackType)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create invalid payload
		rawPayload := []byte(`{invalid json`)

		// Call method
		event, err := service.processSMTPWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("Unsupported Event Type", func(t *testing.T) {
		// Create unsupported event type
		payload := domain.SMTPWebhookPayload{
			Event:     "unknown",
			Timestamp: "2023-01-01T12:00:00Z",
			Recipient: "test@example.com",
			MessageID: "message1",
		}
		rawPayload, err := json.Marshal(payload)
		require.NoError(t, err)

		// Call method
		event, err := service.processSMTPWebhook(integrationID, rawPayload)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, event)
	})
}
