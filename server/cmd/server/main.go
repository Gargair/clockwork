package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apphttp "github.com/Gargair/clockwork/server/internal/http"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/config"
	"github.com/Gargair/clockwork/server/internal/db"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if cfg.AutoMigrate {
		fromV, toV, err := db.RunMigrations(ctx, cfg.DatabaseURL, cfg.MigrationsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrations failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "migrations applied: from %d to %d in %s\n", fromV, toV, cfg.MigrationsDir)
	}

	logger := buildLogger(cfg)

	dbConn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db_open_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = dbConn.Close() }()

	handler := apphttp.NewRouter(cfg, dbConn, clock.NewSystemClock(), logger)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go serveHTTP(srv, logger)
	logger.Info("server_started",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.Port),
		slog.String("static_dir", cfg.StaticDir),
	)

	waitForShutdown(srv, logger)
}

func buildLogger(cfg config.Config) *slog.Logger {
	if cfg.Env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func serveHTTP(srv *http.Server, logger *slog.Logger) {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server_listen_failed", slog.String("error", err.Error()))
	}
}

func waitForShutdown(srv *http.Server, logger *slog.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_error", slog.String("error", err.Error()))
	} else {
		logger.Info("server_stopped")
	}
}
