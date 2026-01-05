package ui

import (
	"testing"
)

func TestPrdCreationState(t *testing.T) {
	state := NewPrdCreationState()

	if !state.IsEmpty() {
		t.Error("New state should be empty")
	}

	// Update state
	state.Title = "Test PRD"
	state.Summary = "A test summary"
	state.Filename = "test.md"

	if state.IsEmpty() {
		t.Error("State with values should not be empty")
	}

	// Convert to form values
	formVals := state.ToFormValues()
	if formVals.Title != "Test PRD" {
		t.Errorf("Expected title 'Test PRD', got '%s'", formVals.Title)
	}

	// Clear and check
	state.Clear()
	if !state.IsEmpty() {
		t.Error("Cleared state should be empty")
	}
}

func TestPrdCreationStateUpdateFromFormValues(t *testing.T) {
	state := NewPrdCreationState()

	formVals := PrdFormValues{
		Title:            "My PRD",
		Summary:          "This is a summary",
		ScopeConstraints: "Limitations",
		OutputFilename:   "my-prd.md",
	}

	state.UpdateFromFormValues(formVals)

	if state.Title != "My PRD" {
		t.Errorf("Expected title 'My PRD', got '%s'", state.Title)
	}
	if state.Summary != "This is a summary" {
		t.Errorf("Expected summary 'This is a summary', got '%s'", state.Summary)
	}
	if state.Scope != "Limitations" {
		t.Errorf("Expected scope 'Limitations', got '%s'", state.Scope)
	}
	if state.Filename != "my-prd.md" {
		t.Errorf("Expected filename 'my-prd.md', got '%s'", state.Filename)
	}
}

func TestPrdCreationStateRoundTrip(t *testing.T) {
	originalState := &PrdCreationState{
		Title:    "Round Trip Test",
		Summary:  "Testing state preservation",
		Scope:    "Test scope",
		Filename: "round-trip.md",
	}

	// Convert to form values
	formVals := originalState.ToFormValues()

	// Create new state and update from form values
	newState := NewPrdCreationState()
	newState.UpdateFromFormValues(formVals)

	// Verify all values match
	if newState.Title != originalState.Title {
		t.Errorf("Title mismatch: %s != %s", newState.Title, originalState.Title)
	}
	if newState.Summary != originalState.Summary {
		t.Errorf("Summary mismatch: %s != %s", newState.Summary, originalState.Summary)
	}
	if newState.Scope != originalState.Scope {
		t.Errorf("Scope mismatch: %s != %s", newState.Scope, originalState.Scope)
	}
	if newState.Filename != originalState.Filename {
		t.Errorf("Filename mismatch: %s != %s", newState.Filename, originalState.Filename)
	}
}
