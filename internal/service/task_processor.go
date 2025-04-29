package service

import (
	"context"
	"fmt"
	"time"

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
		// In a real implementation, we would fetch this from the broadcast service
		// For this example, we'll simulate a broadcast with recipients
		broadcastState.TotalRecipients = 1000
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
	case <-time.After(3 * time.Second): // Simulate work
		// In a real implementation, we would call the broadcast service to send messages
		// Simulate processing by calculating sent and failed counts
		batchSize := broadcastState.BatchSize
		if broadcastState.CurrentBatch == broadcastState.TotalBatches {
			// Last batch might be smaller
			batchSize = broadcastState.TotalRecipients - ((broadcastState.CurrentBatch - 1) * broadcastState.BatchSize)
		}

		// Simulate some failures (e.g., 5% failure rate)
		failuresThisBatch := batchSize / 20
		successesThisBatch := batchSize - failuresThisBatch

		// Update counters
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
		subtask.State.SendBroadcast = &domain.SendBroadcastState{
			BroadcastID:     parentTask.State.SendBroadcast.BroadcastID,
			ChannelType:     parentTask.State.SendBroadcast.ChannelType,
			BatchSize:       parentTask.State.SendBroadcast.BatchSize,
			TotalRecipients: parentTask.State.SendBroadcast.TotalRecipients / subtask.Total,
			SentCount:       0,
			FailedCount:     0,
			CurrentBatch:    1,
			TotalBatches:    parentTask.State.SendBroadcast.TotalBatches / subtask.Total,
		}
	}

	// Process the subtask (similar to the main task, but just for a subset of recipients)
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(2 * time.Second): // Simulate work
		broadcastState := subtask.State.SendBroadcast

		// Simulate sending messages
		batchSize := broadcastState.BatchSize
		if broadcastState.CurrentBatch == broadcastState.TotalBatches {
			// Last batch might be smaller
			batchSize = broadcastState.TotalRecipients - ((broadcastState.CurrentBatch - 1) * broadcastState.BatchSize)
		}

		// Simulate some failures
		failuresThisBatch := batchSize / 20
		successesThisBatch := batchSize - failuresThisBatch

		// Update counters
		broadcastState.SentCount += successesThisBatch
		broadcastState.FailedCount += failuresThisBatch

		// Update subtask progress
		if broadcastState.TotalRecipients > 0 {
			subtask.Progress = float64(broadcastState.CurrentBatch) / float64(broadcastState.TotalBatches) * 100
			subtask.State.Progress = subtask.Progress
		}

		// Check if we've processed all batches for this subtask
		if broadcastState.CurrentBatch >= broadcastState.TotalBatches {
			subtask.Progress = 100
			subtask.State.Progress = 100
			subtask.State.Message = fmt.Sprintf("Completed: %d sent, %d failed",
				broadcastState.SentCount, broadcastState.FailedCount)

			return true, nil
		}

		// Move to next batch
		broadcastState.CurrentBatch++
		subtask.State.Message = fmt.Sprintf("Processing batch %d/%d",
			broadcastState.CurrentBatch, broadcastState.TotalBatches)

		return false, nil
	}
}
