package dialog

import (
	"github.com/adriangreen/tm-tui/internal/taskmaster"
)

// NewComplexityFilterDialog creates a dialog for filtering complexity report results
func NewComplexityFilterDialog(
	currentSettings FilterSettings,
	style *DialogStyle,
) (*FormDialog, error) {
	// Define fields for the filter form
	fields := []FormField{
		{
			ID:       "level_low",
			Label:    "Low Complexity",
			Type:     FormFieldTypeCheckbox,
			Required: false,
			Value:    currentSettings.Levels[taskmaster.ComplexityLow],
		},
		{
			ID:       "level_medium",
			Label:    "Medium Complexity",
			Type:     FormFieldTypeCheckbox,
			Required: false,
			Value:    currentSettings.Levels[taskmaster.ComplexityMedium],
		},
		{
			ID:       "level_high",
			Label:    "High Complexity",
			Type:     FormFieldTypeCheckbox,
			Required: false,
			Value:    currentSettings.Levels[taskmaster.ComplexityHigh],
		},
		{
			ID:       "level_veryhigh",
			Label:    "Very High Complexity",
			Type:     FormFieldTypeCheckbox,
			Required: false,
			Value:    currentSettings.Levels[taskmaster.ComplexityVeryHigh],
		},
		{
			ID:       "tag",
			Label:    "Filter by Tag (optional):",
			Type:     FormFieldTypeText,
			Required: false,
			Value:    currentSettings.Tag,
		},
	}

	// Create the form dialog
	form := NewFormDialog(
		"Filter Complexity Results",
		"Select which complexity levels and tags to display:",
		fields,
		[]string{"Apply", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Apply" {
				return nil, nil // Return nil for cancel
			}

			// Extract values into new FilterSettings
			settings := FilterSettings{
				Levels: map[taskmaster.ComplexityLevel]bool{
					taskmaster.ComplexityLow:      values["level_low"] == true,
					taskmaster.ComplexityMedium:   values["level_medium"] == true,
					taskmaster.ComplexityHigh:     values["level_high"] == true,
					taskmaster.ComplexityVeryHigh: values["level_veryhigh"] == true,
				},
			}

			// Extract tag filter if provided
			if tag, ok := values["tag"].(string); ok {
				settings.Tag = tag
			}

			return settings, nil
		},
	)

	// Add validator to ensure at least one level is selected
	form.AddValidator(func(values map[string]interface{}) error {
		if !(values["level_low"] == true ||
			values["level_medium"] == true ||
			values["level_high"] == true ||
			values["level_veryhigh"] == true) {
			return ErrorFormValidation{
				FieldID: "level_low", // Mark any level field for error
				Message: "At least one complexity level must be selected",
			}
		}
		return nil
	})

	return form, nil
}

// NewComplexityExportDialog creates a dialog for exporting complexity report results
func NewComplexityExportDialog(style *DialogStyle) (*FormDialog, error) {
	// Define options for export format
	formatOptions := []FormOption{
		{Value: "csv", Label: "CSV (Comma-Separated Values)"},
		{Value: "json", Label: "JSON (JavaScript Object Notation)"},
	}

	// Define fields for the export form
	fields := []FormField{
		{
			ID:       "format",
			Label:    "Export Format:",
			Type:     FormFieldTypeRadio,
			Required: true,
			Options:  formatOptions,
			Value:    "csv", // Default to CSV
		},
		{
			ID:       "path",
			Label:    "Output Path (leave empty for current directory):",
			Type:     FormFieldTypeText,
			Required: false,
			Value:    "",
		},
	}

	// Create the form dialog
	form := NewFormDialog(
		"Export Complexity Results",
		"Select export format and location:",
		fields,
		[]string{"Export", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Export" {
				return nil, nil // Return nil for cancel
			}

			// Extract export format and path
			result := map[string]string{
				"format": values["format"].(string),
			}

			// Add path if provided
			if path, ok := values["path"].(string); ok {
				result["path"] = path
			}

			return result, nil
		},
	)

	return form, nil
}
