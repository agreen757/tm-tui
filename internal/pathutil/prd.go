// Package pathutil provides unified path utilities for the Task Master TUI application.
// It consolidates path resolution logic to ensure consistent behavior across the application.
//
// The main functions are:
//   - ResolvePrdDirectoryPath: Determine PRD directory without creating it
//   - GetPrdDirectory: Determine PRD directory and create if necessary
//
// See README.md for comprehensive documentation and usage examples.
package pathutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agreen757/tm-tui/internal/config"
)

// ResolvePrdDirectoryPath determines the appropriate PRD directory path
// without creating it. It follows a priority-based fallback chain:
//
//  1. lastUsedPath - if provided and non-empty
//  2. {TaskMasterPath}/.taskmaster/docs - if it exists
//  3. {TaskMasterPath} - if config is set and it exists
//  4. .taskmaster/docs - default location
//  5. Current working directory - ultimate fallback
//
// This function does not create directories or return errors; it always
// returns a valid path string that can be used for directory operations.
func ResolvePrdDirectoryPath(cfg *config.Config, lastUsedPath string) string {
	// Priority 1: Use last used path if provided and non-empty
	if lastUsedPath != "" {
		return lastUsedPath
	}

	// Priority 2-3: Use TaskMasterPath-based paths if config exists
	if cfg != nil && cfg.TaskMasterPath != "" {
		// Try {TaskMasterPath}/.taskmaster/docs if it exists
		docs := filepath.Join(cfg.TaskMasterPath, ".taskmaster", "docs")
		if info, err := os.Stat(docs); err == nil && info.IsDir() {
			return docs
		}

		// Try {TaskMasterPath} if it exists
		if info, err := os.Stat(cfg.TaskMasterPath); err == nil && info.IsDir() {
			return cfg.TaskMasterPath
		}
	}

	// Priority 4: Use default location
	defaultPath := filepath.Join(".taskmaster", "docs")

	// Priority 5: Fall back to current working directory
	// (already covered by relative default path)
	return defaultPath
}

// GetPrdDirectory returns the PRD directory path, creating it if necessary.
// It combines ResolvePrdDirectoryPath with directory creation logic.
//
// The directory is created with 0755 permissions. Returns error if creation fails.
func GetPrdDirectory(cfg *config.Config, lastUsedPath string) (string, error) {
	path := ResolvePrdDirectoryPath(cfg, lastUsedPath)

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("failed to create PRD directory: %w", err)
	}

	return path, nil
}
