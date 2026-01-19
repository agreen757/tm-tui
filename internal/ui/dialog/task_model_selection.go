package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TaskModelInfo represents an available AI model for task selection
type TaskModelInfo struct {
	ID       string
	Provider string
	Label    string
}

// TaskModelSelectionMsg is sent when a model is selected for a task
type TaskModelSelectionMsg struct {
	TaskID    string
	Model     *TaskModelInfo
	TaskIndex int
	TotalTasks int
}

// TaskModelSelectionDialog displays task information and allows model selection
type TaskModelSelectionDialog struct {
	BaseFocusableDialog
	Task              *taskmaster.Task
	TaskIndex         int
	TotalTasks        int
	AvailableModels   []TaskModelInfo
	SelectedIndex     int
	ViewPort          viewport.Model
	content           string
	result            *TaskModelInfo
	QueueTaskIDs      []string            // All task IDs in the queue
	ModelSelections   map[string]string   // taskID -> model mapping for all queued tasks
	GetTaskByID       func(string) *taskmaster.Task
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
			20,
			DialogKindCustom,
			len(models),
		),
		Task:            task,
		TaskIndex:       index,
		TotalTasks:      total,
		AvailableModels: models,
		SelectedIndex:   selectedIndex,
		result:          nil,
		QueueTaskIDs:    queueTaskIDs,
		ModelSelections: modelSelections,
		GetTaskByID:     getTaskByID,
	}

	// Initialize viewport
	dialog.ViewPort = viewport.New(76, 12)
	dialog.ViewPort.SetContent(dialog.renderContent())

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	return dialog
}

// loadTaskAvailableModels loads all available models from the model configuration
func loadTaskAvailableModels() []TaskModelInfo {
	availableModels := config.ListAvailableModels()
	var models []TaskModelInfo

	// Iterate through providers and models in a consistent order
	providers := []string{"anthropic", "openai", "google", "mistral", "perplexity"}

	for _, provider := range providers {
		if providerModels, exists := availableModels[provider]; exists {
			for _, model := range providerModels {
				models = append(models, TaskModelInfo{
					ID:       model.ModelID,
					Provider: model.Provider,
					Label:    fmt.Sprintf("%s • %s", model.ModelName, strings.ToTitle(model.Provider)),
				})
			}
		}
	}

	return models
}

// renderContent renders the dialog content including task info and model list
func (d *TaskModelSelectionDialog) renderContent() string {
	var sb strings.Builder

	// Task information section
	sb.WriteString(d.renderTaskInfo())
	sb.WriteString("\n\n")

	// Remaining tasks checklist (if available)
	if len(d.QueueTaskIDs) > 0 {
		sb.WriteString(d.renderRemainingTasksChecklist())
		sb.WriteString("\n")
	}

	sb.WriteString("Available Models:\n")
	sb.WriteString(strings.Repeat("─", 76))
	sb.WriteString("\n")

	// Model list
	for i, model := range d.AvailableModels {
		marker := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

		if i == d.SelectedIndex {
			marker = "▶ "
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Bold(true).
				Background(lipgloss.Color("63"))
		}

		line := fmt.Sprintf("%s%s", marker, model.Label)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderTaskInfo renders the task information at the top of the dialog
func (d *TaskModelSelectionDialog) renderTaskInfo() string {
	var sb strings.Builder

	// Task position
	positionText := fmt.Sprintf("Task %d of %d", d.TaskIndex+1, d.TotalTasks)
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(positionText))
	sb.WriteString("\n")

	// Task ID and title
	taskHeader := fmt.Sprintf("📋 #%s - %s", d.Task.ID, d.Task.Title)
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(taskHeader))
	sb.WriteString("\n")

	// Task description (truncated to fit)
	maxDescLen := 72
	description := d.Task.Description
	if len(description) > maxDescLen {
		description = description[:maxDescLen-3] + "..."
	}
	if description != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(description))
		sb.WriteString("\n")
	}

	// Task metadata
	var metadata []string

	// Complexity
	if d.Task.Complexity > 0 {
		complexityText := fmt.Sprintf("Complexity: %d", d.Task.Complexity)
		metadata = append(metadata, complexityText)
	}

	// Priority
	if d.Task.Priority != "" {
		priorityText := fmt.Sprintf("Priority: %s", d.Task.Priority)
		metadata = append(metadata, priorityText)
	}

	// Dependencies
	if len(d.Task.Dependencies) > 0 {
		depsText := fmt.Sprintf("Deps: %s", strings.Join(d.Task.Dependencies, ", "))
		metadata = append(metadata, depsText)
	}

	if len(metadata) > 0 {
		metadataLine := strings.Join(metadata, " • ")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true).Render(metadataLine))
	}

	return sb.String()
}

// renderRemainingTasksChecklist renders a checklist of remaining tasks in the queue
func (d *TaskModelSelectionDialog) renderRemainingTasksChecklist() string {
	// Only render if we have queue information
	if len(d.QueueTaskIDs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render("Queue Status:"))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 76))
	sb.WriteString("\n")

	for i, taskID := range d.QueueTaskIDs {
		// Determine status indicator
		var status string
		var color lipgloss.Color
		if i < d.TaskIndex {
			status = "✓"
			color = lipgloss.Color("40") // Green
		} else if i == d.TaskIndex {
			status = "●"
			color = lipgloss.Color("226") // Yellow
		} else {
			status = "○"
			color = lipgloss.Color("243") // Gray
		}

		// Get task details
		var taskTitle string
		if d.GetTaskByID != nil {
			task := d.GetTaskByID(taskID)
			if task != nil {
				taskTitle = task.Title
			}
		}

		// Get model selection status
		modelStatus := "⏳ awaiting"
		if model, ok := d.ModelSelections[taskID]; ok && model != "" {
			modelStatus = fmt.Sprintf("✓ %s", model)
		}

		// Truncate task title to fit on line
		maxTitleLen := 35
		if len(taskTitle) > maxTitleLen {
			taskTitle = taskTitle[:maxTitleLen-3] + "..."
		}

		// Build the line
		lineStyle := lipgloss.NewStyle().Foreground(color)
		line := fmt.Sprintf("%s %s", status, taskID)
		if taskTitle != "" {
			line += fmt.Sprintf(" - %s", taskTitle)
		}
		line = fmt.Sprintf("%-52s %s", line, modelStatus)

		sb.WriteString(lineStyle.Render(line))
		sb.WriteString("\n")
	}

	return sb.String()
}

// Init initializes the dialog
func (d *TaskModelSelectionDialog) Init() tea.Cmd {
	return nil
}

// Update processes messages for the dialog
func (d *TaskModelSelectionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		result, cmd := d.HandleKey(msg)
		if result != DialogResultNone {
			return d, cmd
		}
	}
	return d, nil
}

// HandleKey processes keyboard input
func (d *TaskModelSelectionDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if d.SelectedIndex > 0 {
			d.SelectedIndex--
			d.ViewPort.SetContent(d.renderContent())
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if d.SelectedIndex < len(d.AvailableModels)-1 {
			d.SelectedIndex++
			d.ViewPort.SetContent(d.renderContent())
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("home"))):
		d.SelectedIndex = 0
		d.ViewPort.SetContent(d.renderContent())
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("end"))):
		if len(d.AvailableModels) > 0 {
			d.SelectedIndex = len(d.AvailableModels) - 1
		}
		d.ViewPort.SetContent(d.renderContent())
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if d.SelectedIndex >= 0 && d.SelectedIndex < len(d.AvailableModels) {
			d.result = &d.AvailableModels[d.SelectedIndex]
			return DialogResultConfirm, nil
		}
		return DialogResultNone, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		return DialogResultCancel, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
		return DialogResultCancel, nil
	}

	return DialogResultNone, nil
}

// View renders the dialog
func (d *TaskModelSelectionDialog) View() string {
	// Update viewport content
	d.ViewPort.SetContent(d.renderContent())

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.Style.TitleColor).
		Padding(0, 1)

	title := titleStyle.Render(fmt.Sprintf("Select Model - Task %d/%d", d.TaskIndex+1, d.TotalTasks))

	// Border style
	borderStyle := lipgloss.NewStyle().
		Border(d.Style.Border).
		BorderForeground(d.Style.BorderColor)

	if d.IsFocused() {
		borderStyle = borderStyle.BorderForeground(d.Style.FocusedBorderColor)
	}

	// Content
	content := d.ViewPort.View()

	// Footer with hints
	footerText := ""
	for i, hint := range d.footerHints {
		if i > 0 {
			footerText += " • "
		}
		footerText += fmt.Sprintf("%s: %s", hint.Key, hint.Label)
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Padding(1, 1, 0, 1)

	footer := footerStyle.Render(footerText)

	// Assemble dialog
	dialogContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		footer,
	)

	dialog := borderStyle.Render(dialogContent)

	return dialog
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
		d.ViewPort.SetContent(d.renderContent())
	}
}
