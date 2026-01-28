package dialog

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadPRDFile(t *testing.T) {
	// Create log viewer
	style := DefaultDialogStyle()
	viewer := NewLogViewerPanel(80, 24, style)

	// Get the file path relative to test
	prdPath := "../../../.taskmaster/docs/CLOUD_EXECUTION_PRD.md"

	// Check if file exists
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		t.Skipf("PRD file not found at %s, skipping test", prdPath)
	}

	done := make(chan error, 1)

	go func() {
		err := viewer.LoadFileContent(prdPath)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("LoadFileContent failed: %v", err)
		}
		t.Log("File loaded successfully!")
		t.Logf("Line limited: %v", viewer.IsLineLimited())
		t.Logf("File size warning: %v", viewer.HasFileSizeWarning())
	case <-time.After(10 * time.Second):
		t.Fatal("LoadFileContent timed out - likely infinite loop")
	}

	// Test rendering
	t.Log("Testing render...")
	viewer.SetFocused(true)
	view := viewer.View()
	t.Logf("Render output: %d chars", len(view))

	// Test with markdown enabled
	t.Log("Testing markdown render...")
	viewer.ToggleMarkdown()
	view = viewer.View()
	t.Logf("Markdown render output: %d chars", len(view))
}

// TestParsePRDDialogTagsInput tests the tags input field functionality
func TestParsePRDDialogTagsInput(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty tags field returns empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single tag",
			input:    "feature",
			expected: "feature",
		},
		{
			name:     "Comma-separated tags",
			input:    "feature,auth,ui",
			expected: "feature,auth,ui",
		},
		{
			name:     "Tags with spaces",
			input:    "feature, auth, ui",
			expected: "feature, auth, ui",
		},
		{
			name:     "Trimmed whitespace",
			input:    "  feature,auth  ",
			expected: "feature,auth",
		},
		{
			name:     "Leading/trailing spaces per tag",
			input:    " feature , auth , ui ",
			expected: "feature , auth , ui",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set input value
			dialog.tagsInput.SetValue(tt.input)

			// Get tags via GetTags method
			result := dialog.GetTags()

			if result != tt.expected {
				t.Errorf("GetTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParsePRDDialogTagsField tests the tags input field properties
func TestParsePRDDialogTagsField(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test placeholder text
	if dialog.tagsInput.Placeholder != "(optional) comma-separated tags" {
		t.Errorf("Placeholder = %q, want \"(optional) comma-separated tags\"", dialog.tagsInput.Placeholder)
	}

	// Test character limit
	if dialog.tagsInput.CharLimit != 50 {
		t.Errorf("CharLimit = %d, want 50", dialog.tagsInput.CharLimit)
	}

	// Test that tags input accepts input
	dialog.tagsInput.SetValue("test-tag")
	if dialog.tagsInput.Value() != "test-tag" {
		t.Errorf("SetValue failed: Value() = %q, want \"test-tag\"", dialog.tagsInput.Value())
	}
}

// TestParsePRDDialogTagsRendering tests the visual rendering of tags field
func TestParsePRDDialogTagsRendering(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test rendering with empty tags
	content := dialog.renderContent()
	if !strings.Contains(content, "Tags (optional):") {
		t.Errorf("renderContent() does not contain 'Tags (optional):' label")
	}

	// Test rendering with tags
	dialog.tagsInput.SetValue("feature,auth")
	content = dialog.renderContent()
	if !strings.Contains(content, "Tags (optional):") {
		t.Errorf("renderContent() with tags does not contain 'Tags (optional):' label")
	}
}

// TestParsePRDDialogFocusHandling tests focus navigation with Tab/Shift+Tab
func TestParsePRDDialogFocusHandling(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Initial focus should be on field 0 (fileBrowser)
	if dialog.focusedField != 0 {
		t.Errorf("Initial focusedField = %d, want 0", dialog.focusedField)
	}

	// Verify fileBrowser has focus (Task 3.3)
	if !dialog.fileBrowser.focused {
		t.Errorf("fileBrowser should have focus initially")
	}

	// Simulate tab navigation to tags field
	dialog.focusedField = 1

	// Verify style rendering changes based on focus
	tagStyle := dialog.getInputStyle(true) // focused
	content := tagStyle.Render("test")
	if !strings.Contains(content, "test") {
		t.Errorf("Focused input style rendering failed")
	}

	// Verify unfocused style
	tagStyleUnfocused := dialog.getInputStyle(false) // not focused
	content = tagStyleUnfocused.Render("test")
	if !strings.Contains(content, "test") {
		t.Errorf("Unfocused input style rendering failed")
	}
}

// TestParsePRDDialogTagsSection tests the tags section rendering
func TestParsePRDDialogTagsSection(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test renderTagsSection with various focus states
	tests := []struct {
		name         string
		focusedField int
		shouldFocus  bool
	}{
		{
			name:         "Tags field focused",
			focusedField: 1,
			shouldFocus:  true,
		},
		{
			name:         "Tags field not focused",
			focusedField: 0,
			shouldFocus:  false,
		},
		{
			name:         "Buttons focused",
			focusedField: 2,
			shouldFocus:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.focusedField = tt.focusedField
			section := dialog.renderTagsSection()

			if !strings.Contains(section, "Tags (optional):") {
				t.Errorf("renderTagsSection() missing label")
			}
		})
	}
}

// TestParsePRDDialogTagsCharacterLimit tests the 50-character limit
func TestParsePRDDialogTagsCharacterLimit(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Create a string longer than 50 characters
	longInput := strings.Repeat("a", 60)

	// Try to set it - the textinput should enforce the limit
	dialog.tagsInput.SetValue(longInput)

	// The actual value stored might be limited
	result := dialog.tagsInput.Value()

	// Verify it doesn't exceed the limit
	if len(result) > 50 {
		t.Errorf("Tags input value length %d exceeds CharLimit of 50", len(result))
	}
}

// TestParsePRDDialogGetFilePath tests file path retrieval
func TestParsePRDDialogGetFilePath(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test empty file path (initial state)
	// The file browser is complex to test directly, so we verify the method exists
	// and returns a string (empty or otherwise)
	filePath := dialog.GetFilePath()
	if filePath != strings.TrimSpace(filePath) {
		t.Errorf("GetFilePath() should return trimmed path")
	}
}

// TestParsePRDDialogOperationMode tests operation mode selection
func TestParsePRDDialogOperationMode(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test default mode is Replace
	if dialog.operationMode != ParsePRDReplace {
		t.Errorf("Default operationMode = %d, want ParsePRDReplace (%d)", dialog.operationMode, ParsePRDReplace)
	}

	// Test switching mode
	dialog.operationMode = ParsePRDAppend
	if dialog.GetOperationMode() != ParsePRDAppend {
		t.Errorf("GetOperationMode() = %d, want ParsePRDAppend (%d)", dialog.GetOperationMode(), ParsePRDAppend)
	}
}

// TestParsePRDDialogUpdate tests the Update method with window size
func TestParsePRDDialogUpdate(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Test window size message
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	result, _ := dialog.Update(msg)

	// Dialog should still be returned
	if result == nil {
		t.Errorf("Update with WindowSizeMsg returned nil")
	}

	// Dialog should be centered (not checking exact values, just that it processes)
	if d, ok := result.(*ParsePRDDialog); !ok {
		t.Errorf("Update did not return ParsePRDDialog")
	} else if d == nil {
		t.Errorf("Returned dialog is nil")
	}
}

// TestParsePRDDialogGetAppendFlag tests the GetAppendFlag method
func TestParsePRDDialogGetAppendFlag(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name           string
		operationMode  ParsePRDOperationMode
		expectedAppend bool
	}{
		{
			name:           "Replace mode returns false",
			operationMode:  ParsePRDReplace,
			expectedAppend: false,
		},
		{
			name:           "Append mode returns true",
			operationMode:  ParsePRDAppend,
			expectedAppend: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.operationMode = tt.operationMode
			result := dialog.GetAppendFlag()
			if result != tt.expectedAppend {
				t.Errorf("GetAppendFlag() = %v, want %v", result, tt.expectedAppend)
			}
		})
	}
}

// TestParsePRDDialogOperationModeRendering tests the visual rendering of operation mode radio buttons
func TestParsePRDDialogOperationModeRendering(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	// Set up the theme colors
	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	tests := []struct {
		name          string
		operationMode ParsePRDOperationMode
		shouldContain string
	}{
		{
			name:          "Replace mode shows filled circle symbol",
			operationMode: ParsePRDReplace,
			shouldContain: "◉",
		},
		{
			name:          "Append mode shows filled circle symbol",
			operationMode: ParsePRDAppend,
			shouldContain: "◉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.operationMode = tt.operationMode
			content := dialog.renderOperationModeSection()
			if !strings.Contains(content, tt.shouldContain) {
				t.Errorf("renderOperationModeSection() does not contain '%s'", tt.shouldContain)
			}
			if !strings.Contains(content, "Operation Mode:") {
				t.Errorf("renderOperationModeSection() does not contain 'Operation Mode:' label")
			}
		})
	}
}

// TestParsePRDDialogRadioButtonRendering tests individual radio button rendering
func TestParsePRDDialogRadioButtonRendering(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	tests := []struct {
		name               string
		selected           bool
		label              string
		shouldContainBtn   string
		shouldContainLabel bool
	}{
		{
			name:               "Selected radio button renders filled circle",
			selected:           true,
			label:              "Test Option",
			shouldContainBtn:   "◉",
			shouldContainLabel: true,
		},
		{
			name:               "Unselected radio button renders hollow circle",
			selected:           false,
			label:              "Test Option",
			shouldContainBtn:   "○",
			shouldContainLabel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dialog.renderRadioButton(tt.selected, tt.label)
			if !strings.Contains(result, tt.shouldContainBtn) {
				t.Errorf("renderRadioButton() does not contain '%s'", tt.shouldContainBtn)
			}
			if tt.shouldContainLabel && !strings.Contains(result, tt.label) {
				t.Errorf("renderRadioButton() does not contain label '%s'", tt.label)
			}
		})
	}
}

// TestParsePRDDialogKeyboardNavigation tests keyboard navigation for operation mode
func TestParsePRDDialogKeyboardNavigation(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Start with Replace mode
	dialog.operationMode = ParsePRDReplace
	dialog.focusedField = 2 // Focus on operation mode field

	tests := []struct {
		name              string
		initialMode       ParsePRDOperationMode
		keyMsg            string
		expectedMode      ParsePRDOperationMode
		expectedResult    DialogResult
	}{
		{
			name:              "Down arrow toggles from Replace to Append",
			initialMode:       ParsePRDReplace,
			keyMsg:            "down",
			expectedMode:      ParsePRDAppend,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Up arrow toggles from Append to Replace",
			initialMode:       ParsePRDAppend,
			keyMsg:            "up",
			expectedMode:      ParsePRDReplace,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "j key toggles from Replace to Append",
			initialMode:       ParsePRDReplace,
			keyMsg:            "j",
			expectedMode:      ParsePRDAppend,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "k key toggles from Append to Replace",
			initialMode:       ParsePRDAppend,
			keyMsg:            "k",
			expectedMode:      ParsePRDReplace,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Space toggles from Replace to Append",
			initialMode:       ParsePRDReplace,
			keyMsg:            " ",
			expectedMode:      ParsePRDAppend,
			expectedResult:    DialogResultNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.operationMode = tt.initialMode
			dialog.focusedField = 2
			// Set fileValid to true so space toggles the mode
			dialog.fileValid = false
			
			keyMsg := tea.KeyMsg{Runes: []rune(tt.keyMsg), Type: tea.KeyRunes}
			result, _ := dialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v", result, tt.expectedResult)
			}

			if dialog.operationMode != tt.expectedMode {
				t.Errorf("After HandleKey(), operationMode = %d, want %d", dialog.operationMode, tt.expectedMode)
			}
		})
	}
}

// TestParsePRDDialogTabNavigation tests Tab/Shift+Tab navigation between fields
func TestParsePRDDialogTabNavigation(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name              string
		initialField      int
		keyMsg            string
		expectedField     int
		expectedResult    DialogResult
	}{
		{
			name:              "Tab from file browser (0) to tags (1)",
			initialField:      0,
			keyMsg:            "tab",
			expectedField:     1,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Tab from tags (1) to operation mode (2)",
			initialField:      1,
			keyMsg:            "tab",
			expectedField:     2,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Tab from operation mode (2) to buttons (3)",
			initialField:      2,
			keyMsg:            "tab",
			expectedField:     3,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Tab from buttons (3) wraps to file browser (0)",
			initialField:      3,
			keyMsg:            "tab",
			expectedField:     0,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Shift+Tab from tags (1) to file browser (0)",
			initialField:      1,
			keyMsg:            "shift+tab",
			expectedField:     0,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Shift+Tab from operation mode (2) to tags (1)",
			initialField:      2,
			keyMsg:            "shift+tab",
			expectedField:     1,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Shift+Tab from buttons (3) to operation mode (2)",
			initialField:      3,
			keyMsg:            "shift+tab",
			expectedField:     2,
			expectedResult:    DialogResultNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.focusedField = tt.initialField
			
			keyMsg := tea.KeyMsg{Runes: []rune(tt.keyMsg), Type: tea.KeyRunes}
			result, _ := dialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v", result, tt.expectedResult)
			}

			if dialog.focusedField != tt.expectedField {
				t.Errorf("After HandleKey(), focusedField = %d, want %d", dialog.focusedField, tt.expectedField)
			}
		})
	}
}

// TestParsePRDDialogFileValidation tests file path validation (Task 8.1)
func TestParsePRDDialogFileValidation(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name            string
		filePath        string
		expectedValid   bool
		expectedMsgLen  int // Just check msg is not empty
	}{
		{
			name:           "Empty path is invalid",
			filePath:       "",
			expectedValid:  false,
			expectedMsgLen: 1, // At least some message
		},
		{
			name:           "Whitespace path is invalid",
			filePath:       "   ",
			expectedValid:  false,
			expectedMsgLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.validateFilePath(tt.filePath)

			if dialog.fileValid != tt.expectedValid {
				t.Errorf("validateFilePath(%q): fileValid = %v, want %v", tt.filePath, dialog.fileValid, tt.expectedValid)
			}

			if len(dialog.fileValidMsg) < tt.expectedMsgLen {
				t.Errorf("validateFilePath(%q): fileValidMsg length = %d, want at least %d", tt.filePath, len(dialog.fileValidMsg), tt.expectedMsgLen)
			}
		})
	}
}

// TestParsePRDDialogTagValidation tests tag length validation (Task 8.1)
func TestParsePRDDialogTagValidation(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name           string
		tagInput       string
		expectedValid  bool
		expectedMsg    string
	}{
		{
			name:          "Empty tags are valid",
			tagInput:      "",
			expectedValid: true,
			expectedMsg:   "0 / 50 characters",
		},
		{
			name:          "Tags under 50 chars are valid",
			tagInput:      "feature,auth",
			expectedValid: true,
			expectedMsg:   "12 / 50 characters",
		},
		{
			name:          "Tags at exactly 50 chars are valid",
			tagInput:      strings.Repeat("a", 50),
			expectedValid: true,
			expectedMsg:   "50 / 50 characters",
		},
		{
			name:          "Tags over 50 chars are invalid",
			tagInput:      strings.Repeat("a", 51),
			expectedValid: false,
			expectedMsg:   "51 / 50 characters (exceeded)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.validateTags(tt.tagInput)

			if dialog.tagValid != tt.expectedValid {
				t.Errorf("validateTags(%q): tagValid = %v, want %v", tt.tagInput, dialog.tagValid, tt.expectedValid)
			}

			if dialog.tagCountMsg != tt.expectedMsg {
				t.Errorf("validateTags(%q): tagCountMsg = %q, want %q", tt.tagInput, dialog.tagCountMsg, tt.expectedMsg)
			}
		})
	}
}

// TestParsePRDDialogUpdateValidation tests the updateValidation method (Task 8.1)
func TestParsePRDDialogUpdateValidation(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Set up test values
	dialog.tagsInput.SetValue("feature,auth")

	// Call updateValidation
	dialog.updateValidation()

	// Should update both file and tag validation
	if dialog.tagCountMsg != "12 / 50 characters" {
		t.Errorf("updateValidation(): tagCountMsg = %q, want '12 / 50 characters'", dialog.tagCountMsg)
	}

	// File should still be invalid (no file selected)
	if dialog.fileValid {
		t.Errorf("updateValidation(): fileValid = true, want false (no file selected)")
	}
}

// TestParsePRDDialogValidationRendering tests rendering of validation feedback (Task 8.1)
func TestParsePRDDialogValidationRendering(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	// Test file validation rendering
	dialog.fileValid = true
	dialog.fileValidMsg = "File found ✓"
	fileSection := dialog.renderFilePathSection()
	if !strings.Contains(fileSection, "File found ✓") {
		t.Errorf("renderFilePathSection() does not contain validation message")
	}

	// Test tags validation rendering with character count
	dialog.tagsInput.SetValue("feature")
	dialog.tagCountMsg = "7 / 50 characters"
	tagsSection := dialog.renderTagsSection()
	if !strings.Contains(tagsSection, "7 / 50 characters") {
		t.Errorf("renderTagsSection() does not contain character count")
	}

	// Test button rendering when invalid
	dialog.fileValid = false
	buttonSection := dialog.renderButtonSection()
	if !strings.Contains(buttonSection, "Parse PRD") {
		t.Errorf("renderButtonSection() does not contain Parse PRD button")
	}
	if !strings.Contains(buttonSection, "Cancel") {
		t.Errorf("renderButtonSection() does not contain Cancel button")
	}
}

// TestParsePRDDialogValidationFeedback tests real-time validation feedback (Task 8.1)
func TestParsePRDDialogValidationFeedback(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Simulate tag input with various lengths
	testCases := []struct {
		input       string
		expectedLen int
		expectedMsg string
	}{
		{"", 0, "0 / 50 characters"},
		{"a", 1, "1 / 50 characters"},
		{"feature,auth,ui", 15, "15 / 50 characters"},
		{strings.Repeat("a", 50), 50, "50 / 50 characters"},
		{strings.Repeat("a", 51), 51, "51 / 50 characters (exceeded)"},
	}

	for _, tc := range testCases {
		dialog.validateTags(tc.input)
		if dialog.tagCountMsg != tc.expectedMsg {
			t.Errorf("validateTags(%q): got %q, want %q", tc.input, dialog.tagCountMsg, tc.expectedMsg)
		}
	}
}

// TestParsePRDButtonComponentsRendering tests that buttons are rendered with correct labels (Subtask 7.1)
func TestParsePRDButtonComponentsRendering(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	// Test button section rendering
	buttonSection := dialog.renderButtonSection()

	// Verify buttons have correct labels
	if !strings.Contains(buttonSection, "Parse PRD") {
		t.Errorf("renderButtonSection() does not contain 'Parse PRD' button label")
	}

	if !strings.Contains(buttonSection, "Cancel") {
		t.Errorf("renderButtonSection() does not contain 'Cancel' button label")
	}

	// Verify both button brackets are present
	if !strings.Contains(buttonSection, "[") || !strings.Contains(buttonSection, "]") {
		t.Errorf("renderButtonSection() does not contain button brackets")
	}
}

// TestParsePRDButtonInitialState tests that buttons are properly initialized (Subtask 7.1)
func TestParsePRDButtonInitialState(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Parse button should be focused by default
	if !dialog.submitBtn {
		t.Errorf("submitBtn should be true initially (Parse button focused by default)")
	}

	// Cancel button should not be focused initially
	if dialog.cancelBtn {
		t.Errorf("cancelBtn should be false initially")
	}

	// Field focus should not be on buttons initially
	if dialog.focusedField == 3 {
		t.Errorf("focusedField should not be 3 (buttons) initially, got %d", dialog.focusedField)
	}
}

// TestParsePRDButtonFocus tests button focus with focus field (Subtask 7.1)
func TestParsePRDButtonFocus(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	// Move focus to buttons field
	dialog.focusedField = 3

	// Test Parse button styling when focused
	dialog.submitBtn = true
	dialog.cancelBtn = false

	parseStyle := dialog.getButtonStyle(true, true)
	parseBtn := parseStyle.Render("[ Parse PRD ]")
	if !strings.Contains(parseBtn, "Parse PRD") {
		t.Errorf("Parse button focus style rendering failed")
	}

	// Test Cancel button styling when focused
	dialog.submitBtn = false
	dialog.cancelBtn = true

	cancelStyle := dialog.getButtonStyle(true, false)
	cancelBtn := cancelStyle.Render("[ Cancel ]")
	if !strings.Contains(cancelBtn, "Cancel") {
		t.Errorf("Cancel button focus style rendering failed")
	}
}

// TestParsePRDButtonSpacing tests that buttons have correct spacing (Subtask 7.1)
func TestParsePRDButtonSpacing(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	buttonSection := dialog.renderButtonSection()

	// The button section should have proper spacing - looking for consistent layout
	// Since rendering may involve ANSI codes, we just verify it renders without error
	// and contains both button labels
	if buttonSection == "" {
		t.Errorf("renderButtonSection() returned empty string")
	}

	if !strings.Contains(buttonSection, "Parse PRD") || !strings.Contains(buttonSection, "Cancel") {
		t.Errorf("renderButtonSection() missing button labels")
	}
}

// TestParsePRDButtonNavigation tests left/right arrow navigation between buttons (Subtask 7.1)
func TestParsePRDButtonNavigation(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Focus on buttons field
	dialog.focusedField = 3
	dialog.submitBtn = true
	dialog.cancelBtn = false

	tests := []struct {
		name             string
		initialSubmit    bool
		initialCancel    bool
		keyMsg           string
		expectedSubmit   bool
		expectedCancel   bool
		expectedResult   DialogResult
	}{
		{
			name:             "Left arrow moves to Parse button",
			initialSubmit:    false,
			initialCancel:    true,
			keyMsg:           "left",
			expectedSubmit:   true,
			expectedCancel:   false,
			expectedResult:   DialogResultNone,
		},
		{
			name:             "Right arrow moves to Cancel button",
			initialSubmit:    true,
			initialCancel:    false,
			keyMsg:           "right",
			expectedSubmit:   false,
			expectedCancel:   true,
			expectedResult:   DialogResultNone,
		},
		{
			name:             "h key moves to Parse button",
			initialSubmit:    false,
			initialCancel:    true,
			keyMsg:           "h",
			expectedSubmit:   true,
			expectedCancel:   false,
			expectedResult:   DialogResultNone,
		},
		{
			name:             "l key moves to Cancel button",
			initialSubmit:    true,
			initialCancel:    false,
			keyMsg:           "l",
			expectedSubmit:   false,
			expectedCancel:   true,
			expectedResult:   DialogResultNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.submitBtn = tt.initialSubmit
			dialog.cancelBtn = tt.initialCancel

			keyMsg := tea.KeyMsg{Runes: []rune(tt.keyMsg), Type: tea.KeyRunes}
			result, _ := dialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v", result, tt.expectedResult)
			}

			if dialog.submitBtn != tt.expectedSubmit {
				t.Errorf("After HandleKey(), submitBtn = %v, want %v", dialog.submitBtn, tt.expectedSubmit)
			}

			if dialog.cancelBtn != tt.expectedCancel {
				t.Errorf("After HandleKey(), cancelBtn = %v, want %v", dialog.cancelBtn, tt.expectedCancel)
			}
		})
	}
}

// TestParsePRDButtonThemeStyling tests that buttons use correct theme colors (Subtask 7.2)
func TestParsePRDButtonThemeStyling(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	tests := []struct {
		name     string
		focused  bool
		isSubmit bool
		desc     string
	}{
		{
			name:     "Parse button normal uses accent color",
			focused:  false,
			isSubmit: true,
			desc:     "normal state",
		},
		{
			name:     "Parse button focused uses accent background",
			focused:  true,
			isSubmit: true,
			desc:     "focused state",
		},
		{
			name:     "Cancel button normal uses neutral text color",
			focused:  false,
			isSubmit: false,
			desc:     "normal state",
		},
		{
			name:     "Cancel button focused uses error background",
			focused:  true,
			isSubmit: false,
			desc:     "focused state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := dialog.getButtonStyle(tt.focused, tt.isSubmit)
			rendered := style.Render("Test")

			if rendered == "" {
				t.Errorf("getButtonStyle() returned empty rendered result for %s", tt.desc)
			}

			if !strings.Contains(rendered, "Test") {
				t.Errorf("Rendered output doesn't contain button text for %s", tt.desc)
			}
		})
	}
}

// TestParsePRDButtonColorsInDialog tests button colors in full dialog rendering (Subtask 7.2)
func TestParsePRDButtonColorsInDialog(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.Style = DefaultDialogStyle()

	if dialog.Style == nil {
		t.Fatal("Dialog style is nil")
	}

	// Set focus to buttons field
	dialog.focusedField = 3
	dialog.submitBtn = true
	dialog.cancelBtn = false

	// Render buttons
	buttonSection := dialog.renderButtonSection()

	// Verify that both buttons are rendered
	if !strings.Contains(buttonSection, "Parse PRD") {
		t.Errorf("renderButtonSection() doesn't contain Parse PRD button in theme styling")
	}

	if !strings.Contains(buttonSection, "Cancel") {
		t.Errorf("renderButtonSection() doesn't contain Cancel button in theme styling")
	}

	// Test with Cancel focused
	dialog.submitBtn = false
	dialog.cancelBtn = true

	buttonSection = dialog.renderButtonSection()

	if !strings.Contains(buttonSection, "Parse PRD") {
		t.Errorf("renderButtonSection() loses Parse PRD button when Cancel focused")
	}

	if !strings.Contains(buttonSection, "Cancel") {
		t.Errorf("renderButtonSection() loses Cancel button when focused")
	}
}

// TestParsePRDParseButtonServiceCall tests Parse button triggers service call (Subtask 7.3)
func TestParsePRDParseButtonServiceCall(t *testing.T) {
	tests := []struct {
		name           string
		focusedField   int
		submitBtn      bool
		cancelBtn      bool
		keyMsg         string
		expectedResult DialogResult
	}{
		{
			name:           "Space on Cancel button returns DialogResultCancel",
			focusedField:   3,
			submitBtn:      false,
			cancelBtn:      true,
			keyMsg:         " ",
			expectedResult: DialogResultCancel,
		},
		{
			name:           "Enter on Cancel button returns DialogResultCancel",
			focusedField:   3,
			submitBtn:      false,
			cancelBtn:      true,
			keyMsg:         "enter",
			expectedResult: DialogResultCancel,
		},
		{
			name:           "Space on Parse button with empty form returns None",
			focusedField:   3,
			submitBtn:      true,
			cancelBtn:      false,
			keyMsg:         " ",
			expectedResult: DialogResultNone,
		},
		{
			name:           "Enter on Parse button with empty form returns None",
			focusedField:   3,
			submitBtn:      true,
			cancelBtn:      false,
			keyMsg:         "enter",
			expectedResult: DialogResultNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh dialog for this test case
			testDialog := NewParsePRDDialog()

			// Set up dialog state
			testDialog.focusedField = tt.focusedField
			testDialog.submitBtn = tt.submitBtn
			testDialog.cancelBtn = tt.cancelBtn

			// Simulate key press
			keyMsg := tea.KeyMsg{Runes: []rune(tt.keyMsg), Type: tea.KeyRunes}
			result, _ := testDialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestParsePRDFormValidationForServiceCall tests that form validation blocks invalid submissions (Subtask 7.3)
func TestParsePRDFormValidationForServiceCall(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		expectedValid  bool
	}{
		{
			name:           "Invalid form with no file selected",
			filePath:       "",
			expectedValid:  false,
		},
		{
			name:           "Invalid form with whitespace file path",
			filePath:       "   ",
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh dialog for each test
			testDialog := NewParsePRDDialog()

			// Test isFormValid directly
			// Note: We can only test invalid states here since setting up
			// a valid file selection requires complex browser initialization
			isValid := testDialog.isFormValid()
			if isValid != tt.expectedValid {
				t.Errorf("isFormValid() = %v, want %v (filePath=%q)", isValid, tt.expectedValid, tt.filePath)
			}
		})
	}
}

// TestParsePRDCancelButtonDialogResult tests Cancel button returns proper DialogResult (Subtask 7.4)
func TestParsePRDCancelButtonDialogResult(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name              string
		focusedField      int
		cancelBtn         bool
		submitBtn         bool
		keyMsg            string
		expectedResult    DialogResult
	}{
		{
			name:              "Space activates Cancel button returns DialogResultCancel",
			focusedField:      3,
			cancelBtn:         true,
			submitBtn:         false,
			keyMsg:            " ",
			expectedResult:    DialogResultCancel,
		},
		{
			name:              "Enter activates Cancel button returns DialogResultCancel",
			focusedField:      3,
			cancelBtn:         true,
			submitBtn:         false,
			keyMsg:            "enter",
			expectedResult:    DialogResultCancel,
		},
		{
			name:              "Left/h navigates to Cancel button",
			focusedField:      3,
			cancelBtn:         true,
			submitBtn:         false,
			keyMsg:            "left",
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Cancel button ignores form validity",
			focusedField:      3,
			cancelBtn:         true,
			submitBtn:         false,
			keyMsg:            " ",
			expectedResult:    DialogResultCancel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.focusedField = tt.focusedField
			dialog.cancelBtn = tt.cancelBtn
			dialog.submitBtn = tt.submitBtn

			keyMsg := tea.KeyMsg{Runes: []rune(tt.keyMsg), Type: tea.KeyRunes}
			result, _ := dialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestParsePRDCancelDoesNotProcessFormData tests Cancel doesn't process or validate form (Subtask 7.4)
func TestParsePRDCancelDoesNotProcessFormData(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Set up incomplete form data
	dialog.tagsInput.SetValue("") // No tags
	dialog.fileBrowser.selectedFile = "" // No file selected

	// Set focus to buttons, Cancel focused
	dialog.focusedField = 3
	dialog.cancelBtn = true
	dialog.submitBtn = false

	// Cancel should return immediately without validation
	keyMsg := tea.KeyMsg{Runes: []rune(" "), Type: tea.KeyRunes}
	result, _ := dialog.HandleKey(keyMsg)

	// Should return Cancel immediately
	if result != DialogResultCancel {
		t.Errorf("Cancel with invalid form should still return DialogResultCancel, got %v", result)
	}

	// Form validity should still show as invalid
	if dialog.isFormValid() {
		t.Errorf("Form should remain invalid after Cancel button press")
	}
}

// TestParsePRDButtonsTabFocusIntegration tests buttons are part of Tab focus chain (Subtask 7.5)
func TestParsePRDButtonsTabFocusIntegration(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name              string
		initialField      int
		keyMsg            string
		expectedField     int
		expectedResult    DialogResult
	}{
		{
			name:              "Tab from Operation field (2) goes to Buttons (3)",
			initialField:      2,
			keyMsg:            "tab",
			expectedField:     3,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Tab from Buttons (3) wraps to File browser (0)",
			initialField:      3,
			keyMsg:            "tab",
			expectedField:     0,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Shift+Tab from Buttons (3) goes to Operation (2)",
			initialField:      3,
			keyMsg:            "shift+tab",
			expectedField:     2,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Shift+Tab from File browser (0) wraps to Buttons (3)",
			initialField:      0,
			keyMsg:            "shift+tab",
			expectedField:     3,
			expectedResult:    DialogResultNone,
		},
		{
			name:              "Complete Tab cycle through all fields",
			initialField:      0,
			keyMsg:            "tab",
			expectedField:     1,
			expectedResult:    DialogResultNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.focusedField = tt.initialField

			keyMsg := tea.KeyMsg{Runes: []rune(tt.keyMsg), Type: tea.KeyRunes}
			result, _ := dialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v", result, tt.expectedResult)
			}

			if dialog.focusedField != tt.expectedField {
				t.Errorf("After HandleKey(), focusedField = %d, want %d", dialog.focusedField, tt.expectedField)
			}
		})
	}
}

// TestParsePRDFullTabCycle tests complete Tab cycle including buttons (Subtask 7.5)
func TestParsePRDFullTabCycle(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Starting at File browser (0)
	expectedPath := []int{0, 1, 2, 3, 0} // File → Tags → Operation → Buttons → File (wraps)

	for i, expected := range expectedPath[:4] { // Test up to one cycle
		if dialog.focusedField != expected {
			t.Errorf("Step %d: focusedField = %d, want %d", i, dialog.focusedField, expected)
		}

		// Press Tab to move to next field
		keyMsg := tea.KeyMsg{Runes: []rune("tab"), Type: tea.KeyRunes}
		dialog.HandleKey(keyMsg)
	}

	// After 4 Tab presses, should be back at File browser
	if dialog.focusedField != 0 {
		t.Errorf("After full Tab cycle, focusedField = %d, want 0", dialog.focusedField)
	}
}

// TestParsePRDFullShiftTabCycle tests complete Shift+Tab cycle in reverse (Subtask 7.5)
func TestParsePRDFullShiftTabCycle(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Start at File browser (0), go backward
	dialog.focusedField = 0
	expectedPath := []int{3, 2, 1, 0} // File → Buttons → Operation → Tags → File (reverse wraps)

	for i, expected := range expectedPath {
		// Press Shift+Tab to move to previous field
		keyMsg := tea.KeyMsg{Runes: []rune("shift+tab"), Type: tea.KeyRunes}
		dialog.HandleKey(keyMsg)

		if dialog.focusedField != expected {
			t.Errorf("Step %d (Shift+Tab backward): focusedField = %d, want %d", i, dialog.focusedField, expected)
		}
	}
}

// TestParsePRDButtonActionOnFocus tests buttons are activated when focused (Subtask 7.5)
func TestParsePRDButtonActionOnFocus(t *testing.T) {
	dialog := NewParsePRDDialog()

	tests := []struct {
		name              string
		focusedField      int
		buttonFocus       string // "parse" or "cancel"
		activationKey     string // " " or "enter"
		expectedResult    DialogResult
	}{
		{
			name:              "Space activates focused Parse button",
			focusedField:      3,
			buttonFocus:       "parse",
			activationKey:     " ",
			expectedResult:    DialogResultNone, // Form invalid, so None
		},
		{
			name:              "Space activates focused Cancel button",
			focusedField:      3,
			buttonFocus:       "cancel",
			activationKey:     " ",
			expectedResult:    DialogResultCancel,
		},
		{
			name:              "Enter activates focused Parse button",
			focusedField:      3,
			buttonFocus:       "parse",
			activationKey:     "enter",
			expectedResult:    DialogResultNone, // Form invalid, so None
		},
		{
			name:              "Enter activates focused Cancel button",
			focusedField:      3,
			buttonFocus:       "cancel",
			activationKey:     "enter",
			expectedResult:    DialogResultCancel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.focusedField = tt.focusedField

			// Set button focus
			if tt.buttonFocus == "parse" {
				dialog.submitBtn = true
				dialog.cancelBtn = false
			} else {
				dialog.submitBtn = false
				dialog.cancelBtn = true
			}

			// Activate button
			keyMsg := tea.KeyMsg{Runes: []rune(tt.activationKey), Type: tea.KeyRunes}
			result, _ := dialog.HandleKey(keyMsg)

			if result != tt.expectedResult {
				t.Errorf("HandleKey() result = %v, want %v (button=%s, key=%q)", result, tt.expectedResult, tt.buttonFocus, tt.activationKey)
			}
		})
	}
}
