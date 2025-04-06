package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type TemplateService struct {
	repo        domain.TemplateRepository
	authService domain.AuthService
	logger      logger.Logger
}

func NewTemplateService(repo domain.TemplateRepository, authService domain.AuthService, logger logger.Logger) domain.TemplateService {
	return &TemplateService{
		repo:        repo,
		authService: authService,
		logger:      logger,
	}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, workspaceID string, template *domain.Template) error {
	// Authenticate user for workspace
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Set initial version *before* validation
	template.Version = 1

	// Validate template
	if err := template.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Create template
	if err := s.repo.CreateTemplate(ctx, workspaceID, template); err != nil {
		s.logger.WithField("template_id", template.ID).Error(fmt.Sprintf("Failed to create template: %v", err))
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

func (s *TemplateService) GetTemplateByID(ctx context.Context, workspaceID string, id string, version int64) (*domain.Template, error) {
	// Authenticate user for workspace
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
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

func (s *TemplateService) GetTemplates(ctx context.Context, workspaceID string) ([]*domain.Template, error) {
	// Authenticate user for workspace
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get templates
	templates, err := s.repo.GetTemplates(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get templates: %v", err))
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	return templates, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, workspaceID string, template *domain.Template) error {
	// Authenticate user for workspace
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
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
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
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
