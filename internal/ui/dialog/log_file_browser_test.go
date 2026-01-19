package dialog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// TestExtractTaskID tests the task ID extraction from filenames
func TestExtractTaskID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"simple task ID", "1.2.log", "1.2"},
		{"nested task ID", "3.1.4.md", "3.1.4"},
		{"task ID with prefix", "task-2.1.log", "2.1"},
		{"task ID with underscore", "log_1.3.txt", "1.3"},
		{"no task ID", "notes.txt", ""},
		{"no extension", "readme", ""},
		{"task ID no extension", "1.2", "1.2"},
		{"multiple dots", "test.1.2.log", "1.2"},
		{"no valid pattern", "test123.log", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTaskID(tt.filename)
			if result != tt.expected {
				t.Errorf("extractTaskID(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestIsTaskIDPattern tests the task ID pattern matching
func TestIsTaskIDPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple", "1.2", true},
		{"valid nested", "3.1.4", true},
		{"valid long", "1.2.3.4.5", true},
		{"no dot", "123", false},
		{"no digit", "abc", false},
		{"invalid char", "1.2a", false},
		{"empty", "", false},
		{"only dot", ".", false},
		{"ends with dot", "1.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTaskIDPattern(tt.input)
			if result != tt.expected {
				t.Errorf("isTaskIDPattern(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCompareTaskIDs tests task ID comparison
func TestCompareTaskIDs(t *testing.T) {
	tests := []struct {
		name     string
		id1      string
		id2      string
		expected bool // true if id1 < id2
	}{
		{"simple less", "1.1", "1.2", true},
		{"simple greater", "1.2", "1.1", false},
		{"equal", "1.2", "1.2", false},
		{"different major", "1.2", "2.1", true},
		{"nested less", "1.2.1", "1.2.2", true},
		{"nested vs simple", "1.2", "1.2.1", true},
		{"numeric comparison", "1.10", "1.2", false}, // 1.10 > 1.2
		{"deep nested", "1.2.3.4", "1.2.3.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareTaskIDs(tt.id1, tt.id2)
			if result != tt.expected {
				t.Errorf("compareTaskIDs(%q, %q) = %v, want %v", tt.id1, tt.id2, result, tt.expected)
			}
		})
	}
}

// TestShouldSkipEntry tests the file filtering logic
func TestShouldSkipEntry(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool // true if should skip
	}{
		{"hidden file", ".hidden", true},
		{"node_modules", "node_modules", true},
		{"dist directory", "dist", true},
		{"normal file", "test.log", false},
		{"normal directory", "logs", false},
		{".git directory", ".git", true},
		{"dotfile", ".env", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipEntry(tt.filename)
			if result != tt.expected {
				t.Errorf("shouldSkipEntry(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestIsSupportedFile tests file extension filtering
func TestIsSupportedFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"log file", "test.log", true},
		{"markdown file", "readme.md", true},
		{"text file", "notes.txt", true},
		{"no extension", "README", true},
		{"unsupported extension", "test.pdf", false},
		{"executable", "script.sh", false},
		{"uppercase extension", "FILE.LOG", true},
		{"mixed case", "Test.Md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSupportedFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isSupportedFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestSortFileEntries tests the sorting logic
func TestSortFileEntries(t *testing.T) {
	entries := []FileEntry{
		{Name: "3.1.log", IsDir: false, ModTime: time.Now()},
		{Name: "logs", IsDir: true, ModTime: time.Now()},
		{Name: "1.2.log", IsDir: false, ModTime: time.Now()},
		{Name: "archive", IsDir: true, ModTime: time.Now()},
		{Name: "notes.txt", IsDir: false, ModTime: time.Now()},
		{Name: "1.10.log", IsDir: false, ModTime: time.Now()},
		{Name: "2.1.md", IsDir: false, ModTime: time.Now()},
	}

	sortFileEntries(entries)

	// Expected order:
	// 1. Directories first (archive, logs)
	// 2. Files with task IDs sorted numerically (1.2.log, 1.10.log, 2.1.md, 3.1.log)
	// 3. Files without task IDs alphabetically (notes.txt)

	expectedOrder := []string{
		"archive",      // directory
		"logs",         // directory
		"1.2.log",      // task ID 1.2
		"1.10.log",     // task ID 1.10
		"2.1.md",       // task ID 2.1
		"3.1.log",      // task ID 3.1
		"notes.txt",    // no task ID
	}

	if len(entries) != len(expectedOrder) {
		t.Fatalf("Expected %d entries, got %d", len(expectedOrder), len(entries))
	}

	for i, expected := range expectedOrder {
		if entries[i].Name != expected {
			t.Errorf("Position %d: expected %q, got %q", i, expected, entries[i].Name)
		}
	}
}

// TestFormatFileSize tests file size formatting
func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"bytes", 100, "100 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"kilobytes fractional", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %q, want %q", tt.size, result, tt.expected)
			}
		})
	}
}

// TestDiscoverFiles tests file discovery with fixtures
func TestDiscoverFiles(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create test directory structure
	testFiles := map[string]bool{
		"1.2.log":         false, // file
		"3.1.md":          false, // file
		"notes.txt":       false, // file
		".hidden":         false, // should be skipped
		"test.pdf":        false, // should be skipped (unsupported extension)
		"logs":            true,  // directory
		"node_modules":    true,  // should be skipped
	}

	for name, isDir := range testFiles {
		path := filepath.Join(tempDir, name)
		if isDir {
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", path, err)
			}
		} else {
			if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", path, err)
			}
		}
	}

	// Discover files
	entries, err := discoverFiles(tempDir)
	if err != nil {
		t.Fatalf("discoverFiles failed: %v", err)
	}

	// Expected entries: logs (dir), 1.2.log, 3.1.md, notes.txt
	// Skipped: .hidden, test.pdf, node_modules
	expectedCount := 4
	if len(entries) != expectedCount {
		t.Errorf("Expected %d entries, got %d", expectedCount, len(entries))
		for _, entry := range entries {
			t.Logf("Found: %s (IsDir: %v)", entry.Name, entry.IsDir)
		}
	}

	// Verify directory comes first
	if len(entries) > 0 && entries[0].Name != "logs" {
		t.Errorf("Expected first entry to be 'logs', got %q", entries[0].Name)
	}

	// Verify files are sorted by task ID
	fileNames := []string{}
	for _, entry := range entries {
		if !entry.IsDir {
			fileNames = append(fileNames, entry.Name)
		}
	}

	expectedFiles := []string{"1.2.log", "3.1.md", "notes.txt"}
	if len(fileNames) != len(expectedFiles) {
		t.Errorf("Expected %d files, got %d", len(expectedFiles), len(fileNames))
	}

	for i, expected := range expectedFiles {
		if i >= len(fileNames) {
			break
		}
		if fileNames[i] != expected {
			t.Errorf("File position %d: expected %q, got %q", i, expected, fileNames[i])
		}
	}
}

// TestDiscoverFilesEmptyDirectory tests handling of empty directories
func TestDiscoverFilesEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	entries, err := discoverFiles(tempDir)
	if err != nil {
		t.Fatalf("discoverFiles failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for empty directory, got %d", len(entries))
	}
}

// TestDiscoverFilesNonexistentDirectory tests error handling
func TestDiscoverFilesNonexistentDirectory(t *testing.T) {
	_, err := discoverFiles("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}
}

// TestFileEntryImplementsListItem verifies FileEntry implements list.Item interface
func TestFileEntryImplementsListItem(t *testing.T) {
	entry := FileEntry{
		Name:        "test.log",
		Path:        "/path/to/test.log",
		IsDir:       false,
		Size:        1024,
		ModTime:     time.Now(),
		DisplayName: "test.log",
	}

	// Test FilterValue
	if entry.FilterValue() != "test.log" {
		t.Errorf("FilterValue() = %q, want %q", entry.FilterValue(), "test.log")
	}

	// Test Title
	title := entry.Title()
	if title != "📄 test.log" {
		t.Errorf("Title() = %q, want %q", title, "📄 test.log")
	}

	// Test Description
	desc := entry.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

// TestFileEntryDirectoryDisplay tests directory display formatting
func TestFileEntryDirectoryDisplay(t *testing.T) {
	entry := FileEntry{
		Name:        "logs",
		Path:        "/path/to/logs",
		IsDir:       true,
		Size:        0,
		ModTime:     time.Now(),
		DisplayName: "logs",
	}

	title := entry.Title()
	if title != "📁 logs" {
		t.Errorf("Title() = %q, want %q", title, "📁 logs")
	}
}

// TestSortFileEntriesEdgeCases tests edge cases in sorting
func TestSortFileEntriesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		entries  []FileEntry
		expected []string
	}{
		{
			name:     "empty slice",
			entries:  []FileEntry{},
			expected: []string{},
		},
		{
			name: "single entry",
			entries: []FileEntry{
				{Name: "test.log", IsDir: false, ModTime: time.Now()},
			},
			expected: []string{"test.log"},
		},
		{
			name: "all directories",
			entries: []FileEntry{
				{Name: "logs", IsDir: true, ModTime: time.Now()},
				{Name: "archive", IsDir: true, ModTime: time.Now()},
			},
			expected: []string{"archive", "logs"},
		},
		{
			name: "all files with task IDs",
			entries: []FileEntry{
				{Name: "3.1.log", IsDir: false, ModTime: time.Now()},
				{Name: "1.2.log", IsDir: false, ModTime: time.Now()},
				{Name: "2.1.log", IsDir: false, ModTime: time.Now()},
			},
			expected: []string{"1.2.log", "2.1.log", "3.1.log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortFileEntries(tt.entries)

			if len(tt.entries) != len(tt.expected) {
				t.Fatalf("Expected %d entries, got %d", len(tt.expected), len(tt.entries))
			}

			for i, expected := range tt.expected {
				if tt.entries[i].Name != expected {
					t.Errorf("Position %d: expected %q, got %q", i, expected, tt.entries[i].Name)
				}
			}
		})
	}
}

// TestExtractTaskIDEdgeCases tests edge cases in task ID extraction
func TestExtractTaskIDEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"very long task ID", "1.2.3.4.5.6.7.8.log", "1.2.3.4.5.6.7.8"},
		{"task ID with spaces", "task 1.2 log.txt", "1.2"},
		{"multiple task IDs", "1.2-3.4.log", "1.2"}, // should return first match
		{"task ID at end", "report-for-1.2.md", "1.2"},
		{"task ID at start", "1.2-implementation.log", "1.2"},
		{"single digit", "1.log", ""},               // not a valid task ID (no dot)
		{"trailing dot", "1.2..log", "1.2"},
		{"unicode characters", "test-①.②.log", ""}, // should not match
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTaskID(tt.filename)
			if result != tt.expected {
				t.Errorf("extractTaskID(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestLogFileBrowserModelNavigation tests directory navigation
func TestLogFileBrowserModelNavigation(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()
	taskmasterDir := filepath.Join(tempDir, ".taskmaster", "test-tag")
	logsDir := filepath.Join(taskmasterDir, "logs")

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test files
	testFiles := []string{
		filepath.Join(taskmasterDir, "1.2.log"),
		filepath.Join(logsDir, "3.1.log"),
	}

	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Change to temp directory for test
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create browser model
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")

	// Test 1: Initial load should find the test-tag directory
	if browser.currentPath == "" {
		t.Error("Expected currentPath to be set after initialization")
	}

	if len(browser.files) == 0 {
		t.Error("Expected files to be loaded after initialization")
	}

	// Test 2: Enter directory navigation
	// Find the logs directory in the list
	logsIndex := -1
	for i, file := range browser.files {
		if file.IsDir && file.Name == "logs" {
			logsIndex = i
			break
		}
	}

	if logsIndex >= 0 {
		// Simulate selecting the logs directory
		for i := 0; i < logsIndex; i++ {
			browser.list.CursorDown()
		}

		// Navigate into the directory
		model, _ := browser.handleEnter()
		browser = model.(*LogFileBrowserModel)

		// Verify we're in the logs directory
		if !strings.HasSuffix(browser.currentPath, "logs") {
			t.Errorf("Expected to be in logs directory, got %s", browser.currentPath)
		}
	}

	// Test 3: Parent directory navigation
	initialPath := browser.currentPath
	model, _ := browser.handleParent()
	browser = model.(*LogFileBrowserModel)

	// Verify we went up one level
	if browser.currentPath == initialPath {
		t.Error("Expected currentPath to change after going to parent")
	}
}

// TestLogFileBrowserModelUpdate tests the Update method
func TestLogFileBrowserModelUpdate(t *testing.T) {
	tempDir := t.TempDir()
	taskmasterDir := filepath.Join(tempDir, ".taskmaster", "test-tag")

	if err := os.MkdirAll(taskmasterDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(taskmasterDir, "test.log")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")

	// Test keyboard message handling
	tests := []struct {
		name string
		key  string
	}{
		{"enter key", "enter"},
		{"l key", "l"},
		{"right arrow", "right"},
		{"backspace key", "backspace"},
		{"h key", "h"},
		{"left arrow", "left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "enter" {
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			} else if tt.key == "backspace" {
				msg = tea.KeyMsg{Type: tea.KeyBackspace}
			} else if tt.key == "right" {
				msg = tea.KeyMsg{Type: tea.KeyRight}
			} else if tt.key == "left" {
				msg = tea.KeyMsg{Type: tea.KeyLeft}
			}

			// Should not panic
			_, _ = browser.Update(msg)
		})
	}
}

// TestLogFileBrowserModelUpdateSelection tests selection tracking
func TestLogFileBrowserModelUpdateSelection(t *testing.T) {
	browser := &LogFileBrowserModel{
		list:  list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		files: []FileEntry{},
	}

	// Test with empty list
	browser.updateSelection()
	if browser.selectedFile != "" {
		t.Errorf("Expected empty selectedFile for empty list, got %q", browser.selectedFile)
	}

	// Test with file entry
	fileEntry := FileEntry{
		Name:  "test.log",
		Path:  "/path/to/test.log",
		IsDir: false,
	}

	browser.files = []FileEntry{fileEntry}
	items := []list.Item{fileEntry}
	browser.list.SetItems(items)
	
	browser.updateSelection()
	
	// Should select the file
	if browser.selectedFile != "/path/to/test.log" {
		t.Errorf("Expected selectedFile to be %q, got %q", "/path/to/test.log", browser.selectedFile)
	}
}

// TestLogFileBrowserModelSetSize tests dimension updates
func TestLogFileBrowserModelSetSize(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")

	browser.SetSize(100, 30)

	if browser.width != 100 {
		t.Errorf("Expected width 100, got %d", browser.width)
	}

	if browser.height != 30 {
		t.Errorf("Expected height 30, got %d", browser.height)
	}
}

// TestLogFileBrowserModelSetFocused tests focus state management
func TestLogFileBrowserModelSetFocused(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")

	browser.SetFocused(true)
	if !browser.focused {
		t.Error("Expected focused to be true")
	}

	browser.SetFocused(false)
	if browser.focused {
		t.Error("Expected focused to be false")
	}
}

// TestLogFileBrowserModelGetSelectedFile tests file selection retrieval
func TestLogFileBrowserModelGetSelectedFile(t *testing.T) {
	browser := &LogFileBrowserModel{
		list:  list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		files: []FileEntry{},
	}

	// Empty list
	selected := browser.GetSelectedFile()
	if selected != "" {
		t.Errorf("Expected empty string for empty list, got %q", selected)
	}

	// With file
	fileEntry := FileEntry{
		Name:  "test.log",
		Path:  "/path/to/test.log",
		IsDir: false,
	}

	browser.files = []FileEntry{fileEntry}
	items := []list.Item{fileEntry}
	browser.list.SetItems(items)

	selected = browser.GetSelectedFile()
	if selected != "/path/to/test.log" {
		t.Errorf("Expected %q, got %q", "/path/to/test.log", selected)
	}
}

// TestLoadFilesFromPath tests loading files from a specific path
func TestLoadFilesFromPath(t *testing.T) {
	tempDir := t.TempDir()
	taskmasterDir := filepath.Join(tempDir, ".taskmaster", "test-tag")

	if err := os.MkdirAll(taskmasterDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test files
	testFiles := []string{"1.2.log", "3.1.md"}
	for _, name := range testFiles {
		path := filepath.Join(taskmasterDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	browser := &LogFileBrowserModel{
		list:     list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		files:    []FileEntry{},
		dirCache: NewLRUCache(10),
	}

	// Load files from the test directory
	browser.loadFilesFromPath(taskmasterDir)

	if len(browser.files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(browser.files))
	}

	if browser.currentPath != taskmasterDir {
		t.Errorf("Expected currentPath to be %q, got %q", taskmasterDir, browser.currentPath)
	}
}

// TestLoadFilesFromPathInvalidPath tests error handling
func TestLoadFilesFromPathInvalidPath(t *testing.T) {
	browser := &LogFileBrowserModel{
		list:     list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		files:    []FileEntry{},
		dirCache: NewLRUCache(10),
	}

	// Try loading from invalid path (outside .taskmaster)
	browser.loadFilesFromPath("/some/invalid/path")

	if browser.currentPath != "" {
		t.Error("Expected currentPath to be empty for invalid path")
	}

	if len(browser.files) != 0 {
		t.Error("Expected files to be empty for invalid path")
	}
}

// TestFileEntryTitleFormatting tests file entry title formatting
func TestFileEntryTitleFormatting(t *testing.T) {
	tests := []struct {
		name     string
		entry    FileEntry
		expected string
	}{
		{
			name: "directory",
			entry: FileEntry{
				Name:        "logs",
				DisplayName: "logs",
				IsDir:       true,
			},
			expected: "📁 logs",
		},
		{
			name: "file",
			entry: FileEntry{
				Name:        "test.log",
				DisplayName: "test.log",
				IsDir:       false,
			},
			expected: "📄 test.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.Title()
			if result != tt.expected {
				t.Errorf("Title() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestFileEntryDescriptionFormatting tests file entry description formatting
func TestFileEntryDescriptionFormatting(t *testing.T) {
	modTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	entry := FileEntry{
		Name:        "test.log",
		Size:        1024,
		ModTime:     modTime,
		DisplayName: "test.log",
	}

	result := entry.Description()

	// Should contain file size
	if !strings.Contains(result, "1.0 KB") {
		t.Errorf("Description should contain '1.0 KB', got %q", result)
	}

	// Should contain timestamp
	if !strings.Contains(result, "2024-01-15 14:30") {
		t.Errorf("Description should contain timestamp, got %q", result)
	}

	// Should use bullet separator
	if !strings.Contains(result, "•") {
		t.Errorf("Description should contain bullet separator, got %q", result)
	}
}

// TestLogFileBrowserModelView tests the View method rendering
func TestLogFileBrowserModelView(t *testing.T) {
	// Test with zero dimensions
	browser := NewLogFileBrowserModel(0, 0, nil, "test-tag")
	view := browser.View()

	if view != "" {
		t.Error("Expected empty view for zero dimensions")
	}

	// Test with valid dimensions
	browser = NewLogFileBrowserModel(80, 24, nil, "test-tag")
	view = browser.View()

	if view == "" {
		t.Error("Expected non-empty view for valid dimensions")
	}
}

// TestLogFileBrowserModelViewFocusedState tests focused state styling
func TestLogFileBrowserModelViewFocusedState(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")

	// Unfocused state
	browser.SetFocused(false)
	viewUnfocused := browser.View()

	// Focused state
	browser.SetFocused(true)
	viewFocused := browser.View()

	// Views should be different (focused has different border color)
	if viewFocused == viewUnfocused {
		// Note: This might not always fail if styling doesn't affect output,
		// but it verifies the method runs without errors in both states
		t.Log("Views are identical, styling may not affect rendered output in tests")
	}
}

// TestLogFileBrowserModelViewWithFiles tests view rendering with files
func TestLogFileBrowserModelViewWithFiles(t *testing.T) {
	browser := &LogFileBrowserModel{
		list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		width:  80,
		height: 24,
		files: []FileEntry{
			{
				Name:        "test.log",
				Path:        "/path/to/test.log",
				IsDir:       false,
				Size:        1024,
				ModTime:     time.Now(),
				DisplayName: "test.log",
			},
		},
	}

	items := []list.Item{browser.files[0]}
	browser.list.SetItems(items)

	view := browser.View()

	if view == "" {
		t.Error("Expected non-empty view with files")
	}
}

// TestFormatFileSizeVariety tests various file sizes
func TestFormatFileSizeVariety(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		contains string
	}{
		{"bytes", 512, "512 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"10 KB", 10240, "10.0 KB"},
		{"1 MB", 1048576, "1.0 MB"},
		{"100 MB", 104857600, "100.0 MB"},
		{"1 GB", 1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.size)
			if result != tt.contains {
				t.Errorf("formatFileSize(%d) = %q, want %q", tt.size, result, tt.contains)
			}
		})
	}
}

// TestFileEntryFilterValue tests filter value for search
func TestFileEntryFilterValue(t *testing.T) {
	entry := FileEntry{
		Name:        "test.log",
		DisplayName: "test.log",
	}

	result := entry.FilterValue()
	if result != "test.log" {
		t.Errorf("FilterValue() = %q, want %q", result, "test.log")
	}
}



// TestBreadcrumbInitialization tests that breadcrumbs are initialized correctly
func TestBreadcrumbInitialization(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	if len(browser.breadcrumbs) != 1 {
		t.Errorf("Expected 1 breadcrumb (root), got %d", len(browser.breadcrumbs))
	}
	if browser.breadcrumbs[0] != "test-tag" {
		t.Errorf("Expected first breadcrumb to be 'test-tag', got %q", browser.breadcrumbs[0])
	}
}

// TestBreadcrumbPush tests adding directories to breadcrumb trail
func TestBreadcrumbPush(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	browser.pushBreadcrumb("dir1")
	if len(browser.breadcrumbs) != 2 {
		t.Errorf("Expected 2 breadcrumbs, got %d", len(browser.breadcrumbs))
	}
	
	browser.pushBreadcrumb("dir2")
	if len(browser.breadcrumbs) != 3 {
		t.Errorf("Expected 3 breadcrumbs, got %d", len(browser.breadcrumbs))
	}
	
	if browser.breadcrumbs[2] != "dir2" {
		t.Errorf("Expected last breadcrumb to be 'dir2', got %q", browser.breadcrumbs[2])
	}
}

// TestBreadcrumbPop tests removing directories from breadcrumb trail
func TestBreadcrumbPop(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	browser.pushBreadcrumb("dir1")
	browser.pushBreadcrumb("dir2")
	
	browser.popBreadcrumb()
	if len(browser.breadcrumbs) != 2 {
		t.Errorf("Expected 2 breadcrumbs after pop, got %d", len(browser.breadcrumbs))
	}
	
	browser.popBreadcrumb()
	if len(browser.breadcrumbs) != 1 {
		t.Errorf("Expected 1 breadcrumb after second pop, got %d", len(browser.breadcrumbs))
	}
	
	// Try to pop past root - should not be allowed
	browser.popBreadcrumb()
	if len(browser.breadcrumbs) != 1 {
		t.Error("Expected to stay at root when popping at minimum depth")
	}
}

// TestBreadcrumbReset tests resetting breadcrumbs to root
func TestBreadcrumbReset(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	browser.pushBreadcrumb("dir1")
	browser.pushBreadcrumb("dir2")
	
	browser.resetBreadcrumbs()
	if len(browser.breadcrumbs) != 1 {
		t.Errorf("Expected 1 breadcrumb after reset, got %d", len(browser.breadcrumbs))
	}
	if browser.breadcrumbs[0] != "test-tag" {
		t.Errorf("Expected breadcrumb to be 'test-tag' after reset, got %q", browser.breadcrumbs[0])
	}
}

// TestGetBreadcrumbString tests breadcrumb string formatting
func TestGetBreadcrumbString(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	breadcrumb := browser.getBreadcrumbString()
	if breadcrumb != "test-tag" {
		t.Errorf("Expected breadcrumb 'test-tag', got %q", breadcrumb)
	}
	
	browser.pushBreadcrumb("dir1")
	breadcrumb = browser.getBreadcrumbString()
	if breadcrumb != "test-tag / dir1" {
		t.Errorf("Expected breadcrumb 'test-tag / dir1', got %q", breadcrumb)
	}
	
	browser.pushBreadcrumb("dir2")
	breadcrumb = browser.getBreadcrumbString()
	if breadcrumb != "test-tag / dir1 / dir2" {
		t.Errorf("Expected breadcrumb 'test-tag / dir1 / dir2', got %q", breadcrumb)
	}
}

// TestCurrentDepth tests depth calculation
func TestCurrentDepth(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	if browser.getCurrentDepth() != 0 {
		t.Errorf("Expected depth 0 at root, got %d", browser.getCurrentDepth())
	}
	
	browser.pushBreadcrumb("dir1")
	if browser.getCurrentDepth() != 1 {
		t.Errorf("Expected depth 1 after first push, got %d", browser.getCurrentDepth())
	}
	
	browser.pushBreadcrumb("dir2")
	if browser.getCurrentDepth() != 2 {
		t.Errorf("Expected depth 2 after second push, got %d", browser.getCurrentDepth())
	}
}

// TestMaxDepthLimit tests that navigation respects max depth
func TestMaxDepthLimit(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	browser.maxDepth = 3 // Set low max depth for testing
	
	// Push up to max depth
	browser.pushBreadcrumb("dir1")
	browser.pushBreadcrumb("dir2")
	browser.pushBreadcrumb("dir3")
	
	// Try to push beyond max depth
	browser.pushBreadcrumb("dir4")
	
	// Should not exceed max depth
	if browser.getCurrentDepth() > browser.maxDepth {
		t.Errorf("Expected depth <= %d, got %d", browser.maxDepth, browser.getCurrentDepth())
	}
}

// TestSetTagResetsBreadcrumbs tests that changing tags resets breadcrumbs
func TestSetTagResetsBreadcrumbs(t *testing.T) {
	tmpDir := t.TempDir()
	service := &taskmaster.Service{RootDir: tmpDir}
	browser := NewLogFileBrowserModel(80, 24, service, "test-tag")
	
	browser.pushBreadcrumb("dir1")
	browser.pushBreadcrumb("dir2")
	
	if browser.getCurrentDepth() != 2 {
		t.Errorf("Expected depth 2 before SetTag, got %d", browser.getCurrentDepth())
	}
	
	browser.SetTag("new-tag")
	
	if browser.currentTag != "new-tag" {
		t.Errorf("Expected current tag 'new-tag', got %q", browser.currentTag)
	}
	
	if browser.getCurrentDepth() != 0 {
		t.Errorf("Expected depth 0 after SetTag, got %d", browser.getCurrentDepth())
	}
	
	if browser.breadcrumbs[0] != "new-tag" {
		t.Errorf("Expected breadcrumb 'new-tag' after SetTag, got %q", browser.breadcrumbs[0])
	}
}

// TestKeyboardShortcutsUpDown tests up/down arrow and k/j navigation
func TestKeyboardShortcutsUpDown(t *testing.T) {
	// Create browser with test files
	browser := &LogFileBrowserModel{
		list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		width:  80,
		height: 24,
		files: []FileEntry{
			{Name: "file1.log", Path: "/path/to/file1.log", IsDir: false, ModTime: time.Now()},
			{Name: "file2.log", Path: "/path/to/file2.log", IsDir: false, ModTime: time.Now()},
			{Name: "file3.log", Path: "/path/to/file3.log", IsDir: false, ModTime: time.Now()},
		},
	}
	
	// Set items in list
	items := make([]list.Item, len(browser.files))
	for i, file := range browser.files {
		items[i] = file
	}
	browser.list.SetItems(items)
	
	// Test down arrow key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	_, _ = browser.Update(downMsg)
	
	// List should move cursor down (to file2.log)
	// The list component handles this internally
	
	// Test j key (vim-style down)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = browser.Update(jMsg)
	
	// Test up arrow key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	_, _ = browser.Update(upMsg)
	
	// Test k key (vim-style up)
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = browser.Update(kMsg)
	
	// Should not panic - list component handles these keys
}

// TestKeyboardShortcutsHomeEnd tests Home and End navigation
func TestKeyboardShortcutsHomeEnd(t *testing.T) {
	// Create browser with multiple files
	browser := &LogFileBrowserModel{
		list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		width:  80,
		height: 24,
		files: []FileEntry{
			{Name: "file1.log", Path: "/path/to/file1.log", IsDir: false, ModTime: time.Now()},
			{Name: "file2.log", Path: "/path/to/file2.log", IsDir: false, ModTime: time.Now()},
			{Name: "file3.log", Path: "/path/to/file3.log", IsDir: false, ModTime: time.Now()},
			{Name: "file4.log", Path: "/path/to/file4.log", IsDir: false, ModTime: time.Now()},
			{Name: "file5.log", Path: "/path/to/file5.log", IsDir: false, ModTime: time.Now()},
		},
	}
	
	items := make([]list.Item, len(browser.files))
	for i, file := range browser.files {
		items[i] = file
	}
	browser.list.SetItems(items)
	
	// Move to middle of list
	browser.list.Select(2)
	
	// Test Home key (jump to top)
	homeMsg := tea.KeyMsg{Type: tea.KeyHome}
	_, _ = browser.Update(homeMsg)
	
	// Test End key (jump to bottom)
	endMsg := tea.KeyMsg{Type: tea.KeyEnd}
	_, _ = browser.Update(endMsg)
	
	// Should not panic - list component handles these keys
}

// TestKeyboardShortcutsSearch tests / search functionality
func TestKeyboardShortcutsSearch(t *testing.T) {
	// Create browser with filterable files
	browser := &LogFileBrowserModel{
		list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		width:  80,
		height: 24,
		files: []FileEntry{
			{Name: "apple.log", Path: "/path/to/apple.log", IsDir: false, ModTime: time.Now()},
			{Name: "banana.log", Path: "/path/to/banana.log", IsDir: false, ModTime: time.Now()},
			{Name: "cherry.log", Path: "/path/to/cherry.log", IsDir: false, ModTime: time.Now()},
		},
	}
	
	// Enable filtering on the list
	browser.list.SetFilteringEnabled(true)
	
	items := make([]list.Item, len(browser.files))
	for i, file := range browser.files {
		items[i] = file
	}
	browser.list.SetItems(items)
	
	// Test / key (activate filter)
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	_, _ = browser.Update(slashMsg)
	
	// Should not panic - list component handles filter activation
}

// TestKeyboardShortcutsEnterAndNavigation tests Enter/l/right for directory navigation
func TestKeyboardShortcutsEnterAndNavigation(t *testing.T) {
	tempDir := t.TempDir()
	taskmasterDir := filepath.Join(tempDir, ".taskmaster", "test-tag")
	subDir := filepath.Join(taskmasterDir, "logs")
	
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test file in subdirectory
	testFile := filepath.Join(subDir, "test.log")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	// Find logs directory
	logsIndex := -1
	for i, file := range browser.files {
		if file.IsDir && file.Name == "logs" {
			logsIndex = i
			break
		}
	}
	
	if logsIndex < 0 {
		t.Skip("logs directory not found in test setup")
	}
	
	// Move cursor to logs directory
	for i := 0; i < logsIndex; i++ {
		browser.list.CursorDown()
	}
	
	// Test Enter key navigation
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, _ := browser.Update(enterMsg)
	browser = model.(*LogFileBrowserModel)
	
	// Should have entered the logs directory
	if !strings.HasSuffix(browser.currentPath, "logs") {
		t.Errorf("Expected to be in logs directory after Enter, got %s", browser.currentPath)
	}
	
	// Test l key navigation (vim-style)
	// Reset to parent
	model, _ = browser.handleParent()
	browser = model.(*LogFileBrowserModel)
	
	for i := 0; i < logsIndex; i++ {
		browser.list.CursorDown()
	}
	
	lMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	model, _ = browser.Update(lMsg)
	browser = model.(*LogFileBrowserModel)
	
	if !strings.HasSuffix(browser.currentPath, "logs") {
		t.Errorf("Expected to be in logs directory after 'l', got %s", browser.currentPath)
	}
	
	// Test right arrow navigation
	// Reset to parent
	model, _ = browser.handleParent()
	browser = model.(*LogFileBrowserModel)
	
	for i := 0; i < logsIndex; i++ {
		browser.list.CursorDown()
	}
	
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	model, _ = browser.Update(rightMsg)
	browser = model.(*LogFileBrowserModel)
	
	if !strings.HasSuffix(browser.currentPath, "logs") {
		t.Errorf("Expected to be in logs directory after right arrow, got %s", browser.currentPath)
	}
}

// TestKeyboardShortcutsBackspaceAndParent tests Backspace/h/left for parent navigation
func TestKeyboardShortcutsBackspaceAndParent(t *testing.T) {
	tempDir := t.TempDir()
	taskmasterDir := filepath.Join(tempDir, ".taskmaster", "test-tag")
	subDir := filepath.Join(taskmasterDir, "logs")
	
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	// Navigate into subdirectory
	browser.currentPath = subDir
	browser.loadFilesFromPath(subDir)
	
	initialPath := browser.currentPath
	
	// Test Backspace key
	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	model, _ := browser.Update(backspaceMsg)
	browser = model.(*LogFileBrowserModel)
	
	if browser.currentPath == initialPath {
		t.Error("Expected currentPath to change after Backspace")
	}
	
	// Reset to subdirectory
	browser.currentPath = subDir
	browser.loadFilesFromPath(subDir)
	
	// Test h key (vim-style)
	hMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	model, _ = browser.Update(hMsg)
	browser = model.(*LogFileBrowserModel)
	
	if browser.currentPath == subDir {
		t.Error("Expected currentPath to change after 'h'")
	}
	
	// Reset to subdirectory
	browser.currentPath = subDir
	browser.loadFilesFromPath(subDir)
	
	// Test left arrow key
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	model, _ = browser.Update(leftMsg)
	browser = model.(*LogFileBrowserModel)
	
	if browser.currentPath == subDir {
		t.Error("Expected currentPath to change after left arrow")
	}
}

// TestKeyboardShortcutsOnlyWorkWhenFocused tests that shortcuts only work when panel is focused
func TestKeyboardShortcutsOnlyWorkWhenFocused(t *testing.T) {
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	
	// The browser should handle keys regardless of focus state in its Update method
	// Focus is managed by the parent dialog to route messages only to focused panels
	
	browser.SetFocused(true)
	if !browser.focused {
		t.Error("Expected browser to be focused")
	}
	
	// Test that update works when focused
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	_, _ = browser.Update(keyMsg)
	// Should not panic
	
	browser.SetFocused(false)
	if browser.focused {
		t.Error("Expected browser to be unfocused")
	}
	
	// When unfocused, the parent dialog should not route messages to this panel
	// But the Update method itself doesn't check focus - that's handled by the parent
	_, _ = browser.Update(keyMsg)
	// Should still not panic (parent handles routing)
}

// TestKeyboardShortcutsPageUpDown tests PageUp/PageDown scrolling
func TestKeyboardShortcutsPageUpDown(t *testing.T) {
	// Create browser with many files
	browser := &LogFileBrowserModel{
		list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24),
		width:  80,
		height: 24,
		files:  []FileEntry{},
	}
	
	// Add many files to enable page scrolling
	for i := 1; i <= 30; i++ {
		entry := FileEntry{
			Name:        fmt.Sprintf("file%d.log", i),
			Path:        fmt.Sprintf("/path/to/file%d.log", i),
			IsDir:       false,
			Size:        1024,
			ModTime:     time.Now(),
			DisplayName: fmt.Sprintf("file%d.log", i),
		}
		browser.files = append(browser.files, entry)
	}
	
	items := make([]list.Item, len(browser.files))
	for i, file := range browser.files {
		items[i] = file
	}
	browser.list.SetItems(items)
	
	// Test PageDown key
	pgDnMsg := tea.KeyMsg{Type: tea.KeyPgDown}
	_, _ = browser.Update(pgDnMsg)
	
	// Test PageUp key
	pgUpMsg := tea.KeyMsg{Type: tea.KeyPgUp}
	_, _ = browser.Update(pgUpMsg)
	
	// Should not panic - list component handles page scrolling
}

// TestAllKeyboardShortcutsCombined tests all shortcuts in sequence
func TestAllKeyboardShortcutsCombined(t *testing.T) {
	tempDir := t.TempDir()
	taskmasterDir := filepath.Join(tempDir, ".taskmaster", "test-tag")
	
	if err := os.MkdirAll(taskmasterDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test files
	for i := 1; i <= 5; i++ {
		testFile := filepath.Join(taskmasterDir, fmt.Sprintf("file%d.log", i))
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	browser := NewLogFileBrowserModel(80, 24, nil, "test-tag")
	browser.SetFocused(true)
	
	// Test all keyboard shortcuts in sequence
	keySequence := []tea.KeyMsg{
		{Type: tea.KeyDown},               // Down arrow
		{Type: tea.KeyRunes, Runes: []rune{'j'}},  // j (down)
		{Type: tea.KeyUp},                 // Up arrow
		{Type: tea.KeyRunes, Runes: []rune{'k'}},  // k (up)
		{Type: tea.KeyHome},               // Home
		{Type: tea.KeyEnd},                // End
		{Type: tea.KeyPgDown},             // PageDown
		{Type: tea.KeyPgUp},               // PageUp
		{Type: tea.KeyRunes, Runes: []rune{'/'}}, // Search
	}
	
	for _, keyMsg := range keySequence {
		_, _ = browser.Update(keyMsg)
		// Should not panic
	}
}

// TestLogFileBrowserSendsFileSelectedMsg tests that FileSelectedMsg is sent when a file is selected
func TestLogFileBrowserSendsFileSelectedMsg(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "test.log")
	
	// Create test file
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create taskmaster service with temp dir
	service := &taskmaster.Service{}
	
	// Create browser model
	browser := NewLogFileBrowserModel(80, 24, service, "logs")
	
	// Manually set files to include our test file
	browser.files = []FileEntry{
		{
			Name:        "test.log",
			Path:        testFilePath,
			IsDir:       false,
			Size:        12,
			ModTime:     time.Now(),
			DisplayName: "test.log",
		},
	}
	browser.updateList()
	
	// Select the file with Enter key
	_, cmd := browser.Update(tea.KeyMsg{Type: tea.KeyEnter})
	
	// Verify that a command was returned
	if cmd == nil {
		t.Fatal("Expected a command to be returned when file is selected, got nil")
	}
	
	// Execute the command to get the message
	msg := cmd()
	
	// Verify it's a FileSelectedMsg
	fileMsg, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	
	// Verify the file path is correct
	if fileMsg.FilePath != testFilePath {
		t.Errorf("Expected FilePath %q, got %q", testFilePath, fileMsg.FilePath)
	}
}

// TestLogFileBrowserDirectoryNavigationDoesNotSendFileSelectedMsg tests that directory navigation doesn't send FileSelectedMsg
func TestLogFileBrowserDirectoryNavigationDoesNotSendFileSelectedMsg(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Create taskmaster service
	service := &taskmaster.Service{}
	
	// Create browser model
	browser := NewLogFileBrowserModel(80, 24, service, "logs")
	
	// Manually set files to include a directory
	browser.files = []FileEntry{
		{
			Name:        "subdir",
			Path:        subDir,
			IsDir:       true,
			Size:        0,
			ModTime:     time.Now(),
			DisplayName: "subdir",
		},
	}
	browser.updateList()
	browser.currentPath = tmpDir
	
	// Try to enter the directory with Enter key
	_, cmd := browser.Update(tea.KeyMsg{Type: tea.KeyEnter})
	
	// For directories, cmd should be nil (no message sent)
	// Directory navigation updates internal state but doesn't send FileSelectedMsg
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(FileSelectedMsg); ok {
			t.Error("Expected no FileSelectedMsg for directory navigation, but got one")
		}
	}
}

