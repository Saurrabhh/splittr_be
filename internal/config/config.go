package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds environment configurations for the application.
type Config struct {
	Port              string `env:"PORT" envDefault:"8080"`
	DatabaseURL       string `env:"DATABASE_URL,required"`
	FirebaseProjectID string `env:"FIREBASE_PROJECT_ID,required"`
}

// Load reads config from environment variables.
func Load() (*Config, error) {
	// Attempt to load .env file, but ignore error if file is missing (e.g. in production)
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env file: %w", err)
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
