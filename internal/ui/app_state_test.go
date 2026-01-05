package ui

import (
	"testing"
)

func TestAppStateInitPrdCreationState(t *testing.T) {
	appState := &AppState{}

	// Initially nil
	if appState.PrdCreationState != nil {
		t.Error("PrdCreationState should be nil initially")
	}

	// After init, should be created
	appState.InitPrdCreationState()
	if appState.PrdCreationState == nil {
		t.Error("PrdCreationState should be created after Init")
	}

	// Should have empty OutputBuffer
	if appState.PrdCreationState.OutputBuffer == nil {
		t.Error("OutputBuffer should be initialized")
	}

	// Calling init again should not overwrite existing state
	appState.PrdCreationState.Title = "Test"
	appState.InitPrdCreationState()
	if appState.PrdCreationState.Title != "Test" {
		t.Error("Init should not overwrite existing state")
	}
}

func TestAppStateUpdatePrdCreationInputs(t *testing.T) {
	appState := &AppState{}

	appState.UpdatePrdCreationInputs("Test Title", "Test Summary", "Test Scope", "test.md")

	if appState.PrdCreationState == nil {
		t.Error("PrdCreationState should be created by UpdatePrdCreationInputs")
	}

	if appState.PrdCreationState.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", appState.PrdCreationState.Title)
	}
	if appState.PrdCreationState.Summary != "Test Summary" {
		t.Errorf("Expected summary 'Test Summary', got '%s'", appState.PrdCreationState.Summary)
	}
	if appState.PrdCreationState.Scope != "Test Scope" {
		t.Errorf("Expected scope 'Test Scope', got '%s'", appState.PrdCreationState.Scope)
	}
	if appState.PrdCreationState.Filename != "test.md" {
		t.Errorf("Expected filename 'test.md', got '%s'", appState.PrdCreationState.Filename)
	}
}

func TestAppStateClearPrdCreationState(t *testing.T) {
	appState := &AppState{}
	appState.InitPrdCreationState()

	if appState.PrdCreationState == nil {
		t.Error("PrdCreationState should exist before clearing")
	}

	appState.ClearPrdCreationState()

	if appState.PrdCreationState != nil {
		t.Error("PrdCreationState should be nil after clearing")
	}
}

func TestAppStateGetPrdCreationState(t *testing.T) {
	appState := &AppState{}

	// Should auto-initialize
	state := appState.GetPrdCreationState()
	if state == nil {
		t.Error("GetPrdCreationState should auto-initialize and return non-nil state")
	}

	// Should return same instance on subsequent calls
	state2 := appState.GetPrdCreationState()
	if state != state2 {
		t.Error("GetPrdCreationState should return same instance")
	}
}

func TestPrdCreationStateOutputBuffer(t *testing.T) {
	state := NewPrdCreationState()

	// Buffer should be initialized and empty
	if state.OutputBuffer == nil {
		t.Error("OutputBuffer should be initialized")
	}

	if state.OutputBuffer.Len() != 0 {
		t.Error("OutputBuffer should be empty initially")
	}

	// Write to buffer
	state.OutputBuffer.WriteString("Test output")
	if state.OutputBuffer.String() != "Test output" {
		t.Error("OutputBuffer should contain written data")
	}

	// Clear should reset buffer
	state.Clear()
	if state.OutputBuffer.Len() != 0 {
		t.Error("OutputBuffer should be empty after Clear()")
	}
}

func TestPrdCreationStateInProgressFlag(t *testing.T) {
	state := NewPrdCreationState()

	// Should start as false
	if state.InProgress {
		t.Error("InProgress should be false initially")
	}

	// Can be set
	state.InProgress = true
	if !state.InProgress {
		t.Error("InProgress should be true after setting")
	}

	// Clear should reset it
	state.Clear()
	if state.InProgress {
		t.Error("InProgress should be false after Clear()")
	}
}

func TestPrdCreationStatePreservationAcrossUpdates(t *testing.T) {
	state := NewPrdCreationState()

	// Set initial values
	state.Title = "Initial Title"
	state.Summary = "Initial Summary"
	state.Scope = "Initial Scope"
	state.Filename = "initial.md"
	state.OutputBuffer.WriteString("Initial output\n")
	state.InProgress = true

	// Create form values for update
	formVals := PrdFormValues{
		Title:            "Updated Title",
		Summary:          "Updated Summary",
		ScopeConstraints: "Updated Scope",
		OutputFilename:   "updated.md",
	}

	// Update from form
	state.UpdateFromFormValues(formVals)

	// Verify updates
	if state.Title != "Updated Title" {
		t.Errorf("Title should be updated, got %s", state.Title)
	}

	// Verify OutputBuffer is NOT cleared by UpdateFromFormValues
	content := state.OutputBuffer.String()
	if content != "Initial output\n" {
		t.Errorf("OutputBuffer should preserve content, got %s", content)
	}

	// Verify InProgress flag is NOT cleared by UpdateFromFormValues
	if !state.InProgress {
		t.Error("InProgress should be preserved by UpdateFromFormValues")
	}
}

func TestAppStateDialogStatePreservation(t *testing.T) {
	// Simulate dialog navigation scenario
	appState := &AppState{}

	// Start PRD creation flow
	appState.InitPrdCreationState()
	state1 := appState.PrdCreationState

	// User enters values in PRD input dialog
	appState.UpdatePrdCreationInputs("User PRD", "User Summary", "User Scope", "user-prd.md")

	// User navigates to model selection dialog (state should be preserved)
	state2 := appState.PrdCreationState
	if state1 != state2 {
		t.Error("State should be same instance during dialog navigation")
	}

	if state2.Title != "User PRD" {
		t.Error("State values should be preserved during dialog navigation")
	}

	// User cancels and returns to PRD input (state should still be preserved)
	state3 := appState.GetPrdCreationState()
	if state3.Title != "User PRD" {
		t.Error("State should be preserved when returning to dialog")
	}
}

func TestPrdCreationStateEmptyCheck(t *testing.T) {
	state := NewPrdCreationState()

	// Should be empty initially
	if !state.IsEmpty() {
		t.Error("State should be empty initially")
	}

	// Should not be empty after setting a field
	state.Title = "Title"
	if state.IsEmpty() {
		t.Error("State should not be empty after setting Title")
	}

	// Should be empty after clearing
	state.Clear()
	if !state.IsEmpty() {
		t.Error("State should be empty after Clear()")
	}
}

func TestPrdCreationStateFormValuesRoundTrip(t *testing.T) {
	// Create state with initial values
	state := NewPrdCreationState()
	appState := &AppState{}
	appState.UpdatePrdCreationInputs("Test Title", "Test Summary", "Test Scope", "test.md")

	// Convert to form values
	state.UpdateFromFormValues(PrdFormValues{
		Title:            "Test Title",
		Summary:          "Test Summary",
		ScopeConstraints: "Test Scope",
		OutputFilename:   "test.md",
	})

	// Convert back to form values
	formVals := state.ToFormValues()

	// Verify round-trip
	if formVals.Title != "Test Title" {
		t.Errorf("Title round-trip failed: %s", formVals.Title)
	}
	if formVals.Summary != "Test Summary" {
		t.Errorf("Summary round-trip failed: %s", formVals.Summary)
	}
	if formVals.ScopeConstraints != "Test Scope" {
		t.Errorf("Scope round-trip failed: %s", formVals.ScopeConstraints)
	}
	if formVals.OutputFilename != "test.md" {
		t.Errorf("Filename round-trip failed: %s", formVals.OutputFilename)
	}
}
