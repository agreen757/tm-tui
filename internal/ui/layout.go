package ui

import (
	"fmt"
	"strings"
	
	"github.com/adriangreen/tm-tui/internal/taskmaster"
)

const (
	headerHeight    = 3
	statusBarHeight = 1
	panelPadding    = 2
	minPanelWidth   = 20
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
	layout.LogWidth = m.width
	layout.LogHeight = logHeight
	
	return layout
}

// renderHeader renders the header with project name and task counts
func (m Model) renderHeader() string {
	// Get project name from config or default
	projectName := "Task Master"
	if m.config != nil && m.config.TaskMasterPath != "" {
		// Extract directory name from path
		parts := strings.Split(strings.TrimSuffix(m.config.TaskMasterPath, "/"), "/")
		if len(parts) > 0 {
			projectName = parts[len(parts)-1]
		}
	}
	
	// Count tasks by status
	counts := make(map[string]int)
	var countTasks func(tasks []taskmaster.Task)
	countTasks = func(tasks []taskmaster.Task) {
		for _, task := range tasks {
			counts[task.Status]++
			if len(task.Subtasks) > 0 {
				countTasks(task.Subtasks)
			}
		}
	}
	countTasks(m.tasks)
	
	// Build status counts string with colors
	var statusParts []string
	
	if count := counts[taskmaster.StatusPending]; count > 0 {
		statusParts = append(statusParts, m.styles.Pending.Render(fmt.Sprintf("%d pending", count)))
	}
	if count := counts[taskmaster.StatusInProgress]; count > 0 {
		statusParts = append(statusParts, m.styles.InProgress.Render(fmt.Sprintf("%d in-progress", count)))
	}
	if count := counts[taskmaster.StatusDone]; count > 0 {
		statusParts = append(statusParts, m.styles.Done.Render(fmt.Sprintf("%d done", count)))
	}
	if count := counts[taskmaster.StatusBlocked]; count > 0 {
		statusParts = append(statusParts, m.styles.Blocked.Render(fmt.Sprintf("%d blocked", count)))
	}
	if count := counts[taskmaster.StatusDeferred]; count > 0 {
		statusParts = append(statusParts, m.styles.Deferred.Render(fmt.Sprintf("%d deferred", count)))
	}
	if count := counts[taskmaster.StatusCancelled]; count > 0 {
		statusParts = append(statusParts, m.styles.Cancelled.Render(fmt.Sprintf("%d cancelled", count)))
	}
	
	statusLine := strings.Join(statusParts, " | ")
	
	// Build header
	titleLine := m.styles.Header.Width(m.width).Render(m.styles.Title.Render(projectName))
	countsLine := m.styles.StatusBar.Width(m.width).Render(statusLine)
	
	// Add search input if in search mode
	if m.searchMode {
		searchLabel := m.styles.Info.Render("Search: ")
		searchBox := m.searchInput.View()
		searchHelp := m.styles.Subtle.Render(" (Enter to search, Esc to cancel)")
		searchLine := searchLabel + searchBox + searchHelp
		return titleLine + "\n" + countsLine + "\n" + searchLine + "\n"
	}
	
	// Build filter/search status line
	var filterParts []string
	
	// Show status filter if active
	if m.statusFilter != "" {
		filterInfo := m.styles.Warning.Render(fmt.Sprintf("📌 Filter: %s", m.statusFilter))
		filterParts = append(filterParts, filterInfo)
	}
	
	// Show search query if active
	if m.searchQuery != "" {
		resultsCount := 0
		if m.searchResults != nil {
			resultsCount = len(m.searchResults)
		}
		searchInfo := m.styles.Info.Render(fmt.Sprintf("🔍 Search: '%s' (%d results)", m.searchQuery, resultsCount))
		filterParts = append(filterParts, searchInfo)
	}
	
	if len(filterParts) > 0 {
		filterLine := strings.Join(filterParts, " | ")
		clearHint := m.styles.Subtle.Render(" (F=filter, /=search, Esc=clear)")
		return titleLine + "\n" + countsLine + "\n" + filterLine + clearHint + "\n"
	}
	
	return titleLine + "\n" + countsLine + "\n"
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
	
	// Normal help text
	helpText := m.helpModel.ShortHelpView(m.keyMap.ShortHelp())
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
	}
	
	// Refresh task list viewport content after resize
	m.updateTaskListViewport()
}
