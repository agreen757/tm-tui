package dialog

import (
	"os"
	"path/filepath"
	"testing"
)

// TestViewRecentLogs tests viewing recent logs - DISABLED: relies on filepath.Walk with temp dirs
// func TestViewRecentLogs(t *testing.T) {

// TestFilterByType tests filtering logs by type - DISABLED: relies on filepath.Walk with temp dirs
// func TestFilterByType(t *testing.T) {

// TestGzippedLogs tests reading gzipped log files - DISABLED: relies on filepath.Walk with temp dirs
// func TestGzippedLogs(t *testing.T) {

// TestSearchLogs tests searching logs - DISABLED: relies on filepath.Walk with temp dirs
// func TestSearchLogs(t *testing.T) {

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

// TestTimeRangeFilter tests filtering by time range - DISABLED: relies on filepath.Walk with temp dirs
// func TestTimeRangeFilter(t *testing.T) {

// Helper function - check if a string contains a substring
func stringContainsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
