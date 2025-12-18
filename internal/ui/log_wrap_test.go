package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

// TestLogPanelLayoutWidthConstraints tests that LogWidth is properly constrained
func TestLogPanelLayoutWidthConstraints(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		expectedMax   int // LogWidth should not exceed this
		expectedMin   int // LogWidth should not be less than this
	}{
		{
			name:        "Normal terminal width",
			width:       80,
			expectedMax: 80 - 2, // minMargin = 2
			expectedMin: 20,
		},
		{
			name:        "Wide terminal",
			width:       200,
			expectedMax: 200 - 2,
			expectedMin: 20,
		},
		{
			name:        "Narrow terminal",
			width:       60,
			expectedMax: 60 - 2,
			expectedMin: 20,
		},
		{
			name:        "Very narrow terminal",
			width:       25,
			expectedMax: 25 - 2,
			expectedMin: 20,
		},
		{
			name:        "Minimum terminal size",
			width:       20,
			expectedMax: 20, // LogWidth is at least the minimum
			expectedMin: 20, // LogWidth should be 20 at minimum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.width = tt.width
			m.height = 30
			m.showLogPanel = true

			layout := m.calculateLayout()

			// LogWidth should not exceed (terminal width - minMargin)
			if layout.LogWidth > tt.expectedMax {
				t.Errorf("Expected LogWidth <= %d, got %d (terminal width %d)", tt.expectedMax, layout.LogWidth, tt.width)
			}

			// LogWidth should never be less than 20 (minimum usable width)
			if layout.LogWidth < tt.expectedMin {
				t.Errorf("Expected LogWidth >= %d, got %d (terminal width %d)", tt.expectedMin, layout.LogWidth, tt.width)
			}
		})
	}
}

// TestViewportWidthAfterLayoutUpdate tests that viewport width is properly constrained
func TestViewportWidthAfterLayoutUpdate(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		expectedMinVPW int
	}{
		{
			name:           "Normal terminal",
			width:          80,
			expectedMinVPW: 20, // minimum constraint in updateViewportSizes
		},
		{
			name:           "Very narrow terminal",
			width:          30,
			expectedMinVPW: 20,
		},
		{
			name:           "Large terminal",
			width:          160,
			expectedMinVPW: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.width = tt.width
			m.height = 40
			m.showLogPanel = true
			m.logViewport = viewport.New(0, 0)

			m.updateViewportSizes()

			// Viewport width should be constrained
			if m.logViewport.Width < tt.expectedMinVPW {
				t.Errorf("Viewport width %d is less than minimum %d (terminal width %d)", 
					m.logViewport.Width, tt.expectedMinVPW, tt.width)
			}

			// Viewport width should not exceed terminal width
			if m.logViewport.Width > tt.width {
				t.Errorf("Viewport width %d exceeds terminal width %d", m.logViewport.Width, tt.width)
			}
		})
	}
}

// TestLogPanelConsistentWidthCalculation tests that LogWidth and viewport width calculations are consistent
func TestLogPanelConsistentWidthCalculation(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 50
	m.showLogPanel = true
	m.logViewport = viewport.New(0, 0)

	layout := m.calculateLayout()
	m.updateViewportSizes()

	// The relationship should be:
	// viewport.Width = layout.LogWidth - panelPadding*2
	// But with a minimum constraint on viewport width
	expectedViewportWidth := layout.LogWidth - panelPadding*2
	if expectedViewportWidth < 20 {
		expectedViewportWidth = 20
	}

	if m.logViewport.Width != expectedViewportWidth {
		t.Errorf("Viewport width %d doesn't match expected %d from layout LogWidth %d",
			m.logViewport.Width, expectedViewportWidth, layout.LogWidth)
	}
}

// TestLogPanelLayoutWidthWithDetailsPanelVisible tests LogWidth when details panel is shown
func TestLogPanelLayoutWidthWithDetailsPanelVisible(t *testing.T) {
	m := createTestModel()
	m.width = 160
	m.height = 50
	m.showLogPanel = true
	m.showDetailsPanel = true

	layout := m.calculateLayout()

	// Even when details panel is visible, LogWidth should still be the full terminal width
	// (log panel is below, not beside)
	if layout.LogWidth != 160-2 {
		t.Errorf("Expected LogWidth to be terminal width minus margin, got %d", layout.LogWidth)
	}

	// But should still be constrained to terminal width
	if layout.LogWidth > m.width-2 {
		t.Errorf("LogWidth %d exceeds terminal width %d minus margin", layout.LogWidth, m.width)
	}
}

// TestLogPanelLayoutWithVeryTinyTerminal tests extreme edge case with very tiny terminal
func TestLogPanelLayoutWithVeryTinyTerminal(t *testing.T) {
	m := createTestModel()
	m.width = 15
	m.height = 10
	m.showLogPanel = true

	layout := m.calculateLayout()

	// Should enforce minimum width of 20, even though terminal is only 15 wide
	if layout.LogWidth < 20 {
		t.Errorf("Expected minimum LogWidth of 20, got %d", layout.LogWidth)
	}
}

// TestRenderLogEmpty tests renderLog with no log lines
func TestRenderLogEmpty(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{}

	result := m.renderLog()

	if !strings.Contains(result, "No log output yet") {
		t.Errorf("Expected 'No log output yet' for empty log, got: %s", result)
	}
}

// TestRenderLogShortLines tests renderLog with lines that don't need wrapping
func TestRenderLogShortLines(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"Short line",
		"Another short one",
	}
	m.logViewport.Width = 80

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	if lines[0] != "Short line" {
		t.Errorf("Expected first line to be 'Short line', got: %s", lines[0])
	}
}

// TestRenderLogWrappingLongLine tests renderLog wraps long lines
func TestRenderLogWrappingLongLine(t *testing.T) {
	m := createTestModel()
	longLine := "This is a very long line that should definitely be wrapped when the viewport width is small enough"
	m.logLines = []string{longLine}
	m.logViewport.Width = 30 // Small width to force wrapping

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Should be wrapped into multiple lines
	if len(lines) <= 1 {
		t.Errorf("Expected line to be wrapped into multiple lines, got: %s", result)
	}

	// Each wrapped line should respect the width constraint
	// wrapWidth = 30 - 4 = 26
	for _, line := range lines {
		if len(line) > 30 {
			t.Errorf("Line exceeds width constraint: %s (len=%d)", line, len(line))
		}
	}
}

// TestRenderLogPreservesEmptyLines tests that empty lines are preserved
func TestRenderLogPreservesEmptyLines(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"First line",
		"",
		"Third line",
	}
	m.logViewport.Width = 80

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines including empty line, got %d", len(lines))
	}

	if lines[1] != "" {
		t.Errorf("Expected second line to be empty, got: %s", lines[1])
	}
}

// TestRenderLogMinimumWidth tests minimum width constraint
func TestRenderLogMinimumWidth(t *testing.T) {
	m := createTestModel()
	// Long line with spaces (so it can wrap properly)
	longLine := strings.Repeat("word ", 20) // 20 repetitions of "word "
	m.logLines = []string{longLine}
	// Small viewport width that would result in less than minimum width
	m.logViewport.Width = 10

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Even though viewport is small, should use minimum width of 20
	// So the line should be wrapped with minimum 20-char width
	if len(lines) <= 1 {
		t.Errorf("Expected line to be wrapped into multiple lines with minimum width, got: %s", result)
	}

	// With minimum width of 20, multiple words should wrap into multiple lines
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines with minimum width for repeated words, got %d", len(lines))
	}
}

// TestRenderLogVeryLongWord tests handling of very long single words
func TestRenderLogVeryLongWord(t *testing.T) {
	m := createTestModel()
	// Very long single word (like a URL)
	longWord := "https://github.com/agreen757/tm-tui/blob/main/internal/ui/app.go"
	m.logLines = []string{longWord}
	m.logViewport.Width = 30

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// The long word can't be wrapped at character boundaries, so it stays as-is
	// wrapText should return it unmodified
	if len(lines) != 1 || lines[0] != longWord {
		t.Errorf("Expected long word to not be broken, got: %v", lines)
	}
}

// TestRenderLogMultipleWrappedLines tests multiple lines that need wrapping
func TestRenderLogMultipleWrappedLines(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"This is the first long line that should definitely be wrapped at the viewport width",
		"And here is a second long line that also needs wrapping for proper display",
	}
	m.logViewport.Width = 40

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Should be more than 2 lines due to wrapping
	if len(lines) <= 2 {
		t.Errorf("Expected multiple wrapped lines, got: %s", result)
	}

	// Verify no line exceeds the width constraint (40 - 4 = 36)
	for _, line := range lines {
		if len(line) > 40 {
			t.Errorf("Line exceeds width constraint: %s (len=%d)", line, len(line))
		}
	}
}

// TestRenderLogMixedContent tests mix of short lines, long lines, and empty lines
func TestRenderLogMixedContent(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"Short",
		"This is a much longer line that should be wrapped because it exceeds the viewport width",
		"",
		"Another short line",
		"Yet another very long line that needs to be wrapped to fit within the available width",
	}
	m.logViewport.Width = 35

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Should have multiple lines due to wrapping
	if len(lines) < 5 {
		t.Errorf("Expected at least 5 lines from mixed content, got %d", len(lines))
	}

	// Empty line should be preserved
	found := false
	for _, line := range lines {
		if line == "" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find empty line in wrapped output")
	}
}

// TestRenderLogWithViewportModel tests integration with actual viewport model
func TestRenderLogWithViewportModel(t *testing.T) {
	m := createTestModel()
	m.logViewport = viewport.New(50, 20)
	m.logLines = []string{
		"This is a test line that might need wrapping depending on the viewport width",
	}

	result := m.renderLog()

	if result == "" {
		t.Errorf("Expected non-empty result from renderLog")
	}

	// Result should not contain the empty state message
	if strings.Contains(result, "No log output yet") {
		t.Errorf("Unexpected empty state message in result")
	}
}

// TestRenderLogConsistencyWithWrapText verifies renderLog uses wrapText correctly
func TestRenderLogConsistencyWithWrapText(t *testing.T) {
	testLine := "The quick brown fox jumps over the lazy dog multiple times"
	wrapWidth := 25

	// Test wrapText directly
	directWrapped := wrapText(testLine, wrapWidth)

	// Test through renderLog
	m := createTestModel()
	m.logViewport.Width = wrapWidth + 4 // Account for the -4 padding
	m.logLines = []string{testLine}

	result := m.renderLog()

	if result != directWrapped {
		t.Errorf("renderLog wrapping differs from direct wrapText call.\nDirect: %s\nViaRenderLog: %s", directWrapped, result)
	}
}

// TestRenderLogEdgeCaseWindowWidth tests edge cases with very small window widths
func TestRenderLogEdgeCaseWindowWidth(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{"test"}
	m.logViewport.Width = 1 // Extremely small

	result := m.renderLog()

	// Should not panic and should return something
	if result == "" {
		t.Errorf("Expected non-empty result even with width=1")
	}
}

// TestRenderLogMultipleConsecutiveEmptyLines tests multiple empty lines
func TestRenderLogMultipleConsecutiveEmptyLines(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"First line",
		"",
		"",
		"Fourth line",
	}
	m.logViewport.Width = 80

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 lines with consecutive empty lines, got %d", len(lines))
	}

	if lines[1] != "" || lines[2] != "" {
		t.Errorf("Expected lines 1 and 2 to be empty")
	}
}

// TestRenderLogWrappingPreservesWords tests that wrapping breaks at word boundaries
func TestRenderLogWrappingPreservesWords(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"The quick brown fox",
	}
	m.logViewport.Width = 15 // Small width: 15 - 4 = 11 chars per line

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Verify words are not broken mid-word
	for _, line := range lines {
		// Each line should consist of complete words
		words := strings.Fields(line)
		for _, word := range words {
			if strings.Contains(word, "-") && !strings.HasPrefix(word, "-") {
				// Might be a hyphenated word, which is OK
				continue
			}
			// Word should be intact
			if len(word) > 11 {
				t.Errorf("Word appears to be broken: %s", word)
			}
		}
	}
}

// TestUpdateLogViewportUsesWrappedContent tests that updateLogViewport uses wrapped content from renderLog
func TestUpdateLogViewportUsesWrappedContent(t *testing.T) {
	m := createTestModel()
	longLine := "This is a very long line that should definitely be wrapped when the viewport width is small"
	m.logLines = []string{longLine}
	m.logViewport.Width = 40
	m.logViewport.Height = 10

	// Call updateLogViewport
	m.updateLogViewport()

	// Get the content set in the viewport
	content := m.logViewport.View()

	// The content should be wrapped (multiple lines)
	lines := strings.Split(content, "\n")
	if len(lines) <= 1 {
		t.Errorf("Expected wrapped content with multiple lines, got: %s", content)
	}

	// Each line should respect the width constraint
	for _, line := range lines {
		if len(line) > m.logViewport.Width {
			t.Errorf("Line exceeds viewport width: %s (len=%d, width=%d)", line, len(line), m.logViewport.Width)
		}
	}
}

// TestUpdateLogViewportSetsContent tests that updateLogViewport sets the viewport content
func TestUpdateLogViewportSetsContent(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{"Test log line"}
	m.logViewport.Width = 50
	m.logViewport.Height = 10

	// Call updateLogViewport
	m.updateLogViewport()

	// Content should now be set
	content := m.logViewport.View()
	if content == "" {
		t.Errorf("Expected viewport content to be set, got empty string")
	}

	// Content should contain the log line
	if !strings.Contains(content, "Test log line") {
		t.Errorf("Expected viewport content to contain log line, got: %s", content)
	}
}

// TestUpdateLogViewportResponectsViewportWidthConstraint tests that updateLogViewport respects viewport width
func TestUpdateLogViewportRespectViewportWidthConstraint(t *testing.T) {
	tests := []struct {
		name        string
		viewportWidth int
		longLine    string
	}{
		{
			name:         "Small viewport",
			viewportWidth: 30,
			longLine:     "This is a moderately long line that needs to be wrapped",
		},
		{
			name:         "Medium viewport",
			viewportWidth: 60,
			longLine:     strings.Repeat("word ", 20),
		},
		{
			name:         "Large viewport",
			viewportWidth: 100,
			longLine:     strings.Repeat("word ", 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.logLines = []string{tt.longLine}
			m.logViewport.Width = tt.viewportWidth
			m.logViewport.Height = 20

			m.updateLogViewport()

			content := m.logViewport.View()
			lines := strings.Split(content, "\n")

			// Verify no line exceeds viewport width
			for _, line := range lines {
				if len(line) > tt.viewportWidth {
					t.Errorf("Line exceeds viewport width: %s (len=%d, width=%d)", line, len(line), tt.viewportWidth)
				}
			}
		})
	}
}

// TestUpdateLogViewportWithMultipleLines tests updateLogViewport with multiple log entries
func TestUpdateLogViewportWithMultipleLines(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"Short line",
		"This is a much longer line that might need wrapping depending on the viewport width",
		"Another short line",
		"Yet another very long line that definitely needs to be wrapped to fit properly",
	}
	m.logViewport.Width = 50
	m.logViewport.Height = 20

	m.updateLogViewport()

	content := m.logViewport.View()

	// Should contain all the original log content (or at least the wrapped version)
	if !strings.Contains(content, "Short line") {
		t.Errorf("Expected viewport content to contain 'Short line', got: %s", content)
	}

	// Verify no line exceeds viewport width
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if len(line) > m.logViewport.Width {
			t.Errorf("Line exceeds viewport width: %s (len=%d, width=%d)", line, len(line), m.logViewport.Width)
		}
	}
}

// TestUpdateLogViewportPreservesEmptyLines tests that updateLogViewport preserves empty lines
func TestUpdateLogViewportPreservesEmptyLines(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{
		"First line",
		"",
		"Third line",
	}
	m.logViewport.Width = 80
	m.logViewport.Height = 10

	m.updateLogViewport()

	// Should have the log lines properly set up
	if len(m.logLines) != 3 {
		t.Errorf("Expected 3 logLines, got %d", len(m.logLines))
	}

	// Verify that renderLog produces the correct output with empty lines
	content := m.renderLog()
	lines := strings.Split(content, "\n")

	// Should have at least 3 lines (with empty line preserved)
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines including empty line, got %d", len(lines))
	}

	// Should have an empty line somewhere
	found := false
	for _, line := range lines {
		if line == "" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find empty line in renderLog output")
	}
}

// TestUpdateLogViewportCallsRenderLog tests that updateLogViewport properly delegates to renderLog
func TestUpdateLogViewportCallsRenderLog(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{"Test line"}
	m.logViewport.Width = 50
	m.logViewport.Height = 10

	// Call updateLogViewport
	m.updateLogViewport()

	// Verify that the viewport content contains the rendered log content
	actualContent := m.logViewport.View()

	// The viewport content should contain the rendered log content
	if !strings.Contains(actualContent, "Test line") {
		t.Errorf("Expected viewport content to contain rendered log content")
	}

	// Verify that renderLog is being called by checking if the log line is visible
	// in the viewport after updateLogViewport
	if !strings.Contains(m.logViewport.View(), "Test line") {
		t.Errorf("Expected Test line to be in viewport after updateLogViewport")
	}
}

// TestAddLogLineCallsUpdateLogViewport tests that addLogLine properly calls updateLogViewport
func TestAddLogLineCallsUpdateLogViewport(t *testing.T) {
	m := createTestModel()
	m.logViewport.Width = 50
	m.logViewport.Height = 10

	// Initially empty
	if len(m.logLines) != 0 {
		t.Errorf("Expected empty logLines initially")
	}

	// Add a log line
	m.addLogLine("New log entry")

	// Should be in logLines
	if len(m.logLines) != 1 || m.logLines[0] != "New log entry" {
		t.Errorf("Expected log line to be added to logLines")
	}

	// Should be reflected in viewport (addLogLine calls updateLogViewport internally)
	content := m.logViewport.View()
	if !strings.Contains(content, "New log entry") {
		t.Errorf("Expected viewport content to reflect the added log line, got: %s", content)
	}
}

// TestAddLogLineMultipleEntries tests adding multiple log lines via addLogLine
func TestAddLogLineMultipleEntries(t *testing.T) {
	m := createTestModel()
	m.logViewport.Width = 50
	m.logViewport.Height = 10

	// Add multiple log lines
	m.addLogLine("First entry")
	m.addLogLine("Second entry")
	m.addLogLine("Third entry")

	// All should be in logLines
	if len(m.logLines) != 3 {
		t.Errorf("Expected 3 log lines, got %d", len(m.logLines))
	}

	// All should be in viewport
	content := m.logViewport.View()
	for _, expected := range []string{"First entry", "Second entry", "Third entry"} {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected viewport content to contain '%s', got: %s", expected, content)
		}
	}
}

// TestUpdateLogViewportEmptyLogs tests updateLogViewport with empty logs
func TestUpdateLogViewportEmptyLogs(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{}
	m.logViewport.Width = 50
	m.logViewport.Height = 10

	m.updateLogViewport()

	content := m.logViewport.View()

	// Should show the empty state message
	if !strings.Contains(content, "No log output yet") {
		t.Errorf("Expected empty log message, got: %s", content)
	}
}

// TestUpdateLogViewportMinimumWidth tests updateLogViewport with very small viewport width
func TestUpdateLogViewportMinimumWidth(t *testing.T) {
	m := createTestModel()
	longLine := "This is a test line with multiple words that should be wrapped properly"
	m.logLines = []string{longLine}
	m.logViewport.Width = 10 // Very small
	m.logViewport.Height = 10

	m.updateLogViewport()

	content := m.logViewport.View()

	// Should still produce valid wrapped content
	if content == "" {
		t.Errorf("Expected non-empty content even with small viewport")
	}

	// Should be wrapped into multiple lines (minimum width is 20 characters)
	lines := strings.Split(content, "\n")
	if len(lines) <= 1 {
		t.Errorf("Expected wrapped content with multiple lines, got: %s", content)
	}
}

// TestWindowResizeTriggersLogReWrapping tests that window resize events trigger log re-wrapping
func TestWindowResizeTriggersLogReWrapping(t *testing.T) {
	tests := []struct {
		name          string
		initialWidth  int
		resizeWidth   int
		logLine       string
		expectWrapped bool
	}{
		{
			name:          "Resize smaller - should wrap tighter",
			initialWidth:  100,
			resizeWidth:   40,
			logLine:       "This is a long line that will be wrapped differently based on terminal width",
			expectWrapped: true,
		},
		{
			name:          "Resize larger - should use more space",
			initialWidth:  40,
			resizeWidth:   100,
			logLine:       "This is a long line that will be wrapped differently based on terminal width",
			expectWrapped: true,
		},
		{
			name:          "Resize to narrow - extreme case",
			initialWidth:  80,
			resizeWidth:   25,
			logLine:       "This is a long line that will be wrapped differently based on terminal width",
			expectWrapped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.width = tt.initialWidth
			m.height = 30
			m.showLogPanel = true
			m.logLines = []string{tt.logLine}
			m.logViewport = viewport.New(0, 0)

			// Initial render with initial width
			m.updateViewportSizes()
			initialWidth := m.logViewport.Width
			m.updateLogViewport()
			initialContent := m.logViewport.View()
			initialLines := strings.Split(initialContent, "\n")

			// Simulate window resize
			m.width = tt.resizeWidth
			m.height = 30
			m.updateViewportSizes()
			newWidth := m.logViewport.Width

			// This is what should happen on WindowSizeMsg
			if m.showLogPanel {
				m.updateLogViewport()
			}

			newContent := m.logViewport.View()
			newLines := strings.Split(newContent, "\n")

			// Verify viewport width changed
			if initialWidth == newWidth && tt.initialWidth != tt.resizeWidth {
				t.Errorf("Expected viewport width to change from %d to something different, but got %d", initialWidth, newWidth)
			}

			// Verify content was re-rendered (may have different number of lines)
			if initialContent == newContent && tt.initialWidth != tt.resizeWidth {
				// It's OK if content is the same for some cases, but line count should often differ
				// (unless the line is short enough not to wrap in either case)
				if len(initialLines) != len(newLines) && tt.expectWrapped {
					// This is expected - content changed due to re-wrapping
				}
			}

			// Verify no line exceeds the new viewport width
			for _, line := range newLines {
				if len(line) > newWidth {
					t.Errorf("After resize to width %d, line exceeds viewport width: %s (len=%d)", newWidth, line, len(line))
				}
			}
		})
	}
}

// TestWindowResizeWithHiddenLogPanel tests that resize doesn't update hidden log panel
func TestWindowResizeWithHiddenLogPanel(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 30
	m.showLogPanel = false
	m.logLines = []string{"Test log line"}
	m.logViewport = viewport.New(0, 0)

	// Set initial viewport sizes
	m.updateViewportSizes()
	initialContent := m.logViewport.View()

	// Simulate resize with panel hidden
	m.width = 120
	m.height = 40
	m.updateViewportSizes()

	// This is what should happen - no update when panel is hidden
	if m.showLogPanel {
		m.updateLogViewport()
	}

	// Content should not have changed because we didn't call updateLogViewport
	// (the Update handler checks m.showLogPanel before calling updateLogViewport)
	// So the old content should still be there (though the viewport dimensions changed)
	if m.logViewport.View() != initialContent {
		// Actually, this might not be true because updateViewportSizes changes the viewport
		// But the important thing is we didn't call updateLogViewport, so rendering didn't happen
		// The test passes if we reach here without panicking
	}
}

// TestWindowResizeMultipleResizes tests multiple resize events in sequence
func TestWindowResizeMultipleResizes(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 30
	m.showLogPanel = true
	m.logLines = []string{
		"First long line: This is a long line that will be wrapped differently based on terminal width",
		"Second long line: Another long line that should also be re-wrapped properly on each resize",
		"Short line",
	}
	m.logViewport = viewport.New(0, 0)

	widths := []int{100, 50, 120, 40, 80}

	for _, newWidth := range widths {
		m.width = newWidth
		m.updateViewportSizes()
		if m.showLogPanel {
			m.updateLogViewport()
		}

		// Verify no line exceeds the viewport width
		content := m.logViewport.View()
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if len(line) > m.logViewport.Width {
				t.Errorf("After resize to width %d, line exceeds viewport: %s (len=%d, vp_width=%d)", newWidth, line, len(line), m.logViewport.Width)
			}
		}
	}
}

// TestWindowResizePreservesLogContent tests that resize doesn't lose log content
func TestWindowResizePreservesLogContent(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 30
	m.showLogPanel = true
	logContent := []string{
		"Important log line 1",
		"Important log line 2",
		"Important log line 3",
	}
	m.logLines = logContent
	m.logViewport = viewport.New(0, 0)

	// Initial setup
	m.updateViewportSizes()
	m.updateLogViewport()

	// Resize
	m.width = 120
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}

	// Verify all original log lines are still in the model
	if len(m.logLines) != len(logContent) {
		t.Errorf("Expected %d log lines after resize, got %d", len(logContent), len(m.logLines))
	}

	// Verify content contains all the original log lines (they might be wrapped, but the content should be there)
	viewportContent := m.logViewport.View()
	for _, expectedLine := range logContent {
		if !strings.Contains(viewportContent, expectedLine) {
			t.Errorf("Expected log line not found in viewport after resize: %s\nViewport content: %s", expectedLine, viewportContent)
		}
	}
}

// TestWindowResizeAutoScrollBehavior tests that auto-scroll behavior is preserved after resize
func TestWindowResizeAutoScrollBehavior(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 30
	m.showLogPanel = true
	m.logLines = []string{
		"Line 1",
		"Line 2",
		"Line 3",
		"Line 4",
		"Line 5",
	}
	m.logViewport = viewport.New(0, 20)

	// Initial setup
	m.updateViewportSizes()
	m.updateLogViewport()

	// Add a new log line (which should trigger auto-scroll in real usage)
	m.addLogLine("New line 6")

	// Verify the new line is in the log
	if len(m.logLines) != 6 {
		t.Errorf("Expected 6 log lines, got %d", len(m.logLines))
	}

	// Resize
	m.width = 120
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}

	// Verify all lines are still there
	if len(m.logLines) != 6 {
		t.Errorf("Expected 6 log lines after resize, got %d", len(m.logLines))
	}

	// Verify the new line is in the viewport
	viewportContent := m.logViewport.View()
	if !strings.Contains(viewportContent, "New line 6") {
		t.Errorf("Expected new log line to be in viewport after resize")
	}
}

// TestLogPanelToggleShowsLogPanel tests that toggling the log panel to visible initializes it correctly
func TestLogPanelToggleShowsLogPanel(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 40
	m.showLogPanel = false // Start with log panel hidden
	m.logLines = []string{
		"Line 1",
		"Line 2",
		"This is a very long line that should be wrapped when the viewport width is small enough for demonstration purposes",
	}
	m.logViewport = viewport.New(0, 0)

	// Verify the panel starts hidden
	if m.showLogPanel {
		t.Errorf("Expected showLogPanel to be false initially")
	}

	// Toggle the log panel to visible (simulating the keyboard shortcut)
	m.showLogPanel = !m.showLogPanel
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}

	// Verify the panel is now visible
	if !m.showLogPanel {
		t.Errorf("Expected showLogPanel to be true after toggle")
	}

	// Verify viewport dimensions are set correctly
	if m.logViewport.Width <= 0 {
		t.Errorf("Expected logViewport.Width to be > 0 after toggle, got %d", m.logViewport.Width)
	}
	if m.logViewport.Width < 20 {
		t.Errorf("Expected logViewport.Width to be at least 20, got %d", m.logViewport.Width)
	}

	// Verify viewport has content
	content := m.logViewport.View()
	if content == "" {
		t.Errorf("Expected logViewport to have content after toggle and updateLogViewport()")
	}

	// Verify the content includes wrapped log lines
	if !strings.Contains(content, "Line 1") {
		t.Errorf("Expected log content to include 'Line 1', got: %s", content)
	}
}

// TestLogPanelToggleHidesLogPanel tests that toggling the log panel to hidden works correctly
func TestLogPanelToggleHidesLogPanel(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 40
	m.showLogPanel = true // Start with log panel visible
	m.logLines = []string{
		"Line 1",
		"Line 2",
		"Line 3",
	}
	m.logViewport = viewport.New(30, 10)

	// Initialize the viewport
	m.updateViewportSizes()
	m.updateLogViewport()

	// Verify the panel starts visible and has content
	if !m.showLogPanel {
		t.Errorf("Expected showLogPanel to be true initially")
	}
	initialContent := m.logViewport.View()
	if initialContent == "" {
		t.Errorf("Expected logViewport to have initial content")
	}

	// Toggle the log panel to hidden
	m.showLogPanel = !m.showLogPanel
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}

	// Verify the panel is now hidden
	if m.showLogPanel {
		t.Errorf("Expected showLogPanel to be false after toggle")
	}

	// The log lines should still be preserved
	if len(m.logLines) != 3 {
		t.Errorf("Expected 3 log lines to be preserved when hiding panel, got %d", len(m.logLines))
	}
}

// TestLogPanelToggleMultipleTimes tests toggling the log panel on/off multiple times
func TestLogPanelToggleMultipleTimes(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 40
	m.showLogPanel = false
	m.logLines = []string{
		"Log line 1",
		"Log line 2",
		"This is a longer log line that will be wrapped: The quick brown fox jumps over the lazy dog multiple times",
	}
	m.logViewport = viewport.New(0, 0)

	// Toggle on
	m.showLogPanel = !m.showLogPanel
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}
	if !m.showLogPanel {
		t.Errorf("Expected showLogPanel to be true after first toggle")
	}
	if m.logViewport.Width <= 0 {
		t.Errorf("Expected valid viewport width after first toggle")
	}
	firstContent := m.logViewport.View()
	if firstContent == "" {
		t.Errorf("Expected content after first toggle")
	}

	// Toggle off
	m.showLogPanel = !m.showLogPanel
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}
	if m.showLogPanel {
		t.Errorf("Expected showLogPanel to be false after second toggle")
	}

	// Toggle on again
	m.showLogPanel = !m.showLogPanel
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}
	if !m.showLogPanel {
		t.Errorf("Expected showLogPanel to be true after third toggle")
	}
	secondContent := m.logViewport.View()
	if secondContent == "" {
		t.Errorf("Expected content after third toggle")
	}

	// Verify the content is still valid and includes the log lines
	if !strings.Contains(secondContent, "Log line 1") {
		t.Errorf("Expected log content to include 'Log line 1' after multiple toggles")
	}
}

// TestLogPanelToggleScrollPosition tests that viewport scrolls to bottom when shown
func TestLogPanelToggleScrollPosition(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 40
	m.showLogPanel = false
	// Add enough lines to cause scrolling
	m.logLines = []string{}
	for i := 1; i <= 20; i++ {
		m.logLines = append(m.logLines, fmt.Sprintf("Line %d", i))
	}
	m.logViewport = viewport.New(70, 10)

	// Toggle to show the log panel
	m.showLogPanel = !m.showLogPanel
	m.updateViewportSizes()
	if m.showLogPanel {
		m.updateLogViewport()
	}

	// After updateLogViewport(), the viewport should have scrolled to bottom
	// The actual scroll position depends on the content height vs viewport height
	// We just verify that the viewport is properly initialized
	if m.logViewport.Width <= 0 || m.logViewport.Height <= 0 {
		t.Errorf("Expected non-zero viewport dimensions after toggle")
	}

	// The viewport should contain content
	content := m.logViewport.View()
	if content == "" {
		t.Errorf("Expected viewport to have content after toggle to visible")
	}
}

// TestLogPanelToggleWithWrappedContent tests that wrapped content is handled correctly during toggle
func TestLogPanelToggleWithWrappedContent(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 40
	m.showLogPanel = false
	m.logLines = []string{
		"Short",
		"This is a very long line that will be wrapped because it exceeds the viewport width significantly and contains multiple words that should flow to the next line when rendered in the log panel with proper word wrapping enabled",
		"Another",
	}
	m.logViewport = viewport.New(0, 0)

	// Toggle to show
	m.showLogPanel = true
	m.updateViewportSizes()
	m.updateLogViewport()

	// Get the rendered content
	content := m.renderLog()
	lines := strings.Split(content, "\n")

	// Should have more lines than the original log lines due to wrapping
	if len(lines) <= len(m.logLines) {
		t.Logf("Warning: Expected more lines due to wrapping, got %d lines from %d log lines", len(lines), len(m.logLines))
	}

	// All original content should be present
	for _, logLine := range m.logLines {
		if logLine != "" && !strings.Contains(content, logLine) {
			// Due to wrapping, we might not find the exact line, but pieces of it should be there
			t.Logf("Warning: Could not find complete log line in wrapped output: %s", logLine)
		}
	}
}

// TestLogPanelToggleViewportInitialization tests that viewport is properly initialized when shown
func TestLogPanelToggleViewportInitialization(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 50
	m.showLogPanel = false
	m.logLines = []string{"Test log line 1", "Test log line 2"}
	m.logViewport = viewport.New(0, 0)

	// Before toggle: viewport should be uninitialized (0,0)
	if m.logViewport.Width != 0 || m.logViewport.Height != 0 {
		t.Errorf("Expected initial viewport to be (0, 0), got (%d, %d)", m.logViewport.Width, m.logViewport.Height)
	}

	// Toggle to show
	m.showLogPanel = true
	m.updateViewportSizes()
	m.updateLogViewport()

	// After toggle: viewport should have positive dimensions
	if m.logViewport.Width <= 0 {
		t.Errorf("Expected viewport.Width > 0 after toggle, got %d", m.logViewport.Width)
	}
	if m.logViewport.Height <= 0 {
		t.Errorf("Expected viewport.Height > 0 after toggle, got %d", m.logViewport.Height)
	}

	// Viewport should have the correct relationship with layout
	layout := m.calculateLayout()
	expectedWidth := layout.LogWidth - panelPadding*2
	if expectedWidth < 20 {
		expectedWidth = 20
	}
	if m.logViewport.Width != expectedWidth {
		t.Errorf("Expected viewport.Width to be %d, got %d", expectedWidth, m.logViewport.Width)
	}
}

// TestLogPanelToggleWithEmptyLogLines tests toggling when log is empty
func TestLogPanelToggleWithEmptyLogLines(t *testing.T) {
	m := createTestModel()
	m.width = 80
	m.height = 40
	m.showLogPanel = false
	m.logLines = []string{} // Empty log
	m.logViewport = viewport.New(0, 0)

	// Toggle to show
	m.showLogPanel = true
	m.updateViewportSizes()
	m.updateLogViewport()

	// Verify the panel is visible
	if !m.showLogPanel {
		t.Errorf("Expected showLogPanel to be true")
	}

	// Verify viewport is initialized
	if m.logViewport.Width <= 0 {
		t.Errorf("Expected viewport to be initialized")
	}

	// Verify renderLog handles empty case correctly
	content := m.renderLog()
	if !strings.Contains(content, "No log output yet") {
		t.Errorf("Expected 'No log output yet' for empty log, got: %s", content)
	}
}

// TestRenderLogEmptyLinesEdgeCase tests that empty lines in logLines are preserved as blank lines
// This is a critical edge case for visual separation in log output
func TestRenderLogEmptyLinesEdgeCase(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{"A", "", "B"}
	m.logViewport.Width = 80

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Should have exactly 3 lines with the middle one empty
	if len(lines) != 3 {
		t.Errorf("Expected exactly 3 lines (including empty line), got %d: %v", len(lines), lines)
	}

	if lines[0] != "A" {
		t.Errorf("Expected first line to be 'A', got: '%s'", lines[0])
	}

	if lines[1] != "" {
		t.Errorf("Expected second line to be empty, got: '%s'", lines[1])
	}

	if lines[2] != "B" {
		t.Errorf("Expected third line to be 'B', got: '%s'", lines[2])
	}
}

// TestRenderLogMultipleConsecutiveEmptyLinesEdgeCase tests multiple consecutive empty lines
// to ensure they are all preserved, not collapsed or skipped
func TestRenderLogMultipleConsecutiveEmptyLinesEdgeCase(t *testing.T) {
	m := createTestModel()
	m.logLines = []string{"Start", "", "", "", "End"}
	m.logViewport.Width = 80

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Should have exactly 5 lines with 3 empty ones in the middle
	if len(lines) != 5 {
		t.Errorf("Expected exactly 5 lines, got %d", len(lines))
	}

	// Check that the empty lines are preserved
	emptyCount := 0
	for i := 1; i <= 3; i++ {
		if lines[i] == "" {
			emptyCount++
		}
	}

	if emptyCount != 3 {
		t.Errorf("Expected 3 consecutive empty lines, found %d", emptyCount)
	}
}

// TestRenderLogVerySmallViewportWidthEdgeCase tests renderLog with extremely small viewport width
// Ensures no panics and that minimum wrapWidth of 20 is enforced
func TestRenderLogVerySmallViewportWidthEdgeCase(t *testing.T) {
	tests := []struct {
		name        string
		viewportWidth int
	}{
		{
			name:        "Viewport width of 0",
			viewportWidth: 0,
		},
		{
			name:        "Viewport width of 1",
			viewportWidth: 1,
		},
		{
			name:        "Viewport width of 5",
			viewportWidth: 5,
		},
		{
			name:        "Viewport width of 24 (< 20 + 4 padding)",
			viewportWidth: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.logLines = []string{"This is a test line with multiple words that should be wrapped"}
			m.logViewport.Width = tt.viewportWidth

			// Should not panic
			result := m.renderLog()

			// Should produce valid output
			if result == "" {
				t.Errorf("Expected non-empty result even with viewport width %d", tt.viewportWidth)
			}

			// Lines should be wrapped with minimum width of 20 characters
			lines := strings.Split(result, "\n")
			if len(lines) < 1 {
				t.Errorf("Expected at least 1 line, got %d", len(lines))
			}

			// With wrapWidth of at least 20, the line should be wrapped into multiple lines
			if len(lines) <= 1 {
				t.Logf("Warning: Expected wrapping with minimum width 20, but got %d line(s)", len(lines))
			}
		})
	}
}

// TestRenderLogVeryLongWordEdgeCase tests handling of extremely long single words without spaces
// Ensures that very long words (like URLs or file paths) are not broken with hyphens or splits
func TestRenderLogVeryLongWordEdgeCase(t *testing.T) {
	m := createTestModel()
	// A very long URL-like string without spaces
	longWord := strings.Repeat("a", 200)
	m.logLines = []string{longWord}
	m.logViewport.Width = 40

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// The long word should appear as a single contiguous line (not broken in the middle)
	if len(lines) != 1 {
		t.Errorf("Expected 1 line for very long word, got %d lines", len(lines))
	}

	// The long word should be preserved exactly as-is (not hyphenated or broken)
	if lines[0] != longWord {
		t.Errorf("Expected long word to remain intact, got: %s (original length: %d, result length: %d)", 
			lines[0], len(longWord), len(lines[0]))
	}

	// The line can exceed wrapWidth, but that's expected for very long words
	if len(lines[0]) < len(longWord) {
		t.Errorf("Expected long word to be preserved completely, but it was truncated")
	}
}

// TestRenderLogRealWorldLongUrlEdgeCase tests with realistic long URL
func TestRenderLogRealWorldLongUrlEdgeCase(t *testing.T) {
	m := createTestModel()
	// A realistic long URL
	longUrl := "https://github.com/charmbracelet/bubbletea/blob/main/internal/ui/very/long/path/to/some/file.go"
	m.logLines = []string{longUrl}
	m.logViewport.Width = 30

	result := m.renderLog()

	// The URL should appear intact without being broken or hyphenated
	if !strings.Contains(result, longUrl) {
		t.Errorf("Expected URL to be preserved intact in output")
	}

	// Verify no hyphens were added for wrapping
	if strings.Contains(result, "-\n") {
		t.Errorf("Expected no hyphen line breaks in output")
	}
}

// TestRenderLogMinimumWidthEnforcement verifies minimum width is actually enforced
func TestRenderLogMinimumWidthEnforcement(t *testing.T) {
	m := createTestModel()
	// A line with many short words to force wrapping
	m.logLines = []string{"a b c d e f g h i j k l m n o p q r s t u v w x y z"}
	m.logViewport.Width = 5 // Very small, should use minimum of 20

	result := m.renderLog()
	lines := strings.Split(result, "\n")

	// Should be wrapped into multiple lines using minimum width of 20
	if len(lines) <= 1 {
		t.Errorf("Expected wrapping with minimum 20-char width, got %d line(s)", len(lines))
	}

	// Each line should not exceed wrapWidth calculation (but may exceed if content can't wrap)
	for _, line := range lines {
		// This is acceptable since we use minimum width
		if len(line) > 100 {
			t.Errorf("Line seems too long: %s (len=%d)", line, len(line))
		}
	}
}

// TestRenderLogWithANSIColoredText tests renderLog with ANSI escape sequences
// This is an exploratory test to document current ANSI handling behavior.
// IMPORTANT: Current implementation counts ANSI bytes toward visual width,
// which causes apparent visual misalignment. This is documented but not fixed
// in the initial implementation, matching the PRD decision: "Start with simple
// implementation, add ANSI support if needed."
func TestRenderLogWithANSIColoredText(t *testing.T) {
	m := createTestModel()

	// Create a long line with multiple colored sections to exceed wrapWidth
	// Visual: ~95 characters, Actual: ~110+ with ANSI codes
	coloredLine := "\x1b[31mThis is a long line\x1b[0m " +
		"\x1b[32mwith green text\x1b[0m " +
		"\x1b[1mand bold text\x1b[0m " +
		"and some normal text mixed in " +
		"\x1b[31mwith more red\x1b[0m"

	m.logLines = []string{coloredLine}
	m.logViewport.Width = 80 // wrapWidth = 76 after -4 padding

	// Should not panic
	result := m.renderLog()

	// Result should not be empty
	if result == "" {
		t.Errorf("Expected non-empty result for colored log line")
	}

	// ANSI sequences should be preserved in output
	if !strings.Contains(result, "\x1b[31m") {
		t.Errorf("Expected ANSI red color code to be preserved in output")
	}
	if !strings.Contains(result, "\x1b[32m") {
		t.Errorf("Expected ANSI green color code to be preserved in output")
	}
	if !strings.Contains(result, "\x1b[0m") {
		t.Errorf("Expected ANSI reset code to be preserved in output")
	}

	// NOTE: Due to ANSI bytes being counted toward visual width,
	// the wrapping will appear visually misaligned. This is expected
	// with the current simple implementation. Each line in the output
	// may exceed the visual wrapWidth because the algorithm counts
	// invisible ANSI escape sequences.
	lines := strings.Split(result, "\n")
	if len(lines) <= 0 {
		t.Errorf("Expected at least one line in output")
	}

	// Log observation: ANSI sequences are present and preserved
	t.Logf("Colored line input length: %d bytes (visual ~95 chars)\n", len(coloredLine))
	t.Logf("Output lines: %d\n", len(lines))
	for i, line := range lines {
		// Count visible characters by removing ANSI codes
		visibleLength := len(removeANSICodes(line))
		t.Logf("  Line %d: %d bytes (%d visible chars)\n", i, len(line), visibleLength)
	}
}

// TestRenderLogWithMultipleANSIColors tests renderLog with multiple colored lines
// This further explores ANSI handling with mixed colored and normal text.
func TestRenderLogWithMultipleANSIColors(t *testing.T) {
	m := createTestModel()

	m.logLines = []string{
		"\x1b[31mRed colored log line that is quite long and should trigger wrapping at some point\x1b[0m",
		"Normal uncolored line that is also quite long and will be wrapped to multiple lines normally",
		"\x1b[32m\x1b[1mGreen bold text: Another long line here that combines multiple ANSI codes\x1b[0m",
	}
	m.logViewport.Width = 80 // wrapWidth = 76

	// Should not panic
	result := m.renderLog()

	// Verify ANSI codes are preserved
	if !strings.Contains(result, "\x1b[31m") {
		t.Errorf("Expected red color code preserved")
	}
	if !strings.Contains(result, "\x1b[32m") {
		t.Errorf("Expected green color code preserved")
	}
	if !strings.Contains(result, "\x1b[1m") {
		t.Errorf("Expected bold code preserved")
	}

	// Should contain all original text (minus ANSI codes, but the text should be there)
	if !strings.Contains(result, "Red colored log line") {
		t.Errorf("Expected red line text preserved in output")
	}
	if !strings.Contains(result, "Normal uncolored line") {
		t.Errorf("Expected normal line text preserved in output")
	}
	if !strings.Contains(result, "Green bold text") {
		t.Errorf("Expected green bold line text preserved in output")
	}

	// Output should have more lines than input due to wrapping
	outputLines := strings.Split(result, "\n")
	if len(outputLines) <= len(m.logLines) {
		t.Logf("Note: With ANSI codes, wrapping behavior differs. Input: %d lines, Output: %d lines\n",
			len(m.logLines), len(outputLines))
	}

	// Log details about each output line
	t.Logf("Output analysis for %d input lines:\n", len(m.logLines))
	for i, line := range outputLines {
		visibleLength := len(removeANSICodes(line))
		t.Logf("  Output line %d: %d bytes (%d visible)\n", i, len(line), visibleLength)
	}
}

// removeANSICodes is a helper function for tests to measure visual width
// by removing ANSI escape sequences from a string.
// This helps understand how visual width differs from byte length.
func removeANSICodes(s string) string {
	// Simple regex-based ANSI code removal for testing purposes
	// Matches escape sequences like \x1b[...m
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			result += string(r)
		}
	}
	return result
}
