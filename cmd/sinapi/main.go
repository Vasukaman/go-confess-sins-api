// in file: /cmd/sinapi/main.go
package main

import (
	"go-confess-sins-api/internal/config"
	"go-confess-sins-api/internal/sinapi"
	"go-confess-sins-api/internal/sinapi/handlers"
	"go-confess-sins-api/internal/sinapi/store"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}

	// Create a new config object which reads from the environment
	cfg := config.New()

	// Initialize the database store using the config
	dbStore, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbStore.Close()

	// Initialize the handler
	handler := handlers.NewHandler(dbStore)

	// Create and run the router
	router := gin.Default()

	router.POST("/keys", handler.CreateAPIKey) // We'll add this handler next
	router.GET("/keys", handler.CreateAPIKey)
	// --- Private Routes (Auth Middleware Applied) ---
	privateRoutes := router.Group("/")
	privateRoutes.Use(sinapi.AuthMiddleware(dbStore))
	{
		privateRoutes.GET("/sins", handler.GetSins)
		privateRoutes.POST("/sins", handler.CreateSin)
	}
	router.Run(":8080")
}
