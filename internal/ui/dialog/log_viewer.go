package dialog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogContentLoadedMsg indicates that content has been lazily loaded
type LogContentLoadedMsg struct {
	Content   string
	StartLine int
	EndLine   int
	Error     error
}

// LogViewerErrorMsg indicates an error during lazy loading
type LogViewerErrorMsg struct {
	Error error
}

// LogViewerPanel displays log file contents with scrolling, word-wrap, and markdown rendering
type LogViewerPanel struct {
	viewport         viewport.Model
	content          string
	renderedContent  string
	filePath         string
	width            int
	height           int
	focused          bool
	showLineNumbers  bool
	wordWrap         bool
	markdownEnabled  bool   // Enable markdown rendering
	style            *DialogStyle
	scrollPos        string // Current scroll position indicator (e.g., "25%")
	loadError        error  // Stores any file loading errors
	isLoaded         bool   // Tracks if content has been loaded
	lineLimited      bool   // Indicates if content was truncated due to line limit
	fileSizeWarning  bool   // Indicates if file was large (>1MB)
	actualFileSize   int64  // Stores the actual file size in bytes

	// Performance optimization fields
	dirty            bool              // Dirty flag indicating content needs re-rendering
	renderCache      map[int]string    // Cache rendered lines: line number -> rendered content
	lineCacheMutex   sync.RWMutex      // Thread-safe access to render cache
	lastViewportSize struct{ width, height int } // Track last viewport size for cache invalidation
	contentLines     []string          // Pre-split content lines for efficient re-rendering
	scrollBuffer     int               // Number of lines to load beyond viewport (±buffer)
	
	// Lazy loading fields
	lazyLoadEnabled  bool              // Enable lazy loading for large files
	totalFileLines   int               // Total number of lines in the file (for lazy loading)
	visibleStartLine int               // Start line of currently loaded visible range
	visibleEndLine   int               // End line of currently loaded visible range
}

// renderCacheEntry holds a cached rendered line with metadata
type renderCacheEntry struct {
	lineNum  int
	content  string
	width    int
	markdown bool
	lineNums bool
	wordWrap bool
}

// NewLogViewerPanel creates a new log viewer instance
func NewLogViewerPanel(width, height int, style *DialogStyle) *LogViewerPanel {
	vp := viewport.New(width, height)
	vp.KeyMap = viewport.KeyMap{} // Disable default viewport keybindings to handle manually
	
	if style == nil {
		style = DefaultDialogStyle()
	}

	return &LogViewerPanel{
		viewport:        vp,
		content:         "",
		renderedContent: "",
		filePath:        "",
		width:           width,
		height:          height,
		focused:         false,
		showLineNumbers: false,
		wordWrap:        true, // Enable by default
		markdownEnabled: false, // Disabled by default
		style:           style,
		scrollPos:       "0%",
		loadError:       nil,
		isLoaded:        false,
		lineLimited:     false,
		fileSizeWarning: false,
		actualFileSize:  0,
		// Performance optimization initialization
		dirty:            true,
		renderCache:      make(map[int]string),
		contentLines:     []string{},
		scrollBuffer:     5, // Load 5 lines before/after visible viewport
		lastViewportSize: struct{ width, height int }{0, 0},
		// Lazy loading initialization
		lazyLoadEnabled:  false,
		totalFileLines:   0,
		visibleStartLine: 0,
		visibleEndLine:   0,
	}
}

// SetContent sets the log content to display
func (lv *LogViewerPanel) SetContent(content, filePath string) {
	lv.content = content
	lv.filePath = filePath
	lv.isLoaded = true
	lv.loadError = nil
	// Pre-split content for efficient rendering
	lv.contentLines = strings.Split(content, "\n")
	lv.invalidateCache()
	lv.renderContent()
	lv.updateScrollPosition()
}

// LoadFileContent loads content from a file with comprehensive error handling
// Returns an error if the file cannot be read
func (lv *LogViewerPanel) LoadFileContent(filePath string) error {
	// Reset state
	lv.filePath = filePath
	lv.content = ""
	lv.loadError = nil
	lv.isLoaded = false
	lv.lineLimited = false
	lv.fileSizeWarning = false
	lv.actualFileSize = 0
	lv.contentLines = []string{}
	lv.invalidateCache()
	
	// Check if path is empty
	if filePath == "" {
		lv.loadError = errors.New("file path is empty")
		lv.renderContent()
		return lv.loadError
	}
	
	// First check file existence and permissions with os.Stat
	info, err := os.Stat(filePath)
	if err != nil {
		// Handle specific filesystem errors with user-friendly messages
		if os.IsNotExist(err) {
			lv.loadError = fmt.Errorf("file not found: %s", filePath)
		} else if os.IsPermission(err) {
			lv.loadError = fmt.Errorf("permission denied reading file: %s", filePath)
		} else if errors.Is(err, syscall.ENOTDIR) {
			lv.loadError = fmt.Errorf("invalid path (not a directory): %s", filePath)
		} else {
			lv.loadError = fmt.Errorf("cannot access file: %w", err)
		}
		lv.renderContent()
		return lv.loadError
	}
	
	// Check if it's a directory
	if info.IsDir() {
		lv.loadError = fmt.Errorf("path is a directory, not a file: %s", filePath)
		lv.renderContent()
		return lv.loadError
	}
	
	// Store actual file size
	lv.actualFileSize = info.Size()
	
	// Check file size - warn if over 1MB
	const maxRecommendedSize = 1 * 1024 * 1024 // 1MB
	if info.Size() > maxRecommendedSize {
		lv.fileSizeWarning = true
	}
	
	// Try to read file with timeout logic
	type readResult struct {
		data []byte
		err  error
	}
	
	readChan := make(chan readResult, 1)
	go func() {
		data, err := os.ReadFile(filePath)
		readChan <- readResult{data, err}
	}()
	
	// Wait for read with timeout (useful for large files on slow systems)
	select {
	case result := <-readChan:
		if result.err != nil {
			// Handle read-time errors
			if os.IsPermission(result.err) {
				lv.loadError = fmt.Errorf("permission denied reading file: %s", filePath)
			} else if os.IsTimeout(result.err) {
				lv.loadError = fmt.Errorf("timeout reading file (file too large): %s", filePath)
			} else {
				lv.loadError = fmt.Errorf("failed to read file: %w", result.err)
			}
			lv.renderContent()
			return lv.loadError
		}
		
		// Convert to string and split into lines
		content := string(result.data)
		lines := strings.Split(content, "\n")
		
		// Apply 10,000 line limit for performance
		const maxLines = 10000
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lv.lineLimited = true
		}
		
		// Rejoin limited lines and pre-split for cache
		lv.content = strings.Join(lines, "\n")
		lv.contentLines = lines
		lv.isLoaded = true
		
	case <-time.After(10 * time.Second):
		// Timeout (file too large to read in reasonable time)
		lv.loadError = fmt.Errorf("timeout reading file - file may be too large or I/O is slow")
		lv.renderContent()
		return lv.loadError
	}
	
	// Render content and update scroll position
	lv.renderContent()
	lv.updateScrollPosition()
	
	return nil
}

// lazyLoadContent implements lazy loading for file contents
// Only loads the visible portion of the file plus a buffer
// Returns a tea.Cmd that loads content asynchronously
func (lv *LogViewerPanel) lazyLoadContent() tea.Cmd {
	if !lv.lazyLoadEnabled || lv.filePath == "" {
		return nil
	}
	
	// Calculate visible range based on viewport position
	startLine := max(0, lv.viewport.YOffset-lv.scrollBuffer)
	endLine := min(lv.totalFileLines, lv.viewport.YOffset+lv.viewport.Height+lv.scrollBuffer)
	
	// Check if we already have this range loaded
	if startLine >= lv.visibleStartLine && endLine <= lv.visibleEndLine {
		return nil // Already loaded
	}
	
	// Return command to load the visible lines
	filePath := lv.filePath
	return func() tea.Msg {
		visibleLines, err := readFileLines(filePath, startLine, endLine-startLine)
		if err != nil {
			return LogViewerErrorMsg{Error: err}
		}
		return LogContentLoadedMsg{
			Content:   strings.Join(visibleLines, "\n"),
			StartLine: startLine,
			EndLine:   endLine,
		}
	}
}

// readFileLines reads a specific range of lines from a file
// Efficiently reads only the requested lines without loading the entire file
func readFileLines(filePath string, startLine, numLines int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, numLines)
	currentLine := 0
	
	// Skip lines until we reach startLine
	for currentLine < startLine && scanner.Scan() {
		currentLine++
	}
	
	// Read the requested number of lines
	for currentLine < startLine+numLines && scanner.Scan() {
		lines = append(lines, scanner.Text())
		currentLine++
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	
	return lines, nil
}

// renderVisible implements windowed rendering
// Only renders the portion of content that's visible in the viewport
func (lv *LogViewerPanel) renderVisible() string {
	if len(lv.content) == 0 {
		return "Select a log file to preview"
	}
	
	lines := lv.contentLines
	if len(lines) == 0 {
		lines = strings.Split(lv.content, "\n")
	}
	
	// Calculate visible range
	startLine := max(0, lv.viewport.YOffset)
	endLine := min(len(lines), startLine+lv.viewport.Height)
	
	// Extract visible lines only
	var visibleLines []string
	if startLine < len(lines) {
		if endLine > len(lines) {
			endLine = len(lines)
		}
		visibleLines = lines[startLine:endLine]
	}
	
	// Apply markdown rendering if enabled
	var renderedLines []string
	if lv.markdownEnabled {
		content := strings.Join(visibleLines, "\n")
		content = lv.renderMarkdown(content)
		renderedLines = strings.Split(content, "\n")
	} else {
		renderedLines = visibleLines
	}
	
	// Apply word-wrap if enabled
	if lv.wordWrap {
		renderedLines = lv.applyWordWrap(renderedLines)
	}
	
	// Apply line numbers if enabled (use original line numbers, not visible indices)
	if lv.showLineNumbers {
		renderedLines = lv.applyLineNumbersWithOffset(renderedLines, startLine)
	}
	
	// Join lines efficiently
	var builder strings.Builder
	for i, line := range renderedLines {
		builder.WriteString(line)
		if i < len(renderedLines)-1 {
			builder.WriteRune('\n')
		}
	}
	
	return builder.String()
}

// applyLineNumbersWithOffset adds line numbers with a specific offset
// Used for windowed rendering to show correct line numbers
func (lv *LogViewerPanel) applyLineNumbersWithOffset(lines []string, offset int) []string {
	numbered := make([]string, len(lines))
	totalLines := len(lv.contentLines)
	if totalLines == 0 {
		totalLines = len(lines)
	}
	numWidth := len(fmt.Sprintf("%d", totalLines))
	
	for i, line := range lines {
		lineNum := offset + i + 1
		numbered[i] = fmt.Sprintf("%*d │ %s", numWidth, lineNum, line)
	}
	
	return numbered
}

// countFileLines counts the total number of lines in a file without loading it all
func countFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	
	return count, nil
}

// ClearContent clears the loaded content
func (lv *LogViewerPanel) ClearContent() {
	lv.content = ""
	lv.filePath = ""
	lv.loadError = nil
	lv.isLoaded = false
	lv.lineLimited = false
	lv.fileSizeWarning = false
	lv.actualFileSize = 0
	lv.contentLines = []string{}
	lv.invalidateCache()
	lv.renderContent()
	lv.updateScrollPosition()
}

// HasError returns true if there was an error loading the file
func (lv *LogViewerPanel) HasError() bool {
	return lv.loadError != nil
}

// GetError returns the file loading error, if any
func (lv *LogViewerPanel) GetError() error {
	return lv.loadError
}

// IsLoaded returns true if content has been loaded
func (lv *LogViewerPanel) IsLoaded() bool {
	return lv.isLoaded
}

// IsLineLimited returns true if content was truncated due to line limit
func (lv *LogViewerPanel) IsLineLimited() bool {
	return lv.lineLimited
}

// HasFileSizeWarning returns true if file was larger than 1MB
func (lv *LogViewerPanel) HasFileSizeWarning() bool {
	return lv.fileSizeWarning
}

// GetActualFileSize returns the actual file size in bytes
func (lv *LogViewerPanel) GetActualFileSize() int64 {
	return lv.actualFileSize
}

// IsPermissionError checks if the error is a permission denied error
func (lv *LogViewerPanel) IsPermissionError() bool {
	if lv.loadError == nil {
		return false
	}
	return os.IsPermission(lv.loadError) || strings.Contains(lv.loadError.Error(), "permission denied")
}

// IsFileNotFoundError checks if the error is a file not found error
func (lv *LogViewerPanel) IsFileNotFoundError() bool {
	if lv.loadError == nil {
		return false
	}
	return os.IsNotExist(lv.loadError) || strings.Contains(lv.loadError.Error(), "not found")
}

// IsTimeoutError checks if the error is a timeout error
func (lv *LogViewerPanel) IsTimeoutError() bool {
	if lv.loadError == nil {
		return false
	}
	return strings.Contains(lv.loadError.Error(), "timeout")
}

// SetFocused sets the focused state
func (lv *LogViewerPanel) SetFocused(focused bool) {
	lv.focused = focused
}

// IsFocused returns the focused state
func (lv *LogViewerPanel) IsFocused() bool {
	return lv.focused
}

// SetSize updates the viewport dimensions
// Note: Borders are now handled by the parent container (wrapPanel)
// The width and height passed here should be the available content area
func (lv *LogViewerPanel) SetSize(width, height int) {
	lv.width = width
	lv.height = height
	
	// Reserve space for footer (1 line)
	footerHeight := 1
	viewportHeight := height - footerHeight
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	
	// Set viewport dimensions
	// Width and height are already the available content area (borders handled externally)
	lv.viewport.Width = width
	lv.viewport.Height = viewportHeight
	
	// Check if viewport size changed (triggers cache invalidation due to word-wrap)
	if lv.lastViewportSize.width != width || lv.lastViewportSize.height != viewportHeight {
		lv.lastViewportSize.width = width
		lv.lastViewportSize.height = viewportHeight
		lv.invalidateCache() // Width changes affect word-wrap caching
	}
	
	lv.renderContent() // Re-render with new dimensions
	lv.updateScrollPosition()
}

// ToggleLineNumbers toggles line number display
func (lv *LogViewerPanel) ToggleLineNumbers() {
	lv.showLineNumbers = !lv.showLineNumbers
	lv.invalidateCache()
	lv.renderContent()
}

// ToggleWordWrap toggles word wrap
func (lv *LogViewerPanel) ToggleWordWrap() {
	lv.wordWrap = !lv.wordWrap
	lv.invalidateCache()
	lv.renderContent()
}

// ToggleMarkdown toggles markdown rendering
func (lv *LogViewerPanel) ToggleMarkdown() {
	lv.markdownEnabled = !lv.markdownEnabled
	lv.invalidateCache()
	lv.renderContent()
}

// Init initializes the model
func (lv *LogViewerPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages and keyboard input
func (lv *LogViewerPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !lv.focused {
		return lv, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return lv.handleKeyMsg(msg)
	case LogContentLoadedMsg:
		// Handle lazy-loaded content
		if msg.Error != nil {
			lv.loadError = msg.Error
			lv.renderContent()
			return lv, nil
		}
		lv.content = msg.Content
		lv.contentLines = strings.Split(msg.Content, "\n")
		lv.visibleStartLine = msg.StartLine
		lv.visibleEndLine = msg.EndLine
		lv.invalidateCache()
		lv.renderContent()
		return lv, nil
	case LogViewerErrorMsg:
		lv.loadError = msg.Error
		lv.renderContent()
		return lv, nil
	}
	
	return lv, nil
}

// handleKeyMsg handles keyboard input for scrolling and toggles
func (lv *LogViewerPanel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Intercept Tab/Shift+Tab to allow dialog focus cycling
	// These keys should NOT be handled by the log viewer
	switch msg.String() {
	case "tab", "shift+tab":
		// Don't handle these keys - let dialog handle focus cycling
		return lv, nil
	}
	
	switch msg.String() {
	// Line scrolling
	case "up", "k":
		lv.viewport.LineUp(1)
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	case "down", "j":
		lv.viewport.LineDown(1)
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	
	// Page scrolling
	case "pgup":
		lv.viewport.HalfPageUp()
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	case "pgdown":
		lv.viewport.HalfPageDown()
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	case "ctrl+u":
		lv.viewport.HalfPageUp()
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	case "ctrl+d":
		lv.viewport.HalfPageDown()
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	
	// Jump to top/bottom
	case "home", "g":
		// Handle "gg" for vim-style jump to top
		if msg.String() == "g" {
			// TODO: Implement "gg" sequence detection in parent model
			// For now, just handle single "g"
			lv.viewport.GotoTop()
			lv.updateScrollPosition()
		} else {
			lv.viewport.GotoTop()
			lv.updateScrollPosition()
		}
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	case "end":
		lv.viewport.GotoBottom()
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	case "G":
		// Vim-style jump to bottom
		lv.viewport.GotoBottom()
		lv.updateScrollPosition()
		cmd = lv.lazyLoadContent() // Trigger lazy load if needed
	
	// Toggle features
	case "n":
		lv.ToggleLineNumbers()
	case "w":
		lv.ToggleWordWrap()
	case "m":
		lv.ToggleMarkdown()
	}
	
	return lv, cmd
}

// View renders the log viewer content without borders
// Borders are added by the parent container (wrapPanel) for consistency
func (lv *LogViewerPanel) View() string {
	if lv.content == "" {
		return lv.renderEmptyState()
	}

	// Render the viewport content directly without borders
	// The parent container (wrapPanel) will add borders based on focus state
	content := lv.viewport.View()
	
	// Add status footer with scroll position and toggle states
	footer := lv.renderFooter()
	
	// Combine content and footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		footer,
	)
}

// renderEmptyState renders the empty state when no file is loaded
func (lv *LogViewerPanel) renderEmptyState() string {
	var message string
	var color lipgloss.Color
	var suggestions string
	
	if lv.loadError != nil {
		// Error state - provide context-aware messages
		color = lipgloss.Color("#F7768E") // Error red
		errorMsg := lv.loadError.Error()
		
		if lv.IsPermissionError() {
			message = "❌ Permission Denied"
			suggestions = "Check file permissions or try with different credentials"
		} else if lv.IsFileNotFoundError() {
			message = "❌ File Not Found"
			suggestions = "The file may have been deleted. Check the file path above."
		} else if lv.IsTimeoutError() {
			message = "⏱️  Timeout Reading File"
			suggestions = "File is too large or I/O is slow. Try opening in external editor."
		} else if strings.Contains(errorMsg, "directory") {
			message = "❌ Path is a Directory"
			suggestions = "Select a file, not a directory"
		} else if strings.Contains(errorMsg, "empty") {
			message = "⚠️  Empty Path"
			suggestions = "Select a file from the browser"
		} else {
			message = "❌ Error Loading File"
			suggestions = "Details: " + errorMsg
		}
		
		// Add full error details
		message = fmt.Sprintf("%s\n\n%s\n\n%s", message, errorMsg, suggestions)
		
	} else if !lv.isLoaded {
		// Not loaded yet
		message = "📄 Select a log file to preview"
		color = lipgloss.Color("#666666") // Gray
		suggestions = "Choose a file from the left panel to view its contents"
		message = fmt.Sprintf("%s\n\n%s", message, suggestions)
	} else {
		// Empty file
		message = "📋 File is empty"
		color = lipgloss.Color("#666666") // Gray
		suggestions = "This file contains no content"
		message = fmt.Sprintf("%s\n\n%s", message, suggestions)
	}
	
	emptyStyle := lipgloss.NewStyle().
		Foreground(color).
		Align(lipgloss.Center).
		Width(lv.width).
		Height(lv.height).
		PaddingTop(lv.height / 3)
	
	return emptyStyle.Render(message)
}

// renderFooter renders the status footer with scroll position and toggles
func (lv *LogViewerPanel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1)
	
	// Build toggle indicators
	toggles := []string{}
	if lv.showLineNumbers {
		toggles = append(toggles, "[n] Line Numbers: ON")
	} else {
		toggles = append(toggles, "[n] Line Numbers: OFF")
	}
	
	if lv.wordWrap {
		toggles = append(toggles, "[w] Word Wrap: ON")
	} else {
		toggles = append(toggles, "[w] Word Wrap: OFF")
	}
	
	if lv.markdownEnabled {
		toggles = append(toggles, "[m] Markdown: ON")
	} else {
		toggles = append(toggles, "[m] Markdown: OFF")
	}
	
	// Add warnings for content truncation and large files
	var warningText string
	warnings := []string{}
	
	if lv.lineLimited {
		warnings = append(warnings, "⚠️  Content limited to 10,000 lines")
	}
	if lv.fileSizeWarning {
		sizeStr := formatFileSizeForDisplay(lv.actualFileSize)
		warnings = append(warnings, fmt.Sprintf("⚠️  Large file (%s)", sizeStr))
	}
	
	if len(warnings) > 0 {
		warningText = " | " + strings.Join(warnings, " | ")
	}
	
	// Build footer text
	footerText := fmt.Sprintf("Scroll: %s | %s | ↑↓/k/j: scroll | PgUp/PgDn: page | Home/End: top/bottom%s",
		lv.scrollPos,
		strings.Join(toggles, " | "),
		warningText,
	)
	
	return footerStyle.Render(footerText)
}

// updateScrollPosition calculates the current scroll position as both line count and percentage
func (lv *LogViewerPanel) updateScrollPosition() {
	// Calculate scroll position
	totalLines := lv.viewport.TotalLineCount()
	currentLine := lv.viewport.YOffset
	
	if totalLines == 0 {
		lv.scrollPos = "0/0 lines (0%)"
		return
	}
	
	// Calculate percentage
	percentage := int(float64(currentLine) / float64(totalLines) * 100)
	if percentage > 100 {
		percentage = 100
	}
	
	// Format as "X/Y lines (Z%)"
	lv.scrollPos = fmt.Sprintf("%d/%d lines (%d%%)", currentLine, totalLines, percentage)
}

// renderContent renders the content with markdown, word-wrap, and line numbers
// Uses caching and dirty flags for efficient re-rendering
func (lv *LogViewerPanel) renderContent() {
	if lv.content == "" {
		lv.viewport.SetContent("")
		return
	}
	
	// Only re-render if dirty
	if !lv.dirty && lv.renderedContent != "" {
		return
	}
	
	content := lv.content
	
	// Apply markdown rendering if enabled
	if lv.markdownEnabled {
		content = lv.renderMarkdown(content)
	}
	
	// Use pre-split lines if available for efficiency
	var lines []string
	if len(lv.contentLines) > 0 && !lv.markdownEnabled {
		// Use pre-split lines when markdown is disabled
		lines = make([]string, len(lv.contentLines))
		copy(lines, lv.contentLines)
	} else {
		lines = strings.Split(content, "\n")
	}
	
	// Apply word-wrap if enabled
	if lv.wordWrap {
		lines = lv.applyWordWrap(lines)
	}
	
	// Apply line numbers if enabled
	if lv.showLineNumbers {
		lines = lv.applyLineNumbers(lines)
	}
	
	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	currentLine := lv.viewport.YOffset // Get the current visible line (top of viewport)
	
	// Apply current line highlighting if focused
	currentLineStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#44475a")).
		Foreground(lipgloss.Color("#f8f8f2"))
	
	for i, line := range lines {
		// Highlight the current line if panel is focused
		if lv.focused && i == currentLine {
			line = currentLineStyle.Render(line)
		}
		builder.WriteString(line)
		if i < len(lines)-1 {
			builder.WriteRune('\n')
		}
	}
	
	lv.renderedContent = builder.String()
	lv.viewport.SetContent(lv.renderedContent)
	lv.dirty = false
}

// applyWordWrap applies word-wrapping to lines that exceed the viewport width
// Uses caching per width to avoid redundant wrapping operations
func (lv *LogViewerPanel) applyWordWrap(lines []string) []string {
	// Calculate available width (subtract padding and borders)
	availableWidth := lv.width - 4
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}
	
	wrapped := make([]string, 0, len(lines)*2) // Pre-allocate with expected growth
	
	for _, line := range lines {
		if len(line) <= availableWidth {
			wrapped = append(wrapped, line)
			continue
		}
		
		// Wrap long line using optimized wrapLine function
		wrappedLines := wrapLineOptimized(line, availableWidth)
		wrapped = append(wrapped, wrappedLines...)
	}
	
	return wrapped
}

// wrapLineOptimized wraps a single line to the specified width
// Optimized for performance with rune-based operations for proper Unicode support
// Preserves leading whitespace and adds continuation indicators
func wrapLineOptimized(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	
	// Convert to runes for proper Unicode handling
	runes := []rune(line)
	lineLen := len(runes)
	
	if lineLen <= width {
		return []string{line}
	}
	
	// Count leading whitespace
	leadingSpace := 0
	for _, r := range runes {
		if r == ' ' || r == '\t' {
			leadingSpace++
		} else {
			break
		}
	}
	
	wrapped := make([]string, 0, (lineLen/width)+1) // Pre-allocate with expected segments
	start := 0
	
	for start < lineLen {
		// Calculate segment end position
		end := start + width
		if end >= lineLen {
			// Last segment - take everything remaining
			wrapped = append(wrapped, string(runes[start:]))
			break
		}
		
		// Find word boundary for breaking
		breakPoint := end
		foundSpace := false
		
		// Search backwards from end position for a space
		for i := end; i > start; i-- {
			if runes[i] == ' ' || runes[i] == '\t' {
				breakPoint = i
				foundSpace = true
				break
			}
		}
		
		// If no space found and we're not at the start, force break
		if !foundSpace && end < lineLen {
			// Leave room for continuation indicator (>)
			breakPoint = end - 1
			if breakPoint <= start {
				breakPoint = start + 1
			}
			
			// Add segment with continuation indicator
			segment := string(runes[start:breakPoint]) + ">"
			wrapped = append(wrapped, segment)
			start = breakPoint
		} else {
			// Break at word boundary
			segment := string(runes[start:breakPoint])
			wrapped = append(wrapped, segment)
			
			// Skip whitespace at break point
			start = breakPoint
			for start < lineLen && (runes[start] == ' ' || runes[start] == '\t') {
				start++
			}
			
			// Add leading whitespace to continuation lines
			if start < lineLen && leadingSpace > 0 {
				indent := string(runes[:leadingSpace])
				remaining := string(runes[start:])
				runes = []rune(indent + remaining)
				lineLen = len(runes)
				start = 0
			}
		}
	}
	
	return wrapped
}

// wrapLine wraps a single line to the specified width
// Preserves leading whitespace on first line and adds continuation indent to wrapped lines
// Uses rune-based operations for proper Unicode support
// Deprecated: Use wrapLineOptimized instead for better performance
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	
	// Convert to runes for proper Unicode handling
	runes := []rune(line)
	lineLen := len(runes)
	
	if lineLen <= width {
		return []string{line}
	}
	
	// Count leading whitespace
	leadingSpace := 0
	for _, r := range runes {
		if r == ' ' || r == '\t' {
			leadingSpace++
		} else {
			break
		}
	}
	
	wrapped := []string{}
	start := 0
	
	for start < lineLen {
		// Calculate segment end position
		end := start + width
		if end >= lineLen {
			// Last segment - take everything remaining
			wrapped = append(wrapped, string(runes[start:]))
			break
		}
		
		// Find word boundary for breaking
		breakPoint := end
		foundSpace := false
		
		// Search backwards from end position for a space
		for i := end; i > start; i-- {
			if runes[i] == ' ' || runes[i] == '\t' {
				breakPoint = i
				foundSpace = true
				break
			}
		}
		
		// If no space found and we're not at the start, force break
		if !foundSpace && end < lineLen {
			// Leave room for continuation indicator (>)
			breakPoint = end - 1
			if breakPoint <= start {
				breakPoint = start + 1
			}
			
			// Add segment with continuation indicator
			segment := string(runes[start:breakPoint]) + ">"
			wrapped = append(wrapped, segment)
			start = breakPoint
		} else {
			// Break at word boundary
			segment := string(runes[start:breakPoint])
			wrapped = append(wrapped, segment)
			
			// Skip whitespace at break point
			start = breakPoint
			for start < lineLen && (runes[start] == ' ' || runes[start] == '\t') {
				start++
			}
			
			// Add leading whitespace to continuation lines
			if start < lineLen && leadingSpace > 0 {
				indent := string(runes[:leadingSpace])
				remaining := string(runes[start:])
				runes = []rune(indent + remaining)
				lineLen = len(runes)
				start = 0
			}
		}
	}
	
	return wrapped
}

// applyLineNumbers adds line numbers to each line
func (lv *LogViewerPanel) applyLineNumbers(lines []string) []string {
	numbered := make([]string, len(lines))
	numWidth := len(fmt.Sprintf("%d", len(lines)))
	
	for i, line := range lines {
		numbered[i] = fmt.Sprintf("%*d │ %s", numWidth, i+1, line)
	}
	
	return numbered
}

// GetFilePath returns the current file path
func (lv *LogViewerPanel) GetFilePath() string {
	return lv.filePath
}

// GetContent returns the raw content
func (lv *LogViewerPanel) GetContent() string {
	return lv.content
}

// invalidateCache clears the render cache and marks content as dirty
func (lv *LogViewerPanel) invalidateCache() {
	lv.lineCacheMutex.Lock()
	defer lv.lineCacheMutex.Unlock()
	lv.renderCache = make(map[int]string)
	lv.dirty = true
}

// getCachedLine retrieves a cached rendered line if available
func (lv *LogViewerPanel) getCachedLine(lineNum int) (string, bool) {
	lv.lineCacheMutex.RLock()
	defer lv.lineCacheMutex.RUnlock()
	content, exists := lv.renderCache[lineNum]
	return content, exists
}

// setCachedLine stores a rendered line in the cache
func (lv *LogViewerPanel) setCachedLine(lineNum int, content string) {
	lv.lineCacheMutex.Lock()
	defer lv.lineCacheMutex.Unlock()
	lv.renderCache[lineNum] = content
}

// getCacheSize returns the number of cached entries
func (lv *LogViewerPanel) getCacheSize() int {
	lv.lineCacheMutex.RLock()
	defer lv.lineCacheMutex.RUnlock()
	return len(lv.renderCache)
}

// formatFileSizeForDisplay formats a file size in bytes to a human-readable string
func formatFileSizeForDisplay(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// renderMarkdown renders markdown syntax with Lipgloss styling
func (lv *LogViewerPanel) renderMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	inCodeBlock := false
	codeBlockLang := ""
	
	// Define styles for markdown elements
	h1Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff79c6"))
	h2Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bd93f9"))
	h3Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8be9fd"))
	h4Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50fa7b"))
	h5Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f1fa8c"))
	h6Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffb86c"))
	codeBlockStyle := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#f8f8f2"))
	codeBlockMarkerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#75b5aa"))
	
	for _, line := range lines {
		// Check for code block markers
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				// Extract language if provided
				codeBlockLang = strings.TrimPrefix(line, "```")
				codeBlockLang = strings.TrimSpace(codeBlockLang)
			} else {
				codeBlockLang = ""
			}
			result.WriteString(codeBlockMarkerStyle.Render(line))
			result.WriteString("\n")
			continue
		}
		
		if inCodeBlock {
			// Render code block content with background
			result.WriteString(codeBlockStyle.Render(line))
			result.WriteString("\n")
			continue
		}
		
		// Handle headers
		if strings.HasPrefix(line, "# ") {
			result.WriteString(h1Style.Render(line))
		} else if strings.HasPrefix(line, "## ") {
			result.WriteString(h2Style.Render(line))
		} else if strings.HasPrefix(line, "### ") {
			result.WriteString(h3Style.Render(line))
		} else if strings.HasPrefix(line, "#### ") {
			result.WriteString(h4Style.Render(line))
		} else if strings.HasPrefix(line, "##### ") {
			result.WriteString(h5Style.Render(line))
		} else if strings.HasPrefix(line, "###### ") {
			result.WriteString(h6Style.Render(line))
		} else if strings.HasPrefix(strings.TrimSpace(line), "- ") || strings.HasPrefix(strings.TrimSpace(line), "* ") {
			// Handle list items - preserve indentation
			result.WriteString(lv.renderInlineFormatting(line))
		} else {
			// Regular text with inline formatting
			result.WriteString(lv.renderInlineFormatting(line))
		}
		
		result.WriteString("\n")
	}
	
	// Remove trailing newline if present
	resultStr := result.String()
	if len(resultStr) > 0 && resultStr[len(resultStr)-1] == '\n' {
		resultStr = resultStr[:len(resultStr)-1]
	}
	
	return resultStr
}

// renderInlineFormatting handles inline markdown formatting (bold, italic, code)
func (lv *LogViewerPanel) renderInlineFormatting(line string) string {
	// Define styles
	boldStyle := lipgloss.NewStyle().Bold(true)
	italicStyle := lipgloss.NewStyle().Italic(true)
	inlineCodeStyle := lipgloss.NewStyle().Background(lipgloss.Color("#2d2d2d")).Foreground(lipgloss.Color("#f8f8f2"))
	
	var result strings.Builder
	i := 0
	runes := []rune(line)
	
	for i < len(runes) {
		// Check for inline code (`code`)
		if runes[i] == '`' {
			// Find closing backtick
			j := i + 1
			for j < len(runes) && runes[j] != '`' {
				j++
			}
			
			if j < len(runes) {
				// Found closing backtick - remove backticks for rendering
				codeContent := string(runes[i+1 : j])
				result.WriteString(inlineCodeStyle.Render(codeContent))
				i = j + 1
				continue
			}
		}
		
		// Check for bold (**text**)
		if i+1 < len(runes) && runes[i] == '*' && runes[i+1] == '*' {
			// Find closing **
			j := i + 2
			for j+1 < len(runes) && !(runes[j] == '*' && runes[j+1] == '*') {
				j++
			}
			
			if j+1 < len(runes) && runes[j] == '*' && runes[j+1] == '*' {
				// Found closing **
				boldContent := string(runes[i+2 : j])
				result.WriteString(boldStyle.Render(boldContent))
				i = j + 2
				continue
			}
		}
		
		// Check for italic (*text*)
		if runes[i] == '*' {
			// Find closing *
			j := i + 1
			for j < len(runes) && runes[j] != '*' {
				j++
			}
			
			if j < len(runes) {
				// Found closing *
				italicContent := string(runes[i+1 : j])
				result.WriteString(italicStyle.Render(italicContent))
				i = j + 1
				continue
			}
		}
		
		// Regular character
		result.WriteRune(runes[i])
		i++
	}
	
	return result.String()
}
