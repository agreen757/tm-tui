package dialog

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TaskExecutionStatus represents the execution state of a task
type TaskExecutionStatus int

const (
	TaskRunning TaskExecutionStatus = iota
	TaskCompleted
	TaskFailed
	TaskCancelled
)

// String returns the string representation of TaskExecutionStatus
func (s TaskExecutionStatus) String() string {
	switch s {
	case TaskRunning:
		return "running"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	case TaskCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// TaskExecutionTab represents an individual task execution tab
type TaskExecutionTab struct {
	taskID    string
	taskTitle string
	status    TaskExecutionStatus
	output    []string
	viewport  viewport.Model
	startTime time.Time
	endTime   *time.Time
	model     string
	style     *DialogStyle
	// Cancellation support
	cmd          *exec.Cmd
	cancelFunc   context.CancelFunc
	cancelReason string // Optional reason for cancellation
}

// NewTaskExecutionTab creates a new task execution tab
func NewTaskExecutionTab(taskID, taskTitle, model string, width, height int, style *DialogStyle) *TaskExecutionTab {
	vp := viewport.New(width, height)
	vp.KeyMap = viewport.KeyMap{} // Disable default viewport keybindings

	if style == nil {
		style = DefaultDialogStyle()
	}

	return &TaskExecutionTab{
		taskID:    taskID,
		taskTitle: taskTitle,
		status:    TaskRunning,
		output:    []string{},
		viewport:  vp,
		startTime: time.Now(),
		endTime:   nil,
		model:     model,
		style:     style,
	}
}

// AddOutputLine appends a line to the output buffer and updates the viewport
func (t *TaskExecutionTab) AddOutputLine(line string) {
	t.output = append(t.output, line)
	t.updateViewportContent()
}

// SetStatus updates the task execution status
func (t *TaskExecutionTab) SetStatus(status TaskExecutionStatus) {
	t.status = status
	if status == TaskCompleted || status == TaskFailed || status == TaskCancelled {
		now := time.Now()
		t.endTime = &now
	}
}

// ElapsedTime returns the formatted elapsed time since task started
func (t *TaskExecutionTab) ElapsedTime() string {
	var elapsed time.Duration
	if t.endTime != nil {
		elapsed = t.endTime.Sub(t.startTime)
	} else {
		elapsed = time.Since(t.startTime)
	}

	if elapsed < time.Second {
		return fmt.Sprintf("%.0fms", elapsed.Seconds()*1000)
	}
	if elapsed < time.Minute {
		return fmt.Sprintf("%.1fs", elapsed.Seconds())
	}
	return fmt.Sprintf("%.1fm", elapsed.Minutes())
}

// GetTaskID returns the task ID
func (t *TaskExecutionTab) GetTaskID() string {
	return t.taskID
}

// GetTaskTitle returns the task title
func (t *TaskExecutionTab) GetTaskTitle() string {
	return t.taskTitle
}

// GetStatus returns the current status
func (t *TaskExecutionTab) GetStatus() TaskExecutionStatus {
	return t.status
}

// GetOutput returns the output buffer
func (t *TaskExecutionTab) GetOutput() []string {
	return t.output
}

// SetRect sets the dimensions of the tab's viewport
func (t *TaskExecutionTab) SetRect(width, height int) {
	t.viewport.Width = width
	t.viewport.Height = height
	t.updateViewportContent()
}

// Update handles messages and updates the tab state
func (t *TaskExecutionTab) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKeyMsg(msg)
	}
	return nil
}

// handleKeyMsg handles keyboard input for viewport scrolling
func (t *TaskExecutionTab) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up":
		t.viewport.LineUp(1)
	case "down":
		t.viewport.LineDown(1)
	case "pgup":
		t.viewport.HalfPageUp()
	case "pgdn":
		t.viewport.HalfPageDown()
	case "home":
		t.viewport.GotoTop()
	case "end":
		t.viewport.GotoBottom()
	}
	return nil
}

// updateViewportContent rebuilds the viewport content from the output buffer
func (t *TaskExecutionTab) updateViewportContent() {
	content := ""
	for i, line := range t.output {
		if i > 0 {
			content += "\n"
		}
		content += line
	}
	t.viewport.SetContent(content)
	// Auto-scroll to bottom
	t.viewport.GotoBottom()
}

// View renders the tab content with status and output
func (t *TaskExecutionTab) View() string {
	// Status indicator
	statusIcon := t.getStatusIcon()
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EEEEEE")).
		Bold(true)

	header := headerStyle.Render(fmt.Sprintf(
		"%s %s | Task: %s | Status: %s | Time: %s | Model: %s",
		statusIcon,
		t.taskTitle,
		t.taskID,
		t.status.String(),
		t.ElapsedTime(),
		t.model,
	))

	// Output viewport
	content := t.viewport.View()

	// Combine
	output := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.NewStyle().
			Height(1).
			Render(""),
		content,
	)

	return output
}

// getStatusIcon returns the appropriate icon for the current status
func (t *TaskExecutionTab) getStatusIcon() string {
	switch t.status {
	case TaskRunning:
		return "⏳"
	case TaskCompleted:
		return "✓"
	case TaskFailed:
		return "❌"
	case TaskCancelled:
		return "⊘"
	default:
		return "?"
	}
}

// SetCancellationContext sets the cmd and cancel function for this tab
// This allows the tab to be cancelled later
func (t *TaskExecutionTab) SetCancellationContext(cmd *exec.Cmd, cancelFunc context.CancelFunc) {
	t.cmd = cmd
	t.cancelFunc = cancelFunc
}

// CancelExecution attempts to cancel the running process
// Returns true if cancellation was initiated, false if nothing to cancel
func (t *TaskExecutionTab) CancelExecution(reason string) bool {
	// Only cancel if the task is still running
	if t.status != TaskRunning {
		return false
	}

	t.cancelReason = reason

	// Cancel the context
	if t.cancelFunc != nil {
		t.cancelFunc()
	}

	// Kill the process if it exists
	if t.cmd != nil && t.cmd.Process != nil {
		// Try to kill the process gracefully first, then forcefully if needed
		if err := t.cmd.Process.Kill(); err != nil {
			// Process may have already exited, which is fine
			return true
		}
	}

	return true
}

// GetCancellationReason returns the reason for cancellation, if any
func (t *TaskExecutionTab) GetCancellationReason() string {
	return t.cancelReason
}
