package taskmaster

import (
	"fmt"
)

// validateTasks performs validation checks on tasks and returns warnings for issues.
// It validates:
// - Status and priority values
// - Dependency references
// - Circular dependencies
func validateTasks(tasks []Task, index map[string]*Task) []ValidationWarning {
	warnings := []ValidationWarning{}
	
	// Validate each task
	for i := range tasks {
		warnings = append(warnings, validateTask(&tasks[i], index)...)
	}
	
	// Check for circular dependencies
	warnings = append(warnings, detectCircularDependencies(index)...)
	
	return warnings
}

// validateTask validates a single task and its subtasks recursively
func validateTask(task *Task, index map[string]*Task) []ValidationWarning {
	warnings := []ValidationWarning{}
	
	// Validate status
	if !task.IsValidStatus() {
		warnings = append(warnings, ValidationWarning{
			TaskID:  task.ID,
			Message: fmt.Sprintf("Invalid status: %s", task.Status),
		})
	}
	
	// Validate priority
	if !task.IsValidPriority() {
		warnings = append(warnings, ValidationWarning{
			TaskID:  task.ID,
			Message: fmt.Sprintf("Invalid priority: %s", task.Priority),
		})
	}
	
	// Validate dependencies exist
	for _, depID := range task.Dependencies {
		if _, exists := index[depID]; !exists {
			warnings = append(warnings, ValidationWarning{
				TaskID:  task.ID,
				Message: fmt.Sprintf("Dependency not found: %s", depID),
			})
		}
	}
	
	// Recursively validate subtasks
	for i := range task.Subtasks {
		warnings = append(warnings, validateTask(&task.Subtasks[i], index)...)
	}
	
	return warnings
}

// detectCircularDependencies uses DFS to detect cycles in the dependency graph
func detectCircularDependencies(index map[string]*Task) []ValidationWarning {
	warnings := []ValidationWarning{}
	
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	var dfs func(taskID string, path []string) bool
	dfs = func(taskID string, path []string) bool {
		if recStack[taskID] {
			// Found a cycle
			cycleStart := -1
			for i, id := range path {
				if id == taskID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cyclePath := append(path[cycleStart:], taskID)
				warnings = append(warnings, ValidationWarning{
					TaskID:  taskID,
					Message: fmt.Sprintf("Circular dependency detected: %v", cyclePath),
				})
			}
			return true
		}
		
		if visited[taskID] {
			return false
		}
		
		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)
		
		task, exists := index[taskID]
		if exists {
			for _, depID := range task.Dependencies {
				if dfs(depID, path) {
					return true
				}
			}
		}
		
		recStack[taskID] = false
		return false
	}
	
	// Check all tasks
	for taskID := range index {
		if !visited[taskID] {
			dfs(taskID, []string{})
		}
	}
	
	return warnings
}
