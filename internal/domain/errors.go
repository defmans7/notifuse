package domain

import (
	"fmt"
)

// Common error types
type ErrNotFound struct {
	Entity string
	ID     string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found with ID: %s", e.Entity, e.ID)
}

// Task-specific errors
type ErrTaskExecution struct {
	TaskID string
	Reason string
	Err    error
}

func (e *ErrTaskExecution) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("task execution failed [%s]: %s - %v", e.TaskID, e.Reason, e.Err)
	}
	return fmt.Sprintf("task execution failed [%s]: %s", e.TaskID, e.Reason)
}

func (e *ErrTaskExecution) Unwrap() error {
	return e.Err
}

type ErrTaskTimeout struct {
	TaskID     string
	MaxRuntime int
}

func (e *ErrTaskTimeout) Error() string {
	return fmt.Sprintf("task timed out [%s] after %d seconds", e.TaskID, e.MaxRuntime)
}

// Broadcast-specific errors
type ErrBroadcastDelivery struct {
	BroadcastID string
	ContactID   string
	Email       string
	Reason      string
	Err         error
}

func (e *ErrBroadcastDelivery) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("broadcast delivery failed [%s] to %s: %s - %v", e.BroadcastID, e.Email, e.Reason, e.Err)
	}
	return fmt.Sprintf("broadcast delivery failed [%s] to %s: %s", e.BroadcastID, e.Email, e.Reason)
}

func (e *ErrBroadcastDelivery) Unwrap() error {
	return e.Err
}

type ErrBroadcastInvalidState struct {
	BroadcastID   string
	CurrentState  string
	ExpectedState string
}

func (e *ErrBroadcastInvalidState) Error() string {
	return fmt.Sprintf("broadcast [%s] in invalid state: current=%s, expected=%s", e.BroadcastID, e.CurrentState, e.ExpectedState)
}

// ValidationError represents an error that occurs due to invalid input or parameters
type ValidationError struct {
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error with the given message
func NewValidationError(message string) error {
	return ValidationError{
		Message: message,
	}
}
