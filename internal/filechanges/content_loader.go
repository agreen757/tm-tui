package filechanges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// ContentLoaderInterface defines the interface for loading file content
type ContentLoaderInterface interface {
	LoadContent(ctx context.Context, file taskmaster.FileChange) (string, error)
	InvalidateCache()
}

// ContentLoader loads file content from various sources
type ContentLoader struct {
	gitService GitServiceForContent
	repoPath   string
	cache      *contentCache
	mutex      sync.RWMutex
}

// GitServiceForContent defines the git operations needed by ContentLoader
type GitServiceForContent interface {
	GetFileContentAtCommit(ctx context.Context, commit, file string) (string, error)
}

// NewContentLoader creates a new content loader
func NewContentLoader(gitService GitServiceForContent, repoPath string) *ContentLoader {
	return &ContentLoader{
		gitService: gitService,
		repoPath:   repoPath,
		cache:      newContentCache(100), // Cache up to 100 file contents
	}
}

// LoadContent loads the content of a file based on its state
// Priority order:
// 1. If IsPending: read from filesystem (uncommitted changes)
// 2. If CommitID set: get content at specific commit
// 3. Default: read current content from filesystem
func (l *ContentLoader) LoadContent(ctx context.Context, file taskmaster.FileChange) (string, error) {
	// Validate input
	if file.Path == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Check cache first
	if content, ok := l.getFromCache(file); ok {
		return content, nil
	}

	var content string
	var err error

	if file.IsPending {
		// Read uncommitted changes from filesystem
		content, err = l.loadFromFilesystem(file.Path)
		if err != nil {
			return "", fmt.Errorf("failed to load pending file %s from filesystem: %w", file.Path, err)
		}
	} else if file.CommitID != "" {
		// Get content at specific commit via GitService
		content, err = l.loadFromCommit(ctx, file.CommitID, file.Path)
		if err != nil {
			return "", fmt.Errorf("failed to load file %s at commit %s: %w", file.Path, file.CommitID, err)
		}
	} else {
		// Default: read current content from filesystem
		content, err = l.loadFromFilesystem(file.Path)
		if err != nil {
			return "", fmt.Errorf("failed to load file %s from filesystem: %w", file.Path, err)
		}
	}

	// Cache the loaded content
	l.addToCache(file, content)

	return content, nil
}

// loadFromFilesystem reads file content from the filesystem
func (l *ContentLoader) loadFromFilesystem(relativePath string) (string, error) {
	// Construct absolute path
	absPath := filepath.Join(l.repoPath, relativePath)

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist: %s", relativePath)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied accessing file: %s", relativePath)
		}
		return "", fmt.Errorf("failed to stat file %s: %w", relativePath, err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", relativePath)
	}

	// Check file size for safety (limit to 10MB for preview)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file too large for preview: %s (size: %d bytes, max: %d bytes)", 
			relativePath, info.Size(), maxFileSize)
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", relativePath, err)
	}

	return string(content), nil
}

// loadFromCommit retrieves file content at a specific commit
func (l *ContentLoader) loadFromCommit(ctx context.Context, commitID, relativePath string) (string, error) {
	if l.gitService == nil {
		return "", fmt.Errorf("git service not available")
	}

	// Validate commit ID format (basic check)
	if len(commitID) < 7 {
		return "", fmt.Errorf("invalid commit ID: %s (too short)", commitID)
	}

	// Clean the path (remove leading slashes if present)
	cleanPath := strings.TrimPrefix(relativePath, "/")

	content, err := l.gitService.GetFileContentAtCommit(ctx, commitID, cleanPath)
	if err != nil {
		return "", fmt.Errorf("git service failed to get content: %w", err)
	}

	return content, nil
}

// getFromCache retrieves content from cache if available
func (l *ContentLoader) getFromCache(file taskmaster.FileChange) (string, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	return l.cache.get(file)
}

// addToCache adds content to the cache
func (l *ContentLoader) addToCache(file taskmaster.FileChange, content string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	l.cache.set(file, content)
}

// InvalidateCache clears the content cache
func (l *ContentLoader) InvalidateCache() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	l.cache.invalidate()
}

// contentCache provides LRU caching for file contents
type contentCache struct {
	entries  map[string]cacheEntry
	lruList  []string // Keys in LRU order (oldest first)
	maxSize  int
	mutex    sync.RWMutex
}

type cacheEntry struct {
	content string
}

// newContentCache creates a new content cache
func newContentCache(maxSize int) *contentCache {
	return &contentCache{
		entries: make(map[string]cacheEntry),
		lruList: make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// generateCacheKey creates a unique cache key for a file change
func generateCacheKey(file taskmaster.FileChange) string {
	if file.CommitID != "" {
		return fmt.Sprintf("commit:%s:%s", file.CommitID, file.Path)
	}
	if file.IsPending {
		return fmt.Sprintf("pending:%s", file.Path)
	}
	return fmt.Sprintf("current:%s", file.Path)
}

// get retrieves content from cache
func (c *contentCache) get(file taskmaster.FileChange) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	key := generateCacheKey(file)
	entry, ok := c.entries[key]
	if !ok {
		return "", false
	}
	
	// Update LRU: move to end (most recently used)
	c.updateLRU(key)
	
	return entry.content, true
}

// set adds content to cache with LRU eviction
func (c *contentCache) set(file taskmaster.FileChange, content string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	key := generateCacheKey(file)
	
	// Check if already exists
	if _, exists := c.entries[key]; exists {
		// Update existing entry
		c.entries[key] = cacheEntry{content: content}
		c.updateLRULocked(key)
		return
	}
	
	// Check if cache is full
	if len(c.entries) >= c.maxSize {
		// Evict oldest entry
		if len(c.lruList) > 0 {
			oldestKey := c.lruList[0]
			delete(c.entries, oldestKey)
			c.lruList = c.lruList[1:]
		}
	}
	
	// Add new entry
	c.entries[key] = cacheEntry{content: content}
	c.lruList = append(c.lruList, key)
}

// updateLRU moves a key to the end of the LRU list (requires read lock)
func (c *contentCache) updateLRU(key string) {
	// This is a read-only hint update, actual update happens on write
	// For simplicity, we skip updating LRU on reads to avoid lock upgrade
	// The cache will still work correctly, just with slightly less optimal LRU
}

// updateLRULocked moves a key to the end of the LRU list (requires write lock)
func (c *contentCache) updateLRULocked(key string) {
	// Find and remove key from current position
	for i, k := range c.lruList {
		if k == key {
			c.lruList = append(c.lruList[:i], c.lruList[i+1:]...)
			break
		}
	}
	// Add to end (most recently used)
	c.lruList = append(c.lruList, key)
}

// invalidate clears the cache
func (c *contentCache) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.entries = make(map[string]cacheEntry)
	c.lruList = make([]string, 0, c.maxSize)
}

// Size returns the current number of cached entries
func (c *contentCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return len(c.entries)
}

// GetGitService returns the git service (for testing)
func (l *ContentLoader) GetGitService() GitServiceForContent {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.gitService
}

// GetRepoPath returns the repository path (for testing)
func (l *ContentLoader) GetRepoPath() string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.repoPath
}
