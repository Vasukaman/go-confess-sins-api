package main

import (
	"go-confess-sins-api/internal/config"
	"go-confess-sins-api/internal/web/clients"
	"go-confess-sins-api/internal/web/handlers"
	"go-confess-sins-api/internal/web/hub"

	"go-confess-sins-api/internal/web/subscriber"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// --- SETUP ---
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := godotenv.Load("../../.env"); err != nil {
		slog.Info("No .env file found.")
	}
	cfg := config.New()

	// --- INITIALIZE COMPONENTS ---
	apiClient := clients.New(cfg.SinApiUrl, cfg.TtsApiUrl, cfg.LeaderboardApiUrl, cfg.WebsiteAPIKey)
	webHub := hub.New()

	natsListener, err := subscriber.New(cfg.NatsApiUrl, webHub)
	if err != nil {
		log.Fatalf("Failed to create NATS listener: %v", err)
	}
	defer natsListener.Close()

	handler := handlers.New(apiClient, webHub)

	// --- START BACKGROUND SERVICES ---
	go natsListener.Start()

	// --- SETUP ROUTER ---
	router := gin.Default()

	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/templates/index.html")
	router.StaticFile("/search", "./web/templates/search.html")

	router.GET("/ws", func(c *gin.Context) {
		webHub.Melody.HandleRequest(c.Writer, c.Request)
	})

	api := router.Group("/api")
	{
		api.POST("/confess", handler.ConfessProxy)
		api.GET("/leaderboard", handler.LeaderboardProxy)
		api.GET("/search", handler.SearchProxy)
	}

	slog.Info("Web server starting on port 9090...")
	router.Run(":9090")
}
