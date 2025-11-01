package postgres

import (
	"database/sql"

	"github.com/Gargair/clockwork/server/internal/repository"
)

// Repositories bundles all repository interfaces backed by Postgres.
type Repositories struct {
	Projects    repository.ProjectRepository
	Categories  repository.CategoryRepository
	TimeEntries repository.TimeEntryRepository
}

// NewRepositories constructs all Postgres-backed repositories using the provided *sql.DB.
func NewRepositories(db *sql.DB) Repositories {
	return Repositories{
		Projects:    NewProjectRepository(db),
		Categories:  NewCategoryRepository(db),
		TimeEntries: NewTimeEntryRepository(db),
	}
}
