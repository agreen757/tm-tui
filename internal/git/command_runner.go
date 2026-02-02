package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CommandRunner defines the interface for executing git commands
type CommandRunner interface {
	// Run executes a git command with the given arguments
	// Returns the combined stdout and stderr output and any error
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// DefaultCommandRunner is the default implementation that executes actual git commands
type DefaultCommandRunner struct {
	repoPath string
}

// NewDefaultCommandRunner creates a new DefaultCommandRunner for the specified repository path
func NewDefaultCommandRunner(repoPath string) *DefaultCommandRunner {
	return &DefaultCommandRunner{
		repoPath: repoPath,
	}
}

// Run executes a git command with the provided arguments
func (r *DefaultCommandRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.repoPath
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		// Include stderr in error message for better debugging
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("git command failed: %w - %s", err, stderr.String())
		}
		return nil, fmt.Errorf("git command failed: %w", err)
	}
	
	return stdout.Bytes(), nil
}

// GitService provides high-level git operations using a CommandRunner
type GitService struct {
	repoPath  string
	cmdRunner CommandRunner
}

// NewGitService creates a new GitService for the specified repository path
func NewGitService(repoPath string) *GitService {
	return &GitService{
		repoPath:  repoPath,
		cmdRunner: NewDefaultCommandRunner(repoPath),
	}
}

// NewGitServiceWithRunner creates a new GitService with a custom CommandRunner (useful for testing)
func NewGitServiceWithRunner(repoPath string, runner CommandRunner) *GitService {
	return &GitService{
		repoPath:  repoPath,
		cmdRunner: runner,
	}
}

// ChangeType represents the type of file change
type ChangeType string

const (
	ChangeTypeAdded      ChangeType = "added"
	ChangeTypeModified   ChangeType = "modified"
	ChangeTypeDeleted    ChangeType = "deleted"
	ChangeTypeRenamed    ChangeType = "renamed"
	ChangeTypeCopied     ChangeType = "copied"
	ChangeTypeUntracked  ChangeType = "untracked"
	ChangeTypeIgnored    ChangeType = "ignored"
	ChangeTypeUnmerged   ChangeType = "unmerged"
	ChangeTypeTypeChange ChangeType = "typechange" // File <-> Directory conversion
)

// FileChange represents a file change in the git repository
type FileChange struct {
	Path       string     // File path relative to repository root
	OldPath    string     // Original path (for renames/copies)
	Status     string     // Raw Git status code (e.g., " M", "R ", "MM")
	ChangeType ChangeType // Parsed change type
	IsStaged   bool       // Whether change is staged (index status)
	IsModified bool       // Whether file is modified in working tree
}

// GetUncommittedChanges returns a list of uncommitted file changes
func (g *GitService) GetUncommittedChanges(ctx context.Context) ([]FileChange, error) {
	output, err := g.cmdRunner.Run(ctx, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to get uncommitted changes: %w", err)
	}
	
	// Don't trim the output before splitting - porcelain format uses specific spacing
	outputStr := string(output)
	if outputStr == "" || strings.TrimSpace(outputStr) == "" {
		return []FileChange{}, nil
	}
	
	lines := strings.Split(strings.TrimRight(outputStr, "\n"), "\n")
	changes := make([]FileChange, 0, len(lines))
	
	for _, line := range lines {
		// Skip empty lines
		if len(line) == 0 {
			continue
		}
		
		// Porcelain format: XY PATH [-> NEWPATH]
		// X = index status, Y = working tree status
		if len(line) < 3 {
			continue
		}
		
		statusCode := line[:2]
		indexStatus := rune(statusCode[0])
		wtStatus := rune(statusCode[1])
		
		// Extract path(s) - handle renamed/copied files
		pathPart := strings.TrimSpace(line[3:])
		var path, oldPath string
		
		// Check for rename/copy (has "->")
		if strings.Contains(pathPart, " -> ") {
			parts := strings.Split(pathPart, " -> ")
			if len(parts) == 2 {
				oldPath = strings.TrimSpace(parts[0])
				path = strings.TrimSpace(parts[1])
			} else {
				path = pathPart
			}
		} else {
			path = pathPart
		}
		
		// Parse status and determine change type
		change := FileChange{
			Path:       path,
			OldPath:    oldPath,
			Status:     statusCode,
			IsStaged:   indexStatus != ' ' && indexStatus != '?',
			IsModified: wtStatus != ' ' && wtStatus != '?',
		}
		
		// Determine change type based on status codes
		switch {
		case indexStatus == 'R' || wtStatus == 'R':
			change.ChangeType = ChangeTypeRenamed
		case indexStatus == 'C' || wtStatus == 'C':
			change.ChangeType = ChangeTypeCopied
		case indexStatus == 'A' || wtStatus == 'A':
			change.ChangeType = ChangeTypeAdded
		case indexStatus == 'D' || wtStatus == 'D':
			change.ChangeType = ChangeTypeDeleted
		case indexStatus == 'M' || wtStatus == 'M':
			change.ChangeType = ChangeTypeModified
		case indexStatus == 'U' || wtStatus == 'U':
			// Unmerged (conflict)
			change.ChangeType = ChangeTypeUnmerged
		case indexStatus == '?' && wtStatus == '?':
			change.ChangeType = ChangeTypeUntracked
		case indexStatus == '!' && wtStatus == '!':
			change.ChangeType = ChangeTypeIgnored
		case indexStatus == 'T' || wtStatus == 'T':
			// Type change (file <-> directory, etc.)
			change.ChangeType = ChangeTypeTypeChange
		default:
			// Unknown or combined status - default to modified
			change.ChangeType = ChangeTypeModified
		}
		
		changes = append(changes, change)
	}
	
	return changes, nil
}

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash    string   // Commit hash
	Message string   // Commit message
	Author  string   // Commit author
	Date    string   // Commit date
	TaskIDs []string // Extracted task IDs from commit message
}

// GetCommitsWithTaskIDs finds commits containing task IDs in their messages
// Returns a map of task ID to list of commit hashes
func (g *GitService) GetCommitsWithTaskIDs(ctx context.Context, limit int) (map[string][]string, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	
	// Get recent commits with full message
	output, err := g.cmdRunner.Run(ctx, "log", fmt.Sprintf("-n%d", limit), "--pretty=format:%H%x09%s")
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	taskMap := make(map[string][]string)
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		
		hash := parts[0]
		message := parts[1]
		
		// Extract task IDs from commit message
		taskIDs := ExtractTaskIDs(message)
		
		// Add this commit to each task ID it references
		for _, taskID := range taskIDs {
			taskMap[taskID] = append(taskMap[taskID], hash)
		}
	}
	
	return taskMap, nil
}

// GetFileContentAtCommit retrieves file content at a specific commit
func (g *GitService) GetFileContentAtCommit(ctx context.Context, commit, file string) (string, error) {
	// Validate inputs
	if commit == "" {
		return "", fmt.Errorf("commit hash cannot be empty")
	}
	if file == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	
	// Use git show to get file content at specific commit
	output, err := g.cmdRunner.Run(ctx, "show", fmt.Sprintf("%s:%s", commit, file))
	if err != nil {
		return "", fmt.Errorf("failed to get file content at commit %s: %w", commit, err)
	}
	
	return string(output), nil
}

// GetFileDiff returns the diff for a specific file between two commits
func (g *GitService) GetFileDiff(ctx context.Context, file, fromCommit, toCommit string) (string, error) {
	// Validate inputs
	if file == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	
	args := []string{"diff"}
	
	// Add commit range if specified
	if fromCommit != "" && toCommit != "" {
		args = append(args, fmt.Sprintf("%s..%s", fromCommit, toCommit))
	} else if fromCommit != "" {
		args = append(args, fromCommit)
	}
	
	args = append(args, "--", file)
	
	output, err := g.cmdRunner.Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff for file %s: %w", file, err)
	}
	
	return string(output), nil
}

// changeCache provides caching for uncommitted changes to improve performance
type changeCache struct {
	changes    []FileChange
	timestamp  time.Time
	ttl        time.Duration
	mutex      sync.RWMutex
}

// newChangeCache creates a new cache with the specified TTL
func newChangeCache(ttl time.Duration) *changeCache {
	return &changeCache{
		ttl: ttl,
	}
}

// get retrieves cached changes if still valid
func (c *changeCache) get() ([]FileChange, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if c.changes == nil {
		return nil, false
	}
	
	if time.Since(c.timestamp) > c.ttl {
		return nil, false
	}
	
	return c.changes, true
}

// set updates the cache with new changes
func (c *changeCache) set(changes []FileChange) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.changes = changes
	c.timestamp = time.Now()
}

// invalidate clears the cache
func (c *changeCache) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.changes = nil
	c.timestamp = time.Time{}
}

// GetUncommittedChangesWithCache retrieves uncommitted changes with caching
// Uses a TTL-based cache to avoid repeated git calls
func (g *GitService) GetUncommittedChangesWithCache(ctx context.Context, cache *changeCache) ([]FileChange, error) {
	// Try to get from cache first
	if cached, ok := cache.get(); ok {
		return cached, nil
	}
	
	// Cache miss - fetch from git
	changes, err := g.GetUncommittedChanges(ctx)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	cache.set(changes)
	
	return changes, nil
}

// CommitFilterOptions provides filtering options for commit history analysis
type CommitFilterOptions struct {
	Limit      int       // Maximum number of commits to analyze (0 = no limit)
	Since      time.Time // Only commits after this time
	Until      time.Time // Only commits before this time
	Author     string    // Filter by author name/email
	Branch     string    // Filter by branch (default: current branch)
	NoMerges   bool      // Exclude merge commits
	MergesOnly bool      // Only include merge commits
}

// TaskCommitMapping provides bidirectional mapping between tasks and commits
type TaskCommitMapping struct {
	TaskToCommits   map[string][]CommitInfo // Task ID -> Commits
	CommitToTasks   map[string][]string     // Commit hash -> Task IDs
	TotalCommits    int                     // Total commits analyzed
	CommitsWithTasks int                    // Commits that reference tasks
	mutex           sync.RWMutex
}

// NewTaskCommitMapping creates a new empty mapping
func NewTaskCommitMapping() *TaskCommitMapping {
	return &TaskCommitMapping{
		TaskToCommits: make(map[string][]CommitInfo),
		CommitToTasks: make(map[string][]string),
	}
}

// AddCommit adds a commit to the mapping
func (m *TaskCommitMapping) AddCommit(commit CommitInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.TotalCommits++
	
	if len(commit.TaskIDs) > 0 {
		m.CommitsWithTasks++
		
		// Add to both directions of the mapping
		for _, taskID := range commit.TaskIDs {
			m.TaskToCommits[taskID] = append(m.TaskToCommits[taskID], commit)
			m.CommitToTasks[commit.Hash] = append(m.CommitToTasks[commit.Hash], taskID)
		}
	}
}

// GetCommitsForTask returns all commits referencing a task ID
func (m *TaskCommitMapping) GetCommitsForTask(taskID string) []CommitInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.TaskToCommits[taskID]
}

// GetTasksForCommit returns all task IDs referenced by a commit
func (m *TaskCommitMapping) GetTasksForCommit(commitHash string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.CommitToTasks[commitHash]
}

// commitHistoryCache caches commit history results
type commitHistoryCache struct {
	mapping    *TaskCommitMapping
	timestamp  time.Time
	ttl        time.Duration
	filterHash string // Hash of filter options for cache invalidation
	mutex      sync.RWMutex
}

// newCommitHistoryCache creates a new cache with the specified TTL
func newCommitHistoryCache(ttl time.Duration) *commitHistoryCache {
	return &commitHistoryCache{
		ttl: ttl,
	}
}

// get retrieves cached mapping if still valid
func (c *commitHistoryCache) get(filterHash string) (*TaskCommitMapping, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if c.mapping == nil || c.filterHash != filterHash {
		return nil, false
	}
	
	if time.Since(c.timestamp) > c.ttl {
		return nil, false
	}
	
	return c.mapping, true
}

// set updates the cache with new mapping
func (c *commitHistoryCache) set(mapping *TaskCommitMapping, filterHash string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.mapping = mapping
	c.filterHash = filterHash
	c.timestamp = time.Now()
}

// invalidate clears the cache
func (c *commitHistoryCache) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.mapping = nil
	c.filterHash = ""
	c.timestamp = time.Time{}
}

// GetCommitsWithTaskIDsFiltered retrieves commits with task IDs using filter options
func (g *GitService) GetCommitsWithTaskIDsFiltered(ctx context.Context, opts CommitFilterOptions) (*TaskCommitMapping, error) {
	// Build git log arguments based on filter options
	args := []string{"log", "--pretty=format:%H%x09%an%x09%ad%x09%s"}
	
	// Add limit
	if opts.Limit > 0 {
		args = append(args, fmt.Sprintf("-n%d", opts.Limit))
	}
	
	// Add time range
	if !opts.Since.IsZero() {
		args = append(args, fmt.Sprintf("--since=%s", opts.Since.Format(time.RFC3339)))
	}
	if !opts.Until.IsZero() {
		args = append(args, fmt.Sprintf("--until=%s", opts.Until.Format(time.RFC3339)))
	}
	
	// Add author filter
	if opts.Author != "" {
		args = append(args, fmt.Sprintf("--author=%s", opts.Author))
	}
	
	// Add merge commit filter
	if opts.NoMerges {
		args = append(args, "--no-merges")
	} else if opts.MergesOnly {
		args = append(args, "--merges")
	}
	
	// Add date format
	args = append(args, "--date=iso")
	
	// Add branch if specified
	if opts.Branch != "" {
		args = append(args, opts.Branch)
	}
	
	// Execute git log
	output, err := g.cmdRunner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered commits: %w", err)
	}
	
	// Parse output and build mapping
	mapping := NewTaskCommitMapping()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Parse line: HASH\tAUTHOR\tDATE\tMESSAGE
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) != 4 {
			continue
		}
		
		commit := CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    parts[2],
			Message: parts[3],
			TaskIDs: ExtractTaskIDs(parts[3]),
		}
		
		mapping.AddCommit(commit)
	}
	
	return mapping, nil
}

// GetCommitsWithTaskIDsFilteredCached retrieves commits with caching
func (g *GitService) GetCommitsWithTaskIDsFilteredCached(ctx context.Context, opts CommitFilterOptions, cache *commitHistoryCache) (*TaskCommitMapping, error) {
	// Create a hash of filter options for cache key
	filterHash := fmt.Sprintf("%d_%s_%s_%s_%v_%v", 
		opts.Limit, 
		opts.Since.Format(time.RFC3339), 
		opts.Until.Format(time.RFC3339),
		opts.Author,
		opts.NoMerges,
		opts.MergesOnly)
	
	// Try to get from cache first
	if cached, ok := cache.get(filterHash); ok {
		return cached, nil
	}
	
	// Cache miss - fetch from git
	mapping, err := g.GetCommitsWithTaskIDsFiltered(ctx, opts)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	cache.set(mapping, filterHash)
	
	return mapping, nil
}

// DiffFormat represents the format for git diff output
type DiffFormat string

const (
	DiffFormatUnified  DiffFormat = "unified"  // Default unified diff
	DiffFormatContext  DiffFormat = "context"  // Context diff
	DiffFormatStat     DiffFormat = "stat"     // Stat summary only
	DiffFormatNameOnly DiffFormat = "nameonly" // File names only
)

// DiffHunk represents a single hunk in a diff
type DiffHunk struct {
	OldStart int      // Start line in old file
	OldCount int      // Number of lines in old file
	NewStart int      // Start line in new file
	NewCount int      // Number of lines in new file
	Lines    []string // Diff lines (with +/- prefixes)
}

// Diff represents a structured git diff
type Diff struct {
	OldPath    string     // Old file path
	NewPath    string     // New file path
	OldMode    string     // Old file mode
	NewMode    string     // New file mode
	IsBinary   bool       // Whether file is binary
	IsNew      bool       // Whether file is new (added)
	IsDeleted  bool       // Whether file is deleted
	IsRenamed  bool       // Whether file is renamed
	Hunks      []DiffHunk // Diff hunks
	Additions  int        // Total additions
	Deletions  int        // Total deletions
	RawDiff    string     // Raw diff output
}

// ParseDiff parses raw git diff output into a structured Diff
func ParseDiff(rawDiff string) (*Diff, error) {
	if strings.TrimSpace(rawDiff) == "" {
		return &Diff{RawDiff: rawDiff}, nil
	}
	
	diff := &Diff{
		RawDiff: rawDiff,
		Hunks:   make([]DiffHunk, 0),
	}
	
	lines := strings.Split(rawDiff, "\n")
	var currentHunk *DiffHunk
	
	for _, line := range lines {
		// Parse diff headers
		if strings.HasPrefix(line, "diff --git") {
			// Extract file paths
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				diff.OldPath = strings.TrimPrefix(parts[2], "a/")
				diff.NewPath = strings.TrimPrefix(parts[3], "b/")
			}
		} else if strings.HasPrefix(line, "old mode") {
			diff.OldMode = strings.TrimPrefix(line, "old mode ")
		} else if strings.HasPrefix(line, "new mode") {
			diff.NewMode = strings.TrimPrefix(line, "new mode ")
		} else if strings.HasPrefix(line, "new file mode") {
			diff.IsNew = true
			diff.NewMode = strings.TrimPrefix(line, "new file mode ")
		} else if strings.HasPrefix(line, "deleted file mode") {
			diff.IsDeleted = true
			diff.OldMode = strings.TrimPrefix(line, "deleted file mode ")
		} else if strings.HasPrefix(line, "rename from") {
			diff.IsRenamed = true
			diff.OldPath = strings.TrimPrefix(line, "rename from ")
		} else if strings.HasPrefix(line, "rename to") {
			diff.NewPath = strings.TrimPrefix(line, "rename to ")
		} else if strings.HasPrefix(line, "Binary files") {
			diff.IsBinary = true
		} else if strings.HasPrefix(line, "@@") {
			// New hunk
			if currentHunk != nil {
				diff.Hunks = append(diff.Hunks, *currentHunk)
			}
			
			// Parse hunk header: @@ -oldStart,oldCount +newStart,newCount @@
			hunk := DiffHunk{Lines: make([]string, 0)}
			fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@", 
				&hunk.OldStart, &hunk.OldCount, 
				&hunk.NewStart, &hunk.NewCount)
			currentHunk = &hunk
		} else if currentHunk != nil {
			// Hunk content
			currentHunk.Lines = append(currentHunk.Lines, line)
			
			// Count additions/deletions
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				diff.Additions++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				diff.Deletions++
			}
		}
	}
	
	// Add final hunk
	if currentHunk != nil {
		diff.Hunks = append(diff.Hunks, *currentHunk)
	}
	
	return diff, nil
}

// GetFileDiffWithFormat returns the diff for a specific file with format options
func (g *GitService) GetFileDiffWithFormat(ctx context.Context, file, fromCommit, toCommit string, format DiffFormat) (string, error) {
	if file == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	
	args := []string{"diff"}
	
	// Add format-specific options
	switch format {
	case DiffFormatContext:
		args = append(args, "--context=3")
	case DiffFormatStat:
		args = append(args, "--stat")
	case DiffFormatNameOnly:
		args = append(args, "--name-only")
	default:
		// Unified is default
	}
	
	// Add commit range
	if fromCommit != "" && toCommit != "" {
		args = append(args, fmt.Sprintf("%s..%s", fromCommit, toCommit))
	} else if fromCommit != "" {
		args = append(args, fromCommit)
	}
	
	args = append(args, "--", file)
	
	output, err := g.cmdRunner.Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff for file %s: %w", file, err)
	}
	
	return string(output), nil
}

// fileContentCache caches file contents at specific commits
type fileContentCache struct {
	contents map[string]string // Key: "commit:path"
	mutex    sync.RWMutex
	maxSize  int // Maximum cache entries
}

// newFileContentCache creates a new cache with the specified max size
func newFileContentCache(maxSize int) *fileContentCache {
	return &fileContentCache{
		contents: make(map[string]string),
		maxSize:  maxSize,
	}
}

// get retrieves cached file content
func (c *fileContentCache) get(commit, path string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	key := fmt.Sprintf("%s:%s", commit, path)
	content, ok := c.contents[key]
	return content, ok
}

// set updates the cache with file content
func (c *fileContentCache) set(commit, path, content string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	key := fmt.Sprintf("%s:%s", commit, path)
	
	// Simple LRU: if cache is full, remove oldest (first) entry
	if len(c.contents) >= c.maxSize {
		for k := range c.contents {
			delete(c.contents, k)
			break
		}
	}
	
	c.contents[key] = content
}

// diffCache caches git diff results
type diffCache struct {
	diffs   map[string]string // Key: "file:from:to:format"
	mutex   sync.RWMutex
	maxSize int
}

// newDiffCache creates a new cache with the specified max size
func newDiffCache(maxSize int) *diffCache {
	return &diffCache{
		diffs:   make(map[string]string),
		maxSize: maxSize,
	}
}

// get retrieves cached diff
func (c *diffCache) get(file, fromCommit, toCommit string, format DiffFormat) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	key := fmt.Sprintf("%s:%s:%s:%s", file, fromCommit, toCommit, format)
	diff, ok := c.diffs[key]
	return diff, ok
}

// set updates the cache with diff
func (c *diffCache) set(file, fromCommit, toCommit string, format DiffFormat, diff string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	key := fmt.Sprintf("%s:%s:%s:%s", file, fromCommit, toCommit, format)
	
	// Simple LRU: if cache is full, remove oldest (first) entry
	if len(c.diffs) >= c.maxSize {
		for k := range c.diffs {
			delete(c.diffs, k)
			break
		}
	}
	
	c.diffs[key] = diff
}

// GetFileContentAtCommitCached retrieves file content with caching
func (g *GitService) GetFileContentAtCommitCached(ctx context.Context, commit, file string, cache *fileContentCache) (string, error) {
	// Try cache first
	if content, ok := cache.get(commit, file); ok {
		return content, nil
	}
	
	// Cache miss - fetch from git
	content, err := g.GetFileContentAtCommit(ctx, commit, file)
	if err != nil {
		return "", err
	}
	
	// Update cache
	cache.set(commit, file, content)
	
	return content, nil
}

// GetFileDiffWithFormatCached retrieves diff with caching
func (g *GitService) GetFileDiffWithFormatCached(ctx context.Context, file, fromCommit, toCommit string, format DiffFormat, cache *diffCache) (string, error) {
	// Try cache first
	if diff, ok := cache.get(file, fromCommit, toCommit, format); ok {
		return diff, nil
	}
	
	// Cache miss - fetch from git
	diff, err := g.GetFileDiffWithFormat(ctx, file, fromCommit, toCommit, format)
	if err != nil {
		return "", err
	}
	
	// Update cache
	cache.set(file, fromCommit, toCommit, format, diff)
	
	return diff, nil
}
