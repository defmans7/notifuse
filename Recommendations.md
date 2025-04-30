# Broadcast System Refactoring Recommendations

## Overview

This document outlines recommendations for refactoring the broadcast/task system in Notifuse to improve code quality, maintainability, and performance. The main focus is on the send_broadcast task flow which handles email campaign delivery.

## Current Architecture

The current system consists of several interconnected components:

- **TaskService**: Manages the lifecycle of background tasks
- **SendBroadcastProcessor**: Handles the actual email sending logic
- **BroadcastService**: Manages broadcast campaigns and their state
- **TaskHandler**: HTTP endpoints for task management

## Key Issues Identified

### 1. Separation of Concerns

The `SendBroadcastProcessor` handles too many responsibilities including template loading, recipient fetching, and message sending. This makes the code difficult to maintain and test.

### 2. Error Handling

Error handling is inconsistent across the codebase with different approaches in different components. Some errors are wrapped in domain-specific errors while others are returned directly.

### 3. State Management

Task state is managed through complex nested structs with nullable fields, making it prone to errors and difficult to reason about.

### 4. Configuration

Configuration values like `maxParallelism` and `maxProcessTime` are hardcoded, limiting flexibility.

### 5. Resource Management

The code creates multiple goroutines without proper resource control, potentially leading to resource exhaustion.

### 6. Progress Tracking

Create a dedicated progress tracker:

```go
type ProgressTracker interface {
    Increment(sent, failed int)
    GetProgress() float64
    GetMessage() string
    Save(ctx context.Context) error
}
```

### 7. Observability

Add structured metrics and tracing:

```go
// Metrics example
metrics.Counter("broadcast.recipients.processed").Inc(processedCount)
metrics.Gauge("broadcast.progress").Set(currentProgress)
metrics.Timer("broadcast.send.duration").Record(elapsed)
```

### 8. Testing Improvements

Restructure for better testability:

- Add more interfaces to facilitate mocking
- Create test helpers for common testing scenarios
- Implement property-based testing for complex logic

## Implementation Plan

1. Create new interfaces and data structures
2. Implement the core components with proper tests
3. Create adapters to integrate with existing code
4. Gradually migrate functionality to new components
5. Remove deprecated code once migration is complete

## Long-term Considerations

- Consider using a dedicated task queue system (e.g., Temporal, Cadence)
- Explore event-sourcing for more robust state management
- Implement a proper workflow engine for complex broadcast scenarios
- Consider separating the broadcast worker from the API service

## Conclusion

Refactoring the broadcast system as outlined above will improve code quality, maintainability, and resilience. The modular approach will also make it easier to extend the system with new features in the future.
