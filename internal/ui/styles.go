package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// Color constants from PRD
const (
	ColorPending    = "#FFD700" // Gold
	ColorInProgress = "#4169E1" // Royal Blue
	ColorDone       = "#32CD32" // Lime Green
	ColorBlocked    = "#DC143C" // Crimson
	ColorDeferred   = "#808080" // Gray
	ColorCancelled  = "#404040" // Dark Gray

	ColorBorder    = "#555555"
	ColorText      = "#FFFFFF"
	ColorSubtle    = "#666666"
	ColorHighlight = "#00FFFF"
)

// Color constants for complexity levels (cool-to-hot gradient)
const (
	ColorComplexityLow      = "#4169E1" // Royal Blue
	ColorComplexityMedium   = "#00CED1" // Dark Turquoise
	ColorComplexityHigh     = "#FFA500" // Orange
	ColorComplexityVeryHigh = "#DC143C" // Crimson
)

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	// Status colors
	Pending    lipgloss.Style
	InProgress lipgloss.Style
	Done       lipgloss.Style
	Blocked    lipgloss.Style
	Deferred   lipgloss.Style
	Cancelled  lipgloss.Style

	// Layout styles
	Header    lipgloss.Style
	StatusBar lipgloss.Style
	Border    lipgloss.Style

	// Panel styles
	Panel       lipgloss.Style
	PanelTitle  lipgloss.Style
	PanelBorder lipgloss.Style

	// Task list styles
	TaskSelected   lipgloss.Style
	TaskUnselected lipgloss.Style
	TaskCursor     lipgloss.Style

	// Help styles
	Help    lipgloss.Style
	HelpKey lipgloss.Style
	HelpSep lipgloss.Style

	// Text styles
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Subtle    lipgloss.Style // For subtle/muted text
	Highlight lipgloss.Style // For highlighted/emphasized text
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Success   lipgloss.Style
	Info      lipgloss.Style
	Key       lipgloss.Style // For keyboard shortcuts
}

// NewStyles creates a new Styles instance with default values
func NewStyles() *Styles {
	return &Styles{
		// Status colors
		Pending:    lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPending)),
		InProgress: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorInProgress)),
		Done:       lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDone)),
		Blocked:    lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlocked)),
		Deferred:   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDeferred)),
		Cancelled:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCancelled)),

		// Layout styles
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorHighlight)).
			Background(lipgloss.Color(ColorBorder)).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtle)).
			Background(lipgloss.Color(ColorBorder)).
			Padding(0, 1),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)),

		// Panel styles
		Panel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1),

		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorHighlight)),

		PanelBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)),

		// Task list styles
		TaskSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Bold(true),

		TaskUnselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)),

		TaskCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Bold(true),

		// Help styles
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtle)),

		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)),

		HelpSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtle)),

		// Text styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorHighlight)),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtle)),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlocked)).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPending)),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDone)),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInProgress)),

		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Background(lipgloss.Color("#222222")).
			Padding(0, 1).
			Bold(true),
	}
}

// GetStatusStyle returns the style for a given status
func (s *Styles) GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "pending":
		return s.Pending
	case "in-progress":
		return s.InProgress
	case "done":
		return s.Done
	case "blocked":
		return s.Blocked
	case "deferred":
		return s.Deferred
	case "cancelled":
		return s.Cancelled
	default:
		return lipgloss.NewStyle()
	}
}

// GetStatusIcon returns the icon for a given status
func GetStatusIcon(status string) string {
	switch status {
	case "pending":
		return "○"
	case "in-progress":
		return "►"
	case "done":
		return "✓"
	case "blocked":
		return "!"
	case "deferred":
		return "⏱"
	case "cancelled":
		return "✗"
	default:
		return "?"
	}
}

// GetStatusLabel returns a text label for a given status
// Accessibility: Provides text alternative to icon-only indicators
func GetStatusLabel(status string) string {
	switch status {
	case "pending":
		return "PENDING"
	case "in-progress":
		return "IN-PROGRESS"
	case "done":
		return "DONE"
	case "blocked":
		return "BLOCKED"
	case "deferred":
		return "DEFERRED"
	case "cancelled":
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// GetStatusIndicator returns both icon and text for status
// This combines visual and text representation for accessibility
func GetStatusIndicator(status string) string {
	icon := GetStatusIcon(status)
	label := GetStatusLabel(status)
	if icon == "" {
		return label
	}
	return icon + " " + label
}

// GetComplexityLabel returns a text label for complexity score
// Accessibility: Provides text alternative to color-only indicators
func GetComplexityLabel(complexity int) string {
	if complexity <= 0 {
		return ""
	}

	// Use taskmaster.DefaultLevelThresholds() for consistency with GetComplexityStyle
	thresholds := taskmaster.DefaultLevelThresholds()

	// Determine label based on complexity score matching the threshold logic
	switch {
	case complexity <= thresholds.Low:
		return "LOW"
	case complexity <= thresholds.Medium:
		return "MEDIUM"
	case complexity <= thresholds.High:
		return "HIGH"
	default:
		return "VERY HIGH"
	}
}

// GetComplexityIndicator returns both icon and text for complexity
// This combines visual and text representation for accessibility
func GetComplexityIndicator(complexity int) string {
	label := GetComplexityLabel(complexity)
	if label == "" {
		return ""
	}
	// Format: LEVEL(numeric) e.g., "HIGH(8)"
	return fmt.Sprintf("%s(%d)", label, complexity)
}

// GetComplexityStyle returns a lipgloss style for rendering complexity indicators based on numeric score
func (s *Styles) GetComplexityStyle(complexity int) lipgloss.Style {
	// Return subtle style for complexity = 0 (no complexity assigned)
	if complexity == 0 {
		return s.Subtle
	}

	// Use taskmaster.DefaultLevelThresholds() to determine score ranges
	thresholds := taskmaster.DefaultLevelThresholds()

	// Determine color based on complexity score
	var color string
	switch {
	case complexity <= thresholds.Low:
		color = ColorComplexityLow
	case complexity <= thresholds.Medium:
		color = ColorComplexityMedium
	case complexity <= thresholds.High:
		color = ColorComplexityHigh
	default:
		color = ColorComplexityVeryHigh
	}

	// Return bold styled text with appropriate foreground color
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color))
}

// GetComplexityLevelStyle returns a lipgloss style for rendering complexity indicators based on level enum
func (s *Styles) GetComplexityLevelStyle(level taskmaster.ComplexityLevel) lipgloss.Style {
	// Determine color based on complexity level
	var color string
	switch level {
	case taskmaster.ComplexityLow:
		color = ColorComplexityLow
	case taskmaster.ComplexityMedium:
		color = ColorComplexityMedium
	case taskmaster.ComplexityHigh:
		color = ColorComplexityHigh
	case taskmaster.ComplexityVeryHigh:
		color = ColorComplexityVeryHigh
	default:
		// Return subtle style for unknown or none
		return s.Subtle
	}

	// Return bold styled text with appropriate foreground color
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color))
}
