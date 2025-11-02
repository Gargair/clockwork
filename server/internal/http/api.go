package http

import (
	"database/sql"

	"log/slog"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/repository"
	repo_pg "github.com/Gargair/clockwork/server/internal/repository/postgres"
	"github.com/Gargair/clockwork/server/internal/service"
	"github.com/go-chi/chi/v5"
)

// HealthzHandler serves the /healthz route.
type ApiHandler struct {
	db     *sql.DB
	clk    clock.Clock
	logger *slog.Logger
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
	projH := NewProjectHandler(svcs.Projects, h.logger)
	catH := NewCategoryHandler(svcs.Categories, h.logger)
	timeH := NewTimeHandler(svcs.Time, h.logger)

	// /api/projects
	api.Route("/projects", func(rp chi.Router) {
		projH.RegisterRoutes(rp)
		// /api/projects/{projectId}/categories
		rp.Route("/{projectId}/categories", catH.RegisterRoutes)
	})

	// /api/time
	api.Route("/time", timeH.RegisterRoutes)
}
