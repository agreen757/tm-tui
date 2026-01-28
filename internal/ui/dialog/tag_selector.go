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

// renderFilterStatus renders a styled visual indicator for filter mode
// Returns empty string if no filter is active
func renderFilterStatus(filterValue string, matchCount int, totalCount int) string {
	if filterValue == "" {
		return ""
	}

	// Truncate filter text to 20 characters if needed
	truncatedFilter := filterValue
	if len(filterValue) > 20 {
		truncatedFilter = filterValue[:20]
	}

	// Format the filter status text
	filterText := fmt.Sprintf("[FILTER: %s] %d/%d", truncatedFilter, matchCount, totalCount)

	// Create styled output using dialog theme colors
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).  // Focused border color (blue)
		Background(lipgloss.Color("238")). // Border color (dark gray)
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	return style.Render(filterText)
}

// TagSelector is a dialog for selecting tags with single or multi-select support.
//
// Features:
// - Single-select mode: pressing Enter immediately selects and closes
// - Multi-select mode: Space toggles selection, Enter confirms batch selection
// - Keyboard-driven navigation: ↑/↓/PgUp/PgDn/Home/End for moving through list
// - Real-time filtering: Press '/' to enter filter mode for case-insensitive tag search
// - Focus management: Automatic focus restoration when exiting filter mode
// - Selection persistence: Selected tags remain marked when filtered out
//
// The dialog embeds both BaseFocusableDialog (for dialog UI) and BaseFilterable (for
// filtering state management), enabling seamless integration with the DialogManager's
// filter-aware key routing system.
//
// Filtering uses case-insensitive substring matching via TagItem.FilterValue(), which
// returns the concatenation of tag Name and Description. This enables searching both
// by tag name and by description text.
//
// Performance is optimized for <100ms filtering on lists of 1000+ tags using efficient
// string matching with strings.Contains().
type TagSelector struct {
	*BaseFocusableDialog
	*BaseFilterable
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

// NewTagSelector creates a new tag selector dialog with the provided configuration.
//
// The returned TagSelector is initialized with:
// - All tags from cfg.TagList converted to selectable items
// - An "Add New Tag" option appended as the last item
// - Default title "Select Tags" if none provided
// - Filtering enabled for keyboard-driven tag search
// - Multi-select or single-select mode based on cfg.MultiSelect
//
// The selector is ready to be added to a DialogManager for display and interaction.
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
	
	// Initialize BaseFilterable with filtering enabled
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)

	selector := &TagSelector{
		BaseFocusableDialog: &bfd,
		BaseFilterable:      bf,
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

// refreshFilteredItems updates the view items list based on current filter value
func (t *TagSelector) refreshFilteredItems() {
	filterValue := t.GetFilterValue()
	
	// If no filter, show all items
	if filterValue == "" {
		t.viewItems = t.items
		t.viewIndices = make([]int, len(t.items))
		for i := range t.items {
			t.viewIndices[i] = i
		}
		log.Printf("[TagSelector.refreshFilteredItems] No filter, showing all %d items", len(t.items))
		return
	}
	
	// Apply filter: case-insensitive substring match using FilterValue()
	// Always include "Add New Tag" (last item with IsNew=true)
	filterLower := strings.ToLower(filterValue)
	filteredItems := make([]TagItem, 0, len(t.items))
	filteredIndices := make([]int, 0, len(t.items))
	
	for i, item := range t.items {
		// Always include the "Add New Tag" item
		if item.IsNew {
			filteredItems = append(filteredItems, item)
			filteredIndices = append(filteredIndices, i)
			continue
		}
		
		// Use FilterValue() for consistent filtering logic
		itemFilterValue := strings.ToLower(item.FilterValue())
		if strings.Contains(itemFilterValue, filterLower) {
			filteredItems = append(filteredItems, item)
			filteredIndices = append(filteredIndices, i)
		}
	}
	
	t.viewItems = filteredItems
	t.viewIndices = filteredIndices
	
	// Ensure selectedIndex is within bounds
	if t.selectedIndex >= len(t.viewItems) && len(t.viewItems) > 0 {
		t.selectedIndex = len(t.viewItems) - 1
		log.Printf("[TagSelector.refreshFilteredItems] Adjusted selectedIndex to %d (filtered count: %d)", 
			t.selectedIndex, len(t.viewItems))
	} else if len(t.viewItems) == 0 {
		t.selectedIndex = 0
	}
	
	log.Printf("[TagSelector.refreshFilteredItems] Filter='%s' matched %d/%d items", 
		filterValue, len(t.viewItems), len(t.items))
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

	// Update title to show filter status if filtering
	originalTitle := t.config.Title
	if t.IsFiltering() {
		// Calculate match count: count items matching current filter
		matchCount := len(t.viewItems)
		totalCount := len(t.items)
		
		// Get filter status view (includes formatting and styling)
		filterStatus := renderFilterStatus(t.GetFilterValue(), matchCount, totalCount)
		
		// Prepend filter status to title if it's not empty
		if filterStatus != "" {
			t.TitleText = filterStatus
		} else {
			t.TitleText = fmt.Sprintf("[FILTER: %s]", t.GetFilterValue())
		}
		log.Printf("[TagSelector.View] In filter mode: filter='%s', matches=%d/%d", 
			t.GetFilterValue(), matchCount, totalCount)
	} else {
		// Restore original title when not filtering
		t.TitleText = originalTitle
	}

	listContent := t.renderItems(contentWidth)
	result := t.RenderBorder(listContent)
	
	// Restore original title for next render
	t.TitleText = originalTitle
	
	return result
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
	log.Printf("[TagSelector.HandleKey] Key pressed: %s, Current selectedIndex: %d, viewItems count: %d, filtering: %v",
		msg.String(), t.selectedIndex, len(t.viewItems), t.IsFiltering())

	// Handle filter mode keys
	switch msg.String() {
	case "/":
		if !t.IsFiltering() && t.IsFilteringEnabled() {
			t.EnterFilterMode()
			log.Printf("[TagSelector.HandleKey] Entered filter mode")
			return DialogResultNone, nil
		}
	case "esc":
		if t.IsFiltering() {
			t.ExitFilterMode()
			t.SetFilterValue("")
			log.Printf("[TagSelector.HandleKey] Exited filter mode via Esc")
			return DialogResultNone, nil
		}
		// Exit dialog if not filtering
		log.Printf("[TagSelector.HandleKey] ESC: cancelling dialog")
		return DialogResultCancel, nil
	case "enter":
		if t.IsFiltering() {
			t.ExitFilterMode()
			log.Printf("[TagSelector.HandleKey] Exited filter mode via Enter")
			return DialogResultNone, nil
		}
		// Normal enter handling if not filtering
		log.Printf("[TagSelector.HandleKey] ENTER: confirming selection at index %d", t.selectedIndex)
		if t.multiSelect {
			return DialogResultConfirm, nil
		}
		// Single-select mode
		if t.selectedIndex < len(t.viewItems) {
			item := t.viewItems[t.selectedIndex]
			if item.IsNew {
				lastIdx := len(t.items) - 1
				t.selectedItems[lastIdx] = true
				log.Printf("[TagSelector.HandleKey] Selected 'Add New Tag' option")
			} else {
				t.selectedItems[item.Index] = true
				log.Printf("[TagSelector.HandleKey] Selected tag: %s", item.Tag.Name)
			}
		}
		return DialogResultConfirm, nil
	}

	// If in filter mode, handle typing characters
	if t.IsFiltering() {
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			// Add typed character to filter
			newFilter := t.GetFilterValue() + string(msg.Runes)
			t.SetFilterValue(newFilter)
			log.Printf("[TagSelector.HandleKey] Filter updated to: %s", newFilter)
			return DialogResultNone, nil
		}
		// Handle backspace in filter mode
		if msg.String() == "backspace" {
			filter := t.GetFilterValue()
			if len(filter) > 0 {
				newFilter := filter[:len(filter)-1]
				t.SetFilterValue(newFilter)
				log.Printf("[TagSelector.HandleKey] Filter backspaced to: %s", newFilter)
			}
			return DialogResultNone, nil
		}
	}

	// Normal navigation keys (when not filtering or after navigation)
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
		oldIndex := t.selectedIndex
		t.selectedIndex -= t.visibleItems
		if t.selectedIndex < 0 {
			t.selectedIndex = 0
		}
		log.Printf("[TagSelector.HandleKey] PGUP: index changed from %d to %d", oldIndex, t.selectedIndex)
		return DialogResultNone, nil

	case "pgdown":
		oldIndex := t.selectedIndex
		t.selectedIndex += t.visibleItems
		if t.selectedIndex >= len(t.viewItems) {
			t.selectedIndex = len(t.viewItems) - 1
		}
		log.Printf("[TagSelector.HandleKey] PGDOWN: index changed from %d to %d", oldIndex, t.selectedIndex)
		return DialogResultNone, nil

	case " ":
		if t.multiSelect {
			item := t.viewItems[t.selectedIndex]
			if !item.IsNew {
				t.selectedItems[item.Index] = !t.selectedItems[item.Index]
				log.Printf("[TagSelector.HandleKey] SPACE: toggled tag %s to %v", item.Tag.Name, t.selectedItems[item.Index])
			}
			return DialogResultNone, nil
		}
		log.Printf("[TagSelector.HandleKey] SPACE: ignored (not multi-select mode)")
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

// FilterableComponent interface implementation
// These methods delegate to the embedded BaseFilterable

// EnableFiltering enables or disables filtering for this tag selector
func (t *TagSelector) EnableFiltering(enabled bool) {
	t.BaseFilterable.EnableFiltering(enabled)
}

// SetFilterValue sets the current filter value and applies filtering
func (t *TagSelector) SetFilterValue(value string) {
	t.BaseFilterable.SetFilterValue(value)
	// Apply filtering based on new filter value
	t.refreshFilteredItems()
	log.Printf("[TagSelector.SetFilterValue] Filter set to '%s', refreshed items", value)
}

// GetFilterValue returns the current filter value
func (t *TagSelector) GetFilterValue() string {
	return t.BaseFilterable.GetFilterValue()
}

// IsFiltering returns whether the selector is currently in filtering mode
func (t *TagSelector) IsFiltering() bool {
	return t.BaseFilterable.IsFiltering()
}

// EnterFilterMode transitions the selector into filtering mode
func (t *TagSelector) EnterFilterMode() {
	t.BaseFilterable.StoreFocusIndex(t.selectedIndex)
	t.BaseFilterable.EnterFilterMode()
	log.Printf("[TagSelector.EnterFilterMode] Stored focus index: %d", t.selectedIndex)
}

// ExitFilterMode transitions the selector out of filtering mode and restores focus
func (t *TagSelector) ExitFilterMode() {
	t.BaseFilterable.ExitFilterMode()
	// Restore previous focus index if available
	if idx, ok := t.BaseFilterable.GetStoredFocusIndex(); ok {
		if idx >= 0 && idx < len(t.viewItems) {
			t.selectedIndex = idx
		}
		t.BaseFilterable.ClearStoredFocusIndex()
	}
	log.Printf("[TagSelector.ExitFilterMode] Filter mode exited, focus restored to: %d", t.selectedIndex)
}

// Compile-time assertion that TagSelector implements FilterableComponent
var _ FilterableComponent = (*TagSelector)(nil)
