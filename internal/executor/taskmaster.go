package executor

import (
	"fmt"
	"strings"
)

// TaskMasterExecutor provides methods for executing task-master commands
type TaskMasterExecutor struct {
	service *Service
}

// NewTaskMasterExecutor creates a new TaskMasterExecutor with the provided Service
func NewTaskMasterExecutor(service *Service) *TaskMasterExecutor {
	return &TaskMasterExecutor{
		service: service,
	}
}

// UpdateTask executes the appropriate task-master update command
// Detects if the task ID is a subtask (contains dots) or a main task,
// and uses the appropriate command type.
// Returns an error if the command fails to start.
func (e *TaskMasterExecutor) UpdateTask(taskID, updateContent string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Determine if this is a subtask (contains dots) or a main task
	isSubtask := strings.Contains(taskID, ".")

	// Prepare command type
	cmdType := "update-task"
	if isSubtask {
		cmdType = "update-subtask"
	}

	// Build arguments
	args := []string{cmdType, fmt.Sprintf("--id=%s", taskID)}

	// Only add prompt flag if updateContent is not empty or whitespace-only
	if strings.TrimSpace(updateContent) != "" {
		args = append(args, fmt.Sprintf("--prompt=%s", escapeShellArg(updateContent)))
	}

	// Execute the command asynchronously via the service
	return e.service.Execute(args[0], args[1:]...)
}

// GetOutput returns the output channel from the underlying service
// This channel receives streamed output lines from the command execution
func (e *TaskMasterExecutor) GetOutput() <-chan string {
	return e.service.GetOutput()
}

// GetDone returns the completion channel from the underlying service
// This channel receives a CommandResult when the command completes
func (e *TaskMasterExecutor) GetDone() <-chan CommandResult {
	return e.service.GetDone()
}

// IsRunning returns whether a command is currently executing
func (e *TaskMasterExecutor) IsRunning() bool {
	return e.service.IsRunning()
}

// escapeShellArg properly escapes a string for shell execution
// Uses single quotes and escapes any single quotes within the argument
func escapeShellArg(arg string) string {
	// Use single quotes and escape single quotes by ending the single-quoted string,
	// adding an escaped single quote, and starting a new single-quoted string
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}
