package dialog

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// FocusableDialog is an interface for dialogs that have focusable elements
type FocusableDialog interface {
	Dialog

	// FocusNext focuses the next element in the dialog
	FocusNext() tea.Cmd

	// FocusPrev focuses the previous element in the dialog
	FocusPrev() tea.Cmd

	// FocusedIndex returns the index of the currently focused element
	FocusedIndex() int

	// SetFocusedIndex sets the focused element by index
	SetFocusedIndex(index int) tea.Cmd

	// NumFocusableElements returns the number of focusable elements
	NumFocusableElements() int

	// IsFocusable returns whether the dialog has focusable elements
	IsFocusable() bool
}

// BaseFocusableDialog implements common functionality for focusable dialogs
type BaseFocusableDialog struct {
	BaseDialog
	focusedIndex int
	numElements  int
}

// NewBaseFocusableDialog creates a new base focusable dialog
func NewBaseFocusableDialog(title string, width, height int, kind DialogKind, numElements int) BaseFocusableDialog {
	return BaseFocusableDialog{
		BaseDialog:   NewBaseDialog(title, width, height, kind),
		focusedIndex: 0,
		numElements:  numElements,
	}
}

// FocusedIndex returns the index of the currently focused element
func (d BaseFocusableDialog) FocusedIndex() int {
	return d.focusedIndex
}

// SetFocusedIndex sets the focused element by index
func (d *BaseFocusableDialog) SetFocusedIndex(index int) tea.Cmd {
	if index >= 0 && index < d.numElements {
		d.focusedIndex = index
	}
	return nil
}

// FocusNext focuses the next element in the dialog
func (d *BaseFocusableDialog) FocusNext() tea.Cmd {
	if d.numElements > 0 {
		d.focusedIndex = (d.focusedIndex + 1) % d.numElements
	}
	return nil
}

// FocusPrev focuses the previous element in the dialog
func (d *BaseFocusableDialog) FocusPrev() tea.Cmd {
	if d.numElements > 0 {
		d.focusedIndex = (d.focusedIndex - 1 + d.numElements) % d.numElements
	}
	return nil
}

// NumFocusableElements returns the number of focusable elements
func (d BaseFocusableDialog) NumFocusableElements() int {
	return d.numElements
}

// IsFocusable returns whether the dialog has focusable elements
func (d BaseFocusableDialog) IsFocusable() bool {
	return d.numElements > 0
}

// HandleBaseFocusableKey handles common key events for focusable dialogs
func (d *BaseFocusableDialog) HandleBaseFocusableKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (like ESC)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Then handle focusable dialog keys
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		return DialogResultNone, d.FocusNext()
	case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
		return DialogResultNone, d.FocusPrev()
	}

	return DialogResultNone, nil
}
