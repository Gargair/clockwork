package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	// TODO: Start HTTP server in Milestone 3
}
