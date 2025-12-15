package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

// createTestModel creates a model with test data
func createTestModel() Model {
	cfg := &config.Config{
		TaskMasterPath: "/tmp/test",
	}

	// Create test tasks
	tasks := []taskmaster.Task{
		{
			ID:     "1",
			Title:  "Task 1",
			Status: "pending",
			Subtasks: []taskmaster.Task{
				{
					ID:     "1.1",
					Title:  "Subtask 1.1",
					Status: "pending",
				},
				{
					ID:     "1.2",
					Title:  "Subtask 1.2",
					Status: "done",
				},
			},
		},
		{
			ID:     "2",
			Title:  "Task 2",
			Status: "in-progress",
		},
	}

	// Create a minimal model without full services
	m := Model{
		config:           cfg,
		tasks:            tasks,
		taskIndex:        make(map[string]*taskmaster.Task),
		visibleTasks:     []*taskmaster.Task{},
		selectedIndex:    0,
		viewMode:         ViewModeTree,
		focusedPanel:     PanelTaskList,
		expandedNodes:    make(map[string]bool),
		selectedIDs:      make(map[string]bool),
		keyMap:           DefaultKeyMap(),
		showDetailsPanel: true,
		showLogPanel:     false,
		showHelp:         false,
		commandMode:      false,
		commandInput:     "",
		styles:           NewStyles(),
		logLines:         []string{},
	}

	m.buildTaskIndex()
	m.rebuildVisibleTasks()

	if len(m.visibleTasks) > 0 {
		m.selectedTask = m.visibleTasks[0]
		m.selectedIndex = 0
	}

	return m
}

// TestNavigationDown tests moving down in the task list
func TestNavigationDown(t *testing.T) {
	m := createTestModel()

	// Start at first task
	if m.selectedIndex != 0 {
		t.Errorf("Expected initial selectedIndex to be 0, got %d", m.selectedIndex)
	}

	// Move down
	m.selectNext()

	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after moving down, got %d", m.selectedIndex)
	}

	// Try to move past the end
	for i := 0; i < 10; i++ {
		m.selectNext()
	}

	if m.selectedIndex >= len(m.visibleTasks) {
		t.Errorf("selectedIndex should not exceed visible tasks length")
	}
}

// TestNavigationUp tests moving up in the task list
func TestNavigationUp(t *testing.T) {
	m := createTestModel()

	// Move to second task
	m.selectNext()
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1, got %d", m.selectedIndex)
	}

	// Move up
	m.selectPrevious()

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0 after moving up, got %d", m.selectedIndex)
	}

	// Try to move before the start
	m.selectPrevious()

	if m.selectedIndex < 0 {
		t.Errorf("selectedIndex should not be negative, got %d", m.selectedIndex)
	}
}

// TestExpandCollapse tests expand and collapse functionality
func TestExpandCollapse(t *testing.T) {
	m := createTestModel()

	// Select task with subtasks (task 1)
	m.selectedIndex = 0
	m.selectedTask = m.visibleTasks[0]

	if m.selectedTask.ID != "1" {
		t.Errorf("Expected to select task 1, got %s", m.selectedTask.ID)
	}

	// Initially should be collapsed
	initialVisibleCount := len(m.visibleTasks)

	// Expand task 1
	m.expandSelected()

	if len(m.visibleTasks) <= initialVisibleCount {
		t.Errorf("Expected more visible tasks after expanding, got %d", len(m.visibleTasks))
	}

	if !m.expandedNodes["1"] {
		t.Error("Expected task 1 to be expanded")
	}

	// Collapse task 1 (should collapse since it's expanded)
	m.collapseSelected()

	if len(m.visibleTasks) != initialVisibleCount {
		t.Errorf("Expected visible tasks to return to %d after collapsing, got %d",
			initialVisibleCount, len(m.visibleTasks))
	}

	if m.expandedNodes["1"] {
		t.Error("Expected task 1 to be collapsed")
	}
}

// TestCollapseNavigatesToParent tests that collapsing an already-collapsed task navigates to parent
func TestCollapseNavigatesToParent(t *testing.T) {
	m := createTestModel()

	// Expand task 1 to make subtasks visible
	m.selectedIndex = 0
	m.selectedTask = m.visibleTasks[0]
	m.expandSelected()

	// Select subtask 1.1 (first subtask)
	m.selectNext() // Move to 1.1
	if m.selectedTask == nil || m.selectedTask.ID != "1.1" {
		t.Errorf("Expected to select task 1.1, got %v", m.selectedTask)
	}

	// Collapse on 1.1 (it has no subtasks, so should navigate to parent "1")
	m.collapseSelected()

	if m.selectedTask == nil || m.selectedTask.ID != "1" {
		t.Errorf("Expected to navigate to parent task 1, got %v", m.selectedTask)
	}
}

// TestGetParentID tests the parent ID extraction
func TestGetParentID(t *testing.T) {
	m := createTestModel()

	tests := []struct {
		taskID   string
		expected string
	}{
		{"1", ""},        // Root level, no parent
		{"2", ""},        // Root level, no parent
		{"1.1", "1"},     // Parent is 1
		{"1.2", "1"},     // Parent is 1
		{"1.1.1", "1.1"}, // Parent is 1.1
		{"2.3.4", "2.3"}, // Parent is 2.3
	}

	for _, test := range tests {
		result := m.getParentID(test.taskID)
		if result != test.expected {
			t.Errorf("getParentID(%s) = %s; expected %s", test.taskID, result, test.expected)
		}
	}
}

// TestToggleExpand tests toggling expand state
func TestToggleExpand(t *testing.T) {
	m := createTestModel()

	// Select task with subtasks (task 1)
	m.selectedIndex = 0
	m.selectedTask = m.visibleTasks[0]

	// Toggle expand
	m.toggleExpanded()

	if !m.expandedNodes["1"] {
		t.Error("Expected task 1 to be expanded after toggle")
	}

	// Toggle collapse
	m.toggleExpanded()

	if m.expandedNodes["1"] {
		t.Error("Expected task 1 to be collapsed after second toggle")
	}
}

// TestTaskSelection tests selecting and deselecting tasks
func TestTaskSelection(t *testing.T) {
	m := createTestModel()

	// Select task 1
	m.selectedIndex = 0
	m.selectedTask = m.visibleTasks[0]

	// Toggle selection
	m.toggleSelection()

	if !m.isTaskSelected("1") {
		t.Error("Expected task 1 to be selected")
	}

	// Toggle deselection
	m.toggleSelection()

	if m.isTaskSelected("1") {
		t.Error("Expected task 1 to be deselected")
	}
}

// TestMultipleSelection tests selecting multiple tasks
func TestMultipleSelection(t *testing.T) {
	m := createTestModel()

	// Select task 1
	m.selectedIndex = 0
	m.selectedTask = m.visibleTasks[0]
	m.toggleSelection()

	// Move to task 2 and select it
	m.selectNext()
	m.toggleSelection()

	selectedTasks := m.getSelectedTasks()

	if len(selectedTasks) != 2 {
		t.Errorf("Expected 2 selected tasks, got %d", len(selectedTasks))
	}

	// Clear selection
	m.clearSelection()

	if len(m.getSelectedTasks()) != 0 {
		t.Errorf("Expected 0 selected tasks after clear, got %d", len(m.getSelectedTasks()))
	}
}

// TestQuickJump tests jumping to a task by ID
func TestQuickJump(t *testing.T) {
	m := createTestModel()

	// Expand task 1 to make subtasks visible
	m.expandSelected()

	// Jump to subtask 1.2
	success := m.selectTaskByID("1.2")

	if !success {
		t.Error("Expected to successfully jump to task 1.2")
	}

	if m.selectedTask == nil || m.selectedTask.ID != "1.2" {
		t.Errorf("Expected selected task to be 1.2, got %v", m.selectedTask)
	}

	// Verify ancestor is expanded
	if !m.expandedNodes["1"] {
		t.Error("Expected parent task 1 to be expanded after jumping to 1.2")
	}
}

// TestQuickJumpInvalid tests jumping to a non-existent task
func TestQuickJumpInvalid(t *testing.T) {
	m := createTestModel()

	// Try to jump to non-existent task
	success := m.selectTaskByID("999")

	if success {
		t.Error("Expected jump to non-existent task to fail")
	}
}

// TestKeyboardMessage tests handling keyboard messages
func TestKeyboardMessage(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Test down key
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)

	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after 'j' key, got %d", m.selectedIndex)
	}

	// Test up key
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0 after 'k' key, got %d", m.selectedIndex)
	}
}

// TestExpandAll tests expanding all tasks
func TestExpandAll(t *testing.T) {
	m := createTestModel()

	// Expand all
	m.expandAll()

	// Check that task 1 is expanded
	if !m.expandedNodes["1"] {
		t.Error("Expected task 1 to be expanded")
	}

	// All tasks with subtasks should be visible
	visibleCount := len(m.visibleTasks)
	expectedCount := 4 // 1, 1.1, 1.2, 2

	if visibleCount != expectedCount {
		t.Errorf("Expected %d visible tasks, got %d", expectedCount, visibleCount)
	}
}

// TestCollapseAll tests collapsing all tasks
func TestCollapseAll(t *testing.T) {
	m := createTestModel()

	// First expand all
	m.expandAll()

	// Then collapse all
	m.collapseAll()

	// No tasks should be expanded
	if len(m.expandedNodes) != 0 {
		t.Errorf("Expected no expanded nodes, got %d", len(m.expandedNodes))
	}

	// Only top-level tasks should be visible
	expectedCount := 2 // 1, 2
	if len(m.visibleTasks) != expectedCount {
		t.Errorf("Expected %d visible tasks, got %d", expectedCount, len(m.visibleTasks))
	}
}

// TestClearUIState tests the state clearing functionality
func TestClearUIState(t *testing.T) {
	m := createTestModel()

	// Set up some custom state
	m.expandedNodes["1"] = true
	m.viewMode = ViewModeList
	m.focusedPanel = PanelDetails
	m.showDetailsPanel = false
	m.showLogPanel = true
	m.selectedIndex = 1

	// Rebuild to reflect expanded state
	m.rebuildVisibleTasks()

	// Clear the state (without actual file operations since we don't have a real state path)
	m.config.StatePath = ""
	err := m.ClearUIState()

	if err != nil {
		t.Errorf("ClearUIState() returned error: %v", err)
	}

	// Verify all state was reset
	if len(m.expandedNodes) != 0 {
		t.Errorf("Expected expandedNodes to be empty, got %d items", len(m.expandedNodes))
	}

	if m.viewMode != ViewModeTree {
		t.Errorf("Expected viewMode to be Tree, got %v", m.viewMode)
	}

	if m.focusedPanel != PanelTaskList {
		t.Errorf("Expected focusedPanel to be TaskList, got %v", m.focusedPanel)
	}

	if !m.showDetailsPanel {
		t.Error("Expected showDetailsPanel to be true")
	}

	if m.showLogPanel {
		t.Error("Expected showLogPanel to be false")
	}

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", m.selectedIndex)
	}

	if m.selectedTask == nil || m.selectedTask.ID != "1" {
		t.Errorf("Expected selectedTask to be task 1, got %v", m.selectedTask)
	}

	if m.confirmingClearState {
		t.Error("Expected confirmingClearState to be false")
	}
}
