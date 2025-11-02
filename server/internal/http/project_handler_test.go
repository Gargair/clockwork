package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"

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
	h := NewProjectHandler(f, slog.Default())

	r := mountRoutes(projectRoute, h.RegisterRoutes)

	body := ProjectCreateRequest{Name: "Proj A"}
	data := mustJSON(t, body)
	w := doRequest(r, stdhttp.MethodPost, projectRoute, data, nil)

	if w.Code != stdhttp.StatusCreated {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusCreated, w.Code)
	}
	var resp ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
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
	h := NewProjectHandler(f, slog.Default())
	r := mountRoutes(projectRoute, h.RegisterRoutes)

	w := doRequest(r, stdhttp.MethodGet, projectRoute, nil, nil)

	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	var resp []ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
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
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodGet, projectRoute+"/"+invalidId, nil)
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
	h := NewProjectHandler(f, slog.Default())
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

func TestProjectHandlerGetByIDHappyPath(t *testing.T) {
	now := time.Now().UTC()
	f := &fakeProjectService{
		getFn: func(id uuid.UUID) (domain.Project, error) {
			return domain.Project{ID: id, Name: "proj", CreatedAt: now, UpdatedAt: now}, nil
		},
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	req := httptest.NewRequest(stdhttp.MethodGet, fmt.Sprintf("%s/%s", projectRoute, id), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}
	var resp ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if resp.ID.String() != id || resp.Name != "proj" {
		t.Fatalf("unexpected body: %+v", resp)
	}
}

func TestProjectHandlerCreateInvalidNameMapsErrorAndIncludesRequestID(t *testing.T) {
	f := &fakeProjectService{
		createFn: func(name string, description *string) (domain.Project, error) {
			return domain.Project{}, service.ErrInvalidProjectName
		},
		listFn: func() ([]domain.Project, error) { return nil, nil },
		getFn:  func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route(projectRoute, h.RegisterRoutes)

	body := []byte(`{"name":"   "}`)
	req := httptest.NewRequest(stdhttp.MethodPost, projectRoute, bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if errResp.Code != string(codeInvalidProjectName) || errResp.RequestID == "" {
		t.Fatalf("unexpected error response: %+v", errResp)
	}
}

func TestProjectHandlerCreateUnknownFieldIsInvalidJSON(t *testing.T) {
	f := &fakeProjectService{
		createFn: func(name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		listFn: func() ([]domain.Project, error) { return nil, nil },
		getFn:  func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route(projectRoute, h.RegisterRoutes)

	body := []byte(`{"name":"proj","bogus":1}`)
	req := httptest.NewRequest(stdhttp.MethodPost, projectRoute, bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if errResp.Code != string(codeInvalidJSON) || errResp.RequestID == "" {
		t.Fatalf("unexpected error response: %+v", errResp)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
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
	h := NewProjectHandler(f, slog.Default())
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
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(projectRoute, h.RegisterRoutes)

	body := ProjectUpdateRequest{Name: ptr("Updated")}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPatch, projectRoute+"/"+invalidId, bytes.NewReader(data))
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
	h := NewProjectHandler(f, slog.Default())
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
	h := NewProjectHandler(f, slog.Default())
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

func TestProjectHandlerUpdateInvalidJSON(t *testing.T) {
	f := &fakeProjectService{
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	body := []byte(`{"name":"Proj","bogus":true}`)
	w := doRequest(r, stdhttp.MethodPatch, projectRoute+"/"+id, body, nil)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if errResp.Code != string(codeInvalidJSON) || errResp.RequestID == "" {
		t.Fatalf("unexpected error response: %+v", errResp)
	}
}

func TestProjectHandlerDeleteInvalidUUID(t *testing.T) {
	f := &fakeProjectService{
		deleteFn: func(id uuid.UUID) error { return nil },
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
	}
	h := NewProjectHandler(f, slog.Default())
	r := mountRoutes(projectRoute, h.RegisterRoutes)
	w := doRequest(r, stdhttp.MethodDelete, projectRoute+"/not-a-uuid", nil, nil)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
}

func TestProjectHandlerDeleteNotFoundMaps404(t *testing.T) {
	f := &fakeProjectService{
		deleteFn: func(id uuid.UUID) error { return repository.ErrNotFound },
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		listFn:   func() ([]domain.Project, error) { return nil, nil },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
	}
	h := NewProjectHandler(f, slog.Default())
	r := mountRoutes(projectRoute, h.RegisterRoutes)

	id := uuid.New().String()
	w := doRequest(r, stdhttp.MethodDelete, projectRoute+"/"+id, nil, nil)
	if w.Code != stdhttp.StatusNotFound {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusNotFound, w.Code)
	}
}

func TestProjectHandlerCreateUnknownRepoErrorMaps500Internal(t *testing.T) {
	f := &fakeProjectService{
		createFn: func(name string, description *string) (domain.Project, error) {
			return domain.Project{}, repository.ErrDuplicate
		},
		listFn: func() ([]domain.Project, error) { return nil, nil },
		getFn:  func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route(projectRoute, h.RegisterRoutes)

	body := mustJSON(t, ProjectCreateRequest{Name: "X"})
	w := doRequest(r, stdhttp.MethodPost, projectRoute, body, nil)
	if w.Code != stdhttp.StatusInternalServerError {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusInternalServerError, w.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if errResp.Code != string(codeInternal) || errResp.RequestID == "" {
		t.Fatalf("unexpected error response: %+v", errResp)
	}
}

func TestProjectHandlerListUnknownErrorMaps500(t *testing.T) {
	f := &fakeProjectService{
		listFn:   func() ([]domain.Project, error) { return nil, repository.ErrDuplicate },
		createFn: func(name string, description *string) (domain.Project, error) { return domain.Project{}, nil },
		getFn:    func(id uuid.UUID) (domain.Project, error) { return domain.Project{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string) (domain.Project, error) {
			return domain.Project{}, nil
		},
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	h := NewProjectHandler(f, slog.Default())
	r := mountRoutes(projectRoute, h.RegisterRoutes)
	w := doRequest(r, stdhttp.MethodGet, projectRoute, nil, nil)
	if w.Code != stdhttp.StatusInternalServerError {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusInternalServerError, w.Code)
	}
}

// helpers
func ptr[T any](v T) *T { return &v }
