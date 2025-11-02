package http

import (
	"bytes"
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/Gargair/clockwork/server/internal/service"
)

type fakeProjectService struct {
	createFn func(name string, description *string) (domain.Project, error)
	updateFn func(id uuid.UUID, name string, description *string) (domain.Project, error)
	deleteFn func(id uuid.UUID) error
	getFn    func(id uuid.UUID) (domain.Project, error)
	listFn   func() ([]domain.Project, error)
}

func (f *fakeProjectService) Create(_ context.Context, name string, description *string) (domain.Project, error) {
	return f.createFn(name, description)
}
func (f *fakeProjectService) Update(_ context.Context, id uuid.UUID, name string, description *string) (domain.Project, error) {
	return f.updateFn(id, name, description)
}
func (f *fakeProjectService) Delete(_ context.Context, id uuid.UUID) error { return f.deleteFn(id) }
func (f *fakeProjectService) GetByID(_ context.Context, id uuid.UUID) (domain.Project, error) {
	return f.getFn(id)
}
func (f *fakeProjectService) List(_ context.Context) ([]domain.Project, error) { return f.listFn() }

// Ensure interface compliance
var _ service.ProjectService = (*fakeProjectService)(nil)

// --- Tests ---

const projectRoute = "/api/projects"

func TestProjectHandlerCreateHappyPath(t *testing.T) {
	f := &fakeProjectService{
		createFn: func(name string, description *string) (domain.Project, error) {
			now := time.Now().UTC()
			return domain.Project{ID: uuid.New(), Name: name, Description: description, CreatedAt: now, UpdatedAt: now}, nil
		},
		listFn: func() ([]domain.Project, error) { return nil, nil },
		getFn:  func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)

	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	body := ProjectCreateRequest{Name: "Proj A"}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(stdhttp.MethodPost, projectRoute, bytes.NewReader(data))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != stdhttp.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Name != body.Name || resp.ID == uuid.Nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestProjectHandlerListHappyPath(t *testing.T) {
	now := time.Now().UTC()
	f := &fakeProjectService{
		listFn: func() ([]domain.Project, error) {
			return []domain.Project{{ID: uuid.New(), Name: "A", CreatedAt: now, UpdatedAt: now}}, nil
		},
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		getFn: func(id uuid.UUID) (domain.Project, error) {
			return domain.Project{}, nil
		},
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodGet, projectRoute, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	var resp []ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp) != 1 || resp[0].Name != "A" {
		t.Fatalf("unexpected list response: %+v", resp)
	}
}

func TestProjectHandlerGetByIDInvalidUUID(t *testing.T) {
	f := &fakeProjectService{
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/projects/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
}

func TestProjectHandlerGetByIDNotFound(t *testing.T) {
	f := &fakeProjectService{
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, repository.ErrNotFound },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	req := httptest.NewRequest(stdhttp.MethodGet, projectRoute+"/"+id, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusNotFound {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusNotFound, w.Code)
	}
}

func TestProjectHandlerUpdateHappyPath(t *testing.T) {
	now := time.Now().UTC()
	f := &fakeProjectService{
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{ID: id, Name: name, Description: description, CreatedAt: now, UpdatedAt: now}, nil
		},
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	body := ProjectUpdateRequest{Name: ptr("Updated")}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPatch, projectRoute+"/"+id, bytes.NewReader(data))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
}

func TestProjectHandlerUpdateInvalidUUID(t *testing.T) {
	f := &fakeProjectService{
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	body := ProjectUpdateRequest{Name: ptr("Updated")}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPatch, projectRoute+"/not-a-uuid", bytes.NewReader(data))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
}

func TestProjectHandlerUpdateEmptyName(t *testing.T) {
	f := &fakeProjectService{
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, service.ErrInvalidProjectName
		},
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	body := ProjectUpdateRequest{Name: ptr("   ")}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPatch, projectRoute+"/"+id, bytes.NewReader(data))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
}

func TestProjectHandlerDeleteHappyPath(t *testing.T) {
	f := &fakeProjectService{
		deleteFn: func(id uuid.UUID) error { return nil },
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
	}
	h := NewProjectHandler(f)
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	req := httptest.NewRequest(stdhttp.MethodDelete, projectRoute+"/"+id, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusNoContent {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusNoContent, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("expected empty body on 204, got %q", w.Body.String())
	}
}

// helpers
func ptr[T any](v T) *T { return &v }
