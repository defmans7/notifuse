package broadcast

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testCircuitBreaker is a simplified implementation of the circuit breaker for testing
type testCircuitBreaker struct {
	failures       int
	threshold      int
	cooldownPeriod time.Duration
	lastFailure    time.Time
	isOpen         bool
	mutex          sync.RWMutex
}

// IsOpen checks if the circuit is open (preventing further calls)
func (cb *testCircuitBreaker) IsOpen() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// If circuit is open, check if cooldown period has passed
	if cb.isOpen {
		if time.Since(cb.lastFailure) > cb.cooldownPeriod {
			// Reset circuit after cooldown
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.isOpen = false
			cb.failures = 0
			cb.mutex.Unlock()
			cb.mutex.RLock()
		}
	}

	return cb.isOpen
}

// RecordSuccess records a successful call
func (cb *testCircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.isOpen = false
}

// RecordFailure records a failed call and opens circuit if threshold is reached
func (cb *testCircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.isOpen = true
	}
}

// TestCircuitBreakerStandalone tests the circuit breaker in isolation
func TestCircuitBreakerStandalone(t *testing.T) {
	// Create a circuit breaker directly
	cb := &testCircuitBreaker{
		threshold:      3,
		cooldownPeriod: 1 * time.Second,
	}

	// Initially circuit should be closed
	assert.False(t, cb.IsOpen())

	// Record failures
	cb.RecordFailure()
	cb.RecordFailure()
	assert.False(t, cb.IsOpen(), "Circuit should still be closed after 2 failures")

	// Third failure should open the circuit
	cb.RecordFailure()
	assert.True(t, cb.IsOpen(), "Circuit should be open after 3 failures")

	// Record success should reset the failure count and close the circuit
	cb.RecordSuccess()
	assert.False(t, cb.IsOpen(), "Circuit should be closed after success")

	// Test cooldown period
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure() // This should open the circuit
	assert.True(t, cb.IsOpen(), "Circuit should be open after 3 failures")

	// Wait for cooldown period to expire
	time.Sleep(1100 * time.Millisecond)
	assert.False(t, cb.IsOpen(), "Circuit should be closed after cooldown period")
}
