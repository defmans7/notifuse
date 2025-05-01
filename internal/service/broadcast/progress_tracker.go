package broadcast

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// ProgressTracker is the interface for tracking progress of broadcast sending
//
//go:generate mockgen -destination=../../mocks/progress_tracker.go -package=mocks github.com/Notifuse/notifuse/internal/service/broadcast ProgressTracker
type ProgressTracker interface {
	// Initialize sets up the tracker with initial state
	Initialize(ctx context.Context, workspaceID, taskID, broadcastID string, totalRecipients int) error

	// Increment updates progress metrics
	Increment(sent, failed int)

	// GetProgress returns the current percentage (0-100)
	GetProgress() float64

	// GetState returns the current state for persisting to storage
	GetState() *domain.TaskState

	// GetMessage returns a human-readable progress message
	GetMessage() string

	// Save persists the current progress to storage
	Save(ctx context.Context, workspaceID, taskID string) error
}

// progressTracker implements the ProgressTracker interface
type progressTracker struct {
	logger logger.Logger
	repo   domain.TaskRepository
	config *Config

	// State tracking
	taskID          string
	broadcastID     string
	totalRecipients int
	sentCount       int
	failedCount     int
	processedCount  int
	startTime       time.Time
	lastSaveTime    time.Time
	lastLogTime     time.Time

	// Locking
	mu sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(logger logger.Logger, repo domain.TaskRepository, config *Config) ProgressTracker {
	if config == nil {
		config = DefaultConfig()
	}

	return &progressTracker{
		logger:       logger,
		repo:         repo,
		config:       config,
		startTime:    time.Now(),
		lastSaveTime: time.Now(),
		lastLogTime:  time.Now(),
	}
}

// Initialize sets up the tracker with initial state
func (t *progressTracker) Initialize(ctx context.Context, workspaceID, taskID, broadcastID string, totalRecipients int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.taskID = taskID
	t.broadcastID = broadcastID
	t.totalRecipients = totalRecipients
	t.sentCount = 0
	t.failedCount = 0
	t.processedCount = 0
	t.startTime = time.Now()
	t.lastSaveTime = time.Now()
	t.lastLogTime = time.Now()

	// Initialize or restore state from task
	task, err := t.repo.Get(ctx, workspaceID, taskID)
	if err != nil {
		t.logger.WithFields(map[string]interface{}{
			"task_id":      taskID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get task for progress tracking")
		return NewBroadcastError(ErrCodeTaskStateInvalid, "failed to get task", false, err)
	}

	// If task already has progress, restore it
	if task.State != nil && task.State.SendBroadcast != nil {
		state := task.State.SendBroadcast
		t.sentCount = state.SentCount
		t.failedCount = state.FailedCount
		t.processedCount = int(state.RecipientOffset)

		t.logger.WithFields(map[string]interface{}{
			"task_id":         taskID,
			"workspace_id":    workspaceID,
			"sent_count":      t.sentCount,
			"failed_count":    t.failedCount,
			"processed_count": t.processedCount,
		}).Info("Restored progress state from task")
	}

	return nil
}

// Increment updates progress metrics
func (t *progressTracker) Increment(sent, failed int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.sentCount += sent
	t.failedCount += failed
	t.processedCount += sent + failed

	// Log progress at regular intervals
	if time.Since(t.lastLogTime) >= t.config.ProgressLogInterval {
		t.logger.WithFields(map[string]interface{}{
			"task_id":         t.taskID,
			"broadcast_id":    t.broadcastID,
			"sent_count":      t.sentCount,
			"failed_count":    t.failedCount,
			"processed_count": t.processedCount,
			"total_count":     t.totalRecipients,
			"progress":        t.calculateProgress(),
			"elapsed":         time.Since(t.startTime).String(),
		}).Info("Broadcast progress update")

		t.lastLogTime = time.Now()
	}
}

// GetProgress returns the current percentage (0-100)
func (t *progressTracker) GetProgress() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.calculateProgress()
}

// calculateProgress calculates the progress percentage (0-100)
func (t *progressTracker) calculateProgress() float64 {
	if t.totalRecipients <= 0 {
		return 100.0 // Avoid division by zero
	}

	progress := float64(t.processedCount) / float64(t.totalRecipients) * 100.0
	if progress > 100.0 {
		progress = 100.0
	}
	return progress
}

// GetState returns the current state for persisting to storage
func (t *progressTracker) GetState() *domain.TaskState {
	t.mu.Lock()
	defer t.mu.Unlock()

	progress := t.calculateProgress()
	message := t.formatMessage()

	return &domain.TaskState{
		Progress: progress,
		Message:  message,
		SendBroadcast: &domain.SendBroadcastState{
			BroadcastID:     t.broadcastID,
			TotalRecipients: t.totalRecipients,
			SentCount:       t.sentCount,
			FailedCount:     t.failedCount,
			ChannelType:     "email", // Hardcoded for now
			RecipientOffset: int64(t.processedCount),
		},
	}
}

// GetMessage returns a human-readable progress message
func (t *progressTracker) GetMessage() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.formatMessage()
}

// formatMessage creates a human-readable progress message
func (t *progressTracker) formatMessage() string {
	progress := t.calculateProgress()

	// Calculate remaining time if we have processed more than 5%
	var eta string
	if progress > 5.0 && t.processedCount > 0 {
		elapsed := time.Since(t.startTime)
		estimatedTotal := elapsed.Seconds() * float64(t.totalRecipients) / float64(t.processedCount)
		remaining := estimatedTotal - elapsed.Seconds()

		if remaining > 0 {
			eta = fmt.Sprintf(", ETA: %s", formatDuration(time.Duration(remaining)*time.Second))
		}
	}

	return fmt.Sprintf("Processed %d/%d recipients (%.1f%%)%s",
		t.processedCount, t.totalRecipients, progress, eta)
}

// formatDuration formats a duration in a human-readable form
func formatDuration(d time.Duration) string {
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

// Save persists the current progress to storage
func (t *progressTracker) Save(ctx context.Context, workspaceID, taskID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Skip saving if not enough time has passed since the last save
	if time.Since(t.lastSaveTime) < 5*time.Second {
		return nil
	}

	state := t.GetState()

	err := t.repo.SaveState(ctx, workspaceID, taskID, state.Progress, state)
	if err != nil {
		t.logger.WithFields(map[string]interface{}{
			"task_id":      taskID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to save progress state")
		return NewBroadcastError(ErrCodeTaskStateInvalid, "failed to save task state", true, err)
	}

	t.lastSaveTime = time.Now()

	t.logger.WithFields(map[string]interface{}{
		"task_id":      taskID,
		"workspace_id": workspaceID,
		"progress":     state.Progress,
		"sent":         t.sentCount,
		"failed":       t.failedCount,
	}).Debug("Saved progress state")

	return nil
}
