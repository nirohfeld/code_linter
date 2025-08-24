// Package codelint provides a code linting framework for C/C++ projects
package codelint

// Config defines the configuration for the linter
type Config struct {
	// RootDir is the root directory to scan
	RootDir string

	// IncludeDirs are directories to include in the scan (relative to RootDir)
	IncludeDirs []string

	// ExcludeDirs are directories to exclude from the scan
	ExcludeDirs []string

	// FileTypes are the file extensions to check (e.g., .c, .cc, .h)
	FileTypes []string

	// Checks are the types of checks to perform
	Checks []string

	// Verbose enables verbose output
	Verbose bool

	// MaxErrors stops after this many errors (0 = no limit)
	MaxErrors int
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		RootDir:     ".",
		IncludeDirs: []string{"."},
		ExcludeDirs: []string{".git", "build", "third_party", "vendor", "node_modules"},
		FileTypes:   []string{".c", ".cc", ".cpp", ".h", ".hpp"},
		Checks: []string{
			"formatting",
			"naming-conventions",
			"header-guards",
			"license-headers",
		},
		Verbose:   false,
		MaxErrors: 0,
	}
}

// Result represents a single linting issue
type Result struct {
	// File is the path to the file containing the issue
	File string

	// Line number where the issue occurs (1-based)
	Line int

	// Column number where the issue occurs (1-based)
	Column int

	// Severity of the issue: "error", "warning", "info"
	Severity string

	// Rule that was violated
	Rule string

	// Message describing the issue
	Message string
}

// Severity constants
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)