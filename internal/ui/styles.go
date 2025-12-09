package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color constants from PRD
const (
	ColorPending   = "#FFD700" // Gold
	ColorInProgress = "#4169E1" // Royal Blue
	ColorDone      = "#32CD32" // Lime Green
	ColorBlocked   = "#DC143C" // Crimson
	ColorDeferred  = "#808080" // Gray
	ColorCancelled = "#404040" // Dark Gray
	
	ColorBorder    = "#555555"
	ColorText      = "#FFFFFF"
	ColorSubtle    = "#666666"
	ColorHighlight = "#00FFFF"
)

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	// Status colors
	Pending   lipgloss.Style
	InProgress lipgloss.Style
	Done      lipgloss.Style
	Blocked   lipgloss.Style
	Deferred  lipgloss.Style
	Cancelled lipgloss.Style
	
	// Layout styles
	Header     lipgloss.Style
	StatusBar  lipgloss.Style
	Border     lipgloss.Style
	
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
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Subtle   lipgloss.Style  // For subtle/muted text
	Error    lipgloss.Style
	Warning  lipgloss.Style
	Success  lipgloss.Style
	Info     lipgloss.Style
	Key      lipgloss.Style  // For keyboard shortcuts
}

// NewStyles creates a new Styles instance with default values
func NewStyles() *Styles {
	return &Styles{
		// Status colors
		Pending:   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPending)),
		InProgress: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorInProgress)),
		Done:      lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDone)),
		Blocked:   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlocked)),
		Deferred:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDeferred)),
		Cancelled: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCancelled)),
		
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
