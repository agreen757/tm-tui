package taskmaster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadTasksFromFile reads and parses tasks.json from the given taskmaster root directory.
// It supports multiple formats:
// 1. Tagged format: { "tag-name": { "tasks": [...] } }
// 2. Direct array format: { "tasks": [...] }
// 3. Simple array format: [...]
// The tag parameter specifies which tag to load; if empty or not found, attempts to use "master",
// and if that's not found, uses the first available tag.
func LoadTasksFromFile(rootDir string, tag string) ([]Task, error) {
	tasksPath := filepath.Join(rootDir, ".taskmaster", "tasks", "tasks.json")
	
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}
	
	// Try to determine the format by unmarshaling to a generic map first
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Maybe it's a direct array?
		var tasks []Task
		if err := json.Unmarshal(data, &tasks); err != nil {
			return nil, fmt.Errorf("failed to parse tasks file: %w", err)
		}
		return tasks, nil
	}
	
	// Determine which tag to use
	targetTag := tag
	if targetTag == "" {
		targetTag = "master"
	}
	
	// Try the requested/default tag first
	if tagData, ok := raw[targetTag]; ok {
		if tagMap, ok := tagData.(map[string]interface{}); ok {
			if tasksData, ok := tagMap["tasks"]; ok {
				return parseTasksData(tasksData)
			}
		}
	}
	
	// If requested tag not found, try "master" if we didn't already
	if targetTag != "master" {
		if tagData, ok := raw["master"]; ok {
			if tagMap, ok := tagData.(map[string]interface{}); ok {
				if tasksData, ok := tagMap["tasks"]; ok {
					return parseTasksData(tasksData)
				}
			}
		}
	}
	
	// If still not found, try to use the first available tag
	for key, value := range raw {
		if key == "tasks" {
			// Skip the direct tasks format, we'll handle it below
			continue
		}
		if tagMap, ok := value.(map[string]interface{}); ok {
			if tasksData, ok := tagMap["tasks"]; ok {
				// Found a tag with tasks, use it
				return parseTasksData(tasksData)
			}
		}
	}
	
	// Check for direct tasks array format: { "tasks": [...] }
	if tasksData, ok := raw["tasks"]; ok {
		return parseTasksData(tasksData)
	}
	
	return nil, fmt.Errorf("unrecognized tasks.json format or no tasks found")
}

// parseTasksData is a helper to convert a generic tasks data structure to []Task
func parseTasksData(tasksData interface{}) ([]Task, error) {
	tasksJSON, err := json.Marshal(tasksData)
	if err != nil {
		return nil, fmt.Errorf("failed to re-marshal tasks: %w", err)
	}
	
	var tasks []Task
	if err := json.Unmarshal(tasksJSON, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks array: %w", err)
	}
	
	// Normalize subtask IDs to use proper dotted notation
	normalizeSubtaskIDs(tasks)
	
	return tasks, nil
}

// normalizeSubtaskIDs fixes subtask IDs to follow proper dotted notation
// Parent tasks keep their numeric IDs (1, 2, 3), but subtasks get
// proper dotted IDs (1.1, 1.2, 2.1.1, etc.) based on their position
func normalizeSubtaskIDs(tasks []Task) {
	for i := range tasks {
		fixSubtaskIDs(&tasks[i])
	}
}

// fixSubtaskIDs recursively fixes subtask IDs for a task and all its descendants
func fixSubtaskIDs(task *Task) {
	for i := range task.Subtasks {
		subtask := &task.Subtasks[i]
		// Generate proper ID: parentID.childIndex
		expectedID := fmt.Sprintf("%s.%d", task.ID, i+1)
		if subtask.ID != expectedID {
			subtask.ID = expectedID
		}
		// Recursively fix subtask's subtasks
		fixSubtaskIDs(subtask)
	}
}

// convertIDToString handles both int and string IDs from JSON
func convertIDToString(id interface{}) string {
	switch v := id.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%d", int(v))
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
