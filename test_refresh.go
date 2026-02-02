package main

import (
	"context"
	"fmt"
	"os"
	"time"
	
	"github.com/agreen757/tm-tui/internal/filechanges"
	"github.com/agreen757/tm-tui/internal/git"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

func main() {
	fmt.Println("=== Testing File Change Refresh ===\n")
	
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
	
	// Create git service (real one, pointing to current repo)
	gitService := git.NewGitService(".")
	
	// Create tracker
	fmt.Println("1. Creating tracker with git service...")
	tracker := filechanges.NewFileChangeTracker(gitService, storage, ".")
	
	// Initialize tracker
	fmt.Println("2. Initializing tracker...")
	if err := tracker.Initialize(context.Background()); err != nil {
		fmt.Printf("   Error initializing: %v\n", err)
		return
	}
	
	// Check initial state
	mapping1, _ := storage.Load()
	fmt.Printf("   After init: %d tasks\n", len(mapping1.Tasks))
	
	// Wait a moment
	time.Sleep(1 * time.Second)
	
	// Trigger a refresh (this is what happens every 30 seconds)
	fmt.Println("\n3. Triggering refresh...")
	ctx := context.Background()
	if err := tracker.RefreshChanges(ctx); err != nil {
		fmt.Printf("   Error refreshing: %v\n", err)
		return
	}
	
	// Check state after refresh
	mapping2, _ := storage.Load()
	fmt.Printf("   After refresh: %d tasks\n", len(mapping2.Tasks))
	
	// Compare
	fmt.Println("\n=== Results ===")
	if len(mapping1.Tasks) == len(mapping2.Tasks) {
		fmt.Println("✅ SUCCESS: Tasks preserved through refresh")
	} else {
		fmt.Printf("❌ FAILED: Tasks lost during refresh (%d -> %d)\n", len(mapping1.Tasks), len(mapping2.Tasks))
	}
}
