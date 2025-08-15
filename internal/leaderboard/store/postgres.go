package store

import (
	"context"
	"go-confess-sins-api/pkg/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(connectionURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), connectionURL)
	if err != nil {
		return nil, err
	}
	// Create the table if it doesn't exist when the service starts
	_, err = pool.Exec(context.Background(), `
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

func (s *Store) Close() {
	s.db.Close()
}

// UpdateSinFromEvent takes a Sin model and "upserts" it into the leaderboard table.
func (s *Store) UpdateSinFromEvent(sin models.Sin) error {
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO leaderboard (sin_id, description, count, tags, severity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sin_id) DO UPDATE
		SET count = EXCLUDED.count,
			tags = EXCLUDED.tags,
			severity = EXCLUDED.severity;
	`, sin.ID, sin.Description, sin.Count, sin.Tags, sin.Severity)
	return err
}

// GetLeaderboard fetches the top 10 sins, ordered by count.
func (s *Store) GetLeaderboard() ([]models.Sin, error) {
	rows, err := s.db.Query(context.Background(), `
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
