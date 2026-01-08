package dialog

import (
	"strings"
)

// CommandPromptResult holds the result of the command runner dialog
type CommandPromptResult struct {
	Prompt string
}

// NewCommandRunnerDialog creates a command runner dialog
func NewCommandRunnerDialog(style *DialogStyle) *FormDialog {
	if style == nil {
		style = DefaultDialogStyle()
	}

	// Create textarea field
	field := FormField{
		ID:          "prompt",
		Label:       "Prompt",
		Type:        FormFieldTypeTextArea,
		Required:    true,
		Placeholder: "Enter your command prompt for Crush AI...",
		Rows:        8,
		Border:      true,
		Help:        "Enter the command you want Crush AI to execute.",
	}

	// Create form dialog
	dialog := NewFormDialog(
		"Run Command with Crush",
		"Execute an ad-hoc command using Crush AI without creating a formal task.",
		[]FormField{field},
		[]string{"Execute", "Cancel"},
		style,
		handleCommandRunnerSubmit,
	)

	// Add custom validator
	dialog.AddValidator(func(values map[string]interface{}) error {
		if prompt, ok := values["prompt"].(string); ok {
			if strings.TrimSpace(prompt) == "" {
				return ErrorFormValidation{
					FieldID: "prompt",
					Message: "Prompt cannot be empty",
				}
			}
		}
		return nil
	})

	return dialog
}

// handleCommandRunnerSubmit processes the form submission
func handleCommandRunnerSubmit(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
	// Handle cancel button (case-insensitive)
	if strings.EqualFold(button, "cancel") {
		return nil, nil
	}

	// Validate and extract prompt
	prompt, ok := values["prompt"].(string)
	if !ok {
		return nil, ErrorFormValidation{
			FieldID: "prompt",
			Message: "Invalid prompt value",
		}
	}

	// Trim whitespace and validate non-empty
	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		return nil, ErrorFormValidation{
			FieldID: "prompt",
			Message: "Prompt cannot be empty",
		}
	}

	// Return the prompt value
	return CommandPromptResult{
		Prompt: trimmedPrompt,
	}, nil
}
