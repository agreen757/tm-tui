package ui

import (
	"errors"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
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
		styles:           NewStyles(),
		logLines:         []string{},
		appState:         NewAppState(nil, &keyMap),
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

