package dialog

import (
	"fmt"
	"strings"

	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SubtaskEditDialog allows editing of proposed subtasks before application
type SubtaskEditDialog struct {
	title            string
	drafts           []taskmaster.SubtaskDraft
	editingIndex     int
	selectedIndex    int
	editingMode      bool // true if editing a subtask, false if navigating list
	editTitleInput   textinput.Model
	editDescInput    textinput.Model
	scrollOffset     int
	width            int
	height           int
	style            *DialogStyle
	focused          bool
	cancelCallback   func()
	confirmCallback  func([]taskmaster.SubtaskDraft)
	maxHeight        int
	editFieldFocus   int // 0=title, 1=description
}

// NewSubtaskEditDialog creates a dialog for editing proposed subtasks
func NewSubtaskEditDialog(
	drafts []taskmaster.SubtaskDraft,
	style *DialogStyle,
) *SubtaskEditDialog {
	if style == nil {
		style = DefaultDialogStyle()
	}

	titleInput := textinput.New()
	titleInput.Placeholder = "Subtask title (required)"
	titleInput.Focus()

	descInput := textinput.New()
	descInput.Placeholder = "Description (optional)"

	d := &SubtaskEditDialog{
		title:          "Edit Subtasks",
		drafts:         copyDrafts(drafts),
		selectedIndex:  0,
		editingIndex:   -1,
		editingMode:    false,
		editTitleInput: titleInput,
		editDescInput:  descInput,
		scrollOffset:   0,
		width:          72,
		height:         20,
		style:          style,
		focused:        true,
		maxHeight:      16,
		editFieldFocus: 0,
	}

	return d
}

func copyDrafts(drafts []taskmaster.SubtaskDraft) []taskmaster.SubtaskDraft {
	result := make([]taskmaster.SubtaskDraft, len(drafts))
	for i, d := range drafts {
		result[i] = copyDraft(d)
	}
	return result
}

func copyDraft(d taskmaster.SubtaskDraft) taskmaster.SubtaskDraft {
	result := taskmaster.SubtaskDraft{
		Title:       d.Title,
		Description: d.Description,
	}
	if len(d.Children) > 0 {
		result.Children = copyDrafts(d.Children)
	}
	return result
}

// Init implements tea.Model
func (d *SubtaskEditDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements Dialog interface
func (d *SubtaskEditDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.focused {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return d.handleKeyMsg(msg)
	}

	return d, nil
}

func (d *SubtaskEditDialog) handleKeyMsg(msg tea.KeyMsg) (Dialog, tea.Cmd) {
	// In edit mode, handle text input
	if d.editingMode {
		return d.handleEditMode(msg)
	}

	// In navigation mode, handle list navigation
	switch msg.String() {
	case "up":
		if d.selectedIndex > 0 {
			d.selectedIndex--
			d.ensureVisible()
		}
		return d, nil

	case "down":
		if d.selectedIndex < len(d.drafts)-1 {
			d.selectedIndex++
			d.ensureVisible()
		}
		return d, nil

	case "enter":
		if d.confirmCallback != nil {
			d.confirmCallback(d.drafts)
		}
		return d, nil

	case "a", "A":
		// Add new subtask
		d.drafts = append(d.drafts, taskmaster.SubtaskDraft{
			Title: "New subtask",
		})
		d.selectedIndex = len(d.drafts) - 1
		d.startEditMode()
		return d, nil

	case "d", "D":
		// Delete current subtask
		if len(d.drafts) > 0 {
			d.drafts = append(d.drafts[:d.selectedIndex], d.drafts[d.selectedIndex+1:]...)
			if d.selectedIndex >= len(d.drafts) && d.selectedIndex > 0 {
				d.selectedIndex--
			}
			d.ensureVisible()
		}
		return d, nil

	case "e", "E":
		// Edit current subtask
		if d.selectedIndex < len(d.drafts) {
			d.startEditMode()
		}
		return d, nil

	case "esc", "ctrl+c":
		if d.cancelCallback != nil {
			d.cancelCallback()
		}
		return d, nil
	}

	return d, nil
}

func (d *SubtaskEditDialog) handleEditMode(msg tea.KeyMsg) (Dialog, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm edit - move to next field or exit edit mode
		if d.editFieldFocus == 0 {
			// Move to description field
			d.editFieldFocus = 1
			d.editDescInput.Focus()
			return d, nil
		} else {
			// Confirm edit
			if d.editTitleInput.Value() == "" {
				// Title is required
				d.showError("Title is required")
				return d, nil
			}
			d.drafts[d.selectedIndex].Title = d.editTitleInput.Value()
			d.drafts[d.selectedIndex].Description = d.editDescInput.Value()
			d.exitEditMode()
			return d, nil
		}

	case "tab":
		// Switch between title and description fields
		if d.editFieldFocus == 0 {
			d.editFieldFocus = 1
			d.editDescInput.Focus()
		} else {
			d.editFieldFocus = 0
			d.editTitleInput.Focus()
		}
		return d, nil

	case "esc":
		// Cancel edit without saving
		d.exitEditMode()
		return d, nil

	default:
		// Handle text input in the appropriate field
		var cmd tea.Cmd
		if d.editFieldFocus == 0 {
			d.editTitleInput, cmd = d.editTitleInput.Update(msg)
		} else {
			d.editDescInput, cmd = d.editDescInput.Update(msg)
		}
		return d, cmd
	}
}

func (d *SubtaskEditDialog) startEditMode() {
	d.editingMode = true
	if d.selectedIndex < len(d.drafts) {
		d.editingIndex = d.selectedIndex
		d.editTitleInput.SetValue(d.drafts[d.selectedIndex].Title)
		d.editDescInput.SetValue(d.drafts[d.selectedIndex].Description)
		d.editFieldFocus = 0
		d.editTitleInput.Focus()
	}
}

func (d *SubtaskEditDialog) exitEditMode() {
	d.editingMode = false
	d.editingIndex = -1
	d.editTitleInput.SetValue("")
	d.editDescInput.SetValue("")
	d.editFieldFocus = 0
	d.editTitleInput.Blur()
	d.editDescInput.Blur()
}

func (d *SubtaskEditDialog) showError(msg string) {
	// For now, just print to log or show briefly
	// This can be enhanced with a transient error display
}

func (d *SubtaskEditDialog) ensureVisible() {
	if d.selectedIndex < d.scrollOffset {
		d.scrollOffset = d.selectedIndex
	} else if d.selectedIndex >= d.scrollOffset+d.maxHeight {
		d.scrollOffset = d.selectedIndex - d.maxHeight + 1
	}
}

// View implements tea.Model
func (d *SubtaskEditDialog) View() string {
	if d.editingMode {
		return d.renderEditMode()
	}
	return d.renderListMode()
}

func (d *SubtaskEditDialog) renderListMode() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(d.style.TitleColor).
		Bold(true)
	title := titleStyle.Render(d.title)
	
	descStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor).
		PaddingBottom(1)
	desc := descStyle.Render(
		"Edit the proposed subtasks. Keys: a=add, d=delete, e=edit, ↑/↓=navigate, Enter=confirm, Esc=cancel",
	)

	// Render list
	listLines := d.renderList()

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
	if endIdx > len(listLines) {
		endIdx = len(listLines)
	}

	visibleLines := listLines[startIdx:endIdx]
	lines = append(lines, visibleLines...)

	// Pad to height
	for len(lines) < d.height-2 {
		lines = append(lines, "")
	}

	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	
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

func (d *SubtaskEditDialog) renderEditMode() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(d.style.TitleColor).
		Bold(true)
	title := titleStyle.Render("Edit Subtask")

	titleLabelStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor)
	titleLabel := "Title (required):"
	if d.editFieldFocus == 0 {
		titleLabelStyle = titleLabelStyle.
			Foreground(d.style.ButtonColor).
			Bold(true)
	}
	titleLabel = titleLabelStyle.Render(titleLabel)

	descLabelStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor)
	descLabel := "Description (optional):"
	if d.editFieldFocus == 1 {
		descLabelStyle = descLabelStyle.
			Foreground(d.style.ButtonColor).
			Bold(true)
	}
	descLabel = descLabelStyle.Render(descLabel)

	footerStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor).
		PaddingTop(1)
	
	lines := []string{
		title,
		"",
		titleLabel,
		d.editTitleInput.View(),
		"",
		descLabel,
		d.editDescInput.View(),
		"",
		footerStyle.Render("Tab:Switch Field Enter:Confirm Esc:Cancel"),
	}

	content := strings.Join(lines, "\n")
	
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

func (d *SubtaskEditDialog) renderList() []string {
	lines := make([]string, 0)

	for i, draft := range d.drafts {
		isSelected := i == d.selectedIndex
		line := d.renderListItem(draft, isSelected)
		lines = append(lines, line)
	}

	return lines
}

func (d *SubtaskEditDialog) renderListItem(draft taskmaster.SubtaskDraft, isSelected bool) string {
	text := fmt.Sprintf("• %s", draft.Title)

	if draft.Description != "" && draft.Description != draft.Title {
		desc := draft.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		text = fmt.Sprintf("%s — %s", text, desc)
	}

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

func (d *SubtaskEditDialog) renderFooter() string {
	info := fmt.Sprintf("[%d/%d] a:Add d:Delete e:Edit ↑↓:Navigate Enter:Confirm Esc:Cancel",
		d.selectedIndex+1, len(d.drafts))
	footerStyle := lipgloss.NewStyle().
		Foreground(d.style.TextColor).
		PaddingTop(1)
	return footerStyle.Render(info)
}

// SetSize sets the size of the dialog
func (d *SubtaskEditDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.maxHeight = height - 5
}

// SetFocused sets whether the dialog is focused
func (d *SubtaskEditDialog) SetFocused(focused bool) {
	d.focused = focused
}

// SetContinueCallback sets the callback when user confirms
func (d *SubtaskEditDialog) SetContinueCallback(cb func([]taskmaster.SubtaskDraft)) {
	d.confirmCallback = cb
}

// SetCancelCallback sets the callback when user cancels
func (d *SubtaskEditDialog) SetCancelCallback(cb func()) {
	d.cancelCallback = cb
}

// GetDrafts returns the current drafts being edited
func (d *SubtaskEditDialog) GetDrafts() []taskmaster.SubtaskDraft {
	return d.drafts
}

// HandleKey implements Dialog interface
func (d *SubtaskEditDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	_, cmd := d.handleKeyMsg(msg)
	return DialogResultNone, cmd
}

// SetRect implements Dialog interface
func (d *SubtaskEditDialog) SetRect(width, height, x, y int) {
	d.SetSize(width, height)
}

// GetRect implements Dialog interface
func (d *SubtaskEditDialog) GetRect() (width, height, x, y int) {
	return d.width, d.height, 0, 0
}

// Title implements Dialog interface
func (d *SubtaskEditDialog) Title() string {
	return d.title
}

// Kind implements Dialog interface
func (d *SubtaskEditDialog) Kind() DialogKind {
	return DialogTypeForm
}

// ZIndex implements Dialog interface
func (d *SubtaskEditDialog) ZIndex() int {
	return 0
}

// SetZIndex implements Dialog interface
func (d *SubtaskEditDialog) SetZIndex(z int) {
	// Placeholder for z-index management
}

// IsCancellable implements Dialog interface
func (d *SubtaskEditDialog) IsCancellable() bool {
	return true
}

// IsFocused implements Dialog interface
func (d *SubtaskEditDialog) IsFocused() bool {
	return d.focused
}
