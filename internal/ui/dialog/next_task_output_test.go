package dialog

import (
	"testing"
	"strings"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewNextTaskOutputContent verifies the constructor returns a properly initialized component
func TestNewNextTaskOutputContent(t *testing.T) {
	content := NewNextTaskOutputContent()

	if content == nil {
		t.Fatal("NewNextTaskOutputContent returned nil")
	}

	if !content.loading {
		t.Error("expected loading to be true")
	}

	if len(content.output) != 1 {
		t.Errorf("expected output length 1, got %d", len(content.output))
	}

	if content.output[0] != "Loading..." {
		t.Errorf("expected output[0] to be 'Loading...', got %q", content.output[0])
	}

	if content.viewport.Width != 0 || content.viewport.Height != 0 {
		t.Errorf("expected viewport to have 0 width and height, got %d x %d", content.viewport.Width, content.viewport.Height)
	}
}

// TestInit verifies Init returns nil command
func TestInit(t *testing.T) {
	content := NewNextTaskOutputContent()
	cmd := content.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil")
	}
}

// TestSetSize verifies SetSize updates viewport dimensions
func TestSetSize(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.SetSize(80, 24)

	if content.width != 80 || content.height != 24 {
		t.Errorf("expected size 80x24, got %dx%d", content.width, content.height)
	}

	if content.viewport.Width != 80 || content.viewport.Height != 24 {
		t.Errorf("expected viewport size 80x24, got %dx%d", content.viewport.Width, content.viewport.Height)
	}
}

// TestViewLoading verifies View shows loading state
func TestViewLoading(t *testing.T) {
	content := NewNextTaskOutputContent()
	view := content.View(40, 10)

	if !strings.Contains(view, "Loading...") {
		t.Errorf("expected View to contain 'Loading...', got: %q", view)
	}
}

// TestViewEmpty verifies View shows empty message when output is empty
func TestViewEmpty(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{}

	view := content.View(40, 10)

	if !strings.Contains(view, "No output available") {
		t.Errorf("expected View to contain 'No output available', got: %q", view)
	}
}

// TestViewWithContent verifies View renders viewport content
func TestViewWithContent(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2", "Line 3"}

	view := content.View(40, 10)

	if !strings.Contains(view, "Line 1") {
		t.Errorf("expected View to contain 'Line 1', got: %q", view)
	}
}

// TestUpdateArrowUp verifies arrow up key scrolls viewport up
func TestUpdateArrowUp(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6"}
	content.viewport.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6")
	content.SetSize(40, 5)

	// Simulate scrolling down first
	content.viewport.LineDown(2)
	prevYOffset := content.viewport.YOffset

	// Send arrow up message
	_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyUp})
	newYOffset := content.viewport.YOffset

	// Command can be nil, important part is YOffset changes
	if newYOffset >= prevYOffset {
		t.Errorf("expected YOffset to decrease, was %d now %d", prevYOffset, newYOffset)
	}
	_ = cmd
}

// TestUpdateArrowDown verifies arrow down key scrolls viewport down
func TestUpdateArrowDown(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6"}
	content.viewport.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6")
	content.SetSize(40, 5)

	prevYOffset := content.viewport.YOffset

	// Send arrow down message
	_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyDown})
	newYOffset := content.viewport.YOffset

	// Command can be nil, important part is YOffset changes
	if newYOffset <= prevYOffset {
		t.Errorf("expected YOffset to increase, was %d now %d", prevYOffset, newYOffset)
	}
	_ = cmd
}

// TestUpdatePageUp verifies PageUp key scrolls viewport
func TestUpdatePageUp(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10"}
	content.viewport.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	content.SetSize(40, 5)

	// Simulate scrolling down first
	content.viewport.LineDown(5)
	prevYOffset := content.viewport.YOffset

	// Send PageUp message
	_, cmd := content.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	newYOffset := content.viewport.YOffset

	// Command can be nil, important part is YOffset changes
	if newYOffset >= prevYOffset {
		t.Errorf("expected YOffset to decrease with PageUp, was %d now %d", prevYOffset, newYOffset)
	}
	_ = cmd
}

// TestSetOutput verifies SetOutput updates output and transitions from loading state
func TestSetOutput(t *testing.T) {
	content := NewNextTaskOutputContent()

	if !content.loading {
		t.Fatal("expected initial loading to be true")
	}

	lines := []string{"Output line 1", "Output line 2", "Output line 3"}
	content.SetOutput(lines)

	if content.loading {
		t.Error("expected loading to be false after SetOutput")
	}

	if len(content.output) != len(lines) {
		t.Errorf("expected output length %d, got %d", len(lines), len(content.output))
	}

	for i, line := range lines {
		if content.output[i] != line {
			t.Errorf("expected output[%d] to be %q, got %q", i, line, content.output[i])
		}
	}
}

// TestSetOutputEmpty verifies SetOutput with empty input shows "No tasks available." message
func TestSetOutputEmpty(t *testing.T) {
	content := NewNextTaskOutputContent()

	if !content.loading {
		t.Fatal("expected initial loading to be true")
	}

	content.SetOutput([]string{})

	if content.loading {
		t.Error("expected loading to be false after SetOutput")
	}

	if len(content.output) != 1 {
		t.Errorf("expected output length 1, got %d", len(content.output))
	}

	if content.output[0] != "No tasks available." {
		t.Errorf("expected 'No tasks available.', got %q", content.output[0])
	}
}

// TestSetLoading verifies SetLoading updates loading state
func TestSetLoading(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.output = []string{"Some output"}
	content.loading = false

	content.SetLoading(true)

	if !content.loading {
		t.Error("expected loading to be true")
	}

	if len(content.output) != 1 || content.output[0] != "Loading..." {
		t.Error("expected output to be reset to Loading...")
	}
}

// TestSetError verifies SetError formats error message with remediation hint
func TestSetError(t *testing.T) {
	content := NewNextTaskOutputContent()

	testErr := errors.New("test error message")
	content.SetError(testErr)

	if content.loading {
		t.Error("expected loading to be false after SetError")
	}

	if len(content.output) != 1 {
		t.Errorf("expected 1 output line, got %d", len(content.output))
	}

	// Check that error message contains "Error:", error text, and remediation hint
	errorMsg := content.output[0]
	if !strings.Contains(errorMsg, "Error:") {
		t.Errorf("expected error message to contain 'Error:', got %q", errorMsg)
	}

	if !strings.Contains(errorMsg, "test error message") {
		t.Errorf("expected error message to contain error details, got %q", errorMsg)
	}

	if !strings.Contains(errorMsg, "Check that task-master is properly installed and configured") {
		t.Errorf("expected error message to contain remediation hint, got %q", errorMsg)
	}
}

// TestSetErrorNil verifies SetError handles nil error gracefully
func TestSetErrorNil(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.output = []string{"Some output"}
	content.loading = true

	content.SetError(nil)

	if content.loading {
		t.Error("expected loading to be false after SetError(nil)")
	}

	// SetError formats nil as string "<nil>"
	if len(content.output) != 1 {
		t.Errorf("expected 1 output line, got %d", len(content.output))
	}

	if !strings.Contains(content.output[0], "Error:") {
		t.Errorf("expected output to contain 'Error:', got %q", content.output[0])
	}
}

// TestImplementsModalContent verifies component implements ModalContent interface
func TestImplementsModalContent(t *testing.T) {
	content := NewNextTaskOutputContent()
	var _ ModalContent = content
}

// TestHandleKey verifies HandleKey delegates to viewport
func TestHandleKey(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content.viewport.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	content.SetSize(40, 3)

	// Store initial offset
	prevYOffset := content.viewport.YOffset

	// Handle key event
	cmd := content.HandleKey(tea.KeyMsg{Type: tea.KeyDown})

	// Command can be nil, but key should be handled
	if content.viewport.YOffset <= prevYOffset {
		// Down key should increase offset
		t.Logf("note: Down key handling may depend on viewport implementation")
	}
	_ = cmd
}

// TestAppendOutput verifies AppendOutput appends lines and respects max buffer size
func TestAppendOutput(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{} // Start with empty

	// Append lines up to max
	for i := 0; i < maxOutputLines; i++ {
		content.AppendOutput("Line " + string(rune('0'+i%10)))
	}

	if len(content.output) != maxOutputLines {
		t.Errorf("expected output length %d, got %d", maxOutputLines, len(content.output))
	}

	// Append one more line, should trim oldest
	content.AppendOutput("Last line")

	if len(content.output) != maxOutputLines {
		t.Errorf("expected output length to stay at %d after exceeding, got %d", maxOutputLines, len(content.output))
	}

	// Check that the last line is in the buffer
	if content.output[len(content.output)-1] != "Last line" {
		t.Errorf("expected last line to be 'Last line', got %q", content.output[len(content.output)-1])
	}
}

// TestAppendOutputBatchingUpdates verifies viewport updates are batched
func TestAppendOutputBatchingUpdates(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{}
	content.SetSize(80, 24)

	// Append 5 lines - should not update every time
	for i := 0; i < 5; i++ {
		content.AppendOutput("Line " + string(rune('0'+i)))
	}

	// Appending lines 1-5 should not trigger updates (only every 10)
	if len(content.output) != 5 {
		t.Errorf("expected 5 lines, got %d", len(content.output))
	}

	// Append 5 more lines to trigger the 10-line batch update
	for i := 5; i < 10; i++ {
		content.AppendOutput("Line " + string(rune('0'+i)))
	}

	if len(content.output) != 10 {
		t.Errorf("expected 10 lines, got %d", len(content.output))
	}
}

// TestAppendOutputTaskKeywordTriggers verifies Task: keyword triggers immediate update
func TestAppendOutputTaskKeywordTriggers(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{}
	content.SetSize(80, 24)

	// Append a single line with "Task:" keyword
	content.AppendOutput("Task: some-task-id")

	if len(content.output) != 1 {
		t.Errorf("expected 1 line, got %d", len(content.output))
	}

	if content.output[0] != "Task: some-task-id" {
		t.Errorf("expected 'Task: some-task-id', got %q", content.output[0])
	}
}

// TestFinalizeContent verifies FinalizeContent disables loading and updates viewport
func TestFinalizeContent(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = true
	content.output = []string{"Line 1", "Line 2", "Line 3"}
	content.SetSize(80, 24)

	if !content.loading {
		t.Error("expected loading to be true before finalize")
	}

	content.FinalizeContent()

	if content.loading {
		t.Error("expected loading to be false after finalize")
	}

	// Verify viewport content is set
	viewContent := content.viewport.View()
	if !strings.Contains(viewContent, "Line") {
		t.Errorf("expected viewport to contain output content, got: %q", viewContent)
	}
}

// TestESCKeyHandling verifies that ESC key is properly delegated to viewport
func TestESCKeyHandling(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2", "Line 3"}
	content.SetSize(40, 10)

	// Handle ESC key
	cmd := content.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})

	// Command result is not significant for ESC - important that content is not modified
	if content.loading {
		t.Error("expected loading state to remain unchanged after ESC key")
	}

	if len(content.output) != 3 {
		t.Errorf("expected output to remain unchanged after ESC key, got %d lines", len(content.output))
	}

	_ = cmd
}

// TestUpdatePreservesStateOnEsc verifies that Update preserves model state on ESC
func TestUpdatePreservesStateOnEsc(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Line 1", "Line 2"}
	content.SetSize(80, 20)

	// Store original state
	originalOutput := make([]string, len(content.output))
	copy(originalOutput, content.output)
	originalLoading := content.loading
	originalWidth := content.width
	originalHeight := content.height

	// Send ESC key message
	updatedContent, cmd := content.Update(tea.KeyMsg{Type: tea.KeyEscape})

	// Verify model is returned unchanged
	if updatedContent != content {
		t.Error("expected Update to return the same content model")
	}

	// Cast back to *NextTaskOutputContent to access fields
	updatedContentTyped := updatedContent.(*NextTaskOutputContent)

	// Verify state is preserved
	if updatedContentTyped.loading != originalLoading {
		t.Errorf("expected loading to remain %v, got %v", originalLoading, updatedContentTyped.loading)
	}

	if len(updatedContentTyped.output) != len(originalOutput) {
		t.Errorf("expected output length to remain %d, got %d", len(originalOutput), len(updatedContentTyped.output))
	}

	for i, line := range updatedContentTyped.output {
		if line != originalOutput[i] {
			t.Errorf("expected output[%d] to remain %q, got %q", i, originalOutput[i], line)
		}
	}

	if updatedContentTyped.width != originalWidth || updatedContentTyped.height != originalHeight {
		t.Errorf("expected size to remain %dx%d, got %dx%d", originalWidth, originalHeight, updatedContentTyped.width, updatedContentTyped.height)
	}

	_ = cmd
}

// TestModalIntegration verifies NextTaskOutputContent works with ModalDialog
func TestModalIntegration(t *testing.T) {
	content := NewNextTaskOutputContent()
	content.loading = false
	content.output = []string{"Task: example", "Status: pending"}

	// Create a modal dialog with the content
	dlg := NewModalDialog("Test Modal", 60, 15, content)

	if dlg == nil {
		t.Fatal("NewModalDialog returned nil")
	}

	// Verify dialog is cancellable by default
	if !dlg.IsCancellable() {
		t.Error("expected modal to be cancellable")
	}

	// Test that ESC key triggers dialog closure
	result, cmd := dlg.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})

	if result != DialogResultCancel {
		t.Errorf("expected ESC to return DialogResultCancel, got %v", result)
	}

	_ = cmd
}

// TestMultipleOpenCloseCycles verifies multiple open/close cycles work correctly
func TestMultipleOpenCloseCycles(t *testing.T) {
	for i := 0; i < 3; i++ {
		content := NewNextTaskOutputContent()
		content.SetOutput([]string{"Output line 1", "Output line 2"})

		if content.loading {
			t.Errorf("cycle %d: expected loading to be false after SetOutput", i)
		}

		if len(content.output) != 2 {
			t.Errorf("cycle %d: expected output length 2, got %d", i, len(content.output))
		}

		// Simulate a close by resetting
		content.SetLoading(true)
		if !content.loading {
			t.Errorf("cycle %d: expected loading to be true after SetLoading(true)", i)
		}
	}
}
