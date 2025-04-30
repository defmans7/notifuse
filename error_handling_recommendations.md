# Error Handling Recommendations

## Current Issues

1. **Inconsistent Error Types**: The codebase uses generic errors in many places, making it difficult to distinguish between different error conditions.

2. **Poor Error Context**: Many errors lack sufficient context about what failed and why.

3. **Inadequate Error Handling in HTTP Layer**: Error responses don't consistently map to appropriate HTTP status codes.

4. **Error Propagation Issues**: Original errors are sometimes lost when wrapping or returning them.

5. **Missing Domain-Specific Error Types**: The system lacks structured error types for common domain failures.

## Recommended Improvements

### 1. Create Domain-Specific Error Types

Create custom error types for common failure scenarios:

- `ErrNotFound`: For missing resources
- `ErrTaskExecution`: For task execution failures
- `ErrTaskTimeout`: For task timeouts
- `ErrBroadcastDelivery`: For message delivery failures
- `ErrBroadcastInvalidState`: For invalid broadcast states

### 2. Implement Proper Error Wrapping

Use Go's error wrapping capabilities to preserve the original error context:

- Utilize `fmt.Errorf("... %w", err)` to wrap errors
- Implement the `Unwrap()` method for custom error types
- Structure errors to support `errors.Is()` and `errors.As()`

### 3. Improve HTTP Error Handling

Map domain errors to appropriate HTTP status codes:

- `ErrNotFound` → 404 Not Found
- `ErrTaskExecution` → 500 Internal Server Error (or other appropriate code)
- `ErrTaskTimeout` → 504 Gateway Timeout
- `ErrBroadcastInvalidState` → 400 Bad Request

### 4. Enhance Logging

Improve error logging to capture more context:

- Log the full error chain with context
- Include relevant IDs (task ID, broadcast ID, etc.)
- Distinguish between expected errors and unexpected failures

### 5. Consistent Error Handling Patterns

Establish consistent patterns for error handling:

- Use type assertions to handle specific error types
- Return meaningful errors from all service methods
- Follow a consistent approach to error propagation

### 6. Add Error Testing

Create comprehensive tests for error scenarios:

- Unit tests for each custom error type
- Tests for error wrapping and unwrapping
- Tests for HTTP status code mapping

## Implementation Plan

1. Create a new file `internal/domain/errors.go` to define custom error types
2. Update service layer to use these error types
3. Enhance HTTP handlers to properly map errors to status codes
4. Update existing error handling in task processors
5. Add comprehensive tests for the new error types and handling
