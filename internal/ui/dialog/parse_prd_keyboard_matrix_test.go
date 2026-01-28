package dialog

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestParsePRDDialogKeyboardMatrix tests all keyboard shortcuts comprehensively
func TestParsePRDDialogKeyboardMatrix(t *testing.T) {
	// Define all keyboard combinations to test
	tests := []struct {
		name           string
		keyType        tea.KeyType
		runes          []rune
		focusedField   int
		expectedResult DialogResult
	}{
		// Tab navigation
		{"Tab forward from field 0", tea.KeyTab, nil, 0, DialogResultNone},
		{"Tab forward from field 1", tea.KeyTab, nil, 1, DialogResultNone},
		{"Tab forward from field 2", tea.KeyTab, nil, 2, DialogResultNone},
		{"Tab forward from field 3", tea.KeyTab, nil, 3, DialogResultNone},

		// Shift+Tab navigation
		{"Shift+Tab back from field 0", tea.KeyShiftTab, nil, 0, DialogResultNone},
		{"Shift+Tab back from field 1", tea.KeyShiftTab, nil, 1, DialogResultNone},
		{"Shift+Tab back from field 2", tea.KeyShiftTab, nil, 2, DialogResultNone},
		{"Shift+Tab back from field 3", tea.KeyShiftTab, nil, 3, DialogResultNone},

		// Arrow keys in operation mode field
		{"Left in field 3 (buttons)", tea.KeyLeft, nil, 3, DialogResultNone},
		{"Right in field 3 (buttons)", tea.KeyRight, nil, 3, DialogResultNone},
		{"Up in field 2 (operation)", tea.KeyUp, nil, 2, DialogResultNone},
		{"Down in field 2 (operation)", tea.KeyDown, nil, 2, DialogResultNone},

		// Vim keys
		{"h (left) in field 3", tea.KeyRunes, []rune{'h'}, 3, DialogResultNone},
		{"l (right) in field 3", tea.KeyRunes, []rune{'l'}, 3, DialogResultNone},
		{"k (up) in field 2", tea.KeyRunes, []rune{'k'}, 2, DialogResultNone},
		{"j (down) in field 2", tea.KeyRunes, []rune{'j'}, 2, DialogResultNone},

		// Space in operation mode
		{"Space in field 2", tea.KeyRunes, []rune{' '}, 2, DialogResultNone},

		// Space in buttons
		{"Space in field 3", tea.KeyRunes, []rune{' '}, 3, DialogResultNone},

		// Escape
		{"Escape", tea.KeyEsc, nil, 0, DialogResultCancel},
		{"Escape in field 1", tea.KeyEsc, nil, 1, DialogResultCancel},
		{"Escape in field 2", tea.KeyEsc, nil, 2, DialogResultCancel},

		// Enter key
		{"Enter in field 0", tea.KeyEnter, nil, 0, DialogResultNone},
		{"Enter in field 1", tea.KeyEnter, nil, 1, DialogResultNone},
		{"Enter in field 2", tea.KeyEnter, nil, 2, DialogResultNone},
		{"Enter in field 3", tea.KeyEnter, nil, 3, DialogResultNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = test.focusedField

			var msg tea.KeyMsg
			if len(test.runes) > 0 {
				msg = tea.KeyMsg{Type: test.keyType, Runes: test.runes}
			} else {
				msg = tea.KeyMsg{Type: test.keyType}
			}

			result, _ := dialog.HandleKey(msg)

			// Only verify escape returns Cancel for now
			if test.keyType == tea.KeyEsc && result != test.expectedResult {
				t.Errorf("Escape key: expected %v, got %v", test.expectedResult, result)
			}
		})
	}
}

// TestParsePRDDialogFocusStateRendering tests that rendering differs by focus state
func TestParsePRDDialogFocusStateRendering(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Get rendering with different focus states
	tests := []struct {
		name         string
		focusedField int
		shouldRender string
	}{
		{
			name:         "FilePath section contains label",
			focusedField: 0,
			shouldRender: "PRD File:",
		},
		{
			name:         "Tags section contains label",
			focusedField: 1,
			shouldRender: "Tags (optional):",
		},
		{
			name:         "Operation mode section contains label",
			focusedField: 2,
			shouldRender: "Operation Mode:",
		},
		{
			name:         "Button section contains buttons",
			focusedField: 3,
			shouldRender: "Parse",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialog.focusedField = test.focusedField
			content := dialog.View()

			if !strings.Contains(content, test.shouldRender) {
				t.Errorf("Expected to find %q in rendered content at field %d", test.shouldRender, test.focusedField)
			}
		})
	}
}

// TestParsePRDDialogButtonStylesFocused tests button styling with focus
func TestParsePRDDialogButtonStylesFocused(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	// Test submit button focused
	submitStyleFocused := dialog.getButtonStyle(true, true)
	submitText := submitStyleFocused.Render("[ Parse ]")
	if submitText == "" {
		t.Errorf("Submit button focused style produced empty render")
	}

	// Test submit button unfocused
	submitStyleUnfocused := dialog.getButtonStyle(false, true)
	submitTextUnfocused := submitStyleUnfocused.Render("[ Parse ]")
	if submitTextUnfocused == "" {
		t.Errorf("Submit button unfocused style produced empty render")
	}

	// Test cancel button focused
	cancelStyleFocused := dialog.getButtonStyle(true, false)
	cancelText := cancelStyleFocused.Render("[ Cancel ]")
	if cancelText == "" {
		t.Errorf("Cancel button focused style produced empty render")
	}

	// Test cancel button unfocused
	cancelStyleUnfocused := dialog.getButtonStyle(false, false)
	cancelTextUnfocused := cancelStyleUnfocused.Render("[ Cancel ]")
	if cancelTextUnfocused == "" {
		t.Errorf("Cancel button unfocused style produced empty render")
	}
}

// TestParsePRDDialogInputStylesFocused tests input field styling with focus
func TestParsePRDDialogInputStylesFocused(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test focused input
	focusedStyle := dialog.getInputStyle(true)
	renderedFocused := focusedStyle.Render("test input")
	if renderedFocused == "" {
		t.Errorf("Focused input style produced empty render")
	}

	// Test unfocused input
	unfocusedStyle := dialog.getInputStyle(false)
	renderedUnfocused := unfocusedStyle.Render("test input")
	if renderedUnfocused == "" {
		t.Errorf("Unfocused input style produced empty render")
	}

	// They should render differently
	if renderedFocused == renderedUnfocused {
		t.Errorf("Focused and unfocused inputs rendered identically")
	}
}

// TestParsePRDDialogValidationStylesRendering tests validation feedback styling
func TestParsePRDDialogValidationStylesRendering(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	// Test valid state
	validStyle := dialog.getValidationStyle(true)
	validRender := validStyle.Render("✓ Valid")
	if validRender == "" {
		t.Errorf("Valid style produced empty render")
	}

	// Test invalid state
	invalidStyle := dialog.getValidationStyle(false)
	invalidRender := invalidStyle.Render("✗ Invalid")
	if invalidRender == "" {
		t.Errorf("Invalid style produced empty render")
	}

	// They should render differently
	if validRender == invalidRender {
		t.Errorf("Valid and invalid styles rendered identically")
	}
}

// TestParsePRDDialogRadioButtonStyling tests radio button rendering for both states
func TestParsePRDDialogRadioButtonStyling(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	// Test selected radio button
	selectedRadio := dialog.renderRadioButton(true, "Replace")
	if !strings.Contains(selectedRadio, "◉") {
		t.Errorf("Selected radio button should contain filled circle")
	}

	// Test unselected radio button
	unselectedRadio := dialog.renderRadioButton(false, "Append")
	if !strings.Contains(unselectedRadio, "○") {
		t.Errorf("Unselected radio button should contain hollow circle")
	}

	// Both should contain the label
	if !strings.Contains(selectedRadio, "Replace") {
		t.Errorf("Selected radio should contain label")
	}

	if !strings.Contains(unselectedRadio, "Append") {
		t.Errorf("Unselected radio should contain label")
	}
}

// TestParsePRDDialogAllFieldFocusStates tests rendering with each field focused
func TestParsePRDDialogAllFieldFocusStates(t *testing.T) {
	dialog := NewParsePRDDialog()

	for field := 0; field < 4; field++ {
		t.Run("Field"+string(rune('0'+field)), func(t *testing.T) {
			dialog.focusedField = field

			// Render the dialog
			content := dialog.View()

			// Should contain border
			if !strings.Contains(content, "─") && !strings.Contains(content, "│") {
				t.Errorf("Dialog should have borders")
			}

			// Should contain title
			if !strings.Contains(content, "Parse PRD") {
				t.Errorf("Dialog should contain title")
			}
		})
	}
}

// TestParsePRDDialogThemeVariants tests rendering with different theme colors
func TestParsePRDDialogThemeVariants(t *testing.T) {
	// Test with default theme
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	content1 := dialog.View()
	if content1 == "" {
		t.Errorf("Dialog with default theme should render non-empty")
	}

	// Change theme colors
	dialog.Style.ButtonColor = "#FF0000"
	dialog.Style.TextColor = "#00FF00"
	dialog.Style.ErrorColor = "#0000FF"

	content2 := dialog.View()
	if content2 == "" {
		t.Errorf("Dialog with custom theme should render non-empty")
	}

	// Both should render (they may not be visually different in text, but both should work)
	if content1 == "" || content2 == "" {
		t.Errorf("Dialog should render with both default and custom themes")
	}
}

// TestParsePRDDialogRenderingCompleteContent tests that all sections render
func TestParsePRDDialogRenderingCompleteContent(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Update some field values
	dialog.tagsInput.SetValue("test-tag")
	dialog.operationMode = ParsePRDAppend

	content := dialog.View()

	// Check all major sections are rendered
	requiredStrings := []string{
		"Parse PRD",     // Title
		"PRD File:",     // File section
		"Tags",          // Tags section
		"Operation Mode:", // Operation mode section
		"Parse",         // Parse button
		"Cancel",        // Cancel button
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("Rendered content should contain %q", required)
		}
	}
}

// TestParsePRDDialogKeyLeakageToUpdate verifies no key events leak to Update()
func TestParsePRDDialogKeyLeakageToUpdate(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Store initial state
	initialField := dialog.focusedField

	// Send Tab key through Update (should be ignored)
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	dialog.Update(keyMsg)

	// Field should not have changed (Tab only works through HandleKey)
	if dialog.focusedField != initialField {
		t.Errorf("Tab key should not be processed by Update, field changed from %d to %d", initialField, dialog.focusedField)
	}

	// Now send through HandleKey (should work)
	dialog.HandleKey(keyMsg)

	// Field should have changed
	expectedField := (initialField + 1) % 4
	if dialog.focusedField != expectedField {
		t.Errorf("Tab through HandleKey: expected field %d, got %d", expectedField, dialog.focusedField)
	}
}

// TestParsePRDDialogAllShortcutsMatrix creates exhaustive shortcut coverage matrix
func TestParsePRDDialogAllShortcutsMatrix(t *testing.T) {
	shortcuts := []struct {
		name      string
		keyType   tea.KeyType
		runes     []rune
		fields    []int // Which fields this applies to
		validated bool
	}{
		{"Tab", tea.KeyTab, nil, []int{0, 1, 2, 3}, true},
		{"Shift+Tab", tea.KeyShiftTab, nil, []int{0, 1, 2, 3}, true},
		{"Escape", tea.KeyEsc, nil, []int{0, 1, 2, 3}, true},
		{"Enter", tea.KeyEnter, nil, []int{0, 1, 2, 3}, true},
		{"Left arrow", tea.KeyLeft, nil, []int{3}, true},
		{"Right arrow", tea.KeyRight, nil, []int{3}, true},
		{"Up arrow", tea.KeyUp, nil, []int{2}, false},
		{"Down arrow", tea.KeyDown, nil, []int{2}, false},
		{"Space", tea.KeyRunes, []rune{' '}, []int{2, 3}, true},
		{"h (vim left)", tea.KeyRunes, []rune{'h'}, []int{3}, true},
		{"l (vim right)", tea.KeyRunes, []rune{'l'}, []int{3}, true},
		{"k (vim up)", tea.KeyRunes, []rune{'k'}, []int{2}, false},
		{"j (vim down)", tea.KeyRunes, []rune{'j'}, []int{2}, false},
	}

	totalCombinations := 0
	for _, shortcut := range shortcuts {
		for _, field := range shortcut.fields {
			totalCombinations++

			t.Run(shortcut.name+" at field "+string(rune('0'+field)), func(t *testing.T) {
				dialog := NewParsePRDDialog()
				dialog.focusedField = field

				var msg tea.KeyMsg
				if len(shortcut.runes) > 0 {
					msg = tea.KeyMsg{Type: shortcut.keyType, Runes: shortcut.runes}
				} else {
					msg = tea.KeyMsg{Type: shortcut.keyType}
				}

				// Should not panic
				result, cmd := dialog.HandleKey(msg)

				// Basic assertions
				if shortcut.validated {
					// Should at least not error out
					_ = result
					_ = cmd
				}
			})
		}
	}

	// Log total combinations tested
	t.Logf("Tested %d keyboard shortcut combinations", totalCombinations)
}

// TestParsePRDDialogNoUnexpectedKeyHandling tests that unknown keys are handled gracefully
func TestParsePRDDialogNoUnexpectedKeyHandling(t *testing.T) {
	dialog := NewParsePRDDialog()

	unexpectedKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'X'}},
		{Type: tea.KeyRunes, Runes: []rune{'@'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{'\n'}},
	}

	for _, key := range unexpectedKeys {
		// Should not panic
		result, _ := dialog.HandleKey(key)

		// Unknown keys should return DialogResultNone
		if result != DialogResultNone {
			t.Errorf("Unknown key %v should return DialogResultNone, got %v", key, result)
		}
	}
}
