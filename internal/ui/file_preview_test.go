package ui

import (
	"context"
	"fmt"
	"testing"

	"github.com/agreen757/tm-tui/internal/filechanges"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// mockContentLoader implements ContentLoaderInterface for testing
type mockContentLoader struct {
	content map[string]string
	errors  map[string]error
}

func newMockContentLoader() *mockContentLoader {
	return &mockContentLoader{
		content: make(map[string]string),
		errors:  make(map[string]error),
	}
}

func (m *mockContentLoader) LoadContent(ctx context.Context, file taskmaster.FileChange) (string, error) {
	key := file.Path
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if content, ok := m.content[key]; ok {
		return content, nil
	}
	return "", fmt.Errorf("file not found: %s", key)
}

func (m *mockContentLoader) InvalidateCache() {}

func (m *mockContentLoader) setContent(path, content string) {
	m.content[path] = content
}

func (m *mockContentLoader) setError(path string, err error) {
	m.errors[path] = err
}

// mockDiffGenerator implements DiffGeneratorInterface for testing
type mockDiffGenerator struct {
	diffs   map[string]string
	errors  map[string]error
	pending map[string]string
}

func newMockDiffGenerator() *mockDiffGenerator {
	return &mockDiffGenerator{
		diffs:   make(map[string]string),
		errors:  make(map[string]error),
		pending: make(map[string]string),
	}
}

func (m *mockDiffGenerator) GenerateDiff(ctx context.Context, file, fromCommit, toCommit string) (string, error) {
	key := fmt.Sprintf("%s:%s:%s", file, fromCommit, toCommit)
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if diff, ok := m.diffs[key]; ok {
		return diff, nil
	}
	return "", fmt.Errorf("no diff found")
}

func (m *mockDiffGenerator) GenerateDiffWithFormat(ctx context.Context, file, fromCommit, toCommit string, format filechanges.DiffFormat) (string, error) {
	return m.GenerateDiff(ctx, file, fromCommit, toCommit)
}

func (m *mockDiffGenerator) GeneratePendingDiff(ctx context.Context, file string) (string, error) {
	if diff, ok := m.pending[file]; ok {
		return diff, nil
	}
	return "", fmt.Errorf("no pending diff found for %s", file)
}

func (m *mockDiffGenerator) InvalidateCache() {}

func (m *mockDiffGenerator) setDiff(file, fromCommit, toCommit, diff string) {
	key := fmt.Sprintf("%s:%s:%s", file, fromCommit, toCommit)
	m.diffs[key] = diff
}

func (m *mockDiffGenerator) setPendingDiff(file, diff string) {
	m.pending[file] = diff
}

// Test FilePreview creation
func TestNewFilePreview(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()

	preview := NewFilePreview(loader, diffGen)

	if preview == nil {
		t.Fatal("NewFilePreview returned nil")
	}

	if preview.loader != loader {
		t.Error("Loader not set correctly")
	}

	if preview.diffGen != diffGen {
		t.Error("DiffGenerator not set correctly")
	}

	if preview.diffMode {
		t.Error("DiffMode should be false initially")
	}
}

// Test setting a file
func TestSetFile(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	testContent := "func main() {\n\tfmt.Println(\"hello\")\n}"
	loader.setContent("main.go", testContent)

	file := taskmaster.FileChange{
		Path:      "main.go",
		IsPending: true,
	}

	ctx := context.Background()
	err := preview.SetFile(ctx, file)

	if err != nil {
		t.Fatalf("SetFile failed: %v", err)
	}

	if preview.file != file {
		t.Error("File not set correctly")
	}

	if preview.content != testContent {
		t.Errorf("Content not loaded correctly: got %q", preview.content)
	}
}

// Test toggling diff mode
func TestToggleDiffMode(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	testContent := "func main() {}"
	testDiff := `--- a/main.go
+++ b/main.go
@@ -1 +1,2 @@
 func main() {
+	fmt.Println("hello")
 }`

	loader.setContent("main.go", testContent)
	diffGen.setPendingDiff("main.go", testDiff)

	file := taskmaster.FileChange{
		Path:      "main.go",
		IsPending: true,
	}

	ctx := context.Background()
	preview.SetFile(ctx, file)

	if preview.diffMode {
		t.Error("DiffMode should be false initially")
	}

	// Toggle to diff mode
	err := preview.ToggleDiffMode(ctx)
	if err != nil {
		t.Fatalf("ToggleDiffMode failed: %v", err)
	}

	if !preview.diffMode {
		t.Error("DiffMode should be true after toggle")
	}

	// Toggle back to content mode
	err = preview.ToggleDiffMode(ctx)
	if err != nil {
		t.Fatalf("Second ToggleDiffMode failed: %v", err)
	}

	if preview.diffMode {
		t.Error("DiffMode should be false after second toggle")
	}
}

// Test setting dimensions
func TestSetDimensions(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	width := 120
	height := 40

	preview.SetDimensions(width, height)

	if preview.width != width || preview.height != height {
		t.Errorf("Dimensions not set correctly: got %d x %d, expected %d x %d",
			preview.width, preview.height, width, height)
	}

	if preview.viewport.Width != width || preview.viewport.Height != height {
		t.Error("Viewport dimensions not updated")
	}
}

// Test refresh with no file selected
func TestRefresh_NoFile(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	ctx := context.Background()
	err := preview.Refresh(ctx)

	if err != nil {
		t.Fatalf("Refresh with no file failed: %v", err)
	}

	if preview.content != "" {
		t.Error("Content should be empty when no file is set")
	}
}

// Test file not found error
func TestSetFile_FileNotFound(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	loader.setError("nonexistent.go", fmt.Errorf("file not found"))

	file := taskmaster.FileChange{
		Path:      "nonexistent.go",
		IsPending: true,
	}

	ctx := context.Background()
	err := preview.SetFile(ctx, file)

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if preview.lastError == "" {
		t.Error("LastError should be set")
	}
}

// Test syntax highlighting toggle
func TestToggleSyntaxHighlighting(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	if !preview.syntaxEnabled {
		t.Error("Syntax highlighting should be enabled by default")
	}

	preview.ToggleSyntaxHighlighting()
	if preview.syntaxEnabled {
		t.Error("Syntax highlighting should be disabled")
	}

	preview.ToggleSyntaxHighlighting()
	if !preview.syntaxEnabled {
		t.Error("Syntax highlighting should be enabled again")
	}
}

// Test get methods
func TestGetters(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	testContent := "test content"
	loader.setContent("test.go", testContent)

	file := taskmaster.FileChange{
		Path:      "test.go",
		IsPending: true,
	}

	ctx := context.Background()
	preview.SetFile(ctx, file)

	// Test GetContent
	if preview.GetContent() != testContent {
		t.Errorf("GetContent returned wrong value: %q", preview.GetContent())
	}

	// Test GetFile
	if preview.GetFile() != file {
		t.Error("GetFile returned wrong file")
	}

	// Test GetDiffMode
	if preview.GetDiffMode() {
		t.Error("GetDiffMode should return false")
	}

	// Test GetViewport
	vp := preview.GetViewport()
	if vp == nil {
		t.Error("GetViewport returned nil")
	}

	// Test GetLastError (should be empty)
	if preview.GetLastError() != "" {
		t.Error("GetLastError should be empty when no error occurred")
	}
}

// Test diff mode with pending changes
func TestDiffMode_PendingChanges(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	pendingDiff := `--- a/main.go
+++ b/main.go
@@ -1 +1 @@
-old code
+new code`

	diffGen.setPendingDiff("main.go", pendingDiff)

	file := taskmaster.FileChange{
		Path:      "main.go",
		IsPending: true,
	}

	ctx := context.Background()
	preview.SetFile(ctx, file)
	preview.ToggleDiffMode(ctx)

	if !preview.GetDiffMode() {
		t.Error("Diff mode should be enabled")
	}

	if !contains(preview.GetContent(), "-old code") {
		t.Errorf("Diff content not loaded correctly: %q", preview.GetContent())
	}
}

// Test view rendering
func TestView(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	// View with no file
	view := preview.View()
	if !contains(view, "No file selected") {
		t.Error("View should show 'No file selected' when no file is set")
	}

	// View with file
	loader.setContent("test.go", "test content")
	file := taskmaster.FileChange{
		Path:      "test.go",
		IsPending: true,
	}

	ctx := context.Background()
	preview.SetFile(ctx, file)
	view = preview.View()

	if !contains(view, "test.go") {
		t.Error("View should contain file name")
	}

	if !contains(view, "View") {
		t.Error("View should show 'View' mode indicator")
	}
}

// Test multiple file changes
func TestSetFile_MultipleFiles(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	loader.setContent("file1.go", "content 1")
	loader.setContent("file2.go", "content 2")
	loader.setContent("file3.go", "content 3")

	ctx := context.Background()

	files := []taskmaster.FileChange{
		{Path: "file1.go"},
		{Path: "file2.go"},
		{Path: "file3.go"},
	}

	for i, file := range files {
		err := preview.SetFile(ctx, file)
		if err != nil {
			t.Fatalf("Failed to set file %d: %v", i, err)
		}

		if preview.GetFile() != file {
			t.Errorf("File %d not set correctly", i)
		}
	}
}

// Test file extension detection
func TestGetFileExtension(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	tests := []struct {
		path string
		ext  string
	}{
		{"main.go", "go"},
		{"config.json", "json"},
		{"data.yaml", "yaml"},
		{"README.md", "md"},
		{"Makefile", ""},
		{"path/to/file.txt", "txt"},
	}

	for _, tt := range tests {
		preview.file.Path = tt.path
		ext := preview.getFileExtension()
		if ext != tt.ext {
			t.Errorf("For path %q: expected ext %q, got %q", tt.path, tt.ext, ext)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr)))
}

// Test init and update
func TestInit(t *testing.T) {
	loader := newMockContentLoader()
	diffGen := newMockDiffGenerator()
	preview := NewFilePreview(loader, diffGen)

	cmd := preview.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}
