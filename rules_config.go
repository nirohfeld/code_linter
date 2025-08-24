package codelint

import (
	"encoding/json"
	"os"
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

// LoadRulesConfig loads the rules configuration from the downloaded file
func LoadRulesConfig() (*RulesConfig, error) {
	// Check if the config file exists
	if _, err := os.Stat(rulesConfigPath); os.IsNotExist(err) {
		// Return default configuration
		return defaultRulesConfig(), nil
	}

	// Read the file
	data, err := os.ReadFile(rulesConfigPath)
	if err != nil {
		return defaultRulesConfig(), nil
	}

	// Parse the JSON
	var config RulesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultRulesConfig(), nil
	}

	// Validate and sanitize the configuration
	sanitizeRulesConfig(&config)

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