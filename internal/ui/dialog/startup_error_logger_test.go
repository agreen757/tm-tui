package dialog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLog_BinaryNotFound tests logging when git binary is not found
func TestLog_BinaryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Log a startup error
	err := LogStartupError("git-cmd-1", []string{"status"}, "test-tag", "git binary not found")
	if err != nil {
		t.Fatalf("Failed to log error: %v", err)
	}

	// Verify log file was created
	expectedPath := filepath.Join(".taskmaster", "test-tag", "git-cmd-1.log")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("Log file not created: %v", err)
	}

	// Read and verify content
	content, _ := os.ReadFile(expectedPath)
	logStr := string(content)

	if !stringContainsSubstring(logStr, "git binary not found") {
		t.Error("Error message not in log")
	}
	if !stringContainsSubstring(logStr, "startup_failed") {
		t.Error("Event type not correct")
	}
	if !stringContainsSubstring(logStr, "git-cmd-1") {
		t.Error("CommandID not in log")
	}
}

// TestLog_PermissionDenied tests logging permission denied errors
func TestLog_PermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := LogStartupError("git-cmd-2", []string{"clone", "repo"}, "feature-branch", "permission denied")
	if err != nil {
		t.Fatalf("Failed to log error: %v", err)
	}

	// Verify log file creation
	expectedPath := filepath.Join(".taskmaster", "feature-branch", "git-cmd-2.log")
	content, _ := os.ReadFile(expectedPath)
	logStr := string(content)

	if !stringContainsSubstring(logStr, "permission denied") {
		t.Error("Permission denied error not logged")
	}
}

// TestLog_InvalidArgs tests logging with invalid arguments
func TestLog_InvalidArgs(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := LogStartupError("git-invalid", []string{"--invalid-flag"}, "", "invalid command line argument")
	if err != nil {
		t.Fatalf("Failed to log error: %v", err)
	}

	// Check log was created in default logs directory
	entries, _ := os.ReadDir(filepath.Join(".taskmaster", "logs"))
	if len(entries) == 0 {
		t.Fatal("No log file created in default logs directory")
	}

	// Read first log file
	logFile := filepath.Join(".taskmaster", "logs", entries[0].Name())
	content, _ := os.ReadFile(logFile)
	logStr := string(content)

	if !stringContainsSubstring(logStr, "invalid command line argument") {
		t.Error("Invalid argument error not logged")
	}
}

// TestStartupErrorJSON tests that startup errors are valid JSON
func TestStartupErrorJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	LogStartupError("test-cmd", []string{"status"}, "test", "test error")

	expectedPath := filepath.Join(".taskmaster", "test", "test-cmd.log")
	content, _ := os.ReadFile(expectedPath)
	logStr := string(content)

	// Verify it's valid JSON by unmarshaling
	var entry map[string]interface{}
	err := json.Unmarshal([]byte(logStr), &entry)
	if err != nil {
		t.Fatalf("Log entry is not valid JSON: %v", err)
	}

	// Verify required fields
	if entry["level"] != "error" {
		t.Error("Level should be error")
	}
	if entry["event"] != "startup_failed" {
		t.Error("Event should be startup_failed")
	}
	if entry["duration_ms"] != float64(0) {
		t.Error("Duration should be 0 for startup failures")
	}
}

// TestStartupErrorWithTagAndWithout tests both with and without tag
func TestStartupErrorWithTagAndWithout(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Test with tag
	LogStartupError("cmd-with-tag", []string{"status"}, "my-tag", "error with tag")

	// Test without tag (empty string)
	LogStartupError("cmd-without-tag", []string{"branch"}, "", "error without tag")

	// Verify both created
	withTagPath := filepath.Join(".taskmaster", "my-tag", "cmd-with-tag.log")
	if _, err := os.Stat(withTagPath); err != nil {
		t.Fatalf("Tagged log not created: %v", err)
	}

	// Find the untagged log
	entries, _ := os.ReadDir(filepath.Join(".taskmaster", "logs"))
	if len(entries) == 0 {
		t.Fatal("Untagged log not created in default logs directory")
	}
}

// TestStartupErrorDirectoryCreation tests that directories are created
func TestStartupErrorDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Use deeply nested tag
	LogStartupError("deep-cmd", []string{"status"}, "deep/nested/tag", "error in deep directory")

	// Verify directory was created
	deepPath := filepath.Join(".taskmaster", "deep", "nested", "tag", "deep-cmd.log")
	if _, err := os.Stat(deepPath); err != nil {
		t.Fatalf("Deeply nested log directory not created: %v", err)
	}
}
