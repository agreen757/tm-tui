package filechanges

import (
	"testing"
)

func TestNewCommitParser(t *testing.T) {
	parser := NewCommitParser()
	
	if parser == nil {
		t.Fatal("Expected parser to be non-nil")
	}
	
	if parser.taskIDPattern == nil {
		t.Fatal("Expected taskIDPattern to be non-nil")
	}
}

func TestNewCommitParserWithPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "valid pattern",
			pattern:     `#([0-9]+)`,
			expectError: false,
		},
		{
			name:        "invalid pattern",
			pattern:     `#([0-9+`,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewCommitParserWithPattern(tt.pattern)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				if parser != nil {
					t.Error("Expected parser to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if parser == nil {
					t.Error("Expected parser to be non-nil")
				}
			}
		})
	}
}

func TestExtractTaskIDs(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected []string
	}{
		{
			name:     "single task ID",
			message:  "Implement feature #1",
			expected: []string{"1"},
		},
		{
			name:     "multiple task IDs",
			message:  "Fix bugs #1 and #2",
			expected: []string{"1", "2"},
		},
		{
			name:     "nested subtask ID",
			message:  "Complete subtask #1.2.3",
			expected: []string{"1.2.3"},
		},
		{
			name:     "task ID at start",
			message:  "#5 - Add new feature",
			expected: []string{"5"},
		},
		{
			name:     "task ID at end",
			message:  "Add new feature #5",
			expected: []string{"5"},
		},
		{
			name:     "multiple nested task IDs",
			message:  "Implement #1.1, #1.2, and #2.3",
			expected: []string{"1.1", "1.2", "2.3"},
		},
		{
			name:     "no task IDs",
			message:  "This commit has no task references",
			expected: []string{},
		},
		{
			name:     "duplicate task IDs",
			message:  "Work on #1 and also #1",
			expected: []string{"1"},
		},
		{
			name:     "task IDs with surrounding text",
			message:  "This is task #3 which relates to #4 and #5.1",
			expected: []string{"3", "4", "5.1"},
		},
		{
			name:     "empty message",
			message:  "",
			expected: []string{},
		},
		{
			name:     "hash but no number",
			message:  "Use #hashtag in code",
			expected: []string{},
		},
		{
			name:     "number without hash",
			message:  "Fix issue 123",
			expected: []string{},
		},
		{
			name:     "complex subtask hierarchy",
			message:  "Complete #1.2.3.4 and #2.1.5",
			expected: []string{"1.2.3.4", "2.1.5"},
		},
		{
			name:     "task ID in brackets",
			message:  "Implement feature [#1]",
			expected: []string{"1"},
		},
		{
			name:     "task ID in parentheses",
			message:  "Fix bug (#2.1)",
			expected: []string{"2.1"},
		},
		{
			name:     "task ID with punctuation",
			message:  "Complete task #3, then #4.",
			expected: []string{"3", "4"},
		},
		{
			name:     "multiple duplicate IDs",
			message:  "Work on #1, #2, #1, #3, #2, #1",
			expected: []string{"1", "2", "3"},
		},
	}
	
	parser := NewCommitParser()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractTaskIDs(tt.message)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d task IDs, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			
			// Check that all expected IDs are present
			for i, expectedID := range tt.expected {
				if result[i] != expectedID {
					t.Errorf("Expected task ID %s at index %d, got %s", expectedID, i, result[i])
				}
			}
		})
	}
}

func TestExtractTaskIDsNilParser(t *testing.T) {
	var parser *CommitParser
	result := parser.ExtractTaskIDs("Test #1")
	
	if len(result) != 0 {
		t.Errorf("Expected empty slice for nil parser, got %v", result)
	}
}

func TestHasTaskIDs(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "has task ID",
			message:  "Implement feature #1",
			expected: true,
		},
		{
			name:     "no task ID",
			message:  "Regular commit message",
			expected: false,
		},
		{
			name:     "multiple task IDs",
			message:  "Fix #1 and #2",
			expected: true,
		},
		{
			name:     "empty message",
			message:  "",
			expected: false,
		},
	}
	
	parser := NewCommitParser()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.HasTaskIDs(tt.message)
			
			if result != tt.expected {
				t.Errorf("Expected HasTaskIDs to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetFirstTaskID(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "single task ID",
			message:  "Implement feature #1",
			expected: "1",
		},
		{
			name:     "multiple task IDs returns first",
			message:  "Fix #2 and #3",
			expected: "2",
		},
		{
			name:     "no task ID",
			message:  "Regular commit message",
			expected: "",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "",
		},
		{
			name:     "nested task ID",
			message:  "Complete #1.2.3",
			expected: "1.2.3",
		},
	}
	
	parser := NewCommitParser()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetFirstTaskID(tt.message)
			
			if result != tt.expected {
				t.Errorf("Expected first task ID to be %q, got %q", tt.expected, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkExtractTaskIDs(b *testing.B) {
	parser := NewCommitParser()
	message := "Complete tasks #1, #2.1, #3.2.1, and #4"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ExtractTaskIDs(message)
	}
}

func BenchmarkExtractTaskIDsNoMatch(b *testing.B) {
	parser := NewCommitParser()
	message := "This is a commit message with no task IDs"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ExtractTaskIDs(message)
	}
}

func BenchmarkHasTaskIDs(b *testing.B) {
	parser := NewCommitParser()
	message := "Complete task #1.2.3"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.HasTaskIDs(message)
	}
}
