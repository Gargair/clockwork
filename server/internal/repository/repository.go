package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/google/uuid"
)

// Package-level errors for common repository conditions.
var (
	ErrNotFound            = errors.New("repository: not found")
	ErrDuplicate           = errors.New("repository: duplicate")
	ErrForeignKeyViolation = errors.New("repository: foreign key violation")
)

// ProjectRepository defines CRUD operations for projects.
type ProjectRepository interface {
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// CategoryRepository defines operations for categories.
// Note: Update must not allow changing ProjectID (enforced by implementations/services).
type CategoryRepository interface {
	Create(ctx context.Context, category domain.Category) (domain.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error)
	Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// TimeEntryRepository defines operations for time entries.
type TimeEntryRepository interface {
	Create(ctx context.Context, entry domain.TimeEntry) (domain.TimeEntry, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.TimeEntry, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error)
	ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error)
	FindActive(ctx context.Context) (*domain.TimeEntry, error)
	Stop(ctx context.Context, id uuid.UUID, stoppedAt time.Time, durationSeconds *int32) (domain.TimeEntry, error)
}
