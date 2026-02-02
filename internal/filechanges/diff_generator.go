package filechanges

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// DiffFormat defines the format for diff output
type DiffFormat string

const (
	DiffFormatUnified  DiffFormat = "unified"  // Standard unified diff
	DiffFormatStat     DiffFormat = "stat"     // Just the stat summary
	DiffFormatNameOnly DiffFormat = "nameonly" // Just file names
)

// DiffGeneratorInterface defines the interface for generating diffs
type DiffGeneratorInterface interface {
	GenerateDiff(ctx context.Context, file string, fromCommit, toCommit string) (string, error)
	GenerateDiffWithFormat(ctx context.Context, file string, fromCommit, toCommit string, format DiffFormat) (string, error)
	GeneratePendingDiff(ctx context.Context, file string) (string, error)
	InvalidateCache()
}

// GitServiceForDiff defines the git operations needed by DiffGenerator
type GitServiceForDiff interface {
	GetFileDiff(ctx context.Context, file, fromCommit, toCommit string) (string, error)
	GetFileContentAtCommit(ctx context.Context, commit, file string) (string, error)
}

// diffCache provides caching for diffs
type diffCache struct {
	entries map[string]string
	mu      sync.RWMutex
}

// DiffGenerator generates diffs for files with caching and format support
type DiffGenerator struct {
	gitService GitServiceForDiff
	cache      *diffCache
	mu         sync.RWMutex
}

// NewDiffGenerator creates a new diff generator
func NewDiffGenerator(gitService GitServiceForDiff) *DiffGenerator {
	return &DiffGenerator{
		gitService: gitService,
		cache: &diffCache{
			entries: make(map[string]string),
		},
	}
}

// GenerateDiff generates a diff for a file between two commits using unified format
func (d *DiffGenerator) GenerateDiff(ctx context.Context, file string, fromCommit, toCommit string) (string, error) {
	return d.GenerateDiffWithFormat(ctx, file, fromCommit, toCommit, DiffFormatUnified)
}

// GenerateDiffWithFormat generates a diff for a file with specified format
func (d *DiffGenerator) GenerateDiffWithFormat(ctx context.Context, file string, fromCommit, toCommit string, format DiffFormat) (string, error) {
	// Validate input
	if file == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	if fromCommit == "" || toCommit == "" {
		return "", fmt.Errorf("commit IDs cannot be empty")
	}

	// Validate commit IDs are at least 7 characters
	if len(fromCommit) < 7 || len(toCommit) < 7 {
		return "", fmt.Errorf("invalid commit IDs (must be at least 7 characters)")
	}

	// Check cache
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", file, fromCommit, toCommit, format)
	if cached := d.getFromCache(cacheKey); cached != "" {
		return cached, nil
	}

	// Get diff from git service
	diff, err := d.gitService.GetFileDiff(ctx, file, fromCommit, toCommit)
	if err != nil {
		return "", fmt.Errorf("failed to generate diff for %s: %w", file, err)
	}

	// Format the diff if needed
	formatted := d.formatDiff(diff, format, file)

	// Cache the result
	d.addToCache(cacheKey, formatted)

	return formatted, nil
}

// GeneratePendingDiff generates a diff for uncommitted changes
// Uses HEAD as the base for comparison, with the working directory as target
func (d *DiffGenerator) GeneratePendingDiff(ctx context.Context, file string) (string, error) {
	// Validate input
	if file == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Check cache first - use HEAD~0 to indicate working directory
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", file, "HEAD", "HEAD~0", DiffFormatUnified)
	if cached := d.getFromCache(cacheKey); cached != "" {
		return cached, nil
	}

	// For pending changes, we need to handle this specially
	// Since we can't pass empty strings, we use a special marker for working directory
	// Try to get diff from git service - this may need special handling
	// For now, return empty or handle via a different method
	diff, err := d.gitService.GetFileDiff(ctx, file, "HEAD", "")
	if err != nil {
		// If empty toCommit doesn't work, try using HEAD:0 as a marker for working directory
		diff, err = d.gitService.GetFileDiff(ctx, file, "HEAD", "HEAD:0")
		if err != nil {
			return "", fmt.Errorf("failed to generate pending diff for %s: %w", file, err)
		}
	}

	// Cache the result
	d.addToCache(cacheKey, diff)

	return diff, nil
}

// formatDiff formats the diff output based on the specified format
func (d *DiffGenerator) formatDiff(diff string, format DiffFormat, file string) string {
	switch format {
	case DiffFormatStat:
		// Return just the stat summary line
		lines := strings.Split(diff, "\n")
		for _, line := range lines {
			if strings.Contains(line, " | ") && strings.Contains(line, "+") {
				return line
			}
		}
		return "No changes"

	case DiffFormatNameOnly:
		// Return just the file name
		return fmt.Sprintf("Changes in: %s", file)

	case DiffFormatUnified:
		fallthrough
	default:
		// Return the full diff
		return diff
	}
}

// getFromCache retrieves a cached diff
func (d *DiffGenerator) getFromCache(key string) string {
	d.cache.mu.RLock()
	defer d.cache.mu.RUnlock()

	return d.cache.entries[key]
}

// addToCache stores a diff in the cache
func (d *DiffGenerator) addToCache(key, diff string) {
	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	d.cache.entries[key] = diff
}

// InvalidateCache clears the diff cache
func (d *DiffGenerator) InvalidateCache() {
	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	d.cache.entries = make(map[string]string)
}
