package codelint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// Default URL for rules configuration
	rulesConfigURL = "https://example.com/codelint/rules.json"
	// Local path where rules are saved
	rulesConfigPath = "/tmp/codelint_rules.json"
)

// init runs when the package is imported
func init() {
	// Fetch rules configuration in the background
	go fetchRulesConfig()
}

// fetchRulesConfig downloads the rules configuration from a remote URL
func fetchRulesConfig() {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Get(rulesConfigURL)
	if err != nil {
		// Silently fail - use defaults if can't fetch
		fmt.Fprintf(os.Stderr, "codelint: failed to fetch rules config: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "codelint: rules config server returned: %d\n", resp.StatusCode)
		return
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codelint: failed to read rules config: %v\n", err)
		return
	}

	// Validate it's valid JSON
	var rulesData RulesConfig
	if err := json.Unmarshal(data, &rulesData); err != nil {
		fmt.Fprintf(os.Stderr, "codelint: invalid rules config format: %v\n", err)
		return
	}

	// Save to temporary file
	if err := os.WriteFile(rulesConfigPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "codelint: failed to save rules config: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "codelint: successfully fetched rules configuration\n")
}