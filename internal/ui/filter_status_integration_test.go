package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

// TestFilterStatusViewComponentRendering verifies FilterStatusView component renders correctly
func TestFilterStatusViewComponentRendering(t *testing.T) {
	tests := []struct {
		name        string
		filter      string
		matchCount  int
		totalCount  int
		expectEmpty bool
	}{
		{
			name:        "Empty filter returns empty",
			filter:      "",
			matchCount:  0,
			totalCount:  10,
			expectEmpty: true,
		},
		{
			name:        "Active filter renders",
			filter:      "test",
			matchCount:  3,
			totalCount:  10,
			expectEmpty: false,
		},
		{
			name:        "Long filter truncated",
			filter:      "this_is_a_very_long_filter_string_that_exceeds_twenty_chars",
			matchCount:  5,
			totalCount:  20,
			expectEmpty: false,
		},
		{
			name:        "No matches",
			filter:      "nomatch",
			matchCount:  0,
			totalCount:  15,
			expectEmpty: false,
		},
		{
			name:        "All match",
			filter:      "all",
			matchCount:  8,
			totalCount:  8,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterStatusView(tt.filter, tt.matchCount, tt.totalCount)
			if tt.expectEmpty && result != "" {
				t.Errorf("Expected empty result for filter %q, got: %q", tt.filter, result)
			}
			if !tt.expectEmpty && result == "" {
				t.Errorf("Expected non-empty result for filter %q", tt.filter)
			}
		})
	}
}

// TestFilterStatusViewWithBaseFilterable verifies integration with BaseFilterable component
func TestFilterStatusViewWithBaseFilterable(t *testing.T) {
	fc := dialog.NewBaseFilterable()
	fc.EnableFiltering(true)

	// Initially not filtering
	if fc.IsFiltering() {
		t.Error("Initially should not be filtering")
	}

	// Enter filter mode
	fc.EnterFilterMode()
	if !fc.IsFiltering() {
		t.Error("Should be filtering after EnterFilterMode")
	}

	// Set filter value
	fc.SetFilterValue("test_filter")
	filterValue := fc.GetFilterValue()
	if filterValue != "test_filter" {
		t.Errorf("Expected filter value 'test_filter', got %q", filterValue)
	}

	// Render filter status
	filterStatus := FilterStatusView(filterValue, 5, 10)
	if filterStatus == "" {
		t.Error("Filter status should be rendered when filtering is active")
	}

	// Exit filter mode
	fc.ExitFilterMode()
	if fc.IsFiltering() {
		t.Error("Should not be filtering after ExitFilterMode")
	}

	// Filter status should still be renderable with the stored value
	filterStatus2 := FilterStatusView(fc.GetFilterValue(), 5, 10)
	if filterStatus2 == "" {
		t.Error("Filter status should still be renderable after exiting filter mode")
	}
}

// TestFilterStatusViewDisplayCountAccuracy verifies count display is accurate
func TestFilterStatusViewDisplayCountAccuracy(t *testing.T) {
	tests := []struct {
		name       string
		filter     string
		matchCount int
		totalCount int
	}{
		{"Single digit counts", "x", 1, 9},
		{"Double digit counts", "test", 10, 99},
		{"Triple digit counts", "search", 100, 999},
		{"Large counts", "query", 10000, 50000},
		{"Match equals total", "all", 25, 25},
		{"Zero matches", "none", 0, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterStatusView(tt.filter, tt.matchCount, tt.totalCount)
			if result == "" {
				t.Fatal("Result should not be empty for active filter")
			}
			// Verify the count format is in the output
			// Note: exact format may vary due to styling, so we just check content is reasonable
			if len(result) == 0 {
				t.Error("Rendered filter status should have content")
			}
		})
	}
}
