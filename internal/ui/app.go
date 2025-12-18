package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/executor"
	"github.com/agreen757/tm-tui/internal/projects"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// formatHelpLine formats a help line with consistent alignment between keys and descriptions
func formatHelpLine(key string, description string) string {
	// Use fixed width for keys to align descriptions
	const keyWidth = 12
	paddedKey := "  " + key
	if len(paddedKey) < keyWidth {
		paddedKey += strings.Repeat(" ", keyWidth-len(paddedKey))
	}
	return paddedKey + "- " + description
}

// ViewMode represents the current view mode of the task list
type ViewMode int

const (
	ViewModeTree ViewMode = iota
	ViewModeList
	ViewModeKanban // Placeholder for future implementation
)

const (
	complexityProgressMinDisplay   = 1 * time.Second
	complexityProgressCompleteHold = 500 * time.Millisecond
)

// Panel represents which panel is currently focused
type Panel int

const (
	PanelTaskList Panel = iota
	PanelDetails
	PanelLog
)

// Model represents the TUI application state
type Model struct {
	// Services
	config        *config.Config
	configManager *config.ConfigManager
	taskService   TaskService
	execService   *executor.Service
	appState      *AppState
	lastPrdPath   string

	// Task data
	tasks        []taskmaster.Task
	taskIndex    map[string]*taskmaster.Task // Quick lookup by ID
	visibleTasks []*taskmaster.Task          // Flattened list of visible tasks (respects expand/collapse)

	// View state
	viewMode             ViewMode
	focusedPanel         Panel
	selectedIndex        int // Index in visibleTasks array
	selectedTask         *taskmaster.Task
	expandedNodes        map[string]bool
	selectedIDs          map[string]bool // Track multi-select for bulk operations
	deleteWorkflow       *DeleteWorkflowState
	undoSession          *UndoSession
	tagActionContext     *taskmaster.TagContext
	projectRegistry      *projects.Registry
	activeProject        *projects.Metadata
	pendingProjectSwitch *projects.Metadata
	pendingProjectTag    string

	// Layout
	width  int
	height int
	ready  bool

	// Panels
	taskListViewport viewport.Model
	detailsViewport  viewport.Model
	logViewport      viewport.Model
	helpModel        help.Model
	keyMap           KeyMap
	commandShortcuts []commandShortcut
	commands         []CommandSpec

	// Dialog stack for modal dialogs
	complexityMsgCh          chan tea.Msg
	complexityCancel         context.CancelFunc
	currentComplexityScope   string
	currentComplexityTags    []string
	complexityStartedAt      time.Time
	waitingForComplexityHold bool
	parsePrdChan             chan tea.Msg
	parsePrdCancel           context.CancelFunc

	// DEPRECATED: Legacy expansion state - being replaced by new CLI-based flow
	expandTaskMsgCh    chan tea.Msg
	expandTaskCancel   context.CancelFunc
	expandTaskDrafts   []taskmaster.SubtaskDraft // Generated drafts waiting for preview/edit
	expandTaskParentID string                     // ID of task being expanded

	// New expansion workflow state
	expansionMsgCh          chan tea.Msg
	expansionCancel         context.CancelFunc
	currentExpansionScope   string
	currentExpansionTags    []string
	expansionStartedAt      time.Time
	waitingForExpansionHold bool

	// Panel visibility
	showDetailsPanel bool
	showLogPanel     bool
	showHelp         bool

	// Command mode state
	commandMode  bool
	commandInput string

	// Search state
	searchMode    bool
	searchInput   textinput.Model
	searchQuery   string
	searchResults []*taskmaster.Task

	// Filter state
	statusFilter string // empty = all, or specific status like "pending", "in-progress", etc.

	// Confirmation mode state
	confirmingClearState bool

	// Task Runner Modal
	taskRunner        *dialog.TaskRunnerModal
	taskRunnerVisible bool

	// Crush run context (for tracking model selection -> crush execution flow)
	crushRunPending   bool
	crushRunTaskID    string
	crushRunTaskTitle string
	crushRunTask      *taskmaster.Task
	crushRunChannels  map[string]chan tea.Msg // taskID -> output channel for active runs

	// Styles
	styles *Styles

	// Log data
	logLines []string

	// Error state
	err error
}

func (m *Model) dialogManager() *dialog.DialogManager {
	if m.appState == nil {
		return nil
	}
	return m.appState.DialogManager()
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, configManager *config.ConfigManager, taskService TaskService, execService *executor.Service) Model {
	tasks, _ := taskService.GetTasks()

	// Initialize viewports (sizes will be set on first WindowSizeMsg)
	taskListVP := viewport.New(0, 0)
	detailsVP := viewport.New(0, 0)
	logVP := viewport.New(0, 0)

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search tasks (ID, title, description)..."
	searchInput.Focus()
	searchInput.CharLimit = 100
	searchInput.Width = 40
	searchInput.Prompt = ""

	// Style the text input for better visibility
	searchInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	searchInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	// Set up dialog theme colors matching the app style
	dialogStyle := dialog.CreateDialogStyleFromAppStyles(
		ColorBorder,    // border
		ColorHighlight, // focused border
		ColorText,      // title
		"#333333",      // background
		ColorText,      // text
		ColorHighlight, // button
		ColorBlocked,   // error
		ColorDone,      // success
		ColorPending,   // warning
	)

	// Initialize dialog manager (size will be set on first WindowSizeMsg)
	dialogManager := dialog.InitializeDialogManager(0, 0, dialogStyle)
	keyMap := NewKeyMap(cfg)
	appState := NewAppState(dialogManager, &keyMap)

	// Initialize TaskRunnerModal (size will be set on first WindowSizeMsg)
	taskRunner := dialog.NewTaskRunnerModal(0, 0, dialogStyle)

	m := Model{
		config:            cfg,
		configManager:     configManager,
		taskService:       taskService,
		execService:       execService,
		tasks:             tasks,
		taskIndex:         make(map[string]*taskmaster.Task),
		visibleTasks:      []*taskmaster.Task{},
		selectedIndex:     0,
		ready:             false,
		viewMode:          ViewModeTree,
		focusedPanel:      PanelTaskList,
		expandedNodes:     make(map[string]bool),
		selectedIDs:       make(map[string]bool),
		taskListViewport:  taskListVP,
		detailsViewport:   detailsVP,
		logViewport:       logVP,
		helpModel:         help.New(),
		keyMap:            keyMap,
		appState:          appState,
		taskRunner:        taskRunner,
		taskRunnerVisible: false,
		crushRunChannels:  make(map[string]chan tea.Msg), // Initialize crush run channels map
		showDetailsPanel:  true,
		showLogPanel:      false,
		showHelp:          false,
		commandMode:       false,
		commandInput:      "",
		searchInput:       searchInput,
		styles:            NewStyles(),
		logLines:          []string{},
		projectRegistry:   taskService.ProjectRegistry(),
		activeProject:     taskService.ActiveProjectMetadata(),
	}

	m.commands = defaultCommandSpecs()
	m.registerDefaultCommandShortcuts()

	// Build task index
	m.buildTaskIndex()

	// Rebuild visible tasks list
	m.rebuildVisibleTasks()

	// Try to load and restore previous UI state
	if cfg != nil && cfg.StatePath != "" {
		if state, err := config.LoadState(cfg.StatePath); err == nil {
			m.restoreUIState(state)
		}
		// If state loading fails, we just use the default initial state (first task selected)
	}

	// If state restoration didn't set a selection, select first task
	if m.selectedTask == nil && len(m.visibleTasks) > 0 {
		// Use task index to get stable pointer
		taskID := m.visibleTasks[0].ID
		if task, ok := m.taskIndex[taskID]; ok {
			m.selectedTask = task
		} else {
			m.selectedTask = m.visibleTasks[0]
		}
		m.selectedIndex = 0
	}

	return m
}

// registerDefaultCommandShortcuts wires built-in shortcuts to commands.
func (m *Model) registerDefaultCommandShortcuts() {
	m.commandShortcuts = []commandShortcut{
		{binding: m.keyMap.ParsePRD, command: CommandParsePRD, help: "Parse PRD"},
		{binding: m.keyMap.AnalyzeComplexity, command: CommandAnalyzeComplexity, help: "Analyze Complexity"},
		{binding: m.keyMap.ExpandTask, command: CommandExpandTask, help: "Expand Task"},
		{binding: m.keyMap.DeleteTask, command: CommandDeleteTask, help: "Delete Task"},
		{binding: m.keyMap.RunTask, command: CommandRunTask, help: "Run Task with Crush"},
		{binding: m.keyMap.ManageTags, command: CommandManageTags, help: "Add Tag Context"},
		{binding: m.keyMap.TagManagement, command: CommandTagManagement, help: "Manage Tags"},
		{binding: m.keyMap.UseTag, command: CommandUseTag, help: "Use Tag"},
		{binding: m.keyMap.ProjectTags, command: CommandProjectTags, help: "Project Tags"},
		{binding: m.keyMap.ProjectQuickSwitch, command: CommandProjectQuickSwitch, help: "Quick Project Switch"},
		{binding: m.keyMap.ProjectSearch, command: CommandProjectSearch, help: "Project Search"},
	}
}

// tryHandleCommandShortcut executes a command if the key matches a registered shortcut.
func (m *Model) tryHandleCommandShortcut(msg tea.KeyMsg) (tea.Cmd, bool) {
	for _, shortcut := range m.commandShortcuts {
		if key.Matches(msg, shortcut.binding) {
			cmd := m.dispatchCommand(shortcut.command)
			return cmd, true
		}
	}
	return nil, false
}

// openCommandPalette displays the palette dialog listing all commands.
func (m *Model) openCommandPalette() {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	items := make([]dialog.ListItem, 0, len(m.commands))
	for _, spec := range m.commands {
		items = append(items, newCommandPaletteItem(spec))
	}
	palette := dialog.NewListDialog("Command Palette", 72, 18, items)
	palette.SetShowDescription(true)
	m.appState.AddDialog(palette, nil)
}

func (m *Model) handleListSelection(msg dialog.ListSelectionMsg) tea.Cmd {
	if item, ok := msg.SelectedItem.(*commandPaletteItem); ok {
		return m.dispatchCommand(item.spec.ID)
	}
	if item, ok := msg.SelectedItem.(*prdResultListItem); ok {
		if item.taskID != "" {
			return func() tea.Msg {
				return SelectTaskMsg{TaskID: item.taskID}
			}
		}
		return nil
	}
	if item, ok := msg.SelectedItem.(*tagListItem); ok && item.ctx != nil {
		m.tagActionContext = item.ctx
		m.openTagActionDialog()
		return nil
	}
	if actionItem, ok := msg.SelectedItem.(*tagActionItem); ok {
		return m.handleTagActionSelection(actionItem.kind)
	}
	if tagItem, ok := msg.SelectedItem.(*projectTagItem); ok && tagItem != nil {
		return m.openProjectSelectionDialog(tagItem.summary.Name)
	}
	if projectItem, ok := msg.SelectedItem.(*projectListItem); ok {
		return m.handleProjectListItemSelection(projectItem)
	}
	// Handle model selection from ModelSelectionDialog
	if modelItem, ok := msg.SelectedItem.(*dialog.ModelSelectionListItem); ok {
		opt := modelItem.GetOption()
		return func() tea.Msg {
			return dialog.ModelSelectionMsg{
				Provider:  opt.Provider,
				ModelName: opt.ModelID,
				ModelID:   opt.ModelID,
			}
		}
	}
	if msg.SelectedItem != nil {
		m.addLogLine(fmt.Sprintf("Selected item: %s", msg.SelectedItem.Title()))
	}
	return nil
}

func (m *Model) handleTagOperationMsg(msg TagOperationMsg) tea.Cmd {
	if msg.Err != nil {
		appErr := NewOperationError("Tag Command", "Tag operation failed", msg.Err).
			WithRecoveryHints(
				"Check if you have permission to perform this action",
				"Verify the tag exists",
				"Try again",
			)
		m.showAppError(appErr)
		return nil
	}
	if msg.Result == nil {
		return nil
	}
	label := strings.ReplaceAll(msg.Operation, "-", " ")
	if label != "" {
		label = strings.ToUpper(label[:1]) + label[1:]
	}
	if msg.TagName != "" {
		label = fmt.Sprintf("%s (%s)", label, msg.TagName)
	}
	m.addLogLine(label)
	if output := strings.TrimSpace(msg.Result.Output); output != "" {
		m.addLogLine(output)
	}
	if msg.Operation == "use-tag" && m.config != nil && msg.TagName != "" {
		m.config.ActiveTag = msg.TagName
	}
	return LoadTasksCmd(m.taskService)
}

func (m *Model) handleProjectListItemSelection(item *projectListItem) tea.Cmd {
	if item == nil || item.meta == nil {
		return nil
	}
	tag := strings.TrimSpace(item.Tag())
	if tag == "" {
		tag = item.meta.PrimaryTag()
	}
	if tag == "" {
		appErr := NewValidationError("Switch Project", fmt.Sprintf("Project '%s' has no associated tags.", item.meta.Name), nil).
			WithRecoveryHints(
				"Add tags to this project first",
				"Try selecting a different project",
			)
		m.showAppError(appErr)
		return nil
	}
	return m.requestProjectSwitch(item.meta, tag)
}

func (m *Model) useProjectTag(tag string) tea.Cmd {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil
	}
	if m.config != nil && strings.EqualFold(strings.TrimSpace(m.config.ActiveTag), tag) {
		m.addLogLine(fmt.Sprintf("Already using tag %s", tag))
		return nil
	}
	return m.runTagOperation("use-tag", tag, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
		return m.taskService.UseTagContext(ctx, tag)
	})
}

// buildTaskIndex creates a flat index of all tasks by ID for quick lookup
func (m *Model) buildTaskIndex() {
	m.taskIndex = make(map[string]*taskmaster.Task)
	
	// Use stable pointer recursion to avoid creating pointers to temporary slice elements
	var indexTask func(task *taskmaster.Task)
	indexTask = func(task *taskmaster.Task) {
		m.taskIndex[task.ID] = task
		for i := range task.Subtasks {
			indexTask(&task.Subtasks[i])
		}
	}
	
	for i := range m.tasks {
		indexTask(&m.tasks[i])
	}
}

// rebuildVisibleTasks rebuilds the visibleTasks slice based on view mode and expanded state
// Note: This function updates the visible tasks array but does NOT automatically
// preserve the selection. Callers should use ensureTaskSelected() if they need
// to maintain the current selection after the rebuild.
func (m *Model) rebuildVisibleTasks() {
	// In list view, show all tasks regardless of expanded state
	if m.viewMode == ViewModeList {
		m.visibleTasks = m.flattenAllTasks()
	} else {
		// In tree view, respect expanded state
		m.visibleTasks = m.flattenTasks()
	}

	// Ensure selectedIndex is valid after rebuild
	if m.selectedIndex >= len(m.visibleTasks) {
		m.selectedIndex = len(m.visibleTasks) - 1
	}
	if m.selectedIndex < 0 && len(m.visibleTasks) > 0 {
		m.selectedIndex = 0
	}

	// Update selectedTask based on selectedIndex
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.visibleTasks) {
		// Use task index to get stable pointer
		taskID := m.visibleTasks[m.selectedIndex].ID
		if task, ok := m.taskIndex[taskID]; ok {
			m.selectedTask = task
		} else {
			m.selectedTask = m.visibleTasks[m.selectedIndex]
		}
	}
}

// extractUIState extracts the current UI state from the model for persistence
func (m *Model) extractUIState() *config.UIState {
	// Extract expanded node IDs
	expandedIDs := make([]string, 0, len(m.expandedNodes))
	for id, expanded := range m.expandedNodes {
		if expanded {
			expandedIDs = append(expandedIDs, id)
		}
	}

	// Get selected task ID
	selectedID := ""
	if m.selectedTask != nil {
		selectedID = m.selectedTask.ID
	}

	// Convert ViewMode to string
	viewModeStr := "tree"
	switch m.viewMode {
	case ViewModeList:
		viewModeStr = "list"
	case ViewModeKanban:
		viewModeStr = "kanban"
	}

	// Convert Panel to string
	focusedPanelStr := "taskList"
	switch m.focusedPanel {
	case PanelDetails:
		focusedPanelStr = "details"
	case PanelLog:
		focusedPanelStr = "log"
	}

	return &config.UIState{
		ExpandedIDs:      expandedIDs,
		SelectedID:       selectedID,
		ViewMode:         viewModeStr,
		FocusedPanel:     focusedPanelStr,
		ShowDetailsPanel: m.showDetailsPanel,
		ShowLogPanel:     m.showLogPanel,
		PanelHeights:     make(map[string]int), // Can be extended later
		LastPrdPath:      m.lastPrdPath,
	}
}

// SaveUIState persists the current UI state to disk
func (m *Model) SaveUIState() error {
	if m.config == nil || m.config.StatePath == "" {
		return nil // No state path configured, skip saving
	}

	state := m.extractUIState()
	return config.SaveState(m.config.StatePath, state)
}

// restoreUIState restores UI state from a UIState object
func (m *Model) restoreUIState(state *config.UIState) {
	if state == nil {
		return
	}

	// Restore expanded nodes
	m.expandedNodes = make(map[string]bool)
	for _, id := range state.ExpandedIDs {
		m.expandedNodes[id] = true
	}

	// Restore view mode
	switch state.ViewMode {
	case "list":
		m.viewMode = ViewModeList
	case "kanban":
		m.viewMode = ViewModeKanban
	default:
		m.viewMode = ViewModeTree
	}

	// Restore focused panel
	switch state.FocusedPanel {
	case "details":
		m.focusedPanel = PanelDetails
	case "log":
		m.focusedPanel = PanelLog
	default:
		m.focusedPanel = PanelTaskList
	}

	// Restore panel visibility
	m.showDetailsPanel = state.ShowDetailsPanel
	m.showLogPanel = state.ShowLogPanel
	m.lastPrdPath = state.LastPrdPath

	// Rebuild visible tasks with restored expanded state
	m.rebuildVisibleTasks()

	// Restore selected task
	if state.SelectedID != "" {
		if task, ok := m.taskIndex[state.SelectedID]; ok {
			m.selectedTask = task
			// Find index in visible tasks
			for i, t := range m.visibleTasks {
				if t.ID == state.SelectedID {
					m.selectedIndex = i
					break
				}
			}
		}
	}
}

// ClearUIState resets all UI state to defaults and deletes the state file
func (m *Model) ClearUIState() error {
	// Delete the state file
	if m.config != nil && m.config.StatePath != "" {
		if err := os.Remove(m.config.StatePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove state file: %w", err)
		}
	}

	// Reset all in-memory state to defaults
	m.expandedNodes = make(map[string]bool)
	m.selectedIndex = 0
	m.viewMode = ViewModeTree
	m.focusedPanel = PanelTaskList
	m.showDetailsPanel = true
	m.showLogPanel = false
	m.commandMode = false
	m.commandInput = ""
	m.confirmingClearState = false

	// Rebuild visible tasks with cleared expanded state
	m.rebuildVisibleTasks()

	// Select first task if available
	if len(m.visibleTasks) > 0 {
		// Use task index to get stable pointer
		taskID := m.visibleTasks[0].ID
		if task, ok := m.taskIndex[taskID]; ok {
			m.selectedTask = task
		} else {
			m.selectedTask = m.visibleTasks[0]
		}
		m.selectedIndex = 0
	} else {
		m.selectedTask = nil
	}

	// Update viewports
	m.updateTaskListViewport()
	m.updateDetailsViewport()

	return nil
}

// Dialog helper methods

// ShowConfirmationDialog shows a confirmation dialog with Yes/No options
func (m *Model) ShowConfirmationDialog(title, message string, dangerMode bool) {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	confirmDialog := dialog.YesNo(title, message, dangerMode)
	m.appState.PushDialog(confirmDialog)
}

// ShowInformationDialog shows a simple information dialog with an OK button
func (m *Model) ShowInformationDialog(title, message string) {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	okDialog := dialog.OkDialog(title, message)
	m.appState.PushDialog(okDialog)
}

// ShowNotificationDialog displays an informational OK dialog and records the message in the log.
func (m *Model) ShowNotificationDialog(title, message, level string, duration time.Duration) {
	if level == "" {
		level = "info"
	}
	levelLabel := strings.ToUpper(level)
	logEntry := fmt.Sprintf("[%s] %s: %s", levelLabel, title, message)
	if duration > 0 {
		logEntry = fmt.Sprintf("%s (hint: %s)", logEntry, duration.Round(time.Second))
	}
	m.addLogLine(logEntry)

	dm := m.dialogManager()
	if dm == nil {
		return
	}
	notification := dialog.OkDialog(title, message)
	m.appState.AddDialog(notification, nil)
}

// ShowFormDialog shows a form dialog with the specified fields

func (m *Model) ShowFormDialog(title string, width, height int, fields []dialog.FormField) {
	if m.dialogManager() == nil {
		return
	}
	formDialog := dialog.NewLegacyFormDialog(title, width, height, fields)
	m.appState.PushDialog(formDialog)
}

// ShowListDialog shows a list dialog with the specified items
func (m *Model) ShowListDialog(title string, width, height int, items []dialog.ListItem, multiSelect bool) {
	if m.dialogManager() == nil {
		return
	}
	listDialog := dialog.NewListDialog(title, width, height, items)
	listDialog.SetMultiSelect(multiSelect)
	m.appState.PushDialog(listDialog)
}

// ShowProgressDialog shows a progress dialog
func (m *Model) ShowProgressDialog(title string, width, height int) *dialog.ProgressDialog {
	if m.dialogManager() == nil {
		return nil
	}
	progressDialog := dialog.NewProgressDialog(title, width, height)
	m.appState.PushDialog(progressDialog)
	return progressDialog
}

// UpdateProgress updates the progress of the active progress dialog
func (m *Model) UpdateProgress(progress float64, label string) tea.Cmd {
	return dialog.UpdateProgress(progress, label)
}

// CloseAllDialogs closes all open dialogs
func (m *Model) CloseAllDialogs() {
	m.appState.ClearDialogs()
}

// ShowModelSelectionDialog opens the model selection dialog
func (m *Model) ShowModelSelectionDialog() {
	if m.dialogManager() == nil {
		return
	}
	modelSelectionDialog := dialog.NewModelSelectionDialogSimple()
	m.appState.PushDialog(modelSelectionDialog)
}

// handleModelSelection handles model selection from the dialog
func (m *Model) handleModelSelection(msg dialog.ModelSelectionMsg) tea.Cmd {
	// Save the selected model to config
	if err := config.SaveModelConfig(msg.Provider, msg.ModelName); err != nil {
		m.addLogLine(fmt.Sprintf("Error saving model config: %v", err))
		return nil
	}

	m.addLogLine(fmt.Sprintf("Model set to: %s (%s)", msg.ModelID, msg.Provider))

	// Close the dialog
	m.appState.PopDialog()

	// Check if this model selection is for a Crush run
	if m.crushRunPending && m.crushRunTask != nil {
		// Clear the pending state
		m.crushRunPending = false
		taskID := m.crushRunTaskID
		taskTitle := m.crushRunTaskTitle
		task := m.crushRunTask
		
		// Clear context
		m.crushRunTaskID = ""
		m.crushRunTaskTitle = ""
		m.crushRunTask = nil
		
		// Start the Crush run
		return m.startCrushRun(taskID, taskTitle, task, msg.ModelID)
	}

	return nil
}

// openModelSelectionForCrushRun opens the model selection dialog specifically for Crush run
// The selected model will be used to execute the task via Crush
func (m *Model) openModelSelectionForCrushRun(task *taskmaster.Task) tea.Cmd {
	if m.dialogManager() == nil {
		return nil
	}

	// Store task context for after model selection
	// Use the exact task object that was selected to ensure consistency
	m.crushRunPending = true
	m.crushRunTaskID = task.ID
	m.crushRunTaskTitle = task.Title
	m.crushRunTask = task
	
	m.addLogLine(fmt.Sprintf("Stored task for Crush run: ID=%s, Title='%s'", task.ID, task.Title))
	// Debug: Log the actual task data we're using
	m.addLogLine(fmt.Sprintf("DEBUG: Task ID=%s, Title='%s', Deps=%v", task.ID, task.Title, task.Dependencies))

	modelSelectionDialog := dialog.NewModelSelectionDialogSimple()
	m.appState.PushDialog(modelSelectionDialog)
	m.addLogLine(fmt.Sprintf("Select AI model to run task %s: %s", task.ID, task.Title))
	
	return nil
}

// startCrushRun initiates a Crush subprocess execution for the given task
func (m *Model) startCrushRun(taskID, taskTitle string, task *taskmaster.Task, modelID string) tea.Cmd {
	// Validate crush binary
	if err := dialog.ValidateCrushBinary(); err != nil {
		appErr := NewDependencyError("Crush Run", err.Error(), err).
			WithRecoveryHints(
				"Install Crush: go install github.com/crush-ai/crush@latest",
				"Ensure Crush is in your PATH",
				"Verify the installation with: crush --version",
			)
		m.showAppError(appErr)
		return nil
	}

	// Generate the prompt for Crush
	prompt, err := dialog.GenerateCrushPrompt(task, modelID)
	if err != nil {
		appErr := NewOperationError("Crush Run", "Failed to generate prompt", err).
			WithRecoveryHints(
				"Check if CRUSH_RUN_INSTRUCTIONS.md is valid",
				"Verify task details are complete",
				"Try again",
			)
		m.showAppError(appErr)
		return nil
	}

	// Log the generated prompt for debugging
	m.addLogLine(fmt.Sprintf("Generated prompt (first 200 chars): %s...", truncateString(prompt, 200)))

	// Ensure task runner modal exists and is visible
	if m.taskRunner == nil {
		m.taskRunner = dialog.NewTaskRunnerModal(m.width, m.height-4, m.appState.DialogStyle())
	}
	m.taskRunnerVisible = true

	// Send TaskStartedMsg to create a new tab
	m.addLogLine(fmt.Sprintf("Starting Crush run for task %s with model %s", taskID, modelID))
	
	// Return commands: first start the tab, then start execution with prompt
	return tea.Sequence(
		dialog.StartCrushExecution(taskID, taskTitle, modelID, prompt, m.taskRunner),
		dialog.ExecuteCrushSubprocess(taskID, modelID, prompt),
	)
}

// renderTaskTree renders the task tree with proper indentation and expand/collapse
func (m Model) renderTaskTree(tasks []taskmaster.Task, depth int) string {
	var b strings.Builder

	for i := range tasks {
		task := &tasks[i]
		indent := strings.Repeat("  ", depth)
		statusIcon := GetStatusIcon(task.Status)
		statusStyle := m.styles.GetStatusStyle(task.Status)

		// Determine if this is the selected task - check by ID only
		isSelected := m.selectedTask != nil && m.selectedTask.ID == task.ID

		// Build the line
		line := ""

		// Selection checkbox
		if m.isTaskSelected(task.ID) {
			line += "[✓] "
		} else {
			line += "[ ] "
		}

		// Expand/collapse indicator
		hasSubtasks := len(task.Subtasks) > 0
		if hasSubtasks {
			if m.expandedNodes[task.ID] {
				line += "▼ "
			} else {
				line += "▶ "
			}
		} else {
			line += "  "
		}

		// Status icon and task info
		line += statusStyle.Render(statusIcon) + " "
		line += fmt.Sprintf("%s: %s", task.ID, task.Title)

		// Add complexity indicator with text for accessibility
		if task.Complexity > 0 {
			complexityIndicator := GetComplexityIndicator(task.Complexity)
			if complexityIndicator != "" {
				line += " " + complexityIndicator
			}
		}

		// Build full line with cursor and indentation
		var fullLine string
		if isSelected {
			// Selected task gets cursor
			fullLine = "> " + indent + line
			fullLine = m.styles.TaskSelected.Render(fullLine)
		} else {
			// Unselected tasks get space for alignment
			fullLine = "  " + indent + line
			fullLine = m.styles.TaskUnselected.Render(fullLine)
		}

		b.WriteString(fullLine)
		b.WriteString("\n")

		// Recursively render subtasks if expanded
		if hasSubtasks && m.expandedNodes[task.ID] {
			b.WriteString(m.renderTaskTree(task.Subtasks, depth+1))
		}
	}

	return b.String()
}

// flattenTasks returns a flat list of all visible tasks in tree order
func (m Model) flattenTasks() []*taskmaster.Task {
	var result []*taskmaster.Task
	var flatten func(tasks []taskmaster.Task)
	flatten = func(tasks []taskmaster.Task) {
		for i := range tasks {
			// Use task index to get stable pointer instead of creating pointer to slice element
			if task, ok := m.taskIndex[tasks[i].ID]; ok {
				result = append(result, task)
				if len(task.Subtasks) > 0 && m.expandedNodes[task.ID] {
					flatten(task.Subtasks)
				}
			}
		}
	}
	flatten(m.tasks)
	return result
}

// selectNext selects the next visible task
func (m *Model) selectNext() {
	if len(m.visibleTasks) == 0 {
		return
	}

	m.selectedIndex++
	if m.selectedIndex >= len(m.visibleTasks) {
		m.selectedIndex = len(m.visibleTasks) - 1
	}

	if m.selectedIndex >= 0 && m.selectedIndex < len(m.visibleTasks) {
		// Ensure we're using the stable pointer from the index
		taskID := m.visibleTasks[m.selectedIndex].ID
		if task, ok := m.taskIndex[taskID]; ok {
			m.selectedTask = task
		} else {
			m.selectedTask = m.visibleTasks[m.selectedIndex]
		}
	}
}

// selectPrevious selects the previous visible task
func (m *Model) selectPrevious() {
	if len(m.visibleTasks) == 0 {
		return
	}

	m.selectedIndex--
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}

	if m.selectedIndex >= 0 && m.selectedIndex < len(m.visibleTasks) {
		// Ensure we're using the stable pointer from the index
		taskID := m.visibleTasks[m.selectedIndex].ID
		if task, ok := m.taskIndex[taskID]; ok {
			m.selectedTask = task
		} else {
			m.selectedTask = m.visibleTasks[m.selectedIndex]
		}
	}
}

// selectParentTask selects the parent task of the currently selected task
func (m *Model) selectParentTask() {
	if m.selectedTask == nil {
		return
	}

	// Get parent ID (e.g., "1.2.3" -> "1.2", "1.2" -> "1")
	parentID := m.getParentID(m.selectedTask.ID)
	if parentID == "" {
		// Already at root level, can't go higher
		return
	}

	// Select the parent task
	m.selectTaskByID(parentID)
}

// getParentID returns the parent task ID for a given task ID
// Returns empty string if the task is at root level
func (m *Model) getParentID(taskID string) string {
	parts := strings.Split(taskID, ".")
	if len(parts) <= 1 {
		// Root level task, no parent
		return ""
	}

	// Return all parts except the last one
	return strings.Join(parts[:len(parts)-1], ".")
}

// selectTaskByID selects a task by ID and ensures its ancestors are expanded
func (m *Model) selectTaskByID(taskID string) bool {
	_, ok := m.taskIndex[taskID]
	if !ok {
		return false
	}

	// Expand all ancestors
	m.expandAncestors(taskID)

	// Rebuild visible tasks
	m.rebuildVisibleTasks()

	// Find the task in visibleTasks and select it
	for i, t := range m.visibleTasks {
		if t.ID == taskID {
			m.selectedIndex = i
			m.selectedTask = t
			return true
		}
	}

	return false
}

// expandAncestors ensures all ancestors of a task are expanded
func (m *Model) expandAncestors(taskID string) {
	// Parse task ID to find ancestors (e.g., "1.2.3" -> ["1", "1.2"])
	ancestors := m.getAncestorIDs(taskID)
	for _, ancestorID := range ancestors {
		m.expandedNodes[ancestorID] = true
	}
}

// getAncestorIDs returns IDs of all ancestors for a given task ID
func (m *Model) getAncestorIDs(taskID string) []string {
	var ancestors []string
	parts := strings.Split(taskID, ".")

	for i := 1; i < len(parts); i++ {
		ancestorID := strings.Join(parts[:i], ".")
		ancestors = append(ancestors, ancestorID)
	}

	return ancestors
}

// toggleExpanded toggles the expanded state of the selected task
func (m *Model) toggleExpanded() {
	if m.selectedTask == nil || len(m.selectedTask.Subtasks) == 0 {
		return
	}

	// Store the selected task ID to restore after rebuild
	selectedTaskID := m.selectedTask.ID

	m.expandedNodes[m.selectedTask.ID] = !m.expandedNodes[m.selectedTask.ID]
	m.rebuildVisibleTasks()

	// Ensure the same task is still selected after rebuild
	m.ensureTaskSelected(selectedTaskID)
}

// expandSelected expands the selected task if it has subtasks
func (m *Model) expandSelected() {
	if m.selectedTask == nil || len(m.selectedTask.Subtasks) == 0 {
		return
	}

	// Store the selected task ID to restore after rebuild
	selectedTaskID := m.selectedTask.ID

	m.expandedNodes[m.selectedTask.ID] = true
	m.rebuildVisibleTasks()

	// Ensure the same task is still selected after rebuild
	m.ensureTaskSelected(selectedTaskID)
}

// ensureTaskSelected ensures a specific task ID is selected
func (m *Model) ensureTaskSelected(taskID string) {
	for i, task := range m.visibleTasks {
		if task.ID == taskID {
			m.selectedIndex = i
			// Use task index to get stable pointer
			if stableTask, ok := m.taskIndex[taskID]; ok {
				m.selectedTask = stableTask
			} else {
				m.selectedTask = task
			}
			return
		}
	}
}

// collapseSelected collapses the selected task if it's expanded,
// otherwise navigates to the parent task (standard tree navigation behavior)
func (m *Model) collapseSelected() {
	if m.selectedTask == nil {
		return
	}

	// If task has subtasks and is expanded, collapse it
	if len(m.selectedTask.Subtasks) > 0 && m.expandedNodes[m.selectedTask.ID] {
		// Store the selected task ID to restore after rebuild
		selectedTaskID := m.selectedTask.ID

		m.expandedNodes[m.selectedTask.ID] = false
		m.rebuildVisibleTasks()

		// Ensure the same task is still selected after rebuild
		m.ensureTaskSelected(selectedTaskID)
		return
	}

	// Otherwise, navigate to parent task
	m.selectParentTask()
}

// expandAll expands all tasks in the tree
func (m *Model) expandAll() {
	var expandRecursive func(tasks []taskmaster.Task)
	expandRecursive = func(tasks []taskmaster.Task) {
		for i := range tasks {
			task := &tasks[i]
			if len(task.Subtasks) > 0 {
				m.expandedNodes[task.ID] = true
				expandRecursive(task.Subtasks)
			}
		}
	}
	expandRecursive(m.tasks)
	m.rebuildVisibleTasks()
}

// collapseAll collapses all tasks in the tree
func (m *Model) collapseAll() {
	m.expandedNodes = make(map[string]bool)
	m.rebuildVisibleTasks()
}

// toggleSelection toggles the selection state of the current task
func (m *Model) toggleSelection() {
	if m.selectedTask == nil {
		return
	}

	if m.selectedIDs[m.selectedTask.ID] {
		delete(m.selectedIDs, m.selectedTask.ID)
	} else {
		m.selectedIDs[m.selectedTask.ID] = true
	}
}

// clearSelection clears all selected tasks
func (m *Model) clearSelection() {
	m.selectedIDs = make(map[string]bool)
}

// getSelectedTasks returns a slice of all selected task IDs
func (m *Model) getSelectedTasks() []string {
	var selected []string
	for id := range m.selectedIDs {
		selected = append(selected, id)
	}
	return selected
}

// isTaskSelected returns true if the task is in the selection
func (m *Model) isTaskSelected(taskID string) bool {
	return m.selectedIDs[taskID]
}

// filterTaskTree filters tasks in tree structure, keeping parents of matching tasks
func (m Model) filterTaskTree(tasks []taskmaster.Task) []taskmaster.Task {
	// Apply status filter first
	var statusFiltered []taskmaster.Task
	if m.statusFilter == "" {
		statusFiltered = tasks
	} else {
		statusFiltered = m.filterTaskTreeByStatus(tasks, m.statusFilter)
	}

	// Then apply search filter
	if m.searchQuery == "" {
		return statusFiltered
	}

	return m.filterTaskTreeBySearch(statusFiltered, m.searchQuery)
}

// filterTaskTreeByStatus filters tasks in tree structure by status
func (m Model) filterTaskTreeByStatus(tasks []taskmaster.Task, status string) []taskmaster.Task {
	var filtered []taskmaster.Task

	for i := range tasks {
		task := tasks[i]
		matches := task.Status == status

		// Check if any subtask matches
		if len(task.Subtasks) > 0 {
			filteredSubtasks := m.filterTaskTreeByStatus(task.Subtasks, status)
			if len(filteredSubtasks) > 0 {
				// Keep this parent and its matching subtasks
				task.Subtasks = filteredSubtasks
				filtered = append(filtered, task)
			} else if matches {
				// This task matches but subtasks don't - keep task without subtasks
				task.Subtasks = nil
				filtered = append(filtered, task)
			}
		} else if matches {
			// Leaf task that matches
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// filterTaskTreeBySearch filters tasks in tree structure by search query
func (m Model) filterTaskTreeBySearch(tasks []taskmaster.Task, query string) []taskmaster.Task {
	if query == "" {
		return tasks
	}

	lowerQuery := strings.ToLower(query)
	var filtered []taskmaster.Task

	for i := range tasks {
		task := tasks[i]
		// Check if this task matches
		matches := strings.Contains(strings.ToLower(task.ID), lowerQuery) ||
			strings.Contains(strings.ToLower(task.Title), lowerQuery) ||
			strings.Contains(strings.ToLower(task.Description), lowerQuery)

		// Check if any subtask matches
		if len(task.Subtasks) > 0 {
			filteredSubtasks := m.filterTaskTreeBySearch(task.Subtasks, query)
			if len(filteredSubtasks) > 0 {
				// Keep this parent and its matching subtasks
				task.Subtasks = filteredSubtasks
				filtered = append(filtered, task)
			} else if matches {
				// This task matches but subtasks don't - keep task without subtasks
				task.Subtasks = nil
				filtered = append(filtered, task)
			}
		} else if matches {
			// Leaf task that matches
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// wrapText wraps text to a specified width
func wrapText(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}

	var lines []string
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return text
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		// Check if adding the next word exceeds the width
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			// Line is full, start a new one
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	// Don't forget the last line
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}

// renderTaskDetails renders the details of the selected task
func (m Model) renderTaskDetails() string {
	if m.selectedTask == nil {
		return m.styles.Info.Render("No task selected")
	}

	// Calculate available width for wrapping text
	// Subtract some padding to ensure text doesn't hit the edge of the viewport
	wrapWidth := m.detailsViewport.Width - 4
	if wrapWidth < 20 {
		wrapWidth = 20 // Minimum reasonable width
	}

	task := m.selectedTask
	// Re-fetch from index to ensure we have the latest stable pointer
	if freshTask, ok := m.taskIndex[task.ID]; ok {
		task = freshTask
	}
	var b strings.Builder

	// Title
	b.WriteString(m.styles.PanelTitle.Render(fmt.Sprintf("Task %s", task.ID)))
	b.WriteString("\n\n")

	// Title field
	b.WriteString(m.styles.Subtitle.Render("Title: "))
	title := wrapText(task.Title, wrapWidth-7) // "Title: " is 7 chars
	// Replace newlines with indented newlines to maintain alignment
	if strings.Contains(title, "\n") {
		lines := strings.Split(title, "\n")
		b.WriteString(lines[0]) // First line follows the "Title: " label
		for _, line := range lines[1:] {
			b.WriteString("\n       " + line) // Indent subsequent lines
		}
	} else {
		b.WriteString(title)
	}
	b.WriteString("\n\n")

	// Status
	statusStyle := m.styles.GetStatusStyle(task.Status)
	b.WriteString(m.styles.Subtitle.Render("Status: "))
	statusIndicator := GetStatusIndicator(task.Status)
	b.WriteString(statusStyle.Render(statusIndicator))
	b.WriteString("\n\n")

	// Priority
	if task.Priority != "" {
		b.WriteString(m.styles.Subtitle.Render("Priority: "))
		b.WriteString(task.Priority)
		b.WriteString("\n\n")
	}

	// Complexity
	if task.Complexity > 0 {
		b.WriteString(m.styles.Subtitle.Render("Complexity: "))
		complexityIndicator := GetComplexityIndicator(task.Complexity)
		b.WriteString(complexityIndicator)
		b.WriteString("\n\n")
	}

	// Dependencies
	if len(task.Dependencies) > 0 {
		b.WriteString(m.styles.Subtitle.Render("Dependencies: "))
		deps := strings.Join(task.Dependencies, ", ")
		wrappedDeps := wrapText(deps, wrapWidth-14) // "Dependencies: " is 14 chars
		// Replace newlines with indented newlines to maintain alignment
		if strings.Contains(wrappedDeps, "\n") {
			lines := strings.Split(wrappedDeps, "\n")
			b.WriteString(lines[0]) // First line follows the "Dependencies: " label
			for _, line := range lines[1:] {
				b.WriteString("\n              " + line) // Indent subsequent lines
			}
		} else {
			b.WriteString(wrappedDeps)
		}
		b.WriteString("\n\n")
	}

	// Description
	if task.Description != "" {
		b.WriteString(m.styles.Subtitle.Render("Description:"))
		b.WriteString("\n")
		wrappedDesc := wrapText(task.Description, wrapWidth)
		b.WriteString(wrappedDesc)
		b.WriteString("\n\n")
	}

	// Details
	if task.Details != "" {
		b.WriteString(m.styles.Subtitle.Render("Details:"))
		b.WriteString("\n")
		wrappedDetails := wrapText(task.Details, wrapWidth)
		b.WriteString(wrappedDetails)
		b.WriteString("\n\n")
	}

	// Test Strategy
	if task.TestStrategy != "" {
		b.WriteString(m.styles.Subtitle.Render("Test Strategy:"))
		b.WriteString("\n")
		wrappedStrategy := wrapText(task.TestStrategy, wrapWidth)
		b.WriteString(wrappedStrategy)
		b.WriteString("\n\n")
	}

	// Subtasks count
	if len(task.Subtasks) > 0 {
		completed := 0
		for _, subtask := range task.Subtasks {
			if subtask.Status == "done" {
				completed++
			}
		}
		b.WriteString(m.styles.Subtitle.Render("Subtasks: "))
		b.WriteString(fmt.Sprintf("%d/%d completed", completed, len(task.Subtasks)))
		b.WriteString("\n")
	}

	return b.String()
}

// updateDetailsViewport updates the details viewport content
func (m *Model) updateDetailsViewport() {
	content := m.renderTaskDetails()
	m.detailsViewport.SetContent(content)
}

// updateTaskListViewport updates the task list viewport content
func (m *Model) updateTaskListViewport() {
	if len(m.tasks) == 0 {
		m.taskListViewport.SetContent(m.styles.Info.Render("No tasks available\n\nPress 'r' to reload"))
		return
	}

	// Render title with view mode indicator
	viewModeStr := ""
	switch m.viewMode {
	case ViewModeTree:
		viewModeStr = " [Tree]"
	case ViewModeList:
		viewModeStr = " [List]"
	case ViewModeKanban:
		viewModeStr = " [Kanban]"
	}
	title := m.styles.PanelTitle.Render("📋 Tasks" + viewModeStr)

	// Render content based on view mode
	var content string
	switch m.viewMode {
	case ViewModeList:
		content = title + "\n\n" + m.renderTaskList()
	case ViewModeKanban:
		// Placeholder for future kanban view
		content = title + "\n\n" + m.styles.Info.Render("Kanban view not yet implemented")
	default: // ViewModeTree
		// Apply filters if any are active
		tasksToRender := m.tasks
		if m.searchQuery != "" || m.statusFilter != "" {
			tasksToRender = m.filterTaskTree(m.tasks)
			if len(tasksToRender) == 0 {
				content = title + "\n\n" + m.styles.Info.Render("No tasks match current filters")
			} else {
				content = title + "\n\n" + m.renderTaskTree(tasksToRender, 0)
			}
		} else {
			content = title + "\n\n" + m.renderTaskTree(tasksToRender, 0)
		}
	}

	m.taskListViewport.SetContent(content)
}

// renderTaskList renders tasks in flat list view
func (m Model) renderTaskList() string {
	var b strings.Builder

	// Use visibleTasks if any filter is active, otherwise flatten all tasks
	var tasksToRender []*taskmaster.Task
	if (m.searchQuery != "" || m.statusFilter != "") && m.visibleTasks != nil {
		tasksToRender = m.visibleTasks
	} else {
		tasksToRender = m.flattenAllTasks()
	}

	for i, task := range tasksToRender {
		statusIcon := GetStatusIcon(task.Status)
		statusStyle := m.styles.GetStatusStyle(task.Status)

		// Determine if this is the selected task
		isSelected := m.selectedTask != nil && m.selectedTask.ID == task.ID

		// Build the line
		line := ""

		// Selection checkbox
		if m.isTaskSelected(task.ID) {
			line += "[✓] "
		} else {
			line += "[ ] "
		}

		// Status icon and task info with indentation to show hierarchy
		depth := strings.Count(task.ID, ".")
		indent := strings.Repeat("  ", depth)
		line += indent + statusStyle.Render(statusIcon) + " "
		line += fmt.Sprintf("%s: %s", task.ID, task.Title)

		// Add priority if high
		if task.Priority == "high" {
			line += " " + m.styles.Warning.Render("[HIGH]")
		} else if task.Priority == "critical" {
			line += " " + m.styles.Error.Render("[CRITICAL]")
		}

		// Build full line with cursor
		var fullLine string
		if isSelected {
			fullLine = "> " + line
			fullLine = m.styles.TaskSelected.Render(fullLine)
		} else {
			fullLine = "  " + line
			fullLine = m.styles.TaskUnselected.Render(fullLine)
		}

		b.WriteString(fullLine)
		if i < len(tasksToRender)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// flattenAllTasks returns all tasks in a flat list (ignoring expanded state)
func (m Model) flattenAllTasks() []*taskmaster.Task {
	var result []*taskmaster.Task
	var flatten func(tasks []taskmaster.Task)
	flatten = func(tasks []taskmaster.Task) {
		for i := range tasks {
			// Use task index to get stable pointer instead of creating pointer to slice element
			if task, ok := m.taskIndex[tasks[i].ID]; ok {
				result = append(result, task)
				if len(task.Subtasks) > 0 {
					flatten(task.Subtasks)
				}
			}
		}
	}
	flatten(m.tasks)
	return result
}

// addLogLine adds a line to the log panel
func (m *Model) addLogLine(line string) {
	m.logLines = append(m.logLines, line)
	m.updateLogViewport()
}

// renderLog renders the log panel content
// renderLog renders the log panel content with word wrapping applied to each line.
//
// IMPORTANT - ANSI Color Code Handling:
// The current implementation counts ANSI escape sequences (e.g., \x1b[31m for red)
// toward the visual width constraint. This means that a line with 70 visible
// characters plus 10 bytes of ANSI codes will be treated as 80 bytes and wrapped
// accordingly, potentially causing visual misalignment.
//
// Example:
// - Input: "\x1b[31mLong colored text here that should wrap\x1b[0m" (60 visible chars, 70 bytes)
// - Configured wrapWidth: 60 characters
// - Actual wrap point: ~50 visible chars (when byte-count reaches 60)
//
// This is acceptable for the initial simple implementation, matching the PRD
// decision: "Start with simple implementation, add ANSI support if needed."
//
// FUTURE ENHANCEMENT:
// To improve visual alignment with colored text, consider using a library like
// github.com/charmbracelet/x/ansi to strip ANSI codes before measuring width:
//   plain := ansi.Strip(line)        // Get visual text without codes
//   width := len(plain)               // Measure actual visible width
//   wrapped := wrapText(plain, width) // Wrap based on visual width
//   reapplyColors(wrapped, line)      // Restore colors to wrapped lines
//
// This has been verified with exploratory tests in log_wrap_test.go:
// - TestRenderLogWithANSIColoredText
// - TestRenderLogWithMultipleANSIColors
// Both tests pass and confirm ANSI codes are preserved without corruption.
func (m Model) renderLog() string {
	if len(m.logLines) == 0 {
		return m.styles.Info.Render("No log output yet")
	}

	// Calculate available width for wrapping text
	wrapWidth := m.logViewport.Width - 4
	if wrapWidth < 20 {
		wrapWidth = 20 // Minimum reasonable width
	}

	// Wrap each log line individually
	var wrappedLines []string
	for _, line := range m.logLines {
		if line == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}
		wrappedLine := wrapText(line, wrapWidth)
		wrappedLines = append(wrappedLines, wrappedLine)
	}

	return strings.Join(wrappedLines, "\n")
}

// FUTURE ENHANCEMENT: ANSI-aware wrapping helper
// This commented-out helper demonstrates how to improve ANSI color handling in the future.
// Do NOT enable this without adding the charmbracelet/x/ansi dependency.
//
// // ansiWrapText wraps text while respecting ANSI escape sequences for proper visual alignment.
// // This is a future enhancement to replace the byte-based width counting in wrapText().
// //
// // Implementation would require: go get github.com/charmbracelet/x/ansi
// //
// // func ansiWrapText(text string, width int) string {
// //     // Strip ANSI codes to get visual content
// //     // This requires importing "github.com/charmbracelet/x/ansi"
// //     plain := ansi.Strip(text)      // Get text without color codes
// //     wrapped := wrapText(plain, width) // Wrap based on visual width
// //
// //     // If no ANSI codes were present, wrapped text is ready to use
// //     if text == plain {
// //         return wrapped
// //     }
// //
// //     // If ANSI codes were present, this would require reapplying them to wrapped lines
// //     // (Implementation detail omitted; this is more complex and depends on code structure)
// //     return wrapped
// // }
//
// Alternative simpler approach without a new dependency:
// 1. Create a simple pattern matcher to find ANSI sequences
// 2. Count only non-ANSI characters when checking line width
// 3. Preserve ANSI codes in output unchanged
//
// This would avoid the visual misalignment without adding external dependencies.

// updateLogViewport updates the log viewport content
func (m *Model) updateLogViewport() {
	content := m.renderLog()
	m.logViewport.SetContent(content)
	// Auto-scroll to bottom
	m.logViewport.GotoBottom()
}

// setTaskStatus sets the status of selected task(s) via executor
func (m *Model) setTaskStatus(status string) {
	if m.execService.IsRunning() {
		m.addLogLine("Command already running, please wait...")
		return
	}

	// Get tasks to update (selected or current)
	var taskIDs []string
	if len(m.selectedIDs) > 0 {
		taskIDs = m.getSelectedTasks()
	} else if m.selectedTask != nil {
		taskIDs = []string{m.selectedTask.ID}
	} else {
		m.addLogLine("No task selected")
		return
	}

	// Execute set-status command for each task
	for _, taskID := range taskIDs {
		m.addLogLine(fmt.Sprintf("Setting task %s to %s", taskID, status))
		if err := m.execService.Execute("set-status", fmt.Sprintf("--id=%s", taskID), fmt.Sprintf("--status=%s", status)); err != nil {
			m.addLogLine(fmt.Sprintf("Error: %v", err))
			return
		}
	}

	// Clear selection after status change
	m.clearSelection()
}

// Init initializes the model and starts watching for file changes
func (m Model) Init() tea.Cmd {
	// Return a batch of startup commands:
	// 1. Load tasks from disk initially
	// 2. Start watching for task file changes
	// 3. Start watching for config changes
	// 4. Start listening for executor output
	return tea.Batch(
		LoadTasksCmd(m.taskService),
		WaitForTasksReload(m.taskService),
		WaitForConfigReload(m.configManager),
		WaitForExecutorOutput(m.execService),
	)
}

// filterTasksBySearch filters tasks based on the search query
func (m *Model) filterTasksBySearch(query string) {
	// Now handled by updateFilteredTasks
	m.updateFilteredTasks()
}

// updateSearchResults updates the visible tasks based on current search query
func (m *Model) updateSearchResults() {
	m.updateFilteredTasks()
}

// cycleStatusFilter cycles through status filters
func (m *Model) cycleStatusFilter() {
	statuses := []string{
		"", // All tasks
		taskmaster.StatusPending,
		taskmaster.StatusInProgress,
		taskmaster.StatusDone,
		taskmaster.StatusBlocked,
		taskmaster.StatusDeferred,
		taskmaster.StatusCancelled,
	}

	// Find current status in the list
	currentIndex := 0
	for i, status := range statuses {
		if status == m.statusFilter {
			currentIndex = i
			break
		}
	}

	// Move to next status
	nextIndex := (currentIndex + 1) % len(statuses)
	m.statusFilter = statuses[nextIndex]

	// Log the change
	if m.statusFilter == "" {
		m.addLogLine("Filter cleared - showing all tasks")
	} else {
		m.addLogLine(fmt.Sprintf("Filtering by status: %s", m.statusFilter))
	}

	// Update the display
	m.updateFilteredTasks()
}

// updateFilteredTasks applies both search and status filters
func (m *Model) updateFilteredTasks() {
	// Start with all tasks
	allTasks := m.flattenAllTasks()

	// Apply status filter
	var statusFiltered []*taskmaster.Task
	if m.statusFilter == "" {
		statusFiltered = allTasks
	} else {
		for _, task := range allTasks {
			if task.Status == m.statusFilter {
				statusFiltered = append(statusFiltered, task)
			}
		}
	}

	// Apply search filter on top of status filter
	if m.searchQuery == "" {
		m.visibleTasks = statusFiltered
		m.searchResults = nil
	} else {
		lowerQuery := strings.ToLower(m.searchQuery)
		var searchFiltered []*taskmaster.Task

		for _, task := range statusFiltered {
			if strings.Contains(strings.ToLower(task.ID), lowerQuery) ||
				strings.Contains(strings.ToLower(task.Title), lowerQuery) ||
				strings.Contains(strings.ToLower(task.Description), lowerQuery) {
				searchFiltered = append(searchFiltered, task)
			}
		}

		m.visibleTasks = searchFiltered
		m.searchResults = searchFiltered
	}

	m.updateTaskListViewport()

	// Update selection if current selection is not visible
	if m.selectedTask != nil {
		found := false
		for _, task := range m.visibleTasks {
			if task.ID == m.selectedTask.ID {
				found = true
				break
			}
		}
		if !found && len(m.visibleTasks) > 0 {
			// Select first visible task using task index
			taskID := m.visibleTasks[0].ID
			if task, ok := m.taskIndex[taskID]; ok {
				m.selectedTask = task
			} else {
				m.selectedTask = m.visibleTasks[0]
			}
			m.selectedIndex = 0
			m.updateDetailsViewport()
		}
	}
}

// Update handles messages and updates the model
func (m Model) Update(incomingMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := incomingMsg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update viewport sizes based on layout
		m.updateViewportSizes()

		// Update dialog manager dimensions
		if dm := m.dialogManager(); dm != nil {
			dm.SetTerminalSize(msg.Width, msg.Height)
		}

		// Re-render log panel with new width if it's visible
		if m.showLogPanel {
			m.updateLogViewport()
		}

		return m, nil

	case TasksLoadedMsg:
		// Initial tasks loaded successfully
		m.tasks = msg.Tasks
		m.buildTaskIndex()

		// Select first task if available
		if len(m.tasks) > 0 {
			m.selectedTask = &m.tasks[0]
		}

		m.updateTaskListViewport()
		m.updateDetailsViewport()
		return m, nil

	case TasksReloadedMsg:
		// Tasks were reloaded from disk, refresh the view
		m.tasks, _ = m.taskService.GetTasks()
		m.buildTaskIndex()

		// Try to maintain selection after reload
		if m.selectedTask != nil {
			if task, ok := m.taskIndex[m.selectedTask.ID]; ok {
				m.selectedTask = task
			}
		}

		m.updateTaskListViewport()
		m.updateDetailsViewport()
		m.addLogLine("Tasks reloaded from disk")

		// Continue listening for next reload
		return m, WaitForTasksReload(m.taskService)

	case ConfigReloadedMsg:
		// Config was reloaded from disk, update local reference
		m.config = m.configManager.GetConfig()
		m.addLogLine("Configuration reloaded")

		// Continue listening for next reload
		return m, WaitForConfigReload(m.configManager)

	case ExecutorOutputMsg:
		// Executor produced output, add to log
		m.addLogLine(msg.Line)

		// Continue listening for next output
		return m, WaitForExecutorOutput(m.execService)

	case CommandCompletedMsg:
		// Command execution completed, log it and potentially reload tasks
		if msg.Success {
			m.addLogLine(fmt.Sprintf("✓ Command '%s' completed successfully", msg.Command))
		} else {
			m.addLogLine(fmt.Sprintf("✗ Command '%s' failed", msg.Command))
		}

		// Reload tasks after command completion
		cmds = append(cmds, LoadTasksCmd(m.taskService))
		return m, tea.Batch(cmds...)

	case ErrorMsg:
		// Handle errors by storing them and displaying in UI
		m.err = msg.Err
		m.addLogLine(fmt.Sprintf("Error: %v", msg.Err))
		return m, nil

	case WatcherErrorMsg:
		// Handle watcher errors (just log for now)
		m.err = msg.Err
		m.addLogLine(fmt.Sprintf("Watcher error: %v", msg.Err))
		return m, nil

	// Complexity Analysis Messages
	case AnalyzeTaskComplexityMsg:
		// Start complexity analysis by opening scope selection dialog
		m.showComplexityScopeDialog()
		return m, nil

	// Tag Dialog Messages
	case ComplexityScopeSelectedMsg:
		// Handle selected complexity scope
		cmd := m.handleComplexityScopeSelected(msg)
		return m, cmd

	case ComplexityAnalysisProgressMsg:
		// Update progress dialog
		if cmd := m.handleComplexityAnalysisProgress(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if waitCmd := m.waitForComplexityMessages(); waitCmd != nil {
			cmds = append(cmds, waitCmd)
		}
		return m, tea.Batch(cmds...)

	case ComplexityAnalysisCompletedMsg:
		// Show complexity report
		if delay := m.complexityCompletionDelay(msg); delay > 0 {
			m.waitingForComplexityHold = true
			m.showComplexityCompletionMessage(msg)
			msgCopy := msg
			return m, tea.Tick(delay, func(time.Time) tea.Msg {
				return msgCopy
			})
		}
		m.waitingForComplexityHold = false
		if cmd := m.handleComplexityAnalysisCompleted(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case parsePrdResultMsg:
		if cmd := m.handleParsePrdResult(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if wait := m.waitForParsePrdMessages(); wait != nil {
			cmds = append(cmds, wait)
		}
		return m, tea.Batch(cmds...)

	case parsePrdStreamClosedMsg:
		m.clearParsePrdRuntimeState()
		return m, nil

	case complexityStreamClosedMsg:
		m.complexityMsgCh = nil
		return m, nil

	case ExpandTaskProgressMsg:
		// Update progress dialog
		if dm := m.dialogManager(); dm != nil {
			if dlg, ok := dm.GetDialogByType(dialog.DialogTypeProgress); ok {
				if pd, ok := dlg.(*dialog.ProgressDialog); ok {
					pd.SetLabel(msg.Stage)
					pd.SetProgress(msg.Progress)
				}
			}
		}
		// Wait for next message
		if wait := m.waitForExpandTaskMessages(); wait != nil {
			cmds = append(cmds, wait)
		}
		return m, tea.Batch(cmds...)

	case ExpandTaskCompletedMsg:
		m.clearExpandTaskRuntimeState()
		if msg.Error != nil {
			appErr := NewOperationError("Expand Task", "Task expansion failed", msg.Error).
				WithRecoveryHints(
					"Check the task description format",
					"Verify all required fields are present",
					"Try again",
				)
			m.showAppError(appErr)
		} else {
			m.addLogLine("Task expanded successfully")
			// Dismiss progress dialog
			if dm := m.dialogManager(); dm != nil {
				dm.PopDialog()
			}
			// Reload tasks
			cmds = append(cmds, LoadTasksCmd(m.taskService))
		}
		return m, tea.Batch(cmds...)

	case expandTaskStreamClosedMsg:
		m.expandTaskMsgCh = nil
		return m, nil

	case ComplexityReportActionMsg:
		// Handle actions from the complexity report dialog
		cmd := m.handleComplexityReportAction(msg)
		return m, cmd

	case ComplexityFilterAppliedMsg:
		// Apply filter settings to report
		cmd := m.handleComplexityFilterApplied(msg)
		return m, cmd

	case ComplexityExportRequestMsg:
		// Handle export request
		cmd := m.handleComplexityExportRequest(msg)
		return m, cmd

	case ComplexityExportCompletedMsg:
		// Handle export completion
		cmd := m.handleComplexityExportCompleted(msg)
		return m, cmd

	// New expansion workflow messages
	case ExpansionScopeSelectedMsg:
		// Handle selected expansion scope
		cmd := m.handleExpansionScopeSelected(msg)
		return m, cmd

	case ExpansionProgressMsg:
		// Update progress dialog during expansion
		if cmd := m.handleExpansionProgress(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case ExpansionCompletedMsg:
		// Handle expansion completion
		if cmd := m.handleExpansionCompleted(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case expansionStreamClosedMsg:
		m.expansionMsgCh = nil
		return m, nil

	case dialog.DialogResultMsg:
		if cmd := m.handleDialogResultMsg(msg); cmd != nil {
			return m, cmd
		}
		return m, nil

	case UndoTickMsg:
		if cmd := m.handleUndoTick(msg); cmd != nil {
			return m, cmd
		}
		return m, nil

	case UndoExpiredMsg:
		m.handleUndoExpired(msg)
		return m, nil

	case dialog.ProgressCancelMsg:
		cmd := m.handleProgressDialogCancel()
		return m, cmd

	case SelectTaskMsg:
		if msg.TaskID != "" {
			if !m.selectTaskByID(msg.TaskID) {
				m.addLogLine(fmt.Sprintf("Task %s not found", msg.TaskID))
			}
		}
		return m, nil

	case dialog.ListSelectionMsg:
		if cmd := m.handleListSelection(msg); cmd != nil {
			return m, cmd
		}
		return m, nil
	case dialog.ModelSelectionMsg:
		if cmd := m.handleModelSelection(msg); cmd != nil {
			return m, cmd
		}
		return m, nil
	case tagListLoadedMsg:
		if msg.Err != nil {
			appErr := NewOperationError("Tag Contexts", "Failed to load tag contexts", msg.Err).
				WithRecoveryHints(
					"Check if Task Master is properly configured",
					"Verify you're in a Task Master workspace",
					"Try again",
				)
			m.showAppError(appErr)
			return m, nil
		}
		m.showTagListDialog(msg.List)
		return m, nil
	case TagOperationMsg:
		if cmd := m.handleTagOperationMsg(msg); cmd != nil {
			return m, cmd
		}
		return m, nil
	case projectSwitchedMsg:
		m.pendingProjectSwitch = nil
		m.pendingProjectTag = ""
		if msg.Err != nil {
			appErr := NewOperationError("Switch Project", "Failed to switch project", msg.Err).
				WithRecoveryHints(
					"Check if the project directory is valid",
					"Verify project configuration",
					"Try again",
				)
			m.showAppError(appErr)
			return m, nil
		}
		if msg.Meta != nil {
			m.activeProject = msg.Meta
			m.projectRegistry = m.taskService.ProjectRegistry()
			m.addLogLine(fmt.Sprintf("Switched to project %s", msg.Meta.Name))
		}
		cmds = append(cmds, LoadTasksCmd(m.taskService))
		tag := strings.TrimSpace(msg.Tag)
		if tag == "" && msg.Meta != nil {
			tag = msg.Meta.PrimaryTag()
		}
		if tag != "" {
			if cmd := m.useProjectTag(tag); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	// TaskRunnerModal message handlers
	case dialog.TaskStartedMsg:
		m.taskRunnerVisible = true
		if m.taskRunner != nil {
			updatedDialog, cmd := m.taskRunner.Update(msg)
			if updatedDialog != nil {
				m.taskRunner = updatedDialog.(*dialog.TaskRunnerModal)
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Start ticker for UI updates (elapsed time, etc.)
		cmds = append(cmds, TickCmd())
		return m, tea.Batch(cmds...)

	case dialog.CrushExecutionSub:
		// Store the output channel for this task
		m.crushRunChannels[msg.TaskID] = msg.OutCh
		// Start listening for messages from this channel
		cmds = append(cmds, dialog.WaitForCrushMsg(msg.OutCh))
		return m, tea.Batch(cmds...)

	case dialog.CrushExecutionContextMsg:
		// Set the cancellation context on the appropriate tab
		if m.taskRunner != nil {
			tab := m.taskRunner.GetTabByTaskID(msg.TaskID)
			if tab != nil {
				tab.SetCancellationContext(msg.Cmd, msg.CancelFunc)
			}
		}
		// Continue listening for more messages from this task's channel
		if ch, ok := m.crushRunChannels[msg.TaskID]; ok {
			cmds = append(cmds, dialog.WaitForCrushMsg(ch))
		}
		return m, tea.Batch(cmds...)

	case dialog.TaskOutputMsg:
		m.addLogLine(fmt.Sprintf("[Task %s] %s", msg.TaskID, msg.Output))
		if m.taskRunner != nil {
			updatedDialog, cmd := m.taskRunner.Update(msg)
			if updatedDialog != nil {
				m.taskRunner = updatedDialog.(*dialog.TaskRunnerModal)
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Continue listening for more output from this task's channel
		if ch, ok := m.crushRunChannels[msg.TaskID]; ok {
			cmds = append(cmds, dialog.WaitForCrushMsg(ch))
		}
		return m, tea.Batch(cmds...)

	case dialog.TaskCompletedMsg:
		if m.taskRunner != nil {
			updatedDialog, cmd := m.taskRunner.Update(msg)
			if updatedDialog != nil {
				m.taskRunner = updatedDialog.(*dialog.TaskRunnerModal)
			}
			// Hide modal if no tasks are running
			if !m.taskRunner.HasRunningTasks() {
				m.taskRunnerVisible = false
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Clean up the channel for this task
		delete(m.crushRunChannels, msg.TaskID)
		return m, tea.Batch(cmds...)

	case dialog.TaskFailedMsg:
		if m.taskRunner != nil {
			updatedDialog, cmd := m.taskRunner.Update(msg)
			if updatedDialog != nil {
				m.taskRunner = updatedDialog.(*dialog.TaskRunnerModal)
			}
			// Hide modal if no tasks are running
			if !m.taskRunner.HasRunningTasks() {
				m.taskRunnerVisible = false
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Clean up the channel for this task
		delete(m.crushRunChannels, msg.TaskID)
		m.addLogLine(fmt.Sprintf("Task %s failed: %s", msg.TaskID, msg.Error))
		return m, tea.Batch(cmds...)

	case dialog.TaskCancelledMsg:
		if m.taskRunner != nil {
			updatedDialog, cmd := m.taskRunner.Update(msg)
			if updatedDialog != nil {
				m.taskRunner = updatedDialog.(*dialog.TaskRunnerModal)
			}
			// Hide modal if no tasks are running
			if !m.taskRunner.HasRunningTasks() {
				m.taskRunnerVisible = false
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Clean up the channel for this task
		delete(m.crushRunChannels, msg.TaskID)
		m.addLogLine(fmt.Sprintf("Task %s cancelled", msg.TaskID))
		return m, tea.Batch(cmds...)

	case TickMsg:
		// Handle periodic UI updates (for elapsed time, etc.)
		// Continue ticking only if there are running tasks
		if m.taskRunner != nil && m.taskRunner.HasRunningTasks() {
			return m, TickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		// Handle help overlay mode first - takes priority
		if m.showHelp {
			switch msg.String() {
			case "?", "esc":
				m.showHelp = false
				return m, nil
			default:
				// Ignore other keys when help is showing
				return m, nil
			}
		}

		// Handle dialog-specific confirmation results
		switch msgTyped := incomingMsg.(type) {
		case dialog.ConfirmationMsg:
			if msgTyped.Result == dialog.ConfirmationResultYes {
				// Clear state
				if err := m.ClearUIState(); err != nil {
					m.addLogLine(fmt.Sprintf("Failed to clear state: %v", err))
				} else {
					m.addLogLine("TUI state cleared successfully")
				}
			} else {
				m.addLogLine("State clear cancelled")
			}
			return m, nil
		}

		// Handle command mode separately
		if m.commandMode {
			switch msg.String() {
			case "esc":
				m.commandMode = false
				m.commandInput = ""
			case "enter":
				// Process command (jump to ID)
				if m.commandInput != "" {
					if m.selectTaskByID(m.commandInput) {
						m.addLogLine(fmt.Sprintf("Jumped to task %s", m.commandInput))
						m.updateTaskListViewport()
						m.updateDetailsViewport()
					} else {
						m.addLogLine(fmt.Sprintf("Task %s not found", m.commandInput))
					}
				}
				m.commandMode = false
				m.commandInput = ""
			case "backspace":
				if len(m.commandInput) > 0 {
					m.commandInput = m.commandInput[:len(m.commandInput)-1]
				}
			default:
				// Append character to command input
				m.commandInput += msg.String()
			}
			return m, nil
		}

		// Handle task runner modal keyboard input when visible
		if m.taskRunnerVisible && m.taskRunner != nil {
			updatedDialog, cmd := m.taskRunner.Update(msg)
			if updatedDialog != nil {
				m.taskRunner = updatedDialog.(*dialog.TaskRunnerModal)
			}
			// Check if modal should be closed (returns nil when closed)
			if updatedDialog == nil {
				m.taskRunnerVisible = false
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Modal handled the key, don't pass to main app UNLESS minimized
			// When minimized, allow keys to pass through so user can interact with main UI
			if !m.taskRunner.GetMinimized() {
				return m, tea.Batch(cmds...)
			}
			// If minimized, fall through to main app key handling
		}

		// Handle search mode separately
		if m.searchMode {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)

			switch msg.String() {
			case "esc":
				// Exit search mode
				m.searchMode = false
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.searchResults = nil
				m.visibleTasks = nil
				m.updateTaskListViewport()
				m.addLogLine("Search cancelled")
			case "enter":
				// Confirm search
				m.searchQuery = m.searchInput.Value()

				// Only exit search mode if there's actual input
				if m.searchQuery != "" {
					m.searchMode = false
					m.updateSearchResults()
					if len(m.searchResults) == 0 {
						m.addLogLine("No tasks found matching search query")
					} else {
						m.addLogLine(fmt.Sprintf("Found %d tasks matching '%s'", len(m.searchResults), m.searchQuery))
					}
				}
			default:
				// Update search results as user types
				m.searchQuery = m.searchInput.Value()
				m.updateSearchResults()
			}
			return m, cmd
		}

		// Check if we have any active dialogs - let dialogue manager handle messages first
		if m.appState != nil && m.appState.HasActiveDialog() {
			// Log that dialog is active and blocking shortcuts
			m.addLogLine("DEBUG: Active dialog is blocking key events")
			
			// Let dialog manager handle the message
			cmd := m.appState.HandleDialogMsg(incomingMsg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			// Check for dialog-specific messages
			switch msgTyped := incomingMsg.(type) {
			case dialog.ListSelectionMsg:
				if cmd := m.handleListSelection(msgTyped); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case dialog.ConfirmationMsg:
				// Handle confirmation result
				if msgTyped.Result == dialog.ConfirmationResultYes {
					m.addLogLine("Confirmed: Yes")
				} else {
					m.addLogLine("Confirmed: No")
				}
			case dialog.ProgressUpdateMsg:
				// Update progress bar
				m.addLogLine(fmt.Sprintf("Progress: %.0f%%", msgTyped.Progress*100))
				if m.parsePrdChan != nil {
					if wait := m.waitForParsePrdMessages(); wait != nil {
						cmds = append(cmds, wait)
					}
				}
			}

			// If message is a key event and we have active dialogs,
			// return immediately - dialogs get priority for key events
			switch msgTyped := incomingMsg.(type) {
			case tea.KeyMsg:
				_ = msgTyped // Use the variable to avoid compiler error
				return m, tea.Batch(cmds...)
			}
		}

		if cmd, handled := m.tryHandleCommandShortcut(msg); handled {
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		}

		// Normal mode keyboard handling
		if keyMsg, ok := incomingMsg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keyMap.Quit):
				// Save UI state before quitting
				if err := m.SaveUIState(); err != nil {
					m.addLogLine(fmt.Sprintf("Warning: failed to save UI state: %v", err))
				}
				return m, tea.Quit

			case key.Matches(msg, m.keyMap.Cancel):
				// Cancel running command or quit
				if m.execService.IsRunning() {
					if err := m.execService.Cancel(); err == nil {
						m.addLogLine("Command cancelled")
					}
				} else {
					// Save UI state before quitting
					if err := m.SaveUIState(); err != nil {
						m.addLogLine(fmt.Sprintf("Warning: failed to save UI state: %v", err))
					}
					return m, tea.Quit
				}

			case key.Matches(msg, m.keyMap.Up):
				m.selectPrevious()
				m.updateTaskListViewport()
				m.updateDetailsViewport()

			case key.Matches(msg, m.keyMap.Down):
				m.selectNext()
				m.updateTaskListViewport()
				m.updateDetailsViewport()

			case key.Matches(msg, m.keyMap.Left):
				m.collapseSelected()
				m.updateTaskListViewport()

			case key.Matches(msg, m.keyMap.Right):
				m.expandSelected()
				m.updateTaskListViewport()

			case key.Matches(msg, m.keyMap.ToggleExpand):
				m.toggleExpanded()
				m.updateTaskListViewport()

			case key.Matches(msg, m.keyMap.Select):
				m.toggleSelection()
				m.updateTaskListViewport()

			case key.Matches(msg, m.keyMap.Refresh):
				// Reload tasks manually
				m.addLogLine("Manually reloading tasks...")
				cmds = append(cmds, LoadTasksCmd(m.taskService))

			case key.Matches(msg, m.keyMap.NextTask):
				// Execute task-master next command via executor
				if !m.execService.IsRunning() {
					m.addLogLine("Executing: task-master next")
					if err := m.execService.Execute("next"); err != nil {
						m.addLogLine(fmt.Sprintf("Error: %v", err))
					}
				} else {
					m.addLogLine("Command already running")
				}

			case key.Matches(msg, m.keyMap.JumpToID):
				// Enter command mode for quick jump
				m.commandMode = true
				m.commandInput = ""
				m.addLogLine("Jump to task ID: (type ID and press Enter)")

			case key.Matches(msg, m.keyMap.Search):
				// Enter search mode
				m.searchMode = true
				m.searchInput.Focus()

				// Update input with previous query if it exists
				if m.searchQuery != "" {
					m.searchInput.SetValue(m.searchQuery)
					// Select all text so typing immediately replaces it
					m.searchInput.CursorEnd()
					// This would ideally select all text, but the API doesn't directly support it
					// We position cursor at end for better UX
				} else {
					m.searchInput.SetValue("")
				}

				m.addLogLine("Search: (type query and press Enter, Esc to cancel)")
				return m, textinput.Blink

			case key.Matches(msg, m.keyMap.Filter):
				// Cycle through status filters
				m.cycleStatusFilter()

			case key.Matches(msg, m.keyMap.SetInProgress):
				// Set task(s) to in-progress
				m.setTaskStatus("in-progress")

			case key.Matches(msg, m.keyMap.SetDone):
				// Set task(s) to done
				m.setTaskStatus("done")

			case key.Matches(msg, m.keyMap.SetBlocked):
				// Set task(s) to blocked
				m.setTaskStatus("blocked")

			case key.Matches(msg, m.keyMap.SetCancelled):
				// Set task(s) to cancelled
				m.setTaskStatus("cancelled")

			case key.Matches(msg, m.keyMap.SetDeferred):
				// Set task(s) to deferred
				m.setTaskStatus("deferred")

			case key.Matches(msg, m.keyMap.SetPending):
				// Set task(s) to pending
				m.setTaskStatus("pending")

			case key.Matches(msg, m.keyMap.CyclePanel):
				// Cycle focus between panels
				switch m.focusedPanel {
				case PanelTaskList:
					if m.showDetailsPanel {
						m.focusedPanel = PanelDetails
					} else if m.showLogPanel {
						m.focusedPanel = PanelLog
					}
				case PanelDetails:
					if m.showLogPanel {
						m.focusedPanel = PanelLog
					} else {
						m.focusedPanel = PanelTaskList
					}
				case PanelLog:
					m.focusedPanel = PanelTaskList
				}

			case key.Matches(msg, m.keyMap.AnalyzeComplexity):
				// Start complexity analysis by opening scope selection dialog
				m.addLogLine("Analyzing task complexity...")
				m.showComplexityScopeDialog()

			case key.Matches(msg, m.keyMap.ToggleDetails):
				// Toggle details panel
				m.showDetailsPanel = !m.showDetailsPanel
				m.updateViewportSizes()

			case key.Matches(msg, m.keyMap.ToggleLog):
				// Toggle log panel
				m.showLogPanel = !m.showLogPanel
				m.updateViewportSizes()
				// Update log viewport when showing the log panel
				if m.showLogPanel {
					m.updateLogViewport()
				}

			case key.Matches(msg, m.keyMap.CommandPalette):
				m.openCommandPalette()
				return m, nil

			case key.Matches(msg, m.keyMap.ViewTree):
				// Switch to tree view
				if m.viewMode != ViewModeTree {
					selectedID := ""
					if m.selectedTask != nil {
						selectedID = m.selectedTask.ID
					}
					m.viewMode = ViewModeTree
					m.rebuildVisibleTasks()
					if selectedID != "" {
						m.ensureTaskSelected(selectedID)
					}
					m.updateTaskListViewport()
					m.addLogLine("Switched to tree view")
				}

			case key.Matches(msg, m.keyMap.ViewList):
				// Switch to list view
				if m.viewMode != ViewModeList {
					selectedID := ""
					if m.selectedTask != nil {
						selectedID = m.selectedTask.ID
					}
					m.viewMode = ViewModeList
					m.rebuildVisibleTasks()
					if selectedID != "" {
						m.ensureTaskSelected(selectedID)
					}
					m.updateTaskListViewport()
					m.addLogLine("Switched to list view")
				}

			case key.Matches(msg, m.keyMap.CycleView):
				// Cycle through view modes
				selectedID := ""
				if m.selectedTask != nil {
					selectedID = m.selectedTask.ID
				}

				switch m.viewMode {
				case ViewModeTree:
					m.viewMode = ViewModeList
					m.addLogLine("Switched to list view")
				case ViewModeList:
					// Skip kanban for now since it's not implemented
					m.viewMode = ViewModeTree
					m.addLogLine("Switched to tree view")
				case ViewModeKanban:
					m.viewMode = ViewModeTree
					m.addLogLine("Switched to tree view")
				}

				m.rebuildVisibleTasks()
				if selectedID != "" {
					m.ensureTaskSelected(selectedID)
				}
				m.updateTaskListViewport()

			case key.Matches(msg, m.keyMap.Help):
				// Toggle help
				m.showHelp = !m.showHelp

			case key.Matches(msg, m.keyMap.Back):
				// Clear search and/or filter if active
				cleared := false
				if m.searchQuery != "" {
					m.searchQuery = ""
					m.searchInput.SetValue("")
					m.searchResults = nil
					cleared = true
				}
				if m.statusFilter != "" {
					m.statusFilter = ""
					cleared = true
				}
				if cleared {
					m.visibleTasks = nil
					m.updateFilteredTasks()
					m.addLogLine("Search and filters cleared")
				}

			case key.Matches(keyMsg, m.keyMap.ClearState):
				// Show confirmation dialog for clearing state
				m.ShowConfirmationDialog("Clear UI State", "Are you sure you want to reset all UI state to defaults?", true)
			}
		}
	}

	// Handle viewport scrolling when focused
	if m.focusedPanel == PanelTaskList {
		var cmd tea.Cmd
		m.taskListViewport, cmd = m.taskListViewport.Update(incomingMsg)
		cmds = append(cmds, cmd)
	} else if m.focusedPanel == PanelDetails {
		var cmd tea.Cmd
		m.detailsViewport, cmd = m.detailsViewport.Update(incomingMsg)
		cmds = append(cmds, cmd)
	} else if m.focusedPanel == PanelLog {
		var cmd tea.Cmd
		m.logViewport, cmd = m.logViewport.Update(incomingMsg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return m.styles.Info.Render("Initializing Task Master TUI...")
	}

	// Check for active dialogs - display them over everything else
	if dm := m.dialogManager(); dm != nil && dm.HasDialogs() {
		// Render main UI, then overlay dialogs
		var sections []string

		// 1. Header
		sections = append(sections, m.renderHeader())

		// 2. Main content area
		layout := m.calculateLayout()
		mainContent := m.renderMainContent(layout)
		sections = append(sections, mainContent)

		// 3. Log panel (if visible)
		if m.showLogPanel {
			sections = append(sections, m.renderLogPanel(layout))
		}

		// 4. Status bar with help
		sections = append(sections, m.renderStatusBar())

		// Render the base UI first (not used for dialogs)
		// Then overlay the active dialog
		dialogContent := dm.View()

		// For now, we just render the dialog content over the middle of the screen
		// A more sophisticated implementation would overlay it properly
		dialogStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center)

		return dialogStyle.Render(dialogContent)
	}

	// Help overlay takes priority over everything except dialogs
	if m.showHelp {
		return m.renderHelpOverlay()
	}

	// Check if taskmaster is not available
	if len(m.tasks) == 0 && !m.taskService.IsAvailable() {
		return m.renderNoTaskmaster()
	}

	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err))
	}

	// Calculate layout
	layout := m.calculateLayout()

	var sections []string

	// 1. Header
	sections = append(sections, m.renderHeader())

	// 2. Main content area (task list + details)
	mainContent := m.renderMainContent(layout)
	sections = append(sections, mainContent)

	// 3. Log panel (if visible)
	if m.showLogPanel {
		sections = append(sections, m.renderLogPanel(layout))
	}

	// 4. Status bar with help
	sections = append(sections, m.renderStatusBar())

	// Join all sections
	baseContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// 5. TaskRunnerModal overlay (if visible)
	if m.taskRunnerVisible && m.taskRunner != nil {
		modalContent := m.taskRunner.View()
		// Overlay the modal on top of the base content
		return baseContent + "\n" + modalContent
	}

	return baseContent
}

// renderNoTaskmaster displays a message when .taskmaster is not found
func (m Model) renderNoTaskmaster() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Task Master TUI"))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Warning.Render("⚠ No .taskmaster directory found"))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Info.Render("To get started with Task Master:"))
	b.WriteString("\n\n")
	b.WriteString("  1. Initialize Task Master in your project:\n")
	b.WriteString("     task-master init\n\n")
	b.WriteString("  2. Create or import tasks:\n")
	b.WriteString("     task-master parse-prd .taskmaster/docs/prd.txt\n\n")
	b.WriteString("  3. Restart the TUI\n\n")
	b.WriteString(m.styles.Help.Render("Press 'q' to quit"))

	return b.String()
}

// renderMainContent renders the main content area with task list and details
func (m Model) renderMainContent(layout LayoutDimensions) string {
	// Get task list content from viewport
	taskListContent := m.taskListViewport.View()

	// Create task list panel with border - let it naturally fit viewport content
	taskListPanel := m.styles.Panel.Render(taskListContent)

	// Add focus indicator for task list
	if m.focusedPanel == PanelTaskList {
		taskListPanel = m.styles.PanelBorder.
			BorderForeground(lipgloss.Color(ColorHighlight)).
			Render(taskListPanel)
	}

	// If details panel is hidden, return just the task list
	if !m.showDetailsPanel {
		return taskListPanel
	}

	// Render details panel - let it naturally fit viewport content
	detailsContent := m.detailsViewport.View()
	detailsPanel := m.styles.Panel.Render(detailsContent)

	// Add focus indicator for details panel
	if m.focusedPanel == PanelDetails {
		detailsPanel = m.styles.PanelBorder.
			BorderForeground(lipgloss.Color(ColorHighlight)).
			Render(detailsPanel)
	}

	// Join task list and details side by side
	return lipgloss.JoinHorizontal(lipgloss.Top, taskListPanel, detailsPanel)
}

// renderLogPanel renders the log panel
func (m Model) renderLogPanel(layout LayoutDimensions) string {
	title := m.styles.PanelTitle.Render("📝 Log")
	logContent := title + "\n\n" + m.logViewport.View()

	// Let panel naturally fit viewport content
	logPanel := m.styles.Panel.Render(logContent)

	// Add focus indicator for log panel
	if m.focusedPanel == PanelLog {
		logPanel = m.styles.PanelBorder.
			BorderForeground(lipgloss.Color(ColorHighlight)).
			Render(logPanel)
	}

	return logPanel
}

// renderHelpOverlay renders a help overlay on top of the main content using bubbles/help
func (m Model) renderHelpOverlay() string {
	var sections []string

	// Title
	title := m.styles.Title.Render("📚 Task Master TUI Help")
	sections = append(sections, title)
	sections = append(sections, "")

	// Navigation section
	navSection := m.styles.Subtitle.Render("Navigation")
	navHelp := []string{
		formatHelpLine(m.renderBinding(m.keyMap.Up), "Move up"),
		formatHelpLine(m.renderBinding(m.keyMap.Down), "Move down"),
		formatHelpLine(m.renderBinding(m.keyMap.Left), "Collapse/Move left"),
		formatHelpLine(m.renderBinding(m.keyMap.Right), "Expand/Move right"),
		formatHelpLine(m.renderBinding(m.keyMap.PageUp), "Page up"),
		formatHelpLine(m.renderBinding(m.keyMap.PageDown), "Page down"),
	}
	sections = append(sections, navSection)
	sections = append(sections, strings.Join(navHelp, "\n"))
	sections = append(sections, "")

	// Task Operations section
	taskSection := m.styles.Subtitle.Render("Task Operations")
	taskHelp := []string{
		formatHelpLine(m.renderBinding(m.keyMap.ToggleExpand), "Toggle expand/collapse"),
		formatHelpLine(m.renderBinding(m.keyMap.Select), "Select/deselect for bulk operations"),
		formatHelpLine(m.renderBinding(m.keyMap.NextTask), "Get next available task"),
		formatHelpLine(m.renderBinding(m.keyMap.Refresh), "Refresh tasks from disk"),
		formatHelpLine(m.renderBinding(m.keyMap.JumpToID), "Jump to task by ID"),
	}
	sections = append(sections, taskSection)
	sections = append(sections, strings.Join(taskHelp, "\n"))
	sections = append(sections, "")

	// Status Changes section
	statusSection := m.styles.Subtitle.Render("Status Changes")
	statusHelp := []string{
		formatHelpLine(m.renderBinding(m.keyMap.SetInProgress), m.styles.InProgress.Render("► Set in-progress")),
		formatHelpLine(m.renderBinding(m.keyMap.SetDone), m.styles.Done.Render("✓ Set done")),
		formatHelpLine(m.renderBinding(m.keyMap.SetBlocked), m.styles.Blocked.Render("! Set blocked")),
		formatHelpLine(m.renderBinding(m.keyMap.SetCancelled), m.styles.Cancelled.Render("✗ Set cancelled")),
		formatHelpLine(m.renderBinding(m.keyMap.SetDeferred), m.styles.Deferred.Render("⏱ Set deferred")),
		formatHelpLine(m.renderBinding(m.keyMap.SetPending), m.styles.Pending.Render("○ Set pending")),
	}
	sections = append(sections, statusSection)
	sections = append(sections, strings.Join(statusHelp, "\n"))
	sections = append(sections, "")

	// Panel & View section
	panelSection := m.styles.Subtitle.Render("Panels & Views")
	panelHelp := []string{
		formatHelpLine(m.renderBinding(m.keyMap.FocusTaskList), "Focus task list panel"),
		formatHelpLine(m.renderBinding(m.keyMap.FocusDetails), "Focus details panel"),
		formatHelpLine(m.renderBinding(m.keyMap.FocusLog), "Focus log panel"),
		formatHelpLine(m.renderBinding(m.keyMap.CyclePanel), "Cycle through panels"),
		formatHelpLine(m.renderBinding(m.keyMap.ToggleDetails), "Toggle details panel"),
		formatHelpLine(m.renderBinding(m.keyMap.ToggleLog), "Toggle log panel"),
		formatHelpLine(m.renderBinding(m.keyMap.ViewTree), "Switch to tree view"),
		formatHelpLine(m.renderBinding(m.keyMap.ViewList), "Switch to list view"),
		formatHelpLine(m.renderBinding(m.keyMap.CycleView), "Cycle view modes"),
	}
	sections = append(sections, panelSection)
	sections = append(sections, strings.Join(panelHelp, "\n"))
	sections = append(sections, "")

	// General section
	generalSection := m.styles.Subtitle.Render("General")
	generalHelp := []string{
		formatHelpLine(m.renderBinding(m.keyMap.Help), "Toggle this help"),
		formatHelpLine(m.renderBinding(m.keyMap.ClearState), "Clear TUI state (reset UI)"),
		formatHelpLine(m.renderBinding(m.keyMap.Back), "Back/Cancel/Close"),
		formatHelpLine(m.renderBinding(m.keyMap.Quit), "Quit application"),
		formatHelpLine(m.renderBinding(m.keyMap.Cancel), "Cancel command or quit"),
	}
	sections = append(sections, generalSection)
	sections = append(sections, strings.Join(generalHelp, "\n"))
	sections = append(sections, "")

	// About section
	aboutSection := m.styles.Subtitle.Render("About")
	aboutHelp := []string{
		"  Task Master TUI - Terminal interface for Task Master AI",
		"  Documentation: https://github.com/task-master-ai/tm-tui",
		"  Version: 1.0.0",
	}
	sections = append(sections, aboutSection)
	sections = append(sections, strings.Join(aboutHelp, "\n"))
	sections = append(sections, "")

	// Footer
	footer := m.styles.Help.Render("Press '?' or 'Esc' to close help")
	sections = append(sections, footer)

	// Join all sections
	content := strings.Join(sections, "\n")

	// Create overlay style with distinct border and centering
	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ColorHighlight)).
		Padding(1, 2).
		Width(80).
		Align(lipgloss.Center)

	// Create backdrop to ensure overlay appears on top
	backdrop := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return backdrop.Render(overlayStyle.Render(content))
}

// renderBinding formats a key binding for display
func (m Model) renderBinding(binding key.Binding) string {
	if len(binding.Keys()) > 0 {
		keys := binding.Keys()
		formatted := make([]string, len(keys))
		for i, k := range keys {
			// Special handling for space key
			if k == " " {
				formatted[i] = m.styles.Key.Render("space")
			} else {
				formatted[i] = m.styles.Key.Render(k)
			}
		}
		if len(formatted) == 1 {
			return formatted[0]
		}
		// Show multiple keys with "/" separator
		return strings.Join(formatted, "/")
	}
	return ""
}

func (m *Model) complexityCompletionDelay(msg ComplexityAnalysisCompletedMsg) time.Duration {
	if m.waitingForComplexityHold || msg.Error != nil {
		return 0
	}
	delay := complexityProgressCompleteHold
	if !m.complexityStartedAt.IsZero() {
		elapsed := time.Since(m.complexityStartedAt)
		if elapsed < complexityProgressMinDisplay {
			delay += complexityProgressMinDisplay - elapsed
		}
	}
	if delay < 0 {
		return 0
	}
	return delay
}

func (m *Model) showComplexityCompletionMessage(msg ComplexityAnalysisCompletedMsg) {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	dlg, ok := dm.GetDialogByType(dialog.DialogTypeProgress)
	if !ok {
		return
	}
	pd, ok := dlg.(*dialog.ProgressDialog)
	if !ok {
		return
	}
	label := "Preparing complexity report..."
	if msg.Error != nil {
		label = fmt.Sprintf("Analysis failed: %v", msg.Error)
	} else if msg.Report != nil {
		count := len(msg.Report.Tasks)
		label = fmt.Sprintf("Analysis complete • %d task(s) analyzed\nPreparing report...", count)
	}
	pd.SetProgress(1.0)
	pd.SetLabel(label)
}

// truncateString truncates a string to the given length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
