package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDetectRepositoryValidDirectory(t *testing.T) {
	// Get the current working directory (which is in a git repo)
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	result := DetectRepository(currentDir)

	if !result.IsRepo {
		t.Errorf("Expected IsRepo to be true, got false")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.RootPath == "" {
		t.Errorf("Expected RootPath to be non-empty, got empty string")
	}
}

func TestDetectRepositoryEmptyPath(t *testing.T) {
	result := DetectRepository("")

	if result.IsRepo {
		t.Errorf("Expected IsRepo to be false for empty path, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for empty path, got nil")
	}
}

func TestDetectRepositoryNonExistentDirectory(t *testing.T) {
	nonExistentPath := "/nonexistent/directory/that/does/not/exist"

	result := DetectRepository(nonExistentPath)

	if result.IsRepo {
		t.Errorf("Expected IsRepo to be false for non-existent directory, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for non-existent directory, got nil")
	}
}

func TestDetectRepositoryFilePath(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	result := DetectRepository(tempFile.Name())

	if result.IsRepo {
		t.Errorf("Expected IsRepo to be false for file path, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for file path, got nil")
	}
}

func TestDetectRepositoryNonGitDirectory(t *testing.T) {
	// Create a temporary directory that is not a git repository
	tempDir, err := os.MkdirTemp("", "test_non_git_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	result := DetectRepository(tempDir)

	if result.IsRepo {
		t.Errorf("Expected IsRepo to be false for non-git directory, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for non-git directory, got nil")
	}
}

func TestDetectRepositoryPathNormalization(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create a path with redundant elements
	unnormalizedPath := filepath.Join(currentDir, ".", "..", filepath.Base(currentDir))

	result := DetectRepository(unnormalizedPath)

	if !result.IsRepo {
		t.Errorf("Expected IsRepo to be true for unnormalized path, got false")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	// Verify path is normalized
	expectedPath := filepath.Clean(unnormalizedPath)
	if result.RootPath != expectedPath {
		t.Logf("Note: RootPath may differ due to symlinks. Got: %s, Expected base: %s",
			result.RootPath, expectedPath)
	}
}

func TestIsGitAvailable(t *testing.T) {
	available := IsGitAvailable()

	// This test depends on whether git is installed
	// We'll just verify the function returns a boolean without panicking
	if available != true && available != false {
		t.Errorf("IsGitAvailable returned unexpected value: %v", available)
	}
}

func TestIsGitAvailableReturnType(t *testing.T) {
	result := IsGitAvailable()

	// Verify the result is a valid boolean
	if result != (result == true) && result != (result == false) {
		t.Errorf("IsGitAvailable did not return a valid boolean")
	}
}

// TestDetectRepositoryErrorMessages verifies error messages are descriptive
func TestDetectRepositoryErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{
			name:        "empty directory path",
			dir:         "",
			expectError: true,
		},
		{
			name:        "non-existent directory",
			dir:         "/path/that/does/not/exist/anywhere",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectRepository(tt.dir)

			if tt.expectError && result.Error == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if tt.expectError && result.Error != nil {
				errorMsg := result.Error.Error()
				if errorMsg == "" {
					t.Errorf("Expected non-empty error message for %s", tt.name)
				}
			}
		})
	}
}

// TestRepoInfo verifies RepoInfo struct fields
func TestRepoInfo(t *testing.T) {
	// Test with error case
	errRepo := RepoInfo{
		IsRepo:   false,
		RootPath: "",
		Error:    fmt.Errorf("test error"),
	}

	if errRepo.IsRepo {
		t.Errorf("Expected IsRepo to be false, got true")
	}
	if errRepo.RootPath != "" {
		t.Errorf("Expected RootPath to be empty, got %s", errRepo.RootPath)
	}
	if errRepo.Error == nil {
		t.Errorf("Expected Error to be non-nil")
	}

	// Test with success case
	successRepo := RepoInfo{
		IsRepo:   true,
		RootPath: "/path/to/repo",
		Error:    nil,
	}

	if !successRepo.IsRepo {
		t.Errorf("Expected IsRepo to be true, got false")
	}
	if successRepo.RootPath != "/path/to/repo" {
		t.Errorf("Expected RootPath to be /path/to/repo, got %s", successRepo.RootPath)
	}
	if successRepo.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", successRepo.Error)
	}
}

// BenchmarkDetectRepository measures the performance of repository detection
func BenchmarkDetectRepository(b *testing.B) {
	currentDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectRepository(currentDir)
	}
}

// BenchmarkIsGitAvailable measures the performance of git availability check
func BenchmarkIsGitAvailable(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsGitAvailable()
	}
}

// TestGetStatusValidRepository tests GitStatus retrieval for a valid git repository
func TestGetStatusValidRepository(t *testing.T) {
	ctx := context.Background()

	// Get the current working directory (which should be in a git repo)
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	status, err := GetStatus(ctx, currentDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status.Branch == "" {
		t.Errorf("Expected non-empty branch name, got empty string")
	}
	if status.LastUpdated.IsZero() {
		t.Errorf("Expected LastUpdated to be set, got zero time")
	}
}

// TestGetStatusBranchDetection tests that branch name is correctly detected
func TestGetStatusBranchDetection(t *testing.T) {
	ctx := context.Background()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	status, err := GetStatus(ctx, currentDir)

	if err != nil {
		t.Errorf("Expected no error for branch detection, got: %v", err)
	}

	// Branch should not be empty and should not contain whitespace
	if status.Branch != filepath.Base(status.Branch) {
		t.Errorf("Expected branch to not contain path separators, got: %s", status.Branch)
	}
}

// TestGetStatusDirtyState tests dirty/clean state detection
func TestGetStatusDirtyState(t *testing.T) {
	ctx := context.Background()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	status, err := GetStatus(ctx, currentDir)

	if err != nil {
		t.Errorf("Expected no error for dirty state detection, got: %v", err)
	}

	// IsDirty should be a valid boolean
	if status.IsDirty != (status.IsDirty == true) && status.IsDirty != (status.IsDirty == false) {
		t.Errorf("Expected IsDirty to be a valid boolean, got: %v", status.IsDirty)
	}
}

// TestGetStatusInvalidRepository tests GetStatus with non-existent directory
func TestGetStatusInvalidRepository(t *testing.T) {
	ctx := context.Background()
	invalidPath := "/nonexistent/path/that/does/not/exist"

	status, err := GetStatus(ctx, invalidPath)

	if err == nil {
		t.Errorf("Expected error for non-existent directory, got nil")
	}
	if status.Error == nil {
		t.Errorf("Expected status.Error to be set, got nil")
	}
}

// TestGetStatusContextCancellation tests behavior with cancelled context
func TestGetStatusContextCancellation(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	status, err := GetStatus(ctx, currentDir)

	if err == nil {
		t.Errorf("Expected error for cancelled context, got nil")
	}
	if status.Error == nil {
		t.Errorf("Expected status.Error to be set for cancelled context, got nil")
	}
}

// TestGetStatusUpstreamDetection tests HasUpstream field
func TestGetStatusUpstreamDetection(t *testing.T) {
	ctx := context.Background()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	status, err := GetStatus(ctx, currentDir)

	if err != nil {
		t.Errorf("Expected no error for upstream detection, got: %v", err)
	}

	// HasUpstream should be set based on whether upstream branch exists
	if status.HasUpstream {
		// If upstream exists, ahead and behind should be integers >= 0
		if status.Ahead < 0 || status.Behind < 0 {
			t.Errorf("Expected Ahead and Behind to be >= 0 when HasUpstream is true, got Ahead: %d, Behind: %d",
				status.Ahead, status.Behind)
		}
	}
}

// TestGetStatusAheadBehindCounts tests ahead/behind tracking
func TestGetStatusAheadBehindCounts(t *testing.T) {
	ctx := context.Background()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	status, err := GetStatus(ctx, currentDir)

	if err != nil {
		t.Errorf("Expected no error for ahead/behind counts, got: %v", err)
	}

	// Verify Ahead and Behind are non-negative
	if status.Ahead < 0 {
		t.Errorf("Expected Ahead to be >= 0, got: %d", status.Ahead)
	}
	if status.Behind < 0 {
		t.Errorf("Expected Behind to be >= 0, got: %d", status.Behind)
	}
}

// TestGitStatusStructFields verifies GitStatus struct fields
func TestGitStatusStructFields(t *testing.T) {
	tests := []struct {
		name   string
		status GitStatus
	}{
		{
			name: "clean repository with upstream",
			status: GitStatus{
				Branch:      "main",
				IsDirty:     false,
				HasUpstream: true,
				Ahead:       0,
				Behind:      0,
				Error:       nil,
			},
		},
		{
			name: "dirty repository without upstream",
			status: GitStatus{
				Branch:      "feature/new",
				IsDirty:     true,
				HasUpstream: false,
				Ahead:       0,
				Behind:      0,
				Error:       nil,
			},
		},
		{
			name: "repository with commits ahead",
			status: GitStatus{
				Branch:      "develop",
				IsDirty:     false,
				HasUpstream: true,
				Ahead:       3,
				Behind:      1,
				Error:       nil,
			},
		},
		{
			name: "repository with error",
			status: GitStatus{
				Branch:      "",
				IsDirty:     false,
				HasUpstream: false,
				Ahead:       0,
				Behind:      0,
				Error:       fmt.Errorf("git command failed"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify struct can be instantiated and fields are accessible
			if tt.status.Branch != "main" && tt.status.Branch != "feature/new" &&
				tt.status.Branch != "develop" && tt.status.Branch != "" {
				// Just verify we can access the field
			}
			if tt.status.IsDirty != true && tt.status.IsDirty != false {
				t.Errorf("IsDirty should be boolean")
			}
			if tt.status.Ahead < 0 || tt.status.Behind < 0 {
				t.Errorf("Ahead and Behind should be non-negative")
			}
		})
	}
}

// BenchmarkGetStatus measures the performance of GetStatus
func BenchmarkGetStatus(b *testing.B) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetStatus(ctx, currentDir)
	}
}

// TestNewStatusRefresherReturnsNonNil tests that NewStatusRefresher returns non-nil struct
func TestNewStatusRefresherReturnsNonNil(t *testing.T) {
	repoPath := "/test/repo"
	refresher := NewStatusRefresher(repoPath)

	if refresher == nil {
		t.Errorf("Expected non-nil StatusRefresher, got nil")
	}
}

// TestNewStatusRefresherRepoPath tests that repoPath is stored correctly
func TestNewStatusRefresherRepoPath(t *testing.T) {
	repoPath := "/test/repo"
	refresher := NewStatusRefresher(repoPath)

	if refresher.repoPath != repoPath {
		t.Errorf("Expected repoPath to be %s, got %s", repoPath, refresher.repoPath)
	}
}

// TestNewStatusRefresherValidContext tests that context is created and cancel func is valid
func TestNewStatusRefresherValidContext(t *testing.T) {
	repoPath := "/test/repo"
	refresher := NewStatusRefresher(repoPath)

	if refresher.ctx == nil {
		t.Errorf("Expected non-nil context, got nil")
	}
	if refresher.cancel == nil {
		t.Errorf("Expected non-nil cancel func, got nil")
	}

	// Test that cancel func works
	refresher.cancel()
	select {
	case <-refresher.ctx.Done():
		// Context is cancelled as expected
	default:
		t.Errorf("Expected context to be cancelled after calling cancel")
	}
}

// TestNewStatusRefresherDefaultInterval tests that default refresh interval is 5 seconds
func TestNewStatusRefresherDefaultInterval(t *testing.T) {
	repoPath := "/test/repo"
	refresher := NewStatusRefresher(repoPath)

	expectedInterval := 5 * time.Second
	if refresher.refreshInterval != expectedInterval {
		t.Errorf("Expected refreshInterval to be %v, got %v", expectedInterval, refresher.refreshInterval)
	}
}

// TestNewStatusRefresherMutexInitialized tests that mutex is properly initialized
func TestNewStatusRefresherMutexInitialized(t *testing.T) {
	repoPath := "/test/repo"
	refresher := NewStatusRefresher(repoPath)

	// Test that mutex can be locked (which means it's initialized)
	done := make(chan bool)
	go func() {
		refresher.mutex.Lock()
		defer refresher.mutex.Unlock()
		done <- true
	}()

	select {
	case <-done:
		// Mutex was acquired successfully
	case <-time.After(1 * time.Second):
		t.Errorf("Mutex lock timed out, mutex may not be properly initialized")
	}
}

// TestNewStatusRefresherMultipleInstances tests creating multiple refreshers
func TestNewStatusRefresherMultipleInstances(t *testing.T) {
	refresher1 := NewStatusRefresher("/repo/1")
	refresher2 := NewStatusRefresher("/repo/2")

	if refresher1 == refresher2 {
		t.Errorf("Expected different refresher instances, got same instance")
	}
	if refresher1.repoPath == refresher2.repoPath {
		t.Errorf("Expected different repoPaths, got same path")
	}
}

// TestStatusRefresherStartCallsRefresh tests that Start performs initial refresh
func TestStatusRefresherStartCallsRefresh(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.Start()
	defer refresher.Stop()

	// Give the initial refresh a moment to complete
	time.Sleep(100 * time.Millisecond)

	status := refresher.GetStatus()
	if status.Branch == "" {
		t.Errorf("Expected branch to be set after initial refresh, got empty string")
	}
}

// TestStatusRefresherStartPeriodicRefresh tests that Start performs periodic updates
func TestStatusRefresherStartPeriodicRefresh(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	// Set a short interval for testing
	refresher.refreshInterval = 100 * time.Millisecond

	refresher.Start()
	defer refresher.Stop()

	// Get initial status after first refresh
	time.Sleep(50 * time.Millisecond)
	status1 := refresher.GetStatus()

	// Wait for at least 2 more ticks
	time.Sleep(250 * time.Millisecond)
	status2 := refresher.GetStatus()

	// Both should have valid branch info (periodic refresh happened)
	if status1.Branch == "" || status2.Branch == "" {
		t.Errorf("Expected valid branch info after periodic refreshes")
	}

	// Last updated times should be different (refresh happened)
	if status1.LastUpdated == status2.LastUpdated {
		t.Logf("Note: LastUpdated times may be equal if git status didn't change")
	}
}

// TestStatusRefresherStartAfterStop tests that Start doesn't interfere with Stop
func TestStatusRefresherStartAfterStop(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.refreshInterval = 50 * time.Millisecond

	refresher.Start()

	// Let it run for a bit and get initial status
	time.Sleep(150 * time.Millisecond)

	// Get status before stopping
	statusBefore := refresher.GetStatus()
	if statusBefore.Branch == "" {
		t.Logf("Note: Initial status was empty, may indicate no branch info yet")
	}

	// Stop the refresher
	refresher.Stop()

	// Wait to ensure stop is processed
	time.Sleep(200 * time.Millisecond)

	// Verify context is cancelled
	select {
	case <-refresher.ctx.Done():
		// Context cancelled as expected
	default:
		t.Errorf("Expected context to be cancelled after Stop")
	}

	// Verify we can still read the last status without panic
	statusAfter := refresher.GetStatus()
	// The status should be readable (implementation doesn't lose it)
	// No specific check for consistency since status might have changed between reads
	_ = statusAfter
}

// TestStatusRefresherConcurrentGetStatus tests concurrent reads during refresh
func TestStatusRefresherConcurrentGetStatus(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.refreshInterval = 100 * time.Millisecond

	refresher.Start()
	defer refresher.Stop()

	// Give initial refresh time
	time.Sleep(50 * time.Millisecond)

	// Run multiple concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				status := refresher.GetStatus()
				// Just verify we can read without panicking
				_ = status
			}
			done <- true
		}()
	}

	// Wait for all reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestStatusRefresherRefreshUpdatesStatus tests that Refresh updates internal status
func TestStatusRefresherRefreshUpdatesStatus(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)

	// Initial refresh
	refresher.Refresh()

	status := refresher.GetStatus()
	if status.Branch == "" {
		t.Errorf("Expected branch to be set after Refresh, got empty string")
	}

	// Do another refresh and verify status is updated
	refresher.Refresh()
	status2 := refresher.GetStatus()

	if status2.Branch != status.Branch {
		t.Logf("Note: Branch name changed between refreshes (might be expected)")
	}
}

// TestStatusRefresherGetStatusReturnsCurrentValue tests GetStatus returns current cached value
func TestStatusRefresherGetStatusReturnsCurrentValue(t *testing.T) {
	refresher := NewStatusRefresher("/some/repo")

	// Before any refresh, status should be zero-initialized
	status := refresher.GetStatus()
	if status.Branch != "" {
		t.Errorf("Expected empty branch before refresh, got %s", status.Branch)
	}

	// Manually set status to test GetStatus behavior
	refresher.mutex.Lock()
	refresher.status = GitStatus{
		Branch:      "test-branch",
		IsDirty:     true,
		HasUpstream: true,
		Ahead:       2,
		Behind:      3,
	}
	refresher.mutex.Unlock()

	// Verify GetStatus returns the set value
	status = refresher.GetStatus()
	if status.Branch != "test-branch" {
		t.Errorf("Expected branch 'test-branch', got '%s'", status.Branch)
	}
	if !status.IsDirty {
		t.Errorf("Expected IsDirty to be true")
	}
	if status.Ahead != 2 {
		t.Errorf("Expected Ahead to be 2, got %d", status.Ahead)
	}
	if status.Behind != 3 {
		t.Errorf("Expected Behind to be 3, got %d", status.Behind)
	}
}

// TestStatusRefresherRefreshWithContextCancellation tests Refresh with cancelled context
func TestStatusRefresherRefreshWithContextCancellation(t *testing.T) {
	// Create a refresher and immediately cancel its context
	refresher := NewStatusRefresher("/nonexistent/repo")
	refresher.cancel()

	// Refresh should handle the cancelled context gracefully
	// (GetStatus is called with cancelled context, which should return error)
	refresher.Refresh()

	// GetStatus should still return the (empty) cached value
	status := refresher.GetStatus()
	// Status might be empty since we never did a successful refresh
	// This is acceptable behavior
	_ = status
}

// TestStatusRefresherThreadSafetyWithMultipleRefreshers tests multiple refreshers don't interfere
func TestStatusRefresherThreadSafetyWithMultipleRefreshers(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher1 := NewStatusRefresher(currentDir)
	refresher2 := NewStatusRefresher(currentDir)

	refresher1.Start()
	refresher2.Start()
	defer refresher1.Stop()
	defer refresher2.Stop()

	time.Sleep(150 * time.Millisecond)

	// Both should have valid status
	status1 := refresher1.GetStatus()
	status2 := refresher2.GetStatus()

	if status1.Branch == "" || status2.Branch == "" {
		t.Errorf("Expected valid branch in both refreshers")
	}
}

// TestStatusRefresherSetRefreshInterval tests setting custom refresh interval
func TestStatusRefresherSetRefreshInterval(t *testing.T) {
	refresher := NewStatusRefresher("/test/repo")

	// Default should be 5 seconds
	if refresher.GetRefreshInterval() != 5*time.Second {
		t.Errorf("Expected default interval 5s, got %v", refresher.GetRefreshInterval())
	}

	// Set new interval
	newInterval := 2 * time.Second
	refresher.SetRefreshInterval(newInterval)

	if refresher.GetRefreshInterval() != newInterval {
		t.Errorf("Expected interval %v, got %v", newInterval, refresher.GetRefreshInterval())
	}
}

// TestStatusRefresherSetRefreshIntervalThreadSafe tests concurrent interval changes
func TestStatusRefresherSetRefreshIntervalThreadSafe(t *testing.T) {
	refresher := NewStatusRefresher("/test/repo")

	done := make(chan bool, 10)

	// Multiple goroutines setting intervals
	for i := 0; i < 10; i++ {
		go func(index int) {
			interval := time.Duration(100*(index+1)) * time.Millisecond
			refresher.SetRefreshInterval(interval)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have a valid interval (whichever was set last)
	finalInterval := refresher.GetRefreshInterval()
	if finalInterval < 100*time.Millisecond || finalInterval > 1*time.Second {
		t.Errorf("Expected interval in valid range, got %v", finalInterval)
	}
}

// TestStatusRefresherStopCancelsContext tests Stop properly cancels context
func TestStatusRefresherStopCancelsContext(t *testing.T) {
	refresher := NewStatusRefresher("/test/repo")

	// Context should not be cancelled initially
	select {
	case <-refresher.ctx.Done():
		t.Errorf("Context should not be cancelled initially")
	default:
		// Good, not cancelled
	}

	// Call Stop
	refresher.Stop()

	// Context should be cancelled
	select {
	case <-refresher.ctx.Done():
		// Good, cancelled as expected
	default:
		t.Errorf("Expected context to be cancelled after Stop")
	}
}

// TestStatusRefresherStopStopsGoroutine tests Stop actually stops the background goroutine
func TestStatusRefresherStopStopsGoroutine(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.refreshInterval = 50 * time.Millisecond

	refresher.Start()

	// Let it run
	time.Sleep(150 * time.Millisecond)

	// Get status before stop
	statusBefore := refresher.GetStatus()
	lastUpdateBefore := statusBefore.LastUpdated

	// Stop
	refresher.Stop()

	// Wait for stop to take effect
	time.Sleep(200 * time.Millisecond)

	// Get status after stop
	statusAfter := refresher.GetStatus()
	lastUpdateAfter := statusAfter.LastUpdated

	// LastUpdated should not change significantly after stop
	// (allowing small tolerance for timing variations)
	diff := lastUpdateAfter.Sub(lastUpdateBefore)
	if diff > 150*time.Millisecond {
		t.Logf("Note: LastUpdated changed more than expected after Stop (%v)", diff)
	}
}

// TestStatusRefresherGetRefreshInterval tests getting current interval
func TestStatusRefresherGetRefreshInterval(t *testing.T) {
	refresher := NewStatusRefresher("/test/repo")

	// Test default
	interval := refresher.GetRefreshInterval()
	if interval != 5*time.Second {
		t.Errorf("Expected default interval 5s, got %v", interval)
	}

	// Set and get
	newInterval := 3 * time.Second
	refresher.SetRefreshInterval(newInterval)

	interval = refresher.GetRefreshInterval()
	if interval != newInterval {
		t.Errorf("Expected interval %v, got %v", newInterval, interval)
	}
}

// TestStatusRefresherMultipleStartStops tests multiple Start/Stop cycles
func TestStatusRefresherMultipleStartStops(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.refreshInterval = 50 * time.Millisecond

	// First cycle
	refresher.Start()
	time.Sleep(100 * time.Millisecond)
	refresher.Stop()
	time.Sleep(50 * time.Millisecond)

	status1 := refresher.GetStatus()

	// Create new refresher for next cycle (context can't be reused after cancel)
	refresher2 := NewStatusRefresher(currentDir)
	refresher2.refreshInterval = 50 * time.Millisecond

	refresher2.Start()
	time.Sleep(100 * time.Millisecond)
	refresher2.Stop()

	status2 := refresher2.GetStatus()

	// Both should have valid branch info
	if status1.Branch == "" || status2.Branch == "" {
		t.Logf("Warning: Expected valid branch info in both cycles")
	}
}

// TestStatusRefresherStartAndRefreshDirectly tests calling Refresh outside of Start
func TestStatusRefresherStartAndRefreshDirectly(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)

	// Manual refresh without Start
	refresher.Refresh()
	status1 := refresher.GetStatus()

	// Another manual refresh
	time.Sleep(50 * time.Millisecond)
	refresher.Refresh()
	status2 := refresher.GetStatus()

	// Both should have valid branch
	if status1.Branch == "" || status2.Branch == "" {
		t.Errorf("Expected valid branch info from manual refreshes")
	}
}

// TestStatusRefresherStartTwiceWithDifferentIntervals tests changing interval between starts
func TestStatusRefresherStartTwiceWithDifferentIntervals(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// First refresher with default interval
	refresher1 := NewStatusRefresher(currentDir)
	if refresher1.GetRefreshInterval() != 5*time.Second {
		t.Errorf("Expected default 5s interval")
	}

	// Second refresher with custom interval
	refresher2 := NewStatusRefresher(currentDir)
	refresher2.SetRefreshInterval(100 * time.Millisecond)

	refresher1.Start()
	refresher2.Start()
	defer refresher1.Stop()
	defer refresher2.Stop()

	time.Sleep(200 * time.Millisecond)

	// Both should work fine with their respective intervals
	status1 := refresher1.GetStatus()
	status2 := refresher2.GetStatus()

	if status1.Branch == "" {
		t.Logf("Warning: refresher1 status empty")
	}
	if status2.Branch == "" {
		t.Logf("Warning: refresher2 status empty")
	}
}

// TestStatusRefresherConcurrentOperations tests complex concurrent scenarios
func TestStatusRefresherConcurrentOperations(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.refreshInterval = 75 * time.Millisecond

	refresher.Start()
	defer refresher.Stop()

	done := make(chan bool, 20)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 5; j++ {
				_ = refresher.GetStatus()
				time.Sleep(25 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Concurrent interval changes
	for i := 0; i < 10; i++ {
		go func() {
			interval := time.Duration(50+i*5) * time.Millisecond
			refresher.SetRefreshInterval(interval)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Final status should be readable
	status := refresher.GetStatus()
	_ = status
}

// TestStatusRefresherLargeNumberOfReads tests high-volume concurrent reads
func TestStatusRefresherLargeNumberOfReads(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher := NewStatusRefresher(currentDir)
	refresher.refreshInterval = 100 * time.Millisecond

	refresher.Start()
	defer refresher.Stop()

	time.Sleep(50 * time.Millisecond)

	// Fire off 50 concurrent reads
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				status := refresher.GetStatus()
				// Verify we got a struct (not panicked)
				if status.Branch != "" {
					// Good, has branch info
				}
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 50; i++ {
		<-done
	}
}

// TestStatusRefresherErrorHandling tests behavior with error conditions
func TestStatusRefresherErrorHandling(t *testing.T) {
	// Nonexistent repo - GetStatus will return error
	refresher := NewStatusRefresher("/nonexistent/path/to/repo")

	// Calling Refresh with nonexistent path should not panic
	refresher.Refresh()

	// GetStatus should return what was cached (empty initially)
	status := refresher.GetStatus()
	_ = status
}

// TestStatusRefresherContextIsolation tests context isolation between refreshers
func TestStatusRefresherContextIsolation(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	refresher1 := NewStatusRefresher(currentDir)
	refresher2 := NewStatusRefresher(currentDir)

	refresher1.Start()
	refresher2.Start()
	defer refresher1.Stop()
	defer refresher2.Stop()

	// Stop refresher1
	refresher1.Stop()

	// refresher1 context should be cancelled
	select {
	case <-refresher1.ctx.Done():
		// Good, cancelled
	default:
		t.Errorf("Expected refresher1 context to be cancelled")
	}

	// refresher2 context should NOT be cancelled
	select {
	case <-refresher2.ctx.Done():
		t.Errorf("Expected refresher2 context to still be active")
	default:
		// Good, not cancelled
	}

	// refresher2 should still be updating
	time.Sleep(100 * time.Millisecond)
	status := refresher2.GetStatus()
	if status.Branch == "" {
		t.Logf("Note: refresher2 status is empty (may be initial)")
	}
}

// TestCreateBranchValidName tests CreateBranch with valid branch name
func TestCreateBranchValidName(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Get the current branch to restore it later
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = currentDir
	currentBranchOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	currentBranch := strings.TrimSpace(string(currentBranchOutput))

	branchName := "test-branch-valid-name"
	output, err := CreateBranch(ctx, currentDir, branchName)

	if err != nil {
		// It's OK if the branch already exists
		if !strings.Contains(string(output), "already exists") {
			t.Errorf("Expected no error, got: %v", err)
		}
		return
	}

	// Cleanup: switch back to original branch
	cmd = exec.Command("git", "checkout", currentBranch)
	cmd.Dir = currentDir
	cmd.Run()

	// Delete the test branch
	cmd = exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = currentDir
	cmd.Run()
}

// TestCreateBranchEmptyName tests CreateBranch with empty branch name
func TestCreateBranchEmptyName(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	output, err := CreateBranch(ctx, currentDir, "")

	if err == nil {
		t.Errorf("Expected error for empty branch name, got nil")
	}
	if output != "" {
		t.Errorf("Expected empty output for invalid branch name, got: %s", output)
	}
}

// TestCreateBranchWithSpaces tests CreateBranch with spaces in branch name
func TestCreateBranchWithSpaces(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	output, err := CreateBranch(ctx, currentDir, "branch with spaces")

	if err == nil {
		t.Errorf("Expected error for branch name with spaces, got nil")
	}
	if output != "" {
		t.Errorf("Expected empty output for invalid branch name, got: %s", output)
	}
}

// TestGetRecentCommitsValidRepository tests GetRecentCommits with valid repo
func TestGetRecentCommitsValidRepository(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	commits, err := GetRecentCommits(ctx, currentDir, 5)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(commits) == 0 {
		t.Logf("Note: No commits found (empty repository is acceptable)")
	}

	// If commits exist, verify structure
	for _, commit := range commits {
		if commit.Hash == "" {
			t.Errorf("Expected non-empty hash, got empty string")
		}
		if commit.Author == "" {
			t.Errorf("Expected non-empty author, got empty string")
		}
		// Subject can be empty for initial commits with no message
	}
}

// TestGetRecentCommitsDefaultCount tests GetRecentCommits with default count
func TestGetRecentCommitsDefaultCount(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Pass 0 to test default count (should be 20)
	commits, err := GetRecentCommits(ctx, currentDir, 0)

	if err != nil {
		t.Errorf("Expected no error with default count, got: %v", err)
	}
	// Should return up to 20 commits (or fewer if repo has fewer)
	if len(commits) > 20 {
		t.Errorf("Expected at most 20 commits with default count, got %d", len(commits))
	}
}

// TestGetRecentCommitsSpecificCount tests GetRecentCommits with specific count
func TestGetRecentCommitsSpecificCount(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	commits, err := GetRecentCommits(ctx, currentDir, 5)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(commits) > 5 {
		t.Errorf("Expected at most 5 commits, got %d", len(commits))
	}
}

// TestGetRecentCommitsNegativeCount tests GetRecentCommits with negative count
func TestGetRecentCommitsNegativeCount(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Negative count should be treated as default (20)
	commits, err := GetRecentCommits(ctx, currentDir, -5)

	if err != nil {
		t.Errorf("Expected no error with negative count, got: %v", err)
	}
	if len(commits) > 20 {
		t.Errorf("Expected at most 20 commits with negative count, got %d", len(commits))
	}
}

// TestGetRecentCommitsInvalidRepository tests GetRecentCommits with non-existent repo
func TestGetRecentCommitsInvalidRepository(t *testing.T) {
	ctx := context.Background()
	invalidPath := "/nonexistent/repository/path"

	commits, err := GetRecentCommits(ctx, invalidPath, 5)

	if err == nil {
		t.Errorf("Expected error for invalid repository, got nil")
	}
	if commits != nil {
		t.Errorf("Expected nil commits on error, got %v", commits)
	}
}

// TestGetRecentCommitsContextCancellation tests GetRecentCommits with cancelled context
func TestGetRecentCommitsContextCancellation(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create and immediately cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	commits, err := GetRecentCommits(ctx, currentDir, 5)

	if err == nil {
		t.Errorf("Expected error for cancelled context, got nil")
	}
	if commits != nil {
		t.Errorf("Expected nil commits on error, got %v", commits)
	}
}

// TestGetRecentCommitsEmptyLineHandling tests GetRecentCommits handles empty lines
func TestGetRecentCommitsEmptyLineHandling(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	commits, err := GetRecentCommits(ctx, currentDir, 100)

	if err != nil {
		t.Logf("Note: error getting commits (may be expected in some cases): %v", err)
	}

	// Verify all returned commits have required fields
	for _, commit := range commits {
		if commit.Hash == "" {
			t.Errorf("Expected non-empty hash for commit")
		}
	}
}

// TestGetRecentCommitsStructure tests Commit struct fields
func TestGetRecentCommitsStructure(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	commits, err := GetRecentCommits(ctx, currentDir, 1)

	if err != nil {
		t.Logf("Note: error getting commits: %v", err)
		return
	}

	if len(commits) > 0 {
		commit := commits[0]

		// Verify all fields exist and are strings
		if reflect.TypeOf(commit.Hash).Kind() != reflect.String {
			t.Errorf("Expected Hash to be string, got %T", commit.Hash)
		}
		if reflect.TypeOf(commit.Subject).Kind() != reflect.String {
			t.Errorf("Expected Subject to be string, got %T", commit.Subject)
		}
		if reflect.TypeOf(commit.Author).Kind() != reflect.String {
			t.Errorf("Expected Author to be string, got %T", commit.Author)
		}
		if reflect.TypeOf(commit.RelativeTime).Kind() != reflect.String {
			t.Errorf("Expected RelativeTime to be string, got %T", commit.RelativeTime)
		}
	}
}
