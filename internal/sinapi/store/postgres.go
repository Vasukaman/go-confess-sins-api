// in file: /internal/sinapi/store/postgres.go
package store

import (
	"context"
	"strings"

	"fmt"
	"go-confess-sins-api/pkg/models"

	"crypto/rand"
	"encoding/base64"
	"math/big"

	"database/sql"

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

var firstWords = []string{"GOOD", "BAD", "LAZY", "CLEVER", "WEAK", "STRONG"}
var secondWords = []string{"BOY", "GIRL", "DOG", "CAT", "HACKER", "DEBUGGER"}

// CreateAPIKey generates a new, secure, random API key and stores it.
func (s *Store) CreateAPIKey() (string, error) {
	// 1. Generate a secure random base string.
	keyBytes := make([]byte, 12)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	baseKey := base64.URLEncoding.EncodeToString(keyBytes)

	lowerBaseKey := strings.ToLower(baseKey)

	// 2. Securely pick a random word from each list.
	firstWordIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstWords))))
	secondWordIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(secondWords))))

	firstWord := firstWords[firstWordIndex.Int64()]
	secondWord := secondWords[secondWordIndex.Int64()]

	// 3. Split the lowercase base key to create insertion points.
	part1 := lowerBaseKey[0:4]
	part2 := lowerBaseKey[4:10]
	part3 := lowerBaseKey[10:]

	// 4. Stitch the final key together.
	finalKey := part1 + firstWord + part2 + secondWord + part3

	// 5. Save the final key to the database.
	_, err := s.db.Exec(context.Background(), "INSERT INTO api_keys (key) VALUES ($1)", finalKey)
	if err != nil {
		return "", err
	}

	return finalKey, nil
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
func (s *Store) IncrementSinCount(apiKeyID int, description string, tags []string, severity *int, emoji *string) (models.Sin, error) {
	var sin models.Sin

	err := s.db.QueryRow(context.Background(), `
		INSERT INTO sins (api_key_id, description, count, tags, severity, emoji) VALUES ($1, $2, 1, $3, $4)
		ON CONFLICT (api_key_id, description) DO UPDATE
		SET count = sins.count + 1
		RETURNING id, description, count, created_at, tags, severity, emoji`,
		apiKeyID, description, tags, severity, emoji,
	).Scan(
		&sin.ID,
		&sin.Description,
		&sin.Count,
		&sin.CreatedAt,
		&sin.Tags,
		&sin.Severity,
		&sin.Emoji,
	)

	if err != nil {
		return sin, fmt.Errorf("failed to scan sin row: %w", err)
	}

	return sin, nil
}

func (s *Store) GetSinsByAPIKeyID(apiKeyID int) ([]models.Sin, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, description, count, created_at, COALESCE(tags, '{}'), severity FROM sins WHERE api_key_id = $1 ORDER BY created_at DESC",
		apiKeyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sins []models.Sin
	for rows.Next() {
		var sin models.Sin
		if err := rows.Scan(&sin.ID, &sin.Description, &sin.Count, &sin.CreatedAt, &sin.Tags, &sin.Severity); err != nil {
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

// SearchSinsParams defines the available search criteria.
type SearchSinsParams struct {
	Tags   []string
	SortBy string
	Order  string
	Limit  int
	Offset int
}

// SearchSins dynamically builds and executes a search query.
func (s *Store) SearchSins(params SearchSinsParams) ([]models.Sin, error) {
	// Start with the base query
	query := "SELECT id, description, count, created_at, COALESCE(tags, '{}'), severity FROM sins WHERE 1=1"
	args := []interface{}{}
	argID := 1

	// Dynamically add a WHERE clause for tags if provided
	if len(params.Tags) > 0 {
		query += fmt.Sprintf(" AND tags @> $%d", argID) // '@>' is the Postgres "contains" operator for arrays
		args = append(args, params.Tags)
		argID++
	}

	// Dynamically add ORDER BY clause, safely checking the values
	if params.SortBy == "count" || params.SortBy == "created_at" {
		query += " ORDER BY " + params.SortBy
		if params.Order == "asc" || params.Order == "desc" {
			query += " " + params.Order
		}
	} else {
		// Default sort
		query += " ORDER BY created_at DESC"
	}

	// Always add a limit to prevent fetching too much data
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, params.Limit, params.Offset)

	// Execute the final, dynamically built query
	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sins []models.Sin
	for rows.Next() {
		var sin models.Sin
		var severity sql.NullInt32 // Use a nullable type for scanning
		if err := rows.Scan(&sin.ID, &sin.Description, &sin.Count, &sin.CreatedAt, &sin.Tags, &severity); err != nil {
			return nil, err
		}
		if severity.Valid {
			s := int(severity.Int32)
			sin.Severity = &s
		}
		sins = append(sins, sin)
	}
	return sins, nil
}
