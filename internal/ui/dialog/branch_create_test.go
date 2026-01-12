package dialog

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestBranchCreateDialogCreation tests dialog initialization
func TestBranchCreateDialogCreation(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "test-tag")

	if dialog == nil {
		t.Errorf("Expected non-nil dialog, got nil")
	}
	if dialog.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath to be /test/repo, got %s", dialog.repoPath)
	}
	if dialog.tagName != "test-tag" {
		t.Errorf("Expected tagName to be test-tag, got %s", dialog.tagName)
	}
	if dialog.TitleText != "Create Branch" {
		t.Errorf("Expected title Create Branch, got %s", dialog.TitleText)
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
			dialog := NewBranchCreateDialog("/test/repo", "")
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
			dialog := NewBranchCreateDialog("/test/repo", "")
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
	dialog := NewBranchCreateDialog("/test/repo", "")
	cmd := dialog.Init()

	// Verify Init returns a command (it should return textinput.Blink)
	if cmd == nil {
		t.Errorf("Expected Init to return a non-nil command")
	}
}

// TestUpdateWithEscapeKey tests that Update does not handle Escape key
// Escape key handling is delegated to HandleKey
func TestUpdateWithEscapeKey(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("test-branch")

	msg := tea.KeyMsg{
		Type: tea.KeyEsc,
	}

	result, _ := dialog.Update(msg)

	// Update should forward to textinput.Update, not handle Esc itself
	// Dialog should remain open
	if result == nil {
		t.Errorf("Expected Update to return non-nil dialog (forwards to textinput)")
	}
}

// TestUpdateWithValidInputAndEnter tests that Update does not handle Enter key
// Enter key handling is delegated to HandleKey
func TestUpdateWithValidInputAndEnter(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("test-branch")

	// Simulate enter key
	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, _ := dialog.Update(enterMsg)

	// Update should forward to textinput.Update, not handle Enter itself
	// Dialog should remain open
	if result == nil {
		t.Errorf("Expected Update to return non-nil dialog (forwards to textinput)")
	}
}

// TestUpdateWithInvalidInputAndEnter tests that Update does not handle Enter key
// Enter key handling is delegated to HandleKey
func TestUpdateWithInvalidInputAndEnter(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("")

	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, _ := dialog.Update(enterMsg)

	// Update should forward to textinput.Update, not validate
	// Dialog should remain open without error message
	if result == nil {
		t.Errorf("Expected Update to return non-nil dialog (forwards to textinput)")
	}
	// errorMsg should be cleared by validateAndUpdateError
	// since the input is empty
	if dialog.errorMsg != "" {
		t.Errorf("Expected error message to be empty for empty input, got %q", dialog.errorMsg)
	}
}

// TestSetRect tests rectangle setting
func TestBranchCreateDialogSetRect(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
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
	dialog := NewBranchCreateDialog("/test/repo", "")

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
	dialog := NewBranchCreateDialog("/test/repo", "")
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
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("invalid branch")

	view := dialog.View()

	if !strings.Contains(view, "Invalid name") {
		t.Errorf("Expected view to show invalid name message")
	}
}

// TestHandleKeyWithValidBranch tests HandleKey creates branch on Enter
func TestBranchCreateDialogHandleKey(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("valid-branch")

	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, cmd := dialog.HandleKey(enterMsg)

	// HandleKey should return DialogResultClose to close dialog
	if result != DialogResultClose {
		t.Errorf("Expected HandleKey to return DialogResultClose, got %v", result)
	}
	// Should return a command to create the branch
	if cmd == nil {
		t.Errorf("Expected HandleKey to return a non-nil command for branch creation")
	}
}

// TestHandleKeyWithInvalidBranch tests HandleKey rejects invalid branch names
func TestHandleKeyWithInvalidBranch(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("invalid branch")

	enterMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	result, cmd := dialog.HandleKey(enterMsg)

	// Should not create command for invalid branch
	if cmd != nil {
		t.Errorf("Expected HandleKey to return nil command for invalid branch")
	}
	// Should return DialogResultNone
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for invalid branch")
	}
	// Should set error message
	if !strings.Contains(dialog.errorMsg, "Invalid") {
		t.Errorf("Expected error message to contain 'Invalid', got %q", dialog.errorMsg)
	}
}

// TestLaunchGitCreateBranch tests launchGitCreateBranch returns a command
func TestLaunchGitCreateBranch(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "test-tag")
	cmd := dialog.launchGitCreateBranch("test-branch")

	// launchGitCreateBranch should return a non-nil tea.Cmd
	if cmd == nil {
		t.Errorf("Expected launchGitCreateBranch to return a non-nil tea.Cmd")
	}
}

// TestLaunchGitCreateBranchWithoutTag tests launchGitCreateBranch with empty tag
func TestLaunchGitCreateBranchWithoutTag(t *testing.T) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	cmd := dialog.launchGitCreateBranch("feature-branch")

	// Should still return a valid command even without a tag
	if cmd == nil {
		t.Errorf("Expected launchGitCreateBranch to return a non-nil tea.Cmd even without tag")
	}
}

// BenchmarkIsValidBranchName benchmarks the validation function
func BenchmarkIsValidBranchName(b *testing.B) {
	dialog := NewBranchCreateDialog("/test/repo", "")
	dialog.input.SetValue("feature-branch-name")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialog.isValidBranchName()
	}
}
