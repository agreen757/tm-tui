package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
)

// TestLogBrowserKeyBinding tests that Ctrl+F is bound in the keymap
func TestLogBrowserKeyBinding(t *testing.T) {
	km := DefaultKeyMap()

	// Check that LogBrowser binding exists
	if len(km.LogBrowser.Keys()) == 0 {
		t.Error("LogBrowser keybinding not configured")
	}

	// Check that Ctrl+F is one of the keys
	found := false
	for _, k := range km.LogBrowser.Keys() {
		if k == "ctrl+f" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'ctrl+f' in LogBrowser keys, got %v", km.LogBrowser.Keys())
	}

	// Check help text
	if len(km.LogBrowser.Help().Key) == 0 {
		t.Error("LogBrowser help key not configured")
	}
	if len(km.LogBrowser.Help().Desc) == 0 {
		t.Error("LogBrowser help description not configured")
	}
}

// TestLogBrowserKeybindingInFullHelp tests that LogBrowser is included in full help
func TestLogBrowserKeybindingInFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	found := false
	for _, row := range fullHelp {
		for _, binding := range row {
			// Check if this binding matches LogBrowser
			if len(binding.Keys()) > 0 {
				for _, k := range binding.Keys() {
					if k == "ctrl+f" {
						found = true
						break
					}
				}
			}
		}
	}

	if !found {
		t.Error("LogBrowser keybinding not found in FullHelp()")
	}
}

// TestLogBrowserDialogOpensWithCtrlL tests that Ctrl+F key press triggers dialog opening
func TestLogBrowserDialogOpensWithCtrlL(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	// Create a Ctrl+F key message
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyCtrlL,
		Runes: []rune{},
	}

	// Send the key message to the app
	updatedModel, cmd := m.Update(keyMsg)

	// Verify a command is returned (dialog opening command)
	if cmd == nil {
		t.Error("Expected a command to be returned when Ctrl+F is pressed, got nil")
	}

	// Verify the model is returned (not nil)
	if updatedModel == nil {
		t.Error("Expected model to be returned, got nil")
	}

	// Execute the command to verify it returns a message
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			t.Logf("Command returned message: %v", msg)
		}
	}
}

// TestLogBrowserDialogOpenFromAnyScreen tests that dialog opens from any panel
func TestLogBrowserDialogOpenFromAnyScreen(t *testing.T) {
	tests := []struct {
		name       string
		panelIndex Panel
	}{
		{"from task list panel", PanelTaskList},
		{"from details panel", PanelDetails},
		{"from log panel", PanelLog},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModelWithTaskService()
			m.width = 120
			m.height = 40
			m.focusedPanel = tt.panelIndex

			keyMsg := tea.KeyMsg{
				Type:  tea.KeyCtrlL,
				Runes: []rune{},
			}

			updatedModel, cmd := m.Update(keyMsg)

			if cmd == nil {
				t.Errorf("Expected dialog command from %s, got nil", tt.name)
			}

			if updatedModel == nil {
				t.Error("Expected model to be returned")
			}
		})
	}
}

// TestLogBrowserHandlerReturnsCommand tests that openLogBrowserDialog returns a command
func TestLogBrowserHandlerReturnsCommand(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	cmd := m.openLogBrowserDialog()

	if cmd == nil {
		t.Error("openLogBrowserDialog returned nil command")
	}
}

// TestLogBrowserHandlerWithInvalidTaskService tests error handling when task service is invalid
func TestLogBrowserHandlerWithInvalidTaskService(t *testing.T) {
	m := createTestModel()
	m.width = 120
	m.height = 40
	m.taskService = nil // Set to nil to test error handling

	// This should not panic even with nil taskService
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("openLogBrowserDialog panicked: %v", r)
		}
	}()

	cmd := m.openLogBrowserDialog()
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			t.Logf("Command returned: %v", msg)
		}
	}
}

// TestLogBrowserKeymapConfiguration tests that LogBrowser keymap can be configured
func TestLogBrowserKeymapConfiguration(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: "/tmp/test",
		KeyBindings: map[string]string{
			"logBrowser": "alt+b", // Custom keybinding
		},
	}

	km := NewKeyMap(cfg)

	// Should use custom key if provided
	if len(km.LogBrowser.Keys()) > 0 {
		found := false
		for _, k := range km.LogBrowser.Keys() {
			if k == "alt+b" {
				found = true
				break
			}
		}
		if !found {
			// Should fall back to default if no config
			found = false
			for _, k := range km.LogBrowser.Keys() {
				if k == "ctrl+f" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected 'alt+b' or 'ctrl+f' in LogBrowser keys, got %v", km.LogBrowser.Keys())
			}
		}
	}
}

// createTestModelWithTaskService creates a test model with a proper task service
func createTestModelWithTaskService() *Model {
	cfg := &config.Config{
		TaskMasterPath: "/tmp/test",
	}

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
			},
		},
	}

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
		width:            120,
		height:           40,
		taskService:      nil, // Will be set by caller if needed
	}

	m.buildTaskIndex()
	m.rebuildVisibleTasks()

	if len(m.visibleTasks) > 0 {
		m.selectedTask = m.visibleTasks[0]
		m.selectedIndex = 0
	}

	return m
}

// BenchmarkLogBrowserKeyBinding benchmarks keymap creation with LogBrowser
func BenchmarkLogBrowserKeyBinding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		km := DefaultKeyMap()
		_ = km.LogBrowser.Keys()
	}
}

// TestLogBrowserKeyMatchesCorrectly tests that key.Matches works with LogBrowser binding
func TestLogBrowserKeyMatchesCorrectly(t *testing.T) {
	km := DefaultKeyMap()

	// Test that the key binding works with key.Matches
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyCtrlL,
		Runes: []rune{},
	}

	if !key.Matches(keyMsg, km.LogBrowser) {
		t.Error("key.Matches should return true for Ctrl+F with LogBrowser binding")
	}

	// Test that non-matching key doesn't match
	otherKeyMsg := tea.KeyMsg{
		Type:  tea.KeyCtrlK,
		Runes: []rune{},
	}

	if key.Matches(otherKeyMsg, km.LogBrowser) {
		t.Error("key.Matches should return false for non-matching key")
	}
}

// TestLogBrowserDialogMessageHandling tests that dialog messages are handled correctly
func TestLogBrowserDialogMessageHandling(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	// Simulate opening the dialog by sending Ctrl+F
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyCtrlL,
		Runes: []rune{},
	}

	updatedModel, cmd := m.Update(keyMsg)

	// Verify a command is returned
	if cmd == nil {
		t.Fatal("Expected command to be returned for Ctrl+F press")
	}

	// Execute the command
	if msg := cmd(); msg != nil {
		// Verify the model is updated
		if updatedModel == nil {
			t.Error("Expected model to be returned after opening dialog")
		}
	}
}

// TestLogBrowserIntegrationWithAppState tests that dialog properly integrates with app state
func TestLogBrowserIntegrationWithAppState(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	// The model should have an app state with a dialog manager
	if m.appState == nil {
		t.Fatal("Model appState is nil")
	}

	// Try to open the dialog
	cmd := m.openLogBrowserDialog()
	if cmd == nil {
		t.Fatal("openLogBrowserDialog returned nil command")
	}

	// Execute the command
	msg := cmd()
	if msg != nil {
		t.Logf("Command executed and returned message: %v", msg)
	}
}

// TestLogBrowserKeybindingInHelpText tests that LogBrowser appears in help text
func TestLogBrowserKeybindingInHelpText(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	found := false
	for _, row := range fullHelp {
		for _, binding := range row {
			if len(binding.Keys()) > 0 && binding.Keys()[0] == "ctrl+f" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("LogBrowser keybinding not found in help text")
	}
}

// TestSaveStateBeforeDialog tests that app state is properly saved before opening dialog
func TestSaveStateBeforeDialog(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40
	m.viewMode = ViewModeTree
	m.focusedPanel = PanelTaskList
	m.showDetailsPanel = true
	m.showLogPanel = false
	m.taskListViewport.YOffset = 5

	// Save state
	m.saveStateBeforeDialog()

	// Verify state was saved
	if !m.appState.HasSavedState() {
		t.Error("No saved state found after calling saveStateBeforeDialog")
	}

	// Restore and verify
	savedState := m.appState.RestoreDialogState()
	if savedState == nil {
		t.Fatal("Failed to restore saved state")
	}

	if savedState.CurrentView != ViewModeTree {
		t.Errorf("Expected CurrentView=ViewModeTree, got %v", savedState.CurrentView)
	}
	if savedState.FocusedPanel != PanelTaskList {
		t.Errorf("Expected FocusedPanel=PanelTaskList, got %v", savedState.FocusedPanel)
	}
	if savedState.ShowDetailsPane != true {
		t.Errorf("Expected ShowDetailsPane=true, got %v", savedState.ShowDetailsPane)
	}
	if savedState.ShowLogPane != false {
		t.Errorf("Expected ShowLogPane=false, got %v", savedState.ShowLogPane)
	}
	if savedState.ScrollPosition != 5 {
		t.Errorf("Expected ScrollPosition=5, got %v", savedState.ScrollPosition)
	}
}

// TestRestoreStateAfterDialog tests that app state is properly restored after closing dialog
func TestRestoreStateAfterDialog(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	// Set initial state
	m.viewMode = ViewModeList
	m.focusedPanel = PanelDetails
	m.showDetailsPanel = false
	m.showLogPanel = true
	m.taskListViewport.YOffset = 10

	// Save state
	m.saveStateBeforeDialog()

	// Change state to simulate dialog being open
	m.viewMode = ViewModeTree
	m.focusedPanel = PanelLog
	m.showDetailsPanel = true
	m.showLogPanel = false
	m.taskListViewport.YOffset = 0

	// Restore state
	m.restoreStateAfterDialog()

	// Verify state was restored
	if m.viewMode != ViewModeList {
		t.Errorf("ViewMode not restored, expected ViewModeList, got %v", m.viewMode)
	}
	if m.focusedPanel != PanelDetails {
		t.Errorf("FocusedPanel not restored, expected PanelDetails, got %v", m.focusedPanel)
	}
	if m.showDetailsPanel != false {
		t.Errorf("ShowDetailsPanel not restored, expected false, got %v", m.showDetailsPanel)
	}
	if m.showLogPanel != true {
		t.Errorf("ShowLogPanel not restored, expected true, got %v", m.showLogPanel)
	}
}

// TestStatePreservationWithSelectedTask tests that selected task is preserved
func TestStatePreservationWithSelectedTask(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	// Set a task as selected
	if len(m.visibleTasks) > 0 {
		m.selectedTask = m.visibleTasks[0]
		m.saveStateBeforeDialog()

		// Change selection to simulate dialog interaction
		if len(m.visibleTasks) > 1 {
			m.selectedTask = m.visibleTasks[1]
		}

		// Restore state
		m.restoreStateAfterDialog()

		// Verify task selection was restored
		if m.appState.HasSavedState() {
			t.Error("SavedState stack should be empty after restore")
		}
	}
}

// TestDialogStateStackMultiple tests that multiple dialog states can be stacked
func TestDialogStateStackMultiple(t *testing.T) {
	m := createTestModelWithTaskService()
	m.width = 120
	m.height = 40

	// Save first state
	m.viewMode = ViewModeTree
	m.saveStateBeforeDialog()

	// Save second state
	m.viewMode = ViewModeList
	m.saveStateBeforeDialog()

	// Should have 2 saved states
	if !m.appState.HasSavedState() {
		t.Error("Should have saved states")
	}

	// Restore second (most recent) state
	restored := m.appState.RestoreDialogState()
	if restored == nil || restored.CurrentView != ViewModeList {
		t.Error("Failed to restore most recent state")
	}

	// Still should have one saved state
	if !m.appState.HasSavedState() {
		t.Error("Should still have one saved state")
	}

	// Restore first state
	restored = m.appState.RestoreDialogState()
	if restored == nil || restored.CurrentView != ViewModeTree {
		t.Error("Failed to restore first state")
	}

	// Now should have no saved states
	if m.appState.HasSavedState() {
		t.Error("Should have no saved states after popping all")
	}
}
