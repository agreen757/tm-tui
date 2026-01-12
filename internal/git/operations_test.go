package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGitCreateBranchOperationValidateArgs tests branch name validation
func TestGitCreateBranchOperationValidateArgs(t *testing.T) {
	tests := []struct {
		name      string
		branchName string
		shouldErr bool
		errMsg    string
	}{
		{
			name:       "valid branch name",
			branchName: "feature/new-feature",
			shouldErr:  false,
		},
		{
			name:       "empty branch name",
			branchName: "",
			shouldErr:  true,
			errMsg:     "branch name cannot be empty",
		},
		{
			name:       "branch name with spaces",
			branchName: "feature new feature",
			shouldErr:  true,
			errMsg:     "branch name cannot contain spaces",
		},
		{
			name:       "complex valid name",
			branchName: "feat/JIRA-123-new-feature",
			shouldErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewGitCreateBranchOperation("/tmp", tt.branchName)
			err := op.ValidateArgs()

			if tt.shouldErr && err == nil {
				t.Errorf("expected error for branch name %q", tt.branchName)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for branch name %q: %v", tt.branchName, err)
			}
			if tt.shouldErr && err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

// TestGitSwitchBranchOperationValidateArgs tests branch name validation for switch
func TestGitSwitchBranchOperationValidateArgs(t *testing.T) {
	tests := []struct {
		name      string
		branchName string
		shouldErr bool
	}{
		{
			name:       "valid branch name",
			branchName: "main",
			shouldErr:  false,
		},
		{
			name:       "empty branch name",
			branchName: "",
			shouldErr:  true,
		},
		{
			name:       "branch with slash",
			branchName: "feature/test",
			shouldErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewGitSwitchBranchOperation("/tmp", tt.branchName)
			err := op.ValidateArgs()

			if tt.shouldErr && err == nil {
				t.Errorf("expected error for branch name %q", tt.branchName)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for branch name %q: %v", tt.branchName, err)
			}
		})
	}
}

// TestGitCreateBranchOperationBuildCommand tests command construction
func TestGitCreateBranchOperationBuildCommand(t *testing.T) {
	repoPath := "/test/repo"
	branchName := "new-branch"
	op := NewGitCreateBranchOperation(repoPath, branchName)

	cmd := op.BuildCommand()

	if cmd.Dir != repoPath {
		t.Errorf("expected repo path %q, got %q", repoPath, cmd.Dir)
	}

	expectedArgs := []string{"git", "checkout", "-b", branchName}
	actualArgs := append([]string{cmd.Path}, cmd.Args[1:]...)
	
	if len(actualArgs) != len(expectedArgs) {
		t.Errorf("expected args length %d, got %d", len(expectedArgs), len(actualArgs))
	}

	for i, expected := range expectedArgs {
		if i < len(actualArgs) && !bytes.Contains([]byte(actualArgs[i]), []byte(expected)) && actualArgs[i] != expected {
			// Allow for path differences in git executable
			if i > 0 {
				t.Errorf("expected arg %q, got %q", expected, actualArgs[i])
			}
		}
	}
}

// TestGitSwitchBranchOperationBuildCommand tests switch command construction
func TestGitSwitchBranchOperationBuildCommand(t *testing.T) {
	repoPath := "/test/repo"
	branchName := "main"
	op := NewGitSwitchBranchOperation(repoPath, branchName)

	cmd := op.BuildCommand()

	if cmd.Dir != repoPath {
		t.Errorf("expected repo path %q, got %q", repoPath, cmd.Dir)
	}

	// Just verify the command was built
	if cmd == nil {
		t.Errorf("expected non-nil command")
	}
}

// TestGitCreateBranchOperationParseOutput tests output parsing
func TestGitCreateBranchOperationParseOutput(t *testing.T) {
	op := NewGitCreateBranchOperation("/tmp", "test-branch")

	outputStr := "Switched to a new branch 'test-branch'\n"
	output := []byte(outputStr)

	err := op.ParseOutput(output)
	if err != nil {
		t.Errorf("unexpected error parsing output: %v", err)
	}

	expected := "Switched to a new branch 'test-branch'"
	if op.GetOutput() != expected {
		t.Errorf("expected output %q, got %q", expected, op.GetOutput())
	}
}

// TestGitCreateBranchOperationHandleError tests error handling
func TestGitCreateBranchOperationHandleError(t *testing.T) {
	op := NewGitCreateBranchOperation("/tmp", "test-branch")

	tests := []struct {
		name     string
		input    error
		hasError bool
		contains string
	}{
		{
			name:     "nil error",
			input:    nil,
			hasError: false,
		},
		{
			name:     "exec error",
			input:    exec.Command("git").Run(),
			hasError: true,
			contains: "failed to create branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := op.HandleError(tt.input)

			if tt.hasError && result == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.hasError && result != nil {
				t.Errorf("expected nil error, got %v", result)
			}
			if tt.contains != "" && result != nil && !bytes.Contains([]byte(result.Error()), []byte(tt.contains)) {
				t.Errorf("expected error containing %q, got %q", tt.contains, result.Error())
			}
		})
	}
}

// TestGitSwitchBranchOperationHandleError tests error handling for switch
func TestGitSwitchBranchOperationHandleError(t *testing.T) {
	op := NewGitSwitchBranchOperation("/tmp", "main")

	err := op.HandleError(nil)
	if err != nil {
		t.Errorf("expected nil for nil input, got %v", err)
	}

	testErr := exec.Command("git").Run()
	result := op.HandleError(testErr)
	if result == nil {
		t.Errorf("expected error for non-nil input")
	}
}

// TestGitCreateBranchOperationGetOperationType tests operation type
func TestGitCreateBranchOperationGetOperationType(t *testing.T) {
	op := NewGitCreateBranchOperation("/tmp", "test")
	expected := "create-branch"
	if op.GetOperationType() != expected {
		t.Errorf("expected operation type %q, got %q", expected, op.GetOperationType())
	}
}

// TestGitSwitchBranchOperationGetOperationType tests operation type
func TestGitSwitchBranchOperationGetOperationType(t *testing.T) {
	op := NewGitSwitchBranchOperation("/tmp", "main")
	expected := "switch-branch"
	if op.GetOperationType() != expected {
		t.Errorf("expected operation type %q, got %q", expected, op.GetOperationType())
	}
}

// TestGitOperationFactoryCreateBranchOp tests factory creates correct operation type
func TestGitOperationFactoryCreateBranchOp(t *testing.T) {
	factory := NewGitOperationFactory()
	op, err := factory.NewGitOperation(OperationCreateBranch, "/tmp", "new-branch")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if op == nil {
		t.Errorf("expected operation, got nil")
	}

	expected := "create-branch"
	if op.GetOperationType() != expected {
		t.Errorf("expected operation type %q, got %q", expected, op.GetOperationType())
	}
}

// TestGitOperationFactorySwitchBranchOp tests factory creates switch operation
func TestGitOperationFactorySwitchBranchOp(t *testing.T) {
	factory := NewGitOperationFactory()
	op, err := factory.NewGitOperation(OperationSwitchBranch, "/tmp", "main")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if op == nil {
		t.Errorf("expected operation, got nil")
	}

	expected := "switch-branch"
	if op.GetOperationType() != expected {
		t.Errorf("expected operation type %q, got %q", expected, op.GetOperationType())
	}
}

// TestGitOperationFactoryMissingArgs tests factory handles missing arguments
func TestGitOperationFactoryMissingArgs(t *testing.T) {
	factory := NewGitOperationFactory()

	tests := []struct {
		name    string
		opType  OperationType
		args    []string
		wantErr bool
	}{
		{
			name:    "create branch without args",
			opType:  OperationCreateBranch,
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "switch branch without args",
			opType:  OperationSwitchBranch,
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := factory.NewGitOperation(tt.opType, "/tmp", tt.args...)
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if tt.wantErr && op != nil {
				t.Errorf("expected nil operation with error")
			}
		})
	}
}

// TestGitOperationFactoryInvalidOperation tests factory handles unknown operation types
func TestGitOperationFactoryInvalidOperation(t *testing.T) {
	factory := NewGitOperationFactory()
	op, err := factory.NewGitOperation(OperationType("invalid-op"), "/tmp", "arg")

	if err == nil {
		t.Errorf("expected error for invalid operation type")
	}
	if op != nil {
		t.Errorf("expected nil operation for invalid type")
	}
}

// TestGitOperationFactoryNotImplementedOperations tests factory for future operations
func TestGitOperationFactoryNotImplementedOperations(t *testing.T) {
	factory := NewGitOperationFactory()

	tests := []struct {
		name   string
		opType OperationType
	}{
		{name: "commit operation", opType: OperationCommit},
		{name: "push operation", opType: OperationPush},
		{name: "pull operation", opType: OperationPull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := factory.NewGitOperation(tt.opType, "/tmp")
			if err == nil {
				t.Errorf("expected error for not-implemented operation")
			}
			if op != nil {
				t.Errorf("expected nil operation for not-implemented type")
			}
		})
	}
}

// TestGitCreateBranchOperationIntegration tests the complete operation lifecycle
func TestGitCreateBranchOperationIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create a commit so we have something to branch from
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = tempDir
	_ = configCmd.Run()

	configName := exec.Command("git", "config", "user.name", "Test User")
	configName.Dir = tempDir
	_ = configName.Run()

	createFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(createFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = tempDir
	_ = addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = tempDir
	_ = commitCmd.Run()

	// Test the operation
	op := NewGitCreateBranchOperation(tempDir, "test-branch")
	err = op.ValidateArgs()
	if err != nil {
		t.Errorf("validation failed: %v", err)
	}

	result, err := ExecuteOperation(op)
	if err != nil {
		t.Errorf("execute operation failed: %v", err)
	}

	if result == "" {
		t.Errorf("expected non-empty result")
	}

	// Verify the branch was created
	getBranchCmd := exec.Command("git", "branch", "--show-current")
	getBranchCmd.Dir = tempDir
	output, err := getBranchCmd.Output()
	if err != nil {
		t.Errorf("failed to get current branch: %v", err)
	}

	if string(bytes.TrimSpace(output)) != "test-branch" {
		t.Errorf("expected branch 'test-branch', got %q", string(bytes.TrimSpace(output)))
	}
}

// TestGitSwitchBranchOperationIntegration tests the complete operation lifecycle
func TestGitSwitchBranchOperationIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-switch-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo and create branches
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = tempDir
	_ = configCmd.Run()

	configName := exec.Command("git", "config", "user.name", "Test User")
	configName.Dir = tempDir
	_ = configName.Run()

	// Create initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = tempDir
	_ = addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = tempDir
	_ = commitCmd.Run()

	// Rename master to main if needed
	renameCmd := exec.Command("git", "branch", "-m", "main")
	renameCmd.Dir = tempDir
	_ = renameCmd.Run()

	// Create another branch
	createCmd := exec.Command("git", "checkout", "-b", "feature")
	createCmd.Dir = tempDir
	if err := createCmd.Run(); err != nil {
		t.Errorf("failed to create feature branch: %v", err)
	}

	// Test switching back to main
	op := NewGitSwitchBranchOperation(tempDir, "main")
	err = op.ValidateArgs()
	if err != nil {
		t.Errorf("validation failed: %v", err)
	}

	result, err := ExecuteOperation(op)
	if err != nil {
		t.Errorf("execute operation failed: %v", err)
	}

	if result == "" {
		t.Errorf("expected non-empty result")
	}

	// Verify we switched to main
	getBranchCmd := exec.Command("git", "branch", "--show-current")
	getBranchCmd.Dir = tempDir
	output, err := getBranchCmd.Output()
	if err != nil {
		t.Errorf("failed to get current branch: %v", err)
	}

	if string(bytes.TrimSpace(output)) != "main" {
		t.Errorf("expected branch 'main', got %q", string(bytes.TrimSpace(output)))
	}
}

// TestExecuteOperation tests the standardized operation execution
func TestExecuteOperation(t *testing.T) {
	// Create operation with invalid args to test validation
	op := NewGitCreateBranchOperation("/tmp", "")

	_, err := ExecuteOperation(op)
	if err == nil {
		t.Errorf("expected error for invalid args")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("validation failed")) {
		t.Errorf("expected validation error, got: %v", err)
	}
}

// TestGitOperationExtensibility verifies new operations can be easily added
func TestGitOperationExtensibility(t *testing.T) {
	// This test demonstrates that new operations can be added
	// by implementing the GitOperation interface

	factory := NewGitOperationFactory()

	// Verify status operation is now implemented
	_, err := factory.NewGitOperation(OperationStatus, "/tmp")
	if err != nil {
		t.Errorf("expected status operation to be available: %v", err)
	}

	// Verify placeholders exist for future operations
	for _, opType := range []OperationType{
		OperationCommit,
		OperationPush,
		OperationPull,
	} {
		_, err := factory.NewGitOperation(opType, "/tmp")
		if err == nil {
			t.Errorf("expected error for unimplemented operation: %s", opType)
		}
		// This is expected - operations aren't implemented yet
		t.Logf("Confirmed %s is a valid placeholder for future implementation", opType)
	}
}

// TestGitStatusOperationValidateArgs tests status operation validation
func TestGitStatusOperationValidateArgs(t *testing.T) {
	op := NewGitStatusOperation("/tmp")
	err := op.ValidateArgs()
	if err != nil {
		t.Errorf("unexpected error for status operation: %v", err)
	}
}

// TestGitStatusOperationBuildCommand tests status command construction
func TestGitStatusOperationBuildCommand(t *testing.T) {
	repoPath := "/test/repo"
	op := NewGitStatusOperation(repoPath)

	cmd := op.BuildCommand()

	if cmd.Dir != repoPath {
		t.Errorf("expected repo path %q, got %q", repoPath, cmd.Dir)
	}

	if cmd == nil {
		t.Errorf("expected non-nil command")
	}
}

// TestGitStatusOperationParseOutput tests output parsing
func TestGitStatusOperationParseOutput(t *testing.T) {
	op := NewGitStatusOperation("/tmp")

	outputStr := " M file1.go\n A file2.go\n"
	output := []byte(outputStr)

	err := op.ParseOutput(output)
	if err != nil {
		t.Errorf("unexpected error parsing output: %v", err)
	}

	expected := "M file1.go\n A file2.go"
	if op.GetStatusOutput() != expected {
		t.Errorf("expected output %q, got %q", expected, op.GetStatusOutput())
	}
}

// TestGitStatusOperationHandleError tests error handling
func TestGitStatusOperationHandleError(t *testing.T) {
	op := NewGitStatusOperation("/tmp")

	err := op.HandleError(nil)
	if err != nil {
		t.Errorf("expected nil for nil input, got %v", err)
	}

	testErr := exec.Command("git").Run()
	result := op.HandleError(testErr)
	if result == nil {
		t.Errorf("expected error for non-nil input")
	}
}

// TestGitStatusOperationGetOperationType tests operation type
func TestGitStatusOperationGetOperationType(t *testing.T) {
	op := NewGitStatusOperation("/tmp")
	expected := "status"
	if op.GetOperationType() != expected {
		t.Errorf("expected operation type %q, got %q", expected, op.GetOperationType())
	}
}

// TestGitOperationFactoryStatusOperation tests factory creates status operation
func TestGitOperationFactoryStatusOperation(t *testing.T) {
	factory := NewGitOperationFactory()
	op, err := factory.NewGitOperation(OperationStatus, "/tmp")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if op == nil {
		t.Errorf("expected operation, got nil")
	}

	expected := "status"
	if op.GetOperationType() != expected {
		t.Errorf("expected operation type %q, got %q", expected, op.GetOperationType())
	}
}

// TestGitOperationFactoryRegisterOperation tests registration mechanism
func TestGitOperationFactoryRegisterOperation(t *testing.T) {
	factory := NewGitOperationFactory()

	// Define a custom operation type
	customOpType := OperationType("custom-op")

	// Register a custom operation constructor
	customConstructor := func(repoPath string, args ...string) (GitOperation, error) {
		// Return a mock operation for testing
		return NewGitStatusOperation(repoPath), nil
	}

	factory.RegisterOperation(customOpType, customConstructor)

	// Verify the custom operation can be created
	op, err := factory.NewGitOperation(customOpType, "/tmp")
	if err != nil {
		t.Errorf("expected custom operation to be registered: %v", err)
	}
	if op == nil {
		t.Errorf("expected operation, got nil")
	}
}

// TestGitStatusOperationIntegration tests the complete operation lifecycle
func TestGitStatusOperationIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-status-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create a test file to make the repo dirty
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test the operation
	op := NewGitStatusOperation(tempDir)
	err = op.ValidateArgs()
	if err != nil {
		t.Errorf("validation failed: %v", err)
	}

	result, err := ExecuteOperation(op)
	if err != nil {
		t.Errorf("execute operation failed: %v", err)
	}

	// Result should show untracked file
	if !strings.Contains(result, "test.txt") {
		t.Errorf("expected status output to contain 'test.txt', got: %q", result)
	}
}
