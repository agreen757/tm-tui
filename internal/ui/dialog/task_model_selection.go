package dialog

import (
	"fmt"
	"strings"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TaskModelInfo represents an available AI model for task selection
type TaskModelInfo struct {
	ID       string
	Provider string
	Label    string
}

// TaskModelSelectionMsg is sent when a model is selected for a task
type TaskModelSelectionMsg struct {
	TaskID     string
	Model      *TaskModelInfo
	TaskIndex  int
	TotalTasks int
}

// TaskModelSelectionDialog displays task information and allows model selection
type TaskModelSelectionDialog struct {
	BaseFocusableDialog
	Task            *taskmaster.Task
	TaskIndex       int
	TotalTasks      int
	AvailableModels []TaskModelInfo
	SelectedIndex   int
	scrollOffset    int // For scrolling the model list
	visibleItems    int // Number of visible items in the list
	result          *TaskModelInfo
	QueueTaskIDs    []string          // All task IDs in the queue
	ModelSelections map[string]string // taskID -> model mapping for all queued tasks
	GetTaskByID     func(string) *taskmaster.Task
}

// NewTaskModelSelectionDialog creates a new TaskModelSelectionDialog
func NewTaskModelSelectionDialog(task *taskmaster.Task, index, total int) *TaskModelSelectionDialog {
	return NewTaskModelSelectionDialogWithQueue(task, index, total, nil, nil, nil)
}

// NewTaskModelSelectionDialogWithQueue creates a new TaskModelSelectionDialog with queue information
func NewTaskModelSelectionDialogWithQueue(
	task *taskmaster.Task,
	index, total int,
	queueTaskIDs []string,
	modelSelections map[string]string,
	getTaskByID func(string) *taskmaster.Task,
) *TaskModelSelectionDialog {
	if task == nil {
		task = &taskmaster.Task{}
	}

	models := loadTaskAvailableModels()

	// Default selected index is 0
	selectedIndex := 0
	if len(models) == 0 {
		selectedIndex = -1
	}

	dialog := &TaskModelSelectionDialog{
		BaseFocusableDialog: NewBaseFocusableDialog(
			fmt.Sprintf("Select Model for Task %s", task.ID),
			80,
			24, // Taller dialog
			DialogKindCustom,
			len(models),
		),
		Task:            task,
		TaskIndex:       index,
		TotalTasks:      total,
		AvailableModels: models,
		SelectedIndex:   selectedIndex,
		scrollOffset:    0,
		visibleItems:    10, // Show 10 models at a time
		result:          nil,
		QueueTaskIDs:    queueTaskIDs,
		ModelSelections: modelSelections,
		GetTaskByID:     getTaskByID,
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	return dialog
}

// loadTaskAvailableModels loads all available models (hardcoded list matching ModelSelectionDialog)
func loadTaskAvailableModels() []TaskModelInfo {
	// Use the same hardcoded model list as ModelSelectionDialog for consistency
	models := []TaskModelInfo{
		// Anthropic Claude 4.x Models
		{ID: "claude-opus-4-5-20251101", Provider: "anthropic", Label: "Claude Opus 4.5 • ANTHROPIC"},
		{ID: "claude-sonnet-4-5-20250929", Provider: "anthropic", Label: "Claude Sonnet 4.5 • ANTHROPIC"},
		{ID: "claude-haiku-4-5-20251001", Provider: "anthropic", Label: "Claude Haiku 4.5 • ANTHROPIC"},
		{ID: "claude-opus-4-1-20250805", Provider: "anthropic", Label: "Claude Opus 4.1 • ANTHROPIC"},
		{ID: "claude-sonnet-4-20250514", Provider: "anthropic", Label: "Claude Sonnet 4 • ANTHROPIC"},
		{ID: "claude-opus-4-20250514", Provider: "anthropic", Label: "Claude Opus 4 • ANTHROPIC"},
		// OpenAI Models
		{ID: "gpt-4o", Provider: "openai", Label: "GPT-4o • OPENAI"},
		{ID: "gpt-4o-mini", Provider: "openai", Label: "GPT-4o Mini • OPENAI"},
		{ID: "gpt-4-turbo", Provider: "openai", Label: "GPT-4 Turbo • OPENAI"},
		{ID: "o1-preview", Provider: "openai", Label: "o1 Preview • OPENAI"},
		{ID: "o1-mini", Provider: "openai", Label: "o1 Mini • OPENAI"},
		// Google Gemini Models
		{ID: "gemini-2.0-flash", Provider: "google", Label: "Gemini 2.0 Flash • GOOGLE"},
		{ID: "gemini-1.5-pro", Provider: "google", Label: "Gemini 1.5 Pro • GOOGLE"},
		{ID: "gemini-1.5-flash", Provider: "google", Label: "Gemini 1.5 Flash • GOOGLE"},
	}
	return models
}

// ensureSelectedVisible adjusts scroll offset to keep selected item visible
func (d *TaskModelSelectionDialog) ensureSelectedVisible() {
	if d.SelectedIndex < d.scrollOffset {
		d.scrollOffset = d.SelectedIndex
	} else if d.SelectedIndex >= d.scrollOffset+d.visibleItems {
		d.scrollOffset = d.SelectedIndex - d.visibleItems + 1
	}
}

// Init initializes the dialog
func (d *TaskModelSelectionDialog) Init() tea.Cmd {
	return nil
}

// Update processes messages for the dialog
// Note: Key messages are handled by HandleKey() which is called by DialogManager
// We don't call HandleKey() here to avoid double-processing
func (d *TaskModelSelectionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	// Don't process key messages here - DialogManager calls HandleKey() separately
	return d, nil
}

// HandleKey processes keyboard input
func (d *TaskModelSelectionDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if d.SelectedIndex > 0 {
			d.SelectedIndex--
			d.ensureSelectedVisible()
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if d.SelectedIndex < len(d.AvailableModels)-1 {
			d.SelectedIndex++
			d.ensureSelectedVisible()
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("home"))):
		d.SelectedIndex = 0
		d.ensureSelectedVisible()
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("end"))):
		if len(d.AvailableModels) > 0 {
			d.SelectedIndex = len(d.AvailableModels) - 1
		}
		d.ensureSelectedVisible()
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("pageup"))):
		d.SelectedIndex -= d.visibleItems
		if d.SelectedIndex < 0 {
			d.SelectedIndex = 0
		}
		d.ensureSelectedVisible()
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("pagedown"))):
		d.SelectedIndex += d.visibleItems
		if d.SelectedIndex >= len(d.AvailableModels) {
			d.SelectedIndex = len(d.AvailableModels) - 1
		}
		d.ensureSelectedVisible()
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if d.SelectedIndex >= 0 && d.SelectedIndex < len(d.AvailableModels) {
			d.result = &d.AvailableModels[d.SelectedIndex]
			// Return DialogResultNone + cmd that emits DialogResultMsg
			// This bypasses the DialogManager's callback system and ensures
			// the proper message flow to handleDialogResultMsg in delete_workflow.go
			cmd := func() tea.Msg {
				return DialogResultMsg{
					ID:     "task_model_selection_dialog",
					Button: "confirm",
					Value:  d.result.ID, // Pass the model ID as a string
				}
			}
			return DialogResultNone, cmd
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		// Emit cancel message
		cmd := func() tea.Msg {
			return DialogResultMsg{
				ID:     "task_model_selection_dialog",
				Button: "cancel",
				Value:  nil,
			}
		}
		return DialogResultNone, cmd

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
		// Emit cancel message
		cmd := func() tea.Msg {
			return DialogResultMsg{
				ID:     "task_model_selection_dialog",
				Button: "cancel",
				Value:  nil,
			}
		}
		return DialogResultNone, cmd
	}

	return DialogResultNone, nil
}

// View renders the dialog
func (d *TaskModelSelectionDialog) View() string {
	var sb strings.Builder

	// Title line
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51"))

	title := fmt.Sprintf("Select Model - Task %d/%d: %s", d.TaskIndex+1, d.TotalTasks, d.Task.Title)
	if len(title) > 74 {
		title = title[:71] + "..."
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 76))
	sb.WriteString("\n\n")

	// Model list with selection indicator
	endIdx := d.scrollOffset + d.visibleItems
	if endIdx > len(d.AvailableModels) {
		endIdx = len(d.AvailableModels)
	}

	// Show scroll indicator if needed
	if d.scrollOffset > 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  ↑ more above"))
		sb.WriteString("\n")
	}

	for i := d.scrollOffset; i < endIdx; i++ {
		model := d.AvailableModels[i]
		marker := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

		if i == d.SelectedIndex {
			marker = "▶ "
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("16")).
				Bold(true).
				Background(lipgloss.Color("51"))
		}

		line := fmt.Sprintf("%s%-70s", marker, model.Label)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}

	// Show scroll indicator if needed
	if endIdx < len(d.AvailableModels) {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  ↓ more below"))
		sb.WriteString("\n")
	}

	// Footer with hints
	sb.WriteString("\n")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	sb.WriteString(footerStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel"))

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)

	return borderStyle.Render(sb.String())
}

// GetSelectedModel returns the selected model or nil if none selected
func (d *TaskModelSelectionDialog) GetSelectedModel() *TaskModelInfo {
	return d.result
}

// Height returns the dialog height
func (d *TaskModelSelectionDialog) Height() int {
	return d.height
}

// GetFocusedIndex returns the currently focused model index
func (d *TaskModelSelectionDialog) GetFocusedIndex() int {
	return d.SelectedIndex
}

// SetFocusedIndex sets the focused model index
func (d *TaskModelSelectionDialog) SetFocusedIndex(index int) {
	if index >= 0 && index < len(d.AvailableModels) {
		d.SelectedIndex = index
		d.ensureSelectedVisible()
	}
}
