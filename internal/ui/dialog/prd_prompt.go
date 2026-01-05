package dialog

import (
	"fmt"
)

// GenerateCrushPrdPrompt creates a simple prompt for PRD generation from a summary
// Takes the user's summary input and requests a complete and concise PRD
func GenerateCrushPrdPrompt(title, summary, scope string) string {
	return fmt.Sprintf("create a complete and concise PRD from this summary: %s", summary)
}
