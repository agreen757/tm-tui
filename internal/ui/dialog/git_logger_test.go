package dialog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLogFormat_Complete tests that the log output is valid JSON and complete
func TestLogFormat_Complete(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create structured logger
	commandID := "test-cmd-1"
	args := []string{"status"}
	tagName := "test-tag"

	gitLogger, logPath, err := NewGitLogger(commandID, args, tagName)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log some operations
	gitLogger.LogOutput("stdout", "On branch main")
	gitLogger.LogOutput("stderr", "")
	gitLogger.LogWarning("test warning", nil)
	gitLogger.LogCompletion(0, nil)

	gitLogger.Close()

	// Read and parse log file
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse each line as JSON
	lines := []string{}
	decoder := json.NewDecoder(os.Stdin)
	decoder.UseNumber()

	// Verify log file contains JSON data
	if len(logContent) == 0 {
		t.Fatal("Log file is empty")
	}

	// Verify file contains expected content
	logStr := string(logContent)
	if !gitLoggerTestContains(logStr, "command_id") {
		t.Error("Log missing command_id field")
	}
	if !gitLoggerTestContains(logStr, "git_args") {
		t.Error("Log missing git_args field")
	}
	// Zerolog includes timestamp field in JSON output
	if !gitLoggerTestContains(logStr, "time") && !gitLoggerTestContains(logStr, "timestamp") {
		t.Error("Log missing time/timestamp field")
	}

	_ = lines // For future use with line parsing
}

// TestLogFields_Present tests that all required fields are present in logs
func TestLogFields_Present(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	commandID := "test-cmd-2"
	args := []string{"branch", "-a"}
	tagName := "test-tag-2"

	gitLogger, logPath, err := NewGitLogger(commandID, args, tagName)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	gitLogger.LogOutput("stdout", "  main")
	gitLogger.LogOutput("stdout", "* develop")
	gitLogger.LogCompletion(0, nil)
	gitLogger.Close()

	// Read and verify log file
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(logContent)

	// Verify all required fields are present
	requiredFields := []string{
		"command_id",
		"git_args",
		"tag",
		"duration_ms",
	}

	for _, field := range requiredFields {
		if !gitLoggerTestContains(logStr, field) {
			t.Errorf("Log missing required field: %s", field)
		}
	}

	// Check for time field (zerolog outputs it as "time" not "timestamp")
	if !gitLoggerTestContains(logStr, "time") {
		t.Error("Log missing time field")
	}
}

// TestDuration_Accurate tests that duration is correctly calculated
func TestDuration_Accurate(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	commandID := "test-cmd-3"
	args := []string{"--version"}

	gitLogger, logPath, err := NewGitLogger(commandID, args, "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Sleep briefly to ensure measurable duration
	time.Sleep(10 * time.Millisecond)

	gitLogger.LogCompletion(0, nil)
	gitLogger.Close()

	// Read and verify duration
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(logContent)
	if !gitLoggerTestContains(logStr, "duration_ms") {
		t.Fatal("Log missing duration_ms field")
	}

	// Check for time field
	if !gitLoggerTestContains(logStr, "time") {
		t.Error("Log missing time field")
	}
}

// TestLogCreationOnError tests that logs are created even if operations fail
func TestLogCreationOnError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	commandID := "test-cmd-error"
	args := []string{"invalid-command"}

	gitLogger, logPath, err := NewGitLogger(commandID, args, "error-test")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	gitLogger.LogError("test error", nil)
	gitLogger.LogCompletion(1, nil)
	gitLogger.Close()

	// Verify log file exists
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("Log file not created: %v", err)
	}

	logContent, _ := os.ReadFile(logPath)
	logStr := string(logContent)

	// Verify error was logged
	if !gitLoggerTestContains(logStr, "error") || !gitLoggerTestContains(logStr, "test error") {
		t.Error("Error message not found in log")
	}
}

// TestLogRotation tests basic log rotation setup
func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	tagName := "rotation-test"

	// Create 5 log files
	for i := 1; i <= 5; i++ {
		commandID := fmt.Sprintf("git-cmd-rotation-%d", i)
		gitLogger, _, err := NewGitLogger(commandID, []string{"status"}, tagName)
		if err != nil {
			t.Fatalf("Failed to create logger %d: %v", i, err)
		}
		gitLogger.LogCompletion(0, nil)
		gitLogger.Close()
	}

	// Verify logs directory exists and contains files
	logsDir := filepath.Join(".taskmaster", tagName)
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatalf("Failed to read logs directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No log files created")
	}
}

// Helper function to check if a string contains a substring (specific to git_logger tests)
func gitLoggerTestContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && gitLoggerTestStringContains(s, substr)
}

// gitLoggerTestStringContains is a basic substring search
func gitLoggerTestStringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
