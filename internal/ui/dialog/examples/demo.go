package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	width           int
	height          int
	ready           bool
	content         string
	dialogManager   *dialog.DialogManager
	showHelp        bool
	progressValue   float64
	progressRunning bool
}

func initialModel() Model {
	return Model{
		content:         "Welcome to Dialog Demo!\n\nPress '?' for help.",
		dialogManager:   nil, // Will be initialized on WindowSizeMsg
		showHelp:        false,
		progressValue:   0.0,
		progressRunning: false,
	}
}

// Init initializes the program
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// First, check if we need to handle a window size message
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Initialize dialog manager with terminal dimensions
		if m.dialogManager == nil {
			m.dialogManager = dialog.NewDialogManager(msg.Width, msg.Height)
		}
		return m, nil
	}

	// Check if we have any active dialogs
	if m.dialogManager != nil && m.dialogManager.HasDialogs() {
		// Let dialog manager handle the message first
		cmd := m.dialogManager.HandleMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check for dialog-specific messages
		switch msg := msg.(type) {
		case dialog.ListSelectionMsg:
			// Handle list selection
			m.content = fmt.Sprintf("Selected item: %s", msg.SelectedItem.Title())
			return m, nil

		case dialog.FormSubmitMsg:
			// Handle form submission
			var result strings.Builder
			result.WriteString("Form submitted:\n")
			for _, field := range msg.Fields {
				switch field.Kind {
				case dialog.FormFieldText:
					result.WriteString(fmt.Sprintf("%s: %s\n", field.Label, field.TextInput.Value()))
				case dialog.FormFieldCheckbox:
					result.WriteString(fmt.Sprintf("%s: %v\n", field.Label, field.Checked))
				case dialog.FormFieldRadioGroup:
					if field.SelectedOption >= 0 && field.SelectedOption < len(field.Options) {
						result.WriteString(fmt.Sprintf("%s: %s\n", field.Label, field.Options[field.SelectedOption]))
					}
				}
			}
			m.content = result.String()
			return m, nil

		case dialog.ConfirmationMsg:
			// Handle confirmation result
			if msg.Result == dialog.ConfirmationResultYes {
				m.content = "Confirmation: Yes"
			} else {
				m.content = "Confirmation: No"
			}
			return m, nil

		case dialog.ProgressUpdateMsg:
			// Update progress bar
			m.progressValue = msg.Progress
			if m.progressValue >= 1.0 {
				m.progressRunning = false
				m.content = "Progress completed!"
			}
			return m, nil

		case dialog.ProgressCompleteMsg:
			// Progress complete
			m.progressRunning = false
			m.content = "Progress completed!"
			return m, nil

		case dialog.ProgressCancelMsg:
			// Progress cancelled
			m.progressRunning = false
			m.content = "Progress cancelled!"
			return m, nil
		}

		// If dialogs are open, they get priority for key events
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Batch(cmds...)
		}
	}

	// If we get here, no dialogs consumed the message
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			// Toggle help
			m.showHelp = !m.showHelp

		case "m":
			// Show modal dialog
			content := dialog.NewSimpleModalContent("This is a simple modal dialog with a message.\n\nPress Esc to close or Enter to confirm.")
			modalDialog := dialog.NewModalDialog("Modal Dialog", 50, 10, content)
			m.dialogManager.PushDialog(modalDialog)

		case "b":
			// Show button modal dialog
			content := dialog.NewSimpleModalContent("This modal has buttons.\n\nChoose one option.")
			buttons := []dialog.ModalButton{
				{
					Kind:  dialog.ButtonOk,
					Label: "OK",
					OnClick: func() (dialog.DialogResult, tea.Cmd) {
						return dialog.DialogResultConfirm, nil
					},
				},
				{
					Kind:  dialog.ButtonCancel,
					Label: "Cancel",
					OnClick: func() (dialog.DialogResult, tea.Cmd) {
						return dialog.DialogResultCancel, nil
					},
				},
			}
			buttonDialog := dialog.NewButtonModalDialog("Button Modal", 50, 15, content, buttons)
			m.dialogManager.PushDialog(buttonDialog)

		case "l":
			// Show list dialog
			items := []dialog.ListItem{
				dialog.NewSimpleListItem("Item 1", "Description for item 1"),
				dialog.NewSimpleListItem("Item 2", "Description for item 2"),
				dialog.NewSimpleListItem("Item 3", "Description for item 3"),
				dialog.NewSimpleListItem("Item 4", "Description for item 4"),
				dialog.NewSimpleListItem("Item 5", "Description for item 5"),
			}
			listDialog := dialog.NewListDialog("List Dialog", 60, 15, items)
			listDialog.SetMultiSelect(true)
			m.dialogManager.PushDialog(listDialog)

		case "f":
			// Show form dialog
			fields := []dialog.FormField{
				dialog.NewTextField("Name", "Enter your name", true),
				dialog.NewTextField("Email", "Enter your email", false),
				dialog.NewCheckboxField("Subscribe to newsletter", false),
				dialog.NewRadioGroupField("Favorite color", []string{"Red", "Green", "Blue", "Yellow"}, 0),
			}
			formDialog := dialog.NewLegacyFormDialog("Form Dialog", 60, 20, fields)
			m.dialogManager.PushDialog(formDialog)

		case "c":
			// Show confirmation dialog
			confirmDialog := dialog.YesNo("Confirmation", "Are you sure you want to continue?", false)
			m.dialogManager.PushDialog(confirmDialog)

		case "d":
			// Show danger confirmation dialog
			dangerDialog := dialog.YesNo("Warning", "This action cannot be undone. Continue?", true)
			dangerDialog.SetYesText("Delete")
			dangerDialog.SetNoText("Cancel")
			dangerDialog.SetYesDefault(false) // No is default for dangerous actions
			m.dialogManager.PushDialog(dangerDialog)

		case "p":
			// Show progress dialog
			if !m.progressRunning {
				progressDialog := dialog.NewProgressDialog("Progress", 60, 10)
				progressDialog.SetLabel("Processing items...")
				m.dialogManager.PushDialog(progressDialog)
				m.progressValue = 0.0
				m.progressRunning = true

				// Start a fake progress updater
				cmds = append(cmds, m.progressUpdater())
			}

		case "o":
			// Show OK dialog (message box)
			okDialog := dialog.OkDialog("Information", "This is a simple OK dialog that shows a message.")
			m.dialogManager.PushDialog(okDialog)
		}
	}

	return m, tea.Batch(cmds...)
}

// progressUpdater simulates a progress bar advancing
func (m Model) progressUpdater() tea.Cmd {
	return func() tea.Msg {
		// Don't update if we're not running
		if !m.progressRunning {
			return nil
		}

		// Update progress in small increments
		m.progressValue += 0.1
		if m.progressValue > 1.0 {
			m.progressValue = 1.0
		}

		// Create a progress update message
		return dialog.ProgressUpdateMsg{
			Progress: m.progressValue,
			Label:    fmt.Sprintf("Processing: %.0f%%", m.progressValue*100),
		}
	}
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// If we have a dialog, render it
	if m.dialogManager != nil && m.dialogManager.HasDialogs() {
		return m.dialogManager.View()
	}

	if m.showHelp {
		return m.renderHelp()
	}

	// Main content
	mainContent := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(1).
		Render(fmt.Sprintf("%s\n\n%s",
			m.content,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6D98BA")).
				Render("Press 'm' for modal, 'l' for list, 'f' for form, 'c' for confirm, 'p' for progress"),
		))

	return mainContent
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(1)

	helpContent := `Dialog Demo Help

Key Commands:
  m - Show modal dialog
  b - Show modal with buttons
  l - Show list dialog
  f - Show form dialog
  c - Show confirmation dialog
  d - Show dangerous confirmation
  p - Show progress dialog
  o - Show OK dialog

  ? - Toggle help
  q - Quit

In dialogs:
  Escape - Cancel/close dialog
  Enter - Confirm
  Tab/Shift+Tab - Navigate form fields
  Arrow keys - Navigate
  Space - Toggle selection`

	return helpStyle.Render(helpContent)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
