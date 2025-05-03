package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// TransactionalNotificationService provides operations for managing and sending transactional notifications
type TransactionalNotificationService struct {
	transactionalRepo  domain.TransactionalNotificationRepository
	messageHistoryRepo domain.MessageHistoryRepository
	templateService    domain.TemplateService
	contactService     domain.ContactService
	emailService       *EmailService
	logger             logger.Logger
}

// NewTransactionalNotificationService creates a new instance of the transactional notification service
func NewTransactionalNotificationService(
	transactionalRepo domain.TransactionalNotificationRepository,
	messageHistoryRepo domain.MessageHistoryRepository,
	templateService domain.TemplateService,
	contactService domain.ContactService,
	emailService *EmailService,
	logger logger.Logger,
) domain.TransactionalNotificationService {
	return &TransactionalNotificationService{
		transactionalRepo:  transactionalRepo,
		messageHistoryRepo: messageHistoryRepo,
		templateService:    templateService,
		contactService:     contactService,
		emailService:       emailService,
		logger:             logger,
	}
}

// CreateNotification creates a new transactional notification
func (s *TransactionalNotificationService) CreateNotification(
	ctx context.Context,
	workspace string,
	params domain.TransactionalNotificationCreateParams,
) (*domain.TransactionalNotification, error) {
	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        params.ID,
		"name":      params.Name,
	}).Debug("Creating new transactional notification")

	// Create the notification object
	notification := &domain.TransactionalNotification{
		ID:          params.ID,
		Name:        params.Name,
		Description: params.Description,
		Channels:    params.Channels,
		Status:      params.Status,
		Metadata:    params.Metadata,
	}

	// Validate the notification templates exist
	for channel, template := range notification.Channels {
		_, err := s.templateService.GetTemplateByID(ctx, workspace, template.TemplateID, int64(template.Version))
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":       err.Error(),
				"channel":     channel,
				"template_id": template.TemplateID,
				"version":     template.Version,
			}).Error("Invalid template for channel")
			return nil, fmt.Errorf("invalid template for channel %s: %w", channel, err)
		}
	}

	// Save the notification to the repository
	if err := s.transactionalRepo.Create(ctx, workspace, notification); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        notification.ID,
		}).Error("Failed to create notification")
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        notification.ID,
		"name":      notification.Name,
	}).Info("Transactional notification created successfully")
	return notification, nil
}

// UpdateNotification updates an existing transactional notification
func (s *TransactionalNotificationService) UpdateNotification(
	ctx context.Context,
	workspace, id string,
	params domain.TransactionalNotificationUpdateParams,
) (*domain.TransactionalNotification, error) {
	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Debug("Updating transactional notification")

	// Get the existing notification
	notification, err := s.transactionalRepo.Get(ctx, workspace, id)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        id,
		}).Error("Failed to get notification for update")
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Update fields if provided
	if params.Name != "" {
		notification.Name = params.Name
	}
	if params.Description != "" {
		notification.Description = params.Description
	}
	if params.Channels != nil {
		// Validate the updated templates exist
		for channel, template := range params.Channels {
			_, err := s.templateService.GetTemplateByID(ctx, workspace, template.TemplateID, int64(template.Version))
			if err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":       err.Error(),
					"channel":     channel,
					"template_id": template.TemplateID,
					"version":     template.Version,
				}).Error("Invalid template for channel in update")
				return nil, fmt.Errorf("invalid template for channel %s: %w", channel, err)
			}
		}
		notification.Channels = params.Channels
	}
	if params.Status != "" {
		notification.Status = params.Status
	}
	if params.Metadata != nil {
		notification.Metadata = params.Metadata
	}

	// Save the updated notification
	if err := s.transactionalRepo.Update(ctx, workspace, notification); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        notification.ID,
		}).Error("Failed to update notification")
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        notification.ID,
	}).Info("Transactional notification updated successfully")
	return notification, nil
}

// GetNotification retrieves a transactional notification by ID
func (s *TransactionalNotificationService) GetNotification(
	ctx context.Context,
	workspace, id string,
) (*domain.TransactionalNotification, error) {
	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Debug("Retrieving transactional notification")

	notification, err := s.transactionalRepo.Get(ctx, workspace, id)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        id,
		}).Error("Failed to get notification")
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return notification, nil
}

// ListNotifications retrieves all transactional notifications with optional filtering
func (s *TransactionalNotificationService) ListNotifications(
	ctx context.Context,
	workspace string,
	filter map[string]interface{},
	limit, offset int,
) ([]*domain.TransactionalNotification, int, error) {
	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"limit":     limit,
		"offset":    offset,
		"filter":    filter,
	}).Debug("Listing transactional notifications")

	notifications, total, err := s.transactionalRepo.List(ctx, workspace, filter, limit, offset)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
		}).Error("Failed to list notifications")
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"count":     len(notifications),
		"total":     total,
	}).Debug("Successfully retrieved notifications list")
	return notifications, total, nil
}

// DeleteNotification soft-deletes a transactional notification
func (s *TransactionalNotificationService) DeleteNotification(
	ctx context.Context,
	workspace, id string,
) error {
	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Debug("Deleting transactional notification")

	if err := s.transactionalRepo.Delete(ctx, workspace, id); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        id,
		}).Error("Failed to delete notification")
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"id":        id,
	}).Info("Transactional notification deleted successfully")
	return nil
}

// SendNotification sends a transactional notification to a contact
func (s *TransactionalNotificationService) SendNotification(
	ctx context.Context,
	workspace string,
	params domain.TransactionalNotificationSendParams,
) (string, error) {
	// Get the notification
	notification, err := s.transactionalRepo.Get(ctx, workspace, params.ID)
	if err != nil {
		return "", fmt.Errorf("notification not found: %w", err)
	}

	// Check if the notification is active
	if notification.Status != domain.TransactionalStatusActive {
		return "", fmt.Errorf("notification is not active")
	}

	// Upsert the contact first
	if params.Contact == nil {
		return "", fmt.Errorf("contact is required")
	}

	contactOperation := s.contactService.UpsertContact(ctx, workspace, params.Contact)
	if contactOperation.Action == domain.UpsertContactOperationError {
		return "", fmt.Errorf("failed to upsert contact: %s", contactOperation.Error)
	}

	// Get the contact with complete information
	contact, err := s.contactService.GetContactByEmail(ctx, workspace, params.Contact.Email)
	if err != nil {
		return "", fmt.Errorf("contact not found after upsert: %w", err)
	}

	// Determine which channels to send through
	channelsToSend := make(map[domain.TransactionalChannel]struct{})
	if len(params.Channels) > 0 {
		// Use the specified channels
		for _, channel := range params.Channels {
			if _, ok := notification.Channels[channel]; ok {
				channelsToSend[channel] = struct{}{}
			}
		}
	} else {
		// Use all configured channels
		for channel := range notification.Channels {
			channelsToSend[channel] = struct{}{}
		}
	}

	if len(channelsToSend) == 0 {
		return "", fmt.Errorf("no valid channels to send notification")
	}

	// Create message history entry
	messageID := uuid.New().String()
	successfulChannels := 0

	// Process each channel
	for channel := range channelsToSend {
		templateConfig := notification.Channels[channel]

		// Prepare message data with contact and custom data
		messageData := domain.MessageData{
			Data: map[string]interface{}{
				"contact": contact,
			},
		}

		// Add custom data if provided
		if params.Data != nil {
			for key, value := range params.Data {
				messageData.Data[key] = value
			}
		}

		// Add metadata if provided
		if params.Metadata != nil {
			messageData.Metadata = params.Metadata
		}

		// Send the message based on channel type
		if channel == domain.TransactionalChannelEmail {
			err = s.sendEmailNotification(
				ctx,
				workspace,
				messageID,
				contact,
				templateConfig,
				messageData,
			)
			if err == nil {
				successfulChannels++
			} else {
				// Log the error but continue with other channels
				s.logger.WithFields(map[string]interface{}{
					"error":        err.Error(),
					"channel":      channel,
					"notification": notification.ID,
					"contact":      contact.Email,
					"message_id":   messageID,
				}).Error("Failed to send email notification")
			}
		}
		// Add other channel handling here as needed
	}

	if successfulChannels == 0 {
		return "", fmt.Errorf("failed to send notification through any channel")
	}

	return messageID, nil
}

// sendEmailNotification handles sending through the email channel
func (s *TransactionalNotificationService) sendEmailNotification(
	ctx context.Context,
	workspace string,
	messageID string,
	contact *domain.Contact,
	templateConfig domain.ChannelTemplate,
	messageData domain.MessageData,
) error {
	s.logger.WithFields(map[string]interface{}{
		"workspace":    workspace,
		"message_id":   messageID,
		"contact":      contact.Email,
		"template_id":  templateConfig.TemplateID,
		"template_ver": templateConfig.Version,
	}).Debug("Preparing to send email notification")

	// Get the template
	template, err := s.templateService.GetTemplateByID(ctx, workspace, templateConfig.TemplateID, int64(templateConfig.Version))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":       err.Error(),
			"template_id": templateConfig.TemplateID,
			"version":     templateConfig.Version,
		}).Error("Failed to get template")
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Compile the template with the message data
	compiledTemplate, err := s.templateService.CompileTemplate(ctx, workspace, template.Email.VisualEditorTree, messageData.Data)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":       err.Error(),
			"template_id": templateConfig.TemplateID,
			"version":     templateConfig.Version,
		}).Error("Failed to compile template")
		return fmt.Errorf("failed to compile template: %w", err)
	}

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Unknown error"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.WithField("error", errMsg).Error("Template compilation failed")
		return fmt.Errorf("template compilation failed: %s", errMsg)
	}

	// Get necessary email information from the template
	fromEmail := template.Email.FromAddress
	fromName := template.Email.FromName
	subject := template.Email.Subject
	htmlContent := *compiledTemplate.HTML

	// Create message history record
	messageHistory := &domain.MessageHistory{
		ID:              messageID,
		ContactID:       contact.Email,
		TemplateID:      templateConfig.TemplateID,
		TemplateVersion: templateConfig.Version,
		Channel:         "email",
		Status:          domain.MessageStatusSent,
		MessageData:     messageData,
		SentAt:          time.Now().UTC(),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Save to message history
	if err := s.messageHistoryRepo.Create(ctx, workspace, messageHistory); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"message_id": messageID,
		}).Error("Failed to create message history")
		return fmt.Errorf("failed to create message history: %w", err)
	}

	// Send the email using the email service
	s.logger.WithFields(map[string]interface{}{
		"to":         contact.Email,
		"from":       fromEmail,
		"subject":    subject,
		"message_id": messageID,
	}).Debug("Sending email")

	err = s.emailService.SendEmail(
		ctx,
		workspace,
		"transactional", // Use transactional provider type
		fromEmail,
		fromName,
		contact.Email, // To address
		subject,
		htmlContent,
	)

	if err != nil {
		// Update message history with error status
		messageHistory.Status = domain.MessageStatusFailed
		messageHistory.UpdatedAt = time.Now().UTC()
		errorMsg := err.Error()
		messageHistory.Error = &errorMsg

		// Attempt to update the message history record
		updateErr := s.messageHistoryRepo.Update(ctx, workspace, messageHistory)
		if updateErr != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      updateErr.Error(),
				"message_id": messageID,
			}).Error("Failed to update message history with error status")
		}

		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"message_id": messageID,
			"to":         contact.Email,
		}).Error("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"message_id": messageID,
		"to":         contact.Email,
	}).Info("Email sent successfully")
	return nil
}
