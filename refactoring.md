## Recommended Refactoring Approach

### 1. Component Decomposition

#### Current Issues

The `SendBroadcastProcessor` currently:

- Loads templates
- Fetches recipients in batches
- Tracks progress
- Manages concurrent sending
- Handles error recovery
- Updates task state

This violates the Single Responsibility Principle and makes the code difficult to test and maintain.

#### Proposed Solution

Split `SendBroadcastProcessor` into smaller, focused components with well-defined interfaces:

```
SendBroadcastProcessor (Orchestrator)
├── BroadcastTemplateLoader
│   ├── LoadTemplatesForBroadcast(ctx, workspaceID, broadcastID)
│   └── ValidateTemplates(templates)
│
├── RecipientFetcher
│   ├── GetTotalRecipientCount(ctx, workspaceID, broadcastID)
│   ├── FetchBatch(ctx, workspaceID, broadcastID, offset, limit)
│   └── GetRecipientIterator(ctx, workspaceID, broadcastID)
│
├── MessageSender
│   ├── SendToRecipient(ctx, recipient, template, data)
│   ├── SendBatch(ctx, recipients, template, data)
│   └── TrackDeliveryStatus(ctx, messageID, status)
│
└── ProgressTracker
    ├── UpdateProgress(sent, failed, total)
    ├── GetProgressPercentage()
    └── GetProgressMessage()
```

#### Implementation Example

```go
// Template Loader
type TemplateLoader interface {
    LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error)
    ValidateTemplates(templates map[string]*domain.Template) error
}

type BroadcastTemplateLoader struct {
    templateService domain.TemplateService
    logger          logger.Logger
}

func (l *BroadcastTemplateLoader) LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error) {
    // Implementation
}

// Recipient Fetcher
type RecipientFetcher interface {
    GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error)
    FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.Contact, error)
}

// Message Sender
type MessageSender interface {
    SendToRecipient(ctx context.Context, workspaceID, broadcastID string, recipient *domain.Contact,
                   template *domain.Template, data map[string]interface{}) error
    SendBatch(ctx context.Context, workspaceID, broadcastID string, recipients []*domain.Contact,
             templates map[string]*domain.Template, data map[string]interface{}) (sent, failed int, err error)
}

// Progress Tracker
type ProgressTracker interface {
    UpdateProgress(sent, failed, total int)
    GetProgressPercentage() float64
    GetProgressMessage() string
    SaveProgress(ctx context.Context, taskID string) error
}

// Orchestrator
type BroadcastSendingOrchestrator struct {
    templateLoader  TemplateLoader
    recipientFetcher RecipientFetcher
    messageSender   MessageSender
    progressTracker ProgressTracker
    logger          logger.Logger
    config          BroadcastProcessorConfig
}

func (o *BroadcastSendingOrchestrator) Process(ctx context.Context, task *domain.Task) (bool, error) {
    // Orchestration logic that uses the component interfaces
}
```

#### Benefits

1. **Testability**: Each component can be tested in isolation
2. **Maintainability**: Changes to one aspect (e.g., template loading) don't affect other components
3. **Flexibility**: Components can be replaced or modified independently
4. **Reusability**: Components can be reused across different processors
5. **Parallelization**: Clear boundaries make it easier to parallelize work

### 2. Consistent Error Handling

Implement a standardized error handling strategy:

```go
// Example of improved error handling
type ErrorCode string

const (
    ErrCodeTemplateMissing ErrorCode = "TEMPLATE_MISSING"
    ErrCodeRecipientFetch  ErrorCode = "RECIPIENT_FETCH_FAILED"
    // ...
)

type BroadcastError struct {
    Code    ErrorCode
    Message string
    Task    string
    Retryable bool
    Err     error
}
```

### 3. State Machine Pattern

Implement a proper state machine for tasks:

```
PendingState → RunningState → (CompletedState | FailedState | PausedState)
```

Each state should encapsulate its own behavior and transition logic.

### 4. Configuration Management

Move configuration to an external system:

```go
type BroadcastProcessorConfig struct {
    MaxParallelism     int           `json:"max_parallelism"`
    MaxProcessTime     time.Duration `json:"max_process_time"`
    FetchBatchSize     int           `json:"fetch_batch_size"`
    ProgressLogInterval time.Duration `json:"progress_log_interval"`
}
```

### 5. Resource Management

Implement proper resource pooling and limits:

```go
// Example connection pool
type ConnectionPool struct {
    maxConnections int
    semaphore      *semaphore.Weighted
    // ...
}

// With circuit breaker
type CircuitBreaker struct {
    failures int
    threshold int
    cooldownPeriod time.Duration
    lastFailure time.Time
}
```

## Implementation Plan

### Phase 1: Core Infrastructure (2-3 weeks)

1. **Create Interfaces and Base Types** (Week 1)

   - Define all interfaces for components (TemplateLoader, RecipientFetcher, etc.)
   - Create error types and codes
   - Define state machine interfaces and states
   - Add configuration structures

2. **Implement Adapter Components** (Week 1-2)

   - Create adapter implementations that wrap existing functionality
   - These adapters will implement the new interfaces but delegate to existing code
   - This allows for incremental migration without breaking current functionality

3. **Add Observability and Metrics** (Week 2)
   - Implement metrics collection
   - Add structured logging
   - Set up tracing infrastructure

### Phase 2: Component Implementation (3-4 weeks)

4. **Template Loader Component** (Week 3)

   ```go
   // internal/service/broadcast/template_loader.go
   package broadcast

   type TemplateLoader struct {
       templateService domain.TemplateService
       logger          logger.Logger
       metrics         metrics.Reporter
   }

   func NewTemplateLoader(templateService domain.TemplateService, logger logger.Logger, metrics metrics.Reporter) *TemplateLoader {
       return &TemplateLoader{
           templateService: templateService,
           logger:          logger,
           metrics:         metrics,
       }
   }

   func (l *TemplateLoader) LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error) {
       startTime := time.Now()
       defer func() {
           l.metrics.Timer("broadcast.template_loading.duration").Record(time.Since(startTime))
       }()

       // Get the broadcast to access its template variations
       broadcast, err := l.templateService.GetBroadcast(ctx, workspaceID, broadcastID)
       if err != nil {
           l.logger.WithFields(map[string]interface{}{
               "broadcast_id": broadcastID,
               "workspace_id": workspaceID,
               "error":        err.Error(),
           }).Error("Failed to get broadcast for templates")
           return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
       }

       // Similar to existing code but with better error handling and metrics
       variationTemplates := make(map[string]*domain.Template)
       // ...implementation details...

       return variationTemplates, nil
   }
   ```

5. **Recipient Fetcher Component** (Week 3-4)

   ```go
   // internal/service/broadcast/recipient_fetcher.go
   package broadcast

   type RecipientFetcher struct {
       contactService domain.ContactRepository
       logger         logger.Logger
       metrics        metrics.Reporter
   }

   func (f *RecipientFetcher) GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
       // Implementation with metrics and error handling
   }

   func (f *RecipientFetcher) FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.Contact, error) {
       // Implementation with metrics and error handling
   }
   ```

6. **Message Sender Component** (Week 4-5)

   ```go
   // internal/service/broadcast/message_sender.go
   // Implementation of message sending with rate limiting and circuit breaking
   ```

7. **Progress Tracker Component** (Week 5)

   ```go
   // internal/service/broadcast/progress_tracker.go
   // Implementation of progress tracking and status updates
   ```

8. **Orchestrator Implementation** (Week 6)
   ```go
   // internal/service/broadcast/orchestrator.go
   // Implementation of the main orchestrator that uses all components
   ```

### Phase 3: Integration and Cleanup (2-3 weeks)

9. **Wire Up Dependencies** (Week 7)

   ```go
   // cmd/api/app.go or relevant dependency initialization
   func (a *App) InitServices() error {
       // Existing code...

       // Create and register new components
       templateLoader := broadcast.NewTemplateLoader(a.templateService, a.logger, a.metrics)
       recipientFetcher := broadcast.NewRecipientFetcher(a.contactRepo, a.logger, a.metrics)
       messageSender := broadcast.NewMessageSender(a.emailService, a.logger, a.metrics, broadcastConfig)
       progressTracker := broadcast.NewProgressTracker(a.logger, a.metrics)

       // Create orchestrator
       orchestrator := broadcast.NewBroadcastOrchestrator(
           templateLoader,
           recipientFetcher,
           messageSender,
           progressTracker,
           a.logger,
           broadcastConfig,
       )

       // Register as a task processor
       a.taskService.RegisterProcessor(orchestrator)

       return nil
   }
   ```

10. **Migrate Existing Tasks** (Week 7-8)

    - Update data migration process to ensure state compatibility
    - Create database migration scripts if needed

11. **Finalize Configuration** (Week 8)

    ```go
    // config/broadcast.go
    package config

    import "time"

    type BroadcastConfig struct {
        MaxParallelism     int           `json:"max_parallelism"`
        MaxProcessTime     time.Duration `json:"max_process_time"`
        FetchBatchSize     int           `json:"fetch_batch_size"`
        ProgressLogInterval time.Duration `json:"progress_log_interval"`
        EnableCircuitBreaker bool         `json:"enable_circuit_breaker"`
        CircuitBreakerThreshold int       `json:"circuit_breaker_threshold"`
        CircuitBreakerCooldown time.Duration `json:"circuit_breaker_cooldown"`
    }

    func LoadBroadcastConfig() (*BroadcastConfig, error) {
        // Load from environment, config file, or database
    }
    ```

12. **Cleanup and Deprecation** (Week 9)
    - Mark old code as deprecated with clear migration instructions
    - Create deprecation timeline for complete removal
    - Document the new system

### Phase 4: Testing and Validation (1-2 weeks)

13. **Create Comprehensive Tests** (Weeks 7-9)

    - Unit tests for each component
    - Integration tests for the orchestrator
    - Performance tests to verify scaling characteristics
    - Chaos tests to verify resilience

14. **Validate in Production** (Week 10)
    - Deploy with feature flag to gradually adopt the new system
    - Monitor performance and error rates
    - Collect user feedback

### Detailed Testing Strategy

```go
// Example unit test for TemplateLoader
func TestTemplateLoader_LoadTemplatesForBroadcast(t *testing.T) {
    // Setup test environment
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockTemplateService := mocks.NewMockTemplateService(ctrl)
    mockLogger := mocks.NewMockLogger(ctrl)
    mockMetrics := mocks.NewMockMetricsReporter(ctrl)

    loader := broadcast.NewTemplateLoader(mockTemplateService, mockLogger, mockMetrics)

    // Setup expectations
    mockBroadcast := &domain.Broadcast{
        ID: "broadcast-123",
        TestSettings: domain.BroadcastTestSettings{
            Variations: []domain.BroadcastVariation{
                {ID: "var-1", TemplateID: "template-1"},
                {ID: "var-2", TemplateID: "template-2"},
            },
        },
    }

    mockTemplate1 := &domain.Template{ID: "template-1", /* other fields */}
    mockTemplate2 := &domain.Template{ID: "template-2", /* other fields */}

    mockTemplateService.EXPECT().
        GetBroadcast(gomock.Any(), "workspace-1", "broadcast-123").
        Return(mockBroadcast, nil)

    mockTemplateService.EXPECT().
        GetTemplateByID(gomock.Any(), "workspace-1", "template-1").
        Return(mockTemplate1, nil)

    mockTemplateService.EXPECT().
        GetTemplateByID(gomock.Any(), "workspace-1", "template-2").
        Return(mockTemplate2, nil)

    mockMetrics.EXPECT().
        Timer("broadcast.template_loading.duration").
        Return(mockMetrics)

    mockMetrics.EXPECT().
        Record(gomock.Any())

    // Execute test
    templates, err := loader.LoadTemplatesForBroadcast(context.Background(), "workspace-1", "broadcast-123")

    // Assert results
    assert.NoError(t, err)
    assert.Equal(t, 2, len(templates))
    assert.Equal(t, mockTemplate1, templates["template-1"])
    assert.Equal(t, mockTemplate2, templates["template-2"])
}
```

## Migration Considerations

### Database Changes

No schema changes are required initially, but for better performance we might consider:

1. Adding indices on commonly queried broadcast and task fields
2. Optimizing task state storage with a more efficient JSON structure
3. Adding caching for template and recipient data

### Backward Compatibility

Maintain backward compatibility by:

1. Using adapters to bridge old and new implementations
2. Keeping task state compatible with both systems
3. Implementing A/B testing between old and new implementations

### Performance Monitoring

Monitor key metrics during and after migration:

1. Task processing time (should improve)
2. Memory usage (should decrease)
3. Error rates (should decrease)
4. Throughput (emails sent per minute)

## Rollout Strategy

1. **Develop and Test**: Implement all components with comprehensive tests
2. **Canary Deployment**: Deploy to a subset of workspaces for initial validation
3. **Gradual Rollout**: Enable the new system for more workspaces over time
4. **Full Deployment**: Switch all workspaces to the new system
5. **Deprecation**: Remove old code after a safe period
