//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestTimeEntryRepositoryCreateAndFindActiveThenStopFlowIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)
	tr := NewTimeEntryRepository(db)

	p, err := pr.Create(ctx, NewProject("te-proj", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}
	c, err := cr.Create(ctx, NewCategory(p.ID, "work", nil, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "category", err)
	}

	start := time.Now().UTC().Add(-1 * time.Minute)
	created, err := tr.Create(ctx, NewTimeEntry(c.ID, start))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "time entry", err)
	}
	if created.StoppedAt != nil || created.DurationSeconds != nil {
		t.Fatalf("expected active entry on create, got stoppedAt=%v duration=%v", created.StoppedAt, created.DurationSeconds)
	}

	active, err := tr.FindActive(ctx)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if active == nil {
		t.Fatalf("expected active entry, got nil")
	}
	if active.ID != created.ID {
		t.Fatalf("unexpected active entry id: got %v want %v", active.ID, created.ID)
	}

	// Stop the active entry
	stoppedAt := time.Now().UTC()
	dur := int32(stoppedAt.Sub(created.StartedAt).Seconds())
	stopped, err := tr.Stop(ctx, created.ID, stoppedAt, &dur)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stopped.StoppedAt == nil || stopped.DurationSeconds == nil {
		t.Fatalf("expected stopped fields to be set: stoppedAt=%v duration=%v", stopped.StoppedAt, stopped.DurationSeconds)
	}

	// GetByID reflects changes
	fetched, err := tr.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.StoppedAt == nil || fetched.DurationSeconds == nil {
		t.Fatalf("expected fetched stopped fields to be set")
	}

	// After stop, no active entry
	none, err := tr.FindActive(ctx)
	if err != nil {
		t.Fatalf("FindActive after stop: %v", err)
	}
	if none != nil {
		t.Fatalf("expected no active entry, got %+v", none)
	}
}

func TestTimeEntryRepositoryListByCategoryOrderDescIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)
	tr := NewTimeEntryRepository(db)

	p, err := pr.Create(ctx, NewProject("order-proj", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}
	c, err := cr.Create(ctx, NewCategory(p.ID, "cat-order", nil, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "category", err)
	}

	t0 := time.Now().UTC().Add(-3 * time.Hour)
	t1 := time.Now().UTC().Add(-2 * time.Hour)
	t2 := time.Now().UTC().Add(-1 * time.Hour)
	if _, err := tr.Create(ctx, NewTimeEntry(c.ID, t0)); err != nil {
		t.Fatalf("create t0: %v", err)
	}
	if _, err := tr.Create(ctx, NewTimeEntry(c.ID, t1)); err != nil {
		t.Fatalf("create t1: %v", err)
	}
	if _, err := tr.Create(ctx, NewTimeEntry(c.ID, t2)); err != nil {
		t.Fatalf("create t2: %v", err)
	}

	list, err := tr.ListByCategory(ctx, c.ID)
	if err != nil {
		t.Fatalf("ListByCategory: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}
	if !(list[0].StartedAt.After(list[1].StartedAt) || list[0].StartedAt.Equal(list[1].StartedAt)) {
		t.Fatalf("expected list[0] >= list[1] by started_at")
	}
	if !(list[1].StartedAt.After(list[2].StartedAt) || list[1].StartedAt.Equal(list[2].StartedAt)) {
		t.Fatalf("expected list[1] >= list[2] by started_at")
	}
}

func TestTimeEntryRepositoryListByCategoryAndRangeInclusiveIntegration(t *testing.T) {
	db := OpenDBFromEnv(t)
	t.Cleanup(func() { _ = db.Close() })
	TruncateAll(t, db)

	ctx := context.Background()
	pr := NewProjectRepository(db)
	cr := NewCategoryRepository(db)
	tr := NewTimeEntryRepository(db)

	p, err := pr.Create(ctx, NewProject("range-proj", nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "project", err)
	}
	c, err := cr.Create(ctx, NewCategory(p.ID, "cat-range", nil, nil))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "category", err)
	}

	base := time.Now().UTC().Add(-5 * time.Hour)
	_, err = tr.Create(ctx, NewTimeEntry(c.ID, base))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "e0", err)
	}
	e1, err := tr.Create(ctx, NewTimeEntry(c.ID, base.Add(1*time.Hour)))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "e1", err)
	}
	e2, err := tr.Create(ctx, NewTimeEntry(c.ID, base.Add(2*time.Hour)))
	if err != nil {
		t.Fatalf(CreateFailedErrorMessage, "e2", err)
	}

	// inclusive range [e1.started, e2.started]
	start := e1.StartedAt
	end := e2.StartedAt
	list, err := tr.ListByCategoryAndRange(ctx, c.ID, start, end)
	if err != nil {
		t.Fatalf("ListByCategoryAndRange: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entries in range, got %d", len(list))
	}
	ids := map[string]bool{e1.ID.String(): false, e2.ID.String(): false}
	for _, e := range list {
		if _, ok := ids[e.ID.String()]; ok {
			ids[e.ID.String()] = true
		}
		if e.StartedAt.Before(start) || e.StartedAt.After(end) {
			t.Fatalf("entry outside range: %v", e.StartedAt)
		}
	}
	for id, seen := range ids {
		if !seen {
			t.Fatalf("expected id in range: %s", id)
		}
	}
}
