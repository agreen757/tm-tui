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

// showExpansionScopeDialog displays the dialog for selecting expansion scope and options
func (m *Model) showExpansionScopeDialog() {
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
	scopeDialog, err := dialog.NewExpansionScopeDialog(selectedTaskID, dm.Style)
	if err != nil {
		m.logLines = append(m.logLines, fmt.Sprintf("Error creating expansion scope dialog: %s", err))
		return
	}

	// Show the dialog and handle the result
	m.appState.AddDialog(scopeDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		if value == nil {
			return nil
		}

		result, ok := value.(dialog.ExpansionScopeResult)
		if !ok {
			return func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("invalid result type from expansion scope dialog")}
			}
		}

		// Create a message with the selected scope
		return func() tea.Msg {
			return ExpansionScopeSelectedMsg{
				Scope:       result.Scope,
				TaskID:      result.TaskID,
				FromID:      result.FromID,
				ToID:        result.ToID,
				Tags:        result.Tags,
				Depth:       result.Depth,
				NumSubtasks: result.NumSubtasks,
				UseAI:       result.UseAI,
			}
		}
	})
}

// handleExpansionScopeSelected handles the selected expansion scope
func (m *Model) handleExpansionScopeSelected(msg ExpansionScopeSelectedMsg) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	// Determine total tasks to expand
	var totalTasks int
	switch msg.Scope {
	case "single":
		totalTasks = 1
	case "all":
		totalTasks = len(m.taskIndex)
	case "range":
		// Count tasks in range
		for id := range m.taskIndex {
			if (msg.FromID == "" || id >= msg.FromID) && (msg.ToID == "" || id <= msg.ToID) {
				totalTasks++
			}
		}
	case "tag":
		// Count tasks with matching tags
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

	m.currentExpansionScope = msg.Scope
	m.currentExpansionTags = append([]string(nil), msg.Tags...)

	// Create and show progress dialog using the expansion progress helper
	progressDialog := dialog.NewExpansionProgressDialog(msg.Scope, totalTasks, dm.Style)

	m.appState.AddDialog(progressDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			return func() tea.Msg {
				return ErrorMsg{Err: err}
			}
		}

		// Handle cancellation
		if value == nil {
			m.cancelExpansion()
			return nil
		}

		return nil
	})

	m.expansionStartedAt = time.Now()

	// Start expansion work
	return m.startExpansion(msg.Scope, msg.TaskID, msg.FromID, msg.ToID, msg.Tags, taskmaster.ExpandTaskOptions{
		NumSubtasks: msg.NumSubtasks,
		UseAI:       msg.UseAI,
		Force:       false,
	})
}

// startExpansion initiates the expansion process
func (m *Model) startExpansion(scope, taskID, fromID, toID string, tags []string, opts taskmaster.ExpandTaskOptions) tea.Cmd {
	if m.expansionCancel != nil {
		m.expansionCancel()
		m.expansionCancel = nil
	}

	progressCh := make(chan tea.Msg, 32)
	m.expansionMsgCh = progressCh

	ctx, cancel := context.WithCancel(context.Background())
	m.expansionCancel = cancel

	go func() {
		defer close(progressCh)

		err := m.taskService.ExecuteExpandWithProgress(
			ctx,
			scope,
			taskID,
			fromID,
			toID,
			tags,
			opts,
			func(state taskmaster.ExpandProgressState) {
				msg := ExpansionProgressMsg{
					Progress:        state.Progress,
					Stage:           state.Stage,
					CurrentTask:     state.CurrentTask,
					TasksExpanded:   state.TasksExpanded,
					TotalTasks:      state.TotalTasks,
					SubtasksCreated: state.SubtasksCreated,
					Message:         state.Message,
				}
				select {
				case progressCh <- msg:
				case <-ctx.Done():
				}
			},
		)

		// Send completion message
		completionMsg := ExpansionCompletedMsg{Error: err}
		if err == nil {
			// Extract stats from final state if available
			// For now, we'll get this from the reload
			completionMsg.TasksExpanded = 1 // Placeholder
		}

		select {
		case progressCh <- completionMsg:
		case <-ctx.Done():
		}
	}()

	return m.waitForExpansionMessages()
}

// handleExpansionProgress updates the progress dialog during expansion
func (m *Model) handleExpansionProgress(msg ExpansionProgressMsg) tea.Cmd {
	// Find the progress dialog
	if dm := m.dialogManager(); dm != nil {
		if progressDialog, ok := dm.GetDialogByType(dialog.DialogTypeProgress); ok {
			if pd, ok := progressDialog.(*dialog.ProgressDialog); ok {
				// Create update struct and use the formatting helper
				update := dialog.ExpansionProgressUpdate{
					Progress:        msg.Progress,
					Stage:           msg.Stage,
					CurrentTask:     msg.CurrentTask,
					TasksExpanded:   msg.TasksExpanded,
					TotalTasks:      msg.TotalTasks,
					SubtasksCreated: msg.SubtasksCreated,
					Scope:           m.currentExpansionScope,
					Message:         msg.Message,
				}
				dialog.UpdateExpansionProgress(pd, update)
			}
		}
	}

	return m.waitForExpansionMessages()
}

// handleExpansionCompleted handles completion of expansion
func (m *Model) handleExpansionCompleted(msg ExpansionCompletedMsg) tea.Cmd {
	dm := m.dialogManager()

	// Close the progress dialog
	if dm != nil {
		dm.RemoveDialogsByType(dialog.DialogTypeProgress)
	}

	m.clearExpansionRuntimeState()

	// Handle errors
	if msg.Error != nil {
		if errors.Is(msg.Error, context.Canceled) {
			m.ShowNotificationDialog("Expansion Cancelled", "Task expansion was cancelled.", "warning", 3*time.Second)
			return nil
		}
		m.ShowNotificationDialog("Expansion Failed", fmt.Sprintf("Error expanding tasks: %s", msg.Error), "error", 5*time.Second)
		return nil
	}

	// Success notification
	duration := time.Since(m.expansionStartedAt)
	message := fmt.Sprintf("Successfully expanded tasks in %s", duration.Round(time.Millisecond))
	m.ShowNotificationDialog("Expansion Complete", message, "success", 3*time.Second)

	// Reload tasks to show new subtasks
	return LoadTasksCmd(m.taskService)
}

// waitForExpansionMessages waits for messages from the expansion channel
func (m *Model) waitForExpansionMessages() tea.Cmd {
	ch := m.expansionMsgCh
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		if msg, ok := <-ch; ok {
			return msg
		}
		return expansionStreamClosedMsg{}
	}
}

// cancelExpansion cancels the ongoing expansion
func (m *Model) cancelExpansion() {
	if m.expansionCancel != nil {
		m.expansionCancel()
		m.expansionCancel = nil
	}
}

// clearExpansionRuntimeState clears all expansion-related runtime state
func (m *Model) clearExpansionRuntimeState() {
	m.cancelExpansion()
	m.expansionMsgCh = nil
	m.currentExpansionScope = ""
	m.currentExpansionTags = nil
	m.expansionStartedAt = time.Time{}
	m.waitingForExpansionHold = false
}
