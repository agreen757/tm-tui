package taskmaster

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTasksFromFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantTasks int
		wantErr   bool
	}{
		{
			name: "tagged format with master tag",
			content: `{
				"master": {
					"tasks": [
						{
							"id": "1",
							"title": "Task 1",
							"description": "Description 1",
							"status": "pending",
							"priority": "high",
							"dependencies": [],
							"details": "",
							"testStrategy": "",
							"subtasks": []
						}
					]
				}
			}`,
			wantTasks: 1,
			wantErr:   false,
		},
		{
			name: "direct tasks array format",
			content: `{
				"tasks": [
					{
						"id": "1",
						"title": "Task 1",
						"description": "Description 1",
						"status": "pending",
						"priority": "high",
						"dependencies": [],
						"details": "",
						"testStrategy": "",
						"subtasks": []
					},
					{
						"id": "2",
						"title": "Task 2",
						"description": "Description 2",
						"status": "done",
						"priority": "medium",
						"dependencies": ["1"],
						"details": "",
						"testStrategy": "",
						"subtasks": []
					}
				]
			}`,
			wantTasks: 2,
			wantErr:   false,
		},
		{
			name: "simple array format",
			content: `[
				{
					"id": "1",
					"title": "Task 1",
					"description": "Description 1",
					"status": "pending",
					"priority": "high",
					"dependencies": [],
					"details": "",
					"testStrategy": "",
					"subtasks": []
				}
			]`,
			wantTasks: 1,
			wantErr:   false,
		},
		{
			name: "nested subtasks",
			content: `{
				"tasks": [
					{
						"id": "1",
						"title": "Parent Task",
						"description": "Description",
						"status": "in-progress",
						"priority": "high",
						"dependencies": [],
						"details": "",
						"testStrategy": "",
						"subtasks": [
							{
								"id": "1.1",
								"title": "Subtask 1",
								"description": "Subtask description",
								"status": "done",
								"priority": "medium",
								"dependencies": [],
								"details": "",
								"testStrategy": "",
								"subtasks": []
							},
							{
								"id": "1.2",
								"title": "Subtask 2",
								"description": "Subtask description",
								"status": "pending",
								"priority": "medium",
								"dependencies": [],
								"details": "",
								"testStrategy": "",
								"subtasks": []
							}
						]
					}
				]
			}`,
			wantTasks: 1,
			wantErr:   false,
		},
		{
			name:      "invalid json",
			content:   `{invalid json}`,
			wantTasks: 0,
			wantErr:   true,
		},
		{
			name:      "unrecognized format",
			content:   `{"foo": "bar"}`,
			wantTasks: 0,
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tmpDir := t.TempDir()
			tasksDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
			if err := os.MkdirAll(tasksDir, 0755); err != nil {
				t.Fatal(err)
			}
			
			// Write test content
			tasksFile := filepath.Join(tasksDir, "tasks.json")
			if err := os.WriteFile(tasksFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			
			// Test loading
			tasks, err := LoadTasksFromFile(tmpDir)
			
			if tt.wantErr {
				if err == nil {
					t.Error("LoadTasksFromFile() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("LoadTasksFromFile() unexpected error = %v", err)
				return
			}
			
			if len(tasks) != tt.wantTasks {
				t.Errorf("LoadTasksFromFile() got %d tasks, want %d", len(tasks), tt.wantTasks)
			}
		})
	}
}

func TestLoadTasksFromFile_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, err := LoadTasksFromFile(tmpDir)
	if err == nil {
		t.Error("LoadTasksFromFile() with missing file should return error")
	}
}

func TestLoadTasksFromFile_ComplexHierarchy(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create a complex task hierarchy
	content := `{
		"tasks": [
			{
				"id": "1",
				"title": "Root Task 1",
				"status": "in-progress",
				"priority": "high",
				"dependencies": [],
				"subtasks": [
					{
						"id": "1.1",
						"title": "Subtask 1.1",
						"status": "done",
						"priority": "medium",
						"dependencies": [],
						"subtasks": [
							{
								"id": "1.1.1",
								"title": "Deep Subtask",
								"status": "done",
								"priority": "low",
								"dependencies": [],
								"subtasks": []
							}
						]
					},
					{
						"id": "1.2",
						"title": "Subtask 1.2",
						"status": "pending",
						"priority": "medium",
						"dependencies": ["1.1"],
						"subtasks": []
					}
				]
			},
			{
				"id": "2",
				"title": "Root Task 2",
				"status": "pending",
				"priority": "medium",
				"dependencies": ["1"],
				"subtasks": []
			}
		]
	}`
	
	tasksFile := filepath.Join(tasksDir, "tasks.json")
	if err := os.WriteFile(tasksFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	
	tasks, err := LoadTasksFromFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadTasksFromFile() error = %v", err)
	}
	
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 root tasks, got %d", len(tasks))
	}
	
	// Check first task has subtasks
	if len(tasks[0].Subtasks) != 2 {
		t.Errorf("Expected 2 subtasks for task 1, got %d", len(tasks[0].Subtasks))
	}
	
	// Check deep nesting
	if len(tasks[0].Subtasks[0].Subtasks) != 1 {
		t.Errorf("Expected 1 deep subtask, got %d", len(tasks[0].Subtasks[0].Subtasks))
	}
	
	// Check task IDs
	if tasks[0].ID != "1" {
		t.Errorf("Expected task ID '1', got '%s'", tasks[0].ID)
	}
	if tasks[0].Subtasks[0].ID != "1.1" {
		t.Errorf("Expected subtask ID '1.1', got '%s'", tasks[0].Subtasks[0].ID)
	}
	if tasks[0].Subtasks[0].Subtasks[0].ID != "1.1.1" {
		t.Errorf("Expected deep subtask ID '1.1.1', got '%s'", tasks[0].Subtasks[0].Subtasks[0].ID)
	}
}

func createTestTasksJSON(t *testing.T, tmpDir string, content interface{}) {
	tasksDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	data, err := json.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	
	tasksFile := filepath.Join(tasksDir, "tasks.json")
	if err := os.WriteFile(tasksFile, data, 0644); err != nil {
		t.Fatal(err)
	}
}
