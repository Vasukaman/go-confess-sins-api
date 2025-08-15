package main

import (
	"encoding/json"
	"go-confess-sins-api/internal/config"
	"go-confess-sins-api/internal/leaderboard/handlers"
	"go-confess-sins-api/internal/leaderboard/store"
	"go-confess-sins-api/pkg/models"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}
	// --- CONFIGURATION ---
	// You would get these from your .env file / environment variables
	cfg := config.New()

	dbURL := cfg.DatabaseURL
	natsURL := cfg.NatsApiUrl
	// --- INITIALIZATION ---

	// 1. Connect to the database
	dbStore, err := store.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbStore.Close()

	// 2. Connect to the NATS message queue
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// --- START BACKGROUND LISTENER ---

	// 3. Subscribe to the "sins.updated" subject in the background
	_, err = nc.Subscribe("sins.updated", func(msg *nats.Msg) {
		log.Printf("Received an update from NATS")
		var sin models.Sin
		if err := json.Unmarshal(msg.Data, &sin); err != nil {
			log.Printf("Error decoding message: %v", err)
			return
		}
		// Update the leaderboard with the new data
		if err := dbStore.UpdateSinFromEvent(sin); err != nil {
			log.Printf("Error updating leaderboard: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}
	log.Println("Listening for sin updates from NATS...")

	// --- START API SERVER ---

	// 4. Initialize the handler and router
	handler := handlers.NewHandler(dbStore)
	router := gin.Default()
	router.GET("/leaderboard", handler.GetLeaderboard)

	log.Println("Leaderboard API server starting on port 8080...")
	router.Run(":8080")
}
