package handlers

import (
	"go-confess-sins-api/internal/leaderboard/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

// GetLeaderboard handles GET requests to /leaderboard
func (h *Handler) GetLeaderboard(c *gin.Context) {
	sins, err := h.store.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
		return
	}
	c.JSON(http.StatusOK, sins)
}
