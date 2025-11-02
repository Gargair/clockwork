//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Gargair/clockwork/server/internal/repository"
)

func TestCategoryRepositoryCreateSucceedsWithValidProjectIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)

	proj := NewProject("cat-proj", nil)
	p, err := pr.Create(ctx, proj)
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}

	cat := NewCategory(p.ID, "alpha", nil, nil)
	created, err := cr.Create(ctx, cat)
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "category", err)
	}
	if created.ProjectID != p.ID || created.Name != cat.Name {
		t.Fatalf("created mismatch: got %+v", created)
	}
}

func TestCategoryRepositoryUniqueCategoryNameWithinProjectEnforcedIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)

	p, err := pr.Create(ctx, NewProject("uniq-proj", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}

	c1 := NewCategory(p.ID, "dup-name", nil, nil)
	if _, err := cr.Create(ctx, c1); err != nil {
		t.Fatalf(CreateFailedErrorMessage, "first category", err)
	}

	c2 := NewCategory(p.ID, "dup-name", nil, nil)
	if _, err := cr.Create(ctx, c2); err == nil {
		t.Fatalf("expected duplicate error, got nil")
	} else if err != repository.ErrDuplicate {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestCategoryRepositoryParentChildAndDeleteParentSetsChildrenNullIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)

	p, err := pr.Create(ctx, NewProject("parent-child-proj", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}

	parent, err := cr.Create(ctx, NewCategory(p.ID, "parent", nil, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "parent category", err)
	}
	childA, err := cr.Create(ctx, NewCategory(p.ID, "child-a", &parent.ID, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "childA", err)
	}
	childB, err := cr.Create(ctx, NewCategory(p.ID, "child-b", &parent.ID, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "childB", err)
	}

	children, err := cr.ListChildren(ctx, parent.ID)
	if err != nil {
		t.Fatalf("ListChildren: %v", err)
	}
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if err := cr.Delete(ctx, parent.ID); err != nil {
		t.Fatalf("delete parent: %v", err)
	}

	aFetched, err := cr.GetByID(ctx, childA.ID)
	if err != nil {
		t.Fatalf("get childA: %v", err)
	}
	if aFetched.ParentCategoryID != nil {
		t.Fatalf("expected childA parent to be NULL, got %v", aFetched.ParentCategoryID)
	}
	bFetched, err := cr.GetByID(ctx, childB.ID)
	if err != nil {
		t.Fatalf("get childB: %v", err)
	}
	if bFetched.ParentCategoryID != nil {
		t.Fatalf("expected childB parent to be NULL, got %v", bFetched.ParentCategoryID)
	}
}

func TestCategoryRepositoryUpdateFieldsButNotProjectIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)

	p, err := pr.Create(ctx, NewProject("update-proj", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}

	c, err := cr.Create(ctx, NewCategory(p.ID, "orig", nil, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "category", err)
	}

	// Ensure updated_at can advance
	time.Sleep(5 * time.Millisecond)

	// create a new parent to attach to
	parent, err := cr.Create(ctx, NewCategory(p.ID, "new-parent", nil, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "new parent category", err)
	}

	newDesc := "desc-updated"
	updated, err := cr.Update(ctx, c.ID, "renamed", &newDesc, &parent.ID)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "renamed" {
		t.Fatalf("expected name change, got %s", updated.Name)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Fatalf("expected description change, got %v", updated.Description)
	}
	if updated.ParentCategoryID == nil || *updated.ParentCategoryID != parent.ID {
		t.Fatalf("expected parent change, got %v", updated.ParentCategoryID)
	}
	if updated.ProjectID != p.ID {
		t.Fatalf("project id should remain unchanged")
	}
}

func TestCategoryRepositoryListByProjectFiltersIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)

	p1, err := pr.Create(ctx, NewProject("p1", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "p1", err)
	}
	p2, err := pr.Create(ctx, NewProject("p2", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "p2", err)
	}

	if _, err := cr.Create(ctx, NewCategory(p1.ID, "a", nil, nil)); err != nil {
		t.Fatalf(CreateFailedErrorMessage, "cat a", err)
	}
	if _, err := cr.Create(ctx, NewCategory(p2.ID, "b", nil, nil)); err != nil {
		t.Fatalf(CreateFailedErrorMessage, "cat b", err)
	}
	if _, err := cr.Create(ctx, NewCategory(p1.ID, "c", nil, nil)); err != nil {
		t.Fatalf(CreateFailedErrorMessage, "cat c", err)
	}

	catsP1, err := cr.ListByProject(ctx, p1.ID)
	if err != nil {
		t.Fatalf("ListByProject p1: %v", err)
	}
	if len(catsP1) != 2 {
		t.Fatalf("expected 2 categories in p1, got %d", len(catsP1))
	}
	for _, c := range catsP1 {
		if c.ProjectID != p1.ID {
			t.Fatalf("unexpected project id in result: %v", c.ProjectID)
		}
	}
}
