package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// BroadcastService handles all broadcast-related operations
type BroadcastService struct {
	logger      logger.Logger
	repo        domain.BroadcastRepository
	contactRepo domain.ContactRepository
	emailSvc    domain.EmailServiceInterface
	templateSvc domain.TemplateService
	taskService domain.TaskService
	authService domain.AuthService
	eventBus    domain.EventBus
	apiEndpoint string
}

// NewBroadcastService creates a new broadcast service
func NewBroadcastService(
	logger logger.Logger,
	repository domain.BroadcastRepository,
	emailService domain.EmailServiceInterface,
	contactRepository domain.ContactRepository,
	templateService domain.TemplateService,
	taskService domain.TaskService,
	authService domain.AuthService,
	eventBus domain.EventBus,
	apiEndpoint string,
) *BroadcastService {
	return &BroadcastService{
		logger:      logger,
		repo:        repository,
		emailSvc:    emailService,
		contactRepo: contactRepository,
		templateSvc: templateService,
		taskService: taskService,
		authService: authService,
		eventBus:    eventBus,
		apiEndpoint: apiEndpoint,
	}
}

// SetTaskService sets the task service (used to avoid circular dependencies)
func (s *BroadcastService) SetTaskService(taskService domain.TaskService) {
	s.taskService = taskService
}

// CreateBroadcast creates a new broadcast
func (s *BroadcastService) CreateBroadcast(ctx context.Context, request *domain.CreateBroadcastRequest) (*domain.Broadcast, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	broadcast, err := request.Validate()
	if err != nil {
		s.logger.Error("Failed to validate broadcast creation request")
		return nil, err
	}

	// Generate a unique ID for the broadcast if not provided
	if broadcast.ID == "" {
		// Create a random ID
		id := make([]byte, 16)
		_, err := rand.Read(id)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
		broadcast.ID = fmt.Sprintf("%x", id)[:32]
	}

	// Set default values
	broadcast.Status = domain.BroadcastStatusDraft
	now := time.Now().UTC()
	broadcast.CreatedAt = now
	broadcast.UpdatedAt = now

	// Set scheduled time if needed
	if broadcast.Schedule.IsScheduled && (broadcast.Schedule.ScheduledDate != "" && broadcast.Schedule.ScheduledTime != "") {
		// Set status to scheduled if the broadcast is scheduled
		broadcast.Status = domain.BroadcastStatusScheduled
	}

	// Persist the broadcast
	err = s.repo.CreateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to create broadcast in repository")
		return nil, err
	}

	s.logger.Info("Broadcast created successfully")

	return broadcast, nil
}

// GetBroadcast retrieves a broadcast by ID
func (s *BroadcastService) GetBroadcast(ctx context.Context, workspaceID, broadcastID string) (*domain.Broadcast, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", broadcastID).Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Fetch the broadcast from the repository
	return s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
}

// GetRecipientCount retrieves the total recipient count for a broadcast
func (s *BroadcastService) GetRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", broadcastID).Error("Failed to authenticate user for workspace")
		return 0, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get the broadcast to retrieve audience settings
	broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient count")
		return 0, fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Use the more efficient count method
	count, err := s.contactRepo.CountContactsForBroadcast(ctx, workspaceID, broadcast.Audience)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"error":        err.Error(),
		}).Error("Failed to count recipients for broadcast")
		return 0, fmt.Errorf("failed to count recipients: %w", err)
	}

	return count, nil
}

// SendBatch sends a batch of messages for a broadcast
func (s *BroadcastService) ProcessRecipients(ctx context.Context, workspaceID, broadcastID string, startOffset, limit int) (int, int, error) {
	// Get the broadcast details
	broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast details for processing recipients")
		return 0, 0, fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Ensure the broadcast is in sending status
	if broadcast.Status != domain.BroadcastStatusSending {
		err := fmt.Errorf("broadcast is not in sending status, current status: %s", broadcast.Status)
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"status":       broadcast.Status,
		}).Error("Cannot process recipients for broadcast with non-sending status")
		return 0, 0, err
	}

	// Fetch contacts for this batch
	contacts, err := s.contactRepo.GetContactsForBroadcast(
		ctx,
		workspaceID,
		broadcast.Audience,
		limit,
		startOffset,
	)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"offset":       startOffset,
			"limit":        limit,
			"error":        err.Error(),
		}).Error("Failed to fetch contacts for broadcast recipients")
		return 0, 0, fmt.Errorf("failed to fetch contacts: %w", err)
	}

	// If no contacts for this batch, we're done
	if len(contacts) == 0 {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"offset":       startOffset,
			"limit":        limit,
		}).Info("No contacts found for this request")
		return 0, 0, nil
	}

	// Track success and failure counts
	successCount := 0
	failureCount := 0

	// Load templates for all variations
	variationTemplates := make(map[string]*domain.Template)
	for _, variation := range broadcast.TestSettings.Variations {
		template, err := s.templateSvc.GetTemplateByID(ctx, workspaceID, variation.TemplateID, 1)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"template_id":  variation.TemplateID,
				"error":        err.Error(),
			}).Error("Failed to load template for variation")
			return 0, 0, fmt.Errorf("failed to load template: %w", err)
		}
		variationTemplates[variation.ID] = template
	}

	// Process each contact
	for _, contact := range contacts {
		// Determine which variation to use for this contact
		var variationID string
		if broadcast.WinningVariation != "" {
			// If there's a winning variation, use it
			variationID = broadcast.WinningVariation
		} else if broadcast.TestSettings.Enabled {
			// A/B testing is enabled but no winner yet, assign a variation
			// Use a deterministic approach based on contact's email
			hashValue := int(contact.Email[0]) % len(broadcast.TestSettings.Variations)
			variationID = broadcast.TestSettings.Variations[hashValue].ID
		} else if len(broadcast.TestSettings.Variations) > 0 {
			// Not A/B testing, use the first variation
			variationID = broadcast.TestSettings.Variations[0].ID
		} else {
			// No variations available, log error and skip
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"email":        contact.Email,
			}).Error("No variations available for contact")
			failureCount++
			continue
		}

		// Get the template for this variation
		template, ok := variationTemplates[variationID]
		if !ok {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"variation_id": variationID,
				"email":        contact.Email,
			}).Error("Template not found for variation")
			failureCount++
			continue
		}

		// Prepare template data
		contactData, err := contact.ToMapOfAny()
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"email":        contact.Email,
				"error":        err.Error(),
			}).Error("Failed to convert contact to template data")
			failureCount++
			continue
		}

		templateData := domain.MapOfAny{
			"contact": contactData,
		}

		// Add UTM parameters if tracking is enabled
		if broadcast.TrackingEnabled && broadcast.UTMParameters != nil {
			templateData["utm_parameters"] = broadcast.UTMParameters
		}

		// Compile the template
		compiledTemplate, err := s.templateSvc.CompileTemplate(ctx, workspaceID, template.Email.VisualEditorTree, templateData)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"email":        contact.Email,
				"error":        err.Error(),
			}).Error("Failed to compile template")
			failureCount++
			continue
		}

		if !compiledTemplate.Success || compiledTemplate.HTML == nil {
			errMsg := "Template compilation failed"
			if compiledTemplate.Error != nil {
				errMsg = compiledTemplate.Error.Message
			}
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"email":        contact.Email,
				"error":        errMsg,
			}).Error("Failed to generate HTML from template")
			failureCount++
			continue
		}

		// Apply rate limiting if configured
		if broadcast.Audience.RateLimitPerMinute > 0 {
			// Simple implementation: sleep for (60 / rate_limit) seconds
			sleepTime := time.Duration(60/broadcast.Audience.RateLimitPerMinute) * time.Second
			time.Sleep(sleepTime)
		}

		// Send the email
		err = s.emailSvc.SendEmail(
			ctx,
			workspaceID,
			"marketing", // Email provider type
			template.Email.FromAddress,
			template.Email.FromName,
			contact.Email,
			template.Email.Subject,
			*compiledTemplate.HTML,
		)

		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"email":        contact.Email,
				"error":        err.Error(),
			}).Error("Failed to send email")
			failureCount++
		} else {
			successCount++

			// Update metrics for the variation
			for i, v := range broadcast.TestSettings.Variations {
				if v.ID == variationID {
					if v.Metrics == nil {
						broadcast.TestSettings.Variations[i].Metrics = &domain.VariationMetrics{}
					}
					broadcast.TestSettings.Variations[i].Metrics.Recipients++
					broadcast.TestSettings.Variations[i].Metrics.Delivered++
					break
				}
			}
		}
	}

	// Update the broadcast with the new totals
	broadcast.TotalSent += successCount
	broadcast.TotalFailed += failureCount
	broadcast.UpdatedAt = time.Now().UTC()

	// Persist the updated metrics
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"error":        err.Error(),
		}).Error("Failed to update broadcast metrics")
		// Continue anyway - we've already sent the emails
	}

	s.logger.WithFields(map[string]interface{}{
		"workspace_id": workspaceID,
		"broadcast_id": broadcastID,
		"offset":       startOffset,
		"limit":        limit,
		"successes":    successCount,
		"failures":     failureCount,
	}).Info("Sent broadcast batch")

	// If this is the last batch or if all messages have been sent,
	// we should check if we need to update the broadcast status
	if len(contacts) < limit {
		// This was the last batch, mark broadcast as sent
		broadcast.Status = domain.BroadcastStatusSent
		now := time.Now().UTC()
		broadcast.CompletedAt = &now
		broadcast.UpdatedAt = now
		broadcast.SentAt = &now

		err = s.repo.UpdateBroadcast(ctx, broadcast)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"error":        err.Error(),
			}).Error("Failed to update broadcast status to sent")
			// Continue anyway - we've already sent the emails
		}

		// Publish broadcast sent event
		eventPayload := domain.EventPayload{
			Type:        domain.EventBroadcastSent,
			WorkspaceID: workspaceID,
			EntityID:    broadcastID,
			Data: map[string]interface{}{
				"broadcast_id": broadcastID,
				"total_sent":   broadcast.TotalSent,
				"total_failed": broadcast.TotalFailed,
			},
		}

		s.eventBus.Publish(ctx, eventPayload)
	}

	return successCount, failureCount, nil
}

// UpdateBroadcast updates an existing broadcast
func (s *BroadcastService) UpdateBroadcast(ctx context.Context, request *domain.UpdateBroadcastRequest) (*domain.Broadcast, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// First, get the existing broadcast
	existingBroadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for update")
		return nil, err
	}

	// Validate and update broadcast fields
	updatedBroadcast, err := request.Validate(existingBroadcast)
	if err != nil {
		s.logger.Error("Failed to validate broadcast update request")
		return nil, err
	}

	// Set the updated time
	updatedBroadcast.UpdatedAt = time.Now().UTC()

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, updatedBroadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return nil, err
	}

	s.logger.Info("Broadcast updated successfully")

	return updatedBroadcast, nil
}

// ListBroadcasts retrieves a list of broadcasts with pagination
func (s *BroadcastService) ListBroadcasts(ctx context.Context, params domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, params.WorkspaceID)
	if err != nil {
		s.logger.Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Apply default values for pagination if not provided
	if params.Limit <= 0 {
		params.Limit = 50 // Default limit
	}
	if params.Limit > 100 {
		params.Limit = 100 // Maximum limit
	}
	if params.Offset < 0 {
		params.Offset = 0 // Ensure offset is not negative
	}

	response, err := s.repo.ListBroadcasts(ctx, params)
	if err != nil {
		s.logger.Error("Failed to list broadcasts from repository")
		return nil, err
	}

	// If WithTemplates is true, fetch template details for each variation
	if params.WithTemplates {
		for _, broadcast := range response.Broadcasts {
			for i, variation := range broadcast.TestSettings.Variations {
				if variation.TemplateID != "" {
					// Fetch the template for this variation
					template, err := s.templateSvc.GetTemplateByID(ctx, params.WorkspaceID, variation.TemplateID, 1)
					if err != nil {
						s.logger.Error("Failed to fetch template for broadcast variation")
						// Continue with the next variation rather than failing the whole request
						continue
					}

					// Assign the template to the variation
					broadcast.SetTemplateForVariation(i, template)
				}
			}
		}
	}

	s.logger.Info("Broadcasts listed successfully")

	return response, nil
}

// ScheduleBroadcast schedules a broadcast for sending
func (s *BroadcastService) ScheduleBroadcast(ctx context.Context, request *domain.ScheduleBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate schedule broadcast request")
		return err
	}

	// Using a channel to wait for the event callback
	done := make(chan error, 1)

	// Use transaction to retrieve, update the broadcast, and publish the event
	err = s.repo.WithTransaction(ctx, request.WorkspaceID, func(tx *sql.Tx) error {
		// Retrieve the broadcast
		broadcast, err := s.repo.GetBroadcastTx(ctx, tx, request.WorkspaceID, request.ID)
		if err != nil {
			s.logger.Error("Failed to get broadcast for scheduling")
			return err
		}

		// Only draft broadcasts can be scheduled
		if broadcast.Status != domain.BroadcastStatusDraft {
			err := fmt.Errorf("only broadcasts with draft status can be scheduled, current status: %s", broadcast.Status)
			s.logger.Error("Cannot schedule broadcast with non-draft status")
			return err
		}

		// Update broadcast status and scheduling info
		broadcast.Status = domain.BroadcastStatusScheduled
		broadcast.UpdatedAt = time.Now().UTC()

		if request.SendNow {
			// If sending immediately, set status to sending
			broadcast.Status = domain.BroadcastStatusSending
			now := time.Now().UTC()
			broadcast.StartedAt = &now
		} else {
			// Update the schedule settings with the requested settings
			broadcast.Schedule.IsScheduled = true
			broadcast.Schedule.ScheduledDate = request.ScheduledDate
			broadcast.Schedule.ScheduledTime = request.ScheduledTime
			broadcast.Schedule.Timezone = request.Timezone
			broadcast.Schedule.UseRecipientTimezone = request.UseRecipientTimezone
		}

		// Persist the changes
		err = s.repo.UpdateBroadcastTx(ctx, tx, broadcast)
		if err != nil {
			s.logger.Error("Failed to update broadcast in repository")
			return err
		}

		// Create event payload
		eventPayload := domain.EventPayload{
			Type:        domain.EventBroadcastScheduled,
			WorkspaceID: request.WorkspaceID,
			EntityID:    request.ID,
			Data: map[string]interface{}{
				"broadcast_id": request.ID,
				"send_now":     request.SendNow,
				"status":       string(broadcast.Status),
			},
		}

		// Publish the event with callback within the transaction
		s.eventBus.PublishWithAck(ctx, eventPayload, func(eventErr error) {
			if eventErr != nil {
				// Event processing failed, log the error
				s.logger.WithFields(map[string]interface{}{
					"broadcast_id": request.ID,
					"workspace_id": request.WorkspaceID,
					"error":        eventErr.Error(),
				}).Error("Failed to process schedule broadcast event")

				// Since we're still in the same transaction, we don't need to rollback explicitly
				// The outer transaction will be rolled back when we return an error

				done <- fmt.Errorf("failed to process schedule event: %w", eventErr)
			} else {
				s.logger.WithField("broadcast_id", request.ID).Info("Schedule broadcast event processed successfully")
				done <- nil
			}
		})

		// Wait for the event processing to complete
		select {
		case eventErr := <-done:
			if eventErr != nil {
				// If the event processing failed, roll back the transaction by returning an error
				return eventErr
			}
			// If event processing succeeded, commit the transaction
			return nil
		case <-ctx.Done():
			// If context is cancelled, roll back transaction by returning an error
			return ctx.Err()
		}
	})

	return err
}

// PauseBroadcast pauses a sending broadcast
func (s *BroadcastService) PauseBroadcast(ctx context.Context, request *domain.PauseBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate pause broadcast request")
		return err
	}

	// Using a channel to wait for the event callback
	done := make(chan error, 1)

	// Use transaction to retrieve, update the broadcast, and publish the event
	err = s.repo.WithTransaction(ctx, request.WorkspaceID, func(tx *sql.Tx) error {
		// Retrieve the broadcast
		broadcast, err := s.repo.GetBroadcastTx(ctx, tx, request.WorkspaceID, request.ID)
		if err != nil {
			s.logger.Error("Failed to get broadcast for pausing")
			return err
		}

		// Only sending broadcasts can be paused
		if broadcast.Status != domain.BroadcastStatusSending {
			err := fmt.Errorf("only broadcasts with sending status can be paused, current status: %s", broadcast.Status)
			s.logger.Error("Cannot pause broadcast with non-sending status")
			return err
		}

		// Update broadcast status and pause info
		broadcast.Status = domain.BroadcastStatusPaused
		now := time.Now().UTC()
		broadcast.PausedAt = &now
		broadcast.UpdatedAt = now

		// Persist the changes
		err = s.repo.UpdateBroadcastTx(ctx, tx, broadcast)
		if err != nil {
			s.logger.Error("Failed to update broadcast in repository")
			return err
		}

		s.logger.Info("Broadcast paused successfully")

		// Create an event with acknowledgment callback
		eventPayload := domain.EventPayload{
			Type:        domain.EventBroadcastPaused,
			WorkspaceID: request.WorkspaceID,
			EntityID:    request.ID,
			Data: map[string]interface{}{
				"broadcast_id": request.ID,
			},
		}

		// Publish the event with callback within the transaction
		s.eventBus.PublishWithAck(ctx, eventPayload, func(eventErr error) {
			if eventErr != nil {
				// Event processing failed, log the error
				s.logger.WithFields(map[string]interface{}{
					"broadcast_id": request.ID,
					"workspace_id": request.WorkspaceID,
					"error":        eventErr.Error(),
				}).Error("Failed to process pause broadcast event")

				// Since we're still in the same transaction, we don't need to rollback explicitly
				// The outer transaction will be rolled back when we return an error

				done <- fmt.Errorf("failed to process pause event: %w", eventErr)
			} else {
				s.logger.WithField("broadcast_id", request.ID).Info("Pause broadcast event processed successfully")
				done <- nil
			}
		})

		// Wait for the event processing to complete
		select {
		case eventErr := <-done:
			if eventErr != nil {
				// If the event processing failed, roll back the transaction by returning an error
				return eventErr
			}
			// If event processing succeeded, commit the transaction
			return nil
		case <-ctx.Done():
			// If context is cancelled, roll back transaction by returning an error
			return ctx.Err()
		}
	})

	return err
}

// ResumeBroadcast resumes a paused broadcast
func (s *BroadcastService) ResumeBroadcast(ctx context.Context, request *domain.ResumeBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate resume broadcast request")
		return err
	}

	// Using a channel to wait for the event callback
	done := make(chan error, 1)

	// Use transaction to retrieve, update the broadcast, and publish the event
	err = s.repo.WithTransaction(ctx, request.WorkspaceID, func(tx *sql.Tx) error {
		// Retrieve the broadcast
		broadcast, err := s.repo.GetBroadcastTx(ctx, tx, request.WorkspaceID, request.ID)
		if err != nil {
			s.logger.Error("Failed to get broadcast for resuming")
			return err
		}

		// Only paused broadcasts can be resumed
		if broadcast.Status != domain.BroadcastStatusPaused {
			err := fmt.Errorf("only broadcasts with paused status can be resumed, current status: %s", broadcast.Status)
			s.logger.Error("Cannot resume broadcast with invalid status")
			return err
		}

		// Update broadcast status
		now := time.Now().UTC()
		broadcast.UpdatedAt = now

		// Determine the new status based on scheduling
		startNow := false

		// If broadcast was originally scheduled and scheduled time is in the future
		if broadcast.Schedule.IsScheduled {
			scheduledTime, err := broadcast.Schedule.ParseScheduledDateTime()
			isScheduledInFuture := err == nil && scheduledTime.After(now) && broadcast.StartedAt == nil

			if isScheduledInFuture {
				broadcast.Status = domain.BroadcastStatusScheduled
				s.logger.Info("Broadcast resumed to scheduled status")
			} else {
				// If scheduled time has passed or there was an error parsing it
				broadcast.Status = domain.BroadcastStatusSending
				startNow = true
				if broadcast.StartedAt == nil {
					broadcast.StartedAt = &now
				}
				s.logger.Info("Broadcast resumed to sending status")
			}
		} else {
			// If broadcast wasn't scheduled, resume sending
			broadcast.Status = domain.BroadcastStatusSending
			startNow = true
			if broadcast.StartedAt == nil {
				broadcast.StartedAt = &now
			}
			s.logger.Info("Broadcast resumed to sending status")
		}

		// Clear the paused timestamp
		broadcast.PausedAt = nil

		// Persist the changes
		err = s.repo.UpdateBroadcastTx(ctx, tx, broadcast)
		if err != nil {
			s.logger.Error("Failed to update broadcast in repository")
			return err
		}

		s.logger.Info("Broadcast resumed successfully")

		// Create an event with acknowledgment callback
		eventPayload := domain.EventPayload{
			Type:        domain.EventBroadcastResumed,
			WorkspaceID: request.WorkspaceID,
			EntityID:    request.ID,
			Data: map[string]interface{}{
				"broadcast_id": request.ID,
				"start_now":    startNow,
			},
		}

		// Publish the event with callback within the transaction
		s.eventBus.PublishWithAck(ctx, eventPayload, func(eventErr error) {
			if eventErr != nil {
				// Event processing failed, log the error
				s.logger.WithFields(map[string]interface{}{
					"broadcast_id": request.ID,
					"workspace_id": request.WorkspaceID,
					"error":        eventErr.Error(),
				}).Error("Failed to process resume broadcast event")

				// Since we're still in the same transaction, we don't need to rollback explicitly
				// The outer transaction will be rolled back when we return an error

				done <- fmt.Errorf("failed to process resume event: %w", eventErr)
			} else {
				s.logger.WithField("broadcast_id", request.ID).Info("Resume broadcast event processed successfully")
				done <- nil
			}
		})

		// Wait for the event processing to complete
		select {
		case eventErr := <-done:
			if eventErr != nil {
				// If the event processing failed, roll back the transaction by returning an error
				return eventErr
			}
			// If event processing succeeded, commit the transaction
			return nil
		case <-ctx.Done():
			// If context is cancelled, roll back transaction by returning an error
			return ctx.Err()
		}
	})

	return err
}

// CancelBroadcast cancels a scheduled broadcast
func (s *BroadcastService) CancelBroadcast(ctx context.Context, request *domain.CancelBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate cancel broadcast request")
		return err
	}

	// Using a channel to wait for the event callback
	done := make(chan error, 1)

	// Use transaction to retrieve, update the broadcast, and publish the event
	err = s.repo.WithTransaction(ctx, request.WorkspaceID, func(tx *sql.Tx) error {
		// Retrieve the broadcast
		broadcast, err := s.repo.GetBroadcastTx(ctx, tx, request.WorkspaceID, request.ID)
		if err != nil {
			s.logger.Error("Failed to get broadcast for cancellation")
			return err
		}

		// Only scheduled or paused broadcasts can be cancelled
		if broadcast.Status != domain.BroadcastStatusScheduled && broadcast.Status != domain.BroadcastStatusPaused {
			err := fmt.Errorf("only broadcasts with scheduled or paused status can be cancelled, current status: %s", broadcast.Status)
			s.logger.Error("Cannot cancel broadcast with invalid status")
			return err
		}

		// Update broadcast status and cancellation info
		broadcast.Status = domain.BroadcastStatusCancelled
		now := time.Now().UTC()
		broadcast.CancelledAt = &now
		broadcast.UpdatedAt = now

		// Persist the changes
		err = s.repo.UpdateBroadcastTx(ctx, tx, broadcast)
		if err != nil {
			s.logger.Error("Failed to update broadcast in repository")
			return err
		}

		s.logger.Info("Broadcast cancelled successfully")

		// Create an event with acknowledgment callback
		eventPayload := domain.EventPayload{
			Type:        domain.EventBroadcastCancelled,
			WorkspaceID: request.WorkspaceID,
			EntityID:    request.ID,
			Data: map[string]interface{}{
				"broadcast_id": request.ID,
			},
		}

		// Publish the event with callback within the transaction
		s.eventBus.PublishWithAck(ctx, eventPayload, func(eventErr error) {
			if eventErr != nil {
				// Event processing failed, log the error
				s.logger.WithFields(map[string]interface{}{
					"broadcast_id": request.ID,
					"workspace_id": request.WorkspaceID,
					"error":        eventErr.Error(),
				}).Error("Failed to process cancel broadcast event")

				// Since we're still in the same transaction, we don't need to rollback explicitly
				// The outer transaction will be rolled back when we return an error

				done <- fmt.Errorf("failed to process cancel event: %w", eventErr)
			} else {
				s.logger.WithField("broadcast_id", request.ID).Info("Cancel broadcast event processed successfully")
				done <- nil
			}
		})

		// Wait for the event processing to complete
		select {
		case eventErr := <-done:
			if eventErr != nil {
				// If the event processing failed, roll back the transaction by returning an error
				return eventErr
			}
			// If event processing succeeded, commit the transaction
			return nil
		case <-ctx.Done():
			// If context is cancelled, roll back transaction by returning an error
			return ctx.Err()
		}
	})

	return err
}

// DeleteBroadcast deletes a broadcast
func (s *BroadcastService) DeleteBroadcast(ctx context.Context, request *domain.DeleteBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate delete broadcast request")
		return err
	}

	// Retrieve the broadcast to check its status
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for deletion")
		return err
	}

	// Broadcasts in 'sending' status cannot be deleted
	if broadcast.Status == domain.BroadcastStatusSending {
		err := fmt.Errorf("broadcasts in 'sending' status cannot be deleted")
		s.logger.Error("Cannot delete broadcast with sending status")
		return err
	}

	// Delete the broadcast
	err = s.repo.DeleteBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to delete broadcast from repository")
		return err
	}

	s.logger.Info("Broadcast deleted successfully")

	return nil
}

// SendToIndividual sends a broadcast to an individual recipient
func (s *BroadcastService) SendToIndividual(ctx context.Context, request *domain.SendToIndividualRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.BroadcastID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate send to individual request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.BroadcastID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for individual sending")
		return err
	}

	// Determine which variation to use
	variationID := request.VariationID
	if variationID == "" && len(broadcast.TestSettings.Variations) > 0 {
		// If no variation ID specified, use the first one
		variationID = broadcast.TestSettings.Variations[0].ID
		s.logger.Debug("No variation specified, using first variation")
	} else if variationID == "" {
		err := fmt.Errorf("broadcast has no variations")
		s.logger.Error("Cannot send broadcast with no variations")
		return err
	}

	// Find the specified variation
	var variation *domain.BroadcastVariation
	for _, v := range broadcast.TestSettings.Variations {
		if v.ID == variationID {
			variation = &v
			break
		}
	}

	if variation == nil {
		err := fmt.Errorf("variation with ID %s not found in broadcast", variationID)
		s.logger.Error("Variation not found in broadcast")
		return err
	}

	// Fetch the contact if it exists, but don't fail if not found
	contact, contactErr := s.contactRepo.GetContactByEmail(ctx, request.WorkspaceID, request.RecipientEmail)
	if contactErr != nil {
		// Just log the error, don't return it
		s.logger.Info("Contact not found, using email address only")
	}

	// Fetch the template
	template, err := s.templateSvc.GetTemplateByID(ctx, request.WorkspaceID, variation.TemplateID, 1)
	if err != nil {
		s.logger.Error("Failed to fetch template for broadcast")
		return err
	}

	// Prepare template data
	templateData := domain.MapOfAny{
		"contact": domain.MapOfAny{
			"email": request.RecipientEmail,
		},
	}

	// Add contact data if available
	if contact != nil {
		contactData, err := contact.ToMapOfAny()
		if err == nil {
			templateData["contact"] = contactData
		}
	}

	// Compile the template
	compiledTemplate, err := s.templateSvc.CompileTemplate(ctx, request.WorkspaceID, template.Email.VisualEditorTree, templateData)
	if err != nil {
		s.logger.Error("Failed to compile template for broadcast")
		return err
	}

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Template compilation failed"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.Error("Failed to generate HTML from template")
		return fmt.Errorf("template compilation failed: %s", errMsg)
	}

	// Send the email
	err = s.emailSvc.SendEmail(
		ctx,
		request.WorkspaceID,
		"marketing", // Email provider type
		template.Email.FromAddress,
		template.Email.FromName,
		request.RecipientEmail,
		template.Email.Subject,
		*compiledTemplate.HTML,
	)
	if err != nil {
		s.logger.Error("Failed to send email to individual recipient")
		return err
	}

	s.logger.Info("Email sent to individual recipient successfully")

	return nil
}

// SendWinningVariation sends the winning variation of an A/B test to remaining recipients
func (s *BroadcastService) SendWinningVariation(ctx context.Context, request *domain.SendWinningVariationRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.BroadcastID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate send winning variation request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.BroadcastID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for sending winning variation")
		return err
	}

	// Verify that the broadcast has A/B testing enabled
	if !broadcast.TestSettings.Enabled {
		err := fmt.Errorf("broadcast does not have A/B testing enabled")
		s.logger.Error("Cannot send winning variation for broadcast without A/B testing")
		return err
	}

	// Find the specified variation
	var winningVariation *domain.BroadcastVariation
	for _, v := range broadcast.TestSettings.Variations {
		if v.ID == request.VariationID {
			winningVariation = &v
			break
		}
	}

	if winningVariation == nil {
		err := fmt.Errorf("variation with ID %s not found in broadcast", request.VariationID)
		s.logger.Error("Winning variation not found in broadcast")
		return err
	}

	// Ensure tracking is enabled when sending the winning variation
	if !request.TrackingEnabled && broadcast.TrackingEnabled {
		// If tracking was enabled for the broadcast but not specified in the request,
		// use the broadcast's tracking setting
		request.TrackingEnabled = broadcast.TrackingEnabled
	}

	// Update the broadcast with the winning variation and winner sent time
	now := time.Now().UTC()
	broadcast.WinningVariation = request.VariationID
	broadcast.WinnerSentAt = &now
	broadcast.TrackingEnabled = request.TrackingEnabled
	broadcast.UpdatedAt = now

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast with winning variation information")
		return err
	}

	s.logger.Info("Broadcast updated with winning variation")

	return nil
}

// GetBroadcastRecipients gets recipients for a broadcast with pagination
func (s *BroadcastService) GetBroadcastRecipients(ctx context.Context, workspaceID, broadcastID string, limit, offset int) ([]*domain.Contact, error) {
	// Get the broadcast to retrieve audience settings
	broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipients")
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Fetch contacts using the repository
	contacts, err := s.contactRepo.GetContactsForBroadcast(
		ctx,
		workspaceID,
		broadcast.Audience,
		limit,
		offset,
	)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"limit":        limit,
			"offset":       offset,
			"error":        err.Error(),
		}).Error("Failed to get contacts for broadcast")
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	return contacts, nil
}

// SendToContact sends a broadcast message to a single contact
func (s *BroadcastService) SendToContact(ctx context.Context, workspaceID, broadcastID string, contact *domain.Contact) error {
	// Get the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for sending to contact")
		return &domain.ErrNotFound{
			Entity: "broadcast",
			ID:     broadcastID,
		}
	}

	// Determine which variation to use for this contact
	var variationID string
	if broadcast.WinningVariation != "" {
		// If there's a winning variation, use it
		variationID = broadcast.WinningVariation
	} else if broadcast.TestSettings.Enabled {
		// A/B testing is enabled but no winner yet, assign a variation
		// Use a deterministic approach based on contact's email
		hashValue := int(contact.Email[0]) % len(broadcast.TestSettings.Variations)
		variationID = broadcast.TestSettings.Variations[hashValue].ID
	} else if len(broadcast.TestSettings.Variations) > 0 {
		// Not A/B testing, use the first variation
		variationID = broadcast.TestSettings.Variations[0].ID
	} else {
		// No variations available
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      "no template variations available for broadcast",
		}
	}

	// Find the specified variation
	var variation *domain.BroadcastVariation
	for _, v := range broadcast.TestSettings.Variations {
		if v.ID == variationID {
			variation = &v
			break
		}
	}

	if variation == nil {
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      "variation not found",
		}
	}

	// Fetch the template
	template, err := s.templateSvc.GetTemplateByID(ctx, workspaceID, variation.TemplateID, 1)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"template_id":  variation.TemplateID,
			"error":        err.Error(),
		}).Error("Failed to get template for broadcast")
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      "template not found",
			Err:         err,
		}
	}

	// Prepare template data
	// Convert contact to map first
	contactData, err := contact.ToMapOfAny()
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to convert contact to template data")
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      "failed to prepare template data",
			Err:         err,
		}
	}

	templateData := domain.MapOfAny{
		"contact": contactData,
		"unsubscribe_url": fmt.Sprintf("%s/api/contacts/unsubscribe?workspace_id=%s&email=%s",
			s.apiEndpoint, workspaceID, url.QueryEscape(contact.Email)),
	}

	// Add UTM parameters if tracking is enabled
	if broadcast.TrackingEnabled && broadcast.UTMParameters != nil {
		templateData["utm"] = broadcast.UTMParameters
	}

	// Compile the template
	compiledTemplate, err := s.templateSvc.CompileTemplate(ctx, workspaceID, template.Email.VisualEditorTree, templateData)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to compile template")
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      "template compilation failed",
			Err:         err,
		}
	}

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Template compilation failed"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        errMsg,
		}).Error("Failed to generate HTML from template")
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      errMsg,
		}
	}

	// Send the email
	err = s.emailSvc.SendEmail(
		ctx,
		workspaceID,
		"marketing", // Email provider type
		template.Email.FromAddress,
		template.Email.FromName,
		contact.Email,
		template.Email.Subject,
		*compiledTemplate.HTML,
	)

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to send email to contact")
		return &domain.ErrBroadcastDelivery{
			BroadcastID: broadcastID,
			Email:       contact.Email,
			Reason:      "failed to send email",
			Err:         err,
		}
	}

	// Record the successful message
	// Create a unique ID for the message
	messageID := fmt.Sprintf("%s-%s-%s", workspaceID, broadcastID, contact.Email)

	broadcastIDPtr := broadcastID
	message := &domain.MessageHistory{
		ID:              messageID,
		ContactID:       contact.Email, // Using email as contact ID
		BroadcastID:     &broadcastIDPtr,
		TemplateID:      template.ID,
		TemplateVersion: 1,
		Channel:         "email",
		Status:          domain.MessageStatusSent,
		MessageData: domain.MessageData{
			Data: map[string]interface{}{
				"broadcast_id": broadcastID,
				"variation_id": variationID,
				"email":        contact.Email,
			},
		},
		SentAt:    time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Record message history - don't fail if this doesn't work
	if err := s.RecordMessageSent(ctx, workspaceID, message); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Warn("Failed to record message history, but email was sent")
		// We don't return the error here since the message was already sent
	}

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"email":        contact.Email,
		"variation_id": variationID,
	}).Debug("Email sent to contact successfully")

	return nil
}

// SendToContactWithTemplates sends a broadcast message to a single contact with pre-loaded templates
func (s *BroadcastService) SendToContactWithTemplates(ctx context.Context, workspaceID, broadcastID string,
	contact *domain.Contact, templates map[string]*domain.Template, templateData map[string]interface{}) error {

	// Get the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for sending to contact")
		return fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Determine which variation to use for this contact
	var variationID string
	if broadcast.WinningVariation != "" {
		// If there's a winning variation, use it
		variationID = broadcast.WinningVariation
	} else if broadcast.TestSettings.Enabled {
		// A/B testing is enabled but no winner yet, assign a variation
		// Use a deterministic approach based on contact's email
		hashValue := int(contact.Email[0]) % len(broadcast.TestSettings.Variations)
		variationID = broadcast.TestSettings.Variations[hashValue].ID
	} else if len(broadcast.TestSettings.Variations) > 0 {
		// Not A/B testing, use the first variation
		variationID = broadcast.TestSettings.Variations[0].ID
	} else {
		// No variations available
		return fmt.Errorf("no template variations available for broadcast")
	}

	// Find the specified variation
	var variation *domain.BroadcastVariation
	for _, v := range broadcast.TestSettings.Variations {
		if v.ID == variationID {
			variation = &v
			break
		}
	}

	if variation == nil {
		return fmt.Errorf("variation with ID %s not found in broadcast", variationID)
	}

	// Get the template from the pre-loaded templates map
	template, exists := templates[variation.TemplateID]
	if !exists {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"template_id":  variation.TemplateID,
			"email":        contact.Email,
		}).Error("Template not found in pre-loaded templates")

		// Fall back to loading the template directly
		var err error
		template, err = s.templateSvc.GetTemplateByID(ctx, workspaceID, variation.TemplateID, 1)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"template_id":  variation.TemplateID,
				"email":        contact.Email,
				"error":        err.Error(),
			}).Error("Failed to fetch template for broadcast")
			return err
		}
	}

	// Use provided template data if available, otherwise generate it
	var finalTemplateData domain.MapOfAny
	if templateData != nil {
		// Use provided template data
		finalTemplateData = templateData
	} else {
		// Generate template data from contact
		contactData, err := contact.ToMapOfAny()
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"email":        contact.Email,
				"error":        err.Error(),
			}).Error("Failed to convert contact to template data")
			return err
		}

		finalTemplateData = domain.MapOfAny{
			"contact": contactData,
		}
	}

	// Add UTM parameters if tracking is enabled and not already in template data
	if broadcast.TrackingEnabled && broadcast.UTMParameters != nil {
		if _, exists := finalTemplateData["utm_parameters"]; !exists {
			finalTemplateData["utm_parameters"] = broadcast.UTMParameters
		}
	}

	// Compile the template
	compiledTemplate, err := s.templateSvc.CompileTemplate(ctx, workspaceID, template.Email.VisualEditorTree, finalTemplateData)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to compile template")
		return err
	}

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Template compilation failed"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        errMsg,
		}).Error("Failed to generate HTML from template")
		return fmt.Errorf("template compilation failed: %s", errMsg)
	}

	// Send the email
	err = s.emailSvc.SendEmail(
		ctx,
		workspaceID,
		"marketing", // Email provider type
		template.Email.FromAddress,
		template.Email.FromName,
		contact.Email,
		template.Email.Subject,
		*compiledTemplate.HTML,
	)

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"email":        contact.Email,
			"error":        err.Error(),
		}).Error("Failed to send email to contact")
		return err
	}

	// Update metrics for the variation (if needed, this can be done in bulk later)
	// This operation is not performed here for performance reasons,
	// metrics are tracked in the task processor instead

	s.logger.WithFields(map[string]interface{}{
		"broadcast_id": broadcastID,
		"email":        contact.Email,
		"variation_id": variationID,
	}).Debug("Email sent to contact successfully")

	return nil
}

// GetTemplateByID gets a template by ID for use with broadcasts
func (s *BroadcastService) GetTemplateByID(ctx context.Context, workspaceID, templateID string) (*domain.Template, error) {
	// Simply delegate to the template service, but only requesting version 1
	return s.templateSvc.GetTemplateByID(ctx, workspaceID, templateID, 1)
}

// RecordMessageSent records a message sent event in the message history
func (s *BroadcastService) RecordMessageSent(ctx context.Context, workspaceID string, message *domain.MessageHistory) error {
	// Check if message history repository is available
	messageHistoryRepo, ok := s.repo.(interface {
		CreateMessageHistory(ctx context.Context, workspaceID string, message *domain.MessageHistory) error
	})

	if !ok {
		s.logger.Error("Repository does not support message history")
		// Don't fail the broadcast if message history tracking is not available
		return nil
	}

	return messageHistoryRepo.CreateMessageHistory(ctx, workspaceID, message)
}

// UpdateMessageStatus updates the status of a message in the message history
func (s *BroadcastService) UpdateMessageStatus(ctx context.Context, workspaceID string, messageID string, status domain.MessageStatus, timestamp time.Time) error {
	// Check if message history repository is available
	messageHistoryRepo, ok := s.repo.(interface {
		UpdateMessageStatus(ctx context.Context, workspaceID string, messageID string, status domain.MessageStatus, timestamp time.Time) error
	})

	if !ok {
		s.logger.Error("Repository does not support message history")
		// Don't fail the broadcast if message history tracking is not available
		return nil
	}

	return messageHistoryRepo.UpdateMessageStatus(ctx, workspaceID, messageID, status, timestamp)
}

// GetAPIEndpoint returns the API endpoint for the broadcast service
func (s *BroadcastService) GetAPIEndpoint() string {
	return s.apiEndpoint
}
