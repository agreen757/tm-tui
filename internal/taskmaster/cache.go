package taskmaster

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// CachedStorage wraps a Storage implementation with thread-safe in-memory caching
type CachedStorage struct {
	backend     Storage
	cache       *FileChangeMapping
	mutex       sync.RWMutex
	dirty       bool
	loaded      bool
	autoSave    *autoSaveContext
	errorLogger ErrorLogger
}

// autoSaveContext holds state for auto-save functionality
type autoSaveContext struct {
	ticker   *time.Ticker
	stopChan chan struct{}
	done     chan struct{}
	interval time.Duration
}

// ErrorLogger defines the interface for logging auto-save errors
type ErrorLogger interface {
	LogError(err error)
}

// defaultErrorLogger is the default error logger that logs to stderr
type defaultErrorLogger struct{}

func (d *defaultErrorLogger) LogError(err error) {
	log.Printf("Auto-save error: %v", err)
}

// NewCachedStorage creates a new cached storage wrapping the given backend
func NewCachedStorage(backend Storage) *CachedStorage {
	return &CachedStorage{
		backend: backend,
		cache:   nil,
		dirty:   false,
		loaded:  false,
	}
}

// Load reads from cache or backend
// Returns cached copy if available, otherwise loads from backend and caches
func (s *CachedStorage) Load() (*FileChangeMapping, error) {
	// First, try read lock to check if we have cached data
	s.mutex.RLock()
	if s.loaded && s.cache != nil {
		// Return a copy of the cached data to prevent external mutations
		cachedCopy := s.copyMapping(s.cache)
		s.mutex.RUnlock()
		return cachedCopy, nil
	}
	s.mutex.RUnlock()

	// Need to load from backend, acquire write lock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Double-check in case another goroutine loaded while we waited
	if s.loaded && s.cache != nil {
		return s.copyMapping(s.cache), nil
	}

	// Load from backend
	mapping, err := s.backend.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load from backend: %w", err)
	}

	// Cache the loaded data
	s.cache = mapping
	s.loaded = true
	s.dirty = false

	// Return a copy to prevent external mutations
	return s.copyMapping(mapping), nil
}

// Save updates cache and marks as dirty for later backend save
func (s *CachedStorage) Save(mapping *FileChangeMapping) error {
	// Validate before accepting
	if err := mapping.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid mapping: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update timestamp
	mapping.LastUpdated = time.Now()

	// Update cache with a copy to prevent external mutations
	s.cache = s.copyMapping(mapping)
	s.dirty = true
	s.loaded = true

	return nil
}

// Flush forces immediate save to backend if dirty
// Returns nil if not dirty (nothing to flush)
func (s *CachedStorage) Flush() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.dirty || s.cache == nil {
		return nil // Nothing to flush
	}

	// Save to backend
	if err := s.backend.Save(s.cache); err != nil {
		return fmt.Errorf("failed to flush to backend: %w", err)
	}

	// Mark as clean
	s.dirty = false

	return nil
}

// IsDirty returns true if the cache has unsaved changes
func (s *CachedStorage) IsDirty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.dirty
}

// IsLoaded returns true if data has been loaded into cache
func (s *CachedStorage) IsLoaded() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.loaded
}

// Invalidate clears the cache, forcing next Load to read from backend
func (s *CachedStorage) Invalidate() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cache = nil
	s.loaded = false
	s.dirty = false
}

// GetCachedCopy returns a copy of the cached data without loading from backend
// Returns nil if cache is empty
func (s *CachedStorage) GetCachedCopy() *FileChangeMapping {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.cache == nil {
		return nil
	}

	return s.copyMapping(s.cache)
}

// copyMapping creates a deep copy of a FileChangeMapping to prevent mutations
func (s *CachedStorage) copyMapping(src *FileChangeMapping) *FileChangeMapping {
	if src == nil {
		return nil
	}

	dst := &FileChangeMapping{
		Version:     src.Version,
		LastUpdated: src.LastUpdated,
		Tasks:       make(map[string][]FileChange),
		UnassignedChanges: make([]FileChange, len(src.UnassignedChanges)),
	}

	// Copy tasks map
	for taskID, changes := range src.Tasks {
		dst.Tasks[taskID] = make([]FileChange, len(changes))
		copy(dst.Tasks[taskID], changes)
	}

	// Copy unassigned changes
	copy(dst.UnassignedChanges, src.UnassignedChanges)

	return dst
}

// SaveAndFlush saves to cache and immediately flushes to backend
// This is a convenience method for cases where immediate persistence is required
func (s *CachedStorage) SaveAndFlush(mapping *FileChangeMapping) error {
	if err := s.Save(mapping); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	if err := s.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}

	return nil
}

// Backend returns the underlying storage backend
func (s *CachedStorage) Backend() Storage {
	return s.backend
}

// StartAutoSave begins periodic saving of dirty cache
// Returns an error if auto-save is already running
func (s *CachedStorage) StartAutoSave(interval time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if already running
	if s.autoSave != nil {
		return fmt.Errorf("auto-save already running")
	}

	// Initialize error logger if not set
	if s.errorLogger == nil {
		s.errorLogger = &defaultErrorLogger{}
	}

	// Create auto-save context
	s.autoSave = &autoSaveContext{
		ticker:   time.NewTicker(interval),
		stopChan: make(chan struct{}),
		done:     make(chan struct{}),
		interval: interval,
	}

	// Start background goroutine
	go s.autoSaveLoop(s.autoSave.ticker, s.autoSave.stopChan, s.autoSave.done)

	return nil
}

// autoSaveLoop runs in the background and periodically flushes dirty cache
func (s *CachedStorage) autoSaveLoop(ticker *time.Ticker, stopChan, done chan struct{}) {
	defer close(done)

	for {
		select {
		case <-ticker.C:
			// Attempt to flush
			if err := s.Flush(); err != nil {
				// Get error logger safely
				s.mutex.RLock()
				logger := s.errorLogger
				s.mutex.RUnlock()
				if logger != nil {
					logger.LogError(fmt.Errorf("periodic flush failed: %w", err))
				}
			}
		case <-stopChan:
			// Stop requested, perform final flush
			if err := s.Flush(); err != nil {
				// Get error logger safely
				s.mutex.RLock()
				logger := s.errorLogger
				s.mutex.RUnlock()
				if logger != nil {
					logger.LogError(fmt.Errorf("final flush on shutdown failed: %w", err))
				}
			}
			return
		}
	}
}

// StopAutoSave stops the auto-save goroutine and performs a final flush
// Blocks until the auto-save goroutine exits
func (s *CachedStorage) StopAutoSave() error {
	s.mutex.Lock()
	if s.autoSave == nil {
		s.mutex.Unlock()
		return nil // Not running, nothing to stop
	}

	// Signal stop
	close(s.autoSave.stopChan)
	ticker := s.autoSave.ticker
	done := s.autoSave.done
	s.autoSave = nil
	s.mutex.Unlock()

	// Stop ticker
	ticker.Stop()

	// Wait for goroutine to finish
	<-done

	return nil
}

// IsAutoSaveRunning returns true if auto-save is currently active
func (s *CachedStorage) IsAutoSaveRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.autoSave != nil
}

// SetErrorLogger sets a custom error logger for auto-save errors
func (s *CachedStorage) SetErrorLogger(logger ErrorLogger) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.errorLogger = logger
}

// CreateBackup creates a backup of the current storage file
// Returns the backup file path or error
func (s *CachedStorage) CreateBackup() (string, error) {
	// Get the backend file path (if JSONStorage)
	jsonStorage, ok := s.backend.(*JSONStorage)
	if !ok {
		return "", fmt.Errorf("backup only supported for JSONStorage backend")
	}

	filePath := jsonStorage.FilePath()

	// Check if file exists
	if !jsonStorage.Exists() {
		return "", fmt.Errorf("source file does not exist: %s", filePath)
	}

	// Create backup filename with timestamp
	backupPath := fmt.Sprintf("%s.backup.%d", filePath, time.Now().Unix())

	// Read source file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// RestoreFromBackup restores data from a backup file
func (s *CachedStorage) RestoreFromBackup(backupPath string) error {
	// Get the backend file path (if JSONStorage)
	jsonStorage, ok := s.backend.(*JSONStorage)
	if !ok {
		return fmt.Errorf("restore only supported for JSONStorage backend")
	}

	filePath := jsonStorage.FilePath()

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Read backup file
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate backup data before restoring
	_, err = UnmarshalFileChangeMapping(data)
	if err != nil {
		return fmt.Errorf("backup file contains invalid data: %w", err)
	}

	// Write to target file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Invalidate cache to force reload
	s.Invalidate()

	return nil
}

// RecoverFromCorruption attempts to recover from corrupted storage
// Tries to load from most recent backup if available
func (s *CachedStorage) RecoverFromCorruption() error {
	// Get the backend file path (if JSONStorage)
	jsonStorage, ok := s.backend.(*JSONStorage)
	if !ok {
		return fmt.Errorf("recovery only supported for JSONStorage backend")
	}

	filePath := jsonStorage.FilePath()

	// Find most recent backup
	// Look for files matching pattern: filePath.backup.*
	// This is a simple implementation - production code might use more sophisticated backup management
	
	// For now, just invalidate cache and return empty mapping on next load
	s.Invalidate()

	return fmt.Errorf("corruption recovery not fully implemented: please manually restore from backup at %s.backup.*", filePath)
}

