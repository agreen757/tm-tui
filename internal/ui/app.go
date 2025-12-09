package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/executor"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view mode of the task list
type ViewMode int

const (
	ViewModeTree ViewMode = iota
	ViewModeList
	ViewModeKanban // Placeholder for future implementation
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
	taskService   *taskmaster.Service
	execService   *executor.Service
	
	// Task data
	tasks         []taskmaster.Task
	taskIndex     map[string]*taskmaster.Task // Quick lookup by ID
	visibleTasks  []*taskmaster.Task          // Flattened list of visible tasks (respects expand/collapse)
	
	// View state
	viewMode      ViewMode
	focusedPanel  Panel
	selectedIndex int                         // Index in visibleTasks array
	selectedTask  *taskmaster.Task
	expandedNodes map[string]bool
	selectedIDs   map[string]bool              // Track multi-select for bulk operations
	
	// Layout
	width         int
	height        int
	ready         bool
	
	// Panels
	taskListViewport viewport.Model
	detailsViewport viewport.Model
	logViewport     viewport.Model
	helpModel       help.Model
	keyMap          KeyMap
	
	// Panel visibility
	showDetailsPanel bool
	showLogPanel     bool
	showHelp         bool
	
	// Command mode state
	commandMode      bool
	commandInput     string
	
	// Search state
	searchMode       bool
	searchInput      textinput.Model
	searchQuery      string
	searchResults    []*taskmaster.Task
	
	// Filter state
	statusFilter     string // empty = all, or specific status like "pending", "in-progress", etc.
	
	// Confirmation mode state
	confirmingClearState bool
	
	// Styles
	styles        *Styles
	
	// Log data
	logLines      []string
	
	// Error state
	err           error
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, configManager *config.ConfigManager, taskService *taskmaster.Service, execService *executor.Service) Model {
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
	
	m := Model{
		config:        cfg,
		configManager: configManager,
		taskService:   taskService,
		execService:   execService,
		tasks:         tasks,
		taskIndex:     make(map[string]*taskmaster.Task),
		visibleTasks:  []*taskmaster.Task{},
		selectedIndex: 0,
		ready:         false,
		viewMode:      ViewModeTree,
		focusedPanel:  PanelTaskList,
		expandedNodes: make(map[string]bool),
		selectedIDs:   make(map[string]bool),
		taskListViewport: taskListVP,
		detailsViewport: detailsVP,
		logViewport:    logVP,
		helpModel:      help.New(),
		keyMap:         NewKeyMap(cfg),
		showDetailsPanel: true,
		showLogPanel:    false,
		showHelp:        false,
		commandMode:     false,
		commandInput:    "",
		styles:         NewStyles(),
		logLines:       []string{},
	}
	
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
		m.selectedTask = m.visibleTasks[0]
		m.selectedIndex = 0
	}
	
	return m
}

// buildTaskIndex creates a flat index of all tasks by ID for quick lookup
func (m *Model) buildTaskIndex() {
	m.taskIndex = make(map[string]*taskmaster.Task)
	var indexTasks func(tasks []taskmaster.Task)
	indexTasks = func(tasks []taskmaster.Task) {
		for i := range tasks {
			task := &tasks[i]
			m.taskIndex[task.ID] = task
			if len(task.Subtasks) > 0 {
				indexTasks(task.Subtasks)
			}
		}
	}
	indexTasks(m.tasks)
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
		m.selectedTask = m.visibleTasks[m.selectedIndex]
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
		m.selectedTask = m.visibleTasks[0]
		m.selectedIndex = 0
	} else {
		m.selectedTask = nil
	}
	
	// Update viewports
	m.updateTaskListViewport()
	m.updateDetailsViewport()
	
	return nil
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
			task := &tasks[i]
			result = append(result, task)
			if len(task.Subtasks) > 0 && m.expandedNodes[task.ID] {
				flatten(task.Subtasks)
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
		m.selectedTask = m.visibleTasks[m.selectedIndex]
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
		m.selectedTask = m.visibleTasks[m.selectedIndex]
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
			m.selectedTask = task
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

// renderTaskDetails renders the details of the selected task
func (m Model) renderTaskDetails() string {
	if m.selectedTask == nil {
		return m.styles.Info.Render("No task selected")
	}
	
	task := m.selectedTask
	var b strings.Builder
	
	// Title
	b.WriteString(m.styles.PanelTitle.Render(fmt.Sprintf("Task %s", task.ID)))
	b.WriteString("\n\n")
	
	// Title field
	b.WriteString(m.styles.Subtitle.Render("Title: "))
	b.WriteString(task.Title)
	b.WriteString("\n\n")
	
	// Status
	statusStyle := m.styles.GetStatusStyle(task.Status)
	b.WriteString(m.styles.Subtitle.Render("Status: "))
	b.WriteString(statusStyle.Render(task.Status))
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
		b.WriteString(fmt.Sprintf("%d", task.Complexity))
		b.WriteString("\n\n")
	}
	
	// Dependencies
	if len(task.Dependencies) > 0 {
		b.WriteString(m.styles.Subtitle.Render("Dependencies: "))
		b.WriteString(strings.Join(task.Dependencies, ", "))
		b.WriteString("\n\n")
	}
	
	// Description
	if task.Description != "" {
		b.WriteString(m.styles.Subtitle.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(task.Description)
		b.WriteString("\n\n")
	}
	
	// Details
	if task.Details != "" {
		b.WriteString(m.styles.Subtitle.Render("Details:"))
		b.WriteString("\n")
		b.WriteString(task.Details)
		b.WriteString("\n\n")
	}
	
	// Test Strategy
	if task.TestStrategy != "" {
		b.WriteString(m.styles.Subtitle.Render("Test Strategy:"))
		b.WriteString("\n")
		b.WriteString(task.TestStrategy)
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
			task := &tasks[i]
			result = append(result, task)
			if len(task.Subtasks) > 0 {
				flatten(task.Subtasks)
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
func (m Model) renderLog() string {
	if len(m.logLines) == 0 {
		return m.styles.Info.Render("No log output yet")
	}
	
	return strings.Join(m.logLines, "\n")
}

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
		"",                            // All tasks
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
			// Select first visible task
			m.selectedTask = m.visibleTasks[0]
			m.selectedIndex = 0
			m.updateDetailsViewport()
		}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		
		// Update viewport sizes based on layout
		m.updateViewportSizes()
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
		
		// Handle clear state confirmation mode
		if m.confirmingClearState {
			switch msg.String() {
			case "y", "Y":
				// Clear state
				if err := m.ClearUIState(); err != nil {
					m.addLogLine(fmt.Sprintf("Failed to clear state: %v", err))
				} else {
					m.addLogLine("TUI state cleared successfully")
				}
				m.confirmingClearState = false
			case "n", "N", "esc":
				// Cancel clear
				m.confirmingClearState = false
				m.addLogLine("State clear cancelled")
			default:
				// Ignore other keys in confirmation mode
				// Don't fall through to normal key handlers
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
				m.updateSearchResults()
				if len(m.searchResults) == 0 {
					m.addLogLine("No tasks found matching search query")
				} else {
					m.addLogLine(fmt.Sprintf("Found %d tasks matching '%s'", len(m.searchResults), m.searchQuery))
				}
			default:
				// Update search results as user types
				m.searchQuery = m.searchInput.Value()
				m.updateSearchResults()
			}
			return m, cmd
		}
		
		// Normal mode keyboard handling
		switch {
		case key.Matches(msg, m.keyMap.Quit):
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
			m.searchInput.SetValue(m.searchQuery) // Preserve previous query
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
			
		case key.Matches(msg, m.keyMap.ToggleDetails):
			// Toggle details panel
			m.showDetailsPanel = !m.showDetailsPanel
			m.updateViewportSizes()
			
		case key.Matches(msg, m.keyMap.ToggleLog):
			// Toggle log panel
			m.showLogPanel = !m.showLogPanel
			m.updateViewportSizes()
			
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
			
		case key.Matches(msg, m.keyMap.ClearState):
			// Enter confirmation mode for clearing state
			m.confirmingClearState = true
			m.addLogLine("Clear TUI state? (y/n)")
		}
	}
	
	// Handle viewport scrolling when focused
	if m.focusedPanel == PanelTaskList {
		var cmd tea.Cmd
		m.taskListViewport, cmd = m.taskListViewport.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.focusedPanel == PanelDetails {
		var cmd tea.Cmd
		m.detailsViewport, cmd = m.detailsViewport.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.focusedPanel == PanelLog {
		var cmd tea.Cmd
		m.logViewport, cmd = m.logViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return m.styles.Info.Render("Initializing Task Master TUI...")
	}

	// Help overlay takes priority over everything
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
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
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
		"  " + m.renderBinding(m.keyMap.Up) + " - Move up",
		"  " + m.renderBinding(m.keyMap.Down) + " - Move down", 
		"  " + m.renderBinding(m.keyMap.Left) + " - Collapse/Move left",
		"  " + m.renderBinding(m.keyMap.Right) + " - Expand/Move right",
		"  " + m.renderBinding(m.keyMap.PageUp) + " - Page up",
		"  " + m.renderBinding(m.keyMap.PageDown) + " - Page down",
	}
	sections = append(sections, navSection)
	sections = append(sections, strings.Join(navHelp, "\n"))
	sections = append(sections, "")
	
	// Task Operations section
	taskSection := m.styles.Subtitle.Render("Task Operations")
	taskHelp := []string{
		"  " + m.renderBinding(m.keyMap.ToggleExpand) + " - Toggle expand/collapse",
		"  " + m.renderBinding(m.keyMap.Select) + " - Select/deselect for bulk operations",
		"  " + m.renderBinding(m.keyMap.NextTask) + " - Get next available task",
		"  " + m.renderBinding(m.keyMap.Refresh) + " - Refresh tasks from disk",
		"  " + m.renderBinding(m.keyMap.JumpToID) + " - Jump to task by ID",
	}
	sections = append(sections, taskSection)
	sections = append(sections, strings.Join(taskHelp, "\n"))
	sections = append(sections, "")
	
	// Status Changes section
	statusSection := m.styles.Subtitle.Render("Status Changes")
	statusHelp := []string{
		"  " + m.renderBinding(m.keyMap.SetInProgress) + " - " + m.styles.InProgress.Render("► Set in-progress"),
		"  " + m.renderBinding(m.keyMap.SetDone) + " - " + m.styles.Done.Render("✓ Set done"),
		"  " + m.renderBinding(m.keyMap.SetBlocked) + " - " + m.styles.Blocked.Render("! Set blocked"),
		"  " + m.renderBinding(m.keyMap.SetCancelled) + " - " + m.styles.Cancelled.Render("✗ Set cancelled"),
		"  " + m.renderBinding(m.keyMap.SetDeferred) + " - " + m.styles.Deferred.Render("⏱ Set deferred"),
		"  " + m.renderBinding(m.keyMap.SetPending) + " - " + m.styles.Pending.Render("○ Set pending"),
	}
	sections = append(sections, statusSection)
	sections = append(sections, strings.Join(statusHelp, "\n"))
	sections = append(sections, "")
	
	// Panel & View section
	panelSection := m.styles.Subtitle.Render("Panels & Views")
	panelHelp := []string{
		"  " + m.renderBinding(m.keyMap.FocusTaskList) + " - Focus task list panel",
		"  " + m.renderBinding(m.keyMap.FocusDetails) + " - Focus details panel",
		"  " + m.renderBinding(m.keyMap.FocusLog) + " - Focus log panel",
		"  " + m.renderBinding(m.keyMap.CyclePanel) + " - Cycle through panels",
		"  " + m.renderBinding(m.keyMap.ToggleDetails) + " - Toggle details panel",
		"  " + m.renderBinding(m.keyMap.ToggleLog) + " - Toggle log panel",
		"  " + m.renderBinding(m.keyMap.ViewTree) + " - Switch to tree view",
		"  " + m.renderBinding(m.keyMap.ViewList) + " - Switch to list view",
		"  " + m.renderBinding(m.keyMap.CycleView) + " - Cycle view modes",
	}
	sections = append(sections, panelSection)
	sections = append(sections, strings.Join(panelHelp, "\n"))
	sections = append(sections, "")
	
	// General section
	generalSection := m.styles.Subtitle.Render("General")
	generalHelp := []string{
		"  " + m.renderBinding(m.keyMap.Help) + " - Toggle this help",
		"  " + m.renderBinding(m.keyMap.ClearState) + " - Clear TUI state (reset UI)",
		"  " + m.renderBinding(m.keyMap.Back) + " - Back/Cancel/Close",
		"  " + m.renderBinding(m.keyMap.Quit) + " - Quit application",
		"  " + m.renderBinding(m.keyMap.Cancel) + " - Cancel command or quit",
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
		MaxHeight(m.height - 4).
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

