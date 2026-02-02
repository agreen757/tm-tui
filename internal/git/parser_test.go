package git

import (
	"reflect"
	"testing"
)

func TestExtractTaskIDs(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected []string
	}{
		{
			name:     "No task IDs",
			message:  "fix: regular commit message",
			expected: []string{},
		},
		{
			name:     "Single task ID",
			message:  "fix: resolve issue #1.2",
			expected: []string{"1.2"},
		},
		{
			name:     "Multiple task IDs",
			message:  "fix: resolve issues #1.2 and #1.3",
			expected: []string{"1.2", "1.3"},
		},
		{
			name:     "Task ID at start",
			message:  "#2.1 implement feature X",
			expected: []string{"2.1"},
		},
		{
			name:     "Task ID at end",
			message:  "implement feature X #2.1",
			expected: []string{"2.1"},
		},
		{
			name:     "Top-level task ID",
			message:  "fix: resolve issue #3",
			expected: []string{"3"},
		},
		{
			name:     "Deep nested task ID",
			message:  "fix: resolve issue #1.2.3.4",
			expected: []string{"1.2.3.4"},
		},
		{
			name:     "Mixed task ID formats",
			message:  "fix: resolve #1, #2.1, and #3.2.1",
			expected: []string{"1", "2.1", "3.2.1"},
		},
		{
			name:     "Duplicate task IDs",
			message:  "fix: work on #1.2 and #1.2 again",
			expected: []string{"1.2"}, // Should deduplicate
		},
		{
			name:     "Hash without task ID",
			message:  "fix: use hash #abc123",
			expected: []string{}, // Should not match non-numeric
		},
		{
			name:     "Multiple hashes, only some are task IDs",
			message:  "fix: #1.2 with commit #abc123",
			expected: []string{"1.2"},
		},
		{
			name:     "Task ID in brackets",
			message:  "fix: [#2.1] resolve issue",
			expected: []string{"2.1"},
		},
		{
			name:     "Task ID in parentheses",
			message:  "fix: (#2.1) resolve issue",
			expected: []string{"2.1"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTaskIDs(tt.message)
			
			// Handle empty slice comparison
			if len(tt.expected) == 0 && len(result) == 0 {
				return
			}
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractTaskIDsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected []string
	}{
		{
			name:     "Empty message",
			message:  "",
			expected: []string{},
		},
		{
			name:     "Only whitespace",
			message:  "   \t\n  ",
			expected: []string{},
		},
		{
			name:     "Only hash symbol",
			message:  "# ",
			expected: []string{},
		},
		{
			name:     "Hash with letters",
			message:  "#abc",
			expected: []string{},
		},
		{
			name:     "Very long task ID",
			message:  "#1.2.3.4.5.6.7.8.9.10",
			expected: []string{"1.2.3.4.5.6.7.8.9.10"},
		},
		{
			name:     "Leading zeros",
			message:  "#01.02",
			expected: []string{"01.02"}, // Regex will match, validation can happen elsewhere
		},
		{
			name:     "Decimal-like but valid",
			message:  "#1.0",
			expected: []string{"1.0"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTaskIDs(tt.message)
			
			// Handle empty slice comparison
			if len(tt.expected) == 0 && len(result) == 0 {
				return
			}
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func BenchmarkExtractTaskIDs(b *testing.B) {
	message := "fix: resolve issues #1.2, #1.3, and #2.1 with commits #abc123 and #def456"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractTaskIDs(message)
	}
}
