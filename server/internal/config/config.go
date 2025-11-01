package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds typed server configuration.
type Config struct {
	// DatabaseURL is the PostgreSQL connection string.
	DatabaseURL string `env:"DATABASE_URL" required:"true"`
	// AutoMigrate controls whether the server runs DB migrations on startup.
	AutoMigrate bool `env:"DB_AUTO_MIGRATE" envDefault:"false"`
	// MigrationsDir is the directory containing goose SQL migrations.
	MigrationsDir string `env:"MIGRATIONS_DIR" envDefault:"server/migrations"`
}

// Load reads configuration from environment (and optional .env) and validates it.
func Load() (Config, error) {
	// Load .env for local dev; no-op if missing
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if err := validateDatabaseURL(cfg.DatabaseURL); err != nil {
		return Config{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	return cfg, nil
}

// validateDatabaseURL ensures the DSN is a parseable postgres URL with host and db name.
func validateDatabaseURL(databaseURL string) error {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errors.New("scheme must be 'postgres' or 'postgresql'")
	}
	if parsed.Host == "" {
		return errors.New("host is required")
	}
	// Path typically contains the database name, e.g. /clockwork
	if strings.Trim(parsed.Path, "/ ") == "" {
		return errors.New("database name is required in URL path")
	}
	return nil
}
