package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

// Message types for the demo application

// ModelSelectedMsg is sent when a model is selected
type ModelSelectedMsg struct {
	Model string
}

// TaskExecutionStartedMsg is sent when a task starts executing
type TaskExecutionStartedMsg struct {
	TaskID    string
	TaskTitle string
	Model     string
}

// readOutputMsg is used to read next output from a task
type readOutputMsg struct {
	taskID string
	runner *dialog.MockTaskRunner
}

// TaskOutputReceivedMsg wraps task output for Bubble Tea
type TaskOutputReceivedMsg dialog.TaskOutputMsg

// DemoState represents different states of the demo
type DemoState int

const (
	StateTaskSelection DemoState = iota
	StateModelSelection
	StateRunning
	StateHelp
	StateExit
)

// MockTask represents a task in the demo
type MockTask struct {
	ID       string
	Title    string
	Scenario dialog.MockScenario
}

// DemoApp is the main model for the demo application
type DemoApp struct {
	state              DemoState
	tasks              []MockTask
	selectedTaskIdx    int
	selectedModel      string
	taskRunner         *dialog.TaskRunnerModal
	mockRunners        map[string]*dialog.MockTaskRunner
	width              int
	height             int
	showHelp           bool
	runningTaskCount   int
	completedTaskCount int
}

// NewDemoApp creates a new demo application
func NewDemoApp(width, height int) *DemoApp {
	tasks := []MockTask{
		{ID: "task-1", Title: "Quick Build", Scenario: dialog.ScenarioQuickSuccess},
		{ID: "task-2", Title: "Long Compilation", Scenario: dialog.ScenarioLongRunning},
		{ID: "task-3", Title: "Build with Error", Scenario: dialog.ScenarioWithError},
		{ID: "task-4", Title: "Build with Warnings", Scenario: dialog.ScenarioWithWarnings},
		{ID: "task-5", Title: "Data Processing", Scenario: dialog.ScenarioLongRunning},
		{ID: "task-6", Title: "Test Suite", Scenario: dialog.ScenarioWithWarnings},
		{ID: "task-7", Title: "Deploy", Scenario: dialog.ScenarioQuickSuccess},
		{ID: "task-8", Title: "Validation", Scenario: dialog.ScenarioWithError},
		{ID: "task-9", Title: "Performance Test", Scenario: dialog.ScenarioLongRunning},
	}

	return &DemoApp{
		state:            StateTaskSelection,
		tasks:            tasks,
		selectedTaskIdx:  0,
		selectedModel:    "claude-3.5-sonnet",
		taskRunner:       dialog.NewTaskRunnerModal(width, height, nil),
		mockRunners:      make(map[string]*dialog.MockTaskRunner),
		width:            width,
		height:           height,
		showHelp:         false,
		runningTaskCount: 0,
	}
}

// Init initializes the demo
func (m *DemoApp) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *DemoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.taskRunner.SetRect(m.width, m.height, 0, 0)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case readOutputMsg:
		return m.handleReadOutput(msg)

	case dialog.TaskOutputMsg:
		m.taskRunner.Update(msg)
		return m, nil

	case dialog.TaskCompletedMsg:
		m.taskRunner.Update(msg)
		m.runningTaskCount--
		m.completedTaskCount++
		if mock, ok := m.mockRunners[msg.TaskID]; ok {
			delete(m.mockRunners, msg.TaskID)
			mock.Cancel()
		}
		return m, nil

	case dialog.TaskFailedMsg:
		m.taskRunner.Update(msg)
		m.runningTaskCount--
		if mock, ok := m.mockRunners[msg.TaskID]; ok {
			delete(m.mockRunners, msg.TaskID)
			mock.Cancel()
		}
		return m, nil

	case dialog.TaskCancelledMsg:
		m.taskRunner.Update(msg)
		m.runningTaskCount--
		if mock, ok := m.mockRunners[msg.TaskID]; ok {
			delete(m.mockRunners, msg.TaskID)
			mock.Cancel()
		}
		return m, nil
	}

	return m, nil
}

// handleReadOutput processes output from a task runner
func (m *DemoApp) handleReadOutput(msg readOutputMsg) (tea.Model, tea.Cmd) {
	select {
	case output, ok := <-msg.runner.OutputChan():
		if !ok {
			// Channel closed, task is done
			m.taskRunner.Update(dialog.TaskCompletedMsg{TaskID: msg.taskID})
			m.runningTaskCount--
			m.completedTaskCount++
			if mock, ok := m.mockRunners[msg.taskID]; ok {
				delete(m.mockRunners, msg.taskID)
				mock.Cancel()
			}
			return m, nil
		}
		// Send output to task runner modal
		m.taskRunner.Update(output)
		// Continue listening for more output
		return m, m.listenToTaskOutput(msg.taskID, msg.runner)
	default:
		// No output available yet, check again soon
		return m, tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
			return msg
		})
	}
}

// handleKeyMsg handles keyboard input
func (m *DemoApp) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateTaskSelection:
		return m.handleTaskSelectionKeys(msg)
	case StateModelSelection:
		return m.handleModelSelectionKeys(msg)
	case StateRunning:
		return m.handleRunningKeys(msg)
	case StateHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// handleTaskSelectionKeys handles keys in task selection state
func (m *DemoApp) handleTaskSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedTaskIdx > 0 {
			m.selectedTaskIdx--
		}
	case "down", "j":
		if m.selectedTaskIdx < len(m.tasks)-1 {
			m.selectedTaskIdx++
		}
	case "enter", "r":
		m.state = StateModelSelection
	case "?":
		m.showHelp = true
		m.state = StateHelp
	case "q", "ctrl+c":
		return m, tea.Quit
	case "a":
		// Run all tasks at once
		m.state = StateRunning
		cmds := m.runMultipleTasks()
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleModelSelectionKeys handles keys in model selection state
func (m *DemoApp) handleModelSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	models := []string{
		"claude-3.5-sonnet",
		"gpt-4-turbo",
		"gpt-4-mini",
		"claude-3-opus",
	}

	idx := -1
	for i, model := range models {
		if model == m.selectedModel {
			idx = i
			break
		}
	}

	switch msg.String() {
	case "up", "k":
		if idx > 0 {
			m.selectedModel = models[idx-1]
		}
	case "down", "j":
		if idx < len(models)-1 {
			m.selectedModel = models[idx+1]
		}
	case "enter":
		// Execute the selected task with the selected model
		m.state = StateRunning
		cmd := m.executeTask(m.selectedTaskIdx)
		return m, cmd
	case "esc":
		m.state = StateTaskSelection
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handleRunningKeys handles keys while tasks are running
func (m *DemoApp) handleRunningKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// Switch between task tabs
		if m.taskRunner.GetTabCount() > 0 {
			// Handled by task runner modal
		}
	case "m":
		// Minimize/maximize modal
		// Handled by task runner modal
	case "ctrl+c":
		// Cancel active task
		// Handled by task runner modal
	case "esc":
		// Check if all tasks are done before closing
		if m.runningTaskCount == 0 && m.taskRunner.GetTabCount() > 0 {
			m.state = StateTaskSelection
		}
	case "?":
		m.showHelp = true
		m.state = StateHelp
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

// handleHelpKeys handles keys in help state
func (m *DemoApp) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.showHelp = false
		if m.runningTaskCount > 0 {
			m.state = StateRunning
		} else {
			m.state = StateTaskSelection
		}
	}
	return m, nil
}

// executeTask starts a selected task
func (m *DemoApp) executeTask(taskIdx int) tea.Cmd {
	if taskIdx < 0 || taskIdx >= len(m.tasks) {
		return nil
	}

	task := m.tasks[taskIdx]

	// Send task started message
	m.taskRunner.Update(dialog.TaskStartedMsg{
		TaskID:    task.ID,
		TaskTitle: task.Title,
		Model:     m.selectedModel,
	})

	// Create and start mock runner
	mockRunner := dialog.NewMockTaskRunner(task.ID, taskIdx, task.Scenario)
	mockRunner.Start()
	m.mockRunners[task.ID] = mockRunner

	m.runningTaskCount++

	// Return command to collect output
	return m.listenToTaskOutput(task.ID, mockRunner)
}

// listenToTaskOutput creates a command that listens to task output
func (m *DemoApp) listenToTaskOutput(taskID string, runner *dialog.MockTaskRunner) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return readOutputMsg{taskID: taskID, runner: runner}
	})
}

// runMultipleTasks starts multiple tasks concurrently
func (m *DemoApp) runMultipleTasks() []tea.Cmd {
	maxTasks := 9
	if len(m.tasks) < maxTasks {
		maxTasks = len(m.tasks)
	}

	var cmds []tea.Cmd
	for i := 0; i < maxTasks; i++ {
		cmd := m.executeTask(i)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

func (m *DemoApp) View() string {
	switch m.state {
	case StateTaskSelection:
		return m.renderTaskSelection()
	case StateModelSelection:
		return m.renderModelSelection()
	case StateRunning:
		return m.renderRunning()
	case StateHelp:
		return m.renderHelp()
	}
	return ""
}

// renderTaskSelection renders the task selection screen
func (m *DemoApp) renderTaskSelection() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(1, 2)

	title := titleStyle.Render("Task Runner Demo - Task Selection")

	taskList := ""
	for i, task := range m.tasks {
		marker := "  "
		if i == m.selectedTaskIdx {
			marker = "▶ "
		}
		scenario := m.getScenarioEmoji(task.Scenario)
		taskList += fmt.Sprintf("%s[%d] %s %s\n", marker, i+1, scenario, task.Title)
	}

	taskListBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4A7C9E")).
		Padding(1, 2).
		Render(taskList)

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2).
		Render("↑/↓: Navigate | Enter/R: Select | A: Run All | ?: Help | Q: Quit")

	stats := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Padding(1, 2).
		Render(fmt.Sprintf("Running: %d | Completed: %d", m.runningTaskCount, m.completedTaskCount))

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		taskListBox,
		stats,
		helpText,
	)
}

// getScenarioEmoji returns an emoji for a scenario type
func (m *DemoApp) getScenarioEmoji(scenario dialog.MockScenario) string {
	switch scenario {
	case dialog.ScenarioQuickSuccess:
		return "✅"
	case dialog.ScenarioLongRunning:
		return "⏳"
	case dialog.ScenarioWithError:
		return "❌"
	case dialog.ScenarioWithWarnings:
		return "⚠️"
	default:
		return "❓"
	}
}

// renderModelSelection renders the model selection screen
func (m *DemoApp) renderModelSelection() string {
	models := []string{
		"claude-3.5-sonnet",
		"gpt-4-turbo",
		"gpt-4-mini",
		"claude-3-opus",
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(1, 2)

	title := titleStyle.Render("Task Runner Demo - Model Selection")

	selectedTask := m.tasks[m.selectedTaskIdx]
	taskInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Padding(1, 2).
		Render(fmt.Sprintf("Task: [%s] %s", selectedTask.ID, selectedTask.Title))

	modelList := ""
	for i, model := range models {
		marker := "  "
		if model == m.selectedModel {
			marker = "▶ "
		}
		modelList += fmt.Sprintf("%s[%d] %s\n", marker, i+1, model)
	}

	modelListBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4A7C9E")).
		Padding(1, 2).
		Render(modelList)

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2).
		Render("↑/↓: Navigate | Enter: Run | Esc: Back | Q: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		taskInfo,
		modelListBox,
		helpText,
	)
}

// renderRunning renders the running state with task runner modal
func (m *DemoApp) renderRunning() string {
	return m.taskRunner.View()
}

// renderHelp renders the help screen
func (m *DemoApp) renderHelp() string {
	helpTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(1, 2).
		Render("Task Runner Demo - Help")

	helpContent := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Padding(1, 2).
		Render(`
KEYBOARD SHORTCUTS:

Task Selection:
  ↑/↓, j/k    Navigate tasks
  Enter, R    Run selected task
  A           Run all 9 tasks
  ?           Show help
  Q, Ctrl+C   Quit

Model Selection:
  ↑/↓, j/k    Navigate models
  Enter       Execute with model
  Esc         Go back
  Q, Ctrl+C   Quit

Running Tasks:
  Tab         Next task tab
  Shift+Tab   Previous tab
  1-9         Jump to tab
  M           Minimize/maximize
  Ctrl+C      Cancel active task
  Esc         Return (when done)
  ?, Q        Help/quit

DEMO SCENARIOS:

✅ Quick Build           Fast execution (~2s)
⏳ Long Compilation      Extended run (~35s)
❌ Build with Error      Fails at step 4
⚠️ Build with Warnings   Succeeds with warnings

FEATURES:

✓ Multiple concurrent tasks
✓ Tab-based task organization
✓ Real-time output display
✓ Task state visualization
✓ Minimize/expand modal
✓ Model selection dialog

Press ESC or ? to close help...`)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		helpTitle,
		helpContent,
	)
}

func main() {
	p := tea.NewProgram(NewDemoApp(80, 30), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
