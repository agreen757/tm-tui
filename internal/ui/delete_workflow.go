package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	deleteConfirmDialogID = "delete_confirm_dialog"
	deleteOptionsDialogID = "delete_options_dialog"
	deleteWarningDialogID = "delete_warning_dialog"
	undoDialogID          = "delete_undo_dialog"
	maxConcurrentTasks    = 9 // Maximum number of tasks that can be run concurrently
)

// DeleteWorkflowState tracks the multi-step delete flow.
type DeleteWorkflowState struct {
	TaskIDs []string
	Options taskmaster.DeleteOptions
	Impact  *taskmaster.DeleteImpact
}

// UndoSession captures countdown state for undo dialogs.
type UndoSession struct {
	Token     *taskmaster.UndoToken
	Remaining time.Duration
}

func (m *Model) startDeleteWorkflow(taskIDs []string) {
	if len(taskIDs) == 0 {
		m.showErrorDialog("Delete Task", "Select at least one task before deleting.")
		return
	}
	m.deleteWorkflow = &DeleteWorkflowState{TaskIDs: taskIDs}
	m.openDeleteConfirmation(len(taskIDs))
}

func (m *Model) openDeleteConfirmation(count int) {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	message := fmt.Sprintf("Delete %d selected task(s)? You will be able to review impacted tasks before confirming.", count)
	content := dialog.NewSimpleModalContent(message)
	buttons := []dialog.ModalButton{
		{
			Kind:  dialog.ButtonYes,
			Label: "Continue",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultConfirm, func() tea.Msg {
					return dialog.DialogResultMsg{ID: deleteConfirmDialogID, Button: "continue"}
				}
			},
		},
		{
			Kind:  dialog.ButtonCancel,
			Label: "Cancel",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultCancel, func() tea.Msg {
					return dialog.DialogResultMsg{ID: deleteConfirmDialogID, Button: "cancel"}
				}
			},
		},
	}
	dlg := dialog.NewButtonModalDialog("Delete Tasks", 70, 9, content, buttons)
	dlg.ModalDialog.BaseDialog.ID = deleteConfirmDialogID
	m.appState.AddDialog(dlg, nil)
}

func (m *Model) openDeleteOptions() {
	dm := m.dialogManager()
	if dm == nil || m.deleteWorkflow == nil {
		return
	}
	fields := []dialog.FormField{
		{
			ID:      "recursive",
			Label:   "Recursive (delete subtasks and dependents)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: m.deleteWorkflow.Options.Recursive,
			Value:   m.deleteWorkflow.Options.Recursive,
		},
		{
			ID:      "force",
			Label:   "Force (skip minor validation errors)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: m.deleteWorkflow.Options.Force,
			Value:   m.deleteWorkflow.Options.Force,
		},
	}
	form := dialog.NewFormDialog(
		"Delete Options",
		"Choose how deletion should behave before continuing.",
		fields,
		[]string{"Review Impact", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Review Impact" {
				return nil, nil
			}
			return taskmaster.DeleteOptions{
				Recursive: boolValue(values, "recursive"),
				Force:     boolValue(values, "force"),
			}, nil
		},
	)
	form.BaseFocusableDialog.BaseDialog.ID = deleteOptionsDialogID
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Delete Options", err.Error())
			return nil
		}
		if value == nil {
			if m.deleteWorkflow != nil && m.deleteWorkflow.Impact == nil {
				m.deleteWorkflow = nil
			}
			return nil
		}
		if opts, ok := value.(taskmaster.DeleteOptions); ok {
			m.deleteWorkflow.Options = opts
			m.evaluateDeleteImpact()
		}
		return nil
	})
}

func (m *Model) evaluateDeleteImpact() {
	if m.taskService == nil || m.deleteWorkflow == nil {
		return
	}
	impact, err := m.taskService.AnalyzeDeleteImpact(m.deleteWorkflow.TaskIDs, m.deleteWorkflow.Options)
	if err != nil {
		m.showErrorDialog("Delete Task", err.Error())
		return
	}
	m.deleteWorkflow.Impact = impact
	m.openDeleteWarningDialog()
}

func (m *Model) openDeleteWarningDialog() {
	dm := m.dialogManager()
	if dm == nil || m.deleteWorkflow == nil || m.deleteWorkflow.Impact == nil {
		return
	}
	impact := m.deleteWorkflow.Impact
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Tasks to delete: %d\n", impact.TotalDeleteCount))
	if impact.BlockingReason != "" {
		b.WriteString(fmt.Sprintf("Blocking issue: %s\n", impact.BlockingReason))
	}
	if len(impact.WarningMessages) > 0 {
		for _, msg := range impact.WarningMessages {
			b.WriteString("- " + msg + "\n")
		}
	}
	content := dialog.NewSimpleModalContent(strings.TrimSpace(b.String()))
	buttons := []dialog.ModalButton{}
	if impact.BlockingReason != "" {
		buttons = append(buttons,
			dialog.ModalButton{
				Kind:  dialog.ButtonYes,
				Label: "Adjust Options",
				OnClick: func() (dialog.DialogResult, tea.Cmd) {
					return dialog.DialogResultConfirm, func() tea.Msg {
						return dialog.DialogResultMsg{ID: deleteWarningDialogID, Button: "adjust"}
					}
				},
			},
		)
	} else {
		buttons = append(buttons,
			dialog.ModalButton{
				Kind:  dialog.ButtonYes,
				Label: "Delete",
				OnClick: func() (dialog.DialogResult, tea.Cmd) {
					return dialog.DialogResultConfirm, func() tea.Msg {
						return dialog.DialogResultMsg{ID: deleteWarningDialogID, Button: "confirm"}
					}
				},
			},
		)
	}
	buttons = append(buttons, dialog.ModalButton{
		Kind:  dialog.ButtonCancel,
		Label: "Cancel",
		OnClick: func() (dialog.DialogResult, tea.Cmd) {
			return dialog.DialogResultCancel, func() tea.Msg {
				return dialog.DialogResultMsg{ID: deleteWarningDialogID, Button: "cancel"}
			}
		},
	})
	dlg := dialog.NewButtonModalDialog("Review Impact", 74, 12, content, buttons)
	dlg.ModalDialog.BaseDialog.ID = deleteWarningDialogID
	m.appState.AddDialog(dlg, nil)
}

func (m *Model) performDelete() tea.Cmd {
	if m.taskService == nil || m.deleteWorkflow == nil {
		return nil
	}
	ctx := context.Background()
	result, err := m.taskService.DeleteTasks(ctx, m.deleteWorkflow.TaskIDs, m.deleteWorkflow.Options)
	if err != nil {
		m.showErrorDialog("Delete Task", err.Error())
		return nil
	}
	m.addLogLine(fmt.Sprintf("Deleted %d task(s)", result.DeletedCount))
	for _, warning := range result.Warnings {
		m.addLogLine("Warning: " + warning)
	}

	// Clear active task if any of the deleted tasks is the active task
	if m.fileChangeTracker != nil {
		activeTask := m.fileChangeTracker.GetActiveTask()
		for _, deletedID := range m.deleteWorkflow.TaskIDs {
			if activeTask == deletedID {
				m.fileChangeTracker.SetActiveTask("")
				break
			}
		}
	}

	m.deleteWorkflow = nil
	if result.Undo != nil {
		return m.showUndoDialog(result.Undo)
	}
	return nil
}

func (m *Model) showUndoDialog(token *taskmaster.UndoToken) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil || token == nil {
		m.undoSession = nil
		return nil
	}
	content := newUndoContent(token)
	buttons := []dialog.ModalButton{
		{
			Kind:  dialog.ButtonYes,
			Label: "Undo",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultConfirm, func() tea.Msg {
					return dialog.DialogResultMsg{ID: undoDialogID, Button: "undo", Value: token}
				}
			},
		},
		{
			Kind:  dialog.ButtonCancel,
			Label: "Dismiss",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultCancel, func() tea.Msg {
					return dialog.DialogResultMsg{ID: undoDialogID, Button: "dismiss"}
				}
			},
		},
	}
	dlg := dialog.NewButtonModalDialog("Undo Delete", 68, 9, content, buttons)
	dlg.ModalDialog.BaseDialog.ID = undoDialogID
	m.appState.AddDialog(dlg, nil)
	m.undoSession = &UndoSession{Token: token, Remaining: token.Duration}
	return StartUndoCountdown(token.ID, token.ExpiresAt)
}

// undoContent renders live countdown text inside the undo dialog.
type undoContent struct {
	actionID  string
	summary   string
	remaining time.Duration
}

func newUndoContent(token *taskmaster.UndoToken) *undoContent {
	remaining := token.Duration
	if remaining <= 0 {
		remaining = time.Until(token.ExpiresAt)
	}
	return &undoContent{
		actionID:  token.ID,
		summary:   token.Summary,
		remaining: remaining,
	}
}

func (c *undoContent) Init() tea.Cmd { return nil }

func (c *undoContent) Update(msg tea.Msg) (dialog.ModalContent, tea.Cmd) {
	switch t := msg.(type) {
	case UndoTickMsg:
		if t.ActionID == c.actionID {
			c.remaining = t.Remaining
		}
	case UndoExpiredMsg:
		if t.ActionID == c.actionID {
			c.remaining = 0
		}
	}
	return c, nil
}

func (c *undoContent) View(width, height int) string {
	secs := int(c.remaining.Round(time.Second) / time.Second)
	if secs < 0 {
		secs = 0
	}
	text := fmt.Sprintf("%s\nUndo available for %d seconds", c.summary, secs)
	style := lipgloss.NewStyle().Width(width).Height(height).Align(lipgloss.Left)
	return style.Render(text)
}

func (c *undoContent) HandleKey(tea.KeyMsg) tea.Cmd { return nil }

func (m *Model) handleDialogResultMsg(msg dialog.DialogResultMsg) tea.Cmd {
	m.addLogLine(fmt.Sprintf("DEBUG: handleDialogResultMsg - ID: %s, Button: %s", msg.ID, msg.Button))
	switch msg.ID {
	case deleteConfirmDialogID:
		if msg.Button == "continue" {
			m.openDeleteOptions()
			return nil
		}
		m.deleteWorkflow = nil
	case deleteWarningDialogID:
		switch msg.Button {
		case "confirm":
			return m.performDelete()
		case "adjust":
			m.openDeleteOptions()
			return nil
		default:
			m.deleteWorkflow = nil
		}
	case undoDialogID:
		if msg.Button == "undo" {
			if token, ok := msg.Value.(*taskmaster.UndoToken); ok {
				return m.executeUndo(token.ID)
			}
			return m.executeUndo("")
		}
		m.undoSession = nil
	case "ready_tasks_dialog":
		m.addLogLine(fmt.Sprintf("DEBUG: ready_tasks_dialog case - Button: %s, Value type: %T", msg.Button, msg.Value))
		if msg.Button == "confirm" {
			// Get selected task IDs from the message value
			if selectedTasks, ok := msg.Value.([]string); ok && len(selectedTasks) > 0 {
				m.addLogLine(fmt.Sprintf("DEBUG: Selected tasks: %v", selectedTasks))
				// Validate the task selection (max 9 tasks, dependencies, etc.)
				if !m.validateReadyTasksExecution(selectedTasks) {
					// Validation failed - error dialog was shown
					m.addLogLine("DEBUG: Validation failed")
					m.appState.PopDialog()
					return nil
				}

				m.addLogLine("DEBUG: Validation passed, initializing execution queue")

				// Close the ready tasks dialog
				m.appState.PopDialog()

				// Create execution queue for multi-task execution
				m.executionQueue = &ExecutionQueue{
					TaskIDs:         selectedTasks,
					CurrentIndex:    0,
					ModelSelections: make(map[string]string),
					TaskStatus:      make(map[string]string),
				}

				// Initialize task statuses
				for _, id := range selectedTasks {
					m.executionQueue.TaskStatus[id] = "pending"
				}

				// Reset selection tracking
				m.taskModelSelectionDone = make(map[string]bool)

				m.addLogLine(fmt.Sprintf("Initialized execution queue with %d tasks, showing TaskModelSelectionDialog", len(selectedTasks)))

				// Show the TaskModelSelectionDialog for the first task
				return m.showTaskModelDialogCmd()
			} else {
				m.addLogLine(fmt.Sprintf("DEBUG: Type assertion failed or empty - ok: %v, len: %d", ok, len(selectedTasks)))
			}
		}
		// Cancellation or empty selection - just close the dialog
		m.addLogLine("DEBUG: Closing ready_tasks_dialog (cancel or empty)")
		m.readyTasksSelectionIDs = nil
		m.appState.PopDialog()
	case "model_selection_dialog":
		if msg.Button == "confirm" {
			// Get selected model from the message value
			if result, ok := msg.Value.(*dialog.ModelSelectionResult); ok {
				// Close the dialog
				m.appState.PopDialog()

				// Check if this model selection is for PRD creation
				if m.prdCreationPending {
					m.addLogLine("PRD creation detected, starting PRD generation")
					m.prdCreationPending = false
					// Execute Crush for PRD creation with selected model
					return m.executeCrushForPrd(result.Provider, result.ModelID)
				}

				// Check if this model selection is for an agent run
				if m.agentRunPending && m.agentRunTask != nil {
					taskID := m.agentRunTaskID
					taskTitle := m.agentRunTaskTitle
					task := m.agentRunTask

					// Clear context
					m.agentRunPending = false
					m.agentRunTaskID = ""
					m.agentRunTaskTitle = ""
					m.agentRunTask = nil

					// Start the agent run with selected agent type
					return m.startAgentRun(taskID, taskTitle, task, m.selectedAgentType, result.ModelID)
				}

				// Check if this model selection is for bulk ready tasks execution
				if len(m.readyTasksSelectionIDs) > 0 {
					taskIDs := m.readyTasksSelectionIDs
					m.readyTasksSelectionIDs = nil // Clear after capturing

					// Validate the model before execution
					if !m.validateModelExecution(result.ModelID) {
						// Validation failed - error dialog was shown
						return nil
					}

					m.addLogLine(fmt.Sprintf("Starting concurrent execution of %d tasks with model %s", len(taskIDs), result.ModelID))

					// Create model selections map with the same model for all tasks
					modelSelections := make(map[string]string)
					for _, taskID := range taskIDs {
						modelSelections[taskID] = result.ModelID
					}

					// Execute the selected tasks concurrently
					return m.executeMultipleTasks(taskIDs, modelSelections)
				}

				m.addLogLine(fmt.Sprintf("Model selected: %s (provider: %s)", result.ModelID, result.Provider))
				return nil
			}
		}
		// Cancellation - clear state and close the dialog
		m.readyTasksSelectionIDs = nil
		m.agentRunPending = false
		m.agentRunTaskID = ""
		m.agentRunTaskTitle = ""
		m.agentRunTask = nil
		m.prdCreationPending = false
		m.appState.PopDialog()
	case "task_model_selection_dialog":
		// Validate execution queue exists
		if m.executionQueue == nil {
			m.cleanupExecutionQueue()
			m.showErrorDialog("Model Selection", "Execution queue is not initialized")
			m.appState.PopDialog()
			return nil
		}

		if msg.Button == "confirm" {
			// Get current task ID from execution queue
			taskID := m.executionQueue.CurrentTask()
			if taskID == "" {
				m.cleanupExecutionQueue()
				m.showErrorDialog("Model Selection", "No current task in execution queue")
				m.appState.PopDialog()
				return nil
			}

			// Extract modelID from message value with type assertion
			modelID, ok := msg.Value.(string)
			if !ok {
				m.cleanupExecutionQueue()
				m.showErrorDialog("Model Selection", "Invalid model selection data")
				m.appState.PopDialog()
				return nil
			}

			// Store model selection in execution queue
			m.executionQueue.SetModelSelection(taskID, modelID)

			// Mark task as having model selection
			if m.taskModelSelectionDone == nil {
				m.taskModelSelectionDone = make(map[string]bool)
			}
			m.taskModelSelectionDone[taskID] = true

			// Update task status to "ready"
			m.executionQueue.TaskStatus[taskID] = "ready"

			// Close current dialog
			m.appState.PopDialog()

			// Check if there are more tasks to configure BEFORE advancing
			if m.executionQueue.HasNext() {
				// Move to next task in queue
				m.executionQueue.Next()
				// Show dialog for the next task
				return m.showTaskModelDialogCmd()
			} else {
				// All tasks have models - start execution
				// Collect all task IDs and their model selections
				var taskIDs []string
				for _, id := range m.executionQueue.TaskIDs {
					taskIDs = append(taskIDs, id)
				}

				// Pass the entire model selections map to executeMultipleTasks
				return m.executeMultipleTasks(taskIDs, m.executionQueue.ModelSelections)
			}
		} else if msg.Button == "skip" {
			// Get current task ID from execution queue
			taskID := m.executionQueue.CurrentTask()
			if taskID == "" {
				m.cleanupExecutionQueue()
				m.addLogLine("Error: No current task to skip in execution queue")
				m.appState.PopDialog()
				m.showErrorDialog("Skip Task", "No current task to skip")
				return nil
			}

			// Mark task status as "skipped"
			m.executionQueue.TaskStatus[taskID] = "skipped"

			// Skip this task (removes from queue)
			m.executionQueue.Skip()

			// Close current dialog
			m.appState.PopDialog()

			// Show next task dialog if there are more tasks
			if m.executionQueue.HasNext() {
				return m.showTaskModelDialogCmd()
			} else if m.executionQueue.RemainingCount() > 0 {
				// Execute remaining tasks (those that already have model selections)
				var taskIDs []string
				for _, id := range m.executionQueue.TaskIDs {
					if m.executionQueue.GetModelSelection(id) != "" {
						taskIDs = append(taskIDs, id)
					}
				}

				if len(taskIDs) > 0 {
					// Create model selections map with only selected tasks
					modelSelections := make(map[string]string)
					for _, taskID := range taskIDs {
						modelSelections[taskID] = m.executionQueue.GetModelSelection(taskID)
					}
					return m.executeMultipleTasks(taskIDs, modelSelections)
				}

				// No tasks with model selections
				m.cleanupExecutionQueue()
				m.addLogLine("No tasks selected for execution")
				return nil
			} else {
				// No tasks remain in queue
				m.cleanupExecutionQueue()
				m.addLogLine("No tasks selected for execution")
				return nil
			}
		} else if msg.Button == "cancel" {
			// Cancel entire queue - reset all state
			m.cleanupExecutionQueue()

			// Close dialog
			m.appState.PopDialog()

			// Return nil to halt flow
			return nil
		} else {
			// Unknown button value - log warning and close dialog with cleanup
			m.addLogLine(fmt.Sprintf("Warning: Unknown button '%s' in task model selection dialog", msg.Button))
			m.cleanupExecutionQueue()
			m.appState.PopDialog()
			return nil
		}
	case commandRunnerDialogID:
		if msg.Button == "Execute" {
			if result, ok := msg.Value.(dialog.CommandPromptResult); ok {
				return m.handleCommandRunnerSubmission(result)
			}
		}
		// Cancel button or other action - just close the dialog
	}
	return nil
}

func (m *Model) executeUndo(actionID string) tea.Cmd {
	if actionID == "" || m.taskService == nil {
		m.showErrorDialog("Undo Delete", "Undo action is no longer available.")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := m.taskService.UndoAction(ctx, actionID); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			m.showErrorDialog("Undo Delete", fmt.Sprintf("Undo operation timed out after 30 seconds: %v", err))
		} else {
			m.showErrorDialog("Undo Delete", err.Error())
		}
	} else {
		m.addLogLine("Undo completed successfully")
	}
	m.undoSession = nil
	return nil
}

func (m *Model) handleUndoTick(msg UndoTickMsg) tea.Cmd {
	if m.undoSession == nil || m.undoSession.Token == nil || m.undoSession.Token.ID != msg.ActionID {
		return nil
	}
	m.undoSession.Remaining = msg.Remaining
	return StartUndoCountdown(msg.ActionID, m.undoSession.Token.ExpiresAt)
}

func (m *Model) handleUndoExpired(msg UndoExpiredMsg) {
	if m.undoSession == nil || m.undoSession.Token == nil || m.undoSession.Token.ID != msg.ActionID {
		return
	}
	m.addLogLine("Undo window expired")
	m.undoSession = nil
	if m.appState != nil {
		m.appState.PopDialog()
	}
}

// validateTaskSelection validates the selected tasks for multi-task execution
// Returns error message if validation fails, empty string if valid
func (m *Model) validateTaskSelection(taskIDs []string) string {
	// Check for empty selection
	if len(taskIDs) == 0 {
		return "No tasks selected. Please select at least one task."
	}

	// Check for max task limit
	if len(taskIDs) > maxConcurrentTasks {
		return fmt.Sprintf("Too many tasks selected (%d). Maximum concurrent tasks is %d. Please select fewer tasks.", len(taskIDs), maxConcurrentTasks)
	}

	return ""
}

// validateModelSelection validates the selected model for task execution
// Returns error message if validation fails, empty string if valid
func (m *Model) validateModelSelection(modelID string) string {
	// Check for empty model ID
	if modelID == "" {
		return "No model selected. Please select a model before executing tasks."
	}

	return ""
}

// validateTaskDependencies validates that selected tasks don't have unmet dependencies
// Returns error message if validation fails, empty string if valid
func (m *Model) validateTaskDependencies(taskIDs []string) string {
	if m.taskService == nil {
		// Can't validate without task service
		return ""
	}

	taskSvc, ok := m.taskService.(*taskmaster.Service)
	if !ok {
		// Can't validate with non-standard service
		return ""
	}

	// Check each task's dependencies - they should either be done OR in the selection
	selectedIDMap := make(map[string]bool)
	for _, id := range taskIDs {
		selectedIDMap[id] = true
	}

	for _, taskID := range taskIDs {
		task, found := taskSvc.GetTaskByID(taskID)
		if !found || task == nil {
			continue
		}

		// Check if task has unmet dependencies
		for _, depID := range task.Dependencies {
			// Skip if dependency is in the selection (will be executed together)
			if selectedIDMap[depID] {
				continue
			}

			// Check if dependency is already done
			depTask, depFound := taskSvc.GetTaskByID(depID)
			if !depFound || depTask == nil {
				// Dependency not found - skip validation for this one
				continue
			}

			// If dependency is not done AND not in selection, that's an error
			if depTask.Status != taskmaster.StatusDone {
				return fmt.Sprintf("Task %s has an incomplete dependency: %s (status: %s). Complete the dependency first or add it to your selection.", taskID, depID, depTask.Status)
			}
		}
	}

	return ""
}

// validateReadyTasksExecution performs comprehensive validation before ready tasks execution
// Shows error dialog if validation fails
// Returns true if validation passed, false if validation failed
func (m *Model) validateReadyTasksExecution(taskIDs []string) bool {
	// Validate task selection (max 9 tasks)
	if errMsg := m.validateTaskSelection(taskIDs); errMsg != "" {
		m.showErrorDialog("Task Selection Validation", errMsg)
		return false
	}

	// Validate dependencies
	if errMsg := m.validateTaskDependencies(taskIDs); errMsg != "" {
		m.showErrorDialog("Dependency Validation", errMsg)
		return false
	}

	return true
}

// validateModelExecution validates the model before task execution
// Shows error dialog if validation fails
// Returns true if validation passed, false if validation failed
func (m *Model) validateModelExecution(modelID string) bool {
	// Validate model selection
	if errMsg := m.validateModelSelection(modelID); errMsg != "" {
		m.showErrorDialog("Model Validation", errMsg)
		return false
	}

	return true
}
