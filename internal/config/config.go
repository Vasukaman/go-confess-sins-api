package config

import "os"

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL string // Add this line
	// ... keep your other config variables
}

// New loads configuration from environment variables.
func New() *Config {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"), // Add this line
		// ... keep your other os.Getenv calls
	}
}
