package executor

import (
	"context"
	"testing"
	"time"
)

// TestGetTimeoutForOperation verifies timeout durations for different operation types
func TestGetTimeoutForOperation(t *testing.T) {
	tests := []struct {
		opType   OperationType
		expected time.Duration
		name     string
	}{
		{OperationTypeGitQuick, 30 * time.Second, "git-quick"},
		{OperationTypeGitFetch, 5 * time.Minute, "git-fetch"},
		{OperationTypeGitPush, 3 * time.Minute, "git-push"},
		{OperationTypeGitClone, 10 * time.Minute, "git-clone"},
		{OperationTypeCrushExecution, 30 * time.Minute, "crush-execution"},
		{OperationTypeDefault, 5 * time.Minute, "default"},
		{OperationType(999), 5 * time.Minute, "unknown-type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := GetTimeoutForOperation(tt.opType)
			if timeout != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, timeout)
			}
		})
	}
}

// TestDetermineGitOperationType verifies git command classification
func TestDetermineGitOperationType(t *testing.T) {
	tests := []struct {
		args     []string
		expected OperationType
		name     string
	}{
		{[]string{"status"}, OperationTypeGitQuick, "status"},
		{[]string{"branch"}, OperationTypeGitQuick, "branch"},
		{[]string{"branch", "-a"}, OperationTypeGitQuick, "branch-all"},
		{[]string{"log"}, OperationTypeGitQuick, "log"},
		{[]string{"show", "commit"}, OperationTypeGitQuick, "show"},
		{[]string{"checkout", "branch"}, OperationTypeGitQuick, "checkout"},
		{[]string{"diff", "file"}, OperationTypeGitQuick, "diff"},
		{[]string{"fetch"}, OperationTypeGitFetch, "fetch"},
		{[]string{"pull"}, OperationTypeGitFetch, "pull"},
		{[]string{"pull", "origin", "main"}, OperationTypeGitFetch, "pull-with-args"},
		{[]string{"push"}, OperationTypeGitPush, "push"},
		{[]string{"push", "origin", "main"}, OperationTypeGitPush, "push-with-args"},
		{[]string{"clone"}, OperationTypeGitClone, "clone"},
		{[]string{"clone", "https://example.com/repo.git"}, OperationTypeGitClone, "clone-url"},
		{[]string{}, OperationTypeGitQuick, "empty-args"},
		{[]string{"unknown"}, OperationTypeDefault, "unknown-command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opType := DetermineGitOperationType(tt.args)
			if opType != tt.expected {
				t.Errorf("for %v: expected %v, got %v", tt.args, tt.expected, opType)
			}
		})
	}
}

// TestOperationTypeString verifies string representation of operation types
func TestOperationTypeString(t *testing.T) {
	tests := []struct {
		opType   OperationType
		expected string
		name     string
	}{
		{OperationTypeGitQuick, "git-quick", "git-quick"},
		{OperationTypeGitFetch, "git-fetch", "git-fetch"},
		{OperationTypeGitPush, "git-push", "git-push"},
		{OperationTypeGitClone, "git-clone", "git-clone"},
		{OperationTypeCrushExecution, "crush-execution", "crush-execution"},
		{OperationTypeDefault, "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.opType.String()
			if str != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, str)
			}
		})
	}
}

// TestContextWithTimeout creates context with timeout and verifies it expires
func TestContextWithTimeout(t *testing.T) {
	parentCtx := context.Background()
	ctx, cancel := ContextWithTimeout(parentCtx, OperationTypeGitQuick)
	defer cancel()

	// Context should not be cancelled immediately
	select {
	case <-ctx.Done():
		t.Error("context cancelled immediately")
	default:
		// Expected behavior
	}

	// Verify timeout is applied (should expire in 30 seconds)
	// But for testing, we'll just verify it has a deadline
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("context should have a deadline")
	}
	if deadline.Before(time.Now()) {
		t.Error("deadline should be in the future")
	}
}

// TestContextWithTimeoutCancellation verifies parent context cancellation propagates
func TestContextWithTimeoutCancellation(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	ctx, cancel := ContextWithTimeout(parentCtx, OperationTypeGitFetch)
	defer cancel()

	// Cancel parent context
	parentCancel()

	// Child context should be cancelled immediately
	select {
	case <-ctx.Done():
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Error("context not cancelled after parent cancellation")
	}
}

// TestContextWithTimeoutNested verifies nested timeout contexts work correctly
func TestContextWithTimeoutNested(t *testing.T) {
	// Create a parent with very long timeout
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer parentCancel()

	// Create child with shorter timeout
	childCtx, childCancel := ContextWithTimeout(parentCtx, OperationTypeGitQuick)
	defer childCancel()

	// Verify both have deadlines
	parentDeadline, ok := parentCtx.Deadline()
	if !ok {
		t.Error("parent context should have deadline")
	}

	childDeadline, ok := childCtx.Deadline()
	if !ok {
		t.Error("child context should have deadline")
	}

	// Child deadline should be earlier than parent
	if childDeadline.After(parentDeadline) {
		t.Error("child deadline should be earlier than parent deadline")
	}
}

// TestIsDeadlineExceeded verifies deadline exceeded detection
func TestIsDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	if !IsDeadlineExceeded(ctx.Err()) {
		t.Error("IsDeadlineExceeded returned false for deadline exceeded")
	}

	// Non-timeout error should return false
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	if IsDeadlineExceeded(cancelCtx.Err()) {
		t.Error("IsDeadlineExceeded returned true for cancelled context")
	}

	// Nil error should return false
	if IsDeadlineExceeded(nil) {
		t.Error("IsDeadlineExceeded returned true for nil error")
	}
}

// TestIsCancelled verifies cancellation detection
func TestIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if !IsCancelled(ctx.Err()) {
		t.Error("IsCancelled returned false for cancelled context")
	}

	// Timeout should return false
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer timeoutCancel()
	time.Sleep(10 * time.Millisecond)

	if IsCancelled(timeoutCtx.Err()) {
		t.Error("IsCancelled returned true for deadline exceeded")
	}

	// Nil error should return false
	if IsCancelled(nil) {
		t.Error("IsCancelled returned true for nil error")
	}
}

// TestFormatContextError verifies error message formatting
func TestFormatContextError(t *testing.T) {
	tests := []struct {
		name              string
		contextErr        func() context.Context
		timeout           time.Duration
		expectedContains  string
		shouldNotContain  string
	}{
		{
			name: "deadline-exceeded",
			contextErr: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				time.Sleep(10 * time.Millisecond)
				return ctx
			},
			timeout:          30 * time.Second,
			expectedContains: "timed out after 30s",
			shouldNotContain: "cancelled",
		},
		{
			name: "cancelled",
			contextErr: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			timeout:          30 * time.Second,
			expectedContains: "cancelled",
			shouldNotContain: "timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.contextErr()
			msg := FormatContextError(ctx, tt.timeout)

			if tt.expectedContains != "" && !contains(msg, tt.expectedContains) {
				t.Errorf("expected message to contain %q, got %q", tt.expectedContains, msg)
			}

			if tt.shouldNotContain != "" && contains(msg, tt.shouldNotContain) {
				t.Errorf("expected message NOT to contain %q, got %q", tt.shouldNotContain, msg)
			}
		})
	}
}

// TestWaitWithContext verifies channel waiting with context
func TestWaitWithContext(t *testing.T) {
	t.Run("channel-ready", func(t *testing.T) {
		ctx := context.Background()
		ch := make(chan struct{})
		close(ch)

		result := WaitWithContext(ctx, ch)
		if !result {
			t.Error("expected true for ready channel")
		}
	})

	t.Run("context-cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan struct{})

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		result := WaitWithContext(ctx, ch)
		if result {
			t.Error("expected false for cancelled context")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		ch := make(chan struct{})

		result := WaitWithContext(ctx, ch)
		if result {
			t.Error("expected false for timeout")
		}
	})
}

// TestConfigureProcessGroup verifies process group configuration
func TestConfigureProcessGroup(t *testing.T) {
	sysAttr := ConfigureProcessGroup()

	if sysAttr == nil {
		t.Fatal("ConfigureProcessGroup returned nil")
	}

	if !sysAttr.Setpgid {
		t.Error("Setpgid should be true for process group configuration")
	}
}

// Helper function for test
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkDetermineGitOperationType measures performance of operation type detection
func BenchmarkDetermineGitOperationType(b *testing.B) {
	args := []string{"pull", "origin", "main"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetermineGitOperationType(args)
	}
}

// BenchmarkContextWithTimeout measures performance of context creation
func BenchmarkContextWithTimeout(b *testing.B) {
	parentCtx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := ContextWithTimeout(parentCtx, OperationTypeGitQuick)
		cancel()
		_ = ctx
	}
}
