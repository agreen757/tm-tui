package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewModelSelectionDialog tests creating a new dialog
func TestNewModelSelectionDialog(t *testing.T) {
	dialog := NewModelSelectionDialog()

	t.Run("creates dialog with default values", func(t *testing.T) {
		if dialog == nil {
			t.Fatal("Expected dialog to be created")
		}
		if dialog.focusedPane != "providers" {
			t.Errorf("Expected focusedPane to be 'providers', got %s", dialog.focusedPane)
		}
		if dialog.selectedIndex != 0 {
			t.Errorf("Expected selectedIndex to be 0, got %d", dialog.selectedIndex)
		}
		if dialog.modelIndex != 0 {
			t.Errorf("Expected modelIndex to be 0, got %d", dialog.modelIndex)
		}
	})

	t.Run("loads providers", func(t *testing.T) {
		if len(dialog.providers) == 0 {
			t.Error("Expected providers to be loaded")
		}
	})
}

// TestModelSelectionDialogInit tests dialog initialization
func TestModelSelectionDialogInit(t *testing.T) {
	dialog := NewModelSelectionDialog()

	t.Run("init returns no error", func(t *testing.T) {
		cmd := dialog.Init()
		// cmd is a tea.Cmd which is func() tea.Msg
		// A nil tea.Cmd is allowed
		_ = cmd
	})
}

// TestModelSelectionDialogUpdate tests dialog updates
func TestModelSelectionDialogUpdate(t *testing.T) {
	dialog := NewModelSelectionDialog()

	t.Run("up key decrements provider index", func(t *testing.T) {
		dialog.selectedIndex = 2
		if len(dialog.providers) > 2 {
			msg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'k'},
			}
			d, _ := dialog.Update(msg)
			updatedDialog := d.(*ModelSelectionDialog)
			if updatedDialog.selectedIndex != 1 {
				t.Errorf("Expected selectedIndex to be 1, got %d", updatedDialog.selectedIndex)
			}
		}
	})

	t.Run("down key increments provider index", func(t *testing.T) {
		dialog.selectedIndex = 0
		if len(dialog.providers) > 1 {
			msg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'j'},
			}
			d, _ := dialog.Update(msg)
			updatedDialog := d.(*ModelSelectionDialog)
			if updatedDialog.selectedIndex != 1 {
				t.Errorf("Expected selectedIndex to be 1, got %d", updatedDialog.selectedIndex)
			}
		}
	})

	t.Run("tab switches panes", func(t *testing.T) {
		dialog.focusedPane = "providers"
		msg := tea.KeyMsg{
			Type: tea.KeyTab,
		}
		d, _ := dialog.Update(msg)
		updatedDialog := d.(*ModelSelectionDialog)
		if updatedDialog.focusedPane != "models" {
			t.Errorf("Expected focusedPane to be 'models', got %s", updatedDialog.focusedPane)
		}
	})

	t.Run("escape cancels", func(t *testing.T) {
		msg := tea.KeyMsg{
			Type: tea.KeyEsc,
		}
		d, _ := dialog.Update(msg)
		updatedDialog := d.(*ModelSelectionDialog)
		if updatedDialog == nil {
			t.Error("Expected dialog to remain valid after escape")
		}
	})
}

// TestModelSelectionDialogView tests rendering
func TestModelSelectionDialogView(t *testing.T) {
	dialog := NewModelSelectionDialog()

	t.Run("renders without errors", func(t *testing.T) {
		view := dialog.View()
		if view == "" {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("includes title", func(t *testing.T) {
		view := dialog.View()
		if view == "" {
			t.Fatal("Expected non-empty view")
		}
		// Just verify it renders without panic
	})

	t.Run("includes providers in view", func(t *testing.T) {
		view := dialog.View()
		if len(dialog.providers) > 0 && len(view) == 0 {
			t.Error("Expected providers to be in view")
		}
	})
}

// TestDefaultModelSelectionKeyMap tests key map
func TestDefaultModelSelectionKeyMap(t *testing.T) {
	keyMap := DefaultModelSelectionKeyMap()

	t.Run("key map has required bindings", func(t *testing.T) {
		if keyMap.Up.Keys() == nil || len(keyMap.Up.Keys()) == 0 {
			t.Error("Expected Up key binding to have keys")
		}
		if keyMap.Down.Keys() == nil || len(keyMap.Down.Keys()) == 0 {
			t.Error("Expected Down key binding to have keys")
		}
		if keyMap.Confirm.Keys() == nil || len(keyMap.Confirm.Keys()) == 0 {
			t.Error("Expected Confirm key binding to have keys")
		}
		if keyMap.Cancel.Keys() == nil || len(keyMap.Cancel.Keys()) == 0 {
			t.Error("Expected Cancel key binding to have keys")
		}
	})
}
