package dialog

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFileSelectionDialog_Init_LoadsDirectory(t *testing.T) {
	startDir := filepath.Join("/Users/adriangreen/Work/taskmaster-crush-fork", ".taskmaster", "docs")
	dialog := NewFileSelectionDialog("Select PRD File", startDir, 78, 20, []string{".md", ".txt"})

	if dialog.loading {
		t.Errorf("Expected loading=false after init, got true")
	}

	cmd := dialog.Init()
	if cmd == nil {
		t.Fatalf("Init() returned nil command")
	}

	// Simulate command execution
	msg := cmd()
	filesMsg, ok := msg.(FileSelectionEntriesMsg)
	if !ok {
		t.Fatalf("Expected FileSelectionEntriesMsg, got %T", msg)
	}

	if filesMsg.Err != nil {
		t.Fatalf("Expected no error, got %v", filesMsg.Err)
	}

	if len(filesMsg.Entries) == 0 {
		t.Fatalf("Expected entries, got empty list")
	}

	// Check if we have .md and .txt files
	hasMarkdown := false
	hasText := false
	for _, entry := range filesMsg.Entries {
		if entry.Name == "feature-additions-prd.md" {
			hasMarkdown = true
		}
		if entry.Name == "tui-advanced-features.txt" {
			hasText = true
		}
	}

	if !hasMarkdown {
		t.Errorf("Expected markdown file in entries")
	}
	if !hasText {
		t.Errorf("Expected text file in entries")
	}
}

func TestFileSelectionDialog_Update_HandlesEntries(t *testing.T) {
	startDir := filepath.Join("/Users/adriangreen/Work/taskmaster-crush-fork", ".taskmaster", "docs")
	dialog := NewFileSelectionDialog("Select PRD File", startDir, 78, 20, []string{".md", ".txt"})

	// First init to load the directory
	cmd := dialog.Init()
	if cmd == nil {
		t.Fatalf("Init() returned nil command")
	}

	// Get the message that would be sent
	msg := cmd()

	// Update with the entries message
	updatedDialog, _ := dialog.Update(msg)

	// Cast back to our dialog type
	d := updatedDialog.(*FileSelectionDialog)

	if d.loading {
		t.Errorf("Expected loading=false after Update with entries, got true")
	}

	if len(d.entries) == 0 {
		t.Errorf("Expected entries after Update, got empty list")
	}

	if d.selected >= len(d.entries) {
		t.Errorf("Selected index out of bounds: %d >= %d", d.selected, len(d.entries))
	}
}

func TestFileSelectionDialog_View_ShowsEntries(t *testing.T) {
	startDir := filepath.Join("/Users/adriangreen/Work/taskmaster-crush-fork", ".taskmaster", "docs")
	dialog := NewFileSelectionDialog("Select PRD File", startDir, 78, 20, []string{".md", ".txt"})

	// Apply default style
	dialog.Style = DefaultDialogStyle()

	// Simulate window resize to set visibleItems
	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedDialog, _ := dialog.Update(windowMsg)
	d := updatedDialog.(*FileSelectionDialog)

	// Initialize and load
	cmd := d.Init()
	msg := cmd()
	updatedDialog, _ = d.Update(msg)
	d = updatedDialog.(*FileSelectionDialog)

	// Render the view
	view := d.View()

	if len(view) == 0 {
		t.Errorf("Expected non-empty view after loading entries")
	}

	// Check that the location is shown
	if !contains(view, d.currentPath) {
		t.Errorf("View should show current path %s", d.currentPath)
	}

	// Check that files are shown
	if !contains(view, "feature-additions-prd.md") && !contains(view, "tui-advanced-features.txt") {
		t.Errorf("View should show at least one of the PRD files")
	}
}

func TestFileSelectionDialog_HandleKey_Navigation(t *testing.T) {
	startDir := filepath.Join("/Users/adriangreen/Work/taskmaster-crush-fork", ".taskmaster", "docs")
	dialog := NewFileSelectionDialog("Select PRD File", startDir, 78, 20, []string{".md", ".txt"})

	dialog.Style = DefaultDialogStyle()

	// Load entries
	cmd := dialog.Init()
	msg := cmd()
	updatedDialog, _ := dialog.Update(msg)
	d := updatedDialog.(*FileSelectionDialog)

	initialSelected := d.selected

	// Test down key
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := d.HandleKey(keyMsg)
	if result != DialogResultNone {
		t.Errorf("Down key should not close dialog")
	}
	if d.selected != initialSelected+1 && len(d.entries) > 1 {
		t.Errorf("Expected selected to move down")
	}

	// Test up key
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	result, _ = d.HandleKey(keyMsg)
	if result != DialogResultNone {
		t.Errorf("Up key should not close dialog")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFileExtensionMatching(t *testing.T) {
	// Skip if running in CI environment without file system access
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Setup a test directory with files
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"document.txt",
		"readme.md",
		"script.js",
		"image.png",
	}

	for _, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filepath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test with various filter combinations
	testCases := []struct {
		name       string
		extensions []string
		expected   int
	}{
		{
			name:       "No filters",
			extensions: nil,
			expected:   len(testFiles),
		},
		{
			name:       "Filter txt only",
			extensions: []string{".txt"},
			expected:   1, // document.txt
		},
		{
			name:       "Filter md and txt",
			extensions: []string{".md", ".txt"},
			expected:   2, // document.txt and readme.md
		},
		{
			name:       "Filter without dot prefix",
			extensions: []string{"md", "txt"},
			expected:   2, // document.txt and readme.md
		},
		{
			name:       "Filter mixed formats",
			extensions: []string{".MD", "TXT", "js"},
			expected:   3, // document.txt, readme.md, and script.js
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create dialog with test extensions
			dialog := NewFileSelectionDialog("Test", tempDir, 50, 20, tc.extensions)

			// Get entries directly using the dialog's filter map
			entries, err := readDirectoryEntries(tempDir, dialog.filters)
			if err != nil {
				t.Fatalf("Error reading directory: %v", err)
			}

			// Filter out directories
			fileCount := 0
			for _, entry := range entries {
				if !entry.IsDir {
					fileCount++
				}
			}

			// Check if we have the expected number of files
			if fileCount != tc.expected {
				t.Errorf("Expected %d files, got %d with filters %v",
					tc.expected, fileCount, tc.extensions)

				// Log the file entries for debugging
				for _, entry := range entries {
					if !entry.IsDir {
						t.Logf("Found file: %s", entry.Name)
					}
				}

				// Log the normalized filters
				filtersStr := "nil"
				if dialog.filters != nil {
					filters := make([]string, 0, len(dialog.filters))
					for f := range dialog.filters {
						filters = append(filters, f)
					}
					filtersStr = fmt.Sprintf("%v", filters)
				}
				t.Logf("Normalized filters: %s", filtersStr)
			}
		})
	}
}

// TestFileSelectionDialog_DirectoryNavigation verifies that directory navigation
// doesn't cause visual jumps (selection stays stable until new entries load)
func TestFileSelectionDialog_DirectoryNavigation(t *testing.T) {
	// Skip if running in CI environment without file system access
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create test directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files in root
	if err := os.WriteFile(filepath.Join(tempDir, "root.txt"), []byte("root"), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	// Create test files in subdirectory
	if err := os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("sub"), 0644); err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	// Create dialog starting in root
	dialog := NewFileSelectionDialog("Test", tempDir, 78, 20, []string{".txt"})
	dialog.Style = DefaultDialogStyle()

	// Load initial entries
	cmd := dialog.Init()
	if cmd == nil {
		t.Fatalf("Init() returned nil command")
	}
	msg := cmd()
	updatedDialog, _ := dialog.Update(msg)
	d := updatedDialog.(*FileSelectionDialog)

	// Verify we have entries (should have subdir and root.txt, possibly parent dir)
	if len(d.entries) < 2 {
		t.Fatalf("Expected at least 2 entries (subdir + root.txt), got %d", len(d.entries))
	}

	// Find the subdirectory entry
	subdirIndex := -1
	for i, entry := range d.entries {
		if entry.IsDir && !entry.IsParent && entry.Name == "subdir" {
			subdirIndex = i
			break
		}
	}
	if subdirIndex == -1 {
		t.Fatalf("Could not find 'subdir' entry in list")
	}

	// Move selection to subdirectory
	d.selected = subdirIndex
	initialSelected := d.selected
	initialPath := d.currentPath

	// Simulate pressing Enter on the subdirectory
	// This should NOT immediately change d.selected before entries load
	result, cmd := d.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultNone {
		t.Errorf("Enter on directory should return DialogResultNone, got %v", result)
	}

	// CRITICAL: At this point, selected should STILL be at subdirIndex
	// because the async load hasn't completed yet
	if d.selected != initialSelected {
		t.Errorf("Selection changed before directory loaded: was %d, now %d", initialSelected, d.selected)
	}

	// CRITICAL: Path should also not change yet
	if d.currentPath != initialPath {
		t.Errorf("Path changed before directory loaded: was %s, now %s", initialPath, d.currentPath)
	}

	// Now execute the async command to load the directory
	if cmd == nil {
		t.Fatalf("HandleKey(Enter) on directory should return a command")
	}
	msg = cmd()
	updatedDialog, _ = d.Update(msg)
	d = updatedDialog.(*FileSelectionDialog)

	// NOW selection should be reset to 0 (first item in new directory)
	if d.selected != 0 {
		t.Errorf("After loading directory, selection should be 0, got %d", d.selected)
	}

	// Path should now be updated to subdirectory
	if !contains(d.currentPath, "subdir") {
		t.Errorf("After loading, path should contain 'subdir', got %s", d.currentPath)
	}

	// Verify we have entries from the subdirectory
	foundSubFile := false
	for _, entry := range d.entries {
		if entry.Name == "sub.txt" {
			foundSubFile = true
			break
		}
	}
	if !foundSubFile {
		t.Errorf("Expected to find 'sub.txt' in subdirectory entries")
	}

	t.Logf("✓ Directory navigation works correctly - no visual jump")
}
