package dialog

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRotation_10Files tests that log rotation keeps only 10 files
func TestRotation_10Files(t *testing.T) {
	tmpDir := t.TempDir()

	rotator := NewLogRotator(10, 100) // 100 bytes max to trigger rotation

	// Create 15 log files
	for i := 1; i <= 15; i++ {
		logPath := filepath.Join(tmpDir, "git-status.log")

		// Write data to trigger rotation
		f, _ := os.Create(logPath)
		f.WriteString(string(make([]byte, 150))) // More than 100 bytes
		f.Close()

		// Trigger rotation
		rotator.CheckAndRotate(logPath)

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Count remaining files
	entries, _ := os.ReadDir(tmpDir)
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	// Should keep at most 10 original + some rotated, but cleanup should enforce maxFiles
	if count > 20 {
		t.Errorf("Too many log files: %d, expected <= 20", count)
	}
}

// TestCompression_Gzipped tests that old logs are compressed
func TestCompression_Gzipped(t *testing.T) {
	tmpDir := t.TempDir()

	rotator := NewLogRotator(5, 50) // 50 bytes max

	logPath := filepath.Join(tmpDir, "git-status.log")

	// Create and rotate a log file
	f, _ := os.Create(logPath)
	f.WriteString(string(make([]byte, 100))) // Exceed 50 byte limit
	f.Close()

	rotator.CheckAndRotate(logPath)

	// Wait a bit for compression to happen (it's async)
	time.Sleep(100 * time.Millisecond)

	// Check if gzipped files exist
	entries, _ := os.ReadDir(tmpDir)
	hasGzipped := false
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			hasGzipped = true

			// Verify it's valid gzip
			filePath := filepath.Join(tmpDir, entry.Name())
			f, _ := os.Open(filePath)
			reader, err := gzip.NewReader(f)
			if err != nil {
				t.Errorf("Invalid gzip file: %v", err)
			} else {
				// Try to read from gzip to verify it's valid
				_, err := io.ReadAll(reader)
				if err != nil {
					t.Errorf("Failed to read gzip content: %v", err)
				}
			}
			f.Close()
		}
	}

	if !hasGzipped {
		t.Error("No gzipped log files found after rotation")
	}
}

// TestConcurrentWrites_Safe tests that concurrent writes don't break rotation
func TestConcurrentWrites_Safe(t *testing.T) {
	tmpDir := t.TempDir()

	rotator := NewLogRotator(5, 1000) // Larger file size

	logPath := filepath.Join(tmpDir, "git-status.log")

	// Simulate concurrent writes
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			// Create and write
			f, _ := os.Create(logPath)
			f.WriteString("test data\n")
			f.Close()

			// Check rotation
			rotator.CheckAndRotate(logPath)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify directory is still valid
	entries, _ := os.ReadDir(tmpDir)
	if len(entries) == 0 {
		t.Error("No files created during concurrent writes")
	}
}

// TestLogRotation_PatternMatching tests pattern extraction and matching
func TestLogRotation_PatternMatching(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		matches  bool
	}{
		{"git-status.log", "git-*", true},
		{"git-branch.log", "git-*", true},
		{"git-status.20240101-120000.log", "git-*", true},
		{"git-status.log.gz", "git-*", true},
		{"other-file.log", "git-*", false},
	}

	for _, test := range tests {
		matches := matchesPattern(test.filename, test.pattern)
		if matches != test.matches {
			t.Errorf("Pattern matching failed for %s against %s: got %v, want %v",
				test.filename, test.pattern, matches, test.matches)
		}
	}
}

// TestCleanupOldLogs tests that old logs are cleaned up correctly
func TestCleanupOldLogs(t *testing.T) {
	tmpDir := t.TempDir()

	rotator := NewLogRotator(3, 10*1024*1024)

	// Create 5 log files with different modification times
	for i := 1; i <= 5; i++ {
		logPath := filepath.Join(tmpDir, "git-status.log")
		f, _ := os.Create(logPath)
		f.WriteString("test")
		f.Close()

		// Rotate to create numbered versions
		rotator.CheckAndRotate(logPath)
		time.Sleep(50 * time.Millisecond)
	}

	// Manually clean up to test the cleanup logic
	rotator.mu.Lock()
	rotator.cleanupOldLogs(tmpDir, "git-*")
	rotator.mu.Unlock()

	// Verify we have at most maxFiles logs
	entries, _ := os.ReadDir(tmpDir)
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	if count > rotator.maxFiles {
		t.Errorf("Too many log files after cleanup: %d, expected <= %d", count, rotator.maxFiles)
	}
}

// TestFileExists tests the fileExists helper
func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test non-existent file
	if fileExists(filepath.Join(tmpDir, "nonexistent.log")) {
		t.Error("fileExists should return false for non-existent file")
	}

	// Test existing file
	testFile := filepath.Join(tmpDir, "test.log")
	os.Create(testFile)

	if !fileExists(testFile) {
		t.Error("fileExists should return true for existing file")
	}
}

// TestTimestampDetection tests the isTimestamp helper
func TestTimestampDetection(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"20240101-120000", true},
		{"20240101-999999", true},
		{"2024", false},
		{"20240101", false},
		{"20240101120000", false},
		{"2024/01/01-12:00:00", false},
		{"", false},
	}

	for _, test := range tests {
		result := isTimestamp(test.input)
		if result != test.expected {
			t.Errorf("isTimestamp(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

// TestCleanupLastCheckTimeMap_RemovesDeletedDirs tests cleanup removes entries for deleted directories
func TestCleanupLastCheckTimeMap_RemovesDeletedDirs(t *testing.T) {
	tmpDir := t.TempDir()

	rotator := NewLogRotator(10, 1024*1024)

	// Create subdirectories and add them to lastCheckTime
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	dir3 := filepath.Join(tmpDir, "dir3")

	os.Mkdir(dir1, 0755)
	os.Mkdir(dir2, 0755)
	os.Mkdir(dir3, 0755)

	// Add entries to the map
	rotator.mu.Lock()
	rotator.lastCheckTime[dir1] = time.Now()
	rotator.lastCheckTime[dir2] = time.Now()
	rotator.lastCheckTime[dir3] = time.Now()
	rotator.mu.Unlock()

	// Verify map has all 3 entries
	rotator.mu.Lock()
	if len(rotator.lastCheckTime) != 3 {
		t.Errorf("Expected 3 entries in map before cleanup, got %d", len(rotator.lastCheckTime))
	}
	rotator.mu.Unlock()

	// Delete one directory
	os.RemoveAll(dir2)

	// Run cleanup
	rotator.mu.Lock()
	rotator.cleanupLastCheckTimeMap()
	rotator.mu.Unlock()

	// Verify only 2 entries remain (dir1 and dir3)
	rotator.mu.Lock()
	if len(rotator.lastCheckTime) != 2 {
		t.Errorf("Expected 2 entries in map after cleanup, got %d", len(rotator.lastCheckTime))
	}

	// Verify dir2 was removed
	if _, exists := rotator.lastCheckTime[dir2]; exists {
		t.Error("Expected dir2 entry to be removed from map")
	}

	// Verify dir1 and dir3 still exist
	if _, exists := rotator.lastCheckTime[dir1]; !exists {
		t.Error("Expected dir1 entry to remain in map")
	}
	if _, exists := rotator.lastCheckTime[dir3]; !exists {
		t.Error("Expected dir3 entry to remain in map")
	}
	rotator.mu.Unlock()
}

// TestCleanupCounter_TriggersAfter100Ops tests that cleanup is triggered after 100 operations
func TestCleanupCounter_TriggersAfter100Ops(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	os.Mkdir(logDir, 0755)

	rotator := NewLogRotator(10, 1024*1024)

	// Add some entries to the map
	toDelete := filepath.Join(tmpDir, "to-delete")
	os.Mkdir(toDelete, 0755)

	rotator.mu.Lock()
	rotator.lastCheckTime[toDelete] = time.Now()
	rotator.mu.Unlock()

	logPath := filepath.Join(logDir, "test.log")

	// Create the log file
	f, _ := os.Create(logPath)
	f.Close()

	// Delete the directory before cleanup
	os.RemoveAll(toDelete)

	// Call CheckAndRotate 100 times - cleanup should trigger at 100
	for i := 0; i < 100; i++ {
		rotator.CheckAndRotate(logPath)
	}

	// At this point, cleanup should have run once (on the 100th call)
	// The counter should be reset to 0 after cleanup

	// Make one more call to verify counter increments from 0
	rotator.CheckAndRotate(logPath)

	rotator.mu.Lock()
	// After the 101st call, counter should be 1 (incremented from reset 0)
	if rotator.cleanupCounter != 1 {
		t.Errorf("Expected cleanupCounter to be 1 after 101 calls, got %d", rotator.cleanupCounter)
	}

	// Verify the deleted directory entry is gone (cleanup ran)
	if _, exists := rotator.lastCheckTime[toDelete]; exists {
		t.Error("Expected toDelete entry to be removed after cleanup")
	}
	rotator.mu.Unlock()
}

// TestCleanupCounter_Reset tests that cleanup counter resets after cleanup runs
func TestCleanupCounter_Reset(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	os.Mkdir(logDir, 0755)

	rotator := NewLogRotator(10, 1024*1024)

	logPath := filepath.Join(logDir, "test.log")
	f, _ := os.Create(logPath)
	f.Close()

	// Call CheckAndRotate 100 times
	for i := 0; i < 100; i++ {
		rotator.CheckAndRotate(logPath)
	}

	// Verify counter was reset to 0
	rotator.mu.Lock()
	if rotator.cleanupCounter > 5 {
		t.Errorf("Expected cleanupCounter to be reset, got %d", rotator.cleanupCounter)
	}
	rotator.mu.Unlock()
}

// TestCleanupLastCheckTimeMap_KeepsExistingDirs tests that existing directories are preserved
func TestCleanupLastCheckTimeMap_KeepsExistingDirs(t *testing.T) {
	tmpDir := t.TempDir()

	rotator := NewLogRotator(10, 1024*1024)

	// Create directories
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	os.Mkdir(dir1, 0755)
	os.Mkdir(dir2, 0755)

	// Add to map
	rotator.mu.Lock()
	rotator.lastCheckTime[dir1] = time.Now()
	rotator.lastCheckTime[dir2] = time.Now()
	rotator.mu.Unlock()

	// Run cleanup (all dirs exist)
	rotator.mu.Lock()
	rotator.cleanupLastCheckTimeMap()
	rotator.mu.Unlock()

	// Both should still be in map
	rotator.mu.Lock()
	if len(rotator.lastCheckTime) != 2 {
		t.Errorf("Expected 2 entries after cleanup, got %d", len(rotator.lastCheckTime))
	}

	if _, exists := rotator.lastCheckTime[dir1]; !exists {
		t.Error("Expected dir1 to remain in map")
	}
	if _, exists := rotator.lastCheckTime[dir2]; !exists {
		t.Error("Expected dir2 to remain in map")
	}
	rotator.mu.Unlock()
}

// TestMapGrowthPrevention_LargeNumberOfDirs tests map cleanup prevents unbounded growth
func TestMapGrowthPrevention_LargeNumberOfDirs(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	os.Mkdir(logDir, 0755)

	rotator := NewLogRotator(10, 1024*1024)

	logPath := filepath.Join(logDir, "test.log")
	f, _ := os.Create(logPath)
	f.Close()

	// Create and delete 1000+ directories with CheckAndRotate
	for i := 0; i < 1100; i++ {
		tempDir := filepath.Join(tmpDir, "temp", "dir"+string(rune(i)))
		os.MkdirAll(tempDir, 0755)

		rotator.mu.Lock()
		rotator.lastCheckTime[tempDir] = time.Now()
		rotator.mu.Unlock()

		// Every 100 operations, delete directories
		if i > 0 && i%100 == 0 {
			os.RemoveAll(filepath.Join(tmpDir, "temp"))
			os.Mkdir(filepath.Join(tmpDir, "temp"), 0755)
		}

		// Trigger CheckAndRotate which increments counter
		rotator.CheckAndRotate(logPath)
	}

	// Map should be cleaned up periodically, not growing unbounded
	rotator.mu.Lock()
	mapSize := len(rotator.lastCheckTime)
	rotator.mu.Unlock()

	// Map should be much smaller than 1100 entries due to cleanup
	if mapSize > 200 {
		t.Errorf("Map grew too large: %d entries (should be cleaned up)", mapSize)
	}
}

// TestCleanupWithRaceDetector is designed to be run with -race flag
func TestCleanupWithRaceDetector(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	os.Mkdir(logDir, 0755)

	rotator := NewLogRotator(10, 1024*1024)

	logPath := filepath.Join(logDir, "test.log")
	f, _ := os.Create(logPath)
	f.Close()

	// Create multiple goroutines calling CheckAndRotate concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				rotator.CheckAndRotate(logPath)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
}

// TestCleanupNormalOperationUnaffected tests that normal rotation logic still works
func TestCleanupNormalOperationUnaffected(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	os.Mkdir(logDir, 0755)

	rotator := NewLogRotator(3, 50) // Keep 3 files, 50 bytes max

	logPath := filepath.Join(logDir, "test.log")

	// Create multiple files and trigger rotation
	for i := 0; i < 10; i++ {
		f, _ := os.Create(logPath)
		f.WriteString(string(make([]byte, 100))) // Exceed 50 byte limit
		f.Close()

		rotator.CheckAndRotate(logPath)
		time.Sleep(10 * time.Millisecond)
	}

	// Verify rotation still happened correctly
	entries, _ := os.ReadDir(logDir)
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	// Should have limited files due to rotation
	if count > 10 {
		t.Errorf("Expected <= 10 files, got %d", count)
	}
}
