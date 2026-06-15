package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	DatabaseDSN string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, reading from environment")
	}

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
	}

	cfg.DatabaseDSN = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "postgres"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "subscriptions"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	port, err := strconv.Atoi(c.AppPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid APP_PORT: %q", c.AppPort)
	}

	if c.DatabaseDSN == "" {
		return fmt.Errorf("database DSN is empty")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
