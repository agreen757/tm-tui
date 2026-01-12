package dialog

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestViewRecentLogs tests viewing recent logs
func TestViewRecentLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test log files directly in tmpDir
	logFile := filepath.Join(tmpDir, "git-status.log")
	f, _ := os.Create(logFile)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:45.123Z","command_id":"test-1","message":"command executed","event":"start"}\n`)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:46.123Z","command_id":"test-1","message":"command completed","event":"completion","duration_ms":1000}\n`)
	f.Close()

	// Create viewer pointing to tmpDir
	viewer := NewGitLogViewer(tmpDir)

	// View recent logs
	entries, err := viewer.ViewRecentLogs(10, "")
	if err != nil {
		t.Fatalf("Failed to view logs: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if !stringContainsSubstring(entries[0].CommandID, "test-1") || !stringContainsSubstring(entries[0].Message, "completed") {
		t.Error("Most recent entry not first")
	}
}

// TestFilterByType tests filtering logs by type
func TestFilterByType(t *testing.T) {
	tmpDir := t.TempDir()

	// Write different log types directly in tmpDir
	statusFile := filepath.Join(tmpDir, "git-status.log")
	branchFile := filepath.Join(tmpDir, "git-branch.log")

	f, _ := os.Create(statusFile)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:45.123Z","command_id":"status-1","message":"status check"}\n`)
	f.Close()

	f, _ = os.Create(branchFile)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:46.123Z","command_id":"branch-1","message":"branch list"}\n`)
	f.Close()

	viewer := NewGitLogViewer(tmpDir)

	// Filter by status type
	entries, err := viewer.ViewRecentLogs(10, "status")
	if err != nil {
		t.Fatalf("Failed to filter logs: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if !stringContainsSubstring(entries[0].Message, "status") {
		t.Error("Filtered entry doesn't match type")
	}
}

// TestGzippedLogs tests reading gzipped log files
func TestGzippedLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create gzipped log file directly in tmpDir
	gzPath := filepath.Join(tmpDir, "git-status.log.gz")
	f, _ := os.Create(gzPath)
	gzWriter := gzip.NewWriter(f)
	gzWriter.Write([]byte(`{"level":"info","time":"2024-01-12T15:30:45.123Z","command_id":"test-1","message":"gzipped entry"}\n`))
	gzWriter.Close()
	f.Close()

	viewer := NewGitLogViewer(tmpDir)

	// View logs
	entries, err := viewer.ViewRecentLogs(10, "")
	if err != nil {
		t.Fatalf("Failed to read gzipped logs: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if !stringContainsSubstring(entries[0].Message, "gzipped") {
		t.Error("Gzipped entry not read correctly")
	}
}

// TestSearchLogs tests searching logs
func TestSearchLogs(t *testing.T) {
	tmpDir := t.TempDir()

	logFile := filepath.Join(tmpDir, "git-status.log")
	f, _ := os.Create(logFile)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:45.123Z","command_id":"test-1","message":"successful branch checkout"}\n`)
	f.WriteString(`{"level":"error","time":"2024-01-12T15:30:46.123Z","command_id":"test-2","message":"failed merge"}\n`)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:47.123Z","command_id":"test-3","message":"status check complete"}\n`)
	f.Close()

	viewer := NewGitLogViewer(tmpDir)

	// Search for "branch"
	entries, err := viewer.SearchLogs("branch")
	if err != nil {
		t.Fatalf("Failed to search logs: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 match for 'branch', got %d", len(entries))
	}

	// Search for "error"
	entries, err = viewer.SearchLogs("error")
	if err != nil {
		t.Fatalf("Failed to search logs: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 match for 'error', got %d", len(entries))
	}
}

// TestFormatAsText tests text formatting
func TestFormatAsText(t *testing.T) {
	entries := []GitLogEntry{
		{
			Time:       "2024-01-12T15:30:45.123Z",
			CommandID:  "test-1",
			Message:    "command executed",
			Command:    "git",
			DurationMs: 1000,
		},
	}

	viewer := NewGitLogViewer("")
	text := viewer.FormatAsText(entries)

	if !stringContainsSubstring(text, "test-1") {
		t.Error("CommandID not in formatted text")
	}
	if !stringContainsSubstring(text, "command executed") {
		t.Error("Message not in formatted text")
	}
	if !stringContainsSubstring(text, "git") {
		t.Error("Command not in formatted text")
	}
	if !stringContainsSubstring(text, "1000ms") {
		t.Error("Duration not in formatted text")
	}
}

// TestGetLogTypes tests retrieving available log types
func TestGetLogTypes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create log files with different types directly in tmpDir
	os.Create(filepath.Join(tmpDir, "git-status.log"))
	os.Create(filepath.Join(tmpDir, "git-branch.log"))
	os.Create(filepath.Join(tmpDir, "git-checkout.log.gz"))

	viewer := NewGitLogViewer(tmpDir)

	types, err := viewer.GetLogTypes()
	if err != nil {
		t.Fatalf("Failed to get log types: %v", err)
	}

	if len(types) < 2 {
		t.Errorf("Expected at least 2 types, got %d", len(types))
	}

	// Check that git-status is in types
	found := false
	for _, t := range types {
		if stringContainsSubstring(t, "status") {
			found = true
			break
		}
	}
	if !found {
		t.Error("git-status type not found")
	}
}

// TestTimeRangeFilter tests filtering by time range
func TestTimeRangeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	logFile := filepath.Join(tmpDir, "git-status.log")
	f, _ := os.Create(logFile)
	f.WriteString(`{"level":"info","time":"2024-01-12T15:30:45.123Z","command_id":"test-1"}\n`)
	f.Close()

	// Set file modification time
	past := time.Now().Add(-2 * time.Hour)
	os.Chtimes(logFile, past, past)

	viewer := NewGitLogViewer(tmpDir)

	// Query with time range that includes the file
	since := time.Now().Add(-3 * time.Hour)
	until := time.Now().Add(1 * time.Hour)

	entries, err := viewer.ViewLogsByTimeRange(since, until)
	if err != nil {
		t.Fatalf("Failed to filter by time range: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Expected entries in time range")
	}
}

// Helper function - check if a string contains a substring
func stringContainsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
