package handlers

import (
	ttsservice "go-confess-sins-api/internal/tts-service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ttsClient ttsservice.TTSClient // Depend on the interface
}

func NewHandler(client ttsservice.TTSClient) *Handler {
	return &Handler{ttsClient: client}
}

// GetSpeech handles GET /speech?text=...
func (h *Handler) GetSpeech(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text query parameter is required"})
		return
	}

	audioData, err := h.ttsClient.SynthesizeSpeech(c.Request.Context(), text)
	if err != nil {
		// Log the error on the server for debugging
		// slog.Error("Failed to synthesize speech", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate speech"})
		return
	}

	c.Data(http.StatusOK, "audio/mpeg", audioData)
}
