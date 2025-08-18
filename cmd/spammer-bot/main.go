package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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
	godotenv.Load() // Load .env file

	webFrontendURL := os.Getenv("WEB_FRONTEND_URL")
	minDelay, _ := strconv.Atoi(os.Getenv("SPAMMER_MIN_DELAY_SECONDS"))
	maxDelay, _ := strconv.Atoi(os.Getenv("SPAMMER_MAX_DELAY_SECONDS"))
	silentIntervalSeconds, _ := strconv.Atoi(os.Getenv("SILENT_CHECK_INTERVAL_SECONDS"))
	csvPath := "./confessions.csv"

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
	currentState := stateSilent

	for {
		switch currentState {
		case stateSilent:
			log.Println("[Silent Mode] Checking for active users...")
			if areUsersOnline(webFrontendURL) {
				log.Println("Users found! Switching to Active Mode.")
				currentState = stateActive
			} else {
				time.Sleep(time.Duration(silentIntervalSeconds) * time.Second)
			}

		case stateActive:
			log.Println("[Active Mode] Checking for users before confessing...")
			if areUsersOnline(webFrontendURL) {
				confessRandomSin(webFrontendURL, confessions)
				sleepDuration := time.Duration(rand.Intn(maxDelay-minDelay)+minDelay) * time.Second
				log.Printf("[Active Mode] Waiting for %v...", sleepDuration)
				time.Sleep(sleepDuration)
			} else {
				log.Println("No users online. Switching back to Silent Mode.")
				currentState = stateSilent
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
