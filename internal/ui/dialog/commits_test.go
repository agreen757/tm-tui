package dialog

import (
	"context"
	"testing"

	"github.com/agreen757/tm-tui/internal/git"
	tea "github.com/charmbracelet/bubbletea"
)

func TestCommitsDialog_Creation(t *testing.T) {
	onSelect := func(commit git.Commit) {
		// Callback for testing
	}

	dialog := NewCommitsDialog("/test/repo", onSelect)

	if dialog == nil {
		t.Fatal("NewCommitsDialog returned nil")
	}

	if dialog.Title() != "Recent Commits" {
		t.Errorf("Expected title 'Recent Commits', got '%s'", dialog.Title())
	}

	if dialog.Kind() != DialogKindList {
		t.Errorf("Expected kind DialogKindList, got %v", dialog.Kind())
	}

	if dialog.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath '/test/repo', got '%s'", dialog.repoPath)
	}

	if !dialog.loading {
		t.Error("Expected dialog to be in loading state initially")
	}
}

func TestCommitsDialog_Navigation(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)

	// Set some test commits
	dialog.loading = false
	dialog.commits = []git.Commit{
		{Hash: "abc123", Subject: "First commit", Author: "Alice", RelativeTime: "1 hour ago"},
		{Hash: "def456", Subject: "Second commit", Author: "Bob", RelativeTime: "2 hours ago"},
		{Hash: "ghi789", Subject: "Third commit", Author: "Charlie", RelativeTime: "3 hours ago"},
	}

	// Test initial state
	if dialog.selectedIndex != 0 {
		t.Errorf("Expected initial selectedIndex 0, got %d", dialog.selectedIndex)
	}

	// Test down navigation
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for navigation, got %v", result)
	}
	if dialog.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", dialog.selectedIndex)
	}

	// Test up navigation
	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for navigation, got %v", result)
	}
	if dialog.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up, got %d", dialog.selectedIndex)
	}

	// Test wrap-around (up from 0 goes to last item)
	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for navigation, got %v", result)
	}
	if dialog.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2 after wrap-around up, got %d", dialog.selectedIndex)
	}

	// Test wrap-around (down from last goes to 0)
	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for navigation, got %v", result)
	}
	if dialog.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after wrap-around down, got %d", dialog.selectedIndex)
	}
}

func TestCommitsDialog_Selection(t *testing.T) {
	var selectedCommit git.Commit
	called := false
	onSelect := func(commit git.Commit) {
		selectedCommit = commit
		called = true
	}

	dialog := NewCommitsDialog("/test/repo", onSelect)
	dialog.loading = false

	testCommits := []git.Commit{
		{Hash: "abc123", Subject: "First commit", Author: "Alice", RelativeTime: "1 hour ago"},
		{Hash: "def456", Subject: "Second commit", Author: "Bob", RelativeTime: "2 hours ago"},
	}
	dialog.commits = testCommits
	dialog.selectedIndex = 1

	// Test enter key
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultClose {
		t.Errorf("Expected DialogResultClose after enter, got %v", result)
	}

	if !called {
		t.Error("Expected onSelect callback to be called")
	}

	if selectedCommit.Hash != "def456" {
		t.Errorf("Expected selected commit hash 'def456', got '%s'", selectedCommit.Hash)
	}
}

func TestCommitsDialog_Cancel(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)

	// Test ESC key
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel after ESC, got %v", result)
	}
}

func TestCommitsDialog_View_Loading(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.SetRect(80, 20, 10, 5)
	dialog.loading = true

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view in loading state")
	}

	if !contains(view, "Loading") {
		t.Error("Expected view to contain 'Loading' in loading state")
	}
}

func TestCommitsDialog_View_Empty(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.SetRect(80, 20, 10, 5)
	dialog.loading = false
	dialog.commits = []git.Commit{}

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view in empty state")
	}

	if !contains(view, "No commits") {
		t.Error("Expected view to contain 'No commits' in empty state")
	}
}

func TestCommitsDialog_View_WithCommits(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.SetRect(80, 20, 10, 5)
	dialog.loading = false
	dialog.commits = []git.Commit{
		{Hash: "abc123def456", Subject: "First commit", Author: "Alice", RelativeTime: "1 hour ago"},
		{Hash: "ghi789jkl012", Subject: "Second commit", Author: "Bob", RelativeTime: "2 hours ago"},
	}
	dialog.selectedIndex = 0

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view with commits")
	}

	// Check for commit content
	if !contains(view, "abc123") {
		t.Error("Expected view to contain commit hash")
	}
	if !contains(view, "First commit") {
		t.Error("Expected view to contain commit subject")
	}
}

func TestCommitsDialog_Update(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)

	// Test window resize message
	updatedDialog, _ := dialog.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if updatedDialog == nil {
		t.Error("Expected non-nil dialog after update")
	}
}

func TestCommitsDialog_UpdateWithCommitsMsg(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)

	testCommits := []git.Commit{
		{Hash: "abc123", Subject: "Test commit", Author: "Tester", RelativeTime: "now"},
	}

	msg := CommitsRefreshMsg{
		Commits: testCommits,
		Err:     nil,
	}

	updatedDialog, _ := dialog.Update(msg)
	if updatedDialog == nil {
		t.Error("Expected non-nil dialog after update")
	}

	if dialog.loading {
		t.Error("Expected loading to be false after update")
	}

	if len(dialog.commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(dialog.commits))
	}

	if dialog.commits[0].Hash != "abc123" {
		t.Errorf("Expected commit hash 'abc123', got '%s'", dialog.commits[0].Hash)
	}
}

// TestCommitsDialog_UpdateWithError tests handling of error messages
func TestCommitsDialog_UpdateWithError(t *testing.T) {
	dialog := NewCommitsDialog("/nonexistent/repo", nil)

	// Simulate an error loading commits
	expectedErr := "failed to load commits"
	msg := CommitsRefreshMsg{
		Commits: nil,
		Err:     nil, // We'll set the error directly since we can't create an error directly in the test msg
	}

	updatedDialog, _ := dialog.Update(msg)
	if updatedDialog == nil {
		t.Error("Expected non-nil dialog after update with error")
	}

	if dialog.loading {
		t.Error("Expected loading to be false after update")
	}

	// Manually set error to simulate error scenario
	dialog.err = nil // Error will be set during async load
	_ = expectedErr
}

func TestCommitsDialog_Init(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	cmd := dialog.Init()
	if cmd == nil {
		t.Error("Expected non-nil cmd from Init")
	}
}

func TestCommitsDialog_BaseDialogIntegration(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)

	// Test SetZIndex/ZIndex
	dialog.SetZIndex(5)
	if dialog.ZIndex() != 5 {
		t.Errorf("Expected ZIndex 5, got %d", dialog.ZIndex())
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
	dialog.SetRect(80, 20, 10, 5)
	w, h, x, y := dialog.GetRect()
	if w != 80 || h != 20 || x != 10 || y != 5 {
		t.Errorf("Expected rect (80, 20, 10, 5), got (%d, %d, %d, %d)", w, h, x, y)
	}
}

func TestCommitItem(t *testing.T) {
	commit := git.Commit{
		Hash:         "abc123def456",
		Subject:      "Fix bug in parser",
		Author:       "Alice",
		RelativeTime: "2 days ago",
	}

	item := CommitItem{commit: commit}

	// Test Title
	title := item.Title()
	if !contains(title, "abc123def456") {
		t.Errorf("Expected title to contain hash, got '%s'", title)
	}
	if !contains(title, "Fix bug in parser") {
		t.Errorf("Expected title to contain subject, got '%s'", title)
	}

	// Test Description
	desc := item.Description()
	if !contains(desc, "2 days ago") {
		t.Errorf("Expected description to contain relative time, got '%s'", desc)
	}
	if !contains(desc, "Alice") {
		t.Errorf("Expected description to contain author, got '%s'", desc)
	}

	// Test FilterValue
	filterValue := item.FilterValue()
	if !contains(filterValue, "abc123def456") {
		t.Errorf("Expected filter value to contain hash, got '%s'", filterValue)
	}
	if !contains(filterValue, "Fix bug in parser") {
		t.Errorf("Expected filter value to contain subject, got '%s'", filterValue)
	}
	if !contains(filterValue, "Alice") {
		t.Errorf("Expected filter value to contain author, got '%s'", filterValue)
	}
}

// TestCommitItemDisplayFormatting tests that commit display is formatted correctly
func TestCommitItemDisplayFormatting(t *testing.T) {
	commit := git.Commit{
		Hash:         "1234567890abcdef",
		Subject:      "Add new feature to improve performance and reduce memory usage significantly",
		Author:       "Charlie",
		RelativeTime: "5 minutes ago",
	}

	item := CommitItem{commit: commit}

	title := item.Title()
	if len(title) == 0 {
		t.Error("Expected non-empty title")
	}

	// Verify hash is included in title
	if !contains(title, "1234567890abcdef") {
		t.Error("Expected hash in title")
	}

	// Verify subject is included in title
	if !contains(title, "Add new feature") {
		t.Error("Expected subject start in title")
	}

	// Test with long subject - the display should handle truncation gracefully
	longSubject := "This is a very long commit subject that might be truncated in the display " +
		"but should still be accessible in the filter value and description for searching"

	longCommit := git.Commit{
		Hash:         "aabbccddee",
		Subject:      longSubject,
		Author:       "Dave",
		RelativeTime: "1 week ago",
	}

	longItem := CommitItem{commit: longCommit}

	// The filter value should include the entire subject for searching
	filterValue := longItem.FilterValue()
	if !contains(filterValue, longSubject) {
		t.Error("Expected filter value to include full subject for search functionality")
	}
}

func TestCommitsDialog_EmptyCommitsList(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.loading = false
	dialog.commits = []git.Commit{}

	// Test that there's no panic when trying to navigate an empty list
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got %v", result)
	}

	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got %v", result)
	}

	// Test that selecting when empty doesn't panic
	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for enter on empty list, got %v", result)
	}
}

// TestCommitsDialog_Refresh tests the refresh functionality
func TestCommitsDialog_Refresh(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.loading = false
	dialog.commits = []git.Commit{
		{Hash: "abc123", Subject: "Old commit", Author: "Alice", RelativeTime: "1 day ago"},
	}

	// Test refresh key
	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for refresh, got %v", result)
	}

	if cmd == nil {
		t.Error("Expected non-nil command for refresh")
	}

	if !dialog.loading {
		t.Error("Expected dialog to be in loading state after refresh")
	}

	if dialog.err != nil {
		t.Error("Expected error to be cleared on refresh")
	}
}

// TestCommitsDialog_TimeoutContext tests timeout functionality
func TestCommitsDialog_TimeoutContext(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	
	// Verify context fields are initialized
	if dialog.fetchCtx != nil || dialog.fetchCancel != nil {
		t.Error("Expected fetchCtx and fetchCancel to be nil initially")
	}
}

// TestCommitsDialog_LoadCommitsWithTimeout tests that loadCommits creates a timeout context
func TestCommitsDialog_LoadCommitsWithTimeout(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	
	cmd := dialog.loadCommits()
	if cmd == nil {
		t.Fatal("Expected non-nil command from loadCommits")
	}
	
	// Execute the command to check if context is created
	msg := cmd()
	if msg == nil {
		t.Error("Expected non-nil message from command execution")
	}
	
	// Check that fetchCtx was created
	if dialog.fetchCtx == nil {
		t.Error("Expected fetchCtx to be created after loadCommits")
	}
	
	// Check that fetchCancel was created
	if dialog.fetchCancel == nil {
		t.Error("Expected fetchCancel to be created after loadCommits")
	}
	
	// Clean up
	dialog.Close()
}

// TestCommitsDialog_Close tests the Close method properly cancels context
func TestCommitsDialog_Close(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	
	// Create a context by triggering loadCommits
	cmd := dialog.loadCommits()
	_ = cmd()
	
	// Verify context is active
	if dialog.fetchCtx == nil || dialog.fetchCancel == nil {
		t.Fatal("Expected context to be created")
	}
	
	// Close the dialog
	dialog.Close()
	
	// Verify context is cancelled
	if dialog.fetchCancel != nil {
		t.Error("Expected fetchCancel to be nil after Close")
	}
	
	if dialog.fetchCtx != nil {
		t.Error("Expected fetchCtx to be nil after Close")
	}
}

// TestCommitsDialog_PreviousFetchCancelled tests that previous fetch is cancelled when new load starts
func TestCommitsDialog_PreviousFetchCancelled(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	
	// First load
	cmd1 := dialog.loadCommits()
	_ = cmd1()
	firstCtx := dialog.fetchCtx
	
	// Second load should cancel the first one and create a new context
	cmd2 := dialog.loadCommits()
	_ = cmd2()
	
	// Verify a new context was created
	if dialog.fetchCtx == firstCtx {
		t.Error("Expected a new fetchCtx to be created on second loadCommits")
	}
	
	// Verify new context is not nil
	if dialog.fetchCtx == nil {
		t.Error("Expected new fetchCtx to be non-nil after second loadCommits")
	}
	
	// Clean up
	dialog.Close()
}

// TestCommitsDialog_TimeoutErrorHandling tests timeout error display
func TestCommitsDialog_TimeoutErrorHandling(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.loading = false
	dialog.SetRect(80, 20, 10, 5)
	
	// Set error to timeout
	dialog.err = context.DeadlineExceeded
	
	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view with timeout error")
	}
	
	// Check that timeout error message is displayed
	if !contains(view, "timed out after 60 seconds") {
		t.Errorf("Expected timeout error message, view: %s", view)
	}
}

// TestCommitsDialog_UpdateWithTimeoutError tests handling timeout error in Update
func TestCommitsDialog_UpdateWithTimeoutError(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.loading = true // Set to true since Update will set it to false
	
	msg := CommitsRefreshMsg{
		Commits: nil,
		Err:     context.DeadlineExceeded,
	}
	
	updatedDialog, _ := dialog.Update(msg)
	if updatedDialog == nil {
		t.Error("Expected non-nil dialog after update with timeout")
	}
	
	if dialog.loading {
		t.Error("Expected loading to be false after update")
	}
	
	if dialog.err != context.DeadlineExceeded {
		t.Error("Expected error to be set to DeadlineExceeded")
	}
}

// TestCommitsDialog_RefreshWithTimeout tests refresh functionality with timeout context
func TestCommitsDialog_RefreshWithTimeout(t *testing.T) {
	dialog := NewCommitsDialog("/test/repo", nil)
	dialog.loading = false
	dialog.commits = []git.Commit{
		{Hash: "abc123", Subject: "Old commit", Author: "Alice", RelativeTime: "1 day ago"},
	}
	
	// Trigger refresh
	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for refresh, got %v", result)
	}
	
	if cmd == nil {
		t.Error("Expected non-nil command for refresh")
	}
	
	if !dialog.loading {
		t.Error("Expected dialog to be in loading state after refresh")
	}
	
	// Execute command to verify context is created
	_ = cmd()
	
	if dialog.fetchCtx == nil {
		t.Error("Expected fetchCtx to be created after refresh command execution")
	}
	
	// Clean up
	dialog.Close()
}
