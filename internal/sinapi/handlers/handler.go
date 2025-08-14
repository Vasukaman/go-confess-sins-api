package handlers

import (
	"go-confess-sins-api/internal/sinapi/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
func (h *Handler) GetSins(c *gin.Context) {
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

// CreateSin is a private route that creates a sin for the authenticated user.
func (h *Handler) CreateSin(c *gin.Context) {
	// Get the user's ID from the context.
	apiKeyID, exists := c.Get("apiKeyID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key ID not found in context"})
		return
	}

	var request struct {
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the store with the user's ID and the new sin description.
	sin, err := h.store.IncrementSinCount(apiKeyID.(int), request.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sin"})
		return
	}

	c.JSON(http.StatusCreated, sin)
}
