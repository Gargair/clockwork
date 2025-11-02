package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCategoryServiceCreateWithParentSameProjectSucceeds(t *testing.T) {
	repo := newFakeCategoryRepo()
	svc := NewCategoryService(repo)
	ctx := context.Background()

	projectID := uuid.New()

	parent, err := svc.Create(ctx, projectID, "Parent", nil, nil)
	if err != nil {
		t.Fatalf("create parent failed: %v", err)
	}

	child, err := svc.Create(ctx, projectID, "Child", nil, &parent.ID)
	if err != nil {
		t.Fatalf("create child failed: %v", err)
	}
	if child.ParentCategoryID == nil || *child.ParentCategoryID != parent.ID {
		t.Fatalf("expected parent id to be set")
	}
}

func TestCategoryServiceCreateCrossProjectParentErr(t *testing.T) {
	repo := newFakeCategoryRepo()
	svc := NewCategoryService(repo)
	ctx := context.Background()

	projectA := uuid.New()
	projectB := uuid.New()

	parent, err := svc.Create(ctx, projectA, "Parent", nil, nil)
	if err != nil {
		t.Fatalf("create parent failed: %v", err)
	}

	if _, err := svc.Create(ctx, projectB, "Child", nil, &parent.ID); err == nil {
		t.Fatalf("expected cross-project error, got nil")
	} else if err != ErrCrossProjectParent {
		t.Fatalf("expected ErrCrossProjectParent, got %v", err)
	}
}

func TestCategoryServiceUpdateParentToDescendantErrCycle(t *testing.T) {
	repo := newFakeCategoryRepo()
	svc := NewCategoryService(repo)
	ctx := context.Background()

	proj := uuid.New()
	a, err := svc.Create(ctx, proj, "A", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.Create(ctx, proj, "B", nil, &a.ID)
	if err != nil {
		t.Fatal(err)
	}
	c, err := svc.Create(ctx, proj, "C", nil, &b.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Update(ctx, a.ID, a.Name, a.Description, &c.ID); err == nil {
		t.Fatalf("expected cycle error, got nil")
	} else if err != ErrCategoryCycle {
		t.Fatalf("expected ErrCategoryCycle, got %v", err)
	}
}

func TestCategoryServiceUpdateNameDescriptionOnlySucceeds(t *testing.T) {
	repo := newFakeCategoryRepo()
	svc := NewCategoryService(repo)
	ctx := context.Background()
	proj := uuid.New()

	a, err := svc.Create(ctx, proj, "A", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	newDesc := "desc"
	updated, err := svc.Update(ctx, a.ID, "A2", &newDesc, a.ParentCategoryID)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "A2" || updated.Description == nil || *updated.Description != "desc" {
		t.Fatalf("unexpected updated values: %+v", updated)
	}
}
