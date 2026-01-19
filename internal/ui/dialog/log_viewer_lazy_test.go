package dialog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadFileLines tests the readFileLines function
func TestReadFileLines(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	// Write test content
	testContent := strings.Join([]string{
		"Line 1",
		"Line 2",
		"Line 3",
		"Line 4",
		"Line 5",
		"Line 6",
		"Line 7",
		"Line 8",
		"Line 9",
		"Line 10",
	}, "\n")

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		startLine int
		numLines  int
		expected  []string
	}{
		{
			name:      "Read first 3 lines",
			startLine: 0,
			numLines:  3,
			expected:  []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:      "Read middle lines",
			startLine: 3,
			numLines:  3,
			expected:  []string{"Line 4", "Line 5", "Line 6"},
		},
		{
			name:      "Read last lines",
			startLine: 7,
			numLines:  3,
			expected:  []string{"Line 8", "Line 9", "Line 10"},
		},
		{
			name:      "Read beyond file end",
			startLine: 8,
			numLines:  5,
			expected:  []string{"Line 9", "Line 10"},
		},
		{
			name:      "Read single line",
			startLine: 4,
			numLines:  1,
			expected:  []string{"Line 5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := readFileLines(testFile, tt.startLine, tt.numLines)
			if err != nil {
				t.Fatalf("readFileLines failed: %v", err)
			}

			if len(lines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(lines))
			}

			for i, line := range lines {
				if i >= len(tt.expected) {
					t.Errorf("Extra line at index %d: %s", i, line)
					continue
				}
				if line != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], line)
				}
			}
		})
	}
}

// TestReadFileLinesErrors tests error handling in readFileLines
func TestReadFileLinesErrors(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Non-existent file",
			filePath: "/nonexistent/file.log",
			wantErr:  true,
		},
		{
			name:     "Empty path",
			filePath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readFileLines(tt.filePath, 0, 10)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFileLines() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCountFileLines tests the countFileLines function
func TestCountFileLines(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	tests := []struct {
		name        string
		lineCount   int
		expectedErr bool
	}{
		{
			name:      "Empty file",
			lineCount: 0,
		},
		{
			name:      "Single line",
			lineCount: 1,
		},
		{
			name:      "Multiple lines",
			lineCount: 100,
		},
		{
			name:      "Large file",
			lineCount: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create file with specified number of lines
			var lines []string
			for i := 0; i < tt.lineCount; i++ {
				lines = append(lines, "Test line")
			}
			content := strings.Join(lines, "\n")
			if tt.lineCount > 0 {
				content += "\n"
			}

			err := os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			count, err := countFileLines(testFile)
			if err != nil {
				t.Fatalf("countFileLines failed: %v", err)
			}

			// Allow for empty file edge case
			expectedCount := tt.lineCount
			if tt.lineCount == 0 {
				expectedCount = 0
			}

			if count != expectedCount {
				t.Errorf("Expected %d lines, got %d", expectedCount, count)
			}
		})
	}
}

// TestRenderVisible tests the windowed rendering function
func TestRenderVisible(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create test content with 100 lines
	var lines []string
	for i := 1; i <= 100; i++ {
		lines = append(lines, "Line "+string(rune('0'+i)))
	}
	content := strings.Join(lines, "\n")
	lv.SetContent(content, "test.log")

	// Test rendering visible portion
	rendered := lv.renderVisible()

	// Should have some content
	if rendered == "" {
		t.Error("renderVisible returned empty string")
	}

	// Test that it doesn't render all 100 lines at once
	renderedLines := strings.Split(rendered, "\n")
	if len(renderedLines) > lv.viewport.Height+10 {
		t.Errorf("Rendered too many lines: %d (expected <= %d)", len(renderedLines), lv.viewport.Height+10)
	}
}

// TestLazyLoadContent tests the lazy loading command generation
func TestLazyLoadContent(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Test with lazy loading disabled
	cmd := lv.lazyLoadContent()
	if cmd != nil {
		t.Error("Expected nil cmd when lazy loading is disabled")
	}

	// Enable lazy loading
	lv.lazyLoadEnabled = true
	lv.totalFileLines = 1000

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	var lines []string
	for i := 1; i <= 1000; i++ {
		lines = append(lines, "Line "+string(rune('0'+(i%10))))
	}
	content := strings.Join(lines, "\n")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lv.filePath = testFile
	lv.visibleStartLine = 0
	lv.visibleEndLine = 0 // Force load

	// Test command generation
	cmd = lv.lazyLoadContent()
	if cmd == nil {
		t.Error("Expected non-nil cmd when lazy loading is enabled and content needs loading")
	}
}

// TestApplyLineNumbersWithOffset tests line numbering with offset
func TestApplyLineNumbersWithOffset(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Set up content with 100 lines
	var contentLines []string
	for i := 1; i <= 100; i++ {
		contentLines = append(contentLines, "Line content")
	}
	lv.contentLines = contentLines

	// Test lines starting at offset 50
	testLines := []string{"Content 1", "Content 2", "Content 3"}
	offset := 50

	numbered := lv.applyLineNumbersWithOffset(testLines, offset)

	if len(numbered) != len(testLines) {
		t.Errorf("Expected %d numbered lines, got %d", len(testLines), len(numbered))
	}

	// Check first line has correct line number (51 = offset 50 + 1)
	if !strings.Contains(numbered[0], "51") {
		t.Errorf("Expected line number 51 in first line, got: %s", numbered[0])
	}

	// Check third line has correct line number (53 = offset 50 + 3)
	if !strings.Contains(numbered[2], "53") {
		t.Errorf("Expected line number 53 in third line, got: %s", numbered[2])
	}
}

// BenchmarkReadFileLines benchmarks the readFileLines function
func BenchmarkReadFileLines(b *testing.B) {
	// Create a temporary test file with 10,000 lines
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	var lines []string
	for i := 1; i <= 10000; i++ {
		lines = append(lines, "This is test line number "+string(rune('0'+(i%10))))
	}
	content := strings.Join(lines, "\n")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = readFileLines(testFile, 1000, 100)
	}
}

// BenchmarkRenderVisible benchmarks the windowed rendering
func BenchmarkRenderVisible(b *testing.B) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create test content with 1000 lines
	var lines []string
	for i := 1; i <= 1000; i++ {
		lines = append(lines, "This is line number "+string(rune('0'+(i%10)))+" with some additional content to make it realistic")
	}
	content := strings.Join(lines, "\n")
	lv.SetContent(content, "test.log")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = lv.renderVisible()
	}
}
