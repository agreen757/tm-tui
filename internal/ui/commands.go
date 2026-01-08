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
	CommandCreatePRD          CommandID = "create_prd"
	CommandAnalyzeComplexity  CommandID = "analyze_complexity"
	CommandExpandTask         CommandID = "expand_task"
	CommandDeleteTask         CommandID = "delete_task"
	CommandManageTags         CommandID = "manage_tags"
	CommandTagManagement      CommandID = "tag_management"
	CommandUseTag             CommandID = "use_tag"
	CommandProjectTags        CommandID = "project_tags"
	CommandProjectQuickSwitch CommandID = "project_quick_switch"
	CommandProjectSearch      CommandID = "project_search"
	CommandRunTask            CommandID = "run_task"
	CommandRunCommand         CommandID = "run_command"
	CommandGitMenu            CommandID = "git.menu"
	CommandGitSwitchBranch    CommandID = "git.switchBranch"
	CommandGitCreateBranch    CommandID = "git.createBranch"
	CommandGitRecentCommits   CommandID = "git.recentCommits"
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
		{ID: CommandCreatePRD, Label: "Create PRD", Description: "Create a new PRD document with AI assistance", Shortcut: "Alt+Shift+P"},
		{ID: CommandAnalyzeComplexity, Label: "Analyze Complexity", Description: "Run complexity analysis via Task Master", Shortcut: "Alt+C"},
		{ID: CommandExpandTask, Label: "Expand Task", Description: "Break down the selected task with AI", Shortcut: "Alt+E"},
		{ID: CommandDeleteTask, Label: "Delete Task", Description: "Open the safe delete workflow for selected tasks", Shortcut: "Alt+D"},
		{ID: CommandRunTask, Label: "Run Task with Crush", Description: "Execute the selected task via Crush AI agent", Shortcut: "Alt+R / Ctrl+R"},
		{ID: CommandRunCommand, Label: "Run Command", Description: "Run a command with Crush AI", Shortcut: "Ctrl+B"},
		{ID: CommandManageTags, Label: "Add Tag Context", Description: "Create a new tag context", Shortcut: "Ctrl+Shift+A"},
		{ID: CommandTagManagement, Label: "Manage Tag Contexts", Description: "View and modify tag contexts", Shortcut: "Ctrl+Shift+M"},
		{ID: CommandUseTag, Label: "Use Tag Context", Description: "Switch the active Task Master tag", Shortcut: "Ctrl+Shift+U"},
		{ID: CommandProjectTags, Label: "Project Tags", Description: "Browse project tags and switch", Shortcut: "Ctrl+T"},
		{ID: CommandProjectQuickSwitch, Label: "Quick Project Switch", Description: "Switch between recent projects", Shortcut: "Ctrl+Q"},
		{ID: CommandProjectSearch, Label: "Search Projects", Description: "Search tags or projects", Shortcut: "Ctrl+Shift+T"},
		{ID: CommandGitMenu, Label: "Git: Open Menu", Description: "Open the Git menu to access Git operations", Shortcut: "g"},
		{ID: CommandGitSwitchBranch, Label: "Git: Switch Branch", Description: "Checkout an existing Git branch", Shortcut: ""},
		{ID: CommandGitCreateBranch, Label: "Git: Create Branch", Description: "Create and checkout a new Git branch", Shortcut: ""},
		{ID: CommandGitRecentCommits, Label: "Git: Recent Commits", Description: "View recent commit history", Shortcut: ""},
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
