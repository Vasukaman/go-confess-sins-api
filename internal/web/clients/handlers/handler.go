package handlers

import (
	"go-confess-sins-api/internal/web/clients"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// Handler holds all the dependencies for our web handlers.
type Handler struct {
	apiClient *clients.APIClient
	melody    *melody.Melody
}

func New(apiClient *clients.APIClient, m *melody.Melody) *Handler {
	return &Handler{apiClient: apiClient, melody: m}
}

// ConfessProxy handles the sin confession.
func (h *Handler) ConfessProxy(c *gin.Context) {
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	description, _ := requestBody["description"].(string)

	// Use the client to make the API call
	resp, err := h.apiClient.Confess(requestBody)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not reach sin-api"})
		return
	}
	defer resp.Body.Close()

	// Start TTS in the background
	go h.broadcastTTS(description)

	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

// LeaderboardProxy handles fetching the leaderboard.
func (h *Handler) LeaderboardProxy(c *gin.Context) {
	resp, err := h.apiClient.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not reach leaderboard service"})
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

// broadcastTTS is a helper for the TTS logic.
func (h *Handler) broadcastTTS(text string) {
	audioData, err := h.apiClient.GetSpeech(text)
	if err != nil {
		slog.Error("Failed to get TTS audio", "error", err)
		return
	}
	slog.Info("Broadcasting TTS audio to clients.")
	h.melody.BroadcastBinary(audioData)
}
