package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testPool, err = pgxpool.New(ctx, os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	code := m.Run()

	os.Exit(code)
}
