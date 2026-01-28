package dialog

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ParsePRDOperationMode represents the mode for parsing PRD files
type ParsePRDOperationMode int

const (
	// ParsePRDReplace replaces all existing tasks with new ones from PRD
	ParsePRDReplace ParsePRDOperationMode = iota
	// ParsePRDAppend appends new tasks to existing ones
	ParsePRDAppend
)

// ParsePRDResultMsg is sent when the ParsePRD dialog completes successfully
type ParsePRDResultMsg struct {
	FilePath string
	Err      error
}

// ParsePRDDialog implements a dialog for parsing PRD files following State-Only Update Pattern
type ParsePRDDialog struct {
	BaseFocusableDialog

	// File browser for PRD file selection (Task 3.3)
	fileBrowser *LogFileBrowserModel
	
	// Form fields
	tagsInput textinput.Model

	// Dialog-specific state
	focusedField  int // Which form field has focus (0=fileBrowser, 1=tags, 2=operationMode, 3=buttons)
	operationMode ParsePRDOperationMode
	submitBtn     bool
	cancelBtn     bool

	// Validation state (Task 8.1)
	fileValid      bool   // Whether the selected file exists
	fileValidMsg   string // Validation message for file
	tagValid       bool   // Whether tags are within 50-char limit
	tagCountMsg    string // Character count feedback for tags

	// Fixed dimensions
	fixedWidth  int
	fixedHeight int

	// Padding
	paddingLeft   int
	paddingRight  int
	paddingTop    int
	paddingBottom int

	// Input area width (70 - 2*padding_left - 2*border)
	inputWidth int
}

// NewParsePRDDialog creates a new ParsePRD dialog with fixed dimensions
func NewParsePRDDialog() *ParsePRDDialog {
	// Create the file browser for PRD file selection (Task 3.3)
	// Initialize with a dummy taskService (we only need the browser for file selection)
	fileBrowser := NewLogFileBrowserModel(54, 10, nil, "")
	
	// Configure the browser for PRD files
	fileBrowser.SetRootPath(".taskmaster/docs")              // Look in .taskmaster/docs
	fileBrowser.SetFileExtensions([]string{".txt", ".md"})   // Only show .txt and .md files

	tagsInput := textinput.New()
	tagsInput.Placeholder = "(optional) comma-separated tags"
	tagsInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6D98BA"))
	tagsInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))
	tagsInput.CharLimit = 50 // Max 50 chars for tags

	dialog := &ParsePRDDialog{
		BaseFocusableDialog: NewBaseFocusableDialog("Parse PRD", 70, 25, DialogKindForm, 4),
		fileBrowser:         fileBrowser,
		tagsInput:           tagsInput,
		fixedWidth:          70,
		fixedHeight:         25,
		paddingLeft:         2,
		paddingRight:        2,
		paddingTop:          1,
		paddingBottom:       1,
		inputWidth:          54, // 70 - 2*2 (padding) - 2 (border)
		focusedField:        0,
		operationMode:       ParsePRDReplace, // Default to Replace mode
		submitBtn:           true,            // Parse button focused by default when on buttons field
		cancelBtn:           false,           // Cancel button not focused initially
		tagValid:            true,            // Tags always valid initially (optional field)
		fileValid:           false,           // No file selected initially
		fileValidMsg:        "No file selected",
		tagCountMsg:         "0 / 50 characters",
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "Tab", Label: "Next"},
		ShortcutHint{Key: "Shift+Tab", Label: "Prev"},
		ShortcutHint{Key: "Enter", Label: "Submit"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	// Focus the first field (file browser)
	dialog.fileBrowser.SetFocused(true)

	return dialog
}

// Init initializes the dialog
func (d *ParsePRDDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes messages and updates dialog state
// Implements State-Only Update Pattern: only handles WindowSizeMsg and custom result messages
func (d *ParsePRDDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Center dialog on window resize, maintain fixed size
		d.Center(msg.Width, msg.Height)
		return d, nil

	case ParsePRDResultMsg:
		// Handle result messages from async operations
		return d, nil
	}

	// Return unchanged - no key handling in state-only pattern
	return d, nil
}

// View renders the dialog
func (d *ParsePRDDialog) View() string {
	content := d.renderContent()
	styledContent := d.applyPadding(content)
	return d.RenderBorder(styledContent)
}

// renderContent renders the main dialog content
func (d *ParsePRDDialog) renderContent() string {
	var sections []string

	// File path section
	sections = append(sections, d.renderFilePathSection())

	// Tags section
	sections = append(sections, d.renderTagsSection())

	// Operation mode section
	sections = append(sections, d.renderOperationModeSection())

	// Button section
	sections = append(sections, d.renderButtonSection())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sections...,
	)
}

// renderFilePathSection renders the file path input section
func (d *ParsePRDDialog) renderFilePathSection() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor).
		Bold(true)

	label := labelStyle.Render("PRD File:")

	// Set focus on the browser based on focusedField (Task 3.3)
	d.fileBrowser.SetFocused(d.focusedField == 0)
	
	// Render the file browser
	browserView := d.fileBrowser.View()

	// Add validation feedback (Task 8.1)
	validationStyle := d.getValidationStyle(d.fileValid)
	validationMsg := validationStyle.Render(d.fileValidMsg)

	return lipgloss.JoinVertical(lipgloss.Left, label, browserView, validationMsg)
}

// renderSummarySection renders the tags input section
func (d *ParsePRDDialog) renderTagsSection() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor).
		Bold(true).
		PaddingTop(1)

	label := labelStyle.Render("Tags (optional):")

	// Apply focus styling to input
	inputStyle := d.getInputStyle(d.focusedField == 1)
	input := inputStyle.Render(d.tagsInput.View())

	// Add character count feedback (Task 8.1)
	countStyle := lipgloss.NewStyle().
		PaddingTop(0).
		Foreground(d.Style.TextColor)
	
	countMsg := countStyle.Render(d.tagCountMsg)

	return lipgloss.JoinVertical(lipgloss.Left, label, input, countMsg)
}

// renderOperationModeSection renders the operation mode radio buttons
func (d *ParsePRDDialog) renderOperationModeSection() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor).
		Bold(true).
		PaddingTop(1)

	label := labelStyle.Render("Operation Mode:")

	// Create radio button styles
	replaceSelected := d.operationMode == ParsePRDReplace
	appendSelected := d.operationMode == ParsePRDAppend

	replaceRadio := d.renderRadioButton(replaceSelected, "Replace all existing tasks")
	appendRadio := d.renderRadioButton(appendSelected, "Append to existing tasks")

	// Combine radio buttons with spacing
	options := lipgloss.JoinVertical(
		lipgloss.Left,
		replaceRadio,
		appendRadio,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		label,
		options,
	)
}

// renderRadioButton renders a single radio button option
func (d *ParsePRDDialog) renderRadioButton(selected bool, label string) string {
	var radioSymbol string
	var symbolStyle lipgloss.Style

	if selected {
		// Use filled circle with accent color for selected
		radioSymbol = "◉"
		symbolStyle = lipgloss.NewStyle().
			Foreground(d.Style.ButtonColor)
	} else {
		// Use hollow circle for unselected
		radioSymbol = "○"
		symbolStyle = lipgloss.NewStyle().
			Foreground(d.Style.TextColor)
	}

	renderedSymbol := symbolStyle.Render(radioSymbol)

	textStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedSymbol,
		" ",
		textStyle.Render(label),
	)
}

// renderButtonSection renders the Parse PRD and Cancel buttons
// Positioned at bottom right with proper alignment and 2-space gap
func (d *ParsePRDDialog) renderButtonSection() string {
	// Get button states based on focus and selection
	parseStyle := d.getButtonStyle(d.focusedField == 3 && d.submitBtn, true)
	cancelStyle := d.getButtonStyle(d.focusedField == 3 && d.cancelBtn, false)

	// Render button labels
	parseBtn := parseStyle.Render("[ Parse PRD ]")
	cancelBtn := cancelStyle.Render("[ Cancel ]")

	// Create horizontal layout with 2-space gap between buttons
	buttonContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		parseBtn,
		"  ", // 2-space gap
		cancelBtn,
	)

	// Right-align buttons using container with width and alignment
	alignedButtons := lipgloss.NewStyle().
		Width(d.inputWidth).
		Align(lipgloss.Right).
		Render(buttonContent)

	return lipgloss.NewStyle().
		PaddingTop(1).
		Render(alignedButtons)
}

// getInputStyle returns the appropriate style for an input field
func (d *ParsePRDDialog) getInputStyle(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(d.inputWidth).
		Foreground(d.Style.TextColor)

	if focused {
		style = style.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.Style.FocusedBorderColor)
	}

	return style
}

// getButtonStyle returns the appropriate style for a button
// Parse button uses accent color, Cancel uses neutral/error colors
func (d *ParsePRDDialog) getButtonStyle(focused bool, isSubmit bool) lipgloss.Style {
	if isSubmit {
		// Parse button: use theme accent color
		style := lipgloss.NewStyle().
			Foreground(d.Style.ButtonColor) // Accent color in normal state

		if focused {
			// Focused Parse button: accent background with contrasting foreground
			style = style.
				Background(d.Style.ButtonColor).
				Foreground(d.Style.BackgroundColor)
		}
		return style
	} else {
		// Cancel button: use neutral colors
		style := lipgloss.NewStyle().
			Foreground(d.Style.TextColor) // Neutral text color in normal state

		if focused {
			// Focused Cancel button: error color (red) background with contrasting foreground
			style = style.
				Background(d.Style.ErrorColor).
				Foreground(d.Style.BackgroundColor)
		}
		return style
	}
}

// getValidationStyle returns the appropriate style for validation feedback (Task 8.1)
func (d *ParsePRDDialog) getValidationStyle(isValid bool) lipgloss.Style {
	style := lipgloss.NewStyle().PaddingTop(0)
	
	if isValid {
		// Success: green border and text
		style = style.
			Foreground(d.Style.SuccessColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.Style.SuccessColor)
	} else {
		// Warning: red border and text
		style = style.
			Foreground(d.Style.ErrorColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.Style.ErrorColor)
	}
	
	return style
}

// applyPadding applies the configured padding to the content
func (d *ParsePRDDialog) applyPadding(content string) string {
	style := lipgloss.NewStyle().
		PaddingLeft(d.paddingLeft).
		PaddingRight(d.paddingRight).
		PaddingTop(d.paddingTop).
		PaddingBottom(d.paddingBottom)

	return style.Render(content)
}

// GetFilePath returns the selected file path from the browser (Task 3.3)
func (d *ParsePRDDialog) GetFilePath() string {
	selectedFile := d.fileBrowser.GetSelectedFile()
	return strings.TrimSpace(selectedFile)
}

// GetTags returns the entered tags
func (d *ParsePRDDialog) GetTags() string {
	return strings.TrimSpace(d.tagsInput.Value())
}

// GetOperationMode returns the current operation mode
func (d *ParsePRDDialog) GetOperationMode() ParsePRDOperationMode {
	return d.operationMode
}

// GetAppendFlag returns whether the --append flag should be used
func (d *ParsePRDDialog) GetAppendFlag() bool {
	return d.operationMode == ParsePRDAppend
}

// isFormValid checks if the form can be submitted
func (d *ParsePRDDialog) isFormValid() bool {
	// File must be selected
	if d.GetFilePath() == "" {
		return false
	}
	// Tags are optional, so always valid
	return true
}

// Validation methods (Task 8.1)
// validateFilePath validates the selected file path (Task 8.1)
func (d *ParsePRDDialog) validateFilePath(filePath string) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		d.fileValid = false
		d.fileValidMsg = "No file selected"
		return
	}

	// Check if file exists using os.Stat
	if _, err := os.Stat(filePath); err == nil {
		d.fileValid = true
		d.fileValidMsg = "File found ✓"
	} else {
		d.fileValid = false
		d.fileValidMsg = "File not found"
	}
}

// validateTags validates the tags input (Task 8.1)
func (d *ParsePRDDialog) validateTags(tagInput string) {
	tagLen := len(tagInput)
	if tagLen > 50 {
		d.tagValid = false
		d.tagCountMsg = strconv.Itoa(tagLen) + " / 50 characters (exceeded)"
	} else {
		d.tagValid = true
		d.tagCountMsg = strconv.Itoa(tagLen) + " / 50 characters"
	}
}

// updateValidation updates all validation states (Task 8.1)
func (d *ParsePRDDialog) updateValidation() {
	d.validateFilePath(d.GetFilePath())
	d.validateTags(d.tagsInput.Value())
}

// HandleKey processes keyboard input with centralized field navigation
// Implements Pattern 1: ALL keyboard logic in HandleKey, nothing in Update
func (d *ParsePRDDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Call d.HandleBaseKey() first for standard dialog keys (Escape to cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	keyStr := msg.String()

	// Handle field cycling with Tab/Shift+Tab
	// Field order: 0=fileBrowser, 1=tags, 2=operationMode, 3=buttons
	switch keyStr {
	case "tab":
		// Move to next field with wrapping
		d.focusedField = (d.focusedField + 1) % 4
		// Blur text input when leaving tags field
		if d.focusedField != 1 {
			d.tagsInput.Blur()
		}
		return DialogResultNone, nil
	case "shift+tab":
		// Move to previous field with wrapping
		d.focusedField = (d.focusedField - 1 + 4) % 4
		// Blur text input when leaving tags field
		if d.focusedField != 1 {
			d.tagsInput.Blur()
		}
		return DialogResultNone, nil
	}

	// Handle Enter key - submit form if valid (except when buttons are focused)
	// When buttons are focused, let the button handler deal with Enter
	if keyStr == "enter" && d.focusedField != 3 {
		if d.isFormValid() {
			return DialogResultConfirm, nil
		}
		// Form is not valid, stay in dialog and return no result
		return DialogResultNone, nil
	}

	// Handle field-specific key input
	switch d.focusedField {
	case 0:
		// File browser field - delegate arrow keys and other navigation to browser
		return d.handleFileBrowserKey(msg)

	case 1:
		// Tags input field - handle Ctrl+A/C/V shortcuts and text input
		return d.handleTagsInputKey(msg)

	case 2:
		// Operation mode field - handle arrow keys for radio navigation
		return d.handleOperationModeKey(msg)

	case 3:
		// Buttons field - handle left/right arrows for button selection and space/enter for activation
		return d.handleButtonsKey(msg)
	}

	return DialogResultNone, nil
}

// handleFileBrowserKey processes keys when file browser is focused
func (d *ParsePRDDialog) handleFileBrowserKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	d.fileBrowser.SetFocused(true)
	// Delegate to file browser for its own key handling
	d.fileBrowser.Update(msg)
	// Update validation after file browser changes
	d.updateValidation()
	return DialogResultNone, nil
}

// handleTagsInputKey processes keys when tags input is focused
func (d *ParsePRDDialog) handleTagsInputKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	keyStr := msg.String()

	// Ctrl+A: Select all
	if keyStr == "ctrl+a" {
		d.tagsInput.SetCursor(len(d.tagsInput.Value()))
		return DialogResultNone, nil
	}

	// Ctrl+C: Copy (would require clipboard integration - stub for now)
	if keyStr == "ctrl+c" {
		// Clipboard handling would go here
		return DialogResultNone, nil
	}

	// Ctrl+V: Paste (would require clipboard integration - stub for now)
	if keyStr == "ctrl+v" {
		// Clipboard handling would go here
		return DialogResultNone, nil
	}

	// All other keys go to textinput for standard text editing
	d.tagsInput.Focus()
	updatedInput, _ := d.tagsInput.Update(msg)
	d.tagsInput = updatedInput

	// Update tag count message
	tagLen := len(d.tagsInput.Value())
	d.tagCountMsg = fmt.Sprintf("%d / 50 characters", tagLen)

	return DialogResultNone, nil
}

// handleOperationModeKey processes keys when operation mode field is focused
func (d *ParsePRDDialog) handleOperationModeKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "up", "k":
		// Move to previous option with wraparound
		if d.operationMode == ParsePRDAppend {
			d.operationMode = ParsePRDReplace
		}
		return DialogResultNone, nil

	case "down", "j":
		// Move to next option with wraparound
		if d.operationMode == ParsePRDReplace {
			d.operationMode = ParsePRDAppend
		}
		return DialogResultNone, nil

	case " ":
		// Space: Toggle selection
		if d.operationMode == ParsePRDReplace {
			d.operationMode = ParsePRDAppend
		} else {
			d.operationMode = ParsePRDReplace
		}
		return DialogResultNone, nil
	}

	return DialogResultNone, nil
}

// handleButtonsKey processes keys when buttons field is focused
func (d *ParsePRDDialog) handleButtonsKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "left", "h":
		// Move to submit button
		d.submitBtn = true
		d.cancelBtn = false
		return DialogResultNone, nil

	case "right", "l":
		// Move to cancel button
		d.submitBtn = false
		d.cancelBtn = true
		return DialogResultNone, nil

	case " ", "enter":
		// Activate button
		if d.submitBtn {
			if d.isFormValid() {
				return DialogResultConfirm, nil
			}
			// Form not valid, stay in dialog
			return DialogResultNone, nil
		}
		if d.cancelBtn {
			return DialogResultCancel, nil
		}
		// Default to submit if neither explicitly set
		if d.isFormValid() {
			return DialogResultConfirm, nil
		}
		return DialogResultNone, nil
	}

	return DialogResultNone, nil
}
