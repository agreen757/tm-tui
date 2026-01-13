package dialog

import (
	"fmt"
)

// ComplexityScopeResult contains the user's selections from the complexity scope dialog
type ComplexityScopeResult struct {
	Scope string // "all", "selected"
}

// NewComplexityScopeDialog creates a dialog for selecting the scope of complexity analysis
func NewComplexityScopeDialog(selectedTaskID string, style *DialogStyle) (*FormDialog, error) {
	// Define options in the form
	options := []FormOption{
		{Value: "all", Label: "All tasks in project"},
		{Value: "selected", Label: fmt.Sprintf("Selected task only (%s)", selectedTaskID)},
	}

	// Create the form fields
	fields := []FormField{
		// Radio selection for scope
		{
			ID:       "scope",
			Label:    "Select analysis scope:",
			Type:     FormFieldTypeRadio,
			Required: true,
			Options:  options,
			Value:    "all", // Default to "all"
		},
	}

	// Create the form dialog
	form := NewFormDialog(
		"Analyze Task Complexity",
		"Select the scope of tasks to analyze (Ctrl+Shift+C).",
		fields,
		[]string{"Analyze", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Analyze" {
				return nil, nil // Return nil for cancel
			}

			// Extract values
			scope, _ := values["scope"].(string)
			if scope == "" {
				return nil, fmt.Errorf("no scope selected")
			}

			result := ComplexityScopeResult{
				Scope: scope,
			}

			return result, nil
		},
	)

	// Validate that selected task ID is provided when scope is "selected"
	form.AddValidator(func(values map[string]interface{}) error {
		scope, _ := values["scope"].(string)
		if scope == "selected" && selectedTaskID == "" {
			return fmt.Errorf("No task is currently selected")
		}
		return nil
	})

	return form, nil
}
