package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
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
	dialog, err := NewBranchSwitchDialog(invalidPath, onSwitch, "")

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
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
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
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
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

// TestBranchSwitchDialog_LaunchGitSwitchBranch tests the launchGitSwitchBranch method
func TestBranchSwitchDialog_LaunchGitSwitchBranch(t *testing.T) {
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
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "test-tag")
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	// Test launchGitSwitchBranch returns a command
	cmd1 := dialog.launchGitSwitchBranch("master")
	if cmd1 == nil {
		t.Error("launchGitSwitchBranch returned nil command")
	}

	// Execute the command and verify it produces a message
	msg := cmd1()
	if msg == nil {
		t.Error("Command execution returned nil message")
	}

	// Verify the message is a sequenceMsg (from tea.Sequence)
	// We just verify it's not nil - tea.Sequence returns a sequenceMsg type
	// which is used internally to chain commands
	if reflect.TypeOf(msg).String() == "<nil>" {
		t.Error("Expected non-nil message from sequenced command")
	}
}

// TestBranchSwitchDialog_Enter_StartsSwitch tests Enter key triggers switch
func TestBranchSwitchDialog_Enter_StartsSwitch(t *testing.T) {
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
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	// Center the dialog to ensure proper size
	dialog.Center(80, 30)

	// Select a branch that's not current
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	dialog.list.Update(keyMsg)

	// Test Enter key
	if dialog.list.SelectedItem() != nil {
		result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
		
		// Should return DialogResultClose to close the dialog
		if result != DialogResultClose {
			t.Errorf("HandleKey(Enter) returned %v, expected DialogResultClose", result)
		}

		// Should return a command
		if cmd == nil {
			t.Error("HandleKey(Enter) returned nil command")
		}

		// Should set switching to true
		if !dialog.switching {
			t.Error("Expected switching to be true after Enter key")
		}
	}
}

// TestBranchSwitchDialog_TaskCompletedMsg_NilSelection tests PHASE 1 FIX
// Ensures no panic when TaskCompletedMsg arrives with nil selection
func TestBranchSwitchDialog_TaskCompletedMsg_NilSelection(t *testing.T) {
	// Create a temporary git repository for testing
	tempDir := t.TempDir()

	// Initialize git repo and create initial commit
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir
	exec.Command("git", "checkout", "-b", "feature/test").Dir = tempDir

	callbackCalled := false
	onSwitch := func(branch, output string, err error) {
		callbackCalled = true
	}
	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	dialog.Center(80, 30)

	// Set switching state and cache selected branch
	dialog.mu.Lock()
	dialog.switching = true
	dialog.currentTaskID = "git-switch-branch"
	dialog.selectedBranch = "feature/test"
	dialog.mu.Unlock()

	// Simulate TaskCompletedMsg arrival
	// This was the critical crash scenario in the bug report
	msg := TaskCompletedMsg{TaskID: "git-switch-branch"}
	updatedDialog, _ := dialog.Update(msg)

	// CRITICAL TEST: Should handle gracefully and return nil to close dialog
	if updatedDialog != nil {
		t.Error("Expected dialog to close (return nil) after TaskCompletedMsg")
	}

	// Callback should have been called with cached branch name
	if !callbackCalled {
		t.Error("Expected onSwitch callback to be called")
	}
}

// TestBranchSwitchDialog_TaskCompletedMsg_WrongTaskID tests PHASE 2 FIX
// Ensures orphaned messages are ignored (prevents stale completion messages)
func TestBranchSwitchDialog_TaskCompletedMsg_WrongTaskID(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir

	onSwitchCalled := false
	onSwitch := func(branch, output string, err error) {
		onSwitchCalled = true
	}

	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	dialog.Center(80, 30)

	// Set state with different task ID
	dialog.mu.Lock()
	dialog.switching = true
	dialog.currentTaskID = "other-task"
	dialog.selectedBranch = "feature/test"
	dialog.mu.Unlock()

	// Send TaskCompletedMsg with mismatched task ID
	// This ensures we don't process orphaned messages from previous operations
	msg := TaskCompletedMsg{TaskID: "git-switch-branch"}
	updatedDialog, _ := dialog.Update(msg)

	// Should NOT call onSwitch callback for mismatched task ID
	if onSwitchCalled {
		t.Error("onSwitch was called for mismatched task ID")
	}

	// Should keep switching state unchanged
	dialog.mu.RLock()
	isSwitching := dialog.switching
	taskID := dialog.currentTaskID
	dialog.mu.RUnlock()

	if !isSwitching {
		t.Error("Expected switching to remain true when task ID doesn't match")
	}

	if taskID != "other-task" {
		t.Error("Expected currentTaskID to remain unchanged")
	}

	// Dialog should still exist
	if updatedDialog == nil {
		t.Error("Expected dialog to remain active for mismatched message")
	}
}

// TestBranchSwitchDialog_MultipleSwitches tests PHASE 1 + PHASE 2 FIX
// Ensures multiple TaskCompletedMsg handlers don't crash
func TestBranchSwitchDialog_MultipleSwitches(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo with multiple branches
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir

	// Create feature branches
	exec.Command("git", "checkout", "-b", "feature/one").Dir = tempDir
	exec.Command("git", "checkout", "-b", "feature/two").Dir = tempDir
	exec.Command("git", "checkout", "-b", "feature/three").Dir = tempDir

	switchCount := 0
	onSwitch := func(branch, output string, err error) {
		switchCount++
	}

	// Test multiple sequential TaskCompletedMsg handlers
	// This tests the critical fix for orphaned messages
	for i := 0; i < 3; i++ {
		dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
		if err != nil {
			t.Fatalf("Iteration %d: Failed to create dialog: %v", i, err)
		}

		dialog.Center(80, 30)

		// Set up state as if a switch was initiated
		dialog.mu.Lock()
		dialog.switching = true
		dialog.currentTaskID = "git-switch-branch"
		dialog.selectedBranch = fmt.Sprintf("feature/test%d", i)
		dialog.mu.Unlock()

		// CRITICAL TEST: Send TaskCompletedMsg - should not crash
		// even if list selection is nil or invalid
		msg := TaskCompletedMsg{TaskID: "git-switch-branch"}
		updatedDialog, _ := dialog.Update(msg)

		// Should close dialog
		if updatedDialog != nil {
			t.Errorf("Iteration %d: Expected dialog to close", i)
		}

		// Should call callback
		if switchCount != i+1 {
			t.Errorf("Iteration %d: Expected switchCount to be %d, got %d", i, i+1, switchCount)
		}
	}

	// Should have completed all without crashing
	if switchCount != 3 {
		t.Errorf("Expected 3 switches, got %d", switchCount)
	}
}

// TestBranchSwitchDialog_ThreadSafety tests PHASE 2 FIX
// Ensures concurrent access doesn't cause data races
func TestBranchSwitchDialog_ThreadSafety(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir
	exec.Command("git", "checkout", "-b", "feature/test").Dir = tempDir

	onSwitch := func(branch, output string, err error) {}

	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	dialog.Center(80, 30)

	// Simulate concurrent operations using goroutines
	done := make(chan bool, 3)

	// Goroutine 1: View rendering (simulates UI updates)
	go func() {
		for i := 0; i < 100; i++ {
			_ = dialog.View()
		}
		done <- true
	}()

	// Goroutine 2: Message processing
	go func() {
		for i := 0; i < 100; i++ {
			msg := TaskOutputMsg{
				TaskID: "git-switch-branch",
				Output: "test output",
			}
			dialog.Update(msg)
		}
		done <- true
	}()

	// Goroutine 3: State modification
	go func() {
		for i := 0; i < 100; i++ {
			keyMsg := tea.KeyMsg{Type: tea.KeyDown}
			dialog.list.Update(keyMsg)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done

	// If we reach here without panic or data race, thread safety is working
	// (Note: This test is best run with -race flag: go test -race)
}

// TestBranchSwitchDialog_HandleKey_SafeCasting tests PHASE 1 FIX
// Ensures safe type assertion in HandleKey
func TestBranchSwitchDialog_HandleKey_SafeCasting(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir

	onSwitch := func(branch, output string, err error) {}

	dialog, err := NewBranchSwitchDialog(tempDir, onSwitch, "")
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	dialog.Center(80, 30)

	// Test 1: No selected item
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultNone {
		t.Error("Expected DialogResultNone when no item selected")
	}

	// Test 2: Select a valid item and press Enter
	dialog.list.Update(tea.KeyMsg{Type: tea.KeyDown})
	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Should either start switch or return none (if already on current branch)
	if result != DialogResultClose && result != DialogResultNone {
		t.Errorf("Expected DialogResultClose or DialogResultNone, got %v", result)
	}

	// Test 3: Escape key should close
	dialog.mu.Lock()
	dialog.switching = false
	dialog.mu.Unlock()

	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel for Escape key, got %v", result)
	}
}
