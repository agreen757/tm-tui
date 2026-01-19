package dialog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

func TestValidateCrushBinary(t *testing.T) {
	// This test will pass if crush is installed, otherwise it should return a helpful error
	err := ValidateCrushBinary()
	if err != nil {
		// Check that the error message is helpful
		if !strings.Contains(err.Error(), "crush binary not found") {
			t.Errorf("Expected helpful error message, got: %v", err)
		}
		if !strings.Contains(err.Error(), "go install") {
			t.Errorf("Expected installation instructions in error, got: %v", err)
		}
		// This is expected if crush is not installed
		t.Skip("Crush binary not installed - this is expected in test environments")
	}
}

func TestGenerateCrushPrompt(t *testing.T) {
	task := &taskmaster.Task{
		ID:           "1.2.3",
		Title:        "Implement user authentication",
		Description:  "Add JWT-based authentication system",
		Details:      "Use bcrypt for password hashing and JWT for tokens",
		TestStrategy: "Unit tests for auth functions, integration tests for login flow",
		Priority:     "high",
		Dependencies: []string{"1.1", "1.2"},
	}

	prompt, err := GenerateCrushPrompt(task, "claude-3-5-sonnet-20241022", "feature-auth")
	if err != nil {
		t.Fatalf("Failed to generate prompt: %v", err)
	}

	// Check that the prompt contains key task information
	requiredContent := []string{
		task.ID,
		task.Title,
		task.Description,
		task.Details,
		task.TestStrategy,
		task.Priority,
	}

	for _, content := range requiredContent {
		if !strings.Contains(prompt, content) {
			t.Errorf("Prompt missing required content: %s", content)
		}
	}
}

func TestGenerateCrushPromptNilTask(t *testing.T) {
	_, err := GenerateCrushPrompt(nil, "test-model", "")
	if err == nil {
		t.Error("Expected error for nil task, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected nil task error, got: %v", err)
	}
}

func TestGenerateCrushPromptWithCustomWorkflowGuide(t *testing.T) {
	// Create a temporary CRUSH_RUN_INSTRUCTIONS.md
	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	customGuide := `# Custom Workflow
Task: {{.TaskID}}
Title: {{.Title}}
Custom field test
`
	if err := os.WriteFile("CRUSH_RUN_INSTRUCTIONS.md", []byte(customGuide), 0644); err != nil {
		t.Fatalf("Failed to create test CRUSH_RUN_INSTRUCTIONS.md: %v", err)
	}

	task := &taskmaster.Task{
		ID:    "test-1",
		Title: "Test Task",
	}

	prompt, err := GenerateCrushPrompt(task, "test-model", "test-tag")
	if err != nil {
		t.Fatalf("Failed to generate prompt with custom guide: %v", err)
	}

	if !strings.Contains(prompt, "Custom field test") {
		t.Error("Prompt should contain custom workflow guide content")
	}
	if !strings.Contains(prompt, "test-1") {
		t.Error("Prompt should contain task ID")
	}
}

func TestGenerateCrushPromptEmptyFields(t *testing.T) {
	task := &taskmaster.Task{
		ID:    "1",
		Title: "Minimal Task",
		// Other fields empty
	}

	prompt, err := GenerateCrushPrompt(task, "test-model", "test-tag")
	if err != nil {
		t.Fatalf("Failed to generate prompt with minimal task: %v", err)
	}

	// Should still contain the basics
	if !strings.Contains(prompt, task.ID) {
		t.Error("Prompt missing task ID")
	}
	if !strings.Contains(prompt, task.Title) {
		t.Error("Prompt missing task title")
	}
}

func TestGetCrushCommand(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		model    string
		expected []string
	}{
		{
			name:     "with model",
			prompt:   "test prompt",
			model:    "claude-3-5-sonnet-20241022",
			expected: []string{"run", "--model", "claude-3-5-sonnet-20241022"},
		},
		{
			name:     "without model",
			prompt:   "test prompt",
			model:    "",
			expected: []string{"run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := GetCrushCommand(tt.prompt, tt.model)

			if len(args) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expected), len(args), args)
				return
			}

			for i, arg := range args {
				if arg != tt.expected[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestCrushBinaryError(t *testing.T) {
	err := &CrushBinaryError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got: %s", err.Error())
	}
}

func TestGenerateCrushPromptInvalidTemplate(t *testing.T) {
	// Create a temporary CRUSH_RUN_INSTRUCTIONS.md with invalid template syntax
	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	invalidGuide := `# Invalid Template
{{.TaskID
`
	if err := os.WriteFile("CRUSH_RUN_INSTRUCTIONS.md", []byte(invalidGuide), 0644); err != nil {
		t.Fatalf("Failed to create test CRUSH_RUN_INSTRUCTIONS.md: %v", err)
	}

	task := &taskmaster.Task{
		ID:    "test-1",
		Title: "Test Task",
	}

	_, err = GenerateCrushPrompt(task, "test-model", "")
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestRunCommandEmptyCommandID(t *testing.T) {
	_, err := RunCommand("", "test prompt", "")
	if err == nil {
		t.Error("Expected error for empty command ID")
	}
	if !strings.Contains(err.Error(), "command ID cannot be empty") {
		t.Errorf("Expected command ID error, got: %v", err)
	}
}

func TestRunCommandEmptyPrompt(t *testing.T) {
	_, err := RunCommand("cmd-1", "", "")
	if err == nil {
		t.Error("Expected error for empty prompt")
	}
	if !strings.Contains(err.Error(), "prompt cannot be empty") {
		t.Errorf("Expected prompt error, got: %v", err)
	}
}

func TestRunCommandWhitespacePrompt(t *testing.T) {
	_, err := RunCommand("cmd-1", "   ", "")
	if err == nil {
		t.Error("Expected error for whitespace-only prompt")
	}
	if !strings.Contains(err.Error(), "prompt cannot be empty") {
		t.Errorf("Expected prompt error, got: %v", err)
	}
}

func TestRunCommandCrushNotInstalled(t *testing.T) {
	// This test expects crush to be installed, skip if not
	err := ValidateCrushBinary()
	if err != nil {
		t.Skip("Crush binary not installed - skipping RunCommand test")
	}

	// If crush is installed, test with valid parameters
	_, err = RunCommand("test-cmd", "echo 'test'", "")
	if err != nil {
		t.Errorf("Expected RunCommand to start without error, got: %v", err)
	}
}

func TestRunCommandParameterValidation(t *testing.T) {
	tests := []struct {
		name      string
		commandID string
		prompt    string
		model     string
		expectErr bool
		errSubstr string
	}{
		{
			name:      "valid parameters",
			commandID: "cmd-1",
			prompt:    "test prompt",
			model:     "test-model",
			expectErr: false,
		},
		{
			name:      "valid without model",
			commandID: "cmd-2",
			prompt:    "test prompt",
			model:     "",
			expectErr: false,
		},
		{
			name:      "empty command ID",
			commandID: "",
			prompt:    "test prompt",
			expectErr: true,
			errSubstr: "command ID cannot be empty",
		},
		{
			name:      "empty prompt",
			commandID: "cmd-1",
			prompt:    "",
			expectErr: true,
			errSubstr: "prompt cannot be empty",
		},
		{
			name:      "whitespace-only prompt",
			commandID: "cmd-1",
			prompt:    "   \t  ",
			expectErr: true,
			errSubstr: "prompt cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RunCommand(tt.commandID, tt.prompt, tt.model)

			// Skip crush-not-found errors for parameter validation tests
			if err != nil && strings.Contains(err.Error(), "crush binary not found") {
				t.Skip("Crush binary not installed")
			}

			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				// Skip if crush is not available
				if strings.Contains(err.Error(), "crush binary not found") {
					t.Skip("Crush binary not installed")
				}
				t.Errorf("Expected no error, got: %v", err)
			}
			if tt.expectErr && err != nil && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("Expected error containing %q, got: %v", tt.errSubstr, err)
			}
		})
	}
}

func TestRunCommandLogFileCreation(t *testing.T) {
	// Skip if crush is not available
	err := ValidateCrushBinary()
	if err != nil {
		t.Skip("Crush binary not installed")
	}

	// Create temp directory for logs
	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Run command with valid parameters
	outputChan, err := RunCommand("test-cmd", "echo 'Hello World'", "")
	if err != nil {
		t.Fatalf("Failed to execute RunCommand: %v", err)
	}

	// Consume output to let goroutine complete
	count := 0
	for range outputChan {
		count++
		if count > 100 { // Prevent infinite loop
			break
		}
	}

	// Wait a bit for file operations to complete
	time.Sleep(100 * time.Millisecond)

	// Check if log file was created
	logsDir := filepath.Join(tempDir, ".taskmaster", "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatalf("Failed to read logs directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Expected log file to be created")
		return
	}

	// Check log file content
	logPath := filepath.Join(logsDir, entries[0].Name())
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	requiredStrings := []string{
		"Crush Command Log",
		"Command ID: test-cmd",
		"Started:",
		"Completed:",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(contentStr, required) {
			t.Errorf("Log file missing required content: %s", required)
		}
	}
}

func TestRunCommandOutputStreaming(t *testing.T) {
	// This test is skipped in non-interactive environments since crush run
	// requires interactive terminal setup
	t.Skip("Crush output streaming test requires interactive terminal - tested via integration tests")
}

func TestRunCommandChannelBuffering(t *testing.T) {
	// This test is skipped in non-interactive environments since crush run
	// requires interactive terminal setup
	t.Skip("Crush buffering test requires interactive terminal - tested via integration tests")
}

// TestSanitizeTaskIDForFilename tests the sanitization of task IDs for use in filenames
func TestSanitizeTaskIDForFilename(t *testing.T) {
	tests := []struct {
		name     string
		taskID   string
		expected string
	}{
		{
			name:     "single digit task",
			taskID:   "1",
			expected: "1",
		},
		{
			name:     "nested task ID",
			taskID:   "1.2",
			expected: "1.2",
		},
		{
			name:     "deeply nested task ID",
			taskID:   "2.1.3",
			expected: "2.1.3",
		},
		{
			name:     "command ID with numeric suffix",
			taskID:   "command-1234567890",
			expected: "command-1234567890",
		},
		{
			name:     "command ID with large timestamp",
			taskID:   "command-9876543210",
			expected: "command-9876543210",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTaskIDForFilename(tt.taskID)
			if result != tt.expected {
				t.Errorf("SanitizeTaskIDForFilename(%q) = %q, want %q", tt.taskID, result, tt.expected)
			}
		})
	}
}

// TestCreateCrushLogFileWithCommandID tests log file creation with command IDs
func TestCreateCrushLogFileWithCommandID(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		taskID  string
		tagName string
		checkFn func(t *testing.T, logPath string)
	}{
		{
			name:    "command ID with tag",
			taskID:  "command-1234567890",
			tagName: "test-tag",
			checkFn: func(t *testing.T, logPath string) {
				if !testContains(logPath, "command-1234567890.log") {
					t.Errorf("Expected log path to contain command ID, got: %s", logPath)
				}
			},
		},
		{
			name:    "task ID with tag",
			taskID:  "1.2",
			tagName: "test-tag",
			checkFn: func(t *testing.T, logPath string) {
				if !testContains(logPath, "1.2.log") {
					t.Errorf("Expected log path to contain task ID, got: %s", logPath)
				}
			},
		},
		{
			name:    "command ID without tag",
			taskID:  "command-9876543210",
			tagName: "",
			checkFn: func(t *testing.T, logPath string) {
				if !testContains(logPath, "crush-run-command-9876543210") {
					t.Errorf("Expected log path to contain crush-run- prefix and command ID, got: %s", logPath)
				}
			},
		},
		{
			name:    "task ID without tag",
			taskID:  "2.1",
			tagName: "",
			checkFn: func(t *testing.T, logPath string) {
				if !testContains(logPath, "crush-run-2.1") {
					t.Errorf("Expected log path to contain crush-run- prefix and task ID, got: %s", logPath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current directory and change to temp directory
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer os.Chdir(origDir)

			// Create log file
			logFile, logPath, err := createCrushLogFile(tt.taskID, tt.tagName)
			if err != nil {
				t.Fatalf("createCrushLogFile failed: %v", err)
			}
			defer logFile.Close()

			// Verify file exists
			if _, err := os.Stat(logPath); err != nil {
				t.Fatalf("Log file was not created at %s: %v", logPath, err)
			}

			// Run the check function
			tt.checkFn(t, logPath)
		})
	}
}

// Helper function
func testContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
