package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalContent is the interface for modal content
type ModalContent interface {
	// Init initializes the content
	Init() tea.Cmd
	// Update processes messages and updates content state
	Update(msg tea.Msg) (ModalContent, tea.Cmd)
	// View renders the content
	View(width, height int) string
	// HandleKey processes a key event
	HandleKey(msg tea.KeyMsg) tea.Cmd
}

// SimpleModalContent is a simple implementation of ModalContent
type SimpleModalContent struct {
	text string
}

// NewSimpleModalContent creates a new simple modal content
func NewSimpleModalContent(text string) *SimpleModalContent {
	return &SimpleModalContent{
		text: text,
	}
}

// Init initializes the content
func (c *SimpleModalContent) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates content state
func (c *SimpleModalContent) Update(msg tea.Msg) (ModalContent, tea.Cmd) {
	return c, nil
}

// View renders the content
func (c *SimpleModalContent) View(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center)

	return style.Render(c.text)
}

// HandleKey processes a key event
func (c *SimpleModalContent) HandleKey(msg tea.KeyMsg) tea.Cmd {
	return nil
}

// ModalDialog is a simple modal dialog
type ModalDialog struct {
	BaseDialog
	content ModalContent
}

// NewModalDialog creates a new modal dialog
func NewModalDialog(title string, width, height int, content ModalContent) *ModalDialog {
	return &ModalDialog{
		BaseDialog: NewBaseDialog(title, width, height, DialogKindModal),
		content:    content,
	}
}

// Init initializes the dialog
func (d *ModalDialog) Init() tea.Cmd {
	return d.content.Init()
}

// Update processes messages and updates dialog state
func (d *ModalDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
	}

	updatedContent, cmd := d.content.Update(msg)
	d.content = updatedContent
	return d, cmd
}

// View renders the dialog
func (d *ModalDialog) View() string {
	// Account for border and padding
	contentWidth := d.width - 4
	contentHeight := d.height - 4

	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render content
	content := d.content.View(contentWidth, contentHeight)

	// Add border and title
	return d.RenderBorder(content)
}

// HandleKey processes a key event
func (d *ModalDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (like ESC)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Then check if content handles the key
	contentCmd := d.content.HandleKey(msg)
	if contentCmd != nil {
		return DialogResultNone, contentCmd
	}

	// Default modal keys
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " ", "space"))):
		return DialogResultConfirm, nil
	}

	return DialogResultNone, nil
}

// ModalButtonKind represents the type of button in a modal
type ModalButtonKind int

const (
	// ButtonNone is no button
	ButtonNone ModalButtonKind = iota
	// ButtonOk is an "OK" button
	ButtonOk
	// ButtonCancel is a "Cancel" button
	ButtonCancel
	// ButtonYes is a "Yes" button
	ButtonYes
	// ButtonNo is a "No" button
	ButtonNo
	// ButtonCustom is a custom button
	ButtonCustom
)

// ModalButton represents a button in a modal dialog
type ModalButton struct {
	Kind    ModalButtonKind
	Label   string
	OnClick func() (DialogResult, tea.Cmd)
}

// ButtonModalDialog is a modal dialog with buttons
type ButtonModalDialog struct {
	ModalDialog
	buttons       []ModalButton
	selectedIndex int
}

// NewButtonModalDialog creates a new button modal dialog
func NewButtonModalDialog(title string, width, height int, content ModalContent, buttons []ModalButton) *ButtonModalDialog {
	dialog := &ButtonModalDialog{
		ModalDialog:   *NewModalDialog(title, width, height, content),
		buttons:       buttons,
		selectedIndex: 0,
	}
	dialog.SetFooterHints(
		ShortcutHint{Key: "←/→", Label: "Change Button"},
		ShortcutHint{Key: "Enter", Label: "Confirm"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)
	return dialog
}

// Update processes messages and updates dialog state
func (d *ButtonModalDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	dialog, cmd := d.ModalDialog.Update(msg)
	d.ModalDialog = *dialog.(*ModalDialog)
	return d, cmd
}

// View renders the dialog
func (d *ButtonModalDialog) View() string {
	// Account for border, padding and button row
	contentWidth := d.width - 4
	contentHeight := d.height - 6

	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render content
	content := d.content.View(contentWidth, contentHeight)

	// Render buttons
	buttonRow := d.renderButtons()

	// Combine content and buttons
	combinedContent := lipgloss.JoinVertical(lipgloss.Left, content, buttonRow)

	// Add border and title
	return d.RenderBorder(combinedContent)
}

// renderButtons renders the button row
func (d *ButtonModalDialog) renderButtons() string {
	if len(d.buttons) == 0 {
		return ""
	}

	buttonStrs := make([]string, len(d.buttons))

	for i, button := range d.buttons {
		label := button.Label
		if label == "" {
			switch button.Kind {
			case ButtonOk:
				label = "OK"
			case ButtonCancel:
				label = "Cancel"
			case ButtonYes:
				label = "Yes"
			case ButtonNo:
				label = "No"
			default:
				label = "Button"
			}
		}

		style := lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(d.Style.ButtonColor)

		if i == d.selectedIndex {
			style = style.
				Bold(true).
				Underline(true)
		}

		buttonStrs[i] = style.Render(label)
	}

	// Center the buttons
	row := lipgloss.JoinHorizontal(lipgloss.Center, buttonStrs...)

	style := lipgloss.NewStyle().
		Width(d.width - 4).
		Align(lipgloss.Center).
		PaddingTop(1)

	return style.Render(row)
}

// HandleKey processes a key event
func (d *ButtonModalDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (like ESC)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Then check if content handles the key
	contentCmd := d.content.HandleKey(msg)
	if contentCmd != nil {
		return DialogResultNone, contentCmd
	}

	// Navigate between buttons
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
		if len(d.buttons) > 0 {
			d.selectedIndex = (d.selectedIndex - 1 + len(d.buttons)) % len(d.buttons)
		}
		return DialogResultNone, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
		if len(d.buttons) > 0 {
			d.selectedIndex = (d.selectedIndex + 1) % len(d.buttons)
		}
		return DialogResultNone, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " ", "space"))):
		if len(d.buttons) > 0 && d.selectedIndex >= 0 && d.selectedIndex < len(d.buttons) {
			if d.buttons[d.selectedIndex].OnClick != nil {
				return d.buttons[d.selectedIndex].OnClick()
			}

			// Default button behavior
			switch d.buttons[d.selectedIndex].Kind {
			case ButtonOk, ButtonYes:
				return DialogResultConfirm, nil
			case ButtonCancel, ButtonNo:
				return DialogResultCancel, nil
			}
		}
		return DialogResultConfirm, nil
	}

	return DialogResultNone, nil
}

// OkDialog creates a simple OK dialog
func OkDialog(title string, message string) *ButtonModalDialog {
	content := NewSimpleModalContent(message)

	buttons := []ModalButton{
		{
			Kind:    ButtonOk,
			Label:   "OK",
			OnClick: func() (DialogResult, tea.Cmd) { return DialogResultConfirm, nil },
		},
	}

	// Calculate width based on message length
	width := len(message) + 10
	if width < 30 {
		width = 30
	}
	if width > 80 {
		width = 80
	}

	// Calculate height based on message length and width
	lines := len(strings.Split(message, "\n"))
	messageChars := len(message)
	estimatedLines := messageChars / (width - 10)
	if estimatedLines > lines {
		lines = estimatedLines
	}
	height := lines + 8

	return NewButtonModalDialog(title, width, height, content, buttons)
}
