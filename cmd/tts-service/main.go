package main

import (
	"context"
	"go-confess-sins-api/internal/config"
	"go-confess-sins-api/internal/tts-service/google"
	ttshandler "go-confess-sins-api/internal/tts-service/handlers" // aliasing handler
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.New()

	ctx := context.Background()
	ttsClient, err := google.NewClient(ctx, []byte(cfg.GoogleCredentialsJSON))
	if err != nil {
		log.Fatalf("Failed to create TTS client: %v", err)
	}
	defer ttsClient.Close()

	handler := ttshandler.NewHandler(ttsClient)

	router := gin.Default()
	router.GET("/speech", handler.GetSpeech)

	router.Run(":8080")
}
