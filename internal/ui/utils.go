package ui

import (
	"regexp"
	"strings"
	"unicode"
)

// Slugify converts a title string into a valid filename by lowercasing,
// replacing spaces with hyphens, and removing special characters.
func Slugify(text string) string {
	// Handle empty string
	if strings.TrimSpace(text) == "" {
		return ""
	}

	// Convert to lowercase
	slug := strings.ToLower(text)

	// Remove accented characters and convert to ASCII equivalents
	slug = removeAccents(slug)

	// Replace spaces and underscores with hyphens
	slug = strings.NewReplacer(
		" ", "-",
		"_", "-",
	).Replace(slug)

	// Remove any character that is not alphanumeric, hyphen, or period
	re := regexp.MustCompile(`[^a-z0-9\-.]`)
	slug = re.ReplaceAllString(slug, "")

	// Replace multiple hyphens with a single hyphen
	re = regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

// removeAccents removes common accented characters and replaces them with ASCII equivalents
func removeAccents(text string) string {
	// Map of accented characters to their ASCII equivalents
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a', 'ã': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'ô': 'o', 'ö': 'o', 'õ': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
	}

	result := make([]rune, 0, len(text))
	for _, r := range text {
		if replacement, exists := replacements[r]; exists {
			result = append(result, replacement)
		} else if !unicode.Is(unicode.Mn, r) { // Skip combining marks
			result = append(result, r)
		}
	}

	return string(result)
}
