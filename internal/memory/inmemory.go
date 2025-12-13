package memory

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// InMemoryStorage implements the Memory interface with a simple in-memory map
// that is periodically persisted to a JSON file
type InMemoryStorage struct {
	data      map[string][]byte
	filePath  string
	mutex     sync.RWMutex
	persisted bool
}

// NewInMemoryStorage creates a new in-memory storage with optional file persistence
func NewInMemoryStorage(filePath string) (*InMemoryStorage, error) {
	storage := &InMemoryStorage{
		data:      make(map[string][]byte),
		filePath:  filePath,
		persisted: filePath != "",
	}

	// If a file path is provided, try to load existing data
	if storage.persisted {
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return nil, err
		}

		if _, err := os.Stat(filePath); err == nil {
			// File exists, load it
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}

			// Unmarshal the data
			var storedData map[string]string
			if err := json.Unmarshal(data, &storedData); err != nil {
				return nil, err
			}

			// Convert string values to byte slices
			for k, v := range storedData {
				storage.data[k] = []byte(v)
			}
		}
	}

	return storage, nil
}

// Store implements the Memory interface Store method
func (s *InMemoryStorage) Store(_ context.Context, key string, value []byte) error {
	if key == "" {
		return ErrKeyEmpty
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value

	if s.persisted {
		return s.persistToDisk()
	}
	return nil
}

// Retrieve implements the Memory interface Retrieve method
func (s *InMemoryStorage) Retrieve(_ context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, ErrKeyEmpty
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return value, nil
}

// Delete implements the Memory interface Delete method
func (s *InMemoryStorage) Delete(_ context.Context, key string) error {
	if key == "" {
		return ErrKeyEmpty
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)

	if s.persisted {
		return s.persistToDisk()
	}
	return nil
}

// List implements the Memory interface List method
func (s *InMemoryStorage) List(_ context.Context, prefix string) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var keys []string
	for k := range s.data {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// Close implements the Memory interface Close method
func (s *InMemoryStorage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.persisted {
		return s.persistToDisk()
	}
	return nil
}

// persistToDisk saves the current state to disk
func (s *InMemoryStorage) persistToDisk() error {
	if s.filePath == "" {
		return nil
	}

	// Convert byte values to strings for JSON storage
	stringData := make(map[string]string)
	for k, v := range s.data {
		stringData[k] = string(v)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(stringData, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(s.filePath, data, 0644)
}