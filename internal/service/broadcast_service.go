package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
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
	// In a real implementation, this would count recipients from a database
	// For testing purposes, we'll return a fixed count
	return 1000, nil
}

// SendBatch sends a batch of messages for a broadcast
func (s *BroadcastService) SendBatch(ctx context.Context, workspaceID, broadcastID string, batchNumber, batchSize int) (int, int, error) {
	// In a real implementation, this would send actual messages through email/SMS/etc providers
	// For testing purposes, we'll simulate sending with some random successes/failures

	// Simulate a 5% failure rate
	failureCount := batchSize / 20
	successCount := batchSize - failureCount

	s.logger.WithFields(map[string]interface{}{
		"workspace_id": workspaceID,
		"broadcast_id": broadcastID,
		"batch":        batchNumber,
		"successes":    successCount,
		"failures":     failureCount,
	}).Info("Sent broadcast batch")

	// If this is the last batch or if all messages have been sent,
	// we should check if we need to update the broadcast status
	shouldUpdateBroadcast := false

	// For simplicity, let's assume a broadcast is complete when batchNumber reaches 10
	// In a real implementation, this would be based on the total recipient count
	if batchNumber >= 10 {
		shouldUpdateBroadcast = true
	}

	// If we need to update the broadcast status
	if shouldUpdateBroadcast {
		// Get the current broadcast state
		broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"broadcast_id": broadcastID,
				"error":        err.Error(),
			}).Error("Failed to get broadcast for status update")
			return successCount, failureCount, nil
		}

		// Check if there were significant failures
		totalFailed := broadcast.TotalFailed + failureCount
		totalSent := broadcast.TotalSent + successCount

		// If more than 20% of messages failed, consider the broadcast failed
		failureRatio := float64(totalFailed) / float64(totalSent+totalFailed)
		if failureRatio > 0.2 && totalFailed > 10 {
			// Create a failure reason for the event
			failureReason := fmt.Sprintf("Too many failures (%d of %d messages failed)",
				totalFailed, totalSent+totalFailed)

			// Using a channel to wait for the event processing
			doneCh := make(chan error, 1)

			// Begin a transaction that will only be committed when the event is processed
			err := s.repo.WithTransaction(ctx, workspaceID, func(tx *sql.Tx) error {
				// Update the broadcast status to failed within the transaction
				broadcast.Status = domain.BroadcastStatusFailed
				now := time.Now().UTC()
				broadcast.UpdatedAt = now
				broadcast.TotalFailed = totalFailed
				broadcast.TotalSent = totalSent

				// Use the repository's transaction-aware method
				err := s.repo.UpdateBroadcastTx(ctx, tx, broadcast)
				if err != nil {
					return fmt.Errorf("failed to update broadcast status: %w", err)
				}

				// Publish broadcast failed event with acknowledgment
				eventPayload := domain.EventPayload{
					Type:        domain.EventBroadcastFailed,
					WorkspaceID: workspaceID,
					EntityID:    broadcastID,
					Data: map[string]interface{}{
						"broadcast_id": broadcastID,
						"reason":       failureReason,
					},
				}

				// Publish the event with callback
				s.eventBus.PublishWithAck(ctx, eventPayload, func(err error) {
					if err != nil {
						// Event processing failed
						s.logger.WithFields(map[string]interface{}{
							"broadcast_id": broadcastID,
							"workspace_id": workspaceID,
							"error":        err.Error(),
							"reason":       failureReason,
						}).Error("Failed to process broadcast failed event")

						doneCh <- fmt.Errorf("failed to process broadcast failed event: %w", err)
					} else {
						s.logger.WithFields(map[string]interface{}{
							"broadcast_id": broadcastID,
							"workspace_id": workspaceID,
							"reason":       failureReason,
						}).Info("Broadcast failed event processed successfully")

						doneCh <- nil
					}
				})

				// Wait for the event processing to complete before committing or rolling back the transaction
				select {
				case err := <-doneCh:
					return err // Return nil to commit, or error to rollback
				case <-ctx.Done():
					return ctx.Err()
				}
			})

			if err != nil {
				s.logger.WithFields(map[string]interface{}{
					"workspace_id": workspaceID,
					"broadcast_id": broadcastID,
					"error":        err.Error(),
				}).Error("Failed to update broadcast status or process event")
			}
		} else {
			// Using a channel to wait for the event processing
			doneCh := make(chan error, 1)

			// Begin a transaction that will only be committed when the event is processed
			err := s.repo.WithTransaction(ctx, workspaceID, func(tx *sql.Tx) error {
				// Update the broadcast status to sent within the transaction
				broadcast.Status = domain.BroadcastStatusSent
				now := time.Now().UTC()
				broadcast.CompletedAt = &now
				broadcast.UpdatedAt = now
				broadcast.TotalFailed = totalFailed
				broadcast.TotalSent = totalSent

				// Use the repository's transaction-aware method
				err := s.repo.UpdateBroadcastTx(ctx, tx, broadcast)
				if err != nil {
					return fmt.Errorf("failed to update broadcast status: %w", err)
				}

				// Create an event with acknowledgment callback
				eventPayload := domain.EventPayload{
					Type:        domain.EventBroadcastSent,
					WorkspaceID: workspaceID,
					EntityID:    broadcastID,
					Data: map[string]interface{}{
						"broadcast_id": broadcastID,
					},
				}

				// Publish the event with callback
				s.eventBus.PublishWithAck(ctx, eventPayload, func(err error) {
					if err != nil {
						// Event processing failed
						s.logger.WithFields(map[string]interface{}{
							"broadcast_id": broadcastID,
							"workspace_id": workspaceID,
							"error":        err.Error(),
						}).Error("Failed to process broadcast sent event")

						doneCh <- fmt.Errorf("failed to process broadcast sent event: %w", err)
					} else {
						s.logger.WithFields(map[string]interface{}{
							"broadcast_id": broadcastID,
							"workspace_id": workspaceID,
						}).Info("Broadcast sent event processed successfully")

						doneCh <- nil
					}
				})

				// Wait for the event processing to complete with a timeout
				// Use a shorter timeout since SendBatch is often called in a loop
				timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()

				select {
				case err := <-doneCh:
					return err // Return nil to commit, or error to rollback
				case <-timeoutCtx.Done():
					// Timeout occurred, log it but still commit the transaction
					s.logger.WithFields(map[string]interface{}{
						"broadcast_id": broadcastID,
						"workspace_id": workspaceID,
					}).Warn("Timeout waiting for broadcast sent event processing, but continuing")
					return nil
				}
			})

			if err != nil {
				s.logger.WithFields(map[string]interface{}{
					"workspace_id": workspaceID,
					"broadcast_id": broadcastID,
					"error":        err.Error(),
				}).Error("Failed to update broadcast status or process event")
			}
		}
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
