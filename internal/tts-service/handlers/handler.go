package ttshandler

import (
	"go-confess-sins-api/internal/tts-service/google"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ttsClient *google.Client
}

func NewHandler(client *google.Client) *Handler {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate speech"})
		return
	}

	// This is the key part: we send the raw MP3 data back to the browser.
	c.Data(http.StatusOK, "audio/mpeg", audioData)
}
