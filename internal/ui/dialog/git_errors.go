package dialog

import (
	"fmt"
	"regexp"
	"strings"
)

// RecoveryOption represents a suggested recovery action for a git error
type RecoveryOption struct {
	Label       string // Display label for the user (e.g., "Force Create Branch")
	Description string // Detailed description of what this option does
	Command     []string // The git command arguments to execute for recovery
	IsDestructive bool // True if this option may lose data (e.g., --force)
}

// GitError is the base interface for all git-specific errors
type GitError interface {
	error
	ErrorType() string
	Suggestion() string
	IsRecoverable() bool
	GetRecoveryOptions() []RecoveryOption
}

// BranchExistsError represents when a branch already exists
type BranchExistsError struct {
	BranchName string
	Message    string
}

func (e *BranchExistsError) Error() string {
	return e.Message
}

func (e *BranchExistsError) ErrorType() string {
	return "BranchExistsError"
}

func (e *BranchExistsError) Suggestion() string {
	return fmt.Sprintf("Branch '%s' already exists. You can:\n• Checkout the existing branch: git checkout %s\n• Force create to overwrite: git checkout -b %s --force", e.BranchName, e.BranchName, e.BranchName)
}

func (e *BranchExistsError) IsRecoverable() bool {
	return true
}

func (e *BranchExistsError) GetRecoveryOptions() []RecoveryOption {
	return []RecoveryOption{
		{
			Label:       "Checkout Existing Branch",
			Description: "Switch to the existing branch instead of creating a new one",
			Command:     []string{"checkout", e.BranchName},
			IsDestructive: false,
		},
		{
			Label:       "Force Create Branch",
			Description: "Create a new branch with the same name, overwriting the existing one",
			Command:     []string{"checkout", "-b", e.BranchName, "--force"},
			IsDestructive: true,
		},
	}
}

// BranchNotFoundError represents when a branch cannot be found
type BranchNotFoundError struct {
	BranchName string
	Message    string
}

func (e *BranchNotFoundError) Error() string {
	return e.Message
}

func (e *BranchNotFoundError) ErrorType() string {
	return "BranchNotFoundError"
}

func (e *BranchNotFoundError) Suggestion() string {
	return fmt.Sprintf("Branch '%s' not found. Try:\n• Fetch latest branches: git fetch\n• List all branches: git branch -a\n• Create a new branch: git checkout -b %s", e.BranchName, e.BranchName)
}

func (e *BranchNotFoundError) IsRecoverable() bool {
	return true
}

func (e *BranchNotFoundError) GetRecoveryOptions() []RecoveryOption {
	return []RecoveryOption{
		{
			Label:       "Fetch Latest Branches",
			Description: "Fetch the latest branches from the remote repository",
			Command:     []string{"fetch"},
			IsDestructive: false,
		},
		{
			Label:       "Create Local Branch",
			Description: fmt.Sprintf("Create a new local branch named '%s'", e.BranchName),
			Command:     []string{"checkout", "-b", e.BranchName},
			IsDestructive: false,
		},
	}
}

// GitPermissionError represents permission-related git errors
type GitPermissionError struct {
	Message    string
	Resource   string
	Operation  string
}

func (e *GitPermissionError) Error() string {
	return e.Message
}

func (e *GitPermissionError) ErrorType() string {
	return "GitPermissionError"
}

func (e *GitPermissionError) Suggestion() string {
	return fmt.Sprintf("Permission denied accessing '%s'.\n• Check file permissions: ls -la %s\n• Verify SSH key is configured (for remote operations)\n• Ensure you have write access to the repository", e.Resource, e.Resource)
}

func (e *GitPermissionError) IsRecoverable() bool {
	return false
}

func (e *GitPermissionError) GetRecoveryOptions() []RecoveryOption {
	// Permission errors are not recoverable through git commands
	return []RecoveryOption{}
}

// GitNetworkError represents network-related git errors
type GitNetworkError struct {
	Message    string
	RemoteName string
}

func (e *GitNetworkError) Error() string {
	return e.Message
}

func (e *GitNetworkError) ErrorType() string {
	return "GitNetworkError"
}

func (e *GitNetworkError) Suggestion() string {
	suggestion := "Network error during git operation.\n• Check internet connection\n• Verify remote URL is correct: git remote -v\n• Try again in a moment"
	if e.RemoteName != "" {
		suggestion += fmt.Sprintf("\n• Remote: %s", e.RemoteName)
	}
	return suggestion
}

func (e *GitNetworkError) IsRecoverable() bool {
	return true
}

func (e *GitNetworkError) GetRecoveryOptions() []RecoveryOption {
	return []RecoveryOption{
		{
			Label:       "Retry Network Operation",
			Description: "Attempt the same operation again after checking network connection",
			Command:     []string{}, // Will be determined by caller
			IsDestructive: false,
		},
		{
			Label:       "Check Repository",
			Description: "Verify the remote repository URL is correct",
			Command:     []string{"remote", "-v"},
			IsDestructive: false,
		},
	}
}

// DetachedHeadError represents when HEAD is detached
type DetachedHeadError struct {
	Message    string
	CurrentSHA string
}

func (e *DetachedHeadError) Error() string {
	return e.Message
}

func (e *DetachedHeadError) ErrorType() string {
	return "DetachedHeadError"
}

func (e *DetachedHeadError) Suggestion() string {
	return fmt.Sprintf("You are in detached HEAD state at %s.\n• Checkout a branch: git checkout <branch>\n• Create a new branch from this state: git checkout -b <new-branch>\n• View the commit: git log -1", e.CurrentSHA)
}

func (e *DetachedHeadError) IsRecoverable() bool {
	return true
}

func (e *DetachedHeadError) GetRecoveryOptions() []RecoveryOption {
	return []RecoveryOption{
		{
			Label:       "Checkout a Branch",
			Description: "Attach HEAD by checking out an existing branch",
			Command:     []string{"checkout", e.CurrentSHA[:7]}, // Show short SHA for context
			IsDestructive: false,
		},
		{
			Label:       "Create New Branch",
			Description: "Create and checkout a new branch from the current commit",
			Command:     []string{},  // Requires user input for branch name
			IsDestructive: false,
		},
	}
}

// MergeConflictError represents when there are merge conflicts
type MergeConflictError struct {
	Message       string
	ConflictFiles []string
}

func (e *MergeConflictError) Error() string {
	return e.Message
}

func (e *MergeConflictError) ErrorType() string {
	return "MergeConflictError"
}

func (e *MergeConflictError) Suggestion() string {
	suggestion := "Merge conflict detected.\n• Resolve conflicts in the affected files\n• Stage resolved files: git add <file>\n• Complete merge: git commit\n\nConflicted files:"
	for _, file := range e.ConflictFiles {
		suggestion += "\n  • " + file
	}
	if len(e.ConflictFiles) == 0 {
		suggestion = "Merge conflict detected.\n• Use your editor to resolve conflicts (look for <<<<<<, ======, >>>>>>)\n• Stage resolved files: git add <file>\n• Complete merge: git commit"
	}
	return suggestion
}

func (e *MergeConflictError) IsRecoverable() bool {
	return true
}

func (e *MergeConflictError) GetRecoveryOptions() []RecoveryOption {
	return []RecoveryOption{
		{
			Label:       "Abort Merge",
			Description: "Abort the current merge operation and return to the previous state",
			Command:     []string{"merge", "--abort"},
			IsDestructive: false,
		},
		{
			Label:       "Resolve Conflicts",
			Description: "After manually resolving conflicts in the editor, commit the merge",
			Command:     []string{"add", "."}, // Stage all resolved files
			IsDestructive: false,
		},
	}
}

// UncommittedChangesError represents when there are uncommitted changes blocking an operation
type UncommittedChangesError struct {
	Message    string
	ChangedFiles []string
}

func (e *UncommittedChangesError) Error() string {
	return e.Message
}

func (e *UncommittedChangesError) ErrorType() string {
	return "UncommittedChangesError"
}

func (e *UncommittedChangesError) Suggestion() string {
	suggestion := "You have uncommitted changes.\n• Commit your changes: git commit -m \"message\"\n• Stash for later: git stash\n• Discard changes: git checkout -- <file>\n\nModified files:"
	for _, file := range e.ChangedFiles {
		suggestion += "\n  • " + file
	}
	if len(e.ChangedFiles) == 0 {
		suggestion = "You have uncommitted changes.\n• Commit your changes: git commit -m \"message\"\n• Stash for later: git stash\n• Discard changes: git checkout -- <file>\n• View status: git status"
	}
	return suggestion
}

func (e *UncommittedChangesError) IsRecoverable() bool {
	return true
}

func (e *UncommittedChangesError) GetRecoveryOptions() []RecoveryOption {
	return []RecoveryOption{
		{
			Label:       "Commit Changes",
			Description: "Stage and commit all outstanding changes",
			Command:     []string{"add", "."},
			IsDestructive: false,
		},
		{
			Label:       "Stash Changes",
			Description: "Save changes to stash and continue with a clean working directory",
			Command:     []string{"stash"},
			IsDestructive: false,
		},
		{
			Label:       "Discard Changes",
			Description: "Discard all uncommitted changes (cannot be undone)",
			Command:     []string{"checkout", "--", "."},
			IsDestructive: true,
		},
	}
}

// AheadBehindError represents when the branch is ahead/behind the remote
type AheadBehindError struct {
	Message    string
	AheadCount int
	BehindCount int
	BranchName string
	RemoteName string
}

func (e *AheadBehindError) Error() string {
	return e.Message
}

func (e *AheadBehindError) ErrorType() string {
	return "AheadBehindError"
}

func (e *AheadBehindError) Suggestion() string {
	suggestion := ""
	if e.BehindCount > 0 {
		suggestion = fmt.Sprintf("Your branch is %d commit(s) behind %s/%s.\n• Pull latest changes: git pull\n• Rebase: git rebase", e.BehindCount, e.RemoteName, e.BranchName)
	}
	if e.AheadCount > 0 {
		if suggestion != "" {
			suggestion += "\n\nYour branch is also " + fmt.Sprintf("%d commit(s) ahead", e.AheadCount)
		} else {
			suggestion = fmt.Sprintf("Your branch is %d commit(s) ahead of %s/%s.\n• Push your changes: git push\n• Push with force (careful!): git push --force-with-lease", e.AheadCount, e.RemoteName, e.BranchName)
		}
	}
	if suggestion == "" {
		// Default case when neither ahead nor behind
		suggestion = "Your branch is synchronized with the remote.\n• Pull latest: git pull\n• Push changes: git push\n• Check status: git status"
	}
	return suggestion
}

func (e *AheadBehindError) IsRecoverable() bool {
	return true
}

func (e *AheadBehindError) GetRecoveryOptions() []RecoveryOption {
	options := []RecoveryOption{}
	
	if e.BehindCount > 0 {
		options = append(options, RecoveryOption{
			Label:       "Pull Latest Changes",
			Description: "Fetch and merge the latest changes from the remote",
			Command:     []string{"pull"},
			IsDestructive: false,
		})
	}
	
	if e.AheadCount > 0 {
		options = append(options, RecoveryOption{
			Label:       "Push Changes",
			Description: "Push your local commits to the remote repository",
			Command:     []string{"push"},
			IsDestructive: false,
		})
	}
	
	if len(options) == 0 {
		options = append(options, RecoveryOption{
			Label:       "Sync with Remote",
			Description: "Ensure your branch is synchronized with the remote",
			Command:     []string{"pull"},
			IsDestructive: false,
		})
	}
	
	return options
}

// GenericGitError is a fallback for unknown git errors
type GenericGitError struct {
	Message    string
	RawOutput  string
}

func (e *GenericGitError) Error() string {
	return e.Message
}

func (e *GenericGitError) ErrorType() string {
	return "GenericGitError"
}

func (e *GenericGitError) Suggestion() string {
	return "An unexpected git error occurred.\n• Check git status: git status\n• Review the error message above\n• Try running the command manually from the terminal for more details"
}

func (e *GenericGitError) IsRecoverable() bool {
	return false
}

func (e *GenericGitError) GetRecoveryOptions() []RecoveryOption {
	// Unknown errors have no recovery options
	return []RecoveryOption{}
}

// ParseGitError parses git stderr output and returns a typed error
// It uses pattern matching to classify errors into specific types
func ParseGitError(stderr string, command string) GitError {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return &GenericGitError{
			Message:   "Unknown git error",
			RawOutput: stderr,
		}
	}

	// Check for branch-related errors
	if strings.Contains(stderr, "already exists") || strings.Contains(stderr, "refname already exists") {
		// Extract branch name from command if possible
		branchName := extractBranchName(command)
		return &BranchExistsError{
			BranchName: branchName,
			Message:    stderr,
		}
	}

	// Check for branch not found / pathspec did not match
	if strings.Contains(stderr, "pathspec did not match any files") ||
		strings.Contains(stderr, "did not match any files in the index") ||
		strings.Contains(stderr, "No such file or directory") ||
		strings.Contains(stderr, "error: unknown switch") {
		branchName := extractBranchName(command)
		return &BranchNotFoundError{
			BranchName: branchName,
			Message:    stderr,
		}
	}

	// Check for permission errors
	if strings.Contains(stderr, "permission denied") ||
		strings.Contains(stderr, "Permission denied") ||
		strings.Contains(stderr, "access denied") ||
		strings.Contains(stderr, "Access denied") {
		resource := extractResource(stderr)
		return &GitPermissionError{
			Message:   stderr,
			Resource:  resource,
			Operation: extractOperation(command),
		}
	}

	// Check for network errors
	if strings.Contains(stderr, "Connection refused") ||
		strings.Contains(stderr, "Connection timed out") ||
		strings.Contains(stderr, "getaddrinfo failed") ||
		strings.Contains(stderr, "Could not resolve") ||
		strings.Contains(stderr, "failed to resolve") ||
		strings.Contains(stderr, "Name or service not known") ||
		strings.Contains(stderr, "no address associated") ||
		strings.Contains(stderr, "unable to access") && strings.Contains(stderr, "Failed to connect") ||
		strings.Contains(stderr, "Could not read from") && strings.Contains(stderr, "remote") {
		remoteName := extractRemoteName(stderr)
		return &GitNetworkError{
			Message:    stderr,
			RemoteName: remoteName,
		}
	}

	// Check for detached HEAD
	if strings.Contains(stderr, "detached HEAD") ||
		strings.Contains(stderr, "Detached HEAD") ||
		strings.Contains(stderr, "You are currently on a detached HEAD") {
		sha := extractSHA(stderr)
		return &DetachedHeadError{
			Message:    stderr,
			CurrentSHA: sha,
		}
	}

	// Check for merge conflicts
	if strings.Contains(stderr, "CONFLICT") ||
		strings.Contains(stderr, "conflict") ||
		strings.Contains(stderr, "Merge conflict") ||
		strings.Contains(stderr, "Auto-merging") && strings.Contains(stderr, "CONFLICT") {
		files := extractConflictFiles(stderr)
		return &MergeConflictError{
			Message:       stderr,
			ConflictFiles: files,
		}
	}

	// Check for uncommitted changes
	if strings.Contains(stderr, "Your local changes to") ||
		strings.Contains(stderr, "Would be overwritten by merge") ||
		strings.Contains(stderr, "uncommitted changes") ||
		strings.Contains(stderr, "Following untracked working tree files would be overwritten") ||
		strings.Contains(stderr, "Please commit your changes or stash them") {
		files := extractChangedFiles(stderr)
		return &UncommittedChangesError{
			Message:      stderr,
			ChangedFiles: files,
		}
	}

	// Check for ahead/behind
	if strings.Contains(stderr, "Your branch is ahead") ||
		strings.Contains(stderr, "Your branch is behind") ||
		strings.Contains(stderr, "Your branch and") && strings.Contains(stderr, "have diverged") {
		ahead, behind, branch, remote := extractAheadBehind(stderr)
		return &AheadBehindError{
			Message:     stderr,
			AheadCount:  ahead,
			BehindCount: behind,
			BranchName:  branch,
			RemoteName:  remote,
		}
	}

	// Default to generic error
	return &GenericGitError{
		Message:   stderr,
		RawOutput: stderr,
	}
}

// extractBranchName tries to extract the branch name from the command or error message
func extractBranchName(command string) string {
	// Try to extract from command like "checkout -b branch-name"
	parts := strings.Fields(command)
	for i, part := range parts {
		if part == "-b" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	
	// If no -b flag, try to extract from "checkout branch-name" pattern
	// Look for the branch name after checkout command
	if len(parts) > 1 && (parts[0] == "checkout" || parts[0] == "switch") {
		// Get the first non-flag argument after checkout/switch
		for i := 1; i < len(parts); i++ {
			if !strings.HasPrefix(parts[i], "-") {
				return parts[i]
			}
		}
	}
	
	return "unknown"
}

// extractResource extracts the resource path from a permission error
func extractResource(message string) string {
	// Try to extract path/file from the error message
	re := regexp.MustCompile(`'([^']+)'`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		return matches[1]
	}
	return "repository"
}

// extractOperation extracts the operation from the git command
func extractOperation(command string) string {
	parts := strings.Fields(command)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// extractRemoteName tries to extract the remote name from error message
func extractRemoteName(message string) string {
	// Look for patterns like "git@github.com:..." or "https://..."
	if strings.Contains(message, "github.com") {
		return "origin (GitHub)"
	}
	if strings.Contains(message, "gitlab") {
		return "origin (GitLab)"
	}
	if strings.Contains(message, "bitbucket") {
		return "origin (Bitbucket)"
	}
	return "origin"
}

// extractSHA tries to extract the current SHA from detached HEAD error
func extractSHA(message string) string {
	// Look for SHA pattern (40 hex chars or 7+ hex chars)
	re := regexp.MustCompile(`([0-9a-f]{7,40})`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

// extractConflictFiles extracts the list of conflicting files
func extractConflictFiles(message string) []string {
	var files []string
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		// Look for lines starting with "CONFLICT" or "<<<<<<" patterns
		if strings.Contains(line, "CONFLICT") {
			// Extract filename after "CONFLICT (content):"
			re := regexp.MustCompile(`CONFLICT.*:\s+(.+)$`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				files = append(files, strings.TrimSpace(matches[1]))
			}
		}
	}
	return files
}

// extractChangedFiles extracts the list of changed files
func extractChangedFiles(message string) []string {
	var files []string
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		// Look for indented filenames (typical git error format)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ".") ||
			strings.HasPrefix(trimmed, "/") ||
			strings.HasPrefix(trimmed, "src/") ||
			strings.HasPrefix(trimmed, "pkg/") {
			files = append(files, trimmed)
		}
	}
	return files
}

// extractAheadBehind extracts ahead/behind counts from status message
func extractAheadBehind(message string) (ahead, behind int, branch, remote string) {
	// Pattern: "Your branch is ahead of 'origin/main' by 3 commits."
	// or: "Your branch is behind 'origin/main' by 2 commits."
	aheadRe := regexp.MustCompile(`behind '[^']+' by (\d+)`)
	behindRe := regexp.MustCompile(`ahead of '[^']+' by (\d+)`)
	branchRe := regexp.MustCompile(`'([^']+)'`)

	if aheadMatches := aheadRe.FindStringSubmatch(message); len(aheadMatches) > 1 {
		fmt.Sscanf(aheadMatches[1], "%d", &behind)
	}

	if behindMatches := behindRe.FindStringSubmatch(message); len(behindMatches) > 1 {
		fmt.Sscanf(behindMatches[1], "%d", &ahead)
	}

	branchMatches := branchRe.FindAllStringSubmatch(message, -1)
	if len(branchMatches) > 0 {
		fullRef := branchMatches[0][1] // First match, usually "origin/branch"
		parts := strings.Split(fullRef, "/")
		if len(parts) > 1 {
			remote = parts[0]
			branch = strings.Join(parts[1:], "/")
		} else {
			branch = fullRef
			remote = "origin"
		}
	}

	return ahead, behind, branch, remote
}
