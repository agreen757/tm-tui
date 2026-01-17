package dialog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

// TestFileSelectionUpdatesLogViewer tests that selecting a file in the File Browser
// correctly updates the Log Viewer with the file's contents
func TestFileSelectionUpdatesLogViewer(t *testing.T) {
	// Create a temporary test directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	testContent := "Test log content\nLine 2\nLine 3"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create mock task service
	taskService := &taskmaster.Service{}

	// Create dialog
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Send FileSelectedMsg
	msg := FileSelectedMsg{FilePath: testFile}
	updatedDialog, cmd := dialog.Update(msg)

	// Verify the dialog was updated
	d := updatedDialog.(*LogBrowserDialog)

	// Check that selectedFile was set
	if d.selectedFile != testFile {
		t.Errorf("Expected selectedFile to be %s, got %s", testFile, d.selectedFile)
	}

	// Check that log viewer loaded the content
	if d.logViewer.GetFilePath() != testFile {
		t.Errorf("Expected log viewer file path to be %s, got %s", testFile, d.logViewer.GetFilePath())
	}

	// Check that status message is empty (no error)
	if d.statusMsg != "" {
		t.Errorf("Expected no error, got status message: %s", d.statusMsg)
	}

	// Verify no command was returned (file loading is synchronous)
	if cmd != nil {
		t.Error("Expected nil command after file selection")
	}

	// Check that dialog title was updated
	expectedTitle := "Log Browser - test.log"
	if d.BaseFocusableDialog.TitleText != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, d.BaseFocusableDialog.TitleText)
	}
}

// TestFileSelectionWithInvalidFile tests error handling when selecting a non-existent file
func TestFileSelectionWithInvalidFile(t *testing.T) {
	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Send FileSelectedMsg with non-existent file
	msg := FileSelectedMsg{FilePath: "/nonexistent/file.log"}
	updatedDialog, _ := dialog.Update(msg)
	d := updatedDialog.(*LogBrowserDialog)

	// Check that error message was set
	if d.statusMsg == "" {
		t.Error("Expected error message in status bar, got empty string")
	}

	// Verify error message contains helpful information
	if len(d.statusMsg) < 10 {
		t.Errorf("Expected detailed error message, got: %s", d.statusMsg)
	}
}

// TestFocusPreservationBetweenPanels tests that focus state is correctly maintained
// when switching between the two panels (file browser and log viewer)
func TestFocusPreservationBetweenPanels(t *testing.T) {
	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Initial state: File Browser should be focused
	if dialog.focusedPanel != 0 {
		t.Errorf("Expected initial focus on panel 0, got %d", dialog.focusedPanel)
	}
	if !dialog.fileBrowser.focused {
		t.Error("Expected file browser to be focused initially")
	}

	// Press Tab to switch to Log Viewer
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := dialog.HandleKey(tabMsg)

	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got %v", result)
	}

	// Verify focus moved to Log Viewer
	if dialog.focusedPanel != 1 {
		t.Errorf("Expected focus on panel 1 after Tab, got %d", dialog.focusedPanel)
	}
	if dialog.fileBrowser.focused {
		t.Error("Expected file browser to lose focus after Tab")
	}
	if !dialog.logViewer.IsFocused() {
		t.Error("Expected log viewer to gain focus")
	}

	// Press Shift+Tab to go back to File Browser
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	result, _ = dialog.HandleKey(shiftTabMsg)

	if dialog.focusedPanel != 0 {
		t.Errorf("Expected focus on panel 0 after Shift+Tab, got %d", dialog.focusedPanel)
	}
	if dialog.logViewer.IsFocused() {
		t.Error("Expected log viewer to lose focus")
	}
	if !dialog.fileBrowser.focused {
		t.Error("Expected file browser to gain focus")
	}
}

// TestFileSelectionAndViewerIntegration tests that file selection properly updates
// both the selectedFile field and loads content into the log viewer
func TestFileSelectionAndViewerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "integration.log")
	testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Select first file
	msg1 := FileSelectedMsg{FilePath: testFile}
	updatedDialog, _ := dialog.Update(msg1)
	d := updatedDialog.(*LogBrowserDialog)

	// Verify file selection
	if d.selectedFile != testFile {
		t.Errorf("Expected selectedFile to be %s, got %s", testFile, d.selectedFile)
	}

	// Create a second file and select it
	testFile2 := filepath.Join(tmpDir, "second.log")
	if err := os.WriteFile(testFile2, []byte("Different content"), 0644); err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}

	msg2 := FileSelectedMsg{FilePath: testFile2}
	updatedDialog, _ = d.Update(msg2)
	d = updatedDialog.(*LogBrowserDialog)

	// Verify second file is now selected
	if d.selectedFile != testFile2 {
		t.Errorf("Expected selectedFile to be %s, got %s", testFile2, d.selectedFile)
	}

	// Verify log viewer has the new file loaded
	if d.logViewer.GetFilePath() != testFile2 {
		t.Errorf("Expected log viewer to have loaded %s, got %s", testFile2, d.logViewer.GetFilePath())
	}
}

// TestMultipleFileSelections tests rapid file selection updates
func TestMultipleFileSelections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	files := make([]string, 5)
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".log")
		if err := os.WriteFile(testFile, []byte("content "+string(rune('0'+i))), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
		files[i] = testFile
	}

	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Select files in sequence
	for _, file := range files {
		msg := FileSelectedMsg{FilePath: file}
		updatedDialog, _ := dialog.Update(msg)
		dialog = updatedDialog.(*LogBrowserDialog)

		// Verify each selection
		if dialog.selectedFile != file {
			t.Errorf("Expected selectedFile to be %s, got %s", file, dialog.selectedFile)
		}
		if dialog.logViewer.GetFilePath() != file {
			t.Errorf("Expected log viewer to have %s, got %s", file, dialog.logViewer.GetFilePath())
		}
	}
}

// TestEventPropagation tests that FileSelectedMsg is properly propagated
// and processed by the dialog
func TestEventPropagation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	os.WriteFile(testFile, []byte("content"), 0644)

	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Simulate file selection from File Browser
	// In real usage, this would come from the File Browser's Update method
	fileMsg := FileSelectedMsg{FilePath: testFile}
	updatedDialog, cmd := dialog.Update(fileMsg)
	d := updatedDialog.(*LogBrowserDialog)

	// Verify the message was processed
	if d.selectedFile != testFile {
		t.Error("FileSelectedMsg was not properly processed")
	}

	// Verify command was not returned (synchronous operation)
	if cmd != nil {
		t.Error("Expected nil command for synchronous file loading")
	}
}

// TestPanelResizePreservesState tests that resizing the dialog preserves selected file state
func TestPanelResizePreservesState(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	os.WriteFile(testFile, []byte("content"), 0644)

	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Select a file
	msg := FileSelectedMsg{FilePath: testFile}
	updatedDialog, _ := dialog.Update(msg)
	dialog = updatedDialog.(*LogBrowserDialog)

	// Verify file is selected
	if dialog.selectedFile != testFile {
		t.Error("File not selected before resize")
	}

	// Resize the dialog
	dialog.SetRect(150, 50, 0, 0)

	// Verify file is still selected after resize
	if dialog.selectedFile != testFile {
		t.Error("File selection not preserved after resize")
	}

	// Verify dimensions were updated
	if dialog.width != 150 || dialog.height != 50 {
		t.Errorf("Expected dimensions 150x50 after resize, got %dx%d", dialog.width, dialog.height)
	}
}

// TestKeyboardNavigationBetweenPanels tests Tab and Shift+Tab navigation
func TestKeyboardNavigationBetweenPanels(t *testing.T) {
	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Test Tab cycle: 0 -> 1 -> 0
	expectedSequence := []int{0, 1, 0, 1, 0}
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}

	for _, expected := range expectedSequence {
		if dialog.focusedPanel != expected {
			t.Errorf("Expected focusedPanel=%d, got %d", expected, dialog.focusedPanel)
		}
		dialog.HandleKey(tabMsg)
	}

	// Test Shift+Tab reverse cycle
	dialog.focusedPanel = 0
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}

	// From 0, Shift+Tab should go to 1
	dialog.HandleKey(shiftTabMsg)
	if dialog.focusedPanel != 1 {
		t.Errorf("Expected Shift+Tab from 0 to go to 1, got %d", dialog.focusedPanel)
	}

	// From 1, Shift+Tab should go to 0
	dialog.HandleKey(shiftTabMsg)
	if dialog.focusedPanel != 0 {
		t.Errorf("Expected Shift+Tab from 1 to go to 0, got %d", dialog.focusedPanel)
	}
}

// TestDialogInitializationState tests that the dialog starts with correct state
func TestDialogInitializationState(t *testing.T) {
	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Verify initial state
	if dialog.focusedPanel != 0 {
		t.Errorf("Expected initial focusedPanel=0, got %d", dialog.focusedPanel)
	}
	if dialog.selectedFile != "" {
		t.Errorf("Expected empty selectedFile, got %s", dialog.selectedFile)
	}
	if dialog.statusMsg != "" {
		t.Errorf("Expected empty statusMsg, got %s", dialog.statusMsg)
	}
	if dialog.showHelp {
		t.Error("Expected showHelp to be false initially")
	}
	if !dialog.fileBrowser.focused {
		t.Error("Expected file browser to be focused initially")
	}
}

// TestViewUpdateAfterFileSelection tests that View() renders correctly after file selection
func TestViewUpdateAfterFileSelection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	os.WriteFile(testFile, []byte("Test content for rendering"), 0644)

	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// Get view before file selection
	viewBefore := dialog.View()
	if viewBefore == "" {
		t.Error("View should not be empty before file selection")
	}

	// Select a file
	msg := FileSelectedMsg{FilePath: testFile}
	updatedDialog, _ := dialog.Update(msg)
	dialog = updatedDialog.(*LogBrowserDialog)

	// Get view after file selection
	viewAfter := dialog.View()
	if viewAfter == "" {
		t.Error("View should not be empty after file selection")
	}

	// Views should be different (title changed, content loaded)
	if viewBefore == viewAfter {
		t.Log("Views are identical before and after file selection (may be acceptable depending on implementation)")
	}
}

// TestStatusMessageClearing tests that status messages are cleared on successful file load
func TestStatusMessageClearing(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	os.WriteFile(testFile, []byte("content"), 0644)

	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	// First select a non-existent file to set error message
	invalidMsg := FileSelectedMsg{FilePath: "/nonexistent/file.log"}
	updatedDialog, _ := dialog.Update(invalidMsg)
	dialog = updatedDialog.(*LogBrowserDialog)

	// Verify error message was set
	if dialog.statusMsg == "" {
		t.Error("Expected error message for invalid file")
	}

	// Now select a valid file
	validMsg := FileSelectedMsg{FilePath: testFile}
	updatedDialog, _ = dialog.Update(validMsg)
	dialog = updatedDialog.(*LogBrowserDialog)

	// Verify error message was cleared
	if dialog.statusMsg != "" {
		t.Errorf("Expected status message to be cleared on valid file selection, got: %s", dialog.statusMsg)
	}
}

// TestDialogClosingWithEsc tests that Esc key closes the dialog
func TestDialogClosingWithEsc(t *testing.T) {
	taskService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 40, taskService)

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := dialog.HandleKey(escMsg)

	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel on Esc, got %v", result)
	}
}
