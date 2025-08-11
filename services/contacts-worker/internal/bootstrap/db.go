package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewSQLDatabase creates and returns a new pgxpool.Pool connection pool
// using the given configuration. It establishes a connection to the PostgreSQL
// database with a timeout and verifies connectivity by pinging the database.
// If any step fails, it terminates the program.
func NewSQLDatabase(config *DBConfig) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	log.Printf("db is running on %v\n", connString)

	return conn
}

// CloseDBConnection safely closes the given pgxpool.Pool database connection pool.
func CloseDBConnection(pool *pgxpool.Pool) {
	pool.Close()
}
