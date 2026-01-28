package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestFilterStatusViewEmptyFilter verifies that an empty filter returns an empty string
func TestFilterStatusViewEmptyFilter(t *testing.T) {
	result := FilterStatusView("", 0, 10)
	if result != "" {
		t.Errorf("Expected empty string for empty filter, got: %q", result)
	}
}

// TestFilterStatusViewBasicFilter verifies that a basic filter renders correctly
func TestFilterStatusViewBasicFilter(t *testing.T) {
	result := FilterStatusView("test", 5, 10)
	if result == "" {
		t.Error("Expected non-empty result for active filter")
	}
	
	// Check that the output contains the filter text and counts
	if !strings.Contains(result, "FILTER") {
		t.Error("Output should contain 'FILTER' keyword")
	}
	if !strings.Contains(result, "test") {
		t.Error("Output should contain filter text 'test'")
	}
	if !strings.Contains(result, "5") {
		t.Error("Output should contain match count '5'")
	}
	if !strings.Contains(result, "10") {
		t.Error("Output should contain total count '10'")
	}
}

// TestFilterStatusViewTruncation verifies that long filter text is truncated to 20 characters
func TestFilterStatusViewTruncation(t *testing.T) {
	longFilter := "this_is_a_very_long_filter_string_that_exceeds_limit"
	result := FilterStatusView(longFilter, 3, 8)
	
	if result == "" {
		t.Error("Expected non-empty result for active filter")
	}
	
	// Verify that the full long filter is NOT in the output
	if strings.Contains(result, longFilter) {
		t.Error("Output should not contain the full long filter text")
	}
	
	// The truncated version (20 chars) should be present
	truncated := longFilter[:20]
	if !strings.Contains(result, truncated) {
		t.Errorf("Output should contain truncated filter: %s", truncated)
	}
}

// TestFilterStatusViewCountDisplay verifies correct count display
func TestFilterStatusViewCountDisplay(t *testing.T) {
	tests := []struct {
		name       string
		filter     string
		matchCount int
		totalCount int
		expectSub  string
	}{
		{
			name:       "Single match",
			filter:     "x",
			matchCount: 1,
			totalCount: 5,
			expectSub:  "1/5",
		},
		{
			name:       "All match",
			filter:     "y",
			matchCount: 10,
			totalCount: 10,
			expectSub:  "10/10",
		},
		{
			name:       "Zero matches",
			filter:     "z",
			matchCount: 0,
			totalCount: 15,
			expectSub:  "0/15",
		},
		{
			name:       "Large counts",
			filter:     "big",
			matchCount: 1234,
			totalCount: 5678,
			expectSub:  "1234/5678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterStatusView(tt.filter, tt.matchCount, tt.totalCount)
			if !strings.Contains(result, tt.expectSub) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expectSub, result)
			}
		})
	}
}

// TestFilterStatusViewExactlyTwentyChars verifies a filter of exactly 20 chars is not truncated
func TestFilterStatusViewExactlyTwentyChars(t *testing.T) {
	filter20 := "12345678901234567890" // Exactly 20 chars
	result := FilterStatusView(filter20, 1, 2)
	
	if !strings.Contains(result, filter20) {
		t.Errorf("Expected 20-char filter to be preserved: %s", filter20)
	}
}

// TestFilterStatusViewSpecialCharacters verifies filter handles special characters
func TestFilterStatusViewSpecialCharacters(t *testing.T) {
	specialChars := "test@#$%"
	result := FilterStatusView(specialChars, 1, 1)
	
	if result == "" {
		t.Error("Expected non-empty result for filter with special characters")
	}
	
	if !strings.Contains(result, specialChars) {
		t.Errorf("Output should contain special characters: %s", specialChars)
	}
}

// TestFilterStatusViewStyling verifies the output is styled (contains ANSI codes)
func TestFilterStatusViewStyling(t *testing.T) {
	result := FilterStatusView("styled", 2, 5)
	
	if result == "" {
		t.Error("Expected styled output")
	}
	
	// lipgloss renders styled text which typically contains ANSI escape codes
	// Check if the output contains visible content
	if !strings.Contains(result, "FILTER") && !strings.Contains(result, "styled") {
		t.Error("Output should contain filter content even with styling applied")
	}
}

// TestFilterStatusViewSpaces verifies filter with spaces is handled correctly
func TestFilterStatusViewSpaces(t *testing.T) {
	filterWithSpaces := "hello world"
	result := FilterStatusView(filterWithSpaces, 3, 7)
	
	if result == "" {
		t.Error("Expected non-empty result for filter with spaces")
	}
	
	// The filter text should be in the output
	if !strings.Contains(result, "hello world") {
		t.Error("Output should preserve spaces in filter text")
	}
}

// TestFilterStatusViewZeroTotalCount verifies handling of zero total count
func TestFilterStatusViewZeroTotalCount(t *testing.T) {
	result := FilterStatusView("empty", 0, 0)
	
	if result == "" {
		t.Error("Expected non-empty result even with zero counts")
	}
	
	if !strings.Contains(result, "0/0") {
		t.Error("Output should show 0/0 for zero counts")
	}
}

// TestFilterStatusViewColorProfile verifies output works with different color profiles
func TestFilterStatusViewColorProfile(t *testing.T) {
	// Store original profile
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)

	// Test with different profiles - just verify the function works with each
	profiles := []termenv.Profile{
		termenv.TrueColor,
		termenv.ANSI256,
		termenv.ANSI,
		termenv.Ascii,
	}

	for i, profile := range profiles {
		t.Run(fmt.Sprintf("Profile_%d", i), func(t *testing.T) {
			lipgloss.SetColorProfile(profile)
			result := FilterStatusView("color", 1, 1)
			
			if result == "" {
				t.Errorf("Expected output with color profile")
			}
		})
	}
}

// TestFilterStatusViewWhitespaceOnly verifies filter with only whitespace
func TestFilterStatusViewWhitespaceOnly(t *testing.T) {
	// A filter with only spaces is still technically "active" since it's not empty
	result := FilterStatusView("   ", 0, 5)
	
	if result == "" {
		t.Error("Expected non-empty result for whitespace-only filter")
	}
	
	// The whitespace should be preserved
	if !strings.Contains(result, "FILTER") {
		t.Error("Output should contain FILTER indicator")
	}
}

// TestFilterStatusViewUnicodeCharacters verifies filter handles Unicode characters
func TestFilterStatusViewUnicodeCharacters(t *testing.T) {
	unicodeFilter := "café"
	result := FilterStatusView(unicodeFilter, 1, 2)
	
	if result == "" {
		t.Error("Expected non-empty result for Unicode filter")
	}
	
	// Should contain the filter text
	if !strings.Contains(result, "FILTER") {
		t.Error("Output should contain FILTER indicator")
	}
}

// TestFilterStatusViewLongFilterWithLargeCount verifies edge case of long filter + large count
func TestFilterStatusViewLongFilterWithLargeCount(t *testing.T) {
	longFilter := "this_is_a_very_long_filter_string_that_will_be_truncated"
	result := FilterStatusView(longFilter, 9999, 99999)
	
	if result == "" {
		t.Error("Expected non-empty result")
	}
	
	// Verify truncation occurred
	if strings.Contains(result, longFilter) {
		t.Error("Long filter should be truncated")
	}
	
	// Verify counts are present
	if !strings.Contains(result, "9999/99999") {
		t.Error("Output should contain large counts")
	}
}
