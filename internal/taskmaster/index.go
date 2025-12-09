package taskmaster

import (
	"fmt"
)

// buildTaskIndex creates a map of task ID -> task pointer for O(1) lookups.
// It also builds parent/child relationships in the task tree.
func buildTaskIndex(tasks []Task) (map[string]*Task, []ValidationWarning) {
	index := make(map[string]*Task)
	warnings := []ValidationWarning{}
	
	// First pass: build index and populate parent relationships
	var indexTask func(task *Task, parent *Task)
	indexTask = func(task *Task, parent *Task) {
		// Check for duplicate IDs
		if _, exists := index[task.ID]; exists {
			warnings = append(warnings, ValidationWarning{
				TaskID:  task.ID,
				Message: fmt.Sprintf("Duplicate task ID found: %s", task.ID),
			})
		}
		
		// Add to index
		index[task.ID] = task
		
		// Set parent relationship
		task.Parent = parent
		if parent != nil {
			task.ParentID = parent.ID
		}
		
		// Build children slice and recurse
		task.Children = make([]*Task, len(task.Subtasks))
		for i := range task.Subtasks {
			task.Children[i] = &task.Subtasks[i]
			indexTask(&task.Subtasks[i], task)
		}
	}
	
	for i := range tasks {
		indexTask(&tasks[i], nil)
	}
	
	return index, warnings
}

// flattenTasks returns a flat list of all tasks including subtasks
func flattenTasks(tasks []Task) []*Task {
	result := []*Task{}
	
	var flatten func(task *Task)
	flatten = func(task *Task) {
		result = append(result, task)
		for i := range task.Subtasks {
			flatten(&task.Subtasks[i])
		}
	}
	
	for i := range tasks {
		flatten(&tasks[i])
	}
	
	return result
}
