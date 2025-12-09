package taskmaster

import (
	"testing"
)

func TestBuildTaskIndex(t *testing.T) {
	tests := []struct {
		name         string
		tasks        []Task
		wantIndexLen int
		wantWarnings int
	}{
		{
			name: "single task",
			tasks: []Task{
				{ID: "1", Title: "Task 1", Status: StatusPending},
			},
			wantIndexLen: 1,
			wantWarnings: 0,
		},
		{
			name: "task with subtasks",
			tasks: []Task{
				{
					ID:     "1",
					Title:  "Parent",
					Status: StatusPending,
					Subtasks: []Task{
						{ID: "1.1", Title: "Child 1", Status: StatusPending},
						{ID: "1.2", Title: "Child 2", Status: StatusDone},
					},
				},
			},
			wantIndexLen: 3, // Parent + 2 children
			wantWarnings: 0,
		},
		{
			name: "duplicate task IDs",
			tasks: []Task{
				{ID: "1", Title: "Task 1", Status: StatusPending},
				{ID: "1", Title: "Task 1 Duplicate", Status: StatusPending},
			},
			wantIndexLen: 1, // Duplicate overwrites
			wantWarnings: 1,
		},
		{
			name: "deeply nested tasks",
			tasks: []Task{
				{
					ID:     "1",
					Title:  "Root",
					Status: StatusPending,
					Subtasks: []Task{
						{
							ID:     "1.1",
							Title:  "Level 1",
							Status: StatusPending,
							Subtasks: []Task{
								{
									ID:     "1.1.1",
									Title:  "Level 2",
									Status: StatusPending,
									Subtasks: []Task{
										{ID: "1.1.1.1", Title: "Level 3", Status: StatusPending},
									},
								},
							},
						},
					},
				},
			},
			wantIndexLen: 4,
			wantWarnings: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, warnings := buildTaskIndex(tt.tasks)
			
			if len(index) != tt.wantIndexLen {
				t.Errorf("buildTaskIndex() index length = %d, want %d", len(index), tt.wantIndexLen)
			}
			
			if len(warnings) != tt.wantWarnings {
				t.Errorf("buildTaskIndex() warnings = %d, want %d", len(warnings), tt.wantWarnings)
			}
			
			// Verify parent relationships
			for id, task := range index {
				if task.ID != id {
					t.Errorf("Index key %s doesn't match task ID %s", id, task.ID)
				}
				
				// Check that children are properly linked
				for i, child := range task.Children {
					if child.Parent != task {
						t.Errorf("Child %s parent pointer doesn't point to parent %s", child.ID, task.ID)
					}
					if child.ParentID != task.ID {
						t.Errorf("Child %s ParentID = %s, want %s", child.ID, child.ParentID, task.ID)
					}
					if &task.Subtasks[i] != child {
						t.Errorf("Children slice doesn't reference correct subtask")
					}
				}
			}
		})
	}
}

func TestFlattenTasks(t *testing.T) {
	tests := []struct {
		name      string
		tasks     []Task
		wantCount int
	}{
		{
			name: "single task",
			tasks: []Task{
				{ID: "1", Title: "Task 1"},
			},
			wantCount: 1,
		},
		{
			name: "task with subtasks",
			tasks: []Task{
				{
					ID:    "1",
					Title: "Parent",
					Subtasks: []Task{
						{ID: "1.1", Title: "Child 1"},
						{ID: "1.2", Title: "Child 2"},
					},
				},
			},
			wantCount: 3,
		},
		{
			name: "multiple root tasks",
			tasks: []Task{
				{ID: "1", Title: "Task 1"},
				{ID: "2", Title: "Task 2"},
				{ID: "3", Title: "Task 3"},
			},
			wantCount: 3,
		},
		{
			name: "complex hierarchy",
			tasks: []Task{
				{
					ID:    "1",
					Title: "Root 1",
					Subtasks: []Task{
						{
							ID:    "1.1",
							Title: "Child 1.1",
							Subtasks: []Task{
								{ID: "1.1.1", Title: "Grandchild"},
							},
						},
						{ID: "1.2", Title: "Child 1.2"},
					},
				},
				{
					ID:    "2",
					Title: "Root 2",
					Subtasks: []Task{
						{ID: "2.1", Title: "Child 2.1"},
					},
				},
			},
			wantCount: 6,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := flattenTasks(tt.tasks)
			
			if len(flattened) != tt.wantCount {
				t.Errorf("flattenTasks() count = %d, want %d", len(flattened), tt.wantCount)
			}
			
			// Verify all tasks are present and unique
			seen := make(map[string]bool)
			for _, task := range flattened {
				if seen[task.ID] {
					t.Errorf("Duplicate task ID in flattened list: %s", task.ID)
				}
				seen[task.ID] = true
			}
		})
	}
}

func TestBuildTaskIndex_ParentRelationships(t *testing.T) {
	tasks := []Task{
		{
			ID:    "1",
			Title: "Parent",
			Subtasks: []Task{
				{ID: "1.1", Title: "Child 1"},
				{ID: "1.2", Title: "Child 2"},
			},
		},
	}
	
	index, _ := buildTaskIndex(tasks)
	
	// Check parent task
	parent, ok := index["1"]
	if !ok {
		t.Fatal("Parent task not in index")
	}
	if parent.Parent != nil {
		t.Error("Root task should have nil parent")
	}
	if parent.ParentID != "" {
		t.Error("Root task should have empty ParentID")
	}
	if len(parent.Children) != 2 {
		t.Errorf("Parent should have 2 children, got %d", len(parent.Children))
	}
	
	// Check child tasks
	child1, ok := index["1.1"]
	if !ok {
		t.Fatal("Child 1 not in index")
	}
	if child1.Parent == nil {
		t.Error("Child should have non-nil parent")
	}
	if child1.Parent.ID != "1" {
		t.Errorf("Child parent ID = %s, want '1'", child1.Parent.ID)
	}
	if child1.ParentID != "1" {
		t.Errorf("Child ParentID = %s, want '1'", child1.ParentID)
	}
	
	child2, ok := index["1.2"]
	if !ok {
		t.Fatal("Child 2 not in index")
	}
	if child2.Parent == nil {
		t.Error("Child should have non-nil parent")
	}
	if child2.Parent.ID != "1" {
		t.Errorf("Child parent ID = %s, want '1'", child2.Parent.ID)
	}
}
