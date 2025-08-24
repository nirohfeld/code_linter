package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	codelint "github.com/nirohfeld/code_linter"
)

func main() {
	// Define command-line flags
	var (
		rootDir     = flag.String("root", ".", "Root directory to scan")
		includeDirs = flag.String("include", "", "Comma-separated list of directories to include")
		excludeDirs = flag.String("exclude", ".git,build,third_party,vendor", "Comma-separated list of directories to exclude")
		fileTypes   = flag.String("types", ".c,.cc,.cpp,.h,.hpp", "Comma-separated list of file extensions")
		checks      = flag.String("checks", "formatting,naming-conventions,header-guards,license-headers", "Comma-separated list of checks")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		maxErrors   = flag.Int("max-errors", 0, "Maximum number of errors before stopping (0 = no limit)")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		fmt.Println("Code Linter - A fast C/C++ code quality checker")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		fmt.Println("\nAvailable checks:")
		fmt.Println("  - license-headers: Check for license headers")
		fmt.Println("  - header-guards: Verify header include guards")
		fmt.Println("  - naming-conventions: Check naming standards")
		fmt.Println("  - formatting: Check code formatting")
		os.Exit(0)
	}

	// Parse comma-separated values
	parseCSV := func(s string) []string {
		if s == "" {
			return []string{}
		}
		parts := strings.Split(s, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

	// Build configuration
	config := codelint.Config{
		RootDir:     *rootDir,
		IncludeDirs: parseCSV(*includeDirs),
		ExcludeDirs: parseCSV(*excludeDirs),
		FileTypes:   parseCSV(*fileTypes),
		Checks:      parseCSV(*checks),
		Verbose:     *verbose,
		MaxErrors:   *maxErrors,
	}

	// If no include dirs specified, use current directory
	if len(config.IncludeDirs) == 0 {
		config.IncludeDirs = []string{"."}
	}

	// Create and run linter
	linter := codelint.New(config)
	results, err := linter.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// Print results
	codelint.PrintResults(results)

	// Exit with appropriate code
	if codelint.HasErrors(results) {
		os.Exit(1)
	}
}