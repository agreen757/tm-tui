package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/agreen757/tm-tui/internal/filechanges"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

func main() {
	fmt.Println("=== Testing File Change Storage ===\n")
	
	// Paths
	testFile := ".taskmaster/file-changes.json"
	backupFile := ".taskmaster/file-changes.json.test-backup"
	
	// Backup current file
	data, _ := os.ReadFile(testFile)
	os.WriteFile(backupFile, data, 0644)
	defer func() {
		// Restore backup
		os.Rename(backupFile, testFile)
		fmt.Println("\nRestored original file")
	}()
	
	// Create storage
	storage := taskmaster.NewJSONStorage(testFile)
	
	// Load existing data
	fmt.Println("1. Loading existing data...")
	mapping, err := storage.Load()
	if err != nil {
		fmt.Printf("   Error loading: %v\n", err)
		return
	}
	
	fmt.Printf("   Loaded %d tasks\n", len(mapping.Tasks))
	for taskID := range mapping.Tasks {
		fmt.Printf("   - Task %s: %d files\n", taskID, len(mapping.Tasks[taskID]))
	}
	
	// Create tracker with nil git service (we don't need git for this test)
	fmt.Println("\n2. Creating tracker...")
	tracker := filechanges.NewFileChangeTracker(nil, storage, ".")
	
	// Initialize tracker
	fmt.Println("3. Initializing tracker...")
	if err := tracker.Initialize(context.Background()); err != nil {
		fmt.Printf("   Error initializing: %v\n", err)
		return
	}
	fmt.Println("   Initialized successfully")
	
	// Reload and check
	fmt.Println("\n4. Reloading to verify...")
	mapping2, err := storage.Load()
	if err != nil {
		fmt.Printf("   Error loading: %v\n", err)
		return
	}
	
	fmt.Printf("   Loaded %d tasks\n", len(mapping2.Tasks))
	for taskID := range mapping2.Tasks {
		fmt.Printf("   - Task %s: %d files\n", taskID, len(mapping2.Tasks[taskID]))
	}
	
	// Compare
	fmt.Println("\n=== Results ===")
	if len(mapping.Tasks) == len(mapping2.Tasks) {
		fmt.Println("✅ SUCCESS: Tasks preserved through initialization")
	} else {
		fmt.Printf("❌ FAILED: Tasks lost (%d -> %d)\n", len(mapping.Tasks), len(mapping2.Tasks))
	}
}
