package broadcast

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

//go:generate mockgen -destination=./mocks/mock_broadcast_orchestrator.go -package=mocks github.com/Notifuse/notifuse/internal/service/broadcast BroadcastOrchestratorInterface

// BroadcastOrchestratorInterface defines the interface for broadcast orchestration
type BroadcastOrchestratorInterface interface {
	// CanProcess returns true if this processor can handle the given task type
	CanProcess(taskType string) bool

	// Process executes or continues a broadcast sending task
	Process(ctx context.Context, task *domain.Task) (bool, error)

	// LoadTemplatesForBroadcast loads all templates for a broadcast's variations
	LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error)

	// ValidateTemplates validates that the required templates are loaded and valid
	ValidateTemplates(templates map[string]*domain.Template) error

	// GetTotalRecipientCount gets the total number of recipients for a broadcast
	GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error)

	// FetchBatch retrieves a batch of recipients for a broadcast
	FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error)
}

// BroadcastOrchestrator is the main processor for sending broadcasts
type BroadcastOrchestrator struct {
	messageSender    MessageSender
	broadcastService domain.BroadcastSender
	templateService  domain.TemplateService
	contactRepo      domain.ContactRepository
	taskRepo         domain.TaskRepository
	logger           logger.Logger
	config           *Config
}

// NewBroadcastOrchestrator creates a new broadcast orchestrator
func NewBroadcastOrchestrator(
	messageSender MessageSender,
	broadcastService domain.BroadcastSender,
	templateService domain.TemplateService,
	contactRepo domain.ContactRepository,
	taskRepo domain.TaskRepository,
	logger logger.Logger,
	config *Config,
) BroadcastOrchestratorInterface {
	if config == nil {
		config = DefaultConfig()
	}

	return &BroadcastOrchestrator{
		messageSender:    messageSender,
		broadcastService: broadcastService,
		templateService:  templateService,
		contactRepo:      contactRepo,
		taskRepo:         taskRepo,
		logger:           logger,
		config:           config,
	}
}

// CanProcess returns true if this processor can handle the given task type
func (o *BroadcastOrchestrator) CanProcess(taskType string) bool {
	return taskType == "send_broadcast"
}

// LoadTemplatesForBroadcast loads all templates for a broadcast's variations
func (o *BroadcastOrchestrator) LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error) {
	startTime := time.Now()
	defer func() {
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Template loading completed")
	}()

	// Get the broadcast to access its template variations
	broadcast, err := o.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
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
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Error("No template variations found in broadcast")
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no template variations found in broadcast", false, nil)
	}

	// Load all templates
	templates := make(map[string]*domain.Template)
	for templateID := range templateIDs {
		template, err := o.templateService.GetTemplateByID(ctx, workspaceID, templateID, 1) // Always use version 1
		if err != nil {
			o.logger.WithFields(map[string]interface{}{
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
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Error("No valid templates found for broadcast")
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no valid templates found for broadcast", false, nil)
	}

	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":    broadcastID,
		"workspace_id":    workspaceID,
		"template_count":  len(templates),
		"variation_count": len(broadcast.TestSettings.Variations),
	}).Info("Templates loaded for broadcast")

	return templates, nil
}

// ValidateTemplates validates that the required templates are loaded and valid
func (o *BroadcastOrchestrator) ValidateTemplates(templates map[string]*domain.Template) error {
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
			o.logger.WithField("template_id", id).Error("Template missing email configuration")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing email configuration", false, nil)
		}

		if template.Email.FromAddress == "" {
			o.logger.WithField("template_id", id).Error("Template missing from address")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing from address", false, nil)
		}

		if template.Email.Subject == "" {
			o.logger.WithField("template_id", id).Error("Template missing subject")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing subject", false, nil)
		}

		if template.Email.VisualEditorTree.Kind == "" {
			o.logger.WithField("template_id", id).Error("Template missing content")
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing content", false, nil)
		}
	}

	return nil
}

// GetTotalRecipientCount gets the total number of recipients for a broadcast
func (o *BroadcastOrchestrator) GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
	startTime := time.Now()
	defer func() {
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Recipient count completed")
	}()

	// Get the broadcast to access audience settings
	broadcast, err := o.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient count")
		return 0, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Use the contact repository to count recipients
	count, err := o.contactRepo.CountContactsForBroadcast(ctx, workspaceID, broadcast.Audience)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to count recipients for broadcast")
		return 0, NewBroadcastError(ErrCodeRecipientFetch, "failed to count recipients", true, err)
	}

	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":      broadcastID,
		"workspace_id":      workspaceID,
		"recipient_count":   count,
		"audience_lists":    len(broadcast.Audience.Lists),
		"audience_segments": len(broadcast.Audience.Segments),
	}).Info("Got recipient count for broadcast")

	return count, nil
}

// FetchBatch retrieves a batch of recipients for a broadcast
func (o *BroadcastOrchestrator) FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error) {
	startTime := time.Now()
	defer func() {
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
		}).Debug("Recipient batch fetch completed")
	}()

	// Get the broadcast to access audience settings
	broadcast, err := o.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient fetch")
		return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Apply the actual batch limit from config if not specified
	if limit <= 0 {
		limit = o.config.FetchBatchSize
	}

	// Fetch contacts based on broadcast audience
	contactsWithList, err := o.contactRepo.GetContactsForBroadcast(ctx, workspaceID, broadcast.Audience, limit, offset)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
			"error":        err.Error(),
		}).Error("Failed to fetch recipients for broadcast")
		return nil, NewBroadcastError(ErrCodeRecipientFetch, "failed to fetch recipients", true, err)
	}

	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":     broadcastID,
		"workspace_id":     workspaceID,
		"offset":           offset,
		"limit":            limit,
		"contacts_fetched": len(contactsWithList),
		"with_list_info":   true,
	}).Info("Fetched recipient batch with list info")

	return contactsWithList, nil
}

// formatDuration formats a duration in a human-readable form
func (o *BroadcastOrchestrator) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	} else {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", h, m)
	}
}

// calculateProgress calculates the progress percentage (0-100)
func (o *BroadcastOrchestrator) calculateProgress(processed, total int) float64 {
	if total <= 0 {
		return 100.0 // Avoid division by zero
	}

	progress := float64(processed) / float64(total) * 100.0
	if progress > 100.0 {
		progress = 100.0
	}
	return progress
}

// formatProgressMessage creates a human-readable progress message
func (o *BroadcastOrchestrator) formatProgressMessage(processed, total int, startTime time.Time) string {
	progress := o.calculateProgress(processed, total)

	// Calculate remaining time if we have processed more than 5%
	var eta string
	if progress > 5.0 && processed > 0 {
		elapsed := time.Since(startTime)
		estimatedTotal := elapsed.Seconds() * float64(total) / float64(processed)
		remaining := estimatedTotal - elapsed.Seconds()

		if remaining > 0 {
			eta = fmt.Sprintf(", ETA: %s", o.formatDuration(time.Duration(remaining)*time.Second))
		}
	}

	return fmt.Sprintf("Processed %d/%d recipients (%.1f%%)%s",
		processed, total, progress, eta)
}

// saveProgressState saves the current task progress to the repository
func (o *BroadcastOrchestrator) saveProgressState(
	ctx context.Context,
	workspaceID, taskID, broadcastID string,
	totalRecipients, sentCount, failedCount, processedCount int,
	lastSaveTime time.Time,
	startTime time.Time,
) (time.Time, error) {
	// Skip saving if not enough time has passed
	if time.Since(lastSaveTime) < 5*time.Second {
		return lastSaveTime, nil
	}

	// Calculate progress
	progress := o.calculateProgress(processedCount, totalRecipients)
	message := o.formatProgressMessage(processedCount, totalRecipients, startTime)

	// Create state
	state := &domain.TaskState{
		Progress: progress,
		Message:  message,
		SendBroadcast: &domain.SendBroadcastState{
			BroadcastID:     broadcastID,
			TotalRecipients: totalRecipients,
			SentCount:       sentCount,
			FailedCount:     failedCount,
			ChannelType:     "email", // Hardcoded for now
			RecipientOffset: int64(processedCount),
		},
	}

	// Save state
	err := o.taskRepo.SaveState(ctx, workspaceID, taskID, progress, state)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"task_id":      taskID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to save progress state")
		return lastSaveTime, NewBroadcastError(ErrCodeTaskStateInvalid, "failed to save task state", true, err)
	}

	newLastSaveTime := time.Now()

	// Log progress
	o.logger.WithFields(map[string]interface{}{
		"task_id":      taskID,
		"workspace_id": workspaceID,
		"progress":     progress,
		"sent":         sentCount,
		"failed":       failedCount,
	}).Debug("Saved progress state")

	return newLastSaveTime, nil
}

// Process executes or continues a broadcast sending task
func (o *BroadcastOrchestrator) Process(ctx context.Context, task *domain.Task) (bool, error) {
	o.logger.WithField("task_id", task.ID).Info("Processing send_broadcast task")

	// Initialize structured state if needed
	if task.State == nil {
		// Initialize a new state for the broadcast task
		task.State = &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0, // Track how many recipients we've processed
			},
		}
	}

	// Initialize the SendBroadcast state if it doesn't exist yet
	if task.State.SendBroadcast == nil {
		task.State.SendBroadcast = &domain.SendBroadcastState{
			SentCount:       0,
			FailedCount:     0,
			RecipientOffset: 0,
		}
	}

	// Extract broadcast ID from task state or context
	broadcastState := task.State.SendBroadcast
	if broadcastState.BroadcastID == "" && task.BroadcastID != nil {
		// If state doesn't have broadcast ID but task does, use it
		broadcastState.BroadcastID = *task.BroadcastID
	}

	if broadcastState.BroadcastID == "" {
		// In a real implementation, we'd expect the broadcast ID to be set when creating the task
		return false, NewBroadcastErrorWithTask(
			ErrCodeTaskStateInvalid,
			"broadcast ID is missing in task state",
			task.ID,
			false,
			nil,
		)
	}

	// Track progress
	sentCount := broadcastState.SentCount
	failedCount := broadcastState.FailedCount
	processedCount := int(broadcastState.RecipientOffset)
	startTime := time.Now()
	lastSaveTime := time.Now()
	lastLogTime := time.Now()

	// Phase 1: Get recipient count if not already set
	if broadcastState.TotalRecipients == 0 {
		count, err := o.GetTotalRecipientCount(ctx, task.WorkspaceID, broadcastState.BroadcastID)
		if err != nil {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        err.Error(),
			}).Error("Failed to get recipient count for broadcast")
			return false, err
		}

		broadcastState.TotalRecipients = count
		broadcastState.ChannelType = "email" // Hardcoded for now

		task.State.Message = "Preparing to send broadcast"
		task.Progress = 0

		o.logger.WithFields(map[string]interface{}{
			"task_id":          task.ID,
			"broadcast_id":     broadcastState.BroadcastID,
			"total_recipients": broadcastState.TotalRecipients,
			"channel_type":     broadcastState.ChannelType,
		}).Info("Broadcast sending initialized")

		// If there are no recipients, we can mark as completed immediately
		if broadcastState.TotalRecipients == 0 {
			task.State.Message = "Broadcast completed: No recipients found"
			task.Progress = 100.0
			task.State.Progress = 100.0

			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
			}).Info("Broadcast completed with no recipients")

			return true, nil
		}

		// Early return to save state before processing
		return false, nil
	}

	// Phase 2: Load templates
	templates, err := o.LoadTemplatesForBroadcast(ctx, task.WorkspaceID, broadcastState.BroadcastID)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Failed to load templates for broadcast")
		return false, err
	}

	// Validate templates
	if err := o.ValidateTemplates(templates); err != nil {
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Templates validation failed")
		return false, err
	}

	// Phase 3: Process recipients in batches with a timeout
	processCtx, cancel := context.WithTimeout(ctx, o.config.MaxProcessTime)
	defer cancel()

	// Whether we've processed all recipients
	allDone := false

	// Process until timeout or completion
	for {
		select {
		case <-processCtx.Done():
			// We've hit the time limit, break out of the loop
			o.logger.WithField("task_id", task.ID).Info("Processing time limit reached")
			allDone = false
			break
		default:
			// Continue processing
		}

		// Fetch the next batch of recipients
		currentOffset := int(broadcastState.RecipientOffset)
		recipients, err := o.FetchBatch(
			ctx,
			task.WorkspaceID,
			broadcastState.BroadcastID,
			currentOffset,
			o.config.FetchBatchSize,
		)
		if err != nil {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"error":        err.Error(),
			}).Error("Failed to fetch recipients for broadcast")
			return false, err
		}

		// If no more recipients, we're done
		if len(recipients) == 0 {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
			}).Info("No more recipients to process")
			allDone = true
			break
		}

		// Process this batch of recipients
		sent, failed, err := o.messageSender.SendBatch(
			ctx,
			task.WorkspaceID,
			broadcastState.BroadcastID,
			recipients,
			templates,
			nil, // Base template data, will be populated per recipient
		)

		// Handle errors during sending
		if err != nil {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"error":        err.Error(),
			}).Error("Error sending batch")
			// Continue despite errors as we want to make progress
		}

		// Update progress counters
		sentCount += sent
		failedCount += failed
		processedCount += sent + failed

		// Log progress at regular intervals
		if time.Since(lastLogTime) >= o.config.ProgressLogInterval {
			o.logger.WithFields(map[string]interface{}{
				"task_id":         task.ID,
				"broadcast_id":    broadcastState.BroadcastID,
				"sent_count":      sentCount,
				"failed_count":    failedCount,
				"processed_count": processedCount,
				"total_count":     broadcastState.TotalRecipients,
				"progress":        o.calculateProgress(processedCount, broadcastState.TotalRecipients),
				"elapsed":         time.Since(startTime).String(),
			}).Info("Broadcast progress update")

			lastLogTime = time.Now()
		}

		// Save progress to the task
		var saveErr error
		lastSaveTime, saveErr = o.saveProgressState(
			ctx,
			task.WorkspaceID,
			task.ID,
			broadcastState.BroadcastID,
			broadcastState.TotalRecipients,
			sentCount,
			failedCount,
			processedCount,
			lastSaveTime,
			startTime,
		)
		if saveErr != nil {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        saveErr.Error(),
			}).Error("Failed to save progress state")
			// Continue processing despite save error
		}

		// Update local counters for state
		broadcastState.SentCount = sentCount
		broadcastState.FailedCount = failedCount
		broadcastState.RecipientOffset = int64(processedCount)

		// If we processed fewer recipients than requested, we're done
		if len(recipients) < o.config.FetchBatchSize {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"count":        len(recipients),
			}).Info("Reached end of recipient list")
			allDone = true
			break
		}
	}

	// Update task state with the latest progress data
	progress := o.calculateProgress(processedCount, broadcastState.TotalRecipients)
	message := o.formatProgressMessage(processedCount, broadcastState.TotalRecipients, startTime)

	task.State.Progress = progress
	task.State.Message = message
	task.Progress = progress

	o.logger.WithFields(map[string]interface{}{
		"task_id":          task.ID,
		"broadcast_id":     broadcastState.BroadcastID,
		"sent_total":       broadcastState.SentCount,
		"failed_total":     broadcastState.FailedCount,
		"total_recipients": broadcastState.TotalRecipients,
		"progress":         task.Progress,
		"all_done":         allDone,
	}).Info("Broadcast processing cycle completed")

	return allDone, nil
}
