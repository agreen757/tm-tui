package filechanges

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/git"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// mockGitService is a mock implementation of git.GitService for testing
type mockGitService struct {
	uncommittedChanges []git.FileChange
	uncommittedError   error
	commitsMapping     *git.TaskCommitMapping
	commitsError       error
}

func newMockGitService() *mockGitService {
	return &mockGitService{
		uncommittedChanges: []git.FileChange{},
		commitsMapping:     git.NewTaskCommitMapping(),
	}
}

func (m *mockGitService) GetUncommittedChanges(ctx context.Context) ([]git.FileChange, error) {
	if m.uncommittedError != nil {
		return nil, m.uncommittedError
	}
	return m.uncommittedChanges, nil
}

func (m *mockGitService) GetCommitsWithTaskIDsFiltered(ctx context.Context, opts git.CommitFilterOptions) (*git.TaskCommitMapping, error) {
	if m.commitsError != nil {
		return nil, m.commitsError
	}
	return m.commitsMapping, nil
}

// mockStorage is a mock implementation of taskmaster.Storage for testing
type mockStorage struct {
	mapping   *taskmaster.FileChangeMapping
	loadError error
	saveError error
	saveCalls int
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		mapping: taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
}

func (m *mockStorage) Load() (*taskmaster.FileChangeMapping, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	return m.mapping, nil
}

func (m *mockStorage) Save(mapping *taskmaster.FileChangeMapping) error {
	m.saveCalls++
	if m.saveError != nil {
		return m.saveError
	}
	m.mapping = mapping
	return nil
}

func TestNewFileChangeTracker(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	
	if tracker.gitService == nil {
		t.Error("expected gitService to be set")
	}
	
	if tracker.storage == nil {
		t.Error("expected storage to be set")
	}
	
	if tracker.repoPath != "/test/repo" {
		t.Errorf("expected repoPath=/test/repo, got %s", tracker.repoPath)
	}
	
	if tracker.stopChan == nil {
		t.Error("expected stopChan to be initialized")
	}
}

func TestInitialize_EmptyStorage(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	// Create tracker with mock git service
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
	}
	
	// Initialize with empty storage
	ctx := context.Background()
	err := tracker.Initialize(ctx)
	
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	if tracker.mapping == nil {
		t.Fatal("expected mapping to be initialized")
	}
	
	if tracker.mapping.Version != taskmaster.SerializationVersion {
		t.Errorf("expected version=%s, got %s", taskmaster.SerializationVersion, tracker.mapping.Version)
	}
	
	if len(tracker.mapping.Tasks) != 0 {
		t.Errorf("expected empty tasks map, got %d entries", len(tracker.mapping.Tasks))
	}
	
	if storage.saveCalls != 1 {
		t.Errorf("expected 1 save call, got %d", storage.saveCalls)
	}
}

func TestInitialize_ExistingData(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	// Pre-populate storage with existing data
	existingMapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc, _ := taskmaster.NewFileChange("src/main.go", "modified", "existing change")
	existingMapping.AddFileChange("1.1", *fc)
	storage.mapping = existingMapping
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
	}
	
	ctx := context.Background()
	err := tracker.Initialize(ctx)
	
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	if tracker.mapping == nil {
		t.Fatal("expected mapping to be loaded")
	}
	
	if len(tracker.mapping.Tasks) != 1 {
		t.Errorf("expected 1 task in mapping, got %d", len(tracker.mapping.Tasks))
	}
	
	changes := tracker.GetFileChangesForTask("1.1")
	if len(changes) != 1 {
		t.Errorf("expected 1 file change for task 1.1, got %d", len(changes))
	}
	
	if changes[0].Path != "src/main.go" {
		t.Errorf("expected path=src/main.go, got %s", changes[0].Path)
	}
}

func TestSetActiveTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initially empty
	if tracker.GetActiveTask() != "" {
		t.Errorf("expected empty active task, got %s", tracker.GetActiveTask())
	}
	
	// Set active task
	tracker.SetActiveTask("2.3")
	
	if tracker.GetActiveTask() != "2.3" {
		t.Errorf("expected active task=2.3, got %s", tracker.GetActiveTask())
	}
	
	// Change active task
	tracker.SetActiveTask("4.1")
	
	if tracker.GetActiveTask() != "4.1" {
		t.Errorf("expected active task=4.1, got %s", tracker.GetActiveTask())
	}
}

func TestGetFileChangesForTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with some data
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc1, _ := taskmaster.NewFileChange("src/a.go", "modified", "change 1")
	fc2, _ := taskmaster.NewFileChange("src/b.go", "added", "change 2")
	mapping.AddFileChange("1.1", *fc1)
	mapping.AddFileChange("1.1", *fc2)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Get changes for task 1.1
	changes := tracker.GetFileChangesForTask("1.1")
	
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	
	if changes[0].Path != "src/a.go" {
		t.Errorf("expected first path=src/a.go, got %s", changes[0].Path)
	}
	
	if changes[1].Path != "src/b.go" {
		t.Errorf("expected second path=src/b.go, got %s", changes[1].Path)
	}
	
	// Get changes for non-existent task
	noChanges := tracker.GetFileChangesForTask("999")
	if len(noChanges) != 0 {
		t.Errorf("expected 0 changes for non-existent task, got %d", len(noChanges))
	}
}

func TestGetAllTaskFileChanges(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with data for multiple tasks
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc1, _ := taskmaster.NewFileChange("src/a.go", "modified", "change 1")
	fc2, _ := taskmaster.NewFileChange("src/b.go", "added", "change 2")
	fc3, _ := taskmaster.NewFileChange("src/c.go", "deleted", "change 3")
	
	mapping.AddFileChange("1.1", *fc1)
	mapping.AddFileChange("1.2", *fc2)
	mapping.AddFileChange("1.2", *fc3)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Get all changes
	allChanges := tracker.GetAllTaskFileChanges()
	
	if len(allChanges) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(allChanges))
	}
	
	if len(allChanges["1.1"]) != 1 {
		t.Errorf("expected 1 change for task 1.1, got %d", len(allChanges["1.1"]))
	}
	
	if len(allChanges["1.2"]) != 2 {
		t.Errorf("expected 2 changes for task 1.2, got %d", len(allChanges["1.2"]))
	}
}

func TestGetUnassignedChanges(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with unassigned changes
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc1, _ := taskmaster.NewFileChange("src/unassigned1.go", "modified", "unassigned 1")
	fc2, _ := taskmaster.NewFileChange("src/unassigned2.go", "added", "unassigned 2")
	
	mapping.AddUnassignedChange(*fc1)
	mapping.AddUnassignedChange(*fc2)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Get unassigned changes
	unassigned := tracker.GetUnassignedChanges()
	
	if len(unassigned) != 2 {
		t.Errorf("expected 2 unassigned changes, got %d", len(unassigned))
	}
	
	if unassigned[0].Path != "src/unassigned1.go" {
		t.Errorf("expected first path=src/unassigned1.go, got %s", unassigned[0].Path)
	}
	
	if unassigned[1].Path != "src/unassigned2.go" {
		t.Errorf("expected second path=src/unassigned2.go, got %s", unassigned[1].Path)
	}
}

func TestStartPeriodicRefresh(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize mapping
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.mutex.Unlock()
	
	// Start periodic refresh with short interval
	interval := 50 * time.Millisecond
	tracker.StartPeriodicRefresh(interval)
	
	// Verify ticker is running
	tracker.mutex.RLock()
	if tracker.refreshTicker == nil {
		t.Error("expected refreshTicker to be started")
	}
	tracker.mutex.RUnlock()
	
	// Wait for at least one tick
	time.Sleep(100 * time.Millisecond)
	
	// Try starting again - should be no-op
	tracker.StartPeriodicRefresh(interval)
	
	// Stop the tracker
	tracker.Stop()
	
	// Verify ticker is stopped
	tracker.mutex.RLock()
	if tracker.refreshTicker != nil {
		t.Error("expected refreshTicker to be stopped")
	}
	tracker.mutex.RUnlock()
}

func TestStop(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize mapping
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.mutex.Unlock()
	
	// Start periodic refresh
	tracker.StartPeriodicRefresh(100 * time.Millisecond)
	
	// Verify it's running
	tracker.mutex.RLock()
	running := tracker.refreshTicker != nil
	tracker.mutex.RUnlock()
	
	if !running {
		t.Error("expected ticker to be running")
	}
	
	// Stop the tracker
	tracker.Stop()
	
	// Verify it's stopped
	tracker.mutex.RLock()
	stopped := tracker.refreshTicker == nil
	tracker.mutex.RUnlock()
	
	if !stopped {
		t.Error("expected ticker to be stopped")
	}
	
	// Stop again should be safe (no panic)
	tracker.Stop()
}

func TestConcurrentAccess(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with some data
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc, _ := taskmaster.NewFileChange("src/concurrent.go", "modified", "concurrent test")
	mapping.AddFileChange("1.1", *fc)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Run concurrent operations
	done := make(chan bool)
	
	// Goroutine 1: Read active task
	go func() {
		for i := 0; i < 100; i++ {
			_ = tracker.GetActiveTask()
		}
		done <- true
	}()
	
	// Goroutine 2: Set active task
	go func() {
		for i := 0; i < 100; i++ {
			tracker.SetActiveTask("1.1")
		}
		done <- true
	}()
	
	// Goroutine 3: Get file changes
	go func() {
		for i := 0; i < 100; i++ {
			_ = tracker.GetFileChangesForTask("1.1")
		}
		done <- true
	}()
	
	// Goroutine 4: Get all changes
	go func() {
		for i := 0; i < 100; i++ {
			_ = tracker.GetAllTaskFileChanges()
		}
		done <- true
	}()
	
	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
	
	// If we got here without deadlock or data race, test passes
}

func TestGetFileChangesForTask_NilMapping(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Don't initialize mapping (leave as nil)
	changes := tracker.GetFileChangesForTask("1.1")
	
	if changes == nil {
		t.Error("expected empty slice, got nil")
	}
	
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestGetAllTaskFileChanges_NilMapping(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Don't initialize mapping (leave as nil)
	allChanges := tracker.GetAllTaskFileChanges()
	
	if allChanges == nil {
		t.Error("expected empty map, got nil")
	}
	
	if len(allChanges) != 0 {
		t.Errorf("expected 0 entries, got %d", len(allChanges))
	}
}

func TestGetUnassignedChanges_NilMapping(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Don't initialize mapping (leave as nil)
	unassigned := tracker.GetUnassignedChanges()
	
	if unassigned == nil {
		t.Error("expected empty slice, got nil")
	}
	
	if len(unassigned) != 0 {
		t.Errorf("expected 0 changes, got %d", len(unassigned))
	}
}

func TestRefreshChanges(t *testing.T) {
	mockGit := newMockGitService()
	storage := newMockStorage()
	
	// Add some uncommitted changes to mock
	mockGit.uncommittedChanges = []git.FileChange{
		{
			Path:       "src/main.go",
			ChangeType: git.ChangeTypeModified,
			IsStaged:   false,
			IsModified: true,
		},
		{
			Path:       "src/new.go",
			ChangeType: git.ChangeTypeAdded,
			IsStaged:   true,
			IsModified: false,
		},
	}
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
		mapping:    taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
	
	// Set active task
	tracker.SetActiveTask("2.1")
	
	// Refresh changes
	ctx := context.Background()
	err := tracker.RefreshChanges(ctx)
	
	if err != nil {
		t.Fatalf("RefreshChanges failed: %v", err)
	}
	
	// Verify changes were added to active task
	changes := tracker.GetFileChangesForTask("2.1")
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes for task 2.1, got %d", len(changes))
	}
	
	// Verify change details
	if changes[0].Path != "src/main.go" {
		t.Errorf("expected first change path=src/main.go, got %s", changes[0].Path)
	}
	if changes[0].ChangeType != "modified" {
		t.Errorf("expected first change type=modified, got %s", changes[0].ChangeType)
	}
	if !changes[0].IsPending {
		t.Error("expected first change to be pending")
	}
	
	if changes[1].Path != "src/new.go" {
		t.Errorf("expected second change path=src/new.go, got %s", changes[1].Path)
	}
	if changes[1].ChangeType != "added" {
		t.Errorf("expected second change type=added, got %s", changes[1].ChangeType)
	}
	
	// Verify storage was saved
	if storage.saveCalls != 1 {
		t.Errorf("expected 1 save call, got %d", storage.saveCalls)
	}
}

func TestRefreshChanges_NoActiveTask(t *testing.T) {
	mockGit := newMockGitService()
	storage := newMockStorage()
	
	// Add uncommitted changes
	mockGit.uncommittedChanges = []git.FileChange{
		{
			Path:       "src/unassigned.go",
			ChangeType: git.ChangeTypeModified,
			IsStaged:   false,
			IsModified: true,
		},
	}
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
		mapping:    taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
	
	// Don't set active task
	
	// Refresh changes
	ctx := context.Background()
	err := tracker.RefreshChanges(ctx)
	
	if err != nil {
		t.Fatalf("RefreshChanges failed: %v", err)
	}
	
	// Verify change was added to unassigned
	unassigned := tracker.GetUnassignedChanges()
	if len(unassigned) != 1 {
		t.Fatalf("expected 1 unassigned change, got %d", len(unassigned))
	}
	
	if unassigned[0].Path != "src/unassigned.go" {
		t.Errorf("expected path=src/unassigned.go, got %s", unassigned[0].Path)
	}
}

func TestRefreshChanges_ClearsPendingChanges(t *testing.T) {
	mockGit := newMockGitService()
	storage := newMockStorage()
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
		mapping:    taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
	
	// Add some pending changes manually
	pendingChange := taskmaster.FileChange{
		Path:        "src/old-pending.go",
		ChangeType:  "modified",
		Description: "old pending",
		LastChanged: time.Now(),
		IsPending:   true,
	}
	tracker.mapping.AddFileChange("1.1", pendingChange)
	
	// Add a committed change that should be preserved
	committedChange := taskmaster.FileChange{
		Path:        "src/committed.go",
		ChangeType:  "modified",
		Description: "committed",
		LastChanged: time.Now(),
		CommitID:    "abc123",
		IsPending:   false,
	}
	tracker.mapping.AddFileChange("1.1", committedChange)
	
	// Mock git returns no uncommitted changes
	mockGit.uncommittedChanges = []git.FileChange{}
	
	// Refresh changes
	ctx := context.Background()
	err := tracker.RefreshChanges(ctx)
	
	if err != nil {
		t.Fatalf("RefreshChanges failed: %v", err)
	}
	
	// Verify pending change was removed
	changes := tracker.GetFileChangesForTask("1.1")
	if len(changes) != 1 {
		t.Fatalf("expected 1 change remaining, got %d", len(changes))
	}
	
	// Verify committed change was preserved
	if changes[0].Path != "src/committed.go" {
		t.Errorf("expected committed change to be preserved, got %s", changes[0].Path)
	}
	if changes[0].IsPending {
		t.Error("expected preserved change to not be pending")
	}
}

func TestRefreshChanges_GitServiceError(t *testing.T) {
	mockGit := newMockGitService()
	storage := newMockStorage()
	
	// Set git service to return error
	mockGit.uncommittedError = fmt.Errorf("git command failed")
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
		mapping:    taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
	
	// Refresh changes
	ctx := context.Background()
	err := tracker.RefreshChanges(ctx)
	
	if err == nil {
		t.Fatal("expected error from RefreshChanges")
	}
	
	if !strings.Contains(err.Error(), "git command failed") {
		t.Errorf("expected error message to contain 'git command failed', got: %v", err)
	}
}

func TestRefreshChanges_StorageSaveError(t *testing.T) {
	mockGit := newMockGitService()
	storage := newMockStorage()
	
	// Set storage to return error on save
	storage.saveError = fmt.Errorf("disk full")
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
		mapping:    taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
	
	// Add some changes to git
	mockGit.uncommittedChanges = []git.FileChange{
		{
			Path:       "src/test.go",
			ChangeType: git.ChangeTypeModified,
		},
	}
	
	tracker.SetActiveTask("1.1")
	
	// Refresh changes
	ctx := context.Background()
	err := tracker.RefreshChanges(ctx)
	
	if err == nil {
		t.Fatal("expected error from RefreshChanges")
	}
	
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("expected error message to contain 'disk full', got: %v", err)
	}
}

func TestGitChangeTypeToString(t *testing.T) {
	tests := []struct {
		input    git.ChangeType
		expected string
	}{
		{git.ChangeTypeAdded, "added"},
		{git.ChangeTypeModified, "modified"},
		{git.ChangeTypeDeleted, "deleted"},
		{git.ChangeTypeRenamed, "modified"},
		{git.ChangeTypeCopied, "added"},
		{git.ChangeTypeUntracked, "added"},
		{git.ChangeTypeIgnored, "modified"},
	}
	
	for _, tt := range tests {
		result := gitChangeTypeToString(tt.input)
		if result != tt.expected {
			t.Errorf("gitChangeTypeToString(%v) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestBuildChangeDescription(t *testing.T) {
	tests := []struct {
		name     string
		change   git.FileChange
		expected string
	}{
		{
			name: "renamed file",
			change: git.FileChange{
				Path:       "new.go",
				OldPath:    "old.go",
				ChangeType: git.ChangeTypeRenamed,
			},
			expected: "Renamed from old.go",
		},
		{
			name: "copied file",
			change: git.FileChange{
				Path:       "copy.go",
				OldPath:    "original.go",
				ChangeType: git.ChangeTypeCopied,
			},
			expected: "Copied from original.go",
		},
		{
			name: "untracked file",
			change: git.FileChange{
				Path:       "new-untracked.go",
				ChangeType: git.ChangeTypeUntracked,
			},
			expected: "Untracked file",
		},
		{
			name: "staged addition",
			change: git.FileChange{
				Path:       "added.go",
				ChangeType: git.ChangeTypeAdded,
				IsStaged:   true,
			},
			expected: "Staged for addition",
		},
		{
			name: "staged and modified",
			change: git.FileChange{
				Path:       "modified.go",
				ChangeType: git.ChangeTypeModified,
				IsStaged:   true,
				IsModified: true,
			},
			expected: "Staged and modified in working tree",
		},
		{
			name: "staged for commit",
			change: git.FileChange{
				Path:       "staged.go",
				ChangeType: git.ChangeTypeModified,
				IsStaged:   true,
			},
			expected: "Staged for commit",
		},
		{
			name: "modified in working tree",
			change: git.FileChange{
				Path:       "working.go",
				ChangeType: git.ChangeTypeModified,
				IsModified: true,
			},
			expected: "Modified in working tree",
		},
		{
			name: "staged deletion",
			change: git.FileChange{
				Path:       "deleted.go",
				ChangeType: git.ChangeTypeDeleted,
				IsStaged:   true,
			},
			expected: "Staged for deletion",
		},
		{
			name: "deleted file",
			change: git.FileChange{
				Path:       "deleted.go",
				ChangeType: git.ChangeTypeDeleted,
			},
			expected: "Deleted",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildChangeDescription(tt.change)
			if result != tt.expected {
				t.Errorf("buildChangeDescription() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestConvertGitChangeToFileChange(t *testing.T) {
	gitChange := git.FileChange{
		Path:       "src/test.go",
		OldPath:    "src/old.go",
		Status:     "R ",
		ChangeType: git.ChangeTypeRenamed,
		IsStaged:   true,
		IsModified: false,
	}
	
	fileChange := convertGitChangeToFileChange(gitChange)
	
	if fileChange.Path != "src/test.go" {
		t.Errorf("expected path=src/test.go, got %s", fileChange.Path)
	}
	
	if fileChange.ChangeType != "modified" {
		t.Errorf("expected changeType=modified (renamed), got %s", fileChange.ChangeType)
	}
	
	if fileChange.Description != "Renamed from src/old.go" {
		t.Errorf("expected description='Renamed from src/old.go', got %s", fileChange.Description)
	}
	
	if !fileChange.IsPending {
		t.Error("expected pending=true")
	}
	
	if fileChange.CommitID != "" {
		t.Errorf("expected empty commitID, got %s", fileChange.CommitID)
	}
}

func TestAssociateChangeWithTask_WithActiveTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize mapping
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.activeTask = "3.2"
	tracker.mutex.Unlock()
	
	// Create a file change
	fc := taskmaster.FileChange{
		Path:        "src/associated.go",
		ChangeType:  "modified",
		Description: "test change",
		LastChanged: time.Now(),
		IsPending:   true,
	}
	
	// Associate the change
	tracker.associateChangeWithTask(fc)
	
	// Verify change was added to active task
	changes := tracker.GetFileChangesForTask("3.2")
	if len(changes) != 1 {
		t.Fatalf("expected 1 change for task 3.2, got %d", len(changes))
	}
	
	if changes[0].Path != "src/associated.go" {
		t.Errorf("expected path=src/associated.go, got %s", changes[0].Path)
	}
}

func TestAssociateChangeWithTask_NoActiveTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize mapping without active task
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.activeTask = ""
	tracker.mutex.Unlock()
	
	// Create a file change
	fc := taskmaster.FileChange{
		Path:        "src/unassigned.go",
		ChangeType:  "added",
		Description: "test unassigned",
		LastChanged: time.Now(),
		IsPending:   true,
	}
	
	// Associate the change (should go to unassigned)
	tracker.associateChangeWithTask(fc)
	
	// Verify change was added to unassigned
	unassigned := tracker.GetUnassignedChanges()
	if len(unassigned) != 1 {
		t.Fatalf("expected 1 unassigned change, got %d", len(unassigned))
	}
	
	if unassigned[0].Path != "src/unassigned.go" {
		t.Errorf("expected path=src/unassigned.go, got %s", unassigned[0].Path)
	}
}

func TestTaskSwitchingDuringUncommittedChanges(t *testing.T) {
	mockGit := newMockGitService()
	storage := newMockStorage()
	
	// First set of uncommitted changes
	mockGit.uncommittedChanges = []git.FileChange{
		{
			Path:       "src/file1.go",
			ChangeType: git.ChangeTypeModified,
			IsStaged:   false,
			IsModified: true,
		},
	}
	
	tracker := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
		mapping:    taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion),
	}
	
	// Set active task to 1.1
	tracker.SetActiveTask("1.1")
	
	// Refresh changes - should associate with task 1.1
	ctx := context.Background()
	err := tracker.RefreshChanges(ctx)
	if err != nil {
		t.Fatalf("First refresh failed: %v", err)
	}
	
	// Verify first change associated with task 1.1
	changes1 := tracker.GetFileChangesForTask("1.1")
	if len(changes1) != 1 {
		t.Fatalf("expected 1 change for task 1.1, got %d", len(changes1))
	}
	if changes1[0].Path != "src/file1.go" {
		t.Errorf("expected path=src/file1.go, got %s", changes1[0].Path)
	}
	
	// Switch active task to 2.1
	tracker.SetActiveTask("2.1")
	
	// New uncommitted changes
	mockGit.uncommittedChanges = []git.FileChange{
		{
			Path:       "src/file2.go",
			ChangeType: git.ChangeTypeAdded,
			IsStaged:   true,
			IsModified: false,
		},
	}
	
	// Refresh changes - should associate with task 2.1
	err = tracker.RefreshChanges(ctx)
	if err != nil {
		t.Fatalf("Second refresh failed: %v", err)
	}
	
	// Verify task 1.1 no longer has pending changes (cleared by refresh)
	changes1After := tracker.GetFileChangesForTask("1.1")
	if len(changes1After) != 0 {
		t.Errorf("expected 0 pending changes for task 1.1 after switch, got %d", len(changes1After))
	}
	
	// Verify second change associated with task 2.1
	changes2 := tracker.GetFileChangesForTask("2.1")
	if len(changes2) != 1 {
		t.Fatalf("expected 1 change for task 2.1, got %d", len(changes2))
	}
	if changes2[0].Path != "src/file2.go" {
		t.Errorf("expected path=src/file2.go, got %s", changes2[0].Path)
	}
}

func TestGetFileChangesForTask_FilteringByPending(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with mixed pending and committed changes
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	
	pendingChange := taskmaster.FileChange{
		Path:        "src/pending.go",
		ChangeType:  "modified",
		Description: "pending change",
		LastChanged: time.Now(),
		IsPending:   true,
	}
	
	committedChange := taskmaster.FileChange{
		Path:        "src/committed.go",
		ChangeType:  "added",
		Description: "committed change",
		LastChanged: time.Now(),
		CommitID:    "abc123",
		IsPending:   false,
	}
	
	mapping.AddFileChange("4.1", pendingChange)
	mapping.AddFileChange("4.1", committedChange)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Get all changes for task 4.1
	allChanges := tracker.GetFileChangesForTask("4.1")
	if len(allChanges) != 2 {
		t.Fatalf("expected 2 total changes, got %d", len(allChanges))
	}
	
	// Verify we can filter by pending status
	var pendingOnly []taskmaster.FileChange
	var committedOnly []taskmaster.FileChange
	
	for _, change := range allChanges {
		if change.IsPending {
			pendingOnly = append(pendingOnly, change)
		} else {
			committedOnly = append(committedOnly, change)
		}
	}
	
	if len(pendingOnly) != 1 {
		t.Errorf("expected 1 pending change, got %d", len(pendingOnly))
	}
	
	if len(committedOnly) != 1 {
		t.Errorf("expected 1 committed change, got %d", len(committedOnly))
	}
	
	if len(pendingOnly) > 0 && pendingOnly[0].Path != "src/pending.go" {
		t.Errorf("expected pending path=src/pending.go, got %s", pendingOnly[0].Path)
	}
	
	if len(committedOnly) > 0 && committedOnly[0].Path != "src/committed.go" {
		t.Errorf("expected committed path=src/committed.go, got %s", committedOnly[0].Path)
	}
}

func TestAssociateChangesWithTask_FromUnassigned(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with unassigned changes
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc1, _ := taskmaster.NewFileChange("src/file1.go", "modified", "unassigned 1")
	fc2, _ := taskmaster.NewFileChange("src/file2.go", "added", "unassigned 2")
	
	mapping.AddUnassignedChange(*fc1)
	mapping.AddUnassignedChange(*fc2)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Associate files with task 5.1
	ctx := context.Background()
	err := tracker.AssociateChangesWithTask(ctx, []string{"src/file1.go", "src/file2.go"}, "5.1")
	
	if err != nil {
		t.Fatalf("AssociateChangesWithTask failed: %v", err)
	}
	
	// Verify changes moved to task 5.1
	changes := tracker.GetFileChangesForTask("5.1")
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes for task 5.1, got %d", len(changes))
	}
	
	// Verify changes removed from unassigned
	unassigned := tracker.GetUnassignedChanges()
	if len(unassigned) != 0 {
		t.Errorf("expected 0 unassigned changes, got %d", len(unassigned))
	}
	
	// Verify storage was updated
	if storage.saveCalls < 1 {
		t.Error("expected storage.Save to be called")
	}
}

func TestAssociateChangesWithTask_NewFiles(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with empty mapping
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Associate new files with task 2.3
	ctx := context.Background()
	err := tracker.AssociateChangesWithTask(ctx, []string{"src/new1.go", "src/new2.go"}, "2.3")
	
	if err != nil {
		t.Fatalf("AssociateChangesWithTask failed: %v", err)
	}
	
	// Verify changes added to task 2.3
	changes := tracker.GetFileChangesForTask("2.3")
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes for task 2.3, got %d", len(changes))
	}
	
	if changes[0].Path != "src/new1.go" {
		t.Errorf("expected first path=src/new1.go, got %s", changes[0].Path)
	}
	
	if changes[1].Path != "src/new2.go" {
		t.Errorf("expected second path=src/new2.go, got %s", changes[1].Path)
	}
	
	// Verify changes are not pending (manual associations)
	if changes[0].IsPending {
		t.Error("expected manual change to not be pending")
	}
}

func TestAssociateChangesWithTask_ConflictResolution(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with file already in task 1.1
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc, _ := taskmaster.NewFileChange("src/conflict.go", "modified", "existing in 1.1")
	mapping.AddFileChange("1.1", *fc)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Try to associate same file with task 2.2 (should move)
	ctx := context.Background()
	err := tracker.AssociateChangesWithTask(ctx, []string{"src/conflict.go"}, "2.2")
	
	if err != nil {
		t.Fatalf("AssociateChangesWithTask failed: %v", err)
	}
	
	// Verify file moved from task 1.1 to 2.2
	changes1 := tracker.GetFileChangesForTask("1.1")
	if len(changes1) != 0 {
		t.Errorf("expected 0 changes for task 1.1, got %d", len(changes1))
	}
	
	changes2 := tracker.GetFileChangesForTask("2.2")
	if len(changes2) != 1 {
		t.Fatalf("expected 1 change for task 2.2, got %d", len(changes2))
	}
	
	if changes2[0].Path != "src/conflict.go" {
		t.Errorf("expected path=src/conflict.go, got %s", changes2[0].Path)
	}
}

func TestAssociateChangesWithTask_EmptyTaskID(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.mutex.Unlock()
	
	ctx := context.Background()
	err := tracker.AssociateChangesWithTask(ctx, []string{"src/test.go"}, "")
	
	if err == nil {
		t.Fatal("expected error for empty taskID")
	}
	
	if err.Error() != "taskID cannot be empty" {
		t.Errorf("expected 'taskID cannot be empty' error, got: %v", err)
	}
}

func TestAssociateChangesWithTask_EmptyFiles(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.mutex.Unlock()
	
	ctx := context.Background()
	err := tracker.AssociateChangesWithTask(ctx, []string{}, "1.1")
	
	if err == nil {
		t.Fatal("expected error for empty files list")
	}
	
	if err.Error() != "files list cannot be empty" {
		t.Errorf("expected 'files list cannot be empty' error, got: %v", err)
	}
}

func TestRemoveFileChangeFromTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with files in task
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc1, _ := taskmaster.NewFileChange("src/keep.go", "modified", "keep this")
	fc2, _ := taskmaster.NewFileChange("src/remove.go", "added", "remove this")
	
	mapping.AddFileChange("3.1", *fc1)
	mapping.AddFileChange("3.1", *fc2)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Remove one file
	ctx := context.Background()
	err := tracker.RemoveFileChangeFromTask(ctx, "src/remove.go", "3.1")
	
	if err != nil {
		t.Fatalf("RemoveFileChangeFromTask failed: %v", err)
	}
	
	// Verify only one file remains
	changes := tracker.GetFileChangesForTask("3.1")
	if len(changes) != 1 {
		t.Fatalf("expected 1 change remaining, got %d", len(changes))
	}
	
	if changes[0].Path != "src/keep.go" {
		t.Errorf("expected path=src/keep.go, got %s", changes[0].Path)
	}
	
	// Verify storage was updated
	if storage.saveCalls < 1 {
		t.Error("expected storage.Save to be called")
	}
}

func TestRemoveFileChangeFromTask_NotFound(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with file in task
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc, _ := taskmaster.NewFileChange("src/exists.go", "modified", "exists")
	mapping.AddFileChange("4.1", *fc)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Try to remove non-existent file
	ctx := context.Background()
	err := tracker.RemoveFileChangeFromTask(ctx, "src/nonexistent.go", "4.1")
	
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestMoveChangeBetweenTasks(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with file in task 1.1
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc, _ := taskmaster.NewFileChange("src/move.go", "modified", "will be moved")
	mapping.AddFileChange("1.1", *fc)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Move file from task 1.1 to 2.1
	ctx := context.Background()
	err := tracker.MoveChangeBetweenTasks(ctx, "src/move.go", "1.1", "2.1")
	
	if err != nil {
		t.Fatalf("MoveChangeBetweenTasks failed: %v", err)
	}
	
	// Verify file removed from task 1.1
	changes1 := tracker.GetFileChangesForTask("1.1")
	if len(changes1) != 0 {
		t.Errorf("expected 0 changes for task 1.1, got %d", len(changes1))
	}
	
	// Verify file added to task 2.1
	changes2 := tracker.GetFileChangesForTask("2.1")
	if len(changes2) != 1 {
		t.Fatalf("expected 1 change for task 2.1, got %d", len(changes2))
	}
	
	if changes2[0].Path != "src/move.go" {
		t.Errorf("expected path=src/move.go, got %s", changes2[0].Path)
	}
	
	// Verify storage was updated
	if storage.saveCalls < 1 {
		t.Error("expected storage.Save to be called")
	}
}

func TestMoveChangeBetweenTasks_SameTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	tracker.mutex.Lock()
	tracker.mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	tracker.mutex.Unlock()
	
	// Try to move with same source and destination
	ctx := context.Background()
	err := tracker.MoveChangeBetweenTasks(ctx, "src/test.go", "1.1", "1.1")
	
	if err == nil {
		t.Fatal("expected error for same source and destination")
	}
	
	if !strings.Contains(err.Error(), "same") {
		t.Errorf("expected 'same' error, got: %v", err)
	}
}

func TestMoveChangeBetweenTasks_FileNotFound(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	
	// Initialize with empty task
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc, _ := taskmaster.NewFileChange("src/other.go", "modified", "other file")
	mapping.AddFileChange("1.1", *fc)
	
	tracker.mutex.Lock()
	tracker.mapping = mapping
	tracker.mutex.Unlock()
	
	// Try to move non-existent file
	ctx := context.Background()
	err := tracker.MoveChangeBetweenTasks(ctx, "src/nonexistent.go", "1.1", "2.1")
	
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestPersistenceAcrossServiceRestarts(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	// First instance: create and save associations
	mapping := taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
	fc1, _ := taskmaster.NewFileChange("src/persistent1.go", "modified", "manual association 1")
	fc2, _ := taskmaster.NewFileChange("src/persistent2.go", "added", "manual association 2")
	
	// Set as not pending (committed) so they won't be cleared
	fc1.IsPending = false
	fc2.IsPending = false
	
	mapping.AddFileChange("6.1", *fc1)
	mapping.AddFileChange("6.1", *fc2)
	
	// Save directly to storage
	err := storage.Save(mapping)
	if err != nil {
		t.Fatalf("storage.Save failed: %v", err)
	}
	
	// Second instance: load from storage (simulating service restart)
	tracker2 := &FileChangeTracker{
		gitService: mockGit,
		storage:    storage,
		repoPath:   "/test/repo",
		stopChan:   make(chan struct{}),
	}
	
	ctx := context.Background()
	err = tracker2.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Verify changes persisted
	changes := tracker2.GetFileChangesForTask("6.1")
	if len(changes) != 2 {
		t.Fatalf("expected 2 persisted changes, got %d", len(changes))
	}
	
	if changes[0].Path != "src/persistent1.go" {
		t.Errorf("expected first path=src/persistent1.go, got %s", changes[0].Path)
	}
	
	if changes[1].Path != "src/persistent2.go" {
		t.Errorf("expected second path=src/persistent2.go, got %s", changes[1].Path)
	}
}

func TestProcessCommit_SingleTaskID(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create test file changes
	fc1, _ := taskmaster.NewFileChange("src/file1.go", "modified", "modified file")
	fc2, _ := taskmaster.NewFileChange("src/file2.go", "added", "new file")
	files := []taskmaster.FileChange{*fc1, *fc2}
	
	// Process commit with single task ID
	commitID := "abc123def456"
	message := "Implement #2.1 for user authentication"
	
	tracker.mutex.Lock()
	err = tracker.ProcessCommit(commitID, message, files)
	tracker.mutex.Unlock()
	
	if err != nil {
		t.Fatalf("ProcessCommit failed: %v", err)
	}
	
	// Verify files are associated with task 2.1
	changes := tracker.GetFileChangesForTask("2.1")
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes for task 2.1, got %d", len(changes))
	}
	
	if changes[0].CommitID != commitID {
		t.Errorf("expected CommitID=%s, got %s", commitID, changes[0].CommitID)
	}
	
	if changes[0].IsPending {
		t.Error("expected IsPending=false for committed changes")
	}
}

func TestProcessCommit_MultipleTaskIDs(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create test file changes
	fc1, _ := taskmaster.NewFileChange("src/auth.go", "modified", "auth modifications")
	files := []taskmaster.FileChange{*fc1}
	
	// Process commit with multiple task IDs
	commitID := "xyz789abc"
	message := "Implement #2.1 and #2.2 for auth system"
	
	tracker.mutex.Lock()
	err = tracker.ProcessCommit(commitID, message, files)
	tracker.mutex.Unlock()
	
	if err != nil {
		t.Fatalf("ProcessCommit failed: %v", err)
	}
	
	// Verify file is associated with both tasks
	changes21 := tracker.GetFileChangesForTask("2.1")
	changes22 := tracker.GetFileChangesForTask("2.2")
	
	if len(changes21) != 1 {
		t.Fatalf("expected 1 change for task 2.1, got %d", len(changes21))
	}
	
	if len(changes22) != 1 {
		t.Fatalf("expected 1 change for task 2.2, got %d", len(changes22))
	}
	
	if changes21[0].Path != "src/auth.go" {
		t.Errorf("expected path=src/auth.go in task 2.1, got %s", changes21[0].Path)
	}
	
	if changes22[0].Path != "src/auth.go" {
		t.Errorf("expected path=src/auth.go in task 2.2, got %s", changes22[0].Path)
	}
}

func TestProcessCommit_NoTaskIDs(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create test file changes
	fc1, _ := taskmaster.NewFileChange("src/file1.go", "modified", "modified file")
	files := []taskmaster.FileChange{*fc1}
	
	// Process commit without task IDs
	commitID := "no_task_commit"
	message := "Fix typo in documentation"
	
	tracker.mutex.Lock()
	err = tracker.ProcessCommit(commitID, message, files)
	tracker.mutex.Unlock()
	
	if err != nil {
		t.Fatalf("ProcessCommit failed: %v", err)
	}
	
	// Verify file is added to unassigned changes
	unassigned := tracker.GetUnassignedChanges()
	if len(unassigned) != 1 {
		t.Fatalf("expected 1 unassigned change, got %d", len(unassigned))
	}
	
	if unassigned[0].Path != "src/file1.go" {
		t.Errorf("expected path=src/file1.go in unassigned, got %s", unassigned[0].Path)
	}
	
	if unassigned[0].CommitID != commitID {
		t.Errorf("expected CommitID=%s, got %s", commitID, unassigned[0].CommitID)
	}
}

func TestProcessCommitHistory(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	// Set up mock git service with commits
	commit1 := git.CommitInfo{
		Hash:    "commit1hash",
		Author:  "Test Author",
		Date:    "2024-01-15",
		Message: "Implement #1.1",
		TaskIDs: []string{"1.1"},
	}
	
	commit2 := git.CommitInfo{
		Hash:    "commit2hash",
		Author:  "Test Author",
		Date:    "2024-01-16",
		Message: "Implement #1.2",
		TaskIDs: []string{"1.2"},
	}
	
	mockGit.commitsMapping.AddCommit(commit1)
	mockGit.commitsMapping.AddCommit(commit2)
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// ProcessCommitHistory should be callable without errors
	// Note: getFilesInCommit returns empty list currently, so no files will be associated
	err = tracker.ProcessCommitHistory(ctx)
	if err != nil {
		t.Fatalf("ProcessCommitHistory failed: %v", err)
	}
}

func TestGetCommitMessageSuggestion_WithActiveTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Set active task
	tracker.SetActiveTask("3.2.1")
	
	// Get suggestion
	suggestion := tracker.GetCommitMessageSuggestion()
	
	expected := "Implement #3.2.1"
	if suggestion != expected {
		t.Errorf("expected suggestion=%s, got %s", expected, suggestion)
	}
}

func TestGetCommitMessageSuggestion_NoActiveTask(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Get suggestion without setting active task
	suggestion := tracker.GetCommitMessageSuggestion()
	
	if suggestion != "" {
		t.Errorf("expected empty suggestion when no active task, got %s", suggestion)
	}
}

func TestProcessCommit_NestedTaskIDs(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Create test file changes
	fc1, _ := taskmaster.NewFileChange("src/config.go", "modified", "configuration changes")
	files := []taskmaster.FileChange{*fc1}
	
	// Process commit with nested task ID
	commitID := "nested_task_commit"
	message := "Configure #4.1.2.1 for database setup"
	
	tracker.mutex.Lock()
	err = tracker.ProcessCommit(commitID, message, files)
	tracker.mutex.Unlock()
	
	if err != nil {
		t.Fatalf("ProcessCommit failed: %v", err)
	}
	
	// Verify file is associated with nested task ID
	changes := tracker.GetFileChangesForTask("4.1.2.1")
	if len(changes) != 1 {
		t.Fatalf("expected 1 change for task 4.1.2.1, got %d", len(changes))
	}
	
	if changes[0].Path != "src/config.go" {
		t.Errorf("expected path=src/config.go, got %s", changes[0].Path)
	}
}

func TestSetCommitMessageTemplate(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Set a custom template
	customTemplate := "feat: implement #{{.TaskID}}"
	tracker.SetCommitMessageTemplate(customTemplate)
	
	// Set active task
	tracker.SetActiveTask("2.1")
	
	// Get suggestion - should use custom template
	suggestion := tracker.GetCommitMessageSuggestion()
	expected := "feat: implement #2.1"
	
	if suggestion != expected {
		t.Errorf("expected suggestion=%s, got %s", expected, suggestion)
	}
}

func TestGetCommitMessageTemplate(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Check default template
	defaultTemplate := tracker.GetCommitMessageTemplate()
	if defaultTemplate != "Implement #{{.TaskID}}" {
		t.Errorf("expected default template=Implement #{{.TaskID}}, got %s", defaultTemplate)
	}
	
	// Set custom template
	customTemplate := "fix: resolve #{{.TaskID}}"
	tracker.SetCommitMessageTemplate(customTemplate)
	
	// Verify get returns the custom template
	retrieved := tracker.GetCommitMessageTemplate()
	if retrieved != customTemplate {
		t.Errorf("expected template=%s, got %s", customTemplate, retrieved)
	}
}

func TestResetCommitMessageTemplate(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Set custom template
	tracker.SetCommitMessageTemplate("custom: #{{.TaskID}}")
	
	// Reset to default
	tracker.ResetCommitMessageTemplate()
	
	// Verify it's back to default
	current := tracker.GetCommitMessageTemplate()
	if current != "Implement #{{.TaskID}}" {
		t.Errorf("expected default template after reset, got %s", current)
	}
}

func TestGetCommitMessageSuggestion_CustomTemplate(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Test various custom templates
	testCases := []struct {
		name        string
		template    string
		taskID      string
		expected    string
	}{
		{
			name:        "feat_template",
			template:    "feat: implement #{{.TaskID}}",
			taskID:      "1.1",
			expected:    "feat: implement #1.1",
		},
		{
			name:        "fix_template",
			template:    "fix: resolve #{{.TaskID}}",
			taskID:      "2.2",
			expected:    "fix: resolve #2.2",
		},
		{
			name:        "docs_template",
			template:    "docs: update for #{{.TaskID}}",
			taskID:      "3.1",
			expected:    "docs: update for #3.1",
		},
		{
			name:        "simple_hash_format",
			template:    "#{{.TaskID}}",
			taskID:      "4.2.1",
			expected:    "#4.2.1",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tracker.SetCommitMessageTemplate(tc.template)
			tracker.SetActiveTask(tc.taskID)
			
			suggestion := tracker.GetCommitMessageSuggestion()
			if suggestion != tc.expected {
				t.Errorf("expected=%s, got=%s", tc.expected, suggestion)
			}
		})
	}
}

func TestGetCommitMessageSuggestion_InvalidTemplate(t *testing.T) {
	storage := newMockStorage()
	mockGit := newMockGitService()
	
	tracker := NewFileChangeTracker(mockGit, storage, "/test/repo")
	ctx := context.Background()
	
	// Initialize the tracker
	err := tracker.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	// Set an invalid template (will fall back to default format)
	tracker.SetCommitMessageTemplate("{{.InvalidField}}")
	tracker.SetActiveTask("5.1")
	
	// Should fall back to simple format
	suggestion := tracker.GetCommitMessageSuggestion()
	if suggestion == "" || suggestion == "{{.InvalidField}}" {
		t.Errorf("expected fallback format, got %s", suggestion)
	}
}


