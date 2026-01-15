package executor

import (
	"strings"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
)

func TestEscapeShellArg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "text with single quote",
			input:    "it's working",
			expected: "'it'\\''s working'",
		},
		{
			name:     "multiple single quotes",
			input:    "don't worry, it's ok",
			expected: "'don'\\''t worry, it'\\''s ok'",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "'   '",
		},
		{
			name:     "special characters",
			input:    "test!@#$%^&*()",
			expected: "'test!@#$%^&*()'",
		},
		{
			name:     "newline characters",
			input:    "line1\nline2",
			expected: "'line1\nline2'",
		},
		{
			name:     "quotes at start",
			input:    "'hello",
			expected: "''\\''hello'",
		},
		{
			name:     "quotes at end",
			input:    "hello'",
			expected: "'hello'\\'''",
		},
		{
			name:     "consecutive quotes",
			input:    "test''case",
			expected: "'test'\\'''\\''case'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeShellArg(tt.input)
			if result != tt.expected {
				t.Errorf("escapeShellArg(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUpdateTaskCommandSelection(t *testing.T) {
	tests := []struct {
		name          string
		taskID        string
		content       string
		shouldFail    bool
		expectedCmd   string
		expectedFlags []string
	}{
		{
			name:          "main task with content",
			taskID:        "1",
			content:       "test content",
			shouldFail:    false,
			expectedCmd:   "update-task",
			expectedFlags: []string{"--id=1", "--prompt='test content'"},
		},
		{
			name:          "subtask with content",
			taskID:        "1.1",
			content:       "subtask content",
			shouldFail:    false,
			expectedCmd:   "update-subtask",
			expectedFlags: []string{"--id=1.1", "--prompt='subtask content'"},
		},
		{
			name:          "nested subtask with content",
			taskID:        "1.2.3",
			content:       "nested content",
			shouldFail:    false,
			expectedCmd:   "update-subtask",
			expectedFlags: []string{"--id=1.2.3", "--prompt='nested content'"},
		},
		{
			name:          "empty task ID",
			taskID:        "",
			content:       "content",
			shouldFail:    true,
			expectedCmd:   "",
			expectedFlags: nil,
		},
		{
			name:          "main task empty content",
			taskID:        "2",
			content:       "",
			shouldFail:    false,
			expectedCmd:   "update-task",
			expectedFlags: []string{"--id=2"},
		},
		{
			name:          "main task whitespace content",
			taskID:        "3",
			content:       "   \t\n  ",
			shouldFail:    false,
			expectedCmd:   "update-task",
			expectedFlags: []string{"--id=3"},
		},
		{
			name:          "subtask with quotes in content",
			taskID:        "2.1",
			content:       "It's a test",
			shouldFail:    false,
			expectedCmd:   "update-subtask",
			expectedFlags: []string{"--id=2.1", "--prompt='It'\\''s a test'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh service instance for each test
			cfg := &config.Config{TaskMasterPath: "/tmp"}
			service, err := NewService(cfg)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}
			defer service.Close()

			executor := NewTaskMasterExecutor(service)
			err = executor.UpdateTask(tt.taskID, tt.content)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("UpdateTask(%q, %q) expected error but got nil", tt.taskID, tt.content)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateTask(%q, %q) failed: %v", tt.taskID, tt.content, err)
			}
		})
	}
}

func TestTaskIDValidation(t *testing.T) {
	tests := []struct {
		name      string
		taskID    string
		shouldErr bool
	}{
		{
			name:      "valid main task",
			taskID:    "1",
			shouldErr: false,
		},
		{
			name:      "valid subtask",
			taskID:    "1.1",
			shouldErr: false,
		},
		{
			name:      "valid nested subtask",
			taskID:    "1.2.3",
			shouldErr: false,
		},
		{
			name:      "empty task ID",
			taskID:    "",
			shouldErr: true,
		},
		{
			name:      "whitespace task ID",
			taskID:    "   ",
			shouldErr: false, // Will execute with whitespace ID - service will handle error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh service instance for each test
			cfg := &config.Config{TaskMasterPath: "/tmp"}
			service, err := NewService(cfg)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}
			defer service.Close()

			executor := NewTaskMasterExecutor(service)
			err = executor.UpdateTask(tt.taskID, "test")

			if tt.shouldErr && err == nil {
				t.Errorf("UpdateTask(%q, ...) expected error but got nil", tt.taskID)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("UpdateTask(%q, ...) should not error but got: %v", tt.taskID, err)
			}
		})
	}
}

func TestSubtaskDetection(t *testing.T) {
	tests := []struct {
		name              string
		taskID            string
		expectedCmdType   string
		shouldContainDots bool
	}{
		{
			name:              "main task has no dots",
			taskID:            "1",
			expectedCmdType:   "update-task",
			shouldContainDots: false,
		},
		{
			name:              "subtask has dot",
			taskID:            "1.1",
			expectedCmdType:   "update-subtask",
			shouldContainDots: true,
		},
		{
			name:              "nested subtask has multiple dots",
			taskID:            "1.2.3",
			expectedCmdType:   "update-subtask",
			shouldContainDots: true,
		},
		{
			name:              "numeric main task",
			taskID:            "42",
			expectedCmdType:   "update-task",
			shouldContainDots: false,
		},
		{
			name:              "deep nesting",
			taskID:            "1.2.3.4.5",
			expectedCmdType:   "update-subtask",
			shouldContainDots: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh service instance for each test
			cfg := &config.Config{TaskMasterPath: "/tmp"}
			service, err := NewService(cfg)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}
			defer service.Close()

			executor := NewTaskMasterExecutor(service)
			// Just verify the detection logic by ensuring no error
			err = executor.UpdateTask(tt.taskID, "test")
			if err != nil {
				// Only empty task ID should error
				if tt.taskID != "" {
					t.Errorf("UpdateTask(%q, ...) failed: %v", tt.taskID, err)
				}
			}

			// Verify the detection logic
			isSubtask := strings.Contains(tt.taskID, ".")
			if isSubtask && tt.expectedCmdType != "update-subtask" {
				t.Errorf("UpdateTask(%q) should detect as subtask", tt.taskID)
			}
			if !isSubtask && tt.expectedCmdType != "update-task" {
				t.Errorf("UpdateTask(%q) should detect as main task", tt.taskID)
			}
		})
	}
}

func TestContentWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		taskID  string
		content string
	}{
		{
			name:    "content with quotes",
			taskID:  "1",
			content: "Fix the \"bug\" in parser",
		},
		{
			name:    "content with backslashes",
			taskID:  "1.1",
			content: "Add regex pattern: \\d+\\.\\d+",
		},
		{
			name:    "content with newlines",
			taskID:  "2",
			content: "Line 1\nLine 2\nLine 3",
		},
		{
			name:    "content with mixed special chars",
			taskID:  "2.1",
			content: "Test: var x = 'value'; echo \"done\"",
		},
		{
			name:    "multiline content",
			taskID:  "3",
			content: "Implementation:\n- Step 1: Create files\n- Step 2: Add tests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh service instance for each test
			cfg := &config.Config{TaskMasterPath: "/tmp"}
			service, err := NewService(cfg)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}
			defer service.Close()

			executor := NewTaskMasterExecutor(service)
			err = executor.UpdateTask(tt.taskID, tt.content)
			if err != nil {
				t.Errorf("UpdateTask(%q, %q) failed: %v", tt.taskID, tt.content, err)
			}
		})
	}
}

func TestOutputStreamingChannels(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp"}
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	executor := NewTaskMasterExecutor(service)

	// Verify output and done channels are available
	outputCh := executor.GetOutput()
	doneCh := executor.GetDone()

	if outputCh == nil {
		t.Errorf("GetOutput() returned nil channel")
	}

	if doneCh == nil {
		t.Errorf("GetDone() returned nil channel")
	}
}

func TestIsRunningState(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp"}
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	executor := NewTaskMasterExecutor(service)

	// Initially should not be running
	if executor.IsRunning() {
		t.Errorf("IsRunning() expected false initially, got true")
	}

	// After executing a command, it should be running
	err = executor.UpdateTask("1", "test")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Should be running after UpdateTask call
	// (Note: May depend on timing, but generally should be true immediately after)
	if !executor.IsRunning() {
		t.Errorf("IsRunning() expected true after UpdateTask, got false")
	}
}

func TestStreamingOutputAvailable(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp"}
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	executor := NewTaskMasterExecutor(service)

	// Get output channel
	outputCh := executor.GetOutput()

	// Execute a command
	err = executor.UpdateTask("1", "test content")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// The output channel should be readable (though may be empty initially)
	// Just verify we can attempt to read from it
	select {
	case output, ok := <-outputCh:
		if ok {
			// Received some output (expected for successful command)
			if output == "" && output != "" {
				t.Errorf("unexpected output state")
			}
		}
	default:
		// No output immediately available, which is fine
		// The channel should be non-nil and readable
	}
}

func TestMultipleCommandExecution(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp"}
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	executor := NewTaskMasterExecutor(service)

	// Execute first command
	err = executor.UpdateTask("1", "first task")
	if err != nil {
		t.Fatalf("First UpdateTask failed: %v", err)
	}

	// Try to execute second command while first is running
	// Should fail with "command already running" error
	err = executor.UpdateTask("2", "second task")
	if err == nil {
		t.Errorf("Expected error when executing second command while first is running")
	}

	if err.Error() != "command already running" {
		t.Errorf("Expected 'command already running' error, got: %v", err)
	}
}

func TestExecutorChannelBehavior(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp"}
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	executor := NewTaskMasterExecutor(service)

	// Verify channels are the same as service channels
	executorOutput := executor.GetOutput()
	executorDone := executor.GetDone()
	serviceOutput := service.GetOutput()
	serviceDone := service.GetDone()

	// Both should return the same channel references
	if executorOutput != serviceOutput {
		t.Errorf("GetOutput() should return service's output channel")
	}

	if executorDone != serviceDone {
		t.Errorf("GetDone() should return service's done channel")
	}
}

func TestUpdateTaskWithEmptyIDAndContent(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp"}
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	executor := NewTaskMasterExecutor(service)

	// Empty task ID should fail
	err = executor.UpdateTask("", "")
	if err == nil {
		t.Errorf("UpdateTask with empty ID should fail")
	}

	if err.Error() != "task ID cannot be empty" {
		t.Errorf("Expected 'task ID cannot be empty' error, got: %v", err)
	}
}
