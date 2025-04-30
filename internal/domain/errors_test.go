package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrNotFound_Error(t *testing.T) {
	err := &ErrNotFound{
		Entity: "broadcast",
		ID:     "12345",
	}

	expected := "broadcast not found with ID: 12345"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestErrTaskExecution_Error(t *testing.T) {
	// Test with nil wrapped error
	err1 := &ErrTaskExecution{
		TaskID: "task123",
		Reason: "processor not found",
	}

	expected1 := "task execution failed [task123]: processor not found"
	if err1.Error() != expected1 {
		t.Errorf("Expected error message '%s', got '%s'", expected1, err1.Error())
	}

	// Test with wrapped error
	underlyingErr := fmt.Errorf("database connection failed")
	err2 := &ErrTaskExecution{
		TaskID: "task456",
		Reason: "database error",
		Err:    underlyingErr,
	}

	expected2 := "task execution failed [task456]: database error - database connection failed"
	if err2.Error() != expected2 {
		t.Errorf("Expected error message '%s', got '%s'", expected2, err2.Error())
	}

	// Test error unwrapping
	if !errors.Is(err2, underlyingErr) {
		t.Error("errors.Is() failed to find the wrapped error")
	}
}

func TestErrTaskTimeout_Error(t *testing.T) {
	err := &ErrTaskTimeout{
		TaskID:     "task789",
		MaxRuntime: 60,
	}

	expected := "task timed out [task789] after 60 seconds"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestErrBroadcastDelivery_Error(t *testing.T) {
	// Test with nil wrapped error
	err1 := &ErrBroadcastDelivery{
		BroadcastID: "broadcast123",
		Email:       "user@example.com",
		Reason:      "template not found",
	}

	expected1 := "broadcast delivery failed [broadcast123] to user@example.com: template not found"
	if err1.Error() != expected1 {
		t.Errorf("Expected error message '%s', got '%s'", expected1, err1.Error())
	}

	// Test with wrapped error
	underlyingErr := fmt.Errorf("SMTP connection refused")
	err2 := &ErrBroadcastDelivery{
		BroadcastID: "broadcast456",
		Email:       "other@example.com",
		Reason:      "email sending failed",
		Err:         underlyingErr,
	}

	expected2 := "broadcast delivery failed [broadcast456] to other@example.com: email sending failed - SMTP connection refused"
	if err2.Error() != expected2 {
		t.Errorf("Expected error message '%s', got '%s'", expected2, err2.Error())
	}

	// Test error unwrapping
	if !errors.Is(err2, underlyingErr) {
		t.Error("errors.Is() failed to find the wrapped error")
	}
}

func TestErrBroadcastInvalidState_Error(t *testing.T) {
	err := &ErrBroadcastInvalidState{
		BroadcastID:   "broadcast789",
		CurrentState:  "draft",
		ExpectedState: "sending",
	}

	expected := "broadcast [broadcast789] in invalid state: current=draft, expected=sending"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorTypeAssertion(t *testing.T) {
	// Test that we can properly use type assertions with these errors
	var err error

	// Create an ErrNotFound
	err = &ErrNotFound{Entity: "task", ID: "123"}

	// Type assertion should work
	if _, ok := err.(*ErrNotFound); !ok {
		t.Error("Type assertion for ErrNotFound failed")
	}

	// Create an ErrTaskExecution
	err = &ErrTaskExecution{TaskID: "456", Reason: "test"}

	// Type assertion should work
	if _, ok := err.(*ErrTaskExecution); !ok {
		t.Error("Type assertion for ErrTaskExecution failed")
	}

	// Negative test - wrong type
	if _, ok := err.(*ErrNotFound); ok {
		t.Error("Type assertion incorrectly succeeded for wrong error type")
	}
}
