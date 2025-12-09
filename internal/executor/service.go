package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/adriangreen/tm-tui/internal/config"
)

// Command represents a single command execution record
type Command struct {
	Cmd      string
	Args     []string
	When     time.Time
	ExitCode int
	Err      error
}

// Service manages command execution state with async execution and logging
type Service struct {
	config   *config.Config
	mu       sync.Mutex
	running  bool
	cancelFn context.CancelFunc
	outCh    chan string
	history  []Command
	logFile  *os.File
}

// NewService creates a new executor service with log file initialization
func NewService(cfg *config.Config) (*Service, error) {
	// Create logs directory if it doesn't exist
	logsDir := filepath.Join(cfg.TaskMasterPath, ".taskmaster", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open log file in append mode
	logPath := filepath.Join(logsDir, "tui-session.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Write session start marker
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(logFile, "\n=== TUI Session Started: %s ===\n", timestamp)

	return &Service{
		config:  cfg,
		running: false,
		outCh:   make(chan string, 100), // Buffered channel for output
		history: []Command{},
		logFile: logFile,
	}, nil
}

// Execute runs a task-master command asynchronously
func (s *Service) Execute(cmd string, args ...string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("command already running")
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	s.running = true
	s.cancelFn = cancel
	s.mu.Unlock()

	// Build full command
	fullArgs := append([]string{cmd}, args...)
	execCmd := exec.CommandContext(ctx, "task-master", fullArgs...)

	// Set working directory to task-master path
	if s.config.TaskMasterPath != "" {
		execCmd.Dir = s.config.TaskMasterPath
	}

	// Get stdout and stderr pipes
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.cancelFn = nil
		s.mu.Unlock()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.cancelFn = nil
		s.mu.Unlock()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := execCmd.Start(); err != nil {
		s.mu.Lock()
		s.running = false
		s.cancelFn = nil
		s.mu.Unlock()
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Log command start
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("\n[%s] Executing: task-master %s %v\n", timestamp, cmd, args)
	s.logFile.WriteString(logEntry)

	// Stream output in separate goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.streamOutput(stdout, "")
	}()

	go func() {
		defer wg.Done()
		s.streamOutput(stderr, "")
	}()

	// Wait for command completion in a goroutine
	go func() {
		wg.Wait() // Wait for output streaming to complete

		err := execCmd.Wait()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		// Log completion
		timestamp := time.Now().Format(time.RFC3339)
		completionMsg := fmt.Sprintf("[%s] Command completed with exit code %d\n", timestamp, exitCode)
		s.logFile.WriteString(completionMsg)

		// Update state
		s.mu.Lock()
		s.running = false
		s.cancelFn = nil
		s.history = append(s.history, Command{
			Cmd:      cmd,
			Args:     args,
			When:     time.Now(),
			ExitCode: exitCode,
			Err:      err,
		})
		s.mu.Unlock()

		// Send completion message to output channel
		if exitCode == 0 {
			s.outCh <- fmt.Sprintf("\n✓ Command completed successfully")
		} else {
			s.outCh <- fmt.Sprintf("\n✗ Command failed with exit code %d", exitCode)
		}
	}()

	return nil
}

// streamOutput reads from a pipe and sends lines to output channel and log file
func (s *Service) streamOutput(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if prefix != "" {
			line = prefix + line
		}

		// Send to output channel (non-blocking)
		select {
		case s.outCh <- line:
		default:
			// Channel full, skip this line
		}

		// Write to log file with timestamp
		timestamp := time.Now().Format("15:04:05")
		s.logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, line))
	}
}

// Cancel stops the currently running command
func (s *Service) Cancel() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cancelFn == nil {
		return fmt.Errorf("no command is running")
	}

	// Call cancel function to trigger context cancellation
	s.cancelFn()

	// Log cancellation
	timestamp := time.Now().Format(time.RFC3339)
	s.logFile.WriteString(fmt.Sprintf("[%s] Command cancelled by user\n", timestamp))

	return nil
}

// GetOutput returns a channel that receives command output
func (s *Service) GetOutput() <-chan string {
	return s.outCh
}

// IsRunning returns whether a command is currently executing
func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetHistory returns a copy of the command history
func (s *Service) GetHistory() []Command {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy to prevent external modification
	history := make([]Command, len(s.history))
	copy(history, s.history)
	return history
}

// Close closes the log file and cleans up resources
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logFile != nil {
		timestamp := time.Now().Format(time.RFC3339)
		fmt.Fprintf(s.logFile, "\n=== TUI Session Ended: %s ===\n", timestamp)
		return s.logFile.Close()
	}
	return nil
}
