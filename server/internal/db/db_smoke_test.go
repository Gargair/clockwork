//go:build integration
// +build integration

package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Gargair/clockwork/server/internal/config"
	"github.com/google/uuid"
)

func TestDBRoundTripIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	conn, err := Open(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	id := uuid.New()
	name := "smoke_test_project_" + id.String()[0:8]
	desc := "integration smoke test"

	_, err = conn.ExecContext(ctx,
		`INSERT INTO project (id, name, description) VALUES ($1, $2, $3)`,
		id, name, desc,
	)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), `DELETE FROM project WHERE id = $1`, id)
	}()

	var gotID uuid.UUID
	var gotName string
	var gotDesc sql.NullString
	var createdAt time.Time
	var updatedAt time.Time

	err = conn.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM project WHERE id = $1`, id,
	).Scan(&gotID, &gotName, &gotDesc, &createdAt, &updatedAt)
	if err != nil {
		t.Fatalf("select project: %v", err)
	}
	if gotID != id {
		t.Fatalf("id mismatch: got %s want %s", gotID, id)
	}
	if gotName != name {
		t.Fatalf("name mismatch: got %s want %s", gotName, name)
	}
	if !gotDesc.Valid || gotDesc.String != desc {
		t.Fatalf("description mismatch: got %v want %s", gotDesc, desc)
	}
	if createdAt.IsZero() || updatedAt.IsZero() {
		t.Fatalf("timestamps should not be zero")
	}
}
