package dialog

import (
	"context"

	"github.com/agreen757/tm-tui/internal/git"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommitItem represents a commit in the list
type CommitItem struct {
	commit git.Commit
}

// Title returns the title of the commit item
func (i CommitItem) Title() string {
	return i.commit.Hash + " " + i.commit.Subject
}

// Description returns the description of the commit item
func (i CommitItem) Description() string {
	return i.commit.RelativeTime + " by " + i.commit.Author
}

// FilterValue returns the value used for filtering
func (i CommitItem) FilterValue() string {
	return i.commit.Hash + " " + i.commit.Subject + " " + i.commit.Author
}

// CommitsDialog displays recent commits in a list
type CommitsDialog struct {
	BaseDialog
	selectedIndex int
	commits       []git.Commit
	repoPath      string
	onSelect      func(git.Commit)
	loading       bool
	err           error
}

// NewCommitsDialog creates a new commits dialog
func NewCommitsDialog(repoPath string, onSelect func(git.Commit)) *CommitsDialog {
	dialog := &CommitsDialog{
		BaseDialog:    NewBaseDialog("Recent Commits", 80, 20, DialogKindList),
		selectedIndex: 0,
		repoPath:      repoPath,
		onSelect:      onSelect,
		commits:       []git.Commit{},
		loading:       true,
	}

	// Set footer hints
	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "r", Label: "Refresh"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)

	return dialog
}

// Init initializes the dialog and loads commits
func (d *CommitsDialog) Init() tea.Cmd {
	return d.loadCommits()
}

// Update processes messages and updates dialog state
func (d *CommitsDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Recalculate position on window resize
		d.Center(msg.Width, msg.Height)

	case CommitsRefreshMsg:
		d.loading = false
		d.commits = msg.Commits
		d.err = msg.Err
		if d.selectedIndex >= len(d.commits) && len(d.commits) > 0 {
			d.selectedIndex = len(d.commits) - 1
		}
	}

	return d, nil
}

// View renders the dialog
func (d *CommitsDialog) View() string {
	var content string

	if d.loading {
		content = d.renderLoading()
	} else if d.err != nil {
		content = d.renderError()
	} else if len(d.commits) == 0 {
		content = d.renderEmpty()
	} else {
		content = d.renderCommits()
	}

	return d.RenderBorder(content)
}

// renderLoading renders the loading state
func (d *CommitsDialog) renderLoading() string {
	style := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Width(d.width - 4).
		Align(lipgloss.Center)

	return style.Render("Loading commits...")
}

// renderError renders the error state
func (d *CommitsDialog) renderError() string {
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
		errorStyle.Render("Error loading commits:"),
		messageStyle.Render(d.err.Error()),
	)
}

// renderEmpty renders the empty state
func (d *CommitsDialog) renderEmpty() string {
	style := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Width(d.width - 4).
		Align(lipgloss.Center)

	return style.Render("No commits found")
}

// renderCommits renders the list of commits
func (d *CommitsDialog) renderCommits() string {
	contentWidth := d.width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	var items []string
	for i, commit := range d.commits {
		items = append(items, d.renderCommitItem(commit, i == d.selectedIndex, contentWidth))
		// Add spacing between items (except after last item)
		if i < len(d.commits)-1 {
			items = append(items, "")
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	return content
}

// renderCommitItem renders a single commit item
func (d *CommitsDialog) renderCommitItem(commit git.Commit, focused bool, width int) string {
	// Style hash with shorter length (8 chars for short hash)
	hashWidth := 8
	hashStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor).
		Width(hashWidth)

	hash := commit.Hash
	if len(hash) > hashWidth {
		hash = hash[:hashWidth]
	}

	// Style subject
	subjectStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Width(width - hashWidth - 2)

	if focused {
		subjectStyle = subjectStyle.
			Bold(true).
			Underline(true).
			Background(lipgloss.Color("#1a1a1a"))
	}

	// Truncate subject if too long
	subject := commit.Subject
	maxSubjectWidth := width - hashWidth - 4
	if len(subject) > maxSubjectWidth {
		subject = subject[:maxSubjectWidth-3] + "..."
	}

	prefix := "  "
	if focused {
		prefix = "▶ "
	}

	titleLine := hashStyle.Render(hash) + " " + prefix + subjectStyle.Render(subject)

	// Render metadata on next line
	metaStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor).
		Faint(true).
		Width(width).
		PaddingLeft(hashWidth + 2)

	if focused {
		metaStyle = metaStyle.Faint(false)
	}

	metadata := metaStyle.Render(commit.RelativeTime + " by " + commit.Author)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		metadata,
	)
}

// HandleKey processes a key event
func (d *CommitsDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle navigation keys
	switch msg.String() {
	case "up", "k":
		if d.selectedIndex > 0 {
			d.selectedIndex--
		} else if len(d.commits) > 0 {
			d.selectedIndex = len(d.commits) - 1
		}
		return DialogResultNone, nil

	case "down", "j":
		if d.selectedIndex < len(d.commits)-1 {
			d.selectedIndex++
		} else {
			d.selectedIndex = 0
		}
		return DialogResultNone, nil

	case "enter":
		if d.selectedIndex >= 0 && d.selectedIndex < len(d.commits) {
			if d.onSelect != nil {
				d.onSelect(d.commits[d.selectedIndex])
			}
			return DialogResultClose, nil
		}
		return DialogResultNone, nil

	case "r":
		// Refresh the commit list
		d.loading = true
		d.err = nil
		return DialogResultNone, d.loadCommits()
	}

	return DialogResultNone, nil
}

// CommitsRefreshMsg is sent when commits are loaded
type CommitsRefreshMsg struct {
	Commits []git.Commit
	Err     error
}

// loadCommits loads recent commits from the repository
func (d *CommitsDialog) loadCommits() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		commits, err := git.GetRecentCommits(ctx, d.repoPath, 20)
		return CommitsRefreshMsg{
			Commits: commits,
			Err:     err,
		}
	}
}
