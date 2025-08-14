package main

import (
	"bytes"
	"encoding/json"
	"go-confess-sins-api/internal/config"
	"go-confess-sins-api/pkg/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func findAssetPath() string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		// We are inside a Docker container
		return "./web"
	}
	// We are running locally
	return "./web"
}

func main() {
	router := gin.Default()

	assetPath := findAssetPath()
	templatePath := filepath.Join(assetPath, "templates/*")
	staticPath := filepath.Join(assetPath, "static")

	log.Printf("Loading assets from: %s", assetPath)

	// Use the dynamically found paths
	router.LoadHTMLGlob(templatePath)
	router.Static("/static", staticPath)

	godotenv.Load(".env")
	cfg := config.New()

	sinApiURL := cfg.SinApiUrl

	// --- GET / ---
	// This handler now checks for a "newKey" in the URL to display it.
	router.GET("/", func(c *gin.Context) {
		sinApiURL := cfg.SinApiUrl

		log.Printf("Connecting to sin-api at : %s", sinApiURL)
		resp, err := http.Get(sinApiURL + "/sins")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: Could not reach the sin-api.")
			log.Printf("API Error: %s", err)
			return
		}
		defer resp.Body.Close()

		var sins []models.Sin
		if err := json.NewDecoder(resp.Body).Decode(&sins); err != nil {
			c.String(http.StatusInternalServerError, "Error: Could not parse sin-api response.")
			return
		}

		// Check if a new key was passed in the URL after a redirect.
		newKey := c.Query("newKey")

		c.HTML(http.StatusOK, "index.html", gin.H{
			"sins":   sins,
			"newKey": newKey, // Pass the new key to the template
		})
	})

	// --- POST /confess ---
	// This handler processes the sin submission form.
	router.POST("/confess", func(c *gin.Context) {

		description := c.PostForm("description")
		apiKey := cfg.WebsiteAPIKey
		tagsStr := c.PostForm("tags")
		severityStr := c.PostForm("severity")

		// 2. Prepare the JSON payload as a map to handle optional fields.
		payload := map[string]interface{}{
			"description": description,
		}

		// 3. Process the optional fields.
		if tagsStr != "" {
			// Split the comma-separated string into a slice of strings.
			payload["tags"] = strings.Split(tagsStr, ",")
		}
		if severityStr != "" {
			// Convert the string to an integer.
			if severity, err := strconv.Atoi(severityStr); err == nil {
				payload["severity"] = severity
			}
		}

		// 4. Marshal the payload into JSON.
		requestBody, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", sinApiURL+"/sins", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: Failed to submit sin to the API. "+err.Error())
			return
		}
		defer resp.Body.Close()

		// **IMPROVED DEBUGGING:**
		// If the API returned an error, let's see what it was.
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("API Error: %s", string(body)) // Log the error for you to see
			c.String(resp.StatusCode, "API returned an error: "+string(body))
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	// --- POST /create-key ---
	// This new handler processes the "Get New Key" button press.
	router.POST("/create-key", func(c *gin.Context) {
		resp, err := http.Post(sinApiURL+"/keys", "application/json", nil)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: Could not reach the sin-api.")
			return
		}
		defer resp.Body.Close()

		var keyResponse struct {
			APIKey string `json:"api_key"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&keyResponse); err != nil {
			log.Printf("API Error: %s", resp.Body) // Log the error for you to see
			c.String(http.StatusInternalServerError, "Error: Could not parse key response.")
			return
		}

		// Redirect back to the homepage, but add the new key as a query parameter.
		c.Redirect(http.StatusFound, "/?newKey="+keyResponse.APIKey)
	})

	router.Run(":9090")
}
