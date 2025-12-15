package ui

import (
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

// AppState centralizes references shared across the UI (dialogs, keymap, etc.).
type AppState struct {
	dialogManager *dialog.DialogManager
	keyMap        *KeyMap
}

// NewAppState constructs an AppState helper.
func NewAppState(manager *dialog.DialogManager, keyMap *KeyMap) *AppState {
	return &AppState{
		dialogManager: manager,
		keyMap:        keyMap,
	}
}

// DialogManager returns the dialog manager reference.
func (s *AppState) DialogManager() *dialog.DialogManager {
	return s.dialogManager
}

// KeyMap returns the active key map reference.
func (s *AppState) KeyMap() *KeyMap {
	return s.keyMap
}

// DialogStyle returns the dialog style when available.
func (s *AppState) DialogStyle() *dialog.DialogStyle {
	if s.dialogManager == nil {
		return nil
	}
	return s.dialogManager.Style
}

// HandleDialogMsg routes a Bubble Tea message through the dialog manager.
func (s *AppState) HandleDialogMsg(msg tea.Msg) tea.Cmd {
	if s.dialogManager == nil {
		return nil
	}
	return s.dialogManager.HandleMsg(msg)
}

// HasActiveDialog reports whether any dialogs are currently visible.
func (s *AppState) HasActiveDialog() bool {
	return s.dialogManager != nil && s.dialogManager.HasDialogs()
}

// ActiveDialog returns the dialog at the top of the stack.
func (s *AppState) ActiveDialog() dialog.Dialog {
	if s.dialogManager == nil {
		return nil
	}
	return s.dialogManager.GetActiveDialog()
}

// PushDialog pushes a dialog without a callback.
func (s *AppState) PushDialog(d dialog.Dialog) {
	if s.dialogManager == nil || d == nil {
		return
	}
	s.dialogManager.PushDialog(d)
}

// AddDialog pushes a dialog with the supplied callback.
func (s *AppState) AddDialog(d dialog.Dialog, cb dialog.DialogCallback) {
	if s.dialogManager == nil || d == nil {
		return
	}
	s.dialogManager.AddDialog(d, cb)
}

// ReplaceDialog swaps the active dialog with the provided instance.
func (s *AppState) ReplaceDialog(d dialog.Dialog, cb dialog.DialogCallback) {
	if s.dialogManager == nil || d == nil {
		return
	}
	if s.dialogManager.HasDialogs() {
		s.dialogManager.PopDialog()
	}
	s.dialogManager.AddDialog(d, cb)
}

// PopDialog removes the active dialog and returns it.
func (s *AppState) PopDialog() dialog.Dialog {
	if s.dialogManager == nil {
		return nil
	}
	return s.dialogManager.PopDialog()
}

// ClearDialogs removes every dialog from the stack.
func (s *AppState) ClearDialogs() {
	if s.dialogManager == nil {
		return
	}
	for s.dialogManager.HasDialogs() {
		s.dialogManager.PopDialog()
	}
}
