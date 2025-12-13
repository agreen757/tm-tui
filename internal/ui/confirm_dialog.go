package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialogModel represents a simple confirmation dialog
type ConfirmDialogModel struct {
	message     string
	confirmText string
	cancelText  string
	confirmed   bool
	width       int
	height      int
	done        bool
}

// NewConfirmDialogModel creates a new confirmation dialog
func NewConfirmDialogModel(message, confirmText, cancelText string, width, height int) *ConfirmDialogModel {
	return &ConfirmDialogModel{
		message:     message,
		confirmText: confirmText,
		cancelText:  cancelText,
		width:       width,
		height:      height,
	}
}

// Init initializes the dialog
func (m ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles user interaction
func (m *ConfirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.confirmed = false
			m.done = true
			return m, tea.Quit

		case "enter":
			m.confirmed = true
			m.done = true
			return m, tea.Quit

		case "tab", "right", "l":
			m.confirmed = !m.confirmed

		case "left", "h":
			m.confirmed = !m.confirmed
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the dialog
func (m ConfirmDialogModel) View() string {
	dialogWidth := m.width / 2

	// Style for the message
	messageStyle := lipgloss.NewStyle().
		Width(dialogWidth - 10).
		Align(lipgloss.Center).
		PaddingTop(1).
		PaddingBottom(2)

	// Button styles
	confirmStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#F7768E")).
		Bold(m.confirmed).
		Underline(m.confirmed)

	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#A9B1D6")).
		Bold(!m.confirmed).
		Underline(!m.confirmed)

	// Render buttons
	confirmBtn := confirmStyle.Render(m.confirmText)
	cancelBtn := cancelStyle.Render(m.cancelText)

	// Join buttons with padding
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, confirmBtn, "   ", cancelBtn)

	// Center the buttons
	buttonStyle := lipgloss.NewStyle().
		Width(dialogWidth - 10).
		Align(lipgloss.Center)

	buttons := buttonStyle.Render(buttonRow)

	// Combine message and buttons
	content := lipgloss.JoinVertical(lipgloss.Center,
		messageStyle.Render(m.message),
		buttons,
		"",
		"↑/↓/tab: navigate • enter: confirm • esc: cancel",
	)

	// Add border
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BB9AF7")).
		Padding(1, 2).
		Width(dialogWidth).
		Align(lipgloss.Center).
		Render(content)

	// Center in available space
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialog,
	)
}

// Done returns whether the dialog is done
func (m ConfirmDialogModel) Done() bool {
	return m.done
}

// Result returns the confirmation result
func (m ConfirmDialogModel) Result() bool {
	return m.confirmed
}
