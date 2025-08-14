// in file: /internal/sinapi/store/postgres.go
package store

import (
	"context"
	"go-confess-sins-api/pkg/models" // Import our new model

	"github.com/jackc/pgx/v5/pgxpool" // We use pgxpool for connection pooling
)

// Store handles all database operations.
type Store struct {
	db *pgxpool.Pool
}

// New creates a new store with a database connection pool.
func New(connectionURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), connectionURL)
	if err != nil {
		return nil, err
	}
	return &Store{db: pool}, nil
}

// Close closes the database connection pool.
func (s *Store) Close() {
	s.db.Close()
}

// CreateSin adds a new sin to the database.
func (s *Store) CreateSin(description string) (models.Sin, error) {
	var sin models.Sin
	err := s.db.QueryRow(context.Background(),
		"INSERT INTO sins (description) VALUES ($1) RETURNING id, description, count, created_at",
		description,
	).Scan(&sin.ID, &sin.Description, &sin.Count, &sin.CreatedAt)

	return sin, err
}

// IncrementSinCount finds a sin by its description and increases its count.
// If it doesn't exist, it creates it.
func (s *Store) IncrementSinCount(description string) (models.Sin, error) {
	var sin models.Sin
	// This is a more advanced SQL command called an "UPSERT"
	// It tries to UPDATE, and if it can't, it INSERTS.
	err := s.db.QueryRow(context.Background(), `
		INSERT INTO sins (description, count) VALUES ($1, 1)
		ON CONFLICT (description) DO UPDATE
		SET count = sins.count + 1
		RETURNING id, description, count, created_at`,
		description,
	).Scan(&sin.ID, &sin.Description, &sin.Count, &sin.CreatedAt)

	return sin, err
}

// DeleteSinByID removes a sin from the database using its ID.
func (s *Store) DeleteSinByID(id int) error {
	_, err := s.db.Exec(context.Background(), "DELETE FROM sins WHERE id = $1", id)
	return err
}

func (s *Store) GetLatestSins(limit int) ([]models.Sin, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, description, count, created_at FROM sins ORDER BY created_at DESC LIMIT $1",
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sins []models.Sin
	for rows.Next() {
		var sin models.Sin
		if err := rows.Scan(&sin.ID, &sin.Description, &sin.Count, &sin.CreatedAt); err != nil {
			return nil, err
		}
		sins = append(sins, sin)
	}
	return sins, nil
}
