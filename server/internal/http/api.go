package http

import (
	"database/sql"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/repository"
	repo_pg "github.com/Gargair/clockwork/server/internal/repository/postgres"
	"github.com/Gargair/clockwork/server/internal/service"
	"github.com/go-chi/chi/v5"
)

// HealthzHandler serves the /healthz route.
type ApiHandler struct {
	db  *sql.DB
	clk clock.Clock
}

func (h ApiHandler) mountAPI(api chi.Router) {
	// Construct repositories and services
	repos := repo_pg.NewRepositories(h.db)
	svcs := service.NewServices(struct {
		Projects    repository.ProjectRepository
		Categories  repository.CategoryRepository
		TimeEntries repository.TimeEntryRepository
	}{
		Projects:    repos.Projects,
		Categories:  repos.Categories,
		TimeEntries: repos.TimeEntries,
	}, h.clk)

	// Handlers
	projH := NewProjectHandler(svcs.Projects)
	catH := NewCategoryHandler(svcs.Categories)
	timeH := NewTimeHandler(svcs.Time)

	// /api/projects
	api.Route("/projects", func(rp chi.Router) {
		projH.RegisterRoutes(rp)
		// /api/projects/{projectId}/categories
		rp.Route("/{projectId}/categories", catH.RegisterRoutes)
	})

	// /api/time
	api.Route("/time", timeH.RegisterRoutes)
}
