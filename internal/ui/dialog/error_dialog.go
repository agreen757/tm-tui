package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorDialogModel represents an enhanced error dialog with recovery hints
type ErrorDialogModel struct {
	BaseFocusableDialog
	message       string
	details       string
	recoveryHints string
	result        ConfirmationResult
}

// NewErrorDialogModel creates a new error dialog with standard layout
func NewErrorDialogModel(title, message string) *ErrorDialogModel {
	dialog := &ErrorDialogModel{
		BaseFocusableDialog: NewBaseFocusableDialog(title, 70, 15, DialogKindCustom, 1),
		message:             message,
		details:             "",
		recoveryHints:       "",
		result:              ConfirmationResultNone,
	}
	if dialog.Style != nil {
		dialog.Style.BorderColor = dialog.Style.ErrorColor
		dialog.Style.FocusedBorderColor = dialog.Style.ErrorColor
		dialog.Style.TitleColor = dialog.Style.ErrorColor
	}
	dialog.SetFooterHints(
		ShortcutHint{Key: "Enter", Label: "Dismiss"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)
	return dialog
}

// SetDetails sets the detailed error information
func (d *ErrorDialogModel) SetDetails(details string) {
	d.details = details
}

// SetRecoveryHints sets recovery hint text
func (d *ErrorDialogModel) SetRecoveryHints(hints string) {
	d.recoveryHints = hints
}

// Init initializes the dialog
func (d *ErrorDialogModel) Init() tea.Cmd {
	return nil
}

// Update processes messages
func (d *ErrorDialogModel) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
		// Adjust height based on content
		lines := d.calculateContentHeight()
		if lines > d.height {
			d.height = lines + 2 // Add padding
			if d.height > msg.Height-4 {
				d.height = msg.Height - 4
			}
			d.Center(msg.Width, msg.Height)
		}
	}
	return d, nil
}

// View renders the error dialog
func (d *ErrorDialogModel) View() string {
	contentWidth := d.width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	// Render message (required)
	messageStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Left).
		PaddingTop(1).
		PaddingBottom(1)
	message := messageStyle.Render(d.message)

	// Render details if available
	var content string
	if d.details != "" {
		detailStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Foreground(lipgloss.Color("8")).
			Align(lipgloss.Left).
			PaddingBottom(1)
		details := detailStyle.Render("Details: " + d.details)
		content = lipgloss.JoinVertical(lipgloss.Left, message, details)
	} else {
		content = message
	}

	// Render recovery hints if available
	if d.recoveryHints != "" {
		hintStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Foreground(lipgloss.Color("11")).
			Align(lipgloss.Left).
			PaddingBottom(1)
		hints := hintStyle.Render("Recovery:\n" + d.recoveryHints)
		content = lipgloss.JoinVertical(lipgloss.Left, content, hints)
	}

	// Render dismiss button
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("1")).
		Bold(true)
	button := buttonStyle.Render("[ Dismiss ]")

	buttonRowStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		PaddingTop(1)
	buttonRow := buttonRowStyle.Render(button)

	content = lipgloss.JoinVertical(lipgloss.Left, content, buttonRow)

	// Add border
	return d.RenderBorder(content)
}

// HandleKey processes key events
func (d *ErrorDialogModel) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc", "q":
		return DialogResultClose, nil
	}
	return DialogResultNone, nil
}

// calculateContentHeight calculates how many lines the content will take
func (d *ErrorDialogModel) calculateContentHeight() int {
	contentWidth := d.width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	lines := 1 // Message
	if d.details != "" {
		lines += len(strings.Split(d.details, "\n"))
		lines += 1
	}
	if d.recoveryHints != "" {
		lines += len(strings.Split(d.recoveryHints, "\n"))
		lines += 2
	}
	lines += 2 // Button and padding

	return lines
}
