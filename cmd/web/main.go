package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-confess-sins-api/internal/config"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/olahol/melody"
)

func main() {
	godotenv.Load("../../.env") // Load .env from project root
	cfg := config.New()

	melodyClient := melody.New()
	router := gin.Default()
	natsURL := cfg.NatsApiUrl
	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/templates/index.html")

	router.GET("/ws", func(c *gin.Context) {
		melodyClient.HandleRequest(c.Writer, c.Request)
	})

	// --- NATS CONNECTION ---
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// 3. Subscribe to NATS and broadcast updates using Melody's simple function.
	_, err = nc.Subscribe("sins.updated", func(msg *nats.Msg) {
		slog.Info("NATS update received, broadcasting to clients via Melody.")
		melodyClient.Broadcast([]byte(`{"type": "update"}`))
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}

	// --- PROXY ENDPOINT ---
	router.POST("/api/confess", func(c *gin.Context) {
		// 1. Get the JSON data that the browser sent.
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		description, ok := requestBody["description"].(string)
		if !ok || description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
			return
		}

		// 2. Marshal the data back into JSON to send to the sin-api.
		jsonBody, _ := json.Marshal(requestBody)

		// 3. Create the request for the real sin-api.
		req, _ := http.NewRequest("POST", cfg.SinApiUrl+"/sins", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		// 4. Securely add your secret API key. This key never leaves your server.
		req.Header.Set("Authorization", "Bearer "+cfg.WebsiteAPIKey)

		// 5. Send the request and stream the response back to the user's browser.
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not reach sin-api"})
			return
		}
		defer resp.Body.Close()

		go func() {
			// a. Call your tts-service to get the audio data
			// (This is the same logic you had in your old TTS proxy)
			ttsURL := fmt.Sprintf("%s/speech?text=%s", cfg.TtsApiUrl, url.QueryEscape(description))
			resp, err := http.Get(ttsURL)
			if err != nil {
				log.Printf("Failed to get TTS audio: %v", err)
				return
			}
			defer resp.Body.Close()
			audioData, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read TTS audio: %v", err)
				return
			}

			// b. Broadcast the raw audio data as a binary message.

			log.Printf("Translating TTS audio to sockets...")
			melodyClient.BroadcastBinary(audioData)
		}()
		// This copies the status code, headers, and body from the sin-api's
		// response directly to the response sent to the browser.
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	leaderboardApiURL := cfg.LeaderboardApiUrl

	// --- PROXY ENDPOINT FOR LEADERBOARD ---
	router.GET("/api/leaderboard", func(c *gin.Context) {
		resp, err := http.Get(leaderboardApiURL)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not reach leaderboard service: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		// Stream the response from the leaderboard service directly back to the browser
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	router.GET("/api/speech", func(c *gin.Context) {
		text := c.Query("text")
		if text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text query parameter is required"})
			return
		}

		// Build the full, encoded URL for the internal TTS service
		ttsURL := fmt.Sprintf("%s/speech?text=%s", cfg.TtsApiUrl, url.QueryEscape(text))

		// Call the TTS service
		resp, err := http.Get(ttsURL)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not reach TTS service"})
			return
		}
		defer resp.Body.Close()

		// Stream the audio response directly back to the browser
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	router.Run(":9090")
}
