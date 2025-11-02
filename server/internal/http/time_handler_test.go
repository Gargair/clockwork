package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/service"
)

type fakeTimeService struct {
	startFn                  func(categoryID uuid.UUID) (domain.TimeEntry, error)
	stopActiveFn             func() (domain.TimeEntry, error)
	getActiveFn              func() (*domain.TimeEntry, error)
	listByCategoryFn         func(categoryID uuid.UUID) ([]domain.TimeEntry, error)
	listByCategoryAndRangeFn func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error)
}

func (f *fakeTimeService) Start(_ context.Context, categoryID uuid.UUID) (domain.TimeEntry, error) {
	return f.startFn(categoryID)
}
func (f *fakeTimeService) StopActive(_ context.Context) (domain.TimeEntry, error) {
	return f.stopActiveFn()
}
func (f *fakeTimeService) GetActive(_ context.Context) (*domain.TimeEntry, error) {
	return f.getActiveFn()
}
func (f *fakeTimeService) ListByCategory(_ context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error) {
	return f.listByCategoryFn(categoryID)
}
func (f *fakeTimeService) ListByCategoryAndRange(_ context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
	return f.listByCategoryAndRangeFn(categoryID, start, end)
}

var _ service.TimeTrackingService = (*fakeTimeService)(nil)

const timeRoute = "/api/time"
const categoryEntriesRoute = timeRoute + "/entries?categoryId="
const invalidJsonErrorMessage = "invalid JSON: %v"

func TestTimeHandlerStartCreated(t *testing.T) {
	now := time.Now().UTC()
	f := &fakeTimeService{
		startFn: func(categoryID uuid.UUID) (domain.TimeEntry, error) {
			return domain.TimeEntry{ID: uuid.New(), CategoryID: categoryID, StartedAt: now, CreatedAt: now, UpdatedAt: now}, nil
		},
		stopActiveFn:     func() (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		getActiveFn:      func() (*domain.TimeEntry, error) { return nil, nil },
		listByCategoryFn: func(categoryID uuid.UUID) ([]domain.TimeEntry, error) { return nil, nil },
		listByCategoryAndRangeFn: func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
			return nil, nil
		},
	}
	h := NewTimeHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(timeRoute, h.RegisterRoutes)

	body := TimeStartRequest{CategoryID: uuid.New().String()}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(stdhttp.MethodPost, timeRoute+"/start", bytes.NewReader(data))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusCreated {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusCreated, w.Code)
	}
	var resp TimeEntryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if resp.CategoryID.String() != body.CategoryID {
		t.Fatalf("expected category echo, got %s", resp.CategoryID)
	}
}

func TestTimeHandlerStopNoActive(t *testing.T) {
	f := &fakeTimeService{
		stopActiveFn:     func() (domain.TimeEntry, error) { return domain.TimeEntry{}, service.ErrNoActiveTimer },
		startFn:          func(categoryID uuid.UUID) (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		getActiveFn:      func() (*domain.TimeEntry, error) { return nil, nil },
		listByCategoryFn: func(categoryID uuid.UUID) ([]domain.TimeEntry, error) { return nil, nil },
		listByCategoryAndRangeFn: func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
			return nil, nil
		},
	}
	h := NewTimeHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(timeRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodPost, timeRoute+"/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusConflict {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusConflict, w.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if errResp.Code != string(codeNoActiveTimer) {
		t.Fatalf("expected code %s, got %s", codeNoActiveTimer, errResp.Code)
	}
}

func TestTimeHandlerStopHappyPath(t *testing.T) {
	now := time.Now().UTC()
	f := &fakeTimeService{
		stopActiveFn: func() (domain.TimeEntry, error) {
			stopped := now
			dur := int32(60)
			return domain.TimeEntry{ID: uuid.New(), CategoryID: uuid.New(), StartedAt: now.Add(-time.Minute), StoppedAt: &stopped, DurationSeconds: &dur, CreatedAt: now, UpdatedAt: now}, nil
		},
		startFn:          func(categoryID uuid.UUID) (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		getActiveFn:      func() (*domain.TimeEntry, error) { return nil, nil },
		listByCategoryFn: func(categoryID uuid.UUID) ([]domain.TimeEntry, error) { return nil, nil },
		listByCategoryAndRangeFn: func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
			return nil, nil
		},
	}
	h := NewTimeHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(timeRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodPost, timeRoute+"/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	var resp TimeEntryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if resp.StoppedAt == nil || resp.DurationSeconds == nil {
		t.Fatalf("expected stopped entry with duration, got %+v", resp)
	}
}

func TestTimeHandlerActiveHasEntry(t *testing.T) {
	now := time.Now().UTC()
	entry := domain.TimeEntry{ID: uuid.New(), CategoryID: uuid.New(), StartedAt: now, CreatedAt: now, UpdatedAt: now}
	f := &fakeTimeService{
		getActiveFn:      func() (*domain.TimeEntry, error) { return &entry, nil },
		stopActiveFn:     func() (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		startFn:          func(categoryID uuid.UUID) (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		listByCategoryFn: func(categoryID uuid.UUID) ([]domain.TimeEntry, error) { return nil, nil },
		listByCategoryAndRangeFn: func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
			return nil, nil
		},
	}
	h := NewTimeHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(timeRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodGet, timeRoute+"/active", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	var resp TimeEntryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if resp.ID == uuid.Nil || resp.CategoryID == uuid.Nil {
		t.Fatalf("unexpected active entry: %+v", resp)
	}
}

func TestTimeHandlerActiveNoneReturnsNull(t *testing.T) {
	f := &fakeTimeService{
		getActiveFn:      func() (*domain.TimeEntry, error) { return nil, nil },
		stopActiveFn:     func() (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		startFn:          func(categoryID uuid.UUID) (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		listByCategoryFn: func(categoryID uuid.UUID) ([]domain.TimeEntry, error) { return nil, nil },
		listByCategoryAndRangeFn: func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
			return nil, nil
		},
	}
	h := NewTimeHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(timeRoute, h.RegisterRoutes)

	req := httptest.NewRequest(stdhttp.MethodGet, timeRoute+"/active", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	// Expect literal JSON null
	if body := w.Body.String(); body != "null\n" && body != "null" {
		t.Fatalf("expected null body, got %q", body)
	}
}

func TestTimeHandlerEntriesParsingValidation(t *testing.T) {
	now := time.Now().UTC()
	catID := uuid.New()
	f := &fakeTimeService{
		listByCategoryFn: func(categoryID uuid.UUID) ([]domain.TimeEntry, error) {
			if categoryID != catID {
				t.Fatalf("unexpected category id: %s", categoryID)
			}
			return []domain.TimeEntry{{ID: uuid.New(), CategoryID: categoryID, StartedAt: now, CreatedAt: now, UpdatedAt: now}}, nil
		},
		listByCategoryAndRangeFn: func(categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
			if start.After(end) {
				t.Fatalf("start should be <= end")
			}
			return []domain.TimeEntry{{ID: uuid.New(), CategoryID: categoryID, StartedAt: start, StoppedAt: &end, CreatedAt: now, UpdatedAt: now}}, nil
		},
		getActiveFn:  func() (*domain.TimeEntry, error) { return nil, nil },
		stopActiveFn: func() (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
		startFn:      func(categoryID uuid.UUID) (domain.TimeEntry, error) { return domain.TimeEntry{}, nil },
	}
	h := NewTimeHandler(f, slog.Default())
	r := chi.NewRouter()
	r.Route(timeRoute, h.RegisterRoutes)

	// invalid UUID
	req := httptest.NewRequest(stdhttp.MethodGet, categoryEntriesRoute+"not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}

	// invalid time
	req = httptest.NewRequest(stdhttp.MethodGet, categoryEntriesRoute+catID.String()+"&from=bad-time", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}

	// from > to
	from := now.Add(2 * time.Hour).Format(time.RFC3339)
	to := now.Add(1 * time.Hour).Format(time.RFC3339)
	req = httptest.NewRequest(stdhttp.MethodGet, categoryEntriesRoute+catID.String()+"&from="+from+"&to="+to, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusBadRequest {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusBadRequest, w.Code)
	}

	// happy path list by category
	req = httptest.NewRequest(stdhttp.MethodGet, categoryEntriesRoute+catID.String(), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	var list []TimeEntryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf(invalidJsonErrorMessage, err)
	}
	if len(list) != 1 || list[0].CategoryID != catID {
		t.Fatalf("unexpected list: %+v", list)
	}

	// happy path with range
	fromT := now.Add(-1 * time.Hour)
	toT := now.Add(1 * time.Hour)
	req = httptest.NewRequest(stdhttp.MethodGet, categoryEntriesRoute+catID.String()+"&from="+fromT.Format(time.RFC3339)+"&to="+toT.Format(time.RFC3339), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != stdhttp.StatusOK {
		t.Fatalf(statusCodeFailedExpectationMessage, stdhttp.StatusOK, w.Code)
	}
	// ensure JSON is valid
	_, _ = io.ReadAll(w.Body)
}
