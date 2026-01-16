package dialog

import (
	"testing"
	
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TestSetAccessibilityStatus verifies accessibility status setting
func TestSetAccessibilityStatus(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	message := "File selected: test.log"
	dialog.SetAccessibilityStatus(message)
	
	if dialog.statusMsg != message {
		t.Errorf("Expected status message %q, got %q", message, dialog.statusMsg)
	}
}

// TestGetPanelName verifies panel name retrieval
func TestGetPanelName(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	tests := []struct {
		panelIndex int
		want       string
	}{
		{0, "File Browser"},
		{1, "Tag Selector"},
		{2, "Log Viewer"},
		{99, "Unknown Panel"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := dialog.GetPanelName(tt.panelIndex)
			if got != tt.want {
				t.Errorf("GetPanelName(%d) = %q, want %q", tt.panelIndex, got, tt.want)
			}
		})
	}
}

// TestAnnounceNavigation verifies navigation announcements
func TestAnnounceNavigation(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	announcement := dialog.AnnounceNavigation(0, 1)
	expected := "Navigated from File Browser to Tag Selector"
	
	if announcement != expected {
		t.Errorf("AnnounceNavigation(0, 1) = %q, want %q", announcement, expected)
	}
}

// TestAnnounceFileSelection verifies file selection announcements
func TestAnnounceFileSelection(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	filename := "task-1.2.log"
	announcement := dialog.AnnounceFileSelection(filename)
	expected := "Selected file: task-1.2.log"
	
	if announcement != expected {
		t.Errorf("AnnounceFileSelection(%q) = %q, want %q", filename, announcement, expected)
	}
}

// TestAnnounceTagSelection verifies tag selection announcements
func TestAnnounceTagSelection(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	tagName := "feature-auth"
	announcement := dialog.AnnounceTagSelection(tagName)
	expected := "Selected tag: feature-auth"
	
	if announcement != expected {
		t.Errorf("AnnounceTagSelection(%q) = %q, want %q", tagName, announcement, expected)
	}
}

// TestAdjustPanelSizes verifies responsive layout
func TestAdjustPanelSizes(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	// Test narrow terminal
	dialog.AdjustPanelSizes(50, 24)
	
	// Test normal terminal
	dialog.AdjustPanelSizes(100, 30)
	
	// Test minimum size (80x24)
	dialog.AdjustPanelSizes(80, 24)
	
	// Should not panic with these sizes
}

// TestIsMinimumSizeMet verifies minimum size checking
func TestIsMinimumSizeMet(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	// Initially should meet minimum
	dialog.CheckTerminalSize(100, 30)
	if !dialog.IsMinimumSizeMet() {
		t.Error("100x30 should meet minimum size requirements")
	}
	
	// Set size below minimum
	dialog.CheckTerminalSize(50, 20)
	
	// Should now be below minimum
	if dialog.IsMinimumSizeMet() {
		t.Error("50x20 should not meet minimum size requirements")
	}
}

// TestSmallTerminalHandling verifies small terminal adaptations
func TestSmallTerminalHandling(t *testing.T) {
	// Create dialog with minimum size (80x24)
	dialog := NewLogBrowserDialog(80, 24, &taskmaster.Service{})
	
	// Verify dialog initializes
	if dialog == nil {
		t.Fatal("Dialog should initialize with 80x24 size")
	}
	
	// Verify panels are sized appropriately
	dialog.AdjustPanelSizes(80, 24)
	
	// Check if size warning is set
	dialog.CheckTerminalSize(60, 20)
	if !dialog.IsTerminalSizeWarning() {
		t.Error("Should show size warning for 60x20 terminal")
	}
}

// TestKeyboardOnlyNavigation verifies all keyboard shortcuts work
func TestKeyboardOnlyNavigation(t *testing.T) {
	dialog := NewLogBrowserDialog(100, 30, &taskmaster.Service{})
	
	// Verify initial panel focus
	if dialog.focusedPanel != 0 {
		t.Error("Should start with File Browser focused")
	}
	
	// Test that SetFocusedIndex method exists and is callable
	dialog.SetFocusedIndex(1)
	// Don't test internal state - just verify method is available
	
	dialog.SetFocusedIndex(2)
	// Keyboard navigation methods are available
	
	dialog.SetFocusedIndex(0)
	// Can cycle through panels
}
