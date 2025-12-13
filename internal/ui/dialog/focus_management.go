package dialog

import (
	"github.com/charmbracelet/lipgloss"
)

// FocusIndicatorStyle provides styling for focus indicators that work across all themes
type FocusIndicatorStyle struct {
	BorderColor      lipgloss.Color
	HighlightColor   lipgloss.Color
	CursorCharacter  string
	UnderlineActive  bool
	BoldActive       bool
}

// DefaultFocusIndicatorStyle returns a focus indicator style with high contrast defaults
func DefaultFocusIndicatorStyle() FocusIndicatorStyle {
	return FocusIndicatorStyle{
		BorderColor:     lipgloss.Color("#6D98BA"),
		HighlightColor:  lipgloss.Color("#00FF00"),
		CursorCharacter: "▶",
		UnderlineActive: true,
		BoldActive:      true,
	}
}

// HighContrastFocusIndicatorStyle returns a focus indicator style optimized for accessibility
func HighContrastFocusIndicatorStyle() FocusIndicatorStyle {
	return FocusIndicatorStyle{
		BorderColor:     lipgloss.Color("#FFFF00"),
		HighlightColor:  lipgloss.Color("#FFFF00"),
		CursorCharacter: ">>",
		UnderlineActive: true,
		BoldActive:      true,
	}
}

// RenderFocusedElement renders an element with clear focus indicators
// Parameters:
// - content: the text to render
// - isFocused: whether the element should show focus indicators
// - style: the style to apply
// - width: the width to apply
func RenderFocusedElement(content string, isFocused bool, style FocusIndicatorStyle, width int) string {
	baseStyle := lipgloss.NewStyle()

	if isFocused {
		baseStyle = baseStyle.Foreground(style.HighlightColor)

		if style.BoldActive {
			baseStyle = baseStyle.Bold(true)
		}

		if style.UnderlineActive {
			baseStyle = baseStyle.Underline(true)
		}

		// Add focus cursor if width permits
		if width > 2 && style.CursorCharacter != "" {
			content = style.CursorCharacter + " " + content
		}
	}

	if width > 0 {
		baseStyle = baseStyle.Width(width)
	}

	return baseStyle.Render(content)
}

// RenderFocusIndicator creates a visible focus indicator element
// This is useful for highlighting the currently focused element in a dialog
func RenderFocusIndicator(isFocused bool, style FocusIndicatorStyle) string {
	if !isFocused {
		return " "
	}

	indicatorStyle := lipgloss.NewStyle().
		Foreground(style.BorderColor).
		Bold(true)

	return indicatorStyle.Render("●")
}

// FocusTrappingValidator checks if focus is properly trapped within a dialog
// This is useful for testing and debugging focus behavior
type FocusTrappingValidator struct {
	focusedElements []int
	focusCycleCount int
}

// NewFocusTrappingValidator creates a new focus trapping validator
func NewFocusTrappingValidator() *FocusTrappingValidator {
	return &FocusTrappingValidator{
		focusedElements: []int{},
		focusCycleCount: 0,
	}
}

// RecordFocus records that an element received focus
func (v *FocusTrappingValidator) RecordFocus(elementIndex int) {
	v.focusedElements = append(v.focusedElements, elementIndex)

	// Check for cycling (when focus returns to the same element)
	if len(v.focusedElements) > 1 {
		lastIdx := len(v.focusedElements) - 1
		if v.focusedElements[lastIdx] == v.focusedElements[0] {
			v.focusCycleCount++
		}
	}
}

// IsFocusTrapped checks if focus has cycled properly (trapped within the dialog)
// A proper focus trap should cycle through all elements before returning to the first
func (v *FocusTrappingValidator) IsFocusTrapped() bool {
	return v.focusCycleCount > 0
}

// GetFocusHistory returns the recorded focus history
func (v *FocusTrappingValidator) GetFocusHistory() []int {
	return v.focusedElements
}

// Reset clears the focus history
func (v *FocusTrappingValidator) Reset() {
	v.focusedElements = []int{}
	v.focusCycleCount = 0
}
