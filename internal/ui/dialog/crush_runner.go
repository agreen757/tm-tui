package dialog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

// CrushBinaryError represents an error when the crush binary is not found
type CrushBinaryError struct {
	Message string
}

func (e *CrushBinaryError) Error() string {
	return e.Message
}

// ValidateCrushBinary checks if the crush CLI binary is available
func ValidateCrushBinary() error {
	path, err := exec.LookPath("crush")
	if err != nil {
		return &CrushBinaryError{
			Message: "crush binary not found. Install via: go install github.com/crush-ai/crush@latest",
		}
	}
	// Verify it's executable
	if _, err := os.Stat(path); err != nil {
		return &CrushBinaryError{
			Message: fmt.Sprintf("crush binary found at %s but not accessible: %v", path, err),
		}
	}
	return nil
}

// CrushPromptContext holds the data for templating the Crush prompt
type CrushPromptContext struct {
	TaskID       string
	Title        string
	Description  string
	Details      string
	TestStrategy string
	Priority     string
	Dependencies string
}

const defaultWorkflowGuide = `# Task Execution Guide

You are executing Task {{.TaskID}}: {{.Title}}

## Description
{{.Description}}

## Implementation Details
{{.Details}}

## Test Strategy
{{.TestStrategy}}

## Priority
{{.Priority}}

{{if .Dependencies}}
## Dependencies
This task depends on: {{.Dependencies}}
{{end}}

## Instructions
1. Review the task description and implementation details
2. Follow the test strategy to ensure quality
3. Complete all requirements before finishing
4. Log your progress and any issues encountered
`

// GenerateCrushPrompt creates the prompt for Crush CLI execution
// It reads CRUSH_RUN_INSTRUCTIONS.md if available, otherwise uses a default template
func GenerateCrushPrompt(task *taskmaster.Task, model string) (string, error) {
	if task == nil {
		return "", fmt.Errorf("task cannot be nil")
	}

	// Read CRUSH_RUN_INSTRUCTIONS.md if it exists
	var templateContent string
	workflowPath := filepath.Join("CRUSH_RUN_INSTRUCTIONS.md")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		// Use default template if file doesn't exist
		templateContent = defaultWorkflowGuide
	} else {
		templateContent = string(content)
	}

	// Parse the template
	tmpl, err := template.New("crush-prompt").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse workflow template: %w", err)
	}

	// Prepare context data
	depsStr := ""
	if len(task.Dependencies) > 0 {
		depsStr = strings.Join(task.Dependencies, ", ")
	}

	context := CrushPromptContext{
		TaskID:       task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Details:      task.Details,
		TestStrategy: task.TestStrategy,
		Priority:     task.Priority,
		Dependencies: depsStr,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return "", fmt.Errorf("failed to execute workflow template: %w", err)
	}

	// Return the prompt (model will be passed as CLI flag separately)
	return buf.String(), nil
}

// GetCrushCommand returns the full command arguments for running Crush
func GetCrushCommand(prompt string, model string) []string {
	args := []string{"run"}
	
	if model != "" {
		args = append(args, "--model", model)
	}
	
	// The prompt will be passed via stdin or as an argument
	// For now, we'll prepare it to be passed via stdin in the actual execution
	
	return args
}

// CrushExecutionSub is a subscription message that indicates a new Crush execution channel is ready
type CrushExecutionSub struct {
	TaskID string
	OutCh  chan tea.Msg
}

// CrushExecutionContextMsg is sent after the process starts to provide cancellation context
type CrushExecutionContextMsg struct {
	TaskID     string
	Cmd        *exec.Cmd
	CancelFunc context.CancelFunc
}

// StartCrushExecution initiates a Crush subprocess and streams output to the modal
// This function returns a tea.Cmd that manages the subprocess lifecycle
func StartCrushExecution(taskID, taskTitle, model, prompt string, modal *TaskRunnerModal) tea.Cmd {
	return func() tea.Msg {
		// Validate that crush exists before starting
		if err := ValidateCrushBinary(); err != nil {
			return TaskFailedMsg{
				TaskID:  taskID,
				Error:   err.Error(),
				Message: "Crush binary not available",
			}
		}

		// Send TaskStartedMsg first to create the tab
		return TaskStartedMsg{
			TaskID:    taskID,
			TaskTitle: taskTitle,
			Model:     model,
		}
	}
}

// ExecuteCrushSubprocess performs the actual subprocess execution with streaming
// It immediately returns a subscription message, then spawns the subprocess in a goroutine
func ExecuteCrushSubprocess(taskID, model, prompt string) tea.Cmd {
	// Create a channel for streaming output messages
	outCh := make(chan tea.Msg, 100)
	
	// Create a cancellable context  
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start the subprocess in a goroutine
	go runCrushProcess(ctx, taskID, model, prompt, outCh, cancel)
	
	// Return subscription message immediately
	return func() tea.Msg {
		return CrushExecutionSub{
			TaskID: taskID,
			OutCh:  outCh,
		}
	}
}

// runCrushProcess executes the Crush subprocess and streams output to the channel
func runCrushProcess(ctx context.Context, taskID, model, prompt string, outCh chan tea.Msg, cancel context.CancelFunc) {
	defer close(outCh)
	defer cancel()
	
	// Create log file for this run
	logFile, logPath, err := createCrushLogFile(taskID)
	var logWriter io.Writer
	if err != nil {
		// Log creation failed, but continue without logging to file
		// Send a warning message but don't fail the entire run
		outCh <- TaskOutputMsg{
			TaskID: taskID,
			Output: fmt.Sprintf("[WARN] Failed to create log file: %v", err),
		}
		logWriter = nil
	} else {
		defer logFile.Close()
		logWriter = logFile
		
		// Send message about log file location
		outCh <- TaskOutputMsg{
			TaskID: taskID,
			Output: fmt.Sprintf("📝 Logging to: %s", logPath),
		}
	}
	
	// Create the command
	// Note: crush run takes the prompt as an argument, not stdin
	// The model selection is stored in crush's config, not passed via CLI flag
	cmd := exec.CommandContext(ctx, "crush", "run", prompt)
	
	// Set up stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		outCh <- TaskFailedMsg{
			TaskID:  taskID,
			Error:   fmt.Sprintf("Failed to create stdout pipe: %v", err),
			Message: "Could not set up subprocess output",
		}
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		outCh <- TaskFailedMsg{
			TaskID:  taskID,
			Error:   fmt.Sprintf("Failed to create stderr pipe: %v", err),
			Message: "Could not set up subprocess error output",
		}
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		outCh <- TaskFailedMsg{
			TaskID:  taskID,
			Error:   fmt.Sprintf("Failed to start crush: %v", err),
			Message: "Could not start Crush subprocess",
		}
		return
	}

	// Send the execution context so the tab can support cancellation
	outCh <- CrushExecutionContextMsg{
		TaskID:     taskID,
		Cmd:        cmd,
		CancelFunc: cancel,
	}

	// Use WaitGroup to coordinate output streaming
	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			
			// Write to log file if available
			if logWriter != nil {
				fmt.Fprintf(logWriter, "[OUT] %s\n", line)
			}
			
			select {
			case <-ctx.Done():
				return
			case outCh <- TaskOutputMsg{
				TaskID: taskID,
				Output: line,
			}:
			}
		}
	}()

	// Stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			
			// Write to log file if available
			if logWriter != nil {
				fmt.Fprintf(logWriter, "[ERR] %s\n", line)
			}
			
			select {
			case <-ctx.Done():
				return
			case outCh <- TaskOutputMsg{
				TaskID: taskID,
				Output: "[ERR] " + line,
			}:
			}
		}
	}()

	// Wait for all output to be consumed
	wg.Wait()

	// Wait for the process to complete
	err = cmd.Wait()
	
	// Write completion status to log file
	if logWriter != nil {
		fmt.Fprintf(logWriter, "\n===================\n")
		fmt.Fprintf(logWriter, "Completed: %s\n", time.Now().Format(time.RFC3339))
		if ctx.Err() != nil {
			fmt.Fprintf(logWriter, "Status: Cancelled\n")
		} else if err != nil {
			fmt.Fprintf(logWriter, "Status: Failed - %v\n", err)
		} else {
			fmt.Fprintf(logWriter, "Status: Success\n")
		}
		fmt.Fprintf(logWriter, "===================\n")
	}
	
	if ctx.Err() != nil {
		// Context was cancelled
		outCh <- TaskCancelledMsg{
			TaskID: taskID,
		}
	} else if err != nil {
		outCh <- TaskFailedMsg{
			TaskID:  taskID,
			Error:   fmt.Sprintf("Crush process failed: %v", err),
			Message: "Subprocess execution failed",
		}
	} else {
		outCh <- TaskCompletedMsg{
			TaskID: taskID,
		}
	}
}

// WaitForCrushMsg returns a tea.Cmd that waits for the next message from the channel
// This is called repeatedly to maintain the subscription
func WaitForCrushMsg(outCh chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-outCh
		if !ok {
			// Channel closed, no more messages
			return nil
		}
		return msg
	}
}

// createCrushLogFile creates a log file for the Crush run
// Returns the file handle and path, or error if creation fails
func createCrushLogFile(taskID string) (*os.File, string, error) {
	// Create logs directory if it doesn't exist
	logsDir := filepath.Join(".taskmaster", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Generate log file path with task ID and timestamp
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("crush-run-%s-%s.log", taskID, timestamp)
	logPath := filepath.Join(logsDir, logFileName)

	// Create the log file
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create log file: %w", err)
	}

	// Write header to log file
	fmt.Fprintf(logFile, "=== Crush Run Log ===\n")
	fmt.Fprintf(logFile, "Task ID: %s\n", taskID)
	fmt.Fprintf(logFile, "Started: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(logFile, "===================\n\n")

	return logFile, logPath, nil
}
