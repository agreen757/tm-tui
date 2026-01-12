# Git Runner API Documentation

## Overview

The Git Runner provides a high-level API for executing git commands with real-time streaming output, proper error handling, and automatic logging. All git operations in Task Master TUI are executed through this API.

## Core Functions

### ValidateGitBinary()

Checks if the git CLI binary is available on the system.

```go
func ValidateGitBinary() error
```

**Returns:**
- `nil` - Git binary found and executable
- `*GitBinaryError` - Git binary not found or not accessible

**Usage Example:**

```go
import "github.com/agreen757/tm-tui/internal/ui/dialog"

// Check if git is available before attempting any git operations
if err := dialog.ValidateGitBinary(); err != nil {
    fmt.Println("Git not available:", err.Error())
    return
}
// Proceed with git operations
```

**Error Handling:**

```go
err := dialog.ValidateGitBinary()
if err != nil {
    if gitErr, ok := err.(*dialog.GitBinaryError); ok {
        fmt.Println("Git error:", gitErr.Message)
    }
}
```

### RunGitCommand()

Executes a git command with streaming output. The output is streamed via a channel of tea.Msg messages, and automatically logged to disk.

```go
func RunGitCommand(commandID string, args ...string) (chan tea.Msg, error)
```

**Parameters:**
- `commandID` (string, required) - Unique identifier for logging correlation. Cannot be empty.
- `args` (variadic string) - Git command arguments (e.g., "status", "branch", "-a"). Must contain at least one argument.

**Returns:**
- `chan tea.Msg` - Channel that receives output messages (TaskOutputMsg, TaskCompletedMsg, TaskFailedMsg). Channel is automatically closed when the command completes.
- `error` - Error if validation or startup fails.

**Message Types Sent on Channel:**

1. **TaskOutputMsg** - Each line of git output
   ```go
   type TaskOutputMsg struct {
       TaskID string  // matches commandID parameter
       Output string  // single line of output
   }
   ```

2. **TaskCompletedMsg** - Command succeeded
   ```go
   type TaskCompletedMsg struct {
       TaskID string  // matches commandID parameter
   }
   ```

3. **TaskFailedMsg** - Command failed
   ```go
   type TaskFailedMsg struct {
       TaskID  string  // matches commandID parameter
       Error   string  // error message
       Message string  // user-friendly message
   }
   ```

**Logging:**
- All output automatically written to `.taskmaster/logs/git-command-<commandID>-<timestamp>.log`
- Log file includes command header, output, and completion status
- Logging continues even if errors occur

**Usage Example - Simple Status Check:**

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/agreen757/tm-tui/internal/ui/dialog"
)

// Execute git status
msgChan, err := dialog.RunGitCommand("get-status", "status", "--porcelain")
if err != nil {
    fmt.Println("Failed to start git command:", err.Error())
    return
}

// Process messages from the channel
for msg := range msgChan {
    switch m := msg.(type) {
    case dialog.TaskOutputMsg:
        fmt.Println("Output:", m.Output)
    case dialog.TaskCompletedMsg:
        fmt.Println("Command completed successfully")
    case dialog.TaskFailedMsg:
        fmt.Println("Command failed:", m.Error)
    }
}
```

**Usage Example - Branch Listing:**

```go
// Get list of all branches
msgChan, err := dialog.RunGitCommand("list-branches", "branch", "-a")
if err != nil {
    return err
}

var branches []string
for msg := range msgChan {
    if output, ok := msg.(dialog.TaskOutputMsg); ok {
        branches = append(branches, output.Output)
    }
}
```

**Error Scenarios:**

1. **Empty Command ID:**
   ```go
   msgChan, err := dialog.RunGitCommand("", "status")
   // err is non-nil: "command ID cannot be empty"
   ```

2. **No Arguments:**
   ```go
   msgChan, err := dialog.RunGitCommand("my-cmd")
   // err is non-nil: "git command arguments cannot be empty"
   ```

3. **Git Binary Not Found:**
   ```go
   msgChan, err := dialog.RunGitCommand("my-cmd", "status")
   // err is *GitBinaryError with message about git installation
   ```

4. **Invalid Git Command (Execution Failure):**
   ```go
   msgChan, err := dialog.RunGitCommand("my-cmd", "invalid-command")
   // err is nil (command started), but messages contain TaskFailedMsg
   ```

5. **Pipe Creation Error:**
   ```go
   msgChan, err := dialog.RunGitCommand("my-cmd", "status")
   // err might be: "failed to create stdout pipe: ..."
   ```

### ExecuteGitCommand()

High-level wrapper around RunGitCommand that integrates with Bubble Tea's command system and returns a proper tea.Cmd for use in Update() methods.

```go
func ExecuteGitCommand(commandID string, args []string, tagName string) tea.Cmd
```

**Parameters:**
- `commandID` (string) - Unique identifier for logging correlation
- `args` ([]string) - Git command arguments
- `tagName` (string) - Tag for organizing logs in `.taskmaster/<tagName>/<commandID>.log`. Empty string uses default `.taskmaster/logs/` directory.

**Returns:**
- `tea.Cmd` - Bubble Tea command that can be returned from Update(). The command returns a CrushExecutionSub message.

**Integration with Bubble Tea:**

This function is designed to be used in Bubble Tea's Update() method with tea.Sequence() to coordinate multiple operations:

```go
func (m *MyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            // Sequence ensures TaskStartedMsg is processed before ExecuteGitCommand
            return m, tea.Sequence(
                func() tea.Msg {
                    return dialog.TaskStartedMsg{
                        TaskID:    "my-git-op",
                        TaskTitle: "Running git operation",
                    }
                },
                dialog.ExecuteGitCommand("my-git-op", []string{"status"}, "my-project"),
            )
        }
    }
    return m, nil
}
```

**Usage Example - Branch Switch with Task Runner:**

```go
// In a dialog's HandleSelection method
return tea.Sequence(
    func() tea.Msg {
        return dialog.TaskStartedMsg{
            TaskID:    "git-switch-branch",
            TaskTitle: "Switch Branch: " + branchName,
        }
    },
    dialog.ExecuteGitCommand("git-switch-branch", []string{"checkout", branchName}, d.tagName),
)
```

## Error Types

### GitBinaryError

Represents an error when the git binary is not found or not accessible.

```go
type GitBinaryError struct {
    Message string
}

func (e *GitBinaryError) Error() string
```

**Properties:**
- `Message` - Human-readable error message explaining the issue and how to resolve it

**Example Usage:**

```go
if err := dialog.ValidateGitBinary(); err != nil {
    if gitErr, ok := err.(*dialog.GitBinaryError); ok {
        fmt.Println("Git setup error:", gitErr.Message)
        // Output: "git binary not found. Please install git: https://git-scm.com/download"
    }
}
```

## Message Types

### TaskOutputMsg

Represents a single line of output from the git command.

```go
type TaskOutputMsg struct {
    TaskID string  // Matches the commandID parameter
    Output string  // Single line of output from git
}
```

### TaskCompletedMsg

Indicates the git command completed successfully.

```go
type TaskCompletedMsg struct {
    TaskID string  // Matches the commandID parameter
}
```

### TaskFailedMsg

Indicates the git command failed.

```go
type TaskFailedMsg struct {
    TaskID  string  // Matches the commandID parameter
    Error   string  // Raw error message from git
    Message string  // User-friendly error description
}
```

## Output Logging

All git command output is automatically logged to disk for auditing and debugging purposes.

**Log File Location:**
```
.taskmaster/logs/git-command-<commandID>-<timestamp>.log
```

**Log File Format:**
```
=== Git Command Log ===
Command ID: my-command
Started: 2026-01-12T22:30:00Z
Command: git status --porcelain
===================

<output lines>

===================
Completed: 2026-01-12T22:30:02Z
Status: Success
===================
```

**Log Directory Creation:**
- Directory is created automatically if it doesn't exist
- If directory creation fails, logging is skipped but execution continues
- If log file creation fails, execution continues without logging to file

## Best Practices

### 1. Always Validate Git Before Use

```go
// Bad: Assumes git is available
msgChan, _ := dialog.RunGitCommand("my-cmd", "status")

// Good: Validates first
if err := dialog.ValidateGitBinary(); err != nil {
    handleError(err)
    return
}
msgChan, err := dialog.RunGitCommand("my-cmd", "status")
if err != nil {
    handleError(err)
    return
}
```

### 2. Use Descriptive Command IDs

```go
// Bad: Generic IDs make logs hard to correlate
dialog.RunGitCommand("cmd1", "status")

// Good: Descriptive IDs help with debugging
dialog.RunGitCommand("git-status-check", "status")
dialog.RunGitCommand("git-switch-branch-main", "checkout", "main")
```

### 3. Handle All Message Types

```go
for msg := range msgChan {
    switch m := msg.(type) {
    case dialog.TaskOutputMsg:
        // Process output line
        processLine(m.Output)
    case dialog.TaskCompletedMsg:
        // Handle success
        showSuccess()
    case dialog.TaskFailedMsg:
        // Handle failure with details
        showError(m.Error, m.Message)
    }
}
```

### 4. Don't Block on Channel Operations

```go
// Good: Process messages in a goroutine or use non-blocking select
go func() {
    for msg := range msgChan {
        // Process messages
    }
}()

// Or with Bubble Tea's built-in message handling
```

### 5. Coordinate with Bubble Tea Using tea.Sequence()

```go
// Good: Use tea.Sequence to ensure proper ordering
return tea.Sequence(
    func() tea.Msg { return TaskStartedMsg{...} },
    ExecuteGitCommand("my-cmd", args, tagName),
)

// This ensures:
// 1. TaskStartedMsg is processed first (shows UI)
// 2. ExecuteGitCommand runs second (starts command execution)
```

## Common Patterns

### Pattern: Wait for Command Completion

```go
msgChan, err := dialog.RunGitCommand("my-cmd", "status")
if err != nil {
    return err
}

completed := false
for msg := range msgChan {
    switch msg.(type) {
    case dialog.TaskCompletedMsg:
        completed = true
    case dialog.TaskFailedMsg:
        completed = true
    }
}

if !completed {
    return fmt.Errorf("command did not complete")
}
```

### Pattern: Collect All Output

```go
msgChan, err := dialog.RunGitCommand("my-cmd", "branch", "-a")
if err != nil {
    return err
}

var output []string
for msg := range msgChan {
    if m, ok := msg.(dialog.TaskOutputMsg); ok {
        output = append(output, m.Output)
    }
}
```

### Pattern: Timeout with Context

The git runner uses context.Background() internally, but you can wrap it for timeouts:

```go
type TimeoutFunc func(context.Context) tea.Cmd

func ExecuteGitCommandWithTimeout(commandID string, args []string, tagName string, timeout time.Duration) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()
        
        // Your implementation here
        return dialog.CrushExecutionSub{TaskID: commandID, OutCh: make(chan tea.Msg)}
    }
}
```

## Integration with Task Runner

The git runner integrates seamlessly with the Task Runner modal:

1. **Initiation**: Call `ExecuteGitCommand()` which returns a tea.Cmd
2. **Task Started**: TaskStartedMsg is sent first (optional, but recommended)
3. **Execution**: ExecuteGitCommand sends CrushExecutionSub to start streaming
4. **Output**: TaskOutputMsg messages are routed to the Task Runner modal
5. **Completion**: TaskCompletedMsg or TaskFailedMsg terminates the operation
6. **UI Update**: Modal shows completion status and allows closing

## Reference: Git Commands Used in Task Master TUI

- **Status**: `git status --porcelain`
- **List Branches**: `git branch -a`
- **Switch Branch**: `git checkout <branch>`
- **Create Branch**: `git checkout -b <branch>`
- **Recent Commits**: `git log --oneline -n <count>`

## Troubleshooting

### "git binary not found"

**Cause**: Git is not installed or not in PATH
**Solution**: 
```bash
# macOS
brew install git

# Linux
sudo apt-get install git

# Windows
choco install git
```

### Command executes but no output appears

**Cause**: Output is being buffered or command has no output
**Solution**: Check the log file at `.taskmaster/logs/git-command-<id>-<timestamp>.log`

### Task Runner doesn't show up

**Cause**: TaskStartedMsg wasn't sent before ExecuteGitCommand
**Solution**: Use tea.Sequence to ensure TaskStartedMsg is sent first

### Large Output Truncation

**Cause**: Scanner buffer limit exceeded (1MB per line)
**Solution**: Git commands shouldn't produce lines this long. If they do, pre-process the output.

## Testing Git Runner Functions

See `git_runner_test.go` for comprehensive test examples covering:
- Git binary validation
- Command execution with output
- Error handling
- Message channel behavior
- Logging functionality
- Concurrent execution
- Large output handling

To run tests:
```bash
go test ./internal/ui/dialog -run TestRunGitCommand -v
go test ./internal/ui/dialog -run TestExecuteGitCommand -v
```
