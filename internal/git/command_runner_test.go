package git

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// MockCommandRunner is a mock implementation of CommandRunner for testing
type MockCommandRunner struct {
	responses map[string]mockResponse
	callLog   []string
}

type mockResponse struct {
	output []byte
	err    error
}

// NewMockCommandRunner creates a new mock command runner
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		responses: make(map[string]mockResponse),
		callLog:   make([]string, 0),
	}
}

// AddResponse adds a mock response for a specific command
func (m *MockCommandRunner) AddResponse(command string, output []byte, err error) {
	m.responses[command] = mockResponse{output: output, err: err}
}

// Run executes a mocked git command
func (m *MockCommandRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	command := argsToString(args)
	m.callLog = append(m.callLog, command)
	
	if resp, ok := m.responses[command]; ok {
		return resp.output, resp.err
	}
	
	return nil, fmt.Errorf("no mock response for command: %s", command)
}

// GetCallLog returns the log of all commands executed
func (m *MockCommandRunner) GetCallLog() []string {
	return m.callLog
}

// argsToString converts command arguments to a single string for lookup
func argsToString(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func TestNewGitService(t *testing.T) {
	service := NewGitService("/test/repo")
	
	if service == nil {
		t.Fatal("Expected GitService to be created, got nil")
	}
	
	if service.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath to be /test/repo, got %s", service.repoPath)
	}
	
	if service.cmdRunner == nil {
		t.Error("Expected cmdRunner to be initialized, got nil")
	}
}

func TestNewGitServiceWithRunner(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	
	if service == nil {
		t.Fatal("Expected GitService to be created, got nil")
	}
	
	if service.cmdRunner != mockRunner {
		t.Error("Expected custom command runner to be set")
	}
}

func TestGetUncommittedChanges(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedCount  int
		expectedError  bool
		expectedStatus map[string]string // path -> status
	}{
		{
			name: "No changes",
			mockOutput: "",
			mockError: nil,
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "Modified files",
			mockOutput: " M file1.go\n M file2.go\n",
			mockError: nil,
			expectedCount: 2,
			expectedError: false,
			expectedStatus: map[string]string{
				"file1.go": " M",
				"file2.go": " M",
			},
		},
		{
			name: "Mixed changes",
			mockOutput: " M modified.go\nA  added.go\nD  deleted.go\n",
			mockError: nil,
			expectedCount: 3,
			expectedError: false,
			expectedStatus: map[string]string{
				"modified.go": " M",
				"added.go":    "A ",
				"deleted.go":  "D ",
			},
		},
		{
			name: "Git command error",
			mockOutput: "",
			mockError: errors.New("git error"),
			expectedCount: 0,
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			mockRunner.AddResponse("status --porcelain", []byte(tt.mockOutput), tt.mockError)
			
			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			changes, err := service.GetUncommittedChanges(context.Background())
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(changes) != tt.expectedCount {
				t.Errorf("Expected %d changes, got %d", tt.expectedCount, len(changes))
			}
			
			// Verify specific file statuses if provided
			if tt.expectedStatus != nil {
				for _, change := range changes {
					expectedStatus, ok := tt.expectedStatus[change.Path]
					if !ok {
						t.Errorf("Unexpected file in changes: %s", change.Path)
						continue
					}
					if change.Status != expectedStatus {
						t.Errorf("Expected status %s for %s, got %s", expectedStatus, change.Path, change.Status)
					}
				}
			}
		})
	}
}

func TestGetCommitsWithTaskIDs(t *testing.T) {
	tests := []struct {
		name          string
		mockOutput    string
		mockError     error
		limit         int
		expectedTasks map[string]int // taskID -> number of commits
		expectedError bool
	}{
		{
			name:          "No commits",
			mockOutput:    "",
			mockError:     nil,
			limit:         100,
			expectedTasks: map[string]int{},
			expectedError: false,
		},
		{
			name: "Commits with task IDs",
			mockOutput: "abc123\tfix: resolve issue #1.2\ndef456\tfeat: implement feature #2.1\n",
			mockError: nil,
			limit: 100,
			expectedTasks: map[string]int{
				"1.2": 1,
				"2.1": 1,
			},
			expectedError: false,
		},
		{
			name: "Multiple task IDs in one commit",
			mockOutput: "abc123\tfix: resolve issues #1.2 and #1.3\n",
			mockError: nil,
			limit: 100,
			expectedTasks: map[string]int{
				"1.2": 1,
				"1.3": 1,
			},
			expectedError: false,
		},
		{
			name: "Same task ID in multiple commits",
			mockOutput: "abc123\tfix: part 1 of #2.1\ndef456\tfeat: part 2 of #2.1\n",
			mockError: nil,
			limit: 100,
			expectedTasks: map[string]int{
				"2.1": 2,
			},
			expectedError: false,
		},
		{
			name: "Commits without task IDs",
			mockOutput: "abc123\tfix: random fix\ndef456\tfeat: new feature\n",
			mockError: nil,
			limit: 100,
			expectedTasks: map[string]int{},
			expectedError: false,
		},
		{
			name:          "Git command error",
			mockOutput:    "",
			mockError:     errors.New("git error"),
			limit:         100,
			expectedTasks: nil,
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			mockRunner.AddResponse(fmt.Sprintf("log -n%d --pretty=format:%%H%%x09%%s", tt.limit), []byte(tt.mockOutput), tt.mockError)
			
			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			taskMap, err := service.GetCommitsWithTaskIDs(context.Background(), tt.limit)
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(taskMap) != len(tt.expectedTasks) {
				t.Errorf("Expected %d tasks, got %d", len(tt.expectedTasks), len(taskMap))
			}
			
			for taskID, expectedCount := range tt.expectedTasks {
				commits, ok := taskMap[taskID]
				if !ok {
					t.Errorf("Expected task ID %s not found in result", taskID)
					continue
				}
				if len(commits) != expectedCount {
					t.Errorf("Expected %d commits for task %s, got %d", expectedCount, taskID, len(commits))
				}
			}
		})
	}
}

func TestGetFileContentAtCommit(t *testing.T) {
	tests := []struct {
		name          string
		commit        string
		file          string
		mockOutput    string
		mockError     error
		expectedError bool
	}{
		{
			name:          "Valid file content",
			commit:        "abc123",
			file:          "main.go",
			mockOutput:    "package main\n\nfunc main() {}\n",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Empty commit hash",
			commit:        "",
			file:          "main.go",
			mockOutput:    "",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "Empty file path",
			commit:        "abc123",
			file:          "",
			mockOutput:    "",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "Git command error",
			commit:        "abc123",
			file:          "main.go",
			mockOutput:    "",
			mockError:     errors.New("file not found"),
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			if tt.commit != "" && tt.file != "" {
				mockRunner.AddResponse(fmt.Sprintf("show %s:%s", tt.commit, tt.file), []byte(tt.mockOutput), tt.mockError)
			}
			
			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			content, err := service.GetFileContentAtCommit(context.Background(), tt.commit, tt.file)
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if content != tt.mockOutput {
				t.Errorf("Expected content %q, got %q", tt.mockOutput, content)
			}
		})
	}
}

func TestGetFileDiff(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		fromCommit    string
		toCommit      string
		mockOutput    string
		mockError     error
		expectedCmd   string
		expectedError bool
	}{
		{
			name:          "Diff between two commits",
			file:          "main.go",
			fromCommit:    "abc123",
			toCommit:      "def456",
			mockOutput:    "diff --git a/main.go b/main.go\n...",
			mockError:     nil,
			expectedCmd:   "diff abc123..def456 -- main.go",
			expectedError: false,
		},
		{
			name:          "Diff from commit to working tree",
			file:          "main.go",
			fromCommit:    "abc123",
			toCommit:      "",
			mockOutput:    "diff --git a/main.go b/main.go\n...",
			mockError:     nil,
			expectedCmd:   "diff abc123 -- main.go",
			expectedError: false,
		},
		{
			name:          "Diff of working tree",
			file:          "main.go",
			fromCommit:    "",
			toCommit:      "",
			mockOutput:    "diff --git a/main.go b/main.go\n...",
			mockError:     nil,
			expectedCmd:   "diff -- main.go",
			expectedError: false,
		},
		{
			name:          "Empty file path",
			file:          "",
			fromCommit:    "abc123",
			toCommit:      "def456",
			mockOutput:    "",
			mockError:     nil,
			expectedCmd:   "",
			expectedError: true,
		},
		{
			name:          "Git command error",
			file:          "main.go",
			fromCommit:    "abc123",
			toCommit:      "def456",
			mockOutput:    "",
			mockError:     errors.New("file not found"),
			expectedCmd:   "diff abc123..def456 -- main.go",
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			if tt.expectedCmd != "" {
				mockRunner.AddResponse(tt.expectedCmd, []byte(tt.mockOutput), tt.mockError)
			}
			
			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			diff, err := service.GetFileDiff(context.Background(), tt.file, tt.fromCommit, tt.toCommit)
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if diff != tt.mockOutput {
				t.Errorf("Expected diff %q, got %q", tt.mockOutput, diff)
			}
		})
	}
}

func TestDefaultCommandRunner(t *testing.T) {
	// This test verifies the DefaultCommandRunner can be created
	// but we won't test actual git commands to avoid test fragility
	runner := NewDefaultCommandRunner("/test/repo")
	
	if runner == nil {
		t.Fatal("Expected DefaultCommandRunner to be created, got nil")
	}
	
	if runner.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath to be /test/repo, got %s", runner.repoPath)
	}
}

func TestGetUncommittedChangesEnhanced(t *testing.T) {
	tests := []struct {
		name               string
		mockOutput         string
		expectedChanges    []FileChange
	}{
		{
			name:       "Renamed file",
			mockOutput: "R  old-name.go -> new-name.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "new-name.go",
					OldPath:    "old-name.go",
					Status:     "R ",
					ChangeType: ChangeTypeRenamed,
					IsStaged:   true,
					IsModified: false,
				},
			},
		},
		{
			name:       "Copied file",
			mockOutput: "C  original.go -> copy.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "copy.go",
					OldPath:    "original.go",
					Status:     "C ",
					ChangeType: ChangeTypeCopied,
					IsStaged:   true,
					IsModified: false,
				},
			},
		},
		{
			name:       "Untracked file",
			mockOutput: "?? untracked.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "untracked.go",
					Status:     "??",
					ChangeType: ChangeTypeUntracked,
					IsStaged:   false,
					IsModified: false,
				},
			},
		},
		{
			name:       "Staged and modified",
			mockOutput: "MM file.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "file.go",
					Status:     "MM",
					ChangeType: ChangeTypeModified,
					IsStaged:   true,
					IsModified: true,
				},
			},
		},
		{
			name:       "Added and modified",
			mockOutput: "AM file.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "file.go",
					Status:     "AM",
					ChangeType: ChangeTypeAdded,
					IsStaged:   true,
					IsModified: true,
				},
			},
		},
		{
			name:       "Deleted from index",
			mockOutput: "D  file.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "file.go",
					Status:     "D ",
					ChangeType: ChangeTypeDeleted,
					IsStaged:   true,
					IsModified: false,
				},
			},
		},
		{
			name:       "Deleted from working tree",
			mockOutput: " D file.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "file.go",
					Status:     " D",
					ChangeType: ChangeTypeDeleted,
					IsStaged:   false,
					IsModified: true,
				},
			},
		},
		{
			name:       "Unmerged (conflict)",
			mockOutput: "UU conflict.go\n",
			expectedChanges: []FileChange{
				{
					Path:       "conflict.go",
					Status:     "UU",
					ChangeType: ChangeTypeUnmerged,
					IsStaged:   true,
					IsModified: true,
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			mockRunner.AddResponse("status --porcelain", []byte(tt.mockOutput), nil)
			
			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			changes, err := service.GetUncommittedChanges(context.Background())
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if len(changes) != len(tt.expectedChanges) {
				t.Fatalf("Expected %d changes, got %d", len(tt.expectedChanges), len(changes))
			}
			
			for i, expected := range tt.expectedChanges {
				actual := changes[i]
				
				if actual.Path != expected.Path {
					t.Errorf("Change %d: Expected Path %q, got %q", i, expected.Path, actual.Path)
				}
				if actual.OldPath != expected.OldPath {
					t.Errorf("Change %d: Expected OldPath %q, got %q", i, expected.OldPath, actual.OldPath)
				}
				if actual.Status != expected.Status {
					t.Errorf("Change %d: Expected Status %q, got %q", i, expected.Status, actual.Status)
				}
				if actual.ChangeType != expected.ChangeType {
					t.Errorf("Change %d: Expected ChangeType %q, got %q", i, expected.ChangeType, actual.ChangeType)
				}
				if actual.IsStaged != expected.IsStaged {
					t.Errorf("Change %d: Expected IsStaged %v, got %v", i, expected.IsStaged, actual.IsStaged)
				}
				if actual.IsModified != expected.IsModified {
					t.Errorf("Change %d: Expected IsModified %v, got %v", i, expected.IsModified, actual.IsModified)
				}
			}
		})
	}
}

func TestChangeCaching(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("status --porcelain", []byte(" M file1.go\n"), nil)
	
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newChangeCache(1 * time.Second)
	
	// First call - should hit git
	changes1, err := service.GetUncommittedChangesWithCache(context.Background(), cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(changes1) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(changes1))
	}
	
	// Check that git was called once
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected 1 git call, got %d", len(callLog))
	}
	
	// Second call within TTL - should use cache
	changes2, err := service.GetUncommittedChangesWithCache(context.Background(), cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(changes2) != 1 {
		t.Fatalf("Expected 1 change from cache, got %d", len(changes2))
	}
	
	// Should still be just 1 git call (cached)
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected still 1 git call (cached), got %d", len(callLog))
	}
	
	// Wait for TTL to expire
	time.Sleep(1100 * time.Millisecond)
	
	// Third call after TTL - should hit git again
	changes3, err := service.GetUncommittedChangesWithCache(context.Background(), cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(changes3) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(changes3))
	}
	
	// Should now be 2 git calls
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls after TTL, got %d", len(callLog))
	}
}

func TestCacheInvalidation(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("status --porcelain", []byte(" M file1.go\n"), nil)
	
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newChangeCache(10 * time.Second) // Long TTL
	
	// First call
	changes1, err := service.GetUncommittedChangesWithCache(context.Background(), cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(changes1) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(changes1))
	}
	
	// Invalidate cache
	cache.invalidate()
	
	// Second call should hit git again (not cached)
	changes2, err := service.GetUncommittedChangesWithCache(context.Background(), cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(changes2) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(changes2))
	}
	
	// Should be 2 git calls
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls after invalidation, got %d", len(callLog))
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("status --porcelain", []byte(" M file1.go\n"), nil)
	
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newChangeCache(5 * time.Second)
	
	// Concurrent reads and writes
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetUncommittedChangesWithCache(context.Background(), cache)
			if err != nil {
				errors <- err
			}
		}()
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestTaskCommitMapping(t *testing.T) {
	mapping := NewTaskCommitMapping()
	
	// Add commits
	commit1 := CommitInfo{
		Hash:    "abc123",
		Author:  "Test Author",
		Date:    "2024-01-01",
		Message: "Fix issue #1.2",
		TaskIDs: []string{"1.2"},
	}
	commit2 := CommitInfo{
		Hash:    "def456",
		Author:  "Test Author",
		Date:    "2024-01-02",
		Message: "Implement #2.1 and #2.2",
		TaskIDs: []string{"2.1", "2.2"},
	}
	commit3 := CommitInfo{
		Hash:    "ghi789",
		Author:  "Test Author",
		Date:    "2024-01-03",
		Message: "Regular commit without task IDs",
		TaskIDs: []string{},
	}
	
	mapping.AddCommit(commit1)
	mapping.AddCommit(commit2)
	mapping.AddCommit(commit3)
	
	// Test total counts
	if mapping.TotalCommits != 3 {
		t.Errorf("Expected 3 total commits, got %d", mapping.TotalCommits)
	}
	if mapping.CommitsWithTasks != 2 {
		t.Errorf("Expected 2 commits with tasks, got %d", mapping.CommitsWithTasks)
	}
	
	// Test task -> commits mapping
	commits := mapping.GetCommitsForTask("1.2")
	if len(commits) != 1 {
		t.Fatalf("Expected 1 commit for task 1.2, got %d", len(commits))
	}
	if commits[0].Hash != "abc123" {
		t.Errorf("Expected commit abc123, got %s", commits[0].Hash)
	}
	
	commits = mapping.GetCommitsForTask("2.1")
	if len(commits) != 1 {
		t.Fatalf("Expected 1 commit for task 2.1, got %d", len(commits))
	}
	if commits[0].Hash != "def456" {
		t.Errorf("Expected commit def456, got %s", commits[0].Hash)
	}
	
	// Test commit -> tasks mapping
	tasks := mapping.GetTasksForCommit("def456")
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks for commit def456, got %d", len(tasks))
	}
	
	tasks = mapping.GetTasksForCommit("ghi789")
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks for commit ghi789, got %d", len(tasks))
	}
}

func TestGetCommitsWithTaskIDsFiltered(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		opts           CommitFilterOptions
		expectedCommits int
		expectedArgs    string
	}{
		{
			name:       "Basic filter with limit",
			mockOutput: "abc123\tAuthor\t2024-01-01\tFix #1.2\n",
			opts: CommitFilterOptions{
				Limit: 10,
			},
			expectedCommits: 1,
			expectedArgs:    "log --pretty=format:%H%x09%an%x09%ad%x09%s -n10 --date=iso",
		},
		{
			name:       "Filter by author",
			mockOutput: "abc123\tJohn Doe\t2024-01-01\tImplement #2.1\n",
			opts: CommitFilterOptions{
				Limit:  50,
				Author: "John",
			},
			expectedCommits: 1,
			expectedArgs:    "log --pretty=format:%H%x09%an%x09%ad%x09%s -n50 --author=John --date=iso",
		},
		{
			name:       "No merges filter",
			mockOutput: "abc123\tAuthor\t2024-01-01\tFix #1.2\n",
			opts: CommitFilterOptions{
				Limit:    10,
				NoMerges: true,
			},
			expectedCommits: 1,
			expectedArgs:    "log --pretty=format:%H%x09%an%x09%ad%x09%s -n10 --no-merges --date=iso",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			mockRunner.AddResponse(tt.expectedArgs, []byte(tt.mockOutput), nil)
			
			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			mapping, err := service.GetCommitsWithTaskIDsFiltered(context.Background(), tt.opts)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if mapping.TotalCommits != tt.expectedCommits {
				t.Errorf("Expected %d commits, got %d", tt.expectedCommits, mapping.TotalCommits)
			}
		})
	}
}

func TestCommitHistoryCaching(t *testing.T) {
	mockOutput := "abc123\tAuthor\t2024-01-01\tFix #1.2\ndef456\tAuthor\t2024-01-02\tImplement #2.1\n"
	mockRunner := NewMockCommandRunner()
	
	// Add response for the specific filter combination
	mockRunner.AddResponse("log --pretty=format:%H%x09%an%x09%ad%x09%s -n10 --date=iso", []byte(mockOutput), nil)
	
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newCommitHistoryCache(1 * time.Second)
	
	opts := CommitFilterOptions{Limit: 10}
	
	// First call - should hit git
	mapping1, err := service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mapping1.TotalCommits != 2 {
		t.Fatalf("Expected 2 commits, got %d", mapping1.TotalCommits)
	}
	
	// Check that git was called once
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected 1 git call, got %d", len(callLog))
	}
	
	// Second call within TTL - should use cache
	mapping2, err := service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mapping2.TotalCommits != 2 {
		t.Fatalf("Expected 2 commits from cache, got %d", mapping2.TotalCommits)
	}
	
	// Should still be just 1 git call (cached)
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected still 1 git call (cached), got %d", len(callLog))
	}
	
	// Wait for TTL to expire
	time.Sleep(1100 * time.Millisecond)
	
	// Third call after TTL - should hit git again
	mapping3, err := service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mapping3.TotalCommits != 2 {
		t.Fatalf("Expected 2 commits, got %d", mapping3.TotalCommits)
	}
	
	// Should now be 2 git calls
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls after TTL, got %d", len(callLog))
	}
}

func TestCommitHistoryCacheInvalidation(t *testing.T) {
	mockOutput := "abc123\tAuthor\t2024-01-01\tFix #1.2\n"
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("log --pretty=format:%H%x09%an%x09%ad%x09%s -n10 --date=iso", []byte(mockOutput), nil)
	
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newCommitHistoryCache(10 * time.Second) // Long TTL
	
	opts := CommitFilterOptions{Limit: 10}
	
	// First call
	_, err := service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Invalidate cache
	cache.invalidate()
	
	// Second call should hit git again (not cached)
	_, err = service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should be 2 git calls
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls after invalidation, got %d", len(callLog))
	}
}

func TestCommitHistoryCacheDifferentFilters(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("log --pretty=format:%H%x09%an%x09%ad%x09%s -n10 --date=iso", []byte("abc123\tAuthor\t2024-01-01\tFix #1.2\n"), nil)
	mockRunner.AddResponse("log --pretty=format:%H%x09%an%x09%ad%x09%s -n20 --date=iso", []byte("abc123\tAuthor\t2024-01-01\tFix #1.2\ndef456\tAuthor\t2024-01-02\tFix #2.1\n"), nil)
	
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newCommitHistoryCache(10 * time.Second)
	
	opts1 := CommitFilterOptions{Limit: 10}
	opts2 := CommitFilterOptions{Limit: 20}
	
	// Call with first filter
	mapping1, err := service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts1, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mapping1.TotalCommits != 1 {
		t.Fatalf("Expected 1 commit, got %d", mapping1.TotalCommits)
	}
	
	// Call with different filter - should not use cache
	mapping2, err := service.GetCommitsWithTaskIDsFilteredCached(context.Background(), opts2, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mapping2.TotalCommits != 2 {
		t.Fatalf("Expected 2 commits, got %d", mapping2.TotalCommits)
	}
	
	// Should be 2 git calls (different filters)
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls (different filters), got %d", len(callLog))
	}
}
