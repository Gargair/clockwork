package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open creates a database/sql connection pool for Postgres using the pgx driver.
// It configures conservative defaults suitable for local development and verifies connectivity.
func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	// TODO: Pool settings tuned for dev; adjust later for prod if needed
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	// Verify connectivity
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

// Health pings the database to confirm connectivity.
func Health(ctx context.Context, conn *sql.DB) error {
	if conn == nil {
		return sql.ErrConnDone
	}
	return conn.PingContext(ctx)
}
