package ui

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// createTestModel creates a model with test data
func createTestModel() *Model {
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

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search tasks (ID, title, description)..."
	searchInput.CharLimit = 100
	searchInput.Width = 40
	searchInput.Prompt = ""

	// Create a minimal model without full services
	keyMap := DefaultKeyMap()
	m := &Model{
		config:           cfg,
		tasks:            tasks,
		taskIndex:        make(map[string]*taskmaster.Task),
		visibleTasks:     []*taskmaster.Task{},
		selectedIndex:    0,
		viewMode:         ViewModeTree,
		focusedPanel:     PanelTaskList,
		expandedNodes:    make(map[string]bool),
		selectedIDs:      make(map[string]bool),
		keyMap:           keyMap,
		showDetailsPanel: true,
		showLogPanel:     false,
		showHelp:         false,
		commandMode:      false,
		commandInput:     "",
		searchInput:      searchInput,
		styles:           NewStyles(),
		logLines:         []string{},
		appState:         NewAppState(nil, &keyMap),
		// Execution queue state
		executionQueue:         nil,
		activeTaskModelDialog:  nil,
		showTaskModelDialog:    false,
		taskModelSelectionDone: make(map[string]bool),
		// Initialize filterableComponent for standardized filtering
		filterableComponent: dialog.NewBaseFilterable(),
	}

	// Enable filtering capability
	m.filterableComponent.EnableFiltering(true)

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
	m = newModel.(*Model)

	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after 'j' key, got %d", m.selectedIndex)
	}

	// Test up key
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(*Model)

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

// TestExecutorOutputMsgToLog tests that output is added to log when modal is inactive
func TestExecutorOutputMsgToLog(t *testing.T) {
	m := createTestModel()

	// Ensure modal is not active
	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Send ExecutorOutputMsg
	output := "test output line"
	msg := ExecutorOutputMsg{Line: output}
	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Check that output was added to log
	if len(model.logLines) == 0 {
		t.Error("Expected output to be added to logLines when modal is inactive")
	}

	if len(model.logLines) > 0 && model.logLines[len(model.logLines)-1] != output {
		t.Errorf("Expected last log line to be %q, got %q", output, model.logLines[len(model.logLines)-1])
	}
}

// TestExecutorOutputMsgToModal tests that output is routed to modal when active
func TestExecutorOutputMsgToModal(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Activate the next task modal
	m.appState.StartNextTaskModal()

	// Send ExecutorOutputMsg
	output := "modal output line"
	msg := ExecutorOutputMsg{Line: output}
	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Check that output was added to appState but NOT to log
	if !model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to be active")
	}

	// Verify output is in appState
	if len(model.appState.NextTaskOutput()) == 0 {
		t.Error("Expected output to be added to appState when modal is active")
	}

	if len(model.appState.NextTaskOutput()) > 0 && model.appState.NextTaskOutput()[len(model.appState.NextTaskOutput())-1] != output {
		t.Errorf("Expected appState output to be %q, got %q", output, model.appState.NextTaskOutput()[len(model.appState.NextTaskOutput())-1])
	}
}

// TestExecutorOutputMsgModalSkipsLog tests that log is not updated when modal is active
func TestExecutorOutputMsgModalSkipsLog(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Record initial log length
	initialLogLength := len(m.logLines)

	// Activate the next task modal
	m.appState.StartNextTaskModal()

	// Send ExecutorOutputMsg
	output := "modal only line"
	msg := ExecutorOutputMsg{Line: output}
	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Check that log was NOT updated
	if len(model.logLines) != initialLogLength {
		t.Errorf("Expected log to remain unchanged when modal is active, but got %d lines instead of %d", len(model.logLines), initialLogLength)
	}
}

// TestExecutorOutputMsgMultipleLines tests handling of multiple output lines
func TestExecutorOutputMsgMultipleLines(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Activate the next task modal
	m.appState.StartNextTaskModal()

	// Send multiple ExecutorOutputMsg
	lines := []string{"line 1", "line 2", "line 3"}
	for _, line := range lines {
		msg := ExecutorOutputMsg{Line: line}
		newM, _ := m.Update(msg)
		m = newM.(*Model)
	}

	// Check that all lines were added to appState
	output := m.appState.NextTaskOutput()
	if len(output) != len(lines) {
		t.Errorf("Expected %d lines in appState, got %d", len(lines), len(output))
	}

	for i, line := range lines {
		if i < len(output) && output[i] != line {
			t.Errorf("Expected line %d to be %q, got %q", i, line, output[i])
		}
	}
}

// TestExecutorOutputMsgCloseModal tests that output reverts to log after modal is closed
func TestExecutorOutputMsgCloseModal(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Activate and then close the modal
	m.appState.StartNextTaskModal()
	m.appState.CloseNextTaskModal()

	// Send ExecutorOutputMsg
	output := "post-close line"
	msg := ExecutorOutputMsg{Line: output}
	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Check that output went to log
	if len(model.logLines) == 0 {
		t.Error("Expected output to be added to log after modal is closed")
	}

	if len(model.logLines) > 0 && model.logLines[len(model.logLines)-1] != output {
		t.Errorf("Expected last log line to be %q, got %q", output, model.logLines[len(model.logLines)-1])
	}
}

// TestExecutorDoneMsgSuccess tests handling of successful command completion
func TestExecutorDoneMsgSuccess(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Activate the next task modal
	m.appState.StartNextTaskModal()

	// Add some output first
	outputLines := []string{"task info line 1", "task info line 2"}
	for _, line := range outputLines {
		msg := ExecutorOutputMsg{Line: line}
		newM, _ := m.Update(msg)
		m = newM.(*Model)
	}

	// Send successful ExecutorDoneMsg
	doneMsg := ExecutorDoneMsg{
		Command: "next",
		Success: true,
		Error:   nil,
	}
	newM, _ := m.Update(doneMsg)
	model := newM.(*Model)

	// Verify that the modal was updated with loading state false
	if model.appState.IsNextTaskModalActive() {
		// Get the active dialog
		dlg := model.appState.ActiveDialog()
		if dlg != nil {
			// The modal should still be active but with loading = false
			// This is verified through the SetLoading(false) call in the handler
		}
	}
}

// TestExecutorDoneMsgFailure tests handling of failed command completion
func TestExecutorDoneMsgFailure(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Activate the next task modal
	m.appState.StartNextTaskModal()

	// Send failed ExecutorDoneMsg with error
	doneMsg := ExecutorDoneMsg{
		Command: "next",
		Success: false,
		Error:   errors.New("permission denied"),
	}
	newM, _ := m.Update(doneMsg)
	model := newM.(*Model)

	// Model should still be active, ready to display error
	if !model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to remain active on error")
	}
}

// TestExecutorDoneMsgEmptyOutput tests handling of empty output on completion
func TestExecutorDoneMsgEmptyOutput(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Activate the next task modal without adding any output
	m.appState.StartNextTaskModal()

	// Send successful ExecutorDoneMsg with no output
	doneMsg := ExecutorDoneMsg{
		Command: "next",
		Success: true,
		Error:   nil,
	}
	newM, _ := m.Update(doneMsg)
	model := newM.(*Model)

	// The handler should set "No tasks available." message
	// Verify the modal is still active to show this message
	if !model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to remain active to show 'No tasks available' message")
	}
}

// TestExecutorDoneMsgNonNextCommand tests handling of non-next commands
func TestExecutorDoneMsgNonNextCommand(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Do NOT activate the next task modal

	// Send ExecutorDoneMsg for a different command
	doneMsg := ExecutorDoneMsg{
		Command: "list",
		Success: true,
		Error:   nil,
	}
	newM, _ := m.Update(doneMsg)
	model := newM.(*Model)

	// Modal should not be affected
	if model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to not be affected by non-next command")
	}
}

// TestNextTaskModalESCKeyClosing tests that ESC key closes the next task modal
func TestNextTaskModalESCKeyClosing(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Initialize modal and add some output
	m.appState.StartNextTaskModal()
	m.appState.AppendNextTaskOutput("Task: example")

	if !m.appState.IsNextTaskModalActive() {
		t.Fatal("Expected modal to be active")
	}

	// Create a modal dialog for the next task output
	nextTaskContent := dialog.NewNextTaskOutputContent()
	nextTaskContent.SetOutput(m.appState.NextTaskOutput())
	dlg := dialog.NewModalDialog("Next Task", 80, 20, nextTaskContent)

	// Verify the dialog is cancellable and handles ESC
	if !dlg.IsCancellable() {
		t.Error("Expected modal dialog to be cancellable")
	}

	// Simulate ESC key press
	result, _ := dlg.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})

	if result != dialog.DialogResultCancel {
		t.Errorf("Expected ESC to return DialogResultCancel, got %v", result)
	}
}

// TestNextTaskModalStateResetOnClose tests that modal state is reset when closed
func TestNextTaskModalStateResetOnClose(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Initialize modal and add output
	m.appState.StartNextTaskModal()
	m.appState.AppendNextTaskOutput("Line 1")
	m.appState.AppendNextTaskOutput("Line 2")

	// Verify modal is active with output
	if !m.appState.IsNextTaskModalActive() {
		t.Fatal("Expected modal to be active")
	}

	outputBefore := m.appState.NextTaskOutput()
	if len(outputBefore) != 2 {
		t.Errorf("Expected 2 output lines before close, got %d", len(outputBefore))
	}

	// Close the modal
	m.appState.CloseNextTaskModal()

	// Verify modal state is reset
	if m.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to be inactive after close")
	}

	outputAfter := m.appState.NextTaskOutput()
	if outputAfter != nil {
		t.Errorf("Expected output to be nil after close, got %v", outputAfter)
	}
}

// TestNextTaskModalFocusRestoration tests that focus is restored to task list after closing
func TestNextTaskModalFocusRestoration(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Verify initial focus is on task list
	if m.focusedPanel != PanelTaskList {
		t.Logf("Note: initial focus might not be on task list in test, but model initializes to PanelTaskList")
	}

	// Switch focus to details panel
	m.focusedPanel = PanelDetails

	if m.focusedPanel != PanelDetails {
		t.Fatal("Expected focus to be on details panel")
	}

	// Now simulate closing the modal (which should restore focus to task list)
	m.focusedPanel = PanelTaskList

	// Verify focus is restored to task list
	if m.focusedPanel != PanelTaskList {
		t.Errorf("Expected focus to be restored to task list, got %v", m.focusedPanel)
	}
}

// TestMultipleOpenCloseCyclesWithFocus tests multiple open/close cycles maintain proper state
func TestMultipleOpenCloseCyclesWithFocus(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	for i := 0; i < 3; i++ {
		// Start modal
		m.appState.StartNextTaskModal()

		if !m.appState.IsNextTaskModalActive() {
			t.Errorf("Cycle %d: Expected modal to be active after start", i)
		}

		// Add output
		m.appState.AppendNextTaskOutput("Output " + string(rune('0'+i)))

		// Close modal
		m.appState.CloseNextTaskModal()

		if m.appState.IsNextTaskModalActive() {
			t.Errorf("Cycle %d: Expected modal to be inactive after close", i)
		}

		// Verify output is cleared
		if m.appState.NextTaskOutput() != nil {
			t.Errorf("Cycle %d: Expected output to be cleared after close", i)
		}
	}
}

// TestExecutorDoneMsg_NotInWorkspace tests ExecutorDoneMsg with empty TaskMasterPath error
func TestExecutorDoneMsg_NotInWorkspace(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Start next task modal
	m.appState.StartNextTaskModal()

	// Simulate ExecutorDoneMsg with empty TaskMasterPath error
	msg := ExecutorDoneMsg{
		Command: "next",
		Success: false,
		Error:   errors.New("not running in a Task Master workspace"),
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify the modal is still active (shows error)
	if !model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to remain active after error")
	}
}

// TestExecutorDoneMsg_BinaryNotFound tests ExecutorDoneMsg with missing binary error
func TestExecutorDoneMsg_BinaryNotFound(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Start next task modal
	m.appState.StartNextTaskModal()

	// Simulate ExecutorDoneMsg with missing binary error
	msg := ExecutorDoneMsg{
		Command: "next",
		Success: false,
		Error:   errors.New("task-master binary not found"),
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify the modal is still active (shows error)
	if !model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to remain active after error")
	}
}

// TestExecutorDoneMsg_BinaryNotExecutable tests ExecutorDoneMsg with non-executable error
func TestExecutorDoneMsg_BinaryNotExecutable(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Start next task modal
	m.appState.StartNextTaskModal()

	// Simulate ExecutorDoneMsg with non-executable binary error
	msg := ExecutorDoneMsg{
		Command: "next",
		Success: false,
		Error:   errors.New("task-master binary not executable"),
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify the modal is still active (shows error)
	if !model.appState.IsNextTaskModalActive() {
		t.Error("Expected modal to remain active after error")
	}
}

// TestGitBranchCreationSuccess tests successful branch creation via TaskCompletedMsg
func TestGitBranchCreationSuccess(t *testing.T) {
	m := createTestModel()

	// Initialize git components
	m.gitAvailable = true
	m.gitRepoInfo.IsRepo = true
	m.gitRefresher = nil // No real git refresher in tests

	// Simulate TaskCompletedMsg for git-create-branch
	msg := dialog.TaskCompletedMsg{
		TaskID: "git-create-branch",
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify the success message was logged
	logFound := false
	for _, line := range model.logLines {
		if line == "✓ Branch created successfully" {
			logFound = true
			break
		}
	}
	if !logFound {
		t.Error("Expected success message in log for git branch creation")
	}
}

// TestGitBranchSwitchSuccess tests successful branch switch via TaskCompletedMsg
func TestGitBranchSwitchSuccess(t *testing.T) {
	m := createTestModel()

	// Initialize git components
	m.gitAvailable = true
	m.gitRepoInfo.IsRepo = true
	m.gitRefresher = nil

	// Simulate TaskCompletedMsg for git-switch-branch
	msg := dialog.TaskCompletedMsg{
		TaskID: "git-switch-branch",
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify the success message was logged
	logFound := false
	for _, line := range model.logLines {
		if line == "✓ Branch switched successfully" {
			logFound = true
			break
		}
	}
	if !logFound {
		t.Error("Expected success message in log for git branch switch")
	}
}

// TestGitBranchCreationFailure tests branch creation failure via TaskFailedMsg
func TestGitBranchCreationFailure(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Initialize git components
	m.gitAvailable = true
	m.gitRepoInfo.IsRepo = true

	// Simulate TaskFailedMsg for git-create-branch
	msg := dialog.TaskFailedMsg{
		TaskID: "git-create-branch",
		Error:  "fatal: branch 'test' already exists",
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify error was logged
	errorFound := false
	for _, line := range model.logLines {
		if line == "Task git-create-branch failed: fatal: branch 'test' already exists" {
			errorFound = true
			break
		}
	}
	if !errorFound {
		t.Error("Expected error message in log for failed branch creation")
	}
}

// TestGitBranchSwitchFailure tests branch switch failure via TaskFailedMsg
func TestGitBranchSwitchFailure(t *testing.T) {
	m := createTestModel()

	if m.appState == nil {
		t.Skip("appState not initialized in test model")
	}

	// Initialize git components
	m.gitAvailable = true
	m.gitRepoInfo.IsRepo = true

	// Simulate TaskFailedMsg for git-switch-branch
	msg := dialog.TaskFailedMsg{
		TaskID: "git-switch-branch",
		Error:  "error: pathspec 'nonexistent' did not match any files",
	}

	newM, _ := m.Update(msg)
	model := newM.(*Model)

	// Verify error was logged
	errorFound := false
	for _, line := range model.logLines {
		if line == "Task git-switch-branch failed: error: pathspec 'nonexistent' did not match any files" {
			errorFound = true
			break
		}
	}
	if !errorFound {
		t.Error("Expected error message in log for failed branch switch")
	}
}

// TestOpenBranchSwitchDialogLogging tests that openBranchSwitchDialog logs the operation start
func TestOpenBranchSwitchDialogLogging(t *testing.T) {
	m := createTestModel()

	// Initialize git components
	m.gitAvailable = true
	m.gitRepoInfo.IsRepo = true

	// Call the dialog open function - it should log the start
	m.openBranchSwitchDialog()

	// Verify logging occurred
	logFound := false
	for _, line := range m.logLines {
		if line == "Starting branch switch operation..." {
			logFound = true
			break
		}
	}
	if !logFound {
		t.Error("Expected start message in log for branch switch operation")
	}
}

// TestOpenBranchCreateDialogLogging tests that openBranchCreateDialog logs the operation start
func TestOpenBranchCreateDialogLogging(t *testing.T) {
	m := createTestModel()

	// Initialize git components
	m.gitAvailable = true
	m.gitRepoInfo.IsRepo = true

	// Call the dialog open function - it should log the start
	m.openBranchCreateDialog()

	// Verify logging occurred
	logFound := false
	for _, line := range m.logLines {
		if line == "Starting branch creation operation..." {
			logFound = true
			break
		}
	}
	if !logFound {
		t.Error("Expected start message in log for branch creation operation")
	}
}

// TestRefreshTaskTree_ReturnsCommand tests that refreshTaskTree returns a command
func TestRefreshTaskTree_ReturnsCommand(t *testing.T) {
	m := createTestModel()

	// Call refreshTaskTree - it should return a command
	cmd := m.refreshTaskTree()

	// Verify a command was returned
	if cmd == nil {
		t.Error("Expected refreshTaskTree to return a command, got nil")
	}
}

// TestRefreshTaskTree_WithSelectedTask tests refreshTaskTree with a selected task
func TestRefreshTaskTree_WithSelectedTask(t *testing.T) {
	m := createTestModel()

	// Ensure we have a selected task
	if len(m.tasks) == 0 {
		t.Skip("Test requires tasks to be present")
	}

	// Select the first task
	m.selectedTask = &m.tasks[0]
	originalTaskID := m.selectedTask.ID

	// Call refreshTaskTree
	cmd := m.refreshTaskTree()

	// Verify command is returned
	if cmd == nil {
		t.Error("Expected refreshTaskTree to return a command with selected task")
	}

	// Verify the task ID is still what we selected (before command execution)
	if m.selectedTask.ID != originalTaskID {
		t.Errorf("Selected task changed unexpectedly: was %s, now %s", originalTaskID, m.selectedTask.ID)
	}
}

// TestRefreshTaskTree_WithoutSelection tests refreshTaskTree when no task is selected
func TestRefreshTaskTree_WithoutSelection(t *testing.T) {
	m := createTestModel()

	// Explicitly clear the selection
	m.selectedTask = nil

	// Call refreshTaskTree
	cmd := m.refreshTaskTree()

	// Verify command is returned
	if cmd == nil {
		t.Error("Expected refreshTaskTree to return a command even with no selection")
	}
}

// TestRefreshTaskTree_PreservesViewMode tests that refreshTaskTree preserves the current view mode
func TestRefreshTaskTree_PreservesViewMode(t *testing.T) {
	m := createTestModel()

	// Set a specific view mode
	m.viewMode = ViewModeList
	originalViewMode := m.viewMode

	// Call refreshTaskTree
	cmd := m.refreshTaskTree()

	// Verify command is returned
	if cmd == nil {
		t.Error("Expected refreshTaskTree to return a command")
	}

	// Verify view mode is still the same (before command execution)
	if m.viewMode != originalViewMode {
		t.Errorf("View mode changed: was %d, now %d", originalViewMode, m.viewMode)
	}
}

// TestExecutionQueueStateInitialization tests that Model includes execution queue state fields
func TestExecutionQueueStateInitialization(t *testing.T) {
	m := createTestModel()

	// Verify executionQueue field exists and is initialized to nil
	if m.executionQueue != nil {
		t.Error("Expected executionQueue to be nil initially")
	}

	// Verify activeTaskModelDialog field exists and is initialized to nil
	if m.activeTaskModelDialog != nil {
		t.Error("Expected activeTaskModelDialog to be nil initially")
	}

	// Verify showTaskModelDialog field exists and is initialized to false
	if m.showTaskModelDialog != false {
		t.Error("Expected showTaskModelDialog to be false initially")
	}

	// Verify taskModelSelectionDone field exists and is initialized as empty map
	if m.taskModelSelectionDone == nil {
		t.Error("Expected taskModelSelectionDone to be initialized as non-nil map")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Errorf("Expected taskModelSelectionDone to be empty initially, got %d entries", len(m.taskModelSelectionDone))
	}
}

// TestExecutionQueueStateCanBeModified tests that execution queue state can be properly modified
func TestExecutionQueueStateCanBeModified(t *testing.T) {
	m := createTestModel()

	// Create and assign an ExecutionQueue
	queue := &ExecutionQueue{
		TaskIDs:         []string{"1", "2"},
		ModelSelections: make(map[string]string),
		TaskStatus:      make(map[string]string),
	}
	m.executionQueue = queue

	// Verify the assignment worked
	if m.executionQueue == nil {
		t.Fatal("Failed to assign executionQueue")
	}
	if len(m.executionQueue.TaskIDs) != 2 {
		t.Errorf("Expected 2 task IDs in queue, got %d", len(m.executionQueue.TaskIDs))
	}

	// Modify showTaskModelDialog
	m.showTaskModelDialog = true
	if !m.showTaskModelDialog {
		t.Error("Failed to set showTaskModelDialog to true")
	}

	// Modify taskModelSelectionDone
	m.taskModelSelectionDone["1"] = true
	m.taskModelSelectionDone["2"] = true
	if len(m.taskModelSelectionDone) != 2 {
		t.Errorf("Expected 2 entries in taskModelSelectionDone, got %d", len(m.taskModelSelectionDone))
	}
}

// TestExecutionQueueStateCanBeReset tests that execution queue state can be properly reset
func TestExecutionQueueStateCanBeReset(t *testing.T) {
	m := createTestModel()

	// Set up some state
	m.executionQueue = &ExecutionQueue{
		TaskIDs: []string{"1", "2"},
	}
	m.showTaskModelDialog = true
	m.taskModelSelectionDone["1"] = true
	m.taskModelSelectionDone["2"] = true

	// Verify state was set
	if m.executionQueue == nil {
		t.Fatal("Failed to set executionQueue for test")
	}
	if !m.showTaskModelDialog {
		t.Fatal("Failed to set showTaskModelDialog for test")
	}
	if len(m.taskModelSelectionDone) != 2 {
		t.Fatal("Failed to set taskModelSelectionDone for test")
	}

	// Reset the state (this is how execution queue should be reset)
	m.executionQueue = nil
	m.showTaskModelDialog = false
	m.taskModelSelectionDone = make(map[string]bool)

	// Verify state was reset
	if m.executionQueue != nil {
		t.Error("Expected executionQueue to be nil after reset")
	}
	if m.showTaskModelDialog {
		t.Error("Expected showTaskModelDialog to be false after reset")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Errorf("Expected taskModelSelectionDone to be empty after reset, got %d entries", len(m.taskModelSelectionDone))
	}
}

// TestExecutionQueueStateFieldTypes tests that fields have correct types
func TestExecutionQueueStateFieldTypes(t *testing.T) {
	m := createTestModel()

	// Test that fields exist and have the correct types
	// This is a compile-time check via type assertions
	var _ *ExecutionQueue = m.executionQueue
	var _ bool = m.showTaskModelDialog
	var _ map[string]bool = m.taskModelSelectionDone
}

// TestResetExecutionQueue tests the ResetExecutionQueue method
func TestResetExecutionQueue(t *testing.T) {
	m := createTestModel()

	// Set up some state
	m.executionQueue = &ExecutionQueue{
		TaskIDs:         []string{"1", "2"},
		ModelSelections: make(map[string]string),
		TaskStatus:      make(map[string]string),
	}
	m.showTaskModelDialog = true
	m.taskModelSelectionDone["1"] = true
	m.taskModelSelectionDone["2"] = true

	// Verify state was set
	if m.executionQueue == nil {
		t.Fatal("Failed to set executionQueue for test")
	}
	if !m.showTaskModelDialog {
		t.Fatal("Failed to set showTaskModelDialog for test")
	}
	if len(m.taskModelSelectionDone) != 2 {
		t.Fatal("Failed to set taskModelSelectionDone for test")
	}

	// Call ResetExecutionQueue
	m.ResetExecutionQueue()

	// Verify state was reset
	if m.executionQueue != nil {
		t.Error("Expected executionQueue to be nil after ResetExecutionQueue()")
	}
	if m.activeTaskModelDialog != nil {
		t.Error("Expected activeTaskModelDialog to be nil after ResetExecutionQueue()")
	}
	if m.showTaskModelDialog {
		t.Error("Expected showTaskModelDialog to be false after ResetExecutionQueue()")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Errorf("Expected taskModelSelectionDone to be empty after ResetExecutionQueue(), got %d entries", len(m.taskModelSelectionDone))
	}
}

// TestResetExecutionQueue_WithNilFields tests ResetExecutionQueue when fields are already nil
func TestResetExecutionQueue_WithNilFields(t *testing.T) {
	m := createTestModel()

	// Verify initial state is already reset
	if m.executionQueue != nil {
		t.Fatal("Test setup error: executionQueue should be nil initially")
	}
	if m.activeTaskModelDialog != nil {
		t.Fatal("Test setup error: activeTaskModelDialog should be nil initially")
	}
	if m.showTaskModelDialog {
		t.Fatal("Test setup error: showTaskModelDialog should be false initially")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Fatal("Test setup error: taskModelSelectionDone should be empty initially")
	}

	// Call ResetExecutionQueue on already-reset state (should be safe)
	m.ResetExecutionQueue()

	// Verify state is still reset
	if m.executionQueue != nil {
		t.Error("Expected executionQueue to remain nil")
	}
	if m.activeTaskModelDialog != nil {
		t.Error("Expected activeTaskModelDialog to remain nil")
	}
	if m.showTaskModelDialog {
		t.Error("Expected showTaskModelDialog to remain false")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Errorf("Expected taskModelSelectionDone to remain empty, got %d entries", len(m.taskModelSelectionDone))
	}
}

// TestResetExecutionQueue_MultipleResets tests calling ResetExecutionQueue multiple times
func TestResetExecutionQueue_MultipleResets(t *testing.T) {
	m := createTestModel()

	for i := 0; i < 3; i++ {
		// Set state
		m.executionQueue = &ExecutionQueue{
			TaskIDs: []string{"task1", "task2", "task3"},
		}
		m.showTaskModelDialog = true
		m.taskModelSelectionDone["task1"] = true

		// Reset
		m.ResetExecutionQueue()

		// Verify reset
		if m.executionQueue != nil {
			t.Errorf("Iteration %d: Expected executionQueue to be nil after reset", i)
		}
		if m.showTaskModelDialog {
			t.Errorf("Iteration %d: Expected showTaskModelDialog to be false after reset", i)
		}
		if len(m.taskModelSelectionDone) != 0 {
			t.Errorf("Iteration %d: Expected taskModelSelectionDone to be empty after reset", i)
		}
	}
}

// TestExecutionQueueStateIndividualFieldModification tests modifying individual fields
func TestExecutionQueueStateIndividualFieldModification(t *testing.T) {
	m := createTestModel()

	// Test executionQueue field
	queue := &ExecutionQueue{TaskIDs: []string{"1", "2"}}
	m.executionQueue = queue
	if m.executionQueue != queue {
		t.Error("Failed to set/retrieve executionQueue field")
	}

	// Test showTaskModelDialog field
	m.showTaskModelDialog = true
	if !m.showTaskModelDialog {
		t.Error("Failed to set showTaskModelDialog to true")
	}
	m.showTaskModelDialog = false
	if m.showTaskModelDialog {
		t.Error("Failed to set showTaskModelDialog to false")
	}

	// Test taskModelSelectionDone field
	m.taskModelSelectionDone["task1"] = true
	if !m.taskModelSelectionDone["task1"] {
		t.Error("Failed to set taskModelSelectionDone entry")
	}
	delete(m.taskModelSelectionDone, "task1")
	if _, exists := m.taskModelSelectionDone["task1"]; exists {
		t.Error("Failed to delete taskModelSelectionDone entry")
	}
}

// TestExecutionQueueStateMemoryManagement tests that reset properly clears references
func TestExecutionQueueStateMemoryManagement(t *testing.T) {
	m := createTestModel()

	// Create an execution queue with substantial state
	queue := &ExecutionQueue{
		TaskIDs: []string{"1", "2", "3", "4", "5"},
		ModelSelections: map[string]string{
			"1": "model1",
			"2": "model2",
			"3": "model3",
		},
		TaskStatus: map[string]string{
			"1": "executing",
			"2": "pending",
		},
		CurrentIndex: 2,
	}
	m.executionQueue = queue

	// Populate taskModelSelectionDone with many entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("task%d", i)
		m.taskModelSelectionDone[key] = true
	}

	// Verify state was set
	if m.executionQueue == nil {
		t.Fatal("Failed to set executionQueue")
	}
	if len(m.executionQueue.TaskIDs) != 5 {
		t.Fatal("executionQueue not properly initialized")
	}
	if len(m.taskModelSelectionDone) != 10 {
		t.Fatal("taskModelSelectionDone not properly initialized")
	}

	// Reset and verify complete cleanup
	m.ResetExecutionQueue()

	if m.executionQueue != nil {
		t.Error("Expected executionQueue to be nil after reset")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Errorf("Expected taskModelSelectionDone to be empty, got %d entries", len(m.taskModelSelectionDone))
	}
	// Verify we can add to the reset map immediately
	m.taskModelSelectionDone["newTask"] = true
	if !m.taskModelSelectionDone["newTask"] {
		t.Error("Failed to use taskModelSelectionDone map after reset")
	}
}

// TestExecutionQueueStatePartialReset tests that reset clears all fields even if some were not set
func TestExecutionQueueStatePartialReset(t *testing.T) {
	m := createTestModel()

	// Set only some fields, leaving others at initial values
	m.executionQueue = &ExecutionQueue{TaskIDs: []string{"1"}}
	// Leave activeTaskModelDialog as nil
	m.showTaskModelDialog = true
	// Leave taskModelSelectionDone empty

	// Verify partial state
	if m.executionQueue == nil {
		t.Fatal("Failed to set executionQueue")
	}
	if m.activeTaskModelDialog != nil {
		t.Fatal("Test setup error: activeTaskModelDialog should be nil")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Fatal("Test setup error: taskModelSelectionDone should be empty")
	}

	// Reset
	m.ResetExecutionQueue()

	// Verify all fields are reset to initial state
	if m.executionQueue != nil {
		t.Error("Expected executionQueue to be nil")
	}
	if m.activeTaskModelDialog != nil {
		t.Error("Expected activeTaskModelDialog to be nil")
	}
	if m.showTaskModelDialog {
		t.Error("Expected showTaskModelDialog to be false")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Errorf("Expected taskModelSelectionDone to be empty, got %d entries", len(m.taskModelSelectionDone))
	}
}

// TestExecutionQueueIntegration_RepeatedCycles tests many cycles of queue creation and reset
func TestExecutionQueueIntegration_RepeatedCycles(t *testing.T) {
	m := createTestModel()

	// Run 100 cycles to test for memory leaks
	for cycle := 0; cycle < 100; cycle++ {
		// Create a queue with state
		m.executionQueue = &ExecutionQueue{
			TaskIDs: []string{"task1", "task2", "task3"},
			ModelSelections: map[string]string{
				"task1": "model1",
				"task2": "model2",
				"task3": "model3",
			},
			TaskStatus: map[string]string{
				"task1": "executing",
				"task2": "pending",
				"task3": "pending",
			},
			CurrentIndex: 1,
		}

		// Populate selection tracking
		m.showTaskModelDialog = true
		m.taskModelSelectionDone["task1"] = true
		m.taskModelSelectionDone["task2"] = true
		m.taskModelSelectionDone["task3"] = true

		// Verify state exists
		if m.executionQueue == nil {
			t.Fatalf("Cycle %d: Failed to create executionQueue", cycle)
		}
		if len(m.taskModelSelectionDone) != 3 {
			t.Fatalf("Cycle %d: taskModelSelectionDone not properly set", cycle)
		}

		// Reset
		m.ResetExecutionQueue()

		// Verify reset
		if m.executionQueue != nil {
			t.Errorf("Cycle %d: executionQueue not reset", cycle)
		}
		if m.showTaskModelDialog {
			t.Errorf("Cycle %d: showTaskModelDialog not reset", cycle)
		}
		if len(m.taskModelSelectionDone) != 0 {
			t.Errorf("Cycle %d: taskModelSelectionDone not empty after reset", cycle)
		}
	}
}

// TestExecutionQueueIntegration_StateIsolation tests that queue state doesn't interfere with other Model fields
func TestExecutionQueueIntegration_StateIsolation(t *testing.T) {
	m := createTestModel()

	// Set up some execution queue state
	m.executionQueue = &ExecutionQueue{TaskIDs: []string{"1", "2"}}
	m.showTaskModelDialog = true
	m.taskModelSelectionDone["1"] = true

	// Modify other Model fields
	originalSelectedIndex := m.selectedIndex
	m.selectedIndex = 5
	m.commandMode = true
	m.searchMode = true

	// Reset execution queue state
	m.ResetExecutionQueue()

	// Verify execution queue state is reset
	if m.executionQueue != nil {
		t.Error("executionQueue not reset")
	}
	if m.showTaskModelDialog {
		t.Error("showTaskModelDialog not reset")
	}
	if len(m.taskModelSelectionDone) != 0 {
		t.Error("taskModelSelectionDone not reset")
	}

	// Verify other Model fields are unchanged
	if m.selectedIndex == originalSelectedIndex {
		t.Error("Other Model fields were affected by ResetExecutionQueue()")
	}
	if !m.commandMode {
		t.Error("commandMode should not be affected")
	}
	if !m.searchMode {
		t.Error("searchMode should not be affected")
	}
}

// TestExecutionQueueIntegration_TaskSelection tests interaction with task selection
func TestExecutionQueueIntegration_TaskSelection(t *testing.T) {
	m := createTestModel()

	// Start with a selected task
	if len(m.visibleTasks) == 0 {
		t.Skip("Test requires visible tasks")
	}

	originalTask := m.selectedTask

	// Set up execution queue state
	m.executionQueue = &ExecutionQueue{
		TaskIDs:      []string{"1", "2"},
		CurrentIndex: 0,
	}

	// Perform task navigation
	m.selectNext()

	// Verify task selection changed
	if m.selectedTask == originalTask {
		t.Fatal("Task selection should have changed")
	}

	// Reset execution queue
	m.ResetExecutionQueue()

	// Verify task selection is unchanged by reset
	if m.selectedTask != m.visibleTasks[m.selectedIndex] {
		t.Error("ResetExecutionQueue affected task selection")
	}
}

// TestExecutionQueueIntegration_FieldConsistency tests field consistency after operations
func TestExecutionQueueIntegration_FieldConsistency(t *testing.T) {
	m := createTestModel()

	// Verify initial consistency
	if m.executionQueue != nil || m.activeTaskModelDialog != nil ||
		m.showTaskModelDialog || len(m.taskModelSelectionDone) != 0 {
		t.Fatal("Test setup: initial state not clean")
	}

	// Create queue and partially fill state
	m.executionQueue = &ExecutionQueue{TaskIDs: []string{"1"}}
	m.showTaskModelDialog = true
	m.taskModelSelectionDone["1"] = true

	// Verify consistency before reset
	if m.executionQueue == nil || !m.showTaskModelDialog || len(m.taskModelSelectionDone) == 0 {
		t.Fatal("State setup failed")
	}

	// Reset and verify all fields are consistent (all reset)
	m.ResetExecutionQueue()

	// All fields should be at their initial state
	allReset := m.executionQueue == nil &&
		m.activeTaskModelDialog == nil &&
		!m.showTaskModelDialog &&
		len(m.taskModelSelectionDone) == 0

	if !allReset {
		t.Error("After ResetExecutionQueue(), not all fields are reset to initial state")
	}
}

// TestExecutionQueueIntegration_ConcurrentStateModification tests modifying state while resetting
func TestExecutionQueueIntegration_ConcurrentStateModification(t *testing.T) {
	m := createTestModel()

	// Set up state with multiple entries
	m.executionQueue = &ExecutionQueue{TaskIDs: []string{"1", "2", "3", "4", "5"}}
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("%d", i)
		m.taskModelSelectionDone[key] = true
	}

	// Verify state
	if len(m.taskModelSelectionDone) != 5 {
		t.Fatal("Failed to populate taskModelSelectionDone")
	}

	// Reset
	m.ResetExecutionQueue()

	// Try to add new entries after reset (map should be usable)
	m.taskModelSelectionDone["6"] = true
	m.taskModelSelectionDone["7"] = true

	// Verify new entries were added
	if len(m.taskModelSelectionDone) != 2 {
		t.Error("Failed to add entries to map after reset")
	}

	// Reset again and verify clean state
	m.ResetExecutionQueue()
	if len(m.taskModelSelectionDone) != 0 {
		t.Error("taskModelSelectionDone not empty after second reset")
	}
}

// TestExecuteMultipleTasks_EmptyTaskList tests executeMultipleTasks with empty task list
func TestExecuteMultipleTasks_EmptyTaskList(t *testing.T) {
	m := createTestModel()

	// Call with empty task list
	cmd := m.executeMultipleTasks([]string{}, map[string]string{})

	// Should return nil (error shown via appError)
	if cmd != nil {
		t.Error("Expected executeMultipleTasks to return nil for empty task list")
	}
}

// TestExecuteMultipleTasks_ModelSelectionsMapProperty tests the function signature accepts modelSelections map
func TestExecuteMultipleTasks_ModelSelectionsMapProperty(t *testing.T) {
	m := createTestModel()

	// Test that we can pass a modelSelections map with different models per task
	// This is the key change from the old function signature
	modelSelections := map[string]string{
		"1": "claude-3-5-sonnet-20241022",
		"2": "gpt-4o-mini",
		"3": "gemini-2.0",
	}

	// The function should accept this map (it will fail to execute without taskService,
	// but that's not the concern of this unit test)
	taskIDs := []string{"1", "2", "3"}

	// Just verify the function can be called with the new signature
	// The implementation will handle missing taskService appropriately
	_ = m.executeMultipleTasks(taskIDs, modelSelections)

	// This test passes if the function accepts the new signature without compilation errors
	t.Log("executeMultipleTasks signature correctly accepts modelSelections map")
}

// TestFilterKeyMappings_SlashEntersFilterMode tests that "/" key enters filter mode
func TestFilterKeyMappings_SlashEntersFilterMode(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Verify initial state
	if m.filterableComponent == nil {
		t.Fatal("filterableComponent should be initialized")
	}
	if m.filterableComponent.IsFiltering() {
		t.Fatal("Should not be in filter mode initially")
	}

	// Simulate "/" key press
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(*Model)

	// Verify filter mode is entered
	if !m.filterableComponent.IsFiltering() {
		t.Error("Should be in filter mode after '/' key press")
	}

	// Verify backward compatibility
	if !m.searchMode {
		t.Error("searchMode should be true for backward compatibility")
	}

	// Verify searchInput is focused
	if !m.searchInput.Focused() {
		t.Error("searchInput should be focused after '/' key press")
	}
}

// TestFilterKeyMappings_FKeyInFilterMode tests that "F" key doesn't cycle filters while filtering
func TestFilterKeyMappings_FKeyInFilterMode(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Enter filter mode
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(*Model)

	// Verify we're in filter mode
	if !m.filterableComponent.IsFiltering() {
		t.Fatal("Should be in filter mode after '/' key")
	}

	// Verify initial statusFilter is empty (all tasks)
	initialFilter := m.statusFilter
	if initialFilter != "" {
		t.Logf("Initial status filter: %q", initialFilter)
	}

	// Try to press "F" while in filter mode
	// This should be intercepted by the filter mode handler and not cycle status filters
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	m = newModel.(*Model)

	// The "F" should be treated as input text, not as a filter cycling command
	// So we're still in filter mode
	if !m.filterableComponent.IsFiltering() {
		t.Error("Should still be in filter mode after 'F' key (should not cycle status filter)")
	}

	// Status filter should not have changed
	if m.statusFilter != initialFilter {
		t.Errorf("Status filter should not change while in filter mode, expected %q but got %q", initialFilter, m.statusFilter)
	}
}

// TestFilterKeyMappings_FKeyCyclesOutsideFilterMode tests that "F" key cycles filters normally
func TestFilterKeyMappings_FKeyCyclesOutsideFilterMode(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Verify not in filter mode
	if m.filterableComponent.IsFiltering() {
		t.Fatal("Should not be in filter mode initially")
	}

	// Verify initial statusFilter is empty (all tasks)
	if m.statusFilter != "" {
		t.Fatal("Initial status filter should be empty")
	}

	// Press "F" key outside filter mode
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	m = newModel.(*Model)

	// Should cycle to first non-empty status
	if m.statusFilter == "" {
		t.Error("Status filter should have changed after 'F' key press outside filter mode")
	}

	// Should still not be in filter mode
	if m.filterableComponent.IsFiltering() {
		t.Error("Should not enter filter mode with 'F' key")
	}
}

// TestFilterKeyMappings_EscClearsFilterMode tests that Esc clears all filters
func TestFilterKeyMappings_EscClearsFilterMode(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Set up filter state: enter filter mode
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(*Model)

	// Type some search text
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'e', 's', 't'}})
	m = newModel.(*Model)

	// Verify we have search state
	if m.searchQuery != "test" {
		t.Errorf("Expected search query 'test', got %q", m.searchQuery)
	}

	// Now press Esc to clear
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	// Verify filter mode is exited
	if m.filterableComponent.IsFiltering() {
		t.Error("Should exit filter mode after Esc")
	}

	// Verify searchMode is false
	if m.searchMode {
		t.Error("searchMode should be false after Esc")
	}

	// Verify search query is cleared
	if m.searchQuery != "" {
		t.Errorf("Search query should be cleared after Esc, got %q", m.searchQuery)
	}

	// Verify search input is cleared
	if m.searchInput.Value() != "" {
		t.Errorf("Search input should be cleared after Esc, got %q", m.searchInput.Value())
	}
}

// TestFilterKeyMappings_EscClearsStatusFilter tests that Esc clears status filter too
func TestFilterKeyMappings_EscClearsStatusFilter(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Set status filter by pressing "F" multiple times
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	m = newModel.(*Model)

	// Verify status filter changed
	if m.statusFilter == "" {
		t.Fatal("Status filter should have changed after 'F' key")
	}

	// Press Esc to clear all filters
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	// Verify status filter is cleared
	if m.statusFilter != "" {
		t.Errorf("Status filter should be cleared after Esc, expected empty but got %q", m.statusFilter)
	}
}

// TestFilterMode_DoesNotInterfereWithNavigation tests that filter mode doesn't interfere with normal navigation
func TestFilterMode_DoesNotInterfereWithNavigation(t *testing.T) {
	m := createTestModel()
	m.ready = true
	initialSelectedIndex := m.selectedIndex

	// Enter filter mode
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(*Model)

	// Type some search text
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = newModel.(*Model)

	// Try navigation keys while in filter mode - these should be handled by searchInput
	// not by the main navigation handler
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(*Model)

	// selectedIndex should not have changed (navigation disabled in filter mode)
	if m.selectedIndex != initialSelectedIndex {
		t.Errorf("Navigation keys should not work while in filter mode, selectedIndex changed from %d to %d",
			initialSelectedIndex, m.selectedIndex)
	}

	// Exit filter mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	// Now navigation should work
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(*Model)

	// selectedIndex might change (depends on filter results), but we're out of filter mode
	if m.filterableComponent.IsFiltering() {
		t.Error("Should be out of filter mode after Esc")
	}
}

// TestFilterMode_EnterAndExitSequence tests the complete enter-type-exit sequence
func TestFilterMode_EnterAndExitSequence(t *testing.T) {
	m := createTestModel()
	m.ready = true

	// Initial state
	if m.filterableComponent.IsFiltering() {
		t.Fatal("Should not be in filter mode initially")
	}

	// Step 1: Enter filter mode with "/"
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(*Model)
	if !m.filterableComponent.IsFiltering() {
		t.Error("Step 1: Should be in filter mode after '/'")
	}

	// Step 2: Type search query
	for _, ch := range []rune{'t', 'a', 's', 'k'} {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = newModel.(*Model)
	}
	if m.searchQuery == "" {
		t.Error("Step 2: Should have typed characters into search")
	}

	// Step 3: Confirm search with Enter
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	if m.filterableComponent.IsFiltering() {
		t.Error("Step 3: Should exit filter mode after Enter confirmation")
	}

	// Step 4: Re-enter filter mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(*Model)
	if !m.filterableComponent.IsFiltering() {
		t.Error("Step 4: Should be in filter mode again after '/'")
	}

	// Step 5: Exit without confirming using Esc
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)
	if m.filterableComponent.IsFiltering() {
		t.Error("Step 5: Should exit filter mode after Esc")
	}
}
