package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorDialogModel represents a simple error dialog
type ErrorDialogModel struct {
	title   string
	message string
	width   int
	height  int
	done    bool
}

// NewErrorDialogModel creates a new error dialog
func NewErrorDialogModel(title, message string, width, height int) *ErrorDialogModel {
	return &ErrorDialogModel{
		title:   title,
		message: message,
		width:   width,
		height:  height,
	}
}

// Init initializes the dialog
func (m ErrorDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles user interaction
func (m *ErrorDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "q":
			m.done = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the dialog
func (m ErrorDialogModel) View() string {
	dialogWidth := m.width / 2

	// Style for the title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F7768E")).
		Bold(true).
		PaddingBottom(1)

	// Style for the message
	messageStyle := lipgloss.NewStyle().
		Width(dialogWidth - 10).
		Align(lipgloss.Center).
		PaddingBottom(2)

	// Button style
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#A9B1D6")).
		Bold(true)

	// Render elements
	title := titleStyle.Render(m.title)
	message := messageStyle.Render(m.message)
	button := buttonStyle.Render("OK")

	// Center the button
	buttonRowStyle := lipgloss.NewStyle().
		Width(dialogWidth - 10).
		Align(lipgloss.Center)

	buttonRow := buttonRowStyle.Render(button)

	// Combine elements
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		message,
		buttonRow,
		"",
		"Press Enter or Esc to close",
	)

	// Add border
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F7768E")).
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
func (m ErrorDialogModel) Done() bool {
	return m.done
}
