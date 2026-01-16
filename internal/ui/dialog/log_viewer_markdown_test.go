package dialog

import (
	"strings"
	"testing"
)

// TestToggleMarkdown tests the markdown toggle functionality
func TestToggleMarkdown(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	// Default should be disabled
	if lv.markdownEnabled {
		t.Error("Markdown should be disabled by default")
	}
	
	// Toggle on
	lv.ToggleMarkdown()
	if !lv.markdownEnabled {
		t.Error("Markdown should be enabled after toggle")
	}
	
	// Toggle back off
	lv.ToggleMarkdown()
	if lv.markdownEnabled {
		t.Error("Markdown should be disabled after second toggle")
	}
}

// TestRenderMarkdownHeaders tests header rendering
func TestRenderMarkdownHeaders(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	tests := []struct {
		name     string
		input    string
		contains []string // Strings that should be in output
	}{
		{
			name:     "H1 header",
			input:    "# Main Title",
			contains: []string{"Main Title"},
		},
		{
			name:     "H2 header",
			input:    "## Section Title",
			contains: []string{"Section Title"},
		},
		{
			name:     "H3 header",
			input:    "### Subsection",
			contains: []string{"Subsection"},
		},
		{
			name:     "H4 header",
			input:    "#### Sub-subsection",
			contains: []string{"Sub-subsection"},
		},
		{
			name:     "H5 header",
			input:    "##### Minor heading",
			contains: []string{"Minor heading"},
		},
		{
			name:     "H6 header",
			input:    "###### Smallest heading",
			contains: []string{"Smallest heading"},
		},
		{
			name:     "Mixed headers",
			input:    "# Title\n## Section\n### Subsection",
			contains: []string{"Title", "Section", "Subsection"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lv.renderMarkdown(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

// TestRenderMarkdownCodeBlocks tests code block rendering
func TestRenderMarkdownCodeBlocks(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "simple code block",
			input:    "```\ncode here\n```",
			contains: []string{"code here"},
		},
		{
			name:     "code block with language",
			input:    "```go\nfunc main() {}\n```",
			contains: []string{"func main"},
		},
		{
			name:     "multiple code blocks",
			input:    "```\nblock1\n```\ntext\n```\nblock2\n```",
			contains: []string{"block1", "block2", "text"},
		},
		{
			name:     "nested code in text",
			input:    "Some text\n```\ncode\n```\nMore text",
			contains: []string{"Some text", "code", "More text"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lv.renderMarkdown(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q", expected)
				}
			}
		})
	}
}

// TestRenderInlineFormatting tests inline markdown formatting
func TestRenderInlineFormatting(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "inline code",
			input:    "Use `func main()` to start",
			contains: []string{"func main()"},
		},
		{
			name:     "bold text",
			input:    "This is **important** text",
			contains: []string{"important"},
		},
		{
			name:     "italic text",
			input:    "This is *emphasized* text",
			contains: []string{"emphasized"},
		},
		{
			name:     "mixed formatting",
			input:    "Use **bold** and `code` together",
			contains: []string{"bold", "code"},
		},
		{
			name:     "multiple inline codes",
			input:    "Compare `foo` with `bar`",
			contains: []string{"foo", "bar"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lv.renderInlineFormatting(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

// TestRenderMarkdownLists tests list rendering
func TestRenderMarkdownLists(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "simple list with dash",
			input:    "- Item 1\n- Item 2\n- Item 3",
			contains: []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name:     "simple list with asterisk",
			input:    "* Item 1\n* Item 2\n* Item 3",
			contains: []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name:     "indented list",
			input:    "  - Indented item\n  - Another indented",
			contains: []string{"Indented item", "Another indented"},
		},
		{
			name:     "mixed markers",
			input:    "- Dash item\n* Asterisk item",
			contains: []string{"Dash item", "Asterisk item"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lv.renderMarkdown(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q", expected)
				}
			}
		})
	}
}

// TestRenderMarkdownPreservesStructure tests that structure is preserved
func TestRenderMarkdownPreservesStructure(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	input := `# Title
## Section 1
Some text here
- List item 1
- List item 2

## Section 2
More text with **bold** and ` + "`code`" + `

` + "```" + `
code block
` + "```"
	
	result := lv.renderMarkdown(input)
	
	// Check that all content is present
	expectedParts := []string{
		"Title",
		"Section 1",
		"Some text here",
		"List item 1",
		"List item 2",
		"Section 2",
		"bold",
		"code",
		"code block",
	}
	
	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain %q", part)
		}
	}
}

// TestRenderMarkdownEdgeCases tests edge cases
func TestRenderMarkdownEdgeCases(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	tests := []struct {
		name   string
		input  string
		verify func(string) bool
	}{
		{
			name:  "empty string",
			input: "",
			verify: func(s string) bool {
				return s == ""
			},
		},
		{
			name:  "no markdown",
			input: "Plain text without any markdown",
			verify: func(s string) bool {
				return strings.Contains(s, "Plain text")
			},
		},
		{
			name:  "unclosed code block",
			input: "```\ncode without closing",
			verify: func(s string) bool {
				return strings.Contains(s, "code without closing")
			},
		},
		{
			name:  "unclosed inline code",
			input: "Text with `unclosed code",
			verify: func(s string) bool {
				return strings.Contains(s, "unclosed code")
			},
		},
		{
			name:  "unclosed bold",
			input: "Text with **unclosed bold",
			verify: func(s string) bool {
				return strings.Contains(s, "unclosed bold")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lv.renderMarkdown(tt.input)
			
			if !tt.verify(result) {
				t.Errorf("Verification failed for input %q, got: %s", tt.input, result)
			}
		})
	}
}

// TestMarkdownWithWordWrap tests markdown rendering combined with word wrap
func TestMarkdownWithWordWrap(t *testing.T) {
	lv := NewLogViewerPanel(40, 24, nil)
	lv.markdownEnabled = true
	lv.wordWrap = true
	
	content := "# Very Long Title That Should Wrap When Rendered\n\nThis is a very long paragraph that should wrap properly even with markdown rendering enabled."
	lv.SetContent(content, "test.md")
	
	// Check that content was rendered
	if lv.renderedContent == "" {
		t.Error("Expected rendered content to be non-empty")
	}
	
	// Check that content contains expected parts
	if !strings.Contains(lv.renderedContent, "Title") {
		t.Error("Expected rendered content to contain 'Title'")
	}
}

// TestFooterShowsMarkdownToggle tests that footer displays markdown toggle status
func TestFooterShowsMarkdownToggle(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	
	// Markdown off
	footer := lv.renderFooter()
	if !strings.Contains(footer, "Markdown: OFF") {
		t.Error("Expected footer to show 'Markdown: OFF'")
	}
	
	// Markdown on
	lv.markdownEnabled = true
	footer = lv.renderFooter()
	if !strings.Contains(footer, "Markdown: ON") {
		t.Error("Expected footer to show 'Markdown: ON'")
	}
}
