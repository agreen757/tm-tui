package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in a list
type ListItem interface {
	// Title returns the title of the item
	Title() string
	// Description returns the description of the item
	Description() string
	// FilterValue returns the value to use for filtering
	FilterValue() string
}

// SimpleListItem is a simple implementation of ListItem
type SimpleListItem struct {
	title       string
	description string
}

// NewSimpleListItem creates a new simple list item
func NewSimpleListItem(title, description string) *SimpleListItem {
	return &SimpleListItem{
		title:       title,
		description: description,
	}
}

// Title returns the title of the item
func (i SimpleListItem) Title() string {
	return i.title
}

// Description returns the description of the item
func (i SimpleListItem) Description() string {
	return i.description
}

// FilterValue returns the value to use for filtering
func (i SimpleListItem) FilterValue() string {
	return i.title
}

// ListSelectionMsg is sent when a list item is selected
type ListSelectionMsg struct {
	SelectedIndex int
	SelectedItem  ListItem
	MultiSelect   bool
	SelectedItems []int
}

// ListDialog is a dialog with a selectable list
type ListDialog struct {
	BaseFocusableDialog
	items           []ListItem
	selectedIndex   int
	offset          int
	multiSelect     bool
	selectedItems   map[int]bool
	showDescription bool
	visibleItems    int
	filterEnabled   bool
	filterFocused   bool
	filterInput     textinput.Model
	filterValue     string
	viewItems       []ListItem
	viewIndices     []int
}

// NewListDialog creates a new list dialog
func NewListDialog(title string, width, height int, items []ListItem) *ListDialog {
	// Calculate visible items based on height
	availHeight := height - 6 // Account for borders, title, etc.
	if availHeight < 1 {
		availHeight = 1
	}

	// Item height depends on whether descriptions are shown
	itemHeight := 1
	visibleItems := availHeight / itemHeight

	dialog := &ListDialog{
		BaseFocusableDialog: NewBaseFocusableDialog(title, width, height, DialogKindList, len(items)),
		items:               items,
		selectedIndex:       0,
		offset:              0,
		multiSelect:         false,
		selectedItems:       make(map[int]bool),
		showDescription:     false,
		visibleItems:        visibleItems,
	}
	dialog.refreshFilteredItems()

	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)

	return dialog
}

// Init initializes the dialog
func (d *ListDialog) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates dialog state
func (d *ListDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)

		// Recalculate visible items
		availHeight := msg.Height - 6
		if availHeight < 1 {
			availHeight = 1
		}

		itemHeight := 1
		if d.showDescription {
			itemHeight = 2
		}

		d.visibleItems = availHeight / itemHeight

		if d.filterEnabled {
			width := msg.Width - 6
			if width < 10 {
				width = 10
			}
			d.filterInput.Width = width
		}
	}

	// Update numElements in case items changed
	d.numElements = len(d.items)

	return d, nil
}

// View renders the dialog
func (d *ListDialog) View() string {
	// Account for border and padding
	contentWidth := d.width - 4

	if contentWidth < 1 {
		contentWidth = 1
	}

	// Render list items
	listContent := d.renderItems(contentWidth)

	// Add border and title
	return d.RenderBorder(listContent)
}

// renderItems renders the list items
func (d *ListDialog) renderItems(width int) string {
	items := d.viewItems
	if len(items) == 0 {
		message := "No items"
		if d.filterEnabled && strings.TrimSpace(d.filterValue) != "" {
			message = "No matches"
		}
		content := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Render(message)
		return d.renderFilterContent(content, width)
	}

	if d.selectedIndex < 0 {
		d.selectedIndex = 0
	} else if d.selectedIndex >= len(items) {
		d.selectedIndex = len(items) - 1
	}

	maxVisible := d.visibleItems
	if maxVisible < 1 {
		maxVisible = 1
	}
	if d.filterEnabled && maxVisible > 2 {
		maxVisible -= 2
	}
	if maxVisible < 1 {
		maxVisible = 1
	}

	if d.selectedIndex < d.offset {
		d.offset = d.selectedIndex
	} else if d.selectedIndex >= d.offset+maxVisible {
		d.offset = d.selectedIndex - maxVisible + 1
	}

	if d.offset < 0 {
		d.offset = 0
	}

	end := d.offset + maxVisible
	if end > len(items) {
		end = len(items)
	}

	itemStrs := make([]string, end-d.offset)
	for i := d.offset; i < end; i++ {
		actualIdx := d.actualIndex(i)
		isFocused := i == d.selectedIndex
		_, isSelected := d.selectedItems[actualIdx]
		itemStrs[i-d.offset] = d.renderItem(items[i], width, isFocused, isSelected)
	}

	if d.offset > 0 || end < len(items) {
		scrollInfo := ""
		if d.offset > 0 {
			scrollInfo += "↑ "
		}
		scrollInfo += lipgloss.NewStyle().
			Foreground(d.Style.TextColor).
			Render("PgUp/PgDn scroll • Space selects")
		if end < len(items) {
			scrollInfo += " ↓"
		}
		scrollStyle := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center)
		if len(itemStrs) < maxVisible {
			itemStrs = append(itemStrs, scrollStyle.Render(scrollInfo))
		} else if len(itemStrs) > 0 {
			itemStrs[len(itemStrs)-1] = scrollStyle.Render(scrollInfo)
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, itemStrs...)
	return d.renderFilterContent(body, width)
}

// renderItem renders a single list item
func (d *ListDialog) renderItem(item ListItem, width int, focused, selected bool) string {
	if item == nil {
		return ""
	}
	// Determine the item prefix (selection indicator)
	prefix := "  "

	if d.multiSelect {
		if selected {
			prefix = "[x] "
		} else {
			prefix = "[ ] "
		}
	}

	// Style based on focus and selection
	style := lipgloss.NewStyle().
		Width(width).
		Foreground(d.Style.TextColor)

	if focused {
		// Enhanced accessibility: Make focus very clear with multiple indicators
		style = style.
			Foreground(d.Style.FocusedBorderColor).
			Bold(true).
			Underline(true) // Add underline for better visibility
		prefix = "> " + prefix[2:]
	}

	// Render the item
	title := prefix + item.Title()

	if d.showDescription && item.Description() != "" {
		// Add description on a new line
		descStyle := style.Copy().
			Foreground(d.Style.TextColor).
			Italic(true).
			Bold(false).
			PaddingLeft(4)

		// Truncate description if too long
		desc := item.Description()
		if len(desc) > width-4 {
			desc = desc[:width-7] + "..."
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			style.Render(title),
			descStyle.Render(desc),
		)
	}

	return style.Render(title)
}

func (d *ListDialog) renderFilterContent(body string, width int) string {
	if !d.filterEnabled {
		return body
	}

	inputWidth := width - 6
	if inputWidth < 10 {
		inputWidth = 10
	}
	d.filterInput.Width = inputWidth

	labelStyle := lipgloss.NewStyle().Foreground(d.Style.TextColor)
	hintStyle := lipgloss.NewStyle().Foreground(d.Style.TextColor).Faint(true)

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Filter: ")+d.filterInput.View(),
		hintStyle.Render("Type to filter • Esc clears • Enter to accept"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (d *ListDialog) actualIndex(viewIndex int) int {
	if viewIndex < 0 {
		return 0
	}
	if d.filterEnabled && len(d.viewIndices) > viewIndex {
		return d.viewIndices[viewIndex]
	}
	if viewIndex >= len(d.items) {
		return len(d.items) - 1
	}
	return viewIndex
}

func (d *ListDialog) activateFilter() tea.Cmd {
	if !d.filterEnabled {
		return nil
	}
	d.filterFocused = true
	d.filterInput.Focus()
	return textinput.Blink
}

func (d *ListDialog) handleFilterInput(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if d.filterInput.Value() == "" {
			d.filterFocused = false
			d.filterInput.Blur()
		} else {
			d.filterInput.SetValue("")
			d.applyFilter("")
		}
		return DialogResultNone, nil
	case "enter":
		d.filterFocused = false
		d.filterInput.Blur()
		return DialogResultNone, nil
	default:
		var cmd tea.Cmd
		d.filterInput, cmd = d.filterInput.Update(msg)
		d.applyFilter(d.filterInput.Value())
		return DialogResultNone, cmd
	}
}

func (d *ListDialog) applyFilter(value string) {
	if !d.filterEnabled {
		return
	}
	d.filterValue = value
	d.refreshFilteredItems()
	d.offset = 0
}

// HandleKey processes a key event
func (d *ListDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	if d.filterEnabled && d.filterFocused {
		return d.handleFilterInput(msg)
	}

	if d.filterEnabled {
		if msg.String() == "/" {
			return DialogResultNone, d.activateFilter()
		}
		if msg.Type == tea.KeyCtrlF {
			return DialogResultNone, d.activateFilter()
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && !msg.Alt {
			cmd := d.activateFilter()
			var updateCmd tea.Cmd
			d.filterInput, updateCmd = d.filterInput.Update(msg)
			d.applyFilter(d.filterInput.Value())
			return DialogResultNone, tea.Batch(cmd, updateCmd)
		}
	}

	// First check base dialog keys (like ESC)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	totalItems := len(d.viewItems)
	if totalItems == 0 {
		return DialogResultNone, nil
	}

	// Handle list-specific keys
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if d.selectedIndex > 0 {
			d.selectedIndex--
		} else {
			d.selectedIndex = totalItems - 1
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if d.selectedIndex < totalItems-1 {
			d.selectedIndex++
		} else {
			d.selectedIndex = 0
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
		d.selectedIndex = 0
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
		d.selectedIndex = totalItems - 1
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("pageup", "pgup"))):
		d.selectedIndex -= d.visibleItems
		if d.selectedIndex < 0 {
			d.selectedIndex = 0
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("pagedown", "pgdown"))):
		d.selectedIndex += d.visibleItems
		if d.selectedIndex >= totalItems {
			d.selectedIndex = totalItems - 1
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys(" ", "space"))):
		if d.multiSelect {
			actualIdx := d.actualIndex(d.selectedIndex)
			if d.selectedItems[actualIdx] {
				delete(d.selectedItems, actualIdx)
			} else {
				d.selectedItems[actualIdx] = true
			}
			return DialogResultNone, nil
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
		// Toggle description
		d.showDescription = !d.showDescription
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if totalItems > 0 {
			var selectedItems []int
			if d.multiSelect {
				for idx := range d.selectedItems {
					selectedItems = append(selectedItems, idx)
				}
			} else {
				selectedItems = []int{d.actualIndex(d.selectedIndex)}
			}

			currentItem := d.viewItems[d.selectedIndex]
			return DialogResultConfirm, func() tea.Msg {
				return ListSelectionMsg{
					SelectedIndex: d.selectedIndex,
					SelectedItem:  currentItem,
					MultiSelect:   d.multiSelect,
					SelectedItems: selectedItems,
				}
			}
		}
	}

	return DialogResultNone, nil
}

// SelectedIndex returns the selected index
func (d *ListDialog) SelectedIndex() int {
	return d.selectedIndex
}

// SetSelectedIndex sets the selected index
func (d *ListDialog) SetSelectedIndex(index int) {
	if index >= 0 && index < len(d.items) {
		d.selectedIndex = index
	}
}

// SelectedItem returns the selected item
func (d *ListDialog) SelectedItem() ListItem {
	if d.selectedIndex >= 0 && d.selectedIndex < len(d.viewItems) {
		return d.viewItems[d.selectedIndex]
	}
	return nil
}

// SelectedItems returns the selected items
func (d *ListDialog) SelectedItems() []ListItem {
	var items []ListItem
	for i := range d.selectedItems {
		if i >= 0 && i < len(d.items) {
			items = append(items, d.items[i])
		}
	}
	return items
}

// SetShowDescription toggles description rendering in the list.
func (d *ListDialog) SetShowDescription(show bool) {
	d.showDescription = show
}

// SetMultiSelect sets whether multiple items can be selected
func (d *ListDialog) SetMultiSelect(multiSelect bool) {
	d.multiSelect = multiSelect
}

// EnableFiltering turns on inline filtering for the dialog.
func (d *ListDialog) EnableFiltering(placeholder string) {
	d.filterEnabled = true
	d.filterInput = textinput.New()
	d.filterInput.Prompt = ""
	d.filterInput.CharLimit = 200
	d.filterInput.Placeholder = placeholder
	d.filterInput.TextStyle = lipgloss.NewStyle().Foreground(d.Style.TextColor)
	d.filterInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(d.Style.WarningColor)
	d.filterInput.CursorStyle = lipgloss.NewStyle().Foreground(d.Style.ButtonColor)
	d.refreshFilteredItems()
}

func (d *ListDialog) refreshFilteredItems() {
	d.viewItems = append([]ListItem{}, d.items...)
	d.viewIndices = make([]int, len(d.items))
	for i := range d.items {
		d.viewIndices[i] = i
	}

	if d.filterEnabled {
		value := strings.ToLower(strings.TrimSpace(d.filterValue))
		if value != "" {
			filtered := make([]ListItem, 0, len(d.items))
			indices := make([]int, 0, len(d.items))
			for idx, item := range d.items {
				if item == nil {
					continue
				}
				if strings.Contains(strings.ToLower(item.FilterValue()), value) {
					filtered = append(filtered, item)
					indices = append(indices, idx)
				}
			}
			d.viewItems = filtered
			d.viewIndices = indices
		}
	}

	d.numElements = len(d.viewItems)

	if d.selectedIndex >= len(d.viewItems) {
		d.selectedIndex = len(d.viewItems) - 1
	}
	if d.selectedIndex < 0 {
		d.selectedIndex = 0
	}
}

// SetItems sets the items in the list
func (d *ListDialog) SetItems(items []ListItem) {
	d.items = items
	d.refreshFilteredItems()

	// Clear selected items
	d.selectedItems = make(map[int]bool)
}
