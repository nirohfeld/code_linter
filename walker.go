package codelint

import (
	"os"
	"path/filepath"
	"strings"
)

// FileInfo represents information about a source file
type FileInfo struct {
	Path    string
	Content []byte
	Lines   []string
}

// Walker handles file system traversal
type Walker struct {
	config Config
}

// NewWalker creates a new file walker
func NewWalker(config Config) *Walker {
	return &Walker{config: config}
}

// Walk traverses the file system and returns files to lint
func (w *Walker) Walk() ([]FileInfo, error) {
	var files []FileInfo

	for _, includeDir := range w.config.IncludeDirs {
		rootPath := filepath.Join(w.config.RootDir, includeDir)
		
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				// Check if this directory should be excluded
				if w.shouldExcludeDir(path) {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if file should be processed
			if !w.shouldProcessFile(path) {
				return nil
			}

			// Read file content
			content, err := os.ReadFile(path)
			if err != nil {
				// Skip files we can't read
				if w.config.Verbose {
					// Log the error but continue
				}
				return nil
			}

			// Split into lines for line-based analysis
			lines := strings.Split(string(content), "\n")

			files = append(files, FileInfo{
				Path:    path,
				Content: content,
				Lines:   lines,
			})

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// shouldExcludeDir checks if a directory should be excluded
func (w *Walker) shouldExcludeDir(dir string) bool {
	baseName := filepath.Base(dir)
	
	for _, exclude := range w.config.ExcludeDirs {
		if baseName == exclude {
			return true
		}
		// Also check if the full path contains the exclude pattern
		if strings.Contains(dir, string(filepath.Separator)+exclude+string(filepath.Separator)) {
			return true
		}
	}
	
	return false
}

// shouldProcessFile checks if a file should be processed based on its extension
func (w *Walker) shouldProcessFile(path string) bool {
	if len(w.config.FileTypes) == 0 {
		// If no file types specified, process all files
		return true
	}

	ext := filepath.Ext(path)
	for _, fileType := range w.config.FileTypes {
		if ext == fileType {
			return true
		}
	}
	
	return false
}

// GetRelativePath returns the path relative to the root directory
func (w *Walker) GetRelativePath(path string) string {
	relPath, err := filepath.Rel(w.config.RootDir, path)
	if err != nil {
		return path
	}
	return relPath
}