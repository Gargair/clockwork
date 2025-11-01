package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations opens a DB connection and applies all up migrations in the given directory.
// It returns the version before and after the operation.
func RunMigrations(ctx context.Context, databaseURL string, migrationsDir string) (int64, int64, error) {
	conn, err := waitForDB(ctx, databaseURL)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = conn.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, 0, err
	}

	fromVersion, err := goose.GetDBVersion(conn)
	if err != nil {
		return 0, 0, err
	}

	if err := goose.Up(conn, migrationsDir); err != nil {
		return fromVersion, 0, err
	}

	toVersion, err := goose.GetDBVersion(conn)
	if err != nil {
		return fromVersion, 0, err
	}
	return fromVersion, toVersion, nil
}

// waitForDB tries to open and ping the database with small backoff until ready or context ends.
func waitForDB(ctx context.Context, databaseURL string) (*sql.DB, error) {
	backoffs := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 3 * time.Second, 5 * time.Second}
	var conn *sql.DB
	var err error
	try := 0
	for {
		conn, err = sql.Open("pgx", databaseURL)
		if err == nil {
			pingErr := conn.PingContext(ctx)
			if pingErr == nil {
				return conn, nil
			}
			err = pingErr
			_ = conn.Close()
		}
		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil, ctx.Err()
			}
			return nil, err
		}
		d := backoffs[min(try, len(backoffs)-1)]
		time.Sleep(d)
		try++
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
