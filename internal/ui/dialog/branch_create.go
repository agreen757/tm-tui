package dialog

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/agreen757/tm-tui/internal/git"
)

// branchCreateMsg is sent when branch creation completes
type branchCreateMsg struct {
	branchName string
	output     string
	err        error
}

// BranchCreateDialog allows creating a new branch
type BranchCreateDialog struct {
	BaseDialog
	input      textinput.Model
	repoPath   string
	loading    bool
	errorMsg   string
	onComplete func(branchName string, output string, err error)
}

// NewBranchCreateDialog creates a new branch creation dialog
func NewBranchCreateDialog(repoPath string, onComplete func(branchName string, output string, err error)) *BranchCreateDialog {
	input := textinput.New()
	input.Placeholder = "new-branch-name"
	input.Focus()

	dialog := &BranchCreateDialog{
		BaseDialog:  NewBaseDialog("Create Branch", 60, 12, DialogKindForm),
		input:       input,
		repoPath:    repoPath,
		loading:     false,
		onComplete:  onComplete,
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
	case tea.KeyMsg:
		if d.loading {
			// Don't process input while loading
			return d, nil
		}

		switch msg.String() {
		case "enter":
			if !d.isValidBranchName() {
				d.errorMsg = "Invalid branch name"
				return d, nil
			}

			branchName := strings.TrimSpace(d.input.Value())
			d.loading = true
			d.errorMsg = ""
			return d, d.createBranchCmd(branchName)

		case "esc":
			return nil, nil
		}

	case branchCreateMsg:
		if msg.err != nil {
			d.loading = false
			d.errorMsg = "Error: " + msg.err.Error()
			return d, nil
		}

		// Success - call callback and close dialog
		if d.onComplete != nil {
			d.onComplete(msg.branchName, msg.output, nil)
		}
		return nil, nil
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

// createBranchCmd creates a command that attempts to create the branch
func (d *BranchCreateDialog) createBranchCmd(branchName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		output, err := git.CreateBranch(ctx, d.repoPath, branchName)
		return branchCreateMsg{
			branchName: branchName,
			output:     output,
			err:        err,
		}
	}
}

// View renders the dialog
func (d *BranchCreateDialog) View() string {
	content := "Enter new branch name:\n\n"

	if d.loading {
		content += "Creating branch... \n\n"
	} else {
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
	}

	// Show command hints
	hintsText := "Press Enter to create, Esc to cancel"
	if d.isValidBranchName() {
		hintsText = "✓ Ready - Press Enter to create, Esc to cancel"
	} else if d.input.Value() != "" {
		hintsText = "Invalid name - Press Esc to cancel"
	}
	content += hintsText + "\n"

	// Show error or success message
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
		if d.loading {
			return DialogResultNone, nil
		}

		if !d.isValidBranchName() {
			d.errorMsg = "Invalid branch name"
			return DialogResultNone, nil
		}

		branchName := strings.TrimSpace(d.input.Value())
		d.loading = true
		d.errorMsg = ""
		return DialogResultNone, d.createBranchCmd(branchName)
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
