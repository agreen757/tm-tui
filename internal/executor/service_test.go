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

	"github.com/agreen757/tm-tui/internal/config"
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
	_, _, cleanup := setupTestService(t)
	defer cleanup()

	// Skip this test - task-master commands may hang waiting for input
	// in test environments without proper configuration
	t.Skip("Skipping task-master integration test - task-master may hang waiting for input")
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

// TestGetDone tests done completion channel
func TestGetDone(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	doneCh := service.GetDone()
	if doneCh == nil {
		t.Fatal("done channel should not be nil")
	}

	// Test sending to done channel
	testResult := CommandResult{
		Command: "test",
		Success: true,
		Error:   nil,
	}

	go func() {
		service.doneCh <- testResult
	}()

	select {
	case result := <-doneCh:
		if result.Command != testResult.Command {
			t.Errorf("expected command '%s', got '%s'", testResult.Command, result.Command)
		}
		if result.Success != testResult.Success {
			t.Errorf("expected success %v, got %v", testResult.Success, result.Success)
		}
		if result.Error != testResult.Error {
			t.Errorf("expected error nil, got %v", result.Error)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for done channel")
	}
}

// TestExecutorDoneMsg_Success tests emitting success message
func TestExecutorDoneMsg_Success(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Simulate successful command completion
	go func() {
		service.doneCh <- CommandResult{
			Command: "next",
			Success: true,
			Error:   nil,
		}
	}()

	timeout := time.After(2 * time.Second)
	select {
	case result := <-service.GetDone():
		if result.Command != "next" {
			t.Errorf("expected command 'next', got '%s'", result.Command)
		}
		if !result.Success {
			t.Error("expected success true")
		}
		if result.Error != nil {
			t.Errorf("expected no error, got %v", result.Error)
		}
	case <-timeout:
		t.Fatal("timeout waiting for done message")
	}
}

// TestExecutorDoneMsg_Failure tests emitting failure message
func TestExecutorDoneMsg_Failure(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	testErr := fmt.Errorf("command failed")

	// Simulate failed command completion
	go func() {
		service.doneCh <- CommandResult{
			Command: "next",
			Success: false,
			Error:   testErr,
		}
	}()

	timeout := time.After(2 * time.Second)
	select {
	case result := <-service.GetDone():
		if result.Command != "next" {
			t.Errorf("expected command 'next', got '%s'", result.Command)
		}
		if result.Success {
			t.Error("expected success false")
		}
		if result.Error != testErr {
			t.Errorf("expected error '%v', got %v", testErr, result.Error)
		}
	case <-timeout:
		t.Fatal("timeout waiting for done message")
	}
}

// TestCommandResult tests CommandResult struct
func TestCommandResult(t *testing.T) {
	result := CommandResult{
		Command: "test-cmd",
		Success: true,
		Error:   nil,
	}

	if result.Command != "test-cmd" {
		t.Errorf("expected command 'test-cmd', got '%s'", result.Command)
	}

	if !result.Success {
		t.Error("expected success true")
	}

	if result.Error != nil {
		t.Errorf("expected nil error, got %v", result.Error)
	}

	// Test with error
	testErr := fmt.Errorf("test error")
	resultWithErr := CommandResult{
		Command: "test-cmd",
		Success: false,
		Error:   testErr,
	}

	if resultWithErr.Error != testErr {
		t.Errorf("expected error '%v', got %v", testErr, resultWithErr.Error)
	}
}

// TestExecuteNextWithEmptyTaskMasterPath tests that 'next' command fails with empty TaskMasterPath
func TestExecuteNextWithEmptyTaskMasterPath(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	// Create a temporary config with actual path for log file purposes
	tempCfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	service, err := NewService(tempCfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	// Now set the TaskMasterPath to empty to test the check
	service.config.TaskMasterPath = ""

	// Try to execute 'next' command
	err = service.Execute("next")
	if err == nil {
		t.Error("expected error when TaskMasterPath is empty for 'next' command")
	}

	if !strings.Contains(err.Error(), "not running in a Task Master workspace") {
		t.Errorf("expected 'not running in a Task Master workspace' error, got: %v", err)
	}
}

// TestExecuteOtherCommandWithEmptyTaskMasterPath tests that other commands work even with empty TaskMasterPath
func TestExecuteOtherCommandWithEmptyTaskMasterPath(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	// Set the TaskMasterPath to empty
	service.config.TaskMasterPath = ""

	// Try to execute a different command (not 'next')
	// This should not fail due to empty TaskMasterPath
	err = service.Execute("list")
	// The error here would be due to task-master not being found in PATH, not TaskMasterPath being empty
	if err != nil && strings.Contains(err.Error(), "not running in a Task Master workspace") {
		t.Error("non-'next' commands should not fail due to empty TaskMasterPath")
	}
}

// TestExecuteNextWithMissingBinary tests that 'next' command fails when binary doesn't exist
func TestExecuteNextWithMissingBinary(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	// Try to execute 'next' command when binary doesn't exist
	err = service.Execute("next")
	if err == nil {
		t.Error("expected error when task-master binary doesn't exist")
	}

	if !strings.Contains(err.Error(), "task-master binary not found") {
		t.Errorf("expected 'task-master binary not found' error, got: %v", err)
	}
}

// TestExecuteNextWithExistingBinary tests that 'next' command passes existence check
func TestExecuteNextWithExistingBinary(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	// Create a fake task-master binary
	binPath := filepath.Join(tmpDir, "task-master")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\necho 'test'\n"), 0755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	// Try to execute 'next' command when binary exists
	// It should not fail due to missing binary (it may fail due to execution, but that's expected)
	err = service.Execute("next")
	// We don't care about the execution error, we care that it didn't fail due to missing binary
	if err != nil && strings.Contains(err.Error(), "task-master binary not found") {
		t.Error("should not fail with 'binary not found' when binary exists")
	}
}

// TestExecuteNextWithNonExecutableBinary tests that 'next' command fails when binary isn't executable
func TestExecuteNextWithNonExecutableBinary(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("failed to create .taskmaster dir: %v", err)
	}

	// Create a binary without execute permissions
	binPath := filepath.Join(tmpDir, "task-master")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\necho 'test'\n"), 0644); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
	}

	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer service.Close()

	// Try to execute 'next' command with non-executable binary
	err = service.Execute("next")
	if err == nil {
		t.Error("expected error when task-master binary is not executable")
	}

	if !strings.Contains(err.Error(), "task-master binary not executable") {
		t.Errorf("expected 'task-master binary not executable' error, got: %v", err)
	}
}

// TestAddToHistoryBounded tests that history is bounded to max size
func TestAddToHistoryBounded(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Set a small max history size
	service.SetMaxHistorySize(5)

	// Add 10 commands to history
	for i := 0; i < 10; i++ {
		service.mu.Lock()
		service.addToHistory(Command{
			Cmd:      fmt.Sprintf("cmd%d", i),
			Args:     []string{},
			When:     time.Now(),
			ExitCode: 0,
			Err:      nil,
		})
		service.mu.Unlock()
	}

	// Verify history is bounded to max size
	history := service.GetHistory()
	if len(history) != 5 {
		t.Errorf("expected history size 5, got %d", len(history))
	}

	// Verify only the most recent commands are kept (commands 5-9)
	if history[0].Cmd != "cmd5" {
		t.Errorf("expected first command to be 'cmd5', got '%s'", history[0].Cmd)
	}
	if history[4].Cmd != "cmd9" {
		t.Errorf("expected last command to be 'cmd9', got '%s'", history[4].Cmd)
	}
}

// TestSetMaxHistorySize tests that max history size can be configured
func TestSetMaxHistorySize(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Change max history size
	service.SetMaxHistorySize(3)

	// Add 5 commands
	for i := 0; i < 5; i++ {
		service.mu.Lock()
		service.addToHistory(Command{
			Cmd:      fmt.Sprintf("cmd%d", i),
			Args:     []string{},
			When:     time.Now(),
			ExitCode: 0,
			Err:      nil,
		})
		service.mu.Unlock()
	}

	// Verify history is bounded to new size
	history := service.GetHistory()
	if len(history) != 3 {
		t.Errorf("expected history size 3, got %d", len(history))
	}

	// Verify the most recent 3 commands are kept
	if history[0].Cmd != "cmd2" {
		t.Errorf("expected first command to be 'cmd2', got '%s'", history[0].Cmd)
	}
	if history[2].Cmd != "cmd4" {
		t.Errorf("expected last command to be 'cmd4', got '%s'", history[2].Cmd)
	}
}

// TestOldestCommandsRemoved tests that oldest commands are removed when limit reached
func TestOldestCommandsRemoved(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	service.SetMaxHistorySize(3)

	// Add commands and verify oldest are removed
	cmds := []string{"cmd1", "cmd2", "cmd3", "cmd4"}

	for i, cmd := range cmds {
		service.mu.Lock()
		service.addToHistory(Command{
			Cmd:      cmd,
			Args:     []string{},
			When:     time.Now().Add(time.Duration(i) * time.Second),
			ExitCode: 0,
			Err:      nil,
		})
		service.mu.Unlock()

		history := service.GetHistory()

		// After first 3 commands, size should be 3
		if i < 3 {
			expectedSize := i + 1
			if len(history) != expectedSize {
				t.Errorf("at step %d: expected history size %d, got %d", i, expectedSize, len(history))
			}
		} else {
			// After 4th command, oldest should be removed
			if len(history) != 3 {
				t.Errorf("at step %d: expected history size 3, got %d", i, len(history))
			}
			// First command should be cmd2 (cmd1 was removed)
			if history[0].Cmd != "cmd2" {
				t.Errorf("at step %d: expected oldest command to be 'cmd2', got '%s'", i, history[0].Cmd)
			}
		}
	}
}

// TestRecentCommandsAccessible tests that most recent commands are accessible
func TestRecentCommandsAccessible(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	service.SetMaxHistorySize(10)

	// Add 20 commands
	for i := 0; i < 20; i++ {
		service.mu.Lock()
		service.addToHistory(Command{
			Cmd:      fmt.Sprintf("cmd%d", i),
			Args:     []string{},
			When:     time.Now(),
			ExitCode: 0,
			Err:      nil,
		})
		service.mu.Unlock()
	}

	history := service.GetHistory()

	// Verify we have the last 10 commands
	if len(history) != 10 {
		t.Errorf("expected 10 commands in history, got %d", len(history))
	}

	// Verify most recent command is cmd19
	if history[9].Cmd != "cmd19" {
		t.Errorf("expected most recent command to be 'cmd19', got '%s'", history[9].Cmd)
	}

	// Verify oldest kept command is cmd10
	if history[0].Cmd != "cmd10" {
		t.Errorf("expected oldest kept command to be 'cmd10', got '%s'", history[0].Cmd)
	}
}

// TestMaxHistorySizeInitialization tests default max history size
func TestMaxHistorySizeInitialization(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Default max history size should be 1000
	if service.maxHistorySize != 1000 {
		t.Errorf("expected default max history size 1000, got %d", service.maxHistorySize)
	}
}

// TestSetMaxHistorySizeInvalid tests that invalid max history size is rejected
func TestSetMaxHistorySizeInvalid(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Try to set invalid (zero or negative) max history size
	service.SetMaxHistorySize(0)
	if service.maxHistorySize != 1000 {
		t.Errorf("expected max history size to remain 1000 after setting 0, got %d", service.maxHistorySize)
	}

	service.SetMaxHistorySize(-10)
	if service.maxHistorySize != 1000 {
		t.Errorf("expected max history size to remain 1000 after setting -10, got %d", service.maxHistorySize)
	}
}

// TestMemoryBehaviorWithLargeHistory benchmarks memory usage with large history
func TestMemoryBehaviorWithLargeHistory(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Test with large history size (1000 instead of 10000 for faster test)
	service.SetMaxHistorySize(1000)

	// Add many commands to test memory behavior
	for i := 0; i < 1000; i++ {
		service.mu.Lock()
		service.addToHistory(Command{
			Cmd:      fmt.Sprintf("cmd%d", i),
			Args:     []string{"arg1", "arg2", "arg3"},
			When:     time.Now(),
			ExitCode: 0,
			Err:      nil,
		})
		service.mu.Unlock()
	}

	history := service.GetHistory()

	// Verify history size is bounded correctly
	if len(history) != 1000 {
		t.Errorf("expected history size 1000, got %d", len(history))
	}

	// Add one more command to trigger trimming
	service.mu.Lock()
	service.addToHistory(Command{
		Cmd:      "cmd1000",
		Args:     []string{},
		When:     time.Now(),
		ExitCode: 0,
		Err:      nil,
	})
	service.mu.Unlock()

	// History should still be bounded to 1000
	history = service.GetHistory()
	if len(history) != 1000 {
		t.Errorf("after exceeding max size, expected history size 1000, got %d", len(history))
	}
}
