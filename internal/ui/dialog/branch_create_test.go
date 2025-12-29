package dialog

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestBranchCreateDialogCreation tests dialog initialization
func TestBranchCreateDialogCreation(t *testing.T) {
	onComplete := func(branchName, output string, err error) {}
	dialog := NewBranchCreateDialog("/test/repo", onComplete)

	if dialog == nil {
		t.Errorf("Expected non-nil dialog, got nil")
	}
	if dialog.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath to be /test/repo, got %s", dialog.repoPath)
	}
	if dialog.TitleText != "Create Branch" {
		t.Errorf("Expected title Create Branch, got %s", dialog.TitleText)
	}
	if dialog.loading {
		t.Errorf("Expected loading to be false initially")
	}
}

// TestIsValidBranchName tests branch name validation
func TestIsValidBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		valid      bool
	}{
		{"valid simple name", "feature-branch", true},
		{"valid with slashes", "feature/new-feature", true},
		{"valid with underscores", "feature_branch", true},
		{"valid with numbers", "feature-123", true},
		{"empty name", "", false},
		{"name with spaces", "feature branch", false},
		{"name with tabs", "feature\tbranch", false},
		{"name with colon", "feature:branch", false},
		{"name with tilde", "feature~branch", false},
		{"name with caret", "feature^branch", false},
		{"name with question mark", "feature?branch", false},
		{"name with asterisk", "feature*branch", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewBranchCreateDialog("/test/repo", nil)
			dialog.input.SetValue(tt.branchName)

			if dialog.isValidBranchName() != tt.valid {
				t.Errorf("Expected isValidBranchName(%q) to be %v, got %v",
					tt.branchName, tt.valid, !tt.valid)
			}
		})
	}
}

// TestValidateAndUpdateError tests error message updates
func TestValidateAndUpdateError(t *testing.T) {
	tests := []struct {
		name            string
		branchName      string
		expectedMessage string
	}{
		{"empty name", "", ""},
		{"valid name", "valid-branch", ""},
		{"name with spaces", "invalid branch", "Branch name cannot contain spaces"},
		{"name with colons", "invalid:branch", "Branch name cannot contain colons"},
		{"name with tilde", "invalid~branch", "Branch name contains invalid characters"},
		{"name with caret", "invalid^branch", "Branch name contains invalid characters"},
		{"name with special chars", "invalid?branch", "Branch name contains invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewBranchCreateDialog("/test/repo", nil)
			dialog.input.SetValue(tt.branchName)
			dialog.validateAndUpdateError()

			if dialog.errorMsg != tt.expectedMessage {
				t.Errorf("Expected error message %q, got %q", tt.expectedMessage, dialog.errorMsg)
			}
		})
	}
}

// TestInit tests the Init method
func TestBranchCreateDialogInit(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	cmd := dialog.Init()

	// Verify Init returns a command (it should return textinput.Blink)
	if cmd == nil {
		t.Errorf("Expected Init to return a non-nil command")
	}
}

// TestUpdateWithEscapeKey tests that Escape closes the dialog
func TestUpdateWithEscapeKey(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.input.SetValue("test-branch")

	msg := tea.KeyMsg{
		Type: tea.KeyEsc,
	}

	result, _ := dialog.Update(msg)

	if result != nil {
		t.Errorf("Expected Update to return nil dialog for Escape key")
	}
}

// TestUpdateWithValidInputAndEnter tests successful branch creation
func TestUpdateWithValidInputAndEnter(t *testing.T) {
	onComplete := func(branchName, output string, err error) {
	}

	dialog := NewBranchCreateDialog("/test/repo", onComplete)
	dialog.input.SetValue("test-branch")

	// Simulate enter key
	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, cmd := dialog.Update(enterMsg)

	// After pressing enter with valid input, dialog should still be open with loading state
	if result == nil {
		t.Errorf("Expected Update to return non-nil dialog during loading")
	}
	if !dialog.loading {
		t.Errorf("Expected loading to be true after pressing Enter")
	}
	if cmd == nil {
		t.Errorf("Expected Update to return a non-nil command")
	}
}

// TestUpdateWithInvalidInputAndEnter tests validation on enter
func TestUpdateWithInvalidInputAndEnter(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.input.SetValue("")

	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, _ := dialog.Update(enterMsg)

	// Dialog should remain open with error message
	if result == nil {
		t.Errorf("Expected Update to return non-nil dialog for invalid input")
	}
	if !strings.Contains(dialog.errorMsg, "Invalid") {
		t.Errorf("Expected error message to contain 'Invalid', got %q", dialog.errorMsg)
	}
}

// TestSetRect tests rectangle setting
func TestBranchCreateDialogSetRect(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.SetRect(80, 24, 10, 5)

	w, h, x, y := dialog.GetRect()

	if w != 80 {
		t.Errorf("Expected width 80, got %d", w)
	}
	if h != 24 {
		t.Errorf("Expected height 24, got %d", h)
	}
	if x != 10 {
		t.Errorf("Expected x 10, got %d", x)
	}
	if y != 5 {
		t.Errorf("Expected y 5, got %d", y)
	}
}

// TestInputWidthAdjustment tests that input width adjusts with dialog size
func TestInputWidthAdjustment(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)

	dialog.SetRect(100, 24, 0, 0)
	inputWidth1 := dialog.input.Width

	dialog.SetRect(50, 24, 0, 0)
	inputWidth2 := dialog.input.Width

	if inputWidth1 == inputWidth2 {
		t.Errorf("Expected input width to change with dialog width, but both are %d", inputWidth1)
	}
	if inputWidth2 < inputWidth1 {
		t.Logf("Input width correctly adjusted: %d -> %d", inputWidth1, inputWidth2)
	}
}

// TestViewWithValidInput tests View method with valid input
func TestViewWithValidInput(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.input.SetValue("valid-branch")

	view := dialog.View()

	if !strings.Contains(view, "valid-branch") {
		t.Errorf("Expected view to contain branch name")
	}
	if !strings.Contains(view, "✓ Ready") {
		t.Errorf("Expected view to show ready indicator for valid input")
	}
}

// TestViewWithInvalidInput tests View method with invalid input
func TestViewWithInvalidInput(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.input.SetValue("invalid branch")

	view := dialog.View()

	if !strings.Contains(view, "Invalid name") {
		t.Errorf("Expected view to show invalid name message")
	}
}

// TestViewWithLoading tests View method during loading
func TestViewWithLoading(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.loading = true

	view := dialog.View()

	if !strings.Contains(view, "Creating branch") {
		t.Errorf("Expected view to show creating message during loading")
	}
}

// TestHandleKey tests HandleKey method
func TestBranchCreateDialogHandleKey(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.input.SetValue("valid-branch")

	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, _ := dialog.HandleKey(enterMsg)

	// HandleKey should return DialogResultNone when starting creation
	if result != DialogResultNone {
		t.Errorf("Expected HandleKey to return DialogResultNone")
	}
}

// BenchmarkIsValidBranchName benchmarks the validation function
func BenchmarkIsValidBranchName(b *testing.B) {
	dialog := NewBranchCreateDialog("/test/repo", nil)
	dialog.input.SetValue("feature-branch-name")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialog.isValidBranchName()
	}
}
