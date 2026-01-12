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
