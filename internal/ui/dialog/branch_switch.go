package dialog

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/agreen757/tm-tui/internal/git"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	list           list.Model
	branches       []string
	currentBranch  string
	onSwitch       func(string, string, error)
	repoPath       string
	tagName        string
	switching      bool
	mu             sync.RWMutex // Protects concurrent access to dialog state
	currentTaskID  string       // Track current git operation task ID
	selectedBranch string       // Cache selected branch name during operation
}

// NewBranchSwitchDialog creates a new branch switch dialog
func NewBranchSwitchDialog(repoPath string, onSwitch func(string, string, error), tagName string) (*BranchSwitchDialog, error) {
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
	delegate.SetHeight(3)
	delegate.SetSpacing(1)
	delegate.ShowDescription = true

	// Create the list model
	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	// Create the dialog with BaseDialog
	baseDialog := NewBaseDialog("Switch Branch", 60, 20, DialogKindList)

	dialog := &BranchSwitchDialog{
		BaseDialog:     baseDialog,
		list:           l,
		branches:       branches,
		currentBranch:  currentBranch,
		onSwitch:       onSwitch,
		repoPath:       repoPath,
		tagName:        tagName,
		switching:      false,
		mu:             sync.RWMutex{},
		currentTaskID:  "",
		selectedBranch: "",
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
		d.mu.RLock()
		isSwitching := d.switching
		d.mu.RUnlock()

		if !isSwitching {
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

	case TaskCompletedMsg:
		// PHASE 1 FIX: Safely handle completion of branch switch
		// Verify this message is for the current operation (prevents orphaned messages)
		d.mu.Lock()
		defer d.mu.Unlock()

		// Check if task ID matches - prevents processing stale completion messages
		if d.currentTaskID != "git-switch-branch" {
			// This message is not for us, ignore it
			return d, nil
		}

		d.switching = false

		// CRITICAL FIX: Get selected branch name that was cached during operation
		// Don't try to read list state as it may be invalid after async operation
		selectedBranch := d.selectedBranch
		d.selectedBranch = ""
		d.currentTaskID = ""

		// Verify we have a valid branch name to report
		if selectedBranch == "" {
			// Branch name was lost, close silently
			return nil, nil
		}

		if d.onSwitch != nil {
			d.onSwitch(selectedBranch, "Branch switched successfully", nil)
		}
		return nil, nil

	case TaskFailedMsg:
		// PHASE 1 FIX: Safely handle failure of branch switch
		d.mu.Lock()
		defer d.mu.Unlock()

		// Check if task ID matches
		if d.currentTaskID != "git-switch-branch" {
			// This message is not for us, ignore it
			return d, nil
		}

		d.switching = false
		d.selectedBranch = ""
		d.currentTaskID = ""

		if d.onSwitch != nil {
			d.onSwitch("", msg.Error, fmt.Errorf("%s", msg.Error))
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
	var content strings.Builder

	// Add title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.Style.TitleColor).
		Render("Select Branch")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Render list
	content.WriteString(d.list.View())

	// Status message - use safe read
	d.mu.RLock()
	isSwitching := d.switching
	d.mu.RUnlock()

	if isSwitching {
		content.WriteString("\n\n")
		statusMsg := lipgloss.NewStyle().
			Foreground(d.Style.ButtonColor).
			Render("Switching branch...")
		content.WriteString(statusMsg)
	}

	return d.RenderBorder(content.String())
}

// launchGitSwitchBranch launches the git branch switch via Task Runner
func (d *BranchSwitchDialog) launchGitSwitchBranch(branch string) tea.Cmd {
	return tea.Sequence(
		// First, send TaskStartedMsg to open the Task Runner modal
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-switch-branch",
				TaskTitle: "Switch Branch: " + branch,
			}
		},
		// Then, execute the git command
		ExecuteGitCommand("git-switch-branch", []string{"checkout", branch}, d.tagName),
	)
}

// HandleKey processes a key event
func (d *BranchSwitchDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	d.mu.RLock()
	isSwitching := d.switching
	d.mu.RUnlock()

	if isSwitching {
		return DialogResultNone, nil
	}

	switch msg.String() {
	case "enter":
		// PHASE 1 FIX: Safely get selected item with null checks
		item := d.list.SelectedItem()
		if item == nil {
			// No item selected, ignore Enter
			return DialogResultNone, nil
		}

		// Safely cast to branchItem - nil check prevents panic
		branchItem, ok := item.(branchItem)
		if !ok {
			// Type assertion failed, shouldn't happen but be defensive
			return DialogResultNone, nil
		}

		selected := branchItem.title

		if selected == d.currentBranch {
			// Already on this branch, ignore
			return DialogResultNone, nil
		}

		// PHASE 2 FIX: Thread-safe state update with task tracking
		d.mu.Lock()
		d.switching = true
		d.currentTaskID = "git-switch-branch"
		d.selectedBranch = selected // Cache branch name for later use
		d.mu.Unlock()

		// Close dialog and launch git switch via Task Runner
		return DialogResultClose, d.launchGitSwitchBranch(selected)
	}

	return DialogResultNone, nil
}
