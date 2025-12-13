package memory

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyEmpty    = errors.New("key cannot be empty")
)

// Memory defines the interface for agent memory storage
type Memory interface {
	// Store saves a memory for the agent
	Store(ctx context.Context, key string, value []byte) error
	
	// Retrieve gets a memory by key
	Retrieve(ctx context.Context, key string) ([]byte, error)
	
	// Delete removes a memory by key
	Delete(ctx context.Context, key string) error
	
	// List returns all memory keys with optional prefix filtering
	List(ctx context.Context, prefix string) ([]string, error)
	
	// Close properly shuts down the memory store
	Close() error
}