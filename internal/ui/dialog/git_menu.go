package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GitMenuSelectionMsg is sent when a git menu item is selected
type GitMenuSelectionMsg struct {
	SelectedIndex int
}

// menuItem implements the list.Item interface for git menu items
type menuItem struct {
	title       string
	description string
}

// Title returns the title of the menu item
func (i menuItem) Title() string {
	return i.title
}

// Description returns the description of the menu item
func (i menuItem) Description() string {
	return i.description
}

// FilterValue returns the value used for filtering
func (i menuItem) FilterValue() string {
	return i.title
}

// GitMenuDialog is a dialog that shows Git actions
type GitMenuDialog struct {
	BaseDialog
	items         []menuItem
	selectedIndex int
	onSelect      func(int)
}

// NewGitMenuDialog creates a new Git menu dialog
func NewGitMenuDialog(onSelect func(int)) *GitMenuDialog {
	items := []menuItem{
		{title: "Show Status", description: "View detailed repository status"},
		{title: "Switch Branch", description: "Checkout an existing branch"},
		{title: "Create Branch", description: "Create and checkout a new branch"},
		{title: "Recent Commits", description: "View recent commit history"},
	}

	dialog := &GitMenuDialog{
		BaseDialog:    NewBaseDialog("Git Menu", 60, 16, DialogKindList),
		items:         items,
		selectedIndex: 0,
		onSelect:      onSelect,
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)

	return dialog
}

// Init initializes the dialog
func (d *GitMenuDialog) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates dialog state
func (d *GitMenuDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Recalculate position on window resize
		d.Center(msg.Width, msg.Height)
	}

	return d, nil
}

// View renders the dialog
func (d *GitMenuDialog) View() string {
	// Calculate content width accounting for borders and padding
	contentWidth := d.width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Render menu items with spacing
	var items []string
	for i, item := range d.items {
		items = append(items, d.renderItem(item, i == d.selectedIndex, contentWidth))
		// Add spacing between items (except after last item)
		if i < len(d.items)-1 {
			items = append(items, "")
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	return d.RenderBorder(content)
}

// renderItem renders a single menu item
func (d *GitMenuDialog) renderItem(item menuItem, focused bool, width int) string {
	// Style based on focus with enhanced visual hierarchy
	titleStyle := lipgloss.NewStyle().
		Width(width).
		Foreground(d.Style.TextColor)

	prefix := "  "
	if focused {
		// Enhanced focus styling with background highlight
		titleStyle = titleStyle.
			Foreground(d.Style.FocusedBorderColor).
			Bold(true).
			Underline(true).
			Background(lipgloss.Color("#1a1a1a")) // Subtle background highlight
		prefix = "▶ " // More visible arrow indicator
	}

	title := prefix + item.title

	// Add description on a new line with better contrast
	descStyle := lipgloss.NewStyle().
		Width(width).
		PaddingLeft(4)

	if focused {
		descStyle = descStyle.
			Foreground(d.Style.TextColor).
			Italic(true).
			Faint(false) // Keep description readable when focused
	} else {
		descStyle = descStyle.
			Foreground(d.Style.TextColor).
			Italic(true).
			Faint(true)
	}

	desc := item.description
	if len(desc) > width-4 {
		desc = desc[:width-7] + "..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		descStyle.Render(desc),
	)
}

// HandleKey processes a key event
func (d *GitMenuDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle menu-specific keys
	switch msg.String() {
	case "up", "k":
		if d.selectedIndex > 0 {
			d.selectedIndex--
		} else {
			d.selectedIndex = len(d.items) - 1
		}
		return DialogResultNone, nil

	case "down", "j":
		if d.selectedIndex < len(d.items)-1 {
			d.selectedIndex++
		} else {
			d.selectedIndex = 0
		}
		return DialogResultNone, nil

	case "enter":
		// Return DialogResultClose and a command that emits the selection message
		selectedIndex := d.selectedIndex
		return DialogResultClose, func() tea.Msg {
			return GitMenuSelectionMsg{SelectedIndex: selectedIndex}
		}
	}

	return DialogResultNone, nil
}
