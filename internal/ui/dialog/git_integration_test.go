package dialog

import (
	"errors"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/git"
	tea "github.com/charmbracelet/bubbletea"
)

// TestGitMenuToStatusDialog_IntegrationFlow tests the complete workflow
// from opening the git menu to viewing the status dialog
func TestGitMenuToStatusDialog_IntegrationFlow(t *testing.T) {
	// Step 1: Create git menu dialog (no callback needed with message-based approach)
	menuDialog := NewGitMenuDialog(nil)
	if menuDialog == nil {
		t.Fatal("Failed to create git menu dialog")
	}

	// Step 2: Verify initial state
	if menuDialog.selectedIndex != 0 {
		t.Errorf("Expected initial selection at index 0, got %d", menuDialog.selectedIndex)
	}

	// Step 3: Simulate navigation to "Show Status" (already at index 0)
	// No navigation needed as it's the first item

	// Step 4: Simulate Enter key press to select
	result, cmd := menuDialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultClose {
		t.Errorf("Expected DialogResultClose after selection, got %v", result)
	}

	// Step 5: Execute the command to get the selection message
	if cmd == nil {
		t.Fatal("Expected command to be returned from HandleKey")
	}
	msg := cmd()
	selectionMsg, ok := msg.(GitMenuSelectionMsg)
	if !ok {
		t.Fatalf("Expected GitMenuSelectionMsg, got %T", msg)
	}

	if selectionMsg.SelectedIndex != 0 {
		t.Errorf("Expected selected ID 0 (Show Status), got %d", selectionMsg.SelectedIndex)
	}

	// Step 6: Create status dialog (simulating what app.go does)
	statusDialog := NewGitStatusDialog("/tmp/test-repo")
	if statusDialog == nil {
		t.Fatal("Failed to create git status dialog")
	}

	// Step 7: Verify initial loading state
	if !statusDialog.loading {
		t.Error("Expected status dialog to be in loading state initially")
	}

	// Step 8: Simulate status refresh completion with mock data
	mockStatus := git.GitStatus{
		Branch:      "main",
		IsDirty:     false,
		HasUpstream: true,
		Ahead:       0,
		Behind:      0,
		LastUpdated: time.Now(),
		Error:       nil,
	}

	refreshMsg := GitStatusRefreshMsg{
		Status: mockStatus,
		Err:    nil,
	}

	// Step 9: Update status dialog with mock data
	updatedDialog, _ := statusDialog.Update(refreshMsg)
	statusDialog = updatedDialog.(*GitStatusDialog)

	// Step 10: Verify status dialog is no longer loading and has correct data
	if statusDialog.loading {
		t.Error("Expected status dialog to finish loading")
	}

	if statusDialog.status.Branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", statusDialog.status.Branch)
	}

	// Step 11: Set dialog size for proper rendering
	statusDialog.SetRect(70, 18, 10, 5)

	// Verify status dialog renders correctly
	view := statusDialog.View()
	if view == "" {
		t.Error("Expected non-empty view from status dialog")
	}

	// Debug: print view content if tests fail
	t.Logf("Status dialog view:\n%s", view)

	if !contains(view, "main") {
		t.Error("Expected view to contain branch name 'main'")
	}

	if !contains(view, "Clean") {
		t.Error("Expected view to show 'Clean' state")
	}

	// Step 11: Test refresh functionality
	result, refreshCmd := statusDialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for refresh, got %v", result)
	}

	if refreshCmd == nil {
		t.Error("Expected non-nil refresh command")
	}

	if !statusDialog.loading {
		t.Error("Expected status dialog to be loading after refresh")
	}

	// Step 12: Test closing status dialog
	result, _ = statusDialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel for ESC, got %v", result)
	}
}

// TestGitMenuToStatusDialog_NavigateToOtherOptions tests navigation through menu
func TestGitMenuToStatusDialog_NavigateToOtherOptions(t *testing.T) {
	selectedIDs := make([]int, 0)

	// Test navigating through all menu items
	for i := 0; i < 4; i++ {
		menuDialog := NewGitMenuDialog(nil)
		
		// Navigate to item i
		for j := 0; j < i; j++ {
			menuDialog.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
		}
		
		if menuDialog.selectedIndex != i {
			t.Errorf("Expected index %d, got %d", i, menuDialog.selectedIndex)
		}

		// Select current item
		result, cmd := menuDialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
		if result != DialogResultClose {
			t.Errorf("Expected DialogResultClose, got %v", result)
		}

		if cmd == nil {
			t.Fatal("Expected command to be returned from HandleKey")
		}
		
		msg := cmd()
		selectionMsg, ok := msg.(GitMenuSelectionMsg)
		if !ok {
			t.Fatalf("Expected GitMenuSelectionMsg, got %T", msg)
		}
		
		selectedIDs = append(selectedIDs, selectionMsg.SelectedIndex)

		if selectionMsg.SelectedIndex != i {
			t.Errorf("Expected selected ID %d, got %d", i, selectionMsg.SelectedIndex)
		}
	}

	// Verify all items were selected
	expectedIDs := []int{0, 1, 2, 3}
	if len(selectedIDs) != len(expectedIDs) {
		t.Errorf("Expected %d selections, got %d", len(expectedIDs), len(selectedIDs))
	}

	for i, expected := range expectedIDs {
		if selectedIDs[i] != expected {
			t.Errorf("Selection %d: expected ID %d, got %d", i, expected, selectedIDs[i])
		}
	}
}

// TestGitMenuToStatusDialog_ErrorHandling tests error scenarios
func TestGitMenuToStatusDialog_ErrorHandling(t *testing.T) {
	// Create status dialog
	statusDialog := NewGitStatusDialog("/tmp/nonexistent-repo")

	// Initialize and get command
	cmd := statusDialog.Init()
	if cmd == nil {
		t.Fatal("Expected non-nil init command")
	}

	// Execute command (should fail for nonexistent repo)
	msg := cmd()
	refreshMsg, ok := msg.(GitStatusRefreshMsg)
	if !ok {
		t.Fatalf("Expected GitStatusRefreshMsg, got %T", msg)
	}

	// Simulate error condition
	refreshMsg.Err = errors.New("not a git repository")
	updatedDialog, _ := statusDialog.Update(refreshMsg)
	statusDialog = updatedDialog.(*GitStatusDialog)

	// Verify error state
	if statusDialog.loading {
		t.Error("Expected loading to be false after error")
	}

	if statusDialog.err == nil {
		t.Error("Expected error to be set")
	}

	// Verify error is displayed
	view := statusDialog.View()
	if !contains(view, "Error") {
		t.Error("Expected view to show error")
	}
}

// TestGitMenuToStatusDialog_DirtyRepository tests displaying dirty state
func TestGitMenuToStatusDialog_DirtyRepository(t *testing.T) {
	// Create and initialize status dialog
	statusDialog := NewGitStatusDialog("/tmp/test-repo")

	// Set dirty status using mock refresh message
	refreshMsg := GitStatusRefreshMsg{
		Status: git.GitStatus{
			Branch:      "feature-123",
			IsDirty:     true,
			HasUpstream: true,
			Ahead:       2,
			Behind:      1,
			LastUpdated: time.Now(),
			Error:       nil,
		},
		Err: nil,
	}

	updatedDialog, _ := statusDialog.Update(refreshMsg)
	statusDialog = updatedDialog.(*GitStatusDialog)

	// Set dialog size for proper rendering
	statusDialog.SetRect(70, 18, 10, 5)

	// Verify dirty state is displayed
	view := statusDialog.View()
	if !contains(view, "Dirty") {
		t.Error("Expected view to show 'Dirty' state")
	}

	if !contains(view, "feature-123") {
		t.Error("Expected view to show branch name")
	}

	if !contains(view, "ahead") {
		t.Error("Expected view to show 'ahead' status")
	}

	if !contains(view, "behind") {
		t.Error("Expected view to show 'behind' status")
	}
}

// TestGitMenuToStatusDialog_MultipleRefreshes tests repeated refresh operations
func TestGitMenuToStatusDialog_MultipleRefreshes(t *testing.T) {
	statusDialog := NewGitStatusDialog("/tmp/test-repo")

	// Initial state
	cmd := statusDialog.Init()
	msg := cmd()
	refreshMsg := msg.(GitStatusRefreshMsg)
	refreshMsg.Status = git.GitStatus{
		Branch:      "main",
		IsDirty:     false,
		HasUpstream: true,
		Ahead:       0,
		Behind:      0,
		LastUpdated: time.Now(),
	}

	updatedDialog, _ := statusDialog.Update(refreshMsg)
	statusDialog = updatedDialog.(*GitStatusDialog)

	// Perform multiple refreshes
	for i := 0; i < 3; i++ {
		// Trigger refresh
		_, refreshCmd := statusDialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		if refreshCmd == nil {
			t.Errorf("Refresh %d: expected non-nil refresh command", i)
			continue
		}

		if !statusDialog.loading {
			t.Errorf("Refresh %d: expected loading state", i)
		}

		// Execute refresh
		refreshMsg := refreshCmd().(GitStatusRefreshMsg)
		refreshMsg.Status = git.GitStatus{
			Branch:      "main",
			IsDirty:     false,
			HasUpstream: true,
			Ahead:       i, // Different ahead count each time
			Behind:      0,
			LastUpdated: time.Now(),
		}

		updatedDialog, _ := statusDialog.Update(refreshMsg)
		statusDialog = updatedDialog.(*GitStatusDialog)

		if statusDialog.loading {
			t.Errorf("Refresh %d: expected loading to be complete", i)
		}

		if statusDialog.status.Ahead != i {
			t.Errorf("Refresh %d: expected ahead=%d, got %d", i, i, statusDialog.status.Ahead)
		}
	}
}
