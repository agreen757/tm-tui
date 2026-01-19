package dialog

import (
	"strings"
	"testing"
)

// TestBranchExistsError tests BranchExistsError functionality
func TestBranchExistsError(t *testing.T) {
	err := &BranchExistsError{
		BranchName: "feature-branch",
		Message:    "error: pathspec 'feature-branch' already exists.",
	}

	if err.ErrorType() != "BranchExistsError" {
		t.Errorf("expected ErrorType to be BranchExistsError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for BranchExistsError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, "feature-branch") {
		t.Errorf("expected suggestion to contain branch name")
	}
}

// TestBranchNotFoundError tests BranchNotFoundError functionality
func TestBranchNotFoundError(t *testing.T) {
	err := &BranchNotFoundError{
		BranchName: "nonexistent",
		Message:    "error: pathspec 'nonexistent' did not match any files",
	}

	if err.ErrorType() != "BranchNotFoundError" {
		t.Errorf("expected ErrorType to be BranchNotFoundError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for BranchNotFoundError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, "nonexistent") {
		t.Errorf("expected suggestion to contain branch name")
	}
}

// TestGitPermissionError tests GitPermissionError functionality
func TestGitPermissionError(t *testing.T) {
	err := &GitPermissionError{
		Message:   "permission denied (.git/refs/heads/main)",
		Resource:  ".git/refs/heads/main",
		Operation: "checkout",
	}

	if err.ErrorType() != "GitPermissionError" {
		t.Errorf("expected ErrorType to be GitPermissionError, got %s", err.ErrorType())
	}

	if err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be false for GitPermissionError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, ".git/refs/heads/main") {
		t.Errorf("expected suggestion to contain resource path")
	}
}

// TestGitNetworkError tests GitNetworkError functionality
func TestGitNetworkError(t *testing.T) {
	err := &GitNetworkError{
		Message:    "fatal: Unable to look up github.com (port 9418) (Name or service not known)",
		RemoteName: "origin (GitHub)",
	}

	if err.ErrorType() != "GitNetworkError" {
		t.Errorf("expected ErrorType to be GitNetworkError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for GitNetworkError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, "Network error") {
		t.Errorf("expected suggestion to mention network error")
	}
}

// TestDetachedHeadError tests DetachedHeadError functionality
func TestDetachedHeadError(t *testing.T) {
	err := &DetachedHeadError{
		Message:    "You are in 'detached HEAD' state.",
		CurrentSHA: "abc123def",
	}

	if err.ErrorType() != "DetachedHeadError" {
		t.Errorf("expected ErrorType to be DetachedHeadError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for DetachedHeadError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, "abc123def") {
		t.Errorf("expected suggestion to contain SHA")
	}
}

// TestMergeConflictError tests MergeConflictError functionality
func TestMergeConflictError(t *testing.T) {
	files := []string{"file1.go", "file2.go"}
	err := &MergeConflictError{
		Message:       "CONFLICT (content): Merge conflict in file1.go",
		ConflictFiles: files,
	}

	if err.ErrorType() != "MergeConflictError" {
		t.Errorf("expected ErrorType to be MergeConflictError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for MergeConflictError")
	}

	suggestion := err.Suggestion()
	for _, f := range files {
		if !strings.Contains(suggestion, f) {
			t.Errorf("expected suggestion to contain file %s", f)
		}
	}
}

// TestUncommittedChangesError tests UncommittedChangesError functionality
func TestUncommittedChangesError(t *testing.T) {
	files := []string{"README.md", "main.go"}
	err := &UncommittedChangesError{
		Message:      "Your local changes to 'README.md' would be overwritten by merge.",
		ChangedFiles: files,
	}

	if err.ErrorType() != "UncommittedChangesError" {
		t.Errorf("expected ErrorType to be UncommittedChangesError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for UncommittedChangesError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, "uncommitted") {
		t.Errorf("expected suggestion to mention uncommitted changes")
	}
}

// TestAheadBehindError tests AheadBehindError functionality
func TestAheadBehindError(t *testing.T) {
	err := &AheadBehindError{
		Message:    "Your branch is ahead of 'origin/main' by 3 commits.",
		AheadCount: 3,
		BranchName: "main",
		RemoteName: "origin",
	}

	if err.ErrorType() != "AheadBehindError" {
		t.Errorf("expected ErrorType to be AheadBehindError, got %s", err.ErrorType())
	}

	if !err.IsRecoverable() {
		t.Errorf("expected IsRecoverable to be true for AheadBehindError")
	}

	suggestion := err.Suggestion()
	if !strings.Contains(suggestion, "Push") {
		t.Errorf("expected suggestion to mention push for ahead case")
	}
}

// TestParseGitErrorBranchExists tests parsing branch exists error
func TestParseGitErrorBranchExists(t *testing.T) {
	stderr := "fatal: A branch named 'feature' already exists."
	err := ParseGitError(stderr, "checkout -b feature")

	if _, ok := err.(*BranchExistsError); !ok {
		t.Errorf("expected BranchExistsError, got %T", err)
	}
}

// TestParseGitErrorBranchNotFound tests parsing branch not found error
func TestParseGitErrorBranchNotFound(t *testing.T) {
	stderr := "error: pathspec 'nonexistent-branch' did not match any files in the index."
	err := ParseGitError(stderr, "checkout nonexistent-branch")

	if _, ok := err.(*BranchNotFoundError); !ok {
		t.Errorf("expected BranchNotFoundError, got %T", err)
	}
}

// TestParseGitErrorPermissionDenied tests parsing permission denied error
func TestParseGitErrorPermissionDenied(t *testing.T) {
	stderr := "fatal: Permission denied (.git/config)"
	err := ParseGitError(stderr, "checkout main")

	if _, ok := err.(*GitPermissionError); !ok {
		t.Errorf("expected GitPermissionError, got %T", err)
	}
}

// TestParseGitErrorNetworkConnectionRefused tests parsing connection refused error
func TestParseGitErrorNetworkConnectionRefused(t *testing.T) {
	stderr := "fatal: unable to access 'https://github.com/user/repo.git/': Connection refused"
	err := ParseGitError(stderr, "fetch")

	if _, ok := err.(*GitNetworkError); !ok {
		t.Errorf("expected GitNetworkError, got %T", err)
	}
}

// TestParseGitErrorNetworkTimeout tests parsing connection timeout error
func TestParseGitErrorNetworkTimeout(t *testing.T) {
	stderr := "fatal: operation timed out after 30 seconds: Connection timed out"
	err := ParseGitError(stderr, "push")

	if _, ok := err.(*GitNetworkError); !ok {
		t.Errorf("expected GitNetworkError, got %T", err)
	}
}

// TestParseGitErrorDetachedHead tests parsing detached HEAD error
func TestParseGitErrorDetachedHead(t *testing.T) {
	stderr := "You are in 'detached HEAD' state. You can look around, make experimental changes and commit them, and you can discard any commits you make in this state without impacting any branches by performing another checkout."
	err := ParseGitError(stderr, "checkout abc123def")

	if _, ok := err.(*DetachedHeadError); !ok {
		t.Errorf("expected DetachedHeadError, got %T", err)
	}
}

// TestParseGitErrorMergeConflict tests parsing merge conflict error
func TestParseGitErrorMergeConflict(t *testing.T) {
	stderr := "CONFLICT (content): Merge conflict in file1.go\nAuto-merging file1.go"
	err := ParseGitError(stderr, "merge feature")

	if _, ok := err.(*MergeConflictError); !ok {
		t.Errorf("expected MergeConflictError, got %T", err)
	}
}

// TestParseGitErrorUncommittedChanges tests parsing uncommitted changes error
func TestParseGitErrorUncommittedChanges(t *testing.T) {
	stderr := "error: Your local changes to 'main.go' would be overwritten by merge.\nPlease commit your changes or stash them before you merge."
	err := ParseGitError(stderr, "merge develop")

	if _, ok := err.(*UncommittedChangesError); !ok {
		t.Errorf("expected UncommittedChangesError, got %T", err)
	}
}

// TestParseGitErrorAheadBehind tests parsing ahead/behind error
func TestParseGitErrorAheadBehind(t *testing.T) {
	stderr := "Your branch is ahead of 'origin/main' by 3 commits."
	err := ParseGitError(stderr, "status")

	if _, ok := err.(*AheadBehindError); !ok {
		t.Errorf("expected AheadBehindError, got %T", err)
	}

	aheadErr := err.(*AheadBehindError)
	if aheadErr.AheadCount != 3 {
		t.Errorf("expected AheadCount to be 3, got %d", aheadErr.AheadCount)
	}
}

// TestParseGitErrorGeneric tests parsing unknown git error
func TestParseGitErrorGeneric(t *testing.T) {
	stderr := "Some unknown error that git produced"
	err := ParseGitError(stderr, "unknown-command")

	if _, ok := err.(*GenericGitError); !ok {
		t.Errorf("expected GenericGitError, got %T", err)
	}
}

// TestParseGitErrorEmptyStderr tests parsing empty stderr
func TestParseGitErrorEmptyStderr(t *testing.T) {
	stderr := ""
	err := ParseGitError(stderr, "status")

	if _, ok := err.(*GenericGitError); !ok {
		t.Errorf("expected GenericGitError for empty stderr, got %T", err)
	}
}

// TestExtractBranchName tests branch name extraction from commands
func TestExtractBranchName(t *testing.T) {
	tests := []struct {
		command      string
		expectedName string
	}{
		{"checkout -b my-feature", "my-feature"},
		{"checkout -b feature/PROJ-123", "feature/PROJ-123"},
		{"checkout main", "main"},
		{"unknown command", "unknown"},
	}

	for _, test := range tests {
		result := extractBranchName(test.command)
		if result != test.expectedName {
			t.Errorf("for command '%s', expected '%s' but got '%s'", test.command, test.expectedName, result)
		}
	}
}

// TestExtractRemoteName tests remote name extraction from error messages
func TestExtractRemoteName(t *testing.T) {
	tests := []struct {
		message      string
		expectedName string
	}{
		{"fatal: unable to access 'https://github.com/user/repo.git/'", "origin (GitHub)"},
		{"Connection refused to git@gitlab.com:user/repo.git", "origin (GitLab)"},
		{"fatal: could not read from remote", "origin"},
	}

	for _, test := range tests {
		result := extractRemoteName(test.message)
		if result != test.expectedName {
			t.Errorf("for message '%s', expected '%s' but got '%s'", test.message, test.expectedName, result)
		}
	}
}

// TestGitErrorInterface tests that all error types implement GitError interface
func TestGitErrorInterface(t *testing.T) {
	errors := []GitError{
		&BranchExistsError{BranchName: "test", Message: "test"},
		&BranchNotFoundError{BranchName: "test", Message: "test"},
		&GitPermissionError{Message: "test", Resource: "test", Operation: "test"},
		&GitNetworkError{Message: "test", RemoteName: "test"},
		&DetachedHeadError{Message: "test", CurrentSHA: "test"},
		&MergeConflictError{Message: "test", ConflictFiles: []string{}},
		&UncommittedChangesError{Message: "test", ChangedFiles: []string{}},
		&AheadBehindError{Message: "test"},
		&GenericGitError{Message: "test", RawOutput: "test"},
	}

	for _, err := range errors {
		// Test that all methods exist and don't panic
		if err.Error() == "" {
			t.Errorf("%T.Error() returned empty string", err)
		}
		if err.ErrorType() == "" {
			t.Errorf("%T.ErrorType() returned empty string", err)
		}
		if err.Suggestion() == "" {
			t.Errorf("%T.Suggestion() returned empty string", err)
		}
		// IsRecoverable() can be true or false, just test it exists
		_ = err.IsRecoverable()
	}
}

// TestGetRecoveryOptions verifies recovery options are provided
func TestGetRecoveryOptions(t *testing.T) {
	tests := []struct {
		name          string
		err           GitError
		expectOptions int
	}{
		{
			name: "BranchExistsError",
			err: &BranchExistsError{
				BranchName: "test",
				Message:    "exists",
			},
			expectOptions: 2,
		},
		{
			name: "BranchNotFoundError",
			err: &BranchNotFoundError{
				BranchName: "test",
				Message:    "not found",
			},
			expectOptions: 2,
		},
		{
			name: "GitPermissionError",
			err: &GitPermissionError{
				Message:   "permission denied",
				Resource:  ".git",
				Operation: "checkout",
			},
			expectOptions: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := test.err.GetRecoveryOptions()
			if len(opts) != test.expectOptions {
				t.Errorf("expected %d recovery options, got %d", test.expectOptions, len(opts))
			}
		})
	}
}
