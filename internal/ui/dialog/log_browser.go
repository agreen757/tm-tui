package dialog

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TEA Message Types for panel integration

// FileSelectedMsg is sent when a file is selected in the File Browser
type FileSelectedMsg struct {
	FilePath string
}

// FileLoadedMsg is sent when a file has been loaded into the Log Viewer
type FileLoadedMsg struct {
	FilePath string
	Error    error
}

// LogBrowserDialog implements the main log browser dialog with two-panel layout
type LogBrowserDialog struct {
	BaseFocusableDialog

	// Panels (using actual models)
	fileBrowser *LogFileBrowserModel
	logViewer   *LogViewerPanel

	// State
	currentPath    string
	selectedFile   string
	focusedPanel   int               // 0=browser, 1=viewer
	statusMsg      string            // Status bar message for errors/info
	showHelp       bool              // Show help overlay
	
	// Size constraints
	showSizeWarning bool // Show warning if terminal is too small
	minWidth        int  // Minimum recommended width (80)
	minHeight       int  // Minimum recommended height (24)
	
	// Dependencies
	taskService    *taskmaster.Service
	style          *DialogStyle
	browserStyles  *LogBrowserStyles // Consistent styling for panels
}

// NewLogBrowserDialog creates a new LogBrowserDialog with initialized panels and state
func NewLogBrowserDialog(width, height int, taskService *taskmaster.Service) *LogBrowserDialog {
	// Calculate panel widths with specified proportions: 40% file browser, 60% log viewer
	browserWidth := (width * 40) / 100
	viewerWidth := width - browserWidth
	
	// Get or create default style
	style := DefaultDialogStyle()
	
	// Create consistent browser styles
	browserStyles := NewLogBrowserStyles()
	
	// Create the dialog with actual panel models
	dialog := &LogBrowserDialog{
		BaseFocusableDialog: NewBaseFocusableDialog("Log Browser", width, height, DialogKindCustom, 2),
		fileBrowser:         NewLogFileBrowserModel(browserWidth, height, taskService, ""),
		logViewer:           NewLogViewerPanel(viewerWidth, height, style),
		currentPath:         "",
		selectedFile:        "",
		focusedPanel:        0,
		statusMsg:           "",
		showSizeWarning:     false,
		minWidth:            80,
		minHeight:           24,
		taskService:         taskService,
		style:               style,
		browserStyles:       browserStyles,
	}
	
	// Set initial focus on file browser
	dialog.SetFocusedIndex(0)
	dialog.fileBrowser.SetFocused(true)
	
	return dialog
}

// Init initializes the dialog state and panels
func (d *LogBrowserDialog) Init() tea.Cmd {
	// Initialize all panels
	var cmds []tea.Cmd
	
	// Initialize panels (they all have Init() methods)
	if cmd := d.fileBrowser.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := d.logViewer.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// Update processes messages and updates the dialog state
func (d *LogBrowserDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	var cmds []tea.Cmd
	
	// First, update the focused panel with the message
	// This allows panels to handle keys (like Enter) before global dialog keys
	switch d.focusedPanel {
	case 0: // File Browser
		var model tea.Model
		var cmd tea.Cmd
		model, cmd = d.fileBrowser.Update(msg)
		d.fileBrowser = model.(*LogFileBrowserModel)
		
		// If the file browser returned a FileSelectedMsg command, execute it immediately
		if cmd != nil {
			// Execute the command to get the message
			resultMsg := cmd()
			if fileMsg, ok := resultMsg.(FileSelectedMsg); ok {
				// Handle the file selection immediately instead of batching
				d.selectedFile = fileMsg.FilePath
				
				// Update dialog title to show selected file name
				fileName := filepath.Base(fileMsg.FilePath)
				d.BaseFocusableDialog.TitleText = fmt.Sprintf("Log Browser - %s", fileName)
				
				// Load file content into Log Viewer
				err := d.logViewer.LoadFileContent(fileMsg.FilePath)
				if err != nil {
					d.statusMsg = fmt.Sprintf("Error loading file: %s", err.Error())
				} else {
					d.statusMsg = ""
				}
			} else {
				// Not a FileSelectedMsg, batch it normally
				cmds = append(cmds, cmd)
			}
		}
		
	case 1: // Log Viewer
		var model tea.Model
		var cmd tea.Cmd
		model, cmd = d.logViewer.Update(msg)
		d.logViewer = model.(*LogViewerPanel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	// Then handle specific message types
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update dialog dimensions from window resize
		d.SetRect(msg.Width, msg.Height, 0, 0)
		// Panel dimensions will be recalculated in View() based on new dialog size
		
	case FileSelectedMsg:
		// Handle file selection from File Browser
		d.selectedFile = msg.FilePath
		
		// Update dialog title to show selected file name
		fileName := filepath.Base(msg.FilePath)
		d.BaseFocusableDialog.TitleText = fmt.Sprintf("Log Browser - %s", fileName)
		
		// Load file content into Log Viewer
		err := d.logViewer.LoadFileContent(msg.FilePath)
		if err != nil {
			// Set error message in status bar
			d.statusMsg = fmt.Sprintf("Error loading file: %s", err.Error())
		} else {
			// Clear status message on success
			d.statusMsg = ""
		}
		
		return d, nil
	}
	
	// Batch all accumulated commands and return
	if len(cmds) > 0 {
		return d, tea.Batch(cmds...)
	}
	return d, nil
}

// View renders the dialog with all three panels
func (d *LogBrowserDialog) View() string {
	if d.width == 0 || d.height == 0 {
		return ""
	}

	// Calculate panel widths (40% file browser, 60% log viewer)
	browserWidth := (d.width * 40) / 100
	viewerWidth := d.width - browserWidth

	// Use dialog height for panels (subtract space for title and status bar)
	panelHeight := d.height - 4 // Reserve space for title and status bar

	// Calculate content area for each panel
	// wrapPanel adds: border (2 chars width, 2 lines height) + padding (2 chars width) + title (1 line)
	// Total overhead: 4 chars width, 3 lines height
	contentWidth := func(panelWidth int) int {
		return panelWidth - 4 // Border (2) + Padding (2)
	}
	contentHeight := panelHeight - 3 // Border (2) + Title (1)

	// Update panel sizes with content area dimensions
	// Panels no longer add their own borders, so they get the full content area
	d.fileBrowser.SetSize(contentWidth(browserWidth), contentHeight)
	d.logViewer.SetSize(contentWidth(viewerWidth), contentHeight)

	// Render panels using their View methods with consistent styling
	fileBrowserView := d.wrapPanel(d.fileBrowser.View(), " File Browser", 0, browserWidth, panelHeight)
	logViewerView := d.wrapPanel(d.logViewer.View(), " Content", 1, viewerWidth, panelHeight)

	// Combine panels horizontally using lipgloss
	panelsView := lipgloss.JoinHorizontal(lipgloss.Top, fileBrowserView, logViewerView)
	
	// Render status bar if there's a status message
	var baseView string
	if d.statusMsg != "" {
		statusStyle := d.browserStyles.ErrorStyle.Copy().
			Width(d.width).
			Padding(0, 1)
		statusBar := statusStyle.Render("⚠️  " + d.statusMsg)
		
		// Combine panels and status bar vertically
		baseView = lipgloss.JoinVertical(lipgloss.Left, panelsView, statusBar)
	} else {
		baseView = panelsView
	}
	
	// If help overlay is shown, render it on top
	if d.showHelp {
		helpOverlay := d.renderHelpOverlay()
		// Use Place to center the overlay over the base view
		return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center, helpOverlay, lipgloss.WithWhitespaceChars(" "))
	}
	
	return baseView
}

// HandleKey processes keyboard input for the dialog
func (d *LogBrowserDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// If help overlay is shown, handle help-specific keys first
	if d.showHelp {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("?", "esc"))):
			// Toggle help overlay off
			d.showHelp = false
			return DialogResultNone, nil
		}
		// Consume all other keys when help is shown
		return DialogResultNone, nil
	}
	
	// Handle global shortcuts that work regardless of focused panel
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("?"))):
		// Toggle help overlay
		d.showHelp = !d.showHelp
		return DialogResultNone, nil
		
	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		// Refresh file list and tag data
		return DialogResultNone, d.refresh()
		
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+f"))):
		// Ctrl+F to open dialog is handled at the UI level (not here)
		// But we include it in the help text
		return DialogResultNone, nil
	}
	
	// Use the base focusable handler for common keys like Esc
	result, cmd := d.HandleBaseFocusableKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle Tab/Shift+Tab for panel focus cycling
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		// Unfocus current panel before switching
		d.unfocusCurrentPanel()
		
		// Move to next panel
		d.focusedPanel = (d.focusedPanel + 1) % 2
		d.SetFocusedIndex(d.focusedPanel)
		
		// Focus new panel
		d.focusCurrentPanel()
		
		return DialogResultNone, nil
		
	case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
		// Unfocus current panel before switching
		d.unfocusCurrentPanel()
		
		// Move to previous panel
		d.focusedPanel = (d.focusedPanel - 1 + 2) % 2
		d.SetFocusedIndex(d.focusedPanel)
		
		// Focus new panel
		d.focusCurrentPanel()
		
		return DialogResultNone, nil
	}

	return DialogResultNone, nil
}

// focusCurrentPanel sets focus on the currently selected panel
func (d *LogBrowserDialog) focusCurrentPanel() {
	switch d.focusedPanel {
	case 0: // File Browser
		d.fileBrowser.SetFocused(true)
	case 1: // Log Viewer
		d.logViewer.SetFocused(true)
	}
}

// unfocusCurrentPanel removes focus from the currently selected panel
func (d *LogBrowserDialog) unfocusCurrentPanel() {
	switch d.focusedPanel {
	case 0: // File Browser
		d.fileBrowser.SetFocused(false)
	case 1: // Log Viewer
		d.logViewer.SetFocused(false)
	}
}

// CheckTerminalSize checks if terminal meets minimum size requirements
func (d *LogBrowserDialog) CheckTerminalSize(width, height int) {
	d.showSizeWarning = (width < d.minWidth) || (height < d.minHeight)
}

// IsTerminalSizeWarning returns true if terminal is below recommended size
func (d *LogBrowserDialog) IsTerminalSizeWarning() bool {
	return d.showSizeWarning
}

// refresh reloads file list and tag data
func (d *LogBrowserDialog) refresh() tea.Cmd {
	var cmds []tea.Cmd
	
	// Refresh file browser
	if d.fileBrowser != nil {
		if cmd := d.fileBrowser.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	// Update status message
	d.statusMsg = "Refreshed file list and tags"
	
	// Always return a command (even if empty) to signal the refresh was executed
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	
	// Return a no-op command to indicate refresh was triggered
	return func() tea.Msg {
		return nil
	}
}

// renderHelpOverlay renders the context-sensitive help overlay
func (d *LogBrowserDialog) renderHelpOverlay() string {
	if !d.showHelp {
		return ""
	}
	
	// Create help content based on focused panel
	var helpContent string
	
	// Global shortcuts (always shown)
	globalShortcuts := []string{
		"Global Shortcuts:",
		"  Ctrl+F   - Open log browser dialog",
		"  Esc      - Close dialog and return",
		"  Tab      - Move focus to next panel",
		"  Shift+Tab - Move focus to previous panel",
		"  ?        - Toggle this help overlay",
		"  r        - Refresh file list",
		"",
	}
	
	// Panel-specific shortcuts
	var panelShortcuts []string
	switch d.focusedPanel {
	case 0: // File Browser
		panelShortcuts = []string{
			"File Browser Shortcuts:",
			"  ↑/k      - Move selection up",
			"  ↓/j      - Move selection down",
			"  Enter    - Select file and display",
			"  Backspace/h - Navigate to parent directory",
			"  →/l      - Enter subdirectory",
			"  /        - Quick search/filter files",
			"  Home     - Jump to top of list",
			"  End      - Jump to bottom of list",
		}
	case 1: // Log Viewer
		panelShortcuts = []string{
			"Log Viewer Shortcuts:",
			"  ↑/k      - Scroll up one line",
			"  ↓/j      - Scroll down one line",
			"  PgUp/Ctrl+U - Scroll up one page",
			"  PgDn/Ctrl+D - Scroll down one page",
			"  Home/gg  - Jump to top of log",
			"  End/G    - Jump to bottom of log",
			"  n        - Toggle line numbers",
			"  w        - Toggle word wrap",
		}
	}
	
	// Combine all shortcuts
	allShortcuts := append(globalShortcuts, panelShortcuts...)
	helpContent = strings.Join(allShortcuts, "\n")
	
	// Calculate overlay dimensions
	overlayWidth := 60
	overlayHeight := len(allShortcuts) + 4 // Add padding
	
	// Center the overlay
	x := (d.width - overlayWidth) / 2
	y := (d.height - overlayHeight) / 2
	
	// Create overlay style
	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.style.FocusedBorderColor).
		Background(d.style.BackgroundColor).
		Foreground(d.style.TextColor).
		Width(overlayWidth).
		Height(overlayHeight).
		Padding(1, 2).
		Align(lipgloss.Left)
	
	// Add title
	titleStyle := lipgloss.NewStyle().
		Foreground(d.style.TitleColor).
		Bold(true)
	
	title := titleStyle.Render("Keyboard Shortcuts Help")
	content := title + "\n\n" + helpContent
	
	overlay := overlayStyle.Render(content)
	
	// Position the overlay (we'll use absolute positioning in View())
	_ = x
	_ = y
	
	return overlay
}

// SetMinimumSize sets the minimum recommended terminal size
func (d *LogBrowserDialog) SetMinimumSize(width, height int) {
	d.minWidth = width
	d.minHeight = height
}

// AdjustPanelSizes dynamically adjusts panel proportions based on available space
func (d *LogBrowserDialog) AdjustPanelSizes(width, height int) {
	// Check if we have enough space for all three panels
	if width < 60 {
		// Too narrow - reduce proportions
		// Browser: 40%, Selector: 20%, Viewer: 40%
		browserWidth := (width * 40) / 100
		tagWidth := (width * 20) / 100
		viewerWidth := width - browserWidth - tagWidth
		d.fileBrowser.SetSize(browserWidth, height)
		d.logViewer.SetSize(viewerWidth, height)
	} else {
		// Normal proportions: Browser: 35%, Selector: 25%, Viewer: 40%
		browserWidth := (width * 35) / 100
		tagWidth := (width * 25) / 100
		viewerWidth := width - browserWidth - tagWidth
		d.fileBrowser.SetSize(browserWidth, height)
		d.logViewer.SetSize(viewerWidth, height)
	}
}

// RecoverFromError attempts to recover dialog state after an error
func (d *LogBrowserDialog) RecoverFromError() {
	// Reset error state while preserving user data
	d.statusMsg = ""
	
	// Ensure panels are in a valid state
	if d.logViewer.HasError() {
		d.logViewer.ClearContent()
	}
	
	// Verify focus is still on a valid panel
	if d.focusedPanel < 0 || d.focusedPanel > 1 {
		d.focusedPanel = 0
	}
}

// GetCurrentSize returns current dialog dimensions
func (d *LogBrowserDialog) GetCurrentSize() (width, height int) {
	return d.width, d.height
}

// IsMinimumSizeMet returns true if current size meets minimum requirements
func (d *LogBrowserDialog) IsMinimumSizeMet() bool {
	return !d.showSizeWarning
}

// wrapPanel wraps a panel's content with consistent styling
func (d *LogBrowserDialog) wrapPanel(content, title string, panelIndex, width, height int) string {
	// Choose style based on focus
	var panelStyle lipgloss.Style
	if d.focusedPanel == panelIndex {
		panelStyle = d.browserStyles.FocusedPanelStyle
	} else {
		panelStyle = d.browserStyles.UnfocusedPanelStyle
	}
	
	// Apply dimensions
	panelStyle = panelStyle.
		Width(width - 2). // Account for border and padding
		Height(height - 2)
	
	// Add title with consistent styling
	titleBar := d.browserStyles.TitleStyle.Render(title)
	
	// Combine title and content
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		content,
	)
	
	return panelStyle.Render(fullContent)
}

// SetAccessibilityStatus sets a status message that can be read by screen readers
// This provides audio feedback for navigation and state changes
func (d *LogBrowserDialog) SetAccessibilityStatus(message string) {
	d.statusMsg = message
}

// GetPanelName returns the accessible name for a panel index
func (d *LogBrowserDialog) GetPanelName(panelIndex int) string {
	switch panelIndex {
	case 0:
		return "File Browser"
	case 1:
		return "Log Viewer"
	default:
		return "Unknown Panel"
	}
}

// AnnounceNavigation creates an accessibility announcement for panel navigation
func (d *LogBrowserDialog) AnnounceNavigation(fromPanel, toPanel int) string {
	return fmt.Sprintf("Navigated from %s to %s", d.GetPanelName(fromPanel), d.GetPanelName(toPanel))
}

// AnnounceFileSelection creates an accessibility announcement for file selection
func (d *LogBrowserDialog) AnnounceFileSelection(filename string) string {
	return fmt.Sprintf("Selected file: %s", filename)
}
