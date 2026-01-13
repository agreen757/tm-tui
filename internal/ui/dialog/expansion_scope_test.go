package dialog

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

func TestNewExpansionScopeDialog_WithoutTagList(t *testing.T) {
	dialog, err := NewExpansionScopeDialog("", nil, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if dialog == nil {
		t.Fatal("Expected dialog, got nil")
	}
	if dialog.TitleText != "Expand Tasks" {
		t.Errorf("Expected title 'Expand Tasks', got: %s", dialog.TitleText)
	}
}

func TestNewExpansionScopeDialog_WithTagList(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "backend",
				TaskCount:      10,
				CompletedCount: 5,
				Description:    "Backend tasks",
			},
		},
	}

	dialog, err := NewExpansionScopeDialog("", nil, tagList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if dialog == nil {
		t.Fatal("Expected dialog, got nil")
	}
}

func TestExpansionScopeDialog_DepthField(t *testing.T) {
	dialog, err := NewExpansionScopeDialog("", nil, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Find the depth field
	depthField, ok := dialog.GetField("depth")
	if !ok {
		t.Fatal("Expected to find depth field")
	}

	// Verify it's a radio field
	if depthField.Type != FormFieldTypeRadio {
		t.Errorf("Expected FormFieldTypeRadio, got: %d", depthField.Type)
	}

	// Verify default value is "2"
	if depthField.Value != "2" {
		t.Errorf("Expected default depth '2', got: %v", depthField.Value)
	}

	// Verify three options exist
	if len(depthField.Options) != 3 {
		t.Errorf("Expected 3 depth options, got: %d", len(depthField.Options))
	}
}

func TestExpansionScopeDialog_ResultExtraction(t *testing.T) {
	dialog, err := NewExpansionScopeDialog("", nil, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Create test values
	values := map[string]interface{}{
		"depth":    "2",
		"num":      "3",
		"research": true,
	}

	// Get the submit handler and test it
	if dialog.handler == nil {
		t.Fatal("Expected handler to be set")
	}

	result, err := dialog.handler(dialog, "Expand", values)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expansionResult, ok := result.(ExpansionScopeResult)
	if !ok {
		t.Errorf("Expected ExpansionScopeResult, got: %T", result)
	}

	if expansionResult.Depth != 2 {
		t.Errorf("Expected depth 2, got: %d", expansionResult.Depth)
	}
	if expansionResult.NumSubtasks != 3 {
		t.Errorf("Expected 3 subtasks, got: %d", expansionResult.NumSubtasks)
	}
	if !expansionResult.UseAI {
		t.Error("Expected UseAI to be true")
	}
}

func TestExpansionScopeDialog_CancelButton(t *testing.T) {
	dialog, err := NewExpansionScopeDialog("", nil, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	values := map[string]interface{}{
		"depth":    "2",
		"num":      "",
		"research": false,
	}

	result, err := dialog.handler(dialog, "Cancel", values)
	if err != nil {
		t.Errorf("Expected no error on cancel, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result on cancel, got: %v", result)
	}
}

func TestExpansionScopeDialog_DefaultValues(t *testing.T) {
	dialog, err := NewExpansionScopeDialog("", nil, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check depth default
	depthField, _ := dialog.GetField("depth")
	if depthField.Value != "2" {
		t.Errorf("Expected default depth '2', got: %v", depthField.Value)
	}

	// Check research default
	researchField, _ := dialog.GetField("research")
	if !researchField.Checked {
		t.Error("Expected research checkbox to be checked by default")
	}
}
