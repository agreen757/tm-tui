package ui

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// Basic cases
		{
			name:     "simple lowercase title",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "simple uppercase title",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "title with numbers",
			input:    "Task 123 Title",
			expected: "task-123-title",
		},
		{
			name:     "title with special characters",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "title with punctuation",
			input:    "Create a PRD: Product Requirements",
			expected: "create-a-prd-product-requirements",
		},
		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "single word",
			input:    "Documentation",
			expected: "documentation",
		},
		{
			name:     "multiple spaces",
			input:    "Hello    World",
			expected: "hello-world",
		},
		{
			name:     "underscores",
			input:    "hello_world_test",
			expected: "hello-world-test",
		},
		{
			name:     "mixed underscores and spaces",
			input:    "hello_world test",
			expected: "hello-world-test",
		},
		// Special characters
		{
			name:     "parentheses",
			input:    "Task (Important) Title",
			expected: "task-important-title",
		},
		{
			name:     "quotes",
			input:    `Create "New" PRD`,
			expected: "create-new-prd",
		},
		{
			name:     "hyphens preserved",
			input:    "pre-existing-title",
			expected: "pre-existing-title",
		},
		{
			name:     "dots preserved",
			input:    "v1.0.0 Release Notes",
			expected: "v1.0.0-release-notes",
		},
		// Multiple hyphens
		{
			name:     "multiple hyphens collapse",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "hyphens at edges",
			input:    "---hello-world---",
			expected: "hello-world",
		},
		// Accented characters
		{
			name:     "accented e",
			input:    "Café Menu",
			expected: "cafe-menu",
		},
		{
			name:     "accented a",
			input:    "Année 2024",
			expected: "annee-2024",
		},
		{
			name:     "accented o",
			input:    "São Paulo",
			expected: "sao-paulo",
		},
		{
			name:     "accented u",
			input:    "Überschrift",
			expected: "uberschrift",
		},
		{
			name:     "accented c",
			input:    "Español",
			expected: "espanol",
		},
		// Real-world examples
		{
			name:     "real PRD title 1",
			input:    "User Authentication System",
			expected: "user-authentication-system",
		},
		{
			name:     "real PRD title 2",
			input:    "API v2.0: Breaking Changes & Migration Guide",
			expected: "api-v2.0-breaking-changes-migration-guide",
		},
		{
			name:     "real PRD title 3",
			input:    "Dashboard Redesign (Phase 3)",
			expected: "dashboard-redesign-phase-3",
		},
		{
			name:     "real PRD title 4",
			input:    "Fix: Performance Issues #123",
			expected: "fix-performance-issues-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkSlugify(b *testing.B) {
	inputs := []string{
		"Hello World",
		"Create a PRD: Product Requirements",
		"User Authentication System",
		"API v2.0: Breaking Changes & Migration Guide",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			Slugify(input)
		}
	}
}
