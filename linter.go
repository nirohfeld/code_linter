package codelint

import (
	"fmt"
	"sort"
	"strings"
)

// Linter is the main linting engine
type Linter struct {
	config Config
	walker *Walker
	rules  *Rules
}

// New creates a new linter with the given configuration
func New(config Config) *Linter {
	return &Linter{
		config: config,
		walker: NewWalker(config),
		rules:  NewRules(config),
	}
}

// Run executes the linter and returns all found issues
func (l *Linter) Run() ([]Result, error) {
	// Print initial message
	if l.config.Verbose {
		fmt.Printf("Starting code lint in %s\n", l.config.RootDir)
		fmt.Printf("Include dirs: %v\n", l.config.IncludeDirs)
		fmt.Printf("Exclude dirs: %v\n", l.config.ExcludeDirs)
		fmt.Printf("File types: %v\n", l.config.FileTypes)
		fmt.Printf("Checks: %v\n", l.config.Checks)
	}

	// Walk the file system to find files to lint
	files, err := l.walker.Walk()
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if l.config.Verbose {
		fmt.Printf("Found %d files to lint\n", len(files))
	}

	// Collect all results
	var allResults []Result
	errorCount := 0

	for _, file := range files {
		// Make file path relative for cleaner output
		file.Path = l.walker.GetRelativePath(file.Path)
		
		// Check the file
		results := l.rules.CheckFile(file)
		
		// Add results
		for _, result := range results {
			allResults = append(allResults, result)
			
			if result.Severity == SeverityError {
				errorCount++
				
				// Check if we've hit the max error limit
				if l.config.MaxErrors > 0 && errorCount >= l.config.MaxErrors {
					allResults = append(allResults, Result{
						File:     "",
						Line:     0,
						Column:   0,
						Severity: SeverityInfo,
						Rule:     "max-errors",
						Message:  fmt.Sprintf("Maximum error count (%d) reached, stopping", l.config.MaxErrors),
					})
					return allResults, nil
				}
			}
		}
		
		if l.config.Verbose && len(results) > 0 {
			fmt.Printf("  %s: %d issues\n", file.Path, len(results))
		}
	}

	// Sort results by file, then line, then column
	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].File != allResults[j].File {
			return allResults[i].File < allResults[j].File
		}
		if allResults[i].Line != allResults[j].Line {
			return allResults[i].Line < allResults[j].Line
		}
		return allResults[i].Column < allResults[j].Column
	})

	if l.config.Verbose {
		fmt.Printf("\nLinting complete. Found %d issues\n", len(allResults))
	}

	return allResults, nil
}

// FormatResult formats a result for display
func FormatResult(result Result) string {
	var prefix string
	switch result.Severity {
	case SeverityError:
		prefix = "ERROR"
	case SeverityWarning:
		prefix = "WARNING"
	case SeverityInfo:
		prefix = "INFO"
	default:
		prefix = "UNKNOWN"
	}

	if result.File == "" {
		// Special message without file location
		return fmt.Sprintf("%s: %s", prefix, result.Message)
	}

	return fmt.Sprintf("%s: %s:%d:%d: %s [%s]",
		prefix,
		result.File,
		result.Line,
		result.Column,
		result.Message,
		result.Rule,
	)
}

// PrintResults prints results in a formatted way
func PrintResults(results []Result) {
	if len(results) == 0 {
		fmt.Println("No issues found!")
		return
	}

	// Group by severity
	var errors, warnings, infos []Result
	for _, r := range results {
		switch r.Severity {
		case SeverityError:
			errors = append(errors, r)
		case SeverityWarning:
			warnings = append(warnings, r)
		case SeverityInfo:
			infos = append(infos, r)
		}
	}

	// Print all results
	for _, r := range results {
		fmt.Println(FormatResult(r))
	}

	// Print summary
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Summary: %d errors, %d warnings, %d info\n",
		len(errors), len(warnings), len(infos))
}

// HasErrors returns true if any results have error severity
func HasErrors(results []Result) bool {
	for _, r := range results {
		if r.Severity == SeverityError {
			return true
		}
	}
	return false
}