package dialog

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationResult represents the result of a confirmation
type ConfirmationResult int

const (
	// ConfirmationResultNone indicates no result yet
	ConfirmationResultNone ConfirmationResult = iota
	// ConfirmationResultYes indicates "Yes" was selected
	ConfirmationResultYes
	// ConfirmationResultNo indicates "No" was selected
	ConfirmationResultNo
)

// ConfirmationMsg is sent when a confirmation is made
type ConfirmationMsg struct {
	Result ConfirmationResult
}

// ConfirmationDialog is a dialog that presents a yes/no choice
type ConfirmationDialog struct {
	BaseFocusableDialog
	message    string
	yesText    string
	noText     string
	yesDefault bool
	dangerMode bool
	result     ConfirmationResult
}

// NewConfirmationDialog creates a new confirmation dialog
func NewConfirmationDialog(title string, message string, width, height int) *ConfirmationDialog {
	dialog := &ConfirmationDialog{
		BaseFocusableDialog: NewBaseFocusableDialog(title, width, height, DialogKindConfirmation, 2),
		message:             message,
		yesText:             "Yes",
		noText:              "No",
		yesDefault:          true,
		dangerMode:          false,
		result:              ConfirmationResultNone,
	}
	dialog.SetFooterHints(
		ShortcutHint{Key: "←/→", Label: "Change Selection"},
		ShortcutHint{Key: "Enter", Label: "Confirm"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)
	return dialog
}

// NewErrorDialog returns a confirmation-styled dialog for error scenarios.
func NewErrorDialog(title, message string) *ConfirmationDialog {
	dialog := NewConfirmationDialog(title, message, 64, 10)
	dialog.SetYesText("Dismiss")
	dialog.SetNoText("Cancel")
	dialog.SetDangerMode(true)
	if dialog.Style != nil {
		dialog.Style.BorderColor = dialog.Style.ErrorColor
		dialog.Style.FocusedBorderColor = dialog.Style.ErrorColor
		dialog.Style.TitleColor = dialog.Style.ErrorColor
	}
	dialog.SetFooterHints(
		ShortcutHint{Key: "Enter", Label: "Dismiss"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)
	return dialog
}

// SetYesText sets the text for the "Yes" button
func (d *ConfirmationDialog) SetYesText(text string) {
	d.yesText = text
}

// SetNoText sets the text for the "No" button
func (d *ConfirmationDialog) SetNoText(text string) {
	d.noText = text
}

// SetYesDefault sets whether "Yes" is the default option
func (d *ConfirmationDialog) SetYesDefault(isDefault bool) {
	d.yesDefault = isDefault
	if isDefault {
		d.SetFocusedIndex(0)
	} else {
		d.SetFocusedIndex(1)
	}
}

// SetDangerMode sets whether the dialog is in danger mode (red Yes button)
func (d *ConfirmationDialog) SetDangerMode(dangerMode bool) {
	d.dangerMode = dangerMode
}

// Init initializes the dialog
func (d *ConfirmationDialog) Init() tea.Cmd {
	// Set the focused index based on default
	if d.yesDefault {
		d.SetFocusedIndex(0)
	} else {
		d.SetFocusedIndex(1)
	}
	return nil
}

// Update processes messages and updates dialog state
func (d *ConfirmationDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
	}

	return d, nil
}

// View renders the dialog
func (d *ConfirmationDialog) View() string {
	// Account for border and padding
	contentWidth := d.width - 4

	if contentWidth < 1 {
		contentWidth = 1
	}

	// Render message
	messageStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		PaddingTop(1).
		PaddingBottom(2)

	message := messageStyle.Render(d.message)

	// Render buttons
	buttonRow := d.renderButtons()

	// Combine everything
	content := lipgloss.JoinVertical(lipgloss.Left, message, buttonRow)

	// Add border and title
	return d.RenderBorder(content)
}

// renderButtons renders the buttons
func (d *ConfirmationDialog) renderButtons() string {
	// Yes button
	yesStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(d.FocusedIndex() == 0)

	if d.dangerMode {
		yesStyle = yesStyle.Foreground(d.Style.ErrorColor)
	} else {
		yesStyle = yesStyle.Foreground(d.Style.ButtonColor)
	}

	if d.FocusedIndex() == 0 {
		yesStyle = yesStyle.Underline(true)
	}

	// No button
	noStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(d.Style.ButtonColor).
		Bold(d.FocusedIndex() == 1)

	if d.FocusedIndex() == 1 {
		noStyle = noStyle.Underline(true)
	}

	// Add focus indicators
	yesText := d.yesText
	noText := d.noText

	// Render buttons
	yesBtn := yesStyle.Render(yesText)
	noBtn := noStyle.Render(noText)

	// Join buttons
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesBtn, noBtn)

	// Center the buttons
	style := lipgloss.NewStyle().
		Width(d.width - 4).
		Align(lipgloss.Center)

	return style.Render(buttons)
}

// HandleKey processes a key event
func (d *ConfirmationDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base focusable dialog keys (like Tab/Shift+Tab)
	result, cmd := d.HandleBaseFocusableKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle confirmation-specific keys
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
		if d.FocusedIndex() == 1 {
			return DialogResultNone, d.SetFocusedIndex(0)
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
		if d.FocusedIndex() == 0 {
			return DialogResultNone, d.SetFocusedIndex(1)
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " ", "space"))):
		if d.FocusedIndex() == 0 {
			d.result = ConfirmationResultYes
			return DialogResultConfirm, func() tea.Msg {
				return ConfirmationMsg{Result: ConfirmationResultYes}
			}
		} else {
			d.result = ConfirmationResultNo
			return DialogResultCancel, func() tea.Msg {
				return ConfirmationMsg{Result: ConfirmationResultNo}
			}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
		d.result = ConfirmationResultYes
		return DialogResultConfirm, func() tea.Msg {
			return ConfirmationMsg{Result: ConfirmationResultYes}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
		d.result = ConfirmationResultNo
		return DialogResultCancel, func() tea.Msg {
			return ConfirmationMsg{Result: ConfirmationResultNo}
		}
	}

	return DialogResultNone, nil
}

// Result returns the confirmation result
func (d *ConfirmationDialog) Result() ConfirmationResult {
	return d.result
}

// YesNo creates a Yes/No confirmation dialog
func YesNo(title, message string, dangerMode bool) *ConfirmationDialog {
	// Calculate width based on message length
	width := len(message) + 10
	if width < 30 {
		width = 30
	}
	if width > 80 {
		width = 80
	}

	dialog := NewConfirmationDialog(title, message, width, 7)
	dialog.SetDangerMode(dangerMode)
	return dialog
}
