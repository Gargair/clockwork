package service

import (
	"context"
	"time"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

// ProjectService defines project-related operations.
type ProjectService interface {
	Create(ctx context.Context, name string, description *string) (domain.Project, error)
	Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
}

// CategoryService defines category-related operations and invariants.
type CategoryService interface {
	Create(ctx context.Context, projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error)
	Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error)
}

// TimeTrackingService defines timer operations and invariants.
type TimeTrackingService interface {
	Start(ctx context.Context, categoryID uuid.UUID) (domain.TimeEntry, error)
	StopActive(ctx context.Context) (domain.TimeEntry, error)
	GetActive(ctx context.Context) (*domain.TimeEntry, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error)
	ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error)
}

// NewProjectService constructs a ProjectService.
func NewProjectService(repo repository.ProjectRepository) ProjectService {
	return nil
}

// NewCategoryService constructs a CategoryService.
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return nil
}

// NewTimeTrackingService constructs a TimeTrackingService.
func NewTimeTrackingService(repo repository.TimeEntryRepository, categoryRepo repository.CategoryRepository, clk clock.Clock) TimeTrackingService {
	return nil
}
