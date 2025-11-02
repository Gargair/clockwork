package http

import (
	"database/sql"
	"log/slog"
	stdhttp "net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/config"
)

// NewRouter wires the HTTP router, middleware, and routes.
func NewRouter(cfg config.Config, dbConn *sql.DB, clk clock.Clock, logger *slog.Logger) stdhttp.Handler {
	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(loggingMiddleware(logger))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Method("GET", "/healthz", HealthzHandler{db: dbConn, clk: clk})

	// API routes
	r.Route("/api", ApiHandler{db: dbConn, clk: clk}.mountAPI)

	// Static files (production only)
	if cfg.Env == "production" {
		if _, err := os.Stat(cfg.StaticDir); err == nil {
			r.Handle("/*", NewStaticHandler(cfg.StaticDir))
		}
	}

	return r
}
