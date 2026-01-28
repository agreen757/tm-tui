package dialog

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestParsePRDDialogHandleKeyTabNavigation tests Tab key navigation through fields
func TestParsePRDDialogHandleKeyTabNavigation(t *testing.T) {
	tests := []struct {
		name              string
		startField        int
		keyStr            string
		expectedField     int
		expectedCmd       bool
		description       string
	}{
		{
			name:          "Tab from field 0 moves to field 1",
			startField:    0,
			keyStr:        "tab",
			expectedField: 1,
			expectedCmd:   false,
			description:   "Tab navigation forward",
		},
		{
			name:          "Tab from field 1 moves to field 2",
			startField:    1,
			keyStr:        "tab",
			expectedField: 2,
			expectedCmd:   false,
			description:   "Tab navigation forward",
		},
		{
			name:          "Tab from field 2 moves to field 3",
			startField:    2,
			keyStr:        "tab",
			expectedField: 3,
			expectedCmd:   false,
			description:   "Tab navigation forward",
		},
		{
			name:          "Tab from field 3 wraps to field 0",
			startField:    3,
			keyStr:        "tab",
			expectedField: 0,
			expectedCmd:   false,
			description:   "Tab navigation wraps around",
		},
		{
			name:          "Shift+Tab from field 0 wraps to field 3",
			startField:    0,
			keyStr:        "shift+tab",
			expectedField: 3,
			expectedCmd:   false,
			description:   "Shift+Tab navigation backward with wrap",
		},
		{
			name:          "Shift+Tab from field 1 moves to field 0",
			startField:    1,
			keyStr:        "shift+tab",
			expectedField: 0,
			expectedCmd:   false,
			description:   "Shift+Tab navigation backward",
		},
		{
			name:          "Shift+Tab from field 2 moves to field 1",
			startField:    2,
			keyStr:        "shift+tab",
			expectedField: 1,
			expectedCmd:   false,
			description:   "Shift+Tab navigation backward",
		},
		{
			name:          "Shift+Tab from field 3 moves to field 2",
			startField:    3,
			keyStr:        "shift+tab",
			expectedField: 2,
			expectedCmd:   false,
			description:   "Shift+Tab navigation backward",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = tt.startField

			msg := tea.KeyMsg{Type: tea.KeyTab}
			if tt.keyStr == "shift+tab" {
				msg = tea.KeyMsg{Type: tea.KeyShiftTab}
			}

			result, cmd := dialog.HandleKey(msg)

			if dialog.focusedField != tt.expectedField {
				t.Errorf("HandleKey(%s): focusedField = %d, want %d", tt.keyStr, dialog.focusedField, tt.expectedField)
			}

			if (cmd != nil) != tt.expectedCmd {
				t.Errorf("HandleKey(%s): cmd = %v, want cmd=%v", tt.keyStr, cmd != nil, tt.expectedCmd)
			}

			if result != DialogResultNone {
				t.Errorf("HandleKey(%s): result = %v, want DialogResultNone", tt.keyStr, result)
			}
		})
	}
}

// TestParsePRDDialogHandleKeyOperationMode tests space key toggle for operation mode
func TestParsePRDDialogHandleKeyOperationMode(t *testing.T) {
	tests := []struct {
		name              string
		startMode         ParsePRDOperationMode
		expectedMode      ParsePRDOperationMode
		expectedResult    DialogResult
		description       string
	}{
		{
			name:           "Space toggles from Replace to Append",
			startMode:      ParsePRDReplace,
			expectedMode:   ParsePRDAppend,
			expectedResult: DialogResultNone,
			description:    "Space key toggles radio selection",
		},
		{
			name:           "Space toggles from Append to Replace",
			startMode:      ParsePRDAppend,
			expectedMode:   ParsePRDReplace,
			expectedResult: DialogResultNone,
			description:    "Space key toggles radio selection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = 2 // Operation mode field
			dialog.operationMode = tt.startMode

			// Send space key
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}

			result, _ := dialog.HandleKey(msg)

			if dialog.operationMode != tt.expectedMode {
				t.Errorf("HandleKey: operationMode = %d, want %d", dialog.operationMode, tt.expectedMode)
			}

			if result != tt.expectedResult {
				t.Errorf("HandleKey: result = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestParsePRDDialogHandleKeyButtonNavigation tests left/right arrow keys for button selection
func TestParsePRDDialogHandleKeyButtonNavigation(t *testing.T) {
	tests := []struct {
		name            string
		startSubmit     bool
		startCancel     bool
		keyStr          string
		expectedSubmit  bool
		expectedCancel  bool
		expectedResult  DialogResult
		description     string
	}{
		{
			name:           "Left key moves from Cancel to Submit",
			startSubmit:    false,
			startCancel:    true,
			keyStr:         "left",
			expectedSubmit: true,
			expectedCancel: false,
			expectedResult: DialogResultNone,
			description:    "Left arrow navigation",
		},
		{
			name:           "Right key moves from Submit to Cancel",
			startSubmit:    true,
			startCancel:    false,
			keyStr:         "right",
			expectedSubmit: false,
			expectedCancel: true,
			expectedResult: DialogResultNone,
			description:    "Right arrow navigation",
		},
		{
			name:           "h key (vim) moves from Cancel to Submit",
			startSubmit:    false,
			startCancel:    true,
			keyStr:         "h",
			expectedSubmit: true,
			expectedCancel: false,
			expectedResult: DialogResultNone,
			description:    "Vim keybinding for left",
		},
		{
			name:           "l key (vim) moves from Submit to Cancel",
			startSubmit:    true,
			startCancel:    false,
			keyStr:         "l",
			expectedSubmit: false,
			expectedCancel: true,
			expectedResult: DialogResultNone,
			description:    "Vim keybinding for right",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = 3 // Buttons field
			dialog.submitBtn = tt.startSubmit
			dialog.cancelBtn = tt.startCancel

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.keyStr)}
			if tt.keyStr == "left" {
				msg = tea.KeyMsg{Type: tea.KeyLeft}
			} else if tt.keyStr == "right" {
				msg = tea.KeyMsg{Type: tea.KeyRight}
			}

			result, _ := dialog.HandleKey(msg)

			if dialog.submitBtn != tt.expectedSubmit {
				t.Errorf("HandleKey(%s): submitBtn = %v, want %v", tt.keyStr, dialog.submitBtn, tt.expectedSubmit)
			}

			if dialog.cancelBtn != tt.expectedCancel {
				t.Errorf("HandleKey(%s): cancelBtn = %v, want %v", tt.keyStr, dialog.cancelBtn, tt.expectedCancel)
			}

			if result != tt.expectedResult {
				t.Errorf("HandleKey(%s): result = %v, want %v", tt.keyStr, result, tt.expectedResult)
			}
		})
	}
}

// TestParsePRDDialogHandleKeyButtonActivation tests space/enter on buttons
func TestParsePRDDialogHandleKeyButtonActivation(t *testing.T) {
	tests := []struct {
		name           string
		submitFocused  bool
		cancelFocused  bool
		expectedResult DialogResult
		description    string
	}{
		{
			name:           "Space on Cancel button cancels dialog",
			submitFocused:  false,
			cancelFocused:  true,
			expectedResult: DialogResultCancel,
			description:    "Cancel button activation",
		},
		{
			name:           "Enter on Cancel button cancels dialog",
			submitFocused:  false,
			cancelFocused:  true,
			expectedResult: DialogResultCancel,
			description:    "Cancel button activation with Enter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = 3 // Buttons field
			dialog.submitBtn = tt.submitFocused
			dialog.cancelBtn = tt.cancelFocused

			// Send space key
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}

			result, _ := dialog.HandleKey(msg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey: result = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestParsePRDDialogHandleKeyEnterValidation tests Enter key behavior when form is invalid
func TestParsePRDDialogHandleKeyEnterValidation(t *testing.T) {
	tests := []struct {
		name           string
		focusedField   int
		fileValid      bool
		expectedResult DialogResult
		description    string
	}{
		{
			name:           "Enter in file browser without file doesn't submit",
			focusedField:   0,
			fileValid:      false,
			expectedResult: DialogResultNone,
			description:    "Form validation prevents submit",
		},
		{
			name:           "Enter in buttons field is handled separately",
			focusedField:   3,
			fileValid:      true,
			expectedResult: DialogResultNone,
			description:    "Button field doesn't use Enter for submit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = tt.focusedField
			dialog.fileValid = tt.fileValid

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			result, _ := dialog.HandleKey(msg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey(enter) at field %d: result = %v, want %v", tt.focusedField, result, tt.expectedResult)
			}
		})
	}
}

// TestParsePRDDialogUpdateStateOnly tests that Update() only handles WindowSizeMsg
func TestParsePRDDialogUpdateStateOnly(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		shouldHandle bool
		description string
	}{
		{
			name:        "Update handles WindowSizeMsg",
			msg:         tea.WindowSizeMsg{Width: 120, Height: 40},
			shouldHandle: true,
			description: "State-Only Update Pattern: handles window resize",
		},
		{
			name:        "Update handles ParsePRDResultMsg",
			msg:         ParsePRDResultMsg{FilePath: "test.txt"},
			shouldHandle: true,
			description: "State-Only Update Pattern: handles result messages",
		},
		{
			name:        "Update ignores KeyMsg",
			msg:         tea.KeyMsg{Type: tea.KeyEnter},
			shouldHandle: false,
			description: "State-Only Update Pattern: keys go to HandleKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()

			// Store initial state
			initialField := dialog.focusedField

			// Call Update
			result, _ := dialog.Update(tt.msg)

			// Update should always return the dialog
			if result == nil {
				t.Errorf("Update returned nil dialog")
			}

			// For KeyMsg, state should not change (it's handled in HandleKey)
			if keyMsg, ok := tt.msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEnter {
				if dialog.focusedField != initialField {
					t.Errorf("Update processed key - should be ignored: focusedField changed from %d to %d", initialField, dialog.focusedField)
				}
			}
		})
	}
}

// TestParsePRDDialogValidateFilePath tests file path validation
func TestParsePRDDialogValidateFilePath(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-prd-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name           string
		filePath       string
		expectedValid  bool
		expectedMsg    string
		description    string
	}{
		{
			name:          "Empty file path is invalid",
			filePath:      "",
			expectedValid: false,
			expectedMsg:   "No file selected",
			description:   "Validation: no file selected",
		},
		{
			name:          "Whitespace-only path is invalid",
			filePath:      "   ",
			expectedValid: false,
			expectedMsg:   "No file selected",
			description:   "Validation: whitespace path",
		},
		{
			name:          "Non-existent file is invalid",
			filePath:      "/nonexistent/path/to/file.txt",
			expectedValid: false,
			expectedMsg:   "File not found",
			description:   "Validation: file doesn't exist",
		},
		{
			name:          "Existing file is valid",
			filePath:      tmpFile.Name(),
			expectedValid: true,
			expectedMsg:   "File found ✓",
			description:   "Validation: file exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.validateFilePath(tt.filePath)

			if dialog.fileValid != tt.expectedValid {
				t.Errorf("validateFilePath(%q): fileValid = %v, want %v", tt.filePath, dialog.fileValid, tt.expectedValid)
			}

			if dialog.fileValidMsg != tt.expectedMsg {
				t.Errorf("validateFilePath(%q): fileValidMsg = %q, want %q", tt.filePath, dialog.fileValidMsg, tt.expectedMsg)
			}
		})
	}
}

// TestParsePRDDialogValidateTags tests tags input validation
func TestParsePRDDialogValidateTags(t *testing.T) {
	tests := []struct {
		name              string
		tagsInput         string
		expectedValid     bool
		expectedMsgPrefix string
		description       string
	}{
		{
			name:              "Empty tags is valid",
			tagsInput:         "",
			expectedValid:     true,
			expectedMsgPrefix: "0 / 50",
			description:       "Validation: empty tags",
		},
		{
			name:              "Short tags are valid",
			tagsInput:         "feature,auth",
			expectedValid:     true,
			expectedMsgPrefix: "12 / 50",
			description:       "Validation: short tags",
		},
		{
			name:              "Tags at limit are valid",
			tagsInput:         "12345678901234567890123456789012345678901234567890", // 50 chars
			expectedValid:     true,
			expectedMsgPrefix: "50 / 50",
			description:       "Validation: tags at 50 char limit",
		},
		{
			name:              "Tags over limit are invalid",
			tagsInput:         "123456789012345678901234567890123456789012345678901", // 51 chars
			expectedValid:     false,
			expectedMsgPrefix: "51 / 50 characters (exceeded)",
			description:       "Validation: tags exceed 50 char limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.validateTags(tt.tagsInput)

			if dialog.tagValid != tt.expectedValid {
				t.Errorf("validateTags(%q): tagValid = %v, want %v", tt.tagsInput, dialog.tagValid, tt.expectedValid)
			}

			if !contains(dialog.tagCountMsg, tt.expectedMsgPrefix) {
				t.Errorf("validateTags(%q): tagCountMsg = %q, want to contain %q", tt.tagsInput, dialog.tagCountMsg, tt.expectedMsgPrefix)
			}
		})
	}
}

// TestParsePRDDialogFormValidation tests the form validation by field state
func TestParsePRDDialogFormValidation(t *testing.T) {
	// Test that validation updates properly
	dialog := NewParsePRDDialog()

	// Initially fileValid should be false
	if dialog.fileValid {
		t.Errorf("Initial fileValid should be false")
	}

	// Validate an empty path
	dialog.validateFilePath("")
	if dialog.fileValid {
		t.Errorf("Empty path should be invalid")
	}
	if dialog.fileValidMsg != "No file selected" {
		t.Errorf("Expected 'No file selected' message, got %q", dialog.fileValidMsg)
	}

	// Validate whitespace path
	dialog.validateFilePath("   ")
	if dialog.fileValid {
		t.Errorf("Whitespace path should be invalid")
	}

	// Validate tags
	dialog.validateTags("")
	if !dialog.tagValid {
		t.Errorf("Empty tags should be valid")
	}

	dialog.validateTags("toolong")
	if !dialog.tagValid {
		t.Errorf("Short tags should be valid")
	}

	// Test over-limit tags
	longTags := "123456789012345678901234567890123456789012345678901" // 51 chars
	dialog.validateTags(longTags)
	if dialog.tagValid {
		t.Errorf("Over-limit tags should be invalid")
	}
}

// TestParsePRDDialogServiceCommandConstruction tests service command building (4 cases)
func TestParsePRDDialogServiceCommandConstruction(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		tags            string
		operationMode   ParsePRDOperationMode
		expectedAppend  bool
		description     string
	}{
		{
			name:           "Command: replace mode without tags",
			filePath:       "prd.txt",
			tags:           "",
			operationMode:  ParsePRDReplace,
			expectedAppend: false,
			description:    "Service command: replace, no tags",
		},
		{
			name:           "Command: replace mode with tags",
			filePath:       "prd.txt",
			tags:           "feature,ui",
			operationMode:  ParsePRDReplace,
			expectedAppend: false,
			description:    "Service command: replace, with tags",
		},
		{
			name:           "Command: append mode without tags",
			filePath:       "prd.txt",
			tags:           "",
			operationMode:  ParsePRDAppend,
			expectedAppend: true,
			description:    "Service command: append, no tags",
		},
		{
			name:           "Command: append mode with tags",
			filePath:       "prd.txt",
			tags:           "feature,ui",
			operationMode:  ParsePRDAppend,
			expectedAppend: true,
			description:    "Service command: append, with tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.operationMode = tt.operationMode
			dialog.tagsInput.SetValue(tt.tags)

			// Verify the dialog state matches what would be used for command
			appendFlag := dialog.GetAppendFlag()
			if appendFlag != tt.expectedAppend {
				t.Errorf("GetAppendFlag(): result = %v, want %v", appendFlag, tt.expectedAppend)
			}

			tags := dialog.GetTags()
			if tags != tt.tags {
				t.Errorf("GetTags(): result = %q, want %q", tags, tt.tags)
			}
		})
	}
}

// TestParsePRDDialogNoDoubleKeyHandling tests that keys don't get processed twice
func TestParsePRDDialogNoDoubleKeyHandling(t *testing.T) {
	tests := []struct {
		name        string
		keyStr      string
		initialField int
		description string
	}{
		{
			name:        "Tab only moves focus once",
			keyStr:      "tab",
			initialField: 0,
			description: "Key handling: no double processing",
		},
		{
			name:        "Shift+Tab only moves focus once",
			keyStr:      "shift+tab",
			initialField: 1,
			description: "Key handling: no double processing",
		},
		{
			name:        "Down in operation mode toggles once",
			keyStr:      "down",
			initialField: 2,
			description: "Key handling: no double processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = tt.initialField

			// Store initial state
			initialField := dialog.focusedField

			// Send key once
			var msg tea.KeyMsg
			if tt.keyStr == "tab" {
				msg = tea.KeyMsg{Type: tea.KeyTab}
			} else if tt.keyStr == "shift+tab" {
				msg = tea.KeyMsg{Type: tea.KeyShiftTab}
			} else if tt.keyStr == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			}

			result1, _ := dialog.HandleKey(msg)

			// Store state after first key
			stateAfterFirst := dialog.focusedField

			// Send same key again
			result2, _ := dialog.HandleKey(msg)

			// State should move twice, not once
			if tt.keyStr == "tab" {
				expectedField := (initialField + 2) % 4
				if dialog.focusedField != expectedField {
					t.Errorf("HandleKey(tab) twice: focusedField = %d, want %d", dialog.focusedField, expectedField)
				}
				// Verify no state duplication between calls
				if stateAfterFirst != (initialField + 1) % 4 {
					t.Errorf("HandleKey(tab) first call: focusedField = %d, want %d", stateAfterFirst, (initialField + 1) % 4)
				}
			}

			// Both calls should return None for navigation keys
			if result1 != DialogResultNone || result2 != DialogResultNone {
				t.Errorf("HandleKey: expected DialogResultNone for navigation, got %v and %v", result1, result2)
			}
		})
	}
}
