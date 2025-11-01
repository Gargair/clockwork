package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// DefaultDatabaseURL is a convenient local default for developers using Docker Compose.
// It is used only when DATABASE_URL is not set (and not provided via .env).
const DefaultDatabaseURL string = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"

// Config holds typed server configuration.
type Config struct {
	// DatabaseURL is the PostgreSQL connection string.
	DatabaseURL string
}

// Load reads configuration from environment (and optional .env) and validates it.
// Precedence: env var → .env → DefaultDatabaseURL.
func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		if v := readDotEnvVar("DATABASE_URL"); v != "" {
			databaseURL = v
		}
	}
	if databaseURL == "" {
		databaseURL = DefaultDatabaseURL
	}

	if err := validateDatabaseURL(databaseURL); err != nil {
		return Config{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	return Config{DatabaseURL: databaseURL}, nil
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

// readDotEnvVar is a minimal .env reader for a single variable.
// It returns the value for key if found in a local .env file (same CWD), otherwise "".
func readDotEnvVar(key string) string {
	f, err := os.Open(".env")
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if name != key {
			continue
		}
		value = strings.TrimSpace(value)
		// Remove optional surrounding quotes
		value = strings.Trim(value, "\"'")
		return value
	}
	return ""
}
