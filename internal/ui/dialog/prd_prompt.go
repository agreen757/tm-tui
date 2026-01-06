package dialog

import (
	"fmt"
)

// GenerateCrushPrdPrompt creates a prompt for PRD generation
// Instructs Crush to output the PRD content directly to stdout without creating files
func GenerateCrushPrdPrompt(title, summary, scope string) string {
	prompt := fmt.Sprintf(`Create a Product Requirements Document (PRD) for: %s

Summary: %s`, title, summary)

	if scope != "" {
		prompt += fmt.Sprintf(`
Scope/Constraints: %s`, scope)
	}

	prompt += `

CRITICAL: Output the PRD markdown content DIRECTLY in your response.
- DO NOT use file creation tools or create any files
- DO NOT include conversational text like "I'll create..." or "Created at..."
- DO NOT include instructions about parsing or next steps
- Start immediately with the markdown PRD content (beginning with # title)

Include these sections:
1. Title (# heading)
2. Overview
3. Goals/Objectives  
4. Functional Requirements
5. Technical Requirements
6. Success Criteria
7. Implementation Details
8. Testing Strategy

Begin the PRD now:`

	return prompt
}
