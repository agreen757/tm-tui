# Log Panel Word Wrap Implementation Plan

## Overview
Implement word wrap functionality for the Log panel to match the behavior of the Task Details panel, ensuring long log lines wrap properly within the panel width instead of being truncated or causing horizontal overflow.

## Current State Analysis

### Task Details Panel (Working Reference)
**File:** `internal/ui/app.go`

**How it works:**
1. Calculates available width: `wrapWidth := m.detailsViewport.Width - 4`
2. Uses `wrapText()` function to wrap content
3. Handles label indentation for wrapped continuation lines
4. Updates viewport with wrapped content

**Key function:**
```go
func wrapText(text string, width int) string
```
- Takes text and width as input
- Splits text into words
- Builds lines that don't exceed width
- Returns wrapped text with newlines

### Log Panel (Current State)
**File:** `internal/ui/app.go`

**Current behavior:**
- Lines are stored in `m.logLines []string`
- Rendered by joining lines with newlines: `strings.Join(m.logLines, "\n")`
- **No word wrapping applied**
- Long lines overflow or get truncated by viewport

**Key functions:**
- `addLogLine(line string)` - Adds raw line to array
- `renderLog()` - Joins lines without processing
- `updateLogViewport()` - Sets content and auto-scrolls to bottom

## Problem Statement

**Issues with current implementation:**
1. Long log lines (e.g., "Generated prompt (first 200 chars): ...") overflow the panel width
2. Error messages with long paths or stack traces are hard to read
3. Debug output gets visually truncated
4. No visual consistency with Task Details panel behavior

**Example problematic log lines:**
- `"Generated prompt (first 200 chars): Implement authentication middleware with JWT validation, session management, and role-based access control..."`
- `"Error: failed to parse PRD file at /Users/username/projects/long-project-name/docs/requirements.txt: unexpected token"`
- `"DEBUG: Task ID=5.3.2, Title='Implement comprehensive error handling for API gateway with retry logic', Deps=[5.3.1, 5.2.4]"`

## Solution Design

### Approach 1: Wrap on Render (Recommended)
**Wrap log lines dynamically when rendering, preserving original lines in storage**

**Advantages:**
- Original log lines preserved for export/copy
- Adapts to terminal resize automatically
- Consistent with Task Details panel approach
- No storage overhead

**Disadvantages:**
- Slight rendering overhead (negligible for typical log sizes)

### Approach 2: Wrap on Add
**Wrap log lines when they're added to the log**

**Advantages:**
- No rendering overhead
- Simple implementation

**Disadvantages:**
- Original lines lost (harder to export/search)
- Doesn't adapt to terminal resize
- Pre-wrapped lines look wrong if panel size changes

**Decision: Use Approach 1 (Wrap on Render)**

## Implementation Plan

### Step 1: Modify `renderLog()` Function
**Location:** `internal/ui/app.go` (around line 1545)

**Changes:**
```go
// renderLog renders the log panel content with word wrapping
func (m Model) renderLog() string {
	if len(m.logLines) == 0 {
		return m.styles.Info.Render("No log output yet")
	}

	// Calculate available width for wrapping text
	// Subtract padding to ensure text doesn't hit the edge
	wrapWidth := m.logViewport.Width - 4
	if wrapWidth < 20 {
		wrapWidth = 20 // Minimum reasonable width
	}

	// Wrap each log line individually
	var wrappedLines []string
	for _, line := range m.logLines {
		wrappedLine := wrapText(line, wrapWidth)
		wrappedLines = append(wrappedLines, wrappedLine)
	}

	return strings.Join(wrappedLines, "\n")
}
```

**Rationale:**
- Reuses existing `wrapText()` function (DRY principle)
- Calculates wrap width based on viewport dimensions
- Wraps each line independently to preserve line boundaries
- Maintains minimum width for readability

### Step 2: Update Log Viewport Dimensions Handling
**Location:** `internal/ui/layout.go` (around line 298)

**Current code:**
```go
// Update log viewport
if m.showLogPanel {
	m.logViewport.Width = layout.LogWidth - panelPadding*2
	m.logViewport.Height = layout.LogHeight - panelPadding
}
```

**Ensure this is called properly on:**
- Initial layout calculation
- Window resize events
- Panel visibility toggle

**No changes needed** - already handles viewport width updates correctly.

### Step 3: Update `updateLogViewport()` to Trigger Re-wrap
**Location:** `internal/ui/app.go` (around line 1554)

**Current code:**
```go
func (m *Model) updateLogViewport() {
	content := m.renderLog()
	m.logViewport.SetContent(content)
	// Auto-scroll to bottom
	m.logViewport.GotoBottom()
}
```

**Changes needed:** None - `renderLog()` will automatically wrap based on current viewport width.

**Ensure this is called when:**
- New log lines are added (`addLogLine()`)
- Window is resized (already handled in `Update()`)
- Log panel is shown/hidden

### Step 4: Handle Terminal Resize Events
**Location:** `internal/ui/app.go` in `Update()` function

**Find where window resize is handled:**
```go
case tea.WindowSizeMsg:
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true
	
	// Update viewport sizes based on layout
	m.updateViewportSizes()
```

**Add log viewport update:**
```go
case tea.WindowSizeMsg:
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true
	
	// Update viewport sizes based on layout
	m.updateViewportSizes()
	
	// Re-render log content with new width
	if m.showLogPanel {
		m.updateLogViewport()
	}
```

**Rationale:**
- Ensures log content re-wraps when terminal is resized
- Only updates if log panel is visible (performance)
- Maintains auto-scroll to bottom behavior

### Step 5: Add Unit Tests
**New file:** `internal/ui/log_wrap_test.go`

**Test cases:**
1. **Test short lines remain unchanged**
   - Input: "Task completed"
   - Width: 80
   - Expected: "Task completed"

2. **Test long lines wrap correctly**
   - Input: "This is a very long log line that should wrap to multiple lines when the width is constrained"
   - Width: 40
   - Expected: Multiple lines, each ≤ 40 chars

3. **Test multiple log lines wrap independently**
   - Input: ["Short line", "Very long line that needs wrapping", "Another short"]
   - Verify each line wraps independently

4. **Test minimum width constraint**
   - Input: Long text
   - Width: 5
   - Expected: Wrapped at minimum width (20)

5. **Test empty log lines**
   - Input: []
   - Expected: "No log output yet"

6. **Test viewport width changes**
   - Add lines, change width, verify re-wrapping

**Example test structure:**
```go
func TestRenderLogWithWordWrap(t *testing.T) {
	tests := []struct {
		name        string
		logLines    []string
		viewportWidth int
		wantLines   int // Number of output lines expected
	}{
		{
			name:        "short line no wrap",
			logLines:    []string{"Task completed"},
			viewportWidth: 80,
			wantLines:   1,
		},
		{
			name:        "long line wraps",
			logLines:    []string{"This is a very long log line that should definitely wrap"},
			viewportWidth: 30,
			wantLines:   2, // Should wrap to 2 lines
		},
		// Add more test cases
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model with test data
			m := Model{
				logLines: tt.logLines,
				logViewport: viewport.Model{Width: tt.viewportWidth},
			}
			
			// Render log
			result := m.renderLog()
			
			// Count lines
			lines := strings.Split(result, "\n")
			if len(lines) != tt.wantLines {
				t.Errorf("got %d lines, want %d", len(lines), tt.wantLines)
			}
			
			// Verify no line exceeds width
			wrapWidth := tt.viewportWidth - 4
			for i, line := range lines {
				if len(line) > wrapWidth {
					t.Errorf("line %d exceeds width: %d > %d", i, len(line), wrapWidth)
				}
			}
		})
	}
}
```

### Step 6: Manual Testing Checklist

**Test scenarios:**
1. ✓ Add short log lines - verify they display normally
2. ✓ Add long log lines (>100 chars) - verify they wrap
3. ✓ Resize terminal smaller - verify logs re-wrap
4. ✓ Resize terminal larger - verify logs expand
5. ✓ Toggle log panel visibility - verify no crashes
6. ✓ Add many log lines (100+) - verify performance
7. ✓ Test with very narrow terminal (40 cols) - verify minimum width
8. ✓ Test with ANSI color codes in logs - verify wrapping preserves formatting
9. ✓ Auto-scroll to bottom still works after adding wrapped lines
10. ✓ Compare visual appearance with Task Details panel wrapping

**Test commands to generate long log lines:**
```go
// In TUI, trigger actions that generate long logs:
- Run task with Crush (generates long prompt preview)
- Trigger errors with long file paths
- Add debug output with long task titles
```

## Edge Cases to Handle

### 1. ANSI Color Codes
**Issue:** Log lines may contain ANSI color codes (e.g., styled error messages)

**Solution:**
- The existing `wrapText()` function doesn't account for ANSI codes
- Consider using `lipgloss.Width()` for accurate character counting
- Or strip ANSI codes before wrapping and re-apply after

**Recommended approach:**
```go
// Use lipgloss/x/ansi for accurate width calculation
import "github.com/charmbracelet/x/ansi"

func wrapTextWithANSI(text string, width int) string {
	// Strip ANSI codes for wrapping logic
	plainText := ansi.Strip(text)
	wrapped := wrapText(plainText, width)
	return wrapped // TODO: Re-apply ANSI if needed
}
```

**Decision:** Start with simple implementation, add ANSI support if needed.

### 2. Empty Lines
**Issue:** Empty log lines should be preserved for visual separation

**Solution:**
- Check for empty strings before wrapping
- Preserve empty lines in output

```go
for _, line := range m.logLines {
	if line == "" {
		wrappedLines = append(wrappedLines, "")
		continue
	}
	wrappedLine := wrapText(line, wrapWidth)
	wrappedLines = append(wrappedLines, wrappedLine)
}
```

### 3. Very Long Words
**Issue:** Single words longer than wrap width (e.g., long URLs or file paths)

**Current `wrapText()` behavior:** 
- Word will appear on its own line and exceed width

**Solution options:**
1. Hard-break long words at width boundary
2. Allow long words to overflow (current behavior)

**Decision:** Keep current behavior (allow overflow) for now - preserves readability of paths/URLs.

### 4. Log Performance with Many Lines
**Issue:** With 1000+ log lines, re-wrapping on every resize could be slow

**Solution:**
- Test with large log files first
- If performance is an issue, implement:
  - Lazy wrapping (only wrap visible lines)
  - Caching of wrapped results
  - Line limit (e.g., keep last 500 lines)

**Decision:** Implement simple version first, optimize if needed.

## Files to Modify

### Primary Changes
1. **`internal/ui/app.go`**
   - Modify `renderLog()` to add word wrapping
   - Update `Update()` to re-render logs on window resize
   - Lines: ~1545, ~1730

### Testing
2. **`internal/ui/log_wrap_test.go`** (new file)
   - Unit tests for log wrapping functionality

### Documentation
3. **`README.md`**
   - Update to mention log panel word wrap feature (if doing release notes)

## Implementation Steps Checklist

- [ ] 1. Modify `renderLog()` to wrap log lines based on viewport width
- [ ] 2. Add log viewport update on window resize in `Update()` function
- [ ] 3. Test with short log lines (no change expected)
- [ ] 4. Test with long log lines (wrapping expected)
- [ ] 5. Test terminal resize behavior
- [ ] 6. Add unit tests in `log_wrap_test.go`
- [ ] 7. Run all existing tests to ensure no regression
- [ ] 8. Manual testing with all edge cases
- [ ] 9. Update documentation if needed
- [ ] 10. Commit and create version tag

## Success Criteria

**Feature is complete when:**
1. Long log lines wrap to fit within log panel width
2. Wrapping adapts dynamically to terminal resize
3. Visual appearance matches Task Details panel quality
4. Auto-scroll to bottom continues to work
5. No performance degradation with typical log sizes (100-500 lines)
6. All unit tests pass
7. Manual testing scenarios all pass

## Estimated Effort

**Development:** 1-2 hours
- Modify `renderLog()`: 30 minutes
- Add resize handling: 15 minutes
- Edge case handling: 30 minutes
- Testing and refinement: 30-45 minutes

**Testing:** 30-45 minutes
- Unit tests: 20 minutes
- Manual testing: 15-25 minutes

**Total:** 1.5-3 hours

## Future Enhancements (Out of Scope)

1. **Log export feature** - Export log to file with original unwrapped lines
2. **Log search/filter** - Search within log content
3. **Log timestamps** - Add timestamps to each log entry
4. **Log levels** - Color-code by severity (INFO, WARN, ERROR)
5. **Log line numbers** - Add line numbers for reference
6. **Persistent logs** - Save logs across TUI sessions
7. **ANSI color preservation** - Preserve styling through wrapping

## References

**Similar implementations to study:**
- Task Details panel: `internal/ui/app.go:1286-1406`
- wrapText function: `internal/ui/app.go:1256-1284`
- Viewport handling: `internal/ui/layout.go:298-302`

**Related issues/PRs:**
- None currently

---

*Plan created: December 2024*
*Target version: v0.1.11*
