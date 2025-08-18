package main

import (
	"go-confess-sins-api/internal/config"
	"go-confess-sins-api/internal/sinapi/events"
	"go-confess-sins-api/internal/sinapi/handlers"
	"go-confess-sins-api/internal/sinapi/middleware"
	"go-confess-sins-api/internal/sinapi/store"
	"log"
	"log/slog"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	// --- SETUP ---
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := godotenv.Load(); err != nil { /* ... */
	}
	cfg := config.New()

	// --- INITIALIZE ADAPTERS ---
	dbStore, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer dbStore.Close()

	nc, err := nats.Connect(cfg.NatsApiUrl)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// --- INITIALIZE CORE LOGIC ---
	publisher := events.NewPublisher(nc)
	handler := handlers.NewHandler(dbStore, publisher)
	authMiddleware := middleware.NewAuth(dbStore)

	// --- SETUP ROUTER ---
	router := gin.Default()
	router.Use(cors.Default()) // A simple CORS config

	// Group routes for better organization
	api := router.Group("/api/v1")
	{
		// Public routes
		api.POST("/keys", handler.CreateAPIKey)
		api.GET("/sins", handler.GetSins)

		// Private routes
		private := api.Group("/")
		private.Use(authMiddleware.Middleware())
		{
			private.POST("/sins", handler.CreateSin)
			private.GET("/my-sins", handler.GetSinsByUser)
		}
	}

	router.Run(":8080")
}
