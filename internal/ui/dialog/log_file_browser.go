package dialog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileSortMode represents the file sorting strategy in the log browser
type FileSortMode int

const (
	// FileSortByDateNewest sorts files by modification time, newest first (default)
	FileSortByDateNewest FileSortMode = iota
	// FileSortByTaskID sorts files by task ID (legacy behavior)
	FileSortByTaskID
)

var (
	// ErrNotHandled indicates the key was not handled by HandleKey method
	ErrNotHandled = errors.New("key not handled")
)

// extractFilenameAndExtension splits a filename into base and extension.
// For example: "task-1.2.log" -> ("task-1.2", ".log"), "README" -> ("README", "")
// This is used by FileEntry to populate Filename and Extension fields for filtering.
func extractFilenameAndExtension(name string) (filename, extension string) {
	ext := filepath.Ext(name)
	if ext == "" {
		return name, ""
	}
	return strings.TrimSuffix(name, ext), ext
}

// FileEntry represents a file or directory entry in the file browser
type FileEntry struct {
	Name        string
	Path        string
	IsDir       bool
	Size        int64
	ModTime     time.Time
	DisplayName string
	Filename    string // Filename without extension
	Extension   string // File extension (e.g., ".log", ".md")
}

// FilterValue implements the bubbles/list.Item interface for filtering support.
// It returns a space-separated string containing the filename, display name, and extension.
//
// Example: For a file "task-1.2.log" with DisplayName "task-1.2.log", FilterValue returns:
//   "task-1.2 task-1.2.log .log"
//
// This enables case-insensitive filtering across all three fields using strings.Contains.
// The bubbles/list component automatically uses:
//   strings.Contains(strings.ToLower(f.FilterValue()), strings.ToLower(filter))
//
// This approach provides:
// - Performance: <100ms for filtering 1000 items (simple string search, no regex)
// - Flexibility: Match by filename, display name, or extension
// - Usability: Case-insensitive matching is intuitive for users
func (f FileEntry) FilterValue() string {
	return fmt.Sprintf("%s %s %s", f.Filename, f.DisplayName, f.Extension)
}

func (f FileEntry) Title() string {
	if f.IsDir {
		return "📁 " + f.DisplayName
	}
	return "📄 " + f.DisplayName
}

func (f FileEntry) Description() string {
	sizeStr := formatFileSize(f.Size)
	timeStr := f.ModTime.Format("2006-01-02 15:04")
	return fmt.Sprintf("%s • %s", sizeStr, timeStr)
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(size int64) string {
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

// LogFileBrowserModel represents the file browser panel
type LogFileBrowserModel struct {
	list         list.Model
	currentPath  string
	files        []FileEntry
	selectedFile string
	width        int
	height       int
	focused      bool
	taskService  *taskmaster.Service
	currentTag   string
	breadcrumbs  []string    // Track navigation path for breadcrumbs
	maxDepth     int         // Maximum depth to display (prevent UI overflow)
	sortMode     FileSortMode // Current sorting strategy (default: FileSortByDateNewest)

	// Caching fields
	dirCache      *LRUCache // LRU cache for directory listings (max 50 entries)
	metadataCache *LRUCache // LRU cache for file metadata

	// Search mode
	searchMode    bool                // Whether in search mode
	searchInput   textinput.Model     // Text input for search
	searchQuery   string              // Current search query
	searchResults []FileEntry         // Filtered results based on search
	visibleFiles  []FileEntry         // Files currently displayed (all or search results)

	// Filtering state (FR4.1-4.4, Phase 2)
	filteringEnabled bool // Whether filtering is enabled for this browser

	// Configuration options for reusability (Task 3.1)
	rootPath       string   // Optional: if set, use this instead of tag-based discovery
	fileExtensions []string // Optional: if set, only show files with these extensions
}

// NewLogFileBrowserModel creates a new file browser model
func NewLogFileBrowserModel(width, height int, taskService *taskmaster.Service, currentTag string) *LogFileBrowserModel {
	// Create list with default delegate
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, width, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Create search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search files..."
	searchInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	searchInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	browser := &LogFileBrowserModel{
		list:             l,
		currentPath:      "",
		files:            []FileEntry{},
		width:            width,
		height:           height,
		focused:          false,
		taskService:      taskService,
		currentTag:       currentTag,
		breadcrumbs:      []string{currentTag}, // Initialize with current tag
		maxDepth:         5,                    // Limit depth to prevent UI overflow
		sortMode:         FileSortByDateNewest, // Default: sort by date, newest first
		dirCache:         NewLRUCache(50),      // LRU cache for directory listings
		metadataCache:    NewLRUCache(100),     // LRU cache for file metadata
		searchMode:       false,                // Not in search mode initially
		searchInput:      searchInput,          // Initialize search input
		searchQuery:      "",                   // No search query initially
		searchResults:    []FileEntry{},        // No results initially
		visibleFiles:     []FileEntry{},        // No files initially
		filteringEnabled: true,                 // Filtering enabled by default
		rootPath:         "",                   // No override initially
		fileExtensions:   []string{},           // Use default extensions
	}

	// Load files from the current tag directory
	browser.loadFiles()

	return browser
}

// Init initializes the model
func (m *LogFileBrowserModel) Init() tea.Cmd {
	return nil
}


// Update handles messages
func (m *LogFileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle search mode first
	if m.searchMode {
		return m.handleSearchMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// "/" enters search mode
		if msg.String() == "/" {
			m.enterSearchMode()
			return m, textinput.Blink
		}

		// Intercept Tab/Shift+Tab/Esc to allow dialog to handle these keys
		switch msg.String() {
		case "tab", "shift+tab", "esc":
			return m, nil

		// Handle directory navigation keys
		case "enter", "l", "right":
			return m.handleEnter()
		case "backspace", "h", "left":
			return m.handleParent()
		}
	}

	// Pass other messages to the list for default handling
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// Update selected file based on list selection
	m.updateSelection()

	return m, cmd
}

// handleSearchMode processes keyboard input while in search mode
func (m *LogFileBrowserModel) handleSearchMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Exit search mode and restore full list
			m.exitSearchMode()
			return m, nil

		case "enter":
			// Confirm search - exit search input mode but keep filtered results
			m.searchMode = false
			m.searchInput.Blur()
			// Don't clear the search query or restore files - keep filtered results
			// The filtered results (m.visibleFiles) stay as they are
			return m, nil

		default:
			// Update search input and results as user types
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.searchQuery = m.searchInput.Value()
			m.updateSearchResults()
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// enterSearchMode activates search mode
func (m *LogFileBrowserModel) enterSearchMode() {
	m.searchMode = true
	m.searchInput.Focus()
	m.searchInput.SetValue("")
	m.searchQuery = ""
}

// exitSearchMode deactivates search mode
func (m *LogFileBrowserModel) exitSearchMode() {
	m.searchMode = false
	m.searchInput.Blur()
	m.searchInput.SetValue("")
	// Restore full file list
	m.visibleFiles = m.files
	m.updateList()
	// updateList() now handles selection sync
}

// updateSearchResults filters files based on search query
func (m *LogFileBrowserModel) updateSearchResults() {
	if m.searchQuery == "" {
		m.visibleFiles = m.files
		m.searchResults = nil
	} else {
		// Filter files by name (case-insensitive)
		query := strings.ToLower(m.searchQuery)
		m.searchResults = []FileEntry{}
		for _, file := range m.files {
			if strings.Contains(strings.ToLower(file.DisplayName), query) ||
				strings.Contains(strings.ToLower(file.Filename), query) {
				m.searchResults = append(m.searchResults, file)
			}
		}
		m.visibleFiles = m.searchResults
	}
	m.updateList()
}

// updateList updates the list display with visible files and syncs selection
func (m *LogFileBrowserModel) updateList() {
	// If visibleFiles is empty, use all files (happens on initial load)
	filesToDisplay := m.visibleFiles
	if len(filesToDisplay) == 0 && len(m.files) > 0 {
		filesToDisplay = m.files
	}

	items := make([]list.Item, len(filesToDisplay))
	for i, file := range filesToDisplay {
		items[i] = file
	}
	
	// SetItems resets selection to 0, so we set items first
	m.list.SetItems(items)
	
	// Ensure first item is selected if we have items
	if len(items) > 0 {
		// Force selection to index 0
		m.list.Select(0)
		// Immediately sync the selected file
		m.updateSelection()
	}
}

// handleEnter processes Enter key to enter a directory or select a file
func (m *LogFileBrowserModel) handleEnter() (tea.Model, tea.Cmd) {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return m, nil
	}

	entry, ok := selectedItem.(FileEntry)
	if !ok {
		return m, nil
	}

	if entry.IsDir {
		// Check if we're at max depth
		if m.getCurrentDepth() >= m.maxDepth {
			// Don't navigate deeper, but allow selection
			return m, nil
		}

		// Enter the directory
		m.currentPath = entry.Path
		m.pushBreadcrumb(entry.Name) // Add to breadcrumb trail
		m.loadFilesFromPath(m.currentPath)
		m.updateSelection()
		return m, nil
	} else {
		// File selected - update selectedFile and send message to parent dialog
		m.selectedFile = entry.Path

		// Send FileSelectedMsg to notify parent dialog to load the file content
		return m, func() tea.Msg {
			return FileSelectedMsg{FilePath: entry.Path}
		}
	}
}

// handleParent processes Backspace key to go to parent directory
func (m *LogFileBrowserModel) handleParent() (tea.Model, tea.Cmd) {
	if m.currentPath == "" {
		// Already at root, can't go up
		return m, nil
	}

	// Get parent directory
	parentPath := filepath.Dir(m.currentPath)

	// Don't go above the .taskmaster directory
	if !strings.Contains(parentPath, ".taskmaster") {
		return m, nil
	}

	m.currentPath = parentPath
	m.popBreadcrumb() // Remove from breadcrumb trail
	m.loadFilesFromPath(m.currentPath)
	m.updateSelection()

	return m, nil
}

// updateSelection updates the selectedFile based on current list selection
func (m *LogFileBrowserModel) updateSelection() {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		m.selectedFile = ""
		return
	}

	if entry, ok := selectedItem.(FileEntry); ok {
		if !entry.IsDir {
			m.selectedFile = entry.Path
		}
	}
}

// View renders the file browser
func (m *LogFileBrowserModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	if m.focused {
		style = style.BorderForeground(lipgloss.Color("39"))
	}

	// Show empty state if no files are loaded
	if len(m.files) == 0 && m.currentPath == "" {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Align(lipgloss.Center).
			Width(m.width).
			Height(m.height).
			PaddingTop(m.height / 3).
			Render("📁 No logs found\n\nCheck directory permissions\nor create log files first")
		return style.Render(emptyMsg)
	}

	// Build breadcrumb trail
	breadcrumbStr := m.getBreadcrumbString()
	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Padding(0, 1)
	breadcrumb := breadcrumbStyle.Render(breadcrumbStr)

	// Show depth warning if at max depth
	var depthWarning string
	if m.getCurrentDepth() >= m.maxDepth {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Padding(0, 1)
		depthWarning = warningStyle.Render("⚠️  Max depth reached")
	}

	// Combine breadcrumb and warning
	header := breadcrumb
	if depthWarning != "" {
		header = lipgloss.JoinHorizontal(lipgloss.Left, breadcrumb, " ", depthWarning)
	}

	// Build search input display if in search mode
	var content string
	if m.searchMode {
		searchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)
		searchPrompt := searchStyle.Render("Search: ")
		searchInput := m.searchInput.View()
		searchBar := lipgloss.JoinHorizontal(lipgloss.Left, searchPrompt, searchInput)
		
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			searchBar,
			m.list.View(),
		)
	} else {
		// Build the full view with breadcrumb header
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			m.list.View(),
		)
	}

	return style.Render(content)
}

// SetSize updates the dimensions of the file browser
func (m *LogFileBrowserModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SetFocused sets the focused state
func (m *LogFileBrowserModel) SetFocused(focused bool) {
	m.focused = focused
}

// getCurrentDepth returns the current navigation depth
func (m *LogFileBrowserModel) getCurrentDepth() int {
	return len(m.breadcrumbs) - 1 // Subtract 1 for the root (tag)
}

// pushBreadcrumb adds a directory name to the breadcrumb trail
func (m *LogFileBrowserModel) pushBreadcrumb(dirName string) {
	if m.getCurrentDepth() < m.maxDepth {
		m.breadcrumbs = append(m.breadcrumbs, dirName)
	}
}

// popBreadcrumb removes the last directory from the breadcrumb trail
func (m *LogFileBrowserModel) popBreadcrumb() {
	if len(m.breadcrumbs) > 1 { // Keep at least the root (tag)
		m.breadcrumbs = m.breadcrumbs[:len(m.breadcrumbs)-1]
	}
}

// resetBreadcrumbs resets the breadcrumb trail to just the tag
func (m *LogFileBrowserModel) resetBreadcrumbs() {
	m.breadcrumbs = []string{m.currentTag}
}

// getBreadcrumbString returns the formatted breadcrumb trail
func (m *LogFileBrowserModel) getBreadcrumbString() string {
	if len(m.breadcrumbs) == 0 {
		return m.currentTag
	}
	return strings.Join(m.breadcrumbs, " / ")
}

// loadFiles discovers and loads files from the appropriate directories
func (m *LogFileBrowserModel) loadFiles() {
	m.files = []FileEntry{}

	var targetPath string

	// Check if rootPath is configured (Task 3.1 - generalization)
	if m.rootPath != "" {
		// Use the configured root path directly
		targetPath = m.rootPath
	} else {
		// Original tag-based path discovery logic
		taskmasterDir := ".taskmaster"

		// Define search paths in priority order
		searchPaths := []string{
			filepath.Join(taskmasterDir, m.currentTag),
			filepath.Join(taskmasterDir, m.currentTag, "logs"),
			filepath.Join(taskmasterDir, "logs", m.currentTag),
			filepath.Join(taskmasterDir, "logs"),
		}

		// Try each path and use the first one that exists and is readable
		for _, path := range searchPaths {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if !info.IsDir() {
				continue
			}

			// Check if we have read permission
			if err := canReadDir(path); err != nil {
				continue
			}

			targetPath = path
			break
		}

		if targetPath == "" {
			// No valid directory found - try fallback to .taskmaster itself
			if info, err := os.Stat(taskmasterDir); err == nil && info.IsDir() {
				if canReadDir(taskmasterDir) == nil {
					targetPath = taskmasterDir
				}
			}

			// Still no path - show empty state
			if targetPath == "" {
				m.currentPath = ""
				m.files = []FileEntry{}
				m.visibleFiles = []FileEntry{}  // Sync visibleFiles
				m.updateList()
				return
			}
		}
	}

	m.currentPath = targetPath

	// Check cache first
	if cached, found := m.dirCache.Get(targetPath); found {
		if entries, ok := cached.([]FileEntry); ok {
			m.files = entries
			m.visibleFiles = m.files  // Initialize visibleFiles when files are loaded
			m.updateList()
			return
		}
	}

	// Discover files in the target directory
	entries, err := m.discoverFilesWithConfig(targetPath, m.sortMode)
	if err != nil {
		// Handle error gracefully - just show empty list
		// In a real implementation, we could store error state for the UI
		m.files = []FileEntry{}
		m.visibleFiles = []FileEntry{}  // Sync visibleFiles
		m.updateList()
		return
	}

	// Cache the result
	m.dirCache.Put(targetPath, entries)

	m.files = entries
	m.visibleFiles = m.files  // Initialize visibleFiles when files are loaded
	m.updateList()
}

// loadFilesFromPath loads files from a specific directory path
func (m *LogFileBrowserModel) loadFilesFromPath(path string) {
	m.files = []FileEntry{}

	// Validate that path is within .taskmaster directory
	if !strings.Contains(path, ".taskmaster") {
		m.currentPath = ""
		m.files = []FileEntry{}
		m.visibleFiles = []FileEntry{}  // Sync visibleFiles
		m.updateList()
		return
	}

	// Check cache first
	if cached, found := m.dirCache.Get(path); found {
		if entries, ok := cached.([]FileEntry); ok {
			m.files = entries
			m.currentPath = path
			m.visibleFiles = m.files  // Initialize visibleFiles when files are loaded
			m.updateList()
			return
		}
	}

	// Discover files in the specified directory
	entries, err := discoverFiles(path, m.sortMode)
	if err != nil {
		// Handle error gracefully - just show empty list
		m.files = []FileEntry{}
		m.visibleFiles = []FileEntry{}  // Sync visibleFiles
		m.currentPath = ""
		m.updateList()
		return
	}

	// Cache the result
	m.dirCache.Put(path, entries)

	m.files = entries
	m.currentPath = path
	m.visibleFiles = m.files  // Initialize visibleFiles when files are loaded
	m.updateList()
}

// canReadDir checks if a directory is readable
func canReadDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	// Successfully read directory, can read it
	_ = entries
	return nil
}

// discoverFilesWithConfig discovers files using this model's configuration
// Task 3.1: Supports custom file extensions from m.fileExtensions
func (m *LogFileBrowserModel) discoverFilesWithConfig(rootPath string, sortMode FileSortMode) ([]FileEntry, error) {
	var entries []FileEntry

	// Read directory entries
	dirEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		name := entry.Name()

		// Filter out hidden files and unwanted directories
		if shouldSkipEntry(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't read
		}

		fullPath := filepath.Join(rootPath, name)
		filename, extension := extractFilenameAndExtension(name)

		if entry.IsDir() {
			// Include directory
			entries = append(entries, FileEntry{
				Name:        name,
				Path:        fullPath,
				IsDir:       true,
				Size:        0,
				ModTime:     info.ModTime(),
				DisplayName: name,
				Filename:    filename,
				Extension:   extension,
			})
		} else {
			// Check if file matches configured extensions or default supported files
			if m.isSupportedFileWithConfig(name) {
				entries = append(entries, FileEntry{
					Name:        name,
					Path:        fullPath,
					IsDir:       false,
					Size:        info.Size(),
					ModTime:     info.ModTime(),
					DisplayName: name,
					Filename:    filename,
					Extension:   extension,
				})
			}
		}
	}

	// Sort entries using the specified sort mode
	sortFileEntries(entries, sortMode)

	return entries, nil
}

// isSupportedFileWithConfig checks if a file is supported using this model's configuration
// Task 3.1: If fileExtensions is configured, uses those; otherwise uses defaults
func (m *LogFileBrowserModel) isSupportedFileWithConfig(name string) bool {
	// If fileExtensions is configured, use it exclusively
	if len(m.fileExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(name))
		for _, supported := range m.fileExtensions {
			// Normalize the extension (ensure it has a dot)
			supportedExt := supported
			if !strings.HasPrefix(supportedExt, ".") {
				supportedExt = "." + supportedExt
			}
			if ext == supportedExt {
				return true
			}
		}
		return false
	}
	// Fall back to default supported files
	return isSupportedFile(name)
}

// discoverFiles recursively discovers files in a directory with filtering
func discoverFiles(rootPath string, sortMode FileSortMode) ([]FileEntry, error) {
	var entries []FileEntry

	// Read directory entries
	dirEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		name := entry.Name()

		// Filter out hidden files and unwanted directories
		if shouldSkipEntry(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't read
		}

		fullPath := filepath.Join(rootPath, name)
		filename, extension := extractFilenameAndExtension(name)

		if entry.IsDir() {
			// Include directory
			entries = append(entries, FileEntry{
				Name:        name,
				Path:        fullPath,
				IsDir:       true,
				Size:        0,
				ModTime:     info.ModTime(),
				DisplayName: name,
				Filename:    filename,
				Extension:   extension,
			})
		} else {
			// Check if file has a supported extension
			if isSupportedFile(name) {
				entries = append(entries, FileEntry{
					Name:        name,
					Path:        fullPath,
					IsDir:       false,
					Size:        info.Size(),
					ModTime:     info.ModTime(),
					DisplayName: name,
					Filename:    filename,
					Extension:   extension,
				})
			}
		}
	}

	// Sort entries using the specified sort mode
	sortFileEntries(entries, sortMode)

	return entries, nil
}

// shouldSkipEntry checks if a file or directory should be skipped
func shouldSkipEntry(name string) bool {
	// Skip hidden files (starting with .)
	if strings.HasPrefix(name, ".") {
		return true
	}

	// Skip unwanted directories
	skipDirs := []string{"node_modules", "dist"}
	for _, dir := range skipDirs {
		if name == dir {
			return true
		}
	}

	return false
}

// isSupportedFile checks if a file has a supported extension
func isSupportedFile(name string) bool {
	// Supported extensions
	supportedExts := []string{".log", ".md", ".txt"}

	ext := strings.ToLower(filepath.Ext(name))

	// Files with no extension are supported
	if ext == "" {
		return true
	}

	// Check if extension is in the supported list
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}

	return false
}

// sortFileEntries sorts file entries based on the specified sort mode
// Directories always come before files
func sortFileEntries(entries []FileEntry, sortMode FileSortMode) {
	sort.Slice(entries, func(i, j int) bool {
		// Directories come before files
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}

		switch sortMode {
		case FileSortByDateNewest:
			// Sort by modification time, newest first (descending)
			if entries[i].ModTime != entries[j].ModTime {
				return entries[i].ModTime.After(entries[j].ModTime)
			}
			// If times are equal, sort alphabetically by name
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)

		case FileSortByTaskID:
			// Legacy behavior: sort by task ID, then alphabetically
			// Try to extract task ID from filename (e.g., "1.2.log" -> "1.2")
			taskIDi := extractTaskID(entries[i].Name)
			taskIDj := extractTaskID(entries[j].Name)

			// If both have task IDs, sort by task ID
			if taskIDi != "" && taskIDj != "" {
				return compareTaskIDs(taskIDi, taskIDj)
			}

			// If only one has a task ID, it comes first
			if taskIDi != "" {
				return true
			}
			if taskIDj != "" {
				return false
			}

			// Otherwise, sort alphabetically
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)

		default:
			// Default to date sorting (newest first)
			if entries[i].ModTime != entries[j].ModTime {
				return entries[i].ModTime.After(entries[j].ModTime)
			}
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		}
	})
}

// extractTaskID attempts to extract a task ID from a filename
// Examples: "1.2.log" -> "1.2", "task-3.1.md" -> "3.1", "notes.txt" -> ""
func extractTaskID(filename string) string {
	// Remove only the known file extension (not arbitrary dots)
	var name string
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExts := []string{".log", ".md", ".txt"}

	hasKnownExt := false
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			hasKnownExt = true
			break
		}
	}

	if hasKnownExt {
		name = strings.TrimSuffix(filename, ext)
	} else {
		name = filename
	}

	// Clean up any trailing dots from malformed filenames
	name = strings.TrimRight(name, ".")

	// If the entire name (after removing extension) is a task ID, return it
	if isTaskIDPattern(name) {
		return name
	}

	// Look for patterns like "1.2", "3.1.4", etc. by splitting on common delimiters
	// First, split by hyphens, underscores, and spaces
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	// Check each part to see if it's a task ID
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Clean up trailing dots
		part = strings.TrimRight(part, ".")

		if isTaskIDPattern(part) {
			return part
		}

		// Also try splitting by dots and looking for task ID patterns
		// This handles cases like "test.1.2" where "1.2" should be extracted
		dotParts := strings.Split(part, ".")
		for i := 0; i < len(dotParts); i++ {
			// Try to build a task ID from consecutive dot-separated numeric parts
			for j := i + 1; j <= len(dotParts); j++ {
				candidate := strings.Join(dotParts[i:j], ".")
				if isTaskIDPattern(candidate) {
					return candidate
				}
			}
		}
	}

	return ""
}

// isTaskIDPattern checks if a string matches the task ID pattern (e.g., "1.2", "3.1.4")
func isTaskIDPattern(s string) bool {
	if s == "" {
		return false
	}

	// Must not start or end with a dot
	if strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}

	// Must contain at least one digit and one dot
	hasDigit := false
	hasDot := false

	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		} else if ch == '.' {
			hasDot = true
		} else {
			return false // Contains invalid characters
		}
	}

	// Must have at least one dot and one digit
	if !hasDigit || !hasDot {
		return false
	}

	// Ensure no consecutive dots
	if strings.Contains(s, "..") {
		return false
	}

	// Verify that each dot-separated part is a valid number
	parts := strings.Split(s, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check if part is numeric
		var num int
		if _, err := fmt.Sscanf(part, "%d", &num); err != nil {
			return false
		}
	}

	return true
}

// compareTaskIDs compares two task IDs (e.g., "1.2" vs "1.10" or "2.1" vs "1.2.3")
func compareTaskIDs(id1, id2 string) bool {
	parts1 := strings.Split(id1, ".")
	parts2 := strings.Split(id2, ".")

	// Compare each part numerically
	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		// Parse as integers for numeric comparison
		var num1, num2 int
		fmt.Sscanf(parts1[i], "%d", &num1)
		fmt.Sscanf(parts2[i], "%d", &num2)

		if num1 != num2 {
			return num1 < num2
		}
	}

	// If all compared parts are equal, shorter ID comes first
	return len(parts1) < len(parts2)
}


// GetSelectedFile returns the currently selected file path
func (m *LogFileBrowserModel) GetSelectedFile() string {
	if len(m.files) == 0 {
		return ""
	}

	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return ""
	}

	if entry, ok := selectedItem.(FileEntry); ok {
		return entry.Path
	}

	return ""
}

// Reload refreshes the file list
func (m *LogFileBrowserModel) Reload() {
	m.loadFiles()
}

// SetPath changes the current directory path and reloads files
// This is used for tag switching and navigation
func (m *LogFileBrowserModel) SetPath(path string) {
	m.loadFilesFromPath(path)
}

// SetTag updates the current tag context and reloads files for that tag
func (m *LogFileBrowserModel) SetTag(tag string) {
	m.currentTag = tag
	m.resetBreadcrumbs() // Reset breadcrumbs when changing tags
	m.loadFiles()
}

// GetCurrentPath returns the current directory path
func (m *LogFileBrowserModel) GetCurrentPath() string {
	return m.currentPath
}

// GetFileMetadata returns file metadata from cache or loads it
// This is useful for hover previews without full file loading
func (m *LogFileBrowserModel) GetFileMetadata(path string) (os.FileInfo, error) {
	// Check cache first
	if cached, found := m.metadataCache.Get(path); found {
		if info, ok := cached.(os.FileInfo); ok {
			return info, nil
		}
	}

	// Load from disk
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Cache the metadata
	m.metadataCache.Put(path, info)

	return info, nil
}

// ClearCache clears both directory and metadata caches
func (m *LogFileBrowserModel) ClearCache() {
	m.dirCache.Clear()
	m.metadataCache.Clear()
}

// GetCacheStats returns cache statistics for monitoring
func (m *LogFileBrowserModel) GetCacheStats() (dirCacheSize, metadataCacheSize int) {
	return m.dirCache.Len(), m.metadataCache.Len()
}

// FilterableComponent interface implementation (FR4.1-4.4, Phase 2)

// EnableFiltering enables or disables the filtering capability for this component
func (m *LogFileBrowserModel) EnableFiltering(enabled bool) {
	m.filteringEnabled = enabled
	// If disabling filtering while user is typing in filter, we could exit here
	// but the list component handles this internally
}

// SetFilterValue sets the current filter string value
func (m *LogFileBrowserModel) SetFilterValue(value string) {
	// This is handled by the underlying list model's filter
	// We maintain this for interface compatibility
}

// GetFilterValue returns the current filter string value
func (m *LogFileBrowserModel) GetFilterValue() string {
	// Return empty string as the list model handles filtering
	return ""
}

// IsFiltering returns whether the component is currently in filtering mode
// Uses the list's internal filter state as the authoritative source
func (m *LogFileBrowserModel) IsFiltering() bool {
	return m.list.SettingFilter()
}

// EnterFilterMode transitions the component into filtering mode
// Simulates a "/" key press to activate filtering
func (m *LogFileBrowserModel) EnterFilterMode() {
	if m.filteringEnabled {
		// Simulate "/" key to enter filter mode
		slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		m.list, _ = m.list.Update(slashMsg)
	}
}

// ExitFilterMode transitions the component out of filtering mode
// Simulates an "esc" key to exit filter mode
func (m *LogFileBrowserModel) ExitFilterMode() {
	if m.list.SettingFilter() {
		escMsg := tea.KeyMsg{Type: tea.KeyEsc}
		m.list, _ = m.list.Update(escMsg)
	}
}

// SetRootPath configures a custom root directory for file browsing (Task 3.2)
// This overrides the default tag-based path discovery
// The path is used directly without tag-based fallback logic
func (m *LogFileBrowserModel) SetRootPath(path string) {
	m.rootPath = path
	// Clear cache to ensure fresh data from new root path
	m.dirCache.Clear()
	// Reload files from the new root path
	m.loadFiles()
}

// SetFileExtensions configures which file extensions to display (Task 3.2)
// If extensions is empty, default file extensions are used (log, md, txt, none)
// Extension strings should include or exclude the leading dot (both ".txt" and "txt" are supported)
func (m *LogFileBrowserModel) SetFileExtensions(extensions []string) {
	m.fileExtensions = extensions
	// Clear cache to ensure fresh data with new filtering
	m.dirCache.Clear()
	// Reload files with new extension filtering
	m.loadFiles()
}

// GetRootPath returns the currently configured root path
func (m *LogFileBrowserModel) GetRootPath() string {
	return m.rootPath
}

// GetFileExtensions returns the currently configured file extensions
func (m *LogFileBrowserModel) GetFileExtensions() []string {
	return m.fileExtensions
}

