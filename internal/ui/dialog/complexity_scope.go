package dialog

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ComplexityScopeResult contains the user's selections from the complexity scope dialog
type ComplexityScopeResult struct {
	Scope string   // "all", "selected", "tag"
	Tags  []string // Only used when scope is "tag"
}

// NewComplexityScopeDialog creates a dialog for selecting the scope of complexity analysis
func NewComplexityScopeDialog(selectedTaskID string, style *DialogStyle) (*FormDialog, error) {
	// Define options in the form
	options := []FormOption{
		{Value: "all", Label: "All tasks in project"},
		{Value: "selected", Label: fmt.Sprintf("Selected task only (%s)", selectedTaskID)},
		{Value: "tag", Label: "Tasks with specific tag"},
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
		// Text input for tag (only shown when "tag" is selected)
		{
			ID:              "tag",
			Label:           "Enter tag:",
			Type:            FormFieldTypeText,
			Required:        false,
			ConditionalShow: "scope=tag",
		},
	}

	// Create the form dialog
	form := NewFormDialog(
		"Analyze Task Complexity",
		"Select the scope of tasks to analyze (Alt+C).",
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

			if scope == "tag" {
				tag, ok := values["tag"].(string)
				if !ok || tag == "" {
					return nil, fmt.Errorf("tag must be provided when 'Tasks with tag' is selected")
				}
				result.Tags = []string{tag}
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
		if scope == "tag" {
			if tag, _ := values["tag"].(string); strings.TrimSpace(tag) == "" {
				return ErrorFormValidation{FieldID: "tag", Message: "Enter at least one tag"}
			}
		}
		return nil
	})

	// Event handler to reset tag field when scope changes
	form.AddEventHandler(func(form *FormDialog, msg tea.Msg) {
		switch msg := msg.(type) {
		case FormValueChangedMsg:
			if msg.FieldID == "scope" {
				// When scope changes, reset the tag field if it's not needed
				if msg.NewValue != "tag" {
					if tagField, ok := form.GetField("tag"); ok {
						tagField.Value = ""
						tagField.input.Reset()
					}
				}
			}
		}
	})

	return form, nil
}
