package dialog

import (
	"testing"
)

// TestBaseFocusableDialogEmbedding verifies that BaseFocusableDialog properly embeds BaseFilterable
func TestBaseFocusableDialogEmbedding(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Verify BaseFilterable is embedded and initialized
	if dialog.BaseFilterable == nil {
		t.Error("BaseFocusableDialog.BaseFilterable should not be nil")
	}

	// Verify that filtering is disabled by default
	if dialog.IsFilteringEnabled() {
		t.Error("Filtering should be disabled by default")
	}
}

// TestBaseFocusableDialogEnableFiltering verifies that EnableFiltering works through delegation
func TestBaseFocusableDialogEnableFiltering(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Enable filtering
	dialog.EnableFiltering(true)
	if !dialog.IsFilteringEnabled() {
		t.Error("EnableFiltering(true) should enable filtering")
	}

	// Disable filtering
	dialog.EnableFiltering(false)
	if dialog.IsFilteringEnabled() {
		t.Error("EnableFiltering(false) should disable filtering")
	}
}

// TestBaseFocusableDialogSetFilterValue verifies that SetFilterValue works through delegation
func TestBaseFocusableDialogSetFilterValue(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	testValue := "test filter"
	dialog.SetFilterValue(testValue)

	if dialog.GetFilterValue() != testValue {
		t.Errorf("SetFilterValue/GetFilterValue roundtrip failed: got %q want %q", dialog.GetFilterValue(), testValue)
	}
}

// TestBaseFocusableDialogFilterValueRoundtrip verifies filter value persistence
func TestBaseFocusableDialogFilterValueRoundtrip(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	testValues := []string{
		"simple",
		"multi word",
		"special-chars_123",
		"",
	}

	for _, testValue := range testValues {
		dialog.SetFilterValue(testValue)
		if dialog.GetFilterValue() != testValue {
			t.Errorf("Roundtrip failed for %q: got %q", testValue, dialog.GetFilterValue())
		}
	}
}

// TestBaseFocusableDialogIsFiltering verifies that IsFiltering returns correct state
func TestBaseFocusableDialogIsFiltering(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Initially not filtering
	if dialog.IsFiltering() {
		t.Error("IsFiltering should return false initially")
	}

	// Enable filtering and enter filter mode
	dialog.EnableFiltering(true)
	dialog.EnterFilterMode()
	if !dialog.IsFiltering() {
		t.Error("IsFiltering should return true after EnterFilterMode")
	}

	// Exit filter mode
	dialog.ExitFilterMode()
	if dialog.IsFiltering() {
		t.Error("IsFiltering should return false after ExitFilterMode")
	}
}

// TestBaseFocusableDialogEnterExitFilterMode verifies mode transitions
func TestBaseFocusableDialogEnterExitFilterMode(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)
	dialog.EnableFiltering(true)

	// Enter filter mode
	dialog.EnterFilterMode()
	if !dialog.IsFiltering() {
		t.Error("EnterFilterMode should set filtering to true")
	}

	// Exit filter mode
	dialog.ExitFilterMode()
	if dialog.IsFiltering() {
		t.Error("ExitFilterMode should set filtering to false")
	}
}

// TestBaseFocusableDialogFilterStatePersistence verifies filter state persists across show/hide cycles
func TestBaseFocusableDialogFilterStatePersistence(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Set up filter state
	dialog.EnableFiltering(true)
	testValue := "persistent filter"
	dialog.SetFilterValue(testValue)

	// Simulate a hide/show cycle by checking state is preserved
	if dialog.GetFilterValue() != testValue {
		t.Errorf("Filter value not persisted: got %q want %q", dialog.GetFilterValue(), testValue)
	}
	if !dialog.IsFilteringEnabled() {
		t.Error("Filtering enabled flag not persisted")
	}

	// Change filter state
	dialog.EnterFilterMode()
	oldFilterValue := dialog.GetFilterValue()

	// Exit filter mode
	dialog.ExitFilterMode()

	// Verify value is preserved
	if dialog.GetFilterValue() != oldFilterValue {
		t.Error("Filter value not preserved after ExitFilterMode")
	}
}

// TestBaseFocusableDialogFilterModeWithDisabledFiltering verifies that EnterFilterMode respects enabled state
func TestBaseFocusableDialogFilterModeWithDisabledFiltering(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Try to enter filter mode without enabling filtering
	dialog.EnterFilterMode()
	if dialog.IsFiltering() {
		t.Error("EnterFilterMode should not enter filter mode if filtering is disabled")
	}

	// Now enable and retry
	dialog.EnableFiltering(true)
	dialog.EnterFilterMode()
	if !dialog.IsFiltering() {
		t.Error("EnterFilterMode should enter filter mode after enabling")
	}
}

// TestBaseFocusableDialogImplementsFilterableComponent verifies interface compliance
func TestBaseFocusableDialogImplementsFilterableComponent(t *testing.T) {
	var _ FilterableComponent = (*BaseFocusableDialog)(nil)
}

// TestBaseFocusableDialogFilteringAndFocusing verifies filtering doesn't interfere with focusing
func TestBaseFocusableDialogFilteringAndFocusing(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Enable filtering
	dialog.EnableFiltering(true)
	dialog.EnterFilterMode()
	dialog.SetFilterValue("filter")

	// Verify focusing still works
	if dialog.FocusedIndex() != 0 {
		t.Errorf("FocusedIndex should be 0 initially, got %d", dialog.FocusedIndex())
	}

	// Focus next
	dialog.FocusNext()
	if dialog.FocusedIndex() != 1 {
		t.Errorf("FocusNext should increment focus, got %d", dialog.FocusedIndex())
	}

	// Verify filter state is preserved
	if dialog.GetFilterValue() != "filter" {
		t.Error("Filter value lost after focusing")
	}
	if !dialog.IsFiltering() {
		t.Error("Filter mode lost after focusing")
	}
}

// TestBaseFocusableDialogDisablingFilteringExitsMode verifies mode is exited when disabled
func TestBaseFocusableDialogDisablingFilteringExitsMode(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	// Set up filtering in filter mode
	dialog.EnableFiltering(true)
	dialog.EnterFilterMode()
	if !dialog.IsFiltering() {
		t.Error("Should be in filter mode")
	}

	// Disable filtering
	dialog.EnableFiltering(false)
	if dialog.IsFiltering() {
		t.Error("DisablingFiltering should exit filter mode")
	}
	if dialog.IsFilteringEnabled() {
		t.Error("Filtering should be disabled")
	}
}

// TestBaseFocusableDialogMultipleCycles verifies stability through repeated cycles
func TestBaseFocusableDialogMultipleCycles(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	for i := 0; i < 3; i++ {
		dialog.EnableFiltering(true)
		dialog.EnterFilterMode()
		dialog.SetFilterValue("test")

		if !dialog.IsFiltering() {
			t.Errorf("Cycle %d: Should be in filter mode", i)
		}

		dialog.ExitFilterMode()
		dialog.EnableFiltering(false)

		if dialog.IsFiltering() {
			t.Errorf("Cycle %d: Should not be in filter mode", i)
		}
		if dialog.IsFilteringEnabled() {
			t.Errorf("Cycle %d: Filtering should be disabled", i)
		}
	}
}

// TestBaseFocusableDialogCallbackDelegation verifies callback is called
func TestBaseFocusableDialogCallbackDelegation(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	callbackCalled := false
	dialog.SetOnFilterChange(func(value string) {
		callbackCalled = true
	})

	dialog.SetFilterValue("test")

	if !callbackCalled {
		t.Error("SetFilterValue should trigger callback")
	}
}

// TestBaseFocusableDialogNumElements verifies focusing with different element counts
func TestBaseFocusableDialogNumElements(t *testing.T) {
	testCases := []int{1, 5, 10, 100}

	for _, numElements := range testCases {
		dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, numElements)

		if dialog.NumFocusableElements() != numElements {
			t.Errorf("NumFocusableElements should be %d, got %d", numElements, dialog.NumFocusableElements())
		}

		// Verify filtering doesn't interfere with element count
		dialog.EnableFiltering(true)
		if dialog.NumFocusableElements() != numElements {
			t.Error("Enabling filtering should not change NumFocusableElements")
		}
	}
}

// TestBaseFocusableDialogClearFilter verifies ClearFilter method works
func TestBaseFocusableDialogClearFilter(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 5)

	dialog.EnableFiltering(true)
	dialog.EnterFilterMode()
	dialog.SetFilterValue("test value")

	// Verify we have a filter
	if dialog.GetFilterValue() != "test value" {
		t.Error("Filter value should be set")
	}
	if !dialog.IsFiltering() {
		t.Error("Should be in filter mode")
	}

	// Clear filter
	dialog.ClearFilter()

	if dialog.GetFilterValue() != "" {
		t.Error("ClearFilter should clear filter value")
	}
	if dialog.IsFiltering() {
		t.Error("ClearFilter should exit filter mode")
	}
}

// TestBaseFocusableDialogBaseDialogIntegration verifies BaseDialog methods still work
func TestBaseFocusableDialogBaseDialogIntegration(t *testing.T) {
	dialog := NewBaseFocusableDialog("TestTitle", 80, 24, DialogKindList, 5)

	// Verify BaseDialog fields are set correctly
	if dialog.Title() != "TestTitle" {
		t.Errorf("Title should be 'TestTitle', got %q", dialog.Title())
	}

	width, height, _, _ := dialog.GetRect()
	if width != 80 {
		t.Errorf("Width should be 80, got %d", width)
	}

	if height != 24 {
		t.Errorf("Height should be 24, got %d", height)
	}

	// Enable filtering and verify BaseDialog still works
	dialog.EnableFiltering(true)
	if dialog.Title() != "TestTitle" {
		t.Error("Title should be preserved after enabling filtering")
	}
}

// TestBaseFocusableDialogWithZeroElements verifies behavior with no focusable elements
func TestBaseFocusableDialogWithZeroElements(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 80, 24, DialogKindList, 0)

	if dialog.IsFocusable() {
		t.Error("IsFocusable should return false with zero elements")
	}

	if dialog.NumFocusableElements() != 0 {
		t.Error("NumFocusableElements should be 0")
	}

	// Verify filtering still works independently
	dialog.EnableFiltering(true)
	dialog.EnterFilterMode()
	if !dialog.IsFiltering() {
		t.Error("Filtering should work even with zero elements")
	}
}
