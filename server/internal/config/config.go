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
	// Port is the HTTP server port to bind.
	Port int `env:"PORT" envDefault:"8080"`
	// Env indicates the environment: development or production.
	Env string `env:"ENV" envDefault:"development"`
	// StaticDir is the directory containing built client assets.
	StaticDir string `env:"STATIC_DIR" envDefault:"client/dist"`
	// AllowedOrigins lists origins allowed by CORS. CSV. Default depends on Env.
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envSeparator:","`
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
	if err := validateEnv(cfg.Env); err != nil {
		return Config{}, err
	}
	if cfg.Port <= 0 {
		return Config{}, errors.New("PORT must be > 0")
	}
	// Default CORS origins: '*' in development when not explicitly set.
	if len(cfg.AllowedOrigins) == 0 && cfg.Env == "development" {
		cfg.AllowedOrigins = []string{"*"}
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

// validateEnv ensures environment string is one of the supported values.
func validateEnv(env string) error {
	switch env {
	case "development", "production":
		return nil
	default:
		return fmt.Errorf("ENV must be one of 'development' or 'production', got %q", env)
	}
}
