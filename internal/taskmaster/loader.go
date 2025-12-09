package taskmaster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadTasksFromFile reads and parses tasks.json from the given taskmaster root directory.
// It supports multiple formats:
// 1. Tagged format: { "master": { "tasks": [...] } }
// 2. Direct array format: { "tasks": [...] }
// 3. Simple array format: [...]
func LoadTasksFromFile(rootDir string) ([]Task, error) {
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
	
	// Check for tagged format (e.g., "master": { "tasks": [...] })
	if masterData, ok := raw["master"]; ok {
		if masterMap, ok := masterData.(map[string]interface{}); ok {
			if tasksData, ok := masterMap["tasks"]; ok {
				// Re-marshal and unmarshal to get proper Task structs
				tasksJSON, err := json.Marshal(tasksData)
				if err != nil {
					return nil, fmt.Errorf("failed to re-marshal tasks: %w", err)
				}
				
				var tasks []Task
				if err := json.Unmarshal(tasksJSON, &tasks); err != nil {
					return nil, fmt.Errorf("failed to parse tasks array: %w", err)
				}
				return tasks, nil
			}
		}
	}
	
	// Check for direct tasks array format: { "tasks": [...] }
	if tasksData, ok := raw["tasks"]; ok {
		tasksJSON, err := json.Marshal(tasksData)
		if err != nil {
			return nil, fmt.Errorf("failed to re-marshal tasks: %w", err)
		}
		
		var tasks []Task
		if err := json.Unmarshal(tasksJSON, &tasks); err != nil {
			return nil, fmt.Errorf("failed to parse tasks array: %w", err)
		}
		return tasks, nil
	}
	
	return nil, fmt.Errorf("unrecognized tasks.json format")
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
