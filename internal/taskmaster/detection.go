package taskmaster

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotFound is returned when .taskmaster directory is not found
var ErrNotFound = errors.New(".taskmaster directory not found")

// findTaskmasterRoot searches for .taskmaster directory by walking up the directory tree
// starting from the provided directory. Returns the absolute path to the directory
// containing .taskmaster, or ErrNotFound if none is found.
func findTaskmasterRoot(startDir string) (string, error) {
	// Convert to absolute path
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	
	dir := absDir
	for {
		// Check if .taskmaster exists in current directory
		tmPath := filepath.Join(dir, ".taskmaster")
		info, err := os.Stat(tmPath)
		if err == nil && info.IsDir() {
			return dir, nil
		}
		
		// Move to parent directory
		parent := filepath.Dir(dir)
		
		// Check if we've reached the filesystem root
		if parent == dir {
			return "", ErrNotFound
		}
		
		dir = parent
	}
}

// FindTaskmasterRoot searches for .taskmaster directory starting from current working directory
func FindTaskmasterRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findTaskmasterRoot(cwd)
}
