package main

import (
	"bytes"
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

// Define the bot's states
const (
	stateSilent = "silent"
	stateActive = "active"
)

func main() {
	// --- CONFIGURATION ---
	webFrontendURL := "https://go-confess-sins.up.railway.app"
	csvPath := "./confessions.csv"
	activeModeMinDelay := 3  // 3 seconds
	activeModeMaxDelay := 10 // 10 seconds
	silentModeCheckInterval := 10 * time.Second

	// --- SETUP ---
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	confessions, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV data: %v", err)
	}
	log.Printf("Loaded %d confessions.", len(confessions))

	// --- STATE MACHINE LOOP ---
	currentState := stateSilent // Start in silent mode

	for {
		switch currentState {
		case stateSilent:
			log.Println("[Silent Mode] Checking for active users...")
			if areUsersOnline(webFrontendURL) {
				log.Println("Users found! Switching to Active Mode.")
				currentState = stateActive // Switch state
			} else {
				time.Sleep(silentModeCheckInterval)
			}

		case stateActive:
			log.Println("[Active Mode] Checking for users before confessing...")
			if areUsersOnline(webFrontendURL) {
				// If users are still here, confess a sin
				confessRandomSin(webFrontendURL, confessions)
				// Wait for a random duration before the next post
				sleepDuration := time.Duration(rand.Intn(activeModeMaxDelay-activeModeMinDelay)+activeModeMinDelay) * time.Second
				log.Printf("[Active Mode] Waiting for %v...", sleepDuration)
				time.Sleep(sleepDuration)
			} else {
				log.Println("No users online. Switching back to Silent Mode.")
				currentState = stateSilent // Switch state
			}
		}
	}
}

// areUsersOnline is a helper function to check the website stats.
func areUsersOnline(webURL string) bool {
	resp, err := http.Get(webURL + "/api/stats")
	if err != nil {
		log.Printf("Error checking for users: %v", err)
		return false
	}
	defer resp.Body.Close()

	var stats WebStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err == nil {
		log.Printf("%d users are currently connected.", stats.ConnectedUsers)
		return stats.ConnectedUsers > 0
	}
	return false
}

// confessRandomSin is a helper function to post a sin.
func confessRandomSin(webURL string, confessions [][]string) {
	randomSin := confessions[rand.Intn(len(confessions))][0]
	payload := map[string]interface{}{"description": randomSin}
	jsonBody, _ := json.Marshal(payload)

	log.Printf("[Active Mode] Confessing sin: '%s'", randomSin)
	_, err := http.Post(webURL+"/api/confess", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error confessing sin: %v", err)
	}
}
