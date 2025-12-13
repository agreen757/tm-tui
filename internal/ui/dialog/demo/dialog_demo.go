package demo

import (
	"fmt"
	"os"
	"time"

	"github.com/adriangreen/tm-tui/internal/ui/dialog"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap defines the keybindings for the demo
type KeyMap struct {
	Modal        key.Binding
	Form         key.Binding
	List         key.Binding
	Confirmation key.Binding
	Progress     key.Binding
	Nested       key.Binding
	Help         key.Binding
	Quit         key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Modal: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "Modal Dialog"),
		),
		Form: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "Form Dialog"),
		),
		List: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "List Dialog"),
		),
		Confirmation: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "Confirmation Dialog"),
		),
		Progress: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "Progress Dialog"),
		),
		Nested: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "Nested Dialogs"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "Toggle Help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q", "Quit"),
		),
	}
}

// Model holds the state for the dialog demo
type Model struct {
	width         int
	height        int
	keyMap        KeyMap
	dialogManager *dialog.DialogManager
	help          help.Model
	showHelp      bool
	lastResult    string
	progressVal   float64
}

// New creates a new dialog demo model
func New() Model {
	helpModel := help.New()
	helpModel.ShowAll = true

	// Initialize the dialog manager
	manager := dialog.NewDialogManager(0, 0)

	return Model{
		keyMap:        DefaultKeyMap(),
		dialogManager: manager,
		help:          helpModel,
		showHelp:      false,
		lastResult:    "",
		progressVal:   0.0,
	}
}

// Init initializes the demo
func (m Model) Init() tea.Cmd {
	// No initialization needed
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update dialog manager dimensions
		m.dialogManager.SetTerminalSize(msg.Width, msg.Height)

		return m, nil

	case tea.KeyMsg:
		// Check if dialogs are active, if so, let the dialog manager handle the message
		if m.dialogManager.HasDialogs() {
			cmd := m.dialogManager.HandleMsg(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Otherwise, handle keys for the main interface
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keyMap.Modal):
			// Show a modal dialog
			content := dialog.NewSimpleModalContent("This is a simple modal dialog.\n\nPress Enter to close or Esc to cancel.")
			modal := dialog.NewModalDialog("Modal Dialog", 60, 10, content)
			m.dialogManager.PushDialog(modal)
			return m, nil

		case key.Matches(msg, m.keyMap.Form):
			// Show a form dialog
			fields := []dialog.FormField{
				dialog.NewTextField("Name", "Enter your name", true),
				dialog.NewTextField("Email", "Enter your email", false),
				dialog.NewCheckboxField("Subscribe", false),
				dialog.NewRadioGroupField("Plan", []string{"Free", "Basic", "Premium"}, 0),
			}
			form := dialog.NewLegacyFormDialog("Form Dialog", 70, 20, fields)
			m.dialogManager.PushDialog(form)
			return m, nil

		case key.Matches(msg, m.keyMap.List):
			// Show a list dialog
			items := []dialog.ListItem{
				dialog.NewSimpleListItem("Option 1", "Description for option 1"),
				dialog.NewSimpleListItem("Option 2", "Description for option 2"),
				dialog.NewSimpleListItem("Option 3", "Description for option 3"),
				dialog.NewSimpleListItem("Option 4", "Description for option 4"),
				dialog.NewSimpleListItem("Option 5", "Description for option 5"),
			}
			list := dialog.NewListDialog("List Dialog", 50, 15, items)
			m.dialogManager.PushDialog(list)
			return m, nil

		case key.Matches(msg, m.keyMap.Confirmation):
			// Show a confirmation dialog
			confirm := dialog.YesNo("Confirmation", "Are you sure you want to proceed?", false)
			m.dialogManager.PushDialog(confirm)
			return m, nil

		case key.Matches(msg, m.keyMap.Progress):
			// Show a progress dialog
			progress := dialog.NewProgressDialog("Progress Dialog", 60, 10)
			progress.SetProgress(0.0)
			progress.SetLabel("Starting...")
			m.dialogManager.PushDialog(progress)

			// Simulate progress updates
			m.progressVal = 0.0
			return m, progressCmd(m.progressVal)

		case key.Matches(msg, m.keyMap.Nested):
			// Show nested dialogs demonstration
			content := dialog.NewSimpleModalContent("This demonstrates stacked dialogs.\n\nPress Enter to see a nested confirmation dialog.")
			modal := dialog.NewModalDialog("Nested Dialogs Demo", 60, 10, content)
			m.dialogManager.PushDialog(modal)
			return m, nil
		}

	// Handle dialog-specific messages
	case dialog.FormSubmitMsg:
		m.lastResult = fmt.Sprintf("Form submitted with %d fields", len(msg.Fields))
		return m, nil

	case dialog.ListSelectionMsg:
		m.lastResult = fmt.Sprintf("List item selected: %s", msg.SelectedItem.Title())
		return m, nil

	case dialog.ConfirmationMsg:
		if msg.Result == dialog.ConfirmationResultYes {
			m.lastResult = "Confirmation: Yes"
		} else {
			m.lastResult = "Confirmation: No"
		}
		return m, nil

	case dialog.ProgressUpdateMsg:
		// Update the progress
		m.progressVal = msg.Progress

		// Schedule the next update if not complete
		if m.progressVal < 1.0 {
			return m, progressCmd(m.progressVal + 0.1)
		}
		m.lastResult = "Progress complete"
		return m, nil

	case dialog.ProgressCompleteMsg:
		m.lastResult = "Progress completed successfully"
		return m, nil

	case dialog.ProgressCancelMsg:
		m.lastResult = "Progress was canceled"
		return m, nil

		// Special case for nested dialog demo handled in regular KeyMsg case above
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// If dialogs are active, render them on top
	if m.dialogManager.HasDialogs() {
		return m.dialogManager.View()
	}

	// Otherwise, render the main interface
	s := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 2).
		Render("Dialog Component Demo")

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render("Press a key to show a dialog:")

	options := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render(`
1: Modal Dialog
2: Form Dialog
3: List Dialog
4: Confirmation Dialog
5: Progress Dialog
6: Nested Dialogs

?: Toggle Help
q: Quit
`)

	result := ""
	if m.lastResult != "" {
		resultStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Padding(1, 0)

		result = resultStyle.Render("Last result: " + m.lastResult)
	}

	var content string
	if m.showHelp {
		// Show help
		helpStyle := lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder())

		// Create help content manually since keyMap doesn't implement help.KeyMap
		helpContent := "Keyboard Shortcuts:\n\n" +
			"1: Modal Dialog\n" +
			"2: Form Dialog\n" +
			"3: List Dialog\n" +
			"4: Confirmation Dialog\n" +
			"5: Progress Dialog\n" +
			"6: Nested Dialogs\n\n" +
			"?: Toggle Help\n" +
			"q: Quit"

		content = helpStyle.Render(helpContent)
	} else {
		// Show main content
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			"",
			instructions,
			options,
			result,
		)
	}

	return s.Render(content)
}

// progressCmd returns a command that updates the progress bar
func progressCmd(value float64) tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg {
		if value > 1.0 {
			value = 1.0
		}
		return dialog.ProgressUpdateMsg{
			Progress: value,
			Label:    fmt.Sprintf("Progress: %.0f%%", value*100),
		}
	})
}

// Run starts the dialog demo application
func Run() {
	p := tea.NewProgram(New(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running dialog demo:", err)
		os.Exit(1)
	}
}
