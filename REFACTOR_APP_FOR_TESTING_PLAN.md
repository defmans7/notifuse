# Plan: Refactor App for Integration Testing

## Problem

The integration tests are trying to import `github.com/Notifuse/notifuse/internal/app`, but this package doesn't exist. The complete `App` implementation is in `cmd/api/app.go` (package main), which cannot be imported by tests.

## Goal

Move the App implementation to `internal/app` package so it can be imported by integration tests while maintaining backward compatibility with the existing `cmd/api` package.

## Step-by-Step Plan

### Step 1: Create internal/app directory

```bash
mkdir -p internal/app
```

### Step 2: Copy and modify app.go to internal/app

1. Copy the current `cmd/api/app.go` to `internal/app/app.go`
2. Change package declaration from `package main` to `package app`
3. Keep all the existing functionality intact

### Step 3: Update cmd/api/app.go to use internal/app

Replace the entire content of `cmd/api/app.go` with:

```go
package main

import (
	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/app"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mailer"
)

// Re-export the types from internal/app for backward compatibility
type AppInterface = app.AppInterface
type App = app.App
type AppOption = app.AppOption

// Re-export functions from internal/app
var NewApp = app.NewApp
var WithMockDB = app.WithMockDB
var WithMockMailer = app.WithMockMailer
var WithLogger = app.WithLogger

// NewRealApp creates a new application instance (backward compatibility)
func NewRealApp(cfg *config.Config, opts ...AppOption) AppInterface {
	return app.NewApp(cfg, opts...)
}
```

### Step 4: Update cmd/api/main.go if needed

Ensure `main.go` still works with the new re-exported types and functions.

### Step 5: Verify integration tests work

The integration tests should now be able to import `github.com/Notifuse/notifuse/internal/app` and use `app.NewApp()`.

### Step 6: Update testutil/server.go if needed

Update the `AppInterface` definition in `testutil/server.go` to match the one in `internal/app` if there are any differences.

## Benefits

1. ✅ Integration tests can import and use the real App implementation
2. ✅ Backward compatibility maintained for cmd/api
3. ✅ No breaking changes to existing code
4. ✅ Follows Go best practices (internal packages for internal use)
5. ✅ Single source of truth for App implementation

## Commands to Execute

```bash
# Step 1: Create directory
mkdir -p internal/app

# Step 2: Copy original app.go
cp cmd/api/app.go internal/app/app.go

# Step 3: Change package declaration in internal/app/app.go
sed -i '' 's/package main/package app/' internal/app/app.go

# Step 4: Replace cmd/api/app.go with wrapper
# (This will be done manually with the content above)

# Step 5: Test the changes
go test ./tests/integration/... -v
```

## Files to be Modified

- `internal/app/app.go` (new file, copied from cmd/api/app.go)
- `cmd/api/app.go` (replaced with wrapper)
- `cmd/api/main.go` (verify compatibility)

## Validation Steps

1. Run integration tests: `go test ./tests/integration/... -v`
2. Run cmd/api tests: `go test ./cmd/api/... -v`
3. Build the main application: `go build ./cmd/api`
4. Verify no import cycles exist: `go mod verify`

## Rollback Plan

If issues arise:

1. Delete `internal/app/app.go`
2. Restore original `cmd/api/app.go` from git
3. Remove `internal/app` directory if empty

## Notes

- The App implementation is well-structured and already has proper interfaces
- The AppInterface is already defined and used in tests
- This refactor maintains all existing functionality while enabling testability
- The internal/app package follows Go conventions for internal packages
