package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ComplexityProgressUpdate represents an update to the progress dialog
type ComplexityProgressUpdate struct {
	Progress      float64  // Progress as a percentage (0-100)
	CurrentTask   string   // Current task being processed
	TasksAnalyzed int      // Number of tasks analyzed so far
	TotalTasks    int      // Total number of tasks to analyze
	Scope         string   // Scope of analysis
	Tags          []string // Tags being analyzed (if scope is tag)
	Error         error    // Error, if any
}

// NewComplexityProgressDialog creates a progress dialog for tracking complexity analysis
func NewComplexityProgressDialog(
	scope string,
	tags []string,
	totalTasks int,
	style *DialogStyle,
) *ProgressDialog {
	pd := NewProgressDialog("Analyzing Task Complexity", 80, 10)
	if style != nil {
		ApplyStyleToDialog(pd, style)
	}
	pd.SetCancellable(true)
	pd.SetProgress(0)
	pd.SetLabel(progressDescription(scope, tags, 0, totalTasks))
	return pd
}

// UpdateComplexityProgress updates the progress dialog with current analysis status
func UpdateComplexityProgress(
	progressDialog *ProgressDialog,
	update ComplexityProgressUpdate,
) {
	progressDialog.SetProgress(update.Progress)
	description := progressDescription(
		update.Scope,
		update.Tags,
		update.TasksAnalyzed,
		update.TotalTasks,
	)

	if update.Error != nil {
		errorText := lipgloss.NewStyle().Foreground(progressDialog.Style.ErrorColor).
			Render(fmt.Sprintf("Error: %s", update.Error))
		description = fmt.Sprintf("%s\n\n%s", description, errorText)
	}

	if update.CurrentTask != "" {
		description = fmt.Sprintf("%s\n\nAnalyzing: %s", description, update.CurrentTask)
	}

	progressDialog.SetLabel(description)
}

// Helper function to format the progress description
func progressDescription(scope string, tags []string, tasksAnalyzed, totalTasks int) string {
	var scopeDescription string

	switch scope {
	case "all":
		scopeDescription = "Analyzing all tasks in project"
	case "selected":
		scopeDescription = "Analyzing selected task only"
	case "tag":
		if len(tags) > 0 {
			scopeDescription = fmt.Sprintf("Analyzing tasks with tag: %s", strings.Join(tags, ", "))
		} else {
			scopeDescription = "Analyzing tasks with specified tag"
		}
	default:
		scopeDescription = "Analyzing tasks"
	}

	var percent float64
	if totalTasks > 0 {
		percent = (float64(tasksAnalyzed) / float64(totalTasks)) * 100
	}

	return fmt.Sprintf(
		"%s\n\nProgress: %d/%d tasks analyzed (%.0f%%)",
		scopeDescription,
		tasksAnalyzed,
		totalTasks,
		percent,
	)
}
