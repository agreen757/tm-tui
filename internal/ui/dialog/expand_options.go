package dialog

import (
	"fmt"
)

// ExpandTaskOptionsResult contains the user's selections from the expand task options dialog
type ExpandTaskOptionsResult struct {
	Depth       int  // 1-3 levels for expansion depth
	AIAssistant bool // Whether to use AI assistance for expansion
}

// NewExpandTaskOptionsDialog creates a dialog for configuring task expansion options
func NewExpandTaskOptionsDialog(style *DialogStyle) (*FormDialog, error) {
	// Define options for expansion depth
	depthOptions := []FormOption{
		{Value: "1", Label: "1 level (immediate subtasks only)"},
		{Value: "2", Label: "2 levels (subtasks and their children)"},
		{Value: "3", Label: "3 levels (deep hierarchy)"},
	}

	// Create the form fields
	fields := []FormField{
		// Radio selection for expansion depth
		{
			ID:       "depth",
			Label:    "Expansion depth:",
			Type:     FormFieldTypeRadio,
			Required: true,
			Options:  depthOptions,
			Value:    "1", // Default to 1 level
		},
		// Checkbox for AI assistance
		{
			ID:       "aiAssistant",
			Label:    "Use AI assistance for expansion",
			Type:     FormFieldTypeCheckbox,
			Required: false,
			Checked:  false,
			Value:    false,
			Help:     "Enable AI to help generate meaningful subtask descriptions",
		},
	}

	// Create the form dialog
	form := NewFormDialog(
		"Expand Task Options",
		"Configure how to expand the selected task into subtasks.",
		fields,
		[]string{"Continue", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Continue" {
				return nil, nil // Return nil for cancel
			}

			// Extract depth value
			depthStr, ok := values["depth"].(string)
			if !ok || depthStr == "" {
				return nil, fmt.Errorf("expansion depth must be selected")
			}

			var depth int
			switch depthStr {
			case "1":
				depth = 1
			case "2":
				depth = 2
			case "3":
				depth = 3
			default:
				return nil, fmt.Errorf("invalid expansion depth: %s", depthStr)
			}

			// Extract AI assistance flag
			aiAssistant := false
			if aiVal, ok := values["aiAssistant"]; ok {
				aiAssistant, _ = aiVal.(bool)
			}

			result := ExpandTaskOptionsResult{
				Depth:       depth,
				AIAssistant: aiAssistant,
			}

			return result, nil
		},
	)

	// Add custom validation
	form.AddValidator(func(values map[string]interface{}) error {
		depth, ok := values["depth"].(string)
		if !ok || depth == "" {
			return fmt.Errorf("expansion depth must be selected")
		}

		// Validate that depth is one of the allowed values
		switch depth {
		case "1", "2", "3":
			// Valid
		default:
			return fmt.Errorf("invalid expansion depth: %s", depth)
		}

		return nil
	})

	return form, nil
}
