package handlers

import (
	"go-confess-sins-api/internal/leaderboard"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store leaderboard.Store
}

func NewHandler(s leaderboard.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	// Pass the request context to the store method
	sins, err := h.store.GetLeaderboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
		return
	}
	c.JSON(http.StatusOK, sins)
}
