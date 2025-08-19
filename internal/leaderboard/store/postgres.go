package store

import (
	"context"
	"go-confess-sins-api/pkg/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store handles all database operations.
type Store struct {
	db *pgxpool.Pool
}

// New creates a new store with a database connection pool.
func New(ctx context.Context, connectionURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		return nil, err
	}
	// Create the table if it doesn't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS leaderboard (
			sin_id INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			count INTEGER,
			tags TEXT[],
			severity SMALLINT
		);
	`)
	if err != nil {
		return nil, err
	}
	return &Store{db: pool}, nil
}

// Close closes the database connection pool.
func (s *Store) Close() {
	s.db.Close()
}

// UpdateSinFromEvent takes a Sin model and "upserts" it into the leaderboard table.
func (s *Store) UpdateSinFromEvent(ctx context.Context, sin models.Sin) error {
	query := `
		INSERT INTO leaderboard (sin_id, description, count, tags, severity, emoji)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (sin_id) DO UPDATE
		SET count = EXCLUDED.count,
			tags = EXCLUDED.tags,
			severity = EXCLUDED.severity,
			emoji = EXCLUDED.emoji;
	`
	_, err := s.db.Exec(ctx, query, sin.ID, sin.Description, sin.Count, sin.Tags, sin.Severity, sin.Emoji)
	return err
}

// GetLeaderboard now selects and scans the emoji.
func (s *Store) GetLeaderboard(ctx context.Context) ([]models.Sin, error) {
	rows, err := s.db.Query(ctx, `
		SELECT sin_id, description, count, COALESCE(tags, '{}'), severity, emoji 
		FROM leaderboard 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sins []models.Sin
	for rows.Next() {
		var sin models.Sin
		// THE FIX IS HERE: Add &sin.Emoji to the Scan
		if err := rows.Scan(&sin.ID, &sin.Description, &sin.Count, &sin.Tags, &sin.Severity, &sin.Emoji); err != nil {
			return nil, err
		}
		sins = append(sins, sin)
	}
	return sins, nil
}
