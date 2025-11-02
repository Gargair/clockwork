package service

import (
	"context"
	"sort"
	"time"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

// In-memory ProjectRepository fake
type fakeProjectRepo struct {
	items map[uuid.UUID]domain.Project
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{items: make(map[uuid.UUID]domain.Project)}
}

func (r *fakeProjectRepo) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now
	r.items[project.ID] = project
	return project, nil
}

func (r *fakeProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	p, ok := r.items[id]
	if !ok {
		return domain.Project{}, repository.ErrNotFound
	}
	return p, nil
}

func (r *fakeProjectRepo) List(ctx context.Context) ([]domain.Project, error) {
	out := make([]domain.Project, 0, len(r.items))
	for _, p := range r.items {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *fakeProjectRepo) Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error) {
	p, ok := r.items[id]
	if !ok {
		return domain.Project{}, repository.ErrNotFound
	}
	p.Name = name
	p.Description = description
	p.UpdatedAt = time.Now().UTC()
	r.items[id] = p
	return p, nil
}

func (r *fakeProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := r.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// In-memory CategoryRepository fake
type fakeCategoryRepo struct {
	items map[uuid.UUID]domain.Category
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	return &fakeCategoryRepo{items: make(map[uuid.UUID]domain.Category)}
}

func (r *fakeCategoryRepo) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	now := time.Now().UTC()
	category.CreatedAt = now
	category.UpdatedAt = now
	r.items[category.ID] = category
	return category, nil
}

func (r *fakeCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	c, ok := r.items[id]
	if !ok {
		return domain.Category{}, repository.ErrNotFound
	}
	return c, nil
}

func (r *fakeCategoryRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error) {
	var out []domain.Category
	for _, c := range r.items {
		if c.ProjectID == projectID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeCategoryRepo) ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error) {
	var out []domain.Category
	for _, c := range r.items {
		if c.ParentCategoryID != nil && *c.ParentCategoryID == parentID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeCategoryRepo) Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
	c, ok := r.items[id]
	if !ok {
		return domain.Category{}, repository.ErrNotFound
	}
	c.Name = name
	c.Description = description
	c.ParentCategoryID = parentCategoryID
	c.UpdatedAt = time.Now().UTC()
	r.items[id] = c
	return c, nil
}

func (r *fakeCategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := r.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// In-memory TimeEntryRepository fake
type fakeTimeEntryRepo struct {
	items map[uuid.UUID]domain.TimeEntry
}

func newFakeTimeEntryRepo() *fakeTimeEntryRepo {
	return &fakeTimeEntryRepo{items: make(map[uuid.UUID]domain.TimeEntry)}
}

func (r *fakeTimeEntryRepo) Create(ctx context.Context, entry domain.TimeEntry) (domain.TimeEntry, error) {
	now := time.Now().UTC()
	entry.CreatedAt = now
	entry.UpdatedAt = now
	r.items[entry.ID] = entry
	return entry, nil
}

func (r *fakeTimeEntryRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.TimeEntry, error) {
	e, ok := r.items[id]
	if !ok {
		return domain.TimeEntry{}, repository.ErrNotFound
	}
	return e, nil
}

func (r *fakeTimeEntryRepo) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error) {
	var out []domain.TimeEntry
	for _, e := range r.items {
		if e.CategoryID == categoryID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

func (r *fakeTimeEntryRepo) ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
	var out []domain.TimeEntry
	for _, e := range r.items {
		if e.CategoryID == categoryID && !e.StartedAt.Before(start) && !e.StartedAt.After(end) {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

func (r *fakeTimeEntryRepo) FindActive(ctx context.Context) (*domain.TimeEntry, error) {
	var candidates []domain.TimeEntry
	for _, e := range r.items {
		if e.StoppedAt == nil {
			candidates = append(candidates, e)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].StartedAt.After(candidates[j].StartedAt) })
	e := candidates[0]
	return &e, nil
}

func (r *fakeTimeEntryRepo) Stop(ctx context.Context, id uuid.UUID, stoppedAt time.Time, durationSeconds *int32) (domain.TimeEntry, error) {
	e, ok := r.items[id]
	if !ok {
		return domain.TimeEntry{}, repository.ErrNotFound
	}
	e.StoppedAt = &stoppedAt
	e.DurationSeconds = durationSeconds
	e.UpdatedAt = time.Now().UTC()
	r.items[id] = e
	return e, nil
}
