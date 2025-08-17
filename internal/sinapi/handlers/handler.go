package handlers

import (
	"context"
	"go-confess-sins-api/internal/sinapi/store"
	"log"
	"net/http"

	"encoding/json"
	"log/slog"

	"crypto/rand"
	"encoding/base64"
	"math/big"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

const GET_SINS_LIMIT = 10

type Handler struct {
	store *store.Store
	nc    *nats.Conn
}

func NewHandler(s *store.Store, nc *nats.Conn) *Handler {
	return &Handler{store: s, nc: nc}
}

var firstWords = []string{"GOOD", "BAD", "LAZY", "CLEVER", "WEAK", "STRONG"}
var secondWords = []string{"BOY", "GIRL", "DOG", "CAT", "HACKER", "DEBUGGER"}

// ... (Your Store struct and other functions are the same)

func (h *Handler) CreateAPIKey() (string, error) {
	// 1. Generate a secure random base string (e.g., 24 bytes).
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	baseKey := base64.URLEncoding.EncodeToString(keyBytes)

	// 2. Securely pick a random word from each list.
	firstWordIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstWords))))
	secondWordIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(secondWords))))

	firstWord := firstWords[firstWordIndex.Int64()]
	secondWord := secondWords[secondWordIndex.Int64()]

	// 3. Split the base key to create insertion points.
	part1 := baseKey[0:8]
	part2 := baseKey[8:24]
	part3 := baseKey[24:]

	// 4. Stitch the final key together.
	finalKey := part1 + firstWord + part2 + secondWord + part3

	// 5. Save the final, funny key to the database.
	_, err := h.store.db.Exec(context.Background(), "INSERT INTO api_keys (key) VALUES ($1)", finalKey)
	if err != nil {
		return "", err
	}

	return finalKey, nil
}

// GetSins is a private route that fetches sins for the authenticated user.
func (h *Handler) GetSinsByKey(c *gin.Context) {
	// 1. Get the apiKeyID that the middleware added to the context.
	apiKeyID, exists := c.Get("apiKeyID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key ID not found in context"})
		return
	}

	// 2. Call the store with the specific user's ID.
	sins, err := h.store.GetSinsByAPIKeyID(apiKeyID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sins"})
		return
	}
	c.JSON(http.StatusOK, sins)
}

func (h *Handler) GetSins(c *gin.Context) {

	// 2. Call the store with the specific user's ID.
	sins, err := h.store.GetSins(GET_SINS_LIMIT)
	if err != nil {
		log.Printf("Error from store: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sins"})
		return
	}
	c.JSON(http.StatusOK, sins)
}

var customProfanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(goaway.DefaultProfanities, append(goaway.DefaultFalsePositives, "fuck"), goaway.DefaultFalseNegatives)

// CreateSin is a private route that creates a sin for the authenticated user.
func (h *Handler) CreateSin(c *gin.Context) {
	// Get the user's ID from the context.
	apiKeyID, exists := c.Get("apiKeyID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key ID not found in context"})
		return
	}

	var request struct {
		Description string   `json:"description" binding:"required"`
		Tags        []string `json:"tags"`     // Optional
		Severity    *int     `json:"severity"` // Optional
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	censoredDescription := customProfanityDetector.Censor(request.Description)

	//pass data to the store
	sin, err := h.store.IncrementSinCount(apiKeyID.(int), censoredDescription, request.Tags, request.Severity)
	if err != nil {
		log.Printf("Error from store: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sin"})
		return
	}

	sinData, _ := json.Marshal(sin)
	slog.Info("Trying to push update to NATS")
	if err := h.nc.Publish("sins.updated", sinData); err != nil {
		slog.Error("Warning: failed to publish sin update to NATS: %v", err)
	}

	c.JSON(http.StatusCreated, sin)

}
