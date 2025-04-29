package service

import (
	"context"
	"fmt"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// SendBroadcastProcessor implements domain.TaskProcessor for sending broadcasts
type SendBroadcastProcessor struct {
	broadcastService domain.BroadcastSender
	logger           logger.Logger
}

// NewSendBroadcastProcessor creates a new SendBroadcastProcessor
func NewSendBroadcastProcessor(broadcastService domain.BroadcastSender, logger logger.Logger) *SendBroadcastProcessor {
	return &SendBroadcastProcessor{
		broadcastService: broadcastService,
		logger:           logger,
	}
}

// CanProcess returns true if this processor can handle the given task type
func (p *SendBroadcastProcessor) CanProcess(taskType string) bool {
	return taskType == "send_broadcast"
}

// SupportsParallelization returns true since broadcast sending can be parallelized
func (p *SendBroadcastProcessor) SupportsParallelization() bool {
	return true
}

// Process executes or continues a broadcast sending task
func (p *SendBroadcastProcessor) Process(ctx context.Context, task *domain.Task) (bool, error) {
	p.logger.WithField("task_id", task.ID).Info("Processing send_broadcast task")

	// Initialize structured state if needed
	if task.State == nil {
		// Initialize a new state for the broadcast task
		task.State = &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				SentCount:    0,
				FailedCount:  0,
				CurrentBatch: 0,
				TotalBatches: 0,
				BatchSize:    100, // Default batch size
			},
		}
	}

	// Initialize the SendBroadcast state if it doesn't exist yet
	if task.State.SendBroadcast == nil {
		task.State.SendBroadcast = &domain.SendBroadcastState{
			SentCount:    0,
			FailedCount:  0,
			CurrentBatch: 0,
			TotalBatches: 0,
			BatchSize:    100, // Default batch size
		}
	}

	// Extract broadcast ID from task state or context
	broadcastState := task.State.SendBroadcast
	if broadcastState.BroadcastID == "" {
		// In a real implementation, we'd expect the broadcast ID to be set when creating the task
		return false, fmt.Errorf("broadcast ID is missing in task state")
	}

	// If we're just starting (batch 0), get the total recipients
	if broadcastState.CurrentBatch == 0 {
		// Get actual recipient count from the broadcast service
		recipientCount, err := p.broadcastService.GetRecipientCount(ctx, task.WorkspaceID, broadcastState.BroadcastID)
		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        err.Error(),
			}).Error("Failed to get recipient count for broadcast")
			return false, fmt.Errorf("failed to get recipient count: %w", err)
		}

		broadcastState.TotalRecipients = recipientCount
		broadcastState.TotalBatches = (broadcastState.TotalRecipients + broadcastState.BatchSize - 1) / broadcastState.BatchSize
		broadcastState.CurrentBatch = 1
		broadcastState.ChannelType = "email" // Or could be "sms", "push", etc.

		task.State.Message = fmt.Sprintf("Sending %s broadcast to %d recipients",
			broadcastState.ChannelType, broadcastState.TotalRecipients)

		p.logger.WithFields(map[string]interface{}{
			"task_id":          task.ID,
			"broadcast_id":     broadcastState.BroadcastID,
			"total_recipients": broadcastState.TotalRecipients,
			"total_batches":    broadcastState.TotalBatches,
			"channel_type":     broadcastState.ChannelType,
		}).Info("Broadcast sending initialized")

		return false, nil
	}

	// Process the current batch
	select {
	case <-ctx.Done():
		// Context was canceled (e.g., timeout)
		p.logger.WithField("task_id", task.ID).Warn("Broadcast sending interrupted")
		return false, ctx.Err()
	default:
		// Calculate the current batch size
		batchSize := broadcastState.BatchSize
		if broadcastState.CurrentBatch == broadcastState.TotalBatches {
			// Last batch might be smaller
			remainingRecipients := broadcastState.TotalRecipients - ((broadcastState.CurrentBatch - 1) * broadcastState.BatchSize)
			if remainingRecipients < batchSize {
				batchSize = remainingRecipients
			}
		}

		// Use real broadcast service to send the batch
		successesThisBatch, failuresThisBatch, err := p.broadcastService.SendBatch(
			ctx,
			task.WorkspaceID,
			broadcastState.BroadcastID,
			broadcastState.CurrentBatch-1, // Use zero-based indexing for batch number
			batchSize,
		)

		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"batch":        broadcastState.CurrentBatch,
				"error":        err.Error(),
			}).Error("Failed to send broadcast batch")
			return false, fmt.Errorf("failed to send batch %d: %w", broadcastState.CurrentBatch, err)
		}

		// Update counters with real results
		broadcastState.SentCount += successesThisBatch
		broadcastState.FailedCount += failuresThisBatch

		p.logger.WithFields(map[string]interface{}{
			"task_id":    task.ID,
			"batch":      broadcastState.CurrentBatch,
			"sent_count": broadcastState.SentCount,
			"fail_count": broadcastState.FailedCount,
			"batch_size": batchSize,
			"success":    successesThisBatch,
			"failed":     failuresThisBatch,
		}).Info("Processed broadcast batch")

		// Update task message and progress
		task.State.Message = fmt.Sprintf("Sent to %d/%d recipients (%d failed)",
			broadcastState.SentCount, broadcastState.TotalRecipients, broadcastState.FailedCount)

		// Check if we've processed all batches
		if broadcastState.CurrentBatch >= broadcastState.TotalBatches {
			// Task is complete
			task.State.Message = fmt.Sprintf("Broadcast completed: %d sent, %d failed",
				broadcastState.SentCount, broadcastState.FailedCount)

			// Calculate final progress
			if broadcastState.TotalRecipients > 0 {
				task.Progress = 100.0
				task.State.Progress = 100.0
			}

			p.logger.WithFields(map[string]interface{}{
				"task_id":          task.ID,
				"broadcast_id":     broadcastState.BroadcastID,
				"total_recipients": broadcastState.TotalRecipients,
				"sent_count":       broadcastState.SentCount,
				"failed_count":     broadcastState.FailedCount,
				"progress":         task.Progress,
			}).Info("Broadcast sending completed")

			return true, nil
		}

		// Move to next batch
		broadcastState.CurrentBatch++

		// Calculate progress percentage
		if broadcastState.TotalRecipients > 0 {
			task.Progress = float64(broadcastState.CurrentBatch-1) / float64(broadcastState.TotalBatches) * 100
			task.State.Progress = task.Progress
		}

		p.logger.WithFields(map[string]interface{}{
			"task_id":  task.ID,
			"progress": task.Progress,
			"batch":    broadcastState.CurrentBatch,
			"of_batch": broadcastState.TotalBatches,
		}).Info("Moving to next batch")

		return false, nil
	}
}

// ProcessSubtask executes a portion of the task as a subtask
func (p *SendBroadcastProcessor) ProcessSubtask(ctx context.Context, subtask *domain.Subtask, parentTask *domain.Task) (bool, error) {
	p.logger.WithFields(map[string]interface{}{
		"subtask_id":     subtask.ID,
		"parent_task_id": subtask.ParentTaskID,
		"index":          subtask.Index,
		"total":          subtask.Total,
	}).Info("Processing broadcast subtask")

	// Initialize subtask state if needed
	if subtask.State.SendBroadcast == nil {
		// Make sure parent task has SendBroadcast state
		if parentTask.State == nil || parentTask.State.SendBroadcast == nil {
			return false, fmt.Errorf("parent task missing SendBroadcast state")
		}

		// Calculate which batches this subtask should process
		totalSubtasks := subtask.Total
		if totalSubtasks <= 0 {
			totalSubtasks = 1
		}

		subtaskIndex := subtask.Index
		if subtaskIndex < 0 {
			subtaskIndex = 0
		}

		// Calculate batch range for this subtask
		parentState := parentTask.State.SendBroadcast
		totalBatches := parentState.TotalBatches
		batchesPerSubtask := (totalBatches + totalSubtasks - 1) / totalSubtasks
		startBatch := subtaskIndex * batchesPerSubtask
		endBatch := startBatch + batchesPerSubtask
		if endBatch > totalBatches {
			endBatch = totalBatches
		}

		// Initialize the subtask state
		subtask.State.SendBroadcast = &domain.SendBroadcastState{
			BroadcastID:     parentState.BroadcastID,
			ChannelType:     parentState.ChannelType,
			BatchSize:       parentState.BatchSize,
			TotalRecipients: parentState.TotalRecipients,
			SentCount:       0,
			FailedCount:     0,
			CurrentBatch:    startBatch, // Start at the first batch for this subtask
			TotalBatches:    endBatch,   // End at the last batch for this subtask
		}

		p.logger.WithFields(map[string]interface{}{
			"subtask_id":   subtask.ID,
			"broadcast_id": parentState.BroadcastID,
			"start_batch":  startBatch,
			"end_batch":    endBatch,
		}).Info("Initialized broadcast subtask")
	}

	// Process the subtask
	select {
	case <-ctx.Done():
		p.logger.WithField("subtask_id", subtask.ID).Warn("Broadcast subtask interrupted")
		return false, ctx.Err()
	default:
		broadcastState := subtask.State.SendBroadcast

		// Check if we've already finished all our batches
		if broadcastState.CurrentBatch >= broadcastState.TotalBatches {
			subtask.Progress = 100
			subtask.State.Progress = 100
			subtask.State.Message = fmt.Sprintf("Completed: %d sent, %d failed",
				broadcastState.SentCount, broadcastState.FailedCount)

			p.logger.WithFields(map[string]interface{}{
				"subtask_id":   subtask.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"sent_count":   broadcastState.SentCount,
				"fail_count":   broadcastState.FailedCount,
			}).Info("Broadcast subtask completed")

			return true, nil
		}

		// Calculate this batch size
		batchSize := broadcastState.BatchSize

		// Use real broadcast service to send the batch
		successesThisBatch, failuresThisBatch, err := p.broadcastService.SendBatch(
			ctx,
			parentTask.WorkspaceID,
			broadcastState.BroadcastID,
			broadcastState.CurrentBatch, // Use the current batch number
			batchSize,
		)

		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"subtask_id":   subtask.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"batch":        broadcastState.CurrentBatch,
				"error":        err.Error(),
			}).Error("Failed to send broadcast batch in subtask")
			return false, fmt.Errorf("failed to send batch %d: %w", broadcastState.CurrentBatch, err)
		}

		// Update counters with real results
		broadcastState.SentCount += successesThisBatch
		broadcastState.FailedCount += failuresThisBatch

		// Update subtask progress
		batchesDone := broadcastState.CurrentBatch - (subtask.Index * ((parentTask.State.SendBroadcast.TotalBatches + subtask.Total - 1) / subtask.Total)) + 1
		totalBatchesForSubtask := broadcastState.TotalBatches - broadcastState.CurrentBatch

		if totalBatchesForSubtask > 0 {
			subtask.Progress = float64(batchesDone) / float64(totalBatchesForSubtask+batchesDone) * 100
		} else {
			subtask.Progress = 100
		}
		subtask.State.Progress = subtask.Progress

		// Update subtask message
		subtask.State.Message = fmt.Sprintf("Sent batch %d, %d messages sent, %d failed",
			broadcastState.CurrentBatch, broadcastState.SentCount, broadcastState.FailedCount)

		p.logger.WithFields(map[string]interface{}{
			"subtask_id":   subtask.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"batch":        broadcastState.CurrentBatch,
			"success":      successesThisBatch,
			"failed":       failuresThisBatch,
			"progress":     subtask.Progress,
		}).Info("Processed broadcast batch in subtask")

		// Move to next batch
		broadcastState.CurrentBatch++

		return false, nil
	}
}
