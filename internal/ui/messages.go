package ui

import (
	"context"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/executor"
)

// TasksLoadedMsg is sent when tasks are initially loaded
type TasksLoadedMsg struct {
	Tasks []taskmaster.Task
}

// TasksReloadedMsg is sent when tasks.json has been reloaded from disk
type TasksReloadedMsg struct{}

// ConfigReloadedMsg is sent when config files have been reloaded from disk
type ConfigReloadedMsg struct{}

// WatcherErrorMsg is sent when a file watcher encounters an error
type WatcherErrorMsg struct {
	Err error
}

// ExecutorOutputMsg is sent when the executor produces output
type ExecutorOutputMsg struct {
	Line string
}

// CommandCompletedMsg is sent when a command execution completes
type CommandCompletedMsg struct {
	Command string
	Success bool
	Output  string
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// WaitForTasksReload returns a command that waits for tasks to be reloaded
// and sends a TasksReloadedMsg when that happens
func WaitForTasksReload(service *taskmaster.Service) tea.Cmd {
	return func() tea.Msg {
		<-service.ReloadEvents()
		return TasksReloadedMsg{}
	}
}

// WaitForConfigReload returns a command that waits for config to be reloaded
// and sends a ConfigReloadedMsg when that happens
func WaitForConfigReload(manager *config.ConfigManager) tea.Cmd {
	return func() tea.Msg {
		<-manager.ReloadEvents()
		return ConfigReloadedMsg{}
	}
}

// WaitForExecutorOutput returns a command that listens for executor output
func WaitForExecutorOutput(service *executor.Service) tea.Cmd {
	return func() tea.Msg {
		// Get the output channel and wait for a line
		outputChan := service.GetOutput()
		line := <-outputChan
		return ExecutorOutputMsg{Line: line}
	}
}

// LoadTasksCmd loads tasks from disk and returns a TasksLoadedMsg
func LoadTasksCmd(service *taskmaster.Service) tea.Cmd {
	return func() tea.Msg {
		// Force load tasks initially
		ctx := context.WithValue(context.Background(), "force", true)
		if err := service.LoadTasks(ctx); err != nil {
			return ErrorMsg{Err: err}
		}
		
		tasks, _ := service.GetTasks()
		
		return TasksLoadedMsg{Tasks: tasks}
	}
}
