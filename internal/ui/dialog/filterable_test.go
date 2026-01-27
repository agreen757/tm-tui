package dialog

import (
	"testing"
)

// TestNewBaseFilterable verifies that NewBaseFilterable creates a properly initialized struct
func TestNewBaseFilterable(t *testing.T) {
	bf := NewBaseFilterable()

	if bf.filteringEnabled {
		t.Error("filteringEnabled should be false by default")
	}
	if bf.filterMode {
		t.Error("filterMode should be false by default")
	}
	if bf.filterValue != "" {
		t.Error("filterValue should be empty string by default")
	}
	if bf.onFilterChange != nil {
		t.Error("onFilterChange should be nil by default")
	}
}

// TestEnableFiltering verifies that EnableFiltering correctly toggles the filteringEnabled state
func TestEnableFiltering(t *testing.T) {
	bf := NewBaseFilterable()

	// Test enabling filtering
	bf.EnableFiltering(true)
	if !bf.filteringEnabled {
		t.Error("EnableFiltering(true) should set filteringEnabled to true")
	}

	// Test disabling filtering
	bf.EnableFiltering(false)
	if bf.filteringEnabled {
		t.Error("EnableFiltering(false) should set filteringEnabled to false")
	}
}

// TestEnableFilteringExitsFilterMode verifies that disabling filtering exits filter mode
func TestEnableFilteringExitsFilterMode(t *testing.T) {
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)
	bf.EnterFilterMode()

	if !bf.filterMode {
		t.Error("EnterFilterMode should set filterMode to true")
	}

	// Disabling filtering while in filter mode should exit filter mode
	bf.EnableFiltering(false)
	if bf.filterMode {
		t.Error("EnableFiltering(false) should exit filter mode when in filter mode")
	}
}

// TestSetFilterValue verifies that SetFilterValue correctly updates the filter value
func TestSetFilterValue(t *testing.T) {
	bf := NewBaseFilterable()

	testValue := "test filter"
	bf.SetFilterValue(testValue)

	if bf.filterValue != testValue {
		t.Errorf("SetFilterValue should update filterValue, got %q want %q", bf.filterValue, testValue)
	}
}

// TestGetFilterValue verifies that GetFilterValue returns the current filter value
func TestGetFilterValue(t *testing.T) {
	bf := NewBaseFilterable()

	testValue := "test filter"
	bf.SetFilterValue(testValue)
	retrieved := bf.GetFilterValue()

	if retrieved != testValue {
		t.Errorf("GetFilterValue should return current value, got %q want %q", retrieved, testValue)
	}
}

// TestFilterValueRoundtrip verifies that SetFilterValue and GetFilterValue work together
func TestFilterValueRoundtrip(t *testing.T) {
	bf := NewBaseFilterable()

	testValues := []string{
		"",
		"a",
		"filter",
		"multi word filter",
		"special chars !@#$%",
		"🎯 emoji filter",
	}

	for _, testValue := range testValues {
		bf.SetFilterValue(testValue)
		retrieved := bf.GetFilterValue()
		if retrieved != testValue {
			t.Errorf("Roundtrip failed for %q: got %q", testValue, retrieved)
		}
	}
}

// TestIsFiltering verifies that IsFiltering returns the correct filter mode state
func TestIsFiltering(t *testing.T) {
	bf := NewBaseFilterable()

	if bf.IsFiltering() {
		t.Error("IsFiltering should return false initially")
	}

	bf.EnableFiltering(true)
	bf.EnterFilterMode()
	if !bf.IsFiltering() {
		t.Error("IsFiltering should return true after EnterFilterMode")
	}

	bf.ExitFilterMode()
	if bf.IsFiltering() {
		t.Error("IsFiltering should return false after ExitFilterMode")
	}
}

// TestEnterFilterMode verifies that EnterFilterMode sets filter mode and clears filter value
func TestEnterFilterMode(t *testing.T) {
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)
	bf.SetFilterValue("old value")

	bf.EnterFilterMode()

	if !bf.filterMode {
		t.Error("EnterFilterMode should set filterMode to true")
	}
	if bf.filterValue != "" {
		t.Error("EnterFilterMode should clear filterValue")
	}
}

// TestEnterFilterModeWithoutEnabling verifies that EnterFilterMode does nothing if filtering is disabled
func TestEnterFilterModeWithoutEnabling(t *testing.T) {
	bf := NewBaseFilterable()
	// Don't enable filtering
	bf.EnterFilterMode()

	if bf.filterMode {
		t.Error("EnterFilterMode should not set filterMode if filtering is not enabled")
	}
}

// TestExitFilterMode verifies that ExitFilterMode exits filter mode
func TestExitFilterMode(t *testing.T) {
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)
	bf.EnterFilterMode()
	bf.SetFilterValue("some value")

	bf.ExitFilterMode()

	if bf.filterMode {
		t.Error("ExitFilterMode should set filterMode to false")
	}
	// Note: ExitFilterMode preserves filterValue per the implementation
	if bf.filterValue != "some value" {
		t.Error("ExitFilterMode should preserve filterValue")
	}
}

// TestSetOnFilterChange verifies that SetOnFilterChange sets the callback
func TestSetOnFilterChange(t *testing.T) {
	bf := NewBaseFilterable()

	callbackCalled := false
	bf.SetOnFilterChange(func(value string) {
		callbackCalled = true
	})

	if bf.onFilterChange == nil {
		t.Error("SetOnFilterChange should set the callback")
	}

	bf.SetFilterValue("test")

	if !callbackCalled {
		t.Error("SetFilterValue should call onFilterChange callback when set")
	}
}

// TestSetOnFilterChangeWithValue verifies that callback receives correct value
func TestSetOnFilterChangeWithValue(t *testing.T) {
	bf := NewBaseFilterable()

	receivedValue := ""
	bf.SetOnFilterChange(func(value string) {
		receivedValue = value
	})

	testValue := "test filter value"
	bf.SetFilterValue(testValue)

	if receivedValue != testValue {
		t.Errorf("Callback should receive correct value, got %q want %q", receivedValue, testValue)
	}
}

// TestSetFilterValueWithoutCallback verifies that SetFilterValue works without callback
func TestSetFilterValueWithoutCallback(t *testing.T) {
	bf := NewBaseFilterable()
	// Don't set a callback
	bf.SetFilterValue("test value")

	if bf.GetFilterValue() != "test value" {
		t.Error("SetFilterValue should work without callback")
	}
}

// TestIsFilteringEnabled verifies that IsFilteringEnabled returns the correct state
func TestIsFilteringEnabled(t *testing.T) {
	bf := NewBaseFilterable()

	if bf.IsFilteringEnabled() {
		t.Error("IsFilteringEnabled should return false initially")
	}

	bf.EnableFiltering(true)
	if !bf.IsFilteringEnabled() {
		t.Error("IsFilteringEnabled should return true after EnableFiltering(true)")
	}

	bf.EnableFiltering(false)
	if bf.IsFilteringEnabled() {
		t.Error("IsFilteringEnabled should return false after EnableFiltering(false)")
	}
}

// TestClearFilter verifies that ClearFilter clears value and exits filter mode
func TestClearFilter(t *testing.T) {
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)
	bf.EnterFilterMode()
	bf.SetFilterValue("test value")

	bf.ClearFilter()

	if bf.GetFilterValue() != "" {
		t.Error("ClearFilter should clear filterValue")
	}
	if bf.IsFiltering() {
		t.Error("ClearFilter should exit filter mode")
	}
}

// TestMultipleEnableDisableCycles verifies correct behavior through multiple cycles
func TestMultipleEnableDisableCycles(t *testing.T) {
	bf := NewBaseFilterable()

	for i := 0; i < 3; i++ {
		bf.EnableFiltering(true)
		if !bf.IsFilteringEnabled() {
			t.Errorf("Cycle %d: EnableFiltering(true) failed", i)
		}

		bf.EnableFiltering(false)
		if bf.IsFilteringEnabled() {
			t.Errorf("Cycle %d: EnableFiltering(false) failed", i)
		}
	}
}

// TestMultipleFilterModeTransitions verifies correct behavior through multiple transitions
func TestMultipleFilterModeTransitions(t *testing.T) {
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)

	for i := 0; i < 3; i++ {
		bf.EnterFilterMode()
		if !bf.IsFiltering() {
			t.Errorf("Cycle %d: EnterFilterMode failed", i)
		}

		bf.ExitFilterMode()
		if bf.IsFiltering() {
			t.Errorf("Cycle %d: ExitFilterMode failed", i)
		}
	}
}

// TestCallbackMultipleCalls verifies callback is called multiple times
func TestCallbackMultipleCalls(t *testing.T) {
	bf := NewBaseFilterable()

	callCount := 0
	bf.SetOnFilterChange(func(value string) {
		callCount++
	})

	bf.SetFilterValue("first")
	bf.SetFilterValue("second")
	bf.SetFilterValue("third")

	if callCount != 3 {
		t.Errorf("Callback should be called 3 times, was called %d times", callCount)
	}
}

// TestEmptyStringFilter verifies that empty string is valid filter value
func TestEmptyStringFilter(t *testing.T) {
	bf := NewBaseFilterable()

	bf.SetFilterValue("something")
	bf.SetFilterValue("")

	if bf.GetFilterValue() != "" {
		t.Error("Empty string should be a valid filter value")
	}
}

// TestFilterValueAfterMultipleSets verifies last value is retained
func TestFilterValueAfterMultipleSets(t *testing.T) {
	bf := NewBaseFilterable()

	values := []string{"first", "second", "third"}
	for _, v := range values {
		bf.SetFilterValue(v)
	}

	if bf.GetFilterValue() != "third" {
		t.Error("GetFilterValue should return the last set value")
	}
}

// TestInterfaceCompliance verifies BaseFilterable implements FilterableComponent
func TestInterfaceCompliance(t *testing.T) {
	var _ FilterableComponent = (*BaseFilterable)(nil)
}

// TestFilteringEnabledPreservesValue verifies that enabling/disabling preserves filter value
func TestFilteringEnabledPreservesValue(t *testing.T) {
	bf := NewBaseFilterable()
	bf.EnableFiltering(true)
	testValue := "important filter"
	bf.SetFilterValue(testValue)

	bf.EnableFiltering(false)

	if bf.GetFilterValue() != testValue {
		t.Error("EnableFiltering should preserve filterValue")
	}
}

// TestComplexStateTransitions verifies complex state changes work correctly
func TestComplexStateTransitions(t *testing.T) {
	bf := NewBaseFilterable()

	// Enable filtering
	bf.EnableFiltering(true)
	if !bf.IsFilteringEnabled() {
		t.Error("Step 1: EnableFiltering failed")
	}

	// Enter filter mode
	bf.EnterFilterMode()
	if !bf.IsFiltering() {
		t.Error("Step 2: EnterFilterMode failed")
	}

	// Set filter value while in filter mode
	bf.SetFilterValue("test")
	if bf.GetFilterValue() != "test" {
		t.Error("Step 3: SetFilterValue failed while in filter mode")
	}

	// Exit filter mode
	bf.ExitFilterMode()
	if bf.IsFiltering() {
		t.Error("Step 4: ExitFilterMode failed")
	}
	if bf.GetFilterValue() != "test" {
		t.Error("Step 4: ExitFilterMode should preserve value")
	}

	// Disable filtering while not in filter mode
	bf.EnableFiltering(false)
	if bf.IsFilteringEnabled() {
		t.Error("Step 5: EnableFiltering(false) failed")
	}
	if bf.GetFilterValue() != "test" {
		t.Error("Step 5: FilterValue should be preserved")
	}
}

// TestCallbackValueMatching verifies callback receives exact value passed to SetFilterValue
func TestCallbackValueMatching(t *testing.T) {
	bf := NewBaseFilterable()

	testCases := []string{
		"simple",
		"with spaces",
		"with-dashes",
		"with_underscores",
		"123numbers",
		"!@#$%special",
		"",
	}

	for _, testCase := range testCases {
		receivedValue := ""
		bf.SetOnFilterChange(func(value string) {
			receivedValue = value
		})

		bf.SetFilterValue(testCase)

		if receivedValue != testCase {
			t.Errorf("Callback received %q but SetFilterValue was called with %q", receivedValue, testCase)
		}
	}
}
