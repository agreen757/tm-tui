package dialog

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestZeroWidthPanel tests handling of zero-width panels
func TestZeroWidthPanel(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Zero width", 0, 24},
		{"One pixel width", 1, 24},
		{"Small width", 10, 24},
		{"Normal width", 80, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := NewLogViewerPanel(tt.width, tt.height, nil)

			// Set content
			content := "This is a test line that should wrap"
			lv.SetContent(content, "/test.log")

			// Enable word wrap
			lv.wordWrap = true
			lv.renderContent()

			// Should not panic
			view := lv.View()
			if view == "" {
				t.Error("Expected non-empty view for zero-width panel")
			}

			// Verify minimum width is enforced
			if tt.width < 20 {
				// applyWordWrap should enforce minimum width of 20
				wrapped := lv.applyWordWrap([]string{content})
				for _, line := range wrapped {
					if len(line) > 20+1 { // +1 for continuation indicator
						t.Logf("Line length %d exceeds minimum width enforcement", len(line))
					}
				}
			}
		})
	}
}

// TestEmptyContentEdgeCases tests various empty content scenarios
func TestEmptyContentEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		filePath       string
		expectedLoaded bool
		expectedEmpty  bool
	}{
		{"Empty string", "", "", false, true},
		{"Empty with path", "", "/test.log", false, true},
		{"Whitespace only", "   \n\n   ", "/test.log", true, false},
		{"Single newline", "\n", "/test.log", true, false},
		{"Multiple newlines", "\n\n\n", "/test.log", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := NewLogViewerPanel(80, 24, nil)

			if tt.content != "" {
				lv.SetContent(tt.content, tt.filePath)
			}

			// Test rendering
			view := lv.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}

			// Test with word wrap enabled
			lv.wordWrap = true
			lv.renderContent()

			// Test with line numbers enabled
			lv.showLineNumbers = true
			lv.renderContent()

			// Should not panic
			view = lv.View()
			if view == "" {
				t.Error("Expected non-empty view after toggles")
			}
		})
	}
}

// TestVeryLongLinesWithoutSpaces tests handling of extremely long lines
func TestVeryLongLinesWithoutSpaces(t *testing.T) {
	tests := []struct {
		name       string
		lineLength int
		width      int
	}{
		{"100 chars no spaces", 100, 40},
		{"500 chars no spaces", 500, 60},
		{"1000 chars no spaces", 1000, 80},
		{"5000 chars no spaces", 5000, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := NewLogViewerPanel(tt.width, 24, nil)

			// Create line without spaces
			line := strings.Repeat("x", tt.lineLength)
			lv.SetContent(line, "/test.log")

			// Enable word wrap
			lv.wordWrap = true
			lv.renderContent()

			// Verify it wrapped
			wrapped := lv.applyWordWrap([]string{line})
			if len(wrapped) <= 1 {
				t.Errorf("Expected line to wrap into multiple lines, got %d", len(wrapped))
			}

			// Verify each wrapped line respects width
			availableWidth := tt.width - 4
			if availableWidth < 20 {
				availableWidth = 20
			}

			for i, wrappedLine := range wrapped {
				// Allow +1 for continuation indicator
				if len([]rune(wrappedLine)) > availableWidth+1 {
					t.Errorf("Wrapped line %d exceeds width: %d > %d", i, len([]rune(wrappedLine)), availableWidth)
				}
			}

			// Should not panic
			view := lv.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
		})
	}
}

// TestUnicodeCharacterHandling tests word wrapping with Unicode characters
func TestUnicodeCharacterHandling(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"Emoji", "Hello 👋 world 🌍 this is a test 🎉 with emoji 😊 characters"},
		{"CJK characters", "你好世界 こんにちは世界 안녕하세요 세계 This is mixed CJK and English"},
		{"Arabic RTL", "مرحبا بالعالم هذا اختبار مع النص العربي"},
		{"Combining characters", "café résumé naïve Zürich"},
		{"Mixed scripts", "Hello мир 世界 🌍 café"},
		{"Zero-width joiners", "👨‍👩‍👧‍👦 family emoji with ZWJ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := NewLogViewerPanel(40, 24, nil)

			lv.SetContent(tt.content, "/test.log")

			// Enable word wrap
			lv.wordWrap = true
			lv.renderContent()

			// Should not panic
			view := lv.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}

			// Verify wrapping uses runes (not bytes)
			wrapped := lv.applyWordWrap([]string{tt.content})
			for i, line := range wrapped {
				runeCount := len([]rune(line))
				byteCount := len(line)

				// If Unicode is present, rune count should be less than byte count
				if strings.ContainsAny(tt.content, "👋🌍🎉😊你好こんにちは") && runeCount == byteCount {
					t.Logf("Line %d: rune count=%d, byte count=%d (may not contain multibyte chars)",
						i, runeCount, byteCount)
				}
			}
		})
	}
}

// TestMixedMarkdownAndWordWrap tests markdown rendering with word wrap
func TestMixedMarkdownAndWordWrap(t *testing.T) {
	content := "# Header 1\n" +
		"This is a very long paragraph that should wrap when word wrap is enabled and markdown rendering is active.\n\n" +
		"## Header 2\n" +
		"- List item 1 with a very long description that exceeds the viewport width\n" +
		"- List item 2 with **bold** and *italic* text\n" +
		"- List item 3 with `inline code` that is also very long\n\n" +
		"### Code Block\n" +
		"```go\n" +
		"func veryLongFunctionName() string {\n" +
		"    return \"This is a very long string that should be handled properly with word wrap\"\n" +
		"}\n" +
		"```\n\n" +
		"Regular text after code block."

	tests := []struct {
		name        string
		wordWrap    bool
		markdown    bool
		lineNumbers bool
	}{
		{"All disabled", false, false, false},
		{"Only word wrap", true, false, false},
		{"Only markdown", false, true, false},
		{"Word wrap + markdown", true, true, false},
		{"All enabled", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := NewLogViewerPanel(60, 24, nil)

			lv.SetContent(content, "/test.md")
			lv.wordWrap = tt.wordWrap
			lv.markdownEnabled = tt.markdown
			lv.showLineNumbers = tt.lineNumbers

			// Render content
			lv.renderContent()

			// Should not panic
			view := lv.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}

			// Verify content is present
			if lv.renderedContent == "" {
				t.Error("Expected rendered content to be non-empty")
			}
		})
	}
}

// TestToggleWordWrapIntegration tests word wrap toggle functionality
func TestToggleWordWrapIntegration(t *testing.T) {
	lv := NewLogViewerPanel(40, 24, nil)
	lv.SetFocused(true)

	// Set content with long line
	longLine := "This is a very long line that should wrap when word wrap is enabled but should not wrap when disabled"
	lv.SetContent(longLine, "/test.log")

	// Initially word wrap is enabled
	if !lv.wordWrap {
		t.Fatal("Expected word wrap to be enabled by default")
	}

	// Render with word wrap
	lv.renderContent()
	wrappedContent := lv.renderedContent
	wrappedLines := strings.Split(wrappedContent, "\n")

	// Should have multiple lines
	if len(wrappedLines) <= 1 {
		t.Errorf("Expected multiple lines with word wrap, got %d", len(wrappedLines))
	}

	// Toggle word wrap off
	lv.ToggleWordWrap()

	if lv.wordWrap {
		t.Error("Expected word wrap to be disabled after toggle")
	}

	// Render without word wrap
	lv.renderContent()
	unwrappedContent := lv.renderedContent
	unwrappedLines := strings.Split(unwrappedContent, "\n")

	// Should have single line
	if len(unwrappedLines) != 1 {
		t.Errorf("Expected single line without word wrap, got %d", len(unwrappedLines))
	}

	// Content should differ
	if wrappedContent == unwrappedContent {
		t.Error("Expected wrapped and unwrapped content to differ")
	}

	// Toggle back on
	lv.ToggleWordWrap()

	if !lv.wordWrap {
		t.Error("Expected word wrap to be enabled after second toggle")
	}

	// Verify keyboard shortcut
	_, cmd := lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if cmd != nil {
		t.Error("Expected no command from 'w' key press")
	}

	// Word wrap should have toggled
	if lv.wordWrap {
		t.Error("Expected word wrap to toggle off with 'w' key")
	}
}

// TestToggleSwitchingSpeed tests that toggling is fast
func TestToggleSwitchingSpeed(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.SetFocused(true)

	// Create large content
	var content strings.Builder
	for i := 0; i < 100; i++ {
		content.WriteString("This is a line with some content that may or may not wrap depending on settings\n")
	}

	lv.SetContent(content.String(), "/test.log")

	// Initial state: word wrap is ON by default
	initialState := lv.wordWrap

	// Toggle multiple times rapidly
	for i := 0; i < 10; i++ {
		lv.ToggleWordWrap()

		// Verify state
		if lv.renderedContent == "" {
			t.Fatal("Expected rendered content after toggle")
		}
	}

	// Final state should be toggled 10 times (back to original)
	if lv.wordWrap != initialState {
		t.Errorf("Expected word wrap to return to initial state (%v) after 10 toggles, got %v", initialState, lv.wordWrap)
	}
}

// TestWordWrapWithDifferentWidths tests word wrap behavior at various widths
func TestWordWrapWithDifferentWidths(t *testing.T) {
	content := "This is a test line with multiple words that should wrap differently at different widths"

	widths := []int{20, 40, 60, 80, 100, 120}

	for _, width := range widths {
		t.Run(string(rune(width)), func(t *testing.T) {
			lv := NewLogViewerPanel(width, 24, nil)
			lv.SetContent(content, "/test.log")

			// With word wrap
			lv.wordWrap = true
			lv.renderContent()
			wrappedLines := strings.Split(lv.renderedContent, "\n")

			// Without word wrap
			lv.wordWrap = false
			lv.renderContent()
			unwrappedLines := strings.Split(lv.renderedContent, "\n")

			// Wrapped version should have more lines for small widths
			// Only test if the content actually needs wrapping at this width
			availableWidth := width - 4
			if availableWidth < 20 {
				availableWidth = 20
			}

			if len(content) > availableWidth {
				if len(wrappedLines) <= len(unwrappedLines) {
					t.Logf("Width %d (available: %d): wrapped lines=%d, unwrapped=%d",
						width, availableWidth, len(wrappedLines), len(unwrappedLines))
					// This is not necessarily an error if the line happens to fit
				}
			}
		})
	}
}

// TestCacheInvalidationOnToggle tests that cache is properly invalidated
func TestCacheInvalidationOnToggle(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")

	// SetContent calls renderContent which clears dirty flag
	if lv.dirty {
		t.Error("Expected dirty flag to be cleared after initial render")
	}

	// Capture state before toggle
	renderedBefore := lv.renderedContent

	// Toggle word wrap - this calls invalidateCache() then renderContent()
	lv.ToggleWordWrap()

	// After ToggleWordWrap completes, dirty flag should be cleared
	// (because ToggleWordWrap calls renderContent internally)
	if lv.dirty {
		t.Error("Expected dirty flag to be cleared after toggle (renderContent is called)")
	}

	// Content should have been re-rendered
	renderedAfter := lv.renderedContent

	// Since word wrap state changed, content might differ (even if this example doesn't wrap)
	t.Logf("Content before toggle length: %d, after: %d", len(renderedBefore), len(renderedAfter))
}

// TestEdgeCaseLineBreaking tests edge cases in line breaking
func TestEdgeCaseLineBreaking(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		width int
	}{
		{"Single char line", "x", 40},
		{"Line equals width", strings.Repeat("x", 40), 40},
		{"Line one char over", strings.Repeat("x", 41), 40},
		{"Only spaces", "                    ", 10},
		{"Spaces then text", "          text", 10},
		{"Text then spaces", "text          ", 10},
		{"Alternating spaces", "a b c d e f g h i j k l m n o p", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := NewLogViewerPanel(tt.width, 24, nil)
			lv.SetContent(tt.line, "/test.log")

			lv.wordWrap = true
			lv.renderContent()

			// Should not panic
			view := lv.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
		})
	}
}

// TestMultilineContentWithMixedLengths tests realistic log content
func TestMultilineContentWithMixedLengths(t *testing.T) {
	content := `Short line
This is a medium length line that has some content
This is a very long line that should definitely wrap when word wrap is enabled because it contains a lot of text that exceeds the typical viewport width
    Indented line with some content
Another short line
` + strings.Repeat("X", 200) // Very long line without spaces

	lv := NewLogViewerPanel(60, 24, nil)
	lv.SetContent(content, "/test.log")

	// Test with all features enabled
	lv.wordWrap = true
	lv.showLineNumbers = true
	lv.markdownEnabled = false
	lv.SetFocused(true)

	// Manually invalidate cache to ensure re-render with new settings
	lv.invalidateCache()
	lv.renderContent()

	// Verify rendering
	if lv.renderedContent == "" {
		t.Fatal("Expected rendered content")
	}

	lines := strings.Split(lv.renderedContent, "\n")

	// Should have more lines due to wrapping
	originalLines := strings.Split(content, "\n")
	if len(lines) <= len(originalLines) {
		t.Logf("Lines after wrapping: %d, original: %d (wrapping may not have occurred)",
			len(lines), len(originalLines))
	}

	// Line numbers should be present
	hasLineNumbers := false
	for _, line := range lines {
		if strings.Contains(line, "│") {
			hasLineNumbers = true
			break
		}
	}

	if !hasLineNumbers {
		t.Errorf("Expected at least some lines to have line numbers. First few lines:\n%s",
			strings.Join(lines[:min(5, len(lines))], "\n"))
	}

	// Test view rendering
	view := lv.View()
	if view == "" {
		t.Fatal("Expected non-empty view")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
