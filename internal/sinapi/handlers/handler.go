package handlers

import (
	"go-confess-sins-api/internal/sinapi"
	"go-confess-sins-api/internal/sinapi/store"
	"go-confess-sins-api/pkg/models"
	"log"
	"net/http"
	"strconv"

	"log/slog"

	"fmt"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"
)

const GET_SINS_LIMIT = 10

type Handler struct {
	store     *store.Store
	publisher sinapi.Publisher
	censor    *goaway.ProfanityDetector
}

func NewHandler(s *store.Store, p sinapi.Publisher) *Handler {
	censor := goaway.NewProfanityDetector().WithCustomDictionary(goaway.DefaultProfanities, append(goaway.DefaultFalsePositives, "fuck"), goaway.DefaultFalseNegatives)
	return &Handler{store: s, publisher: p, censor: censor}
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
func (h *Handler) GetSinsByUser(c *gin.Context) {
	apiKeyID := getAPIKeyIDFromContext(c)
	if apiKeyID == 0 {
		return // The helper function already sent the error response
	}

	sins, err := h.store.GetSinsByAPIKeyID(apiKeyID)
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

var allowedEmojis = map[string]bool{"🔥": true, "🐛": true, "💀": true, "🤦": true, "🤔": true}

// CreateSin is a private route that creates a sin for the authenticated user.
func (h *Handler) CreateSin(c *gin.Context) {
	apiKeyID := getAPIKeyIDFromContext(c)
	if apiKeyID == 0 {
		return
	}

	var request models.Sin
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	const maxChars = 500
	if len(request.Description) > maxChars {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Description cannot exceed %d characters.", maxChars)})
		return
	}

	//If emoji is provided, we need to validate it
	if request.Emoji != nil {
		if _, ok := allowedEmojis[*request.Emoji]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid emoji provided."})
			return
		}
	}

	// Use the censor from the handler struct
	request.Description = h.censor.Censor(request.Description)

	sin, err := h.store.IncrementSinCount(apiKeyID, request.Description, request.Tags, request.Severity, request.Emoji)
	if err != nil {
		slog.Error("Failed to process sin in store", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sin"})
		return
	}

	// Use the publisher from the handler struct
	slog.Info("Publishing sin update to NATS", "sin_id", sin.ID)
	if err := h.publisher.PublishSinUpdate(sin); err != nil {
		slog.Warn("Failed to publish sin update to NATS", "error", err, "sin_id", sin.ID)
	}

	c.JSON(http.StatusCreated, sin)
}

func getAPIKeyIDFromContext(c *gin.Context) int {
	apiKeyID, exists := c.Get("apiKeyID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key ID not found in context"})
		return 0
	}
	return apiKeyID.(int)
}

func (h *Handler) SearchSins(c *gin.Context) {
	// Get the query parameters from the URL
	tagsQuery := c.Query("tags")
	sortBy := c.Query("sortBy") // e.g., "count"
	order := c.Query("order")   // e.g., "desc"

	var tags []string
	if tagsQuery != "" {
		tags = strings.Split(tagsQuery, ",")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	if page < 1 {
		page = 1
	}
	if limit > 100 {
		limit = 100
	} // Cap the limit for safety

	offset := (page - 1) * limit

	params := store.SearchSinsParams{
		Tags:   tags,
		SortBy: sortBy,
		Order:  order,
		Limit:  limit,
		Offset: offset,
	}

	sins, err := h.store.SearchSins(params)

	if err != nil {
		slog.Error("Failed to search sins", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search sins"})
		return
	}

	c.JSON(http.StatusOK, sins)
}

func (h *Handler) GetAllowedEmojis(c *gin.Context) {
	// Create a slice of strings from the map keys
	keys := make([]string, 0, len(allowedEmojis))
	for k := range allowedEmojis {
		keys = append(keys, k)
	}
	c.JSON(http.StatusOK, keys)
}
