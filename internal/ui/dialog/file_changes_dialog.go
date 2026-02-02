package dialog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/agreen757/tm-tui/internal/filechanges"
	"github.com/agreen757/tm-tui/internal/git"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// FileChangesDialog is a dialog for exploring file changes across all tasks
type FileChangesDialog struct {
	BaseDialog
	tracker       *filechanges.FileChangeTracker
	gitService    *git.GitService
	filterTask    string
	filterStatus  string
	searchQuery   string
	groupBy       string // "task", "directory", or "time"
	fileTree      *FileTree
	preview       *Preview
	focusedPanel  string // "tree" or "preview"
	width         int
	height        int
}

// FileTree displays a hierarchical view of file changes
type FileTree struct {
	dialog        *FileChangesDialog
	items         []FileTreeItem
	selectedIdx   int
	expandedItems map[string]bool
	width         int
	height        int
}

// FileTreeItem represents an item in the file tree
type FileTreeItem struct {
	Type      string                  // "task", "directory", or "file"
	ID        string                  // Task ID or directory path
	Path      string                  // File path for file items
	Change    *taskmaster.FileChange  // For file items
	Level     int                     // Indentation level
	Collapsed bool                    // For expandable items
}

// Preview displays file content or diff
type Preview struct {
	file      string
	content   string
	diffMode  bool
	lines     []string
	scrollPos int
	width     int
	height    int
}

// NewFileChangesDialog creates a new file changes dialog
func NewFileChangesDialog(tracker *filechanges.FileChangeTracker, gitService *git.GitService) *FileChangesDialog {
	d := &FileChangesDialog{
		tracker:      tracker,
		gitService:   gitService,
		groupBy:      "task",
		filterStatus: "all",
		focusedPanel: "tree",
		width:        80,
		height:       24,
	}
	d.fileTree = NewFileTree(d)
	d.preview = NewPreview()
	d.BaseDialog = NewBaseDialog("File Changes", 80, 24, DialogKindCustom)
	d.BaseDialog.ID = "file-changes"
	d.BaseDialog.SetCancellable(true)
	return d
}

// NewFileTree creates a new file tree component
func NewFileTree(dialog *FileChangesDialog) *FileTree {
	return &FileTree{
		dialog:        dialog,
		items:         []FileTreeItem{},
		selectedIdx:   0,
		expandedItems: make(map[string]bool),
	}
}

// NewPreview creates a new preview panel
func NewPreview() *Preview {
	return &Preview{
		lines:     []string{},
		scrollPos: 0,
	}
}

// Title returns the dialog title
func (d *FileChangesDialog) Title() string {
	return "File Changes"
}

// Kind returns the dialog kind
func (d *FileChangesDialog) Kind() DialogKind {
	return DialogKindCustom
}

// Init initializes the dialog
func (d *FileChangesDialog) Init() tea.Cmd {
	// Rebuild file tree on initialization
	d.fileTree.buildTree()
	return nil
}

// Update handles messages
func (d *FileChangesDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.updateLayout()
	}
	return d, nil
}

// View renders the dialog
func (d *FileChangesDialog) View() string {
	content := d.Render(d.width, d.height)
	return d.RenderBorder(content)
}

// Render draws the dialog content
func (d *FileChangesDialog) Render(width, height int) string {
	if width < 40 || height < 10 {
		return "Window too small"
	}

	d.width = width
	d.height = height
	d.updateLayout()

	// Calculate layout dimensions
	filterBarHeight := 2
	statusBarHeight := 1
	contentHeight := height - filterBarHeight - statusBarHeight - 2 // -2 for borders

	// Render components
	filterBar := d.renderFilterBar(width)
	treeWidth := width / 2
	previewWidth := width - treeWidth - 1 // -1 for separator

	tree := d.fileTree.Render(treeWidth, contentHeight)
	preview := d.preview.Render(previewWidth, contentHeight)
	statusBar := d.renderStatusBar(width)

	// Combine components
	var content strings.Builder
	content.WriteString(filterBar)
	content.WriteString("\n")

	// Split-view: tree | preview
	treeLines := strings.Split(tree, "\n")
	previewLines := strings.Split(preview, "\n")

	maxLines := contentHeight
	if len(treeLines) < maxLines {
		maxLines = len(treeLines)
	}
	if len(previewLines) < maxLines {
		maxLines = len(previewLines)
	}

	for i := 0; i < contentHeight; i++ {
		var treeLine, previewLine string
		if i < len(treeLines) {
			treeLine = treeLines[i]
		} else {
			treeLine = strings.Repeat(" ", treeWidth)
		}
		if i < len(previewLines) {
			previewLine = previewLines[i]
		} else {
			previewLine = strings.Repeat(" ", previewWidth)
		}

		// Pad tree line to exact width
		treeLine = lipgloss.NewStyle().Width(treeWidth).Render(treeLine)
		content.WriteString(treeLine)
		content.WriteString("│")
		content.WriteString(previewLine)
		content.WriteString("\n")
	}

	content.WriteString(statusBar)

	return content.String()
}

// HandleKey processes keyboard input
func (d *FileChangesDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return DialogResultCancel, nil
	case "tab":
		d.toggleFocusedPanel()
		return DialogResultNone, nil
	case "g":
		// Toggle grouping mode
		d.cycleGroupBy()
		d.fileTree.buildTree()
		return DialogResultNone, nil
	case "f":
		// Cycle through status filters
		d.ToggleStatusFilter()
		return DialogResultNone, nil
	case "F":
		// Clear all filters
		d.ClearFilters()
		return DialogResultNone, nil
	case "/":
		// For now, just show that search is available in status bar
		// Full search input dialog can be implemented as an enhancement
		return DialogResultNone, nil
	default:
		// Delegate to focused panel
		if d.focusedPanel == "tree" {
			return d.fileTree.HandleKeyEvent(msg)
		} else if d.focusedPanel == "preview" {
			return d.preview.HandleKeyEvent(msg)
		}
	}
	return DialogResultNone, nil
}

// renderFilterBar renders the filter bar at the top
func (d *FileChangesDialog) renderFilterBar(width int) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(width)

	filterInfo := fmt.Sprintf("Group: %s | Filter: %s | Search: %s",
		d.groupBy,
		d.getFilterDisplay(),
		d.getSearchDisplay())

	return style.Render(filterInfo)
}

// renderStatusBar renders the status bar at the bottom
func (d *FileChangesDialog) renderStatusBar(width int) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(width)

	var help string
	if d.focusedPanel == "tree" {
		help = "↑/↓: Navigate • ←/→: Collapse/Expand • Tab: Switch Panel • G: Group By • F: Filter • Shift+F: Clear • Esc: Close"
	} else {
		help = "↑/↓: Scroll • Tab: Switch Panel • F: Filter • Esc: Close"
	}

	return style.Render(help)
}

// getFilterDisplay returns the current filter display text
func (d *FileChangesDialog) getFilterDisplay() string {
	if d.filterTask != "" {
		return fmt.Sprintf("Task %s", d.filterTask)
	}
	if d.filterStatus != "" && d.filterStatus != "all" {
		return d.filterStatus
	}
	return "All"
}

// getSearchDisplay returns the current search display text
func (d *FileChangesDialog) getSearchDisplay() string {
	if d.searchQuery != "" {
		return d.searchQuery
	}
	return "None"
}

// toggleFocusedPanel switches focus between tree and preview
func (d *FileChangesDialog) toggleFocusedPanel() {
	if d.focusedPanel == "tree" {
		d.focusedPanel = "preview"
	} else {
		d.focusedPanel = "tree"
	}
}

// cycleGroupBy cycles through grouping modes
func (d *FileChangesDialog) cycleGroupBy() {
	switch d.groupBy {
	case "task":
		d.groupBy = "directory"
	case "directory":
		d.groupBy = "time"
	case "time":
		d.groupBy = "task"
	}
}

// updateLayout updates component dimensions
func (d *FileChangesDialog) updateLayout() {
	filterBarHeight := 2
	statusBarHeight := 1
	contentHeight := d.height - filterBarHeight - statusBarHeight - 2

	treeWidth := d.width / 2
	previewWidth := d.width - treeWidth - 1

	d.fileTree.width = treeWidth
	d.fileTree.height = contentHeight

	d.preview.width = previewWidth
	d.preview.height = contentHeight
}

// FileTree Methods

// Render draws the file tree
func (t *FileTree) Render(width, height int) string {
	t.width = width
	t.height = height

	if len(t.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(width).
			Render("No file changes found")
	}

	var lines []string
	startIdx := 0
	if t.selectedIdx >= height {
		startIdx = t.selectedIdx - height + 1
	}

	for i := startIdx; i < len(t.items) && len(lines) < height; i++ {
		item := t.items[i]
		line := t.renderItem(item, i == t.selectedIdx)
		lines = append(lines, line)
	}

	// Pad remaining lines
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderItem renders a single tree item
func (t *FileTree) renderItem(item FileTreeItem, selected bool) string {
	indent := strings.Repeat("  ", item.Level)
	
	var icon, text string
	switch item.Type {
	case "task":
		if item.Collapsed {
			icon = "▶"
		} else {
			icon = "▼"
		}
		text = fmt.Sprintf("%s Task %s", icon, item.ID)
	case "directory":
		if item.Collapsed {
			icon = "▶"
		} else {
			icon = "▼"
		}
		text = fmt.Sprintf("%s %s/", icon, item.Path)
	case "file":
		icon = t.getChangeIcon(item.Change.ChangeType)
		displayPath := item.Path
		
		// Highlight search matches
		if t.dialog.searchQuery != "" {
			query := strings.ToLower(t.dialog.searchQuery)
			lowerPath := strings.ToLower(item.Path)
			
			if idx := strings.Index(lowerPath, query); idx >= 0 {
				// Split path around the match and highlight the matching part
				before := displayPath[:idx]
				match := displayPath[idx:idx+len(t.dialog.searchQuery)]
				after := displayPath[idx+len(t.dialog.searchQuery):]
				
				highlightStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("220")).
					Foreground(lipgloss.Color("0")).
					Bold(true)
				
				displayPath = before + highlightStyle.Render(match) + after
			}
		}
		
		text = fmt.Sprintf("%s %s", icon, displayPath)
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240")).Bold(true)
	}

	return style.Render(indent + text)
}

// getChangeIcon returns an icon for the change type
func (t *FileTree) getChangeIcon(changeType string) string {
	switch changeType {
	case "added":
		return "+"
	case "modified":
		return "~"
	case "deleted":
		return "-"
	default:
		return "•"
	}
}

// HandleKeyEvent processes keyboard input for the tree
func (t *FileTree) HandleKeyEvent(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		t.moveUp()
	case "down", "j":
		t.moveDown()
	case "left", "h":
		t.collapseItem()
	case "right", "l":
		t.expandItem()
	case "enter":
		t.selectItem()
	}
	return DialogResultNone, nil
}

// moveUp moves selection up
func (t *FileTree) moveUp() {
	if t.selectedIdx > 0 {
		t.selectedIdx--
	}
}

// moveDown moves selection down
func (t *FileTree) moveDown() {
	if t.selectedIdx < len(t.items)-1 {
		t.selectedIdx++
	}
}

// collapseItem collapses the selected item
func (t *FileTree) collapseItem() {
	if t.selectedIdx < len(t.items) {
		item := t.items[t.selectedIdx]
		if item.Type == "task" || item.Type == "directory" {
			t.expandedItems[item.ID] = false
			item.Collapsed = true
			t.buildTree()
		}
	}
}

// expandItem expands the selected item
func (t *FileTree) expandItem() {
	if t.selectedIdx < len(t.items) {
		item := t.items[t.selectedIdx]
		if item.Type == "task" || item.Type == "directory" {
			t.expandedItems[item.ID] = true
			item.Collapsed = false
			t.buildTree()
		}
	}
}

// selectItem handles enter key on selected item
func (t *FileTree) selectItem() {
	if t.selectedIdx < len(t.items) {
		item := t.items[t.selectedIdx]
		if item.Type == "file" && item.Change != nil {
			// Load file in preview
			t.dialog.preview.LoadFile(item.Change.Path, t.dialog.gitService)
		}
	}
}

// buildTree builds the tree structure based on groupBy mode
func (t *FileTree) buildTree() {
	t.items = []FileTreeItem{}

	switch t.dialog.groupBy {
	case "task":
		t.buildTreeByTask()
	case "directory":
		t.buildTreeByDirectory()
	case "time":
		t.buildTreeByTime()
	}
}

// matchesFilters checks if a file change matches the current filters
func (t *FileTree) matchesFilters(change *taskmaster.FileChange, taskID string) bool {
	// Task filter
	if t.dialog.filterTask != "" && taskID != t.dialog.filterTask {
		return false
	}
	
	// Status filter
	if t.dialog.filterStatus != "" && t.dialog.filterStatus != "all" {
		if change.ChangeType != t.dialog.filterStatus {
			return false
		}
	}
	
	// Search query filter
	if t.dialog.searchQuery != "" {
		query := strings.ToLower(t.dialog.searchQuery)
		path := strings.ToLower(change.Path)
		desc := strings.ToLower(change.Description)
		
		if !strings.Contains(path, query) && !strings.Contains(desc, query) {
			return false
		}
	}
	
	return true
}

// buildTreeByTask builds tree grouped by task
func (t *FileTree) buildTreeByTask() {
	allChanges := t.dialog.tracker.GetAllTaskFileChanges()

	for taskID, changes := range allChanges {
		// Skip task if filterTask is set and doesn't match
		if t.dialog.filterTask != "" && taskID != t.dialog.filterTask {
			continue
		}
		
		// Filter changes
		var filteredChanges []taskmaster.FileChange
		for _, change := range changes {
			if t.matchesFilters(&change, taskID) {
				filteredChanges = append(filteredChanges, change)
			}
		}
		
		// Skip task if no changes match filters
		if len(filteredChanges) == 0 {
			continue
		}
		
		// Add task item
		taskItem := FileTreeItem{
			Type:      "task",
			ID:        taskID,
			Level:     0,
			Collapsed: !t.expandedItems[taskID],
		}
		t.items = append(t.items, taskItem)

		// Add file items if expanded
		if t.expandedItems[taskID] {
			for _, change := range filteredChanges {
				fileItem := FileTreeItem{
					Type:   "file",
					Path:   change.Path,
					Change: &change,
					Level:  1,
				}
				t.items = append(t.items, fileItem)
			}
		}
	}
}

// buildTreeByDirectory builds tree grouped by directory
func (t *FileTree) buildTreeByDirectory() {
	allChanges := t.dialog.tracker.GetAllTaskFileChanges()
	
	// Collect all unique files with their changes (applying filters)
	fileMap := make(map[string]*taskmaster.FileChange)
	for taskID, changes := range allChanges {
		for _, change := range changes {
			// Apply filters
			if !t.matchesFilters(&change, taskID) {
				continue
			}
			
			// Use the first occurrence of each file
			if _, exists := fileMap[change.Path]; !exists {
				changeCopy := change
				fileMap[change.Path] = &changeCopy
			}
		}
	}
	
	// Build directory tree structure
	dirTree := t.buildDirectoryTree(fileMap)
	
	// Convert to flat list with proper indentation
	t.items = []FileTreeItem{}
	t.renderDirectoryTree(dirTree, 0)
}

// directoryNode represents a node in the directory tree
type directoryNode struct {
	name     string
	path     string
	files    []*taskmaster.FileChange
	children map[string]*directoryNode
}

// buildDirectoryTree constructs a hierarchical directory tree
func (t *FileTree) buildDirectoryTree(fileMap map[string]*taskmaster.FileChange) *directoryNode {
	root := &directoryNode{
		name:     "",
		path:     "",
		children: make(map[string]*directoryNode),
		files:    []*taskmaster.FileChange{},
	}
	
	for path, change := range fileMap {
		parts := strings.Split(path, "/")
		current := root
		
		// Build directory structure
		for i := 0; i < len(parts)-1; i++ {
			dirName := parts[i]
			if current.children[dirName] == nil {
				dirPath := strings.Join(parts[:i+1], "/")
				current.children[dirName] = &directoryNode{
					name:     dirName,
					path:     dirPath,
					children: make(map[string]*directoryNode),
					files:    []*taskmaster.FileChange{},
				}
			}
			current = current.children[dirName]
		}
		
		// Add file to its parent directory
		current.files = append(current.files, change)
	}
	
	return root
}

// renderDirectoryTree renders directory tree nodes recursively
func (t *FileTree) renderDirectoryTree(node *directoryNode, level int) {
	// Render subdirectories first
	for _, child := range node.children {
		isExpanded := t.expandedItems[child.path]
		
		// Add directory item
		dirItem := FileTreeItem{
			Type:      "directory",
			ID:        child.path,
			Path:      child.name,
			Level:     level,
			Collapsed: !isExpanded,
		}
		t.items = append(t.items, dirItem)
		
		// If expanded, render contents
		if isExpanded {
			// Render child directories
			t.renderDirectoryTree(child, level+1)
			
			// Render files in this directory
			for _, change := range child.files {
				fileItem := FileTreeItem{
					Type:   "file",
					Path:   strings.TrimPrefix(change.Path, child.path+"/"),
					Change: change,
					Level:  level + 1,
				}
				t.items = append(t.items, fileItem)
			}
		}
	}
	
	// Render files at root level (no directory)
	if level == 0 {
		for _, change := range node.files {
			fileItem := FileTreeItem{
				Type:   "file",
				Path:   change.Path,
				Change: change,
				Level:  0,
			}
			t.items = append(t.items, fileItem)
		}
	}
}

// buildTreeByTime builds tree grouped by time
func (t *FileTree) buildTreeByTime() {
	allChanges := t.dialog.tracker.GetAllTaskFileChanges()
	
	// Collect all changes with timestamps (applying filters)
	type timeEntry struct {
		change *taskmaster.FileChange
		taskID string
	}
	
	var entries []timeEntry
	for taskID, changes := range allChanges {
		for _, change := range changes {
			// Apply filters
			if !t.matchesFilters(&change, taskID) {
				continue
			}
			
			changeCopy := change
			entries = append(entries, timeEntry{
				change: &changeCopy,
				taskID: taskID,
			})
		}
	}
	
	// Sort by time (most recent first)
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].change.LastChanged.Before(entries[j].change.LastChanged) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	// Group by time periods
	type timePeriod struct {
		label   string
		entries []timeEntry
	}
	
	now := time.Now()
	periods := []timePeriod{
		{label: "Today", entries: []timeEntry{}},
		{label: "Yesterday", entries: []timeEntry{}},
		{label: "This Week", entries: []timeEntry{}},
		{label: "This Month", entries: []timeEntry{}},
		{label: "Older", entries: []timeEntry{}},
	}
	
	for _, entry := range entries {
		age := now.Sub(entry.change.LastChanged)
		
		if age < 24*time.Hour {
			periods[0].entries = append(periods[0].entries, entry)
		} else if age < 48*time.Hour {
			periods[1].entries = append(periods[1].entries, entry)
		} else if age < 7*24*time.Hour {
			periods[2].entries = append(periods[2].entries, entry)
		} else if age < 30*24*time.Hour {
			periods[3].entries = append(periods[3].entries, entry)
		} else {
			periods[4].entries = append(periods[4].entries, entry)
		}
	}
	
	// Build tree items
	t.items = []FileTreeItem{}
	for _, period := range periods {
		if len(period.entries) == 0 {
			continue
		}
		
		isExpanded := t.expandedItems[period.label]
		
		// Add period header
		periodItem := FileTreeItem{
			Type:      "task", // Reuse task type for period headers
			ID:        period.label,
			Path:      period.label,
			Level:     0,
			Collapsed: !isExpanded,
		}
		t.items = append(t.items, periodItem)
		
		// If expanded, add files
		if isExpanded {
			for _, entry := range period.entries {
				fileItem := FileTreeItem{
					Type:   "file",
					Path:   fmt.Sprintf("%s (Task %s)", entry.change.Path, entry.taskID),
					Change: entry.change,
					Level:  1,
				}
				t.items = append(t.items, fileItem)
			}
		}
	}
}

// Preview Methods

// LoadFile loads file content for preview
func (p *Preview) LoadFile(file string, gitService *git.GitService) error {
	p.file = file
	p.diffMode = false
	p.scrollPos = 0
	
	// Read file content from working directory
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file, err)
	}
	
	p.content = string(content)
	
	// Split content into lines and add line numbers
	rawLines := strings.Split(p.content, "\n")
	p.lines = make([]string, len(rawLines))
	
	// Calculate line number width for alignment
	maxLineNum := len(rawLines)
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))
	
	// Apply syntax highlighting and line numbers
	for i, line := range rawLines {
		lineNum := fmt.Sprintf("%*d", lineNumWidth, i+1)
		
		// Apply basic syntax highlighting
		styledLine := p.applySyntaxHighlighting(line, file)
		
		// Format with line number
		p.lines[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(lineNum) + " │ " + styledLine
	}
	
	return nil
}

// LoadDiff loads diff view for a file
func (p *Preview) LoadDiff(file string, gitService *git.GitService) error {
	p.file = file
	p.diffMode = true
	p.scrollPos = 0
	
	if gitService == nil {
		return fmt.Errorf("git service not available")
	}
	
	// Get diff for uncommitted changes (compare HEAD to working tree)
	ctx := context.Background()
	diff, err := gitService.GetFileDiff(ctx, file, "HEAD", "")
	if err != nil {
		return fmt.Errorf("failed to get diff for %s: %w", file, err)
	}
	
	p.content = diff
	
	// Split diff into lines and apply styling
	rawLines := strings.Split(diff, "\n")
	p.lines = make([]string, len(rawLines))
	
	for i, line := range rawLines {
		p.lines[i] = p.styleDiffLine(line)
	}
	
	return nil
}

// Render draws the preview panel
func (p *Preview) Render(width, height int) string {
	p.width = width
	p.height = height

	if p.file == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(width).
			Render("Select a file to preview")
	}

	var lines []string
	
	// Add header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Render(p.file)
	lines = append(lines, header)
	lines = append(lines, "")

	// Add content lines with scrolling
	endIdx := p.scrollPos + height - 2 // -2 for header
	if endIdx > len(p.lines) {
		endIdx = len(p.lines)
	}

	for i := p.scrollPos; i < endIdx; i++ {
		lines = append(lines, p.lines[i])
	}

	// Pad remaining lines
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// HandleKeyEvent processes keyboard input for the preview
func (p *Preview) HandleKeyEvent(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		p.scrollUp()
	case "down", "j":
		p.scrollDown()
	case "pageup":
		p.pageUp()
	case "pagedown":
		p.pageDown()
	case "home":
		p.scrollPos = 0
	case "end":
		p.scrollPos = len(p.lines) - p.height
		if p.scrollPos < 0 {
			p.scrollPos = 0
		}
	}
	return DialogResultNone, nil
}

// scrollUp scrolls preview up
func (p *Preview) scrollUp() {
	if p.scrollPos > 0 {
		p.scrollPos--
	}
}

// scrollDown scrolls preview down
func (p *Preview) scrollDown() {
	if p.scrollPos < len(p.lines)-p.height {
		p.scrollPos++
	}
}

// pageUp scrolls up by page
func (p *Preview) pageUp() {
	p.scrollPos -= p.height
	if p.scrollPos < 0 {
		p.scrollPos = 0
	}
}

// pageDown scrolls down by page
func (p *Preview) pageDown() {
	p.scrollPos += p.height
	maxScroll := len(p.lines) - p.height
	if p.scrollPos > maxScroll {
		p.scrollPos = maxScroll
	}
	if p.scrollPos < 0 {
		p.scrollPos = 0
	}
}

// applySyntaxHighlighting applies basic syntax highlighting based on file extension
func (p *Preview) applySyntaxHighlighting(line, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	style := lipgloss.NewStyle()
	
	// Apply color based on file type and content
	switch ext {
	case ".go":
		// Highlight Go keywords
		if strings.Contains(line, "func ") || strings.Contains(line, "package ") || 
		   strings.Contains(line, "import ") || strings.Contains(line, "type ") ||
		   strings.Contains(line, "const ") || strings.Contains(line, "var ") {
			style = style.Foreground(lipgloss.Color("99"))  // Purple for keywords
		} else if strings.HasPrefix(strings.TrimSpace(line), "//") {
			style = style.Foreground(lipgloss.Color("241")) // Gray for comments
		}
	case ".json", ".yaml", ".yml":
		// Highlight JSON/YAML keys
		if strings.Contains(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			style = style.Foreground(lipgloss.Color("33")) // Blue for keys
		} else if strings.HasPrefix(strings.TrimSpace(line), "#") {
			style = style.Foreground(lipgloss.Color("241")) // Gray for comments
		}
	case ".md", ".txt":
		// Highlight markdown headers
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			style = style.Foreground(lipgloss.Color("99")).Bold(true) // Purple bold for headers
		}
	case ".sh", ".bash":
		// Highlight shell comments and commands
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			style = style.Foreground(lipgloss.Color("241")) // Gray for comments
		}
	}
	
	// Highlight strings (basic detection)
	if strings.Contains(line, "\"") || strings.Contains(line, "'") {
		style = style.Foreground(lipgloss.Color("150")) // Green for strings
	}
	
	return style.Render(line)
}

// styleDiffLine styles a diff output line based on its prefix
func (p *Preview) styleDiffLine(line string) string {
	if line == "" {
		return line
	}
	
	style := lipgloss.NewStyle()
	
	// Style based on diff indicators
	switch {
	case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
		// File markers
		style = style.Foreground(lipgloss.Color("99")).Bold(true)
	case strings.HasPrefix(line, "@@"):
		// Hunk headers
		style = style.Foreground(lipgloss.Color("33")).Bold(true)
	case strings.HasPrefix(line, "+"):
		// Added lines
		style = style.Foreground(lipgloss.Color("150")).Background(lipgloss.Color("22"))
	case strings.HasPrefix(line, "-"):
		// Deleted lines
		style = style.Foreground(lipgloss.Color("203")).Background(lipgloss.Color("52"))
	case strings.HasPrefix(line, "diff "):
		// Diff command line
		style = style.Foreground(lipgloss.Color("241")).Italic(true)
	case strings.HasPrefix(line, "index "):
		// Index line
		style = style.Foreground(lipgloss.Color("241"))
	default:
		// Context lines (unchanged)
		style = style.Foreground(lipgloss.Color("252"))
	}
	
	return style.Render(line)
}

// Filter and Search Methods

// SetFilterTask sets the task filter and rebuilds the tree
func (d *FileChangesDialog) SetFilterTask(taskID string) {
	d.filterTask = taskID
	d.fileTree.buildTree()
}

// SetFilterStatus sets the status filter and rebuilds the tree
func (d *FileChangesDialog) SetFilterStatus(status string) {
	d.filterStatus = status
	d.fileTree.buildTree()
}

// SetSearchQuery sets the search query and rebuilds the tree
func (d *FileChangesDialog) SetSearchQuery(query string) {
	d.searchQuery = query
	d.fileTree.buildTree()
}

// ClearFilters clears all filters and rebuilds the tree
func (d *FileChangesDialog) ClearFilters() {
	d.filterTask = ""
	d.filterStatus = "all"
	d.searchQuery = ""
	d.fileTree.buildTree()
}

// ToggleStatusFilter cycles through status filter options
func (d *FileChangesDialog) ToggleStatusFilter() {
	switch d.filterStatus {
	case "all", "":
		d.filterStatus = "added"
	case "added":
		d.filterStatus = "modified"
	case "modified":
		d.filterStatus = "deleted"
	case "deleted":
		d.filterStatus = "all"
	}
	d.fileTree.buildTree()
}
