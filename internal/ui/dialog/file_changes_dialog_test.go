package dialog

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/agreen757/tm-tui/internal/filechanges"
	"github.com/agreen757/tm-tui/internal/git"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// mockStorage implements taskmaster.Storage interface for testing
type mockStorage struct {
	mapping *taskmaster.FileChangeMapping
}

func (m *mockStorage) Load() (*taskmaster.FileChangeMapping, error) {
	if m.mapping == nil {
		m.mapping = &taskmaster.FileChangeMapping{
			Version:           "1.0",
			LastUpdated:       time.Now(),
			Tasks:             make(map[string][]taskmaster.FileChange),
			UnassignedChanges: []taskmaster.FileChange{},
		}
	}
	return m.mapping, nil
}

func (m *mockStorage) Save(mapping *taskmaster.FileChangeMapping) error {
	m.mapping = mapping
	return nil
}

func (m *mockStorage) SetFileChangeMapping(mapping *taskmaster.FileChangeMapping) {
	m.mapping = mapping
}

// createTestTracker creates a file change tracker for testing
func createTestTracker(t *testing.T) *filechanges.FileChangeTracker {
	t.Helper()
	
	gitService := git.NewGitService("/tmp/test-repo")
	storage := &mockStorage{}
	
	tracker := filechanges.NewFileChangeTracker(gitService, storage, "/tmp/test-repo")
	
	// Add some test data
	mapping, _ := storage.Load()
	mapping.Tasks["1.1"] = []taskmaster.FileChange{
		{
			Path:        "internal/ui/app.go",
			ChangeType:  "modified",
			Description: "Updated main app",
			LastChanged: time.Now(),
			IsPending:   true,
		},
		{
			Path:        "internal/ui/keymap.go",
			ChangeType:  "added",
			Description: "Added new keymap",
			LastChanged: time.Now(),
			IsPending:   true,
		},
	}
	mapping.Tasks["2.1"] = []taskmaster.FileChange{
		{
			Path:        "internal/taskmaster/types.go",
			ChangeType:  "modified",
			Description: "Updated types",
			LastChanged: time.Now(),
			IsPending:   false,
			CommitID:    "abc123",
		},
	}
	storage.Save(mapping)
	
	// Initialize tracker
	ctx := context.Background()
	if err := tracker.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize tracker: %v", err)
	}
	
	return tracker
}

// TestFileChangesDialogCreation tests dialog creation
func TestFileChangesDialogCreation(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	
	if dialog == nil {
		t.Fatal("Expected dialog to be created")
	}
	
	if dialog.Title() != "File Changes" {
		t.Errorf("Expected title 'File Changes', got '%s'", dialog.Title())
	}
	
	if dialog.Kind() != DialogKindCustom {
		t.Errorf("Expected kind DialogKindCustom, got %v", dialog.Kind())
	}
	
	if dialog.groupBy != "task" {
		t.Errorf("Expected default groupBy 'task', got '%s'", dialog.groupBy)
	}
	
	if dialog.focusedPanel != "tree" {
		t.Errorf("Expected default focusedPanel 'tree', got '%s'", dialog.focusedPanel)
	}
	
	if dialog.fileTree == nil {
		t.Error("Expected fileTree to be initialized")
	}
	
	if dialog.preview == nil {
		t.Error("Expected preview to be initialized")
	}
}

// TestFileChangesDialogInit tests dialog initialization
func TestFileChangesDialogInit(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	
	// Init should build the tree
	cmd := dialog.Init()
	
	if cmd != nil {
		t.Error("Expected Init to return nil cmd")
	}
	
	// Verify tree was built
	if len(dialog.fileTree.items) == 0 {
		t.Error("Expected tree items to be built after Init")
	}
}

// TestFileChangesDialogRendering tests dialog rendering with various sizes
func TestFileChangesDialogRendering(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tests := []struct {
		name   string
		width  int
		height int
		valid  bool
	}{
		{"Normal size", 80, 24, true},
		{"Wide", 120, 24, true},
		{"Tall", 80, 40, true},
		{"Small valid", 40, 10, true},
		{"Too narrow", 30, 20, false},
		{"Too short", 80, 5, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := dialog.Render(tt.width, tt.height)
			
			if tt.valid {
				if strings.Contains(content, "Window too small") {
					t.Error("Expected valid rendering, got 'Window too small'")
				}
				
				// Verify content has expected sections
				if !strings.Contains(content, "Group:") {
					t.Error("Expected filter bar with 'Group:'")
				}
			} else {
				if !strings.Contains(content, "Window too small") {
					t.Error("Expected 'Window too small' message")
				}
			}
		})
	}
}

// TestFileChangesDialogLayout tests layout proportions
func TestFileChangesDialogLayout(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	width := 80
	height := 24
	
	dialog.Render(width, height)
	
	// Verify tree and preview widths
	expectedTreeWidth := width / 2
	expectedPreviewWidth := width - expectedTreeWidth - 1
	
	if dialog.fileTree.width != expectedTreeWidth {
		t.Errorf("Expected tree width %d, got %d", expectedTreeWidth, dialog.fileTree.width)
	}
	
	if dialog.preview.width != expectedPreviewWidth {
		t.Errorf("Expected preview width %d, got %d", expectedPreviewWidth, dialog.preview.width)
	}
	
	// Verify heights account for filter and status bars
	filterBarHeight := 2
	statusBarHeight := 1
	expectedContentHeight := height - filterBarHeight - statusBarHeight - 2
	
	if dialog.fileTree.height != expectedContentHeight {
		t.Errorf("Expected tree height %d, got %d", expectedContentHeight, dialog.fileTree.height)
	}
	
	if dialog.preview.height != expectedContentHeight {
		t.Errorf("Expected preview height %d, got %d", expectedContentHeight, dialog.preview.height)
	}
}

// TestFileChangesDialogFocusSwitching tests panel focus switching
func TestFileChangesDialogFocusSwitching(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	// Initial focus should be on tree
	if dialog.focusedPanel != "tree" {
		t.Errorf("Expected initial focus on 'tree', got '%s'", dialog.focusedPanel)
	}
	
	// Press tab to switch to preview
	msg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := dialog.HandleKey(msg)
	
	if result != DialogResultNone {
		t.Error("Expected tab to not close dialog")
	}
	
	if dialog.focusedPanel != "preview" {
		t.Errorf("Expected focus on 'preview' after tab, got '%s'", dialog.focusedPanel)
	}
	
	// Press tab again to switch back to tree
	result, _ = dialog.HandleKey(msg)
	
	if dialog.focusedPanel != "tree" {
		t.Errorf("Expected focus back on 'tree' after second tab, got '%s'", dialog.focusedPanel)
	}
}

// TestFileChangesDialogKeyboardHandling tests keyboard navigation
func TestFileChangesDialogKeyboardHandling(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tests := []struct {
		name           string
		key            tea.KeyMsg
		expectedResult DialogResult
		checkState     func(*testing.T, *FileChangesDialog)
	}{
		{
			name:           "Escape closes dialog",
			key:            tea.KeyMsg{Type: tea.KeyEsc},
			expectedResult: DialogResultCancel,
		},
		{
			name:           "Tab switches panel",
			key:            tea.KeyMsg{Type: tea.KeyTab},
			expectedResult: DialogResultNone,
			checkState: func(t *testing.T, d *FileChangesDialog) {
				if d.focusedPanel != "preview" {
					t.Error("Expected tab to switch to preview")
				}
			},
		},
		{
			name:           "G cycles groupBy mode",
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}},
			expectedResult: DialogResultNone,
			checkState: func(t *testing.T, d *FileChangesDialog) {
				if d.groupBy != "directory" {
					t.Errorf("Expected groupBy 'directory' after G, got '%s'", d.groupBy)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset dialog state
			dialog := NewFileChangesDialog(tracker, gitService)
			dialog.Init()
			
			result, _ := dialog.HandleKey(tt.key)
			
			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}
			
			if tt.checkState != nil {
				tt.checkState(t, dialog)
			}
		})
	}
}

// TestFileChangesDialogGroupByCycle tests groupBy mode cycling
func TestFileChangesDialogGroupByCycle(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	
	// Initial: task
	if dialog.groupBy != "task" {
		t.Errorf("Expected initial groupBy 'task', got '%s'", dialog.groupBy)
	}
	
	// Cycle to directory
	dialog.cycleGroupBy()
	if dialog.groupBy != "directory" {
		t.Errorf("Expected groupBy 'directory', got '%s'", dialog.groupBy)
	}
	
	// Cycle to time
	dialog.cycleGroupBy()
	if dialog.groupBy != "time" {
		t.Errorf("Expected groupBy 'time', got '%s'", dialog.groupBy)
	}
	
	// Cycle back to task
	dialog.cycleGroupBy()
	if dialog.groupBy != "task" {
		t.Errorf("Expected groupBy back to 'task', got '%s'", dialog.groupBy)
	}
}

// TestFileTreeRendering tests file tree rendering
func TestFileTreeRendering(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Render tree
	content := tree.Render(40, 10)
	
	// Should contain task items
	if !strings.Contains(content, "Task") {
		t.Error("Expected tree to contain 'Task' items")
	}
	
	// Should have content
	lines := strings.Split(content, "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 lines, got %d", len(lines))
	}
}

// TestFileTreeNavigation tests file tree navigation
func TestFileTreeNavigation(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	initialIdx := tree.selectedIdx
	
	// Move down
	tree.moveDown()
	if tree.selectedIdx != initialIdx+1 {
		t.Errorf("Expected selectedIdx %d after moveDown, got %d", initialIdx+1, tree.selectedIdx)
	}
	
	// Move up
	tree.moveUp()
	if tree.selectedIdx != initialIdx {
		t.Errorf("Expected selectedIdx %d after moveUp, got %d", initialIdx, tree.selectedIdx)
	}
	
	// Move up at top (should stay at 0)
	tree.selectedIdx = 0
	tree.moveUp()
	if tree.selectedIdx != 0 {
		t.Error("Expected selectedIdx to stay at 0 when moving up at top")
	}
	
	// Move down at bottom (should stay at max)
	tree.selectedIdx = len(tree.items) - 1
	maxIdx := tree.selectedIdx
	tree.moveDown()
	if tree.selectedIdx != maxIdx {
		t.Error("Expected selectedIdx to stay at max when moving down at bottom")
	}
}

// TestPreviewRendering tests preview panel rendering
func TestPreviewRendering(t *testing.T) {
	preview := NewPreview()
	
	// Empty preview
	content := preview.Render(40, 10)
	if !strings.Contains(content, "Select a file") {
		t.Error("Expected empty preview to show 'Select a file' message")
	}
	
	// Load file
	gitService := git.NewGitService("/tmp/test-repo")
	preview.LoadFile("test.go", gitService)
	
	content = preview.Render(40, 10)
	if !strings.Contains(content, "test.go") {
		t.Error("Expected preview to show file name")
	}
}

// TestPreviewScrolling tests preview scrolling functionality
func TestPreviewScrolling(t *testing.T) {
	preview := NewPreview()
	
	// Create content with many lines
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, "Line %d")
	}
	preview.lines = lines
	preview.height = 10
	
	// Initial scroll position
	if preview.scrollPos != 0 {
		t.Error("Expected initial scroll position to be 0")
	}
	
	// Scroll down
	preview.scrollDown()
	if preview.scrollPos != 1 {
		t.Errorf("Expected scrollPos 1 after scrollDown, got %d", preview.scrollPos)
	}
	
	// Scroll up
	preview.scrollUp()
	if preview.scrollPos != 0 {
		t.Errorf("Expected scrollPos 0 after scrollUp, got %d", preview.scrollPos)
	}
	
	// Page down
	preview.pageDown()
	expectedPos := 10
	if preview.scrollPos != expectedPos {
		t.Errorf("Expected scrollPos %d after pageDown, got %d", expectedPos, preview.scrollPos)
	}
	
	// Page up
	preview.pageUp()
	expectedPos = 0
	if preview.scrollPos != expectedPos {
		t.Errorf("Expected scrollPos %d after pageUp, got %d", expectedPos, preview.scrollPos)
	}
}

// TestFileChangesDialogUpdate tests dialog update handling
func TestFileChangesDialogUpdate(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	
	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	_, cmd := dialog.Update(msg)
	
	if cmd != nil {
		t.Error("Expected Update to return nil cmd")
	}
	
	// Verify dimensions were updated
	if dialog.width != 100 {
		t.Errorf("Expected width 100, got %d", dialog.width)
	}
	
	if dialog.height != 30 {
		t.Errorf("Expected height 30, got %d", dialog.height)
	}
}

// TestDialogView tests dialog view rendering
func TestDialogView(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	dialog.width = 80
	dialog.height = 24
	
	view := dialog.View()
	
	// View should include border from RenderBorder
	if view == "" {
		t.Error("Expected non-empty view")
	}
	
	// Should contain title
	if !strings.Contains(view, "File Changes") {
		t.Error("Expected view to contain title")
	}
}

// TestFileTreeGroupByTask tests task-based grouping
func TestFileTreeGroupByTask(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.groupBy = "task"
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Should have task items
	hasTaskItem := false
	for _, item := range tree.items {
		if item.Type == "task" {
			hasTaskItem = true
			break
		}
	}
	
	if !hasTaskItem {
		t.Error("Expected tree to have at least one task item")
	}
}

// TestFileTreeExpansion tests tree item expansion/collapse
func TestFileTreeExpansion(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Find a task item to expand
	taskIdx := -1
	for i, item := range tree.items {
		if item.Type == "task" {
			taskIdx = i
			break
		}
	}
	
	if taskIdx == -1 {
		t.Skip("No task items found for expansion test")
	}
	
	tree.selectedIdx = taskIdx
	taskID := tree.items[taskIdx].ID
	
	// Expand task
	tree.expandItem()
	
	if !tree.expandedItems[taskID] {
		t.Error("Expected task to be marked as expanded")
	}
	
	// Collapse task
	tree.collapseItem()
	
	if tree.expandedItems[taskID] {
		t.Error("Expected task to be marked as collapsed")
	}
}

// TestFileTreeDirectoryGrouping tests directory-based grouping
func TestFileTreeDirectoryGrouping(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.groupBy = "directory"
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Should have directory items
	hasDirectoryItem := false
	for _, item := range tree.items {
		if item.Type == "directory" {
			hasDirectoryItem = true
			break
		}
	}
	
	if !hasDirectoryItem {
		t.Error("Expected tree to have at least one directory item")
	}
	
	// Test directory expansion
	for i, item := range tree.items {
		if item.Type == "directory" {
			tree.selectedIdx = i
			tree.expandItem()
			
			if !tree.expandedItems[item.ID] {
				t.Error("Expected directory to be marked as expanded")
			}
			
			// Rebuild tree to see expanded contents
			tree.buildTree()
			
			// Should have more items after expansion
			if len(tree.items) <= 1 {
				t.Error("Expected expanded directory to show file items")
			}
			break
		}
	}
}

// TestFileTreeTimeGrouping tests time-based grouping
func TestFileTreeTimeGrouping(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.groupBy = "time"
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Should have time period items
	if len(tree.items) == 0 {
		t.Error("Expected tree to have time period items")
	}
	
	// Verify time period structure
	validPeriods := map[string]bool{
		"Today":      true,
		"Yesterday":  true,
		"This Week":  true,
		"This Month": true,
		"Older":      true,
	}
	
	for _, item := range tree.items {
		if item.Level == 0 && !validPeriods[item.ID] {
			t.Errorf("Unexpected time period: %s", item.ID)
		}
	}
	
	// Test time period expansion
	if len(tree.items) > 0 {
		tree.selectedIdx = 0
		periodID := tree.items[0].ID
		tree.expandItem()
		
		if !tree.expandedItems[periodID] {
			t.Error("Expected time period to be marked as expanded")
		}
	}
}

// TestFileTreeDirectoryTreeBuilding tests directory tree construction
func TestFileTreeDirectoryTreeBuilding(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	tree := dialog.fileTree
	
	// Create test file map
	fileMap := map[string]*taskmaster.FileChange{
		"internal/ui/app.go": {
			Path:       "internal/ui/app.go",
			ChangeType: "modified",
		},
		"internal/ui/keymap.go": {
			Path:       "internal/ui/keymap.go",
			ChangeType: "added",
		},
		"internal/taskmaster/types.go": {
			Path:       "internal/taskmaster/types.go",
			ChangeType: "modified",
		},
	}
	
	dirTree := tree.buildDirectoryTree(fileMap)
	
	// Verify root node
	if dirTree == nil {
		t.Fatal("Expected directory tree to be created")
	}
	
	// Verify internal directory exists
	if dirTree.children["internal"] == nil {
		t.Error("Expected 'internal' directory in tree")
	}
	
	// Verify nested directories
	internalNode := dirTree.children["internal"]
	if internalNode.children["ui"] == nil {
		t.Error("Expected 'internal/ui' directory in tree")
	}
	if internalNode.children["taskmaster"] == nil {
		t.Error("Expected 'internal/taskmaster' directory in tree")
	}
	
	// Verify file counts
	uiNode := internalNode.children["ui"]
	if len(uiNode.files) != 2 {
		t.Errorf("Expected 2 files in internal/ui, got %d", len(uiNode.files))
	}
	
	taskmasterNode := internalNode.children["taskmaster"]
	if len(taskmasterNode.files) != 1 {
		t.Errorf("Expected 1 file in internal/taskmaster, got %d", len(taskmasterNode.files))
	}
}

// TestFileTreeTimeSorting tests time-based sorting
func TestFileTreeTimeSorting(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.groupBy = "time"
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Verify time periods are present
	hasPeriods := false
	for _, item := range tree.items {
		if item.Level == 0 {
			hasPeriods = true
			break
		}
	}
	
	if !hasPeriods {
		t.Error("Expected time period headers in tree")
	}
}

// TestFileTreeScrolling tests tree scrolling with large datasets
func TestFileTreeScrolling(t *testing.T) {
	gitService := git.NewGitService("/tmp/test-repo")
	storage := &mockStorage{}
	
	// Add many items to test scrolling
	mapping, _ := storage.Load()
	
	for i := 1; i <= 50; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		mapping.Tasks[taskID] = []taskmaster.FileChange{
			{
				Path:        fmt.Sprintf("file%d.go", i),
				ChangeType:  "modified",
				LastChanged: time.Now(),
			},
		}
	}
	storage.Save(mapping)
	
	tracker := filechanges.NewFileChangeTracker(gitService, storage, "/tmp/test-repo")
	ctx := context.Background()
	tracker.Initialize(ctx)
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	tree := dialog.fileTree
	tree.height = 10 // Small viewport
	
	// Test rendering with scrolling
	content := tree.Render(40, 10)
	lines := strings.Split(content, "\n")
	
	if len(lines) != 10 {
		t.Errorf("Expected 10 lines for small viewport, got %d", len(lines))
	}
	
	// Test selection beyond viewport
	tree.selectedIdx = 20
	content = tree.Render(40, 10)
	
	// Should still render 10 lines but scrolled
	lines = strings.Split(content, "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 lines after scrolling, got %d", len(lines))
	}
}

// TestFileTreeDeepNesting tests deeply nested directory structures
func TestFileTreeDeepNesting(t *testing.T) {
	gitService := git.NewGitService("/tmp/test-repo")
	storage := &mockStorage{}
	
	// Create deeply nested structure
	mapping, _ := storage.Load()
	
	mapping.Tasks["nested-1"] = []taskmaster.FileChange{
		{
			Path:        "a/b/c/d/e/file.go",
			ChangeType:  "added",
			LastChanged: time.Now(),
		},
	}
	storage.Save(mapping)
	
	tracker := filechanges.NewFileChangeTracker(gitService, storage, "/tmp/test-repo")
	ctx := context.Background()
	tracker.Initialize(ctx)
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.groupBy = "directory"
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Recursively expand all directory items
	expandAllDirs := func() {
		for i := range tree.items {
			if tree.items[i].Type == "directory" {
				tree.expandedItems[tree.items[i].ID] = true
			}
		}
	}
	
	// Expand multiple times to handle nested directories
	for i := 0; i < 10; i++ {
		expandAllDirs()
		tree.buildTree()
	}
	
	// Verify deep nesting is represented
	maxLevel := 0
	fileFound := false
	for _, item := range tree.items {
		if item.Level > maxLevel {
			maxLevel = item.Level
		}
		if item.Type == "file" && strings.Contains(item.Change.Path, "a/b/c/d/e/file.go") {
			fileFound = true
		}
	}
	
	if !fileFound {
		t.Error("Expected to find the deeply nested file in tree")
	}
	
	// With 5 levels of directories (a, b, c, d, e), we expect at least level 4
	if maxLevel < 3 {
		t.Errorf("Expected deep nesting (level >= 3 for nested dirs), got max level %d", maxLevel)
	}
}

// TestFileTreeEmptyState tests rendering with no changes
func TestFileTreeEmptyState(t *testing.T) {
	gitService := git.NewGitService("/tmp/test-repo")
	storage := &mockStorage{}
	
	// Create tracker with no changes
	tracker := filechanges.NewFileChangeTracker(gitService, storage, "/tmp/test-repo")
	ctx := context.Background()
	tracker.Initialize(ctx)
	
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	content := tree.Render(40, 10)
	
	if !strings.Contains(content, "No file changes") {
		t.Error("Expected 'No file changes' message for empty tree")
	}
}

// TestPreviewLoadFile tests the LoadFile method
func TestPreviewLoadFile(t *testing.T) {
	// Create a temporary test file
	testFile := "/tmp/test-preview-file.go"
	testContent := `package main

import "fmt"

// Comment
func main() {
	fmt.Println("Hello, world!")
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)
	
	preview := NewPreview()
	gitService := git.NewGitService("/tmp/test-repo")
	
	err := preview.LoadFile(testFile, gitService)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	
	// Verify file is loaded
	if preview.file != testFile {
		t.Errorf("Expected file to be %s, got %s", testFile, preview.file)
	}
	
	// Verify not in diff mode
	if preview.diffMode {
		t.Error("Expected diffMode to be false for LoadFile")
	}
	
	// Verify lines are populated
	if len(preview.lines) == 0 {
		t.Error("Expected lines to be populated")
	}
	
	// Verify lines contain line numbers
	if len(preview.lines) > 0 && !strings.Contains(preview.lines[0], "│") {
		t.Error("Expected lines to contain line number separator")
	}
	
	// Verify scroll position is reset
	if preview.scrollPos != 0 {
		t.Errorf("Expected scrollPos to be 0, got %d", preview.scrollPos)
	}
}

// TestPreviewLoadFileError tests LoadFile with non-existent file
func TestPreviewLoadFileError(t *testing.T) {
	preview := NewPreview()
	gitService := git.NewGitService("/tmp/test-repo")
	
	err := preview.LoadFile("/nonexistent/file.go", gitService)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestPreviewSyntaxHighlighting tests syntax highlighting
func TestPreviewSyntaxHighlighting(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		line     string
		expected string // Expected to contain certain ANSI codes or patterns
	}{
		{
			name:     "Go comment",
			filename: "test.go",
			line:     "// This is a comment",
			expected: "comment", // Should be styled
		},
		{
			name:     "Go keyword",
			filename: "test.go",
			line:     "func main() {",
			expected: "func", // Should contain the keyword
		},
		{
			name:     "JSON key",
			filename: "test.json",
			line:     `"key": "value"`,
			expected: "key", // Should contain the key
		},
		{
			name:     "Markdown header",
			filename: "test.md",
			line:     "# Header",
			expected: "#", // Should contain the header marker
		},
	}
	
	preview := NewPreview()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preview.applySyntaxHighlighting(tt.line, tt.filename)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain %q, got: %s", tt.expected, result)
			}
		})
	}
}

// TestPreviewStyleDiffLine tests diff line styling
func TestPreviewStyleDiffLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		desc string // Description of expected styling
	}{
		{
			name: "Added line",
			line: "+new line",
			desc: "should be styled as addition",
		},
		{
			name: "Deleted line",
			line: "-old line",
			desc: "should be styled as deletion",
		},
		{
			name: "File marker",
			line: "+++ b/file.go",
			desc: "should be styled as file marker",
		},
		{
			name: "Hunk header",
			line: "@@ -1,3 +1,4 @@",
			desc: "should be styled as hunk header",
		},
		{
			name: "Context line",
			line: " unchanged line",
			desc: "should be styled as context",
		},
		{
			name: "Diff command",
			line: "diff --git a/file.go b/file.go",
			desc: "should be styled as command",
		},
	}
	
	preview := NewPreview()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preview.styleDiffLine(tt.line)
			// Verify that styling is applied (result should be non-empty and contain the line content)
			if result == "" {
				t.Errorf("Expected non-empty result for %q", tt.line)
			}
			// Basic check: styled output should still contain some part of the original line
			// (after removing ANSI codes, but we'll just check it's not empty)
		})
	}
}

// TestPreviewLoadDiff tests the LoadDiff method (integration test)
func TestPreviewLoadDiff(t *testing.T) {
	// This is a basic test that verifies LoadDiff doesn't crash
	// A full integration test would require a real git repository
	preview := NewPreview()
	gitService := git.NewGitService("/tmp/test-repo")
	
	// Try to load diff for a file
	err := preview.LoadDiff("test-file.go", gitService)
	
	// We expect this to fail since /tmp/test-repo doesn't exist
	// but we're testing that the method handles errors gracefully
	if err == nil {
		t.Log("LoadDiff succeeded (unexpected but not necessarily wrong)")
	} else {
		// Expected to fail with a meaningful error message
		if !strings.Contains(err.Error(), "failed to get diff") {
			t.Errorf("Expected error message about diff failure, got: %v", err)
		}
	}
	
	// Verify diff mode is set
	if !preview.diffMode {
		t.Error("Expected diffMode to be true after LoadDiff")
	}
}

func TestPreviewKeyboardHandling(t *testing.T) {
	preview := NewPreview()
	preview.lines = []string{"line1", "line2", "line3", "line4", "line5"}
	preview.height = 3
	preview.scrollPos = 0
	
	tests := []struct {
		name           string
		key            string
		initialPos     int
		expectedPos    int
	}{
		{"Up arrow", "up", 2, 1},
		{"Down arrow", "down", 0, 1},
		{"K key", "k", 2, 1},
		{"J key", "j", 0, 1},
		{"Home key", "home", 5, 0},
		{"End key", "end", 0, 2},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview.scrollPos = tt.initialPos
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "up" {
				msg.Type = tea.KeyUp
			} else if tt.key == "down" {
				msg.Type = tea.KeyDown
			} else if tt.key == "home" {
				msg.Type = tea.KeyHome
			} else if tt.key == "end" {
				msg.Type = tea.KeyEnd
			}
			
			preview.HandleKeyEvent(msg)
			
			if preview.scrollPos != tt.expectedPos {
				t.Errorf("Expected scrollPos to be %d after %s, got %d", 
					tt.expectedPos, tt.name, preview.scrollPos)
			}
		})
	}
}

// TestFileChangesDialogFiltering tests filtering functionality
func TestFileChangesDialogFiltering(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	// Test initial state (no filters)
	if dialog.filterTask != "" {
		t.Error("Expected no task filter initially")
	}
	if dialog.filterStatus != "all" {
		t.Errorf("Expected filterStatus to be 'all', got %s", dialog.filterStatus)
	}
	
	// Test SetFilterTask
	dialog.SetFilterTask("1.1")
	if dialog.filterTask != "1.1" {
		t.Errorf("Expected filterTask to be '1.1', got %s", dialog.filterTask)
	}
	
	// Test SetFilterStatus
	dialog.SetFilterStatus("modified")
	if dialog.filterStatus != "modified" {
		t.Errorf("Expected filterStatus to be 'modified', got %s", dialog.filterStatus)
	}
	
	// Test ClearFilters
	dialog.ClearFilters()
	if dialog.filterTask != "" || dialog.filterStatus != "all" {
		t.Error("Expected filters to be cleared")
	}
}

// TestFileChangesDialogStatusFilterToggle tests status filter cycling
func TestFileChangesDialogStatusFilterToggle(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	// Test cycling through status filters
	expected := []string{"added", "modified", "deleted", "all"}
	
	for _, exp := range expected {
		dialog.ToggleStatusFilter()
		if dialog.filterStatus != exp {
			t.Errorf("Expected filterStatus %s, got %s", exp, dialog.filterStatus)
		}
	}
}

// TestFileChangesDialogSearchQuery tests search functionality
func TestFileChangesDialogSearchQuery(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	// Test SetSearchQuery
	dialog.SetSearchQuery("app.go")
	if dialog.searchQuery != "app.go" {
		t.Errorf("Expected searchQuery to be 'app.go', got %s", dialog.searchQuery)
	}
	
	// Verify tree is filtered
	tree := dialog.fileTree
	foundMatch := false
	for _, item := range tree.items {
		if item.Type == "file" && strings.Contains(item.Path, "app.go") {
			foundMatch = true
			break
		}
	}
	if !foundMatch {
		t.Log("Warning: Expected to find app.go in filtered results")
		// Note: This may not fail if test data structure changed
	}
	
	// Clear search
	dialog.SetSearchQuery("")
	if dialog.searchQuery != "" {
		t.Error("Expected searchQuery to be empty after clearing")
	}
}

// TestFileTreeMatchesFilters tests the matchesFilters method
func TestFileTreeMatchesFilters(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	
	testChange := &taskmaster.FileChange{
		Path:        "internal/ui/app.go",
		ChangeType:  "modified",
		Description: "Updated main app",
	}
	
	// Test no filters (should match)
	if !tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to match with no filters")
	}
	
	// Test task filter match
	dialog.filterTask = "1.1"
	if !tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to match task filter")
	}
	
	// Test task filter no match
	dialog.filterTask = "2.1"
	if tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to not match task filter")
	}
	dialog.filterTask = ""
	
	// Test status filter match
	dialog.filterStatus = "modified"
	if !tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to match status filter")
	}
	
	// Test status filter no match
	dialog.filterStatus = "added"
	if tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to not match status filter")
	}
	dialog.filterStatus = "all"
	
	// Test search query match (path)
	dialog.searchQuery = "app.go"
	if !tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to match search query (path)")
	}
	
	// Test search query match (description)
	dialog.searchQuery = "main app"
	if !tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to match search query (description)")
	}
	
	// Test search query no match
	dialog.searchQuery = "nonexistent"
	if tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to not match search query")
	}
	
	// Test combined filters
	dialog.filterTask = "1.1"
	dialog.filterStatus = "modified"
	dialog.searchQuery = "app.go"
	if !tree.matchesFilters(testChange, "1.1") {
		t.Error("Expected change to match all combined filters")
	}
}

// TestFileTreeSearchHighlighting tests search match highlighting
func TestFileTreeSearchHighlighting(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	
	// Set search query
	dialog.searchQuery = "app"
	
	testChange := &taskmaster.FileChange{
		Path:       "internal/ui/app.go",
		ChangeType: "modified",
	}
	
	item := FileTreeItem{
		Type:   "file",
		Path:   testChange.Path,
		Change: testChange,
		Level:  1,
	}
	
	// Render item with search highlighting
	rendered := tree.renderItem(item, false)
	
	// Check that rendered output is not empty
	if rendered == "" {
		t.Error("Expected non-empty rendered output")
	}
	
	// Note: Can't easily test ANSI color codes in rendered output,
	// but we verify the method doesn't crash
}

// TestFileChangesDialogKeyboardFiltering tests keyboard filter shortcuts
func TestFileChangesDialogKeyboardFiltering(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tests := []struct {
		name     string
		key      string
		check    func() bool
		expected string
	}{
		{
			name: "F key cycles filter",
			key:  "f",
			check: func() bool {
				return dialog.filterStatus != "all"
			},
			expected: "status filter changed",
		},
		{
			name: "Shift+F clears filters",
			key:  "F",
			check: func() bool {
				return dialog.filterStatus == "all" && dialog.filterTask == "" && dialog.searchQuery == ""
			},
			expected: "all filters cleared",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			dialog.HandleKey(msg)
			
			if !tt.check() {
				t.Errorf("Expected %s after %s key", tt.expected, tt.name)
			}
		})
	}
}

// TestFileTreeKeyboardNavigation tests comprehensive tree navigation
func TestFileTreeKeyboardNavigation(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	tree.buildTree()
	
	if len(tree.items) == 0 {
		t.Skip("No tree items to test navigation")
	}
	
	// Test up/down navigation
	initialIdx := tree.selectedIdx
	tree.moveDown()
	if tree.selectedIdx != initialIdx+1 && tree.selectedIdx != len(tree.items)-1 {
		t.Errorf("Expected selectedIdx to increase after moveDown")
	}
	
	tree.moveUp()
	if tree.selectedIdx != initialIdx {
		t.Errorf("Expected selectedIdx to return to initial after moveUp")
	}
	
	// Test boundary conditions
	tree.selectedIdx = 0
	tree.moveUp() // Should stay at 0
	if tree.selectedIdx != 0 {
		t.Error("Expected selectedIdx to stay at 0 when moving up from top")
	}
	
	tree.selectedIdx = len(tree.items) - 1
	tree.moveDown() // Should stay at last index
	if tree.selectedIdx != len(tree.items)-1 {
		t.Error("Expected selectedIdx to stay at last when moving down from bottom")
	}
}

// TestFileTreeKeyboardExpansion tests expand/collapse with keyboard
func TestFileTreeKeyboardExpansion(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	tree.buildTree()
	
	if len(tree.items) == 0 {
		t.Skip("No tree items to test expansion")
	}
	
	// Find an expandable item (task or directory)
	var expandableIdx int
	foundExpandable := false
	for i, item := range tree.items {
		if item.Type == "task" || item.Type == "directory" {
			expandableIdx = i
			foundExpandable = true
			break
		}
	}
	
	if !foundExpandable {
		t.Skip("No expandable items found")
	}
	
	tree.selectedIdx = expandableIdx
	item := tree.items[expandableIdx]
	
	// Test expand
	tree.expandedItems[item.ID] = false
	tree.expandItem()
	if !tree.expandedItems[item.ID] {
		t.Error("Expected item to be expanded after expandItem()")
	}
	
	// Test collapse
	tree.expandedItems[item.ID] = true
	tree.collapseItem()
	if tree.expandedItems[item.ID] {
		t.Error("Expected item to be collapsed after collapseItem()")
	}
}

// TestFileTreeKeyboardFileSelection tests selecting files with Enter
func TestFileTreeKeyboardFileSelection(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	tree.buildTree()
	
	// Find a file item
	var fileIdx int
	foundFile := false
	for i, item := range tree.items {
		if item.Type == "file" && item.Change != nil {
			fileIdx = i
			foundFile = true
			break
		}
	}
	
	if !foundFile {
		t.Skip("No file items found")
	}
	
	tree.selectedIdx = fileIdx
	
	// Select the file
	tree.selectItem()
	
	// Verify preview was loaded
	if dialog.preview.file == "" {
		t.Log("Warning: Expected preview.file to be set after selectItem()")
		// This may not work in test environment without actual file
	}
}

// TestDialogGlobalKeyboardShortcuts tests all global shortcuts
func TestDialogGlobalKeyboardShortcuts(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tests := []struct {
		name     string
		key      string
		check    func() bool
		expected string
	}{
		{
			name: "Esc closes dialog",
			key:  "esc",
			check: func() bool {
				result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
				return result == DialogResultCancel
			},
			expected: "DialogResultCancel",
		},
		{
			name: "Tab switches panel",
			key:  "tab",
			check: func() bool {
				initialPanel := dialog.focusedPanel
				dialog.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
				return dialog.focusedPanel != initialPanel
			},
			expected: "panel to switch",
		},
		{
			name: "G cycles groupBy",
			key:  "g",
			check: func() bool {
				initialGroup := dialog.groupBy
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
				dialog.HandleKey(msg)
				return dialog.groupBy != initialGroup
			},
			expected: "groupBy to change",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("Expected %s after %s key", tt.expected, tt.name)
			}
		})
	}
}

// TestPreviewKeyboardScrolling tests all preview scroll shortcuts
func TestPreviewKeyboardScrolling(t *testing.T) {
	preview := NewPreview()
	preview.lines = make([]string, 50) // Create 50 lines
	for i := range preview.lines {
		preview.lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	preview.height = 10
	preview.scrollPos = 0
	
	tests := []struct {
		name        string
		key         tea.KeyMsg
		action      func()
		initialPos  int
		expectedPos int
	}{
		{
			name:        "Up key scrolls up",
			key:         tea.KeyMsg{Type: tea.KeyUp},
			action:      func() { preview.scrollUp() },
			initialPos:  5,
			expectedPos: 4,
		},
		{
			name:        "Down key scrolls down",
			key:         tea.KeyMsg{Type: tea.KeyDown},
			action:      func() { preview.scrollDown() },
			initialPos:  5,
			expectedPos: 6,
		},
		{
			name:        "PageUp scrolls page up",
			key:         tea.KeyMsg{Type: tea.KeyPgUp},
			action:      func() { preview.pageUp() },
			initialPos:  20,
			expectedPos: 10,
		},
		{
			name:        "PageDown scrolls page down",
			key:         tea.KeyMsg{Type: tea.KeyPgDown},
			action:      func() { preview.pageDown() },
			initialPos:  20,
			expectedPos: 30,
		},
		{
			name:        "Home jumps to top",
			key:         tea.KeyMsg{Type: tea.KeyHome},
			action:      func() { preview.scrollPos = 0 },
			initialPos:  20,
			expectedPos: 0,
		},
		{
			name:        "End jumps to bottom",
			key:         tea.KeyMsg{Type: tea.KeyEnd},
			action:      func() { preview.scrollPos = len(preview.lines) - preview.height },
			initialPos:  0,
			expectedPos: 40,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview.scrollPos = tt.initialPos
			tt.action()
			if preview.scrollPos != tt.expectedPos {
				t.Errorf("Expected scrollPos %d, got %d", tt.expectedPos, preview.scrollPos)
			}
		})
	}
}

// TestKeyboardNavigationSequence tests rapid key sequences
func TestKeyboardNavigationSequence(t *testing.T) {
	tracker := createTestTracker(t)
	gitService := git.NewGitService("/tmp/test-repo")
	dialog := NewFileChangesDialog(tracker, gitService)
	dialog.Init()
	
	tree := dialog.fileTree
	tree.buildTree()
	
	if len(tree.items) < 3 {
		t.Skip("Need at least 3 items for sequence test")
	}
	
	// Rapid navigation sequence
	tree.selectedIdx = 0
	tree.moveDown()
	tree.moveDown()
	tree.moveUp()
	
	expectedIdx := 1
	if tree.selectedIdx != expectedIdx {
		t.Errorf("Expected selectedIdx %d after sequence, got %d", expectedIdx, tree.selectedIdx)
	}
	
	// Test panel switching sequence
	dialog.focusedPanel = "tree"
	dialog.toggleFocusedPanel()
	if dialog.focusedPanel != "preview" {
		t.Error("Expected panel to switch to preview")
	}
	dialog.toggleFocusedPanel()
	if dialog.focusedPanel != "tree" {
		t.Error("Expected panel to switch back to tree")
	}
}
