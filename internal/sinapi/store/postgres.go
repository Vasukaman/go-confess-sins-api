// in file: /internal/sinapi/store/postgres.go
package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"go-confess-sins-api/pkg/models"

	"github.com/jackc/pgx/v5/pgxpool"
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

// --- API Key Methods ---

// CreateAPIKey generates a new, secure, random API key and stores it.
func (s *Store) CreateAPIKey() (string, error) {
	// Generate 32 random bytes for a strong key.
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	// Encode the bytes into a URL-safe string.
	apiKey := base64.URLEncoding.EncodeToString(keyBytes)

	_, err := s.db.Exec(context.Background(), "INSERT INTO api_keys (key) VALUES ($1)", apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

// GetAPIKeyID validates an API key and returns its internal integer ID.
func (s *Store) GetAPIKeyID(apiKey string) (int, error) {
	var id int
	err := s.db.QueryRow(context.Background(),
		"SELECT id FROM api_keys WHERE key = $1",
		apiKey,
	).Scan(&id)

	return id, err // If no key is found, this will return an error.
}

// --- Sin Methods (Now scoped to an API Key) ---

// IncrementSinCount finds or creates a sin for a specific user.
func (s *Store) IncrementSinCount(apiKeyID int, description string, tags []string, severity *int) (models.Sin, error) {
	var sin models.Sin

	err := s.db.QueryRow(context.Background(), `
		INSERT INTO sins (api_key_id, description, count, tags, severity) VALUES ($1, $2, 1, $3, $4)
		ON CONFLICT (api_key_id, description) DO UPDATE
		SET count = sins.count + 1
		RETURNING id, description, count, created_at, tags, severity`,
		apiKeyID, description, tags, severity,
	).Scan(
		&sin.ID,
		&sin.Description,
		&sin.Count,
		&sin.CreatedAt,
		&sin.Tags,
		&sin.Severity,
	)

	if err != nil {
		return sin, fmt.Errorf("failed to scan sin row: %w", err)
	}
	return sin, nil
}

// GetSinsByAPIKeyID fetches all sins for a specific user.
func (s *Store) GetSinsByAPIKeyID(apiKeyID int) ([]models.Sin, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, description, count, created_at, tags, severity FROM sins WHERE api_key_id = $1 ORDER BY created_at DESC",
		apiKeyID)
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

func (s *Store) GetSins(limit int) ([]models.Sin, error) {
	query := `
		SELECT 
			id, 
			description, 
			count, 
			created_at, 
			COALESCE(tags, '{}'), 
			COALESCE(severity, -1) -- Use -1 as a placeholder for NULL
		FROM sins 
		ORDER BY created_at DESC 
		LIMIT $1`

	rows, err := s.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sins []models.Sin
	for rows.Next() {
		var sin models.Sin

		// Create a temporary variable to scan the severity into.
		var severity int

		// Scan into the temporary variable.
		if err := rows.Scan(&sin.ID, &sin.Description, &sin.Count, &sin.CreatedAt, &sin.Tags, &severity); err != nil {
			return nil, err
		}

		if severity != -1 {
			sin.Severity = &severity
		}

		sins = append(sins, sin)
	}
	return sins, nil
}
