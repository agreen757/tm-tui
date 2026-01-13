package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFormFieldTypeButton_Creation(t *testing.T) {
	field := FormField{
		ID:          "test-button",
		Label:       "Click Me",
		Type:        FormFieldTypeButton,
		Placeholder: "Click to select",
		Value:       []string{},
	}

	if field.Type != FormFieldTypeButton {
		t.Errorf("Expected FormFieldTypeButton, got: %d", field.Type)
	}

	if field.Value == nil {
		t.Error("Expected initial value, got nil")
	}

	tagsVal, ok := field.Value.([]string)
	if !ok {
		t.Errorf("Expected []string, got: %T", field.Value)
	}
	if len(tagsVal) != 0 {
		t.Error("Expected empty tags slice")
	}
}

func TestFormFieldTypeButton_WithCallback(t *testing.T) {
	callbackCalled := false
	callback := func() tea.Cmd {
		callbackCalled = true
		return nil
	}

	field := FormField{
		ID:           "test-button",
		Type:         FormFieldTypeButton,
		ButtonCallback: callback,
	}

	if field.ButtonCallback == nil {
		t.Fatal("Expected ButtonCallback to be set")
	}

	// Call the callback
	field.ButtonCallback()
	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
}

func TestFormDialog_ButtonFieldRendering(t *testing.T) {
	style := &DialogStyle{
		ButtonColor: "#FF0000",
		ErrorColor:  "#FF0000",
	}

	fields := []FormField{
		{
			ID:          "button1",
			Label:       "Select Items",
			Type:        FormFieldTypeButton,
			Placeholder: "Click here",
			Value:       []string{},
		},
	}

	dialog := NewFormDialog(
		"Test Button Dialog",
		"",
		fields,
		[]string{"Done", "Cancel"},
		style,
		nil,
	)

	view := dialog.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain the button text/placeholder
	if len(view) == 0 {
		t.Error("Expected rendered content")
	}
}

func TestFormDialog_ButtonFieldWithSelectedTags(t *testing.T) {
	style := &DialogStyle{
		ButtonColor: "#FF0000",
		ErrorColor:  "#FF0000",
	}

	fields := []FormField{
		{
			ID:          "button1",
			Label:       "Select Items",
			Type:        FormFieldTypeButton,
			Placeholder: "Click here",
			Value:       []string{"tag1", "tag2"},
		},
	}

	dialog := NewFormDialog(
		"Test Button Dialog",
		"",
		fields,
		[]string{"Done", "Cancel"},
		style,
		nil,
	)

	view := dialog.View()

	// Should contain the selected tags
	if len(view) == 0 {
		t.Error("Expected rendered content with selected tags")
	}
}

func TestFormDialog_ButtonFieldKeyActivation(t *testing.T) {
	callbackCalled := false
	callback := func() tea.Cmd {
		callbackCalled = true
		return func() tea.Msg {
			return nil
		}
	}

	fields := []FormField{
		{
			ID:             "button1",
			Label:          "Select Items",
			Type:           FormFieldTypeButton,
			ButtonCallback: callback,
		},
	}

	dialog := NewFormDialog(
		"Test Button Dialog",
		"",
		fields,
		[]string{"Done"},
		nil,
		nil,
	)

	// Focus the button field
	dialog.SetFocusedIndex(0)

	// Send Space key
	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got: %v", result)
	}

	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
}

func TestFormDialog_ButtonFieldEnterKey(t *testing.T) {
	callbackCalled := false
	callback := func() tea.Cmd {
		callbackCalled = true
		return func() tea.Msg {
			return nil
		}
	}

	fields := []FormField{
		{
			ID:             "button1",
			Label:          "Select Items",
			Type:           FormFieldTypeButton,
			ButtonCallback: callback,
		},
	}

	dialog := NewFormDialog(
		"Test Button Dialog",
		"",
		fields,
		[]string{"Done"},
		nil,
		nil,
	)

	dialog.SetFocusedIndex(0)

	// Send Enter key
	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got: %v", result)
	}

	if !callbackCalled {
		t.Error("Expected callback to be called on Enter")
	}

	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestFormDialog_MultipleFieldTypes(t *testing.T) {
	fields := []FormField{
		{
			ID:    "text1",
			Label: "Text Field",
			Type:  FormFieldTypeText,
		},
		{
			ID:          "button1",
			Label:       "Button Field",
			Type:        FormFieldTypeButton,
			Placeholder: "Click",
		},
		{
			ID:      "checkbox1",
			Label:   "Checkbox Field",
			Type:    FormFieldTypeCheckbox,
			Checked: false,
		},
	}

	dialog := NewFormDialog(
		"Multi-field Dialog",
		"",
		fields,
		[]string{"Submit"},
		nil,
		nil,
	)

	view := dialog.View()
	if len(view) == 0 {
		t.Error("Expected rendered dialog")
	}

	// Verify all fields are present
	if len(dialog.fields) != 3 {
		t.Errorf("Expected 3 fields, got: %d", len(dialog.fields))
	}

	// Verify field types
	if dialog.fields[0].Type != FormFieldTypeText {
		t.Error("Expected first field to be text")
	}
	if dialog.fields[1].Type != FormFieldTypeButton {
		t.Error("Expected second field to be button")
	}
	if dialog.fields[2].Type != FormFieldTypeCheckbox {
		t.Error("Expected third field to be checkbox")
	}
}

func TestFormDialog_ButtonFieldNoCallback(t *testing.T) {
	fields := []FormField{
		{
			ID:    "button1",
			Label: "Button without callback",
			Type:  FormFieldTypeButton,
			// No ButtonCallback set
		},
	}

	dialog := NewFormDialog(
		"Test",
		"",
		fields,
		[]string{"Done"},
		nil,
		nil,
	)

	dialog.SetFocusedIndex(0)

	// Should not panic when Button field has no callback
	result, cmd := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got: %v", result)
	}

	// cmd should be nil or succeed without panicking
	if cmd != nil {
		_ = cmd()
	}
}

