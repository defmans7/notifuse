package broadcast

import (
	"context"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

//go:generate mockgen -destination=./mocks/mock_template_loader.go -package=mocks github.com/Notifuse/notifuse/internal/service/broadcast TemplateLoader

// TemplateLoader is the interface for loading templates for broadcasts
type TemplateLoader interface {
	// LoadTemplatesForBroadcast loads all templates for a broadcast's variations
	LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error)

	// ValidateTemplates validates that the required templates are loaded and valid
	ValidateTemplates(templates map[string]*domain.Template) error
}

// templateLoader implements the TemplateLoader interface
type templateLoader struct {
	broadcastService domain.BroadcastSender
	templateService  domain.TemplateService
	logger           logger.Logger
	config           *Config
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(broadcastService domain.BroadcastSender, templateService domain.TemplateService, logger logger.Logger, config *Config) TemplateLoader {
	if config == nil {
		config = DefaultConfig()
	}
	return &templateLoader{
		broadcastService: broadcastService,
		templateService:  templateService,
		logger:           logger,
		config:           config,
	}
}

// LoadTemplatesForBroadcast loads all templates for a broadcast's variations
func (l *templateLoader) LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error) {
	startTime := time.Now()
	defer func() {
		l.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Template loading completed")
	}()

	// Get the broadcast to access its template variations
	broadcast, err := l.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		l.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for templates")
		return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Process the broadcast's variations to get template IDs
	templateIDs := make(map[string]bool)
	for _, variation := range broadcast.TestSettings.Variations {
		templateIDs[variation.TemplateID] = true
	}

	if len(templateIDs) == 0 {
		l.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Error("No template variations found in broadcast")
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no template variations found in broadcast", false, nil)
	}

	// Load all templates
	templates := make(map[string]*domain.Template)
	for templateID := range templateIDs {
		template, err := l.templateService.GetTemplateByID(ctx, workspaceID, templateID, 1) // Always use version 1
		if err != nil {
			l.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcastID,
				"workspace_id": workspaceID,
				"template_id":  templateID,
				"error":        err.Error(),
			}).Error("Failed to load template for broadcast")
			continue // Don't fail the whole broadcast for one template
		}
		templates[templateID] = template
	}

	// Validate that we found at least one template
	if len(templates) == 0 {
		l.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Error("No valid templates found for broadcast")
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no valid templates found for broadcast", false, nil)
	}

	l.logger.WithFields(map[string]interface{}{
		"broadcast_id":    broadcastID,
		"workspace_id":    workspaceID,
		"template_count":  len(templates),
		"variation_count": len(broadcast.TestSettings.Variations),
	}).Info("Templates loaded for broadcast")

	return templates, nil
}

// ValidateTemplates validates that the required templates are loaded and valid
func (l *templateLoader) ValidateTemplates(templates map[string]*domain.Template) error {
	if len(templates) == 0 {
		return NewBroadcastError(ErrCodeTemplateMissing, "no templates provided for validation", false, nil)
	}

	// Validate each template
	for id, template := range templates {
		if template == nil {
			return NewBroadcastError(ErrCodeTemplateInvalid, "template is nil", false, nil)
		}

		// Ensure the template has the required fields for sending emails
		if template.Email == nil {
			l.logger.WithField("template_id", id).Error("Template missing email configuration")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing email configuration", false, nil)
		}

		if template.Email.FromAddress == "" {
			l.logger.WithField("template_id", id).Error("Template missing from address")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing from address", false, nil)
		}

		if template.Email.Subject == "" {
			l.logger.WithField("template_id", id).Error("Template missing subject")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing subject", false, nil)
		}

		if template.Email.VisualEditorTree.Kind == "" {
			l.logger.WithField("template_id", id).Error("Template missing content")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing content", false, nil)
		}
	}

	return nil
}
