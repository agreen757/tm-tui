package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderCompactHelp renders a compact help menu as shown in the screenshot
func (m Model) renderCompactHelp() string {
	// Title centered with the leaf emoji from the screenshot
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHighlight)).
		Bold(true).
		Align(lipgloss.Center)
	
	title := titleStyle.Render("🍃 Task Master TUI Help")
	
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")

	// Width for the help dialog (centered)
	helpWidth := 70
	if m.width > 80 {
		helpWidth = 70
	} else if m.width > 60 {
		helpWidth = m.width - 10
	} else {
		helpWidth = m.width - 4
	}

	// Navigation keys - align in two columns
	b.WriteString(formatTableRow("up", "Move up", "down", "Move down", helpWidth))
	b.WriteString(formatTableRow("left", "Collapse/Move left", "right", "Expand/Move right", helpWidth))
	b.WriteString(formatTableRow("pgup", "Page up", "pgdown", "Page down", helpWidth))
	b.WriteString("\n")

	// Action keys - use one column for longer descriptions
	b.WriteString(formatCompactKey("e", "Toggle expand/collapse", helpWidth))
	b.WriteString(formatCompactKey("space", "Select/deselect for bulk operations", helpWidth))
	b.WriteString(formatCompactKey("n", "Get next available task", helpWidth))
	b.WriteString(formatCompactKey("r", "Refresh tasks from disk", helpWidth))
	b.WriteString(formatCompactKey(":", "Jump to task by ID", helpWidth))
	b.WriteString("\n")

	// Status keys in two columns
	b.WriteString(formatTableRow("i", "► Set in-progress", "D", "✓ Set done", helpWidth))
	b.WriteString(formatTableRow("b", "! Set blocked", "c", "✗ Set cancelled", helpWidth))
	b.WriteString(formatTableRow("f", "⏱ Set deferred", "p", "○ Set pending", helpWidth))
	b.WriteString("\n")

	// Panel keys in columns where possible
	b.WriteString(formatTableRow("1", "Focus task list panel", "2", "Focus details panel", helpWidth))
	b.WriteString(formatTableRow("3", "Focus log panel", "tab", "Cycle through panels", helpWidth))
	b.WriteString(formatTableRow("d", "Toggle details panel", "L", "Toggle log panel", helpWidth))
	b.WriteString(formatTableRow("t", "Switch to tree view", "T", "Switch to list view", helpWidth))
	b.WriteString(formatCompactKey("v", "Cycle view modes", helpWidth))
	b.WriteString("\n")

	// General keys in columns where possible
	b.WriteString(formatTableRow("?", "Toggle this help", "C", "Clear TUI state (reset UI)", helpWidth))
	b.WriteString(formatTableRow("esc", "Back/Cancel/Close", "q", "Quit application", helpWidth))
	b.WriteString(formatCompactKey("ctrl+c", "Cancel command or quit", helpWidth))
	b.WriteString("\n")

	// Footer
	footerStyle := lipgloss.NewStyle().Align(lipgloss.Center)
	footer := footerStyle.Render("Task Master TUI - Terminal interface for Task Master AI\nDocumentation: https://github.com/task-master-ai/tm-tui\nVersion: 1.0.0")
	b.WriteString(footer)
	b.WriteString("\n\n")

	helpPromptStyle := lipgloss.NewStyle().Align(lipgloss.Center)
	helpPrompt := helpPromptStyle.Render("Press '?' or 'Esc' to close help")
	b.WriteString(helpPrompt)

	// Create a box with the help content
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorHighlight)).
		Width(helpWidth).  // Give some margin
		Align(lipgloss.Center)

	// Center the box in the available width
	boxedContent := boxStyle.Render(b.String())
	
	// Wrapper style to center the boxed content in the terminal
	wrapperStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	return wrapperStyle.Render(boxedContent)
}

// formatCompactKey formats a help item to match the original layout
func formatCompactKey(keyName string, description string, width int) string {
	// Create key with background
	keyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color(ColorHighlight)).
		Padding(0, 1).
		Bold(true)

	// Render the key with background
	key := keyStyle.Render(keyName)
	
	// Format the line with centered content
	lineContent := fmt.Sprintf("%s - %s", key, description)
	
	// Center the entire line
	centerStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)
	
	return centerStyle.Render(lineContent) + "\n"
}

// formatTableRow formats two keys and their descriptions in a table-like row
func formatTableRow(key1 string, desc1 string, key2 string, desc2 string, width int) string {
	// Create key with background
	keyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color(ColorHighlight)).
		Padding(0, 1).
		Bold(true)

	// Render the keys with background
	keyBox1 := keyStyle.Render(key1)
	keyBox2 := keyStyle.Render(key2)
	
	// Calculate column widths
	halfWidth := (width - 4) / 2 // 4 for some padding between columns
	
	// Create the left and right columns
	leftCol := fmt.Sprintf("%s - %s", keyBox1, desc1)
	rightCol := fmt.Sprintf("%s - %s", keyBox2, desc2)
	
	// Style for each column, right-pad the left column to ensure spacing
	leftStyle := lipgloss.NewStyle().Width(halfWidth).PaddingRight(2)
	rightStyle := lipgloss.NewStyle().Width(halfWidth)
	
	// Join the columns side by side
	row := lipgloss.JoinHorizontal(lipgloss.Top, 
		leftStyle.Render(leftCol), 
		rightStyle.Render(rightCol))
	
	// Center the entire row
	centerStyle := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	return centerStyle.Render(row) + "\n"
}