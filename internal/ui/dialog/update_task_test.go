package dialog

import (
	"testing"
)

func TestNewUpdateTaskDialog(t *testing.T) {
	tests := []struct {
		name      string
		taskID    string
		style     *DialogStyle
		wantTitle string
	}{
		{
			name:      "simple task ID",
			taskID:    "1",
			style:     nil,
			wantTitle: "Update Task [1]",
		},
		{
			name:      "dotted task ID",
			taskID:    "2.3.1",
			style:     nil,
			wantTitle: "Update Task [2.3.1]",
		},
		{
			name:      "with custom style",
			taskID:    "5",
			style:     DefaultDialogStyle(),
			wantTitle: "Update Task [5]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewUpdateTaskDialog(tt.taskID, tt.style)

			if dialog == nil {
				t.Errorf("NewUpdateTaskDialog() returned nil")
				return
			}

			if dialog.Title() != tt.wantTitle {
				t.Errorf("Title() = %q, want %q", dialog.Title(), tt.wantTitle)
			}

			if dialog.Description != "Update the selected task with new information or progress notes." {
				t.Errorf("Description not set correctly")
			}

			// Verify fields
			if len(dialog.fields) != 1 {
				t.Errorf("Expected 1 field, got %d", len(dialog.fields))
				return
			}

			field := dialog.fields[0]
			if field.ID != "update" {
				t.Errorf("Field ID = %q, want %q", field.ID, "update")
			}

			if field.Type != FormFieldTypeTextArea {
				t.Errorf("Field Type = %v, want %v", field.Type, FormFieldTypeTextArea)
			}

			if field.Rows != 8 {
				t.Errorf("Field Rows = %d, want 8", field.Rows)
			}

			if !field.Border {
				t.Errorf("Field Border = false, want true")
			}

			if field.Required {
				t.Errorf("Field Required = true, want false")
			}

			// Verify buttons
			if len(dialog.buttons) != 2 {
				t.Errorf("Expected 2 buttons, got %d", len(dialog.buttons))
			}

			if dialog.buttons[0] != "Update" {
				t.Errorf("First button = %q, want %q", dialog.buttons[0], "Update")
			}

			if dialog.buttons[1] != "Cancel" {
				t.Errorf("Second button = %q, want %q", dialog.buttons[1], "Cancel")
			}
		})
	}
}

func TestUpdateTaskDialog_DialogProperties(t *testing.T) {
	dialog := NewUpdateTaskDialog("1", nil)

	if dialog.Kind() != DialogKindForm {
		t.Errorf("Dialog Kind() = %v, want %v", dialog.Kind(), DialogKindForm)
	}

	// Verify placeholder text
	if dialog.fields[0].Placeholder == "" {
		t.Errorf("Placeholder should not be empty")
	}

	if dialog.fields[0].Help == "" {
		t.Errorf("Help text should not be empty")
	}
}

func TestHandleUpdateTaskSubmit_Cancel(t *testing.T) {
	taskID := "1"
	button := "Cancel"
	values := map[string]interface{}{
		"update": "Some update text",
	}

	result, err := handleUpdateTaskSubmit(taskID, button, values)

	if err != nil {
		t.Errorf("handleUpdateTaskSubmit() returned error: %v", err)
	}

	if result != nil {
		t.Errorf("handleUpdateTaskSubmit() for cancel should return nil, got %v", result)
	}
}

func TestHandleUpdateTaskSubmit_CaseInsensitiveCancel(t *testing.T) {
	tests := []string{"cancel", "CANCEL", "Cancel", "cAnCeL"}

	for _, cancelBtn := range tests {
		t.Run(cancelBtn, func(t *testing.T) {
			result, err := handleUpdateTaskSubmit("1", cancelBtn, map[string]interface{}{})

			if err != nil {
				t.Errorf("handleUpdateTaskSubmit() returned error: %v", err)
			}

			if result != nil {
				t.Errorf("handleUpdateTaskSubmit() should return nil for %q", cancelBtn)
			}
		})
	}
}

func TestHandleUpdateTaskSubmit_Update(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		button     string
		values     map[string]interface{}
		wantError  bool
		wantTaskID string
		wantUpdate string
		wantEmpty  bool
	}{
		{
			name:       "simple update",
			taskID:     "1",
			button:     "Update",
			values:     map[string]interface{}{"update": "Test update"},
			wantError:  false,
			wantTaskID: "1",
			wantUpdate: "Test update",
			wantEmpty:  false,
		},
		{
			name:       "multiline update",
			taskID:     "2.1",
			button:     "Update",
			values:     map[string]interface{}{"update": "Line 1\nLine 2\nLine 3"},
			wantError:  false,
			wantTaskID: "2.1",
			wantUpdate: "Line 1\nLine 2\nLine 3",
			wantEmpty:  false,
		},
		{
			name:       "empty update",
			taskID:     "3",
			button:     "Update",
			values:     map[string]interface{}{"update": ""},
			wantError:  false,
			wantTaskID: "3",
			wantUpdate: "",
			wantEmpty:  true,
		},
		{
			name:       "whitespace trimming",
			taskID:     "4",
			button:     "Update",
			values:     map[string]interface{}{"update": "  \n  Test  \n  "},
			wantError:  false,
			wantTaskID: "4",
			wantUpdate: "Test",
			wantEmpty:  false,
		},
		{
			name:       "whitespace only",
			taskID:     "4b",
			button:     "Update",
			values:     map[string]interface{}{"update": "   \t\n  \t "},
			wantError:  false,
			wantTaskID: "4b",
			wantUpdate: "",
			wantEmpty:  true,
		},
		{
			name:       "missing update key",
			taskID:     "5",
			button:     "Update",
			values:     map[string]interface{}{},
			wantError:  false,
			wantTaskID: "5",
			wantUpdate: "",
			wantEmpty:  true,
		},
		{
			name:       "invalid update type",
			taskID:     "6",
			button:     "Update",
			values:     map[string]interface{}{"update": 123},
			wantError:  false,
			wantTaskID: "6",
			wantUpdate: "",
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleUpdateTaskSubmit(tt.taskID, tt.button, tt.values)

			if (err != nil) != tt.wantError {
				t.Errorf("handleUpdateTaskSubmit() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if result == nil {
				t.Errorf("handleUpdateTaskSubmit() returned nil result")
				return
			}

			utr, ok := result.(UpdateTaskResult)
			if !ok {
				t.Errorf("handleUpdateTaskSubmit() result type is %T, want UpdateTaskResult", result)
				return
			}

			if utr.TaskID != tt.wantTaskID {
				t.Errorf("TaskID = %q, want %q", utr.TaskID, tt.wantTaskID)
			}

			if utr.Update != tt.wantUpdate {
				t.Errorf("Update = %q, want %q", utr.Update, tt.wantUpdate)
			}

			if utr.IsEmpty != tt.wantEmpty {
				t.Errorf("IsEmpty = %v, want %v", utr.IsEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestUpdateTaskResult_String(t *testing.T) {
	tests := []struct {
		name     string
		taskID   string
		update   string
		isEmpty  bool
	}{
		{
			name:    "normal update result",
			taskID:  "1.2",
			update:  "Test update",
			isEmpty: false,
		},
		{
			name:    "empty update result",
			taskID:  "2.1",
			update:  "",
			isEmpty: true,
		},
		{
			name:    "multiline update",
			taskID:  "3",
			update:  "Line 1\nLine 2",
			isEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpdateTaskResult{
				TaskID:  tt.taskID,
				Update:  tt.update,
				IsEmpty: tt.isEmpty,
			}

			if result.TaskID != tt.taskID {
				t.Errorf("TaskID = %q, want %q", result.TaskID, tt.taskID)
			}

			if result.Update != tt.update {
				t.Errorf("Update = %q, want %q", result.Update, tt.update)
			}

			if result.IsEmpty != tt.isEmpty {
				t.Errorf("IsEmpty = %v, want %v", result.IsEmpty, tt.isEmpty)
			}
		})
	}
}

func TestHandleUpdateTaskSubmit_DifferentButtons(t *testing.T) {
	taskID := "1"
	values := map[string]interface{}{"update": "Some text"}

	tests := []struct {
		button string
		wantNil bool
	}{
		{"Update", false},
		{"Cancel", true},
		{"Submit", false},
		{"Confirm", false},
	}

	for _, tt := range tests {
		t.Run(tt.button, func(t *testing.T) {
			result, _ := handleUpdateTaskSubmit(taskID, tt.button, values)

			isNil := result == nil
			if isNil != tt.wantNil {
				t.Errorf("For button %q: got nil=%v, want nil=%v", tt.button, isNil, tt.wantNil)
			}
		})
	}
}

func TestNewUpdateTaskDialog_WithTaskIDFormatting(t *testing.T) {
	tests := []struct {
		taskID          string
		expectedInTitle string
	}{
		{"1", "[1]"},
		{"1.1", "[1.1]"},
		{"2.3.4", "[2.3.4]"},
		{"999", "[999]"},
	}

	for _, tt := range tests {
		t.Run(tt.taskID, func(t *testing.T) {
			dialog := NewUpdateTaskDialog(tt.taskID, nil)
			title := dialog.Title()
			
			// Check if title contains expected substring
			found := false
			for i := 0; i < len(title)-len(tt.expectedInTitle)+1; i++ {
				if title[i:i+len(tt.expectedInTitle)] == tt.expectedInTitle {
					found = true
					break
				}
			}
			
			if !found {
				t.Errorf("Title() = %q, expected to contain %q", title, tt.expectedInTitle)
			}
		})
	}
}
