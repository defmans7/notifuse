# Server Architecture Documentation

## Project Structure

```
server/
├── cmd/
│   └── api/
│       └── main.go           # Main application entry point
├── internal/
│   ├── domain/              # Domain models and interfaces
│   │   └── user.go         # User and Session models, Repository interface
│   ├── repository/         # Data access layer
│   │   ├── user_postgres.go    # PostgreSQL implementation
│   │   ├── user_postgres_test.go # Repository tests
│   │   └── test_helper.go      # Test utilities
│   ├── service/           # Business logic layer
│   │   ├── user.go       # User service implementation
│   │   └── user_test.go  # Service tests
│   └── http/             # HTTP handlers layer
│       ├── user_handler.go    # User endpoints handlers
│       └── user_handler_test.go # Handler tests
├── migrations/
│   └── 001_create_tables.sql  # Database schema
├── go.mod                     # Go module file
├── go.sum                     # Go dependencies checksum
├── Makefile                  # Build and development commands
└── README.md                 # Project documentation

```

## Dependencies

```go
// Core dependencies
"database/sql"      // Standard database interface
"github.com/lib/pq" // PostgreSQL driver

// Authentication
"github.com/o1egl/paseto" // Platform-Agnostic Security Tokens

// Testing
"github.com/stretchr/testify/assert"  // Testing assertions
"github.com/stretchr/testify/mock"    // Mocking framework
"github.com/stretchr/testify/require" // Required assertions

// Utilities
"github.com/google/uuid" // UUID generation
```

## Architecture

The project follows a clean architecture pattern with the following layers:

### 1. Domain Layer (`internal/domain/`)

- Contains core business models and interfaces
- Defines repository interfaces
- No external dependencies

```go
type User struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Session struct {
    ID        string
    UserID    string
    ExpiresAt time.Time
    CreatedAt time.Time
}

type UserRepository interface {
    CreateUser(ctx context.Context, user *User) error
    GetUserByEmail(ctx context.Context, email string) (*User, error)
    // ... other methods
}
```

### 2. Repository Layer (`internal/repository/`)

- Implements data access logic
- Uses standard `database/sql` package
- Handles database operations and mapping
- PostgreSQL-specific implementation

### 3. Service Layer (`internal/service/`)

- Implements business logic
- Handles authentication flows
- Manages sessions and tokens
- Uses PASETO for secure tokens

### 4. HTTP Layer (`internal/http/`)

- HTTP handlers for the API endpoints
- Request/response handling
- Route registration
- Basic validation

## Authentication Flow

1. **Sign Up**

   - Endpoint: `POST /api/auth/signup`
   - Accepts email and name
   - Sends verification token via email
   - Token expires in 15 minutes

2. **Sign In**

   - Endpoint: `POST /api/auth/signin`
   - Accepts email
   - Sends magic link token via email
   - Token expires in 15 minutes

3. **Verify Token**
   - Endpoint: `POST /api/auth/verify`
   - Validates magic link token
   - Creates user session
   - Returns PASETO authentication token
   - Session expires in 15 days

## Database Schema

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

## Environment Variables

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5432/notifuse?sslmode=disable"
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/notifuse_test?sslmode=disable"
PASETO_PRIVATE_KEY="your-private-key"
PASETO_PUBLIC_KEY="your-public-key"
PORT="8080"  # Optional, defaults to 8080
```

## Testing Strategy

1. **Repository Tests**

   - Integration tests with real PostgreSQL database
   - Clean database before each test
   - Test all CRUD operations

2. **Service Tests**

   - Unit tests with mocked repository
   - Test authentication flows
   - Test token generation and validation

3. **Handler Tests**
   - Unit tests with mocked service
   - Test request/response handling
   - Test error scenarios

## Code Generation Guidelines

When generating code for this project:

1. Follow the established package structure
2. Use standard library `database/sql` for database operations
3. Implement proper error handling and custom error types
4. Include comprehensive tests with mocks where appropriate
5. Follow Go best practices and idioms
6. Use contexts for cancellation and timeouts
7. Implement proper validation and security measures
