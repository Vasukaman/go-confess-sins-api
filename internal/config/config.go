package config

import "os"

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL   string
	WebsiteAPIKey string
	SinApiUrl     string
}

// New loads configuration from environment variables.
func New() *Config {
	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		WebsiteAPIKey: os.Getenv("WEBSITE_API_KEY"),
		SinApiUrl:     os.Getenv("SIN_API_URL"),
	}
}
