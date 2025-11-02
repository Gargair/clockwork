//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Gargair/clockwork/server/internal/db"
	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/google/uuid"
)

const CreateFailedErrorMessage = "Create %s failed: %v"

// OpenDBFromEnv opens a Postgres connection using DATABASE_URL.
// Skips the test if DATABASE_URL is not set.
func OpenDBFromEnv(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	return conn
}

// TruncateAll deletes rows from all tables in FK-safe order.
func TruncateAll(t *testing.T, conn *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Delete in order to satisfy FK constraints
	if _, err := conn.ExecContext(ctx, "DELETE FROM time_entry"); err != nil {
		t.Fatalf("failed to delete from time_entry: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "DELETE FROM category"); err != nil {
		t.Fatalf("failed to delete from category: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "DELETE FROM project"); err != nil {
		t.Fatalf("failed to delete from project: %v", err)
	}
}

// NewProject creates a minimal domain.Project ready for insertion.
func NewProject(name string, description *string) domain.Project {
	if name == "" {
		name = "project-" + uuid.New().String()
	}
	return domain.Project{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
	}
}

// NewCategory creates a minimal domain.Category ready for insertion.
func NewCategory(projectID uuid.UUID, name string, parentCategoryID *uuid.UUID, description *string) domain.Category {
	if name == "" {
		name = "category-" + uuid.New().String()
	}
	return domain.Category{
		ID:               uuid.New(),
		ProjectID:        projectID,
		ParentCategoryID: parentCategoryID,
		Name:             name,
		Description:      description,
	}
}

// NewTimeEntry creates an active time entry (no stopped_at, no duration seconds).
func NewTimeEntry(categoryID uuid.UUID, startedAt time.Time) domain.TimeEntry {
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	return domain.TimeEntry{
		ID:         uuid.New(),
		CategoryID: categoryID,
		StartedAt:  startedAt,
	}
}
