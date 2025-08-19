package sinapi

import (
	"go-confess-sins-api/pkg/models"
)

// This file defines the interfaces our core logic depends on.

type Store interface {
	CreateAPIKey() (string, error)
	GetAPIKeyID(apiKey string) (int, error)
	IncrementSinCount(apiKeyID int, description string, tags []string, severity *int, emoji *string) (models.Sin, error)
	GetSinsByAPIKeyID(apiKeyID int) ([]models.Sin, error)
	GetSins(limit int) ([]models.Sin, error)
}

type Publisher interface {
	PublishSinUpdate(sin models.Sin) error
}
