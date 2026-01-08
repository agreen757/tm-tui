package dialog

import (
	"testing"
)

func TestNewCommandRunnerDialog_Creation(t *testing.T) {
	// Test with nil style (should default)
	dialog := NewCommandRunnerDialog(nil)

	if dialog == nil {
		t.Fatal("NewCommandRunnerDialog returned nil")
	}

	if dialog.Title() != "Run Command with Crush" {
		t.Errorf("Expected title 'Run Command with Crush', got '%s'", dialog.Title())
	}

	if dialog.Description != "Execute an ad-hoc command using Crush AI without creating a formal task." {
		t.Errorf("Expected correct description, got '%s'", dialog.Description)
	}

	if dialog.Kind() != DialogKindForm {
		t.Errorf("Expected kind DialogKindForm, got %v", dialog.Kind())
	}
}

func TestNewCommandRunnerDialog_Fields(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)

	// Get the fields - check that there's exactly one field
	if len(dialog.fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(dialog.fields))
	}

	field := dialog.fields[0]
	if field.ID != "prompt" {
		t.Errorf("Expected field ID 'prompt', got '%s'", field.ID)
	}

	if field.Label != "Prompt" {
		t.Errorf("Expected label 'Prompt', got '%s'", field.Label)
	}

	if field.Type != FormFieldTypeTextArea {
		t.Errorf("Expected type FormFieldTypeTextArea, got %v", field.Type)
	}

	if !field.Required {
		t.Error("Expected field to be required")
	}

	if field.Placeholder != "Enter your command prompt for Crush AI..." {
		t.Errorf("Expected correct placeholder, got '%s'", field.Placeholder)
	}

	if field.Rows != 8 {
		t.Errorf("Expected 8 rows, got %d", field.Rows)
	}

	if !field.Border {
		t.Error("Expected border to be enabled")
	}

	if field.Help != "Enter the command you want Crush AI to execute." {
		t.Errorf("Expected correct help text, got '%s'", field.Help)
	}
}

func TestNewCommandRunnerDialog_Buttons(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)

	expectedButtons := []string{"Execute", "Cancel"}
	if len(dialog.buttons) != len(expectedButtons) {
		t.Errorf("Expected %d buttons, got %d", len(expectedButtons), len(dialog.buttons))
		return
	}

	for i, expected := range expectedButtons {
		if dialog.buttons[i] != expected {
			t.Errorf("Button %d: expected '%s', got '%s'", i, expected, dialog.buttons[i])
		}
	}
}

func TestNewCommandRunnerDialog_WithStyle(t *testing.T) {
	// Test with custom style
	style := DefaultDialogStyle()
	dialog := NewCommandRunnerDialog(style)

	if dialog == nil {
		t.Fatal("NewCommandRunnerDialog with custom style returned nil")
	}

	if dialog.Style != style {
		t.Error("Expected dialog to use provided style")
	}
}

func TestCommandPromptResultStruct(t *testing.T) {
	// Test that the struct can be instantiated
	result := CommandPromptResult{Prompt: "test prompt"}

	if result.Prompt != "test prompt" {
		t.Errorf("Expected 'test prompt', got '%s'", result.Prompt)
	}
}

func TestHandleCommandRunnerSubmit_Cancel(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": "test"}

	// Test cancel button (case-insensitive)
	result, err := handleCommandRunnerSubmit(dialog, "cancel", values)

	if err != nil {
		t.Errorf("Expected no error on cancel, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result on cancel, got %v", result)
	}
}

func TestHandleCommandRunnerSubmit_CancelUppercase(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": "test"}

	// Test cancel button in uppercase
	result, err := handleCommandRunnerSubmit(dialog, "CANCEL", values)

	if err != nil {
		t.Errorf("Expected no error on CANCEL, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result on CANCEL, got %v", result)
	}
}

func TestHandleCommandRunnerSubmit_MixedCaseCancel(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": "test"}

	// Test cancel button in mixed case
	result, err := handleCommandRunnerSubmit(dialog, "Cancel", values)

	if err != nil {
		t.Errorf("Expected no error on Cancel, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result on Cancel, got %v", result)
	}
}

func TestHandleCommandRunnerSubmit_ValidPrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": "test prompt"}

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	commandResult, ok := result.(CommandPromptResult)
	if !ok {
		t.Fatalf("Expected CommandPromptResult, got %T", result)
	}

	if commandResult.Prompt != "test prompt" {
		t.Errorf("Expected 'test prompt', got '%s'", commandResult.Prompt)
	}
}

func TestHandleCommandRunnerSubmit_PromptWithWhitespace(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": "  test prompt  "}

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	commandResult, ok := result.(CommandPromptResult)
	if !ok {
		t.Fatalf("Expected CommandPromptResult, got %T", result)
	}

	// Should be trimmed
	if commandResult.Prompt != "test prompt" {
		t.Errorf("Expected 'test prompt', got '%s'", commandResult.Prompt)
	}
}

func TestHandleCommandRunnerSubmit_EmptyPrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": ""}

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err == nil {
		t.Fatal("Expected error for empty prompt")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	validationErr, ok := err.(ErrorFormValidation)
	if !ok {
		t.Fatalf("Expected ErrorFormValidation, got %T", err)
	}

	if validationErr.FieldID != "prompt" {
		t.Errorf("Expected FieldID 'prompt', got '%s'", validationErr.FieldID)
	}

	if validationErr.Message != "Prompt cannot be empty" {
		t.Errorf("Expected correct message, got '%s'", validationErr.Message)
	}
}

func TestHandleCommandRunnerSubmit_WhitespaceOnlyPrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": "   \t\n  "}

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err == nil {
		t.Fatal("Expected error for whitespace-only prompt")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	validationErr, ok := err.(ErrorFormValidation)
	if !ok {
		t.Fatalf("Expected ErrorFormValidation, got %T", err)
	}

	if validationErr.Message != "Prompt cannot be empty" {
		t.Errorf("Expected 'Prompt cannot be empty', got '%s'", validationErr.Message)
	}
}

func TestHandleCommandRunnerSubmit_MissingPrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{}

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err == nil {
		t.Fatal("Expected error for missing prompt")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

func TestHandleCommandRunnerSubmit_InvalidPromptType(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	values := map[string]interface{}{"prompt": 123} // Wrong type

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err == nil {
		t.Fatal("Expected error for invalid prompt type")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	validationErr, ok := err.(ErrorFormValidation)
	if !ok {
		t.Fatalf("Expected ErrorFormValidation, got %T", err)
	}

	if validationErr.FieldID != "prompt" {
		t.Errorf("Expected FieldID 'prompt', got '%s'", validationErr.FieldID)
	}
}

func TestHandleCommandRunnerSubmit_MultilinePrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)
	multilinePrompt := `Build a REST API with:
- User authentication
- Role-based access control
- Database integration`

	values := map[string]interface{}{"prompt": multilinePrompt}

	result, err := handleCommandRunnerSubmit(dialog, "Execute", values)

	if err != nil {
		t.Errorf("Expected no error for multiline prompt, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	commandResult, ok := result.(CommandPromptResult)
	if !ok {
		t.Fatalf("Expected CommandPromptResult, got %T", result)
	}

	if commandResult.Prompt != multilinePrompt {
		t.Errorf("Expected prompt to be preserved, got '%s'", commandResult.Prompt)
	}
}

func TestValidatorRejectsEmptyPrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)

	// Manually collect values and test validator
	values := map[string]interface{}{"prompt": ""}

	// Apply validators
	err := dialog.applyValidators(values)

	if err == nil {
		t.Fatal("Expected validator to reject empty prompt")
	}

	validationErr, ok := err.(ErrorFormValidation)
	if !ok {
		t.Fatalf("Expected ErrorFormValidation, got %T", err)
	}

	if validationErr.FieldID != "prompt" {
		t.Errorf("Expected FieldID 'prompt', got '%s'", validationErr.FieldID)
	}
}

func TestValidatorRejectsWhitespacePrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)

	values := map[string]interface{}{"prompt": "   \n\t  "}

	// Apply validators
	err := dialog.applyValidators(values)

	if err == nil {
		t.Fatal("Expected validator to reject whitespace-only prompt")
	}
}

func TestValidatorAcceptsValidPrompt(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)

	values := map[string]interface{}{"prompt": "Valid prompt"}

	// Apply validators
	err := dialog.applyValidators(values)

	if err != nil {
		t.Errorf("Expected validator to accept valid prompt, got error: %v", err)
	}
}

func TestCommandRunnerDialog_Integration(t *testing.T) {
	dialog := NewCommandRunnerDialog(nil)

	// Verify dialog is ready for use
	if dialog.Kind() != DialogKindForm {
		t.Errorf("Expected DialogKindForm, got %v", dialog.Kind())
	}

	if len(dialog.fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(dialog.fields))
	}

	if len(dialog.buttons) != 2 {
		t.Errorf("Expected 2 buttons, got %d", len(dialog.buttons))
	}

	// Verify button labels
	if dialog.buttons[0] != "Execute" || dialog.buttons[1] != "Cancel" {
		t.Errorf("Unexpected button labels: %v", dialog.buttons)
	}
}
