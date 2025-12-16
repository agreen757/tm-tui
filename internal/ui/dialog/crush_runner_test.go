package dialog

import (
	"os"
	"strings"
	"testing"

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

	prompt, err := GenerateCrushPrompt(task, "claude-3-5-sonnet-20241022")
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
	_, err := GenerateCrushPrompt(nil, "test-model")
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

	prompt, err := GenerateCrushPrompt(task, "test-model")
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

	prompt, err := GenerateCrushPrompt(task, "test-model")
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

	_, err = GenerateCrushPrompt(task, "test-model")
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}
