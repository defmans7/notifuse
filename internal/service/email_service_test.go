package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Setup function for EmailService tests
func setupEmailServiceTest(ctrl *gomock.Controller) (*service.EmailService, *domainmocks.MockAuthService, *domainmocks.MockWorkspaceRepository, *domainmocks.MockTemplateRepository, *pkgmocks.MockLogger) {
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockTemplateRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	secretKey := "test-secret-key"

	emailService := service.NewEmailService(
		mockLogger,
		mockAuthService,
		secretKey,
		mockWorkspaceRepo,
		mockTemplateRepo,
	)

	return emailService, mockAuthService, mockWorkspaceRepo, mockTemplateRepo, mockLogger
}

func TestEmailService_TestEmailProvider_SMTP(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	recipientEmail := "test@example.com"

	// Create SMTP provider for testing
	smtpProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Test Sender",
		SMTP: &domain.SMTPSettings{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "testuser",
			Password: "testpassword", // This would be encrypted in production
		},
	}

	t.Run("Success", func(t *testing.T) {
		// Since we can't actually connect to an SMTP server in the test, skip this test
		t.Skip("Skipping due to external SMTP dependency")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		// We can't easily mock the actual SMTP client call since it's created directly in the code
		// In a real implementation, we would refactor to use dependency injection for the mail client

		err := emailService.TestEmailProvider(ctx, workspaceID, smtpProvider, recipientEmail)
		assert.NoError(t, err)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)
		authErr := errors.New("authentication error")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, authErr)

		err := emailService.TestEmailProvider(ctx, workspaceID, smtpProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user for workspace")
	})

	t.Run("Validation Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Create invalid SMTP provider (missing host)
		invalidProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSMTP,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SMTP: &domain.SMTPSettings{
				Port:     587,
				Username: "testuser",
				Password: "testpassword",
			},
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		err := emailService.TestEmailProvider(ctx, workspaceID, invalidProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host is required for SMTP configuration")
	})

	t.Run("Missing SMTP Settings", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Create provider with nil SMTP settings
		invalidProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSMTP,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SMTP:               nil,
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		err := emailService.TestEmailProvider(ctx, workspaceID, invalidProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP settings required")
	})
}

func TestEmailService_TestEmailProvider_SES(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	recipientEmail := "test@example.com"

	// Create SES provider for testing
	sesProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSES,
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Test Sender",
		SES: &domain.AmazonSES{
			Region:    "us-east-1",
			AccessKey: "AKIAXXXXXXXXXXXXXXXX",
			SecretKey: "secret-key", // This would be encrypted in production
		},
	}

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)
		authErr := errors.New("authentication error")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, authErr)

		err := emailService.TestEmailProvider(ctx, workspaceID, sesProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user for workspace")
	})

	t.Run("Validation Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Create invalid SES provider (missing region)
		invalidProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SES: &domain.AmazonSES{
				AccessKey: "AKIAXXXXXXXXXXXXXXXX",
				SecretKey: "secret-key",
			},
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		err := emailService.TestEmailProvider(ctx, workspaceID, invalidProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region is required when Amazon SES is configured")
	})

	t.Run("Missing SES Settings", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Create provider with nil SES settings
		invalidProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSES,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SES:                nil,
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		err := emailService.TestEmailProvider(ctx, workspaceID, invalidProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SES settings required")
	})
}

func TestEmailService_TestEmailProvider_SparkPost(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	recipientEmail := "test@example.com"

	// Create SparkPost provider for testing
	sparkPostProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSparkPost,
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Test Sender",
		SparkPost: &domain.SparkPostSettings{
			Endpoint: "https://api.sparkpost.com",
			APIKey:   "sparkpost-api-key", // This would be encrypted in production
		},
	}

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)
		authErr := errors.New("authentication error")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, authErr)

		err := emailService.TestEmailProvider(ctx, workspaceID, sparkPostProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user for workspace")
	})

	t.Run("Missing SparkPost Settings", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Create provider with nil SparkPost settings
		invalidProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindSparkPost,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			SparkPost:          nil,
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		err := emailService.TestEmailProvider(ctx, workspaceID, invalidProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SparkPost settings required")
	})
}

func TestEmailService_TestEmailProvider_Postmark(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	recipientEmail := "test@example.com"

	// Create Postmark provider for testing
	postmarkProvider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindPostmark,
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Test Sender",
		Postmark: &domain.PostmarkSettings{
			ServerToken: "postmark-server-token", // This would be encrypted in production
		},
	}

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)
		authErr := errors.New("authentication error")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, authErr)

		err := emailService.TestEmailProvider(ctx, workspaceID, postmarkProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user for workspace")
	})

	t.Run("Missing Postmark Settings", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

		// Create provider with nil Postmark settings
		invalidProvider := domain.EmailProvider{
			Kind:               domain.EmailProviderKindPostmark,
			DefaultSenderEmail: "sender@example.com",
			DefaultSenderName:  "Test Sender",
			Postmark:           nil,
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		err := emailService.TestEmailProvider(ctx, workspaceID, invalidProvider, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Postmark settings required")
	})
}

func TestEmailService_TestEmailProvider_UnsupportedProvider(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	recipientEmail := "test@example.com"

	// Create unsupported provider
	unsupportedProvider := domain.EmailProvider{
		Kind:               "unsupported",
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Test Sender",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)

	// Set expectations
	mockAuthService.EXPECT().
		AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
		Return(&domain.User{ID: userID}, nil)

	err := emailService.TestEmailProvider(ctx, workspaceID, unsupportedProvider, recipientEmail)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email provider kind: unsupported")
}

func TestEmailService_TestTemplate(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	templateID := "tmpl-abc"
	providerType := "marketing"
	recipientEmail := "test@example.com"

	// Create workspace with email provider
	workspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			EmailMarketingProvider: domain.EmailProvider{
				Kind:               domain.EmailProviderKindSES,
				DefaultSenderEmail: "sender@example.com",
				DefaultSenderName:  "Test Sender",
				SES: &domain.AmazonSES{
					Region:    "us-east-1",
					AccessKey: "AKIAXXXXXXXXXXXXXXXX",
					SecretKey: "secret-key", // This would be encrypted in production
				},
			},
		},
	}

	// Create template
	template := &domain.Template{
		ID:       templateID,
		Name:     "Test Template",
		Channel:  "email",
		Category: "marketing",
		Email: &domain.EmailTemplate{
			FromAddress:     "sender@example.com",
			FromName:        "Test Sender",
			Subject:         "Test Subject",
			CompiledPreview: "<html><body><h1>Test Email</h1></body></html>",
			VisualEditorTree: mjml.EmailBlock{
				Kind: "root",
				Data: map[string]interface{}{"styles": map[string]interface{}{}},
			},
		},
		TestData: domain.MapOfAny{
			"name":    "Test User",
			"company": "Test Company",
		},
	}

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, _, _, _ := setupEmailServiceTest(ctrl)
		authErr := errors.New("authentication error")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(nil, authErr)

		err := emailService.TestTemplate(ctx, workspaceID, templateID, providerType, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
	})

	t.Run("Workspace Not Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, mockWorkspaceRepo, _, _ := setupEmailServiceTest(ctrl)
		workspaceErr := errors.New("workspace not found")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(nil, workspaceErr)

		err := emailService.TestTemplate(ctx, workspaceID, templateID, providerType, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace")
	})

	t.Run("Template Not Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, mockWorkspaceRepo, mockTemplateRepo, _ := setupEmailServiceTest(ctrl)
		templateErr := errors.New("template not found")

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(nil, templateErr)

		err := emailService.TestTemplate(ctx, workspaceID, templateID, providerType, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get template")
	})

	t.Run("Invalid Provider Type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, mockWorkspaceRepo, mockTemplateRepo, _ := setupEmailServiceTest(ctrl)
		invalidProviderType := "invalid"

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(template, nil)

		err := emailService.TestTemplate(ctx, workspaceID, templateID, invalidProviderType, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid provider type")
	})

	t.Run("No Email Provider Configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, mockWorkspaceRepo, mockTemplateRepo, _ := setupEmailServiceTest(ctrl)

		// Create workspace with no email provider
		workspaceWithoutProvider := &domain.Workspace{
			ID:       workspaceID,
			Name:     "Test Workspace",
			Settings: domain.WorkspaceSettings{},
		}

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspaceWithoutProvider, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(template, nil)

		err := emailService.TestTemplate(ctx, workspaceID, templateID, providerType, recipientEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no email provider configured")
	})

	t.Run("Success with SES Provider", func(t *testing.T) {
		// Skip test due to external AWS dependency
		t.Skip("Skipping due to external AWS dependency")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		emailService, mockAuthService, mockWorkspaceRepo, mockTemplateRepo, _ := setupEmailServiceTest(ctrl)

		// Set expectations
		mockAuthService.EXPECT().
			AuthenticateUserForWorkspace(gomock.Any(), workspaceID).
			Return(&domain.User{ID: userID}, nil)

		mockWorkspaceRepo.EXPECT().
			GetByID(gomock.Any(), workspaceID).
			Return(workspace, nil)

		mockTemplateRepo.EXPECT().
			GetTemplateByID(gomock.Any(), workspaceID, templateID, int64(0)).
			Return(template, nil)

		err := emailService.TestTemplate(ctx, workspaceID, templateID, providerType, recipientEmail)
		assert.NoError(t, err)
	})
}
