package ui

import (
	"context"
	"time"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/executor"
	"github.com/adriangreen/tm-tui/internal/projects"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

// TasksLoadedMsg is sent when tasks are initially loaded
type TasksLoadedMsg struct {
	Tasks []taskmaster.Task
}

// TasksReloadedMsg is sent when tasks.json has been reloaded from disk
type TasksReloadedMsg struct{}

// ConfigReloadedMsg is sent when config files have been reloaded from disk
type ConfigReloadedMsg struct{}

// WatcherErrorMsg is sent when a file watcher encounters an error
type WatcherErrorMsg struct {
	Err error
}

// ExecutorOutputMsg is sent when the executor produces output
type ExecutorOutputMsg struct {
	Line string
}

// CommandCompletedMsg is sent when a command execution completes
type CommandCompletedMsg struct {
	Command string
	Success bool
	Output  string
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// WaitForTasksReload returns a command that waits for tasks to be reloaded
// and sends a TasksReloadedMsg when that happens
func WaitForTasksReload(service TaskService) tea.Cmd {
	return func() tea.Msg {
		<-service.ReloadEvents()
		return TasksReloadedMsg{}
	}
}

// WaitForConfigReload returns a command that waits for config to be reloaded
// and sends a ConfigReloadedMsg when that happens
func WaitForConfigReload(manager *config.ConfigManager) tea.Cmd {
	return func() tea.Msg {
		<-manager.ReloadEvents()
		return ConfigReloadedMsg{}
	}
}

// WaitForExecutorOutput returns a command that listens for executor output
func WaitForExecutorOutput(service *executor.Service) tea.Cmd {
	return func() tea.Msg {
		// Get the output channel and wait for a line
		outputChan := service.GetOutput()
		line := <-outputChan
		return ExecutorOutputMsg{Line: line}
	}
}

// LoadTasksCmd loads tasks from disk and returns a TasksLoadedMsg
func LoadTasksCmd(service TaskService) tea.Cmd {
	return func() tea.Msg {
		// Force load tasks initially
		ctx := context.WithValue(context.Background(), "force", true)
		if err := service.LoadTasks(ctx); err != nil {
			return ErrorMsg{Err: err}
		}

		tasks, _ := service.GetTasks()

		return TasksLoadedMsg{Tasks: tasks}
	}
}

// AnalyzeTaskComplexityMsg is sent when the user requests complexity analysis
type AnalyzeTaskComplexityMsg struct{}

// SelectTaskMsg requests that the UI jump to a specific task ID.
type SelectTaskMsg struct {
	TaskID string
}

// tagListLoadedMsg delivers the parsed tag contexts for management dialogs.
type tagListLoadedMsg struct {
	List *taskmaster.TagList
	Err  error
}

// TagOperationMsg reports the outcome of a CLI-driven tag command.
type TagOperationMsg struct {
	Operation string
	TagName   string
	Result    *taskmaster.TagOperationResult
	Err       error
}

type projectSwitchedMsg struct {
	Meta *projects.Metadata
	Err  error
	Tag  string
}

// ComplexityScopeSelectedMsg is sent when a scope is selected for analysis
type ComplexityScopeSelectedMsg struct {
	Scope  string   // "all", "selected", "tag"
	TaskID string   // Only used when scope is "selected"
	Tags   []string // Only used when scope is "tag"
}

// ComplexityAnalysisProgressMsg is sent during analysis to update progress
type ComplexityAnalysisProgressMsg struct {
	Progress      float64 // Progress percentage (0-100)
	TasksAnalyzed int     // Number of tasks analyzed
	TotalTasks    int     // Total tasks to analyze
	CurrentTask   string  // ID of current task being analyzed
	Error         error   // Error, if any
}

// ComplexityAnalysisCompletedMsg is sent when analysis is complete
type ComplexityAnalysisCompletedMsg struct {
	Report *taskmaster.ComplexityReport
	Error  error
}

// ComplexityReportActionMsg is sent when the user interacts with the report
type ComplexityReportActionMsg struct {
	Action string // "select", "filter", "export", "close"
	TaskID string // Only set when Action is "select"
}

// ComplexityFilterAppliedMsg is sent when filter settings are applied
type ComplexityFilterAppliedMsg struct {
	Settings interface{} // FilterSettings from dialog.complexity_report.go
}

// ComplexityExportRequestMsg is sent when export is requested
type ComplexityExportRequestMsg struct {
	Format string // "json", "csv"
	Path   string // Output path
}

// ComplexityExportCompletedMsg is sent when export is complete
type ComplexityExportCompletedMsg struct {
	FilePath string
	Error    error
}

// UndoTickMsg updates the countdown timer shown in the undo dialog.
type UndoTickMsg struct {
	ActionID  string
	Remaining time.Duration
}

// UndoExpiredMsg signals that the undo window has elapsed.
type UndoExpiredMsg struct {
	ActionID string
}

// DEPRECATED: Legacy expansion messages - being replaced by new CLI-based flow
// ExpandTaskProgressMsg is sent during task expansion to update progress
type ExpandTaskProgressMsg struct {
	Stage    string
	Progress float64
	Error    error
}

// DEPRECATED: Legacy expansion messages - being replaced by new CLI-based flow
// ExpandTaskCompletedMsg is sent when task expansion is complete
type ExpandTaskCompletedMsg struct {
	Error error
}

type expandTaskStreamClosedMsg struct{}

// DEPRECATED: Legacy expansion messages - being replaced by new CLI-based flow
// ExpandTaskDraftsGeneratedMsg is sent when subtask drafts are ready for preview
type ExpandTaskDraftsGeneratedMsg struct {
	Drafts   []taskmaster.SubtaskDraft
	ParentID string
}

// DEPRECATED: Legacy expansion messages - being replaced by new CLI-based flow
// ExpandTaskDraftsConfirmedMsg is sent when the user confirms the final drafts
type ExpandTaskDraftsConfirmedMsg struct {
	Drafts   []taskmaster.SubtaskDraft
	ParentID string
}

// ExpansionScopeSelectedMsg is sent when expansion scope is selected
type ExpansionScopeSelectedMsg struct {
	Scope       string   // "single", "all", "range", "tag"
	TaskID      string   // for single task expansion
	FromID      string   // for range expansion
	ToID        string   // for range expansion
	Tags        []string // for tag-based expansion
	Depth       int      // 1-3 levels
	NumSubtasks int      // optional, 0 = auto
	UseAI       bool     // --research flag
}

// ExpansionProgressMsg is sent during CLI expansion to update progress
type ExpansionProgressMsg struct {
	Progress        float64
	Stage           string
	CurrentTask     string
	TasksExpanded   int
	TotalTasks      int
	SubtasksCreated int
	Message         string
	Error           error
}

// ExpansionCompletedMsg is sent when CLI expansion is complete
type ExpansionCompletedMsg struct {
	TasksExpanded   int
	SubtasksCreated int
	Error           error
}

type expansionStreamClosedMsg struct{}

// AnalyzeTaskComplexityCmd starts the complexity analysis process and returns an AnalyzeTaskComplexityMsg
func AnalyzeTaskComplexityCmd() tea.Cmd {
	return func() tea.Msg {
		return AnalyzeTaskComplexityMsg{}
	}
}

// StartUndoCountdown schedules countdown updates for undo dialogs.
func StartUndoCountdown(actionID string, expiresAt time.Time) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		remaining := time.Until(expiresAt)
		if remaining <= 0 {
			return UndoExpiredMsg{ActionID: actionID}
		}
		return UndoTickMsg{ActionID: actionID, Remaining: remaining}
	})
}
