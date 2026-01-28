package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// FilterStatusView renders a styled visual indicator for the active filter state.
// It displays the current filter text (truncated to 20 chars) and count information.
// Returns an empty string if no filter is active.
//
// Example output: "[FILTER: search] 5/12"
func FilterStatusView(filterValue string, matchCount int, totalCount int) string {
	// Return empty string if no filter is active
	if filterValue == "" {
		return ""
	}

	// Safely truncate filter text to 20 characters
	truncatedFilter := filterValue
	if len(filterValue) > 20 {
		truncatedFilter = filterValue[:20]
	}

	// Format the filter status text
	filterText := fmt.Sprintf("[FILTER: %s] %d/%d", truncatedFilter, matchCount, totalCount)

	// Create lipgloss style with border and padding
	// Use ColorHighlight for foreground to match existing theme
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHighlight)).
		Background(lipgloss.Color(ColorBorder)).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorHighlight))

	return style.Render(filterText)
}
