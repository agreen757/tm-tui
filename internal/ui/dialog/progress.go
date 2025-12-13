package dialog

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressUpdateMsg is sent when progress is updated
type ProgressUpdateMsg struct {
	Progress float64
	Label    string
}

// ProgressCompleteMsg is sent when progress is complete
type ProgressCompleteMsg struct{}

// ProgressCancelMsg is sent when progress is canceled
type ProgressCancelMsg struct{}

// ProgressDialog is a dialog with a progress bar
type ProgressDialog struct {
	BaseDialog
	progress   float64
	label      string
	width      int
	lastUpdate time.Time
	autoClose  bool
	canceled   bool
	completed  bool
}

// NewProgressDialog creates a new progress dialog
func NewProgressDialog(title string, width, height int) *ProgressDialog {
	dialog := &ProgressDialog{
		BaseDialog: NewBaseDialog(title, width, height, DialogKindProgress),
		progress:   0.0,
		label:      "",
		width:      width,
		lastUpdate: time.Now(),
		autoClose:  true,
		canceled:   false,
		completed:  false,
	}
	dialog.SetFooterHints(
		ShortcutHint{Key: "Esc/C", Label: "Cancel"},
	)
	return dialog
}

// Init initializes the dialog
func (d *ProgressDialog) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates dialog state
func (d *ProgressDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)

	case ProgressUpdateMsg:
		d.progress = msg.Progress
		if msg.Label != "" {
			d.label = msg.Label
		}
		d.lastUpdate = time.Now()

		// Check if progress is complete
		if d.progress >= 1.0 && !d.completed {
			d.completed = true
			if d.autoClose {
				return d, func() tea.Msg {
					return ProgressCompleteMsg{}
				}
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *ProgressDialog) View() string {
	// Account for border and padding
	contentWidth := d.width - 4

	if contentWidth < 1 {
		contentWidth = 1
	}

	// Render label
	labelStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Left)

	label := d.label
	if label == "" {
		label = "Processing..."
	}

	labelText := labelStyle.Render(label)

	// Render progress bar
	progressText := d.renderProgress(contentWidth)

	// Render percentage
	percentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Right)

	percent := fmt.Sprintf("%.0f%%", d.progress*100)
	percentText := percentStyle.Render(percent)

	// Render cancel hint if cancellable
	var cancelText string
	if d.IsCancellable() && !d.canceled && !d.completed {
		cancelStyle := lipgloss.NewStyle().
			Foreground(d.Style.TextColor).
			Italic(true).
			Width(contentWidth).
			Align(lipgloss.Left).
			PaddingTop(1)

		cancelText = cancelStyle.Render("Press ESC or C to cancel")
	}

	// Combine everything
	content := lipgloss.JoinVertical(lipgloss.Left,
		labelText,
		progressText,
		percentText,
		cancelText)

	// Add border and title
	return d.RenderBorder(content)
}

// renderProgress renders the progress bar
func (d *ProgressDialog) renderProgress(width int) string {
	// Calculate the number of filled blocks
	barWidth := width - 2 // Account for brackets
	filled := int(float64(barWidth) * d.progress)

	if filled > barWidth {
		filled = barWidth
	}

	// Create the progress bar
	filledChar := "█"
	emptyChar := "░"

	bar := strings.Repeat(filledChar, filled) + strings.Repeat(emptyChar, barWidth-filled)

	// Style the bar
	barStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor)

	return "[" + barStyle.Render(bar) + "]"
}

// HandleKey processes a key event
func (d *ProgressDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (like ESC)
	result, _ := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		d.canceled = true
		return result, func() tea.Msg {
			return ProgressCancelMsg{}
		}
	}

	// Handle progress-specific keys
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		if d.IsCancellable() && !d.canceled {
			d.canceled = true
			return DialogResultCancel, func() tea.Msg {
				return ProgressCancelMsg{}
			}
		}
	}

	return DialogResultNone, nil
}

// Progress returns the current progress
func (d *ProgressDialog) Progress() float64 {
	return d.progress
}

// SetProgress sets the current progress
func (d *ProgressDialog) SetProgress(progress float64) {
	d.progress = progress
	if d.progress < 0.0 {
		d.progress = 0.0
	}
	if d.progress > 1.0 {
		d.progress = 1.0
	}
}

// Label returns the current label
func (d *ProgressDialog) Label() string {
	return d.label
}

// SetLabel sets the label text
func (d *ProgressDialog) SetLabel(label string) {
	d.label = label
}

// IsCanceled returns whether the dialog was canceled
func (d *ProgressDialog) IsCanceled() bool {
	return d.canceled
}

// IsCompleted returns whether the progress is completed
func (d *ProgressDialog) IsCompleted() bool {
	return d.completed
}

// SetAutoClose sets whether the dialog should automatically close when complete
func (d *ProgressDialog) SetAutoClose(autoClose bool) {
	d.autoClose = autoClose
}

// UpdateProgress sends a progress update message
func UpdateProgress(progress float64, label string) tea.Cmd {
	return func() tea.Msg {
		return ProgressUpdateMsg{
			Progress: progress,
			Label:    label,
		}
	}
}
