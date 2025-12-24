package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
	"github.com/stretchr/testify/assert"
)

// TestNextTaskModalStateInitialization verifies initial state is correct
func TestNextTaskModalStateInitialization(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)

	assert.False(t, state.nextTaskModalActive, "nextTaskModalActive should be false initially")
	assert.Nil(t, state.nextTaskOutput, "nextTaskOutput should be nil initially")
}

// TestStartNextTaskModal verifies modal activation and output initialization
func TestStartNextTaskModal(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)

	state.StartNextTaskModal()

	assert.True(t, state.nextTaskModalActive, "nextTaskModalActive should be true after StartNextTaskModal")
	assert.NotNil(t, state.nextTaskOutput, "nextTaskOutput should not be nil after StartNextTaskModal")
	assert.Equal(t, 0, len(state.nextTaskOutput), "nextTaskOutput should be empty slice after StartNextTaskModal")
}

// TestAppendNextTaskOutputWhenActive verifies lines are appended when modal is active
func TestAppendNextTaskOutputWhenActive(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)
	state.StartNextTaskModal()

	state.AppendNextTaskOutput("line 1")
	state.AppendNextTaskOutput("line 2")
	state.AppendNextTaskOutput("line 3")

	assert.Equal(t, 3, len(state.nextTaskOutput), "expected 3 lines in output")
	assert.Equal(t, "line 1", state.nextTaskOutput[0], "expected first line")
	assert.Equal(t, "line 2", state.nextTaskOutput[1], "expected second line")
	assert.Equal(t, "line 3", state.nextTaskOutput[2], "expected third line")
}

// TestAppendNextTaskOutputWhenInactive verifies lines are NOT appended when modal is inactive
func TestAppendNextTaskOutputWhenInactive(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)

	state.AppendNextTaskOutput("line 1")
	state.AppendNextTaskOutput("line 2")

	assert.Nil(t, state.nextTaskOutput, "nextTaskOutput should remain nil when modal is inactive")
}

// TestAppendNextTaskOutputAfterClose verifies lines are not appended after close
func TestAppendNextTaskOutputAfterClose(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)
	state.StartNextTaskModal()
	state.AppendNextTaskOutput("line 1")

	state.CloseNextTaskModal()
	state.AppendNextTaskOutput("line 2")

	assert.Nil(t, state.nextTaskOutput, "nextTaskOutput should be nil after CloseNextTaskModal")
}

// TestCloseNextTaskModal verifies state is properly reset
func TestCloseNextTaskModal(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)
	state.StartNextTaskModal()
	state.AppendNextTaskOutput("line 1")
	state.AppendNextTaskOutput("line 2")

	state.CloseNextTaskModal()

	assert.False(t, state.nextTaskModalActive, "nextTaskModalActive should be false after CloseNextTaskModal")
	assert.Nil(t, state.nextTaskOutput, "nextTaskOutput should be nil after CloseNextTaskModal")
}

// TestMultipleStartClose verifies state can transition correctly multiple times
func TestMultipleStartClose(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)

	// First cycle
	state.StartNextTaskModal()
	state.AppendNextTaskOutput("cycle 1 - line 1")
	assert.Equal(t, 1, len(state.nextTaskOutput), "expected 1 line in first cycle")

	state.CloseNextTaskModal()
	assert.Nil(t, state.nextTaskOutput, "expected nil after close")

	// Second cycle
	state.StartNextTaskModal()
	state.AppendNextTaskOutput("cycle 2 - line 1")
	state.AppendNextTaskOutput("cycle 2 - line 2")
	assert.Equal(t, 2, len(state.nextTaskOutput), "expected 2 lines in second cycle")

	assert.Equal(t, "cycle 2 - line 1", state.nextTaskOutput[0], "expected fresh output in second cycle")
	assert.Equal(t, "cycle 2 - line 2", state.nextTaskOutput[1], "expected fresh output in second cycle")
}

// TestAppendAfterStartResets verifies output is cleared when starting a new modal
func TestAppendAfterStartResets(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)

	state.StartNextTaskModal()
	state.AppendNextTaskOutput("old line")

	// Start again without closing
	state.StartNextTaskModal()
	state.AppendNextTaskOutput("new line")

	assert.Equal(t, 1, len(state.nextTaskOutput), "expected output to be reset")
	assert.Equal(t, "new line", state.nextTaskOutput[0], "expected only new line in output")
}

// TestStatePreservesOtherAppStateFields verifies modal state doesn't affect other fields
func TestStatePreservesOtherAppStateFields(t *testing.T) {
	manager := &dialog.DialogManager{}
	km := DefaultKeyMap()
	state := NewAppState(manager, &km)

	// Verify other fields are unchanged by modal operations
	state.StartNextTaskModal()
	assert.Equal(t, manager, state.DialogManager(), "DialogManager should be unchanged")
	assert.Equal(t, &km, state.KeyMap(), "KeyMap should be unchanged")

	state.AppendNextTaskOutput("test")
	assert.Equal(t, manager, state.DialogManager(), "DialogManager should be unchanged after append")
	assert.Equal(t, &km, state.KeyMap(), "KeyMap should be unchanged after append")

	state.CloseNextTaskModal()
	assert.Equal(t, manager, state.DialogManager(), "DialogManager should be unchanged after close")
	assert.Equal(t, &km, state.KeyMap(), "KeyMap should be unchanged after close")
}

// TestEmptyLineAppend verifies empty strings can be appended
func TestEmptyLineAppend(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)
	state.StartNextTaskModal()

	state.AppendNextTaskOutput("")
	state.AppendNextTaskOutput("line")
	state.AppendNextTaskOutput("")

	assert.Equal(t, 3, len(state.nextTaskOutput), "expected 3 lines including empty ones")
	assert.Equal(t, "", state.nextTaskOutput[0], "expected empty string at index 0")
	assert.Equal(t, "line", state.nextTaskOutput[1], "expected line at index 1")
	assert.Equal(t, "", state.nextTaskOutput[2], "expected empty string at index 2")
}

// TestLargeOutputAppend verifies many lines can be appended
func TestLargeOutputAppend(t *testing.T) {
	km := DefaultKeyMap()
	state := NewAppState(nil, &km)
	state.StartNextTaskModal()

	for i := 0; i < 1000; i++ {
		state.AppendNextTaskOutput("line")
	}

	assert.Equal(t, 1000, len(state.nextTaskOutput), "expected 1000 lines appended")
}

// TestModalStateInheritance verifies state transfers correctly with new dialogs
func TestModalStateInheritance(t *testing.T) {
	manager1 := &dialog.DialogManager{}
	km := DefaultKeyMap()
	state := NewAppState(manager1, &km)

	state.StartNextTaskModal()
	state.AppendNextTaskOutput("test output")

	// Replace dialog manager (simulating new dialog context)
	manager2 := &dialog.DialogManager{}
	state.dialogManager = manager2

	// Modal state should be preserved
	assert.True(t, state.nextTaskModalActive, "modal state should be preserved")
	assert.Equal(t, 1, len(state.nextTaskOutput), "output should be preserved")
	assert.Equal(t, "test output", state.nextTaskOutput[0], "output content should be preserved")
}
