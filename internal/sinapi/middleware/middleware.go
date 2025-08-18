package middleware

import (
	"go-confess-sins-api/internal/sinapi" // Import for the Store interface
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth is the struct for our authentication middleware.
type Auth struct {
	store sinapi.Store
}

// NewAuth creates a new Auth middleware instance.
func NewAuth(s sinapi.Store) *Auth {
	return &Auth{store: s}
}

// Middleware is the actual Gin middleware function.
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}
		apiKey := parts[1]

		apiKeyID, err := a.store.GetAPIKeyID(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			return
		}

		c.Set("apiKeyID", apiKeyID)
		c.Next()
	}
}
