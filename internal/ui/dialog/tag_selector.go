package dialog

import (
	"fmt"
	"log"
	"strings"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TagSelectorConfig configures the tag selection dialog
type TagSelectorConfig struct {
	Title       string
	MultiSelect bool
	TagList     *taskmaster.TagList
}

// TagSelectorResult represents the result of tag selection
type TagSelectorResult struct {
	SelectedTags []string
	AddNewTag    bool
}

// TagItem represents a tag in the selector for display and selection
type TagItem struct {
	Tag   taskmaster.TagContext
	Index int
	IsNew bool
}

// Title returns the title of the tag item
func (t TagItem) Title() string {
	if t.IsNew {
		return "➕ Add New Tag..."
	}
	return t.Tag.Name
}

// Description returns metadata about the tag
func (t TagItem) Description() string {
	if t.IsNew {
		return ""
	}

	// Format: "N tasks (M completed) | created on DATE | DESCRIPTION"
	parts := []string{
		fmt.Sprintf("%d task%s", t.Tag.TaskCount, pluralize(t.Tag.TaskCount)),
	}

	if t.Tag.CompletedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d completed", t.Tag.CompletedCount))
	}

	desc := strings.Join(parts, ", ")

	if t.Tag.CreatedLabel != "" {
		desc = fmt.Sprintf("%s | %s", desc, t.Tag.CreatedLabel)
	}

	if t.Tag.Description != "" {
		desc = fmt.Sprintf("%s | %s", desc, t.Tag.Description)
	}

	return desc
}

// FilterValue returns the value to use for filtering
func (t TagItem) FilterValue() string {
	if t.IsNew {
		return "add new"
	}
	return strings.ToLower(t.Tag.Name + " " + t.Tag.Description)
}

// pluralize returns "s" for counts != 1
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// TagSelector is a dialog for selecting tags with single or multi-select support
type TagSelector struct {
	*BaseFocusableDialog
	config        TagSelectorConfig
	items         []TagItem
	selectedIndex int
	offset        int
	multiSelect   bool
	selectedItems map[int]bool
	visibleItems  int
	viewItems     []TagItem
	viewIndices   []int
}

// NewTagSelector creates a new tag selector dialog
func NewTagSelector(cfg TagSelectorConfig) *TagSelector {
	if cfg.Title == "" {
		cfg.Title = "Select Tags"
	}

	// Convert tag list to items and add "Add New Tag..." at the end
	items := make([]TagItem, 0)
	if cfg.TagList != nil && len(cfg.TagList.Tags) > 0 {
		for i, tag := range cfg.TagList.Tags {
			items = append(items, TagItem{
				Tag:   tag,
				Index: i,
				IsNew: false,
			})
		}
	}

	// Add "Add New Tag..." option
	items = append(items, TagItem{
		Index: len(items),
		IsNew: true,
	})

	// Initialize visibleItems to a reasonable default
	// This will be updated by SetRect() called by the dialog manager
	// Start with minimum visible items to avoid incorrect clipping
	visibleItems := 5

	bfd := NewBaseFocusableDialog(cfg.Title, 60, 30, DialogKindList, len(items))

	selector := &TagSelector{
		BaseFocusableDialog: &bfd,
		config:              cfg,
		items:               items,
		selectedIndex:       0,
		offset:              0,
		multiSelect:         cfg.MultiSelect,
		selectedItems:       make(map[int]bool),
		visibleItems:        visibleItems,
	}

	// Set footer hints
	if cfg.MultiSelect {
		selector.SetFooterHints(
			ShortcutHint{Key: "↑/↓", Label: "Navigate"},
			ShortcutHint{Key: "PgUp/PgDn", Label: "Page"},
			ShortcutHint{Key: "Space", Label: "Toggle"},
			ShortcutHint{Key: "Enter", Label: "Confirm"},
			ShortcutHint{Key: "Esc", Label: "Cancel"},
		)
	} else {
		selector.SetFooterHints(
			ShortcutHint{Key: "↑/↓", Label: "Navigate"},
			ShortcutHint{Key: "PgUp/PgDn", Label: "Page"},
			ShortcutHint{Key: "Enter", Label: "Select"},
			ShortcutHint{Key: "Esc", Label: "Cancel"},
		)
	}

	selector.refreshFilteredItems()
	return selector
}

// refreshFilteredItems updates the view items list
func (t *TagSelector) refreshFilteredItems() {
	t.viewItems = t.items
	t.viewIndices = make([]int, len(t.items))
	for i := range t.items {
		t.viewIndices[i] = i
	}
}

// Init initializes the dialog
func (t *TagSelector) Init() tea.Cmd {
	return nil
}

// Update processes messages
func (t *TagSelector) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	log.Printf("[TagSelector.Update] Received message type: %T", msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Printf("[TagSelector.Update] WindowSizeMsg: %dx%d", msg.Width, msg.Height)
		// Only center the dialog, don't change its size
		// The dialog manager handles sizing via SetRect()
		t.BaseFocusableDialog.Center(msg.Width, msg.Height)
		log.Printf("[TagSelector.Update] visibleItems: %d, selectedIndex: %d", 
			t.visibleItems, t.selectedIndex)
	case tea.KeyMsg:
		log.Printf("[TagSelector.Update] KeyMsg received: %s (should not process this - DialogManager handles it)", msg.String())
	}

	log.Printf("[TagSelector.Update] Returning dialog with selectedIndex=%d", t.selectedIndex)
	return t, nil
}

// View renders the dialog
func (t *TagSelector) View() string {
	contentWidth := t.width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	listContent := t.renderItems(contentWidth)
	return t.RenderBorder(listContent)
}

// renderItems renders the tag list items
func (t *TagSelector) renderItems(width int) string {
	log.Printf("[TagSelector.renderItems] Called with width=%d, selectedIndex=%d, viewItems=%d",
		width, t.selectedIndex, len(t.viewItems))

	if len(t.viewItems) == 0 {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Render("No tags available")
	}

	if t.selectedIndex < 0 {
		t.selectedIndex = 0
	} else if t.selectedIndex >= len(t.viewItems) {
		t.selectedIndex = len(t.viewItems) - 1
	}

	maxVisible := t.visibleItems
	if maxVisible < 1 {
		// Failsafe: should never happen, but ensure we show at least 3 items
		maxVisible = 3
		log.Printf("[TagSelector.renderItems] WARNING: visibleItems=%d, using failsafe=%d", 
			t.visibleItems, maxVisible)
	}

	// Check if we need to reserve space for scroll indicator
	// If scrolling is needed, reserve 1 line for indicator
	needsScrolling := len(t.viewItems) > maxVisible
	itemsToShow := maxVisible
	if needsScrolling {
		itemsToShow = maxVisible - 1 // Reserve 1 line for scroll indicator
		if itemsToShow < 1 {
			itemsToShow = 1
		}
	}

	if t.selectedIndex < t.offset {
		t.offset = t.selectedIndex
	} else if t.selectedIndex >= t.offset+itemsToShow {
		t.offset = t.selectedIndex - itemsToShow + 1
		if t.offset < 0 {
			t.offset = 0
		}
	}

	var lines []string
	endIdx := t.offset + itemsToShow
	if endIdx > len(t.viewItems) {
		endIdx = len(t.viewItems)
	}

	log.Printf("[TagSelector.renderItems] Rendering items %d to %d (offset=%d, itemsToShow=%d, needsScrolling=%v)",
		t.offset, endIdx, t.offset, itemsToShow, needsScrolling)

	for i := t.offset; i < endIdx; i++ {
		item := t.viewItems[i]
		selected := i == t.selectedIndex
		log.Printf("[TagSelector.renderItems] Item %d: '%s', selected=%v", i, item.Title(), selected)
		line := t.renderItemLine(item, selected, width)
		lines = append(lines, line)
	}

	// Add scroll indicator on new line if scrolling is needed
	if needsScrolling {
		scrollInfo := ""
		if t.offset > 0 {
			scrollInfo += "↑ "
		}
		scrollInfo += fmt.Sprintf("(%d-%d of %d)", t.offset+1, endIdx, len(t.viewItems))
		if endIdx < len(t.viewItems) {
			scrollInfo += " ↓"
		}

		scrollStyle := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("241")). // Dim gray
			Italic(true)

		lines = append(lines, scrollStyle.Render(scrollInfo))
	}

	return strings.Join(lines, "\n")
}

// renderItemLine renders a single tag item
func (t *TagSelector) renderItemLine(item TagItem, selected bool, width int) string {
	log.Printf("[TagSelector.renderItemLine] Rendering '%s', selected=%v", item.Title(), selected)

	title := item.Title()
	description := item.Description()

	// Visual indicators
	var indicator string
	if item.IsNew {
		indicator = "  "
	} else {
		if t.selectedItems[item.Index] {
			indicator = "☑ "
		} else {
			indicator = "☐ "
		}

		// Add active indicator if applicable
		if item.Tag.Active {
			title = fmt.Sprintf("%s (active)", title)
		}
	}

	// Build the line
	content := indicator + title
	if description != "" {
		content = fmt.Sprintf("%s\n    %s", content, description)
	}

	// Add cursor prefix for selected item
	var prefix string
	if selected {
		prefix = "> "
		log.Printf("[TagSelector.renderItemLine] APPLYING SELECTION (with cursor) to '%s'", title)
	} else {
		prefix = "  "
		log.Printf("[TagSelector.renderItemLine] Normal style for '%s'", title)
	}
	content = prefix + content

	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1)

	if selected {
		style = style.
			Background(lipgloss.Color("62")).  // Blue background
			Foreground(lipgloss.Color("230")). // Light yellow foreground
			Bold(true)
	}

	rendered := style.Render(content)
	log.Printf("[TagSelector.renderItemLine] Rendered output length: %d chars, selected=%v", len(rendered), selected)
	return rendered
}

// HandleKey processes keyboard input
func (t *TagSelector) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	log.Printf("[TagSelector.HandleKey] Key pressed: %s, Current selectedIndex: %d, viewItems count: %d",
		msg.String(), t.selectedIndex, len(t.viewItems))

	switch msg.String() {
	case "up", "k":
		oldIndex := t.selectedIndex
		if t.selectedIndex > 0 {
			t.selectedIndex--
		}
		log.Printf("[TagSelector.HandleKey] UP: index changed from %d to %d", oldIndex, t.selectedIndex)
		return DialogResultNone, nil

	case "down", "j":
		oldIndex := t.selectedIndex
		if t.selectedIndex < len(t.viewItems)-1 {
			t.selectedIndex++
		}
		log.Printf("[TagSelector.HandleKey] DOWN: index changed from %d to %d", oldIndex, t.selectedIndex)
		return DialogResultNone, nil

	case "pgup":
		// Page up - move up by visibleItems
		oldIndex := t.selectedIndex
		t.selectedIndex -= t.visibleItems
		if t.selectedIndex < 0 {
			t.selectedIndex = 0
		}
		log.Printf("[TagSelector.HandleKey] PGUP: index changed from %d to %d", oldIndex, t.selectedIndex)
		return DialogResultNone, nil

	case "pgdown":
		// Page down - move down by visibleItems
		oldIndex := t.selectedIndex
		t.selectedIndex += t.visibleItems
		if t.selectedIndex >= len(t.viewItems) {
			t.selectedIndex = len(t.viewItems) - 1
		}
		log.Printf("[TagSelector.HandleKey] PGDOWN: index changed from %d to %d", oldIndex, t.selectedIndex)
		return DialogResultNone, nil

	case "enter":
		log.Printf("[TagSelector.HandleKey] ENTER: confirming selection at index %d", t.selectedIndex)
		if t.multiSelect {
			// In multi-select mode, confirm selection
			return DialogResultConfirm, nil
		}

		// In single-select mode, immediately select
		item := t.viewItems[t.selectedIndex]
		if item.IsNew {
			// Mark the "Add New Tag" option as selected (last item)
			lastIdx := len(t.items) - 1
			t.selectedItems[lastIdx] = true
			log.Printf("[TagSelector.HandleKey] Selected 'Add New Tag' option")
		} else {
			t.selectedItems[item.Index] = true
			log.Printf("[TagSelector.HandleKey] Selected tag: %s", item.Tag.Name)
		}
		return DialogResultConfirm, nil

	case " ":
		if t.multiSelect {
			// Toggle selection
			item := t.viewItems[t.selectedIndex]
			if !item.IsNew {
				t.selectedItems[item.Index] = !t.selectedItems[item.Index]
				log.Printf("[TagSelector.HandleKey] SPACE: toggled tag %s to %v", item.Tag.Name, t.selectedItems[item.Index])
			}
			return DialogResultNone, nil
		}
		log.Printf("[TagSelector.HandleKey] SPACE: ignored (not multi-select mode)")

	case "esc":
		log.Printf("[TagSelector.HandleKey] ESC: cancelling dialog")
		return DialogResultCancel, nil
	}

	log.Printf("[TagSelector.HandleKey] Unhandled key: %s", msg.String())
	return DialogResultNone, nil
}

// GetSelectedTags returns the currently selected tag names
func (t *TagSelector) GetSelectedTags() []string {
	var selected []string
	for idx := range t.selectedItems {
		if idx < len(t.items) && !t.items[idx].IsNew {
			selected = append(selected, t.items[idx].Tag.Name)
		}
	}
	return selected
}

// HasAddNewTag returns true if the "Add New Tag" option was selected
func (t *TagSelector) HasAddNewTag() bool {
	// Check if the last item (Add New Tag) is selected
	lastIdx := len(t.items) - 1
	return t.selectedItems[lastIdx]
}

// GetResult returns the final selection result
func (t *TagSelector) GetResult() TagSelectorResult {
	return TagSelectorResult{
		SelectedTags: t.GetSelectedTags(),
		AddNewTag:    t.HasAddNewTag(),
	}
}

// DialogResultValue implements DialogResultProvider interface
func (t *TagSelector) DialogResultValue() (interface{}, error) {
	return t.GetResult(), nil
}

// SetRect sets the dialog's position and size
func (t *TagSelector) SetRect(width, height, x, y int) {
	t.BaseFocusableDialog.SetRect(width, height, x, y)
	
	// Calculate visible items same as ListDialog: height - 6 (borders, title, padding)
	// The scroll indicator (if needed) will be handled gracefully in renderItems
	availHeight := height - 15
	if availHeight < 1 {
		availHeight = 1
	}
	
	// Only update if there's a significant change
	if availHeight != t.visibleItems {
		oldVisible := t.visibleItems
		t.visibleItems = availHeight
		
		log.Printf("[TagSelector.SetRect] visibleItems: %d → %d", oldVisible, t.visibleItems)
		
		// Ensure selectedIndex is within valid bounds after visibility change
		if t.selectedIndex >= len(t.viewItems) && len(t.viewItems) > 0 {
			t.selectedIndex = len(t.viewItems) - 1
			log.Printf("[TagSelector.SetRect] Adjusted selectedIndex to %d", t.selectedIndex)
		}
		if t.selectedIndex < 0 && len(t.viewItems) > 0 {
			t.selectedIndex = 0
		}
	}
}
