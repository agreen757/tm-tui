package ui

import (
	"fmt"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandID represents an action that can be triggered from shortcuts or the palette.
type CommandID string

const (
	CommandParsePRD           CommandID = "parse_prd"
	CommandAnalyzeComplexity  CommandID = "analyze_complexity"
	CommandExpandTask         CommandID = "expand_task"
	CommandDeleteTask         CommandID = "delete_task"
	CommandManageTags         CommandID = "manage_tags"
	CommandTagManagement      CommandID = "tag_management"
	CommandUseTag             CommandID = "use_tag"
	CommandProjectTags        CommandID = "project_tags"
	CommandProjectQuickSwitch CommandID = "project_quick_switch"
	CommandProjectSearch      CommandID = "project_search"
)

// CommandSpec captures palette metadata for a command.
type CommandSpec struct {
	ID          CommandID
	Label       string
	Description string
	Shortcut    string
}

func defaultCommandSpecs() []CommandSpec {
	return []CommandSpec{
		{ID: CommandParsePRD, Label: "Parse PRD", Description: "Parse a PRD file and generate tasks", Shortcut: "Alt+P"},
		{ID: CommandAnalyzeComplexity, Label: "Analyze Complexity", Description: "Run complexity analysis via Task Master", Shortcut: "Alt+C"},
		{ID: CommandExpandTask, Label: "Expand Task", Description: "Break down the selected task with AI", Shortcut: "Alt+E"},
		{ID: CommandDeleteTask, Label: "Delete Task", Description: "Open the safe delete workflow for selected tasks", Shortcut: "Alt+D"},
		{ID: CommandManageTags, Label: "Add Tag Context", Description: "Create a new tag context", Shortcut: "Ctrl+Shift+A"},
		{ID: CommandTagManagement, Label: "Manage Tag Contexts", Description: "View and modify tag contexts", Shortcut: "Ctrl+Shift+M"},
		{ID: CommandUseTag, Label: "Use Tag Context", Description: "Switch the active Task Master tag", Shortcut: "Ctrl+Shift+U"},
		{ID: CommandProjectTags, Label: "Project Tags", Description: "Browse project tags and switch", Shortcut: "Ctrl+T"},
		{ID: CommandProjectQuickSwitch, Label: "Quick Project Switch", Description: "Switch between recent projects", Shortcut: "Ctrl+Q"},
		{ID: CommandProjectSearch, Label: "Search Projects", Description: "Search tags or projects", Shortcut: "Ctrl+Shift+T"},
	}
}

// commandPaletteItem adapts a command spec to a dialog.ListItem.
type commandPaletteItem struct {
	spec CommandSpec
}

func newCommandPaletteItem(spec CommandSpec) dialog.ListItem {
	return &commandPaletteItem{spec: spec}
}

// Title implements dialog.ListItem.
func (i *commandPaletteItem) Title() string {
	if i.spec.Shortcut == "" {
		return i.spec.Label
	}
	return fmt.Sprintf("%s (%s)", i.spec.Label, i.spec.Shortcut)
}

// Description implements dialog.ListItem.
func (i *commandPaletteItem) Description() string {
	return i.spec.Description
}

// FilterValue implements dialog.ListItem.
func (i *commandPaletteItem) FilterValue() string {
	return i.spec.Label
}

// commandShortcut associates a key binding with a command.
type commandShortcut struct {
	binding key.Binding
	command CommandID
	help    string
}

func (cs commandShortcut) matches(msg tea.KeyMsg) bool {
	return key.Matches(msg, cs.binding)
}
