package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adriangreen/tm-tui/internal/config"
)

// setupTestService creates a test service with a temporary directory
func setupTestService(t *testing.T) (*Service, string, func()) {
	t.Helper()

	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "executor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create .taskmaster directory
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	service, err := NewService(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create service: %v", err)
	}

	cleanup := func() {
		service.Close()
		os.RemoveAll(tmpDir)
	}

	return service, tmpDir, cleanup
}

// TestNewService verifies service initialization
func TestNewService(t *testing.T) {
	service, tmpDir, cleanup := setupTestService(t)
	defer cleanup()

	// Verify log file was created
	logPath := filepath.Join(tmpDir, ".taskmaster", "logs", "tui-session.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("log file was not created at %s", logPath)
	}

	// Verify log file contains session start marker
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "TUI Session Started") {
		t.Errorf("log file does not contain session start marker")
	}

	// Verify service is not running initially
	if service.IsRunning() {
		t.Errorf("service should not be running initially")
	}

	// Verify history is empty
	if len(service.GetHistory()) != 0 {
		t.Errorf("history should be empty initially")
	}
}

// TestExecuteCommand tests basic command execution
func TestExecuteCommand(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Use echo command to test output streaming
	err := service.Execute("echo", "test output")
	if err != nil {
		// Note: This will fail if task-master is not installed
		// For testing purposes, we verify the error handling
		if !strings.Contains(err.Error(), "executable file not found") {
			t.Errorf("unexpected error: %v", err)
		}
		t.Skip("task-master not found in PATH, skipping command execution test")
	}

	// Verify service is running
	if !service.IsRunning() {
		t.Errorf("service should be running after Execute")
	}

	// Wait for command to complete (with timeout)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("command did not complete within timeout")
		case <-ticker.C:
			if !service.IsRunning() {
				// Command completed
				goto commandDone
			}
		}
	}

commandDone:
	// Verify history was updated
	history := service.GetHistory()
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
}

// TestExecuteWhileRunning tests concurrent execution prevention
func TestExecuteWhileRunning(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Create a helper script that sleeps
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "sleep.sh")
	script := "#!/bin/sh\nsleep 1\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	// Start a long-running command using sh
	cmd := exec.Command("sh", scriptPath)
	
	// Manually set service as running to simulate a running command
	service.mu.Lock()
	service.running = true
	_, cancel := context.WithCancel(context.Background())
	service.cancelFn = cancel
	service.mu.Unlock()

	// Try to execute another command
	err := service.Execute("echo", "should fail")
	if err == nil {
		t.Error("expected error when executing while another command is running")
	}

	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("expected 'already running' error, got: %v", err)
	}

	// Cleanup
	cancel()
	_ = cmd.Wait()
	service.mu.Lock()
	service.running = false
	service.cancelFn = nil
	service.mu.Unlock()
}

// TestCancel tests command cancellation
func TestCancel(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Test canceling when nothing is running
	err := service.Cancel()
	if err == nil {
		t.Error("expected error when canceling with no running command")
	}

	// Create a helper script that sleeps
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "long-sleep.sh")
	script := "#!/bin/sh\nsleep 10\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	// Manually start a long-running command
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "sh", scriptPath)
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test command: %v", err)
	}

	// Set service as running
	service.mu.Lock()
	service.running = true
	service.cancelFn = cancel
	service.mu.Unlock()

	// Verify it's running
	if !service.IsRunning() {
		t.Error("service should be running")
	}

	// Cancel the command
	if err := service.Cancel(); err != nil {
		t.Errorf("failed to cancel command: %v", err)
	}

	// Wait for cancellation to take effect
	_ = cmd.Wait()

	// Cleanup
	service.mu.Lock()
	service.running = false
	service.cancelFn = nil
	service.mu.Unlock()
}

// TestGetOutput tests output channel
func TestGetOutput(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	outCh := service.GetOutput()
	if outCh == nil {
		t.Fatal("output channel should not be nil")
	}

	// Test sending to output channel
	testLine := "test output line"
	go func() {
		service.outCh <- testLine
	}()

	select {
	case line := <-outCh:
		if line != testLine {
			t.Errorf("expected '%s', got '%s'", testLine, line)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for output")
	}
}

// TestGetHistory tests command history
func TestGetHistory(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Initially empty
	history := service.GetHistory()
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d entries", len(history))
	}

	// Add some history entries
	service.mu.Lock()
	service.history = append(service.history, Command{
		Cmd:      "list",
		Args:     []string{},
		When:     time.Now(),
		ExitCode: 0,
		Err:      nil,
	})
	service.history = append(service.history, Command{
		Cmd:      "show",
		Args:     []string{"1"},
		When:     time.Now(),
		ExitCode: 0,
		Err:      nil,
	})
	service.mu.Unlock()

	// Get history
	history = service.GetHistory()
	if len(history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(history))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect internal state)
	history[0].ExitCode = 999
	
	service.mu.Lock()
	if service.history[0].ExitCode == 999 {
		t.Error("GetHistory should return a copy, not the original slice")
	}
	service.mu.Unlock()
}

// TestLogFileAppend tests that log file appends rather than truncates
func TestLogFileAppend(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "executor-append-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .taskmaster directory
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	// Create first service
	service1, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create first service: %v", err)
	}
	service1.Close()

	// Read log file content after first session
	logPath := filepath.Join(tmpDir, ".taskmaster", "logs", "tui-session.log")
	content1, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Create second service (should append)
	service2, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create second service: %v", err)
	}
	service2.Close()

	// Read log file content after second session
	content2, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Verify second content contains first content (append mode)
	if !strings.Contains(string(content2), string(content1)) {
		t.Error("log file should append, not truncate")
	}

	// Verify there are two session markers
	sessionCount := strings.Count(string(content2), "TUI Session Started")
	if sessionCount != 2 {
		t.Errorf("expected 2 session start markers, got %d", sessionCount)
	}
}

// TestStreamOutput tests output streaming to channel and log
func TestStreamOutput(t *testing.T) {
	service, tmpDir, cleanup := setupTestService(t)
	defer cleanup()

	// Create a pipe to simulate command output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()

	// Start streaming in goroutine
	done := make(chan bool)
	go func() {
		service.streamOutput(r, "")
		done <- true
	}()

	// Write test lines
	testLines := []string{"line 1", "line 2", "line 3"}
	for _, line := range testLines {
		fmt.Fprintln(w, line)
	}
	w.Close()

	// Wait for streaming to complete
	<-done

	// Verify lines were sent to output channel
	receivedLines := []string{}
	timeout := time.After(1 * time.Second)
readLoop:
	for i := 0; i < len(testLines); i++ {
		select {
		case line := <-service.GetOutput():
			receivedLines = append(receivedLines, line)
		case <-timeout:
			break readLoop
		}
	}

	if len(receivedLines) != len(testLines) {
		t.Errorf("expected %d lines, got %d", len(testLines), len(receivedLines))
	}

	// Verify lines were written to log file
	logPath := filepath.Join(tmpDir, ".taskmaster", "logs", "tui-session.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	for _, line := range testLines {
		if !strings.Contains(string(content), line) {
			t.Errorf("log file should contain line: %s", line)
		}
	}
}

// TestClose tests cleanup on service close
func TestClose(t *testing.T) {
	service, tmpDir, _ := setupTestService(t)
	// Don't defer cleanup here since we're testing Close explicitly

	logPath := filepath.Join(tmpDir, ".taskmaster", "logs", "tui-session.log")

	// Close the service
	if err := service.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Read log file and verify session end marker
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file after close: %v", err)
	}

	if !strings.Contains(string(content), "TUI Session Ended") {
		t.Error("log file should contain session end marker")
	}

	// Cleanup
	os.RemoveAll(tmpDir)
}
