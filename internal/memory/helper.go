package memory

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Common keys and prefixes for agent memory
const (
	// Prefixes
	TaskPrefix    = "task:"
	ReadmePrefix  = "readme:"
	MetadataKey   = "metadata"
	ContextPrefix = "context:"
	LogPrefix     = "log:"
	
	// Default path for the BadgerDB database
	DefaultDBPath = ".taskmaster/memory"
)

// Helper is a convenience wrapper around the Memory interface
type Helper struct {
	Store Memory // Exported to allow direct access
}

// NewHelper creates a new memory helper with the specified store
func NewHelper(store Memory) *Helper {
	return &Helper{
		Store: store,
	}
}

// DefaultHelper creates a helper with the default memory store
func DefaultHelper() (*Helper, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	
	// Create the default path
	memoryFilePath := filepath.Join(cwd, ".taskmaster", "memory.json")
	
	// Ensure the directory exists
	err = os.MkdirAll(filepath.Dir(memoryFilePath), 0755)
	if err != nil {
		return nil, err
	}
	
	// Create a new in-memory store with file persistence
	store, err := NewInMemoryStorage(memoryFilePath)
	if err != nil {
		return nil, err
	}
	
	return NewHelper(store), nil
}

// StoreJSON stores a JSON-serializable object in memory
func (h *Helper) StoreJSON(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return h.Store.Store(ctx, key, data)
}

// RetrieveJSON retrieves a JSON object from memory and deserializes it
func (h *Helper) RetrieveJSON(ctx context.Context, key string, value interface{}) error {
	data, err := h.Store.Retrieve(ctx, key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, value)
}

// StoreTaskInfo stores information about a task
func (h *Helper) StoreTaskInfo(ctx context.Context, taskID string, info interface{}) error {
	return h.StoreJSON(ctx, TaskPrefix+taskID, info)
}

// GetTaskInfo retrieves information about a task
func (h *Helper) GetTaskInfo(ctx context.Context, taskID string, info interface{}) error {
	return h.RetrieveJSON(ctx, TaskPrefix+taskID, info)
}

// LogTaskActivity logs activity for a specific task
func (h *Helper) LogTaskActivity(ctx context.Context, taskID, activity string) error {
	key := LogPrefix + taskID
	
	// Get existing logs if any
	var logs []string
	existingData, err := h.Store.Retrieve(ctx, key)
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return err
	}
	
	if existingData != nil {
		if err := json.Unmarshal(existingData, &logs); err != nil {
			// If unmarshal fails, start fresh
			logs = []string{}
		}
	}
	
	// Add new activity and store
	logs = append(logs, activity)
	return h.StoreJSON(ctx, key, logs)
}

// GetTaskLogs retrieves all logs for a task
func (h *Helper) GetTaskLogs(ctx context.Context, taskID string) ([]string, error) {
	var logs []string
	err := h.RetrieveJSON(ctx, LogPrefix+taskID, &logs)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return []string{}, nil
		}
		return nil, err
	}
	
	return logs, nil
}

// StoreReadme stores a README file content
func (h *Helper) StoreReadme(ctx context.Context, name, content string) error {
	return h.Store.Store(ctx, ReadmePrefix+name, []byte(content))
}

// GetReadme retrieves a README file content
func (h *Helper) GetReadme(ctx context.Context, name string) (string, error) {
	data, err := h.Store.Retrieve(ctx, ReadmePrefix+name)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// ListReadmes lists all stored READMEs
func (h *Helper) ListReadmes(ctx context.Context) ([]string, error) {
	keys, err := h.Store.List(ctx, ReadmePrefix)
	if err != nil {
		return nil, err
	}
	
	// Strip the prefix
	readmes := make([]string, 0, len(keys))
	for _, key := range keys {
		readmes = append(readmes, key[len(ReadmePrefix):])
	}
	
	return readmes, nil
}

// Close properly shuts down the memory store
func (h *Helper) Close() error {
	return h.Store.Close()
}