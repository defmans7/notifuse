package broadcast

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

//go:generate mockgen -destination=../../mocks/broadcast_orchestrator.go -package=mocks github.com/Notifuse/notifuse/internal/service/broadcast BroadcastOrchestrator

// BroadcastOrchestrator is the main processor for sending broadcasts
type BroadcastOrchestrator struct {
	templateLoader   TemplateLoader
	recipientFetcher RecipientFetcher
	messageSender    MessageSender
	progressTracker  ProgressTracker
	logger           logger.Logger
	config           *Config
}

// NewBroadcastOrchestrator creates a new broadcast orchestrator
func NewBroadcastOrchestrator(
	templateLoader TemplateLoader,
	recipientFetcher RecipientFetcher,
	messageSender MessageSender,
	progressTracker ProgressTracker,
	logger logger.Logger,
	config *Config,
) *BroadcastOrchestrator {
	if config == nil {
		config = DefaultConfig()
	}

	return &BroadcastOrchestrator{
		templateLoader:   templateLoader,
		recipientFetcher: recipientFetcher,
		messageSender:    messageSender,
		progressTracker:  progressTracker,
		logger:           logger,
		config:           config,
	}
}

// CanProcess returns true if this processor can handle the given task type
func (o *BroadcastOrchestrator) CanProcess(taskType string) bool {
	return taskType == "send_broadcast"
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

	// Phase 1: Get recipient count if not already set
	if broadcastState.TotalRecipients == 0 {
		count, err := o.recipientFetcher.GetTotalRecipientCount(ctx, task.WorkspaceID, broadcastState.BroadcastID)
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

	// Initialize progress tracker
	err := o.progressTracker.Initialize(
		ctx,
		task.WorkspaceID,
		task.ID,
		broadcastState.BroadcastID,
		broadcastState.TotalRecipients,
	)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Failed to initialize progress tracker")
		return false, err
	}

	// Phase 2: Load templates
	templates, err := o.templateLoader.LoadTemplatesForBroadcast(ctx, task.WorkspaceID, broadcastState.BroadcastID)
	if err != nil {
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Failed to load templates for broadcast")
		return false, err
	}

	// Validate templates
	if err := o.templateLoader.ValidateTemplates(templates); err != nil {
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
		recipients, err := o.recipientFetcher.FetchBatch(
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

		// Update progress
		o.progressTracker.Increment(sent, failed)

		// Save progress to the task
		if err := o.progressTracker.Save(ctx, task.WorkspaceID, task.ID); err != nil {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        err.Error(),
			}).Error("Failed to save progress state")
			// Continue processing despite save error
		}

		// Update local counters for state
		broadcastState.SentCount += sent
		broadcastState.FailedCount += failed
		broadcastState.RecipientOffset += int64(len(recipients))

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
	state := o.progressTracker.GetState()
	task.State = state
	task.Progress = state.Progress

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
