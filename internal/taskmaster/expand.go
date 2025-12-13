package taskmaster

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ExpandTaskOptions captures configuration for generating expansion drafts and invoking the CLI.
type ExpandTaskOptions struct {
	Depth       int
	NumSubtasks int
	UseAI       bool
	Force       bool
}

// SubtaskDraft represents a proposed subtask hierarchy that can be previewed or edited before applying.
type SubtaskDraft struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Children    []SubtaskDraft `json:"children,omitempty"`
}

// ExpandProgressState reports coarse CLI progress updates back to the UI.
type ExpandProgressState struct {
	Stage    string
	Progress float64
}

var bulletPattern = regexp.MustCompile(`^([\-*+]\s+|\d+[\.)]\s+)`)

// ExpandTaskDrafts builds an initial draft hierarchy using a deterministic rule-based strategy.
func ExpandTaskDrafts(task *Task, opts ExpandTaskOptions) []SubtaskDraft {
	if task == nil {
		return nil
	}
	depth := opts.Depth
	if depth <= 0 {
		depth = 1
	}
	segments := deriveSegments(task, opts)
	if len(segments) == 0 {
		fallback := fmt.Sprintf("Break down %s into actionable steps", safeTitle(task.Title))
		segments = []string{fallback}
	}

	drafts := make([]SubtaskDraft, 0, len(segments))
	for idx, segment := range segments {
		draft := buildDraft(segment, depth-1, opts, idx)
		drafts = append(drafts, draft)
	}
	return drafts
}

func buildDraft(segment string, remainingDepth int, opts ExpandTaskOptions, idx int) SubtaskDraft {
	segment = strings.TrimSpace(segment)
	title := summarizeSegment(segment, idx)
	draft := SubtaskDraft{Title: title, Description: segment}
	if remainingDepth <= 0 {
		return draft
	}
	childSegments := splitChildSegments(segment)
	limit := opts.NumSubtasks
	if limit > 0 && limit < len(childSegments) {
		childSegments = childSegments[:limit]
	}
	for childIdx, child := range childSegments {
		childDraft := buildDraft(child, remainingDepth-1, opts, childIdx)
		draft.Children = append(draft.Children, childDraft)
	}
	return draft
}

func deriveSegments(task *Task, opts ExpandTaskOptions) []string {
	source := strings.TrimSpace(task.Description)
	if source == "" {
		source = strings.TrimSpace(task.Details)
	}
	if source == "" {
		source = task.Title
	}
	lines := splitNonEmptyLines(source)
	bulletCandidates := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if bulletPattern.MatchString(trimmed) {
			bulletCandidates = append(bulletCandidates, bulletPattern.ReplaceAllString(trimmed, ""))
		}
	}
	if len(bulletCandidates) >= 2 {
		return applySegmentLimit(bulletCandidates, opts.NumSubtasks)
	}

	sentences := splitSentences(source)
	if len(sentences) == 0 {
		return []string{}
	}
	if opts.NumSubtasks > 0 && len(sentences) > opts.NumSubtasks {
		sentences = sentences[:opts.NumSubtasks]
	}
	return sentences
}

func splitNonEmptyLines(input string) []string {
	parts := strings.Split(input, "\n")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func splitSentences(input string) []string {
	separators := func(r rune) bool {
		switch r {
		case '.', '!', '?':
			return true
		default:
			return false
		}
	}
	raw := strings.FieldsFunc(input, separators)
	sentences := make([]string, 0, len(raw))
	for _, chunk := range raw {
		trimmed := strings.TrimSpace(chunk)
		if trimmed == "" {
			continue
		}
		sentences = append(sentences, trimmed)
	}
	return sentences
}

func splitChildSegments(segment string) []string {
	if strings.Contains(segment, "\n") {
		lines := splitNonEmptyLines(segment)
		if len(lines) > 1 {
			return lines
		}
	}
	parts := strings.Split(segment, ";")
	if len(parts) > 1 {
		cleaned := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			return cleaned
		}
	}
	words := strings.Split(segment, ",")
	if len(words) > 1 {
		cleaned := make([]string, 0, len(words))
		for _, word := range words {
			trimmed := strings.TrimSpace(word)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 1 {
			return cleaned
		}
	}
	return splitSentences(segment)
}

func applySegmentLimit(segments []string, limit int) []string {
	if limit <= 0 || len(segments) <= limit {
		return segments
	}
	return append([]string{}, segments[:limit]...)
}

func summarizeSegment(segment string, idx int) string {
	trimmed := strings.TrimSpace(segment)
	if trimmed == "" {
		return fmt.Sprintf("Task Segment %d", idx+1)
	}
	trimmed = strings.TrimSuffix(trimmed, ".")
	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return fmt.Sprintf("Task Segment %d", idx+1)
	}
	maxWords := 10
	if len(words) > maxWords {
		words = words[:maxWords]
	}
	title := strings.Join(words, " ")
	if len(words) == maxWords {
		title += "…"
	}
	return titleCase(title)
}

// FlattenDrafts flattens the hierarchy into a list with depth metadata for editing and preview purposes.
type FlattenedDraft struct {
	Draft *SubtaskDraft
	Level int
	Path  []int
}

func FlattenDrafts(drafts []SubtaskDraft) []FlattenedDraft {
	flattened := make([]FlattenedDraft, 0)
	var walk func(items []SubtaskDraft, level int, path []int)
	walk = func(items []SubtaskDraft, level int, path []int) {
		for idx := range items {
			currentPath := append(append([]int{}, path...), idx)
			flattened = append(flattened, FlattenedDraft{Draft: &items[idx], Level: level, Path: currentPath})
			if len(items[idx].Children) > 0 {
				walk(items[idx].Children, level+1, currentPath)
			}
		}
	}
	walk(drafts, 0, nil)
	return flattened
}

// FormatDraftsAsPrompt renders the drafts as a newline-delimited plan for CLI consumption.
func FormatDraftsAsPrompt(drafts []SubtaskDraft) string {
	lines := make([]string, 0)
	var walk func(items []SubtaskDraft, level int)
	walk = func(items []SubtaskDraft, level int) {
		indent := strings.Repeat("  ", level)
		for _, item := range items {
			line := fmt.Sprintf("%s- %s", indent, item.Title)
			if item.Description != "" && !strings.EqualFold(item.Description, item.Title) {
				line = fmt.Sprintf("%s (%s)", line, item.Description)
			}
			lines = append(lines, line)
			if len(item.Children) > 0 {
				walk(item.Children, level+1)
			}
		}
	}
	walk(drafts, 0)
	return strings.Join(lines, "\n")
}

func safeTitle(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "the selected task"
	}
	return trimmed
}

// NormalizeDrafts sorts child drafts alphabetically for deterministic previews when editing.
func NormalizeDrafts(drafts []SubtaskDraft) []SubtaskDraft {
	for i := range drafts {
		if len(drafts[i].Children) > 0 {
			drafts[i].Children = NormalizeDrafts(drafts[i].Children)
		}
	}
	sort.SliceStable(drafts, func(i, j int) bool {
		return strings.ToLower(drafts[i].Title) < strings.ToLower(drafts[j].Title)
	})
	return drafts
}

func titleCase(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		if len(word) == 0 {
			continue
		}
		if len(word) == 1 {
			words[i] = strings.ToUpper(word)
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
	}
	return strings.Join(words, " ")
}

// ApplySubtaskDrafts converts drafts into actual Task objects and applies them to a parent task.
// It returns the new subtask IDs in order, or an error if the operation fails.
func ApplySubtaskDrafts(parentTask *Task, drafts []SubtaskDraft) ([]string, error) {
	if parentTask == nil {
		return nil, fmt.Errorf("parent task is nil")
	}

	newSubtasks := make([]Task, 0, len(drafts))
	newIDs := make([]string, 0, len(drafts))

	// Generate subtask IDs based on parent ID
	for i, draft := range drafts {
		subtaskID := generateSubtaskID(parentTask.ID, i)
		newIDs = append(newIDs, subtaskID)

		subtask := draftToTask(draft, subtaskID, parentTask.ID, i)
		newSubtasks = append(newSubtasks, subtask)
	}

	// Apply the subtasks to the parent
	parentTask.Subtasks = append(parentTask.Subtasks, newSubtasks...)

	return newIDs, nil
}

// generateSubtaskID creates a proper subtask ID based on the parent ID and child index
func generateSubtaskID(parentID string, childIndex int) string {
	// Handle nested IDs like "1.2.3" by appending the next level
	return fmt.Sprintf("%s.%d", parentID, childIndex+1)
}

// draftToTask converts a SubtaskDraft into a Task object
func draftToTask(draft SubtaskDraft, id, parentID string, index int) Task {
	now := time.Now()

	task := Task{
		ID:          id,
		Title:       draft.Title,
		Description: draft.Description,
		Status:      StatusPending,
		ParentID:    parentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Recursively apply children
	if len(draft.Children) > 0 {
		childSubtasks := make([]Task, 0, len(draft.Children))
		for i, childDraft := range draft.Children {
			childID := generateSubtaskID(id, i)
			childTask := draftToTask(childDraft, childID, id, i)
			childSubtasks = append(childSubtasks, childTask)
		}
		task.Subtasks = childSubtasks
	}

	return task
}

