package dialog

import (
	"fmt"
	"strings"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ExpandTaskPreviewDialog displays a preview of the task expansion as a tree view
type ExpandTaskPreviewDialog struct {
	title           string
	description     string
	drafts          []taskmaster.SubtaskDraft
	flattened       []taskmaster.FlattenedDraft
	selectedIndex   int
	scrollOffset    int
	width           int
	height          int
	style           *DialogStyle
	focused         bool
	cancelCallback  func()
	continueCallback func()
	expandedNodes   map[string]bool
	maxHeight       int // Max displayable lines
}

// NewExpandTaskPreviewDialog creates a new preview dialog for expanded tasks
func NewExpandTaskPreviewDialog(
	drafts []taskmaster.SubtaskDraft,
	style *DialogStyle,
) *ExpandTaskPreviewDialog {
	if style == nil {
		style = DefaultDialogStyle()
	}

	d := &ExpandTaskPreviewDialog{
		title:         "Preview Expanded Tasks",
		description:   "Review the proposed subtask hierarchy. Use arrow keys to navigate, Enter to continue, Esc to cancel.",
		drafts:        drafts,
		flattened:     taskmaster.FlattenDrafts(drafts),
		selectedIndex: 0,
		scrollOffset:  0,
		width:         72,
		height:        20,
		style:         style,
		focused:       true,
		expandedNodes: make(map[string]bool),
		maxHeight:     18, // Leave room for title, description, and footer
	}

	// Expand all nodes by default
	d.expandAllNodes()

	return d
}

func (d *ExpandTaskPreviewDialog) expandAllNodes() {
	for i := range d.flattened {
		d.expandedNodes[fmt.Sprintf("%d", i)] = true
	}
}

// Init implements tea.Model
func (d *ExpandTaskPreviewDialog) Init() tea.Cmd {
	return nil
}

// Update implements Dialog interface
func (d *ExpandTaskPreviewDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.focused {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return d.handleKeyMsg(msg)
	}

	return d, nil
}

func (d *ExpandTaskPreviewDialog) handleKeyMsg(msg tea.KeyMsg) (Dialog, tea.Cmd) {
	switch msg.String() {
	case "up":
		if d.selectedIndex > 0 {
			d.selectedIndex--
			d.ensureVisible()
		}
		return d, nil

	case "down":
		if d.selectedIndex < len(d.flattened)-1 {
			d.selectedIndex++
			d.ensureVisible()
		}
		return d, nil

	case "enter":
		if d.continueCallback != nil {
			d.continueCallback()
		}
		return d, nil

	case "esc", "ctrl+c":
		if d.cancelCallback != nil {
			d.cancelCallback()
		}
		return d, nil

	case "home":
		d.selectedIndex = 0
		d.scrollOffset = 0
		return d, nil

	case "end":
		d.selectedIndex = len(d.flattened) - 1
		d.ensureVisible()
		return d, nil

	case "pgup":
		d.selectedIndex -= d.maxHeight / 2
		if d.selectedIndex < 0 {
			d.selectedIndex = 0
		}
		d.ensureVisible()
		return d, nil

	case "pgdn":
		d.selectedIndex += d.maxHeight / 2
		if d.selectedIndex >= len(d.flattened) {
			d.selectedIndex = len(d.flattened) - 1
		}
		d.ensureVisible()
		return d, nil
	}

	return d, nil
}

func (d *ExpandTaskPreviewDialog) ensureVisible() {
	if d.selectedIndex < d.scrollOffset {
		d.scrollOffset = d.selectedIndex
	} else if d.selectedIndex >= d.scrollOffset+d.maxHeight {
		d.scrollOffset = d.selectedIndex - d.maxHeight + 1
	}
}

// View implements tea.Model
func (d *ExpandTaskPreviewDialog) View() string {
	content := d.renderContent()
	
	borderColor := d.style.BorderColor
	if d.focused {
		borderColor = d.style.FocusedBorderColor
	}

	style := lipgloss.NewStyle().
		Padding(0, 1).
		BorderStyle(d.style.Border).
		BorderForeground(borderColor).
		Width(d.width - 2).
		Height(d.height).
		Background(d.style.BackgroundColor).
		Foreground(d.style.TextColor)

	return style.Render(content)
}

func (d *ExpandTaskPreviewDialog) renderContent() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(d.style.TitleColor).
		Bold(true)
	title := titleStyle.Render(d.title)

	descStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor).
		PaddingBottom(1)
	desc := descStyle.Render(d.description)

	// Render tree view
	treeLines := d.renderTreeView()

	// Render footer
	footer := d.renderFooter()

	// Build content
	lines := []string{
		title,
		desc,
		"",
	}

	// Calculate visible lines
	startIdx := d.scrollOffset
	endIdx := startIdx + d.maxHeight
	if endIdx > len(treeLines) {
		endIdx = len(treeLines)
	}

	visibleLines := treeLines[startIdx:endIdx]
	lines = append(lines, visibleLines...)

	// Pad to height
	for len(lines) < d.height-2 {
		lines = append(lines, "")
	}

	lines = append(lines, footer)

	return strings.Join(lines, "\n")
}

func (d *ExpandTaskPreviewDialog) renderTreeView() []string {
	lines := make([]string, 0)

	for i, fd := range d.flattened {
		// Determine if this item is selected
		isSelected := i == d.selectedIndex
		line := d.renderTreeItem(fd, isSelected)
		lines = append(lines, line)
	}

	return lines
}

func (d *ExpandTaskPreviewDialog) renderTreeItem(fd taskmaster.FlattenedDraft, isSelected bool) string {
	// Build indentation based on level
	indent := strings.Repeat("  ", fd.Level)

	// Add tree glyph
	glyph := "├─"
	if fd.Level == 0 {
		glyph = "•"
	}

	// Build the display text
	text := fmt.Sprintf("%s %s%s", glyph, indent, fd.Draft.Title)

	// Add description if available
	if fd.Draft.Description != "" && fd.Draft.Description != fd.Draft.Title {
		desc := fd.Draft.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		text = fmt.Sprintf("%s — %s", text, desc)
	}

	// Apply selection styling
	if isSelected {
		text = " > " + text
		highlightStyle := lipgloss.NewStyle().
			Foreground(d.style.ButtonColor).
			Bold(true)
		return highlightStyle.Render(text)
	}

	textStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor)
	return textStyle.Render(text)
}

func (d *ExpandTaskPreviewDialog) renderFooter() string {
	// Show item count and navigation hints
	info := fmt.Sprintf("[%d/%d] ↑↓:Navigate Enter:Continue Esc:Cancel", 
		d.selectedIndex+1, len(d.flattened))

	// Right-align additional info
	childCount := 0
	if d.selectedIndex < len(d.flattened) {
		childCount = len(d.flattened[d.selectedIndex].Draft.Children)
	}
	if childCount > 0 {
		info = fmt.Sprintf("%s | %d child items", info, childCount)
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor).
		PaddingTop(1)
	return footerStyle.Render(info)
}

// SetSize sets the size of the dialog
func (d *ExpandTaskPreviewDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.maxHeight = height - 5 // Account for title, description, and footer
}

// SetFocused sets whether the dialog is focused
func (d *ExpandTaskPreviewDialog) SetFocused(focused bool) {
	d.focused = focused
}

// SetContinueCallback sets the callback when user confirms
func (d *ExpandTaskPreviewDialog) SetContinueCallback(cb func()) {
	d.continueCallback = cb
}

// SetCancelCallback sets the callback when user cancels
func (d *ExpandTaskPreviewDialog) SetCancelCallback(cb func()) {
	d.cancelCallback = cb
}

// GetSelectedDrafts returns the currently selected drafts (all of them for now)
func (d *ExpandTaskPreviewDialog) GetSelectedDrafts() []taskmaster.SubtaskDraft {
	return d.drafts
}

// HandleKey implements Dialog interface
func (d *ExpandTaskPreviewDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	_, cmd := d.handleKeyMsg(msg)
	return DialogResultNone, cmd
}

// SetRect implements Dialog interface
func (d *ExpandTaskPreviewDialog) SetRect(width, height, x, y int) {
	d.SetSize(width, height)
}

// GetRect implements Dialog interface
func (d *ExpandTaskPreviewDialog) GetRect() (width, height, x, y int) {
	return d.width, d.height, 0, 0
}

// Title implements Dialog interface
func (d *ExpandTaskPreviewDialog) Title() string {
	return d.title
}

// Kind implements Dialog interface
func (d *ExpandTaskPreviewDialog) Kind() DialogKind {
	return DialogTypeConfirmation
}

// ZIndex implements Dialog interface
func (d *ExpandTaskPreviewDialog) ZIndex() int {
	return 0
}

// SetZIndex implements Dialog interface
func (d *ExpandTaskPreviewDialog) SetZIndex(z int) {
	// Placeholder for z-index management
}

// IsCancellable implements Dialog interface
func (d *ExpandTaskPreviewDialog) IsCancellable() bool {
	return true
}

// IsFocused implements Dialog interface
func (d *ExpandTaskPreviewDialog) IsFocused() bool {
	return d.focused
}
