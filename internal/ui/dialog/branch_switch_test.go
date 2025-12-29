package dialog

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestBranchItem tests the branchItem struct and its interface implementation
func TestBranchItem_Interface(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
	}{
		{
			name:        "main branch",
			title:       "main",
			description: "(current)",
		},
		{
			name:        "feature branch",
			title:       "feature/new-api",
			description: "",
		},
		{
			name:        "branch with special chars",
			title:       "bugfix/issue-123",
			description: "(current)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := branchItem{
				title:       test.title,
				description: test.description,
			}

			// Test Title()
			if item.Title() != test.title {
				t.Errorf("Title() = %q, want %q", item.Title(), test.title)
			}

			// Test Description()
			if item.Description() != test.description {
				t.Errorf("Description() = %q, want %q", item.Description(), test.description)
			}

			// Test FilterValue() - should return title for filtering
			if item.FilterValue() != test.title {
				t.Errorf("FilterValue() = %q, want %q", item.FilterValue(), test.title)
			}
		})
	}
}

// TestNewBranchSwitchDialog tests dialog creation with a real git repository
func TestNewBranchSwitchDialog_Creation(t *testing.T) {
	// Create a temporary git repository for testing
	tempDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Configure git for commits
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	// Create an initial commit so we have a branch
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Test successful dialog creation
	onSwitch := func(branch, output string, err error) {
		// Callback for testing
	}

	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch)
	if err != nil {
		t.Fatalf("NewBranchSwitchDialog failed: %v", err)
	}

	if dialog == nil {
		t.Fatal("NewBranchSwitchDialog returned nil")
	}

	// Verify dialog properties
	if dialog.Title() != "Switch Branch" {
		t.Errorf("Expected title 'Switch Branch', got %q", dialog.Title())
	}

	if dialog.repoPath != tempDir {
		t.Errorf("Expected repoPath %q, got %q", tempDir, dialog.repoPath)
	}

	if dialog.currentBranch == "" {
		t.Error("currentBranch should not be empty")
	}

	// Verify branches were loaded
	if len(dialog.branches) == 0 {
		t.Error("Expected at least one branch (master or main)")
	}

	// Verify current branch is in the list
	found := false
	for _, b := range dialog.branches {
		if b == dialog.currentBranch {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("currentBranch %q not found in branches list", dialog.currentBranch)
	}
}

// TestNewBranchSwitchDialog_InvalidPath tests error handling for invalid repository
func TestNewBranchSwitchDialog_InvalidPath(t *testing.T) {
	invalidPath := "/nonexistent/path"
	
	onSwitch := func(branch, output string, err error) {}
	dialog, err := NewBranchSwitchDialog(invalidPath, onSwitch)

	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}

	if dialog != nil {
		t.Error("Expected nil dialog for invalid path")
	}
}

// TestBranchSwitchDialog_HandleKey tests keyboard input handling
func TestBranchSwitchDialog_HandleKey(t *testing.T) {
	tempDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Configure git and create initial commit
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir

	// Create a feature branch
	exec.Command("git", "checkout", "-b", "feature/test").Dir = tempDir

	onSwitch := func(branch, output string, err error) {}
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch)
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected DialogResult
	}{
		{"escape key", "esc", DialogResultCancel},
		{"up arrow", "up", DialogResultNone},
		{"down arrow", "down", DialogResultNone},
		{"k key", "k", DialogResultNone},
		{"j key", "j", DialogResultNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(test.key[0])}}
			if test.key == "esc" {
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			} else if test.key == "up" {
				keyMsg = tea.KeyMsg{Type: tea.KeyUp}
			} else if test.key == "down" {
				keyMsg = tea.KeyMsg{Type: tea.KeyDown}
			}

			result, _ := dialog.HandleKey(keyMsg)
			if result != test.expected {
				t.Errorf("HandleKey(%q) returned %v, expected %v", test.key, result, test.expected)
			}
		})
	}
}

// TestBranchSwitchDialog_View tests the view rendering
func TestBranchSwitchDialog_View(t *testing.T) {
	tempDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Configure git and create initial commit
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir

	onSwitch := func(branch, output string, err error) {}
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch)
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	// Set dialog size through Center method
	dialog.Center(80, 30)

	// Test view renders without panic
	view := dialog.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Test view with switching state
	dialog.switching = true
	view = dialog.View()
	if view == "" {
		t.Error("View() returned empty string when switching")
	}
}

// TestBranchItem_FilterValue tests branch filtering
func TestBranchItem_FilterValue(t *testing.T) {
	tests := []struct {
		title       string
		description string
		expected    string
	}{
		{"main", "(current)", "main"},
		{"feature/api", "", "feature/api"},
		{"bugfix/issue-456", "", "bugfix/issue-456"},
	}

	for _, test := range tests {
		item := branchItem{
			title:       test.title,
			description: test.description,
		}

		if item.FilterValue() != test.expected {
			t.Errorf("FilterValue() for title %q returned %q, expected %q",
				test.title, item.FilterValue(), test.expected)
		}
	}
}
