package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	deleteConfirmDialogID = "delete_confirm_dialog"
	deleteOptionsDialogID = "delete_options_dialog"
	deleteWarningDialogID = "delete_warning_dialog"
	undoDialogID          = "delete_undo_dialog"
)

// DeleteWorkflowState tracks the multi-step delete flow.
type DeleteWorkflowState struct {
	TaskIDs []string
	Options taskmaster.DeleteOptions
	Impact  *taskmaster.DeleteImpact
}

// UndoSession captures countdown state for undo dialogs.
type UndoSession struct {
	Token     *taskmaster.UndoToken
	Remaining time.Duration
}

func (m *Model) startDeleteWorkflow(taskIDs []string) {
	if len(taskIDs) == 0 {
		m.showErrorDialog("Delete Task", "Select at least one task before deleting.")
		return
	}
	m.deleteWorkflow = &DeleteWorkflowState{TaskIDs: taskIDs}
	m.openDeleteConfirmation(len(taskIDs))
}

func (m *Model) openDeleteConfirmation(count int) {
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	message := fmt.Sprintf("Delete %d selected task(s)? You will be able to review impacted tasks before confirming.", count)
	content := dialog.NewSimpleModalContent(message)
	buttons := []dialog.ModalButton{
		{
			Kind:  dialog.ButtonYes,
			Label: "Continue",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultConfirm, func() tea.Msg {
					return dialog.DialogResultMsg{ID: deleteConfirmDialogID, Button: "continue"}
				}
			},
		},
		{
			Kind:  dialog.ButtonCancel,
			Label: "Cancel",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultCancel, func() tea.Msg {
					return dialog.DialogResultMsg{ID: deleteConfirmDialogID, Button: "cancel"}
				}
			},
		},
	}
	dlg := dialog.NewButtonModalDialog("Delete Tasks", 70, 9, content, buttons)
	dlg.ModalDialog.BaseDialog.ID = deleteConfirmDialogID
	m.appState.AddDialog(dlg, nil)
}

func (m *Model) openDeleteOptions() {
	dm := m.dialogManager()
	if dm == nil || m.deleteWorkflow == nil {
		return
	}
	fields := []dialog.FormField{
		{
			ID:      "recursive",
			Label:   "Recursive (delete subtasks and dependents)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: m.deleteWorkflow.Options.Recursive,
			Value:   m.deleteWorkflow.Options.Recursive,
		},
		{
			ID:      "force",
			Label:   "Force (skip minor validation errors)",
			Type:    dialog.FormFieldTypeCheckbox,
			Checked: m.deleteWorkflow.Options.Force,
			Value:   m.deleteWorkflow.Options.Force,
		},
	}
	form := dialog.NewFormDialog(
		"Delete Options",
		"Choose how deletion should behave before continuing.",
		fields,
		[]string{"Review Impact", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Review Impact" {
				return nil, nil
			}
			return taskmaster.DeleteOptions{
				Recursive: boolValue(values, "recursive"),
				Force:     boolValue(values, "force"),
			}, nil
		},
	)
	form.BaseFocusableDialog.BaseDialog.ID = deleteOptionsDialogID
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Delete Options", err.Error())
			return nil
		}
		if value == nil {
			if m.deleteWorkflow != nil && m.deleteWorkflow.Impact == nil {
				m.deleteWorkflow = nil
			}
			return nil
		}
		if opts, ok := value.(taskmaster.DeleteOptions); ok {
			m.deleteWorkflow.Options = opts
			m.evaluateDeleteImpact()
		}
		return nil
	})
}

func (m *Model) evaluateDeleteImpact() {
	if m.taskService == nil || m.deleteWorkflow == nil {
		return
	}
	impact, err := m.taskService.AnalyzeDeleteImpact(m.deleteWorkflow.TaskIDs, m.deleteWorkflow.Options)
	if err != nil {
		m.showErrorDialog("Delete Task", err.Error())
		return
	}
	m.deleteWorkflow.Impact = impact
	m.openDeleteWarningDialog()
}

func (m *Model) openDeleteWarningDialog() {
	dm := m.dialogManager()
	if dm == nil || m.deleteWorkflow == nil || m.deleteWorkflow.Impact == nil {
		return
	}
	impact := m.deleteWorkflow.Impact
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Tasks to delete: %d\n", impact.TotalDeleteCount))
	if impact.BlockingReason != "" {
		b.WriteString(fmt.Sprintf("Blocking issue: %s\n", impact.BlockingReason))
	}
	if len(impact.WarningMessages) > 0 {
		for _, msg := range impact.WarningMessages {
			b.WriteString("- " + msg + "\n")
		}
	}
	content := dialog.NewSimpleModalContent(strings.TrimSpace(b.String()))
	buttons := []dialog.ModalButton{}
	if impact.BlockingReason != "" {
		buttons = append(buttons,
			dialog.ModalButton{
				Kind:  dialog.ButtonYes,
				Label: "Adjust Options",
				OnClick: func() (dialog.DialogResult, tea.Cmd) {
					return dialog.DialogResultConfirm, func() tea.Msg {
						return dialog.DialogResultMsg{ID: deleteWarningDialogID, Button: "adjust"}
					}
				},
			},
		)
	} else {
		buttons = append(buttons,
			dialog.ModalButton{
				Kind:  dialog.ButtonYes,
				Label: "Delete",
				OnClick: func() (dialog.DialogResult, tea.Cmd) {
					return dialog.DialogResultConfirm, func() tea.Msg {
						return dialog.DialogResultMsg{ID: deleteWarningDialogID, Button: "confirm"}
					}
				},
			},
		)
	}
	buttons = append(buttons, dialog.ModalButton{
		Kind:  dialog.ButtonCancel,
		Label: "Cancel",
		OnClick: func() (dialog.DialogResult, tea.Cmd) {
			return dialog.DialogResultCancel, func() tea.Msg {
				return dialog.DialogResultMsg{ID: deleteWarningDialogID, Button: "cancel"}
			}
		},
	})
	dlg := dialog.NewButtonModalDialog("Review Impact", 74, 12, content, buttons)
	dlg.ModalDialog.BaseDialog.ID = deleteWarningDialogID
	m.appState.AddDialog(dlg, nil)
}

func (m *Model) performDelete() tea.Cmd {
	if m.taskService == nil || m.deleteWorkflow == nil {
		return nil
	}
	ctx := context.Background()
	result, err := m.taskService.DeleteTasks(ctx, m.deleteWorkflow.TaskIDs, m.deleteWorkflow.Options)
	if err != nil {
		m.showErrorDialog("Delete Task", err.Error())
		return nil
	}
	m.addLogLine(fmt.Sprintf("Deleted %d task(s)", result.DeletedCount))
	for _, warning := range result.Warnings {
		m.addLogLine("Warning: " + warning)
	}
	m.deleteWorkflow = nil
	if result.Undo != nil {
		return m.showUndoDialog(result.Undo)
	}
	return nil
}

func (m *Model) showUndoDialog(token *taskmaster.UndoToken) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil || token == nil {
		m.undoSession = nil
		return nil
	}
	content := newUndoContent(token)
	buttons := []dialog.ModalButton{
		{
			Kind:  dialog.ButtonYes,
			Label: "Undo",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultConfirm, func() tea.Msg {
					return dialog.DialogResultMsg{ID: undoDialogID, Button: "undo", Value: token}
				}
			},
		},
		{
			Kind:  dialog.ButtonCancel,
			Label: "Dismiss",
			OnClick: func() (dialog.DialogResult, tea.Cmd) {
				return dialog.DialogResultCancel, func() tea.Msg {
					return dialog.DialogResultMsg{ID: undoDialogID, Button: "dismiss"}
				}
			},
		},
	}
	dlg := dialog.NewButtonModalDialog("Undo Delete", 68, 9, content, buttons)
	dlg.ModalDialog.BaseDialog.ID = undoDialogID
	m.appState.AddDialog(dlg, nil)
	m.undoSession = &UndoSession{Token: token, Remaining: token.Duration}
	return StartUndoCountdown(token.ID, token.ExpiresAt)
}

// undoContent renders live countdown text inside the undo dialog.
type undoContent struct {
	actionID  string
	summary   string
	remaining time.Duration
}

func newUndoContent(token *taskmaster.UndoToken) *undoContent {
	remaining := token.Duration
	if remaining <= 0 {
		remaining = time.Until(token.ExpiresAt)
	}
	return &undoContent{
		actionID:  token.ID,
		summary:   token.Summary,
		remaining: remaining,
	}
}

func (c *undoContent) Init() tea.Cmd { return nil }

func (c *undoContent) Update(msg tea.Msg) (dialog.ModalContent, tea.Cmd) {
	switch t := msg.(type) {
	case UndoTickMsg:
		if t.ActionID == c.actionID {
			c.remaining = t.Remaining
		}
	case UndoExpiredMsg:
		if t.ActionID == c.actionID {
			c.remaining = 0
		}
	}
	return c, nil
}

func (c *undoContent) View(width, height int) string {
	secs := int(c.remaining.Round(time.Second) / time.Second)
	if secs < 0 {
		secs = 0
	}
	text := fmt.Sprintf("%s\nUndo available for %d seconds", c.summary, secs)
	style := lipgloss.NewStyle().Width(width).Height(height).Align(lipgloss.Left)
	return style.Render(text)
}

func (c *undoContent) HandleKey(tea.KeyMsg) tea.Cmd { return nil }

func (m *Model) handleDialogResultMsg(msg dialog.DialogResultMsg) tea.Cmd {
	switch msg.ID {
	case deleteConfirmDialogID:
		if msg.Button == "continue" {
			m.openDeleteOptions()
			return nil
		}
		m.deleteWorkflow = nil
	case deleteWarningDialogID:
		switch msg.Button {
		case "confirm":
			return m.performDelete()
		case "adjust":
			m.openDeleteOptions()
			return nil
		default:
			m.deleteWorkflow = nil
		}
	case undoDialogID:
		if msg.Button == "undo" {
			if token, ok := msg.Value.(*taskmaster.UndoToken); ok {
				return m.executeUndo(token.ID)
			}
			return m.executeUndo("")
		}
		m.undoSession = nil
	}
	return nil
}

func (m *Model) executeUndo(actionID string) tea.Cmd {
	if actionID == "" || m.taskService == nil {
		m.showErrorDialog("Undo Delete", "Undo action is no longer available.")
		return nil
	}
	if err := m.taskService.UndoAction(context.Background(), actionID); err != nil {
		m.showErrorDialog("Undo Delete", err.Error())
	} else {
		m.addLogLine("Undo completed successfully")
	}
	m.undoSession = nil
	return nil
}

func (m *Model) handleUndoTick(msg UndoTickMsg) tea.Cmd {
	if m.undoSession == nil || m.undoSession.Token == nil || m.undoSession.Token.ID != msg.ActionID {
		return nil
	}
	m.undoSession.Remaining = msg.Remaining
	return StartUndoCountdown(msg.ActionID, m.undoSession.Token.ExpiresAt)
}

func (m *Model) handleUndoExpired(msg UndoExpiredMsg) {
	if m.undoSession == nil || m.undoSession.Token == nil || m.undoSession.Token.ID != msg.ActionID {
		return
	}
	m.addLogLine("Undo window expired")
	m.undoSession = nil
	if m.appState != nil {
		m.appState.PopDialog()
	}
}
