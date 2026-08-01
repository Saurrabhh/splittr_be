package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidLocalEnv(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("expected DATABASE_URL %q, got %q", "postgres://user:pass@localhost:5432/testdb?sslmode=disable", cfg.DatabaseURL)
	}
}

func TestLoad_MissingAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")

	cfg, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing APP_ENV, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %v", cfg)
	}

	expectedMsg := "APP_ENV environment variable is not set"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestLoad_InvalidAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "staging")

	cfg, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid APP_ENV, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %v", cfg)
	}

	expectedMsg := "must be one of 'local', 'dev', or 'prod'"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestLoad_MissingRequiredDatabaseURL(t *testing.T) {
	// Change working directory to isolated temp dir so env/local/.env is not loaded by godotenv
	t.Chdir(t.TempDir())

	t.Setenv("APP_ENV", "local")
	require.NoError(t, os.Unsetenv("DATABASE_URL"))

	cfg, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %v", cfg)
	}

	if !strings.Contains(err.Error(), "parse config") {
		t.Errorf("expected error message to contain parse error, got %q", err.Error())
	}
}
