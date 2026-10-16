package config

import (
	"os"
)

var Cfg *Config

type Config struct {
	API         *APIConfig
	Database    *DatabaseConfig
	Environment string
}

type APIConfig struct {
	Address string
}

type DatabaseConfig struct {
	// URL is a libpq connection string (DSN or postgres:// URL).
	URL string
}

func Load() error {
	addr := os.Getenv("API_ADDRESS")
	if addr == "" {
		addr = "localhost:8080"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Development-only fallback, matching the database shop-db-models'
		// Makefile provisions. Production deployments must set DATABASE_URL
		// to point at the real database.
		dbURL = "postgres://madman@localhost:5433/postgres?sslmode=disable"
	}

	Cfg = &Config{
		API: &APIConfig{
			Address: addr,
		},
		Database: &DatabaseConfig{
			URL: dbURL,
		},
		Environment: env,
	}
	return nil
}
