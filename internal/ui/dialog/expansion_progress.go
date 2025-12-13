package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ExpansionProgressUpdate represents an update to the expansion progress dialog
type ExpansionProgressUpdate struct {
	Progress        float64  // Progress as a percentage (0.0 to 1.0)
	Stage           string   // Current stage: "Analyzing", "Generating", "Applying", "Complete"
	CurrentTask     string   // Task ID being processed
	TasksExpanded   int      // Number of tasks expanded so far
	TotalTasks      int      // Total number of tasks to expand
	SubtasksCreated int      // Total subtasks created
	Scope           string   // Scope of expansion: "single", "all", "range", "tag"
	Message         string   // Status message from CLI
	Error           error    // Error, if any
}

// NewExpansionProgressDialog creates a progress dialog for tracking task expansion
func NewExpansionProgressDialog(
	scope string,
	totalTasks int,
	style *DialogStyle,
) *ProgressDialog {
	pd := NewProgressDialog("Expanding Tasks", 80, 10)
	if style != nil {
		ApplyStyleToDialog(pd, style)
	}
	pd.SetCancellable(true)
	pd.SetProgress(0)
	pd.SetLabel(expansionProgressDescription(scope, "", 0, totalTasks, 0))
	return pd
}

// UpdateExpansionProgress updates the progress dialog with current expansion status
func UpdateExpansionProgress(
	progressDialog *ProgressDialog,
	update ExpansionProgressUpdate,
) {
	progressDialog.SetProgress(update.Progress)
	
	description := expansionProgressDescription(
		update.Scope,
		update.Stage,
		update.TasksExpanded,
		update.TotalTasks,
		update.SubtasksCreated,
	)

	// Add error styling if present
	if update.Error != nil {
		errorText := lipgloss.NewStyle().Foreground(progressDialog.Style.ErrorColor).
			Render(fmt.Sprintf("Error: %s", update.Error))
		description = fmt.Sprintf("%s\n\n%s", description, errorText)
	}

	// Add current task information if available
	if update.CurrentTask != "" {
		description = fmt.Sprintf("%s\n\nCurrent task: %s", description, update.CurrentTask)
	}

	// Add custom message if provided and not redundant
	if update.Message != "" && !strings.Contains(description, update.Message) {
		// Filter out raw file paths and CLI noise
		if !strings.Contains(update.Message, "/.taskmaster/") && 
		   !strings.HasPrefix(update.Message, "/Users/") &&
		   !strings.HasPrefix(update.Message, "/home/") &&
		   len(update.Message) < 200 {
			description = fmt.Sprintf("%s\n\n%s", description, update.Message)
		}
	}

	progressDialog.SetLabel(description)
}

// expansionProgressDescription formats the main progress description
func expansionProgressDescription(
	scope string,
	stage string,
	tasksExpanded int,
	totalTasks int,
	subtasksCreated int,
) string {
	var scopeDescription string

	// Build scope description
	switch scope {
	case "all":
		scopeDescription = "Expanding all tasks in project"
	case "single":
		scopeDescription = "Expanding selected task"
	case "range":
		scopeDescription = "Expanding task range"
	case "tag":
		scopeDescription = "Expanding tasks by tag"
	default:
		scopeDescription = "Expanding tasks"
	}

	// Add stage information if available
	if stage != "" && stage != "Processing" {
		scopeDescription = fmt.Sprintf("%s - %s", scopeDescription, stage)
	}

	// Calculate percentage
	var percent float64
	if totalTasks > 0 && tasksExpanded > 0 {
		percent = (float64(tasksExpanded) / float64(totalTasks)) * 100
	}

	// Build progress details
	var progressLine string
	if totalTasks > 0 {
		progressLine = fmt.Sprintf("Progress: %d/%d tasks expanded (%.0f%%)",
			tasksExpanded,
			totalTasks,
			percent,
		)
	} else {
		progressLine = "Initializing expansion..."
	}

	// Add subtasks information if available
	if subtasksCreated > 0 {
		progressLine = fmt.Sprintf("%s\nSubtasks created: %d", progressLine, subtasksCreated)
	}

	return fmt.Sprintf("%s\n\n%s", scopeDescription, progressLine)
}
