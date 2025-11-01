//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	repo "github.com/Gargair/clockwork/server/internal/repository/postgres"
)

const CreateFailedErrorMessage = "Create failed: %v"

func TestProjectRepositoryCreateAndGetByIDIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	r := repo.NewProjectRepository(db)
	ctx := context.Background()

	desc := "first project"
	toCreate := NewProject("proj-a", &desc)

	created, err := r.Create(ctx, toCreate)
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, err)
	}
	if created.ID != toCreate.ID || created.Name != toCreate.Name {
		t.Fatalf("Created mismatch: got %+v", created)
	}

	fetched, err := r.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.ID != created.ID || fetched.Name != created.Name {
		t.Fatalf("Fetched mismatch: got %+v want %+v", fetched, created)
	}
}

func TestProjectRepositoryListIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	r := repo.NewProjectRepository(db)
	ctx := context.Background()

	p1 := NewProject("proj-1", nil)
	p2 := NewProject("proj-2", nil)
	if _, err := r.Create(ctx, p1); err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if _, err := r.Create(ctx, p2); err != nil {
		t.Fatalf("create p2: %v", err)
	}

	list, err := r.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(list))
	}
	ids := map[string]bool{p1.ID.String(): false, p2.ID.String(): false}
	for _, p := range list {
		if _, ok := ids[p.ID.String()]; ok {
			ids[p.ID.String()] = true
		}
	}
	for id, seen := range ids {
		if !seen {
			t.Fatalf("expected project id %s in list", id)
		}
	}
}

func TestProjectRepositoryUpdateIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	r := repo.NewProjectRepository(db)
	ctx := context.Background()

	orig := NewProject("proj-update", nil)
	created, err := r.Create(ctx, orig)
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, err)
	}

	// Small sleep to ensure updated_at can advance
	time.Sleep(5 * time.Millisecond)

	newDesc := "updated desc"
	updated, err := r.Update(ctx, created.ID, "proj-updated", &newDesc)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "proj-updated" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Fatalf("expected updated description, got %v", updated.Description)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("expected updated_at to advance: before=%v after=%v", created.UpdatedAt, updated.UpdatedAt)
	}
}

func TestProjectRepositoryDeleteIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	r := repo.NewProjectRepository(db)
	ctx := context.Background()

	p := NewProject("proj-delete", nil)
	created, err := r.Create(ctx, p)
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, err)
	}
	if err := r.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := r.GetByID(ctx, created.ID); err == nil {
		t.Fatalf("expected ErrNotFound after delete")
	}
}
