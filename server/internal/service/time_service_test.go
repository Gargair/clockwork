package service

import (
	"context"
	"testing"
	"time"

	"github.com/Gargair/clockwork/server/internal/domain"
    "github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

func seedCategory(t *testing.T, repo *fakeCategoryRepo) domain.Category {
	t.Helper()
	c := domain.Category{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		Name:      "Cat",
	}
	out, err := repo.Create(context.Background(), c)
	if err != nil {
		t.Fatalf("seed category failed: %v", err)
	}
	return out
}

func TestTimeTrackingServiceStartNoActiveCreatesActive(t *testing.T) {
	ctx := context.Background()
	catRepo := newFakeCategoryRepo()
	timeRepo := newFakeTimeEntryRepo()
	start := time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC)
	clk := newTestClock(start)
	svc := NewTimeTrackingService(timeRepo, catRepo, clk)

	cat := seedCategory(t, catRepo)
	entry, err := svc.Start(ctx, cat.ID)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if !entry.StartedAt.Equal(start) || entry.StoppedAt == nil && entry.DurationSeconds != nil {
		// No duration or stoppedAt for a new active entry
	}
	active, err := svc.GetActive(ctx)
	if err != nil {
		t.Fatalf("get active failed: %v", err)
	}
	if active == nil || active.ID != entry.ID {
		t.Fatalf("active mismatch")
	}
}

func TestTimeTrackingServiceStartStopsPreviousAndStartsNewAtSameNow(t *testing.T) {
	ctx := context.Background()
	catRepo := newFakeCategoryRepo()
	timeRepo := newFakeTimeEntryRepo()
	t0 := time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC)
	clk := newTestClock(t0)
	svc := NewTimeTrackingService(timeRepo, catRepo, clk)

	cat1 := seedCategory(t, catRepo)
	cat2 := seedCategory(t, catRepo)

	first, err := svc.Start(ctx, cat1.ID)
	if err != nil {
		t.Fatalf("first start failed: %v", err)
	}

	t1 := t0.Add(5 * time.Minute)
	clk.Set(t1)

	second, err := svc.Start(ctx, cat2.ID)
	if err != nil {
		t.Fatalf("second start failed: %v", err)
	}

	if !second.StartedAt.Equal(t1) {
		t.Fatalf("expected second startedAt %v, got %v", t1, second.StartedAt)
	}
	// Check first got stopped exactly at t1
	updatedFirst, err := timeRepo.GetByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("get first failed: %v", err)
	}
	if updatedFirst.StoppedAt == nil || !updatedFirst.StoppedAt.Equal(t1) {
		t.Fatalf("expected first stoppedAt %v, got %v", t1, updatedFirst.StoppedAt)
	}
}

func TestTimeTrackingServiceStopActiveComputesDurationAndClearsActive(t *testing.T) {
	ctx := context.Background()
	catRepo := newFakeCategoryRepo()
	timeRepo := newFakeTimeEntryRepo()
	t0 := time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC)
	clk := newTestClock(t0)
	svc := NewTimeTrackingService(timeRepo, catRepo, clk)

	cat := seedCategory(t, catRepo)
	_, err := svc.Start(ctx, cat.ID)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	clk.Advance(1*time.Hour + 30*time.Second)
	stopped, err := svc.StopActive(ctx)
	if err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if stopped.DurationSeconds == nil || *stopped.DurationSeconds != int32(3600+30) {
		t.Fatalf("expected duration 3630, got %v", stopped.DurationSeconds)
	}

	active, err := svc.GetActive(ctx)
	if err != nil {
		t.Fatalf("get active failed: %v", err)
	}
	if active != nil {
		t.Fatalf("expected no active entry after stop")
	}
}

// --- Error handling tests ---

type errCategoryRepo struct{ err error }

func (r errCategoryRepo) GetByID(context.Context, uuid.UUID) (domain.Category, error) { return domain.Category{}, r.err }
func (r errCategoryRepo) Create(context.Context, domain.Category) (domain.Category, error) { return domain.Category{}, nil }
func (r errCategoryRepo) ListByProject(context.Context, uuid.UUID) ([]domain.Category, error) {
    return nil, nil
}
func (r errCategoryRepo) ListChildren(context.Context, uuid.UUID) ([]domain.Category, error) { return nil, nil }
func (r errCategoryRepo) Update(context.Context, uuid.UUID, string, *string, *uuid.UUID) (domain.Category, error) {
    return domain.Category{}, nil
}
func (r errCategoryRepo) Delete(context.Context, uuid.UUID) error { return nil }

type stubTimeRepo struct{
    active *domain.TimeEntry
    findErr error
    stopErr error
    createErr error
}

func (r stubTimeRepo) Create(context.Context, domain.TimeEntry) (domain.TimeEntry, error) {
    return domain.TimeEntry{}, r.createErr
}
func (r stubTimeRepo) GetByID(context.Context, uuid.UUID) (domain.TimeEntry, error) { return domain.TimeEntry{}, nil }
func (r stubTimeRepo) ListByCategory(context.Context, uuid.UUID) ([]domain.TimeEntry, error) { return nil, nil }
func (r stubTimeRepo) ListByCategoryAndRange(context.Context, uuid.UUID, time.Time, time.Time) ([]domain.TimeEntry, error) {
    return nil, nil
}
func (r stubTimeRepo) FindActive(context.Context) (*domain.TimeEntry, error) { return r.active, r.findErr }
func (r stubTimeRepo) Stop(context.Context, uuid.UUID, time.Time, *int32) (domain.TimeEntry, error) {
    return domain.TimeEntry{}, r.stopErr
}

func TestTimeTrackingServiceStartReturnsCategoryError(t *testing.T) {
    ctx := context.Background()
    clk := newTestClock(time.Now().UTC())
    svc := NewTimeTrackingService(stubTimeRepo{}, errCategoryRepo{err: repository.ErrNotFound}, clk)
    _, err := svc.Start(ctx, uuid.New())
    if err == nil || err != repository.ErrNotFound {
        t.Fatalf("expected ErrNotFound, got %v", err)
    }
}

func TestTimeTrackingServiceStartPropagatesFindActiveError(t *testing.T) {
    ctx := context.Background()
    catRepo := newFakeCategoryRepo()
    seedCategory(t, catRepo)
    findErr := repository.ErrDuplicate
    svc := NewTimeTrackingService(stubTimeRepo{findErr: findErr}, catRepo, newTestClock(time.Now().UTC()))
    _, err := svc.Start(ctx, seedCategory(t, catRepo).ID)
    if err == nil || err != findErr {
        t.Fatalf("expected findErr, got %v", err)
    }
}

func TestTimeTrackingServiceStartPropagatesStopError(t *testing.T) {
    ctx := context.Background()
    catRepo := newFakeCategoryRepo()
    cat := seedCategory(t, catRepo)
    now := time.Now().UTC()
    active := domain.TimeEntry{ID: uuid.New(), CategoryID: cat.ID, StartedAt: now.Add(-time.Minute)}
    stopErr := repository.ErrForeignKeyViolation
    svc := NewTimeTrackingService(stubTimeRepo{active: &active, stopErr: stopErr}, catRepo, newTestClock(now))
    _, err := svc.Start(ctx, cat.ID)
    if err == nil || err != stopErr {
        t.Fatalf("expected stopErr, got %v", err)
    }
}

func TestTimeTrackingServiceStopActiveNoActiveReturnsErr(t *testing.T) {
    ctx := context.Background()
    svc := NewTimeTrackingService(stubTimeRepo{}, newFakeCategoryRepo(), newTestClock(time.Now().UTC()))
    _, err := svc.StopActive(ctx)
    if err == nil || err != ErrNoActiveTimer {
        t.Fatalf("expected ErrNoActiveTimer, got %v", err)
    }
}

func TestTimeTrackingServiceStopActivePropagatesFindError(t *testing.T) {
    ctx := context.Background()
    findErr := repository.ErrDuplicate
    svc := NewTimeTrackingService(stubTimeRepo{findErr: findErr}, newFakeCategoryRepo(), newTestClock(time.Now().UTC()))
    _, err := svc.StopActive(ctx)
    if err == nil || err != findErr {
        t.Fatalf("expected findErr, got %v", err)
    }
}

func TestTimeTrackingServiceStopActivePropagatesStopError(t *testing.T) {
    ctx := context.Background()
    now := time.Now().UTC()
    active := &domain.TimeEntry{ID: uuid.New(), StartedAt: now.Add(-time.Minute)}
    stopErr := repository.ErrForeignKeyViolation
    svc := NewTimeTrackingService(stubTimeRepo{active: active, stopErr: stopErr}, newFakeCategoryRepo(), newTestClock(now))
    _, err := svc.StopActive(ctx)
    if err == nil || err != stopErr {
        t.Fatalf("expected stopErr, got %v", err)
    }
}

// --- List propagation tests ---

func TestTimeTrackingServiceListByCategoryPropagatesResults(t *testing.T) {
    ctx := context.Background()
    repo := newFakeTimeEntryRepo()
    catA := uuid.New()
    catB := uuid.New()

    t0 := time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC)
    t1 := t0.Add(1 * time.Hour)

    // Two entries for catA (different times), one for catB
    if _, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: catA, StartedAt: t0}); err != nil { t.Fatalf("create: %v", err) }
    e2, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: catA, StartedAt: t1})
    if err != nil { t.Fatalf("create: %v", err) }
    if _, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: catB, StartedAt: t1}); err != nil { t.Fatalf("create: %v", err) }

    svc := NewTimeTrackingService(repo, newFakeCategoryRepo(), newTestClock(t0))
    got, err := svc.ListByCategory(ctx, catA)
    if err != nil { t.Fatalf("ListByCategory: %v", err) }
    if len(got) != 2 { t.Fatalf("expected 2 entries, got %d", len(got)) }
    // Repo sorts by StartedAt desc → e2 first
    if got[0].ID != e2.ID {
        t.Fatalf("expected first id %s, got %s", e2.ID, got[0].ID)
    }
}

func TestTimeTrackingServiceListByCategoryAndRangePropagatesResults(t *testing.T) {
    ctx := context.Background()
    repo := newFakeTimeEntryRepo()
    cat := uuid.New()

    start := time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC)
    mid := start.Add(30 * time.Minute)
    end := start.Add(1 * time.Hour)

    before := start.Add(-1 * time.Minute)
    after := end.Add(1 * time.Minute)

    // Create entries around the range boundaries
    if _, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: cat, StartedAt: before}); err != nil { t.Fatalf("create: %v", err) }
    eStart, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: cat, StartedAt: start})
    if err != nil { t.Fatalf("create: %v", err) }
    eMid, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: cat, StartedAt: mid})
    if err != nil { t.Fatalf("create: %v", err) }
    eEnd, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: cat, StartedAt: end})
    if err != nil { t.Fatalf("create: %v", err) }
    if _, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: cat, StartedAt: after}); err != nil { t.Fatalf("create: %v", err) }

    // Distractor in another category within range
    if _, err := repo.Create(ctx, domain.TimeEntry{ID: uuid.New(), CategoryID: uuid.New(), StartedAt: mid}); err != nil { t.Fatalf("create: %v", err) }

    svc := NewTimeTrackingService(repo, newFakeCategoryRepo(), newTestClock(start))
    got, err := svc.ListByCategoryAndRange(ctx, cat, start, end)
    if err != nil { t.Fatalf("ListByCategoryAndRange: %v", err) }
    if len(got) != 3 { t.Fatalf("expected 3 entries in range, got %d", len(got)) }
    // Repo sorts desc by StartedAt: end, mid, start
    if got[0].ID != eEnd.ID || got[1].ID != eMid.ID || got[2].ID != eStart.ID {
        t.Fatalf("unexpected order: got [%s,%s,%s]", got[0].ID, got[1].ID, got[2].ID)
    }
}
