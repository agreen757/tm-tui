package git

import (
	"regexp"
)

// Regular expression to match task IDs in format #X, #X.Y, #X.Y.Z, etc.
var taskIDRegex = regexp.MustCompile(`#([0-9]+(?:\.[0-9]+)*)`)

// ExtractTaskIDs extracts task IDs from a commit message
// Task IDs are expected in the format #X, #X.Y, #X.Y.Z, etc.
// Returns a slice of task IDs without the # prefix
func ExtractTaskIDs(message string) []string {
	matches := taskIDRegex.FindAllStringSubmatch(message, -1)
	
	var taskIDs []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			taskID := match[1]
			// Avoid duplicates
			if !seen[taskID] {
				taskIDs = append(taskIDs, taskID)
				seen[taskID] = true
			}
		}
	}
	
	return taskIDs
}
