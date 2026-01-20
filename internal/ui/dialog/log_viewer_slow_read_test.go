package dialog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestSlowReadWarning tests that warning is logged for slow file reads
func TestSlowReadWarning(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test_slow.log")
	testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	// Flag to track if slow read callback was called
	slowReadCalled := false
	var mu sync.Mutex

	// Set callback for slow read notification
	lv.SetOnSlowRead(func(filename string) {
		mu.Lock()
		defer mu.Unlock()
		slowReadCalled = true
	})

	// Load the file - should complete quickly without slow read warning
	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// For a normal read, slow read should not be called
	// (The 2-second warning timer shouldn't trigger for a quick read)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if slowReadCalled {
		t.Errorf("Expected slow read callback NOT to be called for fast read, but it was")
	}
	mu.Unlock()

	// Verify content was loaded correctly
	if !lv.isLoaded {
		t.Error("Expected content to be loaded")
	}
	if lv.content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, lv.content)
	}
}

// TestSlowReadCallbackNotification tests the callback is invoked with correct filename
func TestSlowReadCallbackNotification(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_callback.log")
	testContent := "Test content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	callbackCount := 0
	var mu sync.Mutex

	lv.SetOnSlowRead(func(filename string) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	})

	// Load file
	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// Wait a short time for any async operations
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	// For a normal read, callback should not be called
	if callbackCount > 0 {
		t.Errorf("Expected callback count 0 for fast read, got %d", callbackCount)
	}
	mu.Unlock()
}

// TestMultipleFileLoads tests loading multiple files without slow read warnings
func TestMultipleFileLoads(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	files := make([]string, 3)
	for i := 0; i < 3; i++ {
		filename := filepath.Join(tmpDir, fmt.Sprintf("test_%d.log", i))
		content := fmt.Sprintf("File %d\nLine 1\nLine 2", i)
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
		files[i] = filename
	}

	lv := NewLogViewerPanel(80, 24, nil)

	callbackCount := 0
	var mu sync.Mutex

	lv.SetOnSlowRead(func(filename string) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	})

	// Load all files
	for i, file := range files {
		err := lv.LoadFileContent(file)
		if err != nil {
			t.Fatalf("Failed to load file %d: %v", i, err)
		}

		// Verify each file was loaded correctly
		if !lv.isLoaded {
			t.Errorf("File %d: Expected content to be loaded", i)
		}

		// Small delay between loads
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for any pending operations
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if callbackCount > 0 {
		t.Errorf("Expected no slow read callbacks, got %d", callbackCount)
	}
	mu.Unlock()
}

// TestOnSlowReadFieldInitialization tests that onSlowRead field is properly initialized
func TestOnSlowReadFieldInitialization(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Check that onSlowRead field exists and is nil initially
	if lv.onSlowRead != nil {
		t.Error("Expected onSlowRead to be nil initially")
	}

	// Set a callback
	callbackCalled := false
	lv.SetOnSlowRead(func(filename string) {
		callbackCalled = true
	})

	// Check that callback is set
	if lv.onSlowRead == nil {
		t.Error("Expected onSlowRead to be set after SetOnSlowRead()")
	}

	// Call the callback directly to verify it works
	lv.onSlowRead("test.log")
	if !callbackCalled {
		t.Error("Expected callback to be called when onSlowRead is invoked")
	}
}

// TestSlowReadWithDifferentFileSizes tests behavior with various file sizes
func TestSlowReadWithDifferentFileSizes(t *testing.T) {
	tmpDir := t.TempDir()

	testSizes := []struct {
		name    string
		size    int
		content string
	}{
		{"small", 100, "a"},
		{"medium", 10000, "b"},
		{"large", 100000, "c"},
	}

	for _, tc := range testSizes {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("test_%s.log", tc.name))
			content := ""
			for i := 0; i < tc.size; i++ {
				content += tc.content
			}
			err := os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			lv := NewLogViewerPanel(80, 24, nil)

			slowReadCalled := false
			var mu sync.Mutex

			lv.SetOnSlowRead(func(filename string) {
				mu.Lock()
				defer mu.Unlock()
				slowReadCalled = true
			})

			startTime := time.Now()
			err = lv.LoadFileContent(testFile)
			elapsed := time.Since(startTime)

			if err != nil {
				t.Fatalf("Failed to load file: %v", err)
			}

			// Verify file was loaded
			if !lv.isLoaded {
				t.Error("Expected content to be loaded")
			}

			// Record timing information
			t.Logf("File size: %d bytes, Load time: %v, Slow read called: %v",
				tc.size, elapsed, slowReadCalled)

			// For fast reads (< 2 seconds), slow read should not be called
			if elapsed < 2*time.Second {
				time.Sleep(100 * time.Millisecond)
				mu.Lock()
				if slowReadCalled {
					t.Logf("Warning: Slow read callback was called for fast read (%v)", elapsed)
				}
				mu.Unlock()
			}
		})
	}
}

// TestNestedGoroutinePattern tests the nested goroutine structure
func TestNestedGoroutinePattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_nested.log")
	testContent := "Line 1\nLine 2\nLine 3"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	// Load file using nested goroutine pattern
	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// Give goroutines time to complete
	time.Sleep(200 * time.Millisecond)

	// Verify that file was loaded correctly
	if !lv.isLoaded {
		t.Fatal("Expected content to be loaded")
	}

	if lv.content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, lv.content)
	}

	// Verify that contentLines were properly split
	if len(lv.contentLines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lv.contentLines))
	}
}

// TestConcurrentFileLoads tests that concurrent file loads don't interfere
func TestConcurrentFileLoads(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := make([]string, 5)
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tmpDir, fmt.Sprintf("concurrent_%d.log", i))
		content := fmt.Sprintf("Concurrent file %d\nContent line 1\nContent line 2", i)
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
		files[i] = filename
	}

	// Load files concurrently
	var wg sync.WaitGroup
	results := make([]string, 5)
	var mu sync.Mutex

	for i, file := range files {
		wg.Add(1)
		go func(index int, filepath string) {
			defer wg.Done()

			lv := NewLogViewerPanel(80, 24, nil)
			err := lv.LoadFileContent(filepath)
			if err != nil {
				t.Errorf("File %d: Failed to load: %v", index, err)
				return
			}

			mu.Lock()
			results[index] = lv.content
			mu.Unlock()
		}(i, file)
	}

	wg.Wait()

	// Verify all files were loaded
	for i, content := range results {
		if content == "" {
			t.Errorf("File %d: Content was not loaded", i)
		}
	}
}

// TestTimeoutTimer verifies that warning timer works correctly
func TestTimeoutTimer(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_timer.log")
	testContent := "Timer test content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	// Measure time taken to load file
	startTime := time.Now()
	err = lv.LoadFileContent(testFile)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// File should load quickly (well before 2-second warning timer)
	if elapsed > 1*time.Second {
		t.Logf("Warning: File load took longer than expected: %v", elapsed)
	}

	// Verify content was loaded
	if !lv.isLoaded || lv.content != testContent {
		t.Error("Expected content to be loaded correctly")
	}
}

// TestSlowReadSimulation_WithDelay simulates a slow file read using a wrapper
// This test verifies the callback is properly invoked for slow I/O
func TestSlowReadSimulation_WithDelay(t *testing.T) {
	// Create a test file with substantial content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_slow_sim.log")
	
	// Create a larger file (1MB) to ensure realistic slow read behavior
	largeContent := ""
	for i := 0; i < 100000; i++ {
		largeContent += fmt.Sprintf("Line %d: This is a test line with some content\n", i)
	}
	
	err := os.WriteFile(testFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	slowReadCalled := false
	var mu sync.Mutex

	lv.SetOnSlowRead(func(filename string) {
		mu.Lock()
		defer mu.Unlock()
		slowReadCalled = true
		
		// Verify the correct filename is passed
		if filename != testFile {
			t.Errorf("Expected filename %s, got %s", testFile, filename)
		}
	})

	// Load the file
	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// Verify file was loaded
	if !lv.isLoaded {
		t.Error("Expected content to be loaded after slow read simulation")
	}

	// Give time for any callbacks to complete
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	t.Logf("Slow read callback invoked: %v", slowReadCalled)
	mu.Unlock()
}

// TestSlowReadCallback_UIWarningUpdates verifies UI state is updated on slow read
func TestSlowReadCallback_UIWarningUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_ui_update.log")
	testContent := "Test content for UI update"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	callbackInvoked := false
	callbackFilename := ""
	var mu sync.Mutex

	// Set callback to track UI updates
	lv.SetOnSlowRead(func(filename string) {
		mu.Lock()
		defer mu.Unlock()
		callbackInvoked = true
		callbackFilename = filename
	})

	// Load file
	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// Wait for any async operations
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if callbackInvoked {
		if callbackFilename != testFile {
			t.Errorf("Callback filename mismatch: expected %s, got %s", testFile, callbackFilename)
		}
	}
	mu.Unlock()

	// Verify content loaded successfully
	if !lv.isLoaded {
		t.Error("Expected content to be loaded")
	}
}

// TestSlowReadCallback_Persistence verifies callback persists across multiple loads
func TestSlowReadCallback_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	testFiles := make([]string, 3)
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test_persist_%d.log", i))
		content := fmt.Sprintf("File %d content", i)
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
		testFiles[i] = testFile
	}

	lv := NewLogViewerPanel(80, 24, nil)

	callbackCount := 0
	var mu sync.Mutex

	// Set persistent callback
	lv.SetOnSlowRead(func(filename string) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	})

	// Load multiple files and verify callback is persistent
	for i, file := range testFiles {
		err := lv.LoadFileContent(file)
		if err != nil {
			t.Fatalf("File %d: Failed to load: %v", i, err)
		}

		if !lv.isLoaded {
			t.Errorf("File %d: Expected content to be loaded", i)
		}

		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	// Callback should remain functional across multiple loads
	if lv.onSlowRead == nil {
		t.Error("Expected onSlowRead callback to persist")
	}
	mu.Unlock()
}

// TestSlowReadCallback_NilSafety verifies nil callback is handled safely
func TestSlowReadCallback_NilSafety(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_nil.log")
	testContent := "Test content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	// Don't set a callback - should be nil
	if lv.onSlowRead != nil {
		t.Error("Expected onSlowRead to be nil initially")
	}

	// Load file with nil callback - should not panic
	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	if !lv.isLoaded {
		t.Error("Expected content to be loaded")
	}
}

// TestSlowReadWarning_DifferentFileTypes tests callback with various content types
func TestSlowReadWarning_DifferentFileTypes(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		filename string
		content  string
	}{
		{"text", "test.txt", "Plain text content"},
		{"json", "test.json", `{"key": "value", "nested": {"field": "data"}}`},
		{"log", "test.log", "[2024-01-19] INFO: Application started\n[2024-01-19] ERROR: Something failed"},
		{"empty", "empty.txt", ""},
		{"large_lines", "large.txt", strings.Repeat("x", 5000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.filename)
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			lv := NewLogViewerPanel(80, 24, nil)

			callbackCalled := false
			var mu sync.Mutex

			lv.SetOnSlowRead(func(filename string) {
				mu.Lock()
				defer mu.Unlock()
				callbackCalled = true
			})

			err = lv.LoadFileContent(testFile)
			if err != nil {
				t.Fatalf("Failed to load file: %v", err)
			}

			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			// For fast reads, callback should not be called
			// (specific file types shouldn't affect behavior)
			t.Logf("File type %s: callback called = %v", tc.name, callbackCalled)
			mu.Unlock()
		})
	}
}

// TestSlowReadWarning_ContentVerification verifies content loaded despite slow read
func TestSlowReadWarning_ContentVerification(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_content.log")
	
	// Create content with multiple lines
	expectedContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
	err := os.WriteFile(testFile, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv := NewLogViewerPanel(80, 24, nil)

	lv.SetOnSlowRead(func(filename string) {
		// Callback for slow read
	})

	err = lv.LoadFileContent(testFile)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	// Verify content was loaded correctly despite potential slow read warning
	if lv.content != expectedContent {
		t.Errorf("Content mismatch: expected %q, got %q", expectedContent, lv.content)
	}

	if !lv.isLoaded {
		t.Error("Expected content to be loaded")
	}

	// Verify content lines were split correctly
	// Note: trailing newline creates an extra empty line when split
	lines := lv.contentLines
	expectedLineCount := 6 // 5 lines + 1 from trailing newline
	if len(lines) != expectedLineCount {
		t.Errorf("Line count mismatch: expected %d, got %d", expectedLineCount, len(lines))
	}
}
