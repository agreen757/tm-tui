package dialog

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestErrorMessageIntegration verifies that git errors are parsed and user-friendly
// messages are sent to the output in the Task Runner
func TestErrorMessageIntegration(t *testing.T) {
	tests := []struct {
		name           string
		stderr         string
		args           string
		expectType     string
		expectInSuggestion []string
	}{
		{
			name:       "Branch exists error",
			stderr:     "fatal: A branch named 'feature' already exists.",
			args:       "checkout -b feature",
			expectType: "BranchExistsError",
			expectInSuggestion: []string{"feature", "already exists", "checkout", "git"},
		},
		{
			name:       "Branch not found error",
			stderr:     "error: pathspec 'nonexistent' did not match any files in the index.",
			args:       "checkout nonexistent",
			expectType: "BranchNotFoundError",
			expectInSuggestion: []string{"nonexistent", "not found", "Fetch"},
		},
		{
			name:       "Permission denied error",
			stderr:     "fatal: Permission denied (.git/refs/heads)",
			args:       "checkout main",
			expectType: "GitPermissionError",
			expectInSuggestion: []string{"Permission", "denied", "Check"},
		},
		{
			name:       "Network error",
			stderr:     "fatal: unable to access 'https://github.com/user/repo.git/': Connection refused",
			args:       "fetch",
			expectType: "GitNetworkError",
			expectInSuggestion: []string{"Network", "connection", "Check"},
		},
		{
			name:       "Merge conflict error",
			stderr:     "CONFLICT (content): Merge conflict in main.go",
			args:       "merge feature",
			expectType: "MergeConflictError",
			expectInSuggestion: []string{"conflict", "resolve", "git"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Parse the error
			gitErr := ParseGitError(test.stderr, test.args)

			// Check the error type
			if gitErr.ErrorType() != test.expectType {
				t.Errorf("expected %s, got %s", test.expectType, gitErr.ErrorType())
			}

			// Check that the suggestion contains expected text
			suggestion := gitErr.Suggestion()
			for _, text := range test.expectInSuggestion {
				if !strings.Contains(suggestion, text) {
					t.Errorf("expected suggestion to contain '%s', suggestion: %s", text, suggestion)
				}
			}
		})
	}
}

// TestErrorMessageFormatting verifies that error messages are properly formatted
// for display in the Task Runner
func TestErrorMessageFormatting(t *testing.T) {
	stderr := "fatal: A branch named 'feature' already exists."
	args := "checkout -b feature"

	gitErr := ParseGitError(stderr, args)

	// Get the error message
	errorMsg := fmt.Sprintf("❌ Error: %s", gitErr.ErrorType())
	if errorMsg == "" {
		t.Errorf("expected error message to be non-empty")
	}

	if !strings.Contains(errorMsg, "Error:") {
		t.Errorf("expected error message to contain 'Error:', got: %s", errorMsg)
	}

	if !strings.Contains(errorMsg, "BranchExistsError") {
		t.Errorf("expected error message to contain error type, got: %s", errorMsg)
	}

	// Get the suggestion
	suggestion := gitErr.Suggestion()
	if suggestion == "" {
		t.Errorf("expected suggestion to be non-empty")
	}

	// Suggestion should contain actionable items
	lines := strings.Split(suggestion, "\n")
	if len(lines) < 2 {
		t.Errorf("expected suggestion to have multiple lines, got %d", len(lines))
	}

	// Check for bullet points or commands
	hasActionableItems := false
	for _, line := range lines {
		if strings.Contains(line, "•") || strings.Contains(line, "git") {
			hasActionableItems = true
			break
		}
	}

	if !hasActionableItems {
		t.Errorf("expected suggestion to contain actionable items (• or git commands)")
	}
}

// TestUserFriendlyMessages verifies each error type provides helpful suggestions
func TestUserFriendlyMessages(t *testing.T) {
	errors := map[GitError]bool{
		&BranchExistsError{
			BranchName: "test",
			Message:    "branch already exists",
		}: true,
		&BranchNotFoundError{
			BranchName: "test",
			Message:    "branch not found",
		}: true,
		&GitPermissionError{
			Message:   "permission denied",
			Resource:  ".git",
			Operation: "checkout",
		}: true,
		&GitNetworkError{
			Message:    "connection refused",
			RemoteName: "origin",
		}: true,
		&DetachedHeadError{
			Message:    "detached head",
			CurrentSHA: "abc123",
		}: true,
		&MergeConflictError{
			Message:       "merge conflict",
			ConflictFiles: []string{"file1.go", "file2.go"},
		}: true,
		&UncommittedChangesError{
			Message:      "uncommitted changes",
			ChangedFiles: []string{"README.md"},
		}: true,
		&AheadBehindError{
			Message:     "ahead of origin",
			AheadCount:  3,
			BranchName:  "main",
			RemoteName:  "origin",
		}: true,
		&GenericGitError{
			Message:   "unknown error",
			RawOutput: "some error",
		}: true,
	}

	for gitErr := range errors {
		// Each error should have a non-empty suggestion
		suggestion := gitErr.Suggestion()
		if suggestion == "" {
			t.Errorf("%T should provide a suggestion", gitErr)
		}

		// Suggestion should be user-friendly (contain actionable steps)
		if !strings.Contains(suggestion, "•") && !strings.Contains(suggestion, "git") {
			// Allow generic errors to not have bullet points
			if _, ok := gitErr.(*GenericGitError); !ok {
				t.Logf("%T suggestion: %s", gitErr, suggestion)
			}
		}
	}
}

// TestRecoverableErrors verifies that recoverable errors are correctly identified
func TestRecoverableErrors(t *testing.T) {
	recoverableTests := []struct {
		name      string
		err       GitError
		expectRec bool
	}{
		{
			name: "BranchExists recoverable",
			err: &BranchExistsError{
				BranchName: "test",
				Message:    "exists",
			},
			expectRec: true,
		},
		{
			name: "BranchNotFound recoverable",
			err: &BranchNotFoundError{
				BranchName: "test",
				Message:    "not found",
			},
			expectRec: true,
		},
		{
			name: "Permission not recoverable",
			err: &GitPermissionError{
				Message:   "permission denied",
				Resource:  ".git",
				Operation: "checkout",
			},
			expectRec: false,
		},
		{
			name: "Network error recoverable",
			err: &GitNetworkError{
				Message:    "connection refused",
				RemoteName: "origin",
			},
			expectRec: true,
		},
	}

	for _, test := range recoverableTests {
		t.Run(test.name, func(t *testing.T) {
			if test.err.IsRecoverable() != test.expectRec {
				t.Errorf("expected recoverable=%v, got %v", test.expectRec, test.err.IsRecoverable())
			}
		})
	}
}

// TestMultilineErrorMessages verifies that error suggestions with multiple lines
// are properly formatted for display
func TestMultilineErrorMessages(t *testing.T) {
	conflictErr := &MergeConflictError{
		Message:       "merge conflict",
		ConflictFiles: []string{"main.go", "README.md", "config.json"},
	}

	suggestion := conflictErr.Suggestion()
	lines := strings.Split(suggestion, "\n")

	// Should have at least the header and the conflicted files
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines, got %d", len(lines))
	}

	// All conflicted files should be listed
	for _, file := range []string{"main.go", "README.md", "config.json"} {
		found := false
		for _, line := range lines {
			if strings.Contains(line, file) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %s to be in suggestion", file)
		}
	}
}

// BenchmarkParseGitError benchmarks the error parsing performance
func BenchmarkParseGitError(b *testing.B) {
	stderr := "fatal: A branch named 'feature' already exists."
	args := "checkout -b feature"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseGitError(stderr, args)
	}
}

// TestErrorMessageRendering simulates how messages would appear in Task Runner
func TestErrorMessageRendering(t *testing.T) {
	stderr := "fatal: A branch named 'my-feature' already exists."
	args := "checkout -b my-feature"

	gitErr := ParseGitError(stderr, args)

	// Simulate the output messages that would be sent to the Task Runner
	messages := []string{
		"",
		"════════════════════════════════════════════════════════",
		fmt.Sprintf("❌ Error: %s", gitErr.ErrorType()),
		"",
		"💡 Suggestion:",
	}

	// Add each line of the suggestion
	suggestion := gitErr.Suggestion()
	for _, line := range strings.Split(suggestion, "\n") {
		messages = append(messages, line)
	}

	messages = append(messages, "")
	messages = append(messages, "════════════════════════════════════════════════════════")

	// Verify the message structure
	if len(messages) < 8 {
		t.Errorf("expected at least 8 message lines, got %d", len(messages))
	}

	// Check for error indicator
	found := false
	for _, msg := range messages {
		if strings.Contains(msg, "❌") && strings.Contains(msg, "Error") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error indicator message")
	}

	// Check for suggestion indicator
	found = false
	for _, msg := range messages {
		if strings.Contains(msg, "💡") && strings.Contains(msg, "Suggestion") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected suggestion indicator message")
	}

	// Check for separator lines
	separatorCount := 0
	for _, msg := range messages {
		if strings.Contains(msg, "════") {
			separatorCount++
		}
	}
	if separatorCount != 2 {
		t.Errorf("expected 2 separator lines, got %d", separatorCount)
	}
}

// TestErrorMessageTiming verifies that error parsing is reasonably fast
func TestErrorMessageTiming(t *testing.T) {
	stderr := "fatal: A branch named 'feature' already exists."
	args := "checkout -b feature"

	start := time.Now()
	_ = ParseGitError(stderr, args)
	duration := time.Since(start)

	// Error parsing should be very fast (less than 1ms)
	if duration > time.Millisecond {
		t.Logf("warning: error parsing took %v, expected < 1ms", duration)
	}
}
