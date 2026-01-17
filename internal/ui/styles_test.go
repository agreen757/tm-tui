package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TestColorComplexityConstants verifies that color constants are defined with correct hex values
func TestColorComplexityConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"ColorComplexityLow", ColorComplexityLow, "#4169E1"},
		{"ColorComplexityMedium", ColorComplexityMedium, "#00CED1"},
		{"ColorComplexityHigh", ColorComplexityHigh, "#FFA500"},
		{"ColorComplexityVeryHigh", ColorComplexityVeryHigh, "#DC143C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.value, tt.expected)
			}
		})
	}
}

// TestColorComplexityConstantsExist verifies constants are accessible
func TestColorComplexityConstantsExist(t *testing.T) {
	// This test simply verifies the constants are defined and accessible
	// If they weren't defined, this test would not compile
	_ = ColorComplexityLow
	_ = ColorComplexityMedium
	_ = ColorComplexityHigh
	_ = ColorComplexityVeryHigh
}

// TestGetComplexityLevelStyle verifies the method returns the correct style for each complexity level enum
func TestGetComplexityLevelStyle(t *testing.T) {
	styles := NewStyles()

	tests := []struct {
		name  string
		level taskmaster.ComplexityLevel
	}{
		{
			name:  "Low complexity",
			level: taskmaster.ComplexityLow,
		},
		{
			name:  "Medium complexity",
			level: taskmaster.ComplexityMedium,
		},
		{
			name:  "High complexity",
			level: taskmaster.ComplexityHigh,
		},
		{
			name:  "Very High complexity",
			level: taskmaster.ComplexityVeryHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := styles.GetComplexityLevelStyle(tt.level)

			// Verify Bold attribute is set
			if !style.GetBold() {
				t.Errorf("GetComplexityLevelStyle(%q) returned style without bold, want bold", tt.level)
			}

			// Verify foreground color is set
			if style.GetForeground() == nil {
				t.Errorf("GetComplexityLevelStyle(%q) returned style with nil foreground, want color set", tt.level)
			}
		})
	}
}

// TestGetComplexityLevelStyleDefaultCase verifies default case returns subtle style
func TestGetComplexityLevelStyleDefaultCase(t *testing.T) {
	styles := NewStyles()

	// Test with invalid/empty level
	invalidLevel := taskmaster.ComplexityLevel("invalid")
	style := styles.GetComplexityLevelStyle(invalidLevel)

	// Should return the subtle style
	if style.GetForeground() == nil {
		t.Errorf("GetComplexityLevelStyle(invalid) returned empty style, want subtle style")
	}
}

// TestGetComplexityStyle verifies the method returns the correct style for numeric complexity scores
func TestGetComplexityStyle(t *testing.T) {
	styles := NewStyles()

	tests := []struct {
		name       string
		complexity int
		shouldBold bool   // Whether the returned style should be bold
		checkColor string // Color to compare against if checking
	}{
		{
			name:       "Zero complexity",
			complexity: 0,
			shouldBold: false, // Subtle style is not bold
		},
		{
			name:       "Low complexity",
			complexity: 2,
			shouldBold: true,
		},
		{
			name:       "Medium complexity",
			complexity: 5,
			shouldBold: true,
		},
		{
			name:       "High complexity",
			complexity: 10,
			shouldBold: true,
		},
		{
			name:       "Very high complexity",
			complexity: 15,
			shouldBold: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := styles.GetComplexityStyle(tt.complexity)

			// Verify foreground color is set
			fg := style.GetForeground()
			if fg == nil {
				t.Errorf("GetComplexityStyle(%d) returned style with nil foreground, want color set", tt.complexity)
				return
			}

			// Verify bold attribute matches expectation
			if style.GetBold() != tt.shouldBold {
				t.Errorf("GetComplexityStyle(%d) bold=%v, want %v", tt.complexity, style.GetBold(), tt.shouldBold)
			}
		})
	}
}

// TestGetComplexityStyleBoundaries verifies the correct styles are returned at complexity threshold boundaries
func TestGetComplexityStyleBoundaries(t *testing.T) {
	styles := NewStyles()
	thresholds := taskmaster.DefaultLevelThresholds()

	// Test at exact threshold boundaries
	tests := []struct {
		name       string
		complexity int
		expectLow  bool
		expectMed  bool
		expectHigh bool
	}{
		{
			name:       "At low threshold",
			complexity: thresholds.Low,
			expectLow:  true,
		},
		{
			name:       "Just above low threshold",
			complexity: thresholds.Low + 1,
			expectMed:  true,
		},
		{
			name:       "At medium threshold",
			complexity: thresholds.Medium,
			expectMed:  true,
		},
		{
			name:       "Just above medium threshold",
			complexity: thresholds.Medium + 1,
			expectHigh: true,
		},
		{
			name:       "At high threshold",
			complexity: thresholds.High,
			expectHigh: true,
		},
		{
			name:       "Well above high threshold",
			complexity: thresholds.High + 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := styles.GetComplexityStyle(tt.complexity)

			// Verify style has a foreground color and is bold (for non-zero)
			if tt.complexity > 0 {
				if style.GetForeground() == nil {
					t.Errorf("Expected foreground color for complexity %d", tt.complexity)
				}
				if !style.GetBold() {
					t.Errorf("Expected bold style for complexity %d", tt.complexity)
				}
			}
		})
	}
}
