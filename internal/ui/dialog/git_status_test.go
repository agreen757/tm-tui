package dialog

import (
	"context"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/git"
	tea "github.com/charmbracelet/bubbletea"
)

func TestGitStatusDialog_Creation(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")

	if dialog == nil {
		t.Fatal("NewGitStatusDialog returned nil")
	}

	if dialog.Title() != "Git Status" {
		t.Errorf("Expected title 'Git Status', got '%s'", dialog.Title())
	}

	if dialog.Kind() != DialogKindModal {
		t.Errorf("Expected kind DialogKindModal, got %v", dialog.Kind())
	}

	if dialog.repoPath != "/tmp/test-repo" {
		t.Errorf("Expected repoPath '/tmp/test-repo', got '%s'", dialog.repoPath)
	}

	if !dialog.loading {
		t.Error("Expected initial loading state to be true")
	}
}

func TestGitStatusDialog_Init(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	cmd := dialog.Init()

	if cmd == nil {
		t.Error("Expected non-nil cmd from Init")
	}

	// Execute the command to get the message
	msg := cmd()
	if msg == nil {
		t.Error("Expected non-nil message from Init command")
	}

	// Should be GitStatusRefreshMsg
	if _, ok := msg.(GitStatusRefreshMsg); !ok {
		t.Errorf("Expected GitStatusRefreshMsg, got %T", msg)
	}
}

func TestGitStatusDialog_Update_RefreshMsg(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.loading = true

	// Create a mock status
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

	updatedDialog, _ := dialog.Update(refreshMsg)
	gitDialog := updatedDialog.(*GitStatusDialog)

	if gitDialog.loading {
		t.Error("Expected loading to be false after refresh")
	}

	if gitDialog.status.Branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", gitDialog.status.Branch)
	}

	if gitDialog.err != nil {
		t.Errorf("Expected no error, got %v", gitDialog.err)
	}
}

func TestGitStatusDialog_Update_WindowSizeMsg(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")

	updatedDialog, _ := dialog.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if updatedDialog == nil {
		t.Error("Expected non-nil dialog after update")
	}
}

func TestGitStatusDialog_View_Loading(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.SetRect(70, 18, 10, 5)
	dialog.loading = true

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "Loading") {
		t.Error("Expected view to contain 'Loading'")
	}
}

func TestGitStatusDialog_View_Error(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.SetRect(70, 18, 10, 5)
	dialog.loading = false
	dialog.err = context.DeadlineExceeded

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "Error") {
		t.Error("Expected view to contain 'Error'")
	}
}

func TestGitStatusDialog_View_CleanStatus(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.SetRect(70, 18, 10, 5)
	dialog.loading = false
	dialog.status = git.GitStatus{
		Branch:      "main",
		IsDirty:     false,
		HasUpstream: true,
		Ahead:       0,
		Behind:      0,
		LastUpdated: time.Now(),
		Error:       nil,
	}

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "main") {
		t.Error("Expected view to contain branch 'main'")
	}

	if !contains(view, "Clean") {
		t.Error("Expected view to contain 'Clean' state")
	}

	if !contains(view, "Tracked") {
		t.Error("Expected view to contain 'Tracked' upstream")
	}
}

func TestGitStatusDialog_View_DirtyStatus(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.SetRect(70, 18, 10, 5)
	dialog.loading = false
	dialog.status = git.GitStatus{
		Branch:      "feature-branch",
		IsDirty:     true,
		HasUpstream: false,
		Ahead:       0,
		Behind:      0,
		LastUpdated: time.Now(),
		Error:       nil,
	}

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "feature-branch") {
		t.Error("Expected view to contain branch 'feature-branch'")
	}

	if !contains(view, "Dirty") {
		t.Error("Expected view to contain 'Dirty' state")
	}

	if !contains(view, "Not tracked") {
		t.Error("Expected view to contain 'Not tracked' upstream")
	}
}

func TestGitStatusDialog_View_AheadBehind(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.SetRect(70, 18, 10, 5)
	dialog.loading = false
	dialog.status = git.GitStatus{
		Branch:      "develop",
		IsDirty:     false,
		HasUpstream: true,
		Ahead:       3,
		Behind:      2,
		LastUpdated: time.Now(),
		Error:       nil,
	}

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "ahead") {
		t.Error("Expected view to contain 'ahead'")
	}

	if !contains(view, "behind") {
		t.Error("Expected view to contain 'behind'")
	}
}

func TestGitStatusDialog_HandleKey_Refresh(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")
	dialog.loading = false

	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for refresh, got %v", result)
	}

	if cmd == nil {
		t.Error("Expected non-nil cmd for refresh")
	}

	if !dialog.loading {
		t.Error("Expected loading to be true after refresh key")
	}

	// Execute the command
	msg := cmd()
	if _, ok := msg.(GitStatusRefreshMsg); !ok {
		t.Errorf("Expected GitStatusRefreshMsg from refresh command, got %T", msg)
	}
}

func TestGitStatusDialog_HandleKey_Cancel(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")

	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})

	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel after ESC, got %v", result)
	}
}

func TestGitStatusDialog_BaseDialogIntegration(t *testing.T) {
	dialog := NewGitStatusDialog("/tmp/test-repo")

	// Test SetZIndex/ZIndex
	dialog.SetZIndex(3)
	if dialog.ZIndex() != 3 {
		t.Errorf("Expected ZIndex 3, got %d", dialog.ZIndex())
	}

	// Test SetFocused/IsFocused
	dialog.SetFocused(false)
	if dialog.IsFocused() {
		t.Error("Expected IsFocused to be false")
	}
	dialog.SetFocused(true)
	if !dialog.IsFocused() {
		t.Error("Expected IsFocused to be true")
	}

	// Test IsCancellable
	if !dialog.IsCancellable() {
		t.Error("Expected dialog to be cancellable")
	}

	// Test GetRect
	dialog.SetRect(70, 18, 10, 5)
	w, h, x, y := dialog.GetRect()
	if w != 70 || h != 18 || x != 10 || y != 5 {
		t.Errorf("Expected rect (70, 18, 10, 5), got (%d, %d, %d, %d)", w, h, x, y)
	}
}
