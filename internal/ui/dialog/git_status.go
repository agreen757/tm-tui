package dialog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/git"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GitStatusRefreshMsg is sent when git status is refreshed
type GitStatusRefreshMsg struct {
	Status git.GitStatus
	Err    error
}

// GitStatusDialog displays detailed git repository status
type GitStatusDialog struct {
	BaseDialog
	repoPath string
	status   git.GitStatus
	err      error
	loading  bool
}

// NewGitStatusDialog creates a new git status dialog
func NewGitStatusDialog(repoPath string) *GitStatusDialog {
	dialog := &GitStatusDialog{
		BaseDialog: NewBaseDialog("Git Status", 70, 18, DialogKindModal),
		repoPath:   repoPath,
		loading:    true,
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "r", Label: "Refresh"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)

	return dialog
}

// Init initializes the dialog and fetches initial status
func (d *GitStatusDialog) Init() tea.Cmd {
	return d.refreshStatus()
}

// Update processes messages and updates dialog state
func (d *GitStatusDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
		return d, nil

	case GitStatusRefreshMsg:
		d.loading = false
		d.status = msg.Status
		d.err = msg.Err
		return d, nil
	}

	return d, nil
}

// View renders the dialog
func (d *GitStatusDialog) View() string {
	var content string

	if d.loading {
		content = d.renderLoading()
	} else if d.err != nil {
		content = d.renderError()
	} else {
		content = d.renderStatus()
	}

	return d.RenderBorder(content)
}

// renderLoading renders the loading state
func (d *GitStatusDialog) renderLoading() string {
	style := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Width(d.width - 4).
		Align(lipgloss.Center)

	return style.Render("Loading git status...")
}

// renderError renders the error state
func (d *GitStatusDialog) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(d.Style.ErrorColor).
		Bold(true).
		Width(d.width - 4)

	messageStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Width(d.width - 4).
		PaddingTop(1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		errorStyle.Render("Error fetching git status:"),
		messageStyle.Render(d.err.Error()),
	)
}

// renderStatus renders the git status information
func (d *GitStatusDialog) renderStatus() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor).
		Bold(true).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor)

	successStyle := lipgloss.NewStyle().
		Foreground(d.Style.SuccessColor)

	warningStyle := lipgloss.NewStyle().
		Foreground(d.Style.WarningColor)

	var lines []string

	// Branch
	lines = append(lines, labelStyle.Render("Branch:")+" "+valueStyle.Render(d.status.Branch))

	// Working directory state
	stateLabel := labelStyle.Render("State:")
	var stateValue string
	if d.status.IsDirty {
		stateValue = warningStyle.Render("Dirty (uncommitted changes)")
	} else {
		stateValue = successStyle.Render("Clean")
	}
	lines = append(lines, stateLabel+" "+stateValue)

	// Upstream tracking
	upstreamLabel := labelStyle.Render("Upstream:")
	var upstreamValue string
	if d.status.HasUpstream {
		upstreamValue = successStyle.Render("Tracked")
		
		// Show ahead/behind if tracked
		if d.status.Ahead > 0 || d.status.Behind > 0 {
			var parts []string
			if d.status.Ahead > 0 {
				parts = append(parts, fmt.Sprintf("↑ %d ahead", d.status.Ahead))
			}
			if d.status.Behind > 0 {
				parts = append(parts, fmt.Sprintf("↓ %d behind", d.status.Behind))
			}
			syncStatus := strings.Join(parts, ", ")
			lines = append(lines, upstreamLabel+" "+upstreamValue)
			lines = append(lines, labelStyle.Render("Sync:")+" "+warningStyle.Render(syncStatus))
		} else {
			lines = append(lines, upstreamLabel+" "+upstreamValue+" "+successStyle.Render("(up to date)"))
		}
	} else {
		upstreamValue = warningStyle.Render("Not tracked")
		lines = append(lines, upstreamLabel+" "+upstreamValue)
	}

	// Last updated
	lines = append(lines, "")
	timeStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Faint(true)
	lines = append(lines, timeStyle.Render("Last updated: "+d.status.LastUpdated.Format("15:04:05")))

	content := strings.Join(lines, "\n")
	
	containerStyle := lipgloss.NewStyle().
		Width(d.width - 4).
		PaddingTop(1)

	return containerStyle.Render(content)
}

// HandleKey processes a key event
func (d *GitStatusDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle refresh key
	switch msg.String() {
	case "r", "R":
		d.loading = true
		return DialogResultNone, d.refreshStatus()
	}

	return DialogResultNone, nil
}

// refreshStatus fetches the latest git status
func (d *GitStatusDialog) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, err := git.GetStatus(ctx, d.repoPath)
		return GitStatusRefreshMsg{
			Status: status,
			Err:    err,
		}
	}
}
