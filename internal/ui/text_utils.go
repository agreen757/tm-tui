package ui

import "strings"

// wrapText wraps the given text to fit within the specified width.
// It performs simple word-wrapping by breaking lines at word boundaries.
func wrapText(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}

	var lines []string
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return text
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		// Check if adding the next word exceeds the width
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			// Line is full, start a new one
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	// Don't forget the last line
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}
