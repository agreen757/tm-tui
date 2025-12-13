package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormFieldType represents the supported input types.
type FormFieldType int

const (
	FormFieldTypeText FormFieldType = iota
	FormFieldTypeCheckbox
	FormFieldTypeRadio
)

// Backwards-compatible aliases for legacy demo/tests.
type FormFieldKind = FormFieldType

const (
	FormFieldText       FormFieldKind = FormFieldTypeText
	FormFieldCheckbox   FormFieldKind = FormFieldTypeCheckbox
	FormFieldRadioGroup FormFieldKind = FormFieldTypeRadio
)

// SubmittedFormField describes a field included in FormSubmitMsg for legacy demos.
type SubmittedFormField struct {
	Kind           FormFieldKind
	Label          string
	TextInput      textinput.Model
	Checked        bool
	Options        []string
	SelectedOption int
}

// FormSubmitMsg is emitted when a form dialog is submitted (for demo compatibility).
type FormSubmitMsg struct {
	DialogID string
	Fields   []SubmittedFormField
}

// FormOption represents a single choice for radio fields.
type FormOption struct {
	Value       string
	Label       string
	Description string
}

// FormField describes a row in the dialog form.
type FormField struct {
	ID              string
	Label           string
	Type            FormFieldType
	Required        bool
	Value           interface{}
	Checked         bool
	Options         []FormOption
	SelectedOption  int
	Placeholder     string
	ConditionalShow string
	Help            string

	ValidationError string

	input textinput.Model
}

// Helper constructors retained for backwards compatibility.
func NewTextField(label, placeholder string, required bool) FormField {
	return FormField{
		ID:          fieldIDFromLabel(label),
		Label:       label,
		Type:        FormFieldTypeText,
		Required:    required,
		Placeholder: placeholder,
	}
}

func NewCheckboxField(label string, initialValue bool) FormField {
	return FormField{
		ID:       fieldIDFromLabel(label),
		Label:    label,
		Type:     FormFieldTypeCheckbox,
		Value:    initialValue,
		Checked:  initialValue,
		Required: false,
	}
}

func NewRadioGroupField(label string, options []string, selected int) FormField {
	opts := make([]FormOption, len(options))
	for i, option := range options {
		opts[i] = FormOption{Value: option, Label: option}
	}
	field := FormField{
		ID:       fieldIDFromLabel(label),
		Label:    label,
		Type:     FormFieldTypeRadio,
		Options:  opts,
		Required: false,
	}
	if len(opts) > 0 {
		idx := selected
		if idx < 0 || idx >= len(opts) {
			idx = 0
		}
		field.Value = opts[idx].Value
		field.SelectedOption = idx
	}
	return field
}

// FormSubmitHandler handles form submissions.
type FormSubmitHandler func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error)

// FormValidator defines a custom validation hook.
type FormValidator func(values map[string]interface{}) error

// FormEventHandler receives intermediate input events.
type FormEventHandler func(form *FormDialog, msg tea.Msg)

// ErrorFormValidation represents a validation failure tied to a field.
type ErrorFormValidation struct {
	FieldID string
	Message string
}

func (e ErrorFormValidation) Error() string {
	if e.FieldID == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.FieldID, e.Message)
}

// FormValueChangedMsg is emitted when a field value changes.
type FormValueChangedMsg struct {
	FieldID  string
	NewValue interface{}
}

// FormDialog is an interactive dialog with labeled inputs and buttons.
type FormDialog struct {
	BaseFocusableDialog

	Description string
	fields      []FormField
	buttons     []string
	handler     FormSubmitHandler
	validators  []FormValidator
	events      []FormEventHandler

	selectedButton int
	resultValue    interface{}
	resultErr      error
	lastError      *ErrorFormValidation
	submitted      bool
}

// NewFormDialog builds a new form dialog with the supplied configuration.
func NewFormDialog(
	title string,
	description string,
	fields []FormField,
	buttons []string,
	style *DialogStyle,
	handler FormSubmitHandler,
) *FormDialog {
	if len(buttons) == 0 {
		buttons = []string{"Submit", "Cancel"}
	}

	width := 64
	height := 6 + len(fields)*2 + len(buttons)

	dialog := &FormDialog{
		BaseFocusableDialog: NewBaseFocusableDialog(title, width, height, DialogKindForm, len(fields)+len(buttons)),
		Description:         description,
		fields:              make([]FormField, len(fields)),
		buttons:             append([]string{}, buttons...),
		handler:             handler,
	}

	if style != nil {
		dialog.Style = style
	}

	dialog.SetFooterHints(
		ShortcutHint{Key: "Tab", Label: "Next Field"},
		ShortcutHint{Key: "Shift+Tab", Label: "Previous"},
		ShortcutHint{Key: "Enter", Label: "Submit"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	autoFocused := false
	for i, field := range fields {
		if field.Type == FormFieldTypeText {
			input := textinput.New()
			input.Placeholder = field.Placeholder
			input.Prompt = ""
			input.CharLimit = 256
			if value, ok := field.Value.(string); ok {
				input.SetValue(value)
			}
			if !autoFocused {
				input.Focus()
				autoFocused = true
			}
			field.input = input
		}
		if field.Type == FormFieldTypeCheckbox {
			field.Checked = field.Value == true
		}
		dialog.fields[i] = field
	}

	if len(fields) > 0 {
		dialog.SetFocusedIndex(0)
	}

	return dialog
}

// NewLegacyFormDialog preserves the original constructor signature for legacy callers.
func NewLegacyFormDialog(title string, width, height int, fields []FormField) *FormDialog {
	dialog := NewFormDialog(title, "", fields, nil, nil, nil)
	dialog.SetRect(width, height, 0, 0)
	return dialog
}

// Init focuses the first text field if one exists.
func (d *FormDialog) Init() tea.Cmd {
	for i := range d.fields {
		if d.fields[i].Type == FormFieldTypeText {
			return d.fields[i].input.Focus()
		}
	}
	return nil
}

// Update propagates messages to specialized handlers.
func (d *FormDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(m.Width, m.Height)
	}
	for _, handler := range d.events {
		handler(d, msg)
	}
	return d, nil
}

// HandleKey routes keyboard input to the focused element.
func (d *FormDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	if result, cmd := d.HandleBaseFocusableKey(msg); result != DialogResultNone {
		return result, cmd
	}

	focused := d.FocusedIndex()
	if focused < len(d.fields) {
		return d.handleFieldKey(focused, msg)
	}

	return d.handleButtonKey(msg)
}

func (d *FormDialog) handleFieldKey(index int, msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	field := &d.fields[index]
	switch field.Type {
	case FormFieldTypeText:
		var cmd tea.Cmd
		field.input, cmd = field.input.Update(msg)
		d.emitValueChanged(field.ID, strings.TrimSpace(field.input.Value()))
		return DialogResultNone, cmd
	case FormFieldTypeCheckbox:
		if key.Matches(msg, key.NewBinding(key.WithKeys(" ", "space", "enter"))) {
			field.Checked = !field.Checked
			field.Value = field.Checked
			d.emitValueChanged(field.ID, field.Checked)
		}
	case FormFieldTypeRadio:
		if len(field.Options) == 0 {
			return DialogResultNone, nil
		}
		current := d.currentOptionIndex(field)
		switch msg.String() {
		case "left", "up":
			current = (current - 1 + len(field.Options)) % len(field.Options)
		case "right", "down":
			current = (current + 1) % len(field.Options)
		case " ", "enter":
			// pressing enter on radio moves to buttons
			d.selectedButton = 0
			d.SetFocusedIndex(len(d.fields))
			return DialogResultNone, nil
		}
		field.Value = field.Options[current].Value
		field.SelectedOption = current
		d.emitValueChanged(field.ID, field.Value)
	}

	return DialogResultNone, nil
}

func (d *FormDialog) handleButtonKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch msg.String() {
	case "left", "up":
		if len(d.buttons) == 0 {
			return DialogResultNone, nil
		}
		d.selectedButton = (d.selectedButton - 1 + len(d.buttons)) % len(d.buttons)
		d.SetFocusedIndex(len(d.fields) + d.selectedButton)
	case "right", "down":
		if len(d.buttons) == 0 {
			return DialogResultNone, nil
		}
		d.selectedButton = (d.selectedButton + 1) % len(d.buttons)
		d.SetFocusedIndex(len(d.fields) + d.selectedButton)
	case "enter":
		if d.selectedButton < len(d.buttons) {
			button := d.buttons[d.selectedButton]
			return d.submit(button)
		}
	}

	return DialogResultNone, nil
}

func (d *FormDialog) submit(button string) (DialogResult, tea.Cmd) {
	values := d.collectValues()

	if err := d.applyValidators(values); err != nil {
		if vf, ok := err.(ErrorFormValidation); ok {
			d.lastError = &vf
		} else {
			d.lastError = &ErrorFormValidation{FieldID: "", Message: err.Error()}
		}
		return DialogResultNone, nil
	}

	d.lastError = nil

	if d.handler != nil {
		result, err := d.handler(d, button, values)
		if err != nil {
			d.lastError = &ErrorFormValidation{FieldID: "", Message: err.Error()}
			return DialogResultNone, nil
		}
		d.resultValue = result
		d.resultErr = nil
	} else {
		d.resultValue = values
		d.resultErr = nil
	}

	if strings.EqualFold(button, "cancel") {
		d.submitted = false
		d.resultValue = nil
		return DialogResultClose, nil
	}

	d.submitted = true
	return DialogResultConfirm, d.buildFormSubmitMsg()
}

func (d *FormDialog) collectValues() map[string]interface{} {
	values := make(map[string]interface{})
	for i, field := range d.fields {
		key := field.ID
		if key == "" {
			key = fmt.Sprintf("field_%d", i)
		}
		switch field.Type {
		case FormFieldTypeText:
			values[key] = strings.TrimSpace(field.input.Value())
		case FormFieldTypeCheckbox:
			values[key] = field.Checked
		case FormFieldTypeRadio:
			values[key] = fmt.Sprintf("%v", field.Value)
		}
	}
	return values
}

func (d *FormDialog) applyValidators(values map[string]interface{}) error {
	for i, field := range d.fields {
		if !field.Required {
			continue
		}
		key := field.ID
		if key == "" {
			key = fmt.Sprintf("field_%d", i)
		}
		switch field.Type {
		case FormFieldTypeText:
			if val, _ := values[key].(string); strings.TrimSpace(val) == "" {
				return ErrorFormValidation{FieldID: field.ID, Message: "This field is required"}
			}
		case FormFieldTypeCheckbox:
			if !field.Checked {
				return ErrorFormValidation{FieldID: field.ID, Message: "This field must be selected"}
			}
		case FormFieldTypeRadio:
			if val, _ := values[key].(string); strings.TrimSpace(val) == "" {
				return ErrorFormValidation{FieldID: field.ID, Message: "Select an option"}
			}
		}
	}

	for _, validator := range d.validators {
		if err := validator(values); err != nil {
			return err
		}
	}

	return nil
}

func (d *FormDialog) currentOptionIndex(field *FormField) int {
	current := fmt.Sprintf("%v", field.Value)
	for i, option := range field.Options {
		if option.Value == current {
			field.SelectedOption = i
			return i
		}
	}
	if len(field.Options) > 0 {
		field.Value = field.Options[0].Value
		field.SelectedOption = 0
	}
	return 0
}

// View renders the dialog contents.
func (d *FormDialog) View() string {
	var rows []string
	if strings.TrimSpace(d.Description) != "" {
		rows = append(rows, lipgloss.NewStyle().PaddingBottom(1).Render(d.Description))
	}

	for i := range d.fields {
		rows = append(rows, d.renderField(i))
	}

	rows = append(rows, d.renderButtons())

	if d.lastError != nil {
		errStyle := lipgloss.NewStyle().Foreground(d.Style.ErrorColor).PaddingTop(1)
		rows = append(rows, errStyle.Render(d.lastError.Message))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return d.RenderBorder(content)
}

func (d *FormDialog) renderField(index int) string {
	field := d.fields[index]
	label := lipgloss.NewStyle().Bold(true).Render(field.Label)
	value := ""

	switch field.Type {
	case FormFieldTypeText:
		value = field.input.View()
	case FormFieldTypeCheckbox:
		marker := "[ ]"
		if field.Value == true {
			marker = "[x]"
		}
		value = marker
	case FormFieldTypeRadio:
		var rendered []string
		current := fmt.Sprintf("%v", field.Value)
		for _, option := range field.Options {
			prefix := "( )"
			if option.Value == current {
				prefix = "(x)"
			}
			rendered = append(rendered, fmt.Sprintf("%s %s", prefix, option.Label))
		}
		value = strings.Join(rendered, "  ")
	}

	body := lipgloss.JoinHorizontal(lipgloss.Left, label, " ", value)
	if field.Help != "" {
		help := lipgloss.NewStyle().Foreground(lipgloss.Color("#7d7d7d")).Render(field.Help)
		body = lipgloss.JoinVertical(lipgloss.Left, body, help)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(body)
}

func (d *FormDialog) renderButtons() string {
	if len(d.buttons) == 0 {
		return ""
	}

	var rendered []string
	for i, button := range d.buttons {
		style := lipgloss.NewStyle().Padding(0, 2).BorderStyle(lipgloss.RoundedBorder())
		if i == d.selectedButton {
			style = style.Background(d.Style.ButtonColor).Foreground(lipgloss.Color("#000000"))
		}
		rendered = append(rendered, style.Render(button))
	}

	return lipgloss.NewStyle().PaddingTop(1).Render(lipgloss.JoinHorizontal(lipgloss.Left, rendered...))
}

// DialogResultValue implements DialogResultProvider.
func (d *FormDialog) DialogResultValue() (interface{}, error) {
	return d.resultValue, d.resultErr
}

// AddValidator appends a validator to the dialog.
func (d *FormDialog) AddValidator(validator FormValidator) {
	d.validators = append(d.validators, validator)
}

// AddEventHandler registers an additional event handler.
func (d *FormDialog) AddEventHandler(handler FormEventHandler) {
	d.events = append(d.events, handler)
}

// GetField returns a form field by ID.
func (d *FormDialog) GetField(id string) (*FormField, bool) {
	for i := range d.fields {
		if d.fields[i].ID == id {
			return &d.fields[i], true
		}
	}
	return nil, false
}

func (d *FormDialog) emitValueChanged(id string, value interface{}) {
	msg := FormValueChangedMsg{FieldID: id, NewValue: value}
	for _, handler := range d.events {
		handler(d, msg)
	}
}

func (d *FormDialog) buildFormSubmitMsg() tea.Cmd {
	fields := d.snapshotSubmittedFields()
	if len(fields) == 0 {
		return nil
	}

	dialogID := d.ID
	if dialogID == "" {
		dialogID = d.Title()
	}

	return func() tea.Msg {
		return FormSubmitMsg{
			DialogID: dialogID,
			Fields:   fields,
		}
	}
}

func (d *FormDialog) snapshotSubmittedFields() []SubmittedFormField {
	snapshot := make([]SubmittedFormField, len(d.fields))
	for i := range d.fields {
		snapshot[i] = exportSubmittedField(&d.fields[i])
	}
	return snapshot
}

func exportSubmittedField(field *FormField) SubmittedFormField {
	result := SubmittedFormField{
		Kind:           FormFieldKind(field.Type),
		Label:          field.Label,
		SelectedOption: field.SelectedOption,
	}

	switch field.Type {
	case FormFieldTypeText:
		result.TextInput = field.input
	case FormFieldTypeCheckbox:
		result.Checked = field.Checked
	case FormFieldTypeRadio:
		options := make([]string, len(field.Options))
		for i, option := range field.Options {
			if option.Label != "" {
				options[i] = option.Label
			} else {
				options[i] = option.Value
			}
		}
		result.Options = options
	}

	return result
}

// IsSubmitted indicates if the dialog was confirmed via a non-cancel button.
func (d *FormDialog) IsSubmitted() bool {
	return d.submitted
}

// SetSubmitLabel customizes the primary/confirm button label.
func (d *FormDialog) SetSubmitLabel(label string) {
	if label == "" {
		return
	}

	if len(d.buttons) == 0 {
		d.buttons = []string{label, "Cancel"}
		return
	}

	d.buttons[0] = label
}

// SetCancelLabel customizes the cancel button label, adding one if necessary.
func (d *FormDialog) SetCancelLabel(label string) {
	if label == "" {
		return
	}

	switch len(d.buttons) {
	case 0:
		d.buttons = []string{"Submit", label}
	case 1:
		d.buttons = append(d.buttons, label)
	default:
		d.buttons[1] = label
	}
}

// GetFieldValue returns the value for the field at the provided index.
func (d *FormDialog) GetFieldValue(index int) interface{} {
	if index < 0 || index >= len(d.fields) {
		return nil
	}
	return d.accessorValue(&d.fields[index])
}

// GetFieldValueByLabel returns the value for the field matching the label.
func (d *FormDialog) GetFieldValueByLabel(label string) interface{} {
	id := fieldIDFromLabel(label)
	field, ok := d.GetField(id)
	if !ok {
		return nil
	}
	return d.accessorValue(field)
}

func (d *FormDialog) accessorValue(field *FormField) interface{} {
	switch field.Type {
	case FormFieldTypeText:
		return strings.TrimSpace(field.input.Value())
	case FormFieldTypeCheckbox:
		return field.Checked
	case FormFieldTypeRadio:
		return field.SelectedOption
	default:
		return nil
	}
}

// ValidateField runs built-in validation for the field and records the result.
func (f *FormField) ValidateField() {
	f.ValidationError = ""
	if !f.Required {
		return
	}

	switch f.Type {
	case FormFieldTypeText:
		if strings.TrimSpace(f.input.Value()) == "" {
			f.ValidationError = "This field is required"
		}
	case FormFieldTypeCheckbox:
		if f.Value != true {
			f.ValidationError = "This field must be selected"
		}
	case FormFieldTypeRadio:
		if len(f.Options) == 0 {
			f.ValidationError = "No options available"
		} else if f.SelectedOption < 0 || f.SelectedOption >= len(f.Options) {
			f.ValidationError = "Select an option"
		}
	}
}

func fieldIDFromLabel(label string) string {
	cleaned := strings.ToLower(label)
	cleaned = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == ' ' || r == '-':
			return '_'
		default:
			return -1
		}
	}, cleaned)
	cleaned = strings.Trim(cleaned, "_")
	if cleaned == "" {
		cleaned = "field"
	}
	return cleaned
}
