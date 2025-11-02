package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/Gargair/clockwork/server/internal/service"
)

// Minimal fakes for end-to-end routing tests (no DB)

type e2eProjectService struct{ items map[uuid.UUID]domain.Project }

func (f *e2eProjectService) Create(_ context.Context, name string, description *string) (domain.Project, error) {
	now := time.Now().UTC()
	p := domain.Project{ID: uuid.New(), Name: name, Description: description, CreatedAt: now, UpdatedAt: now}
	if f.items == nil {
		f.items = make(map[uuid.UUID]domain.Project)
	}
	f.items[p.ID] = p
	return p, nil
}
func (f *e2eProjectService) Update(_ context.Context, id uuid.UUID, name string, description *string) (domain.Project, error) {
	p, ok := f.items[id]
	if !ok {
		return domain.Project{}, repository.ErrNotFound
	}
	p.Name, p.Description, p.UpdatedAt = name, description, time.Now().UTC()
	f.items[id] = p
	return p, nil
}
func (f *e2eProjectService) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := f.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.items, id)
	return nil
}
func (f *e2eProjectService) GetByID(_ context.Context, id uuid.UUID) (domain.Project, error) {
	p, ok := f.items[id]
	if !ok {
		return domain.Project{}, repository.ErrNotFound
	}
	return p, nil
}
func (f *e2eProjectService) List(_ context.Context) ([]domain.Project, error) {
	out := make([]domain.Project, 0, len(f.items))
	for _, p := range f.items {
		out = append(out, p)
	}
	return out, nil
}

var _ service.ProjectService = (*e2eProjectService)(nil)

type e2eCategoryService struct{}

func (e *e2eCategoryService) Create(context.Context, uuid.UUID, string, *string, *uuid.UUID) (domain.Category, error) {
	return domain.Category{}, nil
}
func (e *e2eCategoryService) Update(context.Context, uuid.UUID, string, *string, *uuid.UUID) (domain.Category, error) {
	return domain.Category{}, nil
}
func (e *e2eCategoryService) Delete(context.Context, uuid.UUID) error { return nil }
func (e *e2eCategoryService) GetByID(context.Context, uuid.UUID) (domain.Category, error) {
	return domain.Category{}, repository.ErrNotFound
}
func (e *e2eCategoryService) ListByProject(context.Context, uuid.UUID) ([]domain.Category, error) {
	return nil, nil
}
func (e *e2eCategoryService) ListChildren(context.Context, uuid.UUID) ([]domain.Category, error) {
	return nil, nil
}

var _ service.CategoryService = (*e2eCategoryService)(nil)

type e2eTimeService struct{}

func (e *e2eTimeService) Start(context.Context, uuid.UUID) (domain.TimeEntry, error) {
	return domain.TimeEntry{}, nil
}
func (e *e2eTimeService) StopActive(context.Context) (domain.TimeEntry, error) {
	return domain.TimeEntry{}, service.ErrNoActiveTimer
}
func (e *e2eTimeService) GetActive(context.Context) (*domain.TimeEntry, error) { return nil, nil }
func (e *e2eTimeService) ListByCategory(context.Context, uuid.UUID) ([]domain.TimeEntry, error) {
	return nil, nil
}
func (e *e2eTimeService) ListByCategoryAndRange(context.Context, uuid.UUID, time.Time, time.Time) ([]domain.TimeEntry, error) {
	return nil, nil
}

var _ service.TimeTrackingService = (*e2eTimeService)(nil)

func newTestMux(ps service.ProjectService, cs service.CategoryService, ts service.TimeTrackingService) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(loggingMiddleware(slog.Default()))

	api := chi.NewRouter()
	projH := NewProjectHandler(ps)
	catH := NewCategoryHandler(cs)
	timeH := NewTimeHandler(ts)

	projectsR := chi.NewRouter()
	projH.RegisterRoutes(projectsR)
	categoriesR := chi.NewRouter()
	catH.RegisterRoutes(categoriesR)
	projectsR.Mount("/{projectId}/categories", categoriesR)
	api.Mount("/projects", projectsR)

	timeR := chi.NewRouter()
	timeH.RegisterRoutes(timeR)
	api.Mount("/time", timeR)

	r.Mount("/api", api)
	return r
}

func TestAPIEndToEndProjectCreateAndList(t *testing.T) {
	ps := &e2eProjectService{}
	r := newTestMux(ps, &e2eCategoryService{}, &e2eTimeService{})

	// Create
	body := ProjectCreateRequest{Name: "My Project", Description: ptr("desc")}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPost, projectRoute, bytes.NewReader(data))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusCreated {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusCreated, w.Code)
	}
	var created ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if created.ID == uuid.Nil || created.Name != body.Name || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("unexpected project response: %+v", created)
	}

	// List
	req = httptest.NewRequest(stdhttp.MethodGet, projectRoute, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	var list []ProjectResponse
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if len(list) != 1 || list[0].ID != created.ID || list[0].Name != created.Name {
		t.Fatalf("unexpected list response: %+v", list)
	}
}

func TestAPIEndToEndTimeActiveReturnsNullAndStop409(t *testing.T) {
	r := newTestMux(&e2eProjectService{}, &e2eCategoryService{}, &e2eTimeService{})

	// Active should return null
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/time/active", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	if body := w.Body.String(); body != "null\n" && body != "null" {
		t.Fatalf("expected null body, got %q", body)
	}

	// Stop should return 409 with code
	req = httptest.NewRequest(stdhttp.MethodPost, "/api/time/stop", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusConflict {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusConflict, w.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if errResp.Code != string(codeNoActiveTimer) || errResp.RequestID == "" {
		t.Fatalf("unexpected error response: %+v", errResp)
	}
}
