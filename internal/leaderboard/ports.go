package leaderboard

import (
	"context"
	"go-confess-sins-api/pkg/models"
)

// Store is an interface that describes what our handler needs from a database.
type Store interface {
	GetLeaderboard(ctx context.Context) ([]models.Sin, error)
}
