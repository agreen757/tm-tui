package dialog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TestLogViewerPermissionError tests that permission denied errors are handled gracefully
func TestLogViewerPermissionError(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)

	// Test permission error detection
	err := errors.New("permission denied reading file: /root/.secret")
	viewer.loadError = err
	viewer.renderContent()

	if !viewer.IsPermissionError() {
		t.Error("Expected IsPermissionError() to return true")
	}

	view := viewer.View()
	if !containsHelper(view, "Permission Denied") {
		t.Error("Expected 'Permission Denied' in view output")
	}
	if !containsHelper(view, "Check file permissions") {
		t.Error("Expected helpful suggestion in view output")
	}
}

// TestLogViewerFileNotFoundError tests that file not found errors are handled gracefully
func TestLogViewerFileNotFoundError(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)

	// Test file not found error detection
	err := errors.New("file not found: /nonexistent/file.log")
	viewer.loadError = err
	viewer.renderContent()

	if !viewer.IsFileNotFoundError() {
		t.Error("Expected IsFileNotFoundError() to return true")
	}

	view := viewer.View()
	if !containsHelper(view, "File Not Found") {
		t.Error("Expected 'File Not Found' in view output")
	}
	if !containsHelper(view, "may have been deleted") {
		t.Error("Expected helpful suggestion in view output")
	}
}

// TestLogViewerTimeoutError tests that timeout errors are handled gracefully
func TestLogViewerTimeoutError(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)

	// Test timeout error detection
	err := errors.New("timeout reading file - file may be too large or I/O is slow")
	viewer.loadError = err
	viewer.renderContent()

	if !viewer.IsTimeoutError() {
		t.Error("Expected IsTimeoutError() to return true")
	}

	view := viewer.View()
	if !containsHelper(view, "Timeout") {
		t.Error("Expected 'Timeout' in view output")
	}
	if !containsHelper(view, "external editor") {
		t.Error("Expected helpful suggestion in view output")
	}
}

// TestLogViewerLargeFileSizeWarning tests that large files show warning
func TestLogViewerLargeFileSizeWarning(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)
	viewer.actualFileSize = 2 * 1024 * 1024 // 2MB
	viewer.fileSizeWarning = true
	viewer.content = "test content"
	viewer.isLoaded = true

	view := viewer.View()
	if !containsHelper(view, "Large file") {
		t.Error("Expected 'Large file' warning in footer")
	}
	if !containsHelper(view, "2.0 MB") {
		t.Error("Expected file size in footer")
	}
}

// TestLogViewerLineLimitWarning tests that line limit warning is shown
func TestLogViewerLineLimitWarning(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)
	viewer.content = "line1\nline2\nline3"
	viewer.lineLimited = true
	viewer.isLoaded = true

	view := viewer.View()
	if !containsHelper(view, "10,000 lines") {
		t.Error("Expected line limit warning in footer")
	}
}

// TestLogViewerEmptyStateMessages tests all empty state messages
func TestLogViewerEmptyStateMessages(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*LogViewerPanel)
		wantMsg string
	}{
		{
			name: "No file selected",
			setup: func(lv *LogViewerPanel) {
				lv.loadError = nil
				lv.isLoaded = false
			},
			wantMsg: "Select a log file",
		},
		{
			name: "File is empty",
			setup: func(lv *LogViewerPanel) {
				lv.loadError = nil
				lv.isLoaded = true
				lv.content = ""
			},
			wantMsg: "File is empty",
		},
		{
			name: "Directory error",
			setup: func(lv *LogViewerPanel) {
				lv.loadError = errors.New("path is a directory")
				lv.isLoaded = false
			},
			wantMsg: "Directory",
		},
		{
			name: "Empty path error",
			setup: func(lv *LogViewerPanel) {
				lv.loadError = errors.New("file path is empty")
				lv.isLoaded = false
			},
			wantMsg: "Empty Path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewer := NewLogViewerPanel(80, 24, nil)
			tt.setup(viewer)
			view := viewer.View()

			if !containsHelper(view, tt.wantMsg) {
				t.Errorf("Expected %q in view output, got: %s", tt.wantMsg, view)
			}
		})
	}
}

// TestFileSizeFormatting tests that file sizes are formatted correctly
func TestFileSizeFormatting(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 5, "5.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d bytes", tt.size), func(t *testing.T) {
			result := formatFileSizeForDisplay(tt.size)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLogViewerErrorTypesDetection tests error type detection methods
func TestLogViewerErrorTypesDetection(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)

	// Test with no error
	if viewer.IsPermissionError() || viewer.IsFileNotFoundError() || viewer.IsTimeoutError() {
		t.Error("Expected no errors when loadError is nil")
	}

	// Test with permission error (os.IsPermission wrapper)
	viewer.loadError = fmt.Errorf("permission denied: %w", syscall.EACCES)
	if !viewer.IsPermissionError() {
		t.Error("Expected IsPermissionError() to detect permission denied")
	}

	// Test with file not found error
	viewer.loadError = errors.New("file not found")
	if !viewer.IsFileNotFoundError() {
		t.Error("Expected IsFileNotFoundError() to detect file not found")
	}

	// Test with timeout error
	viewer.loadError = errors.New("timeout reading file")
	if !viewer.IsTimeoutError() {
		t.Error("Expected IsTimeoutError() to detect timeout")
	}
}

// TestCanReadDir tests the canReadDir helper function
func TestCanReadDir(t *testing.T) {
	// Create a temporary readable directory
	tmpDir := t.TempDir()

	if err := canReadDir(tmpDir); err != nil {
		t.Errorf("Expected no error for readable directory, got: %v", err)
	}

	// Test with non-existent directory
	nonExistent := filepath.Join(tmpDir, "nonexistent")
	if err := canReadDir(nonExistent); err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

// TestLoadFileContentWithErrors tests file loading error handling
func TestLoadFileContentWithErrors(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)

	// Test with empty path
	err := viewer.LoadFileContent("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
	if !containsHelper(err.Error(), "empty") {
		t.Errorf("Expected error message about empty path, got: %v", err)
	}

	// Test with non-existent file
	err = viewer.LoadFileContent("/nonexistent/file.log")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !viewer.IsFileNotFoundError() {
		t.Error("Expected file not found error to be detected")
	}

	// Test with directory instead of file
	tmpDir := t.TempDir()
	err = viewer.LoadFileContent(tmpDir)
	if err == nil {
		t.Error("Expected error when loading directory as file")
	}
	if !containsHelper(err.Error(), "directory") {
		t.Errorf("Expected error message about directory, got: %v", err)
	}
}

// TestLoadFileContentSuccess tests successful file loading
func TestLoadFileContentSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	// Create test file with content
	testContent := "Line 1\nLine 2\nLine 3\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	viewer := NewLogViewerPanel(80, 24, nil)
	err := viewer.LoadFileContent(testFile)

	if err != nil {
		t.Errorf("Expected no error loading valid file, got: %v", err)
	}
	if !viewer.IsLoaded() {
		t.Error("Expected file to be marked as loaded")
	}
	if !containsHelper(viewer.GetContent(), "Line 1") {
		t.Error("Expected content to be loaded")
	}
}

// TestLoadLargeFileSize tests large file size detection
func TestLoadLargeFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.log")

	// Create a large file (2MB)
	largeContent := make([]byte, 2*1024*1024)
	if err := os.WriteFile(testFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	viewer := NewLogViewerPanel(80, 24, nil)
	err := viewer.LoadFileContent(testFile)

	if err != nil {
		t.Errorf("Expected no error loading large file, got: %v", err)
	}
	if !viewer.HasFileSizeWarning() {
		t.Error("Expected file size warning for large file")
	}
	if viewer.GetActualFileSize() <= 0 {
		t.Error("Expected actual file size to be tracked")
	}
}

// TestGetActualFileSize tests file size tracking
func TestGetActualFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "size-test.log")
	testContent := "Test content"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	viewer := NewLogViewerPanel(80, 24, nil)
	viewer.LoadFileContent(testFile)

	actualSize := viewer.GetActualFileSize()
	expectedSize := int64(len(testContent))

	if actualSize != expectedSize {
		t.Errorf("Expected file size %d, got %d", expectedSize, actualSize)
	}
}

// TestClearContentResetsErrorState tests that ClearContent properly resets all state
func TestClearContentResetsErrorState(t *testing.T) {
	viewer := NewLogViewerPanel(80, 24, nil)

	// Set up error state
	viewer.loadError = errors.New("test error")
	viewer.isLoaded = true
	viewer.lineLimited = true
	viewer.fileSizeWarning = true
	viewer.actualFileSize = 2048
	viewer.content = "test"

	// Clear content
	viewer.ClearContent()

	// Verify all state is reset
	if viewer.GetError() != nil {
		t.Error("Expected loadError to be cleared")
	}
	if viewer.IsLoaded() {
		t.Error("Expected isLoaded to be false")
	}
	if viewer.IsLineLimited() {
		t.Error("Expected lineLimited to be false")
	}
	if viewer.HasFileSizeWarning() {
		t.Error("Expected fileSizeWarning to be false")
	}
	if viewer.GetActualFileSize() != 0 {
		t.Error("Expected actualFileSize to be reset")
	}
	if viewer.GetContent() != "" {
		t.Error("Expected content to be empty")
	}
}

// TestLogFileBrowserEmptyState tests that file browser shows empty state when no files
func TestLogFileBrowserEmptyState(t *testing.T) {
	tmpDir := t.TempDir()
	service := &taskmaster.Service{RootDir: tmpDir}
	browser := NewLogFileBrowserModel(80, 24, service, "test")
	// Don't load any files, browser.files should be empty

	view := browser.View()
	// The view might show the list or an empty state message
	// Just verify it doesn't panic and returns something
	if view == "" {
		t.Error("Expected non-empty view output")
	}
}

// TestLogTagSelectorEmptyState tests that tag selector shows empty state when no tags
func TestLogTagSelectorEmptyState(t *testing.T) {
	tmpDir := t.TempDir()
	service := &taskmaster.Service{RootDir: tmpDir}
	selector := NewLogTagSelectorModel(80, 24, service, "")
	// Don't discover any tags, selector.tags should be empty

	view := selector.View()
	if !containsHelper(view, "No tags available") {
		t.Error("Expected 'No tags available' message in empty state")
	}
}
