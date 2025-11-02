package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestProjectServiceRejectsEmptyNameOnCreate(t *testing.T) {
	svc := NewProjectService(newFakeProjectRepo())
	if _, err := svc.Create(context.Background(), "   ", nil); err == nil {
		t.Fatalf("expected error for empty name, got nil")
	}
}

func TestProjectServiceCreateAndListGetByID(t *testing.T) {
	repo := newFakeProjectRepo()
	svc := NewProjectService(repo)

	desc := "test"
	created, err := svc.Create(context.Background(), " Project A ", &desc)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatalf("expected non-nil ID")
	}
	if created.Name != "Project A" {
		t.Fatalf("expected trimmed name 'Project A', got %q", created.Name)
	}
	if created.Description == nil || *created.Description != "test" {
		t.Fatalf("expected description 'test', got %v", created.Description)
	}

	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", got, created)
	}

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
}
