package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReadyTaskItem represents a parsed ready task with selection state
type ReadyTaskItem struct {
	ID             string
	TaskTitle      string
	Status         string
	Priority       string
	Dependencies   []string
	Blocks         []string
	Complexity     int
	Selected       bool
	TitleTruncated bool // True if title ends with '...' indicating truncation
}

// TaskDetails represents the complete details of a task parsed from task-master show
type TaskDetails struct {
	ID           string
	Title        string
	Status       string
	Priority     string
	Dependencies []string
	Blocks       []string
	Complexity   int
}

// Title returns the title of the item (implements ListItem interface)
func (r ReadyTaskItem) Title() string {
	return fmt.Sprintf("[%s] %s (%s)", r.ID, r.TaskTitle, r.Priority)
}

// Description returns the description of the item (implements ListItem interface)
func (r ReadyTaskItem) Description() string {
	parts := []string{}
	if len(r.Dependencies) > 0 {
		parts = append(parts, fmt.Sprintf("deps: %s", strings.Join(r.Dependencies, ", ")))
	}
	if len(r.Blocks) > 0 {
		parts = append(parts, fmt.Sprintf("blocks: %s", strings.Join(r.Blocks, ", ")))
	}
	if r.Complexity > 0 {
		parts = append(parts, fmt.Sprintf("complexity: %d", r.Complexity))
	}
	return strings.Join(parts, " | ")
}

// FilterValue returns the value to use for filtering (implements ListItem interface)
func (r ReadyTaskItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s", r.ID, r.TaskTitle, r.Status)
}

// LineType represents the classification of a line in CLI table output
type LineType int

const (
	// LineTypeUnknown represents an unclassified or invalid line
	LineTypeUnknown LineType = iota
	// LineTypeMetadata represents informational lines (tag, file path)
	LineTypeMetadata
	// LineTypeDecorative represents box-drawing separator lines
	LineTypeDecorative
	// LineTypeHeader represents the column header line
	LineTypeHeader
	// LineTypeData represents a data row with task information
	LineTypeData
	// LineTypeEmpty represents empty or whitespace-only lines
	LineTypeEmpty
)

// classifyLine determines the type of a line from CLI table output
func classifyLine(line string) LineType {
	// Empty or whitespace-only lines
	if strings.TrimSpace(line) == "" {
		return LineTypeEmpty
	}

	trimmed := strings.TrimSpace(line)

	// Metadata lines (tag info, file paths)
	if strings.HasPrefix(trimmed, "🏷") || 
	   strings.Contains(trimmed, "Listing tasks from:") {
		return LineTypeMetadata
	}

	// Decorative lines - contain only box-drawing characters and whitespace
	if isDecorativeLine(trimmed) {
		return LineTypeDecorative
	}

	// Header line - contains column headers
	if isHeaderLine(trimmed) {
		return LineTypeHeader
	}

	// Data rows - contain pipe separators and data
	if isDataRow(trimmed) {
		return LineTypeData
	}

	return LineTypeUnknown
}

// isDecorativeLine checks if a line contains only box-drawing characters
func isDecorativeLine(line string) bool {
	// Box-drawing characters used in table formatting
	// Decorative lines use corners, junctions, and horizontal lines
	decorativeChars := "┌┬┐├┼┤└┴┘─"
	
	// Must contain at least one decorative character
	if !strings.ContainsAny(line, decorativeChars) {
		return false
	}
	
	// All box-drawing chars for checking
	boxChars := "┌┬┐├┼┤└┴┘─│"
	
	for _, char := range line {
		// Skip whitespace
		if char == ' ' || char == '\t' {
			continue
		}
		// If character is not a box-drawing char, it's not decorative
		if !strings.ContainsRune(boxChars, char) {
			return false
		}
	}
	
	return true
}

// isHeaderLine checks if a line contains column headers
func isHeaderLine(line string) bool {
	// Header lines contain pipe separators and key column names
	if !strings.Contains(line, "│") {
		return false
	}
	
	// Check for common header keywords
	lowerLine := strings.ToLower(line)
	hasID := strings.Contains(lowerLine, "id")
	hasTitle := strings.Contains(lowerLine, "title")
	hasStatus := strings.Contains(lowerLine, "status")
	
	// Header must have at least ID and Title columns
	return hasID && hasTitle && hasStatus
}

// isDataRow checks if a line is a data row with task information
func isDataRow(line string) bool {
	// Data rows must contain pipe separators
	if !strings.Contains(line, "│") {
		return false
	}
	
	// Must not be a decorative line
	if isDecorativeLine(line) {
		return false
	}
	
	// Must not be a header line
	if isHeaderLine(line) {
		return false
	}
	
	// Data rows contain pipes but not exclusively box-drawing chars
	return true
}

// ColumnInfo represents information about a table column
type ColumnInfo struct {
	Name  string // Column name from header
	Start int    // Starting character position (inclusive)
	End   int    // Ending character position (exclusive)
}

// extractColumnBoundaries parses a header line to extract column positions
func extractColumnBoundaries(headerLine string) []ColumnInfo {
	if strings.TrimSpace(headerLine) == "" {
		return nil
	}

	var columns []ColumnInfo
	
	// The pipe character "│" is 3 bytes in UTF-8
	const pipeSeparator = "│"
	pipeLen := len(pipeSeparator)

	// Split by pipe character to find column boundaries
	parts := strings.Split(headerLine, pipeSeparator)

	// Track our position in the original string
	currentPos := 0

	// Skip first empty part (before first │) and last empty part (after last │)
	for i := 1; i < len(parts)-1; i++ {
		part := parts[i]

		// Calculate start and end positions
		// Account for the │ separator (3 bytes) before this part
		start := currentPos + pipeLen
		end := start + len(part)

		// Extract and clean column name
		name := strings.TrimSpace(part)

		// Only add columns with non-empty names
		if name != "" {
			columns = append(columns, ColumnInfo{
				Name:  name,
				Start: start,
				End:   end,
			})
		}

		// Move position forward (length of part + pipe separator)
		currentPos += len(part) + pipeLen
	}

	return columns
}

// getColumnValue extracts a column value from a data row using column boundaries
func getColumnValue(line string, column ColumnInfo) string {
	// Ensure we don't go out of bounds
	if column.Start >= len(line) {
		return ""
	}

	end := column.End
	if end > len(line) {
		end = len(line)
	}

	// Extract the substring and trim whitespace
	value := line[column.Start:end]
	return strings.TrimSpace(value)
}

// findColumnByName searches for a column by name (case-insensitive)
func findColumnByName(columns []ColumnInfo, name string) *ColumnInfo {
	lowerName := strings.ToLower(name)
	for i := range columns {
		if strings.ToLower(columns[i].Name) == lowerName {
			return &columns[i]
		}
		// Also check for partial matches (e.g., "Complex…" matches "complexity")
		if strings.HasPrefix(strings.ToLower(columns[i].Name), lowerName) ||
			strings.HasPrefix(lowerName, strings.ToLower(columns[i].Name)) {
			return &columns[i]
		}
	}
	return nil
}

// ReadyTasksDialog is a dialog for displaying and selecting ready tasks from CLI output
type ReadyTasksDialog struct {
	*ListDialog
	tasks              []ReadyTaskItem
	selectedIDs        []string
	rawOutput          string
	parseError         error
	mu                 sync.Mutex // Protects concurrent access to tasks during async fetching
	cache              map[string]*TaskDetails // Cache for fetched task details keyed by task ID
	showRawOutput      bool                   // Flag to show raw output when parsing fails
	cliExecutionError  error                  // Stores CLI execution error with context
	emptyResultMessage string                 // Custom message when no tasks available
	statusMessage      string                 // Status message shown during operations (e.g., "Configuring models...")
}

// NewReadyTasksDialog creates a new ready tasks dialog
func NewReadyTasksDialog() *ReadyTasksDialog {
	dialog := &ReadyTasksDialog{
		ListDialog:        NewListDialog("Ready Tasks", 80, 20, []ListItem{}),
		tasks:             []ReadyTaskItem{},
		selectedIDs:       []string{},
		rawOutput:         "",
		parseError:        nil,
		cache:             make(map[string]*TaskDetails),
		showRawOutput:     false,
		cliExecutionError: nil,
		emptyResultMessage: "",
		statusMessage:     "",
	}

	// Assign ID for dialog result handling
	dialog.ListDialog.BaseDialog.ID = "ready_tasks_dialog"

	// Enable multi-select for this dialog
	dialog.ListDialog.SetMultiSelect(true)
	dialog.ListDialog.SetShowDescription(true)

	// Enhanced footer hints with keyboard navigation instructions
	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Space", Label: "Toggle"},
		ShortcutHint{Key: "Alt+A", Label: "Select All"},
		ShortcutHint{Key: "Enter", Label: "Confirm"},
		ShortcutHint{Key: "Esc", Label: "Close"},
	)

	return dialog
}

// SetContent sets the raw CLI output and triggers parsing
func (d *ReadyTasksDialog) SetContent(content string) {
	d.rawOutput = content
	d.parseError = nil
	d.tasks = []ReadyTaskItem{}
	d.showRawOutput = false
	d.statusMessage = "" // Reset status message when new content is set

	// Check for empty results
	if strings.TrimSpace(content) == "" {
		d.parseError = fmt.Errorf("empty output")
		d.emptyResultMessage = "No ready tasks available"
		d.UpdateListItems()
		return
	}

	// Attempt to parse the content
	d.parseTasks()

	// Update list items
	d.UpdateListItems()

	// Asynchronously fetch full details for truncated tasks
	d.FetchAllTruncatedTaskDetails()
}

// UpdateListItems updates the list dialog items from the current tasks
func (d *ReadyTasksDialog) UpdateListItems() {
	items := make([]ListItem, len(d.tasks))
	for i := range d.tasks {
		// Create a copy for the list item
		item := d.tasks[i]
		items[i] = item
	}
	d.ListDialog.SetItems(items)
}

// parseTasks parses the raw CLI output into ReadyTaskItem structs
func (d *ReadyTasksDialog) parseTasks() {
	if strings.TrimSpace(d.rawOutput) == "" {
		d.parseError = fmt.Errorf("empty output")
		return
	}

	// Try JSON parsing first
	var jsonData []interface{}
	err := json.Unmarshal([]byte(d.rawOutput), &jsonData)
	if err == nil {
		// Valid JSON parsed
		if len(jsonData) > 0 {
			d.parseTasksFromJSON(jsonData)
		} else {
			// Empty JSON array - set message and error
			d.parseError = fmt.Errorf("no ready tasks found")
			d.emptyResultMessage = "No ready tasks available"
		}
		// Empty JSON array is valid - no error for parsing
		return
	}

	// Fall back to text parsing
	d.parseTasksFromText()

	// If text parsing also failed and no tasks were parsed, trigger fallback to raw output
	if d.parseError != nil && len(d.tasks) == 0 {
		d.showRawOutput = true
	}
}

// parseTasksFromJSON parses tasks from JSON output
func (d *ReadyTasksDialog) parseTasksFromJSON(data []interface{}) {
	for _, item := range data {
		if mapItem, ok := item.(map[string]interface{}); ok {
			task := d.parseTaskFromMap(mapItem)
			if task != nil {
				d.tasks = append(d.tasks, *task)
			}
		}
	}
}

// parseTaskFromMap extracts a task from a map
func (d *ReadyTasksDialog) parseTaskFromMap(m map[string]interface{}) *ReadyTaskItem {
	task := &ReadyTaskItem{}

	if id, ok := m["id"].(string); ok {
		task.ID = id
	}
	if title, ok := m["title"].(string); ok {
		cleanedTitle, isTruncated := detectTitleTruncation(title)
		task.TaskTitle = cleanedTitle
		task.TitleTruncated = isTruncated
	}
	if status, ok := m["status"].(string); ok {
		task.Status = status
	}
	if priority, ok := m["priority"].(string); ok {
		task.Priority = priority
	}
	if complexity, ok := m["complexity"].(float64); ok {
		task.Complexity = int(complexity)
	}

	// Parse dependencies array
	if deps, ok := m["dependencies"].([]interface{}); ok {
		for _, dep := range deps {
			if depStr, ok := dep.(string); ok {
				task.Dependencies = append(task.Dependencies, depStr)
			}
		}
	}

	// Parse blocks array
	if blocks, ok := m["blocks"].([]interface{}); ok {
		for _, block := range blocks {
			if blockStr, ok := block.(string); ok {
				task.Blocks = append(task.Blocks, blockStr)
			}
		}
	}

	if task.ID == "" && task.TaskTitle == "" {
		return nil
	}

	return task
}

// parseComplexity extracts complexity value from a string
// Handles formats like "● 8", "8", "● ", or empty
func parseComplexity(value string) int {
	// Remove common symbols
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "●", "")
	value = strings.ReplaceAll(value, "○", "")
	value = strings.TrimSpace(value)
	
	// Try to parse as integer
	var complexity int
	fmt.Sscanf(value, "%d", &complexity)
	return complexity
}

// parseList parses a comma-separated list into a slice
// Handles formats like "1, 2, 3" or "1.1, 1.2" or empty
func parseList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	
	parts := strings.Split(value, ",")
	var result []string
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	
	return result
}

// detectTitleTruncation checks if a title ends with '...' indicating truncation.
// Returns the cleaned title (without '...') and a boolean indicating if truncation was detected.
func detectTitleTruncation(title string) (string, bool) {
	trimmed := strings.TrimSpace(title)
	if strings.HasSuffix(trimmed, "...") {
		// Remove the trailing '...'
		cleaned := strings.TrimSuffix(trimmed, "...")
		cleaned = strings.TrimSpace(cleaned)
		return cleaned, true
	}
	return trimmed, false
}

// parseTasksFromTable parses tasks from table-formatted CLI output
// Returns an error if the table format is invalid or cannot be parsed
func (d *ReadyTasksDialog) parseTasksFromTable() error {
	lines := strings.Split(d.rawOutput, "\n")
	
	var headerCols []string
	var currentTask *ReadyTaskItem
	var foundHeader bool
	
	for _, line := range lines {
		lineType := classifyLine(line)
		
		switch lineType {
		case LineTypeHeader:
			// Split by pipe and extract column names
			parts := strings.Split(line, "│")
			headerCols = []string{}
			
			// Skip first and last parts (empty before first │ and after last │)
			for i := 1; i < len(parts)-1; i++ {
				headerCols = append(headerCols, strings.TrimSpace(parts[i]))
			}
			
			if len(headerCols) == 0 {
				return fmt.Errorf("failed to parse header: no columns found")
			}
			
			foundHeader = true
			
		case LineTypeData:
			if len(headerCols) == 0 {
				continue // Skip data rows before header is found
			}
			
			// Split data row by pipes
			parts := strings.Split(line, "│")
			if len(parts) < 3 { // Need at least opening, data, and closing
				continue
			}
			
			// Extract column values (skip first and last which are empty)
			values := make([]string, 0, len(parts)-2)
			for i := 1; i < len(parts)-1; i++ {
				values = append(values, strings.TrimSpace(parts[i]))
			}
			
			// Create a map of column name to value
			colMap := make(map[string]string)
			for i, col := range headerCols {
				if i < len(values) {
					colMap[strings.ToLower(col)] = values[i]
				}
			}
			
			// Helper to find value by column name (case-insensitive, partial match)
			getValue := func(name string) string {
				name = strings.ToLower(name)
				// Try exact match first
				if val, ok := colMap[name]; ok {
					return val
				}
				// Try partial match
				for k, v := range colMap {
					if strings.HasPrefix(k, name) || strings.HasPrefix(name, k) {
						return v
					}
				}
				return ""
			}
			
			id := getValue("id")
			
			// If ID is present, start a new task
			if id != "" {
				// Save previous task if exists
				if currentTask != nil && currentTask.ID != "" {
					// Clean up trailing ellipses in status before saving
					if strings.HasSuffix(currentTask.Status, "…") {
						statusPrefix := strings.TrimSuffix(currentTask.Status, "…")
						switch statusPrefix {
						case "in-progre":
							currentTask.Status = "in-progress"
						case "pen":
							currentTask.Status = "pending"
						default:
							currentTask.Status = statusPrefix
						}
					}
					d.tasks = append(d.tasks, *currentTask)
				}
				
				// Create new task
				currentTask = &ReadyTaskItem{
					ID: id,
				}
				
				// Extract all other fields
				titleRaw := getValue("title")
				cleanedTitle, isTruncated := detectTitleTruncation(titleRaw)
				currentTask.TaskTitle = cleanedTitle
				currentTask.TitleTruncated = isTruncated
				
				// Parse status (handle symbols that might be alone or with text)
				status := getValue("status")
				// Remove symbols
				status = strings.ReplaceAll(status, "▶", "")
				status = strings.ReplaceAll(status, "○", "")
				status = strings.ReplaceAll(status, "✓", "")
				status = strings.TrimSpace(status)
				// If status is empty after removing symbols, it might continue on next line
				currentTask.Status = status
				
				currentTask.Priority = getValue("priority")
				currentTask.Dependencies = parseList(getValue("dependencies"))
				currentTask.Blocks = parseList(getValue("blocks"))
				currentTask.Complexity = parseComplexity(getValue("complex"))
				
			} else if currentTask != nil {
				// This is a continuation line (empty ID column)
				// Append to title if it continues on next line
				titleContinuation := getValue("title")
				if titleContinuation != "" {
					cleanedContinuation, isTruncated := detectTitleTruncation(titleContinuation)
					if currentTask.TaskTitle != "" {
						currentTask.TaskTitle += " " + cleanedContinuation
					} else {
						currentTask.TaskTitle = cleanedContinuation
					}
					// Update truncation flag
					currentTask.TitleTruncated = isTruncated
				}
				
				// Status might continue on the next line (e.g., symbol on line 1, text on line 2)
				statusContinuation := getValue("status")
				// Remove any symbols
				statusContinuation = strings.ReplaceAll(statusContinuation, "▶", "")
				statusContinuation = strings.ReplaceAll(statusContinuation, "○", "")
				statusContinuation = strings.ReplaceAll(statusContinuation, "✓", "")
				statusContinuation = strings.TrimSpace(statusContinuation)
				
				if statusContinuation != "" {
					if currentTask.Status == "" {
						// First line had only symbol, this line has the actual status
						currentTask.Status = statusContinuation
					} else if strings.HasSuffix(currentTask.Status, "…") {
						// Handle truncation: "in-progre…" in previous line
						currentTask.Status = strings.TrimSuffix(currentTask.Status, "…") + statusContinuation
					}
				}
			}
		}
	}
	
	// Add the last task if exists
	if currentTask != nil && currentTask.ID != "" {
		// Clean up trailing ellipses in status if this is the last line
		if strings.HasSuffix(currentTask.Status, "…") {
			// Try common status expansions
			statusPrefix := strings.TrimSuffix(currentTask.Status, "…")
			switch statusPrefix {
			case "in-progre":
				currentTask.Status = "in-progress"
			case "pen":
				currentTask.Status = "pending"
			default:
				// Keep the ellipsis if we don't know the full word
				currentTask.Status = statusPrefix
			}
		}
		d.tasks = append(d.tasks, *currentTask)
	}
	
	// Check if we found a valid table structure
	if !foundHeader {
		return fmt.Errorf("no table header found in output")
	}
	
	if len(d.tasks) == 0 {
		return fmt.Errorf("no tasks parsed from table output")
	}
	
	return nil
}

// parseTasksFromText parses tasks from text output
func (d *ReadyTasksDialog) parseTasksFromText() {
	// Try table parsing first
	if err := d.parseTasksFromTable(); err == nil {
		// Successfully parsed as table
		return
	}
	
	// Fall back to regex-based parsing
	lines := strings.Split(d.rawOutput, "\n")

	// Regular expressions for parsing common formats
	idRegex := regexp.MustCompile(`^\s*#?\s*([0-9.]+)\s*[-–:]\s*(.+)$`)
	priorityRegex := regexp.MustCompile(`(?i:priority|severity):\s*(\w+)`)
	depsRegex := regexp.MustCompile(`(?i:dep|depends|depends-on|dependencies):\s*(.+)`)
	statusRegex := regexp.MustCompile(`(?i:status):\s*(\w+)`)
	complexityRegex := regexp.MustCompile(`(?i:complexity|score):\s*(\d+)`)

	var currentTask *ReadyTaskItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a task ID line
		if matches := idRegex.FindStringSubmatch(line); len(matches) > 2 {
			if currentTask != nil && currentTask.ID != "" {
				d.tasks = append(d.tasks, *currentTask)
			}
			taskTitleRaw := strings.TrimSpace(matches[2])
			cleanedTitle, isTruncated := detectTitleTruncation(taskTitleRaw)
			currentTask = &ReadyTaskItem{
				ID:             matches[1],
				TaskTitle:      cleanedTitle,
				TitleTruncated: isTruncated,
			}
			continue
		}

		if currentTask == nil {
			continue
		}

		// Parse additional fields from the current line
		if matches := priorityRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Priority = strings.ToLower(matches[1])
		}
		if matches := statusRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Status = strings.ToLower(matches[1])
		}
		if matches := complexityRegex.FindStringSubmatch(line); len(matches) > 1 {
			fmt.Sscanf(matches[1], "%d", &currentTask.Complexity)
		}
		if matches := depsRegex.FindStringSubmatch(line); len(matches) > 1 {
			deps := strings.Split(matches[1], ",")
			for _, dep := range deps {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					currentTask.Dependencies = append(currentTask.Dependencies, dep)
				}
			}
		}
	}

	// Add the last task if exists
	if currentTask != nil && currentTask.ID != "" {
		d.tasks = append(d.tasks, *currentTask)
	}
	
	// If no tasks were parsed by either method, set error with context
	if len(d.tasks) == 0 {
		d.parseError = fmt.Errorf("failed to parse tasks: unrecognized output format. Output starts with: %s", truncateStrLen(strings.TrimSpace(d.rawOutput), 50))
		d.showRawOutput = true
	}
}

// truncateStrLen is a helper to truncate strings for error messages
func truncateStrLen(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// GetSelectedTasks returns the IDs of selected tasks
func (d *ReadyTasksDialog) GetSelectedTasks() []string {
	selected := []string{}
	// Iterate through indices in order to preserve task order
	for i := range d.tasks {
		if d.ListDialog.selectedItems[i] {
			selected = append(selected, d.tasks[i].ID)
		}
	}
	return selected
}

// GetSelectionCount returns the count of selected tasks and total tasks
func (d *ReadyTasksDialog) GetSelectionCount() (int, int) {
	count := 0
	for _, selected := range d.ListDialog.selectedItems {
		if selected {
			count++
		}
	}
	return count, len(d.tasks)
}

// AllSelected returns true if all tasks are selected
func (d *ReadyTasksDialog) AllSelected() bool {
	count, total := d.GetSelectionCount()
	return total > 0 && count == total
}

// SelectAll selects or deselects all tasks
func (d *ReadyTasksDialog) SelectAll(selectAll bool) {
	if selectAll {
		// Select all tasks - add all indices to selectedItems
		for i := range d.tasks {
			d.ListDialog.selectedItems[i] = true
		}
	} else {
		// Deselect all tasks - clear the selectedItems map
		d.ListDialog.selectedItems = make(map[int]bool)
	}
}

// GetTruncatedTaskIDs returns the IDs of tasks with truncated titles
// These tasks will need full details fetching via task-master show
func (d *ReadyTasksDialog) GetTruncatedTaskIDs() []string {
	var truncatedIDs []string
	for _, task := range d.tasks {
		if task.TitleTruncated {
			truncatedIDs = append(truncatedIDs, task.ID)
		}
	}
	return truncatedIDs
}

// HasTruncatedTitles returns true if any tasks have truncated titles
func (d *ReadyTasksDialog) HasTruncatedTitles() bool {
	for _, task := range d.tasks {
		if task.TitleTruncated {
			return true
		}
	}
	return false
}

// GetCachedTaskDetails returns the cached task details for a given task ID if available
// Returns nil if the task is not in the cache
func (d *ReadyTasksDialog) GetCachedTaskDetails(taskID string) *TaskDetails {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.cache[taskID]
}

// SetCachedTaskDetails stores task details in the cache with proper locking
// This is the preferred method for setting cache entries to ensure thread safety
func (d *ReadyTasksDialog) SetCachedTaskDetails(taskID string, details *TaskDetails) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache[taskID] = details
}

// SetCLIExecutionError sets the CLI execution error with context for better error reporting
func (d *ReadyTasksDialog) SetCLIExecutionError(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err != nil {
		d.cliExecutionError = fmt.Errorf("CLI execution failed: %w", err)
		d.parseError = d.cliExecutionError
		d.showRawOutput = true
	}
}

// GetCLIExecutionError returns the stored CLI execution error
func (d *ReadyTasksDialog) GetCLIExecutionError() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.cliExecutionError
}

// IsEmpty returns true if no tasks are available
func (d *ReadyTasksDialog) IsEmpty() bool {
	return len(d.tasks) == 0
}

// HasParseError returns true if there was a parse error
func (d *ReadyTasksDialog) HasParseError() bool {
	return d.parseError != nil
}

// ShowRawOutput enables or disables showing raw output
func (d *ReadyTasksDialog) SetShowRawOutput(show bool) {
	d.showRawOutput = show
}

// SetEmptyResultMessage sets a custom message for when no tasks are available
func (d *ReadyTasksDialog) SetEmptyResultMessage(msg string) {
	d.emptyResultMessage = msg
}

// GetStatusMessage returns the current status message
func (d *ReadyTasksDialog) GetStatusMessage() string {
	return d.statusMessage
}

// SetStatusMessage sets a status message to display (e.g., "Configuring models for X tasks...")
func (d *ReadyTasksDialog) SetStatusMessage(msg string) {
	d.statusMessage = msg
}

// ClearStatusMessage clears the status message
func (d *ReadyTasksDialog) ClearStatusMessage() {
	d.statusMessage = ""
}

// InvalidateCache clears the cache for a specific task ID
// Useful when a task has been updated and needs to be refetched
func (d *ReadyTasksDialog) InvalidateCache(taskID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.cache, taskID)
}

// ClearCache clears the entire cache
// Useful for a full refresh of all task details
func (d *ReadyTasksDialog) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]*TaskDetails)
}

// GetCacheSize returns the number of cached tasks
func (d *ReadyTasksDialog) GetCacheSize() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.cache)
}

// FetchFullTaskDetails synchronously fetches full task details for a given task ID
// It executes task-master show {id} and parses all details from the output.
// Results are cached to avoid redundant fetches.
func (d *ReadyTasksDialog) FetchFullTaskDetails(taskID string) (string, error) {
	d.mu.Lock()
	// Check if details are already cached
	if cached, exists := d.cache[taskID]; exists && cached != nil && cached.Title != "" {
		d.mu.Unlock()
		return cached.Title, nil
	}
	d.mu.Unlock()

	// Execute task-master show command synchronously
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "task-master", "show", taskID)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch task details for %s: %w", taskID, err)
	}

	// Parse the output to extract complete task details
	details := d.ParseTaskDetails(string(output))
	if details == nil || details.Title == "" {
		return "", fmt.Errorf("could not extract details from task-master show %s", taskID)
	}

	// Store in cache for future use
	d.mu.Lock()
	d.cache[taskID] = details
	d.mu.Unlock()

	return details.Title, nil
}

// ParseTaskDetails parses task-master show output and extracts all task fields
func (d *ReadyTasksDialog) ParseTaskDetails(output string) *TaskDetails {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	details := &TaskDetails{}
	lines := strings.Split(output, "\n")

	// Parse line by line looking for field markers
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and decorative lines
		if trimmed == "" || isDecorativeLine(trimmed) {
			continue
		}

		// Extract ID from task header like "│ ID:                │ 9.1                                                                        │"
		if strings.Contains(line, "ID:") && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				details.ID = strings.TrimSpace(parts[2])
			}
		}

		// Extract Title
		if strings.Contains(trimmed, "Title:") && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				title := strings.TrimSpace(parts[2])
				// Remove trailing pipe and clean up
				title = strings.TrimRight(title, "│")
				title = strings.TrimSpace(title)
				if title != "" {
					details.Title = title
				}
			}
		}

		// Extract Status
		if strings.Contains(trimmed, "Status:") && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				status := strings.TrimSpace(parts[2])
				// Remove symbols like ○, ▶, ✓
				status = strings.ReplaceAll(status, "○", "")
				status = strings.ReplaceAll(status, "▶", "")
				status = strings.ReplaceAll(status, "✓", "")
				status = strings.TrimSpace(status)
				if status != "" {
					details.Status = status
				}
			}
		}

		// Extract Priority
		if strings.Contains(trimmed, "Priority:") && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				details.Priority = strings.TrimSpace(parts[2])
			}
		}

		// Extract Dependencies
		if strings.Contains(trimmed, "Dependencies:") && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				depStr := strings.TrimSpace(parts[2])
				if depStr != "" && depStr != "-" {
					details.Dependencies = parseList(depStr)
				}
			}
		}

		// Extract Blocks
		if strings.Contains(trimmed, "Blocks:") && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				blockStr := strings.TrimSpace(parts[2])
				if blockStr != "" && blockStr != "-" {
					details.Blocks = parseList(blockStr)
				}
			}
		}

		// Extract Complexity
		if (strings.Contains(trimmed, "Complexity:") || strings.Contains(trimmed, "Complex")) && strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				complexStr := strings.TrimSpace(parts[2])
				details.Complexity = parseComplexity(complexStr)
			}
		}
	}

	// Return nil if we didn't get any meaningful data
	if details.ID == "" && details.Title == "" {
		return nil
	}

	return details
}

// extractTitleFromOutput parses task-master show output and extracts the title
// Looks for lines containing "Title:" followed by the title text
func (d *ReadyTasksDialog) extractTitleFromOutput(output string) string {
	lines := strings.Split(output, "\n")

	// Look for the title line in the output
	// Format is typically: "│ Title:             │ [title text]                                                             │"
	for i, line := range lines {
		if strings.Contains(line, "Title") && strings.Contains(line, "│") {
			// Extract the title value from this line
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				// The title is typically in the third pipe-separated section
				title := strings.TrimSpace(parts[2])
				if title != "" {
					return title
				}
			}
		}

		// Also check if title continues on the same line after the ID
		if strings.HasPrefix(strings.TrimSpace(line), "Title:") {
			// Extract everything after "Title:"
			colon := strings.Index(line, ":")
			if colon != -1 {
				title := strings.TrimSpace(line[colon+1:])
				// Remove trailing pipe characters and other decorations
				title = strings.TrimRight(title, "│")
				title = strings.TrimSpace(title)
				if title != "" {
					return title
				}
			}
		}

		// Check if this line contains task title information
		if i > 0 && (strings.Contains(lines[i-1], "Title") || strings.Contains(line, " - ")) {
			// Extract title from pattern like "# Task: #9 - Task Title"
			dashIdx := strings.Index(line, " - ")
			if dashIdx != -1 {
				title := strings.TrimSpace(line[dashIdx+3:])
				// Remove any trailing decorations
				title = strings.TrimRight(title, "│")
				title = strings.TrimSpace(title)
				if title != "" {
					return title
				}
			}
		}
	}

	return ""
}

// FetchAllTruncatedTaskDetails asynchronously fetches details for all truncated tasks
// Uses goroutines with WaitGroup to manage concurrency and safely update the task list.
// Skips tasks that are already in the cache to avoid redundant fetches.
func (d *ReadyTasksDialog) FetchAllTruncatedTaskDetails() {
	// Get list of truncated task IDs
	truncatedIDs := d.GetTruncatedTaskIDs()
	if len(truncatedIDs) == 0 {
		return
	}

	// Filter out tasks that are already in cache
	var tasksToPrefetch []string
	d.mu.Lock()
	for _, id := range truncatedIDs {
		if cached, exists := d.cache[id]; !exists || cached == nil || cached.Title == "" {
			tasksToPrefetch = append(tasksToPrefetch, id)
		}
	}
	d.mu.Unlock()

	if len(tasksToPrefetch) == 0 {
		// All tasks are already cached, update UI with cached data
		d.updateUIWithCachedDetails(truncatedIDs)
		return
	}

	// Use WaitGroup to manage goroutine lifecycle
	var wg sync.WaitGroup
	// Limit concurrent fetches to avoid overwhelming system
	maxConcurrent := 3
	semaphore := make(chan struct{}, maxConcurrent)

	for _, taskID := range tasksToPrefetch {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Fetch full task details (will use cache if available)
			fullTitle, err := d.FetchFullTaskDetails(id)
			if err != nil {
				// Log error but continue with other tasks
				return
			}

			// Update the task with full title
			d.mu.Lock()
			for i, task := range d.tasks {
				if task.ID == id {
					d.tasks[i].TaskTitle = fullTitle
					d.tasks[i].TitleTruncated = false // Mark as no longer truncated
					break
				}
			}
			d.mu.Unlock()
		}(taskID)
	}

	// Wait for all goroutines to complete in the background
	go func() {
		wg.Wait()
		// After all fetches complete, update the list display
		d.updateUIAfterFetch()
	}()
}

// updateUIWithCachedDetails updates the UI with cached task details
// This is called when all truncated tasks are already cached
func (d *ReadyTasksDialog) updateUIWithCachedDetails(truncatedIDs []string) {
	d.mu.Lock()
	for _, id := range truncatedIDs {
		if cached, exists := d.cache[id]; exists && cached != nil && cached.Title != "" {
			for i, task := range d.tasks {
				if task.ID == id {
					d.tasks[i].TaskTitle = cached.Title
					d.tasks[i].TitleTruncated = false
					break
				}
			}
		}
	}
	d.mu.Unlock()
	
	// Update the list display
	d.updateUIAfterFetch()
}

// updateUIAfterFetch updates the list display with the latest task data
// This should be called after async fetch operations complete
func (d *ReadyTasksDialog) updateUIAfterFetch() {
	d.mu.Lock()
	items := make([]ListItem, len(d.tasks))
	for i := range d.tasks {
		item := d.tasks[i]
		items[i] = item
	}
	d.mu.Unlock()
	d.ListDialog.SetItems(items)
}

// HandleKey handles key presses in the dialog, including Alt+A for Select All
func (d *ReadyTasksDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Handle Alt+A for Select All / Deselect All
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'a' {
		if msg.Alt {
			// Toggle between select all and deselect all
			d.SelectAll(!d.AllSelected())
			return DialogResultNone, nil
		}
	}
	
	// Handle Enter key directly - emit DialogResultMsg with selected tasks
	// This fixes the issue where ListSelectionMsg from ListDialog.HandleKey()
	// was being emitted as a command AFTER the dialog was closed due to 
	// DialogResultConfirm, preventing the message from ever reaching Update()
	if msg.String() == "enter" {
		selectedTasks := d.GetSelectedTasks()
		if len(selectedTasks) == 0 {
			// No tasks selected - cancel the dialog
			return DialogResultCancel, nil
		}
		
		// Show configuration message when multiple tasks are selected
		if len(selectedTasks) > 1 {
			d.statusMessage = fmt.Sprintf("Configuring models for %d tasks...", len(selectedTasks))
		}
		
		// Return DialogResultNone + cmd that emits DialogResultMsg
		// This prevents immediate dialog close by DialogManager but triggers
		// the proper message flow to handleDialogResultMsg in delete_workflow.go
		return DialogResultNone, func() tea.Msg {
			return DialogResultMsg{
				ID:     d.ListDialog.BaseDialog.ID,
				Button: "confirm",
				Value:  selectedTasks,
			}
		}
	}
	
	// Handle Escape key - cancel the dialog
	if msg.String() == "esc" {
		return DialogResultCancel, nil
	}
	
	// Delegate other key presses to ListDialog (navigation, space for toggle, etc.)
	return d.ListDialog.HandleKey(msg)
}

// Update processes messages and updates dialog state
func (d *ReadyTasksDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	// Note: Enter key is now handled directly in HandleKey() to fix the issue where
	// ListSelectionMsg from ListDialog was emitted as a command AFTER the dialog
	// was closed due to DialogResultConfirm. This Update() method now only handles
	// delegation to the embedded ListDialog for non-key messages (like window resize).
	
	// Delegate all messages to ListDialog, but return self to preserve custom View()
	_, cmd := d.ListDialog.Update(msg)
	return d, cmd
}

// View renders the dialog with custom table-like layout
func (d *ReadyTasksDialog) View() string {
	// If there's a parse error and we should show raw output, display it
	if d.showRawOutput && d.parseError != nil {
		return d.renderRawOutput()
	}

	// If there's a parse error but no raw output flag, show error message
	if d.parseError != nil && !d.showRawOutput {
		return d.RenderBorder(fmt.Sprintf("Error parsing tasks: %v", d.parseError))
	}

	if len(d.tasks) == 0 {
		if d.emptyResultMessage != "" {
			emptyContent := lipgloss.NewStyle().
				Foreground(d.ListDialog.Style.TextColor).
				Align(lipgloss.Center).
				Width(76).
				Render(d.emptyResultMessage)
			return d.RenderBorder(emptyContent)
		}
		emptyContent := lipgloss.NewStyle().
			Foreground(d.ListDialog.Style.TextColor).
			Align(lipgloss.Center).
			Width(76).
			Render("No ready tasks found in output")
		return d.RenderBorder(emptyContent)
	}

	return d.renderTaskTable()
}

// renderRawOutput renders the raw CLI output when parsing fails
func (d *ReadyTasksDialog) renderRawOutput() string {
	var lines []string

	// Error header
	errorMsg := fmt.Sprintf("Error parsing task list: %v", d.parseError)
	errorStyle := lipgloss.NewStyle().
		Foreground(d.ListDialog.Style.ErrorColor).
		Bold(true)
	lines = append(lines, errorStyle.Render(errorMsg))

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(d.ListDialog.Style.BorderColor).
		Render(strings.Repeat("─", 76))
	lines = append(lines, separator)
	
	// Add spacing
	lines = append(lines, "")

	// Raw output
	rawLines := strings.Split(d.rawOutput, "\n")
	maxLines := 15 // Show up to 15 lines of raw output
	for i, line := range rawLines {
		if i >= maxLines {
			if len(rawLines) > maxLines {
				infoStyle := lipgloss.NewStyle().
					Foreground(d.ListDialog.Style.TextColor).
					Faint(true)
				lines = append(lines, infoStyle.Render(fmt.Sprintf("... [%d more lines]", len(rawLines)-maxLines)))
			}
			break
		}
		// Truncate long lines
		if len(line) > 76 {
			line = line[:73] + "..."
		}
		textStyle := lipgloss.NewStyle().
			Foreground(d.ListDialog.Style.TextColor)
		lines = append(lines, textStyle.Render(line))
	}

	// Footer with instructions
	lines = append(lines, "")
	footerStyle := lipgloss.NewStyle().
		Foreground(d.ListDialog.Style.BorderColor).
		Faint(true)
	lines = append(lines, separator)
	lines = append(lines, footerStyle.Render("Press Esc to close"))

	content := strings.Join(lines, "\n")
	return d.RenderBorder(content)
}

// renderTaskTable renders a table of ready tasks with checkboxes and highlighting
func (d *ReadyTasksDialog) renderTaskTable() string {
	var lines []string

	// Header row with selection status
	header := d.formatTaskRow("ID", "Title", "Priority", "Complexity", true)
	headerStyle := lipgloss.NewStyle().
		Foreground(d.ListDialog.Style.FocusedBorderColor).
		Bold(true)
	lines = append(lines, headerStyle.Render(header))

	// Separator below header
	separator := lipgloss.NewStyle().
		Foreground(d.ListDialog.Style.BorderColor).
		Render(strings.Repeat("─", 76))
	lines = append(lines, separator)

	// Selection status line
	selected, total := d.GetSelectionCount()
	
	// Format selection count display
	selectionCountText := fmt.Sprintf("Selected: %d task(s)", selected)
	statusStyle := lipgloss.NewStyle().
		Foreground(d.ListDialog.Style.TextColor)
	
	// Add additional help text based on selection state
	var helpText string
	if selected == 0 {
		helpText = fmt.Sprintf("%s / %d total (Space to select, Alt+A for all)", selectionCountText, total)
		statusStyle = statusStyle.Faint(true)
	} else if d.AllSelected() {
		helpText = fmt.Sprintf("%s / %d total (Alt+A to deselect all)", selectionCountText, total)
		statusStyle = statusStyle.Bold(true)
	} else {
		helpText = fmt.Sprintf("%s / %d total (Space to toggle, Alt+A for all)", selectionCountText, total)
		statusStyle = statusStyle.Bold(true)
	}
	
	lines = append(lines, statusStyle.Render(helpText))
	
	// Add spacing between header area and content
	lines = append(lines, "")

	// Content rows
	maxVisible := 15 // Show up to 15 tasks
	offset := 0

	// Ensure selected index is visible
	if d.ListDialog.selectedIndex < offset {
		offset = d.ListDialog.selectedIndex
	} else if d.ListDialog.selectedIndex >= offset+maxVisible {
		offset = d.ListDialog.selectedIndex - maxVisible + 1
	}

	end := offset + maxVisible
	if end > len(d.tasks) {
		end = len(d.tasks)
	}

	for i := offset; i < end; i++ {
		task := d.tasks[i]
		isFocused := i == d.ListDialog.selectedIndex
		isSelected := d.ListDialog.selectedItems[i]

		line := d.formatTaskRowWithCheckbox(task, isFocused, isSelected)

		// Apply styling - just use color and bold, no background changes
		if isFocused {
			// Highlight focused item with accent color and bold text (no background)
			style := lipgloss.NewStyle().
				Foreground(d.ListDialog.Style.FocusedBorderColor).
				Bold(true)
			line = style.Render(line)
		} else {
			style := lipgloss.NewStyle().
				Foreground(d.ListDialog.Style.TextColor)
			line = style.Render(line)
		}

		lines = append(lines, line)
	}

	// Add spacing before scroll indicator
	if offset > 0 || end < len(d.tasks) {
		lines = append(lines, "")
		scrollInfo := ""
		if offset > 0 {
			scrollInfo += "↑ "
		}
		scrollInfo += fmt.Sprintf("↓↑ [%d/%d] items ", d.ListDialog.selectedIndex+1, len(d.tasks))
		if end < len(d.tasks) {
			scrollInfo += "↓"
		}
		scrollStyle := lipgloss.NewStyle().
			Foreground(d.ListDialog.Style.TextColor).
			Faint(true).
			Align(lipgloss.Center).
			Width(76)
		lines = append(lines, scrollStyle.Render(scrollInfo))
	}

	// Add status message if present (e.g., "Configuring models for X tasks...")
	if d.statusMessage != "" {
		lines = append(lines, "")
		statusStyle := lipgloss.NewStyle().
			Foreground(d.ListDialog.Style.FocusedBorderColor).
			Bold(true).
			Align(lipgloss.Center).
			Width(76)
		lines = append(lines, statusStyle.Render(d.statusMessage))
	}

	content := strings.Join(lines, "\n")
	return d.RenderBorder(content)
}

// formatTaskRow formats a task row without checkbox (for header or display)
func (d *ReadyTasksDialog) formatTaskRow(id, title, priority string, complexity interface{}, isHeader bool) string {
	// Column widths: ID(5) | Title(30) | Priority(10) | Complexity(8) + separators(3)
	const (
		idWidth        = 5
		titleWidth     = 30
		priorityWidth  = 10
		complexityWidth = 8
	)

	// Format complexity
	complexityStr := ""
	if c, ok := complexity.(int); ok && c > 0 {
		complexityStr = fmt.Sprintf("%d", c)
	} else if c, ok := complexity.(string); ok {
		complexityStr = c
	}

	// Apply consistent padding to match checkbox area in formatTaskRowWithCheckbox
	padding := "      " // Matches width of "► [✓] " (checkbox area)
	
	// For the header, ensure it lines up with the content rows that include checkboxes
	row := fmt.Sprintf("%s%-5s │ %-30s │ %-10s │ %8s",
		padding,
		truncateStr(id, idWidth),
		truncateStr(title, titleWidth),
		truncateStr(priority, priorityWidth),
		truncateStr(complexityStr, complexityWidth),
	)

	// Use visible width for consistency
	const totalWidth = 76
	displayWidth := lipgloss.Width(row)
	
	if displayWidth < totalWidth {
		row = row + strings.Repeat(" ", totalWidth-displayWidth)
	}
	
	return row
}

// formatTaskRowWithCheckbox formats a task row with checkbox and selection indicator
func (d *ReadyTasksDialog) formatTaskRowWithCheckbox(task ReadyTaskItem, focused, selected bool) string {
	// Create checkbox indicator with checkmark
	checkbox := "[ ]"
	if selected {
		checkbox = "[✓]"
	}

	// Add focus indicator - show arrow for focused item
	prefix := ""
	if focused {
		prefix = "► "
	}

	// Format task data with consistent spacing regardless of selection state
	priority := task.Priority
	if priority == "" {
		priority = "-"
	}

	// Consistent alignment with fixed-width checkbox area
	checkboxArea := fmt.Sprintf("%s%s ", prefix, checkbox)
	
	// Ensure checkbox area has consistent width whether focused or not
	if !focused {
		// Add padding to align with focused items that have the arrow
		checkboxArea = strings.Repeat(" ", 2) + checkbox + " "
	}
	
	row := fmt.Sprintf("%s%-5s │ %-30s │ %-10s │ %8d",
		checkboxArea,
		truncateStr(task.ID, 5),
		truncateStr(task.TaskTitle, 30),
		truncateStr(priority, 10),
		task.Complexity,
	)

	// Use visible width (accounting for ANSI codes) rather than byte length
	displayWidth := lipgloss.Width(row)
	const totalWidth = 76
	
	if displayWidth < totalWidth {
		row = row + strings.Repeat(" ", totalWidth - displayWidth)
	} else if displayWidth > totalWidth {
		// This shouldn't happen with our formatting, but handle it safely
		// We can't just truncate bytes as it may cut ANSI codes
		// Instead, rebuild with smaller column widths
		row = fmt.Sprintf("%s%-5s │ %-25s │ %-8s │ %6d",
			checkboxArea,
			truncateStr(task.ID, 5),
			truncateStr(task.TaskTitle, 25),
			truncateStr(priority, 8),
			task.Complexity,
		)
	}

	return row
}

// truncateStr truncates a string to a maximum length
func truncateStr(s string, maxLen int) string {
	if len(s) > maxLen {
		if maxLen > 3 {
			return s[:maxLen-3] + "..."
		}
		return s[:maxLen]
	}
	return s
}
