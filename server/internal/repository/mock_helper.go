package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// SetupMockDB creates a new mock database and returns the db, mock, and a cleanup function
func SetupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err, "Failed to create mock database")

	cleanup := func() {
		db.Close()
	}

	return db, mock, cleanup
}
