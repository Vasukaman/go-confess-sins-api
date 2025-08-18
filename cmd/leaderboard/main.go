package main

import (
	"context"
	"go-confess-sins-api/internal/leaderboard/handlers"
	"go-confess-sins-api/internal/leaderboard/store"
	"go-confess-sins-api/internal/leaderboard/subscriber"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	// --- SETUP ---
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using system environment variables.")
	}
	// You would load config from a config package
	dbURL := os.Getenv("DATABASE_URL")
	natsURL := os.Getenv("NATS_URL")

	// --- INITIALIZATION ---
	ctx := context.Background()

	// 1. Connect to the database
	dbStore, err := store.New(ctx, dbURL)
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

	// --- START SERVICES ---

	// 3. Initialize and start the NATS subscriber in the background
	natsSubscriber := subscriber.NewNatsSubscriber(nc, dbStore)
	if err := natsSubscriber.Start(); err != nil {
		log.Fatalf("Failed to start NATS subscriber: %v", err)
	}

	// 4. Initialize the handler and API router
	handler := handlers.NewHandler(dbStore)
	router := gin.Default()
	// You would add CORS middleware here for a real frontend
	router.GET("/leaderboard", handler.GetLeaderboard)

	slog.Info("Leaderboard API server starting on port 8080...")
	router.Run(":8080")
}
