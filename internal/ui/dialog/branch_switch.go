package dialog

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/agreen757/tm-tui/internal/git"
)

// branchItem implements the list.Item interface for branch list items
type branchItem struct {
	title       string
	description string
}

// Title returns the title of the branch item
func (i branchItem) Title() string {
	return i.title
}

// Description returns the description of the branch item
func (i branchItem) Description() string {
	return i.description
}

// FilterValue returns the value used for filtering
func (i branchItem) FilterValue() string {
	return i.title
}

// BranchSwitchDialog is a dialog that lists and allows switching between branches
type BranchSwitchDialog struct {
	BaseDialog
	list          list.Model
	branches      []string
	currentBranch string
	onSwitch      func(string, string, error)
	repoPath      string
	switching     bool
}

// NewBranchSwitchDialog creates a new branch switch dialog
func NewBranchSwitchDialog(repoPath string, onSwitch func(string, string, error)) (*BranchSwitchDialog, error) {
	ctx := context.Background()
	branches, currentBranch, err := git.GetBranches(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	// Handle case where no branches exist
	if len(branches) == 0 {
		branches = []string{}
	}

	// Create list items from branches
	items := make([]list.Item, len(branches))
	for i, branch := range branches {
		description := ""
		if branch == currentBranch {
			description = "(current)"
		}
		items[i] = branchItem{title: branch, description: description}
	}

	// Create the list delegate with proper configuration
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)
	delegate.SetSpacing(0)

	// Create the list model
	l := list.New(items, delegate, 0, 0)
	l.Title = "Switch Branch"
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)

	// Create the dialog with BaseDialog
	baseDialog := NewBaseDialog("Switch Branch", 60, 20, DialogKindList)

	dialog := &BranchSwitchDialog{
		BaseDialog:    baseDialog,
		list:          l,
		branches:      branches,
		currentBranch: currentBranch,
		onSwitch:      onSwitch,
		repoPath:      repoPath,
		switching:     false,
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Switch"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)

	// Set initial list size based on dialog dimensions
	dialog.list.SetSize(baseDialog.width-4, baseDialog.height-6)

	return dialog, nil
}

// Init initializes the dialog
func (d *BranchSwitchDialog) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates dialog state
func (d *BranchSwitchDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Recalculate position on window resize
		d.Center(msg.Width, msg.Height)
		d.list.SetSize(d.width-4, d.height-6)
		return d, nil

	case tea.KeyMsg:
		if !d.switching {
			// Handle base dialog keys first (ESC for cancel)
			result, cmd := d.HandleBaseKey(msg)
			if result != DialogResultNone {
				return nil, cmd
			}

			// Handle branch-specific keys
			result, cmd = d.HandleKey(msg)
			if result != DialogResultNone {
				return nil, cmd
			}

			// Pass remaining keys to list
			var cmd2 tea.Cmd
			d.list, cmd2 = d.list.Update(msg)
			return d, cmd2
		}

	case branchSwitchResult:
		d.switching = false
		if d.onSwitch != nil {
			d.onSwitch(msg.branch, msg.output, msg.err)
		}
		return nil, nil
	}

	// Default: pass through to list
	var cmd tea.Cmd
	d.list, cmd = d.list.Update(msg)
	return d, cmd
}

// View renders the dialog
func (d *BranchSwitchDialog) View() string {
	content := d.list.View()
	if d.switching {
		content += lipgloss.NewStyle().
			PaddingTop(1).
			Render("Switching branch...")
	}
	return d.RenderBorder(content)
}

// branchSwitchResult is a message sent when branch switching completes
type branchSwitchResult struct {
	branch string
	output string
	err    error
}

// switchBranchCmd returns a command that switches to the specified branch
func (d *BranchSwitchDialog) switchBranchCmd(branch string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		output, err := git.SwitchBranch(ctx, d.repoPath, branch)
		return branchSwitchResult{
			branch: branch,
			output: output,
			err:    err,
		}
	}
}

// HandleKey processes a key event
func (d *BranchSwitchDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	if d.switching {
		return DialogResultNone, nil
	}

	switch msg.String() {
	case "up", "k":
		d.list.CursorUp()
		return DialogResultNone, nil

	case "down", "j":
		d.list.CursorDown()
		return DialogResultNone, nil

	case "enter":
		if d.list.SelectedItem() == nil {
			return DialogResultNone, nil
		}
		selected := d.list.SelectedItem().(branchItem).title
		if selected == d.currentBranch {
			return DialogResultNone, nil
		}

		d.switching = true
		return DialogResultNone, d.switchBranchCmd(selected)
	}

	return DialogResultNone, nil
}
