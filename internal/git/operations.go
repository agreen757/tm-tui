package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitOperation is an interface that standardizes git operations
// It defines methods for command construction, argument handling, output parsing, and error handling
type GitOperation interface {
	// BuildCommand constructs the exec.Cmd for this operation
	// Returns the command with all necessary arguments set
	BuildCommand() *exec.Cmd

	// ValidateArgs validates the arguments for this operation
	// Returns an error if the arguments are invalid
	ValidateArgs() error

	// ParseOutput parses the output from the git command
	// Takes the raw output bytes and returns any parsing errors
	ParseOutput(output []byte) error

	// HandleError handles errors from git command execution
	// Wraps the error with operation-specific context
	HandleError(err error) error

	// GetOperationType returns the type of git operation (for logging/debugging)
	GetOperationType() string
}

// BaseGitOperation provides common functionality for all git operations
type BaseGitOperation struct {
	repoPath string
}

// NewBaseGitOperation creates a new base git operation
func NewBaseGitOperation(repoPath string) BaseGitOperation {
	return BaseGitOperation{
		repoPath: repoPath,
	}
}

// GitCreateBranchOperation represents a git branch creation operation
type GitCreateBranchOperation struct {
	BaseGitOperation
	branchName string
	output     string
}

// NewGitCreateBranchOperation creates a new create branch operation
func NewGitCreateBranchOperation(repoPath, branchName string) *GitCreateBranchOperation {
	return &GitCreateBranchOperation{
		BaseGitOperation: NewBaseGitOperation(repoPath),
		branchName:       branchName,
	}
}

// BuildCommand constructs the exec.Cmd for branch creation
func (op *GitCreateBranchOperation) BuildCommand() *exec.Cmd {
	cmd := exec.Command("git", "checkout", "-b", op.branchName)
	cmd.Dir = op.repoPath
	return cmd
}

// ValidateArgs validates the branch name
func (op *GitCreateBranchOperation) ValidateArgs() error {
	if op.branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.Contains(op.branchName, " ") {
		return fmt.Errorf("branch name cannot contain spaces")
	}
	return nil
}

// ParseOutput parses the output from branch creation command
func (op *GitCreateBranchOperation) ParseOutput(output []byte) error {
	op.output = strings.TrimSpace(string(output))
	return nil
}

// HandleError handles errors specific to branch creation
func (op *GitCreateBranchOperation) HandleError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to create branch '%s': %w", op.branchName, err)
}

// GetOperationType returns the operation type
func (op *GitCreateBranchOperation) GetOperationType() string {
	return "create-branch"
}

// GetBranchName returns the created branch name
func (op *GitCreateBranchOperation) GetBranchName() string {
	return op.branchName
}

// GetOutput returns the command output
func (op *GitCreateBranchOperation) GetOutput() string {
	return op.output
}

// GitSwitchBranchOperation represents a git branch switch operation
type GitSwitchBranchOperation struct {
	BaseGitOperation
	branchName string
	output     string
}

// NewGitSwitchBranchOperation creates a new switch branch operation
func NewGitSwitchBranchOperation(repoPath, branchName string) *GitSwitchBranchOperation {
	return &GitSwitchBranchOperation{
		BaseGitOperation: NewBaseGitOperation(repoPath),
		branchName:       branchName,
	}
}

// BuildCommand constructs the exec.Cmd for branch switching
func (op *GitSwitchBranchOperation) BuildCommand() *exec.Cmd {
	cmd := exec.Command("git", "checkout", op.branchName)
	cmd.Dir = op.repoPath
	return cmd
}

// ValidateArgs validates the branch name
func (op *GitSwitchBranchOperation) ValidateArgs() error {
	if op.branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	return nil
}

// ParseOutput parses the output from branch switch command
func (op *GitSwitchBranchOperation) ParseOutput(output []byte) error {
	op.output = strings.TrimSpace(string(output))
	return nil
}

// HandleError handles errors specific to branch switching
func (op *GitSwitchBranchOperation) HandleError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to switch to branch '%s': %w", op.branchName, err)
}

// GetOperationType returns the operation type
func (op *GitSwitchBranchOperation) GetOperationType() string {
	return "switch-branch"
}

// GetBranchName returns the target branch name
func (op *GitSwitchBranchOperation) GetBranchName() string {
	return op.branchName
}

// GetOutput returns the command output
func (op *GitSwitchBranchOperation) GetOutput() string {
	return op.output
}

// GitStatusOperation represents a git status operation
// This is a sample implementation demonstrating extensibility
type GitStatusOperation struct {
	BaseGitOperation
	statusOutput string
}

// NewGitStatusOperation creates a new git status operation
func NewGitStatusOperation(repoPath string) *GitStatusOperation {
	return &GitStatusOperation{
		BaseGitOperation: NewBaseGitOperation(repoPath),
	}
}

// BuildCommand constructs the exec.Cmd for git status
func (op *GitStatusOperation) BuildCommand() *exec.Cmd {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = op.repoPath
	return cmd
}

// ValidateArgs validates the operation (no specific args needed)
func (op *GitStatusOperation) ValidateArgs() error {
	// Status operation doesn't require arguments
	return nil
}

// ParseOutput parses the status output
func (op *GitStatusOperation) ParseOutput(output []byte) error {
	op.statusOutput = strings.TrimSpace(string(output))
	return nil
}

// HandleError handles errors specific to status checking
func (op *GitStatusOperation) HandleError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to get git status: %w", err)
}

// GetOperationType returns the operation type
func (op *GitStatusOperation) GetOperationType() string {
	return "status"
}

// GetStatusOutput returns the status output
func (op *GitStatusOperation) GetStatusOutput() string {
	return op.statusOutput
}

// GitOperationRegistry stores constructors for registered git operations
// This allows for dynamic registration of new operations at runtime
var GitOperationRegistry = make(map[OperationType]func(string, ...string) (GitOperation, error))

// GitOperationFactory creates git operations based on type
type GitOperationFactory struct {
	registry map[OperationType]func(string, ...string) (GitOperation, error)
}

// NewGitOperationFactory creates a new git operation factory
func NewGitOperationFactory() *GitOperationFactory {
	// Initialize with built-in operations
	factory := &GitOperationFactory{
		registry: make(map[OperationType]func(string, ...string) (GitOperation, error)),
	}
	factory.registerBuiltinOperations()
	return factory
}

// registerBuiltinOperations registers the built-in operations
func (f *GitOperationFactory) registerBuiltinOperations() {
	f.registry[OperationCreateBranch] = func(repoPath string, args ...string) (GitOperation, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("create-branch operation requires branch name argument")
		}
		op := NewGitCreateBranchOperation(repoPath, args[0])
		if err := op.ValidateArgs(); err != nil {
			return nil, err
		}
		return op, nil
	}

	f.registry[OperationSwitchBranch] = func(repoPath string, args ...string) (GitOperation, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("switch-branch operation requires branch name argument")
		}
		op := NewGitSwitchBranchOperation(repoPath, args[0])
		if err := op.ValidateArgs(); err != nil {
			return nil, err
		}
		return op, nil
	}

	f.registry[OperationStatus] = func(repoPath string, args ...string) (GitOperation, error) {
		op := NewGitStatusOperation(repoPath)
		if err := op.ValidateArgs(); err != nil {
			return nil, err
		}
		return op, nil
	}
}

// RegisterOperation registers a custom git operation constructor
// This allows external code to add new operation types to the factory
func (f *GitOperationFactory) RegisterOperation(opType OperationType, constructor func(string, ...string) (GitOperation, error)) {
	f.registry[opType] = constructor
}

// RegisterGlobalOperation registers a custom git operation in the global registry
// Use this for registering operations before factory creation
func RegisterGlobalOperation(opType OperationType, constructor func(string, ...string) (GitOperation, error)) {
	GitOperationRegistry[opType] = constructor
}

// OperationType defines the type of git operation
type OperationType string

const (
	// OperationCreateBranch is the create branch operation type
	OperationCreateBranch OperationType = "create-branch"
	// OperationSwitchBranch is the switch branch operation type
	OperationSwitchBranch OperationType = "switch-branch"
	// OperationCommit is a placeholder for future commit operations
	OperationCommit OperationType = "commit"
	// OperationPush is a placeholder for future push operations
	OperationPush OperationType = "push"
	// OperationPull is a placeholder for future pull operations
	OperationPull OperationType = "pull"
	// OperationStatus is a placeholder for future status operations
	OperationStatus OperationType = "status"
)

// NewGitOperation creates a new git operation of the specified type
// Returns an error if the operation type is unknown or if required arguments are missing
// This method checks the factory registry and then the global registry
func (f *GitOperationFactory) NewGitOperation(opType OperationType, repoPath string, args ...string) (GitOperation, error) {
	// First check factory registry
	if constructor, exists := f.registry[opType]; exists {
		return constructor(repoPath, args...)
	}

	// Then check global registry
	if constructor, exists := GitOperationRegistry[opType]; exists {
		return constructor(repoPath, args...)
	}

	return nil, fmt.Errorf("unknown operation type: %s", opType)
}

// ExecuteOperation executes a git operation and handles the complete lifecycle
// This provides a standardized way to run any git operation
func ExecuteOperation(op GitOperation) (string, error) {
	// Validate arguments
	if err := op.ValidateArgs(); err != nil {
		return "", fmt.Errorf("[%s] validation failed: %w", op.GetOperationType(), err)
	}

	// Build and execute command
	cmd := op.BuildCommand()
	output, err := cmd.CombinedOutput()

	// Handle execution errors
	if err != nil {
		return "", op.HandleError(err)
	}

	// Parse output
	if err := op.ParseOutput(output); err != nil {
		return "", fmt.Errorf("[%s] output parsing failed: %w", op.GetOperationType(), err)
	}

	return strings.TrimSpace(string(output)), nil
}
