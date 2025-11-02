package service

import (
    "context"
    "testing"

    "github.com/Gargair/clockwork/server/internal/domain"
    "github.com/Gargair/clockwork/server/internal/repository"
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

// --- Error propagation and edge cases ---

type stubCategoryRepo struct {
    items            map[uuid.UUID]domain.Category
    errGetByID       map[uuid.UUID]error
    createErr        error
    updateErr        error
    deleteErr        error
    listByProjectErr error
    listChildrenErr  error
}

func (r stubCategoryRepo) Create(context.Context, domain.Category) (domain.Category, error) {
    return domain.Category{}, r.createErr
}
func (r stubCategoryRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Category, error) {
    if r.errGetByID != nil {
        if e, ok := r.errGetByID[id]; ok {
            return domain.Category{}, e
        }
    }
    if r.items != nil {
        if c, ok := r.items[id]; ok {
            return c, nil
        }
    }
    return domain.Category{}, repository.ErrNotFound
}
func (r stubCategoryRepo) ListByProject(context.Context, uuid.UUID) ([]domain.Category, error) {
    if r.listByProjectErr != nil {
        return nil, r.listByProjectErr
    }
    return nil, nil
}
func (r stubCategoryRepo) ListChildren(context.Context, uuid.UUID) ([]domain.Category, error) {
    if r.listChildrenErr != nil {
        return nil, r.listChildrenErr
    }
    return nil, nil
}
func (r stubCategoryRepo) Update(context.Context, uuid.UUID, string, *string, *uuid.UUID) (domain.Category, error) {
    return domain.Category{}, r.updateErr
}
func (r stubCategoryRepo) Delete(context.Context, uuid.UUID) error { return r.deleteErr }

func TestCategoryServiceCreateInvalidParentWhenMissing(t *testing.T) {
    missingParentID := uuid.New()
    svc := NewCategoryService(stubCategoryRepo{errGetByID: map[uuid.UUID]error{missingParentID: repository.ErrNotFound}})
    if _, err := svc.Create(context.Background(), uuid.New(), "Child", nil, &missingParentID); err == nil || err != ErrInvalidParent {
        t.Fatalf("expected ErrInvalidParent, got %v", err)
    }
}

func TestCategoryServiceCreatePropagatesParentLookupError(t *testing.T) {
    parentID := uuid.New()
    svc := NewCategoryService(stubCategoryRepo{errGetByID: map[uuid.UUID]error{parentID: repository.ErrDuplicate}})
    if _, err := svc.Create(context.Background(), uuid.New(), "Child", nil, &parentID); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected parent lookup error, got %v", err)
    }
}

func TestCategoryServiceCreatePropagatesCreateError(t *testing.T) {
    svc := NewCategoryService(stubCategoryRepo{createErr: repository.ErrDuplicate})
    if _, err := svc.Create(context.Background(), uuid.New(), "Child", nil, nil); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected createErr, got %v", err)
    }
}

func TestCategoryServiceUpdatePropagatesGetCurrentError(t *testing.T) {
    id := uuid.New()
    svc := NewCategoryService(stubCategoryRepo{errGetByID: map[uuid.UUID]error{id: repository.ErrNotFound}})
    if _, err := svc.Update(context.Background(), id, "X", nil, nil); err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected ErrNotFound, got %v", err)
    }
}

func TestCategoryServiceUpdateInvalidParentWhenMissing(t *testing.T) {
    id := uuid.New()
    parentID := uuid.New()
    items := map[uuid.UUID]domain.Category{id: {ID: id, ProjectID: uuid.New(), Name: "cur"}}
    svc := NewCategoryService(stubCategoryRepo{items: items, errGetByID: map[uuid.UUID]error{parentID: repository.ErrNotFound}})
    if _, err := svc.Update(context.Background(), id, "X", nil, &parentID); err == nil || err != ErrInvalidParent {
        t.Fatalf("expected ErrInvalidParent, got %v", err)
    }
}

func TestCategoryServiceUpdatePropagatesParentLookupError(t *testing.T) {
    id := uuid.New()
    parentID := uuid.New()
    proj := uuid.New()
    items := map[uuid.UUID]domain.Category{id: {ID: id, ProjectID: proj, Name: "cur"}}
    svc := NewCategoryService(stubCategoryRepo{items: items, errGetByID: map[uuid.UUID]error{parentID: repository.ErrDuplicate}})
    if _, err := svc.Update(context.Background(), id, "X", nil, &parentID); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected parent lookup error, got %v", err)
    }
}

func TestCategoryServiceUpdatePropagatesListChildrenError(t *testing.T) {
    id := uuid.New()
    parentID := uuid.New()
    proj := uuid.New()
    // Set parent to same project, not self
    items := map[uuid.UUID]domain.Category{id: {ID: id, ProjectID: proj, Name: "cur"}, parentID: {ID: parentID, ProjectID: proj, Name: "par"}}
    svc := NewCategoryService(stubCategoryRepo{items: items, listChildrenErr: repository.ErrDuplicate})
    if _, err := svc.Update(context.Background(), id, "X", nil, &parentID); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected listChildren error, got %v", err)
    }
}

func TestCategoryServiceUpdatePropagatesUpdateError(t *testing.T) {
    id := uuid.New()
    proj := uuid.New()
    items := map[uuid.UUID]domain.Category{id: {ID: id, ProjectID: proj, Name: "cur"}}
    svc := NewCategoryService(stubCategoryRepo{items: items, updateErr: repository.ErrDuplicate})
    if _, err := svc.Update(context.Background(), id, "X", nil, nil); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected updateErr, got %v", err)
    }
}

func TestCategoryServiceDeletePropagatesError(t *testing.T) {
    svc := NewCategoryService(stubCategoryRepo{deleteErr: repository.ErrNotFound})
    if err := svc.Delete(context.Background(), uuid.New()); err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected deleteErr, got %v", err)
    }
}

func TestCategoryServiceGetByIDPropagatesError(t *testing.T) {
    svc := NewCategoryService(stubCategoryRepo{errGetByID: map[uuid.UUID]error{uuid.Nil: repository.ErrNotFound}})
    if _, err := svc.GetByID(context.Background(), uuid.Nil); err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected GetByID error, got %v", err)
    }
}

func TestCategoryServiceListByProjectPropagatesError(t *testing.T) {
    svc := NewCategoryService(stubCategoryRepo{listByProjectErr: repository.ErrDuplicate})
    if _, err := svc.ListByProject(context.Background(), uuid.New()); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected ListByProject error, got %v", err)
    }
}

func TestCategoryServiceListChildrenPropagatesError(t *testing.T) {
    svc := NewCategoryService(stubCategoryRepo{listChildrenErr: repository.ErrDuplicate})
    if _, err := svc.ListChildren(context.Background(), uuid.New()); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected ListChildren error, got %v", err)
    }
}
