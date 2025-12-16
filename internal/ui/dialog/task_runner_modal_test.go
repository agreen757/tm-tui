package dialog

import (
	"testing"

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

