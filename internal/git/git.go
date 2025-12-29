package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GitStatus represents the current state of a git repository
type GitStatus struct {
	Branch      string    `json:"branch"`
	IsDirty     bool      `json:"is_dirty"`
	HasUpstream bool      `json:"has_upstream"`
	Ahead       int       `json:"ahead"`
	Behind      int       `json:"behind"`
	LastUpdated time.Time `json:"last_updated"`
	Error       error     `json:"error"`
}

// RepoInfo stores basic information about a git repository
type RepoInfo struct {
	IsRepo   bool
	RootPath string
	Error    error
}

// StatusRefresher manages the asynchronous refreshing of git status
type StatusRefresher struct {
	repoPath        string
	status          GitStatus
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	refreshInterval time.Duration
}

// NewStatusRefresher creates a new status refresher for the given repository
func NewStatusRefresher(repoPath string) *StatusRefresher {
	ctx, cancel := context.WithCancel(context.Background())
	return &StatusRefresher{
		repoPath:        repoPath,
		ctx:             ctx,
		cancel:          cancel,
		refreshInterval: 5 * time.Second,
	}
}

// DetectRepository checks if the provided path is within a git repository
// and returns the repository root path if found
func DetectRepository(dir string) RepoInfo {
	// Validate input directory
	if dir == "" {
		return RepoInfo{
			IsRepo: false,
			Error:  fmt.Errorf("directory path cannot be empty"),
		}
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return RepoInfo{
			IsRepo: false,
			Error:  fmt.Errorf("failed to access directory: %w", err),
		}
	}

	// Verify it's a directory
	if !info.IsDir() {
		return RepoInfo{
			IsRepo: false,
			Error:  fmt.Errorf("path is not a directory"),
		}
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return RepoInfo{IsRepo: false, Error: err}
	}

	// Validate output is not empty
	rootPath := strings.TrimSpace(string(output))
	if rootPath == "" {
		return RepoInfo{
			IsRepo: false,
			Error:  fmt.Errorf("git command returned empty output"),
		}
	}

	// Normalize the path to ensure cross-platform compatibility
	rootPath = filepath.Clean(rootPath)

	return RepoInfo{IsRepo: true, RootPath: rootPath}
}

// IsGitAvailable checks if git is installed and available in the system PATH
// Returns true if git --version command succeeds, false otherwise
func IsGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

// GetStatus retrieves the current git status information for a repository
func GetStatus(ctx context.Context, repoPath string) (GitStatus, error) {
	status := GitStatus{LastUpdated: time.Now()}

	// Get current branch
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		status.Error = err
		return status, err
	}
	status.Branch = strings.TrimSpace(string(output))

	// Check if working directory is dirty
	cmd = exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err = cmd.Output()
	if err != nil {
		status.Error = err
		return status, err
	}
	status.IsDirty = len(strings.TrimSpace(string(output))) > 0

	// Check ahead/behind counts
	cmd = exec.CommandContext(ctx, "git", "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	cmd.Dir = repoPath
	output, err = cmd.Output()
	if err != nil {
		// No upstream set, not an error
		status.HasUpstream = false
		return status, nil
	}

	status.HasUpstream = true
	counts := strings.Fields(strings.TrimSpace(string(output)))
	if len(counts) == 2 {
		status.Behind, _ = strconv.Atoi(counts[0])
		status.Ahead, _ = strconv.Atoi(counts[1])
	}

	return status, nil
}

// Start begins the periodic refresh of git status
func (r *StatusRefresher) Start() {
	// Initial refresh
	r.Refresh()

	// Read interval under lock to avoid race with SetRefreshInterval
	r.mutex.RLock()
	interval := r.refreshInterval
	r.mutex.RUnlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.Refresh()
			case <-r.ctx.Done():
				return
			}
		}
	}()
}

// Stop halts the periodic refresh
func (r *StatusRefresher) Stop() {
	r.cancel()
}

// SetRefreshInterval updates the refresh interval
func (r *StatusRefresher) SetRefreshInterval(interval time.Duration) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.refreshInterval = interval
}

// GetRefreshInterval returns the current refresh interval
func (r *StatusRefresher) GetRefreshInterval() time.Duration {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.refreshInterval
}

// Refresh immediately updates the git status
func (r *StatusRefresher) Refresh() {
	status, _ := GetStatus(r.ctx, r.repoPath)

	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.status = status
}

// GetStatus returns the current git status
func (r *StatusRefresher) GetStatus() GitStatus {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.status
}

// CreateBranch creates a new branch from the current HEAD
func CreateBranch(ctx context.Context, repoPath, branchName string) (string, error) {
	// Validate branch name
	if strings.Contains(branchName, " ") || branchName == "" {
		return "", fmt.Errorf("invalid branch name: cannot contain spaces and must not be empty")
	}

	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Commit represents a git commit
type Commit struct {
	Hash         string
	Subject      string
	Author       string
	RelativeTime string
}

// GetRecentCommits returns the N most recent commits
func GetRecentCommits(ctx context.Context, repoPath string, count int) ([]Commit, error) {
	if count <= 0 {
		count = 20 // Default to 20 commits
	}

	cmd := exec.CommandContext(
		ctx,
		"git",
		"log",
		"-n",
		strconv.Itoa(count),
		"--pretty=format:%h%x09%ad%x09%an%x09%s",
		"--date=relative",
	)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split by tab character (the delimiter from git format)
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			commits = append(commits, Commit{
				Hash:         parts[0],
				RelativeTime: parts[1],
				Author:       parts[2],
				Subject:      parts[3],
			})
		}
	}

	return commits, nil
}

// GetBranches returns a list of local branches and the current branch
func GetBranches(ctx context.Context, repoPath string) ([]string, string, error) {
	// Get current branch
	currentCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	currentCmd.Dir = repoPath
	currentOutput, err := currentCmd.Output()
	if err != nil {
		return nil, "", err
	}
	currentBranch := strings.TrimSpace(string(currentOutput))

	// Get all branches
	cmd := exec.CommandContext(ctx, "git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, currentBranch, err
	}

	branchesStr := strings.TrimSpace(string(output))
	if branchesStr == "" {
		return []string{}, currentBranch, nil
	}

	branches := strings.Split(branchesStr, "\n")
	return branches, currentBranch, nil
}

// SwitchBranch checks out the specified branch
func SwitchBranch(ctx context.Context, repoPath, branch string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "checkout", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	return string(output), err
}
