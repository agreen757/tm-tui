package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight    = 3
	statusBarHeight = 1
	panelPadding    = 2
	minPanelWidth   = 20
	tagMaxLength    = 12
	tagEllipsis     = "..."
)

// LayoutDimensions holds the calculated dimensions for each panel
type LayoutDimensions struct {
	// Total dimensions
	Width  int
	Height int

	// Header
	HeaderHeight int

	// Main content area
	ContentWidth  int
	ContentHeight int

	// Task list (left side)
	TaskListWidth  int
	TaskListHeight int

	// Details panel (right side)
	DetailsWidth  int
	DetailsHeight int

	// Log panel (bottom)
	LogWidth  int
	LogHeight int

	// Status bar
	StatusBarHeight int
}

// calculateLayout computes the layout dimensions based on terminal size and panel visibility
func (m Model) calculateLayout() LayoutDimensions {
	layout := LayoutDimensions{
		Width:           m.width,
		Height:          m.height,
		HeaderHeight:    headerHeight,
		StatusBarHeight: statusBarHeight,
	}

	// Calculate available content height
	contentHeight := m.height - headerHeight - statusBarHeight
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Determine log panel height
	logHeight := 0
	if m.showLogPanel {
		logHeight = contentHeight / 3 // Log takes 1/3 of content height
		if logHeight < 5 {
			logHeight = 5
		}
	}

	// Main content area (task list + details)
	mainHeight := contentHeight - logHeight
	if mainHeight < 5 {
		mainHeight = 5
	}

	// Calculate widths
	taskListWidth := m.width / 2
	detailsWidth := m.width / 2

	if !m.showDetailsPanel {
		taskListWidth = m.width
		detailsWidth = 0
	} else {
		// Ensure minimum widths
		if taskListWidth < minPanelWidth {
			taskListWidth = minPanelWidth
		}
		if detailsWidth < minPanelWidth {
			detailsWidth = minPanelWidth
		}

		// Adjust if total is too wide
		if taskListWidth+detailsWidth > m.width {
			taskListWidth = m.width / 2
			detailsWidth = m.width - taskListWidth
		}
	}

	layout.ContentWidth = m.width
	layout.ContentHeight = contentHeight
	layout.TaskListWidth = taskListWidth
	layout.TaskListHeight = mainHeight
	layout.DetailsWidth = detailsWidth
	layout.DetailsHeight = mainHeight
	
	// Calculate LogWidth with proper constraints to ensure it never exceeds terminal width
	logWidth := m.width
	const minMargin = 2
	if logWidth > m.width-minMargin {
		logWidth = m.width - minMargin
	}
	if logWidth < 20 { // minimum usable width
		logWidth = 20
	}
	layout.LogWidth = logWidth
	layout.LogHeight = logHeight

	return layout
}

// abbreviateTag shortens long tags to fit narrow terminals
// Tags longer than tagMaxLength characters are truncated with ellipsis
func abbreviateTag(tag string, maxLength int) string {
	if len(tag) <= maxLength {
		return tag
	}
	// Calculate prefix length (leave room for ellipsis)
	prefixLen := maxLength - len(tagEllipsis)
	if prefixLen < 1 {
		prefixLen = 1
	}
	return tag[:prefixLen] + tagEllipsis
}

// renderHeader renders the header with application name, tag, and progress.
// It calculates task progress, resolves the active tag from configuration,
// and selects an appropriate layout (wide, medium, or narrow) based on terminal width.
func (m Model) renderHeader() string {
	// Calculate progress
	done, total, percentage := m.calculateTaskProgress()
	
	// Get active tag or default
	activeTag := "master" // Default
	if m.config != nil && m.config.ActiveTag != "" {
		activeTag = m.config.ActiveTag
	}
	
	// Determine layout based on terminal width
	if m.width >= 100 {
		return m.renderWideHeader(activeTag, done, total, percentage)
	} else if m.width >= 80 {
		return m.renderMediumHeader(activeTag, done, total, percentage)
	} else {
		return m.renderNarrowHeader(activeTag, percentage)
	}
}

// renderWideHeader renders a full-featured header for wide terminals (≥100 columns)
// Displays application name, tag, and detailed progress in a bordered layout
func (m Model) renderWideHeader(tag string, done, total int, percentage float64) string {
	// Application name in bold
	appName := m.styles.Title.Render("Task Master TUI")

	// Tag with brackets and highlight color
	tagDisplay := m.styles.Highlight.Render(fmt.Sprintf("[%s]", tag))
	tagInfo := lipgloss.JoinHorizontal(lipgloss.Left,
		m.styles.Subtle.Render("Tag: "),
		tagDisplay,
	)

	// Progress text and bar
	progressText := fmt.Sprintf("Progress: %d/%d (%.0f%%)", done, total, percentage)
	progressBar := m.renderProgressBar(percentage, 10) // 10-character bar
	progressDisplay := lipgloss.JoinHorizontal(lipgloss.Left,
		m.styles.Info.Render(progressText),
		" ",
		progressBar,
	)

	// Combine all components with separators
	headerContent := lipgloss.JoinHorizontal(lipgloss.Left,
		appName,
		m.styles.Subtle.Render(" │ "),
		tagInfo,
		m.styles.Subtle.Render(" │ "),
		progressDisplay,
	)

	// Apply border style
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Bold(true).
		Padding(0, 1).
		Width(m.width - 2). // Account for border width
		Align(lipgloss.Left)

	return headerStyle.Render(headerContent)
}

// renderMediumHeader renders compact header for medium terminals (80-99 columns)
// Displays application name, tag, and simplified progress with a rounded border
func (m Model) renderMediumHeader(tag string, done, total int, percentage float64) string {
	// Application name in bold
	appName := m.styles.Title.Bold(true).Render("Task Master TUI")

	// Tag with brackets and highlight color (no "Tag:" prefix)
	tagDisplay := m.styles.Highlight.Render(fmt.Sprintf("[%s]", tag))

	// Shorter progress format - just percentage
	progressText := fmt.Sprintf("%.0f%%", percentage)
	progressBar := m.renderProgressBar(percentage, 5) // 5-character compact bar

	// Combine all components with separators
	headerContent := lipgloss.JoinHorizontal(lipgloss.Left,
		appName,
		m.styles.Subtle.Render(" │ "),
		tagDisplay,
		m.styles.Subtle.Render(" │ "),
		m.styles.Info.Render(progressText),
		" ",
		progressBar,
	)

	// Apply border style - rounded border for medium terminals
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Bold(true).
		Padding(0, 1).
		Width(m.width - 2). // Account for border width
		Align(lipgloss.Left)

	return headerStyle.Render(headerContent)
}

// renderNarrowHeader renders minimal header for narrow terminals (<80 columns)
// Displays abbreviated app name, abbreviated tag, and percentage only with normal border
func (m Model) renderNarrowHeader(tag string, percentage float64) string {
	// Abbreviated application name
	appName := m.styles.Title.Bold(true).Render("TM-TUI")

	// Shorten tag if too long
	shortTag := abbreviateTag(tag, tagMaxLength)
	tagDisplay := m.styles.Highlight.Render(fmt.Sprintf("[%s]", shortTag))

	// Progress text - percentage only (no progress bar)
	progressText := fmt.Sprintf("%.0f%%", percentage)

	// Combine all components with minimal separators
	headerContent := lipgloss.JoinHorizontal(lipgloss.Left,
		appName,
		m.styles.Subtle.Render(" │ "),
		tagDisplay,
		m.styles.Subtle.Render(" │ "),
		m.styles.Info.Render(progressText),
	)

	// Apply border style - normal border for narrow terminals
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(0, 1).
		Width(m.width - 2). // Account for border width
		Align(lipgloss.Left)

	return headerStyle.Render(headerContent)
}

func (m Model) activeProjectStatus() string {
	name := ""
	if m.activeProject != nil && strings.TrimSpace(m.activeProject.Name) != "" {
		name = m.activeProject.Name
	} else if m.config != nil && strings.TrimSpace(m.config.TaskMasterPath) != "" {
		name = filepath.Base(strings.TrimSpace(m.config.TaskMasterPath))
	}
	tag := ""
	if m.config != nil {
		tag = strings.TrimSpace(m.config.ActiveTag)
	}
	if name == "" && tag == "" {
		return ""
	}
	if name == "" {
		name = "Project"
	}
	if tag != "" {
		return fmt.Sprintf("Active: %s [%s]", name, tag)
	}
	return fmt.Sprintf("Active: %s", name)
}

// formatGitInfo formats git status information for display in the status bar
func (m Model) formatGitInfo() string {
	if !m.gitAvailable {
		return ""
	}

	if !m.gitRepoInfo.IsRepo {
		// Show "No Git repo" in gray when git is available but not in a repo
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4C566A")).
			Render("[No Git repo]")
	}

	status := m.GitStatus()
	if status.Branch == "" {
		return ""
	}

	// Format branch name with dirty indicator
	branchInfo := status.Branch
	if status.IsDirty {
		branchInfo += "*"
	}

	// Add ahead/behind counts if upstream exists
	if status.HasUpstream {
		aheadBehind := ""
		if status.Behind > 0 {
			aheadBehind += fmt.Sprintf("↓%d", status.Behind)
		}
		if status.Ahead > 0 {
			aheadBehind += fmt.Sprintf("↑%d", status.Ahead)
		}
		if aheadBehind != "" {
			branchInfo += " " + aheadBehind
		}
	} else {
		branchInfo += " (no upstream)"
	}

	// Apply green styling for repo info
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A3BE8C")).
		Render(fmt.Sprintf("[%s]", branchInfo))
}

// renderStatusBar renders the bottom status bar with keyboard hints
func (m Model) renderStatusBar() string {
	// Show confirmation prompt if confirming clear state
	if m.confirmingClearState {
		prompt := "Clear TUI state? (y/n): "
		return m.styles.StatusBar.Width(m.width).Render(prompt)
	}

	// Show command input if in command mode
	if m.commandMode {
		prompt := fmt.Sprintf("Jump to task ID: %s", m.commandInput)
		return m.styles.StatusBar.Width(m.width).Render(prompt)
	}

	// Show search mode status if in search mode
	if m.searchMode {
		searchStatus := fmt.Sprintf("🔍 SEARCH MODE: Type search query and press Enter, or Esc to cancel")
		return m.styles.StatusBar.
			Foreground(lipgloss.Color("#00FFFF")). // ColorHighlight
			Width(m.width).
			Render(searchStatus)
	}

	// Normal help text
	helpText := m.helpModel.ShortHelpView(m.keyMap.ShortHelp())
	if active := m.activeProjectStatus(); active != "" {
		helpText = fmt.Sprintf("%s | %s", helpText, active)
	}

	// Integrate git info into status bar
	gitInfo := m.formatGitInfo()
	if gitInfo != "" {
		statusContent := lipgloss.JoinHorizontal(lipgloss.Left, gitInfo, " ", helpText)
		return m.styles.StatusBar.Width(m.width).Render(statusContent)
	}

	return m.styles.StatusBar.Width(m.width).Render(helpText)
}

// updateViewportSizes updates the viewport sizes based on current layout
func (m *Model) updateViewportSizes() {
	layout := m.calculateLayout()

	// Update task list viewport
	m.taskListViewport.Width = layout.TaskListWidth - panelPadding*2
	m.taskListViewport.Height = layout.TaskListHeight - panelPadding

	// Update details viewport
	if m.showDetailsPanel {
		m.detailsViewport.Width = layout.DetailsWidth - panelPadding*2
		m.detailsViewport.Height = layout.DetailsHeight - panelPadding
	}

	// Update log viewport
	if m.showLogPanel {
		m.logViewport.Width = layout.LogWidth - panelPadding*2
		m.logViewport.Height = layout.LogHeight - panelPadding
		// Ensure viewport width is never less than 20
		if m.logViewport.Width < 20 {
			m.logViewport.Width = 20
		}
	}

	// Update task runner modal dimensions
	if m.taskRunner != nil {
		// Modal should take up most of the screen but leave some border
		modalWidth := m.width - 4
		modalHeight := m.height - 4
		if modalWidth < 40 {
			modalWidth = 40
		}
		if modalHeight < 10 {
			modalHeight = 10
		}
		m.taskRunner.SetRect(2, 2, modalWidth, modalHeight)
	}

	// Refresh task list viewport content after resize
	m.updateTaskListViewport()
}

// renderProgressBar creates a visual progress indicator using block characters.
// Returns empty string if width is too narrow (<5 characters) for a meaningful bar.
func (m Model) renderProgressBar(percentage float64, width int) string {
	// Too narrow for meaningful bar
	if width < 5 {
		return ""
	}

	// Calculate filled portion of the bar
	filledWidth := int(float64(width) * percentage / 100.0)
	if filledWidth < 0 {
		filledWidth = 0
	}
	if filledWidth > width {
		filledWidth = width
	}

	// Build progress bar with filled (▓) and empty (░) portions
	empty := width - filledWidth
	bar := strings.Repeat("▓", filledWidth) + strings.Repeat("░", empty)

	return bar
}
