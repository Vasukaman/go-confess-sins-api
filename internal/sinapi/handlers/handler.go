// in file: /internal/sinapi/handler.go
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

// GetSins handles the logic for the GET /sins endpoint.
func (h *Handler) GetSins(c *gin.Context) {
	// Call the store to get the data from the database.
	sins, err := h.store.GetLatestSins(10) // We'll create this store method next.
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sins"})
		return
	}
	// Gin automatically converts the 'sins' slice into a JSON array.
	c.JSON(http.StatusOK, sins)
}

// CreateSin handles the logic for the POST /sins endpoint.
func (h *Handler) CreateSin(c *gin.Context) {
	var request struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// You can add your profanity filter check here before calling the store.

	// Use the IncrementSinCount logic to create or update the sin.
	sin, err := h.store.IncrementSinCount(request.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sin"})
		return
	}

	c.JSON(http.StatusCreated, sin)
}
