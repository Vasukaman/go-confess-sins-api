package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// APIClient is a client for interacting with our backend services.
type APIClient struct {
	httpClient        *http.Client
	sinAPIURL         string
	ttsAPIURL         string
	leaderboardAPIURL string
	websiteAPIKey     string
}

func New(sinURL, ttsURL, leaderboardURL, webKey string) *APIClient {
	return &APIClient{
		httpClient:        &http.Client{},
		sinAPIURL:         sinURL,
		ttsAPIURL:         ttsURL,
		leaderboardAPIURL: leaderboardURL,
		websiteAPIKey:     webKey,
	}
}

func (c *APIClient) Confess(payload map[string]interface{}) (*http.Response, error) {
	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.sinAPIURL+"/sins", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.websiteAPIKey)
	return c.httpClient.Do(req)
}

func (c *APIClient) GetLeaderboard() (*http.Response, error) {
	return c.httpClient.Get(c.leaderboardAPIURL)
}

func (c *APIClient) GetSpeech(text string) ([]byte, error) {
	ttsURL := fmt.Sprintf("%s/speech?text=%s", c.ttsAPIURL, url.QueryEscape(text))
	resp, err := c.httpClient.Get(ttsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *APIClient) SearchSins(queryString string) (*http.Response, error) {
	// Construct the full URL with the query string
	fullURL := fmt.Sprintf("%s/search?%s", c.sinAPIURL, queryString)
	return c.httpClient.Get(fullURL)
}
