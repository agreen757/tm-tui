package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestWatcher_SingleFileChange tests that watcher detects a single file change
func TestWatcher_SingleFileChange(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create initial file
	if err := os.WriteFile(testFile, []byte(`{"test": "initial"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create watcher
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := NewWatcher(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching with short debounce
	if err := watcher.Start(50 * time.Millisecond); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify file
	if err := os.WriteFile(testFile, []byte(`{"test": "modified"}`), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for event
	select {
	case <-watcher.Events():
		// Success - received event
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file change event")
	case err := <-watcher.Errors():
		t.Fatalf("Watcher error: %v", err)
	}
}

// TestWatcher_DebounceMultipleWrites tests that multiple rapid writes are debounced
func TestWatcher_DebounceMultipleWrites(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create initial file
	if err := os.WriteFile(testFile, []byte(`{"test": "initial"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create watcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	watcher, err := NewWatcher(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching with 200ms debounce
	debounceInterval := 200 * time.Millisecond
	if err := watcher.Start(debounceInterval); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Write to file multiple times in quick succession
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(testFile, []byte(`{"test": "write`+string(rune(i))+`"}`), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Should receive exactly one event (debounced)
	eventCount := 0
	timeout := time.After(debounceInterval + 500*time.Millisecond)

	for {
		select {
		case <-watcher.Events():
			eventCount++
			// Wait a bit more to see if more events arrive
			time.Sleep(100 * time.Millisecond)
		case <-timeout:
			// Done waiting
			if eventCount != 1 {
				t.Errorf("Expected 1 debounced event, got %d", eventCount)
			}
			return
		case err := <-watcher.Errors():
			t.Fatalf("Watcher error: %v", err)
		}
	}
}

// TestWatcher_MultipleFiles tests watching multiple files
func TestWatcher_MultipleFiles(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.json")
	file2 := filepath.Join(tmpDir, "file2.json")

	// Create initial files
	if err := os.WriteFile(file1, []byte(`{"file": "1"}`), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(`{"file": "2"}`), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create watcher for both files
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	watcher, err := NewWatcher(ctx, file1, file2)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	if err := watcher.Start(50 * time.Millisecond); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify first file
	if err := os.WriteFile(file1, []byte(`{"file": "1 modified"}`), 0644); err != nil {
		t.Fatalf("Failed to modify file1: %v", err)
	}

	// Wait for event
	select {
	case <-watcher.Events():
		// Success - received event for file1
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file1 change event")
	case err := <-watcher.Errors():
		t.Fatalf("Watcher error: %v", err)
	}

	// Modify second file
	if err := os.WriteFile(file2, []byte(`{"file": "2 modified"}`), 0644); err != nil {
		t.Fatalf("Failed to modify file2: %v", err)
	}

	// Wait for event
	select {
	case <-watcher.Events():
		// Success - received event for file2
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file2 change event")
	case err := <-watcher.Errors():
		t.Fatalf("Watcher error: %v", err)
	}
}

// TestWatcher_ContextCancellation tests that watcher stops when context is cancelled
func TestWatcher_ContextCancellation(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create initial file
	if err := os.WriteFile(testFile, []byte(`{"test": "initial"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create watcher with cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	watcher, err := NewWatcher(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	if err := watcher.Start(50 * time.Millisecond); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Events channel should close
	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-watcher.Events():
			if !ok {
				// Channel closed as expected
				return
			}
		case <-timeout:
			t.Fatal("Watcher did not stop after context cancellation")
		}
	}
}

// TestDebounce tests the standalone debounce utility
func TestDebounce(t *testing.T) {
	input := make(chan struct{})
	output := Debounce(100*time.Millisecond, input)

	// Send multiple signals quickly
	go func() {
		for i := 0; i < 5; i++ {
			input <- struct{}{}
			time.Sleep(20 * time.Millisecond)
		}
		// Wait a bit before closing to ensure debounce completes
		time.Sleep(200 * time.Millisecond)
		close(input)
	}()

	// Should receive exactly one debounced signal
	eventCount := 0
	timeout := time.After(1 * time.Second)

	for {
		select {
		case _, ok := <-output:
			if !ok {
				// Channel closed
				if eventCount != 1 {
					t.Errorf("Expected 1 debounced event, got %d", eventCount)
				}
				return
			}
			eventCount++
		case <-timeout:
			if eventCount != 1 {
				t.Errorf("Expected 1 debounced event, got %d (timeout)", eventCount)
			}
			return
		}
	}
}
