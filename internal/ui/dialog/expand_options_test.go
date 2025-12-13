package dialog

import (
	"testing"
)

func TestNewExpandTaskOptionsDialog(t *testing.T) {
	style := DefaultDialogStyle()

	dialog, err := NewExpandTaskOptionsDialog(style)
	if err != nil {
		t.Fatalf("Failed to create ExpandTaskOptionsDialog: %v", err)
	}

	if dialog == nil {
		t.Error("Expected dialog to not be nil")
	}

	if dialog.Title() != "Expand Task Options" {
		t.Errorf("Expected title 'Expand Task Options', got '%s'", dialog.Title())
	}

	if len(dialog.fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(dialog.fields))
	}
}

func TestExpandTaskOptionsDialogDepthField(t *testing.T) {
	style := DefaultDialogStyle()
	dialog, _ := NewExpandTaskOptionsDialog(style)

	depthField, ok := dialog.GetField("depth")
	if !ok {
		t.Fatal("Expected to find 'depth' field")
	}

	if depthField.Type != FormFieldTypeRadio {
		t.Errorf("Expected depth field to be radio type, got %v", depthField.Type)
	}

	if len(depthField.Options) != 3 {
		t.Errorf("Expected 3 depth options, got %d", len(depthField.Options))
	}

	expectedValues := []string{"1", "2", "3"}
	for i, opt := range depthField.Options {
		if opt.Value != expectedValues[i] {
			t.Errorf("Expected option %d to have value '%s', got '%s'", i, expectedValues[i], opt.Value)
		}
	}
}

func TestExpandTaskOptionsDialogAIAssistantField(t *testing.T) {
	style := DefaultDialogStyle()
	dialog, _ := NewExpandTaskOptionsDialog(style)

	aiField, ok := dialog.GetField("aiAssistant")
	if !ok {
		t.Fatal("Expected to find 'aiAssistant' field")
	}

	if aiField.Type != FormFieldTypeCheckbox {
		t.Errorf("Expected aiAssistant field to be checkbox type, got %v", aiField.Type)
	}

	if aiField.Checked {
		t.Error("Expected aiAssistant checkbox to be unchecked by default")
	}
}

func TestExpandTaskOptionsDialogValidation(t *testing.T) {
	style := DefaultDialogStyle()
	dialog, _ := NewExpandTaskOptionsDialog(style)

	// Test that validators are properly registered
	if len(dialog.validators) == 0 {
		t.Error("Expected validators to be registered")
	}

	// Test validator function directly with valid values
	validValues := map[string]interface{}{
		"depth":       "1",
		"aiAssistant": false,
	}

	// Call the first validator to test it
	if len(dialog.validators) > 0 {
		err := dialog.validators[0](validValues)
		if err != nil {
			t.Errorf("Expected validation to pass for valid values, got error: %v", err)
		}

		// Test invalid depth value
		invalidValues := map[string]interface{}{
			"depth":       "5",
			"aiAssistant": false,
		}
		err = dialog.validators[0](invalidValues)
		if err == nil {
			t.Error("Expected validation to fail for invalid depth value")
		}
	}
}

func TestExpandTaskOptionsDialogSubmitHandler(t *testing.T) {
	style := DefaultDialogStyle()
	dialog, err := NewExpandTaskOptionsDialog(style)
	if err != nil {
		t.Fatalf("Failed to create dialog: %v", err)
	}

	// Test cancel button
	result, err := dialog.handler(dialog, "Cancel", map[string]interface{}{})
	if err != nil {
		t.Errorf("Cancel button should not return error, got: %v", err)
	}
	if result != nil {
		t.Error("Cancel button should return nil result")
	}

	// Test continue with valid values
	testCases := []struct {
		depthStr    string
		expectedInt int
		description string
	}{
		{"1", 1, "single level"},
		{"2", 2, "two levels"},
		{"3", 3, "three levels"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			values := map[string]interface{}{
				"depth":       tc.depthStr,
				"aiAssistant": true,
			}

			result, err := dialog.handler(dialog, "Continue", values)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result to not be nil")
			}

			expandResult, ok := result.(ExpandTaskOptionsResult)
			if !ok {
				t.Fatalf("Expected ExpandTaskOptionsResult, got %T", result)
			}

			if expandResult.Depth != tc.expectedInt {
				t.Errorf("Expected depth %d, got %d", tc.expectedInt, expandResult.Depth)
			}

			if !expandResult.AIAssistant {
				t.Error("Expected AIAssistant to be true")
			}
		})
	}
}

func TestExpandTaskOptionsDialogDefaultValues(t *testing.T) {
	style := DefaultDialogStyle()
	dialog, _ := NewExpandTaskOptionsDialog(style)

	depthField, _ := dialog.GetField("depth")
	if depthField.Value != "1" {
		t.Errorf("Expected default depth to be '1', got '%v'", depthField.Value)
	}

	aiField, _ := dialog.GetField("aiAssistant")
	if aiField.Checked {
		t.Error("Expected default aiAssistant to be false")
	}
}

func TestExpandTaskOptionsResultStructure(t *testing.T) {
	result := ExpandTaskOptionsResult{
		Depth:       2,
		AIAssistant: true,
	}

	if result.Depth != 2 {
		t.Errorf("Expected Depth to be 2, got %d", result.Depth)
	}

	if !result.AIAssistant {
		t.Error("Expected AIAssistant to be true")
	}
}
