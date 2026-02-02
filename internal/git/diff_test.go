package git

import (
	"context"
	"strings"
	"testing"
)

func TestParseDiff(t *testing.T) {
	tests := []struct {
		name           string
		rawDiff        string
		expectedOldPath string
		expectedNewPath string
		expectedIsBinary bool
		expectedIsNew   bool
		expectedIsDeleted bool
		expectedIsRenamed bool
		expectedHunks   int
		expectedAdditions int
		expectedDeletions int
	}{
		{
			name:    "Empty diff",
			rawDiff: "",
			expectedHunks: 0,
		},
		{
			name: "Simple unified diff",
			rawDiff: `diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main
 
+import "fmt"
+
 func main() {
-    println("hello")
+    fmt.Println("hello")
 }`,
			expectedOldPath: "main.go",
			expectedNewPath: "main.go",
			expectedHunks: 1,
			expectedAdditions: 3,
			expectedDeletions: 1,
		},
		{
			name: "New file",
			rawDiff: `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..abc123
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+package main
+
+func foo() {}`,
			expectedOldPath: "newfile.go",
			expectedNewPath: "newfile.go",
			expectedIsNew: true,
			expectedHunks: 1,
			expectedAdditions: 3,
		},
		{
			name: "Deleted file",
			rawDiff: `diff --git a/oldfile.go b/oldfile.go
deleted file mode 100644
index abc123..0000000
--- a/oldfile.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func bar() {}`,
			expectedOldPath: "oldfile.go",
			expectedNewPath: "oldfile.go",
			expectedIsDeleted: true,
			expectedHunks: 1,
			expectedDeletions: 3,
		},
		{
			name: "Renamed file",
			rawDiff: `diff --git a/old.go b/new.go
similarity index 100%
rename from old.go
rename to new.go`,
			expectedOldPath: "old.go",
			expectedNewPath: "new.go",
			expectedIsRenamed: true,
			expectedHunks: 0,
		},
		{
			name: "Binary file",
			rawDiff: `diff --git a/image.png b/image.png
index abc123..def456 100644
Binary files a/image.png and b/image.png differ`,
			expectedOldPath: "image.png",
			expectedNewPath: "image.png",
			expectedIsBinary: true,
			expectedHunks: 0,
		},
		{
			name: "Multiple hunks",
			rawDiff: `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
 
+import "fmt"
 func foo() {
@@ -10,5 +11,6 @@ func foo() {
     x := 1
+    y := 2
     return
 }`,
			expectedOldPath: "file.go",
			expectedNewPath: "file.go",
			expectedHunks: 2,
			expectedAdditions: 2,
		},
		{
			name: "File mode change",
			rawDiff: `diff --git a/script.sh b/script.sh
old mode 100644
new mode 100755
index abc123..def456`,
			expectedOldPath: "script.sh",
			expectedNewPath: "script.sh",
			expectedHunks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := ParseDiff(tt.rawDiff)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if diff.OldPath != tt.expectedOldPath {
				t.Errorf("Expected OldPath %q, got %q", tt.expectedOldPath, diff.OldPath)
			}
			if diff.NewPath != tt.expectedNewPath {
				t.Errorf("Expected NewPath %q, got %q", tt.expectedNewPath, diff.NewPath)
			}
			if diff.IsBinary != tt.expectedIsBinary {
				t.Errorf("Expected IsBinary %v, got %v", tt.expectedIsBinary, diff.IsBinary)
			}
			if diff.IsNew != tt.expectedIsNew {
				t.Errorf("Expected IsNew %v, got %v", tt.expectedIsNew, diff.IsNew)
			}
			if diff.IsDeleted != tt.expectedIsDeleted {
				t.Errorf("Expected IsDeleted %v, got %v", tt.expectedIsDeleted, diff.IsDeleted)
			}
			if diff.IsRenamed != tt.expectedIsRenamed {
				t.Errorf("Expected IsRenamed %v, got %v", tt.expectedIsRenamed, diff.IsRenamed)
			}
			if len(diff.Hunks) != tt.expectedHunks {
				t.Errorf("Expected %d hunks, got %d", tt.expectedHunks, len(diff.Hunks))
			}
			if diff.Additions != tt.expectedAdditions {
				t.Errorf("Expected %d additions, got %d", tt.expectedAdditions, diff.Additions)
			}
			if diff.Deletions != tt.expectedDeletions {
				t.Errorf("Expected %d deletions, got %d", tt.expectedDeletions, diff.Deletions)
			}
			if diff.RawDiff != tt.rawDiff {
				t.Error("RawDiff should be preserved")
			}
		})
	}
}

func TestParseDiffHunkDetails(t *testing.T) {
	rawDiff := `diff --git a/test.go b/test.go
index abc123..def456 100644
--- a/test.go
+++ b/test.go
@@ -10,7 +10,8 @@ func test() {
     line1
     line2
-    old line
+    new line
+    added line
     line3
     line4
`

	diff, err := ParseDiff(rawDiff)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}

	hunk := diff.Hunks[0]
	
	// Check hunk header
	if hunk.OldStart != 10 {
		t.Errorf("Expected OldStart 10, got %d", hunk.OldStart)
	}
	if hunk.OldCount != 7 {
		t.Errorf("Expected OldCount 7, got %d", hunk.OldCount)
	}
	if hunk.NewStart != 10 {
		t.Errorf("Expected NewStart 10, got %d", hunk.NewStart)
	}
	if hunk.NewCount != 8 {
		t.Errorf("Expected NewCount 8, got %d", hunk.NewCount)
	}

	// Check hunk has lines
	if len(hunk.Lines) == 0 {
		t.Error("Expected hunk to have lines")
	}

	// Verify some line content
	hasAddition := false
	hasDeletion := false
	for _, line := range hunk.Lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			hasAddition = true
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			hasDeletion = true
		}
	}

	if !hasAddition {
		t.Error("Expected to find addition lines")
	}
	if !hasDeletion {
		t.Error("Expected to find deletion lines")
	}
}

func TestGetFileDiffWithFormat(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		fromCommit    string
		toCommit      string
		format        DiffFormat
		mockOutput    string
		expectedCmd   string
		expectedError bool
	}{
		{
			name:        "Unified format (default)",
			file:        "main.go",
			fromCommit:  "abc123",
			toCommit:    "def456",
			format:      DiffFormatUnified,
			mockOutput:  "diff --git a/main.go b/main.go\n...",
			expectedCmd: "diff abc123..def456 -- main.go",
		},
		{
			name:        "Context format",
			file:        "main.go",
			fromCommit:  "abc123",
			toCommit:    "def456",
			format:      DiffFormatContext,
			mockOutput:  "*** main.go\n--- main.go\n...",
			expectedCmd: "diff --context=3 abc123..def456 -- main.go",
		},
		{
			name:        "Stat format",
			file:        "main.go",
			fromCommit:  "abc123",
			toCommit:    "def456",
			format:      DiffFormatStat,
			mockOutput:  " main.go | 5 +++--\n 1 file changed, 3 insertions(+), 2 deletions(-)",
			expectedCmd: "diff --stat abc123..def456 -- main.go",
		},
		{
			name:        "Name only format",
			file:        "main.go",
			fromCommit:  "abc123",
			toCommit:    "def456",
			format:      DiffFormatNameOnly,
			mockOutput:  "main.go",
			expectedCmd: "diff --name-only abc123..def456 -- main.go",
		},
		{
			name:          "Empty file path",
			file:          "",
			fromCommit:    "abc123",
			toCommit:      "def456",
			format:        DiffFormatUnified,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := NewMockCommandRunner()
			if tt.expectedCmd != "" {
				mockRunner.AddResponse(tt.expectedCmd, []byte(tt.mockOutput), nil)
			}

			service := NewGitServiceWithRunner("/test/repo", mockRunner)
			diff, err := service.GetFileDiffWithFormat(context.Background(), tt.file, tt.fromCommit, tt.toCommit, tt.format)

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

func TestFileContentCaching(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("show abc123:main.go", []byte("package main\n\nfunc main() {}"), nil)

	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newFileContentCache(10)

	// First call - should hit git
	content1, err := service.GetFileContentAtCommitCached(context.Background(), "abc123", "main.go", cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify content
	if !strings.Contains(content1, "package main") {
		t.Error("Expected content to contain 'package main'")
	}

	// Check that git was called once
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected 1 git call, got %d", len(callLog))
	}

	// Second call - should use cache
	content2, err := service.GetFileContentAtCommitCached(context.Background(), "abc123", "main.go", cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if content2 != content1 {
		t.Error("Cached content should match original")
	}

	// Should still be just 1 git call (cached)
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected still 1 git call (cached), got %d", len(callLog))
	}

	// Different file - should hit git again
	mockRunner.AddResponse("show abc123:other.go", []byte("package other"), nil)
	content3, err := service.GetFileContentAtCommitCached(context.Background(), "abc123", "other.go", cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(content3, "package other") {
		t.Error("Expected content to contain 'package other'")
	}

	// Should now be 2 git calls
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls, got %d", len(callLog))
	}
}

func TestFileContentCacheLRU(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newFileContentCache(2) // Small cache size

	// Add first file
	mockRunner.AddResponse("show abc123:file1.go", []byte("content1"), nil)
	_, err := service.GetFileContentAtCommitCached(context.Background(), "abc123", "file1.go", cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Add second file
	mockRunner.AddResponse("show abc123:file2.go", []byte("content2"), nil)
	_, err = service.GetFileContentAtCommitCached(context.Background(), "abc123", "file2.go", cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Add third file - should evict first (LRU)
	mockRunner.AddResponse("show abc123:file3.go", []byte("content3"), nil)
	_, err = service.GetFileContentAtCommitCached(context.Background(), "abc123", "file3.go", cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify cache size limit
	if len(cache.contents) > 2 {
		t.Errorf("Cache size should be limited to 2, got %d", len(cache.contents))
	}
}

func TestDiffCaching(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddResponse("diff abc123..def456 -- main.go", []byte("diff output"), nil)

	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newDiffCache(10)

	// First call - should hit git
	diff1, err := service.GetFileDiffWithFormatCached(context.Background(), "main.go", "abc123", "def456", DiffFormatUnified, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if diff1 != "diff output" {
		t.Errorf("Expected 'diff output', got %q", diff1)
	}

	// Check that git was called once
	callLog := mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected 1 git call, got %d", len(callLog))
	}

	// Second call - should use cache
	diff2, err := service.GetFileDiffWithFormatCached(context.Background(), "main.go", "abc123", "def456", DiffFormatUnified, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if diff2 != diff1 {
		t.Error("Cached diff should match original")
	}

	// Should still be just 1 git call (cached)
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected still 1 git call (cached), got %d", len(callLog))
	}

	// Different format - should hit git again
	mockRunner.AddResponse("diff --stat abc123..def456 -- main.go", []byte("stat output"), nil)
	diff3, err := service.GetFileDiffWithFormatCached(context.Background(), "main.go", "abc123", "def456", DiffFormatStat, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if diff3 != "stat output" {
		t.Errorf("Expected 'stat output', got %q", diff3)
	}

	// Should now be 2 git calls (different format)
	callLog = mockRunner.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 git calls, got %d", len(callLog))
	}
}

func TestDiffCacheLRU(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	service := NewGitServiceWithRunner("/test/repo", mockRunner)
	cache := newDiffCache(2) // Small cache size

	// Add first diff
	mockRunner.AddResponse("diff abc..def -- file1.go", []byte("diff1"), nil)
	_, err := service.GetFileDiffWithFormatCached(context.Background(), "file1.go", "abc", "def", DiffFormatUnified, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Add second diff
	mockRunner.AddResponse("diff abc..def -- file2.go", []byte("diff2"), nil)
	_, err = service.GetFileDiffWithFormatCached(context.Background(), "file2.go", "abc", "def", DiffFormatUnified, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Add third diff - should evict first (LRU)
	mockRunner.AddResponse("diff abc..def -- file3.go", []byte("diff3"), nil)
	_, err = service.GetFileDiffWithFormatCached(context.Background(), "file3.go", "abc", "def", DiffFormatUnified, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify cache size limit
	if len(cache.diffs) > 2 {
		t.Errorf("Cache size should be limited to 2, got %d", len(cache.diffs))
	}
}

func TestParseDiffEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		rawDiff string
	}{
		{
			name:    "Whitespace only",
			rawDiff: "   \n\t\n  ",
		},
		{
			name: "Incomplete hunk header",
			rawDiff: `diff --git a/test.go b/test.go
@@ incomplete`,
		},
		{
			name: "Malformed diff header",
			rawDiff: `diff --git a/test.go
index abc..def`,
		},
		{
			name: "Mixed line endings",
			rawDiff: "diff --git a/test.go b/test.go\r\n--- a/test.go\n+++ b/test.go\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic or return error
			diff, err := ParseDiff(tt.rawDiff)
			if err != nil {
				t.Errorf("Should handle edge case gracefully: %v", err)
			}
			if diff == nil {
				t.Error("Should return non-nil diff even for edge cases")
			}
		})
	}
}
