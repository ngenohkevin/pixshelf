package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	ServerPort   int
	DatabaseURL  string
	Environment  string
	ImageStorage string
	BaseURL      string
}

// Load returns the application configuration
func Load() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		ServerPort:   getEnvInt("PORT", 8080),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/pixshelf?sslmode=disable"),
		Environment:  getEnv("ENV", "development"),
		ImageStorage: getEnv("IMAGE_STORAGE", "./static/images"),
		BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
	}

	// Print the config for debugging
	log.Printf("Config: %+v", cfg)

	// Create image storage directory if it doesn't exist
	if err := os.MkdirAll(cfg.ImageStorage, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image storage directory: %w", err)
	}

	// Convert relative path to absolute path
	if !filepath.IsAbs(cfg.ImageStorage) {
		absPath, err := filepath.Abs(cfg.ImageStorage)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		cfg.ImageStorage = absPath
	}

	return cfg, nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
