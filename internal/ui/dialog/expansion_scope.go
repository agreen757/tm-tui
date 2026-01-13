package dialog

import (
	"fmt"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// ExpansionScopeResult contains the user's expansion configuration
type ExpansionScopeResult struct {
	Depth       int  // 1-3 levels
	NumSubtasks int  // optional, 0 = auto
	UseAI       bool // --research flag
}

// NewExpansionScopeDialog creates a dialog for selecting expansion options
func NewExpansionScopeDialog(selectedTaskID string, style *DialogStyle, tagList *taskmaster.TagList) (*FormDialog, error) {
	// Create form fields with vertical layout and clear grouping
	fields := []FormField{
		// === EXPANSION DEPTH GROUP ===
		// Depth selection
		{
			ID:    "depth",
			Label: "Expansion depth:",
			Type:  FormFieldTypeRadio,
			Options: []FormOption{
				{Value: "1", Label: "1 level", Description: "Single layer of subtasks"},
				{Value: "2", Label: "2 levels (recommended)", Description: "Most balanced option"},
				{Value: "3", Label: "3 levels", Description: "Deep nested hierarchy"},
			},
			Value: "2",
			Help:  "How many levels of subtasks to create",
		},
		// === OPTIONS GROUP ===
		// Number of subtasks
		{
			ID:          "num",
			Label:       "Number of subtasks per task:",
			Type:        FormFieldTypeText,
			Placeholder: "Leave blank for auto-detection",
			Help:        "Optional: specify exact subtask count",
		},
		// AI research flag
		{
			ID:      "research",
			Label:   "Enable AI-powered expansion (--research)",
			Type:    FormFieldTypeCheckbox,
			Value:   true,
			Checked: true,
			Help:    "Use AI research for more intelligent task breakdown",
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

			// Extract depth
			depthStr, _ := values["depth"].(string)
			depth := parseIntValue(depthStr, 2)

			// Extract num
			numStr, _ := values["num"].(string)
			num := parseIntValue(numStr, 0)

			// Extract research flag
			research, _ := values["research"].(bool)

			result := ExpansionScopeResult{
				Depth:       depth,
				NumSubtasks: num,
				UseAI:       research,
			}

			return result, nil
		},
	)

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
