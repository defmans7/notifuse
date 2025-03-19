## Testing

### Unit Tests with go-sqlmock

The repository layer uses [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) to test database interactions without requiring a real database connection. This approach provides several advantages:

- **Fast execution**: Tests run without the overhead of database connections
- **Isolation**: Tests don't depend on database state or availability
- **Controlled testing**: Full control over database responses and error scenarios

Tests using go-sqlmock follow these naming conventions:

- Test functions have the suffix `_WithMock` (e.g., `TestUserRepository_GetUserByEmail_WithMock`)
- They use the `setupMockTestDB(t)` helper function to create a mock database

Run mock tests with:

```bash
go test -run ".*WithMock" ./... -v
```

### Integration Tests

Integration tests that require a real database connection use the `integration` build tag:

```bash
go test -tags=integration ./... -v
```

These tests require a PostgreSQL database configured according to the variables in `.env.test`.
