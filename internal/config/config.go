package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds environment configurations for the application.
type Config struct {
	Environment     string `env:"APP_ENV"`
	Port            string `env:"PORT" envDefault:"8080"`
	DatabaseURL     string `env:"DATABASE_URL,required"`
	SwaggerUser     string `env:"SWAGGER_USER" envDefault:"admin"`
	SwaggerPassword string `env:"SWAGGER_PASSWORD" envDefault:"admin"`
	CloudinaryCloudName string `env:"CLOUDINARY_CLOUD_NAME"`
	CloudinaryAPIKey    string `env:"CLOUDINARY_API_KEY"`
	CloudinaryAPISecret string `env:"CLOUDINARY_API_SECRET"`
}

// Load reads config from environment variables.
func Load() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		return nil, fmt.Errorf("APP_ENV environment variable is not set")
	}

	// Validate APP_ENV
	switch appEnv {
	case "local", "dev", "prod":
		// Valid environments
	default:
		return nil, fmt.Errorf("invalid APP_ENV %q: must be one of 'local', 'dev', or 'prod'", appEnv)
	}

	envFile := fmt.Sprintf("env/%s/.env", appEnv)
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("load env file %s: %w", envFile, err)
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
