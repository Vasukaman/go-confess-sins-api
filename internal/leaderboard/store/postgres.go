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
	// Your complex SQL query is good.
	query := `
		INSERT INTO leaderboard (sin_id, description, count, tags, severity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sin_id) DO UPDATE
		SET count = EXCLUDED.count,
			tags = EXCLUDED.tags,
			severity = EXCLUDED.severity;
	`
	_, err := s.db.Exec(ctx, query, sin.ID, sin.Description, sin.Count, sin.Tags, sin.Severity)
	return err
}

// GetLeaderboard fetches the top 10 sins, ordered by count.
func (s *Store) GetLeaderboard(ctx context.Context) ([]models.Sin, error) {
	// Your GetLeaderboard logic is also correct.
	rows, err := s.db.Query(ctx, `
		SELECT sin_id, description, count, COALESCE(tags, '{}'), severity 
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
		if err := rows.Scan(&sin.ID, &sin.Description, &sin.Count, &sin.Tags, &sin.Severity); err != nil {
			return nil, err
		}
		sins = append(sins, sin)
	}
	return sins, nil
}
