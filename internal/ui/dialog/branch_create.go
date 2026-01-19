package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BranchCreateDialog allows creating a new branch
type BranchCreateDialog struct {
	BaseDialog
	input    textinput.Model
	repoPath string
	tagName  string
	errorMsg string
}

// NewBranchCreateDialog creates a new branch creation dialog
func NewBranchCreateDialog(repoPath string, tagName string) *BranchCreateDialog {
	input := textinput.New()
	input.Placeholder = "new-branch-name"
	input.Focus()

	dialog := &BranchCreateDialog{
		BaseDialog: NewBaseDialog("Create Branch", 60, 12, DialogKindForm),
		input:      input,
		repoPath:   repoPath,
		tagName:    tagName,
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "Enter", Label: "Create"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	return dialog
}

// Init initializes the dialog
func (d *BranchCreateDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes messages and updates dialog state
func (d *BranchCreateDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
		return d, nil
	}

	// Handle text input and validate in real-time
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)

	// Clear error message on new input
	if _, ok := msg.(tea.KeyMsg); ok {
		d.validateAndUpdateError()
	}

	return d, cmd
}

// isValidBranchName checks if the current input is a valid branch name
func (d *BranchCreateDialog) isValidBranchName() bool {
	branchName := strings.TrimSpace(d.input.Value())

	if branchName == "" {
		return false
	}

	if strings.Contains(branchName, " ") {
		return false
	}

	// Check for other invalid characters
	invalidChars := []string{"\t", "\n", "\r", ":", "~", "^", "?", "*", "[", "\\"}
	for _, char := range invalidChars {
		if strings.Contains(branchName, char) {
			return false
		}
	}

	return true
}

// validateAndUpdateError updates error message based on current input
func (d *BranchCreateDialog) validateAndUpdateError() {
	branchName := d.input.Value()

	if strings.TrimSpace(branchName) == "" {
		d.errorMsg = ""
		return
	}

	// Check for spaces (works with trimmed value)
	if strings.Contains(branchName, " ") {
		d.errorMsg = "Branch name cannot contain spaces"
		return
	}

	// Check for other invalid characters
	if strings.Contains(branchName, ":") {
		d.errorMsg = "Branch name cannot contain colons"
		return
	}
	if strings.Contains(branchName, "~") || strings.Contains(branchName, "^") ||
		strings.Contains(branchName, "?") || strings.Contains(branchName, "*") {
		d.errorMsg = "Branch name contains invalid characters"
		return
	}

	d.errorMsg = ""
}

// launchGitCreateBranch launches the git branch creation via Task Runner
func (d *BranchCreateDialog) launchGitCreateBranch(branchName string) tea.Cmd {
	return tea.Sequence(
		// First, send TaskStartedMsg to open the Task Runner modal
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-create-branch",
				TaskTitle: "Create Branch: " + branchName,
			}
		},
		// Then, execute the git command
		ExecuteGitCommand("git-create-branch", []string{"checkout", "-b", branchName}, d.tagName),
	)
}

// View renders the dialog
func (d *BranchCreateDialog) View() string {
	content := "Enter new branch name:\n\n"

	// Show input with validation feedback
	inputView := d.input.View()
	if !d.isValidBranchName() && d.input.Value() != "" {
		inputView = lipgloss.NewStyle().
			Foreground(d.Style.ErrorColor).
			Render(inputView)
	} else if d.isValidBranchName() {
		inputView = lipgloss.NewStyle().
			Foreground(d.Style.SuccessColor).
			Render(inputView)
	}
	content += inputView + "\n\n"

	// Show command hints
	hintsText := "Press Enter to create, Esc to cancel"
	if d.isValidBranchName() {
		hintsText = "✓ Ready - Press Enter to create, Esc to cancel"
	} else if d.input.Value() != "" {
		hintsText = "Invalid name - Press Esc to cancel"
	}
	content += hintsText + "\n"

	// Show error message
	if d.errorMsg != "" {
		content += "\n" + lipgloss.NewStyle().
			Foreground(d.Style.ErrorColor).
			Render(d.errorMsg)
	}

	return d.RenderBorder(content)
}

// HandleKey processes a key event
func (d *BranchCreateDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle dialog-specific keys
	switch msg.String() {
	case "enter":
		if !d.isValidBranchName() {
			d.errorMsg = "Invalid branch name"
			return DialogResultNone, nil
		}

		branchName := strings.TrimSpace(d.input.Value())
		// Close dialog immediately and launch git command via Task Runner
		return DialogResultClose, d.launchGitCreateBranch(branchName)
	}

	return DialogResultNone, nil
}

// SetRect sets the dialog's position and dimensions
func (d *BranchCreateDialog) SetRect(width, height, x, y int) {
	d.BaseDialog.SetRect(width, height, x, y)
	// Adjust input width to fit inside the dialog borders
	inputWidth := width - 4 // Account for borders and padding
	if inputWidth < 10 {
		inputWidth = 10
	}
	d.input.Width = inputWidth
}

// GetRect returns the dialog's position and dimensions
func (d *BranchCreateDialog) GetRect() (width, height, x, y int) {
	return d.BaseDialog.GetRect()
}
