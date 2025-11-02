package service

import (
    "context"
    "testing"

    "github.com/Gargair/clockwork/server/internal/domain"
    "github.com/Gargair/clockwork/server/internal/repository"
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

// --- Error handling and propagation ---

type stubProjectRepo struct {
    createErr error
    updateErr error
    deleteErr error
    getErr    error
    listErr   error
}

func (r stubProjectRepo) Create(context.Context, domain.Project) (domain.Project, error) {
    return domain.Project{}, r.createErr
}
func (r stubProjectRepo) GetByID(context.Context, uuid.UUID) (domain.Project, error) {
    return domain.Project{}, r.getErr
}
func (r stubProjectRepo) List(context.Context) ([]domain.Project, error) { return nil, r.listErr }
func (r stubProjectRepo) Update(context.Context, uuid.UUID, string, *string) (domain.Project, error) {
    return domain.Project{}, r.updateErr
}
func (r stubProjectRepo) Delete(context.Context, uuid.UUID) error { return r.deleteErr }

func TestProjectServiceUpdateRejectsEmptyName(t *testing.T) {
    svc := NewProjectService(newFakeProjectRepo())
    if _, err := svc.Update(context.Background(), uuid.New(), "   ", nil); err == nil || err != ErrInvalidProjectName {
        t.Fatalf("expected ErrInvalidProjectName, got %v", err)
    }
}

func TestProjectServiceCreatePropagatesRepoError(t *testing.T) {
    svc := NewProjectService(stubProjectRepo{createErr: repository.ErrDuplicate})
    _, err := svc.Create(context.Background(), "Valid Name", nil)
    if err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected ErrDuplicate, got %v", err)
    }
}

func TestProjectServiceUpdatePropagatesRepoError(t *testing.T) {
    svc := NewProjectService(stubProjectRepo{updateErr: repository.ErrNotFound})
    _, err := svc.Update(context.Background(), uuid.New(), "Valid Name", nil)
    if err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected ErrNotFound, got %v", err)
    }
}

func TestProjectServiceDeletePropagatesRepoError(t *testing.T) {
    svc := NewProjectService(stubProjectRepo{deleteErr: repository.ErrNotFound})
    if err := svc.Delete(context.Background(), uuid.New()); err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected ErrNotFound, got %v", err)
    }
}

func TestProjectServiceGetByIDPropagatesRepoError(t *testing.T) {
    svc := NewProjectService(stubProjectRepo{getErr: repository.ErrNotFound})
    if _, err := svc.GetByID(context.Background(), uuid.New()); err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected ErrNotFound, got %v", err)
    }
}

func TestProjectServiceListPropagatesRepoError(t *testing.T) {
    svc := NewProjectService(stubProjectRepo{listErr: repository.ErrDuplicate})
    if _, err := svc.List(context.Background()); err == nil || err != repository.ErrDuplicate {
        t.Fatalf("expected listErr, got %v", err)
    }
}
