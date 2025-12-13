package dialog

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ExpansionScopeResult contains the user's expansion configuration
type ExpansionScopeResult struct {
	Scope       string   // "single", "all", "range", "tag"
	TaskID      string   // for single task expansion
	FromID      string   // for range expansion
	ToID        string   // for range expansion
	Tags        []string // for tag-based expansion
	Depth       int      // 1-3 levels
	NumSubtasks int      // optional, 0 = auto
	UseAI       bool     // --research flag
}

// NewExpansionScopeDialog creates a dialog for selecting expansion scope and options
func NewExpansionScopeDialog(selectedTaskID string, style *DialogStyle) (*FormDialog, error) {
	// Determine default scope and options
	defaultScope := "all"
	scopeOptions := []FormOption{
		{Value: "all", Label: "All tasks in project"},
		{Value: "range", Label: "Task range (from/to)"},
		{Value: "tag", Label: "Tasks with specific tag"},
	}

	if selectedTaskID != "" {
		defaultScope = "single"
		scopeOptions = append([]FormOption{
			{Value: "single", Label: fmt.Sprintf("Selected task only (%s)", selectedTaskID)},
		}, scopeOptions...)
	}

	// Create form fields
	fields := []FormField{
		// Scope selection
		{
			ID:       "scope",
			Label:    "Expansion scope:",
			Type:     FormFieldTypeRadio,
			Required: true,
			Options:  scopeOptions,
			Value:    defaultScope,
		},
		// Task ID (for single scope)
		{
			ID:              "taskID",
			Label:           "Task ID:",
			Type:            FormFieldTypeText,
			Value:           selectedTaskID,
			Placeholder:     "e.g., 1.2",
			ConditionalShow: "scope=single",
		},
		// From ID (for range scope)
		{
			ID:              "fromID",
			Label:           "From task ID:",
			Type:            FormFieldTypeText,
			Placeholder:     "e.g., 1",
			ConditionalShow: "scope=range",
		},
		// To ID (for range scope)
		{
			ID:              "toID",
			Label:           "To task ID:",
			Type:            FormFieldTypeText,
			Placeholder:     "e.g., 5",
			ConditionalShow: "scope=range",
		},
		// Tags (for tag scope)
		{
			ID:              "tags",
			Label:           "Tags (comma-separated):",
			Type:            FormFieldTypeText,
			Placeholder:     "e.g., backend,api",
			ConditionalShow: "scope=tag",
		},
		// Depth selection
		{
			ID:    "depth",
			Label: "Expansion depth:",
			Type:  FormFieldTypeRadio,
			Options: []FormOption{
				{Value: "1", Label: "1 level"},
				{Value: "2", Label: "2 levels (recommended)"},
				{Value: "3", Label: "3 levels"},
			},
			Value: "2",
		},
		// Number of subtasks
		{
			ID:          "num",
			Label:       "Number of subtasks per task:",
			Type:        FormFieldTypeText,
			Placeholder: "Leave blank for auto-detection",
		},
		// AI research flag
		{
			ID:      "research",
			Label:   "Enable AI-powered expansion (--research)",
			Type:    FormFieldTypeCheckbox,
			Checked: true,
		},
	}

	// Create form dialog
	form := NewFormDialog(
		"Expand Tasks",
		"Configure task expansion options. This will execute 'task-master expand' command.",
		fields,
		[]string{"Expand", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Expand" {
				return nil, nil
			}

			// Extract scope
			scope, ok := values["scope"].(string)
			if !ok || scope == "" {
				return nil, fmt.Errorf("scope is required")
			}

			// Extract depth
			depthStr, _ := values["depth"].(string)
			depth := parseIntValue(depthStr, 2)

			// Extract num
			numStr, _ := values["num"].(string)
			num := parseIntValue(numStr, 0)

			// Extract research flag
			research, _ := values["research"].(bool)

			result := ExpansionScopeResult{
				Scope:       scope,
				Depth:       depth,
				NumSubtasks: num,
				UseAI:       research,
			}

			// Validate and extract scope-specific fields
			switch scope {
			case "single":
				taskID, _ := values["taskID"].(string)
				taskID = strings.TrimSpace(taskID)
				if taskID == "" {
					return nil, fmt.Errorf("task ID is required for single scope")
				}
				result.TaskID = taskID

			case "range":
				fromID, _ := values["fromID"].(string)
				toID, _ := values["toID"].(string)
				fromID = strings.TrimSpace(fromID)
				toID = strings.TrimSpace(toID)
				if fromID == "" && toID == "" {
					return nil, fmt.Errorf("at least one of from/to ID is required for range")
				}
				result.FromID = fromID
				result.ToID = toID

			case "tag":
				tagsStr, _ := values["tags"].(string)
				tagsStr = strings.TrimSpace(tagsStr)
				if tagsStr == "" {
					return nil, fmt.Errorf("tags are required for tag scope")
				}
				result.Tags = parseTagList(tagsStr)
			}

			return result, nil
		},
	)

	// Add validator
	form.AddValidator(func(values map[string]interface{}) error {
		scope, _ := values["scope"].(string)

		switch scope {
		case "single":
			if selectedTaskID == "" {
				taskID, _ := values["taskID"].(string)
				if strings.TrimSpace(taskID) == "" {
					return ErrorFormValidation{FieldID: "taskID", Message: "Task ID is required"}
				}
			}

		case "range":
			fromID, _ := values["fromID"].(string)
			toID, _ := values["toID"].(string)
			if strings.TrimSpace(fromID) == "" && strings.TrimSpace(toID) == "" {
				return fmt.Errorf("at least one of from/to ID is required")
			}

		case "tag":
			tags, _ := values["tags"].(string)
			if strings.TrimSpace(tags) == "" {
				return ErrorFormValidation{FieldID: "tags", Message: "At least one tag is required"}
			}
		}

		return nil
	})

	// Event handler to reset fields when scope changes
	form.AddEventHandler(func(form *FormDialog, msg tea.Msg) {
		switch msg := msg.(type) {
		case FormValueChangedMsg:
			if msg.FieldID == "scope" {
				// Reset fields when scope changes
				if msg.NewValue != "single" {
					if field, ok := form.GetField("taskID"); ok {
						field.Value = selectedTaskID
					}
				}
				if msg.NewValue != "range" {
					if field, ok := form.GetField("fromID"); ok {
						field.Value = ""
						field.input.Reset()
					}
					if field, ok := form.GetField("toID"); ok {
						field.Value = ""
						field.input.Reset()
					}
				}
				if msg.NewValue != "tag" {
					if field, ok := form.GetField("tags"); ok {
						field.Value = ""
						field.input.Reset()
					}
				}
			}
		}
	})

	return form, nil
}

// parseIntValue parses a string to int with fallback
func parseIntValue(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return fallback
	}
	if result <= 0 {
		return fallback
	}
	return result
}

// parseTagList parses comma-separated tags
func parseTagList(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper to convert interface{} to int
func intValue(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

// Helper to convert interface{} to bool
func boolValue(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// Helper to convert interface{} to string
func stringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
