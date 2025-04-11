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
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mjml"
	notifusemjml "github.com/Notifuse/notifuse/pkg/mjml"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks" // Corrected import path
	"github.com/golang/mock/gomock"                   // Added gomock import
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// Keep testify/assert
)

// Updated setup function to use gomock controller
func setupTemplateServiceTest(ctrl *gomock.Controller) (*service.TemplateService, *domainmocks.MockTemplateRepository, *domainmocks.MockAuthService, *pkgmocks.MockLogger) {
	mockRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	templateService := service.NewTemplateService(mockRepo, mockAuthService, mockLogger)
	return templateService, mockRepo, mockAuthService, mockLogger
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
	templateToCreate := &domain.Template{
		ID:       templateID,
		Name:     "Test Template",
		Channel:  "email",
		Category: "transactional",
		Email: &domain.EmailTemplate{
			FromAddress:     "test@example.com",
			FromName:        "Test Sender",
			Subject:         "Test Email",
			CompiledPreview: "<p>Test</p>",
			VisualEditorTree: mjml.EmailBlock{
				Kind: "root",
				Data: map[string]interface{}{"styles": map[string]interface{}{}},
			},
		},
		// Version should be set to 1 by the service
		// CreatedAt and UpdatedAt should be set by the service
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, mockRepo, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		templateToPass := *templateToCreate // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		// Expect CreateTemplate with Version 1 set
		mockRepo.EXPECT().CreateTemplate(ctx, workspaceID, EqTemplateWithVersion1(&templateToPass)).Return(nil)

		err := templateService.CreateTemplate(ctx, workspaceID, &templateToPass)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), templateToPass.Version)
		assert.WithinDuration(t, time.Now().UTC(), templateToPass.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, time.Now().UTC(), templateToPass.UpdatedAt, 5*time.Second)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		authErr := errors.New("auth error")
		templateToPass := *templateToCreate // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, authErr)

		err := templateService.CreateTemplate(ctx, workspaceID, &templateToPass)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate user")
		assert.ErrorIs(t, err, authErr)
	})

	t.Run("Validation Failure - Missing Name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		invalidTemplate := *templateToCreate // Copy
		invalidTemplate.Name = ""            // Make invalid

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)

		err := templateService.CreateTemplate(ctx, workspaceID, &invalidTemplate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template: name is required")
	})

	t.Run("Validation Failure - Missing Email Details", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		templateService, _, mockAuthService, _ := setupTemplateServiceTest(ctrl)
		invalidTemplate := *templateToCreate // Copy
		invalidTemplate.Email = nil          // Make invalid

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
		templateToPass := *templateToCreate // Use a copy

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{ID: userID}, nil)
		mockRepo.EXPECT().CreateTemplate(ctx, workspaceID, gomock.Any()).Return(repoErr)
		mockLogger.EXPECT().WithField("template_id", templateID).Return(mockLogger)
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to create template: %v", repoErr)).Return()

		err := templateService.CreateTemplate(ctx, workspaceID, &templateToPass)

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
			Subject:          "Test Email",
			CompiledPreview:  "<html><body>Test</body></html>",
			VisualEditorTree: mjml.EmailBlock{},
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
	existingCreatedAt := time.Now().Add(-1 * time.Hour).UTC()

	existingTemplate := &domain.Template{
		ID:        templateID,
		Name:      "Old Name",
		Version:   1,
		Channel:   "email",
		Category:  "transactional",
		CreatedAt: existingCreatedAt,
		Email: &domain.EmailTemplate{
			FromAddress:     "old@example.com",
			FromName:        "Old Sender",
			Subject:         "Old Subject",
			CompiledPreview: "<p>Old</p>",
			VisualEditorTree: mjml.EmailBlock{
				Kind: "root",
				Data: map[string]interface{}{"styles": map[string]interface{}{}},
			},
		},
	}

	updatedTemplateData := &domain.Template{
		ID:       templateID,
		Name:     "New Name", // Updated field
		Channel:  "email",
		Category: "transactional",
		Email: &domain.EmailTemplate{
			FromAddress:     "new@example.com", // Updated field
			FromName:        "New Sender",      // Updated field
			Subject:         "New Subject",     // Updated field
			CompiledPreview: "<h1>New</h1>",    // Updated field
			VisualEditorTree: mjml.EmailBlock{
				Kind: "root",
				Data: map[string]interface{}{"styles": map[string]interface{}{}},
			},
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

// Helper types/funcs from other tests or define locally if needed
type MockLogger struct{}

func (l *MockLogger) Debug(msg string)                                       {}
func (l *MockLogger) Info(msg string)                                        {}
func (l *MockLogger) Warn(msg string)                                        {}
func (l *MockLogger) Error(msg string)                                       {}
func (l *MockLogger) Fatal(msg string)                                       {}
func (l *MockLogger) WithField(key string, value interface{}) logger.Logger  { return l }
func (l *MockLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }

// --- Helper to create a basic text block ---
func createTestTextBlock(id, textContent string) notifusemjml.EmailBlock {
	return notifusemjml.EmailBlock{
		ID:   id,
		Kind: "text",
		Data: map[string]interface{}{
			"align": "left",
			"editorData": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"children": []interface{}{
						map[string]interface{}{"text": textContent},
					},
				},
			},
		},
	}
}

// --- Helper to create a valid nested structure for testing success ---
func createValidTestTree(textBlock notifusemjml.EmailBlock) notifusemjml.EmailBlock {
	columnBlock := notifusemjml.EmailBlock{
		ID:       "col1",
		Kind:     "column",
		Data:     map[string]interface{}{"styles": map[string]interface{}{"verticalAlign": "top"}},
		Children: []notifusemjml.EmailBlock{textBlock},
	}
	sectionBlock := notifusemjml.EmailBlock{
		ID:       "sec1",
		Kind:     "oneColumn", // Acts as mj-section
		Data:     map[string]interface{}{"styles": map[string]interface{}{"textAlign": "left"}},
		Children: []notifusemjml.EmailBlock{columnBlock},
	}
	rootStyles := map[string]interface{}{ // Basic root styles
		"body":      map[string]interface{}{"width": "600px", "backgroundColor": "#ffffff"},
		"paragraph": map[string]interface{}{"color": "#000000", "fontSize": "16px", "margin": "0px", "fontFamily": "Arial"},
	}
	return notifusemjml.EmailBlock{
		ID: "root", Kind: "root", Data: map[string]interface{}{"styles": rootStyles}, Children: []notifusemjml.EmailBlock{sectionBlock},
	}
}

func TestCompileTemplate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockLogger := &MockLogger{}

	svc := service.NewTemplateService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "ws_123"
	userID := "user_abc"
	testTree := createValidTestTree(createTestTextBlock("txt1", "Hello {{name}}"))
	testData := domain.MapOfAny{"name": "Tester"}

	// Mock expectations
	mockAuthService.EXPECT().AuthenticateUserForWorkspace(gomock.Any(), workspaceID).Return(&domain.User{ID: userID}, nil)

	// --- Act ---
	resp, err := svc.CompileTemplate(ctx, workspaceID, testTree, testData)

	// --- Assert ---
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.True(t, resp.Success, "Success should be true")
	assert.Nil(t, resp.Error, "Error should be nil on success")
	require.NotNil(t, resp.MJML, "MJML should not be nil on success")
	require.NotNil(t, resp.HTML, "HTML should not be nil on success")

	assert.Contains(t, *resp.MJML, "<mj-section", "MJML should contain <mj-section>")
	assert.Contains(t, *resp.MJML, "<mj-column", "MJML should contain <mj-column>")
	assert.Contains(t, *resp.MJML, "<mj-text", "MJML should contain <mj-text>")
	assert.Contains(t, *resp.MJML, "Hello Tester", "MJML should contain processed liquid variable")

	assert.Contains(t, *resp.HTML, "<html", "HTML should contain <html>")
	assert.Contains(t, *resp.HTML, "Hello Tester", "HTML should contain processed liquid variable")

	// t.Logf("Generated MJML:\n%s", *resp.MJML)
	// t.Logf("Generated HTML:\n%s", *resp.HTML)
}

// Renamed test to focus on TreeToMjml internal errors (like Liquid)
func TestCompileTemplate_TreeToMjmlError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockLogger := &MockLogger{}
	svc := service.NewTemplateService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "ws_123"
	userID := "user_abc"

	// Create a tree containing a block that will cause TreeToMjml to return an error (e.g., bad liquid)
	badLiquidBlock := notifusemjml.EmailBlock{
		ID: "badliq", Kind: "liquid", Data: map[string]interface{}{"liquidCode": "{% invalid tag %}"},
	}
	badLiquidTree := createValidTestTree(badLiquidBlock) // Embed the bad block in a valid structure

	// Mock Auth
	mockAuthService.EXPECT().AuthenticateUserForWorkspace(gomock.Any(), workspaceID).Return(&domain.User{ID: userID}, nil)

	// --- Act ---
	resp, err := svc.CompileTemplate(ctx, workspaceID, badLiquidTree, nil)

	// --- Assert ---
	require.Error(t, err, "Expected a standard Go error for TreeToMjml failure (bad liquid)")
	require.Nil(t, resp, "Response should be nil when a standard Go error occurs")
	assert.Contains(t, err.Error(), "failed to generate MJML from tree", "Error should indicate MJML generation failure")
	assert.Contains(t, err.Error(), "liquid rendering error", "Error should wrap the liquid error")

	// Note: Testing the specific mjmlgo.Error path (where err is nil but resp.Success is false)
	// would ideally involve mocking mjmlgo.ToHTML or using specific input known to cause mjmlgo.Error.
}

func TestCompileTemplate_AuthError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockLogger := &MockLogger{}

	svc := service.NewTemplateService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "ws_123"
	// Use a valid tree for this auth error test
	testTree := createValidTestTree(createTestTextBlock("txt1", "Test"))
	authErr := errors.New("authentication failed")

	// Mock expectations
	mockAuthService.EXPECT().AuthenticateUserForWorkspace(gomock.Any(), workspaceID).Return(nil, authErr)

	// --- Act ---
	resp, err := svc.CompileTemplate(ctx, workspaceID, testTree, nil)

	// --- Assert ---
	require.Error(t, err)
	require.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to authenticate user", "Error message should indicate auth failure")
	assert.ErrorIs(t, err, authErr, "Original auth error should be wrapped")
}

func TestCompileTemplate_InvalidTreeData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockRepo := domainmocks.NewMockTemplateRepository(ctrl)
	mockLogger := &MockLogger{}

	svc := service.NewTemplateService(mockRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "ws_123"
	userID := "user_abc"
	invalidTree := notifusemjml.EmailBlock{
		ID: "root_invalid", Kind: "root", Data: nil, Children: nil,
	}

	// Mock expectations
	mockAuthService.EXPECT().AuthenticateUserForWorkspace(gomock.Any(), workspaceID).Return(&domain.User{ID: userID}, nil)

	// --- Act ---
	resp, err := svc.CompileTemplate(ctx, workspaceID, invalidTree, nil)

	// --- Assert ---
	require.Error(t, err)
	require.Nil(t, resp)
	// Update assertion to match the actual error message when Data is nil
	assert.Contains(t, err.Error(), "invalid root block data format", "Error message should indicate invalid root data")

}
