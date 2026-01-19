package ui

import (
	"fmt"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

// TestValidateTaskSelection tests the validateTaskSelection method
func TestValidateTaskSelection_Valid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	// Valid selection with 5 tasks
	taskIDs := []string{"1", "2", "3", "4", "5"}
	errMsg := m.validateTaskSelection(taskIDs)

	if errMsg != "" {
		t.Errorf("Expected no error for valid task selection, got: %s", errMsg)
	}
}

// TestValidateTaskSelection_Empty tests empty task selection
func TestValidateTaskSelection_Empty(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	taskIDs := []string{}
	errMsg := m.validateTaskSelection(taskIDs)

	if errMsg == "" {
		t.Error("Expected error for empty task selection")
	}
	if errMsg != "No tasks selected. Please select at least one task." {
		t.Errorf("Unexpected error message: %s", errMsg)
	}
}

// TestValidateTaskSelection_MaxTasks tests max task limit
func TestValidateTaskSelection_MaxTasks(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	// Create 10 task IDs (exceeds limit of 9)
	taskIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		taskIDs[i] = fmt.Sprintf("%d", i+1)
	}

	errMsg := m.validateTaskSelection(taskIDs)

	if errMsg == "" {
		t.Error("Expected error for too many tasks")
	}
	if errMsg != fmt.Sprintf("Too many tasks selected (%d). Maximum concurrent tasks is %d. Please select fewer tasks.", 10, maxConcurrentTasks) {
		t.Errorf("Unexpected error message: %s", errMsg)
	}
}

// TestValidateTaskSelection_AtLimit tests exact max limit
func TestValidateTaskSelection_AtLimit(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	// Create exactly 9 task IDs (should be valid)
	taskIDs := make([]string, 9)
	for i := 0; i < 9; i++ {
		taskIDs[i] = fmt.Sprintf("%d", i+1)
	}

	errMsg := m.validateTaskSelection(taskIDs)

	if errMsg != "" {
		t.Errorf("Expected no error for exactly %d tasks, got: %s", maxConcurrentTasks, errMsg)
	}
}

// TestValidateModelSelection_Valid tests valid model selection
func TestValidateModelSelection_Valid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	errMsg := m.validateModelSelection("claude-3-5-sonnet-20241022")

	if errMsg != "" {
		t.Errorf("Expected no error for valid model selection, got: %s", errMsg)
	}
}

// TestValidateModelSelection_Empty tests empty model selection
func TestValidateModelSelection_Empty(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	errMsg := m.validateModelSelection("")

	if errMsg == "" {
		t.Error("Expected error for empty model selection")
	}
	if errMsg != "No model selected. Please select a model before executing tasks." {
		t.Errorf("Unexpected error message: %s", errMsg)
	}
}

// TestValidateTaskDependencies tests dependency validation
func TestValidateTaskDependencies_Valid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	// Create mock tasks with proper dependencies
	taskSvc := &taskmaster.Service{}
	m.taskService = taskSvc

	// Simple case: no dependencies
	taskIDs := []string{"1", "2", "3"}
	errMsg := m.validateTaskDependencies(taskIDs)

	if errMsg != "" {
		t.Errorf("Expected no error for valid dependencies, got: %s", errMsg)
	}
}

// TestValidateTaskDependencies_NilService tests when service is nil
func TestValidateTaskDependencies_NilService(t *testing.T) {
	m := &Model{
		config:      &config.Config{},
		taskService: nil,
	}

	taskIDs := []string{"1", "2", "3"}
	errMsg := m.validateTaskDependencies(taskIDs)

	// Should not error when service is nil
	if errMsg != "" {
		t.Errorf("Expected no error when service is nil, got: %s", errMsg)
	}
}

// TestValidateReadyTasksExecution tests comprehensive ready tasks validation
func TestValidateReadyTasksExecution_Valid(t *testing.T) {
	m := &Model{
		config:      &config.Config{},
		taskService: &taskmaster.Service{},
	}

	// Valid selection
	taskIDs := []string{"1", "2", "3"}
	result := m.validateReadyTasksExecution(taskIDs)

	if !result {
		t.Error("Expected validation to pass for valid tasks")
	}
}

// TestValidateReadyTasksExecution_TooManyTasks tests max task limit in ready tasks validation
func TestValidateReadyTasksExecution_TooManyTasks(t *testing.T) {
	m := &Model{
		config:      &config.Config{},
		taskService: &taskmaster.Service{},
	}

	// Create 10 task IDs
	taskIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		taskIDs[i] = fmt.Sprintf("%d", i+1)
	}

	result := m.validateReadyTasksExecution(taskIDs)

	if result {
		t.Error("Expected validation to fail for > 9 tasks")
	}
}

// TestValidateModelExecution_Valid tests model execution validation
func TestValidateModelExecution_Valid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	result := m.validateModelExecution("claude-3-5-sonnet-20241022")

	if !result {
		t.Error("Expected model validation to pass for valid model")
	}
}

// TestValidateModelExecution_Invalid tests model execution validation fails
func TestValidateModelExecution_Invalid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
	}

	result := m.validateModelExecution("")

	if result {
		t.Error("Expected model validation to fail for empty model")
	}
}

// TestMaxConcurrentTasksConstant verifies max concurrent tasks is 9
func TestMaxConcurrentTasksConstant(t *testing.T) {
	if maxConcurrentTasks != 9 {
		t.Errorf("Expected maxConcurrentTasks to be 9, got %d", maxConcurrentTasks)
	}
}

// TestHandleTaskModelSelectionDialog_NilQueue tests nil execution queue handling
func TestHandleTaskModelSelectionDialog_NilQueue(t *testing.T) {
	m := &Model{
		config:         &config.Config{},
		executionQueue: nil,
		appState:       newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "confirm",
		Value:  "claude-3-5-sonnet-20241022",
	}

	cmd := m.handleDialogResultMsg(msg)

	// Should return nil and close dialog
	if cmd != nil {
		t.Error("Expected nil command for nil execution queue")
	}
}

// TestHandleTaskModelSelectionDialog_ConfirmValid tests confirm button with valid data
func TestHandleTaskModelSelectionDialog_ConfirmValid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{"1", "2", "3"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		},
		taskModelSelectionDone: make(map[string]bool),
		appState:               newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "confirm",
		Value:  "claude-3-5-sonnet-20241022",
	}

	_ = m.handleDialogResultMsg(msg)

	// Verify model selection stored
	if m.executionQueue.GetModelSelection("1") != "claude-3-5-sonnet-20241022" {
		t.Error("Model selection not stored correctly")
	}

	// Verify task marked as having selection
	if !m.taskModelSelectionDone["1"] {
		t.Error("Task not marked in taskModelSelectionDone map")
	}

	// Verify task status updated
	if m.executionQueue.TaskStatus["1"] != "ready" {
		t.Errorf("Task status not updated to ready, got: %s", m.executionQueue.TaskStatus["1"])
	}

	// Verify queue advanced
	if m.executionQueue.CurrentIndex != 1 {
		t.Errorf("Queue not advanced, CurrentIndex: %d", m.executionQueue.CurrentIndex)
	}
}

// TestHandleTaskModelSelectionDialog_ConfirmInvalidType tests confirm with invalid modelID type
func TestHandleTaskModelSelectionDialog_ConfirmInvalidType(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{"1"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		},
		appState: newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "confirm",
		Value:  123, // Invalid type (should be string)
	}

	cmd := m.handleDialogResultMsg(msg)

	// Should return nil and not store selection
	if cmd != nil {
		t.Error("Expected nil command for invalid model type")
	}

	if m.executionQueue.GetModelSelection("1") != "" {
		t.Error("Model selection should not be stored for invalid type")
	}
}

// TestHandleTaskModelSelectionDialog_ConfirmEmptyTaskID tests confirm with empty task ID
func TestHandleTaskModelSelectionDialog_ConfirmEmptyTaskID(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		},
		appState: newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "confirm",
		Value:  "claude-3-5-sonnet-20241022",
	}

	cmd := m.handleDialogResultMsg(msg)

	// Should return nil
	if cmd != nil {
		t.Error("Expected nil command for empty task ID")
	}
}

// TestHandleTaskModelSelectionDialog_SkipValid tests skip button with valid data
func TestHandleTaskModelSelectionDialog_SkipValid(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{"1", "2", "3"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		},
		appState: newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "skip",
	}

	_ = m.handleDialogResultMsg(msg)

	// Verify task removed from queue
	if len(m.executionQueue.TaskIDs) != 2 {
		t.Errorf("Task not removed from queue, length: %d", len(m.executionQueue.TaskIDs))
	}

	// Verify task not in queue anymore
	for _, id := range m.executionQueue.TaskIDs {
		if id == "1" {
			t.Error("Skipped task still in queue")
		}
	}
}

// TestHandleTaskModelSelectionDialog_SkipEmptyQueue tests skip with empty queue
func TestHandleTaskModelSelectionDialog_SkipEmptyQueue(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		},
		appState: newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "skip",
	}

	cmd := m.handleDialogResultMsg(msg)

	// Should return nil
	if cmd != nil {
		t.Error("Expected nil command for empty queue skip")
	}
}

// TestHandleTaskModelSelectionDialog_Cancel tests cancel button
func TestHandleTaskModelSelectionDialog_Cancel(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{"1", "2", "3"},
			CurrentIndex:    1,
			ModelSelections: map[string]string{"1": "model1"},
			TaskStatus:      map[string]string{"1": "ready"},
		},
		taskModelSelectionDone: map[string]bool{"1": true},
		appState:               newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "cancel",
	}

	cmd := m.handleDialogResultMsg(msg)

	// Verify queue is completely cleaned up (set to nil)
	if m.executionQueue != nil {
		t.Error("executionQueue should be nil after cancel")
	}

	// Verify taskModelSelectionDone reset
	if len(m.taskModelSelectionDone) != 0 {
		t.Error("taskModelSelectionDone not reset")
	}

	// Should return nil
	if cmd != nil {
		t.Error("Expected nil command for cancel")
	}
}

// TestHandleTaskModelSelectionDialog_UnknownButton tests unknown button value
func TestHandleTaskModelSelectionDialog_UnknownButton(t *testing.T) {
	m := &Model{
		config: &config.Config{},
		executionQueue: &ExecutionQueue{
			TaskIDs:         []string{"1"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		},
		taskModelSelectionDone: make(map[string]bool),
		appState:               newTestAppState(),
	}

	msg := dialog.DialogResultMsg{
		ID:     "task_model_selection_dialog",
		Button: "unknown_button",
	}

	cmd := m.handleDialogResultMsg(msg)

	// Should return nil
	if cmd != nil {
		t.Error("Expected nil command for unknown button")
	}

	// Verify state is cleaned up (queue set to nil) for unknown button
	if m.executionQueue != nil {
		t.Error("executionQueue should be nil after unknown button (cleanup triggered)")
	}

	// Verify taskModelSelectionDone is cleaned up
	if len(m.taskModelSelectionDone) != 0 {
		t.Error("taskModelSelectionDone not reset")
	}
}

// Helper function to create test app state
func newTestAppState() *AppState {
	keyMap := KeyMap{}
	return NewAppState(nil, &keyMap)
}
