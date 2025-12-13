package dialog

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyboardNavigationHelper provides consistent keyboard navigation across dialogs
type KeyboardNavigationHelper struct {
	focusedIndex int
	numElements  int
}

// NewKeyboardNavigationHelper creates a new keyboard navigation helper
func NewKeyboardNavigationHelper(numElements int) *KeyboardNavigationHelper {
	return &KeyboardNavigationHelper{
		focusedIndex: 0,
		numElements:  numElements,
	}
}

// FocusedIndex returns the currently focused element index
func (h *KeyboardNavigationHelper) FocusedIndex() int {
	return h.focusedIndex
}

// SetFocusedIndex sets the focused element by index
func (h *KeyboardNavigationHelper) SetFocusedIndex(index int) {
	if index >= 0 && index < h.numElements {
		h.focusedIndex = index
	}
}

// MoveFocusNext moves focus to the next element (wraps around)
func (h *KeyboardNavigationHelper) MoveFocusNext() {
	if h.numElements > 0 {
		h.focusedIndex = (h.focusedIndex + 1) % h.numElements
	}
}

// MoveFocusPrev moves focus to the previous element (wraps around)
func (h *KeyboardNavigationHelper) MoveFocusPrev() {
	if h.numElements > 0 {
		h.focusedIndex = (h.focusedIndex - 1 + h.numElements) % h.numElements
	}
}

// HandleTabKey handles Tab/Shift+Tab navigation
func (h *KeyboardNavigationHelper) HandleTabKey(msg tea.KeyMsg) bool {
	if key.Matches(msg, key.NewBinding(key.WithKeys("tab"))) {
		h.MoveFocusNext()
		return true
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))) {
		h.MoveFocusPrev()
		return true
	}
	return false
}

// IsConfirmKey checks if the key is a confirmation key (Enter)
func IsConfirmKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("enter")))
}

// IsCancelKey checks if the key is a cancellation key (Esc)
func IsCancelKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("esc")))
}

// IsToggleKey checks if the key is a toggle key (Space)
func IsToggleKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("space", " ")))
}

// IsNavigationUp checks if the key moves up (Up or k)
func IsNavigationUp(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("up", "k")))
}

// IsNavigationDown checks if the key moves down (Down or j)
func IsNavigationDown(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("down", "j")))
}

// IsNavigationLeft checks if the key moves left (Left or h)
func IsNavigationLeft(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("left", "h")))
}

// IsNavigationRight checks if the key moves right (Right or l)
func IsNavigationRight(msg tea.KeyMsg) bool {
	return key.Matches(msg, key.NewBinding(key.WithKeys("right", "l")))
}

// StandardFooterHints returns the standard footer hints for common dialog types
func StandardFooterHints(dialogType string) []ShortcutHint {
	switch dialogType {
	case "form":
		return []ShortcutHint{
			{Key: "Tab", Label: "Next Field"},
			{Key: "Shift+Tab", Label: "Previous"},
			{Key: "Enter", Label: "Submit"},
			{Key: "Esc", Label: "Cancel"},
		}
	case "list":
		return []ShortcutHint{
			{Key: "↑/↓", Label: "Navigate"},
			{Key: "Enter", Label: "Select"},
			{Key: "Esc", Label: "Close"},
		}
	case "confirmation":
		return []ShortcutHint{
			{Key: "←/→", Label: "Change"},
			{Key: "Enter", Label: "Confirm"},
			{Key: "Esc", Label: "Cancel"},
		}
	case "file":
		return []ShortcutHint{
			{Key: "↑/↓", Label: "Navigate"},
			{Key: "Enter", Label: "Open"},
			{Key: "Backspace", Label: "Parent"},
			{Key: "Esc", Label: "Cancel"},
		}
	default:
		return []ShortcutHint{
			{Key: "Tab", Label: "Navigate"},
			{Key: "Enter", Label: "Confirm"},
			{Key: "Esc", Label: "Cancel"},
		}
	}
}
