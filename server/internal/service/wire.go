package service

import (
	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/repository"
)

// Services bundles all domain services.
type Services struct {
	Projects   ProjectService
	Categories CategoryService
	Time       TimeTrackingService
}

// NewServices constructs all services from repositories and a clock.
func NewServices(repos struct {
	Projects    repository.ProjectRepository
	Categories  repository.CategoryRepository
	TimeEntries repository.TimeEntryRepository
}, clk clock.Clock) Services {
	return Services{
		Projects:   NewProjectService(repos.Projects),
		Categories: NewCategoryService(repos.Categories),
		Time:       NewTimeTrackingService(repos.TimeEntries, repos.Categories, clk),
	}
}
