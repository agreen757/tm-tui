package dialog

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// LogBrowserStyles contains consistent styling for the Log Browser dialog
type LogBrowserStyles struct {
	// Panel styles
	PanelStyle         lipgloss.Style
	FocusedPanelStyle  lipgloss.Style
	UnfocusedPanelStyle lipgloss.Style
	
	// Text styles
	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style
	HintStyle     lipgloss.Style
	
	// State styles
	EmptyStateStyle   lipgloss.Style
	LoadingStyle      lipgloss.Style
	ErrorStyle        lipgloss.Style
	
	// Line highlight style
	CurrentLineStyle lipgloss.Style
	
	// Spinner for loading states
	Spinner spinner.Model
	
	// Theme configuration
	HighContrast bool
}

// ThemeConfig holds color configuration for different themes
type ThemeConfig struct {
	FocusedBorder   string
	UnfocusedBorder string
	Title           string
	Subtitle        string
	Hint            string
	Empty           string
	Loading         string
	Error           string
	LineHighlight   string
	LineText        string
	SpinnerColor    string
}

// GetDefaultTheme returns the standard Dracula-inspired theme
func GetDefaultTheme() ThemeConfig {
	return ThemeConfig{
		FocusedBorder:   "#8be9fd", // Bright cyan
		UnfocusedBorder: "#6272a4", // Muted gray
		Title:           "#f8f8f2", // White
		Subtitle:        "#bd93f9", // Purple
		Hint:            "#6272a4", // Gray
		Empty:           "#6272a4", // Gray
		Loading:         "#8be9fd", // Cyan
		Error:           "#ff5555", // Red
		LineHighlight:   "#44475a", // Subtle background
		LineText:        "#f8f8f2", // White
		SpinnerColor:    "#ff79c6", // Pink
	}
}

// GetHighContrastTheme returns a high-contrast theme for accessibility
func GetHighContrastTheme() ThemeConfig {
	return ThemeConfig{
		FocusedBorder:   "#FFFF00", // Bright yellow
		UnfocusedBorder: "#FFFFFF", // White
		Title:           "#FFFFFF", // White
		Subtitle:        "#FFFFFF", // White
		Hint:            "#CCCCCC", // Light gray
		Empty:           "#FFFFFF", // White
		Loading:         "#00FFFF", // Bright cyan
		Error:           "#FF0000", // Pure red
		LineHighlight:   "#FFFF00", // Yellow background
		LineText:        "#000000", // Black text
		SpinnerColor:    "#00FF00", // Bright green
	}
}

// NewLogBrowserStyles creates a new set of consistent styles for the Log Browser
func NewLogBrowserStyles() *LogBrowserStyles {
	return NewLogBrowserStylesWithTheme(false)
}

// NewLogBrowserStylesWithTheme creates styles with a specific theme configuration
func NewLogBrowserStylesWithTheme(highContrast bool) *LogBrowserStyles {
	// Select theme based on highContrast flag
	var theme ThemeConfig
	if highContrast {
		theme = GetHighContrastTheme()
	} else {
		theme = GetDefaultTheme()
	}
	
	// Base panel style
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	
	// Focused panel - use theme color
	focusedPanelStyle := panelStyle.Copy().
		BorderForeground(lipgloss.Color(theme.FocusedBorder))
	
	// Unfocused panel - use theme color
	unfocusedPanelStyle := panelStyle.Copy().
		BorderForeground(lipgloss.Color(theme.UnfocusedBorder))
	
	// Title style - bold with theme color
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Title))
	
	// Subtitle style - theme color
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtitle))
	
	// Hint style - muted with theme color
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Hint)).
		Italic(true)
	
	// Empty state style - centered, muted
	emptyStateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Empty)).
		Italic(true).
		Align(lipgloss.Center)
	
	// Loading style - theme color
	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Loading))
	
	// Error style - theme color, bold
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Error)).
		Bold(true)
	
	// Current line highlight - theme colors
	currentLineStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.LineHighlight)).
		Foreground(lipgloss.Color(theme.LineText))
	
	// Create spinner for loading states
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SpinnerColor))
	
	return &LogBrowserStyles{
		PanelStyle:          panelStyle,
		FocusedPanelStyle:   focusedPanelStyle,
		UnfocusedPanelStyle: unfocusedPanelStyle,
		TitleStyle:          titleStyle,
		SubtitleStyle:       subtitleStyle,
		HintStyle:           hintStyle,
		EmptyStateStyle:     emptyStateStyle,
		LoadingStyle:        loadingStyle,
		ErrorStyle:          errorStyle,
		CurrentLineStyle:    currentLineStyle,
		Spinner:             s,
		HighContrast:        highContrast,
	}
}

// GetEmptyStateMessage returns a helpful message for empty states
func (s *LogBrowserStyles) GetEmptyStateMessage(panelType string) string {
	messages := map[string]string{
		"file_browser": "No log files found\n\nTip: Log files are saved when tasks are run with Crush",
		"tag_selector": "No tags available\n\nTip: Create tags with task-master or run tasks to generate logs",
		"log_viewer":   "Select a log file to preview\n\nUse ← → to navigate between panels",
	}
	
	if msg, ok := messages[panelType]; ok {
		return msg
	}
	return "No content available"
}

// RenderEmptyState renders an empty state message with consistent styling
func (s *LogBrowserStyles) RenderEmptyState(panelType string, width, height int) string {
	message := s.GetEmptyStateMessage(panelType)
	
	// Center the message
	style := s.EmptyStateStyle.
		Width(width - 4).
		Height(height - 4).
		AlignVertical(lipgloss.Center)
	
	return style.Render(message)
}

// RenderLoadingState renders a loading indicator
func (s *LogBrowserStyles) RenderLoadingState(message string, spinnerView string) string {
	return s.LoadingStyle.Render(spinnerView + " " + message)
}
