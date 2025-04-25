package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mjmlgo "github.com/Boostport/mjml-go"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	notifusemjml "github.com/Notifuse/notifuse/pkg/mjml"
)

type TemplateService struct {
	repo        domain.TemplateRepository
	authService domain.AuthService
	logger      logger.Logger
}

func NewTemplateService(repo domain.TemplateRepository, authService domain.AuthService, logger logger.Logger) *TemplateService {
	return &TemplateService{
		repo:        repo,
		authService: authService,
		logger:      logger,
	}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, workspaceID string, template *domain.Template) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Set initial version and timestamps
	template.Version = 1
	now := time.Now().UTC()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Validate template after setting required fields
	if err := template.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Create template in repository
	if err := s.repo.CreateTemplate(ctx, workspaceID, template); err != nil {
		s.logger.WithField("template_id", template.ID).Error(fmt.Sprintf("Failed to create template: %v", err))
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

func (s *TemplateService) GetTemplateByID(ctx context.Context, workspaceID string, id string, version int64) (*domain.Template, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get template by ID
	template, err := s.repo.GetTemplateByID(ctx, workspaceID, id, version)
	if err != nil {
		if _, ok := err.(*domain.ErrTemplateNotFound); ok {
			return nil, err
		}
		s.logger.WithField("template_id", id).Error(fmt.Sprintf("Failed to get template: %v", err))
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return template, nil
}

func (s *TemplateService) GetTemplates(ctx context.Context, workspaceID string, category string) ([]*domain.Template, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get templates
	templates, err := s.repo.GetTemplates(ctx, workspaceID, category)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get templates: %v", err))
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	return templates, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, workspaceID string, template *domain.Template) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Check if template exists
	existingTemplate, err := s.repo.GetTemplateByID(ctx, workspaceID, template.ID, 0)
	if err != nil {
		if _, ok := err.(*domain.ErrTemplateNotFound); ok {
			return err
		}
		s.logger.WithField("template_id", template.ID).Error(fmt.Sprintf("Failed to check if template exists: %v", err))
		return fmt.Errorf("failed to check if template exists: %w", err)
	}

	// Set version from existing template *before* validation to satisfy the check
	template.Version = existingTemplate.Version

	// Validate template
	if err := template.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Preserve creation time from existing template
	template.CreatedAt = existingTemplate.CreatedAt
	template.UpdatedAt = time.Now().UTC()

	// Update template (this will create a new version in the repo)
	if err := s.repo.UpdateTemplate(ctx, workspaceID, template); err != nil {
		s.logger.WithField("template_id", template.ID).Error(fmt.Sprintf("Failed to update template: %v", err))
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, workspaceID string, id string) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Delete template
	if err := s.repo.DeleteTemplate(ctx, workspaceID, id); err != nil {
		if _, ok := err.(*domain.ErrTemplateNotFound); ok {
			return err
		}
		s.logger.WithField("template_id", id).Error(fmt.Sprintf("Failed to delete template: %v", err))
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

func (s *TemplateService) CompileTemplate(ctx context.Context, workspaceID string, tree notifusemjml.EmailBlock, testData domain.MapOfAny) (*domain.CompileTemplateResponse, error) {
	// Check if user is already authenticated in context
	if user := ctx.Value("authenticated_user"); user == nil {
		// Authenticate user for workspace
		var user *domain.User
		var err error
		ctx, user, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
		if err != nil {
			// Return standard Go error for non-compilation issues
			return nil, fmt.Errorf("failed to authenticate user: %w", err)
		}

		// Store user in context for future use
		ctx = context.WithValue(ctx, "authenticated_user", user)
	}

	// Extract root styles from the tree data
	rootDataMap, ok := tree.Data.(map[string]interface{})
	if !ok {
		s.logger.Error("CompileTemplate: Root block data is not a map")
		// Return standard Go error for non-compilation issues
		return nil, fmt.Errorf("invalid root block data format")
	}
	rootStyles, _ := rootDataMap["styles"].(map[string]interface{})
	if rootStyles == nil {
		s.logger.Error("CompileTemplate: Root block styles are missing")
		// Return standard Go error for non-compilation issues
		return nil, fmt.Errorf("root block styles are required for compilation")
	}

	// Prepare template data JSON string
	var templateDataStr string
	if testData != nil && len(testData) > 0 {
		jsonDataBytes, err := json.Marshal(testData)
		if err != nil {
			s.logger.WithField("error", err).Error("Failed to marshal test_data to JSON")
			// Return standard Go error for non-compilation issues
			return nil, fmt.Errorf("failed to marshal test_data: %w", err)
		}
		templateDataStr = string(jsonDataBytes)
	}

	// Compile tree to MJML using our pkg/mjml function
	mjmlResult, err := notifusemjml.TreeToMjml(rootStyles, tree, templateDataStr, map[string]string{}, 0, nil)
	if err != nil {
		return &domain.CompileTemplateResponse{
			Success: false,
			MJML:    nil,
			HTML:    nil,
			Error: &mjmlgo.Error{
				Message: err.Error(),
			},
		}, nil
	}

	// Compile MJML to HTML using mjml-go library
	htmlResult, err := mjmlgo.ToHTML(ctx, mjmlResult)
	if err != nil {
		// Return the response struct with Success=false and the Error details
		return &domain.CompileTemplateResponse{
			Success: false,
			MJML:    &mjmlResult, // Include original MJML for context if desired
			HTML:    nil,
			Error: &mjmlgo.Error{
				Message: err.Error(),
			},
		}, nil
	}

	// Return successful response
	return &domain.CompileTemplateResponse{
		Success: true,
		MJML:    &mjmlResult,
		HTML:    &htmlResult,
		Error:   nil,
	}, nil
}
