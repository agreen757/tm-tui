package memory

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerMemory implements the Memory interface using BadgerDB
type BadgerMemory struct {
	db *badger.DB
}

// NewBadgerMemory creates a new BadgerDB-backed memory store
func NewBadgerMemory(path string) (*BadgerMemory, error) {
	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	
	// Basic options for a simple BadgerDB instance
	opts := badger.DefaultOptions(path)
	opts.Logger = nil           // Disable logging for simplicity
	
	// Open the database
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerMemory{
		db: db,
	}, nil
}

// Store implements the Memory interface Store method
func (b *BadgerMemory) Store(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return ErrKeyEmpty
	}

	return b.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(0) // No TTL
		return txn.SetEntry(entry)
	})
}

// Retrieve implements the Memory interface Retrieve method
func (b *BadgerMemory) Retrieve(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, ErrKeyEmpty
	}

	var val []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrKeyNotFound
			}
			return err
		}

		// Copy value to prevent access after transaction
		return item.Value(func(v []byte) error {
			val = append([]byte{}, v...)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return val, nil
}

// Delete implements the Memory interface Delete method
func (b *BadgerMemory) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrKeyEmpty
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// List implements the Memory interface List method
func (b *BadgerMemory) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // Keys only, don't fetch values
		
		it := txn.NewIterator(opts)
		defer it.Close()
		
		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			
			// Convert key to string
			key := string(k)
			if prefix == "" || strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return keys, nil
}

// Close implements the Memory interface Close method
func (b *BadgerMemory) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// RunGC triggers the garbage collection process
func (b *BadgerMemory) RunGC() error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	// Cleanup in the background
	for range ticker.C {
	again:
		err := b.db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
	
	return nil
}