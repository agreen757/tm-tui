package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGitMenuDialog_Creation(t *testing.T) {
	onSelect := func(id int) {
		// Callback for testing
	}

	dialog := NewGitMenuDialog(onSelect)

	if dialog == nil {
		t.Fatal("NewGitMenuDialog returned nil")
	}

	if dialog.Title() != "Git Menu" {
		t.Errorf("Expected title 'Git Menu', got '%s'", dialog.Title())
	}

	if dialog.Kind() != DialogKindList {
		t.Errorf("Expected kind DialogKindList, got %v", dialog.Kind())
	}

	if len(dialog.items) != 4 {
		t.Errorf("Expected 4 menu items, got %d", len(dialog.items))
	}

	// Verify menu items
	expectedItems := []struct {
		title       string
		description string
	}{
		{"Show Status", "View detailed repository status"},
		{"Switch Branch", "Checkout an existing branch"},
		{"Create Branch", "Create and checkout a new branch"},
		{"Recent Commits", "View recent commit history"},
	}

	for i, expected := range expectedItems {
		if dialog.items[i].title != expected.title {
			t.Errorf("Item %d: expected title '%s', got '%s'", i, expected.title, dialog.items[i].title)
		}
		if dialog.items[i].description != expected.description {
			t.Errorf("Item %d: expected description '%s', got '%s'", i, expected.description, dialog.items[i].description)
		}
	}
}

func TestGitMenuDialog_Navigation(t *testing.T) {
	dialog := NewGitMenuDialog(nil)

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
	if dialog.selectedIndex != 3 {
		t.Errorf("Expected selectedIndex 3 after wrap-around up, got %d", dialog.selectedIndex)
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

func TestGitMenuDialog_Selection(t *testing.T) {
	dialog := NewGitMenuDialog(nil)
	dialog.selectedIndex = 2

	// Test enter key
	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultClose {
		t.Errorf("Expected DialogResultClose after enter, got %v", result)
	}

	if cmd == nil {
		t.Fatal("Expected command to be returned from HandleKey")
	}

	// Execute the command to get the selection message
	msg := cmd()
	selectionMsg, ok := msg.(GitMenuSelectionMsg)
	if !ok {
		t.Fatalf("Expected GitMenuSelectionMsg, got %T", msg)
	}

	if selectionMsg.SelectedIndex != 2 {
		t.Errorf("Expected selected ID 2, got %d", selectionMsg.SelectedIndex)
	}
}

func TestGitMenuDialog_Cancel(t *testing.T) {
	dialog := NewGitMenuDialog(nil)

	// Test ESC key
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel after ESC, got %v", result)
	}
}

func TestGitMenuDialog_View(t *testing.T) {
	dialog := NewGitMenuDialog(nil)
	dialog.SetRect(60, 12, 10, 5)

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Check that all menu items appear in the view
	if !contains(view, "Show Status") {
		t.Error("Expected view to contain 'Show Status'")
	}
	if !contains(view, "Switch Branch") {
		t.Error("Expected view to contain 'Switch Branch'")
	}
	if !contains(view, "Create Branch") {
		t.Error("Expected view to contain 'Create Branch'")
	}
	if !contains(view, "Recent Commits") {
		t.Error("Expected view to contain 'Recent Commits'")
	}
}

func TestGitMenuDialog_Update(t *testing.T) {
	dialog := NewGitMenuDialog(nil)
	
	// Test window resize message
	updatedDialog, _ := dialog.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if updatedDialog == nil {
		t.Error("Expected non-nil dialog after update")
	}
}

func TestGitMenuDialog_Init(t *testing.T) {
	dialog := NewGitMenuDialog(nil)
	cmd := dialog.Init()
	if cmd != nil {
		t.Error("Expected nil cmd from Init")
	}
}

func TestGitMenuDialog_BaseDialogIntegration(t *testing.T) {
	dialog := NewGitMenuDialog(nil)

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
	dialog.SetRect(60, 12, 10, 5)
	w, h, x, y := dialog.GetRect()
	if w != 60 || h != 12 || x != 10 || y != 5 {
		t.Errorf("Expected rect (60, 12, 10, 5), got (%d, %d, %d, %d)", w, h, x, y)
	}
}
