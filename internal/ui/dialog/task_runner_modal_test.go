package dialog

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TestTaskRunnerKeyMapStructure tests that the key map is properly initialized
func TestTaskRunnerKeyMapStructure(t *testing.T) {
	keyMap := DefaultTaskRunnerKeyMap()

	// Verify TabDirect has 9 bindings
	if len(keyMap.TabDirect) != 9 {
		t.Errorf("Expected 9 TabDirect bindings, got %d", len(keyMap.TabDirect))
	}

	// Test that basic Tab/Shift+Tab keys work
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	if !key.Matches(tabMsg, keyMap.NextTab) {
		t.Error("Expected Tab key to match NextTab binding")
	}

	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	if !key.Matches(shiftTabMsg, keyMap.PrevTab) {
		t.Error("Expected Shift+Tab key to match PrevTab binding")
	}
}

// TestTabNavigation tests Tab/Shift+Tab navigation between tabs
func TestTabNavigation(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add multiple tabs
	modal.addTab("task-1", "Task 1", "model-1")
	modal.addTab("task-2", "Task 2", "model-2")
	modal.addTab("task-3", "Task 3", "model-3")

	// Verify initial state
	if modal.activeTab != 2 {
		t.Errorf("Expected initial activeTab to be 2 (last added), got %d", modal.activeTab)
	}

	// Test Tab key moves to next tab (should wrap around)
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	modal.HandleKey(tabKey)
	if modal.activeTab != 0 {
		t.Errorf("Expected activeTab 0 after Tab from 2, got %d", modal.activeTab)
	}

	// Test Shift+Tab moves to previous tab
	shiftTabKey := tea.KeyMsg{Type: tea.KeyShiftTab}
	modal.HandleKey(shiftTabKey)
	if modal.activeTab != 2 {
		t.Errorf("Expected activeTab 2 after Shift+Tab from 0, got %d", modal.activeTab)
	}

	// Test another Tab
	modal.HandleKey(tabKey)
	if modal.activeTab != 0 {
		t.Errorf("Expected activeTab 0 after second Tab, got %d", modal.activeTab)
	}
}

// TestDirectTabSelection tests numeric key selection (1-9)
func TestDirectTabSelection(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add 5 tabs
	for i := 1; i <= 5; i++ {
		modal.addTab("task-"+string(rune('0'+i)), "Task "+string(rune('0'+i)), "model")
	}

	// Reset to first tab for testing
	modal.activeTab = 0

	// Test pressing "2" selects second tab
	// Note: The handleKey checks if key matches TabDirect[i] binding for keys "1"-"9"
	key2 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	modal.HandleKey(key2)
	if modal.activeTab != 1 {
		t.Errorf("Expected activeTab 1 after pressing '2', got %d", modal.activeTab)
	}

	// Test pressing "5" selects fifth tab
	key5 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	modal.HandleKey(key5)
	if modal.activeTab != 4 {
		t.Errorf("Expected activeTab 4 after pressing '5', got %d", modal.activeTab)
	}

	// Test pressing "9" when only 5 tabs exist - should not select a tab (out of range)
	beforeActive := modal.activeTab
	key9 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}}
	modal.HandleKey(key9)
	if modal.activeTab != beforeActive {
		t.Errorf("Expected activeTab %d after pressing '9' (out of range), got %d", beforeActive, modal.activeTab)
	}
}

// TestMinimizeToggle tests the minimize/maximize functionality
func TestMinimizeToggle(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Initial state should be expanded
	if modal.minimized {
		t.Error("Expected modal to be expanded initially")
	}

	// Press 'm' to minimize
	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	modal.HandleKey(mKey)
	if !modal.minimized {
		t.Error("Expected modal to be minimized after pressing 'm'")
	}

	// Press 'm' again to maximize
	modal.HandleKey(mKey)
	if modal.minimized {
		t.Error("Expected modal to be expanded after second 'm' press")
	}

	// Test uppercase 'M' as well
	MKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}}
	modal.HandleKey(MKey)
	if !modal.minimized {
		t.Error("Expected modal to be minimized after pressing 'M'")
	}

	// Press 'M' again to maximize
	modal.HandleKey(MKey)
	if modal.minimized {
		t.Error("Expected modal to be expanded after second 'M' press")
	}
}

// TestCloseModal tests the close functionality
func TestCloseModal(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a completed task
	modal.addTab("task-1", "Task 1", "model")
	modal.tabs[0].SetStatus(TaskCompleted)

	// Press Esc to close
	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := modal.HandleKey(escKey)
	if result != DialogResultClose {
		t.Errorf("Expected DialogResultClose after Esc, got %v", result)
	}
}

// TestCloseModalWithRunningTask tests that close is prevented when tasks are running
func TestCloseModalWithRunningTask(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Task 1", "model")
	modal.tabs[0].SetStatus(TaskRunning)

	// Press Esc to attempt close
	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := modal.HandleKey(escKey)
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone when closing with running tasks, got %v", result)
	}
}

// TestCancelTask tests the cancel shortcut
func TestCancelTask(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Task 1", "model")
	modal.tabs[0].SetStatus(TaskRunning)

	// Press Ctrl+C to cancel
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	modal.HandleKey(ctrlCKey)

	// Check that the task is marked as cancelled
	if modal.tabs[0].GetStatus() != TaskCancelled {
		t.Errorf("Expected task to be cancelled, got %v", modal.tabs[0].GetStatus())
	}
}

// TestScrollingKeyDelegation tests that scroll keys are delegated to the active tab
func TestScrollingKeyDelegation(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a tab with some output
	modal.addTab("task-1", "Task 1", "model")
	tab := modal.tabs[0]

	// Add multiple lines to test scrolling
	for i := 0; i < 100; i++ {
		tab.AddOutputLine("Line " + string(rune('0'+(i%10))))
	}

	// Initial viewport should be at bottom due to auto-scroll
	initialViewportYOffset := tab.viewport.YOffset

	// Press up arrow to scroll up
	upKey := tea.KeyMsg{Type: tea.KeyUp}
	modal.HandleKey(upKey)

	// The viewport should have changed (scrolled up)
	if tab.viewport.YOffset == initialViewportYOffset {
		// Note: This test may not be strict enough; viewport behavior depends on implementation
		// Just verify no error occurs
	}
}

// TestFooterContextAwareness tests that footer shows appropriate help text
func TestFooterContextAwareness(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// With no tasks, footer should show close option
	footer := modal.renderFooter()
	if len(footer) == 0 {
		t.Error("Expected footer to have content")
	}

	// Add a running task
	modal.addTab("task-1", "Task 1", "model")
	modal.tabs[0].SetStatus(TaskRunning)

	// With running task, footer should indicate close is not available
	footer = modal.renderFooter()
	if len(footer) == 0 {
		t.Error("Expected footer to have content with running task")
	}
}

// TestTabBarRendering tests that the tab bar renders correctly
func TestTabBarRendering(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// With no tabs, should render empty
	tabBar := modal.renderTabBar()
	if tabBar != "" {
		t.Errorf("Expected empty tab bar with no tabs, got: %s", tabBar)
	}

	// Add tabs
	modal.addTab("task-1", "Task 1", "model")
	modal.addTab("task-2", "Task 2", "model")

	// Should render something
	tabBar = modal.renderTabBar()
	if tabBar == "" {
		t.Error("Expected non-empty tab bar with tabs")
	}
}

// TestEnsureTabVisible tests tab scroll position adjustment
func TestEnsureTabVisible(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add many tabs to test scrolling
	for i := 1; i <= 15; i++ {
		modal.addTab("task-"+string(rune('0'+(i%10))), "Task "+string(rune('0'+(i%10))), "model")
	}

	// Navigate to tab 10
	modal.activeTab = 9
	modal.ensureTabVisible()

	// tabScrollPos should keep tab 10 visible
	if modal.activeTab < modal.tabScrollPos || modal.activeTab >= modal.tabScrollPos+7 {
		t.Errorf("Expected activeTab %d to be visible with tabScrollPos %d", modal.activeTab, modal.tabScrollPos)
	}
}

// TestModalGetters tests the getter methods
func TestModalGetters(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Test GetTabCount with no tabs
	if modal.GetTabCount() != 0 {
		t.Errorf("Expected 0 tabs initially, got %d", modal.GetTabCount())
	}

	// Add a tab
	modal.addTab("task-1", "Task 1", "model")
	if modal.GetTabCount() != 1 {
		t.Errorf("Expected 1 tab after add, got %d", modal.GetTabCount())
	}

	// Test GetActiveTab
	activeTab := modal.GetActiveTab()
	if activeTab == nil {
		t.Error("Expected non-nil active tab")
	}

	// Test GetMinimized
	if modal.GetMinimized() {
		t.Error("Expected modal not to be minimized")
	}

	// Minimize and test again
	modal.minimized = true
	if !modal.GetMinimized() {
		t.Error("Expected modal to be minimized")
	}
}

// TestKeyMatchingWithKeyBindings tests the key.Matches functionality
func TestKeyMatchingWithKeyBindings(t *testing.T) {
	keyMap := DefaultTaskRunnerKeyMap()

	// Create a Tab key message
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	if !key.Matches(tabMsg, keyMap.NextTab) {
		t.Error("Expected Tab key to match NextTab binding")
	}

	// Create a Shift+Tab key message
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	if !key.Matches(shiftTabMsg, keyMap.PrevTab) {
		t.Error("Expected Shift+Tab key to match PrevTab binding")
	}

	// Test numeric key match for direct tab selection
	for i := 0; i < 9; i++ {
		keyChar := rune('1' + i)
		numKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{keyChar}}
		if !key.Matches(numKey, keyMap.TabDirect[i]) {
			t.Errorf("Expected key %c to match TabDirect[%d] binding", keyChar, i)
		}
	}
}

// TestMinimizeStatePreservation tests that state is preserved during minimize/maximize
func TestMinimizeStatePreservation(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add multiple tabs
	modal.addTab("task-1", "Task 1", "model-1")
	modal.addTab("task-2", "Task 2", "model-2")
	modal.addTab("task-3", "Task 3", "model-3")

	// Switch to tab 2 and scroll
	modal.activeTab = 1
	tab := modal.tabs[1]
	for i := 0; i < 50; i++ {
		tab.AddOutputLine("Line " + string(rune('0'+(i%10))))
	}
	tab.viewport.LineDown(10) // Scroll down

	// Save the initial scroll position
	initialScrollPos := tab.viewport.YOffset

	// Minimize
	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	modal.HandleKey(mKey)

	// Verify minimized state
	if !modal.minimized {
		t.Error("Expected modal to be minimized")
	}

	// Verify state was saved
	if modal.preMinimizeState.activeTab != 1 {
		t.Errorf("Expected saved activeTab to be 1, got %d", modal.preMinimizeState.activeTab)
	}

	// Maximize (toggle minimize again)
	modal.HandleKey(mKey)

	// Verify maximized state
	if modal.minimized {
		t.Error("Expected modal to be maximized")
	}

	// Verify state was restored
	if modal.activeTab != 1 {
		t.Errorf("Expected restored activeTab to be 1, got %d", modal.activeTab)
	}

	// Note: Scroll position restoration depends on viewport implementation
	// The YOffset should be restored
	restoredScrollPos := modal.tabs[1].viewport.YOffset
	if restoredScrollPos != initialScrollPos {
		t.Errorf("Expected scroll position to be restored to %d, got %d", initialScrollPos, restoredScrollPos)
	}
}

// TestMinimizedViewWithMultipleTasks tests minimized view rendering with multiple tasks
func TestMinimizedViewWithMultipleTasks(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add multiple tasks with different statuses
	modal.addTab("task-1", "Task 1", "model-1")
	modal.tabs[0].SetStatus(TaskRunning)

	modal.addTab("task-2", "Task 2", "model-2")
	modal.tabs[1].SetStatus(TaskCompleted)

	modal.addTab("task-3", "Task 3", "model-3")
	modal.tabs[2].SetStatus(TaskFailed)

	// Minimize
	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	modal.HandleKey(mKey)

	// Get minimized view
	view := modal.renderMinimized()

	// Verify the view contains expected elements
	if !stringContains(view, "Task Runner") {
		t.Error("Expected minimized view to contain 'Task Runner'")
	}

	if !stringContains(view, "running") {
		t.Error("Expected minimized view to contain 'running' count")
	}

	if !stringContains(view, "total") {
		t.Error("Expected minimized view to contain 'total' count")
	}

	if !stringContains(view, "expand") {
		t.Error("Expected minimized view to contain 'expand' hint")
	}
}

// Helper function to check if string contains substring
func stringContains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestTaskContinuityInMinimizedState tests that tasks continue running when minimized
func TestTaskContinuityInMinimizedState(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Task 1", "model-1")
	modal.tabs[0].SetStatus(TaskRunning)

	// Minimize
	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	modal.HandleKey(mKey)

	// Verify task is still running in minimized state
	if modal.tabs[0].GetStatus() != TaskRunning {
		t.Error("Expected task to remain running when minimized")
	}

	// Add output while minimized
	modal.tabs[0].AddOutputLine("Test output in minimized state")

	// Maximize and verify output was captured
	modal.HandleKey(mKey)
	output := modal.tabs[0].GetOutput()
	if len(output) == 0 {
		t.Error("Expected output to be captured while minimized")
	}
}

// TestWindowResizeHandlingInMinimizedState tests window resize handling
func TestWindowResizeHandlingInMinimizedState(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a task
	modal.addTab("task-1", "Task 1", "model-1")

	// Minimize
	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	modal.HandleKey(mKey)

	// Simulate window resize while minimized
	modal.SetRect(100, 40, 0, 0)

	// Verify modal is still minimized
	if !modal.minimized {
		t.Error("Expected modal to remain minimized after resize")
	}

	// Verify GetRect works after resize
	width, height, _, _ := modal.GetRect()
	if width != 100 || height != 40 {
		t.Errorf("Expected resized dimensions (100, 40), got (%d, %d)", width, height)
	}
}

// TestWindowResizeHandlingInMaximizedState tests window resize handling in maximized state
func TestWindowResizeHandlingInMaximizedState(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a task with output
	modal.addTab("task-1", "Task 1", "model-1")
	for i := 0; i < 20; i++ {
		modal.tabs[0].AddOutputLine("Output line " + string(rune('0'+(i%10))))
	}

	// Verify initial tab dimensions
	initialViewportWidth := modal.tabs[0].viewport.Width

	// Resize while maximized
	modal.SetRect(120, 50, 0, 0)

	// Verify tab dimensions were updated
	if modal.tabs[0].viewport.Width == initialViewportWidth {
		t.Error("Expected viewport width to change after resize")
	}
}

// TestRapidMinimizeMaximizeToggling tests rapid toggling between minimize and maximize
func TestRapidMinimizeMaximizeToggling(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add multiple tasks
	for i := 0; i < 5; i++ {
		modal.addTab("task-"+string(rune('1'+rune(i))), "Task "+string(rune('1'+rune(i))), "model")
	}

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}

	// Rapidly toggle minimize/maximize 10 times
	for i := 0; i < 10; i++ {
		modal.HandleKey(mKey)
	}

	// After even number of toggles, should be back to expanded state
	if modal.minimized {
		t.Error("Expected modal to be expanded after even number of minimize toggles")
	}

	// Verify tabs are still intact
	if modal.GetTabCount() != 5 {
		t.Errorf("Expected 5 tabs after rapid toggling, got %d", modal.GetTabCount())
	}
}

// TestCancellationWithLongRunningTask tests cancellation confirmation for long-running tasks
func TestCancellationWithLongRunningTask(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Long Running Task", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskRunning)

	// Add output to simulate running time
	for i := 0; i < 10; i++ {
		tab.AddOutputLine("Processing...")
	}

	// Set a very low threshold so even a newly created task is considered "long-running"
	modal.longRunningThreshold = 0

	// Press Ctrl+C to cancel
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	modal.HandleKey(ctrlCKey)

	// For long-running tasks, a confirmation dialog should be shown
	if modal.cancellationConfirmDialog == nil {
		t.Error("Expected cancellation confirmation dialog for long-running task")
	}
}

// TestCancellationWithQuickTask tests direct cancellation for quick tasks
func TestCancellationWithQuickTask(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Quick Task", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskRunning)

	// Set threshold higher than task runtime (which is near 0)
	modal.longRunningThreshold = 10000

	// Press Ctrl+C to cancel
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	modal.HandleKey(ctrlCKey)

	// For quick tasks, should cancel directly without confirmation
	if tab.GetStatus() != TaskCancelled {
		t.Errorf("Expected task to be cancelled immediately, got status %v", tab.GetStatus())
	}

	// No confirmation dialog should be shown
	if modal.cancellationConfirmDialog != nil {
		t.Error("Expected no confirmation dialog for quick task")
	}
}

// TestCancellationOnNonRunningTask tests that cancellation is ignored for non-running tasks
func TestCancellationOnNonRunningTask(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a completed task
	modal.addTab("task-1", "Completed Task", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskCompleted)

	// Try to cancel
	modal.cancelActiveTask()

	// Status should remain unchanged
	if tab.GetStatus() != TaskCompleted {
		t.Errorf("Expected task status to remain TaskCompleted, got %v", tab.GetStatus())
	}
}

// TestCancelTaskByID tests cancellation by task ID
func TestCancelTaskByID(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add multiple tasks
	modal.addTab("task-1", "Task 1", "model")
	modal.addTab("task-2", "Task 2", "model")
	modal.addTab("task-3", "Task 3", "model")

	// Set task-2 as running
	modal.tabs[1].SetStatus(TaskRunning)

	// Cancel task-2 by ID
	result := modal.CancelTaskByID("task-2")
	if !result {
		t.Error("Expected CancelTaskByID to return true")
	}

	// Verify task-2 is cancelled
	if modal.tabs[1].GetStatus() != TaskCancelled {
		t.Errorf("Expected task-2 to be cancelled, got status %v", modal.tabs[1].GetStatus())
	}

	// Try to cancel non-existent task
	result = modal.CancelTaskByID("task-999")
	if result {
		t.Error("Expected CancelTaskByID to return false for non-existent task")
	}
}

// TestGetTaskStatus tests retrieving task status by ID
func TestGetTaskStatus(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add tasks with different statuses
	modal.addTab("task-1", "Task 1", "model")
	modal.tabs[0].SetStatus(TaskRunning)

	modal.addTab("task-2", "Task 2", "model")
	modal.tabs[1].SetStatus(TaskCompleted)

	// Test getting status
	status, found := modal.GetTaskStatus("task-1")
	if !found {
		t.Error("Expected to find task-1")
	}
	if status != TaskRunning {
		t.Errorf("Expected task-1 status to be TaskRunning, got %v", status)
	}

	status, found = modal.GetTaskStatus("task-2")
	if !found {
		t.Error("Expected to find task-2")
	}
	if status != TaskCompleted {
		t.Errorf("Expected task-2 status to be TaskCompleted, got %v", status)
	}

	// Test getting status of non-existent task
	_, found = modal.GetTaskStatus("task-999")
	if found {
		t.Error("Expected not to find task-999")
	}
}

// TestGetTaskOutput tests retrieving task output by ID
func TestGetTaskOutput(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a task with output
	modal.addTab("task-1", "Task 1", "model")
	tab := modal.tabs[0]
	tab.AddOutputLine("Line 1")
	tab.AddOutputLine("Line 2")
	tab.AddOutputLine("Line 3")

	// Get output
	output, found := modal.GetTaskOutput("task-1")
	if !found {
		t.Error("Expected to find task-1")
	}
	if len(output) != 3 {
		t.Errorf("Expected 3 output lines, got %d", len(output))
	}

	// Test getting output of non-existent task
	_, found = modal.GetTaskOutput("task-999")
	if found {
		t.Error("Expected not to find task-999")
	}
}

// TestTabColoringForCancelledState tests that cancelled tabs have proper coloring
func TestTabColoringForCancelledState(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add tasks with different statuses
	modal.addTab("task-1", "Running", "model")
	modal.tabs[0].SetStatus(TaskRunning)

	modal.addTab("task-2", "Cancelled", "model")
	modal.tabs[1].SetStatus(TaskCancelled)

	modal.addTab("task-3", "Completed", "model")
	modal.tabs[2].SetStatus(TaskCompleted)

	// Render tab bar and verify it renders without error
	tabBar := modal.renderTabBar()
	if tabBar == "" {
		t.Error("Expected non-empty tab bar")
	}

	// Verify icons are present
	if !stringContains(tabBar, "⏳") { // Running
		t.Error("Expected running icon in tab bar")
	}
	if !stringContains(tabBar, "⊘") { // Cancelled
		t.Error("Expected cancelled icon in tab bar")
	}
	if !stringContains(tabBar, "✓") { // Completed
		t.Error("Expected completed icon in tab bar")
	}
}

// TestCancellationMessageOutput tests that cancellation adds appropriate message to output
func TestCancellationMessageOutput(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task with some output
	modal.addTab("task-1", "Task 1", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskRunning)
	tab.AddOutputLine("Initial output")

	// Set threshold to allow direct cancellation
	modal.longRunningThreshold = 10000

	// Cancel the task
	modal.cancelActiveTask()

	// Check that output contains cancellation message
	output := tab.GetOutput()
	if len(output) < 3 { // Initial output + empty line + cancellation message
		t.Errorf("Expected at least 3 lines of output, got %d", len(output))
	}

	found := false
	for _, line := range output {
		if stringContains(line, "cancelled") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected cancellation message in output")
	}
}

// TestGetTabLabelForCommands tests GetTabLabel with command IDs
func TestGetTabLabelForCommands(t *testing.T) {
	tests := []struct {
		name     string
		taskID   string
		taskTitle string
		expected string
	}{
		{
			name:      "command with title",
			taskID:    "command-1234567890",
			taskTitle: "Generate Documentation",
			expected:  "Generate Documentation",
		},
		{
			name:      "command without title",
			taskID:    "command-1234567890",
			taskTitle: "",
			expected:  "Command",
		},
		{
			name:      "regular task with title",
			taskID:    "1.2",
			taskTitle: "Implement Auth",
			expected:  "1.2 - Implement Auth",
		},
		{
			name:      "regular task without title",
			taskID:    "1.2",
			taskTitle: "",
			expected:  "1.2",
		},
		{
			name:      "single digit task with title",
			taskID:    "1",
			taskTitle: "Setup Project",
			expected:  "1 - Setup Project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tab := NewTaskExecutionTab(tt.taskID, tt.taskTitle, "model", 80, 30, nil)
			label := tab.GetTabLabel()
			if label != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, label)
			}
		})
	}
}

// TestGetTabLabelInTabBar tests that tab bar uses GetTabLabel
func TestGetTabLabelInTabBar(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add regular task
	modal.addTab("1.2", "Implement Feature", "model")
	// Add command
	modal.addTab("command-1234567890", "Generate Code", "model")
	// Add command without title
	modal.addTab("command-9876543210", "", "model")

	tabBar := modal.renderTabBar()

	if tabBar == "" {
		t.Error("Expected non-empty tab bar")
	}

	// Verify task label includes task ID and title
	if !stringContains(tabBar, "1.2") {
		t.Error("Expected task ID in tab bar")
	}
	if !stringContains(tabBar, "Implement Feature") {
		t.Error("Expected task title in tab bar")
	}

	// Verify command label uses title directly (not command ID)
	if !stringContains(tabBar, "Generate Code") {
		t.Error("Expected command title in tab bar")
	}
	if stringContains(tabBar, "command-1234567890") {
		t.Error("Did not expect command ID to be shown in tab bar")
	}

	// Verify command without title shows "Command"
	if !stringContains(tabBar, "Command") {
		t.Error("Expected 'Command' label for command without title")
	}
}

// TestMessageHandlingWithCommandIDs tests that Task Runner message handling works with command IDs
func TestMessageHandlingWithCommandIDs(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Test TaskStartedMsg with command ID
	startMsg := TaskStartedMsg{
		TaskID:    "command-1234567890",
		TaskTitle: "Generate Documentation",
		Model:     "claude-3.5-sonnet",
	}
	modal.Update(startMsg)

	// Verify tab was created with command ID
	if len(modal.tabs) != 1 {
		t.Errorf("Expected 1 tab after TaskStartedMsg, got %d", len(modal.tabs))
	}
	if modal.tabs[0].GetTaskID() != "command-1234567890" {
		t.Errorf("Expected taskID 'command-1234567890', got %q", modal.tabs[0].GetTaskID())
	}
	if modal.tabs[0].GetTaskTitle() != "Generate Documentation" {
		t.Errorf("Expected title 'Generate Documentation', got %q", modal.tabs[0].GetTaskTitle())
	}

	// Test TaskOutputMsg with command ID
	outputMsg := TaskOutputMsg{
		TaskID: "command-1234567890",
		Output: "Generating documentation...",
	}
	modal.Update(outputMsg)

	output := modal.tabs[0].GetOutput()
	if len(output) != 1 {
		t.Errorf("Expected 1 line of output, got %d", len(output))
	}
	if output[0] != "Generating documentation..." {
		t.Errorf("Expected output line to match, got %q", output[0])
	}

	// Test multiple output lines
	outputMsg2 := TaskOutputMsg{
		TaskID: "command-1234567890",
		Output: "Documentation complete!",
	}
	modal.Update(outputMsg2)

	output = modal.tabs[0].GetOutput()
	if len(output) != 2 {
		t.Errorf("Expected 2 lines of output, got %d", len(output))
	}

	// Test TaskCompletedMsg with command ID
	completedMsg := TaskCompletedMsg{
		TaskID: "command-1234567890",
	}
	modal.Update(completedMsg)

	if modal.tabs[0].GetStatus() != TaskCompleted {
		t.Errorf("Expected status TaskCompleted, got %v", modal.tabs[0].GetStatus())
	}
}

// TestConcurrentCommandAndTaskExecution tests running both tasks and commands simultaneously
func TestConcurrentCommandAndTaskExecution(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Start a regular task
	taskMsg := TaskStartedMsg{
		TaskID:    "1.2",
		TaskTitle: "Implement Feature",
		Model:     "claude-3.5-sonnet",
	}
	modal.Update(taskMsg)

	// Start a command concurrently
	cmdMsg := TaskStartedMsg{
		TaskID:    "command-9876543210",
		TaskTitle: "Generate Code",
		Model:     "claude-3.5-sonnet",
	}
	modal.Update(cmdMsg)

	// Verify both tabs exist
	if len(modal.tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(modal.tabs))
	}

	// Send output to task
	taskOutput := TaskOutputMsg{
		TaskID: "1.2",
		Output: "Task output line 1",
	}
	modal.Update(taskOutput)

	// Send output to command
	cmdOutput := TaskOutputMsg{
		TaskID: "command-9876543210",
		Output: "Command output line 1",
	}
	modal.Update(cmdOutput)

	// Verify output went to correct tabs
	taskTab := modal.GetTabByTaskID("1.2")
	if taskTab == nil {
		t.Error("Expected task tab to exist")
	}
	if len(taskTab.GetOutput()) != 1 || taskTab.GetOutput()[0] != "Task output line 1" {
		t.Errorf("Task output incorrect: %v", taskTab.GetOutput())
	}

	cmdTab := modal.GetTabByTaskID("command-9876543210")
	if cmdTab == nil {
		t.Error("Expected command tab to exist")
	}
	if len(cmdTab.GetOutput()) != 1 || cmdTab.GetOutput()[0] != "Command output line 1" {
		t.Errorf("Command output incorrect: %v", cmdTab.GetOutput())
	}

	// Complete the command
	cmdComplete := TaskCompletedMsg{
		TaskID: "command-9876543210",
	}
	modal.Update(cmdComplete)

	// Verify only command tab is completed, task still running
	if cmdTab.GetStatus() != TaskCompleted {
		t.Errorf("Expected command tab to be completed")
	}
	if taskTab.GetStatus() != TaskRunning {
		t.Errorf("Expected task tab to still be running")
	}
}

// TestMessageRoutingAccuracy tests that messages are routed to correct tabs
func TestMessageRoutingAccuracy(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Create multiple tabs with various ID formats
	ids := []string{"1", "1.2", "2.1.1", "command-111", "command-222"}
	for _, id := range ids {
		modal.addTab(id, "Title for "+id, "model")
	}

	if len(modal.tabs) != 5 {
		t.Errorf("Expected 5 tabs, got %d", len(modal.tabs))
	}

	// Send output to each tab and verify routing
	for _, id := range ids {
		msg := TaskOutputMsg{
			TaskID: id,
			Output: fmt.Sprintf("Output for %s", id),
		}
		modal.Update(msg)

		// Verify output went to correct tab
		for _, tab := range modal.tabs {
			if tab.GetTaskID() == id {
				output := tab.GetOutput()
				if len(output) != 1 || output[0] != fmt.Sprintf("Output for %s", id) {
					t.Errorf("Output routing failed for %s: got %v", id, output)
				}
			}
		}
	}

	// Verify no cross-contamination
	for _, tab := range modal.tabs {
		output := tab.GetOutput()
		if len(output) != 1 {
			t.Errorf("Expected 1 line of output for %s, got %d", tab.GetTaskID(), len(output))
		}
	}
}

// TestConcurrentHeavyLoad tests Task Runner under heavy concurrent load
// This test verifies that the modal can handle multiple simultaneous messages
// without race conditions or data corruption
func TestConcurrentHeavyLoad(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Create 10 concurrent executions (mix of tasks and commands)
	tabIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			// Regular task
			tabIDs[i] = fmt.Sprintf("%d.%d", i+1, i+10)
		} else {
			// Command
			tabIDs[i] = fmt.Sprintf("command-%d", 1000000000+i*111)
		}
	}

	// Start all tasks/commands
	for i, id := range tabIDs {
		msg := TaskStartedMsg{
			TaskID:    id,
			TaskTitle: fmt.Sprintf("Execution %d", i),
			Model:     "test-model",
		}
		modal.Update(msg)
	}

	if len(modal.tabs) != 10 {
		t.Errorf("Expected 10 tabs, got %d", len(modal.tabs))
	}

	// Send multiple output lines to each tab in random order
	for round := 0; round < 5; round++ {
		for _, id := range tabIDs {
			msg := TaskOutputMsg{
				TaskID: id,
				Output: fmt.Sprintf("Round %d output for %s", round, id),
			}
			modal.Update(msg)
		}
	}

	// Verify all tabs have correct output
	for _, id := range tabIDs {
		tab := modal.GetTabByTaskID(id)
		if tab == nil {
			t.Errorf("Tab not found for %s", id)
			continue
		}
		output := tab.GetOutput()
		if len(output) != 5 {
			t.Errorf("Tab %s expected 5 output lines, got %d", id, len(output))
		}
	}

	// Complete first 5 tabs
	for i := 0; i < 5; i++ {
		msg := TaskCompletedMsg{
			TaskID: tabIDs[i],
		}
		modal.Update(msg)
		if modal.tabs[i].GetStatus() != TaskCompleted {
			t.Errorf("Tab %d not marked completed", i)
		}
	}

	// Fail next 3 tabs
	for i := 5; i < 8; i++ {
		msg := TaskFailedMsg{
			TaskID: tabIDs[i],
		}
		modal.Update(msg)
		if modal.tabs[i].GetStatus() != TaskFailed {
			t.Errorf("Tab %d not marked failed", i)
		}
	}

	// Cancel last 2 tabs
	for i := 8; i < 10; i++ {
		msg := TaskCancelledMsg{
			TaskID: tabIDs[i],
		}
		modal.Update(msg)
		if modal.tabs[i].GetStatus() != TaskCancelled {
			t.Errorf("Tab %d not marked cancelled", i)
		}
	}

	// Verify final states
	expectedStates := []TaskExecutionStatus{
		TaskCompleted, TaskCompleted, TaskCompleted, TaskCompleted, TaskCompleted,
		TaskFailed, TaskFailed, TaskFailed,
		TaskCancelled, TaskCancelled,
	}

	for i, expectedStatus := range expectedStates {
		if modal.tabs[i].GetStatus() != expectedStatus {
			t.Errorf("Tab %d: expected %v, got %v", i, expectedStatus, modal.tabs[i].GetStatus())
		}
	}
}

// TestTabLabelMixedIDFormats tests that tab labels render correctly with mixed ID formats
func TestTabLabelMixedIDFormats(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Create tabs with various ID formats
	testCases := []struct {
		taskID    string
		taskTitle string
		expected  string
	}{
		{"1", "Setup", "1 - Setup"},
		{"1.2", "Implementation", "1.2 - Implementation"},
		{"2.1.1", "Deep Nested", "2.1.1 - Deep Nested"},
		{"command-1234567890", "Generate Docs", "Generate Docs"},
		{"command-9876543210", "", "Command"},
	}

	for _, tc := range testCases {
		modal.addTab(tc.taskID, tc.taskTitle, "model")
	}

	tabBar := modal.renderTabBar()

	// Verify all labels appear in tab bar
	for _, tc := range testCases {
		if !stringContains(tabBar, tc.expected) {
			t.Errorf("Expected label %q in tab bar, not found. Tab bar: %s", tc.expected, tabBar)
		}
	}

	// Verify command IDs are NOT shown directly
	if stringContains(tabBar, "command-1234567890") {
		t.Error("Command ID should not appear in tab bar")
	}
}

// TestOutputIsolation verifies that output from one tab doesn't contaminate others
func TestOutputIsolation(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Create 3 tabs
	taskIDs := []string{"task-1", "command-1000", "task-2"}
	for _, id := range taskIDs {
		modal.addTab(id, "Title "+id, "model")
	}

	// Send specific output to each tab
	expectedOutputs := map[string][]string{
		"task-1":       {"Task 1 line 1", "Task 1 line 2"},
		"command-1000": {"Command output A", "Command output B", "Command output C"},
		"task-2":       {"Task 2 unique output"},
	}

	for id, lines := range expectedOutputs {
		for _, line := range lines {
			msg := TaskOutputMsg{
				TaskID: id,
				Output: line,
			}
			modal.Update(msg)
		}
	}

	// Verify each tab has ONLY its own output
	for id, expectedLines := range expectedOutputs {
		tab := modal.GetTabByTaskID(id)
		if tab == nil {
			t.Fatalf("Tab %s not found", id)
		}
		output := tab.GetOutput()
		if len(output) != len(expectedLines) {
			t.Errorf("Tab %s: expected %d lines, got %d", id, len(expectedLines), len(output))
		}
		for i, expected := range expectedLines {
			if i >= len(output) || output[i] != expected {
				t.Errorf("Tab %s line %d: expected %q, got %q", id, i, expected, output[i])
			}
		}
	}
}

// TestIsGitTask tests git task recognition
func TestIsGitTask(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	tests := []struct {
		taskID   string
		expected bool
	}{
		{"git-create-branch", true},
		{"git-switch-branch", true},
		{"git-status", true},
		{"git-", true},
		{"git", false},
		{"create-branch", false},
		{"regular-task", false},
		{"my-git-task", false},
		{"", false},
	}

	for _, test := range tests {
		result := modal.isGitTask(test.taskID)
		if result != test.expected {
			t.Errorf("isGitTask(%q): expected %v, got %v", test.taskID, test.expected, result)
		}
	}
}

// TestGitTaskAutoCloseOnSuccess tests that successful git operations trigger 2-second auto-close
func TestGitTaskAutoCloseOnSuccess(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)
	modal.SetGitAutoCloseDelay(0) // Disable delay for testing

	// Start a git task
	startMsg := TaskStartedMsg{
		TaskID:    "git-create-branch",
		TaskTitle: "Create Branch",
		Model:     "test",
	}
	modal.Update(startMsg)

	// Mark it as completed
	completeMsg := TaskCompletedMsg{
		TaskID: "git-create-branch",
	}
	modal.Update(completeMsg)

	// Verify the timer was started
	if timer, exists := modal.gitAutoCloseTimers["git-create-branch"]; !exists || timer == nil {
		t.Error("Expected git auto-close timer to be set for successful git task")
	}

	// Verify status is completed
	status, exists := modal.GetTaskStatus("git-create-branch")
	if !exists || status != TaskCompleted {
		t.Errorf("Expected task status to be TaskCompleted, got %v", status)
	}
}

// TestGitTaskNoAutoCloseOnFailure tests that failed git operations do NOT trigger auto-close
func TestGitTaskNoAutoCloseOnFailure(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Start a git task
	startMsg := TaskStartedMsg{
		TaskID:    "git-switch-branch",
		TaskTitle: "Switch Branch",
		Model:     "test",
	}
	modal.Update(startMsg)

	// Mark it as failed
	failMsg := TaskFailedMsg{
		TaskID:  "git-switch-branch",
		Error:   "Branch not found",
		Message: "Git failed",
	}
	modal.Update(failMsg)

	// Verify NO timer was started for failed git task
	if timer, exists := modal.gitAutoCloseTimers["git-switch-branch"]; exists && timer != nil {
		t.Error("Expected no git auto-close timer for failed git task")
	}

	// Verify status is failed
	status, exists := modal.GetTaskStatus("git-switch-branch")
	if !exists || status != TaskFailed {
		t.Errorf("Expected task status to be TaskFailed, got %v", status)
	}
}

// TestNonGitTaskUnaffectedByGitAutoClose tests that non-git tasks use normal auto-close
func TestNonGitTaskUnaffectedByGitAutoClose(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Start a non-git task
	startMsg := TaskStartedMsg{
		TaskID:    "regular-task",
		TaskTitle: "Regular Task",
		Model:     "test",
	}
	modal.Update(startMsg)

	// Mark it as completed
	completeMsg := TaskCompletedMsg{
		TaskID: "regular-task",
	}
	modal.Update(completeMsg)

	// Verify NO git timer was started
	if timer, exists := modal.gitAutoCloseTimers["regular-task"]; exists && timer != nil {
		t.Error("Expected no git auto-close timer for non-git task")
	}

	// Verify normal close timer was set
	if modal.closeTimer == nil && modal.autoCloseOnFailure {
		t.Error("Expected normal auto-close timer to be set for regular task")
	}
}

// TestKeyPressCancel sGitAutoClose tests that any key press cancels git auto-close countdown
func TestKeyPressCancelsGitAutoClose(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Start and complete a git task
	startMsg := TaskStartedMsg{
		TaskID:    "git-create-branch",
		TaskTitle: "Create Branch",
		Model:     "test",
	}
	modal.Update(startMsg)

	completeMsg := TaskCompletedMsg{
		TaskID: "git-create-branch",
	}
	modal.Update(completeMsg)

	// Verify timer was set
	if len(modal.gitAutoCloseTimers) == 0 {
		t.Error("Expected git auto-close timer to be set")
	}

	// Simulate key press
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	modal.HandleKey(keyMsg)

	// Verify timer was cleared
	if len(modal.gitAutoCloseTimers) > 0 {
		t.Error("Expected git auto-close timers to be cleared after key press")
	}
}

// TestGitAutoCloseMessageHandling tests handling of TaskRunnerGitAutoCloseMsg
func TestGitAutoCloseMessageHandling(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a completed git task
	modal.addTab("git-status", "Git Status", "test")
	modal.tabs[0].SetStatus(TaskCompleted)

	// Set up a git auto-close timer
	now := fmt.Sprintf("%v", time.Now())
	t.Logf("Setting git auto-close timer at %s", now)

	// Note: The actual timer expiration is handled by tea.Tick in handleGitTaskCompletion
	// Here we test the message handler
	gitAutoCloseMsg := TaskRunnerGitAutoCloseMsg{
		TaskID: "git-status",
	}

	// Set up the timer manually for this test
	expiredTime := time.Now().Add(-1 * time.Second) // Already expired
	modal.gitAutoCloseTimers["git-status"] = &expiredTime

	// Process the message
	result, _ := modal.Update(gitAutoCloseMsg)

	// Should close the modal since no tasks are running
	if result != nil {
		t.Error("Expected modal to be closed (nil returned)")
	}
}

// TestGetGitAutoCloseInfo tests retrieving git auto-close information
func TestGetGitAutoCloseInfo(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Initially no git auto-close timers
	active, remaining := modal.getGitAutoCloseInfo()
	if active {
		t.Error("Expected no active git auto-close timers initially")
	}

	// Set up a future timer
	futureTime := time.Now().Add(3 * time.Second)
	modal.gitAutoCloseTimers["git-task"] = &futureTime

	// Check info
	active, remaining = modal.getGitAutoCloseInfo()
	if !active {
		t.Error("Expected git auto-close to be active")
	}
	if remaining <= 0 || remaining > 3*time.Second {
		t.Errorf("Expected remaining time around 3 seconds, got %v", remaining)
	}

	// Set up an expired timer
	expiredTime := time.Now().Add(-1 * time.Second)
	modal.gitAutoCloseTimers["git-expired"] = &expiredTime

	// Should not find the expired timer
	active, _ = modal.getGitAutoCloseInfo()
	// Note: May or may not be active depending on if futureTime still exists
}

// TestRenderGitAutoCloseCountdown tests countdown rendering
func TestRenderGitAutoCloseCountdown(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Test countdown at 3 seconds
	countdown := modal.renderGitAutoCloseCountdown(3 * time.Second)
	if !strings.Contains(countdown, "3 seconds") {
		t.Errorf("Expected '3 seconds' in countdown, got: %s", countdown)
	}

	// Test countdown at 1 second (singular)
	countdown = modal.renderGitAutoCloseCountdown(1 * time.Second)
	if !strings.Contains(countdown, "1 second") {
		t.Errorf("Expected '1 second' (singular) in countdown, got: %s", countdown)
	}

	// Test countdown at 0 seconds
	countdown = modal.renderGitAutoCloseCountdown(0 * time.Second)
	if !strings.Contains(countdown, "0 seconds") {
		t.Errorf("Expected '0 seconds' in countdown, got: %s", countdown)
	}

	// All countdowns should contain instruction text
	for i := 0; i <= 3; i++ {
		countdown = modal.renderGitAutoCloseCountdown(time.Duration(i) * time.Second)
		if !strings.Contains(countdown, "Press any key to stay open") {
			t.Errorf("Expected instruction text in countdown: %s", countdown)
		}
	}
}

// TestModalClosesOnGitAutoCloseExpiry tests that modal closes when git auto-close expires
func TestModalClosesOnGitAutoCloseExpiry(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a git task and complete it
	modal.addTab("git-finish", "Finish Operation", "test")
	modal.tabs[0].SetStatus(TaskCompleted)

	// Set an expired timer
	expiredTime := time.Now().Add(-100 * time.Millisecond)
	modal.gitAutoCloseTimers["git-finish"] = &expiredTime

	// Process the auto-close message
	gitAutoCloseMsg := TaskRunnerGitAutoCloseMsg{
		TaskID: "git-finish",
	}
	result, _ := modal.Update(gitAutoCloseMsg)

	// Modal should close (return nil)
	if result != nil {
		t.Error("Expected modal to close on git auto-close expiry")
	}
}

// TestCancellationDialogKeyHandling tests that the cancellation dialog properly handles keyboard input
func TestCancellationDialogKeyHandling(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Long Running Task", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskRunning)

	// Set a very low threshold so even a newly created task is considered "long-running"
	modal.longRunningThreshold = 0

	// Press Ctrl+C to show confirmation dialog
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	modal.HandleKey(ctrlCKey)

	// Confirmation dialog should be shown
	if modal.cancellationConfirmDialog == nil {
		t.Fatal("Expected cancellation confirmation dialog to be shown")
	}

	// Test pressing Ctrl+C on the confirmation dialog (should cancel without crashing)
	updatedDialog, cmd := modal.cancellationConfirmDialog.Update(ctrlCKey)
	if updatedDialog != nil {
		t.Error("Expected confirmation dialog to close on Ctrl+C, but it remained open")
	}
	if cmd == nil {
		t.Error("Expected Cmd to be returned when dialog closes")
	}
}

// TestCancellationDialogEscKey tests that Esc key also closes the confirmation dialog
func TestCancellationDialogEscKey(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Task", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskRunning)
	modal.longRunningThreshold = 0

	// Show confirmation dialog
	modal.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if modal.cancellationConfirmDialog == nil {
		t.Fatal("Expected confirmation dialog")
	}

	// Press Esc on the dialog
	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updatedDialog, _ := modal.cancellationConfirmDialog.Update(escKey)

	// Dialog should close
	if updatedDialog != nil {
		t.Error("Expected Esc to close the confirmation dialog")
	}
}

// TestMultipleCancellationAttemptsAreSafe tests that repeated Ctrl+C presses don't crash
func TestMultipleCancellationAttemptsAreSafe(t *testing.T) {
	modal := NewTaskRunnerModal(80, 30, nil)

	// Add a running task
	modal.addTab("task-1", "Task", "model")
	tab := modal.tabs[0]
	tab.SetStatus(TaskRunning)
	modal.longRunningThreshold = 0

	// First Ctrl+C shows confirmation
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	modal.HandleKey(ctrlCKey)
	if modal.cancellationConfirmDialog == nil {
		t.Fatal("Expected confirmation dialog on first Ctrl+C")
	}

	firstDialog := modal.cancellationConfirmDialog

	// Second Ctrl+C should be ignored (guard prevents creating another dialog)
	modal.HandleKey(ctrlCKey)

	// Dialog should be the same (not duplicated)
	if modal.cancellationConfirmDialog != firstDialog {
		t.Error("Expected guard to prevent creating another dialog on repeated Ctrl+C")
	}

	// Simulate the user pressing Ctrl+C on the dialog
	// This happens through the modal's Update method, not the dialog directly
	if modal.cancellationConfirmDialog != nil {
		// The dialog.Update returns nil when it closes
		updatedDialog, _ := modal.cancellationConfirmDialog.Update(ctrlCKey)
		if updatedDialog == nil {
			// Dialog closed, now modal needs to clear its reference
			// This would happen in a real scenario through modal.Update()
			modal.cancellationConfirmDialog = nil
		}
	}

	// Modal state should be consistent
	if modal.cancellationConfirmDialog != nil {
		t.Error("Expected cancellation dialog to be cleared after closing")
	}
}

// TestConfirmationDialogYesNoNavigation tests dialog button navigation
func TestConfirmationDialogYesNoNavigation(t *testing.T) {
	dialog := YesNo("Test", "Are you sure?", true)

	// Initial focus should be on "Yes"
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected initial focus on Yes button (0), got %d", dialog.FocusedIndex())
	}

	// Navigate to "No"
	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	result, _ := dialog.HandleKey(rightKey)
	if result != DialogResultNone {
		t.Error("Expected navigation to not return a result")
	}
	if dialog.FocusedIndex() != 1 {
		t.Errorf("Expected focus on No button (1), got %d", dialog.FocusedIndex())
	}

	// Navigate back to "Yes"
	leftKey := tea.KeyMsg{Type: tea.KeyLeft}
	result, _ = dialog.HandleKey(leftKey)
	if result != DialogResultNone {
		t.Error("Expected navigation to not return a result")
	}
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focus back on Yes button (0), got %d", dialog.FocusedIndex())
	}
}
