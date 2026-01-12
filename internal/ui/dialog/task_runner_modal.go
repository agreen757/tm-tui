package dialog

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types for TaskRunnerModal

// TaskStartedMsg is sent when a task starts execution
type TaskStartedMsg struct {
	TaskID    string
	TaskTitle string
	Model     string
}

// TaskOutputMsg is sent when a task produces output
type TaskOutputMsg struct {
	TaskID string
	Output string
}

// TaskCompletedMsg is sent when a task completes successfully
type TaskCompletedMsg struct {
	TaskID string
}

// TaskFailedMsg is sent when a task fails
type TaskFailedMsg struct {
	TaskID  string
	Error   string
	Message string
}

// TaskCancelledMsg is sent when a task is cancelled
type TaskCancelledMsg struct {
	TaskID string
}

// TaskRunnerAutoCloseMsg is sent when the auto-close timer expires
type TaskRunnerAutoCloseMsg struct{}

// ModalMinimizedMsg is sent when the modal is minimized/maximized
type ModalMinimizedMsg struct {
	Minimized bool
}

// TaskRunnerKeyMap defines keybindings for the task runner modal
type TaskRunnerKeyMap struct {
	NextTab    key.Binding
	PrevTab    key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	ScrollTop  key.Binding
	ScrollBottom key.Binding
	Minimize   key.Binding
	Cancel     key.Binding
	Close      key.Binding
	TabDirect  []key.Binding // 1-9 keys
}

// DefaultTaskRunnerKeyMap returns the default keybindings
func DefaultTaskRunnerKeyMap() TaskRunnerKeyMap {
	directTabBindings := make([]key.Binding, 9)
	for i := 0; i < 9; i++ {
		keyStr := string(rune('1' + i))
		directTabBindings[i] = key.NewBinding(
			key.WithKeys(keyStr),
			key.WithHelp(keyStr, fmt.Sprintf("jump to tab %d", i+1)),
		)
	}

	return TaskRunnerKeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdn"),
			key.WithHelp("pgdn", "page down"),
		),
		ScrollTop: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "top"),
		),
		ScrollBottom: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "bottom"),
		),
		Minimize: key.NewBinding(
			key.WithKeys("m", "M"),
			key.WithHelp("m", "minimize"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "cancel"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		TabDirect: directTabBindings,
	}
}

// PreMinimizeState captures the state before minimizing the modal
type PreMinimizeState struct {
	activeTab     int
	scrollOffsets map[int]int // Maps tab index to viewport Y offset
}

// TaskRunnerModal is the main container for task execution
type TaskRunnerModal struct {
	BaseDialog
	tabs                      []*TaskExecutionTab
	activeTab                 int
	minimized                 bool
	preMinimizeState          *PreMinimizeState
	keyMap                    TaskRunnerKeyMap
	tabScrollPos              int // For handling tab bar overflow
	cancellationConfirmDialog Dialog
	pendingCancellationTabIdx int
	longRunningThreshold      int       // milliseconds to consider a task "long-running"
	autoCloseOnFailure        bool      // automatically close modal when all tasks fail/complete
	autoCloseDelay            time.Duration // delay before auto-closing
	closeTimer                *time.Time // tracks when to auto-close
	gitAutoCloseDelay         time.Duration // delay before auto-closing git operations (default 2s)
	gitAutoCloseTimers       map[string]*time.Time // tracks git-specific auto-close timers by task ID
}

// NewTaskRunnerModal creates a new task runner modal
func NewTaskRunnerModal(width, height int, style *DialogStyle) *TaskRunnerModal {
	if style == nil {
		style = DefaultDialogStyle()
	}

	modal := &TaskRunnerModal{
		BaseDialog:   NewBaseDialog("Task Runner", width, height, DialogKindCustom),
		tabs:         []*TaskExecutionTab{},
		activeTab:    0,
		minimized:    false,
		preMinimizeState: &PreMinimizeState{
			activeTab:     0,
			scrollOffsets: make(map[int]int),
		},
		keyMap:                   DefaultTaskRunnerKeyMap(),
		tabScrollPos:             0,
		cancellationConfirmDialog: nil,
		pendingCancellationTabIdx: -1,
		longRunningThreshold:     5000,               // 5 seconds
		autoCloseOnFailure:       true,               // auto-close when all tasks stop
		autoCloseDelay:           3 * time.Second,    // wait 3 seconds before closing
		closeTimer:               nil,
		gitAutoCloseDelay:        2 * time.Second,    // git operations auto-close after 2 seconds
		gitAutoCloseTimers:       make(map[string]*time.Time), // per-task timers for git operations
	}
	modal.Style = style
	modal.SetFocused(true)

	return modal
}

// Init satisfies Dialog interface
func (m *TaskRunnerModal) Init() tea.Cmd {
	return nil
}

// Update processes messages
func (m *TaskRunnerModal) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	// Handle cancellation confirmation dialog
	if m.cancellationConfirmDialog != nil {
		updatedDialog, cmd := m.cancellationConfirmDialog.Update(msg)
		if updatedDialog == nil {
			// Dialog was closed
			m.cancellationConfirmDialog = nil
			m.pendingCancellationTabIdx = -1
			return m, cmd
		}

		if confirmMsg, ok := msg.(ConfirmationMsg); ok {
			if confirmMsg.Result == ConfirmationResultYes {
				// Perform the cancellation
				m.performCancellation(m.pendingCancellationTabIdx)
			}
			m.cancellationConfirmDialog = nil
			m.pendingCancellationTabIdx = -1
			return m, cmd
		}

		m.cancellationConfirmDialog = updatedDialog
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		result, cmd := m.HandleKey(msg)
		if result == DialogResultClose {
			return nil, cmd
		}
		return m, cmd
	case TaskStartedMsg:
		m.addTab(msg.TaskID, msg.TaskTitle, msg.Model)
		// Cancel any pending auto-close when a new task starts
		m.closeTimer = nil
	case TaskOutputMsg:
		// Route output to the correct tab by TaskID
		for _, tab := range m.tabs {
			if tab.taskID == msg.TaskID {
				tab.AddOutputLine(msg.Output)
				break
			}
		}
	case TaskCompletedMsg:
		m.setTabStatus(msg.TaskID, TaskCompleted)
		// For git tasks, use git-specific auto-close (2 seconds)
		if m.isGitTask(msg.TaskID) {
			return m, m.handleGitTaskCompletion(msg.TaskID)
		}
		// For other tasks, use normal auto-close logic
		return m, m.checkAutoClose()
	case TaskFailedMsg:
		m.setTabStatus(msg.TaskID, TaskFailed)
		// For git tasks, keep the modal open on failure
		if m.isGitTask(msg.TaskID) {
			m.handleGitTaskFailure(msg.TaskID)
			return m, nil
		}
		// For other tasks, use normal auto-close logic
		return m, m.checkAutoClose()
	case TaskCancelledMsg:
		m.setTabStatus(msg.TaskID, TaskCancelled)
		// For git tasks, clean up any pending timers
		if m.isGitTask(msg.TaskID) {
			delete(m.gitAutoCloseTimers, msg.TaskID)
		}
		return m, m.checkAutoClose()
	case TaskRunnerAutoCloseMsg:
		// Timer expired and no tasks are running - close the modal
		if !m.hasRunningTasks() && m.closeTimer != nil && time.Now().After(*m.closeTimer) {
			return nil, nil
		}
	case TaskRunnerGitAutoCloseMsg:
		// Git auto-close timer expired - close the modal only if no other tasks are running
		if timer, exists := m.gitAutoCloseTimers[msg.TaskID]; exists && timer != nil && time.Now().After(*timer) {
			delete(m.gitAutoCloseTimers, msg.TaskID)
			// Only close if no tasks are running
			if !m.hasRunningTasks() {
				return nil, nil
			}
		}
	}

	return m, nil
}

// HandleKey handles keyboard input
func (m *TaskRunnerModal) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Cancel any pending git auto-close timers on any key press
	if len(m.gitAutoCloseTimers) > 0 {
		m.gitAutoCloseTimers = make(map[string]*time.Time)
	}

	// Pass key to active tab for scrolling control
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		m.tabs[m.activeTab].Update(msg)
	}

	// Check for tab navigation
	if key.Matches(msg, m.keyMap.NextTab) {
		m.nextTab()
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.PrevTab) {
		m.prevTab()
		return DialogResultNone, nil
	}

	// Check for direct tab selection (1-9)
	for i, binding := range m.keyMap.TabDirect {
		if key.Matches(msg, binding) {
			if i < len(m.tabs) {
				m.activeTab = i
			}
			return DialogResultNone, nil
		}
	}

	// Check for scrolling controls
	if key.Matches(msg, m.keyMap.ScrollUp) {
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.ScrollDown) {
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.PageUp) {
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.PageDown) {
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.ScrollTop) {
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.ScrollBottom) {
		return DialogResultNone, nil
	}

	// Check for action shortcuts
	if key.Matches(msg, m.keyMap.Minimize) {
		m.ToggleMinimize()
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.Cancel) {
		m.cancelActiveTask()
		return DialogResultNone, nil
	}
	if key.Matches(msg, m.keyMap.Close) {
		if !m.hasRunningTasks() {
			return DialogResultClose, nil
		}
		return DialogResultNone, nil
	}

	return DialogResultNone, nil
}

// addTab creates a new execution tab
func (m *TaskRunnerModal) addTab(taskID, taskTitle, model string) {
	// Calculate tab width based on modal width
	width, height, _, _ := m.GetRect()
	tabWidth := width - 2
	tabHeight := height - 8 // Account for header, footer, and margins

	if tabHeight < 1 {
		tabHeight = 1
	}

	tab := NewTaskExecutionTab(taskID, taskTitle, model, tabWidth, tabHeight, m.Style)
	m.tabs = append(m.tabs, tab)
	m.activeTab = len(m.tabs) - 1
}

// setTabStatus updates the status of a tab by task ID
func (m *TaskRunnerModal) setTabStatus(taskID string, status TaskExecutionStatus) {
	for _, tab := range m.tabs {
		if tab.GetTaskID() == taskID {
			tab.SetStatus(status)
			return
		}
	}
}

// nextTab switches to the next tab
func (m *TaskRunnerModal) nextTab() {
	if len(m.tabs) > 0 {
		m.activeTab = (m.activeTab + 1) % len(m.tabs)
		m.ensureTabVisible()
	}
}

// prevTab switches to the previous tab
func (m *TaskRunnerModal) prevTab() {
	if len(m.tabs) > 0 {
		m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		m.ensureTabVisible()
	}
}

// closeTab removes a tab if it's not running
func (m *TaskRunnerModal) closeTab(index int) bool {
	if index < 0 || index >= len(m.tabs) {
		return false
	}

	tab := m.tabs[index]
	if tab.GetStatus() == TaskRunning {
		return false
	}

	m.tabs = append(m.tabs[:index], m.tabs[index+1:]...)
	if m.activeTab >= len(m.tabs) && m.activeTab > 0 {
		m.activeTab--
	}
	return true
}

// cancelActiveTask handles cancellation of the active task with optional confirmation
func (m *TaskRunnerModal) cancelActiveTask() {
	if m.activeTab < 0 || m.activeTab >= len(m.tabs) {
		return
	}

	tab := m.tabs[m.activeTab]

	// Only allow cancellation of running tasks
	if tab.GetStatus() != TaskRunning {
		return
	}

	// Check if task is long-running and needs confirmation
	elapsed := time.Since(tab.startTime).Milliseconds()
	if elapsed >= int64(m.longRunningThreshold) {
		m.showCancellationConfirmation(m.activeTab)
		return
	}

	// For quick tasks, cancel directly
	m.performCancellation(m.activeTab)
}

// showCancellationConfirmation displays a confirmation dialog for task cancellation
func (m *TaskRunnerModal) showCancellationConfirmation(tabIndex int) {
	if tabIndex < 0 || tabIndex >= len(m.tabs) {
		return
	}

	tab := m.tabs[tabIndex]
	message := fmt.Sprintf("Cancel task '%s'?\n\nThis may result in incomplete work being discarded.", tab.GetTaskTitle())

	confirmDialog := YesNo("Cancel Task", message, true)
	m.cancellationConfirmDialog = confirmDialog
	m.pendingCancellationTabIdx = tabIndex
}

// performCancellation executes the actual task cancellation
func (m *TaskRunnerModal) performCancellation(tabIndex int) {
	if tabIndex < 0 || tabIndex >= len(m.tabs) {
		return
	}

	tab := m.tabs[tabIndex]
	if tab.GetStatus() != TaskRunning {
		return
	}

	// Attempt to cancel the subprocess execution
	cancelled := tab.CancelExecution("User requested cancellation")

	// Add cancellation message to output
	tab.AddOutputLine("")
	if cancelled {
		tab.AddOutputLine("⊘ Task cancelled by user - terminating subprocess...")
	} else {
		tab.AddOutputLine("⊘ Task cancelled by user")
	}

	// Update status and record end time
	// Note: The subprocess may send TaskCancelledMsg which will also set status,
	// but we set it here immediately for UI feedback
	tab.SetStatus(TaskCancelled)
}

// hasRunningTasks checks if any task is still running
func (m *TaskRunnerModal) hasRunningTasks() bool {
	for _, tab := range m.tabs {
		if tab.GetStatus() == TaskRunning {
			return true
		}
	}
	return false
}

// isGitTask checks if a task ID is a git operation (prefixed with "git-")
func (m *TaskRunnerModal) isGitTask(taskID string) bool {
	return strings.HasPrefix(taskID, "git-")
}

// handleGitTaskCompletion manages auto-close for successful git operations
func (m *TaskRunnerModal) handleGitTaskCompletion(taskID string) tea.Cmd {
	if !m.isGitTask(taskID) {
		return nil
	}

	// For successful git operations, start a 2-second auto-close timer
	closeTime := time.Now().Add(m.gitAutoCloseDelay)
	m.gitAutoCloseTimers[taskID] = &closeTime

	// Schedule the close message for this git task
	return tea.Tick(m.gitAutoCloseDelay, func(time.Time) tea.Msg {
		return TaskRunnerGitAutoCloseMsg{
			TaskID: taskID,
		}
	})
}

// handleGitTaskFailure keeps git operations open on failure
func (m *TaskRunnerModal) handleGitTaskFailure(taskID string) {
	if !m.isGitTask(taskID) {
		return
	}

	// Remove any pending auto-close timer for this git task
	delete(m.gitAutoCloseTimers, taskID)
}

// TaskRunnerGitAutoCloseMsg is sent when a git auto-close timer expires
type TaskRunnerGitAutoCloseMsg struct {
	TaskID string
}

// HasRunningTasks is the public interface for checking if any tasks are running
func (m *TaskRunnerModal) HasRunningTasks() bool {
	return m.hasRunningTasks()
}

// checkAutoClose schedules auto-close if enabled and no tasks are running
func (m *TaskRunnerModal) checkAutoClose() tea.Cmd {
	if !m.autoCloseOnFailure {
		return nil
	}

	if m.hasRunningTasks() {
		// Still have running tasks, cancel any pending close
		m.closeTimer = nil
		return nil
	}

	// No running tasks - start the auto-close timer
	if m.closeTimer == nil {
		closeTime := time.Now().Add(m.autoCloseDelay)
		m.closeTimer = &closeTime
	}

	// Schedule the close message
	return tea.Tick(m.autoCloseDelay, func(time.Time) tea.Msg {
		return TaskRunnerAutoCloseMsg{}
	})
}

// SetAutoCloseOnFailure enables or disables automatic closing when tasks stop
func (m *TaskRunnerModal) SetAutoCloseOnFailure(enabled bool) {
	m.autoCloseOnFailure = enabled
	if !enabled {
		m.closeTimer = nil
	}
}

// SetAutoCloseDelay sets the delay before auto-closing
func (m *TaskRunnerModal) SetAutoCloseDelay(delay time.Duration) {
	m.autoCloseDelay = delay
}

// SetGitAutoCloseDelay sets the delay before auto-closing git operations (default 2 seconds)
func (m *TaskRunnerModal) SetGitAutoCloseDelay(delay time.Duration) {
	m.gitAutoCloseDelay = delay
}

// ensureTabVisible adjusts scroll position so active tab is visible
func (m *TaskRunnerModal) ensureTabVisible() {
	// For simplicity, adjust scroll position if tab bar would overflow
	if m.activeTab < m.tabScrollPos {
		m.tabScrollPos = m.activeTab
	} else if m.activeTab >= m.tabScrollPos+7 { // Show at most 8 visible tabs
		m.tabScrollPos = m.activeTab - 6
	}
}

// ToggleMinimize toggles between minimized and maximized states
// Preserves active tab and scroll positions when minimizing/maximizing
func (m *TaskRunnerModal) ToggleMinimize() {
	if !m.minimized {
		// Save state before minimizing
		m.preMinimizeState.activeTab = m.activeTab
		m.preMinimizeState.scrollOffsets = make(map[int]int)
		for i, tab := range m.tabs {
			m.preMinimizeState.scrollOffsets[i] = tab.viewport.YOffset
		}
	} else {
		// Restore state after maximizing
		if m.preMinimizeState.activeTab >= 0 && m.preMinimizeState.activeTab < len(m.tabs) {
			m.activeTab = m.preMinimizeState.activeTab
			// Restore scroll positions for all tabs
			for i, tab := range m.tabs {
				if offset, exists := m.preMinimizeState.scrollOffsets[i]; exists {
					tab.viewport.SetYOffset(offset)
				}
			}
		}
	}
	m.minimized = !m.minimized
}

// View renders the modal
func (m *TaskRunnerModal) View() string {
	// If there's a confirmation dialog, render it on top
	if m.cancellationConfirmDialog != nil {
		baseView := m.renderModalContent()
		return lipgloss.JoinVertical(
			lipgloss.Center,
			baseView,
			m.cancellationConfirmDialog.View(),
		)
	}

	return m.renderModalContent()
}

// renderModalContent renders the main modal content
func (m *TaskRunnerModal) renderModalContent() string {
	if m.minimized {
		return m.renderMinimized()
	}

	// Render tab bar
	tabBar := m.renderTabBar()

	// Render active tab content
	var content string
	if len(m.tabs) > 0 && m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		content = m.tabs[m.activeTab].View()
	} else {
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Render("No tasks running. Start a task to see output here.")
	}

	// Render git auto-close countdown if active
	var countdownIndicator string
	if gitAutoCloseActive, timeRemaining := m.getGitAutoCloseInfo(); gitAutoCloseActive {
		countdownIndicator = m.renderGitAutoCloseCountdown(timeRemaining)
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			countdownIndicator,
		)
	}

	// Render footer
	footer := m.renderFooter()

	// Combine all parts
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
		footer,
	)

	// Apply border
	bordered := lipgloss.NewStyle().
		Border(m.Style.Border).
		BorderForeground(m.Style.BorderColor).
		Padding(1, 2).
		Render(modalContent)

	return bordered
}

// renderMinimized renders the minimized state
func (m *TaskRunnerModal) renderMinimized() string {
	content := "📊 Task Runner: "
	if len(m.tabs) == 0 {
		content += "No tasks"
	} else {
		runningCount := 0
		statusBar := ""
		maxTasksInBar := 9 // Limit to 9 task icons

		// Build status icons for running tasks
		for i, tab := range m.tabs {
			if i >= maxTasksInBar {
				break
			}
			statusBar += tab.getStatusIcon() + " "
		}

		if len(m.tabs) > maxTasksInBar {
			statusBar = statusBar[:len(statusBar)-1] + "..."
		} else if statusBar != "" {
			statusBar = statusBar[:len(statusBar)-1] // Remove trailing space
		}

		for _, tab := range m.tabs {
			if tab.GetStatus() == TaskRunning {
				runningCount++
			}
		}
		content += fmt.Sprintf("%d running | %d total | %s", runningCount, len(m.tabs), statusBar)
	}

	content += " | Press 'M' to expand"

	return lipgloss.NewStyle().
		Border(m.Style.Border).
		BorderForeground(m.Style.BorderColor).
		Padding(0, 1).
		Render(content)
}

// renderTabBar renders the tab bar showing all task tabs
func (m *TaskRunnerModal) renderTabBar() string {
	if len(m.tabs) == 0 {
		return ""
	}

	var tabStrings []string
	for i := 0; i < len(m.tabs); i++ {
		tab := m.tabs[i]
		indicator := tab.getStatusIcon()
		label := fmt.Sprintf("%s %s", indicator, tab.GetTabLabel())

		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Padding(0, 1)

		// Apply color based on status
		status := tab.GetStatus()
		switch status {
		case TaskCancelled:
			style = style.Foreground(lipgloss.Color("#FF9999"))
		case TaskFailed:
			style = style.Foreground(lipgloss.Color("#FF6666"))
		case TaskCompleted:
			style = style.Foreground(lipgloss.Color("#99FF99"))
		case TaskRunning:
			style = style.Foreground(lipgloss.Color("#FFFF99"))
		}

		// Highlight active tab
		if i == m.activeTab {
			style = style.
				Background(lipgloss.Color("#6D98BA")).
				Bold(true)
		}

		tabStrings = append(tabStrings, style.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabStrings...)
}

// getGitAutoCloseInfo checks if any git auto-close timers are active and returns remaining time
func (m *TaskRunnerModal) getGitAutoCloseInfo() (bool, time.Duration) {
	for _, timer := range m.gitAutoCloseTimers {
		if timer != nil {
			remaining := time.Until(*timer)
			if remaining > 0 {
				return true, remaining
			}
		}
	}
	return false, 0
}

// renderGitAutoCloseCountdown renders a visual countdown indicator for git auto-close
func (m *TaskRunnerModal) renderGitAutoCloseCountdown(timeRemaining time.Duration) string {
	seconds := int(timeRemaining.Seconds())
	if seconds < 0 {
		seconds = 0
	}

	// Create a visual countdown bar
	countdownText := fmt.Sprintf("⏱  Closing in %d second%s...", seconds, func() string {
		if seconds == 1 {
			return ""
		}
		return "s"
	}())

	// Style the countdown with an accent color
	countdownStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF99")). // Yellow accent
		Bold(true).
		Padding(0, 1)

	// Add instruction text
	instructionText := "Press any key to stay open"
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Padding(0, 1)

	// Combine countdown and instruction
	combined := lipgloss.JoinHorizontal(
		lipgloss.Center,
		countdownStyle.Render(countdownText),
		instructionStyle.Render(instructionText),
	)

	return combined
}

// renderFooter renders the footer with helpful shortcuts
func (m *TaskRunnerModal) renderFooter() string {
	shortcuts := []string{
		"Tab/Shift+Tab: switch",
		"1-9: jump",
		"↑/↓: scroll",
		"PgUp/PgDn: page",
		"M: minimize",
		"Ctrl+C: cancel",
	}

	// Only show close option if no running tasks
	if !m.hasRunningTasks() {
		// Check if any git auto-close timers are active
		gitAutoCloseActive := false
		var gitTimeRemaining time.Duration
		for _, timer := range m.gitAutoCloseTimers {
			if timer != nil {
				remaining := time.Until(*timer)
				if remaining > 0 {
					gitAutoCloseActive = true
					gitTimeRemaining = remaining
					break
				}
			}
		}

		// Show git countdown if active
		if gitAutoCloseActive && gitTimeRemaining > 0 {
			shortcuts = append(shortcuts, fmt.Sprintf("Any key to keep open, Esc to close (auto in %ds)", int(gitTimeRemaining.Seconds())+1))
		} else if m.closeTimer != nil && m.autoCloseOnFailure {
			// Show normal auto-close countdown
			remaining := time.Until(*m.closeTimer)
			if remaining > 0 {
				shortcuts = append(shortcuts, fmt.Sprintf("Esc: close (auto in %ds)", int(remaining.Seconds())+1))
			} else {
				shortcuts = append(shortcuts, "Esc: close")
			}
		} else {
			shortcuts = append(shortcuts, "Esc: close")
		}
	} else {
		shortcuts = append(shortcuts, "Esc: (running)")
	}

	// Condense help text for display
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		shortcuts...,
	)

	footerText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render(footerContent)

	return footerText
}

// SetRect sets the modal dimensions
func (m *TaskRunnerModal) SetRect(width, height, x, y int) {
	m.BaseDialog.SetRect(width, height, x, y)
	// Update all tabs with new dimensions
	tabWidth := width - 4
	tabHeight := height - 8

	if tabHeight < 1 {
		tabHeight = 1
	}

	for _, tab := range m.tabs {
		tab.SetRect(tabWidth, tabHeight)
	}
}

// GetActiveTab returns the currently active tab
func (m *TaskRunnerModal) GetActiveTab() *TaskExecutionTab {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		return m.tabs[m.activeTab]
	}
	return nil
}

// GetTabCount returns the number of tabs
func (m *TaskRunnerModal) GetTabCount() int {
	return len(m.tabs)
}

// GetMinimized returns whether the modal is minimized
func (m *TaskRunnerModal) GetMinimized() bool {
	return m.minimized
}

// IsCancellable returns whether the dialog can be cancelled (satisfies Dialog interface)
func (m *TaskRunnerModal) IsCancellable() bool {
	return m.BaseDialog.IsCancellable()
}

// CancelTaskByID attempts to cancel a specific task by its ID
func (m *TaskRunnerModal) CancelTaskByID(taskID string) bool {
	for i, tab := range m.tabs {
		if tab.GetTaskID() == taskID {
			if tab.GetStatus() == TaskRunning {
				m.performCancellation(i)
				return true
			}
		}
	}
	return false
}

// GetTaskStatus returns the status of a task by ID
func (m *TaskRunnerModal) GetTaskStatus(taskID string) (TaskExecutionStatus, bool) {
	for _, tab := range m.tabs {
		if tab.GetTaskID() == taskID {
			return tab.GetStatus(), true
		}
	}
	return -1, false
}

// GetTaskOutput returns the output of a task by ID
func (m *TaskRunnerModal) GetTaskOutput(taskID string) ([]string, bool) {
	for _, tab := range m.tabs {
		if tab.GetTaskID() == taskID {
			return tab.GetOutput(), true
		}
	}
	return nil, false
}

// GetCloseTimer returns the close timer (for checking if auto-close is pending)
func (m *TaskRunnerModal) GetCloseTimer() *time.Time {
	return m.closeTimer
}

// GetTabByTaskID returns a pointer to the tab with the given task ID, or nil if not found
func (m *TaskRunnerModal) GetTabByTaskID(taskID string) *TaskExecutionTab {
	for _, tab := range m.tabs {
		if tab.GetTaskID() == taskID {
			return tab
		}
	}
	return nil
}
