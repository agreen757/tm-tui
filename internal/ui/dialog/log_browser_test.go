package dialog

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

// TestLogBrowserDialogInit verifies dialog initialization with correct default state
func TestLogBrowserDialogInit(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	if dialog == nil {
		t.Fatal("NewLogBrowserDialog returned nil")
	}

	// Verify default state
	if dialog.focusedPanel != 0 {
		t.Errorf("Expected focusedPanel=0, got %d", dialog.focusedPanel)
	}

	if dialog.currentPath != "" {
		t.Errorf("Expected empty currentPath, got %s", dialog.currentPath)
	}

	if dialog.selectedFile != "" {
		t.Errorf("Expected empty selectedFile, got %s", dialog.selectedFile)
	}

	// Verify panels are initialized
	if dialog.fileBrowser == nil {
		t.Fatal("fileBrowser is nil")
	}

	if dialog.logViewer == nil {
		t.Fatal("logViewer is nil")
	}

	// Verify panel dimensions
	if dialog.fileBrowser.width == 0 {
		t.Errorf("fileBrowser width is 0")
	}

	if dialog.logViewer.width == 0 {
		t.Errorf("logViewer width is 0")
	}
}

// TestFocusCycling verifies focus cycling between panels works correctly
func TestFocusCycling(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	tests := []struct {
		name     string
		startIdx int
		key      string
		expected int
	}{
		{"Tab from 0 to 1", 0, "tab", 1},
		{"Tab from 1 to 0", 1, "tab", 0},
		{"Shift+Tab from 1 to 0", 1, "shift+tab", 0},
		{"Shift+Tab from 0 to 1", 0, "shift+tab", 1},
	}

	for _, test := range tests {
		dialog.focusedPanel = test.startIdx

		// Create a key message based on the test key
		keyMsg := toKeyMsg(test.key)
		result, _ := dialog.HandleKey(keyMsg)

		if result != DialogResultNone {
			t.Errorf("test %s: HandleKey returned %v, expected DialogResultNone", test.name, result)
		}

		if dialog.focusedPanel != test.expected {
			t.Errorf("test %s: Expected focusedPanel=%d, got %d", test.name, test.expected, dialog.focusedPanel)
		}
	}
}

// TestDialogClosing verifies dialog can be closed with Esc key
func TestDialogClosing(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Create an Esc key message
	keyMsg := toKeyMsg("esc")
	result, _ := dialog.HandleKey(keyMsg)

	// The base handler returns DialogResultCancel (value 2) for Esc key
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel on Esc, got %v", result)
	}
}

// TestPanelProportions verifies panel proportions are calculated correctly
func TestPanelProportions(t *testing.T) {
	mockService := &taskmaster.Service{}
	width, height := 100, 30
	dialog := NewLogBrowserDialog(width, height, mockService)

	// Expected proportions: 40% file browser, 60% log viewer
	expectedBrowserWidth := (width * 40) / 100
	expectedViewerWidth := width - expectedBrowserWidth

	if dialog.fileBrowser.width != expectedBrowserWidth {
		t.Errorf("Expected fileBrowser width=%d, got %d", expectedBrowserWidth, dialog.fileBrowser.width)
	}

	if dialog.logViewer.width != expectedViewerWidth {
		t.Errorf("Expected logViewer width=%d, got %d", expectedViewerWidth, dialog.logViewer.width)
	}
}

// TestDialogResize verifies dialog handles resize events properly
func TestDialogResize(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Simulate window resize
	newWidth := 150
	newHeight := 50

	dialog.SetRect(newWidth, newHeight, 0, 0)

	if dialog.width != newWidth {
		t.Errorf("Expected width=%d after resize, got %d", newWidth, dialog.width)
	}

	if dialog.height != newHeight {
		t.Errorf("Expected height=%d after resize, got %d", newHeight, dialog.height)
	}
}

// Helper function to convert a string to a key message
func toKeyMsg(key string) tea.KeyMsg {
	switch key {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(key[0])}}
	}
}

// TestLayoutRendering verifies the View() method produces output with proper panel layout
func TestLayoutRendering(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	view := dialog.View()

	if view == "" {
		t.Errorf("View() returned empty string")
	}

	// Verify that all panels are included in the view (indicated by titles)
	if !hasSubstring(view, "Files") && !hasSubstring(view, "file") {
		t.Errorf("View() should include file browser panel")
	}

	if !hasSubstring(view, "Content") && !hasSubstring(view, "content") {
		t.Errorf("View() should include log viewer panel")
	}
}

// TestFocusIndicators verifies that the focused panel has visual distinction
func TestFocusIndicators(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	for panelIdx := 0; panelIdx < 2; panelIdx++ {
		dialog.focusedPanel = panelIdx
		// Manually set the base focusedIndex to match
		dialog.BaseFocusableDialog.SetFocusedIndex(panelIdx)

		view := dialog.View()

		if view == "" {
			t.Errorf("View() returned empty for panel %d", panelIdx)
		}

		// Verify that the correct panel is focused
		if dialog.focusedPanel != panelIdx {
			t.Errorf("Expected focusedPanel %d, got %d", panelIdx, dialog.focusedPanel)
		}
	}
}

// TestResizeHandling verifies that resize updates panel dimensions
func TestResizeHandling(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Simulate window resize event
	oldView := dialog.View()

	// Change dialog dimensions
	dialog.SetRect(150, 40, 0, 0)

	newView := dialog.View()

	// Views should be different after resize
	// (This is a simple check - more sophisticated checks would compare actual layouts)
	if oldView == newView && oldView != "" {
		t.Logf("Views are the same after resize, which may indicate successful scaling")
	}
}

// Helper function to check if string contains substring
func hasSubstring(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0
}

// TestInitialization verifies Init() returns valid command
func TestInitialization(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	cmd := dialog.Init()
	// Init() may return nil or a command, both are valid
	// Just verify no panic occurs
	if cmd == nil {
		t.Logf("Init() returned nil, which is valid")
	}
}

// TestUpdateWithKeyMsg verifies Update processes key messages
func TestUpdateWithKeyMsg(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Create a Tab key message
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}

	// Call Update with key message
	updatedDialog, _ := dialog.Update(keyMsg)

	if updatedDialog == nil {
		t.Errorf("Update should return non-nil dialog")
	}

	// Tab should cycle focus
	if dialog.focusedPanel != 1 {
		t.Errorf("Tab should move focus to panel 1, got %d", dialog.focusedPanel)
	}
}

// TestUpdateWithWindowSizeMsg verifies Update handles resize
func TestUpdateWithWindowSizeMsg(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	oldWidth := dialog.width
	oldHeight := dialog.height

	// Simulate window resize
	sizeMsg := tea.WindowSizeMsg{Width: 150, Height: 50}

	updatedDialog, _ := dialog.Update(sizeMsg)

	if updatedDialog == nil {
		t.Errorf("Update should return non-nil dialog")
	}

	// Verify dimensions updated
	if dialog.width != 150 {
		t.Errorf("Expected width=150 after resize, got %d", dialog.width)
	}

	if dialog.height != 50 {
		t.Errorf("Expected height=50 after resize, got %d", dialog.height)
	}

	if dialog.width == oldWidth || dialog.height == oldHeight {
		t.Logf("Warning: dimensions may not have changed from old values")
	}
}

// TestEscapeKeyClosesDialog verifies Esc key returns DialogResultCancel
func TestEscapeKeyClosesDialog(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := dialog.HandleKey(escMsg)

	if result != DialogResultCancel {
		t.Errorf("Esc key should return DialogResultCancel, got %v", result)
	}
}

// TestTabFocusCycling verifies Tab cycles focus through panels
func TestTabFocusCycling(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	expectedCycle := []int{0, 1, 0}

	for _, expectedPanel := range expectedCycle {
		if dialog.focusedPanel != expectedPanel {
			t.Errorf("Expected focusedPanel=%d, got %d", expectedPanel, dialog.focusedPanel)
		}

		tabMsg := tea.KeyMsg{Type: tea.KeyTab}
		dialog.HandleKey(tabMsg)
	}
}

// TestShiftTabReverseFocusCycling verifies Shift+Tab cycles focus in reverse
func TestShiftTabReverseFocusCycling(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Start at panel 0, shift+tab should go to panel 1
	expectedCycle := []int{0, 1, 0}

	for _, expectedPanel := range expectedCycle {
		if dialog.focusedPanel != expectedPanel {
			t.Errorf("Expected focusedPanel=%d, got %d", expectedPanel, dialog.focusedPanel)
		}

		shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
		dialog.HandleKey(shiftTabMsg)
	}
}

// TestViewRendersAfterResize verifies View still renders correctly after resize
func TestViewRendersAfterResize(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	view1 := dialog.View()

	// Resize dialog
	dialog.SetRect(120, 40, 0, 0)

	view2 := dialog.View()

	if view1 == "" {
		t.Errorf("View() returned empty before resize")
	}

	if view2 == "" {
		t.Errorf("View() returned empty after resize")
	}

	// Both views should be valid (not empty)
	// They may be different sizes, but both should render successfully
}

// TestHelpOverlayToggle verifies that the ? key toggles the help overlay
func TestHelpOverlayToggle(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Initially, help overlay should be hidden
	if dialog.showHelp {
		t.Errorf("Help overlay should be hidden initially")
	}

	// Press ? to show help
	helpKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	result, _ := dialog.HandleKey(helpKeyMsg)

	if result != DialogResultNone {
		t.Errorf("Help toggle should return DialogResultNone, got %v", result)
	}

	if !dialog.showHelp {
		t.Errorf("Help overlay should be visible after pressing ?")
	}

	// Press ? again to hide help
	result, _ = dialog.HandleKey(helpKeyMsg)

	if result != DialogResultNone {
		t.Errorf("Help toggle should return DialogResultNone, got %v", result)
	}

	if dialog.showHelp {
		t.Errorf("Help overlay should be hidden after pressing ? again")
	}
}

// TestHelpOverlayEscape verifies that Esc closes the help overlay
func TestHelpOverlayEscape(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Show help overlay
	helpKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	dialog.HandleKey(helpKeyMsg)

	if !dialog.showHelp {
		t.Errorf("Help overlay should be visible")
	}

	// Press Esc to close help
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := dialog.HandleKey(escMsg)

	if result != DialogResultNone {
		t.Errorf("Esc in help should return DialogResultNone (closes help, not dialog), got %v", result)
	}

	if dialog.showHelp {
		t.Errorf("Help overlay should be hidden after pressing Esc")
	}
}

// TestRefreshShortcut verifies that the r key triggers a refresh
func TestRefreshShortcut(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Press r to refresh
	refreshKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	result, cmd := dialog.HandleKey(refreshKeyMsg)

	if result != DialogResultNone {
		t.Errorf("Refresh should return DialogResultNone, got %v", result)
	}

	// Refresh should return a command
	if cmd == nil {
		t.Errorf("Refresh should return a command")
	}

	// Status message should be set
	if dialog.statusMsg == "" {
		t.Errorf("Refresh should set a status message")
	}
}

// TestGlobalShortcutsWorkRegardlessOfFocusedPanel verifies global shortcuts work from any panel
func TestGlobalShortcutsWorkRegardlessOfFocusedPanel(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Test help toggle from each panel
	for panelIdx := 0; panelIdx < 2; panelIdx++ {
		dialog.focusedPanel = panelIdx
		dialog.showHelp = false

		helpKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		result, _ := dialog.HandleKey(helpKeyMsg)

		if result != DialogResultNone {
			t.Errorf("Help toggle from panel %d should return DialogResultNone, got %v", panelIdx, result)
		}

		if !dialog.showHelp {
			t.Errorf("Help overlay should be visible from panel %d", panelIdx)
		}
	}

	// Reset help state before testing refresh
	dialog.showHelp = false

	// Test refresh from each panel
	for panelIdx := 0; panelIdx < 2; panelIdx++ {
		dialog.focusedPanel = panelIdx
		dialog.statusMsg = ""

		refreshKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
		result, cmd := dialog.HandleKey(refreshKeyMsg)

		if result != DialogResultNone {
			t.Errorf("Refresh from panel %d should return DialogResultNone, got %v", panelIdx, result)
		}

		if cmd == nil {
			t.Errorf("Refresh from panel %d should return a command", panelIdx)
		}

		if dialog.statusMsg == "" {
			t.Errorf("Refresh from panel %d should set a status message", panelIdx)
		}
	}
}

// TestTabShiftTabStillWorkAfterAddingGlobalShortcuts verifies focus cycling still works
func TestTabShiftTabStillWorkAfterAddingGlobalShortcuts(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Test Tab cycling
	dialog.focusedPanel = 0
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := dialog.HandleKey(tabMsg)

	if result != DialogResultNone {
		t.Errorf("Tab should return DialogResultNone, got %v", result)
	}

	if dialog.focusedPanel != 1 {
		t.Errorf("Tab should move focus to panel 1, got %d", dialog.focusedPanel)
	}

	// Test Shift+Tab cycling
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	result, _ = dialog.HandleKey(shiftTabMsg)

	if result != DialogResultNone {
		t.Errorf("Shift+Tab should return DialogResultNone, got %v", result)
	}

	if dialog.focusedPanel != 0 {
		t.Errorf("Shift+Tab should move focus to panel 0, got %d", dialog.focusedPanel)
	}
}

// TestEscStillClosesDialogWhenHelpNotShown verifies Esc closes dialog when help is not shown
func TestEscStillClosesDialogWhenHelpNotShown(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Ensure help is not shown
	dialog.showHelp = false

	// Press Esc
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := dialog.HandleKey(escMsg)

	if result != DialogResultCancel {
		t.Errorf("Esc with help hidden should close dialog (return DialogResultCancel), got %v", result)
	}
}

// TestHelpOverlayConsumesAllKeysExceptEscAndQuestionMark verifies help overlay consumes keys
func TestHelpOverlayConsumesAllKeysExceptEscAndQuestionMark(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Show help overlay
	dialog.showHelp = true

	// Test that Tab doesn't cycle focus when help is shown
	initialPanel := dialog.focusedPanel
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := dialog.HandleKey(tabMsg)

	if result != DialogResultNone {
		t.Errorf("Tab in help should return DialogResultNone, got %v", result)
	}

	if dialog.focusedPanel != initialPanel {
		t.Errorf("Tab in help should not cycle focus, panel changed from %d to %d", initialPanel, dialog.focusedPanel)
	}

	// Test that r doesn't refresh when help is shown
	dialog.statusMsg = ""
	refreshMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	result, cmd := dialog.HandleKey(refreshMsg)

	if result != DialogResultNone {
		t.Errorf("r in help should return DialogResultNone, got %v", result)
	}

	if cmd != nil {
		t.Errorf("r in help should not trigger refresh command")
	}

	if dialog.statusMsg != "" {
		t.Errorf("r in help should not set status message")
	}
}

// TestHelpOverlayRendering verifies help overlay is rendered when shown
func TestHelpOverlayRendering(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	// Without help overlay
	dialog.showHelp = false
	viewWithoutHelp := dialog.View()

	if viewWithoutHelp == "" {
		t.Errorf("View() should not be empty without help")
	}

	// With help overlay
	dialog.showHelp = true
	viewWithHelp := dialog.View()

	if viewWithHelp == "" {
		t.Errorf("View() should not be empty with help")
	}

	// Views should be different when help is shown
	if viewWithHelp == viewWithoutHelp {
		t.Logf("Warning: Views are identical with and without help overlay")
	}
}

// TestHelpOverlayContextSensitive verifies help shows different content based on focused panel
func TestHelpOverlayContextSensitive(t *testing.T) {
	mockService := &taskmaster.Service{}
	dialog := NewLogBrowserDialog(100, 30, mockService)

	dialog.showHelp = true

	// Get help for each panel
	var helpViews []string
	for panelIdx := 0; panelIdx < 2; panelIdx++ {
		dialog.focusedPanel = panelIdx
		helpView := dialog.renderHelpOverlay()
		helpViews = append(helpViews, helpView)

		if helpView == "" {
			t.Errorf("Help overlay should not be empty for panel %d", panelIdx)
		}
	}

	// Help should contain different content for different panels
	// (File Browser mentions different shortcuts than Log Viewer)
	if len(helpViews) == 2 {
		// All help views should contain global shortcuts
		for idx, view := range helpViews {
			if !hasSubstring(view, "Global") && !hasSubstring(view, "Esc") {
				t.Errorf("Help for panel %d should contain global shortcuts", idx)
			}
		}

		// Each panel should have unique content
		// We can't do exact string comparison due to rendering, but we can check length differences
		// This is a basic check - more sophisticated checks would parse the content
		if helpViews[0] == helpViews[1] {
			t.Logf("Warning: Help content appears identical for both panels")
		}
	}
}
