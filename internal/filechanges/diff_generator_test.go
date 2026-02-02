package filechanges

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// MockGitServiceForDiff implements GitServiceForDiff for testing
type mockGitServiceForDiff struct {
	diffs    map[string]string // Key: "file:fromCommit:toCommit"
	errors   map[string]error
	contents map[string]string // Key: "commit:file"
}

func newMockGitServiceForDiff() *mockGitServiceForDiff {
	return &mockGitServiceForDiff{
		diffs:    make(map[string]string),
		errors:   make(map[string]error),
		contents: make(map[string]string),
	}
}

func (m *mockGitServiceForDiff) GetFileDiff(ctx context.Context, file, fromCommit, toCommit string) (string, error) {
	key := fmt.Sprintf("%s:%s:%s", file, fromCommit, toCommit)
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if diff, ok := m.diffs[key]; ok {
		return diff, nil
	}
	return "", fmt.Errorf("no diff found for %s", key)
}

func (m *mockGitServiceForDiff) GetFileContentAtCommit(ctx context.Context, commit, file string) (string, error) {
	key := fmt.Sprintf("%s:%s", commit, file)
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if content, ok := m.contents[key]; ok {
		return content, nil
	}
	return "", fmt.Errorf("no content found for %s", key)
}

func (m *mockGitServiceForDiff) setDiff(file, fromCommit, toCommit, diff string) {
	key := fmt.Sprintf("%s:%s:%s", file, fromCommit, toCommit)
	m.diffs[key] = diff
}

func (m *mockGitServiceForDiff) setError(file, fromCommit, toCommit string, err error) {
	key := fmt.Sprintf("%s:%s:%s", file, fromCommit, toCommit)
	m.errors[key] = err
}

func (m *mockGitServiceForDiff) setContent(commit, file, content string) {
	key := fmt.Sprintf("%s:%s", commit, file)
	m.contents[key] = content
}

// Test DiffGenerator creation
func TestNewDiffGenerator(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	if gen == nil {
		t.Fatal("NewDiffGenerator returned nil")
	}

	if gen.gitService != gitService {
		t.Error("GitService not set correctly")
	}

	if gen.cache == nil {
		t.Error("Cache not initialized")
	}
}

// Test generating diff between commits
func TestGenerateDiff(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	testDiff := `--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 func main() {
+	fmt.Println("hello")
 }
`
	gitService.setDiff("main.go", "abc1234567", "def1234567", testDiff)

	ctx := context.Background()
	diff, err := gen.GenerateDiff(ctx, "main.go", "abc1234567", "def1234567")

	if err != nil {
		t.Fatalf("GenerateDiff failed: %v", err)
	}

	if diff != testDiff {
		t.Errorf("Expected diff %q, got %q", testDiff, diff)
	}
}

// Test diff with format option
func TestGenerateDiffWithFormat(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	testDiff := `main.go | 1 +
	1 file changed, 1 insertion(+)`
	gitService.setDiff("main.go", "abc1234567", "def1234567", testDiff)

	ctx := context.Background()
	diff, err := gen.GenerateDiffWithFormat(ctx, "main.go", "abc1234567", "def1234567", DiffFormatStat)

	if err != nil {
		t.Fatalf("GenerateDiffWithFormat failed: %v", err)
	}

	// Should contain the stat line
	if len(diff) == 0 {
		t.Error("Expected non-empty diff output")
	}
}

// Test diff format options
func TestDiffFormats(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	fullDiff := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 line 1
+line 2
 line 3
file.go | 1 +`

	gitService.setDiff("file.go", "abc1234567", "def1234567", fullDiff)

	ctx := context.Background()

	tests := []struct {
		name      string
		format    DiffFormat
		expectLen func(string) bool
	}{
		{
			name:      "unified format",
			format:    DiffFormatUnified,
			expectLen: func(s string) bool { return len(s) > 0 },
		},
		{
			name:      "stat format",
			format:    DiffFormatStat,
			expectLen: func(s string) bool { return strings.Contains(s, "file.go") || strings.Contains(s, "No changes") },
		},
		{
			name:      "nameonly format",
			format:    DiffFormatNameOnly,
			expectLen: func(s string) bool { return strings.Contains(s, "file.go") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := gen.GenerateDiffWithFormat(ctx, "file.go", "abc1234567", "def1234567", tt.format)
			if err != nil {
				t.Fatalf("GenerateDiffWithFormat failed: %v", err)
			}
			if !tt.expectLen(diff) {
				t.Errorf("Expected diff to match format, got: %q", diff)
			}
		})
	}
}

// Test empty file path error
func TestGenerateDiff_EmptyFilePath(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	ctx := context.Background()
	_, err := gen.GenerateDiff(ctx, "", "abc1234567", "def1234567")

	if err == nil {
		t.Error("Expected error for empty file path")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

// Test empty commit ID error
func TestGenerateDiff_EmptyCommitID(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	ctx := context.Background()
	_, err := gen.GenerateDiff(ctx, "main.go", "", "def1234567")

	if err == nil {
		t.Error("Expected error for empty fromCommit")
	}

	_, err = gen.GenerateDiff(ctx, "main.go", "abc1234567", "")
	if err == nil {
		t.Error("Expected error for empty toCommit")
	}
}

// Test invalid commit ID length
func TestGenerateDiff_InvalidCommitIDLength(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	ctx := context.Background()
	_, err := gen.GenerateDiff(ctx, "main.go", "abc123", "def1234567") // Too short

	if err == nil {
		t.Error("Expected error for short commit ID")
	}
	if !strings.Contains(err.Error(), "at least 7 characters") {
		t.Errorf("Expected '7 characters' error, got: %v", err)
	}
}

// Test git service error
func TestGenerateDiff_GitServiceError(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	gitService.setError("main.go", "abc1234567", "def1234567", fmt.Errorf("git command failed"))

	ctx := context.Background()
	_, err := gen.GenerateDiff(ctx, "main.go", "abc1234567", "def1234567")

	if err == nil {
		t.Error("Expected git service error")
	}
	if !strings.Contains(err.Error(), "failed to generate diff") {
		t.Errorf("Expected 'failed to generate diff' error, got: %v", err)
	}
}

// Test caching mechanism
func TestDiffGenerator_Caching(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	testDiff := "diff content"
	gitService.setDiff("main.go", "abc1234567", "def1234567", testDiff)

	ctx := context.Background()

	// First call should fetch from git service
	diff1, err := gen.GenerateDiff(ctx, "main.go", "abc1234567", "def1234567")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Remove from git service to verify cache is used
	gitService.diffs = make(map[string]string)

	// Second call should use cache
	diff2, err := gen.GenerateDiff(ctx, "main.go", "abc1234567", "def1234567")
	if err != nil {
		t.Fatalf("Second call (from cache) failed: %v", err)
	}

	if diff1 != diff2 {
		t.Errorf("Cached diff differs from first call: %q vs %q", diff1, diff2)
	}
}

// Test cache invalidation
func TestDiffGenerator_InvalidateCache(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	testDiff := "diff content"
	gitService.setDiff("main.go", "abc1234567", "def1234567", testDiff)

	ctx := context.Background()

	// Populate cache
	_, err := gen.GenerateDiff(ctx, "main.go", "abc1234567", "def1234567")
	if err != nil {
		t.Fatalf("Initial call failed: %v", err)
	}

	// Invalidate cache
	gen.InvalidateCache()

	// Remove from git service and try again
	gitService.diffs = make(map[string]string)

	_, err = gen.GenerateDiff(ctx, "main.go", "abc1234567", "def1234567")
	if err == nil {
		t.Error("Expected error after cache invalidation")
	}
}

// Test pending diff generation
func TestGeneratePendingDiff(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	pendingDiff := `--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 func main() {
+	// new code
 }
`
	// Use HEAD:0 as a marker for working directory changes
	gitService.setDiff("main.go", "HEAD", "HEAD:0", pendingDiff)

	ctx := context.Background()
	diff, err := gen.GeneratePendingDiff(ctx, "main.go")

	if err != nil {
		t.Fatalf("GeneratePendingDiff failed: %v", err)
	}

	if diff != pendingDiff {
		t.Errorf("Expected pending diff, got: %q", diff)
	}
}

// Test pending diff with empty file path
func TestGeneratePendingDiff_EmptyFilePath(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	ctx := context.Background()
	_, err := gen.GeneratePendingDiff(ctx, "")

	if err == nil {
		t.Error("Expected error for empty file path")
	}
}

// Test multiple file diffs with different formats
func TestMultipleDiffs(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	files := []string{"main.go", "config.json", "test.go"}
	for _, file := range files {
		diff := fmt.Sprintf("diff for %s", file)
		gitService.setDiff(file, "abc1234567", "def1234567", diff)
	}

	ctx := context.Background()
	for _, file := range files {
		diff, err := gen.GenerateDiff(ctx, file, "abc1234567", "def1234567")
		if err != nil {
			t.Fatalf("Failed to get diff for %s: %v", file, err)
		}
		if !strings.Contains(diff, file) {
			t.Errorf("Expected diff for %s, got: %q", file, diff)
		}
	}
}

// Test diff with special characters in path
func TestDiffWithSpecialPath(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	specialPath := "src/package/my-file.go"
	testDiff := "special diff content"
	gitService.setDiff(specialPath, "abc1234567", "def1234567", testDiff)

	ctx := context.Background()
	diff, err := gen.GenerateDiff(ctx, specialPath, "abc1234567", "def1234567")

	if err != nil {
		t.Fatalf("GenerateDiff with special path failed: %v", err)
	}

	if diff != testDiff {
		t.Errorf("Expected %q, got %q", testDiff, diff)
	}
}

// Test large file diff (performance)
func TestLargeFileDiff(t *testing.T) {
	gitService := newMockGitServiceForDiff()
	gen := NewDiffGenerator(gitService)

	// Simulate a large diff
	largeDiff := ""
	for i := 0; i < 1000; i++ {
		largeDiff += fmt.Sprintf("+line %d\n", i)
	}
	gitService.setDiff("large.go", "abc1234567", "def1234567", largeDiff)

	ctx := context.Background()
	diff, err := gen.GenerateDiff(ctx, "large.go", "abc1234567", "def1234567")

	if err != nil {
		t.Fatalf("Large diff generation failed: %v", err)
	}

	if len(diff) < 5000 {
		t.Error("Expected large diff content")
	}
}
