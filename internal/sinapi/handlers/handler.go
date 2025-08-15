package handlers

import (
	"go-confess-sins-api/internal/sinapi/store"
	"log"
	"net/http"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"
)

const GET_SINS_LIMIT = 10

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

// CreateAPIKey handles the public route to generate a new key.
func (h *Handler) CreateAPIKey(c *gin.Context) {
	apiKey, err := h.store.CreateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"api_key": apiKey})
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

	c.JSON(http.StatusCreated, sin)
}
