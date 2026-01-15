package dialog

import (
	"fmt"
	"strings"
)

// UpdateTaskResult holds the result of the update task dialog
type UpdateTaskResult struct {
	TaskID  string
	Update  string
	IsEmpty bool
}

// NewUpdateTaskDialog creates an update task dialog for the given task ID
func NewUpdateTaskDialog(taskID string, style *DialogStyle) *FormDialog {
	if style == nil {
		style = DefaultDialogStyle()
	}

	// Create textarea field
	field := FormField{
		ID:          "update",
		Label:       "Update",
		Type:        FormFieldTypeTextArea,
		Required:    false,
		Placeholder: "Enter your update for this task...\n\nExamples:\n- Add implementation notes\n- Log progress on subtasks\n- Update with new findings\n- Record challenges and solutions",
		Rows:        8,
		Border:      true,
		Help:        "Enter the update content. Press Tab to navigate to buttons.",
	}

	// Create form dialog with task ID in title
	dialog := NewFormDialog(
		fmt.Sprintf("Update Task [%s]", taskID),
		"Update the selected task with new information or progress notes.",
		[]FormField{field},
		[]string{"Update", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			return handleUpdateTaskSubmit(taskID, button, values)
		},
	)

	return dialog
}

// handleUpdateTaskSubmit processes the form submission
func handleUpdateTaskSubmit(taskID, button string, values map[string]interface{}) (interface{}, error) {
	// Handle cancel button (case-insensitive)
	if strings.EqualFold(button, "cancel") {
		return nil, nil
	}

	// Extract update content
	update, ok := values["update"].(string)
	if !ok {
		update = ""
	}

	// Trim whitespace for empty check
	trimmedUpdate := strings.TrimSpace(update)

	// Check for empty update
	if trimmedUpdate == "" {
		return UpdateTaskResult{
			TaskID:  taskID,
			Update:  trimmedUpdate,
			IsEmpty: true,
		}, nil
	}

	// Return the update content and task ID
	return UpdateTaskResult{
		TaskID:  taskID,
		Update:  trimmedUpdate,
		IsEmpty: false,
	}, nil
}
