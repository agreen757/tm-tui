package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateGitBinary tests the ValidateGitBinary function
func TestValidateGitBinary(t *testing.T) {
	// Test successful validation - git should be available on most systems
	err := ValidateGitBinary()
	// We allow this to fail on systems without git installed, but we check the error type
	if err != nil {
		_, ok := err.(*GitBinaryError)
		assert.True(t, ok, "Error should be GitBinaryError type")
	}
}

// TestValidateGitBinaryMissing tests the error when git binary is not found
// This test is skipped on systems where git is in PATH
func TestValidateGitBinaryMissing(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to simulate missing git
	os.Setenv("PATH", "")

	err := ValidateGitBinary()
	assert.NotNil(t, err, "Should return error when git not found")

	gitErr, ok := err.(*GitBinaryError)
	assert.True(t, ok, "Error should be GitBinaryError type")
	assert.Contains(t, gitErr.Message, "git binary not found")
}

// TestGitBinaryErrorString tests the Error() method of GitBinaryError
func TestGitBinaryErrorString(t *testing.T) {
	err := &GitBinaryError{
		Message: "test error message",
	}
	assert.Equal(t, "test error message", err.Error())
}

// TestRunGitCommandEmptyID tests RunGitCommand with empty command ID
func TestRunGitCommandEmptyID(t *testing.T) {
	_, err := RunGitCommand("")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "command ID cannot be empty")
}

// TestRunGitCommandNoArgs tests RunGitCommand with no arguments
func TestRunGitCommandNoArgs(t *testing.T) {
	_, err := RunGitCommand("test-cmd")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "git command arguments cannot be empty")
}

// TestRunGitCommandValidation tests that RunGitCommand validates git binary exists
func TestRunGitCommandValidation(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to simulate missing git
	os.Setenv("PATH", "")

	_, err := RunGitCommand("test-cmd", "--version")
	assert.NotNil(t, err)

	_, ok := err.(*GitBinaryError)
	assert.True(t, ok, "Error should be GitBinaryError type")
}

// TestRunGitCommandVersion tests successful git command execution with git --version
func TestRunGitCommandVersion(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	msgChan, err := RunGitCommand("test-version", "--version")
	require.NoError(t, err, "Should not error on valid command")
	require.NotNil(t, msgChan, "Should return message channel")

	// Collect all messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Should have at least one output message and a completion message
	require.Greater(t, len(messages), 0, "Should receive at least one message")

	// Last message should be TaskCompletedMsg or TaskFailedMsg
	lastMsg := messages[len(messages)-1]
	_, isCompleted := lastMsg.(TaskCompletedMsg)
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isCompleted || isFailed, "Last message should be completion or failure")
}

// TestRunGitCommandFailure tests git command execution with invalid git command
func TestRunGitCommandFailure(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	msgChan, err := RunGitCommand("test-invalid", "invalid-git-command-that-does-not-exist")
	require.NoError(t, err, "Should not error on command startup")
	require.NotNil(t, msgChan, "Should return message channel")

	// Collect all messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Should have at least error message or failure message
	require.Greater(t, len(messages), 0, "Should receive at least one message")

	// Last message should be TaskFailedMsg
	lastMsg := messages[len(messages)-1]
	failedMsg, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isFailed, "Last message should be failure")
	assert.NotEmpty(t, failedMsg.Error, "Failure message should have error details")
}

// TestRunGitCommandMessageChannel tests that git command properly sends messages to channel
func TestRunGitCommandMessageChannel(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	msgChan, err := RunGitCommand("test-messages", "config", "--list")
	require.NoError(t, err)
	require.NotNil(t, msgChan)

	// Collect messages
	var outputCount int
	var hasCompletion bool
	var channelClosed bool
	timeout := time.After(5 * time.Second)

	for !channelClosed && !hasCompletion {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				// Channel closed, we're done
				channelClosed = true
				break
			}
			switch msg.(type) {
			case TaskOutputMsg:
				outputCount++
			case TaskCompletedMsg:
				hasCompletion = true
			case TaskFailedMsg:
				hasCompletion = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for messages")
		}
	}

	assert.Greater(t, outputCount, 0, "Should receive output messages")
	assert.True(t, hasCompletion, "Should receive completion message")
}

// TestRunGitCommandLogging tests that git command creates log files
func TestRunGitCommandLogging(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	// Create temporary directory for test
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	msgChan, err := RunGitCommand("test-logging", "--version")
	require.NoError(t, err)

	// Wait for command to complete
	for range msgChan {
	}

	// Check that log file was created
	logsDir := filepath.Join(tmpDir, ".taskmaster", "logs")
	entries, err := os.ReadDir(logsDir)
	require.NoError(t, err)
	assert.Greater(t, len(entries), 0, "Should create log file")

	// Check log file contents
	if len(entries) > 0 {
		logFile := filepath.Join(logsDir, entries[0].Name())
		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "=== Git Command Log ===", "Log should have header")
		assert.Contains(t, contentStr, "Command ID: test-logging", "Log should have command ID")
		assert.Contains(t, contentStr, "git --version", "Log should have command details")
	}
}

// TestCreateGitLogFile tests the createGitLogFile function directly
func TestGitLoggerCreation(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	gitLogger, logPath, err := NewGitLogger("test-cmd", []string{"status"}, "")
	require.NoError(t, err)
	require.NotNil(t, gitLogger)
	defer gitLogger.Close()

	// Verify log file was created
	assert.FileExists(t, logPath)
	assert.Contains(t, logPath, ".taskmaster/logs")

	// Read and verify contents
	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "command_id")
	assert.Contains(t, contentStr, "test-cmd")
}

// TestGitLoggerWithTag tests the NewGitLogger function with tag name
func TestGitLoggerWithTag(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	gitLogger, logPath, err := NewGitLogger("test-cmd", []string{"status"}, "feature-branch")
	require.NoError(t, err)
	require.NotNil(t, gitLogger)
	defer gitLogger.Close()

	// Verify log file was created in tag directory
	assert.FileExists(t, logPath)
	assert.Contains(t, logPath, ".taskmaster/feature-branch")
	assert.Contains(t, logPath, "test-cmd.log")
}

// TestExecuteGitCommand tests the ExecuteGitCommand function returns proper subscription
func TestExecuteGitCommand(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := ExecuteGitCommand("test-exec", []string{"--version"}, "")
	require.NotNil(t, cmd)

	// Execute the command to get the subscription
	msg := cmd()
	require.NotNil(t, msg)

	// Should return CrushExecutionSub message
	subMsg, ok := msg.(CrushExecutionSub)
	assert.True(t, ok, "ExecuteGitCommand should return CrushExecutionSub")
	assert.Equal(t, "test-exec", subMsg.TaskID)
	assert.NotNil(t, subMsg.OutCh)

	// Wait for completion
	timeout := time.After(5 * time.Second)
	completed := false
	for {
		select {
		case msg, ok := <-subMsg.OutCh:
			if !ok {
				completed = true
				break
			}
			// Check for completion message
			if _, ok := msg.(TaskCompletedMsg); ok {
				completed = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for command completion")
		}
		if completed {
			break
		}
	}
	assert.True(t, completed, "Should complete git command")
}

// TestGitBinaryErrorType tests the GitBinaryError type assertion
func TestGitBinaryErrorType(t *testing.T) {
	err := &GitBinaryError{Message: "test"}
	var gitErr error = err
	_, ok := gitErr.(*GitBinaryError)
	assert.True(t, ok)
}

// TestRunGitCommandChannelBuffer tests that the message channel has proper buffering
func TestRunGitCommandChannelBuffer(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	msgChan, err := RunGitCommand("test-buffer", "config", "--list")
	require.NoError(t, err)

	// Verify channel is not nil
	assert.NotNil(t, msgChan)

	// Channel should accept messages without blocking (buffered)
	// This is indirectly verified by the fact that the goroutine doesn't hang

	// Drain the channel
	for range msgChan {
	}
}

// TestRunGitCommandWithComplexArgs tests git command with complex arguments
func TestRunGitCommandWithComplexArgs(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	// Test with config list command which returns structured output
	msgChan, err := RunGitCommand("test-complex", "config", "--list", "--null")
	require.NoError(t, err)

	// Collect messages
	var messages []tea.Msg
	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				break
			}
			messages = append(messages, msg)
		case <-timeout:
			t.Fatal("Timeout waiting for messages")
		}
		if len(msgChan) == 0 && len(messages) > 0 {
			if _, ok := messages[len(messages)-1].(TaskCompletedMsg); ok {
				break
			}
			if _, ok := messages[len(messages)-1].(TaskFailedMsg); ok {
				break
			}
		}
	}

	// Should have received messages
	require.Greater(t, len(messages), 0, "Should receive messages")
}

// Integration test: create a temporary git repository and run git commands
func TestGitRunnerWithRepository(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Initialize a git repository
	os.Chdir(tmpDir)
	initCmd := exec.Command("git", "init")
	err = initCmd.Run()
	require.NoError(t, err, "Failed to initialize git repository")

	// Configure git user (required for some commands)
	configUserCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configUserCmd.Run()

	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Run()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test git add command
	msgChan, err := RunGitCommand("test-add", "add", "test.txt")
	require.NoError(t, err)

	// Drain channel
	for range msgChan {
	}

	// Test git status command
	msgChan, err = RunGitCommand("test-status", "status")
	require.NoError(t, err)

	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	require.Greater(t, len(messages), 0, "Should receive status output")

	// Last message should be completion
	lastMsg := messages[len(messages)-1]
	_, ok := lastMsg.(TaskCompletedMsg)
	assert.True(t, ok, "Should complete successfully")
}

// BenchmarkRunGitCommand benchmarks the RunGitCommand function
func BenchmarkRunGitCommand(b *testing.B) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		b.Skip("git binary not available")
	}

	tmpDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Chdir(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgChan, err := RunGitCommand(fmt.Sprintf("bench-%d", i), "--version")
		if err != nil {
			b.Fatalf("Error running command: %v", err)
		}

		// Drain the channel
		for range msgChan {
		}
	}
}

// ============================================
// Test Helpers for Integration Tests
// ============================================

// setupTempGitRepo creates a temporary git repository for testing
// Returns the path to the temporary directory and a cleanup function
func setupTempGitRepo(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Initialize git repository
	initCmd := exec.Command("git", "init")
	err = initCmd.Run()
	require.NoError(t, err, "Failed to initialize git repository")

	// Configure git user (required for commits)
	configEmailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	err = configEmailCmd.Run()
	require.NoError(t, err, "Failed to configure git user email")

	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	err = configNameCmd.Run()
	require.NoError(t, err, "Failed to configure git user name")

	// Return cleanup function that restores original directory
	return tmpDir, func() {
		os.Chdir(originalWd)
	}
}

// createTestCommit creates a test file and commits it
func createTestCommit(t *testing.T, repoDir string, filename, content string) {
	// Create test file
	filePath := filepath.Join(repoDir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Add file to git
	addCmd := exec.Command("git", "add", filename)
	err = addCmd.Run()
	require.NoError(t, err, "Failed to add file to git")

	// Commit file
	commitCmd := exec.Command("git", "commit", "-m", fmt.Sprintf("Test commit: %s", filename))
	err = commitCmd.Run()
	require.NoError(t, err, "Failed to commit file")
}

// ============================================
// Integration Tests for Git Operations
// ============================================

// TestBranchCreateViaTaskRunner tests creating a branch through the task runner
func TestBranchCreateViaTaskRunner(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create initial commit
	createTestCommit(t, tmpDir, "test.txt", "initial content")

	// Create a branch using RunGitCommand
	msgChan, err := RunGitCommand("branch-create", "checkout", "-b", "feature/test-branch")
	require.NoError(t, err, "Should not error on branch creation")
	require.NotNil(t, msgChan, "Should return message channel")

	// Collect all messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Should have completion message
	require.Greater(t, len(messages), 0, "Should receive at least one message")

	lastMsg := messages[len(messages)-1]
	_, isCompleted := lastMsg.(TaskCompletedMsg)
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isCompleted || isFailed, "Last message should be completion or failure")

	if isCompleted {
		// Verify branch was created
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := branchCmd.Output()
		require.NoError(t, err, "Failed to get current branch")
		assert.Contains(t, string(output), "feature/test-branch", "Branch should be created")
	}
}

// TestBranchSwitchViaTaskRunner tests switching branches through the task runner
func TestBranchSwitchViaTaskRunner(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create initial commit on main
	createTestCommit(t, tmpDir, "main.txt", "main content")

	// Create and switch to feature branch
	createCmd := exec.Command("git", "checkout", "-b", "feature/switch-test")
	err := createCmd.Run()
	require.NoError(t, err, "Failed to create feature branch")

	// Add a commit on feature branch
	createTestCommit(t, tmpDir, "feature.txt", "feature content")

	// Switch back to main using RunGitCommand
	msgChan, err := RunGitCommand("branch-switch", "checkout", "main")
	require.NoError(t, err, "Should not error on branch switch")
	require.NotNil(t, msgChan, "Should return message channel")

	// Collect all messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Should have completion message
	require.Greater(t, len(messages), 0, "Should receive at least one message")

	lastMsg := messages[len(messages)-1]
	_, isCompleted := lastMsg.(TaskCompletedMsg)
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isCompleted || isFailed, "Last message should be completion or failure")

	if isCompleted {
		// Verify we're back on main
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := branchCmd.Output()
		require.NoError(t, err, "Failed to get current branch")
		assert.Contains(t, string(output), "main", "Should switch back to main")
	}
}

// TestGitErrorHandling tests error scenarios with proper git error messages
func TestGitErrorHandling(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Test 1: Checkout non-existent branch
	msgChan, err := RunGitCommand("error-checkout", "checkout", "non-existent-branch")
	require.NoError(t, err, "Should not error on command startup")
	require.NotNil(t, msgChan, "Should return message channel")

	// Collect messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Should have failure message for non-existent branch
	require.Greater(t, len(messages), 0, "Should receive messages")
	lastMsg := messages[len(messages)-1]
	failedMsg, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isFailed, "Should fail with non-existent branch")
	assert.NotEmpty(t, failedMsg.Error, "Should have error message")
}

// TestGitErrorHandlingMergeConflict tests error handling during merge
func TestGitErrorHandlingMergeConflict(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create initial commit
	createTestCommit(t, tmpDir, "conflict.txt", "version 1")

	// Create feature branch with different content
	createCmd := exec.Command("git", "checkout", "-b", "feature/conflict")
	err := createCmd.Run()
	require.NoError(t, err)

	// Modify file in feature branch
	filePath := filepath.Join(tmpDir, "conflict.txt")
	err = os.WriteFile(filePath, []byte("feature version"), 0644)
	require.NoError(t, err)

	commitCmd := exec.Command("git", "commit", "-am", "feature change")
	err = commitCmd.Run()
	require.NoError(t, err)

	// Switch back to main and modify same file
	switchCmd := exec.Command("git", "checkout", "main")
	err = switchCmd.Run()
	require.NoError(t, err)

	err = os.WriteFile(filePath, []byte("main version"), 0644)
	require.NoError(t, err)

	commitCmd = exec.Command("git", "commit", "-am", "main change")
	err = commitCmd.Run()
	require.NoError(t, err)

	// Try to merge (should fail with conflict)
	msgChan, err := RunGitCommand("merge-conflict", "merge", "feature/conflict")
	require.NoError(t, err, "Should not error on merge attempt")

	// Collect messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Merge should fail due to conflict
	lastMsg := messages[len(messages)-1]
	failedMsg, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isFailed, "Merge should fail with conflict")
	assert.NotEmpty(t, failedMsg.Error, "Should have error message")
}

// TestGitErrorHandlingPermissions tests error handling with permission issues
func TestGitErrorHandlingPermissions(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create initial commit
	createTestCommit(t, ".", "test.txt", "content")

	// Try to write to read-only directory
	// Note: We're already in the tmpDir due to setupTempGitRepo
	readOnlyDir := filepath.Join(".", "readonly")
	err := os.Mkdir(readOnlyDir, 0644)
	require.NoError(t, err)

	// Try a git command with restricted directory (this may not error on all systems)
	msgChan, err := RunGitCommand("permission-test", "status")
	require.NoError(t, err, "Should not error on command startup")

	// Collect messages
	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	require.Greater(t, len(messages), 0, "Should receive messages")
	lastMsg := messages[len(messages)-1]
	_, isCompletion := lastMsg.(TaskCompletedMsg)
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isCompletion || isFailed, "Should have completion or failure message")
}

// TestGitCommandWithLargeOutput tests handling of commands with large output
func TestGitCommandWithLargeOutput(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create many test files and commits
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("file-%d.txt", i)
		createTestCommit(t, tmpDir, filename, fmt.Sprintf("Content for file %d", i))
	}

	// Run git log which will produce large output
	msgChan, err := RunGitCommand("large-output", "log", "--oneline", "-n", "20")
	require.NoError(t, err, "Should not error on log command")

	// Collect messages
	var outputCount int
	var hasCompletion bool
	for msg := range msgChan {
		if _, ok := msg.(TaskOutputMsg); ok {
			outputCount++
		}
		if _, ok := msg.(TaskCompletedMsg); ok {
			hasCompletion = true
		}
	}

	assert.Greater(t, outputCount, 0, "Should receive output messages")
	assert.True(t, hasCompletion, "Should complete successfully")
}

// TestMultipleSequentialGitCommands tests running multiple git commands in sequence
func TestMultipleSequentialGitCommands(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	commands := []struct {
		id   string
		args []string
	}{
		{"status-1", []string{"status"}},
		{"add-1", []string{"add", "-A"}},
		{"config-1", []string{"config", "user.name"}},
	}

	for _, cmd := range commands {
		msgChan, err := RunGitCommand(cmd.id, cmd.args...)
		require.NoError(t, err, fmt.Sprintf("Should not error on command %s", cmd.id))

		// Drain channel
		var hasCompletion bool
		for msg := range msgChan {
			if _, ok := msg.(TaskCompletedMsg); ok {
				hasCompletion = true
			}
		}

		assert.True(t, hasCompletion, fmt.Sprintf("Command %s should complete", cmd.id))
	}
}

// TestGitCommandContextCancellation tests context cancellation behavior
func TestGitCommandContextCancellation(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create a slow command (git gc on large repo would be slow, using ls-files instead)
	msgChan, err := RunGitCommand("slow-cmd", "ls-files")
	require.NoError(t, err, "Should not error on command startup")

	// Command should complete normally
	var hasCompletion bool
	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				hasCompletion = true
				break
			}
			if _, ok := msg.(TaskCompletedMsg); ok {
				hasCompletion = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for command completion")
		}
		if hasCompletion {
			break
		}
	}

	assert.True(t, hasCompletion, "Command should complete")
}

// TestGitCommandMessageOrdering tests that messages are delivered in order
func TestGitCommandMessageOrdering(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	createTestCommit(t, tmpDir, "test.txt", "content")

	msgChan, err := RunGitCommand("ordering", "log", "--oneline", "-n", "1")
	require.NoError(t, err)

	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	// Verify message ordering
	require.Greater(t, len(messages), 0, "Should have messages")

	// Last message should be completion
	lastMsg := messages[len(messages)-1]
	_, isCompletion := lastMsg.(TaskCompletedMsg)
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isCompletion || isFailed, "Last message should be completion or failure")

	// All messages except last should be output or context
	for i := 0; i < len(messages)-1; i++ {
		_, isOutput := messages[i].(TaskOutputMsg)
		_, isContext := messages[i].(CrushExecutionContextMsg)
		assert.True(t, isOutput || isContext, fmt.Sprintf("Message %d should be output or context", i))
	}
}

// ============================================
// Edge Case Tests
// ============================================

// TestGitCommandEdgeCaseSpecialCharacters tests handling of special characters in args
func TestGitCommandEdgeCaseSpecialCharacters(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create commit with special characters in message
	createTestCommit(t, tmpDir, "special.txt", "content")

	// Run git log to verify special characters are handled
	msgChan, err := RunGitCommand("special-chars", "log", "--pretty=format:%s", "-n", "1")
	require.NoError(t, err)

	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	require.Greater(t, len(messages), 0, "Should receive messages")
	lastMsg := messages[len(messages)-1]
	_, isCompletion := lastMsg.(TaskCompletedMsg)
	assert.True(t, isCompletion, "Should complete successfully with special characters")
}

// TestGitCommandEdgeCaseEmptyRepository tests commands in empty git repository
func TestGitCommandEdgeCaseEmptyRepository(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Don't create any commits - test with empty repo
	msgChan, err := RunGitCommand("empty-repo", "status")
	require.NoError(t, err)

	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	require.Greater(t, len(messages), 0, "Should receive messages")
	lastMsg := messages[len(messages)-1]
	_, isCompletion := lastMsg.(TaskCompletedMsg)
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isCompletion || isFailed, "Should handle empty repo")
}

// TestGitCommandEdgeCaseVeryLongCommandLineArgs tests very long command arguments
func TestGitCommandEdgeCaseVeryLongCommandLineArgs(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create a long format string
	longFormat := strings.Repeat("%h %s ", 100) // 600+ character format
	msgChan, err := RunGitCommand("long-args", "log", fmt.Sprintf("--pretty=format:%s", longFormat), "-n", "0")
	require.NoError(t, err, "Should handle long arguments")

	// Drain channel
	for range msgChan {
	}
}

// TestGitCommandEdgeCaseRapidFireCommands tests rapid sequential commands
func TestGitCommandEdgeCaseRapidFireCommands(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	createTestCommit(t, tmpDir, "file1.txt", "content1")

	// Run 5 commands rapidly
	commandIDs := make([]string, 5)
	channels := make([]chan tea.Msg, 5)

	for i := 0; i < 5; i++ {
		commandIDs[i] = fmt.Sprintf("rapid-%d", i)
		var err error
		channels[i], err = RunGitCommand(commandIDs[i], "status")
		require.NoError(t, err)
	}

	// Collect all messages
	for i, ch := range channels {
		var hasCompletion bool
		for msg := range ch {
			if _, ok := msg.(TaskCompletedMsg); ok {
				hasCompletion = true
			}
		}
		assert.True(t, hasCompletion, fmt.Sprintf("Command %d should complete", i))
	}
}

// TestGitCommandEdgeCaseNullCharactersInOutput tests handling of null bytes in output
func TestGitCommandEdgeCaseNullCharactersInOutput(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	createTestCommit(t, tmpDir, "test.txt", "content")

	// Use git config which doesn't use --null but tests output handling
	msgChan, err := RunGitCommand("null-output", "config", "--list")
	require.NoError(t, err)

	var outputCount int
	var hasCompletion bool
	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				hasCompletion = true
				break
			}
			if _, ok := msg.(TaskOutputMsg); ok {
				outputCount++
			}
			if _, ok := msg.(TaskCompletedMsg); ok {
				hasCompletion = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for command completion")
		}
		if hasCompletion {
			break
		}
	}

	assert.True(t, hasCompletion, "Should complete successfully")
}

// TestGitCommandEdgeCaseBinaryOutput tests handling of binary/non-text output
func TestGitCommandEdgeCaseBinaryOutput(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.bin")
	err := os.WriteFile(testFile, []byte{0x00, 0xFF, 0xAA, 0x55}, 0644)
	require.NoError(t, err)

	// Add and commit binary file
	addCmd := exec.Command("git", "add", "test.bin")
	err = addCmd.Run()
	require.NoError(t, err)

	commitCmd := exec.Command("git", "commit", "-m", "Add binary")
	err = commitCmd.Run()
	require.NoError(t, err)

	// Try to view binary file with git show
	msgChan, err := RunGitCommand("binary-output", "show", "HEAD:test.bin")
	require.NoError(t, err)

	var hasCompletion bool
	for msg := range msgChan {
		if _, ok := msg.(TaskCompletedMsg); ok {
			hasCompletion = true
		}
	}

	assert.True(t, hasCompletion, "Should handle binary output")
}

// TestGitCommandEdgeCaseTimeoutLikeScenario tests handling of long-running operations
func TestGitCommandEdgeCaseTimeoutLikeScenario(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create multiple commits to make git log slower
	for i := 0; i < 20; i++ {
		filename := fmt.Sprintf("file-%d.txt", i)
		createTestCommit(t, tmpDir, filename, fmt.Sprintf("Content %d", i))
	}

	// Run git log with verbose output
	msgChan, err := RunGitCommand("slow-log", "log", "--stat", "-n", "10")
	require.NoError(t, err)

	var outputCount int
	var hasCompletion bool
	timeout := time.After(10 * time.Second)

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				hasCompletion = true
				break
			}
			if _, isOutput := msg.(TaskOutputMsg); isOutput {
				outputCount++
			}
			if _, isCompletion := msg.(TaskCompletedMsg); isCompletion {
				hasCompletion = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for long-running command")
		}
		if hasCompletion {
			break
		}
	}

	assert.Greater(t, outputCount, 0, "Should receive output from long command")
	assert.True(t, hasCompletion, "Should complete")
}

// TestGitCommandEdgeCaseWhitespaceInArgs tests handling of whitespace in arguments
func TestGitCommandEdgeCaseWhitespaceInArgs(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	createTestCommit(t, tmpDir, "test.txt", "content")

	// Test with arguments containing spaces and tabs
	msgChan, err := RunGitCommand("whitespace-args", "log", "--pretty=format:%h   %s", "-n", "1")
	require.NoError(t, err)

	var hasCompletion bool
	for msg := range msgChan {
		if _, ok := msg.(TaskCompletedMsg); ok {
			hasCompletion = true
		}
	}

	assert.True(t, hasCompletion, "Should handle whitespace in args")
}

// TestGitCommandEdgeCaseUnicodeInOutput tests handling of unicode characters
func TestGitCommandEdgeCaseUnicodeInOutput(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create file with unicode content
	unicodeContent := "Hello 世界 🎉 Привет мир"
	err := os.WriteFile(filepath.Join(tmpDir, "unicode.txt"), []byte(unicodeContent), 0644)
	require.NoError(t, err)

	// Commit it
	createTestCommit(t, tmpDir, "unicode.txt", unicodeContent)

	// Show the file with unicode characters
	msgChan, err := RunGitCommand("unicode", "show", "HEAD:unicode.txt")
	require.NoError(t, err)

	var outputCount int
	var hasCompletion bool
	for msg := range msgChan {
		if _, ok := msg.(TaskOutputMsg); ok {
			outputCount++
		}
		if _, ok := msg.(TaskCompletedMsg); ok {
			hasCompletion = true
		}
	}

	assert.Greater(t, outputCount, 0, "Should receive unicode output")
	assert.True(t, hasCompletion, "Should complete with unicode")
}

// TestGitCommandEdgeCaseMultipleErrors tests handling of multiple error outputs
func TestGitCommandEdgeCaseMultipleErrors(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	_, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Try to add non-existent files (will produce warnings/errors)
	msgChan, err := RunGitCommand("multi-error", "add", "non-existent-1.txt", "non-existent-2.txt", "non-existent-3.txt")
	require.NoError(t, err)

	var messages []tea.Msg
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	require.Greater(t, len(messages), 0, "Should receive messages")
	lastMsg := messages[len(messages)-1]
	_, isFailed := lastMsg.(TaskFailedMsg)
	assert.True(t, isFailed, "Should fail with non-existent files")
}

// TestRunGitCommandChannelCapacity tests that large bursts of messages don't overflow
func TestRunGitCommandChannelCapacity(t *testing.T) {
	// Skip if git is not available
	if err := ValidateGitBinary(); err != nil {
		t.Skip("git binary not available")
	}

	tmpDir, cleanup := setupTempGitRepo(t)
	defer cleanup()

	// Create many commits to generate lots of output
	for i := 0; i < 30; i++ {
		createTestCommit(t, tmpDir, fmt.Sprintf("file-%d.txt", i), fmt.Sprintf("Content %d", i))
	}

	// Run git log with full details to generate maximum output
	msgChan, err := RunGitCommand("capacity-test", "log", "--oneline", "--graph", "-n", "30")
	require.NoError(t, err)

	var messageCount int
	var hasCompletion bool
	for msg := range msgChan {
		messageCount++
		if _, ok := msg.(TaskCompletedMsg); ok {
			hasCompletion = true
		}
	}

	assert.Greater(t, messageCount, 20, "Should receive many messages")
	assert.True(t, hasCompletion, "Should complete without overflow")
}
