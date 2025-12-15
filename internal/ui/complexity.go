package ui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

type complexityStreamClosedMsg struct{}

func dialogEnqueuedCmd() tea.Cmd {
	return func() tea.Msg { return nil }
}

// showComplexityScopeDialog displays the dialog for selecting complexity analysis scope
func (m *Model) showComplexityScopeDialog() {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	// Get the ID of the selected task
	selectedTaskID := ""
	if m.selectedTask != nil {
		selectedTaskID = m.selectedTask.ID
	}

	// Create the dialog
	scopeDialog, err := dialog.NewComplexityScopeDialog(selectedTaskID, dm.Style)
	if err != nil {
		m.logLines = append(m.logLines, fmt.Sprintf("Error creating complexity scope dialog: %s", err))
		return
	}

	// Show the dialog and handle the result
	m.appState.AddDialog(scopeDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		// If dialog was canceled
		if value == nil {
			return nil
		}

		// Process the result
		result, ok := value.(dialog.ComplexityScopeResult)
		if !ok {
			return func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("invalid result type from complexity scope dialog")}
			}
		}

		// Create a message with the selected scope
		return func() tea.Msg {
			return ComplexityScopeSelectedMsg{
				Scope:  result.Scope,
				TaskID: selectedTaskID,
				Tags:   result.Tags,
			}
		}
	})
}

// handleComplexityScopeSelected handles the selected complexity scope
func (m *Model) handleComplexityScopeSelected(msg ComplexityScopeSelectedMsg) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}
	// Prepare for analysis
	var totalTasks int

	// Determine the tasks to analyze based on scope
	switch msg.Scope {
	case "all":
		// All indexed tasks
		totalTasks = len(m.taskIndex)
	case "selected":
		// Just the selected task
		totalTasks = 1
	case "tag":
		// Count tasks with the selected tags
		for _, task := range m.taskIndex {
			for _, taskTag := range task.Tags {
				for _, selectedTag := range msg.Tags {
					if taskTag == selectedTag {
						totalTasks++
						break
					}
				}
			}
		}
	}

	m.currentComplexityScope = msg.Scope
	m.currentComplexityTags = append([]string(nil), msg.Tags...)

	// Create and show progress dialog
	progressDialog := dialog.NewComplexityProgressDialog(
		msg.Scope,
		msg.Tags,
		totalTasks,
		dm.Style,
	)

	// Add the dialog
	m.appState.AddDialog(progressDialog, func(value interface{}, err error) tea.Cmd {
		// This callback is triggered when the dialog is closed
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		// If cancelled, return nil
		if value == nil {
			return nil
		}

		return nil
	})

	m.complexityStartedAt = time.Now()

	// Kick off analysis work
	return m.startComplexityAnalysis(msg.Scope, msg.TaskID, msg.Tags, totalTasks)
}

// handleComplexityAnalysisProgress updates the progress dialog during analysis
func (m *Model) handleComplexityAnalysisProgress(msg ComplexityAnalysisProgressMsg) tea.Cmd {
	// Find the progress dialog
	if dm := m.dialogManager(); dm != nil {
		if progressDialog, ok := dm.GetDialogByType(dialog.DialogTypeProgress); ok {
			// Update progress
			if pd, ok := progressDialog.(*dialog.ProgressDialog); ok {
				update := dialog.ComplexityProgressUpdate{
					Progress:      msg.Progress,
					CurrentTask:   msg.CurrentTask,
					TasksAnalyzed: msg.TasksAnalyzed,
					TotalTasks:    msg.TotalTasks,
					Scope:         m.currentComplexityScope,
					Tags:          m.currentComplexityTags,
					Error:         msg.Error,
				}
				dialog.UpdateComplexityProgress(pd, update)
			}
		}
	}

	return nil
}

// handleComplexityAnalysisCompleted handles the completion of complexity analysis
func (m *Model) handleComplexityAnalysisCompleted(msg ComplexityAnalysisCompletedMsg) tea.Cmd {
	dm := m.dialogManager()
	// Close the progress dialog
	if dm != nil {
		dm.RemoveDialogsByType(dialog.DialogTypeProgress)
	}
	m.clearComplexityRuntimeState()

	// If there was an error, show it
	if msg.Error != nil {
		if errors.Is(msg.Error, context.Canceled) {
			m.ShowNotificationDialog("Analysis Cancelled", "Complexity analysis was cancelled.", "warning", 3*time.Second)
			return nil
		}
		m.ShowNotificationDialog("Analysis Failed", fmt.Sprintf("Error analyzing complexity: %s", msg.Error), "error", 5*time.Second)
		return nil
	}

	// If no tasks were analyzed or report is empty, show notification
	if msg.Report == nil || len(msg.Report.Tasks) == 0 {
		m.ShowNotificationDialog("Analysis Complete", "No tasks were found in the selected scope.", "warning", 3*time.Second)
		return nil
	}

	// Show the report dialog
	if dm == nil {
		return nil
	}
	reportDialog := dialog.NewComplexityReportDialog(msg.Report, dm.Style)
	reportDialog.SetSize(m.width, m.height)

	// Add the dialog
	m.appState.AddDialog(reportDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		// If closed without action, return nil
		if value == nil {
			return nil
		}

		// Process the result based on action type
		result, ok := value.(dialog.ComplexityReportResultMsg)
		if !ok {
			return nil
		}

		switch result.Action {
		case "select":
			// Jump to the selected task
			return func() tea.Msg {
				return SelectTaskMsg{TaskID: result.TaskID}
			}
		case "filter":
			// Show filter dialog
			return func() tea.Msg {
				return ComplexityReportActionMsg{Action: "filter"}
			}
		case "export":
			// Show export dialog
			return func() tea.Msg {
				return ComplexityReportActionMsg{Action: "export"}
			}
		}

		return dialogEnqueuedCmd()
	})

	return dialogEnqueuedCmd()
}

func (m *Model) startComplexityAnalysis(scope string, taskID string, tags []string, totalTasks int) tea.Cmd {
	if m.complexityCancel != nil {
		m.complexityCancel()
		m.complexityCancel = nil
	}

	progressCh := make(chan tea.Msg, 32)
	m.complexityMsgCh = progressCh

	ctx, cancel := context.WithCancel(context.Background())
	m.complexityCancel = cancel

	go func() {
		defer close(progressCh)
		report, err := m.taskService.AnalyzeComplexityWithProgress(ctx, scope, taskID, tags, func(state taskmaster.ComplexityProgressState) {
			progress := 0.0
			if state.TotalTasks > 0 {
				progress = float64(state.TasksAnalyzed) / float64(state.TotalTasks)
			}
			msg := ComplexityAnalysisProgressMsg{
				Progress:      progress,
				TasksAnalyzed: state.TasksAnalyzed,
				TotalTasks:    state.TotalTasks,
				CurrentTask:   state.CurrentTaskID,
			}
			select {
			case progressCh <- msg:
			case <-ctx.Done():
			}
		})

		select {
		case progressCh <- ComplexityAnalysisCompletedMsg{Report: report, Error: err}:
		case <-ctx.Done():
		}
	}()

	return m.waitForComplexityMessages()
}

func (m *Model) waitForComplexityMessages() tea.Cmd {
	ch := m.complexityMsgCh
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		if msg, ok := <-ch; ok {
			return msg
		}
		return complexityStreamClosedMsg{}
	}
}

func (m *Model) cancelComplexityAnalysis() {
	if m.complexityCancel != nil {
		m.complexityCancel()
		m.complexityCancel = nil
	}
}

func (m *Model) clearComplexityRuntimeState() {
	m.cancelComplexityAnalysis()
	m.complexityMsgCh = nil
	m.currentComplexityScope = ""
	m.currentComplexityTags = nil
	m.complexityStartedAt = time.Time{}
	m.waitingForComplexityHold = false
}

func (m *Model) handleProgressDialogCancel() tea.Cmd {
	if m.complexityCancel == nil {
		if m.parsePrdCancel != nil {
			m.cancelParsePrdJob()
			m.dismissActiveProgressDialog()
			m.addLogLine("PRD parsing cancelled by user")
		}
		return nil
	}

	m.cancelComplexityAnalysis()
	m.addLogLine("Complexity analysis cancelled by user")
	return nil
}

// handleComplexityReportAction handles actions from the complexity report dialog
func (m *Model) handleComplexityReportAction(msg ComplexityReportActionMsg) tea.Cmd {
	switch msg.Action {
	case "select":
		// Jump to the selected task via a dedicated message so regular selection logic can run
		if msg.TaskID == "" {
			return nil
		}
		return func() tea.Msg {
			return SelectTaskMsg{TaskID: msg.TaskID}
		}

	case "filter":
		// Show filter dialog
		// Find the report dialog to get current filter settings
		dm := m.dialogManager()
		if dm == nil {
			return nil
		}
		reportDialog, ok := dm.GetDialogByType(dialog.DialogTypeCustom)
		if !ok {
			return nil
		}

		complexityReport, ok := reportDialog.(*dialog.ComplexityReportDialog)
		if !ok {
			return nil
		}

		// Create and show filter dialog with current settings
		filterDialog, err := dialog.NewComplexityFilterDialog(
			complexityReport.FilterSettings,
			dm.Style,
		)
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		// Show dialog and handle result
		m.appState.AddDialog(filterDialog, func(value interface{}, err error) tea.Cmd {
			if err != nil {
				return func() tea.Msg {
					return ErrorMsg{Err: err}
				}
			}

			// If cancelled, return nil
			if value == nil {
				return nil
			}

			// Apply the filter settings
			return func() tea.Msg {
				return ComplexityFilterAppliedMsg{Settings: value}
			}
		})
		return dialogEnqueuedCmd()

	case "export":
		// Show export dialog
		dm := m.dialogManager()
		if dm == nil {
			return nil
		}
		exportDialog, err := dialog.NewComplexityExportDialog(dm.Style)
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		// Show dialog and handle result
		m.appState.AddDialog(exportDialog, func(value interface{}, err error) tea.Cmd {
			if err != nil {
				return func() tea.Msg {
					return ErrorMsg{Err: err}
				}
			}

			// If cancelled, return nil
			if value == nil {
				return nil
			}

			// Extract export settings
			settings, ok := value.(map[string]string)
			if !ok {
				return func() tea.Msg {
					return ErrorMsg{Err: fmt.Errorf("invalid export settings")}
				}
			}

			// Export the report
			return func() tea.Msg {
				return ComplexityExportRequestMsg{
					Format: settings["format"],
					Path:   settings["path"],
				}
			}
		})
		return dialogEnqueuedCmd()
	}

	return nil
}

// handleComplexityFilterApplied applies filter settings to the complexity report dialog
func (m *Model) handleComplexityFilterApplied(msg ComplexityFilterAppliedMsg) tea.Cmd {
	// Ensure the report dialog is still mounted before dispatching filter updates
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}
	if _, ok := dm.GetDialogByModel("ComplexityReportDialog"); !ok {
		return nil
	}

	// Send filter settings to the report dialog
	return func() tea.Msg {
		return dialog.DialogSetFilterMsg{Settings: msg.Settings}
	}
}

// handleComplexityExportRequest handles export requests from the complexity report dialog
func (m *Model) handleComplexityExportRequest(msg ComplexityExportRequestMsg) tea.Cmd {
	// Get the latest complexity report from the service
	report := m.taskService.GetLatestComplexityReport()
	if report == nil {
		return func() tea.Msg {
			return ComplexityExportCompletedMsg{
				FilePath: "",
				Error:    fmt.Errorf("no complexity report available for export"),
			}
		}
	}

	// Perform the export
	return func() tea.Msg {
		// Create context for export
		ctx := context.Background()

		// Export the report
		filePath, err := m.taskService.ExportComplexityReport(ctx, msg.Format, msg.Path)

		// Return completion message
		return ComplexityExportCompletedMsg{
			FilePath: filePath,
			Error:    err,
		}
	}
}

// handleComplexityExportCompleted handles export completion
func (m *Model) handleComplexityExportCompleted(msg ComplexityExportCompletedMsg) tea.Cmd {
	// Show notification based on result
	if msg.Error != nil {
		m.ShowNotificationDialog("Export Failed", fmt.Sprintf("Error exporting complexity report: %s", msg.Error), "error", 5*time.Second)
		return nil
	}

	m.ShowNotificationDialog("Export Successful", fmt.Sprintf("Complexity report exported to: %s", msg.FilePath), "success", 5*time.Second)
	return nil
}
