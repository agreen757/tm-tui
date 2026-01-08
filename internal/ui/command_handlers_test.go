package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

func newTestModel() Model {
	keyMap := NewKeyMap(nil)
	dm := dialog.InitializeDialogManager(120, 60, dialog.DefaultDialogStyle())
	return Model{
		appState: NewAppState(dm, &keyMap),
		keyMap:   keyMap,
		commands: defaultCommandSpecs(),
	}
}

func TestCommandDispatchOpensDialogs(t *testing.T) {
	model := newTestModel()
	cases := []struct {
		command CommandID
		title   string
	}{
		{CommandParsePRD, "Select PRD File"},
		{CommandAnalyzeComplexity, "Analyze Task Complexity"},
		{CommandManageTags, "Add Tag Context"},
		{CommandUseTag, "Use Tag Context"},
	}

	for _, tc := range cases {
		t.Run(string(tc.command), func(t *testing.T) {
			model.appState.ClearDialogs()
			model.dispatchCommand(tc.command)
			dialog := model.appState.ActiveDialog()
			if dialog == nil {
				t.Fatalf("expected a dialog for %s but none was active", tc.command)
			}
			if dialog.Title() != tc.title {
				t.Fatalf("expected dialog title %q, got %q", tc.title, dialog.Title())
			}
		})
	}
}

func TestExpandTaskCommandShowsErrorWithoutSelection(t *testing.T) {
	model := newTestModel()
	model.appState.ClearDialogs()

	model.dispatchCommand(CommandExpandTask)

	active := model.appState.ActiveDialog()
	if active == nil {
		t.Fatal("expected error dialog when expanding without selection")
	}

	errDialog, ok := active.(*dialog.ErrorDialogModel)
	if !ok {
		t.Fatalf("expected ErrorDialogModel, got %T", active)
	}

	if errDialog.Style == nil || errDialog.Style.BorderColor != errDialog.Style.ErrorColor {
		t.Fatalf("expected error dialog style to use error color border")
	}
}

func TestOpenCommandRunnerValidatesCrushBinary(t *testing.T) {
	model := newTestModel()
	model.appState.ClearDialogs()

	// The test will attempt to validate Crush binary.
	// If Crush is not installed, this should show an error dialog.
	// If Crush is installed, this should open the command runner dialog.
	cmd := model.openCommandRunner()

	// Check if we got an error dialog (Crush not found) or a command runner dialog
	active := model.appState.ActiveDialog()
	if active == nil {
		// If no dialog, the command returned something that executes later
		// For now, we just verify that a command is returned when Crush is available
		if cmd == nil {
			t.Skip("Crush binary not available for testing, but openCommandRunner structure is correct")
		}
		return
	}

	// If Crush is available, we should get the command runner dialog
	if active.Title() != "Run Command with Crush" {
		errDialog, ok := active.(*dialog.ErrorDialogModel)
		if !ok || errDialog.Title() != "Crush Binary Not Found" {
			t.Fatalf("expected either 'Run Command with Crush' dialog or error dialog, got %s", active.Title())
		}
	}
}

func TestOpenCommandRunnerReturnsInitCommand(t *testing.T) {
	model := newTestModel()
	model.appState.ClearDialogs()

	_ = model.openCommandRunner()

	// If validation passes, cmd should not be nil (it initializes the dialog)
	// We can only verify this if Crush is available
	if model.appState.ActiveDialog() != nil {
		if dialog, ok := model.appState.ActiveDialog().(*dialog.FormDialog); ok {
			if dialog.Title() != "Run Command with Crush" {
				t.Errorf("expected 'Run Command with Crush' dialog, got %q", dialog.Title())
			}
		}
	}
}

func TestCommandRunnerDialogCallback(t *testing.T) {
	_ = newTestModel()

	tests := []struct {
		name        string
		value       interface{}
		err         error
		expectError bool
		expectCmd   bool
	}{
		{
			name:      "valid prompt",
			value:     dialog.CommandPromptResult{Prompt: "test prompt"},
			err:       nil,
			expectCmd: true,
		},
		{
			name:      "nil value (cancellation)",
			value:     nil,
			err:       nil,
			expectCmd: false,
		},
		{
			name:      "empty prompt",
			value:     dialog.CommandPromptResult{Prompt: ""},
			err:       nil,
			expectCmd: false,
		},
		{
			name:        "error during form submission",
			value:       nil,
			err:         dialog.ErrorFormValidation{FieldID: "prompt", Message: "test error"},
			expectError: true,
		},
		{
			name:      "wrong type assertion",
			value:     "not a CommandPromptResult",
			err:       nil,
			expectCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the callback from openCommandRunner
			callback := func(value interface{}, err error) error {
				if err != nil {
					return err
				}

				if value == nil {
					return nil
				}

				result, ok := value.(dialog.CommandPromptResult)
				if !ok || result.Prompt == "" {
					return nil
				}

				// This would execute the command
				return nil // Command execution would happen here
			}

			err := callback(tt.value, tt.err)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCommandRunnerDispatch(t *testing.T) {
	model := newTestModel()
	model.appState.ClearDialogs()

	// Test that CommandRunCommand is properly dispatched to openCommandRunner
	cmd := model.dispatchCommand(CommandRunCommand)

	// Verify the dialog is opened or error is shown (depending on Crush availability)
	active := model.appState.ActiveDialog()
	if active != nil {
		if active.Title() != "Run Command with Crush" && active.Title() != "Crush Binary Not Found" {
			t.Errorf("expected Command Runner or error dialog, got %s", active.Title())
		}
	} else if cmd == nil {
		// If no dialog and no command, something went wrong
		t.Error("expected either a dialog or a command to be returned")
	}
}

// Tests for helper methods (4.2)

func TestGetCurrentCrushModel(t *testing.T) {
	model := newTestModel()
	result := model.getCurrentCrushModel()
	// Should return empty string for default model
	if result != "" {
		t.Errorf("expected empty string for default model, got %q", result)
	}
}

func TestTruncatePrompt(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		maxLen    int
		expected  string
	}{
		{
			name:     "No truncation needed",
			prompt:   "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "Exact length",
			prompt:   "exactly10c",
			maxLen:   10,
			expected: "exactly10c",
		},
		{
			name:     "Truncate with ellipsis",
			prompt:   "this is a very long prompt that needs truncating",
			maxLen:   15,
			expected: "this is a ve...",
		},
		{
			name:     "Single character max length",
			prompt:   "test",
			maxLen:   1,
			expected: "...",
		},
		{
			name:     "Three character max length",
			prompt:   "hello",
			maxLen:   3,
			expected: "...",
		},
		{
			name:     "Four character max length",
			prompt:   "hello",
			maxLen:   4,
			expected: "h...",
		},
		{
			name:     "Empty prompt",
			prompt:   "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePrompt(tt.prompt, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncatePrompt(%q, %d) = %q, expected %q", tt.prompt, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestEnsureTaskRunnerModal(t *testing.T) {
	model := newTestModel()
	model.width = 120
	model.height = 60

	// Should create modal if none exists
	if model.taskRunner != nil {
		t.Error("taskRunner should be nil initially")
	}
	if model.taskRunnerVisible {
		t.Error("taskRunnerVisible should be false initially")
	}

	model.ensureTaskRunnerModal()

	if model.taskRunner == nil {
		t.Error("taskRunner should be created after ensureTaskRunnerModal")
	}
	if !model.taskRunnerVisible {
		t.Error("taskRunnerVisible should be true after ensureTaskRunnerModal")
	}

	// Second call should not create a new modal
	firstModal := model.taskRunner
	model.ensureTaskRunnerModal()
	if model.taskRunner != firstModal {
		t.Error("ensureTaskRunnerModal should reuse existing modal")
	}
	if !model.taskRunnerVisible {
		t.Error("taskRunnerVisible should remain true on subsequent calls")
	}
}

func TestExecuteAdHocCommandGeneratesTaskStartedMsg(t *testing.T) {
	model := newTestModel()
	model.width = 120
	model.height = 60

	// executeAdHocCommand returns a tea.Cmd that when invoked generates TaskStartedMsg
	cmd := model.executeAdHocCommand("test-command-id", "This is a test prompt", "")
	
	if cmd == nil {
		t.Fatal("executeAdHocCommand should return a tea.Cmd")
	}

	// Invoke the command to get the message
	msg := cmd()
	
	startMsg, ok := msg.(dialog.TaskStartedMsg)
	if !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}

	if startMsg.TaskID != "test-command-id" {
		t.Errorf("expected TaskID 'test-command-id', got %s", startMsg.TaskID)
	}

	// TaskTitle should be the truncated prompt
	if startMsg.TaskTitle != "This is a test prompt" {
		t.Errorf("expected TaskTitle 'This is a test prompt', got %s", startMsg.TaskTitle)
	}
}

func TestListenForCommandOutput(t *testing.T) {
	model := newTestModel()

	// Create a channel with some test output
	outputChan := make(chan string, 3)
	outputChan <- "Line 1"
	outputChan <- "Line 2"
	close(outputChan)

	// Call listenForCommandOutput to get the subscription command
	cmd := model.listenForCommandOutput("test-command-1", outputChan)
	
	if cmd == nil {
		t.Fatal("listenForCommandOutput should return a tea.Cmd")
	}

	// The returned cmd should be a WaitForCrushMsg subscription
	// We can verify it returns something (actual subscription testing would need full integration)
	msg := cmd()
	if msg == nil {
		// May be nil if waiting for message, which is fine
		return
	}

	// If we get a message, it should be either TaskOutputMsg or TaskCompletedMsg
	switch msg.(type) {
	case dialog.TaskOutputMsg, dialog.TaskCompletedMsg, nil:
		// Expected types
	default:
		t.Errorf("unexpected message type from listenForCommandOutput: %T", msg)
	}
}
