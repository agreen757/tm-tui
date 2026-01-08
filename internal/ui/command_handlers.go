package ui

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	commandRunnerDialogID = "command_runner_dialog"
)

type addTagFormResult struct {
	Name            string
	CopyFromCurrent bool
	CopyFrom        string
	Description     string
}

type deleteTagFormResult struct {
	Name string
	Yes  bool
}

type useTagFormResult struct {
	Name string
}

// dispatchCommand routes a command to its handler.
func (m *Model) dispatchCommand(id CommandID) tea.Cmd {
	switch id {
	case CommandParsePRD:
		return m.openParsePrdWorkflow()
	case CommandCreatePRD:
		return m.openCreatePrdWorkflow()
	case CommandAnalyzeComplexity:
		m.showComplexityScopeDialog()
	case CommandExpandTask:
		return m.handleExpandTaskCommand()
	case CommandDeleteTask:
		m.handleDeleteTaskCommand()
	case CommandRunTask:
		return m.handleRunTaskCommand()
	case CommandRunCommand:
		return m.openCommandRunner()
	case CommandManageTags:
		m.openAddTagDialog()
	case CommandTagManagement:
		return m.handleTagManagement()
	case CommandUseTag:
		return m.openUseTagDialog()
	case CommandProjectTags:
		return m.openProjectTagsDialog()
	case CommandProjectQuickSwitch:
		return m.openQuickProjectSwitchDialog()
	case CommandProjectSearch:
		return m.openProjectSearchDialog()
	case CommandGitMenu:
		m.openGitMenu()
	case CommandGitSwitchBranch:
		m.openBranchSwitchDialog()
	case CommandGitCreateBranch:
		m.openBranchCreateDialog()
	case CommandGitRecentCommits:
		m.openCommitsDialog()
	default:
		m.addLogLine(fmt.Sprintf("Command %s not implemented", id))
	}
	return nil
}

func (m *Model) handleExpandTaskCommand() tea.Cmd {
	if m.taskService == nil || !m.taskService.IsAvailable() {
		appErr := NewDependencyError("Expand Task", "Task Master CLI is not available in this workspace.", nil).
			WithRecoveryHints(
				"Restart the TUI inside a Task Master workspace",
				"Check Task Master CLI installation",
			)
		m.showAppError(appErr)
		return nil
	}

	// Show scope dialog (new approach)
	m.showExpansionScopeDialog()
	return nil
}

func (m *Model) handleRunTaskCommand() tea.Cmd {
	// Log that the command was triggered
	m.addLogLine("Alt+R pressed - Run Task with Crush")
	
	// Check if a task is selected
	if m.selectedTask == nil {
		m.addLogLine("ERROR: No task selected")
		appErr := NewValidationError("Run Task", "No task selected. Please select a task to run with Crush.", nil).
			WithRecoveryHints(
				"Use arrow keys to navigate and select a task",
				"Press Alt+R again after selecting a task",
			)
		m.showAppError(appErr)
		return nil
	}

	// Pass the entire selected task object to ensure consistency
	m.addLogLine(fmt.Sprintf("Opening model selection for task %s: %s", m.selectedTask.ID, m.selectedTask.Title))

	// Open model selection dialog
	// When model is selected, it will trigger the Crush run via the model selection callback
	return m.openModelSelectionForCrushRun(m.selectedTask)
}

// DEPRECATED: Replaced by ExpansionScopeDialog
// This function is kept for backward compatibility only.
func (m *Model) openTaskSelectionDialog() tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	items := make([]dialog.ListItem, 0, len(m.taskIndex))
	for _, task := range m.sortedTasksForSelection() {
		items = append(items, newTaskListItem(task))
	}
	if len(items) == 0 {
		appErr := NewValidationError("Expand Task", "No tasks available to expand.", nil).
			WithRecoveryHints(
				"Create tasks first using the Add Task feature",
				"Load tasks from a PRD file",
			)
		m.showAppError(appErr)
		return nil
	}

	list := dialog.NewListDialog("Select Task to Expand", 72, 20, items)
	list.SetShowDescription(true)
	list.EnableFiltering("Filter tasks (/)")
	list.SetFooterHints(
		dialog.ShortcutHint{Key: "/", Label: "Search"},
		dialog.ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		dialog.ShortcutHint{Key: "Enter", Label: "Choose"},
	)
	if dm.Style != nil {
		dialog.ApplyStyleToDialog(list, dm.Style)
	}

	m.appState.AddDialog(list, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Expand Task", err.Error())
			return nil
		}
		msg, ok := value.(dialog.ListSelectionMsg)
		if !ok || msg.SelectedItem == nil {
			return nil
		}
		if taskItem, ok := msg.SelectedItem.(*taskListItem); ok {
			return m.startExpandWorkflow(taskItem.id)
		}
		return nil
	})

	return list.Init()
}

type expandOptions struct {
	TaskID string
	Depth  int
	Num    int
	UseAI  bool
}

func (m *Model) sortedTasksForSelection() []*taskmaster.Task {
	if len(m.taskIndex) == 0 {
		return nil
	}
	ids := make([]string, 0, len(m.taskIndex))
	for id := range m.taskIndex {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	tasks := make([]*taskmaster.Task, 0, len(ids))
	for _, id := range ids {
		if task, ok := m.taskIndex[id]; ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

type taskListItem struct {
	id          string
	title       string
	description string
}

func newTaskListItem(task *taskmaster.Task) *taskListItem {
	if task == nil {
		return &taskListItem{}
	}
	desc := fmt.Sprintf("ID %s · %d subtasks", task.ID, len(task.Subtasks))
	return &taskListItem{id: task.ID, title: task.Title, description: desc}
}

func (i *taskListItem) Title() string       { return i.title }
func (i *taskListItem) Description() string { return i.description }
func (i *taskListItem) FilterValue() string { return i.title }

// DEPRECATED: Replaced by handleExpansionScopeSelected
// This function is kept for backward compatibility only.
func (m *Model) startExpandWorkflow(taskID string) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	fields := []dialog.FormField{
		{
			ID:    "depth",
			Label: "Expansion depth",
			Type:  dialog.FormFieldTypeRadio,
			Options: []dialog.FormOption{
				{Value: "1", Label: "1 level"},
				{Value: "2", Label: "2 levels"},
				{Value: "3", Label: "3 levels"},
			},
			Value: "2",
		},
		{
			ID:          "num",
			Label:       "Number of subtasks",
			Type:        dialog.FormFieldTypeText,
			Placeholder: "Leave blank for auto",
		},
		{
			ID:      "ai",
			Label:   "Enable AI assistance (--research)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: true,
		},
	}

	form := dialog.NewFormDialog(
		"Expand Task Options",
		fmt.Sprintf("Configure expansion for task %s", taskID),
		fields,
		[]string{"Expand", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Expand" {
				return nil, nil
			}
			depth := parseDepthValue(stringValue(values, "depth"), 2)
			if depth < 1 {
				depth = 1
			} else if depth > 3 {
				depth = 3
			}
			num := parseOptionalInt(stringValue(values, "num"))
			if num < 0 {
				num = 0
			}
			return expandOptions{
				TaskID: taskID,
				Depth:  depth,
				Num:    num,
				UseAI:  boolValue(values, "ai"),
			}, nil
		},
	)

	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Expand Task", err.Error())
			return nil
		}
		opts, ok := value.(expandOptions)
		if !ok || opts.TaskID == "" {
			return nil
		}
		return m.runExpandTask(opts)
	})

	return form.Init()
}

func parseDepthValue(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	if v, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		return v
	}
	return fallback
}

func parseOptionalInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return 0
}

func (m *Model) handleDeleteTaskCommand() {
	ids := m.selectedOrCurrentTaskIDs()
	if len(ids) == 0 {
		appErr := NewValidationError("Delete Task", "Select at least one task before deleting.", nil).
			WithRecoveryHints(
				"Use arrow keys to navigate and space to select tasks",
				"Or click on a task to select it",
			)
		m.showAppError(appErr)
		return
	}
	m.startDeleteWorkflow(ids)
}

func (m *Model) openAddTagDialog() {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	fields := []dialog.FormField{
		{
			ID:       "name",
			Label:    "Tag name",
			Type:     dialog.FormFieldTypeText,
			Required: true,
		},
		{
			ID:      "copyCurrent",
			Label:   "Copy from current context (--copy-from-current)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: false,
		},
		{
			ID:    "copyFrom",
			Label: "Copy from tag (--copy-from=<tag>)",
			Type:  dialog.FormFieldTypeText,
		},
		{
			ID:    "description",
			Label: "Description (-d=<desc>)",
			Type:  dialog.FormFieldTypeText,
		},
	}
	form := dialog.NewFormDialog(
		"Add Tag Context",
		"Create a new Task Master tag context (Ctrl+Shift+A).",
		fields,
		[]string{"Add", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Add" {
				return nil, nil
			}
			name := strings.TrimSpace(stringValue(values, "name"))
			if name == "" {
				return nil, fmt.Errorf("Tag name is required")
			}
			return addTagFormResult{
				Name:            name,
				CopyFromCurrent: boolValue(values, "copyCurrent"),
				CopyFrom:        strings.TrimSpace(stringValue(values, "copyFrom")),
				Description:     strings.TrimSpace(stringValue(values, "description")),
			}, nil
		},
	)
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			appErr := NewOperationError("Add Tag", "Failed to create tag", err).
				WithRecoveryHints(
					"Check if the tag name is valid",
					"Ensure the source context exists if copying",
					"Try again",
				)
			m.showAppError(appErr)
			return nil
		}
		if value == nil {
			return nil
		}
		result, ok := value.(addTagFormResult)
		if !ok {
			return nil
		}
		opts := taskmaster.TagAddOptions{
			Name:            result.Name,
			CopyFromCurrent: result.CopyFromCurrent,
			CopyFrom:        result.CopyFrom,
			Description:     result.Description,
		}
		return m.runTagOperation("add-tag", result.Name, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.AddTagContext(ctx, opts)
		})
	})
}

func (m *Model) openDeleteTagDialog() {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	fields := []dialog.FormField{
		{
			ID:       "name",
			Label:    "Tag name",
			Type:     dialog.FormFieldTypeText,
			Required: true,
		},
		{
			ID:      "confirm",
			Label:   "Skip confirmation (--yes)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: true,
			Value:   true,
		},
	}
	form := dialog.NewFormDialog(
		"Delete Tag Context",
		"Remove an existing tag context (Ctrl+Shift+M).",
		fields,
		[]string{"Delete", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Delete" {
				return nil, nil
			}
			name := strings.TrimSpace(stringValue(values, "name"))
			if name == "" {
				return nil, fmt.Errorf("Tag name is required")
			}
			return deleteTagFormResult{
				Name: name,
				Yes:  boolValue(values, "confirm"),
			}, nil
		},
	)
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			appErr := NewOperationError("Delete Tag", "Failed to delete tag", err).
				WithRecoveryHints(
					"Check if the tag exists",
					"Ensure you have permission to delete",
					"Try again",
				)
			m.showAppError(appErr)
			return nil
		}
		if value == nil {
			return nil
		}
		result, ok := value.(deleteTagFormResult)
		if !ok {
			return nil
		}
		return m.runTagOperation("delete-tag", result.Name, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.DeleteTagContext(ctx, result.Name, result.Yes)
		})
	})
}

func (m *Model) openUseTagDialog() tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}
	fields := []dialog.FormField{
		{
			ID:       "name",
			Label:    "Tag name",
			Type:     dialog.FormFieldTypeText,
			Required: true,
		},
	}
	form := dialog.NewFormDialog(
		"Use Tag Context",
		"Switch the active Task Master tag (Ctrl+T).",
		fields,
		[]string{"Switch", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Switch" {
				return nil, nil
			}
			name := strings.TrimSpace(stringValue(values, "name"))
			if name == "" {
				return nil, fmt.Errorf("Tag name is required")
			}
			return useTagFormResult{Name: name}, nil
		},
	)
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			appErr := NewOperationError("Use Tag", "Failed to switch tag", err).
				WithRecoveryHints(
					"Check if the tag exists",
					"Verify the tag name is correct",
					"Try again",
				)
			m.showAppError(appErr)
			return nil
		}
		if value == nil {
			return nil
		}
		result, ok := value.(useTagFormResult)
		if !ok {
			return nil
		}
		return m.runTagOperation("use-tag", result.Name, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.UseTagContext(ctx, result.Name)
		})
	})
	return nil
}

func (m *Model) executeTaskMasterCommand(label, command string, args ...string) {
	if m.execService == nil {
		return
	}
	if m.execService.IsRunning() {
		appErr := NewOperationError(label, "Another command is already running. Please wait for it to finish.", nil).
			WithRecoveryHints(
				"Wait for the current operation to complete",
				"Check the status in the log area",
			)
		m.showAppError(appErr)
		return
	}
	prettyArgs := strings.Join(args, " ")
	m.addLogLine(fmt.Sprintf("Executing: task-master %s %s", command, prettyArgs))
	if err := m.execService.Execute(command, args...); err != nil {
		appErr := NewOperationError(label, "Failed to start command", err).
			WithRecoveryHints(
				"Check if task-master CLI is properly installed",
				"Verify the command syntax",
				"Try again",
			)
		m.showAppError(appErr)
	}
}

func (m *Model) selectedOrCurrentTaskIDs() []string {
	if len(m.selectedIDs) > 0 {
		return m.getSelectedTasks()
	}
	if m.selectedTask != nil {
		return []string{m.selectedTask.ID}
	}
	return nil
}

func (m *Model) showErrorDialog(title, message string) {
	dm := m.dialogManager()
	if dm == nil {
		m.addLogLine(fmt.Sprintf("%s: %s", title, message))
		return
	}
	errDialog := dialog.NewConfirmationDialog(title, message, 70, 12)
	errDialog.SetYesText("Dismiss")
	errDialog.SetNoText("Cancel")
	errDialog.SetDangerMode(true)
	if errDialog.Style != nil {
		errDialog.Style.BorderColor = errDialog.Style.ErrorColor
		errDialog.Style.FocusedBorderColor = errDialog.Style.ErrorColor
		errDialog.Style.TitleColor = errDialog.Style.ErrorColor
	}
	errDialog.SetFooterHints(
		dialog.ShortcutHint{Key: "Enter", Label: "Dismiss"},
		dialog.ShortcutHint{Key: "Esc", Label: "Cancel"},
	)
	m.appState.PushDialog(errDialog)
}

// showAppError displays an AppError with recovery hints
func (m *Model) showAppError(appErr *AppError) {
	dm := m.dialogManager()
	if dm == nil {
		m.addLogLine(appErr.Error())
		return
	}
	errDialog := dialog.NewErrorDialogModel(appErr.Title, appErr.GetDisplayMessage())
	errDialog.SetDetails(appErr.Details)
	errDialog.SetRecoveryHints(appErr.GetRecoveryMessage())
	m.appState.PushDialog(errDialog)
}

// showCrushDependencyError displays a Crush binary dependency error with recovery hints
func (m *Model) showCrushDependencyError(message string) {
	dm := m.dialogManager()
	if dm == nil {
		m.addLogLine(fmt.Sprintf("Crush Binary Error: %s", message))
		return
	}
	recoveryHint := "Install Crush: go install github.com/charmbracelet/crush@latest\nOr ensure crush is in your PATH"
	errDialog := dialog.NewErrorDialogModel("Crush Binary Not Found", message)
	errDialog.SetRecoveryHints(recoveryHint)
	m.appState.PushDialog(errDialog)
}

// Helper methods for tag commands
func (m *Model) handleTagManagement() tea.Cmd {
	if m.taskService == nil || !m.taskService.IsAvailable() {
		appErr := NewDependencyError("Tag Contexts", "Task Master CLI is not available in this workspace.", nil).
			WithRecoveryHints(
				"Restart the TUI inside a Task Master workspace",
				"Check Task Master CLI installation",
			)
		m.showAppError(appErr)
		return nil
	}
	m.addLogLine("Loading tag contexts from task-master...")
	return loadTagListCmd(m.taskService, false)
}

func stringValue(values map[string]interface{}, key string) string {
	if v, ok := values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func boolValue(values map[string]interface{}, key string) bool {
	if v, ok := values[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func loadTagListCmd(service TaskService, includeMetadata bool) tea.Cmd {
	if service == nil {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		list, err := service.ListTagContexts(ctx, includeMetadata)
		return tagListLoadedMsg{List: list, Err: err}
	}
}

func (m *Model) runTagOperation(operation string, tagName string, fn func(ctx context.Context) (*taskmaster.TagOperationResult, error)) tea.Cmd {
	if m.taskService == nil {
		m.showErrorDialog("Tag Command", "Task service is not available.")
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		result, err := fn(ctx)
		if err != nil {
			return TagOperationMsg{Operation: operation, TagName: tagName, Err: err}
		}
		return TagOperationMsg{Operation: operation, TagName: tagName, Result: result}
	}
}

// DEPRECATED: Replaced by startExpansion using CLI
// This function is kept for backward compatibility only.
func (m *Model) runExpandTask(opts expandOptions) tea.Cmd {
	if m.taskService == nil {
		m.showErrorDialog("Expand Task", "Task service is not available.")
		return nil
	}

	// Get the task to expand
	task, ok := m.taskIndex[opts.TaskID]
	if !ok {
		m.showErrorDialog("Expand Task", fmt.Sprintf("Task %s not found", opts.TaskID))
		return nil
	}

	// Generate drafts using the expansion algorithm
	expandOpts := taskmaster.ExpandTaskOptions{
		Depth:       opts.Depth,
		NumSubtasks: opts.Num,
		UseAI:       opts.UseAI,
	}

	drafts := taskmaster.ExpandTaskDrafts(task, expandOpts)
	if len(drafts) == 0 {
		m.showErrorDialog("Expand Task", "No subtasks were generated. Check the task description.")
		return nil
	}

	// Store drafts and parent ID in app state for workflow
	m.expandTaskDrafts = drafts
	m.expandTaskParentID = opts.TaskID

	// Show preview dialog
	return m.showExpandPreviewDialog(drafts)
}

// DEPRECATED: CLI handles expansion directly
// This function is kept for backward compatibility only.
func (m *Model) showExpandPreviewDialog(drafts []taskmaster.SubtaskDraft) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	previewDialog := dialog.NewExpandTaskPreviewDialog(drafts, dm.Style)

	m.appState.AddDialog(previewDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Expand Task", err.Error())
			return nil
		}
		// User confirmed preview, now show edit dialog
		return m.showExpandEditDialog(m.expandTaskDrafts)
	})

	return previewDialog.Init()
}

// DEPRECATED: CLI handles expansion directly
// This function is kept for backward compatibility only.
func (m *Model) showExpandEditDialog(drafts []taskmaster.SubtaskDraft) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	editDialog := dialog.NewSubtaskEditDialog(drafts, dm.Style)

	m.appState.AddDialog(editDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Expand Task", err.Error())
			return nil
		}

		// Get edited drafts from the dialog
		finalDrafts, ok := value.([]taskmaster.SubtaskDraft)
		if !ok {
			return nil
		}

		// Apply the drafted subtasks
		return m.applyExpandTaskDrafts(m.expandTaskParentID, finalDrafts)
	})

	return editDialog.Init()
}

// DEPRECATED: CLI handles task application automatically
// This function is kept for backward compatibility only.
func (m *Model) applyExpandTaskDrafts(parentID string, drafts []taskmaster.SubtaskDraft) tea.Cmd {
	if m.taskService == nil {
		m.showErrorDialog("Expand Task", "Task service is not available.")
		return nil
	}

	// Get the parent task
	parentTask, ok := m.taskIndex[parentID]
	if !ok {
		m.showErrorDialog("Expand Task", fmt.Sprintf("Parent task %s not found", parentID))
		return nil
	}

	// Apply the drafts
	newIDs, err := taskmaster.ApplySubtaskDrafts(parentTask, drafts)
	if err != nil {
		m.showErrorDialog("Expand Task", fmt.Sprintf("Failed to apply subtasks: %v", err))
		return nil
	}

	// Persist the changes by reloading all tasks
	m.addLogLine(fmt.Sprintf("Successfully created %d subtasks for task %s: %v", len(newIDs), parentID, newIDs))

	// Reload tasks to refresh the UI
	return LoadTasksCmd(m.taskService)
}

// DEPRECATED: Replaced by waitForExpansionMessages
// This function is kept for backward compatibility only.
func (m *Model) waitForExpandTaskMessages() tea.Cmd {
	ch := m.expandTaskMsgCh
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		if msg, ok := <-ch; ok {
			return msg
		}
		return expandTaskStreamClosedMsg{}
	}
}

// DEPRECATED: Replaced by cancelExpansion
// This function is kept for backward compatibility only.
func (m *Model) cancelExpandTask() {
	if m.expandTaskCancel != nil {
		m.expandTaskCancel()
		m.expandTaskCancel = nil
	}
}

// DEPRECATED: Replaced by clearExpansionRuntimeState
// This function is kept for backward compatibility only.
func (m *Model) clearExpandTaskRuntimeState() {
	m.cancelExpandTask()
	m.expandTaskMsgCh = nil
}

// openCommandRunner opens the command runner dialog
func (m *Model) openCommandRunner() tea.Cmd {
	m.addLogLine("Ctrl+B pressed - Open Command Runner")
	
	// Validate Crush binary is available before opening dialog
	if err := dialog.ValidateCrushBinary(); err != nil {
		m.addLogLine(fmt.Sprintf("ERROR: Crush binary not available: %v", err))
		appErr := NewDependencyError("Run Command", "Crush CLI is not available.", err).
			WithRecoveryHints(
				"Install Crush: go install github.com/charmbracelet/crush@latest",
				"Ensure crush is in your PATH",
			)
		m.showAppError(appErr)
		return nil
	}
	
	// Create the command runner dialog
	cmdDialog := dialog.NewCommandRunnerDialog(dialog.CreateDialogStyleFromAppStyles(
		ColorBorder,    // border
		ColorHighlight, // focused border
		ColorText,      // title
		"#333333",      // background
		ColorText,      // text
		ColorHighlight, // button
		ColorBlocked,   // error
		ColorDone,      // success
		ColorPending,   // warning
	))
	
	// Set the dialog ID so it can be identified in handleDialogResultMsg
	cmdDialog.BaseDialog.ID = commandRunnerDialogID
	
	// Add dialog with callback to handle submission
	m.appState.AddDialog(cmdDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			appErr := NewOperationError("Command Runner", "Failed to submit command", err).
				WithRecoveryHints(
					"Check your input and try again",
					"Ensure the prompt is not empty",
				)
			m.showAppError(appErr)
			return nil
		}
		
		// Handle form cancellation
		if value == nil {
			return nil
		}
		
		// Extract prompt from result
		result, ok := value.(dialog.CommandPromptResult)
		if !ok || result.Prompt == "" {
			return nil
		}
		
		// Execute command
		return m.handleCommandRunnerSubmission(result)
	})
	
	return cmdDialog.Init()
}

// handleCommandRunnerSubmission handles the result when user submits the command runner form
func (m *Model) handleCommandRunnerSubmission(result dialog.CommandPromptResult) tea.Cmd {
	m.addLogLine(fmt.Sprintf("Command runner submitted with prompt (length=%d)", len(result.Prompt)))
	
	// Check if Crush binary is available
	if err := dialog.ValidateCrushBinary(); err != nil {
		m.addLogLine(fmt.Sprintf("ERROR: Crush binary not available: %v", err))
		appErr := NewDependencyError("Run Command", "Crush CLI is not available.", err).
			WithRecoveryHints(
				"Install Crush: go install github.com/charmbracelet/crush@latest",
				"Ensure crush is in your PATH",
			)
		m.showAppError(appErr)
		return nil
	}
	
	// Generate a unique command ID based on timestamp
	commandID := fmt.Sprintf("cmd-%d", time.Now().UnixNano())
	
	// Create or show the task runner modal if not already present
	if m.taskRunner == nil {
		m.taskRunner = dialog.NewTaskRunnerModal(m.width, m.height-2, dialog.CreateDialogStyleFromAppStyles(
			ColorBorder,    // border
			ColorHighlight, // focused border
			ColorText,      // title
			"#333333",      // background
			ColorText,      // text
			ColorHighlight, // button
			ColorBlocked,   // error
			ColorDone,      // success
			ColorPending,   // warning
		))
		m.dialogManager().PushDialog(m.taskRunner)
	}
	
	// Execute the ad-hoc command using RunCommand (no specific model)
	// Use tea.Sequence to ensure tab is created before execution starts
	return tea.Sequence(
		m.executeAdHocCommand(commandID, result.Prompt, ""),
		m.continueAdHocCommand(commandID, result.Prompt, ""),
	)
}

// executeAdHocCommand executes an ad-hoc command with Crush
// Creates both the tab initialization and the command execution
func (m *Model) executeAdHocCommand(commandID, prompt, model string) tea.Cmd {
	return func() tea.Msg {
		// First, send TaskStartedMsg to create the tab in the task runner modal
		// This happens immediately
		return dialog.TaskStartedMsg{
			TaskID:    commandID,
			TaskTitle: truncatePrompt(prompt, 30),
			Model:     model,
		}
	}
}

// continueAdHocCommand handles the actual command execution after the tab is created
// This is called after TaskStartedMsg is processed
func (m *Model) continueAdHocCommand(commandID, prompt, model string) tea.Cmd {
	return func() tea.Msg {
		// Call RunCommand to start the execution
		// RunCommand now returns chan tea.Msg directly (includes TaskOutputMsg, TaskCompletedMsg, TaskFailedMsg)
		outputChan, err := dialog.RunCommand(commandID, prompt, model)
		if err != nil {
			m.addLogLine(fmt.Sprintf("ERROR: Failed to start command: %v", err))
			return tea.Msg(dialog.TaskFailedMsg{
				TaskID:  commandID,
				Error:   err.Error(),
				Message: "Failed to start command",
			})
		}
		
		m.addLogLine(fmt.Sprintf("Started ad-hoc command %s", commandID))
		
		// Return a message to set up the subscription to the output channel
		// No conversion needed since RunCommand returns chan tea.Msg directly
		return dialog.CrushExecutionSub{
			TaskID: commandID,
			OutCh:  outputChan,
		}
	}
}

// convertOutputChannelToMsgChannel converts a string channel to a tea.Msg channel
func (m *Model) convertOutputChannelToMsgChannel(taskID string, outputChan <-chan string) chan tea.Msg {
	msgChan := make(chan tea.Msg, 1000)
	go func() {
		defer close(msgChan)
		for line := range outputChan {
			msgChan <- dialog.TaskOutputMsg{
				TaskID: taskID,
				Output: line,
			}
		}
	}()
	return msgChan
}

// Helper methods for ad-hoc command execution

// getCurrentCrushModel returns the Crush model from configuration or empty string for default
func (m *Model) getCurrentCrushModel() string {
	if m.config == nil {
		return ""
	}
	// For now, return empty string to let Crush use its configured default
	// In the future, this could read from config.CrushModel
	return ""
}

// truncatePrompt truncates a prompt to a maximum length, adding "..." if truncated
func truncatePrompt(prompt string, maxLen int) string {
	if len(prompt) <= maxLen {
		return prompt
	}
	if maxLen <= 3 {
		return "..."
	}
	return prompt[:maxLen-3] + "..."
}

// ensureTaskRunnerModal ensures the Task Runner modal is created and visible
// Creates a new modal if one doesn't exist, and sets the visible flag
func (m *Model) ensureTaskRunnerModal() {
	if m.taskRunner == nil {
		// Create new TaskRunnerModal with current dimensions
		m.taskRunner = dialog.NewTaskRunnerModal(m.width, m.height-4, m.appState.DialogStyle())
	}
	m.taskRunnerVisible = true
}

// listenForCommandOutput creates a subscription command that streams output from a command
// Takes a command ID and a channel of output strings, returns a tea.Cmd that:
// - On first call, reads one line from the channel and emits TaskOutputMsg
// - When the channel closes, emits TaskCompletedMsg
// - Designed to be called repeatedly by Update handler to maintain subscription
func (m *Model) listenForCommandOutput(commandID string, outputChan <-chan string) tea.Cmd {
	// Convert string channel to tea.Msg channel for proper subscription
	msgChan := m.convertOutputChannelToMsgChannel(commandID, outputChan)
	// Subscribe to the channel using the built-in WaitForCrushMsg subscription
	return dialog.WaitForCrushMsg(msgChan)
}
