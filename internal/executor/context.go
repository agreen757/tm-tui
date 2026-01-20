package executor

import (
	"context"
	"os"
	"syscall"
	"time"
)

// OperationType defines the category of operations for timeout purposes
type OperationType int

const (
	// OperationTypeGitQuick - quick git operations like status, branch listing (30 seconds)
	OperationTypeGitQuick OperationType = iota
	// OperationTypeGitFetch - git fetch/pull operations (5 minutes)
	OperationTypeGitFetch
	// OperationTypeGitPush - git push operations (3 minutes)
	OperationTypeGitPush
	// OperationTypeGitClone - git clone operations (10 minutes)
	OperationTypeGitClone
	// OperationTypeCrushExecution - Crush AI execution (30 minutes)
	OperationTypeCrushExecution
	// OperationTypeDefault - default timeout (5 minutes)
	OperationTypeDefault
)

// String returns a string representation of the OperationType
func (o OperationType) String() string {
	switch o {
	case OperationTypeGitQuick:
		return "git-quick"
	case OperationTypeGitFetch:
		return "git-fetch"
	case OperationTypeGitPush:
		return "git-push"
	case OperationTypeGitClone:
		return "git-clone"
	case OperationTypeCrushExecution:
		return "crush-execution"
	default:
		return "default"
	}
}

// GetTimeoutForOperation returns the appropriate timeout for the given operation type
func GetTimeoutForOperation(opType OperationType) time.Duration {
	switch opType {
	case OperationTypeGitQuick:
		return 30 * time.Second
	case OperationTypeGitFetch:
		return 5 * time.Minute
	case OperationTypeGitPush:
		return 3 * time.Minute
	case OperationTypeGitClone:
		return 10 * time.Minute
	case OperationTypeCrushExecution:
		return 30 * time.Minute
	default:
		return 5 * time.Minute
	}
}

// DetermineGitOperationType identifies the timeout category for a git command
// based on its arguments
func DetermineGitOperationType(args []string) OperationType {
	if len(args) == 0 {
		return OperationTypeGitQuick
	}

	cmd := args[0]
	switch cmd {
	case "fetch", "pull":
		return OperationTypeGitFetch
	case "push":
		return OperationTypeGitPush
	case "clone":
		return OperationTypeGitClone
	case "status", "branch", "log", "show", "diff", "checkout":
		return OperationTypeGitQuick
	default:
		return OperationTypeDefault
	}
}

// ContextWithTimeout creates a context with the appropriate timeout for the operation
// Accepts a parent context (which can be cancelled externally) and operation type
// Returns the context and cancel function
func ContextWithTimeout(parentCtx context.Context, opType OperationType) (context.Context, context.CancelFunc) {
	timeout := GetTimeoutForOperation(opType)
	return context.WithTimeout(parentCtx, timeout)
}

// ConfigureProcessGroup sets up the process to be in its own process group
// for graceful termination on Unix-like systems
func ConfigureProcessGroup() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// IsDeadlineExceeded checks if the context error is a deadline exceeded error
func IsDeadlineExceeded(err error) bool {
	return err != nil && err == context.DeadlineExceeded
}

// IsCancelled checks if the context error is a cancellation error
func IsCancelled(err error) bool {
	return err != nil && err == context.Canceled
}

// FormatContextError returns a user-friendly error message for context errors
func FormatContextError(ctx context.Context, timeout time.Duration) string {
	if ctx.Err() == context.DeadlineExceeded {
		return "command timed out after " + timeout.String()
	}
	if ctx.Err() == context.Canceled {
		return "command was cancelled"
	}
	return ""
}

// WaitWithContext waits for a channel to receive a value or for the context to be cancelled
// Returns true if the channel had data, false if context was cancelled
func WaitWithContext(ctx context.Context, ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	case <-ctx.Done():
		return false
	}
}

// ProcessGroupSignal sends a signal to a process group for graceful termination
// On Unix systems, this kills the entire process group
// On Windows, it terminates the process (process groups work differently)
func ProcessGroupSignal(pid int, sig os.Signal) error {
	// On Unix, negative pid targets the process group
	// On Windows, negative pid is invalid but Kill() handles it
	if pid <= 0 {
		return nil // Already terminated
	}

	// For Unix systems, use process group kill
	// For Windows, this will just terminate the process
	pgid := -pid
	return syscall.Kill(pgid, sig.(syscall.Signal))
}
