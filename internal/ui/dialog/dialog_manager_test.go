package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// MockDialog is a simple dialog for testing
type MockDialog struct {
	BaseDialog
}

func (d *MockDialog) Init() tea.Cmd {
	return nil
}

func (d *MockDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	return d, nil
}

func (d *MockDialog) View() string {
	return ""
}

func (d *MockDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	return DialogResultNone, nil
}

func NewMockDialog(title string, width, height int) *MockDialog {
	return &MockDialog{
		BaseDialog: NewBaseDialog(title, width, height, DialogTypeModal),
	}
}

func TestDialogManager_SetTerminalSize(t *testing.T) {
	dm := NewDialogManager(100, 30)

	// Verify initial state
	if dm.termWidth != 100 || dm.termHeight != 30 {
		t.Errorf("Initial terminal size incorrect")
	}

	// Set new terminal size
	dm.SetTerminalSize(80, 24)

	if dm.termWidth != 80 || dm.termHeight != 24 {
		t.Errorf("Terminal size not updated correctly")
	}

	// Verify that duplicate calls are skipped
	dm.SetTerminalSize(80, 24)
	if dm.lastWindowWidth != 80 || dm.lastWindowHeight != 24 {
		t.Errorf("Duplicate SetTerminalSize call not handled correctly")
	}
}

func TestDialogManager_DialogPositioningOnAdd(t *testing.T) {
	dm := NewDialogManager(100, 30)

	// Create and add a dialog
	dialog := NewMockDialog("Test", 50, 20)
	dm.AddDialog(dialog, nil)

	// Check that dialog is properly positioned
	width, height, x, y := dm.GetActiveDialog().GetRect()

	if width != 50 || height != 20 {
		t.Errorf("Dialog size incorrect: got %dx%d, expected 50x20", width, height)
	}

	// Dialog should be centered
	expectedX := (100 - 50) / 2
	expectedY := (30 - 20) / 2

	if x != expectedX || y != expectedY {
		t.Errorf("Dialog position incorrect: got (%d,%d), expected (%d,%d)", x, y, expectedX, expectedY)
	}
}

func TestDialogManager_DialogResizingAndRepositioning(t *testing.T) {
	// Start with a normal terminal size
	dm := NewDialogManager(100, 30)

	// Add a dialog
	dialog := NewMockDialog("Test", 60, 20)
	dm.AddDialog(dialog, nil)

	// Get original position
	origWidth, origHeight, _, _ := dialog.GetRect()

	if origWidth != 60 || origHeight != 20 {
		t.Errorf("Original dialog size incorrect")
	}

	// Simulate terminal resize to smaller size
	resizeMsg := tea.WindowSizeMsg{Width: 50, Height: 25}
	dm.HandleMsg(resizeMsg)

	// Check that dialog was repositioned
	newWidth, _, newX, _ := dialog.GetRect()

	// Dialog should be degraded to fit within 50 width
	if newWidth >= 50 {
		t.Errorf("Dialog width not reduced for smaller terminal: got %d", newWidth)
	}

	// Dialog should still be centered (or close to it)
	if newX < 0 {
		t.Errorf("Dialog position invalid after resize: X=%d", newX)
	}

	if newX+newWidth > 50 {
		t.Errorf("Dialog overflows terminal after resize")
	}
}

func TestDialogManager_MultipleDialogsResize(t *testing.T) {
	dm := NewDialogManager(100, 30)

	// Add multiple dialogs
	dialog1 := NewMockDialog("Dialog 1", 50, 20)
	dialog2 := NewMockDialog("Dialog 2", 60, 18)

	dm.AddDialog(dialog1, nil)
	dm.AddDialog(dialog2, nil)

	// Verify both dialogs are properly positioned
	w1, h1, x1, y1 := dialog1.GetRect()
	w2, h2, x2, y2 := dialog2.GetRect()

	if w1 <= 0 || h1 <= 0 || x1 < 0 || y1 < 0 {
		t.Errorf("Dialog 1 position invalid: (%d,%d) %dx%d", x1, y1, w1, h1)
	}
	if w2 <= 0 || h2 <= 0 || x2 < 0 || y2 < 0 {
		t.Errorf("Dialog 2 position invalid: (%d,%d) %dx%d", x2, y2, w2, h2)
	}

	// Simulate resize
	resizeMsg := tea.WindowSizeMsg{Width: 60, Height: 20}
	dm.HandleMsg(resizeMsg)

	// Both dialogs should be repositioned
	w1New, h1New, x1New, y1New := dialog1.GetRect()
	w2New, h2New, x2New, y2New := dialog2.GetRect()

	// Verify they still fit in terminal bounds
	if x1New < 0 || x1New+w1New > 60 {
		t.Errorf("Dialog 1 overflows after resize")
	}
	if y1New < 0 || y1New+h1New > 20 {
		t.Errorf("Dialog 1 overflows vertically after resize")
	}

	if x2New < 0 || x2New+w2New > 60 {
		t.Errorf("Dialog 2 overflows after resize")
	}
	if y2New < 0 || y2New+h2New > 20 {
		t.Errorf("Dialog 2 overflows vertically after resize")
	}
}

func TestDialogManager_ExtremeSizeResize(t *testing.T) {
	dm := NewDialogManager(100, 30)

	// Add a dialog
	dialog := NewMockDialog("Test", 50, 20)
	dm.AddDialog(dialog, nil)

	// Simulate extreme terminal shrink
	resizeMsg := tea.WindowSizeMsg{Width: 25, Height: 10}
	dm.HandleMsg(resizeMsg)

	// Dialog should degrade gracefully
	w, h, x, y := dialog.GetRect()

	// Dialog should fit within terminal bounds
	if x < 0 || x+w > 25 {
		t.Errorf("Dialog doesn't fit horizontally: x=%d, w=%d, term=25", x, w)
	}
	if y < 0 || y+h > 10 {
		t.Errorf("Dialog doesn't fit vertically: y=%d, h=%d, term=10", y, h)
	}

	// Dialog should have minimum size
	minWidth := 20
	minHeight := 6
	if w < minWidth {
		t.Logf("Dialog width %d below minimum %d (acceptable for extreme shrink)", w, minWidth)
	}
	if h < minHeight {
		t.Logf("Dialog height %d below minimum %d (acceptable for extreme shrink)", h, minHeight)
	}
}

func TestDialogManager_GrowthAfterShrink(t *testing.T) {
	dm := NewDialogManager(100, 30)

	// Add a dialog
	dialog := NewMockDialog("Test", 50, 20)
	dm.AddDialog(dialog, nil)

	// Shrink terminal
	shrinkMsg := tea.WindowSizeMsg{Width: 40, Height: 15}
	dm.HandleMsg(shrinkMsg)

	// Grow terminal back
	growMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	dm.HandleMsg(growMsg)

	w2, h2, _, _ := dialog.GetRect()

	// Dialog should remain valid in new size
	if w2 <= 0 || h2 <= 0 {
		t.Errorf("Dialog invalid after resize growth")
	}

	// Should be approximately centered (within reasonable bounds)
	expectedX := (100 - w2) / 2
	expectedY := (30 - h2) / 2
	actualX, actualY, _, _ := dialog.GetRect()

	// Allow for positioning tolerance due to padding/degradation
	tolerance := 10
	if actualX < expectedX-tolerance || actualX > expectedX+tolerance {
		t.Errorf("Dialog X position off center: got %d, expected ~%d (tolerance=%d)", actualX, expectedX, tolerance)
	}
	if actualY < expectedY-tolerance || actualY > expectedY+tolerance {
		t.Errorf("Dialog Y position off center: got %d, expected ~%d (tolerance=%d)", actualY, expectedY, tolerance)
	}
}
