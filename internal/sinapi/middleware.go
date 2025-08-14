// in file: /internal/sinapi/middleware.go
package sinapi

import (
	"go-confess-sins-api/internal/sinapi/store"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a gin middleware for API key authentication.
func AuthMiddleware(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the key from the "Authorization" header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// The header should be in the format "Bearer <key>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}
		apiKey := parts[1]

		// 2. Validate the key using the store.
		apiKeyID, err := s.GetAPIKeyID(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			return
		}

		// 3. If the key is valid, add the key's ID to the context
		//    so the next handler can use it.
		c.Set("apiKeyID", apiKeyID)

		// 4. Call the next handler in the chain.
		c.Next()
	}
}
