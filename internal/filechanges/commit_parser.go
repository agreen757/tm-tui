package filechanges

import (
	"regexp"
)

// CommitParser extracts task information from commit messages
type CommitParser struct {
	taskIDPattern *regexp.Regexp
}

// NewCommitParser creates a new commit parser with the default task ID pattern
func NewCommitParser() *CommitParser {
	return &CommitParser{
		taskIDPattern: regexp.MustCompile(`#([0-9]+(?:\.[0-9]+)*)`),
	}
}

// NewCommitParserWithPattern creates a new commit parser with a custom pattern
func NewCommitParserWithPattern(pattern string) (*CommitParser, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &CommitParser{
		taskIDPattern: regex,
	}, nil
}

// ExtractTaskIDs extracts task IDs from a commit message
// Task IDs are expected in the format #X, #X.Y, #X.Y.Z, etc.
// Returns a slice of task IDs without the # prefix
// Duplicates are automatically removed
func (p *CommitParser) ExtractTaskIDs(message string) []string {
	if p == nil || p.taskIDPattern == nil {
		return []string{}
	}

	matches := p.taskIDPattern.FindAllStringSubmatch(message, -1)
	
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

// HasTaskIDs checks if a commit message contains any task IDs
func (p *CommitParser) HasTaskIDs(message string) bool {
	taskIDs := p.ExtractTaskIDs(message)
	return len(taskIDs) > 0
}

// GetFirstTaskID returns the first task ID found in a commit message
// Returns an empty string if no task IDs are found
func (p *CommitParser) GetFirstTaskID(message string) string {
	taskIDs := p.ExtractTaskIDs(message)
	if len(taskIDs) > 0 {
		return taskIDs[0]
	}
	return ""
}
