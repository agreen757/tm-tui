package filechanges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// Mock GitService for content loading tests
type mockContentGitService struct {
	contentMap map[string]string // Key: "commit:path"
	errors     map[string]error  // Key: "commit:path"
}

func newMockContentGitService() *mockContentGitService {
	return &mockContentGitService{
		contentMap: make(map[string]string),
		errors:     make(map[string]error),
	}
}

func (m *mockContentGitService) GetFileContentAtCommit(ctx context.Context, commit, file string) (string, error) {
	key := fmt.Sprintf("%s:%s", commit, file)
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if content, ok := m.contentMap[key]; ok {
		return content, nil
	}
	return "", fmt.Errorf("file not found at commit")
}

func (m *mockContentGitService) setContent(commit, file, content string) {
	key := fmt.Sprintf("%s:%s", commit, file)
	m.contentMap[key] = content
}

func (m *mockContentGitService) setError(commit, file string, err error) {
	key := fmt.Sprintf("%s:%s", commit, file)
	m.errors[key] = err
}

// Test helper to create a temporary file
func createTempFile(t *testing.T, dir, relativePath, content string) string {
	fullPath := filepath.Join(dir, relativePath)
	
	// Create parent directories if needed
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	return fullPath
}

func TestNewContentLoader(t *testing.T) {
	gitService := newMockContentGitService()
	repoPath := "/test/repo"
	
	loader := NewContentLoader(gitService, repoPath)
	
	if loader == nil {
		t.Fatal("NewContentLoader returned nil")
	}
	
	if loader.GetGitService() != gitService {
		t.Error("GitService not set correctly")
	}
	
	if loader.GetRepoPath() != repoPath {
		t.Error("RepoPath not set correctly")
	}
	
	if loader.cache == nil {
		t.Error("Cache not initialized")
	}
}

func TestLoadContent_PendingFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file
	testContent := "pending file content\nline 2\nline 3"
	relativePath := "src/main.go"
	createTempFile(t, tempDir, relativePath, testContent)
	
	// Create loader
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	// Create file change representing pending file
	file := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: true,
	}
	
	// Load content
	ctx := context.Background()
	content, err := loader.LoadContent(ctx, file)
	
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}
	
	if content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}
}

func TestLoadContent_CommittedFile(t *testing.T) {
	gitService := newMockContentGitService()
	
	// Set up mock response
	commitID := "abc1234"
	filePath := "src/main.go"
	expectedContent := "committed file content"
	gitService.setContent(commitID, filePath, expectedContent)
	
	// Create loader
	loader := NewContentLoader(gitService, "/test/repo")
	
	// Create file change representing committed file
	file := taskmaster.FileChange{
		Path:      filePath,
		CommitID:  commitID,
		IsPending: false,
	}
	
	// Load content
	ctx := context.Background()
	content, err := loader.LoadContent(ctx, file)
	
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}
	
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestLoadContent_CurrentFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file
	testContent := "current file content"
	relativePath := "README.md"
	createTempFile(t, tempDir, relativePath, testContent)
	
	// Create loader
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	// Create file change representing current file (not pending, no commit ID)
	file := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: false,
		CommitID:  "",
	}
	
	// Load content
	ctx := context.Background()
	content, err := loader.LoadContent(ctx, file)
	
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}
	
	if content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}
}

func TestLoadContent_FileNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      "nonexistent.txt",
		IsPending: true,
	}
	
	ctx := context.Background()
	_, err = loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
	
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestLoadContent_Directory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      "subdir",
		IsPending: true,
	}
	
	ctx := context.Background()
	_, err = loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected error for directory, got nil")
	}
	
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Expected 'directory' error, got: %v", err)
	}
}

func TestLoadContent_LargeFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a file larger than 10MB
	largePath := filepath.Join(tempDir, "large.bin")
	f, err := os.Create(largePath)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	
	// Write 11MB of data
	data := make([]byte, 11*1024*1024)
	if _, err := f.Write(data); err != nil {
		f.Close()
		t.Fatalf("Failed to write large file: %v", err)
	}
	f.Close()
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      "large.bin",
		IsPending: true,
	}
	
	ctx := context.Background()
	_, err = loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected error for large file, got nil")
	}
	
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' error, got: %v", err)
	}
}

func TestLoadContent_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}
	
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a file with no read permissions
	restrictedPath := filepath.Join(tempDir, "restricted.txt")
	if err := os.WriteFile(restrictedPath, []byte("restricted"), 0000); err != nil {
		t.Fatalf("Failed to create restricted file: %v", err)
	}
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      "restricted.txt",
		IsPending: true,
	}
	
	ctx := context.Background()
	_, err = loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected permission error, got nil")
	}
}

func TestLoadContent_GitServiceError(t *testing.T) {
	gitService := newMockContentGitService()
	gitService.setError("abc1234567", "main.go", fmt.Errorf("git command failed"))
	
	loader := NewContentLoader(gitService, "/test/repo")
	
	file := taskmaster.FileChange{
		Path:      "main.go",
		CommitID:  "abc1234567",
		IsPending: false,
	}
	
	ctx := context.Background()
	_, err := loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected git service error, got nil")
	}
	
	if !strings.Contains(err.Error(), "git service failed") {
		t.Errorf("Expected 'git service failed' error, got: %v", err)
	}
}

func TestLoadContent_EmptyPath(t *testing.T) {
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, "/test/repo")
	
	file := taskmaster.FileChange{
		Path:      "",
		IsPending: true,
	}
	
	ctx := context.Background()
	_, err := loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
	
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

func TestLoadContent_InvalidCommitID(t *testing.T) {
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, "/test/repo")
	
	file := taskmaster.FileChange{
		Path:      "main.go",
		CommitID:  "abc", // Too short
		IsPending: false,
	}
	
	ctx := context.Background()
	_, err := loader.LoadContent(ctx, file)
	
	if err == nil {
		t.Error("Expected error for invalid commit ID, got nil")
	}
	
	if !strings.Contains(err.Error(), "invalid commit ID") {
		t.Errorf("Expected 'invalid commit ID' error, got: %v", err)
	}
}

func TestContentCache_HitAndMiss(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testContent := "cached content"
	relativePath := "cache-test.txt"
	createTempFile(t, tempDir, relativePath, testContent)
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: true,
	}
	
	ctx := context.Background()
	
	// First load - cache miss
	content1, err := loader.LoadContent(ctx, file)
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}
	
	if content1 != testContent {
		t.Errorf("Expected %q, got %q", testContent, content1)
	}
	
	// Second load - should hit cache
	content2, err := loader.LoadContent(ctx, file)
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}
	
	if content2 != testContent {
		t.Errorf("Expected %q, got %q", testContent, content2)
	}
	
	// Verify cache size
	if loader.cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", loader.cache.Size())
	}
}

func TestContentCache_LRUEviction(t *testing.T) {
	gitService := newMockContentGitService()
	
	// Create loader with small cache (3 entries)
	loader := NewContentLoader(gitService, "/test/repo")
	loader.cache = newContentCache(3)
	
	// Add content for multiple commits
	for i := 0; i < 5; i++ {
		commitID := fmt.Sprintf("commit%d", i)
		filePath := "file.txt"
		content := fmt.Sprintf("content %d", i)
		
		gitService.setContent(commitID, filePath, content)
		
		file := taskmaster.FileChange{
			Path:      filePath,
			CommitID:  commitID,
			IsPending: false,
		}
		
		ctx := context.Background()
		_, err := loader.LoadContent(ctx, file)
		if err != nil {
			t.Fatalf("Load failed for commit %d: %v", i, err)
		}
	}
	
	// Cache should only have last 3 entries
	cacheSize := loader.cache.Size()
	if cacheSize != 3 {
		t.Errorf("Expected cache size 3, got %d", cacheSize)
	}
}

func TestContentCache_InvalidateCache(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testContent := "content to invalidate"
	relativePath := "invalidate-test.txt"
	createTempFile(t, tempDir, relativePath, testContent)
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: true,
	}
	
	ctx := context.Background()
	
	// Load content to populate cache
	_, err = loader.LoadContent(ctx, file)
	if err != nil {
		t.Fatalf("Initial load failed: %v", err)
	}
	
	if loader.cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", loader.cache.Size())
	}
	
	// Invalidate cache
	loader.InvalidateCache()
	
	if loader.cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after invalidation, got %d", loader.cache.Size())
	}
}

func TestContentCache_DifferentFileStates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test file
	testContent := "multi-state content"
	relativePath := "state-test.txt"
	createTempFile(t, tempDir, relativePath, testContent)
	
	gitService := newMockContentGitService()
	gitService.setContent("abc1234567", relativePath, "committed content")
	
	loader := NewContentLoader(gitService, tempDir)
	ctx := context.Background()
	
	// Load as pending
	filePending := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: true,
	}
	contentPending, err := loader.LoadContent(ctx, filePending)
	if err != nil {
		t.Fatalf("Load pending failed: %v", err)
	}
	
	// Load as committed
	fileCommitted := taskmaster.FileChange{
		Path:      relativePath,
		CommitID:  "abc1234567",
		IsPending: false,
	}
	contentCommitted, err := loader.LoadContent(ctx, fileCommitted)
	if err != nil {
		t.Fatalf("Load committed failed: %v", err)
	}
	
	// Load as current
	fileCurrent := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: false,
		CommitID:  "",
	}
	contentCurrent, err := loader.LoadContent(ctx, fileCurrent)
	if err != nil {
		t.Fatalf("Load current failed: %v", err)
	}
	
	// Verify different contents and caching
	if contentPending != testContent {
		t.Errorf("Pending content mismatch")
	}
	
	if contentCommitted != "committed content" {
		t.Errorf("Committed content mismatch")
	}
	
	if contentCurrent != testContent {
		t.Errorf("Current content mismatch")
	}
	
	// Should have 3 separate cache entries
	if loader.cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", loader.cache.Size())
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		file     taskmaster.FileChange
		expected string
	}{
		{
			name: "Committed file",
			file: taskmaster.FileChange{
				Path:      "main.go",
				CommitID:  "abc123",
				IsPending: false,
			},
			expected: "commit:abc123:main.go",
		},
		{
			name: "Pending file",
			file: taskmaster.FileChange{
				Path:      "test.go",
				IsPending: true,
			},
			expected: "pending:test.go",
		},
		{
			name: "Current file",
			file: taskmaster.FileChange{
				Path:      "readme.md",
				IsPending: false,
				CommitID:  "",
			},
			expected: "current:readme.md",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generateCacheKey(tt.file)
			if key != tt.expected {
				t.Errorf("Expected key %q, got %q", tt.expected, key)
			}
		})
	}
}

func TestLoadContent_NestedPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testContent := "nested file content"
	relativePath := "src/internal/pkg/file.go"
	createTempFile(t, tempDir, relativePath, testContent)
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: true,
	}
	
	ctx := context.Background()
	content, err := loader.LoadContent(ctx, file)
	
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}
	
	if content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}
}

func TestLoadContent_Concurrent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create multiple test files
	numFiles := 10
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf("file %d content", i)
		path := fmt.Sprintf("file%d.txt", i)
		createTempFile(t, tempDir, path, content)
	}
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	ctx := context.Background()
	
	// Load files concurrently
	done := make(chan bool, numFiles)
	for i := 0; i < numFiles; i++ {
		go func(index int) {
			file := taskmaster.FileChange{
				Path:      fmt.Sprintf("file%d.txt", index),
				IsPending: true,
			}
			
			_, err := loader.LoadContent(ctx, file)
			if err != nil {
				t.Errorf("Concurrent load failed for file %d: %v", index, err)
			}
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < numFiles; i++ {
		<-done
	}
}

func TestLoadContent_ContextCancellation(t *testing.T) {
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, "/test/repo")
	
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	file := taskmaster.FileChange{
		Path:      "main.go",
		CommitID:  "abc123",
		IsPending: false,
	}
	
	// This should still attempt the operation
	// (Context cancellation is handled by git service, not loader)
	_, err := loader.LoadContent(ctx, file)
	
	// Expect an error since file isn't in mock service
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestLoadContent_Performance(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "content-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a reasonably sized file
	testContent := strings.Repeat("performance test content\n", 1000)
	relativePath := "perf-test.txt"
	createTempFile(t, tempDir, relativePath, testContent)
	
	gitService := newMockContentGitService()
	loader := NewContentLoader(gitService, tempDir)
	
	file := taskmaster.FileChange{
		Path:      relativePath,
		IsPending: true,
	}
	
	ctx := context.Background()
	
	// First load (cold cache)
	start := time.Now()
	_, err = loader.LoadContent(ctx, file)
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}
	firstLoadTime := time.Since(start)
	
	// Second load (warm cache)
	start = time.Now()
	_, err = loader.LoadContent(ctx, file)
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}
	secondLoadTime := time.Since(start)
	
	// Cache hit should be faster
	if secondLoadTime >= firstLoadTime {
		t.Logf("Warning: Cache hit not faster (first: %v, second: %v)", 
			firstLoadTime, secondLoadTime)
	}
}
