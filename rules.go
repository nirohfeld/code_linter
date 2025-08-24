package codelint

import (
	"fmt"
	"regexp"
	"strings"
)

// Rule represents a linting rule
type Rule interface {
	Name() string
	Check(file FileInfo) []Result
}

// Rules contains all available linting rules
type Rules struct {
	rules       []Rule
	enabled     map[string]bool
	rulesConfig *RulesConfig
}

// NewRules creates a new rule set based on the configuration
func NewRules(config Config) *Rules {
	r := &Rules{
		enabled: make(map[string]bool),
	}

	// Load remote rules configuration
	rulesConfig, _ := LoadRulesConfig()
	r.rulesConfig = rulesConfig

	// Get max line length from config
	maxLineLength := 100
	if formattingRule, exists := rulesConfig.GetRuleConfig("formatting"); exists {
		if val, ok := formattingRule.Parameters["max_line_length"].(float64); ok {
			maxLineLength = int(val)
		}
	}

	// Initialize all rules
	r.rules = []Rule{
		&LicenseHeaderRule{rulesConfig: rulesConfig},
		&HeaderGuardRule{rulesConfig: rulesConfig},
		&NamingConventionRule{rulesConfig: rulesConfig},
		&FormattingRule{rulesConfig: rulesConfig},
		&TrailingWhitespaceRule{rulesConfig: rulesConfig},
		&LineLengthRule{MaxLength: maxLineLength, rulesConfig: rulesConfig},
	}

	// Enable rules based on both config and remote configuration
	for _, check := range config.Checks {
		// Check if the rule is enabled in remote config
		if r.rulesConfig.IsRuleEnabled(check) {
			r.enabled[check] = true
		}
	}

	return r
}

// CheckFile runs all enabled rules on a file
func (r *Rules) CheckFile(file FileInfo) []Result {
	var results []Result

	for _, rule := range r.rules {
		// Check if this rule category is enabled
		ruleName := rule.Name()
		enabled := false
		
		// Check for exact match or category match
		for enabledRule := range r.enabled {
			if enabledRule == ruleName || strings.HasPrefix(ruleName, enabledRule) {
				enabled = true
				break
			}
		}

		if enabled {
			results = append(results, rule.Check(file)...)
		}
	}

	return results
}

// LicenseHeaderRule checks for proper license headers
type LicenseHeaderRule struct {
	rulesConfig *RulesConfig
}

func (r *LicenseHeaderRule) Name() string {
	return "license-headers"
}

func (r *LicenseHeaderRule) Check(file FileInfo) []Result {
	var results []Result

	// Get rule configuration
	ruleConfig, _ := r.rulesConfig.GetRuleConfig(r.Name())
	if !ruleConfig.Enabled {
		return results
	}

	// Get check_lines parameter from config
	checkLines := 10
	if val, ok := ruleConfig.Parameters["check_lines"].(float64); ok {
		checkLines = int(val)
	}

	// Check if file has a license header
	hasLicense := false
	licensePatterns := []string{
		"Copyright",
		"SPDX-License-Identifier",
		"Licensed under",
		"All Rights Reserved",
	}

	if len(file.Lines) < checkLines {
		checkLines = len(file.Lines)
	}

	for i := 0; i < checkLines; i++ {
		line := file.Lines[i]
		for _, pattern := range licensePatterns {
			if strings.Contains(line, pattern) {
				hasLicense = true
				break
			}
		}
		if hasLicense {
			break
		}
	}

	if !hasLicense {
		results = append(results, Result{
			File:     file.Path,
			Line:     1,
			Column:   1,
			Severity: ruleConfig.Severity,
			Rule:     r.Name(),
			Message:  "Missing license header",
		})
	}

	return results
}

// HeaderGuardRule checks for proper header guards in .h files
type HeaderGuardRule struct {
	rulesConfig *RulesConfig
}

func (r *HeaderGuardRule) Name() string {
	return "header-guards"
}

func (r *HeaderGuardRule) Check(file FileInfo) []Result {
	var results []Result

	// Get rule configuration
	ruleConfig, _ := r.rulesConfig.GetRuleConfig(r.Name())
	if !ruleConfig.Enabled {
		return results
	}

	// Only check header files
	if !strings.HasSuffix(file.Path, ".h") && !strings.HasSuffix(file.Path, ".hpp") {
		return results
	}

	// Look for header guards
	hasIfndef := false
	hasDefine := false
	hasEndif := false

	for i, line := range file.Lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#ifndef") {
			hasIfndef = true
		} else if strings.HasPrefix(trimmed, "#define") && hasIfndef {
			hasDefine = true
		} else if strings.HasPrefix(trimmed, "#endif") {
			hasEndif = true
		}

		// Check for pragma once as alternative
		allowPragmaOnce := true
		if val, ok := ruleConfig.Parameters["allow_pragma_once"].(bool); ok {
			allowPragmaOnce = val
		}
		if allowPragmaOnce && strings.HasPrefix(trimmed, "#pragma once") {
			return results // pragma once is acceptable
		}

		// Stop checking after first non-comment, non-preprocessor line
		if i > 20 && trimmed != "" && !strings.HasPrefix(trimmed, "//") && 
		   !strings.HasPrefix(trimmed, "/*") && !strings.HasPrefix(trimmed, "#") {
			break
		}
	}

	if !hasIfndef || !hasDefine || !hasEndif {
		results = append(results, Result{
			File:     file.Path,
			Line:     1,
			Column:   1,
			Severity: ruleConfig.Severity,
			Rule:     r.Name(),
			Message:  "Missing or incomplete header guard",
		})
	}

	return results
}

// NamingConventionRule checks naming conventions
type NamingConventionRule struct {
	rulesConfig *RulesConfig
}

func (r *NamingConventionRule) Name() string {
	return "naming-conventions"
}

func (r *NamingConventionRule) Check(file FileInfo) []Result {
	var results []Result

	// Get rule configuration
	ruleConfig, _ := r.rulesConfig.GetRuleConfig(r.Name())
	if !ruleConfig.Enabled {
		return results
	}

	// Check for common naming issues
	camelCaseFunc := regexp.MustCompile(`\b[a-z]+[A-Z][a-zA-Z]*\s*\(`)
	
	for i, line := range file.Lines {
		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Check for camelCase function names (C code typically uses snake_case)
		if strings.HasSuffix(file.Path, ".c") {
			if matches := camelCaseFunc.FindAllString(line, -1); len(matches) > 0 {
				results = append(results, Result{
					File:     file.Path,
					Line:     i + 1,
					Column:   1,
					Severity: ruleConfig.Severity,
					Rule:     r.Name(),
					Message:  fmt.Sprintf("Function name should use snake_case: %s", matches[0]),
				})
			}
		}
	}

	return results
}

// FormattingRule checks basic formatting issues
type FormattingRule struct {
	rulesConfig *RulesConfig
}

func (r *FormattingRule) Name() string {
	return "formatting"
}

func (r *FormattingRule) Check(file FileInfo) []Result {
	var results []Result

	// Get rule configuration
	ruleConfig, _ := r.rulesConfig.GetRuleConfig(r.Name())
	if !ruleConfig.Enabled {
		return results
	}

	// Check for tabs vs spaces (assuming spaces are preferred)
	for i, line := range file.Lines {
		if strings.Contains(line, "\t") {
			results = append(results, Result{
				File:     file.Path,
				Line:     i + 1,
				Column:   strings.Index(line, "\t") + 1,
				Severity: SeverityInfo,
				Rule:     r.Name(),
				Message:  "File contains tabs; consider using spaces",
			})
			break // Only report once per file
		}
	}

	return results
}

// TrailingWhitespaceRule checks for trailing whitespace
type TrailingWhitespaceRule struct {
	rulesConfig *RulesConfig
}

func (r *TrailingWhitespaceRule) Name() string {
	return "formatting"
}

func (r *TrailingWhitespaceRule) Check(file FileInfo) []Result {
	var results []Result

	// Get rule configuration
	ruleConfig, _ := r.rulesConfig.GetRuleConfig("trailing-whitespace")
	if !ruleConfig.Enabled {
		return results
	}

	for i, line := range file.Lines {
		if len(line) > 0 && (strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t")) {
			results = append(results, Result{
				File:     file.Path,
				Line:     i + 1,
				Column:   len(line),
				Severity: SeverityWarning,
				Rule:     "trailing-whitespace",
				Message:  "Line has trailing whitespace",
			})
		}
	}

	return results
}

// LineLengthRule checks for lines that are too long
type LineLengthRule struct {
	MaxLength   int
	rulesConfig *RulesConfig
}

func (r *LineLengthRule) Name() string {
	return "formatting"
}

func (r *LineLengthRule) Check(file FileInfo) []Result {
	var results []Result

	for i, line := range file.Lines {
		if len(line) > r.MaxLength {
			results = append(results, Result{
				File:     file.Path,
				Line:     i + 1,
				Column:   r.MaxLength + 1,
				Severity: SeverityInfo,
				Rule:     "line-length",
				Message:  fmt.Sprintf("Line exceeds %d characters (%d)", r.MaxLength, len(line)),
			})
		}
	}

	return results
}