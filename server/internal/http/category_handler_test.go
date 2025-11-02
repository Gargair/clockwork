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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/Gargair/clockwork/server/internal/service"
)

type fakeCategoryService struct {
	createFn        func(projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error)
	updateFn        func(id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error)
	deleteFn        func(id uuid.UUID) error
	getFn           func(id uuid.UUID) (domain.Category, error)
	listByProjectFn func(projectID uuid.UUID) ([]domain.Category, error)
	listChildrenFn  func(parentID uuid.UUID) ([]domain.Category, error)
}

func (f *fakeCategoryService) Create(_ context.Context, projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
	return f.createFn(projectID, name, description, parentCategoryID)
}
func (f *fakeCategoryService) Update(_ context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
	return f.updateFn(id, name, description, parentCategoryID)
}
func (f *fakeCategoryService) Delete(_ context.Context, id uuid.UUID) error { return f.deleteFn(id) }
func (f *fakeCategoryService) GetByID(_ context.Context, id uuid.UUID) (domain.Category, error) {
	return f.getFn(id)
}
func (f *fakeCategoryService) ListByProject(_ context.Context, projectID uuid.UUID) ([]domain.Category, error) {
	return f.listByProjectFn(projectID)
}
func (f *fakeCategoryService) ListChildren(_ context.Context, parentID uuid.UUID) ([]domain.Category, error) {
	return f.listChildrenFn(parentID)
}

var _ service.CategoryService = (*fakeCategoryService)(nil)

const categoriesRoute = "/api/projects/%s/categories"

func TestCategoryHandlerCreateWithValidParentSameProject(t *testing.T) {
	now := time.Now().UTC()
	f := &fakeCategoryService{
		createFn: func(projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{ID: uuid.New(), ProjectID: projectID, Name: name, Description: description, ParentCategoryID: parentCategoryID, CreatedAt: now, UpdatedAt: now}, nil
		},
		listByProjectFn: func(projectID uuid.UUID) ([]domain.Category, error) { return nil, nil },
		getFn:           func(id uuid.UUID) (domain.Category, error) { return domain.Category{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, nil
		},
		deleteFn:       func(id uuid.UUID) error { return nil },
		listChildrenFn: func(parentID uuid.UUID) ([]domain.Category, error) { return nil, nil },
	}
	h := NewCategoryHandler(f)

	r := chi.NewRouter()
	projectID := uuid.New()
	r.Route(sprintf(categoriesRoute, projectID.String()), h.RegisterRoutes)

	parentID := uuid.New().String()
	body := CategoryCreateRequest{Name: "Frontend", ParentCategoryID: &parentID}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(stdhttp.MethodPost, sprintf(categoriesRoute, projectID.String()), bytes.NewReader(data))
	// Inject path parameter for projectId since we're not using chi URL building here
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusCreated {
		t.Fatalf(statisCodeFailedExpectationMessage, stdhttp.StatusCreated, w.Code)
	}
}

func TestCategoryHandlerCreateCrossProjectParent(t *testing.T) {
	f := &fakeCategoryService{
		createFn: func(projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, service.ErrCrossProjectParent
		},
		listByProjectFn: func(projectID uuid.UUID) ([]domain.Category, error) { return nil, nil },
		getFn:           func(id uuid.UUID) (domain.Category, error) { return domain.Category{}, nil },
		updateFn: func(id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, nil
		},
		deleteFn:       func(id uuid.UUID) error { return nil },
		listChildrenFn: func(parentID uuid.UUID) ([]domain.Category, error) { return nil, nil },
	}
	h := NewCategoryHandler(f)

	r := chi.NewRouter()
	projectID := uuid.New()
	r.Route(sprintf(categoriesRoute, projectID.String()), h.RegisterRoutes)

	parentID := uuid.New().String()
	body := CategoryCreateRequest{Name: "Frontend", ParentCategoryID: &parentID}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(stdhttp.MethodPost, sprintf(categoriesRoute, projectID.String()), bytes.NewReader(data))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statisCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}
}

func TestCategoryHandlerUpdateCycle(t *testing.T) {
	f := &fakeCategoryService{
		updateFn: func(id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, service.ErrCategoryCycle
		},
		createFn: func(projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, nil
		},
		getFn:           func(id uuid.UUID) (domain.Category, error) { return domain.Category{}, nil },
		listByProjectFn: func(projectID uuid.UUID) ([]domain.Category, error) { return nil, nil },
		deleteFn:        func(id uuid.UUID) error { return nil },
		listChildrenFn:  func(parentID uuid.UUID) ([]domain.Category, error) { return nil, nil },
	}
	h := NewCategoryHandler(f)

	r := chi.NewRouter()
	projectID := uuid.New()
	r.Route(sprintf(categoriesRoute, projectID.String()), h.RegisterRoutes)

	id := uuid.New().String()
	body := CategoryUpdateRequest{Name: ptr("Frontend")}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPatch, sprintf(categoriesRoute, projectID.String())+"/"+id, bytes.NewReader(data))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", projectID.String())
	rctx.URLParams.Add("categoryId", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusConflict {
		t.Fatalf(statisCodeFailedExpectationMessage, stdhttp.StatusConflict, w.Code)
	}
}

func TestCategoryHandlerGetNotFound(t *testing.T) {
	f := &fakeCategoryService{
		getFn:           func(id uuid.UUID) (domain.Category, error) { return domain.Category{}, repository.ErrNotFound },
		listByProjectFn: func(projectID uuid.UUID) ([]domain.Category, error) { return nil, nil },
		createFn: func(projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, nil
		},
		updateFn: func(id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
			return domain.Category{}, nil
		},
		deleteFn:       func(id uuid.UUID) error { return nil },
		listChildrenFn: func(parentID uuid.UUID) ([]domain.Category, error) { return nil, nil },
	}
	h := NewCategoryHandler(f)

	r := chi.NewRouter()
	projectID := uuid.New()
	r.Route(sprintf(categoriesRoute, projectID.String()), h.RegisterRoutes)

	id := uuid.New().String()
	req := httptest.NewRequest(stdhttp.MethodGet, sprintf(categoriesRoute, projectID.String())+"/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", projectID.String())
	rctx.URLParams.Add("categoryId", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusNotFound {
		t.Fatalf(statisCodeFailedExpectationMessage, stdhttp.StatusNotFound, w.Code)
	}
}

func sprintf(format string, a ...any) string { return fmt.Sprintf(format, a...) }
