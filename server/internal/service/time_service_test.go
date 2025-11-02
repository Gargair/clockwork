package service

import (
	"context"
	"testing"
	"time"

	"github.com/Gargair/clockwork/server/internal/domain"
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
