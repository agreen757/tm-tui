package ui

import (
	"testing"
)

func TestExecutionQueue_CurrentTask(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		expected string
	}{
		{
			name:     "nil queue returns empty string",
			queue:    nil,
			expected: "",
		},
		{
			name:     "empty queue returns empty string",
			queue:    &ExecutionQueue{TaskIDs: []string{}, CurrentIndex: 0},
			expected: "",
		},
		{
			name: "valid index returns task ID",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2", "task3"},
				CurrentIndex: 1,
			},
			expected: "task2",
		},
		{
			name: "first task",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2"},
				CurrentIndex: 0,
			},
			expected: "task1",
		},
		{
			name: "last task",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2", "task3"},
				CurrentIndex: 2,
			},
			expected: "task3",
		},
		{
			name: "index out of bounds",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2"},
				CurrentIndex: 5,
			},
			expected: "",
		},
		{
			name: "negative index",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2"},
				CurrentIndex: -1,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.queue.CurrentTask()
			if got != tt.expected {
				t.Errorf("CurrentTask() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExecutionQueue_HasNext(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		expected bool
	}{
		{
			name:     "nil queue returns false",
			queue:    nil,
			expected: false,
		},
		{
			name:     "empty queue returns false",
			queue:    &ExecutionQueue{TaskIDs: []string{}},
			expected: false,
		},
		{
			name: "at last task returns false",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2"},
				CurrentIndex: 1,
			},
			expected: false,
		},
		{
			name: "not at last task returns true",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2", "task3"},
				CurrentIndex: 1,
			},
			expected: true,
		},
		{
			name: "at first task with multiple tasks returns true",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2"},
				CurrentIndex: 0,
			},
			expected: true,
		},
		{
			name: "single task returns false",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1"},
				CurrentIndex: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.queue.HasNext()
			if got != tt.expected {
				t.Errorf("HasNext() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExecutionQueue_Next(t *testing.T) {
	tests := []struct {
		name          string
		queue         *ExecutionQueue
		expectedIndex int
	}{
		{
			name:          "nil queue is no-op",
			queue:         nil,
			expectedIndex: 0,
		},
		{
			name:          "empty queue is no-op",
			queue:         &ExecutionQueue{TaskIDs: []string{}, CurrentIndex: 0},
			expectedIndex: 0,
		},
		{
			name: "advance from first to second task",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2", "task3"},
				CurrentIndex: 0,
			},
			expectedIndex: 1,
		},
		{
			name: "advance at last task does not go beyond",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2"},
				CurrentIndex: 1,
			},
			expectedIndex: 1,
		},
		{
			name: "multiple advances",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2", "task3", "task4"},
				CurrentIndex: 0,
			},
			expectedIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.queue != nil {
				tt.queue.Next()
			}
			if tt.queue != nil && tt.queue.CurrentIndex != tt.expectedIndex {
				t.Errorf("After Next(), CurrentIndex = %d, want %d", tt.queue.CurrentIndex, tt.expectedIndex)
			}
		})
	}
}

func TestExecutionQueue_Skip(t *testing.T) {
	tests := []struct {
		name              string
		queue             *ExecutionQueue
		expectedTaskCount int
		expectedIndex     int
		expectedCurrentID string
	}{
		{
			name:              "nil queue is no-op",
			queue:             nil,
			expectedTaskCount: 0,
			expectedIndex:     0,
			expectedCurrentID: "",
		},
		{
			name:              "empty queue is no-op",
			queue:             &ExecutionQueue{TaskIDs: []string{}, CurrentIndex: 0},
			expectedTaskCount: 0,
			expectedIndex:     0,
			expectedCurrentID: "",
		},
		{
			name: "skip from first task of three",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1", "task2", "task3"},
				CurrentIndex:    0,
				ModelSelections: map[string]string{"task1": "model-a"},
				TaskStatus:      map[string]string{"task1": "pending"},
			},
			expectedTaskCount: 2,
			expectedIndex:     0,
			expectedCurrentID: "task2",
		},
		{
			name: "skip from middle task",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1", "task2", "task3"},
				CurrentIndex:    1,
				ModelSelections: map[string]string{"task2": "model-b"},
				TaskStatus:      map[string]string{"task2": "executing"},
			},
			expectedTaskCount: 2,
			expectedIndex:     1,
			expectedCurrentID: "task3",
		},
		{
			name: "skip from last task of multiple",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1", "task2", "task3"},
				CurrentIndex:    2,
				ModelSelections: map[string]string{"task3": "model-c"},
				TaskStatus:      map[string]string{"task3": "done"},
			},
			expectedTaskCount: 2,
			expectedIndex:     1,
			expectedCurrentID: "task2",
		},
		{
			name: "skip single task",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1"},
				CurrentIndex:    0,
				ModelSelections: map[string]string{"task1": "model-a"},
				TaskStatus:      map[string]string{"task1": "pending"},
			},
			expectedTaskCount: 0,
			expectedIndex:     0,
			expectedCurrentID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.queue != nil {
				tt.queue.Skip()
				if len(tt.queue.TaskIDs) != tt.expectedTaskCount {
					t.Errorf("After Skip(), TaskIDs count = %d, want %d", len(tt.queue.TaskIDs), tt.expectedTaskCount)
				}
				if tt.queue.CurrentIndex != tt.expectedIndex {
					t.Errorf("After Skip(), CurrentIndex = %d, want %d", tt.queue.CurrentIndex, tt.expectedIndex)
				}
				if tt.queue.CurrentTask() != tt.expectedCurrentID {
					t.Errorf("After Skip(), CurrentTask() = %q, want %q", tt.queue.CurrentTask(), tt.expectedCurrentID)
				}
			}
		})
	}
}

func TestExecutionQueue_RemainingCount(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		expected int
	}{
		{
			name:     "nil queue returns 0",
			queue:    nil,
			expected: 0,
		},
		{
			name:     "empty queue returns 0",
			queue:    &ExecutionQueue{TaskIDs: []string{}},
			expected: 0,
		},
		{
			name: "single task returns 1",
			queue: &ExecutionQueue{
				TaskIDs: []string{"task1"},
			},
			expected: 1,
		},
		{
			name: "three tasks returns 3",
			queue: &ExecutionQueue{
				TaskIDs: []string{"task1", "task2", "task3"},
			},
			expected: 3,
		},
		{
			name: "many tasks",
			queue: &ExecutionQueue{
				TaskIDs: []string{"t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9"},
			},
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.queue.RemainingCount()
			if got != tt.expected {
				t.Errorf("RemainingCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExecutionQueue_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		expected bool
	}{
		{
			name:     "nil queue returns false",
			queue:    nil,
			expected: false,
		},
		{
			name:     "empty queue returns false",
			queue:    &ExecutionQueue{TaskIDs: []string{}},
			expected: false,
		},
		{
			name: "single task with model selection returns true",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1"},
				ModelSelections: map[string]string{"task1": "model-a"},
			},
			expected: true,
		},
		{
			name: "single task without model selection returns false",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1"},
				ModelSelections: map[string]string{},
			},
			expected: false,
		},
		{
			name: "all tasks have selections returns true",
			queue: &ExecutionQueue{
				TaskIDs: []string{"task1", "task2", "task3"},
				ModelSelections: map[string]string{
					"task1": "model-a",
					"task2": "model-b",
					"task3": "model-c",
				},
			},
			expected: true,
		},
		{
			name: "some tasks missing selections returns false",
			queue: &ExecutionQueue{
				TaskIDs: []string{"task1", "task2", "task3"},
				ModelSelections: map[string]string{
					"task1": "model-a",
					"task3": "model-c",
					// task2 missing
				},
			},
			expected: false,
		},
		{
			name: "no model selections returns false",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1", "task2"},
				ModelSelections: map[string]string{},
			},
			expected: false,
		},
		{
			name: "nil model selections map returns false",
			queue: &ExecutionQueue{
				TaskIDs:         []string{"task1", "task2"},
				ModelSelections: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.queue.IsComplete()
			if got != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExecutionQueue_GetModelSelection(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		taskID   string
		expected string
	}{
		{
			name:     "nil queue returns empty string",
			queue:    nil,
			taskID:   "task1",
			expected: "",
		},
		{
			name:     "nil model selections returns empty string",
			queue:    &ExecutionQueue{ModelSelections: nil},
			taskID:   "task1",
			expected: "",
		},
		{
			name: "existing selection returns model ID",
			queue: &ExecutionQueue{
				ModelSelections: map[string]string{"task1": "model-a"},
			},
			taskID:   "task1",
			expected: "model-a",
		},
		{
			name: "non-existing selection returns empty string",
			queue: &ExecutionQueue{
				ModelSelections: map[string]string{"task1": "model-a"},
			},
			taskID:   "task2",
			expected: "",
		},
		{
			name: "empty model selections returns empty string",
			queue: &ExecutionQueue{
				ModelSelections: map[string]string{},
			},
			taskID:   "task1",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.queue.GetModelSelection(tt.taskID)
			if got != tt.expected {
				t.Errorf("GetModelSelection(%q) = %q, want %q", tt.taskID, got, tt.expected)
			}
		})
	}
}

func TestExecutionQueue_SetModelSelection(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		taskID   string
		modelID  string
		validate func(*ExecutionQueue) bool
	}{
		{
			name:    "nil queue is no-op",
			queue:   nil,
			taskID:  "task1",
			modelID: "model-a",
			validate: func(q *ExecutionQueue) bool {
				return q == nil
			},
		},
		{
			name:    "set on nil map initializes map and sets value",
			queue:   &ExecutionQueue{ModelSelections: nil},
			taskID:  "task1",
			modelID: "model-a",
			validate: func(q *ExecutionQueue) bool {
				return q.ModelSelections != nil && q.ModelSelections["task1"] == "model-a"
			},
		},
		{
			name:    "set on empty map adds entry",
			queue:   &ExecutionQueue{ModelSelections: map[string]string{}},
			taskID:  "task1",
			modelID: "model-a",
			validate: func(q *ExecutionQueue) bool {
				return q.ModelSelections["task1"] == "model-a"
			},
		},
		{
			name: "set overwrites existing selection",
			queue: &ExecutionQueue{
				ModelSelections: map[string]string{"task1": "model-old"},
			},
			taskID:  "task1",
			modelID: "model-new",
			validate: func(q *ExecutionQueue) bool {
				return q.ModelSelections["task1"] == "model-new"
			},
		},
		{
			name: "set multiple different tasks",
			queue: &ExecutionQueue{
				ModelSelections: map[string]string{"task1": "model-a"},
			},
			taskID:  "task2",
			modelID: "model-b",
			validate: func(q *ExecutionQueue) bool {
				return q.ModelSelections["task1"] == "model-a" && q.ModelSelections["task2"] == "model-b"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.queue != nil {
				tt.queue.SetModelSelection(tt.taskID, tt.modelID)
			}
			if !tt.validate(tt.queue) {
				t.Errorf("SetModelSelection validation failed")
			}
		})
	}
}

func TestExecutionQueue_Reset(t *testing.T) {
	tests := []struct {
		name     string
		queue    *ExecutionQueue
		validate func(*ExecutionQueue) bool
	}{
		{
			name:  "nil queue is no-op",
			queue: nil,
			validate: func(q *ExecutionQueue) bool {
				return q == nil
			},
		},
		{
			name: "reset clears all state",
			queue: &ExecutionQueue{
				TaskIDs:      []string{"task1", "task2", "task3"},
				CurrentIndex: 2,
				ModelSelections: map[string]string{
					"task1": "model-a",
					"task2": "model-b",
					"task3": "model-c",
				},
				TaskStatus: map[string]string{
					"task1": "done",
					"task2": "executing",
					"task3": "pending",
				},
			},
			validate: func(q *ExecutionQueue) bool {
				return len(q.TaskIDs) == 0 &&
					q.CurrentIndex == 0 &&
					len(q.ModelSelections) == 0 &&
					len(q.TaskStatus) == 0
			},
		},
		{
			name: "reset on already empty queue",
			queue: &ExecutionQueue{
				TaskIDs:         []string{},
				CurrentIndex:    0,
				ModelSelections: map[string]string{},
				TaskStatus:      map[string]string{},
			},
			validate: func(q *ExecutionQueue) bool {
				return len(q.TaskIDs) == 0 &&
					q.CurrentIndex == 0 &&
					len(q.ModelSelections) == 0 &&
					len(q.TaskStatus) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.queue != nil {
				tt.queue.Reset()
			}
			if !tt.validate(tt.queue) {
				t.Errorf("Reset validation failed")
			}
		})
	}
}

// Integration tests for complex workflows

func TestExecutionQueue_QueueWorkflow(t *testing.T) {
	t.Run("complete task execution workflow", func(t *testing.T) {
		// Create a queue with multiple tasks
		queue := &ExecutionQueue{
			TaskIDs:         []string{"task1", "task2", "task3"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		}

		// Initial state: should have first task
		if queue.CurrentTask() != "task1" {
			t.Errorf("Expected current task to be task1, got %q", queue.CurrentTask())
		}

		// Set model for first task
		queue.SetModelSelection("task1", "model-a")
		if queue.GetModelSelection("task1") != "model-a" {
			t.Errorf("Expected model-a, got %q", queue.GetModelSelection("task1"))
		}

		// Not complete yet (missing models for task2 and task3)
		if queue.IsComplete() {
			t.Errorf("Expected IsComplete() to be false, got true")
		}

		// Move to next task
		queue.Next()
		if queue.CurrentTask() != "task2" {
			t.Errorf("Expected current task to be task2, got %q", queue.CurrentTask())
		}

		// Set model for second task
		queue.SetModelSelection("task2", "model-b")

		// Move to next task
		queue.Next()
		if queue.CurrentTask() != "task3" {
			t.Errorf("Expected current task to be task3, got %q", queue.CurrentTask())
		}

		// Set model for third task
		queue.SetModelSelection("task3", "model-c")

		// Now should be complete
		if !queue.IsComplete() {
			t.Errorf("Expected IsComplete() to be true, got false")
		}

		// Should not have next
		if queue.HasNext() {
			t.Errorf("Expected HasNext() to be false at last task, got true")
		}
	})

	t.Run("skip task in middle", func(t *testing.T) {
		queue := &ExecutionQueue{
			TaskIDs:         []string{"task1", "task2", "task3"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		}

		// Move to task2
		queue.Next()
		if queue.CurrentTask() != "task2" {
			t.Errorf("Expected task2, got %q", queue.CurrentTask())
		}

		// Skip it
		queue.Skip()
		if queue.CurrentTask() != "task3" {
			t.Errorf("Expected task3 after skipping task2, got %q", queue.CurrentTask())
		}

		if queue.RemainingCount() != 2 {
			t.Errorf("Expected 2 remaining tasks, got %d", queue.RemainingCount())
		}
	})

	t.Run("reset queue clears everything", func(t *testing.T) {
		queue := &ExecutionQueue{
			TaskIDs:      []string{"task1", "task2"},
			CurrentIndex: 1,
			ModelSelections: map[string]string{
				"task1": "model-a",
				"task2": "model-b",
			},
			TaskStatus: map[string]string{
				"task1": "done",
				"task2": "executing",
			},
		}

		queue.Reset()

		if len(queue.TaskIDs) != 0 || queue.CurrentIndex != 0 ||
			len(queue.ModelSelections) != 0 || len(queue.TaskStatus) != 0 {
			t.Errorf("Reset did not properly clear queue state")
		}
	})
}

func TestExecutionQueue_EdgeCases(t *testing.T) {
	t.Run("operations on queue with one task", func(t *testing.T) {
		queue := &ExecutionQueue{
			TaskIDs:         []string{"only-task"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		}

		if queue.CurrentTask() != "only-task" {
			t.Errorf("Expected only-task, got %q", queue.CurrentTask())
		}

		if queue.HasNext() {
			t.Errorf("Expected HasNext() to be false for single task, got true")
		}

		if queue.RemainingCount() != 1 {
			t.Errorf("Expected 1 remaining task, got %d", queue.RemainingCount())
		}

		queue.Next()
		if queue.CurrentTask() != "only-task" {
			t.Errorf("Expected to stay on only-task after Next(), got %q", queue.CurrentTask())
		}
	})

	t.Run("multiple skips", func(t *testing.T) {
		queue := &ExecutionQueue{
			TaskIDs:         []string{"task1", "task2", "task3", "task4"},
			CurrentIndex:    0,
			ModelSelections: make(map[string]string),
			TaskStatus:      make(map[string]string),
		}

		queue.Skip()
		if queue.CurrentTask() != "task2" {
			t.Errorf("After first skip, expected task2, got %q", queue.CurrentTask())
		}

		queue.Skip()
		if queue.CurrentTask() != "task3" {
			t.Errorf("After second skip, expected task3, got %q", queue.CurrentTask())
		}

		if queue.RemainingCount() != 2 {
			t.Errorf("Expected 2 remaining tasks after 2 skips, got %d", queue.RemainingCount())
		}
	})

	t.Run("model selection on non-existent task", func(t *testing.T) {
		queue := &ExecutionQueue{
			TaskIDs:         []string{"task1", "task2"},
			ModelSelections: make(map[string]string),
		}

		// Set model for task that doesn't exist in queue
		queue.SetModelSelection("task3", "model-a")

		// Should still set it (the method doesn't validate task existence)
		if queue.GetModelSelection("task3") != "model-a" {
			t.Errorf("Expected model-a, got %q", queue.GetModelSelection("task3"))
		}

		// But queue should still not be complete
		if queue.IsComplete() {
			t.Errorf("Expected IsComplete() to be false when some tasks in TaskIDs don't have models")
		}
	})
}
