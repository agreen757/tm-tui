package dialog

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// GitBinaryError represents an error when the git binary is not found
type GitBinaryError struct {
	Message string
}

func (e *GitBinaryError) Error() string {
	return e.Message
}

// ValidateGitBinary checks if the git CLI binary is available
func ValidateGitBinary() error {
	path, err := exec.LookPath("git")
	if err != nil {
		return &GitBinaryError{
			Message: "git binary not found. Please install git: https://git-scm.com/download",
		}
	}
	// Verify it's executable
	if _, err := os.Stat(path); err != nil {
		return &GitBinaryError{
			Message: fmt.Sprintf("git binary found at %s but not accessible: %v", path, err),
		}
	}
	return nil
}

// RunGitCommand executes a git command with streaming output.
// commandID should be a unique identifier for logging correlation.
// args are the git command arguments (e.g., "status", "branch", "-a").
// Returns a channel of tea.Msg (TaskOutputMsg, TaskCompletedMsg, TaskFailedMsg) and an error if startup fails.
// The channel is closed when the command completes.
func RunGitCommand(commandID string, args ...string) (chan tea.Msg, error) {
	if commandID == "" {
		return nil, fmt.Errorf("command ID cannot be empty")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("git command arguments cannot be empty")
	}

	// Validate that git binary exists
	if err := ValidateGitBinary(); err != nil {
		return nil, err
	}

	// Create command with background context (no inherent timeout)
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", args...)

	// Set up stdout/stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start git command: %w", err)
	}

	// Create output channel with large buffer to handle bursts
	outputChan := make(chan tea.Msg, 1000)

	// Create log file
	logPath := filepath.Join(".taskmaster", "logs", fmt.Sprintf("git-command-%s-%d.log", commandID, time.Now().Unix()))
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		// Continue without logging if directory creation fails
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		// Continue without logging if file creation fails
		logFile = nil
	}

	// Write header to log file
	if logFile != nil {
		fmt.Fprintf(logFile, "=== Git Command Log ===\n")
		fmt.Fprintf(logFile, "Command ID: %s\n", commandID)
		fmt.Fprintf(logFile, "Started: %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(logFile, "Command: git %s\n", strings.Join(args, " "))
		fmt.Fprintf(logFile, "===================\n\n")
	}

	// Stream output in goroutine
	go func() {
		defer close(outputChan)
		if logFile != nil {
			defer logFile.Close()
		}

		// Merge stdout and stderr
		merged := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(merged)

		// Increase scanner buffer size to handle long lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 1MB max line size

		for scanner.Scan() {
			line := scanner.Text()

			// Write to log file if available
			if logFile != nil {
				fmt.Fprintf(logFile, "%s\n", line)
			}

			// Send to output channel as TaskOutputMsg
			outputChan <- TaskOutputMsg{
				TaskID: commandID,
				Output: line,
			}
		}

		// Wait for command to finish
		err := cmd.Wait()

		// Write completion status to log file
		if logFile != nil {
			fmt.Fprintf(logFile, "\n===================\n")
			fmt.Fprintf(logFile, "Completed: %s\n", time.Now().Format(time.RFC3339))
			if err != nil {
				fmt.Fprintf(logFile, "Status: Failed - %v\n", err)
			} else {
				fmt.Fprintf(logFile, "Status: Success\n")
			}
			fmt.Fprintf(logFile, "===================\n")
		}

		// Send completion or failure message
		if err != nil {
			outputChan <- TaskFailedMsg{
				TaskID:  commandID,
				Error:   fmt.Sprintf("Command failed: %v", err),
				Message: "Git command execution failed",
			}
		} else {
			outputChan <- TaskCompletedMsg{
				TaskID: commandID,
			}
		}
	}()

	return outputChan, nil
}

// ExecuteGitCommand performs git command execution with streaming output
// It returns a subscription message with the output channel
// tagName is used for organizing log files in tag-specific directories
func ExecuteGitCommand(commandID string, args []string, tagName string) tea.Cmd {
	// Create a larger buffered channel to handle bursts of output
	outCh := make(chan tea.Msg, 1000)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start the subprocess in a goroutine
	go runGitProcess(ctx, commandID, args, tagName, outCh, cancel)

	// Return subscription message immediately
	return func() tea.Msg {
		return CrushExecutionSub{
			TaskID: commandID,
			OutCh:  outCh,
		}
	}
}

// runGitProcess executes the git command and streams output to the channel
// tagName determines the log directory structure (.taskmaster/<tagName>/<commandID>.log)
func runGitProcess(ctx context.Context, commandID string, args []string, tagName string, outCh chan tea.Msg, cancel context.CancelFunc) {
	defer close(outCh)
	defer cancel()

	// Create structured logger for this git operation
	gitLogger, logPath, err := NewGitLogger(commandID, args, tagName)
	if err != nil {
		// Log creation failed, but continue without logging to file
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: fmt.Sprintf("[WARN] Failed to create logger: %v", err),
		}
	} else {
		defer gitLogger.Close()

		// Send message about log file location
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: fmt.Sprintf("📝 Logging to: %s", logPath),
		}
	}

	// Store stderr lines for error parsing if command fails
	stderrLines := make([]string, 0)
	stderrMutex := &sync.Mutex{}

	// Create the command
	cmd := exec.CommandContext(ctx, "git", args...)

	// Set up stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		// Log startup failure
		LogStartupError(commandID, args, tagName, fmt.Sprintf("Failed to create stdout pipe: %v", err))

		outCh <- TaskFailedMsg{
			TaskID:  commandID,
			Error:   fmt.Sprintf("Failed to create stdout pipe: %v", err),
			Message: "Could not set up git command output",
		}
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		// Log startup failure
		LogStartupError(commandID, args, tagName, fmt.Sprintf("Failed to create stderr pipe: %v", err))

		outCh <- TaskFailedMsg{
			TaskID:  commandID,
			Error:   fmt.Sprintf("Failed to create stderr pipe: %v", err),
			Message: "Could not set up git command error output",
		}
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		// Log startup failure
		LogStartupError(commandID, args, tagName, fmt.Sprintf("Failed to start git: %v", err))

		outCh <- TaskFailedMsg{
			TaskID:  commandID,
			Error:   fmt.Sprintf("Failed to start git: %v", err),
			Message: "Could not start git command",
		}
		return
	}

	// Send the execution context so the tab can support cancellation
	outCh <- CrushExecutionContextMsg{
		TaskID:     commandID,
		Cmd:        cmd,
		CancelFunc: cancel,
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		// Increase scanner buffer size to handle long lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 1MB max line size

		for scanner.Scan() {
			line := scanner.Text()

			// Log output if logger is available
			if gitLogger != nil {
				gitLogger.LogOutput("stdout", line)
			}

			// Non-blocking send with timeout to prevent goroutine from hanging
			select {
			case <-ctx.Done():
				return
			case outCh <- TaskOutputMsg{
				TaskID: commandID,
				Output: line,
			}:
				// Successfully sent
			case <-time.After(100 * time.Millisecond):
				// Channel is full and TUI is slow - drop message but continue
				// Log file will still have the complete output
				if gitLogger != nil {
					gitLogger.LogWarning("Output channel full, message dropped from UI (logged to file)", nil)
				}
			}
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		// Increase scanner buffer size to handle long lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 1MB max line size

		for scanner.Scan() {
			line := scanner.Text()

			// Store the stderr line for error parsing later
			stderrMutex.Lock()
			stderrLines = append(stderrLines, line)
			stderrMutex.Unlock()

			// Log output if logger is available
			if gitLogger != nil {
				gitLogger.LogOutput("stderr", line)
			}

			// Non-blocking send with timeout to prevent goroutine from hanging
			select {
			case <-ctx.Done():
				return
			case outCh <- TaskOutputMsg{
				TaskID: commandID,
				Output: "[ERR] " + line,
			}:
				// Successfully sent
			case <-time.After(100 * time.Millisecond):
				// Channel is full and TUI is slow - drop message but continue
				// Error messages are critical, so we still log them
				if gitLogger != nil {
					gitLogger.LogWarning("Error output channel full, message dropped from UI (logged to file)", nil)
				}
			}
		}
	}()

	// Wait for the process to complete
	err = cmd.Wait()

	// Log completion status
	if gitLogger != nil {
		if ctx.Err() != nil {
			gitLogger.LogCancellation()
		} else {
			// Extract exit code for logging
			exitCode := 0
			if err != nil {
				exitCode = 1
			}
			gitLogger.LogCompletion(exitCode, err)
		}
	}

	if ctx.Err() != nil {
		// Context was cancelled
		outCh <- TaskCancelledMsg{
			TaskID: commandID,
		}
	} else if err != nil {
		// Collect all stderr lines with proper synchronization
		stderrMutex.Lock()
		stderrOutput := strings.Join(stderrLines, "\n")
		stderrMutex.Unlock()

		// Parse the error using our custom error parser
		gitErr := ParseGitError(stderrOutput, strings.Join(args, " "))
		
		// Send user-friendly error message
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: "",
		}
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: "════════════════════════════════════════════════════════",
		}
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: fmt.Sprintf("❌ Error: %s", gitErr.ErrorType()),
		}
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: "",
		}
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: "💡 Suggestion:",
		}
		for _, line := range strings.Split(gitErr.Suggestion(), "\n") {
			outCh <- TaskOutputMsg{
				TaskID: commandID,
				Output: line,
			}
		}
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: "",
		}
		outCh <- TaskOutputMsg{
			TaskID: commandID,
			Output: "════════════════════════════════════════════════════════",
		}
		
		// Send failure message
		outCh <- TaskFailedMsg{
			TaskID:  commandID,
			Error:   fmt.Sprintf("Git process failed: %v", err),
			Message: "Subprocess execution failed",
		}
	} else {
		outCh <- TaskCompletedMsg{
			TaskID: commandID,
		}
	}
}

