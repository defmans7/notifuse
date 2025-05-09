package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/tracing"
	"go.opencensus.io/trace"
)

// TransactionalNotificationService provides operations for managing and sending transactional notifications
type TransactionalNotificationService struct {
	transactionalRepo  domain.TransactionalNotificationRepository
	messageHistoryRepo domain.MessageHistoryRepository
	templateService    domain.TemplateService
	contactService     domain.ContactService
	emailService       domain.EmailServiceInterface
	logger             logger.Logger
	workspaceRepo      domain.WorkspaceRepository
	apiEndpoint        string
}

// NewTransactionalNotificationService creates a new instance of the transactional notification service
func NewTransactionalNotificationService(
	transactionalRepo domain.TransactionalNotificationRepository,
	messageHistoryRepo domain.MessageHistoryRepository,
	templateService domain.TemplateService,
	contactService domain.ContactService,
	emailService domain.EmailServiceInterface,
	logger logger.Logger,
	workspaceRepo domain.WorkspaceRepository,
	apiEndpoint string,
) *TransactionalNotificationService {
	return &TransactionalNotificationService{
		transactionalRepo:  transactionalRepo,
		messageHistoryRepo: messageHistoryRepo,
		templateService:    templateService,
		contactService:     contactService,
		emailService:       emailService,
		logger:             logger,
		workspaceRepo:      workspaceRepo,
		apiEndpoint:        apiEndpoint,
	}
}

// CreateNotification creates a new transactional notification
func (s *TransactionalNotificationService) CreateNotification(
	ctx context.Context,
	workspace string,
	params domain.TransactionalNotificationCreateParams,
) (*domain.TransactionalNotification, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "CreateNotification")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspace),
		trace.StringAttribute("notification_id", params.ID),
		trace.StringAttribute("notification_name", params.Name),
	)

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
		Metadata:    params.Metadata,
	}

	// Validate the notification templates exist
	for channel, template := range notification.Channels {
		tracing.AddAttribute(ctx, fmt.Sprintf("channel.%s.template_id", channel), template.TemplateID)

		_, err := s.templateService.GetTemplateByID(ctx, workspace, template.TemplateID, 0)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":       err.Error(),
				"channel":     channel,
				"template_id": template.TemplateID,
			}).Error("Invalid template for channel")

			tracing.MarkSpanError(ctx, err)
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

		tracing.MarkSpanError(ctx, err)
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
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "UpdateNotification")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspace),
		trace.StringAttribute("notification_id", id),
	)

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

		tracing.MarkSpanError(ctx, err)
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Add existing details to span
	span.AddAttributes(
		trace.StringAttribute("notification.name", notification.Name),
	)

	// Update fields if provided
	if params.Name != "" {
		tracing.AddAttribute(ctx, "update.name", params.Name)
		notification.Name = params.Name
	}
	if params.Description != "" {
		tracing.AddAttribute(ctx, "update.description", "updated")
		notification.Description = params.Description
	}
	if params.Channels != nil {
		tracing.AddAttribute(ctx, "update.channels", "updated")

		// Validate the updated templates exist
		for channel, template := range params.Channels {
			tracing.AddAttribute(ctx, fmt.Sprintf("channel.%s.template_id", channel), template.TemplateID)

			_, err := s.templateService.GetTemplateByID(ctx, workspace, template.TemplateID, int64(0))
			if err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":       err.Error(),
					"channel":     channel,
					"template_id": template.TemplateID,
				}).Error("Invalid template for channel in update")

				tracing.MarkSpanError(ctx, err)
				return nil, fmt.Errorf("invalid template for channel %s: %w", channel, err)
			}
		}
		notification.Channels = params.Channels
	}
	if params.Metadata != nil {
		tracing.AddAttribute(ctx, "update.metadata", "updated")
		notification.Metadata = params.Metadata
	}

	// Save the updated notification
	if err := s.transactionalRepo.Update(ctx, workspace, notification); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"workspace": workspace,
			"id":        notification.ID,
		}).Error("Failed to update notification")

		tracing.MarkSpanError(ctx, err)
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
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "GetNotification")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspace),
		trace.StringAttribute("notification_id", id),
	)

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

		tracing.MarkSpanError(ctx, err)
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Add notification details to span
	span.AddAttributes(
		trace.StringAttribute("notification.name", notification.Name),
		trace.Int64Attribute("notification.channels_count", int64(len(notification.Channels))),
	)

	return notification, nil
}

// ListNotifications retrieves all transactional notifications with optional filtering
func (s *TransactionalNotificationService) ListNotifications(
	ctx context.Context,
	workspace string,
	filter map[string]interface{},
	limit, offset int,
) ([]*domain.TransactionalNotification, int, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "ListNotifications")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspace),
		trace.Int64Attribute("limit", int64(limit)),
		trace.Int64Attribute("offset", int64(offset)),
	)

	// Add filter keys to span
	if filter != nil {
		filterKeys := make([]string, 0, len(filter))
		for k := range filter {
			filterKeys = append(filterKeys, k)
		}
		tracing.AddAttribute(ctx, "filter.keys", filterKeys)
	}

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

		tracing.MarkSpanError(ctx, err)
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace": workspace,
		"count":     len(notifications),
		"total":     total,
	}).Debug("Successfully retrieved notifications list")

	span.AddAttributes(
		trace.Int64Attribute("result.count", int64(len(notifications))),
		trace.Int64Attribute("result.total", int64(total)),
	)

	return notifications, total, nil
}

// DeleteNotification soft-deletes a transactional notification
func (s *TransactionalNotificationService) DeleteNotification(
	ctx context.Context,
	workspace, id string,
) error {
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "DeleteNotification")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspace),
		trace.StringAttribute("notification_id", id),
	)

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

		tracing.MarkSpanError(ctx, err)
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
	workspaceID string,
	params domain.TransactionalNotificationSendParams,
) (string, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "SendNotification")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspaceID),
		trace.StringAttribute("notification_id", params.ID),
	)

	// Add contact info to span if available
	if params.Contact != nil {
		span.AddAttributes(
			trace.StringAttribute("contact.email", params.Contact.Email),
		)
	}

	// Add channel info to span
	if len(params.Channels) > 0 {
		channelList := make([]string, 0, len(params.Channels))
		for _, ch := range params.Channels {
			channelList = append(channelList, string(ch))
		}
		tracing.AddAttribute(ctx, "channels", channelList)
	}

	// Get the workspace to retrieve email provider settings
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get the notification
	notification, err := s.transactionalRepo.Get(ctx, workspaceID, params.ID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		return "", fmt.Errorf("notification not found: %w", err)
	}

	span.AddAttributes(
		trace.StringAttribute("notification.name", notification.Name),
	)

	// Upsert the contact first
	if params.Contact == nil {
		err := fmt.Errorf("contact is required")
		tracing.MarkSpanError(ctx, err)
		return "", err
	}

	contactOperation := s.contactService.UpsertContact(ctx, workspaceID, params.Contact)
	if contactOperation.Action == domain.UpsertContactOperationError {
		err := fmt.Errorf("failed to upsert contact: %s", contactOperation.Error)
		tracing.MarkSpanError(ctx, err)
		return "", err
	}

	tracing.AddAttribute(ctx, "contact.operation", string(contactOperation.Action))

	// Get the contact with complete information
	contact, err := s.contactService.GetContactByEmail(ctx, workspaceID, params.Contact.Email)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
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
		err := fmt.Errorf("no valid channels to send notification")
		tracing.MarkSpanError(ctx, err)
		return "", err
	}

	// Create message history entry
	messageID := uuid.New().String()
	successfulChannels := 0

	span.AddAttributes(
		trace.StringAttribute("message_id", messageID),
		trace.Int64Attribute("channels_to_send", int64(len(channelsToSend))),
	)

	// Process each channel
	for channel := range channelsToSend {
		templateConfig := notification.Channels[channel]

		childCtx, childSpan := tracing.StartSpan(ctx, fmt.Sprintf("Send.%s", channel))
		childSpan.AddAttributes(
			trace.StringAttribute("channel", string(channel)),
			trace.StringAttribute("template_id", templateConfig.TemplateID),
		)

		// Prepare message data with contact and custom data
		apiEndpoint := s.apiEndpoint // Use the service's configured API endpoint

		contactWithList := domain.ContactWithList{
			Contact: contact,
		}

		templateData, err := domain.BuildTemplateData(workspace.ID, contactWithList, messageID, apiEndpoint, nil)
		if err != nil {
			tracing.MarkSpanError(childCtx, err)
			childSpan.End()
			return "", fmt.Errorf("failed to build template data: %w", err)
		}

		// Add custom data if provided
		if params.Data != nil {
			for key, value := range params.Data {
				templateData[key] = value
			}
		}

		messageData := domain.MessageData{
			Data: templateData,
		}

		// Add metadata if provided
		if params.Metadata != nil {
			messageData.Metadata = params.Metadata
		}

		// Send the message based on channel type
		if channel == domain.TransactionalChannelEmail {

			// Get the email provider using the workspace's GetEmailProvider method
			emailProvider, err := workspace.GetEmailProvider(false)
			if err != nil {
				tracing.MarkSpanError(childCtx, err)
				childSpan.End()
				return "", err
			}

			// Validate that the provider is configured
			if emailProvider == nil || emailProvider.Kind == "" {
				err := fmt.Errorf("no email provider configured for transactional notifications")
				tracing.MarkSpanError(childCtx, err)
				childSpan.End()
				return "", err
			}

			childSpan.AddAttributes(
				trace.StringAttribute("provider.kind", string(emailProvider.Kind)),
			)

			err = s.DoSendEmailNotification(
				childCtx,
				workspaceID,
				messageID,
				contact,
				templateConfig,
				messageData,
				emailProvider,
			)
			if err == nil {
				successfulChannels++
				childSpan.End()
			} else {
				// Log the error but continue with other channels
				s.logger.WithFields(map[string]interface{}{
					"error":        err.Error(),
					"channel":      channel,
					"notification": notification.ID,
					"contact":      contact.Email,
					"message_id":   messageID,
				}).Error("Failed to send email notification")

				tracing.MarkSpanError(childCtx, err)
				childSpan.End()
			}
		}
		// Add other channel handling here as needed
	}

	if successfulChannels == 0 {
		err := fmt.Errorf("failed to send notification through any channel")
		tracing.MarkSpanError(ctx, err)
		return "", err
	}

	span.AddAttributes(
		trace.Int64Attribute("successful_channels", int64(successfulChannels)),
	)

	return messageID, nil
}

// DoSendEmailNotification handles sending through the email channel
func (s *TransactionalNotificationService) DoSendEmailNotification(
	ctx context.Context,
	workspace string,
	messageID string,
	contact *domain.Contact,
	templateConfig domain.ChannelTemplate,
	messageData domain.MessageData,
	emailProvider *domain.EmailProvider,
) error {
	ctx, span := tracing.StartServiceSpan(ctx, "TransactionalNotificationService", "DoSendEmailNotification")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspace),
		trace.StringAttribute("message_id", messageID),
		trace.StringAttribute("contact.email", contact.Email),
		trace.StringAttribute("template_id", templateConfig.TemplateID),
	)

	s.logger.WithFields(map[string]interface{}{
		"workspace":   workspace,
		"message_id":  messageID,
		"contact":     contact.Email,
		"template_id": templateConfig.TemplateID,
	}).Debug("Preparing to send email notification")

	// Get the template
	template, err := s.templateService.GetTemplateByID(ctx, workspace, templateConfig.TemplateID, int64(0))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":       err.Error(),
			"template_id": templateConfig.TemplateID,
		}).Error("Failed to get template")

		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to get template: %w", err)
	}

	span.AddAttributes(
		trace.StringAttribute("template.subject", template.Email.Subject),
		trace.StringAttribute("template.from_email", template.Email.FromAddress),
	)

	// Compile the template with the message data
	compiledTemplate, err := s.templateService.CompileTemplate(ctx, workspace, template.Email.VisualEditorTree, messageData.Data)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":       err.Error(),
			"template_id": templateConfig.TemplateID,
		}).Error("Failed to compile template")

		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to compile template: %w", err)
	}

	tracing.AddAttribute(ctx, "template.compilation_success", compiledTemplate.Success)

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Unknown error"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.WithField("error", errMsg).Error("Template compilation failed")

		err := fmt.Errorf("template compilation failed: %s", errMsg)
		tracing.MarkSpanError(ctx, err)
		return err
	}

	// Get necessary email information from the template
	fromEmail := template.Email.FromAddress
	fromName := template.Email.FromName
	subject := template.Email.Subject
	htmlContent := *compiledTemplate.HTML

	// Create message history record
	messageHistory := &domain.MessageHistory{
		ID:          messageID,
		ContactID:   contact.Email,
		TemplateID:  templateConfig.TemplateID,
		Channel:     "email",
		Status:      domain.MessageStatusSent,
		MessageData: messageData,
		SentAt:      time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Save to message history
	if err := s.messageHistoryRepo.Create(ctx, workspace, messageHistory); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"message_id": messageID,
		}).Error("Failed to create message history")

		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to create message history: %w", err)
	}

	tracing.AddAttribute(ctx, "message_history.created", true)

	// Send the email using the email service
	s.logger.WithFields(map[string]interface{}{
		"to":         contact.Email,
		"from":       fromEmail,
		"subject":    subject,
		"message_id": messageID,
	}).Debug("Sending email")

	tracing.AddAttribute(ctx, "email.sending", true)

	err = s.emailService.SendEmail(
		ctx,
		workspace,
		false, // Use transactional provider type
		fromEmail,
		fromName,
		contact.Email, // To address
		subject,
		htmlContent,
		emailProvider,
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

			tracing.AddAttribute(ctx, "message_history.update_error", updateErr.Error())
		}

		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"message_id": messageID,
			"to":         contact.Email,
		}).Error("Failed to send email")

		tracing.MarkSpanError(ctx, err)
		tracing.AddAttribute(ctx, "email.error", err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"message_id": messageID,
		"to":         contact.Email,
	}).Info("Email sent successfully")

	tracing.AddAttribute(ctx, "email.sent", true)
	return nil
}
