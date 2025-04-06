package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks" // Corrected import path
	"github.com/Notifuse/notifuse/internal/service"                  // Added logger import
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"                // Corrected import path
	"github.com/golang/mock/gomock"                                  // Added gomock import
	"github.com/stretchr/testify/assert"
	// Keep testify/assert
)

// Updated setup function to use gomock controller
func setupTemplateServiceTest(ctrl *gomock.Controller) (*service.TemplateService, *domainmocks.MockTemplateRepository, *domainmocks.MockAuthService, *pkgmocks.MockLogger) {
	mockRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	templateService := service.NewTemplateService(mockRepo, mockAuthService, mockLogger)
	return templateService.(*service.TemplateService), mockRepo, mockAuthService, mockLogger
}

// Gomock matcher for validating the template passed to CreateTemplate
type templateMatcher struct {
	expected *domain.Template
}

func (m *templateMatcher) Matches(x interface{}) bool {
	tmpl, ok := x.(*domain.Template)
	if !ok {
		return false
	}
	// Check essential fields and that Version is set to 1
	return tmpl.ID == m.expected.ID &&
		tmpl.Name == m.expected.Name &&
		tmpl.Channel == m.expected.Channel &&
		tmpl.Category == m.expected.Category &&
		tmpl.Email != nil &&
		tmpl.Email.Subject == m.expected.Email.Subject &&
		tmpl.Version == 1 // Crucial check
}

func (m *templateMatcher) String() string {
	return fmt.Sprintf("is a template with ID %s and version 1", m.expected.ID)
}

func EqTemplateWithVersion1(expected *domain.Template) gomock.Matcher {
	return &templateMatcher{expected: expected}
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	templateID := "tmpl-abc"

	// Base valid template structure (without ID, CreatedAt, UpdatedAt for create)
	baseTemplate := &domain.Template{
		ID:       templateID, // Service expects ID to be set by caller
		Name:     "Test Template",
		Channel:  "email",
		Category: "transactional",
		Email: &domain.EmailTemplate{
			FromAddress:      "test@example.com",
			FromName:         "Test Sender",
			Subject:          "Test Subject",
			Content:          "<h1>Hello {{name}}</h1>",
			VisualEditorTree: "{}",
		},
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)

		// Create a fresh copy for this test case to avoid mutation across tests
		templateToCreate := *baseTemplate // Shallow copy is fine here

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().CreateTemplate(ctx, workspaceID, EqTemplateWithVersion1(&templateToCreate)).Return(nil)

		err := templateService.CreateTemplate(ctx, workspaceID, &templateToCreate)

		assert.NoError(t, err)
		// Assert Version is set on the object passed to the service
		assert.Equal(t, int64(1), templateToCreate.Version)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		authErr := errors.New("auth error")
		templateToCreate := *baseTemplate // Use a copy

		// Return nil user and error
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, authErr)

		err := templateService.CreateTemplate(ctx, workspaceID, &templateToCreate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
		assert.ErrorIs(t, err, authErr)
	})

	t.Run("Validation Failure - Missing Name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)

		// Create invalid template directly
		invalidTemplate := *baseTemplate // Copy
		invalidTemplate.Name = ""        // Make invalid

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)

		err := templateService.CreateTemplate(ctx, workspaceID, &invalidTemplate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template: name is required")
	})

	t.Run("Validation Failure - Missing Email Details", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)

		invalidTemplate := *baseTemplate // Copy
		invalidTemplate.Email = nil      // Make invalid

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)

		err := templateService.CreateTemplate(ctx, workspaceID, &invalidTemplate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template: email is required")
	})

	t.Run("Repository Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, mockLogger := setupTemplateServiceTest(ctrl)
		repoErr := errors.New("db error")
		templateToCreate := *baseTemplate // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		// Match any template here, as we expect the repo call itself to fail
		mockRepo.EXPECT().CreateTemplate(ctx, workspaceID, gomock.Any()).Return(repoErr)
		// Expect logging on error (match template ID and specific error message)
		mockLogger.EXPECT().WithField("template_id", templateToCreate.ID).Return(mockLogger)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to create template: %v", repoErr)).Return()

		err := templateService.CreateTemplate(ctx, workspaceID, &templateToCreate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create template")
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestTemplateService_GetTemplateByID(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	templateID := "tmpl-abc"
	version := int64(1)
	now := time.Now().UTC()

	expectedTemplate := &domain.Template{
		ID:        templateID,
		Name:      "Test Template",
		Version:   version,
		Channel:   "email",
		Category:  "transactional",
		CreatedAt: now,
		UpdatedAt: now,
		Email: &domain.EmailTemplate{
			FromAddress:      "test@example.com",
			FromName:         "Test Sender",
			Subject:          "Test Subject",
			Content:          "<h1>Hello {{name}}</h1>",
			VisualEditorTree: "{}",
		},
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, version).Return(expectedTemplate, nil)

		template, err := templateService.GetTemplateByID(ctx, workspaceID, templateID, version)

		assert.NoError(t, err)
		assert.Equal(t, expectedTemplate, template)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		authErr := errors.New("auth error")

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, authErr)

		template, err := templateService.GetTemplateByID(ctx, workspaceID, templateID, version)

		assert.Error(t, err)
		assert.Nil(t, template)
		assert.Contains(t, err.Error(), "failed to authenticate user")
		assert.ErrorIs(t, err, authErr)
	})

	t.Run("Not Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		notFoundErr := &domain.ErrTemplateNotFound{Message: "not found"}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, version).Return(nil, notFoundErr)

		template, err := templateService.GetTemplateByID(ctx, workspaceID, templateID, version)

		assert.Error(t, err)
		assert.Nil(t, template)
		assert.ErrorIs(t, err, notFoundErr)
	})

	t.Run("Repository Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, mockLogger := setupTemplateServiceTest(ctrl)
		repoErr := errors.New("db error")

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, version).Return(nil, repoErr)
		mockLogger.EXPECT().WithField("template_id", templateID).Return(mockLogger)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to get template: %v", repoErr)).Return()

		template, err := templateService.GetTemplateByID(ctx, workspaceID, templateID, version)

		assert.Error(t, err)
		assert.Nil(t, template)
		assert.Contains(t, err.Error(), "failed to get template")
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestTemplateService_GetTemplates(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	now := time.Now().UTC()

	expectedTemplates := []*domain.Template{
		{
			ID:        "tmpl-abc",
			Name:      "Test Template 1",
			Version:   1,
			Channel:   "email",
			Category:  "transactional",
			CreatedAt: now,
			UpdatedAt: now,
			Email: &domain.EmailTemplate{
				Subject: "Subject 1",
			},
		},
		{
			ID:        "tmpl-def",
			Name:      "Test Template 2",
			Version:   2,
			Channel:   "email",
			Category:  "marketing",
			CreatedAt: now,
			UpdatedAt: now,
			Email: &domain.EmailTemplate{
				Subject: "Subject 2",
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplates(ctx, workspaceID).Return(expectedTemplates, nil)

		templates, err := templateService.GetTemplates(ctx, workspaceID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTemplates, templates)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		authErr := errors.New("auth error")

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, authErr)

		templates, err := templateService.GetTemplates(ctx, workspaceID)

		assert.Error(t, err)
		assert.Nil(t, templates)
		assert.Contains(t, err.Error(), "failed to authenticate user")
		assert.ErrorIs(t, err, authErr)
	})

	t.Run("Repository Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, mockLogger := setupTemplateServiceTest(ctrl)
		repoErr := errors.New("db error")

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplates(ctx, workspaceID).Return(nil, repoErr)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to get templates: %v", repoErr)).Return()

		templates, err := templateService.GetTemplates(ctx, workspaceID)

		assert.Error(t, err)
		assert.Nil(t, templates)
		assert.Contains(t, err.Error(), "failed to get templates")
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestTemplateService_UpdateTemplate(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	templateID := "tmpl-abc"
	now := time.Now().UTC()

	existingTemplate := &domain.Template{
		ID:        templateID,
		Name:      "Old Name",
		Version:   1,
		Channel:   "email",
		Category:  "transactional",
		CreatedAt: now.Add(-time.Hour), // Ensure CreatedAt is preserved
		UpdatedAt: now.Add(-time.Hour),
		Email: &domain.EmailTemplate{
			FromAddress:      "old@example.com",
			FromName:         "Old Sender",
			Subject:          "Old Subject",
			Content:          "<p>Old</p>",
			VisualEditorTree: "{\"old\": true}",
		},
	}

	updatedTemplateData := &domain.Template{
		ID:       templateID,
		Name:     "New Name", // Updated field
		Channel:  "email",
		Category: "transactional",
		Email: &domain.EmailTemplate{
			FromAddress:      "new@example.com", // Updated field
			FromName:         "New Sender",      // Updated field
			Subject:          "New Subject",     // Updated field
			Content:          "<h1>New</h1>",    // Updated field
			VisualEditorTree: "{\"new\": true}", // Updated field
		},
		// Version, CreatedAt, UpdatedAt should be handled by the service
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		templateToUpdate := *updatedTemplateData // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		// GetByID is called first to check existence and preserve CreatedAt (version 0 means latest)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, int64(0)).Return(existingTemplate, nil)
		// Expect UpdateTemplate call with correct fields preserved/updated
		mockRepo.EXPECT().UpdateTemplate(ctx, workspaceID, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, tmpl *domain.Template) error {
			assert.Equal(t, templateToUpdate.ID, tmpl.ID)
			assert.Equal(t, templateToUpdate.Name, tmpl.Name)
			assert.Equal(t, templateToUpdate.Channel, tmpl.Channel)
			assert.Equal(t, templateToUpdate.Category, tmpl.Category)
			assert.Equal(t, templateToUpdate.Email, tmpl.Email)                       // Check nested struct
			assert.Equal(t, existingTemplate.CreatedAt, tmpl.CreatedAt)               // Check CreatedAt preserved
			assert.WithinDuration(t, time.Now().UTC(), tmpl.UpdatedAt, 5*time.Second) // Check UpdatedAt is recent
			// Version should be handled by the repository layer during update (not checked here)
			return nil
		})

		err := templateService.UpdateTemplate(ctx, workspaceID, &templateToUpdate)

		assert.NoError(t, err)
		// Check that the passed-in template's CreatedAt and UpdatedAt were updated by the service
		assert.Equal(t, existingTemplate.CreatedAt, templateToUpdate.CreatedAt)
		assert.WithinDuration(t, time.Now().UTC(), templateToUpdate.UpdatedAt, 5*time.Second)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		authErr := errors.New("auth error")
		templateToUpdate := *updatedTemplateData // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, authErr)

		err := templateService.UpdateTemplate(ctx, workspaceID, &templateToUpdate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
		assert.ErrorIs(t, err, authErr)
	})

	t.Run("Get Existing Template Not Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		notFoundErr := &domain.ErrTemplateNotFound{Message: "not found"}
		templateToUpdate := *updatedTemplateData // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, int64(0)).Return(nil, notFoundErr)

		err := templateService.UpdateTemplate(ctx, workspaceID, &templateToUpdate)

		assert.Error(t, err)
		assert.ErrorIs(t, err, notFoundErr) // Service should return the exact error
	})

	t.Run("Get Existing Template Repository Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, mockLogger := setupTemplateServiceTest(ctrl)
		repoErr := errors.New("get db error")
		templateToUpdate := *updatedTemplateData // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, int64(0)).Return(nil, repoErr)
		mockLogger.EXPECT().WithField("template_id", templateID).Return(mockLogger)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to check if template exists: %v", repoErr)).Return()

		err := templateService.UpdateTemplate(ctx, workspaceID, &templateToUpdate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check if template exists")
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		invalidTemplate := *updatedTemplateData // Copy
		invalidTemplate.Name = ""               // Make invalid

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		// Expect GetByID to be called and succeed before validation happens
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, int64(0)).Return(existingTemplate, nil)

		err := templateService.UpdateTemplate(ctx, workspaceID, &invalidTemplate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template: name is required")
	})

	t.Run("Update Repository Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, mockLogger := setupTemplateServiceTest(ctrl)
		repoErr := errors.New("update db error")
		templateToUpdate := *updatedTemplateData // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().GetTemplateByID(ctx, workspaceID, templateID, int64(0)).Return(existingTemplate, nil)
		mockRepo.EXPECT().UpdateTemplate(ctx, workspaceID, gomock.Any()).Return(repoErr)
		mockLogger.EXPECT().WithField("template_id", templateID).Return(mockLogger)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to update template: %v", repoErr)).Return()

		err := templateService.UpdateTemplate(ctx, workspaceID, &templateToUpdate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update template")
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	ctx := context.Background()
	workspaceID := "ws-123"
	userID := "user-456"
	templateID := "tmpl-abc"

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().DeleteTemplate(ctx, workspaceID, templateID).Return(nil)

		err := templateService.DeleteTemplate(ctx, workspaceID, templateID)

		assert.NoError(t, err)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		authErr := errors.New("auth error")

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, authErr)

		err := templateService.DeleteTemplate(ctx, workspaceID, templateID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
		assert.ErrorIs(t, err, authErr)
	})

	t.Run("Not Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		notFoundErr := &domain.ErrTemplateNotFound{Message: "not found"}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().DeleteTemplate(ctx, workspaceID, templateID).Return(notFoundErr)

		err := templateService.DeleteTemplate(ctx, workspaceID, templateID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, notFoundErr) // Service should return the exact error
	})

	t.Run("Repository Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, mockLogger := setupTemplateServiceTest(ctrl)
		repoErr := errors.New("db error")

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().DeleteTemplate(ctx, workspaceID, templateID).Return(repoErr)
		mockLogger.EXPECT().WithField("template_id", templateID).Return(mockLogger)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to delete template: %v", repoErr)).Return()

		err := templateService.DeleteTemplate(ctx, workspaceID, templateID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete template")
		assert.ErrorIs(t, err, repoErr)
	})
}
