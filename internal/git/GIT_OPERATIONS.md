# Adding New Git Operations

This document explains the extensible framework for adding new git operations to Task Master TUI.

## Architecture Overview

The git operations framework consists of:

1. **GitOperation Interface** - Standardized contract for all git operations
2. **Concrete Implementations** - Operations that implement the interface
3. **GitOperationFactory** - Factory pattern for creating operations
4. **Registration System** - Dynamic registration for extensibility

## Step-by-Step Guide: Adding a New Git Operation

### Step 1: Implement the GitOperation Interface

Create a new struct that embeds `BaseGitOperation` and implements the required methods:

```go
package git

type GitMyNewOperation struct {
    BaseGitOperation
    // Operation-specific fields
    parameter string
    output    string
}

// NewGitMyNewOperation creates a new operation
func NewGitMyNewOperation(repoPath, parameter string) *GitMyNewOperation {
    return &GitMyNewOperation{
        BaseGitOperation: NewBaseGitOperation(repoPath),
        parameter:        parameter,
    }
}

// BuildCommand constructs the git command to execute
func (op *GitMyNewOperation) BuildCommand() *exec.Cmd {
    cmd := exec.Command("git", "my-command", op.parameter)
    cmd.Dir = op.repoPath
    return cmd
}

// ValidateArgs validates the operation's arguments
func (op *GitMyNewOperation) ValidateArgs() error {
    if op.parameter == "" {
        return fmt.Errorf("parameter cannot be empty")
    }
    return nil
}

// ParseOutput parses the command output
func (op *GitMyNewOperation) ParseOutput(output []byte) error {
    op.output = strings.TrimSpace(string(output))
    return nil
}

// HandleError wraps errors with operation-specific context
func (op *GitMyNewOperation) HandleError(err error) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("failed to perform operation: %w", err)
}

// GetOperationType returns the operation type identifier
func (op *GitMyNewOperation) GetOperationType() string {
    return "my-new-operation"
}

// GetOutput returns the command output (operation-specific accessor)
func (op *GitMyNewOperation) GetOutput() string {
    return op.output
}
```

### Step 2: Define an OperationType Constant

Add a new constant to the OperationType enumeration:

```go
const (
    // OperationMyNewOperation identifies the custom operation
    OperationMyNewOperation OperationType = "my-new-operation"
)
```

### Step 3: Register the Operation

Two registration approaches are available:

#### Option A: Register with Factory Instance

```go
// During application initialization
factory := NewGitOperationFactory()
factory.RegisterOperation(OperationMyNewOperation, func(repoPath string, args ...string) (GitOperation, error) {
    if len(args) < 1 {
        return nil, fmt.Errorf("operation requires a parameter")
    }
    op := NewGitMyNewOperation(repoPath, args[0])
    if err := op.ValidateArgs(); err != nil {
        return nil, err
    }
    return op, nil
})
```

#### Option B: Register Globally

```go
// At package init time or before factory creation
func init() {
    RegisterGlobalOperation(OperationMyNewOperation, func(repoPath string, args ...string) (GitOperation, error) {
        if len(args) < 1 {
            return nil, fmt.Errorf("operation requires a parameter")
        }
        op := NewGitMyNewOperation(repoPath, args[0])
        if err := op.ValidateArgs(); err != nil {
            return nil, err
        }
        return op, nil
    })
}
```

### Step 4: Write Tests

Create comprehensive tests for your operation:

```go
func TestGitMyNewOperationValidateArgs(t *testing.T) {
    op := NewGitMyNewOperation("/tmp", "valid-param")
    if err := op.ValidateArgs(); err != nil {
        t.Errorf("unexpected validation error: %v", err)
    }
}

func TestGitMyNewOperationBuildCommand(t *testing.T) {
    op := NewGitMyNewOperation("/test/repo", "param")
    cmd := op.BuildCommand()
    if cmd.Dir != "/test/repo" {
        t.Errorf("incorrect repo path")
    }
}

func TestGitMyNewOperationIntegration(t *testing.T) {
    // Real git integration test
    tempDir := createTempGitRepo(t)
    defer cleanupTempDir(t, tempDir)
    
    op := NewGitMyNewOperation(tempDir, "test-param")
    result, err := ExecuteOperation(op)
    if err != nil {
        t.Errorf("operation failed: %v", err)
    }
    
    // Verify results
    if result == "" {
        t.Errorf("expected non-empty output")
    }
}

// Test factory registration
func TestGitOperationFactoryMyNewOperation(t *testing.T) {
    factory := NewGitOperationFactory()
    op, err := factory.NewGitOperation(OperationMyNewOperation, "/tmp", "param")
    if err != nil {
        t.Errorf("factory failed to create operation: %v", err)
    }
    if op == nil {
        t.Errorf("expected operation, got nil")
    }
}
```

### Step 5: Integrate with UI Components

Use the operation in UI dialogs:

```go
package dialog

func (d *MyDialog) executeGitOperation() {
    factory := git.NewGitOperationFactory()
    
    op, err := factory.NewGitOperation(
        git.OperationMyNewOperation,
        d.repoPath,
        d.parameter,
    )
    if err != nil {
        d.handleError(err)
        return
    }
    
    result, err := git.ExecuteOperation(op)
    if err != nil {
        d.handleError(err)
        return
    }
    
    d.handleSuccess(result)
}
```

## Architecture Patterns

### The Interface Contract

Every `GitOperation` must implement these methods:

- **BuildCommand()**: Constructs and returns an `*exec.Cmd` ready to execute
- **ValidateArgs()**: Returns an error if arguments are invalid
- **ParseOutput([]byte)**: Processes the command output
- **HandleError(error)**: Wraps execution errors with context
- **GetOperationType()**: Returns a string identifier for the operation

### Execution Lifecycle

The `ExecuteOperation` helper function manages the complete lifecycle:

```
1. Validate arguments      → op.ValidateArgs()
2. Build command          → op.BuildCommand()
3. Execute command        → cmd.CombinedOutput()
4. Handle errors          → op.HandleError()
5. Parse output           → op.ParseOutput()
6. Return result          → result string
```

This ensures consistent error handling and validation across all operations.

### BaseGitOperation Struct

The `BaseGitOperation` provides common functionality:

- Stores the repository path
- Can be embedded to share code
- Reduces boilerplate in concrete implementations

Example:

```go
type GitMyOperation struct {
    BaseGitOperation  // Embeds repoPath
    specificField     string
}

// Automatically has access to: op.repoPath
```

## Future Operations (Placeholders)

The following operations are registered but not yet implemented:

- `OperationCommit` - Future commit functionality
- `OperationPush` - Future push functionality
- `OperationPull` - Future pull functionality

These are ready for implementation following the steps above.

## Best Practices

### 1. Validation First
Always validate arguments before attempting to execute:

```go
func (op *GitMyOperation) ValidateArgs() error {
    if op.param == "" {
        return fmt.Errorf("parameter cannot be empty")
    }
    if strings.Contains(op.param, " ") {
        return fmt.Errorf("parameter cannot contain spaces")
    }
    return nil
}
```

### 2. Meaningful Error Messages
Wrap errors with operation-specific context:

```go
func (op *GitMyOperation) HandleError(err error) error {
    return fmt.Errorf("failed to my-operation for param '%s': %w", op.param, err)
}
```

### 3. Output Storage
Store parsed output for retrieval:

```go
func (op *GitMyOperation) ParseOutput(output []byte) error {
    op.parsedOutput = strings.TrimSpace(string(output))
    op.outputLines = strings.Split(op.parsedOutput, "\n")
    return nil
}

func (op *GitMyOperation) GetParsedOutput() string {
    return op.parsedOutput
}
```

### 4. Integration Tests
Always include integration tests with real git repositories:

```go
func TestGitMyOperationIntegration(t *testing.T) {
    tempDir, _ := os.MkdirTemp("", "git-test-")
    defer os.RemoveAll(tempDir)
    
    // Initialize git repo
    exec.Command("git", "init").Dir = tempDir
    
    // Test your operation
    op := NewGitMyOperation(tempDir, "args")
    result, err := ExecuteOperation(op)
    // Assertions...
}
```

## Example: Implementing GitStatusOperation

Here's the actual implementation of `GitStatusOperation` to use as reference:

```go
type GitStatusOperation struct {
    BaseGitOperation
    statusOutput string
}

func NewGitStatusOperation(repoPath string) *GitStatusOperation {
    return &GitStatusOperation{
        BaseGitOperation: NewBaseGitOperation(repoPath),
    }
}

func (op *GitStatusOperation) BuildCommand() *exec.Cmd {
    cmd := exec.Command("git", "status", "--porcelain")
    cmd.Dir = op.repoPath
    return cmd
}

func (op *GitStatusOperation) ValidateArgs() error {
    return nil // No args needed
}

func (op *GitStatusOperation) ParseOutput(output []byte) error {
    op.statusOutput = strings.TrimSpace(string(output))
    return nil
}

func (op *GitStatusOperation) HandleError(err error) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("failed to get git status: %w", err)
}

func (op *GitStatusOperation) GetOperationType() string {
    return "status"
}

func (op *GitStatusOperation) GetStatusOutput() string {
    return op.statusOutput
}
```

## Testing Your Implementation

Run all git tests to ensure your implementation works:

```bash
go test ./internal/git/... -v
```

Run only your operation tests:

```bash
go test ./internal/git/... -run YourOperationName -v
```

## Documentation and Code Comments

When adding operations, include:

1. **Godoc comments** for all exported types and functions
2. **Inline comments** explaining complex logic
3. **Examples** in documentation or tests
4. **Error scenarios** in test cases

Example:

```go
// GitMyNewOperation represents a git operation that does something specific.
// It implements the GitOperation interface and can be used with the factory.
type GitMyNewOperation struct {
    BaseGitOperation
    param string
}

// NewGitMyNewOperation creates a new operation instance.
// It returns an error if the parameter is invalid.
func NewGitMyNewOperation(repoPath, param string) (*GitMyNewOperation, error) {
    // Implementation...
}
```

## Troubleshooting

### Operation not found in factory
- Check that the OperationType constant is defined
- Verify the registration call includes the correct OperationType
- Ensure the constructor function has the correct signature

### Validation always fails
- Review ValidateArgs() logic
- Check for off-by-one errors in string length checks
- Verify regex patterns if using string matching

### Integration tests fail
- Ensure git is installed and available
- Check that temp directories are created with proper permissions
- Verify git repo initialization succeeds before testing operation

## Contributing New Operations

When submitting new git operations:

1. Follow the interface contract exactly
2. Include comprehensive tests (validation, error handling, integration)
3. Update this documentation with example usage
4. Register placeholders for operations that will be implemented later
5. Ensure backward compatibility with existing operations

