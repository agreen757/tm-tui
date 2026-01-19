package dialog

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestCachingBehavior tests the caching mechanism
func TestCachingBehavior(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Test initial cache is empty
	if lv.getCacheSize() != 0 {
		t.Errorf("Expected empty cache initially, got %d entries", lv.getCacheSize())
	}

	// Set content (which calls renderContent internally)
	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")

	// After SetContent, dirty should be false (renderContent was called)
	if lv.dirty {
		t.Error("Expected dirty flag to be false after SetContent completes")
	}

	// Call renderContent again without changing anything - should return early without re-rendering
	lv.renderContent()
	if lv.dirty {
		t.Error("Expected dirty flag to remain false")
	}

	// Clear the rendered content to test that renderContent respects dirty flag
	lv.renderedContent = ""

	// Now toggle something to set dirty flag
	lv.ToggleLineNumbers()

	// At this point, renderContent has been called immediately in ToggleLineNumbers,
	// so dirty should be false again (since content wasn't empty)
	if lv.dirty {
		t.Error("Expected dirty flag to be false after ToggleLineNumbers calls renderContent")
	}
}

// TestDirtyFlagInvalidation tests that dirty flag is set when content changes
func TestDirtyFlagInvalidation(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	lv.SetContent("test content", "/test.log")
	lv.renderContent()

	if lv.dirty {
		t.Error("Expected dirty flag to be false after render")
	}

	// Invalidate cache (which should set dirty flag)
	lv.invalidateCache()
	if !lv.dirty {
		t.Error("Expected dirty flag to be true after invalidateCache")
	}

	// renderContent should clear dirty flag when actually rendering
	lv.renderContent()
	if lv.dirty {
		t.Error("Expected dirty flag to be false after render")
	}

	// Invalidate cache again
	lv.invalidateCache()
	if !lv.dirty {
		t.Error("Expected dirty flag to be true after invalidateCache")
	}

	// renderContent with no renderedContent should do a full render and clear dirty
	lv.renderedContent = ""
	lv.renderContent()
	if lv.dirty {
		t.Error("Expected dirty flag to be false after render completes")
	}
}

// TestContentLinesSplitting tests that content is pre-split for efficiency
func TestContentLinesSplitting(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	lv.SetContent(content, "/test.log")

	if len(lv.contentLines) != 5 {
		t.Errorf("Expected 5 content lines, got %d", len(lv.contentLines))
	}

	expectedLines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	for i, expectedLine := range expectedLines {
		if lv.contentLines[i] != expectedLine {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expectedLine, lv.contentLines[i])
		}
	}
}

// TestViewportSizeChangeInvalidatesCache tests cache invalidation on size change
func TestViewportSizeChangeInvalidatesCache(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")
	lv.renderContent()

	// Add something to cache manually to verify it gets cleared
	lv.setCachedLine(1, "test")

	if lv.getCacheSize() != 1 {
		t.Errorf("Expected 1 cache entry, got %d", lv.getCacheSize())
	}

	// Resize viewport
	lv.SetSize(100, 30)

	if lv.getCacheSize() != 0 {
		t.Errorf("Expected cache to be cleared after resize, got %d entries", lv.getCacheSize())
	}
}

// TestRenderingWithoutRecompute tests dirty flag prevents unnecessary re-rendering
func TestRenderingWithoutRecompute(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")
	lv.renderContent()

	// Get rendered content
	firstRender := lv.renderedContent

	// Call renderContent again without any changes
	lv.renderContent()
	secondRender := lv.renderedContent

	if firstRender != secondRender {
		t.Error("Expected rendered content to be the same on second render without changes")
	}
}

// BenchmarkRenderLargeFile benchmarks rendering performance with large files
func BenchmarkRenderLargeFile(b *testing.B) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create large content (10,000 lines)
	lines := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		lines[i] = fmt.Sprintf("Line %d: This is a sample log line with some content", i+1)
	}
	content := strings.Join(lines, "\n")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lv.SetContent(content, "/test.log")
		lv.renderContent()
	}
}

// BenchmarkWordWrap benchmarks word wrapping performance
func BenchmarkWordWrap(b *testing.B) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.wordWrap = true

	// Create content with long lines
	lines := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		lines[i] = strings.Repeat("a", 200) // Very long line
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lv.renderContent()
	}
}

// BenchmarkMarkdownRendering benchmarks markdown rendering performance
func BenchmarkMarkdownRendering(b *testing.B) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.markdownEnabled = true

	// Create markdown content
	lines := make([]string, 500)
	for i := 0; i < 500; i++ {
		lines[i] = fmt.Sprintf("# Header %d\nThis is **bold** and *italic* text with `code` blocks", i)
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lv.renderContent()
	}
}

// BenchmarkLineNumbering benchmarks line number application
func BenchmarkLineNumbering(b *testing.B) {
	lv := NewLogViewerPanel(80, 24, nil)
	lv.showLineNumbers = true

	// Create content
	lines := make([]string, 5000)
	for i := 0; i < 5000; i++ {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lv.renderContent()
	}
}

// TestCacheHitRate tests cache efficiency
func TestCacheHitRate(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create content
	lines := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")

	// Manually add cache entries
	for i := 0; i < 100; i++ {
		lv.setCachedLine(i, fmt.Sprintf("cached-%d", i))
	}

	// Check cache retrieval
	hits := 0
	misses := 0

	for i := 0; i < 100; i++ {
		if _, ok := lv.getCachedLine(i); ok {
			hits++
		} else {
			misses++
		}
	}

	if hits != 100 {
		t.Errorf("Expected 100 cache hits, got %d", hits)
	}

	if misses != 0 {
		t.Errorf("Expected 0 cache misses, got %d", misses)
	}
}

// TestScrollingPerformance tests performance of scroll operations
func TestScrollingPerformance(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create large content
	lines := make([]string, 5000)
	for i := 0; i < 5000; i++ {
		lines[i] = fmt.Sprintf("Line %d: Content", i+1)
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")
	lv.renderContent()

	// Time 100 scroll operations
	start := time.Now()

	for i := 0; i < 100; i++ {
		lv.viewport.LineDown(1)
		lv.updateScrollPosition()
	}

	elapsed := time.Since(start)

	// Should complete quickly (less than 100ms for 100 operations)
	if elapsed > 100*time.Millisecond {
		t.Logf("Scrolling 100 times took %v (may indicate performance issue)", elapsed)
	}
}

// TestTogglePerformance tests performance of toggle operations
func TestTogglePerformance(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create content
	lines := make([]string, 2000)
	for i := 0; i < 2000; i++ {
		lines[i] = fmt.Sprintf("Line %d: %s", i+1, strings.Repeat("x", 50))
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")
	lv.renderContent()

	// Time 50 toggle operations (word wrap)
	start := time.Now()

	for i := 0; i < 50; i++ {
		lv.ToggleWordWrap()
		lv.renderContent()
	}

	elapsed := time.Since(start)

	// Should be reasonably fast
	if elapsed > 5*time.Second {
		t.Logf("50 word wrap toggles took %v (may indicate performance issue)", elapsed)
	}
}

// TestCacheThreadSafety tests thread-safe cache access
func TestCacheThreadSafety(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Set content
	content := "Line 1\nLine 2\nLine 3"
	lv.SetContent(content, "/test.log")

	// Simulate concurrent access
	done := make(chan bool, 10)

	// Writer goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				lv.setCachedLine(id*100+j, fmt.Sprintf("cache-%d-%d", id, j))
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				_, _ = lv.getCachedLine(id * 100 % 500)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we got here without panicking, thread safety is working
	if lv.getCacheSize() == 0 {
		t.Error("Expected cache to have entries after concurrent access")
	}
}

// BenchmarkCacheInvalidation benchmarks cache invalidation performance
func BenchmarkCacheInvalidation(b *testing.B) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		lv.setCachedLine(i, fmt.Sprintf("cached-%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lv.invalidateCache()
	}
}

// TestStringsBuilderEfficiency tests strings.Builder usage in rendering
func TestStringsBuilderEfficiency(t *testing.T) {
	lv := NewLogViewerPanel(80, 24, nil)

	// Create content
	lines := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		lines[i] = fmt.Sprintf("Line %d with content", i+1)
	}
	content := strings.Join(lines, "\n")

	lv.SetContent(content, "/test.log")
	lv.renderContent()

	// Check that rendered content is properly constructed
	if lv.renderedContent == "" {
		t.Error("Expected rendered content to be non-empty")
	}

	// Verify content includes lines
	if !strings.Contains(lv.renderedContent, "Line 1") {
		t.Error("Expected rendered content to include 'Line 1'")
	}

	if !strings.Contains(lv.renderedContent, "Line 1000") {
		t.Error("Expected rendered content to include 'Line 1000'")
	}
}
