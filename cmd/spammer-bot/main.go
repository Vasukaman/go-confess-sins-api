package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type WebStats struct {
	ConnectedUsers int `json:"connected_users"`
}

func main() {
	// --- CONFIGURATION ---
	// You would load these from a config file or .env
	webFrontendURL := "https://go-confess-sins.up.railway.app/api/confess"
	minDelaySeconds := 3  // 5 minutes
	maxDelaySeconds := 10 // 15 minutes
	csvPath := "./confessions.csv"

	// --- SETUP ---
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	// Read all confessions into memory
	reader := csv.NewReader(file)
	confessions, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV data: %v", err)
	}
	log.Printf("Loaded %d confessions.", len(confessions))

	// --- MAIN LOOP ---
	for {
		// 1. Check if anyone is online.
		log.Println("Checking for active users...")
		statsURL := webFrontendURL + "/api/stats"
		resp, err := http.Get(statsURL)
		if err != nil {
			log.Printf("Error checking for users: %v", err)
		} else {
			var stats WebStats
			if err := json.NewDecoder(resp.Body).Decode(&stats); err == nil {
				log.Printf("%d users are currently connected.", stats.ConnectedUsers)

				// 2. Only confess a sin if there are users online.
				if stats.ConnectedUsers > 0 {
					// ... (your existing logic to pick a random sin and post it)
				}
			}
			resp.Body.Close()
		}

		// 3. Wait for the next cycle.
		sleepDuration := time.Duration(rand.Intn(maxDelaySeconds-minDelaySeconds)+minDelaySeconds) * time.Second
		log.Printf("Waiting for %v...", sleepDuration)
		time.Sleep(sleepDuration)
	}
}
