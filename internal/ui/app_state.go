package ui

import (
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

// SavedDialogState captures the app state before opening a dialog
// to enable restoration when the dialog closes
type SavedDialogState struct {
	CurrentView     ViewMode
	ScrollPosition  int
	SelectedTaskID  string
	FocusedPanel    Panel
	ShowDetailsPane bool
	ShowLogPane     bool
}

// AppState centralizes references shared across the UI (dialogs, keymap, etc.).
type AppState struct {
	dialogManager        *dialog.DialogManager
	keyMap               *KeyMap
	nextTaskModalActive  bool
	nextTaskOutput       []string
	PrdCreationState     *PrdCreationState
	SavedDialogStates    []SavedDialogState // Stack of saved states for dialog nesting
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

// StartNextTaskModal initializes the next task modal state.
func (s *AppState) StartNextTaskModal() {
	s.nextTaskModalActive = true
	s.nextTaskOutput = []string{}
}

// AppendNextTaskOutput appends a line to the output only when modal is active.
func (s *AppState) AppendNextTaskOutput(line string) {
	if s.nextTaskModalActive {
		s.nextTaskOutput = append(s.nextTaskOutput, line)
	}
}

// CloseNextTaskModal resets the next task modal state.
func (s *AppState) CloseNextTaskModal() {
	s.nextTaskModalActive = false
	s.nextTaskOutput = nil
}

// NextTaskOutput returns the current next task output lines.
func (s *AppState) NextTaskOutput() []string {
	return s.nextTaskOutput
}

// IsNextTaskModalActive returns whether the next task modal is active.
func (s *AppState) IsNextTaskModalActive() bool {
	return s.nextTaskModalActive
}

// InitPrdCreationState initializes the PRD creation state if not already initialized.
func (s *AppState) InitPrdCreationState() {
	if s.PrdCreationState == nil {
		s.PrdCreationState = NewPrdCreationState()
	}
}

// UpdatePrdCreationInputs updates the PRD creation state with form input values.
func (s *AppState) UpdatePrdCreationInputs(title, summary, scope, filename string) {
	s.InitPrdCreationState()
	s.PrdCreationState.Title = title
	s.PrdCreationState.Summary = summary
	s.PrdCreationState.Scope = scope
	s.PrdCreationState.Filename = filename
}

// GetPrdCreationState returns the current PRD creation state, initializing if needed.
func (s *AppState) GetPrdCreationState() *PrdCreationState {
	s.InitPrdCreationState()
	return s.PrdCreationState
}

// ClearPrdCreationState clears the PRD creation state.
func (s *AppState) ClearPrdCreationState() {
	s.PrdCreationState = nil
}

// SaveDialogState saves the current app state for restoration after dialog closes
func (s *AppState) SaveDialogState(state SavedDialogState) {
	s.SavedDialogStates = append(s.SavedDialogStates, state)
}

// RestoreDialogState pops and returns the most recently saved app state
func (s *AppState) RestoreDialogState() *SavedDialogState {
	if len(s.SavedDialogStates) == 0 {
		return nil
	}
	// Pop from the end of the slice
	state := s.SavedDialogStates[len(s.SavedDialogStates)-1]
	s.SavedDialogStates = s.SavedDialogStates[:len(s.SavedDialogStates)-1]
	return &state
}

// HasSavedState checks if there are any saved dialog states
func (s *AppState) HasSavedState() bool {
	return len(s.SavedDialogStates) > 0
}
