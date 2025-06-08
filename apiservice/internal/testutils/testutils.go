package testutils

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testConn *pgxpool.Pool
	once     sync.Once
)

// SetupTestDB initializes and returns a connection pool to the test database.
// It reads the connection URL from the TEST_DATABASE_URL environment variable.
// The setup is run only once using sync.Once to avoid multiple initializations.
func SetupTestDB() (*pgxpool.Pool, error) {
	var err error
	once.Do(func() {
		dbURL := os.Getenv("TEST_DATABASE_URL")
		if dbURL == "" {
			err = errors.New("TEST_DATABASE_URL must be set")
			return
		}
		testConn, err = pgxpool.New(context.Background(), dbURL)
		if err != nil {
			return
		}
	})
	return testConn, err
}

// TeardownTestDB closes the test database connection pool if it was initialized.
func TeardownTestDB() {
	if testConn != nil {
		testConn.Close()
	}
}

// WithRollback executes the provided test function within a database transaction,
// and automatically rolls it back after execution to isolate test changes.
// It fails the test if starting the transaction fails.
func WithRollback(t *testing.T, fn func(ctx context.Context, tx domain.DBConn)) {
	ctx := context.Background()
	tx, err := testConn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			panic(err)
		}
	}(tx, ctx)
	fn(ctx, tx)
}
