package dialog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TagEntry represents a tag in the log tag selector
type LogTagEntry struct {
	Name      string
	LogCount  int
	LastMod   time.Time
	IsActive  bool
	IsSpecial bool // For 'logs' and 'archive' tags
}

// Implement list.Item interface for LogTagEntry
func (t LogTagEntry) FilterValue() string {
	return strings.ToLower(t.Name)
}

func (t LogTagEntry) Title() string {
	indicator := "  "
	if t.IsActive {
		indicator = "● "
	}
	return indicator + t.Name
}

func (t LogTagEntry) Description() string {
	timeStr := t.LastMod.Format("Jan 02 15:04")
	return fmt.Sprintf("%d log%s • %s", t.LogCount, pluralizeLogCount(t.LogCount), timeStr)
}

// pluralizeLogCount returns "s" for counts != 1
func pluralizeLogCount(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// LogTagSelectorModel represents the tag selector dialog for viewing logs from different tags
type LogTagSelectorModel struct {
	*BaseFocusableDialog
	list          list.Model
	tags          []LogTagEntry
	currentTag    string
	width         int
	height        int
	focused       bool
	taskService   *taskmaster.Service
	taskmasterDir string
	onTagSelected func(tagName string) tea.Cmd
}

// NewLogTagSelectorModel creates a new log tag selector model
func NewLogTagSelectorModel(width, height int, taskService *taskmaster.Service, currentTag string) *LogTagSelectorModel {
	// Create delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62"))

	// Create list model
	l := list.New([]list.Item{}, delegate, width-4, height-8)
	l.Title = "Available Tags"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Find taskmaster directory
	taskmasterDir := ""
	if taskService.IsAvailable() {
		taskmasterDir = filepath.Join(taskService.RootDir, ".taskmaster")
	}

	baseBfd := NewBaseFocusableDialog("Select Log Tag", width, height, DialogKindList, 0)
	selector := &LogTagSelectorModel{
		BaseFocusableDialog: &baseBfd,
		list:                l,
		currentTag:          currentTag,
		width:               width,
		height:              height,
		focused:             false,
		taskService:         taskService,
		taskmasterDir:       taskmasterDir,
	}

	// Set footer hints
	selector.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	// Discover and load tags
	selector.discoverTags()

	return selector
}

// discoverTags discovers all available tags from the .taskmaster directory
func (m *LogTagSelectorModel) discoverTags() {
	m.tags = []LogTagEntry{}

	if m.taskmasterDir == "" {
		log.Printf("[LogTagSelector] No .taskmaster directory available")
		m.updateListItems()
		return
	}

	// Check if taskmasterDir exists and is readable
	info, err := os.Stat(m.taskmasterDir)
	if err != nil {
		log.Printf("[LogTagSelector] Cannot access .taskmaster directory: %v", err)
		m.updateListItems()
		return
	}

	if !info.IsDir() {
		log.Printf("[LogTagSelector] .taskmaster is not a directory")
		m.updateListItems()
		return
	}

	// Add special 'logs' tag (default untagged logs)
	logsDir := filepath.Join(m.taskmasterDir, "logs")
	logCount, lastMod := m.calculateTagMetadata(logsDir)
	m.tags = append(m.tags, LogTagEntry{
		Name:      "logs",
		LogCount:  logCount,
		LastMod:   lastMod,
		IsActive:  m.currentTag == "logs",
		IsSpecial: true,
	})

	// Read .taskmaster directory to find subdirectories
	entries, err := os.ReadDir(m.taskmasterDir)
	if err != nil {
		log.Printf("[LogTagSelector] Failed to read .taskmaster directory: %v", err)
		m.updateListItems()
		return
	}

	// Excluded directories that shouldn't be treated as tags
	excludedDirs := map[string]bool{
		"tasks":   true,
		"docs":    true,
		"reports": true,
		"memory":  true,
		"logs":    true, // Already added above
	}

	// Excluded files
	excludedFiles := map[string]bool{
		"config.json": true,
	}

	// Collect all subdirectories as potential tags
	var tagDirs []os.DirEntry

	for _, entry := range entries {
		if !entry.IsDir() {
			// Skip files
			if excludedFiles[entry.Name()] {
				continue
			}
			continue
		}

		// Skip excluded directories
		if excludedDirs[entry.Name()] {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Try to check if directory is readable
		tagPath := filepath.Join(m.taskmasterDir, entry.Name())
		if _, err := os.ReadDir(tagPath); err != nil {
			// Skip directories we can't read
			log.Printf("[LogTagSelector] Cannot read tag directory %s: %v", tagPath, err)
			continue
		}

		tagDirs = append(tagDirs, entry)
	}

	// Sort tag directories by name
	sort.Slice(tagDirs, func(i, j int) bool {
		return tagDirs[i].Name() < tagDirs[j].Name()
	})

	// Convert directories to tags
	for _, entry := range tagDirs {
		tagPath := filepath.Join(m.taskmasterDir, entry.Name())
		logCount, lastMod := m.calculateTagMetadata(tagPath)

		m.tags = append(m.tags, LogTagEntry{
			Name:      entry.Name(),
			LogCount:  logCount,
			LastMod:   lastMod,
			IsActive:  m.currentTag == entry.Name(),
			IsSpecial: false,
		})
	}

	// Add 'archive' tag (special archived logs)
	archiveDir := filepath.Join(m.taskmasterDir, "archive")
	if _, err := os.Stat(archiveDir); err == nil {
		logCount, lastMod := m.calculateTagMetadata(archiveDir)
		m.tags = append(m.tags, LogTagEntry{
			Name:      "archive",
			LogCount:  logCount,
			LastMod:   lastMod,
			IsActive:  m.currentTag == "archive",
			IsSpecial: true,
		})
	}

	m.updateListItems()
}

// calculateTagMetadata calculates log count and last modified timestamp for a tag directory
func (m *LogTagSelectorModel) calculateTagMetadata(tagPath string) (int, time.Time) {
	logCount := 0
	var lastMod time.Time

	entries, err := os.ReadDir(tagPath)
	if err != nil {
		// Directory doesn't exist or can't be read
		return 0, time.Now()
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		// Count files with .log extension
		if strings.HasSuffix(entry.Name(), ".log") {
			logCount++

			// Get file info for last modified time
			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime()
			if lastMod.IsZero() || modTime.After(lastMod) {
				lastMod = modTime
			}
		}
	}

	if lastMod.IsZero() {
		lastMod = time.Now()
	}

	return logCount, lastMod
}

// updateListItems updates the list model with tag entries
func (m *LogTagSelectorModel) updateListItems() {
	items := make([]list.Item, 0, len(m.tags))
	for _, tag := range m.tags {
		items = append(items, tag)
	}

	m.list.SetItems(items)
}

// Init initializes the model
func (m *LogTagSelectorModel) Init() tea.Cmd {
	return nil
}

// Update processes messages
func (m *LogTagSelectorModel) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.Center(msg.Width, msg.Height)
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)

	case tea.KeyMsg:
		log.Printf("[LogTagSelector.Update] KeyMsg received: %s (should not process this - DialogManager handles it)", msg.String())
	}

	return m, nil
}

// View renders the dialog
func (m *LogTagSelectorModel) View() string {
	contentWidth := m.width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Show empty state if no tags are available
	if len(m.tags) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Align(lipgloss.Center).
			Width(m.width - 4).
			Height(m.height - 4).
			PaddingTop(5).
			Render("🏷️  No tags available\n\nCreate a .taskmaster directory\nwith log files to see tags")
		return m.RenderBorder(emptyMsg)
	}

	listContent := m.list.View()
	return m.RenderBorder(listContent)
}

// HandleKey processes keyboard input
func (m *LogTagSelectorModel) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	log.Printf("[LogTagSelector.HandleKey] Key pressed: %s", msg.String())

	// Check if the list model is in filtering mode
	// If so, delegate the key to the list model for search handling
	if m.list.FilterState() == list.Filtering {
		switch msg.String() {
		case "esc":
			// Allow Escape to cancel filter
			m.list.ResetFilter()
			return DialogResultNone, nil
		default:
			// Let the list handle search input
			// The list model handles search internally with "/" key
			return DialogResultNone, nil
		}
	}

	switch msg.String() {
	case "/":
		// Activate search mode in the list
		// The list model handles "/" automatically when filtering is enabled
		return DialogResultNone, nil

	case "up", "k":
		m.list.CursorUp()
		return DialogResultNone, nil

	case "down", "j":
		m.list.CursorDown()
		return DialogResultNone, nil

	case "pgup":
		m.list.CursorUp()
		m.list.CursorUp()
		m.list.CursorUp()
		return DialogResultNone, nil

	case "pgdown":
		m.list.CursorDown()
		m.list.CursorDown()
		m.list.CursorDown()
		return DialogResultNone, nil

	case "home":
		// Jump to current active tag
		for i, tag := range m.tags {
			if tag.IsActive {
				m.list.Select(i)
				return DialogResultNone, nil
			}
		}
		// If no active tag found, go to first
		m.list.Select(0)
		return DialogResultNone, nil

	case "enter":
		if len(m.tags) == 0 {
			return DialogResultCancel, nil
		}

		selectedIdx := m.list.Index()
		if selectedIdx >= 0 && selectedIdx < len(m.tags) {
			selectedTag := m.tags[selectedIdx]
			log.Printf("[LogTagSelector.HandleKey] Selected tag: %s", selectedTag.Name)

			// Call the callback if set
			if m.onTagSelected != nil {
				return DialogResultConfirm, m.onTagSelected(selectedTag.Name)
			}

			return DialogResultConfirm, nil
		}
		return DialogResultCancel, nil

	case "esc":
		log.Printf("[LogTagSelector.HandleKey] ESC: cancelling dialog")
		return DialogResultCancel, nil
	}

	return DialogResultNone, nil
}

// GetSelectedTag returns the currently selected tag name
func (m *LogTagSelectorModel) GetSelectedTag() string {
	selectedIdx := m.list.Index()
	if selectedIdx >= 0 && selectedIdx < len(m.tags) {
		return m.tags[selectedIdx].Name
	}
	return ""
}

// DialogResultValue implements DialogResultProvider interface
func (m *LogTagSelectorModel) DialogResultValue() (interface{}, error) {
	return m.GetSelectedTag(), nil
}

// SetRect sets the dialog's position and size
func (m *LogTagSelectorModel) SetRect(width, height, x, y int) {
	m.width = width
	m.height = height
	m.BaseFocusableDialog.SetRect(width, height, x, y)
	m.list.SetWidth(width - 4)
	m.list.SetHeight(height - 8)
}

// SetOnTagSelected sets the callback function when a tag is selected
func (m *LogTagSelectorModel) SetOnTagSelected(callback func(tagName string) tea.Cmd) {
	m.onTagSelected = callback
}

// RefreshTags refreshes the tag list (useful after file system changes)
func (m *LogTagSelectorModel) RefreshTags() {
	m.discoverTags()
}
