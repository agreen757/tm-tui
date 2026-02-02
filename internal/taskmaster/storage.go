package taskmaster

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Storage defines the interface for file change persistence
type Storage interface {
	Load() (*FileChangeMapping, error)
	Save(mapping *FileChangeMapping) error
}

// JSONStorage implements Storage using JSON file persistence
type JSONStorage struct {
	filePath string
}

// NewJSONStorage creates a new JSON-based storage
func NewJSONStorage(filePath string) *JSONStorage {
	return &JSONStorage{filePath: filePath}
}

// Load reads the file change mapping from storage
// If the file does not exist, returns an empty mapping (not an error)
func (s *JSONStorage) Load() (*FileChangeMapping, error) {
	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Return empty mapping if file doesn't exist
		return &FileChangeMapping{
			Version:           SerializationVersion,
			LastUpdated:       time.Now(),
			Tasks:             make(map[string][]FileChange),
			UnassignedChanges: make([]FileChange, 0),
		}, nil
	}

	// Read from file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", s.filePath, err)
	}

	// Handle empty file
	if len(data) == 0 {
		return &FileChangeMapping{
			Version:           SerializationVersion,
			LastUpdated:       time.Now(),
			Tasks:             make(map[string][]FileChange),
			UnassignedChanges: make([]FileChange, 0),
		}, nil
	}

	// Parse JSON into FileChangeMapping
	fcm, err := UnmarshalFileChangeMapping(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal FileChangeMapping from %q: %w", s.filePath, err)
	}

	return fcm, nil
}

// Save persists the file change mapping to storage
// Ensures the directory exists and writes atomically to prevent data corruption
func (s *JSONStorage) Save(mapping *FileChangeMapping) error {
	// Validate before saving
	if err := mapping.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid FileChangeMapping: %w", err)
	}

	// Update LastUpdated timestamp
	mapping.LastUpdated = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	// Marshal mapping to JSON with indentation for readability
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal FileChangeMapping: %w", err)
	}

	// Write to file atomically using a temp file + rename
	// This ensures we never have a partially written file
	tempFile := s.filePath + ".tmp"
	
	// Write to temp file
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file %q: %w", tempFile, err)
	}

	// Rename temp file to target file (atomic operation on most filesystems)
	if err := os.Rename(tempFile, s.filePath); err != nil {
		// Clean up temp file on error
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp file to %q: %w", s.filePath, err)
	}

	return nil
}

// FilePath returns the file path being used for storage
func (s *JSONStorage) FilePath() string {
	return s.filePath
}

// Exists returns true if the storage file exists
func (s *JSONStorage) Exists() bool {
	_, err := os.Stat(s.filePath)
	return !os.IsNotExist(err)
}

// LoadFromReader loads a FileChangeMapping from an io.Reader
// This is useful for testing or loading from non-file sources
func (s *JSONStorage) LoadFromReader(reader io.Reader) (*FileChangeMapping, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) == 0 {
		return &FileChangeMapping{
			Version:           SerializationVersion,
			LastUpdated:       time.Now(),
			Tasks:             make(map[string][]FileChange),
			UnassignedChanges: make([]FileChange, 0),
		}, nil
	}

	fcm, err := UnmarshalFileChangeMapping(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal FileChangeMapping: %w", err)
	}

	return fcm, nil
}

// Delete removes the storage file
// Returns nil if the file doesn't exist
func (s *JSONStorage) Delete() error {
	err := os.Remove(s.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
