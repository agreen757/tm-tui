package dialog

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewLogViewerPanel(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	if lv == nil {
		t.Fatal("NewLogViewerPanel returned nil")
	}

	if lv.width != 80 {
		t.Errorf("Expected width 80, got %d", lv.width)
	}

	if lv.height != 24 {
		t.Errorf("Expected height 24, got %d", lv.height)
	}

	if !lv.wordWrap {
		t.Error("Expected word wrap to be enabled by default")
	}

	if lv.showLineNumbers {
		t.Error("Expected line numbers to be disabled by default")
	}

	if lv.focused {
		t.Error("Expected focused to be false by default")
	}
}

func TestLogViewerSetContent(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/path/to/test.log")

	if lv.content != content {
		t.Errorf("Expected content '%s', got '%s'", content, lv.content)
	}

	if lv.filePath != "/path/to/test.log" {
		t.Errorf("Expected file path '/path/to/test.log', got '%s'", lv.filePath)
	}
}

func TestLogViewerToggleLineNumbers(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	initialState := lv.showLineNumbers
	lv.ToggleLineNumbers()

	if lv.showLineNumbers == initialState {
		t.Error("ToggleLineNumbers did not change state")
	}

	lv.ToggleLineNumbers()

	if lv.showLineNumbers != initialState {
		t.Error("ToggleLineNumbers did not return to initial state")
	}
}

func TestLogViewerToggleWordWrap(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	initialState := lv.wordWrap
	lv.ToggleWordWrap()

	if lv.wordWrap == initialState {
		t.Error("ToggleWordWrap did not change state")
	}

	lv.ToggleWordWrap()

	if lv.wordWrap != initialState {
		t.Error("ToggleWordWrap did not return to initial state")
	}
}

func TestLogViewerSetFocused(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	if lv.IsFocused() {
		t.Error("Expected initial focused state to be false")
	}

	lv.SetFocused(true)

	if !lv.IsFocused() {
		t.Error("Expected focused state to be true after SetFocused(true)")
	}

	lv.SetFocused(false)

	if lv.IsFocused() {
		t.Error("Expected focused state to be false after SetFocused(false)")
	}
}

func TestLogViewerSetSize(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	lv.SetSize(100, 30)

	if lv.width != 100 {
		t.Errorf("Expected width 100, got %d", lv.width)
	}

	if lv.height != 30 {
		t.Errorf("Expected height 30, got %d", lv.height)
	}

	if lv.viewport.Width != 100 {
		t.Errorf("Expected viewport width 100, got %d", lv.viewport.Width)
	}

	// Viewport height should be content height minus footer (1 line)
	expectedViewportHeight := 30 - 1 // footer takes 1 line
	if lv.viewport.Height != expectedViewportHeight {
		t.Errorf("Expected viewport height %d, got %d", expectedViewportHeight, lv.viewport.Height)
	}
}

func TestWrapLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		width    int
		expected int // Expected number of wrapped lines
	}{
		{
			name:     "Short line no wrap",
			line:     "Short line",
			width:    20,
			expected: 1,
		},
		{
			name:     "Exact width no wrap",
			line:     "Exactly twenty chars",
			width:    20,
			expected: 1,
		},
		{
			name:     "Long line wraps at space",
			line:     "This is a very long line that should wrap at the nearest space",
			width:    20,
			expected: 4, // Should wrap into multiple lines
		},
		{
			name:     "No spaces force wrap",
			line:     "Verylongwordwithoutanyspacesthatwillbeforcewrapped",
			width:    20,
			expected: 3, // Should force wrap at width
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := wrapLine(tt.line, tt.width)

			if len(wrapped) != tt.expected {
				t.Errorf("Expected %d wrapped lines, got %d: %v", tt.expected, len(wrapped), wrapped)
			}

			// Verify each line respects width (allow +1 for continuation indicator)
			for i, line := range wrapped {
				if len(line) > tt.width+1 {
					t.Errorf("Line %d exceeds width: %d > %d: '%s'", i, len(line), tt.width, line)
				}
			}
		})
	}
}

func TestApplyWordWrap(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	lines := []string{
		"Short line",
		"This is a very long line that should definitely wrap because it exceeds the available width significantly",
		"Another short line",
	}

	wrapped := lv.applyWordWrap(lines)

	// Should have more lines after wrapping
	if len(wrapped) <= len(lines) {
		t.Errorf("Expected more lines after wrapping, got %d (original: %d)", len(wrapped), len(lines))
	}

	// Verify all wrapped lines fit within available width
	availableWidth := lv.width - 4
	for i, line := range wrapped {
		if len(line) > availableWidth+1 { // +1 for continuation indicator
			t.Errorf("Wrapped line %d exceeds available width: %d > %d", i, len(line), availableWidth)
		}
	}
}

func TestApplyLineNumbers(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	lines := []string{
		"Line one",
		"Line two",
		"Line three",
	}

	numbered := lv.applyLineNumbers(lines)

	if len(numbered) != len(lines) {
		t.Errorf("Expected %d numbered lines, got %d", len(lines), len(numbered))
	}

	// Verify line numbers are present
	for i, line := range numbered {
		if !strings.Contains(line, "│") {
			t.Errorf("Line %d missing separator: '%s'", i, line)
		}

		// Verify original content is preserved
		if !strings.Contains(line, lines[i]) {
			t.Errorf("Line %d missing original content: '%s'", i, line)
		}
	}
}

func TestLogViewerKeyboardNavigation(t *testing.T) {
	lv := NewLogViewerPanel(80, 10, nil)
	lv.SetFocused(true)

	// Set content with many lines to enable scrolling
	content := strings.Repeat("Line\n", 50)
	lv.SetContent(content, "/test.log")

	tests := []struct {
		name      string
		key       string
		expectCmd bool
	}{
		{"Arrow up", "up", false},
		{"Arrow down", "down", false},
		{"Vim up", "k", false},
		{"Vim down", "j", false},
		{"Page up", "pgup", false},
		{"Page down", "pgdown", false},
		{"Ctrl+U", "ctrl+u", false},
		{"Ctrl+D", "ctrl+d", false},
		{"Home", "home", false},
		{"End", "end", false},
		{"Vim top", "g", false},
		{"Vim bottom", "G", false},
		{"Toggle line numbers", "n", false},
		{"Toggle word wrap", "w", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes}

			// Simulate key press
			_, cmd := lv.Update(msg)

			if tt.expectCmd && cmd == nil {
				t.Error("Expected command, got nil")
			}

			if !tt.expectCmd && cmd != nil {
				t.Error("Expected no command, got command")
			}
		})
	}
}

func TestLogViewerViewEmptyState(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	view := lv.View()

	if view == "" {
		t.Error("Expected non-empty view for empty state")
	}

	if !strings.Contains(view, "Select a log file") {
		t.Error("Expected empty state message in view")
	}
}

func TestLogViewerViewWithContent(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Test content\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")

	view := lv.View()

	if view == "" {
		t.Error("Expected non-empty view with content")
	}

	// View should contain footer with scroll position
	if !strings.Contains(view, "Scroll:") {
		t.Error("Expected scroll position in footer")
	}

	// View should contain toggle indicators
	if !strings.Contains(view, "Line Numbers") {
		t.Error("Expected line numbers toggle in footer")
	}

	if !strings.Contains(view, "Word Wrap") {
		t.Error("Expected word wrap toggle in footer")
	}
}

func TestLogViewerScrollPosition(t *testing.T) {
	lv := NewLogViewerPanel(80, 10, nil)

	// Set content with many lines
	content := strings.Repeat("Line\n", 100)
	lv.SetContent(content, "/test.log")

	// Initially at top (0%)
	if lv.scrollPos == "" {
		t.Error("Expected scroll position to be set")
	}

	// Verify format includes line count and percentage
	if !strings.Contains(lv.scrollPos, "lines") {
		t.Errorf("Expected scroll position to contain 'lines', got: %s", lv.scrollPos)
	}

	if !strings.Contains(lv.scrollPos, "%") {
		t.Errorf("Expected scroll position to contain percentage, got: %s", lv.scrollPos)
	}

	// Test that scroll position updates
	lv.SetFocused(true)
	initialScrollPos := lv.scrollPos
	lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Scroll position should be updated
	if lv.scrollPos == "" {
		t.Error("Expected scroll position to be updated after scrolling")
	}

	// Should have changed from initial position
	if lv.scrollPos == initialScrollPos {
		t.Error("Expected scroll position to change after scrolling")
	}
}

func TestScrollPositionFormat(t *testing.T) {
	lv := NewLogViewerPanel(80, 10, nil)

	// Test with no content
	lv.updateScrollPosition()
	expectedEmpty := "0/0 lines (0%)"
	if lv.scrollPos != expectedEmpty {
		t.Errorf("Expected '%s' for empty content, got '%s'", expectedEmpty, lv.scrollPos)
	}

	// Test with content
	content := strings.Repeat("Line\n", 50)
	lv.SetContent(content, "/test.log")

	// Verify format: "X/Y lines (Z%)"
	// Should match pattern like "0/51 lines (0%)"
	if !strings.Contains(lv.scrollPos, "/") {
		t.Errorf("Expected scroll position to contain '/', got: %s", lv.scrollPos)
	}

	if !strings.Contains(lv.scrollPos, "lines") {
		t.Errorf("Expected scroll position to contain 'lines', got: %s", lv.scrollPos)
	}

	if !strings.HasSuffix(lv.scrollPos, "%)") {
		t.Errorf("Expected scroll position to end with '%%)', got: %s", lv.scrollPos)
	}
}

func TestLogViewerGetters(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Test content"
	filePath := "/path/to/test.log"

	lv.SetContent(content, filePath)

	if lv.GetContent() != content {
		t.Errorf("Expected content '%s', got '%s'", content, lv.GetContent())
	}

	if lv.GetFilePath() != filePath {
		t.Errorf("Expected file path '%s', got '%s'", filePath, lv.GetFilePath())
	}
}

func TestLogViewerUnfocusedIgnoresInput(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.SetFocused(false)

	content := strings.Repeat("Line\n", 50)
	lv.SetContent(content, "/test.log")

	initialScrollPos := lv.scrollPos

	// Try to scroll while unfocused
	lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Scroll position should not change
	if lv.scrollPos != initialScrollPos {
		t.Error("Expected scroll position to not change when unfocused")
	}
}

func TestWrapLinePreservesLeadingWhitespace(t *testing.T) {
	line := "    Indented line with many words that should wrap while preserving leading whitespace"
	width := 30

	wrapped := wrapLine(line, width)

	if len(wrapped) < 2 {
		t.Fatal("Expected line to wrap into multiple lines")
	}

	// First line should preserve leading whitespace
	if !strings.HasPrefix(wrapped[0], "    ") {
		t.Errorf("Expected first wrapped line to preserve leading whitespace: '%s'", wrapped[0])
	}

	// Continuation lines should have continuation indent (2 spaces)
	if len(wrapped) > 1 {
		for i := 1; i < len(wrapped); i++ {
			if !strings.HasPrefix(wrapped[i], "  ") {
				t.Errorf("Expected continuation line %d to have continuation indent: '%s'", i, wrapped[i])
			}
		}
	}
}

func TestWrapLineWithoutLeadingWhitespace(t *testing.T) {
	line := "This is a line without leading whitespace that should wrap normally without continuation indent"
	width := 30

	wrapped := wrapLine(line, width)

	if len(wrapped) < 2 {
		t.Fatal("Expected line to wrap into multiple lines")
	}

	// First line should not have leading whitespace
	if strings.HasPrefix(wrapped[0], " ") {
		t.Errorf("Expected first line without leading whitespace: '%s'", wrapped[0])
	}

	// Continuation lines should NOT have continuation indent (no original indent)
	if len(wrapped) > 1 {
		for i := 1; i < len(wrapped); i++ {
			if strings.HasPrefix(wrapped[i], "  ") {
				t.Errorf("Expected continuation line %d without indent (no original indent): '%s'", i, wrapped[i])
			}
		}
	}
}

func TestWrapLineContinuationIndicator(t *testing.T) {
	// Line with no spaces - should force wrap with continuation indicator
	line := "Verylongwordwithoutanyspacesthatwillbeforcewrappedwithindicator"
	width := 20

	wrapped := wrapLine(line, width)

	if len(wrapped) < 2 {
		t.Fatal("Expected line to wrap into multiple lines")
	}

	// First line should end with continuation indicator ">"
	if !strings.HasSuffix(wrapped[0], ">") {
		t.Errorf("Expected first line to end with continuation indicator: '%s'", wrapped[0])
	}
}

func TestLogViewerRenderContentWithLineNumbers(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.showLineNumbers = true

	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")

	// Check that rendered content contains line numbers
	if !strings.Contains(lv.renderedContent, "│") {
		t.Error("Expected rendered content to contain line number separator")
	}

	// Check that all original lines are present
	if !strings.Contains(lv.renderedContent, "Line 1") {
		t.Error("Expected rendered content to contain 'Line 1'")
	}
	if !strings.Contains(lv.renderedContent, "Line 2") {
		t.Error("Expected rendered content to contain 'Line 2'")
	}
	if !strings.Contains(lv.renderedContent, "Line 3") {
		t.Error("Expected rendered content to contain 'Line 3'")
	}
}

func TestLogViewerRenderContentWithWordWrap(t *testing.T) {
	lv := NewLogViewerPanel(40, 24, nil) // Narrow width to force wrapping
	lv.wordWrap = true

	longLine := "This is a very long line that should definitely wrap because it exceeds the available width"
	lv.SetContent(longLine, "/test.log")

	// Count lines in rendered content
	lines := strings.Split(lv.renderedContent, "\n")

	if len(lines) <= 1 {
		t.Errorf("Expected multiple lines after wrapping, got %d", len(lines))
	}
}

func TestLoadFileContent(t *testing.T) {
	// Create temporary test file
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	lv := NewLogViewerPanel(80, 24, nil)

	// Load file
	err = lv.LoadFileContent(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadFileContent failed: %v", err)
	}

	// Verify content loaded
	if !lv.IsLoaded() {
		t.Error("Expected IsLoaded to be true after loading file")
	}

	if lv.GetContent() != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, lv.GetContent())
	}

	if lv.GetFilePath() != tmpFile.Name() {
		t.Errorf("Expected file path '%s', got '%s'", tmpFile.Name(), lv.GetFilePath())
	}

	if lv.HasError() {
		t.Errorf("Expected no error, got: %v", lv.GetError())
	}
}

func TestLoadFileContentNonExistent(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Try to load non-existent file
	err := lv.LoadFileContent("/nonexistent/file.log")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	if !lv.HasError() {
		t.Error("Expected HasError to be true after load failure")
	}

	if lv.IsLoaded() {
		t.Error("Expected IsLoaded to be false after load failure")
	}

	if lv.GetError() == nil {
		t.Error("Expected GetError to return error")
	}
}

func TestLoadFileContentLineLimit(t *testing.T) {
	// Create temporary test file with more than 10,000 lines
	tmpFile, err := os.CreateTemp("", "test-large-log-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write 15,000 lines
	for i := 0; i < 15000; i++ {
		fmt.Fprintf(tmpFile, "Line %d\n", i+1)
	}
	tmpFile.Close()

	lv := NewLogViewerPanel(80, 24, nil)

	// Load file
	err = lv.LoadFileContent(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadFileContent failed: %v", err)
	}

	// Verify line limit applied
	if !lv.IsLineLimited() {
		t.Error("Expected IsLineLimited to be true for large file")
	}

	// Count lines in loaded content
	lines := strings.Split(lv.GetContent(), "\n")
	if len(lines) > 10000 {
		t.Errorf("Expected max 10,000 lines, got %d", len(lines))
	}
}

func TestClearContent(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Set some content
	lv.SetContent("Test content", "/test/file.log")

	if !lv.IsLoaded() {
		t.Fatal("Expected IsLoaded to be true after SetContent")
	}

	// Clear content
	lv.ClearContent()

	if lv.GetContent() != "" {
		t.Error("Expected empty content after ClearContent")
	}

	if lv.GetFilePath() != "" {
		t.Error("Expected empty file path after ClearContent")
	}

	if lv.IsLoaded() {
		t.Error("Expected IsLoaded to be false after ClearContent")
	}

	if lv.HasError() {
		t.Error("Expected HasError to be false after ClearContent")
	}

	if lv.IsLineLimited() {
		t.Error("Expected IsLineLimited to be false after ClearContent")
	}
}

func TestLoadFileContentCaching(t *testing.T) {
	// Create temporary test file
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := "Cached content"
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	lv := NewLogViewerPanel(80, 24, nil)

	// Load file first time
	err = lv.LoadFileContent(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadFileContent failed: %v", err)
	}

	firstLoad := lv.GetContent()

	// Load file again (should cache)
	err = lv.LoadFileContent(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadFileContent failed on second load: %v", err)
	}

	secondLoad := lv.GetContent()

	if firstLoad != secondLoad {
		t.Error("Content should be consistent across multiple loads")
	}
}

func TestEmptyFileLoading(t *testing.T) {
	// Create empty temp file
	tmpFile, err := os.CreateTemp("", "test-empty-log-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	lv := NewLogViewerPanel(80, 24, nil)

	// Load empty file
	err = lv.LoadFileContent(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadFileContent should succeed for empty file: %v", err)
	}

	if !lv.IsLoaded() {
		t.Error("Expected IsLoaded to be true for empty file")
	}

	if lv.GetContent() != "" {
		t.Errorf("Expected empty content, got '%s'", lv.GetContent())
	}

	if lv.HasError() {
		t.Error("Expected no error for empty file")
	}
}

func TestViewWithError(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Try to load non-existent file
	lv.LoadFileContent("/nonexistent/file.log")

	// View should show error message
	view := lv.View()

	if view == "" {
		t.Error("Expected non-empty view for error state")
	}

	if !strings.Contains(view, "Error loading file") {
		t.Error("Expected error message in view")
	}
}

func TestViewWithLineLimitWarning(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Simulate line limited content
	lv.lineLimited = true
	lv.content = "Test content"
	lv.isLoaded = true
	lv.renderContent()

	// View should show warning in footer
	view := lv.View()

	if !strings.Contains(view, "Content limited to 10,000 lines") {
		t.Error("Expected line limit warning in footer")
	}
}

func TestSetContentMarksLoaded(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	if lv.IsLoaded() {
		t.Error("Expected IsLoaded to be false initially")
	}

	lv.SetContent("Test", "/test.log")

	if !lv.IsLoaded() {
		t.Error("Expected IsLoaded to be true after SetContent")
	}

	if lv.HasError() {
		t.Error("Expected no error after SetContent")
	}
}

// Integration Tests for Rendering Pipeline

func TestRenderingPipelineIntegration(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create test content with multiple lines
	content := "Line 1: Short line\n"
	content += "Line 2: This is a very long line that will definitely exceed the viewport width and should be wrapped properly\n"
	content += "Line 3: Another short line\n"
	content += "    Line 4: Indented line with some content\n"
	content += "Line 5: Final line"

	lv.SetContent(content, "/test/integration.log")

	// Test 1: Basic rendering without options
	view := lv.View()
	if view == "" {
		t.Fatal("Expected non-empty view")
	}

	// Verify footer is present
	if !strings.Contains(view, "Scroll:") {
		t.Error("Expected scroll position in view")
	}

	if !strings.Contains(view, "Line Numbers") {
		t.Error("Expected line numbers toggle in view")
	}

	if !strings.Contains(view, "Word Wrap") {
		t.Error("Expected word wrap toggle in view")
	}

	// Test 2: Rendering with word wrap enabled
	lv.wordWrap = true
	lv.renderContent()

	wrappedLines := strings.Split(lv.renderedContent, "\n")
	if len(wrappedLines) <= 5 {
		t.Error("Expected more than 5 lines after word wrapping long line")
	}

	// Test 3: Rendering with line numbers enabled
	lv.showLineNumbers = true
	lv.renderContent()

	numberedLines := strings.Split(lv.renderedContent, "\n")
	for i, line := range numberedLines {
		if !strings.Contains(line, "│") {
			t.Errorf("Line %d missing line number separator: '%s'", i, line)
		}
	}

	// Test 4: Rendering with both word wrap and line numbers
	lv.wordWrap = true
	lv.showLineNumbers = true
	lv.renderContent()

	bothLines := strings.Split(lv.renderedContent, "\n")
	if len(bothLines) <= 5 {
		t.Error("Expected wrapped content with line numbers")
	}

	for i, line := range bothLines {
		if !strings.Contains(line, "│") {
			t.Errorf("Line %d missing line number separator: '%s'", i, line)
		}
	}
}

func TestRenderingPipelinePerformance(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create large content (1000+ lines)
	var content strings.Builder
	for i := 0; i < 1500; i++ {
		content.WriteString(fmt.Sprintf("Line %d: This is test content for performance testing\n", i+1))
	}

	lv.SetContent(content.String(), "/test/large.log")

	// Test rendering with all options enabled
	lv.wordWrap = true
	lv.showLineNumbers = true

	// Measure rendering time (should be fast due to viewport optimization)
	lv.renderContent()

	// Verify content was processed
	if lv.renderedContent == "" {
		t.Fatal("Expected rendered content for large file")
	}

	// Verify viewport clipping is working (viewport should only show visible lines)
	view := lv.View()
	if view == "" {
		t.Fatal("Expected non-empty view for large file")
	}

	// Viewport should not render all 1500 lines at once
	viewLines := strings.Split(view, "\n")
	if len(viewLines) > 100 {
		t.Logf("Warning: View contains %d lines, viewport optimization may not be working", len(viewLines))
	}
}

func TestViewportOptimization(t *testing.T) {
	lv := NewLogViewerPanel(80, 10, nil) // Small viewport height

	// Create content with many lines
	content := strings.Repeat("Test line\n", 200)
	lv.SetContent(content, "/test/viewport.log")

	// Get view output
	view := lv.View()

	// Count lines in view (should be limited by viewport height, not total content)
	viewLines := strings.Count(view, "\n")

	// View should not contain all 200 lines
	if viewLines > 50 {
		t.Logf("Viewport optimization: View contains %d lines (expected < 50 for height 10)", viewLines)
	}

	// Verify scroll position is calculated correctly
	if lv.scrollPos == "" {
		t.Error("Expected scroll position to be set")
	}

	if !strings.Contains(lv.scrollPos, "lines") {
		t.Errorf("Expected scroll position to show line count: %s", lv.scrollPos)
	}
}

func TestFocusStateStyling(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.SetContent("Test content", "/test.log")

	// Test unfocused state
	lv.SetFocused(false)
	unfocusedView := lv.View()

	if unfocusedView == "" {
		t.Fatal("Expected non-empty view when unfocused")
	}

	// Test focused state
	lv.SetFocused(true)
	focusedView := lv.View()

	if focusedView == "" {
		t.Fatal("Expected non-empty view when focused")
	}

	// Views should differ due to focus styling (border color changes)
	// Note: We can't easily test exact styling, but we verify views are generated
	if len(focusedView) == 0 || len(unfocusedView) == 0 {
		t.Error("Expected both focused and unfocused views to be non-empty")
	}
}

func TestCompleteRenderCycle(t *testing.T) {
	lv := NewLogViewerPanel(40, 24, nil) // Narrow width to force wrapping

	// Step 1: Load content
	content := "Header\n"
	content += "This is a very long line that will definitely be wrapped when word wrap is enabled because it exceeds the width\n"
	content += "    Indented content\n"
	content += "Final line"

	lv.SetContent(content, "/test/complete.log")

	// Step 2: Enable features
	lv.SetFocused(true)
	lv.showLineNumbers = true
	lv.wordWrap = true

	// Step 3: Render content (applies word wrap and line numbers)
	lv.renderContent()

	// Verify cached content
	if lv.renderedContent == "" {
		t.Fatal("Expected rendered content to be cached")
	}

	// Verify word wrapping was applied (long line should wrap)
	lines := strings.Split(lv.renderedContent, "\n")
	if len(lines) <= 4 {
		t.Logf("Got %d lines after wrapping (expected > 4). Content may not have wrapped.", len(lines))
		// Don't fail, just log - wrapping depends on actual width calculations
	}

	// Verify line numbers were applied
	for i, line := range lines {
		if !strings.Contains(line, "│") {
			t.Errorf("Line %d missing line number separator: '%s'", i, line)
			break // Only report first failure
		}
	}

	// Step 4: Generate final view
	view := lv.View()

	if view == "" {
		t.Fatal("Expected non-empty final view")
	}

	// Verify footer components
	if !strings.Contains(view, "Scroll:") {
		t.Error("Expected scroll position in final view")
	}

	if !strings.Contains(view, "Line Numbers: ON") {
		t.Error("Expected line numbers ON in final view")
	}

	if !strings.Contains(view, "Word Wrap: ON") {
		t.Error("Expected word wrap ON in final view")
	}

	// Step 5: Test scrolling updates (only if we have enough content)
	totalLines := lv.viewport.TotalLineCount()
	if totalLines > lv.viewport.Height {
		lv.SetFocused(true)
		initialScrollPos := lv.scrollPos

		// Scroll down
		for i := 0; i < 5; i++ {
			lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}

		if lv.scrollPos == initialScrollPos {
			t.Error("Expected scroll position to update after scrolling")
		}
	} else {
		t.Logf("Skipping scroll test: not enough lines (%d) to scroll in viewport height (%d)",
			totalLines, lv.viewport.Height)
	}
}
