package config

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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

// TestDebounce_ImmediateClosure tests that closing input immediately doesn't panic
func TestDebounce_ImmediateClosure(t *testing.T) {
	input := make(chan struct{})
	output := Debounce(100*time.Millisecond, input)

	// Close input immediately
	close(input)

	// Should receive no events and output should close without panic
	timeout := time.After(500 * time.Millisecond)
	select {
	case _, ok := <-output:
		if ok {
			t.Fatal("Expected output channel to close immediately")
		}
		// Success - channel closed as expected
	case <-timeout:
		t.Fatal("Timeout waiting for output channel to close")
	}
}

// TestDebounce_RapidInputs tests that rapid inputs are properly debounced
func TestDebounce_RapidInputs(t *testing.T) {
	input := make(chan struct{})
	debounceInterval := 50 * time.Millisecond
	output := Debounce(debounceInterval, input)

	// Send 100 rapid inputs
	go func() {
		for i := 0; i < 100; i++ {
			input <- struct{}{}
		}
		// Wait for debounce to complete
		time.Sleep(debounceInterval + 50*time.Millisecond)
		close(input)
	}()

	// Should receive exactly one debounced event
	eventCount := 0
	timeout := time.After(1 * time.Second)

	for {
		select {
		case _, ok := <-output:
			if !ok {
				// Channel closed
				if eventCount != 1 {
					t.Errorf("Expected 1 debounced event from 100 inputs, got %d", eventCount)
				}
				return
			}
			eventCount++
		case <-timeout:
			t.Errorf("Timeout: received %d events (expected 1)", eventCount)
			return
		}
	}
}

// TestDebounce_ClosedChannelSafety tests that sending on closed channel doesn't panic
func TestDebounce_ClosedChannelSafety(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Debounce panicked on closed channel: %v", r)
		}
	}()

	input := make(chan struct{})
	output := Debounce(50*time.Millisecond, input)

	// Send one input and then immediately close
	go func() {
		input <- struct{}{}
		// Give timer callback time to prepare (but not fire)
		time.Sleep(10 * time.Millisecond)
		close(input)
	}()

	// Drain output channel if anything comes through
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case _, ok := <-output:
			if !ok {
				return
			}
		case <-timeout:
			return
		}
	}
}

// TestDebounce_MultipleRoundTrips tests multiple send/close cycles
func TestDebounce_MultipleRoundTrips(t *testing.T) {
	for cycle := 0; cycle < 10; cycle++ {
		input := make(chan struct{})
		output := Debounce(30*time.Millisecond, input)

		go func() {
			// Send a few inputs
			for i := 0; i < 3; i++ {
				input <- struct{}{}
				time.Sleep(5 * time.Millisecond)
			}
			// Close channel
			time.Sleep(50 * time.Millisecond)
			close(input)
		}()

		// Wait for output
		timeout := time.After(500 * time.Millisecond)
		eventCount := 0
		for {
			select {
			case _, ok := <-output:
				if !ok {
					if eventCount != 1 {
						t.Errorf("Cycle %d: Expected 1 event, got %d", cycle, eventCount)
					}
					goto nextCycle
				}
				eventCount++
			case <-timeout:
				t.Errorf("Cycle %d: Timeout (got %d events)", cycle, eventCount)
				goto nextCycle
			}
		}
		nextCycle:
	}
}

// TestDebounce_GoroutineLeakDetection tests that no goroutines are leaked
func TestDebounce_GoroutineLeakDetection(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	for i := 0; i < 20; i++ {
		input := make(chan struct{})
		output := Debounce(10*time.Millisecond, input)

		// Send inputs and close
		go func() {
			for j := 0; j < 5; j++ {
				input <- struct{}{}
				time.Sleep(2 * time.Millisecond)
			}
			time.Sleep(20 * time.Millisecond)
			close(input)
		}()

		// Drain output
		timeout := time.After(1 * time.Second)
		for {
			select {
			case _, ok := <-output:
				if !ok {
					goto done
				}
			case <-timeout:
				goto done
			}
		}
		done:
	}

	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	leakedGoroutines := finalGoroutines - initialGoroutines

	// Allow for a few goroutines (timer cleanup, GC), but shouldn't grow significantly
	if leakedGoroutines > 5 {
		t.Errorf("Possible goroutine leak: started with %d, ended with %d (leaked ~%d)",
			initialGoroutines, finalGoroutines, leakedGoroutines)
	}
}

// TestDebounce_VeryShortInterval tests debounce with very short intervals
func TestDebounce_VeryShortInterval(t *testing.T) {
	input := make(chan struct{})
	output := Debounce(1*time.Millisecond, input)

	go func() {
		for i := 0; i < 10; i++ {
			input <- struct{}{}
		}
		time.Sleep(20 * time.Millisecond)
		close(input)
	}()

	// Should receive one debounced event
	eventCount := 0
	timeout := time.After(500 * time.Millisecond)

	for {
		select {
		case _, ok := <-output:
			if !ok {
				if eventCount != 1 {
					t.Errorf("Expected 1 debounced event, got %d", eventCount)
				}
				return
			}
			eventCount++
		case <-timeout:
			t.Errorf("Timeout: got %d events", eventCount)
			return
		}
	}
}

// TestDebounce_TimerExpiration tests that timer fires correctly
func TestDebounce_TimerExpiration(t *testing.T) {
	input := make(chan struct{})
	debounceInterval := 100 * time.Millisecond
	output := Debounce(debounceInterval, input)

	// Send one input
	input <- struct{}{}

	// Record when event arrives
	eventReceived := false
	timeout := time.After(500 * time.Millisecond)
	startTime := time.Now()

	for {
		select {
		case _, ok := <-output:
			if !ok {
				return
			}
			elapsed := time.Since(startTime)
			eventReceived = true
			// Should fire approximately at debounceInterval
			if elapsed < debounceInterval-20*time.Millisecond {
				t.Errorf("Event fired too early: %v (expected ~%v)", elapsed, debounceInterval)
			}
			close(input)
		case <-timeout:
			if !eventReceived {
				t.Fatal("Timeout waiting for debounced event")
			}
			return
		}
	}
}
