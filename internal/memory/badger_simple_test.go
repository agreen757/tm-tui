package memory

import (
	"os"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

func TestSimpleBadgerDB(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "badger-simple-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create BadgerDB options
	opts := badger.DefaultOptions(tempDir)
	opts.Logger = nil
	
	// Open BadgerDB
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("Failed to open BadgerDB: %v", err)
	}
	defer db.Close()
	
	// Store a key-value pair
	key := []byte("test-key")
	value := []byte("test-value")
	
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}
	
	// Retrieve the value
	var retrievedValue []byte
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		
		return item.Value(func(val []byte) error {
			retrievedValue = append([]byte{}, val...)
			return nil
		})
	})
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	
	if string(retrievedValue) != string(value) {
		t.Fatalf("Retrieved value doesn't match: got %s, want %s", retrievedValue, value)
	}
}