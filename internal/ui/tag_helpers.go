package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	tagListDialogID   = "tag_context_list_dialog"
	tagActionDialogID = "tag_context_actions_dialog"
)

// GetVisibleTasks returns the currently visible tasks
func (m *Model) GetVisibleTasks() []taskmaster.Task {
	// Convert pointers to actual tasks
	tasks := make([]taskmaster.Task, 0, len(m.visibleTasks))
	for _, task := range m.visibleTasks {
		if task != nil {
			tasks = append(tasks, *task)
		}
	}
	return tasks
}

// tagListItem adapts a TagContext to the dialog list interface.
type tagListItem struct {
	ctx *taskmaster.TagContext
}

func newTagListItem(ctx taskmaster.TagContext) *tagListItem {
	copy := ctx
	return &tagListItem{ctx: &copy}
}

func (i *tagListItem) Title() string {
	if i.ctx == nil {
		return ""
	}
	prefix := "  "
	if i.ctx.Active {
		prefix = "● "
	}
	return fmt.Sprintf("%s%s", prefix, i.ctx.Name)
}

func (i *tagListItem) Description() string {
	if i.ctx == nil {
		return ""
	}
	segments := []string{
		fmt.Sprintf("%d task(s)", i.ctx.TaskCount),
		fmt.Sprintf("%d done", i.ctx.CompletedCount),
	}
	if i.ctx.CreatedLabel != "" {
		segments = append(segments, fmt.Sprintf("created %s", i.ctx.CreatedLabel))
	}
	if i.ctx.Description != "" {
		segments = append(segments, i.ctx.Description)
	}
	return strings.Join(segments, " • ")
}

func (i *tagListItem) FilterValue() string {
	if i.ctx == nil {
		return ""
	}
	return i.ctx.Name
}

type tagActionKind string

const (
	tagActionUse    tagActionKind = "use"
	tagActionRename tagActionKind = "rename"
	tagActionCopy   tagActionKind = "copy"
	tagActionDelete tagActionKind = "delete"
)

type tagActionItem struct {
	label       string
	description string
	kind        tagActionKind
}

func newTagActionItem(kind tagActionKind, label, description string) *tagActionItem {
	return &tagActionItem{
		label:       label,
		description: description,
		kind:        kind,
	}
}

func (i *tagActionItem) Title() string {
	return i.label
}

func (i *tagActionItem) Description() string {
	return i.description
}

func (i *tagActionItem) FilterValue() string {
	return i.label
}

func (m *Model) showTagListDialog(list *taskmaster.TagList) {
	dm := m.dialogManager()
	if dm == nil || list == nil {
		return
	}

	if len(list.Tags) == 0 {
		m.showErrorDialog("Tag Contexts", "No tag contexts available. Use Add Tag to create one.")
		return
	}

	items := make([]dialog.ListItem, 0, len(list.Tags))
	for _, ctx := range list.Tags {
		items = append(items, newTagListItem(ctx))
	}

	width := 72
	height := 22
	if m.width > 0 {
		width = m.width - 20
		if width < 60 {
			width = 60
		}
	}
	if m.height > 0 {
		height = m.height - 10
		if height < 16 {
			height = 16
		}
	}

	dialogList := dialog.NewListDialog("Tag Contexts", width, height, items)
	dialogList.SetShowDescription(true)
	dialogList.BaseFocusableDialog.BaseDialog.ID = tagListDialogID
	dialogList.SetCancellable(true)

	m.appState.AddDialog(dialogList, nil)
}

func (m *Model) openTagActionDialog() {
	if m.tagActionContext == nil {
		return
	}
	dm := m.dialogManager()
	if dm == nil {
		return
	}

	items := []dialog.ListItem{
		newTagActionItem(tagActionUse, "Switch to this tag", "Set this tag context as active"),
		newTagActionItem(tagActionRename, "Rename tag", "Update the tag context name"),
		newTagActionItem(tagActionCopy, "Copy tag", "Duplicate tasks into a new tag"),
		newTagActionItem(tagActionDelete, "Delete tag", "Remove this tag context and its tasks"),
	}

	width := 60
	height := 16
	if m.width > 0 && m.height > 0 {
		width = m.width - 30
		if width < 50 {
			width = 50
		}
		height = m.height - 12
		if height < 14 {
			height = 14
		}
	}

	dlg := dialog.NewListDialog(fmt.Sprintf("Actions for %s", m.tagActionContext.Name), width, height, items)
	dlg.SetShowDescription(true)
	dlg.BaseFocusableDialog.BaseDialog.ID = tagActionDialogID
	m.appState.AddDialog(dlg, nil)
}

func (m *Model) openRenameTagDialog() {
	if m.tagActionContext == nil {
		return
	}
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	fields := []dialog.FormField{
		{
			ID:       "newName",
			Label:    "New tag name",
			Type:     dialog.FormFieldTypeText,
			Required: true,
		},
	}
	form := dialog.NewFormDialog(
		fmt.Sprintf("Rename %s", m.tagActionContext.Name),
		"Enter the new name for this tag context.",
		fields,
		[]string{"Rename", "Cancel"},
		dm.Style,
		func(_ *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Rename" {
				return nil, nil
			}
			name := strings.TrimSpace(stringValue(values, "newName"))
			if name == "" {
				return nil, fmt.Errorf("Tag name is required")
			}
			return name, nil
		},
	)
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Rename Tag", err.Error())
			return nil
		}
		newName, ok := value.(string)
		if !ok || newName == "" {
			return nil
		}
		current := m.tagActionContext.Name
		return m.runTagOperation("rename-tag", current, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.RenameTagContext(ctx, current, newName)
		})
	})
}

func (m *Model) openCopyTagDialog() {
	if m.tagActionContext == nil {
		return
	}
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	fields := []dialog.FormField{
		{
			ID:       "target",
			Label:    "New tag name",
			Type:     dialog.FormFieldTypeText,
			Required: true,
		},
		{
			ID:    "description",
			Label: "Description (optional)",
			Type:  dialog.FormFieldTypeText,
		},
	}
	form := dialog.NewFormDialog(
		fmt.Sprintf("Copy %s", m.tagActionContext.Name),
		"Duplicate tasks into a new tag context.",
		fields,
		[]string{"Copy", "Cancel"},
		dm.Style,
		func(_ *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Copy" {
				return nil, nil
			}
			name := strings.TrimSpace(stringValue(values, "target"))
			if name == "" {
				return nil, fmt.Errorf("Tag name is required")
			}
			description := strings.TrimSpace(stringValue(values, "description"))
			return map[string]string{"target": name, "description": description}, nil
		},
	)
	m.appState.AddDialog(form, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Copy Tag", err.Error())
			return nil
		}
		data, ok := value.(map[string]string)
		if !ok {
			return nil
		}
		target := data["target"]
		if target == "" {
			return nil
		}
		opts := taskmaster.TagCopyOptions{Description: data["description"]}
		source := m.tagActionContext.Name
		return m.runTagOperation("copy-tag", source, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.CopyTagContext(ctx, source, target, opts)
		})
	})
}

func (m *Model) openDeleteTagConfirmation() {
	if m.tagActionContext == nil {
		return
	}
	dm := m.dialogManager()
	if dm == nil {
		return
	}
	message := fmt.Sprintf("Delete tag '%s' and all of its tasks?", m.tagActionContext.Name)
	confirm := dialog.YesNo("Delete Tag Context", message, true)
	m.appState.AddDialog(confirm, func(value interface{}, err error) tea.Cmd {
		_ = value
		_ = err
		if confirm.Result() != dialog.ConfirmationResultYes {
			return nil
		}
		name := m.tagActionContext.Name
		return m.runTagOperation("delete-tag", name, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.DeleteTagContext(ctx, name, true)
		})
	})
}

func (m *Model) handleTagActionSelection(kind tagActionKind) tea.Cmd {
	if m.tagActionContext == nil || m.taskService == nil {
		m.showErrorDialog("Tag Actions", "No tag context selected.")
		return nil
	}
	switch kind {
	case tagActionUse:
		name := m.tagActionContext.Name
		return m.runTagOperation("use-tag", name, func(ctx context.Context) (*taskmaster.TagOperationResult, error) {
			return m.taskService.UseTagContext(ctx, name)
		})
	case tagActionRename:
		m.openRenameTagDialog()
	case tagActionCopy:
		m.openCopyTagDialog()
	case tagActionDelete:
		m.openDeleteTagConfirmation()
	}
	return nil
}
