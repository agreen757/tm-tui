package ui

import (
	"testing"
)

// TestStatusIndicator tests that status indicators provide text alternatives for accessibility
func TestStatusIndicator(t *testing.T) {
	tests := []struct {
		status       string
		expectedIcon string
		expectedText string
	}{
		{"pending", "○", "PENDING"},
		{"in-progress", "►", "IN-PROGRESS"},
		{"done", "✓", "DONE"},
		{"blocked", "!", "BLOCKED"},
		{"deferred", "⏱", "DEFERRED"},
		{"cancelled", "✗", "CANCELLED"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			icon := GetStatusIcon(tt.status)
			if icon != tt.expectedIcon {
				t.Errorf("GetStatusIcon(%q) = %q, want %q", tt.status, icon, tt.expectedIcon)
			}

			label := GetStatusLabel(tt.status)
			if label != tt.expectedText {
				t.Errorf("GetStatusLabel(%q) = %q, want %q", tt.status, label, tt.expectedText)
			}

			indicator := GetStatusIndicator(tt.status)
			expected := icon + " " + label
			if indicator != expected {
				t.Errorf("GetStatusIndicator(%q) = %q, want %q", tt.status, indicator, expected)
			}
		})
	}
}

// TestComplexityIndicator tests that complexity indicators provide text alternatives for accessibility
func TestComplexityIndicator(t *testing.T) {
	tests := []struct {
		complexity        int
		expectedLabel     string
		expectedIndicator string
	}{
		{1, "LOW", "LOW(1)"},
		{3, "LOW", "LOW(3)"},
		{4, "MEDIUM", "MEDIUM(4)"},
		{6, "MEDIUM", "MEDIUM(6)"},
		{7, "MEDIUM", "MEDIUM(7)"}, // Fixed: 7 is at Medium threshold (4-7)
		{8, "HIGH", "HIGH(8)"},     // Added: 8 starts High range (8-12)
		{10, "HIGH", "HIGH(10)"},
		{12, "HIGH", "HIGH(12)"},           // Added: 12 is at High threshold
		{13, "VERY HIGH", "VERY HIGH(13)"}, // Added: 13+ is Very High
		{15, "VERY HIGH", "VERY HIGH(15)"}, // Added: test well above threshold
		{0, "", ""},
		{-1, "", ""},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.complexity)), func(t *testing.T) {
			label := GetComplexityLabel(tt.complexity)
			if label != tt.expectedLabel {
				t.Errorf("GetComplexityLabel(%d) = %q, want %q", tt.complexity, label, tt.expectedLabel)
			}

			indicator := GetComplexityIndicator(tt.complexity)
			if indicator != tt.expectedIndicator {
				t.Errorf("GetComplexityIndicator(%d) = %q, want %q", tt.complexity, indicator, tt.expectedIndicator)
			}
		})
	}
}
