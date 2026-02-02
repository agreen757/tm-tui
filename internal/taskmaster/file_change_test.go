package taskmaster

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFileChangeValidation(t *testing.T) {
	tests := []struct {
		name    string
		fc      FileChange
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid file change",
			fc: FileChange{
				Path:        "src/main.go",
				ChangeType:  "modified",
				Description: "Updated main logic",
				LastChanged: time.Now(),
				IsPending:   true,
			},
			wantErr: false,
		},
		{
			name: "valid added file",
			fc: FileChange{
				Path:       "src/new_file.go",
				ChangeType: "added",
			},
			wantErr: false,
		},
		{
			name: "valid deleted file",
			fc: FileChange{
				Path:       "src/old_file.go",
				ChangeType: "deleted",
			},
			wantErr: false,
		},
		{
			name: "empty path",
			fc: FileChange{
				Path:       "",
				ChangeType: "modified",
			},
			wantErr: true,
			errMsg:  "path cannot be empty",
		},
		{
			name: "empty change type",
			fc: FileChange{
				Path:       "src/main.go",
				ChangeType: "",
			},
			wantErr: true,
			errMsg:  "type cannot be empty",
		},
		{
			name: "invalid change type",
			fc: FileChange{
				Path:       "src/main.go",
				ChangeType: "renamed",
			},
			wantErr: true,
			errMsg:  "invalid change type",
		},
		{
			name: "special characters in path",
			fc: FileChange{
				Path:       "src/file-with-dashes_and_underscores.go",
				ChangeType: "modified",
			},
			wantErr: false,
		},
		{
			name: "nested directory path",
			fc: FileChange{
				Path:       "pkg/internal/deeply/nested/file.go",
				ChangeType: "added",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestNewFileChange(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		changeType  string
		description string
		wantErr     bool
	}{
		{
			name:        "valid file change creation",
			path:        "src/main.go",
			changeType:  "modified",
			description: "Updated logic",
			wantErr:     false,
		},
		{
			name:       "invalid change type",
			path:       "src/main.go",
			changeType: "invalid",
			wantErr:    true,
		},
		{
			name:       "empty path",
			path:       "",
			changeType: "modified",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := NewFileChange(tt.path, tt.changeType, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileChange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if fc.Path != tt.path {
					t.Errorf("Path = %q, want %q", fc.Path, tt.path)
				}
				if fc.ChangeType != tt.changeType {
					t.Errorf("ChangeType = %q, want %q", fc.ChangeType, tt.changeType)
				}
				if fc.Description != tt.description {
					t.Errorf("Description = %q, want %q", fc.Description, tt.description)
				}
				if !fc.IsPending {
					t.Error("IsPending should be true for new changes")
				}
				if fc.LastChanged.IsZero() {
					t.Error("LastChanged should be set to current time")
				}
			}
		})
	}
}

func TestFileChangeJSONMarshaling(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	fc := FileChange{
		Path:        "src/main.go",
		ChangeType:  "modified",
		Description: "Updated main function",
		LastChanged: now,
		CommitID:    "abc123def456",
		IsPending:   false,
	}

	// Marshal to JSON
	data, err := json.Marshal(fc)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var fc2 FileChange
	if err := json.Unmarshal(data, &fc2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify all fields are preserved
	if fc2.Path != fc.Path {
		t.Errorf("Path mismatch: got %q, want %q", fc2.Path, fc.Path)
	}
	if fc2.ChangeType != fc.ChangeType {
		t.Errorf("ChangeType mismatch: got %q, want %q", fc2.ChangeType, fc.ChangeType)
	}
	if fc2.Description != fc.Description {
		t.Errorf("Description mismatch: got %q, want %q", fc2.Description, fc.Description)
	}
	if fc2.CommitID != fc.CommitID {
		t.Errorf("CommitID mismatch: got %q, want %q", fc2.CommitID, fc.CommitID)
	}
	if fc2.IsPending != fc.IsPending {
		t.Errorf("IsPending mismatch: got %v, want %v", fc2.IsPending, fc.IsPending)
	}
}

func TestFileChangeMappingValidation(t *testing.T) {
	tests := []struct {
		name    string
		fcm     *FileChangeMapping
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid empty mapping",
			fcm: &FileChangeMapping{
				Version:           "1.0",
				LastUpdated:       time.Now(),
				Tasks:             make(map[string][]FileChange),
				UnassignedChanges: make([]FileChange, 0),
			},
			wantErr: false,
		},
		{
			name: "empty version",
			fcm: &FileChangeMapping{
				Version:           "",
				Tasks:             make(map[string][]FileChange),
				UnassignedChanges: make([]FileChange, 0),
			},
			wantErr: true,
			errMsg:  "version cannot be empty",
		},
		{
			name: "nil tasks map",
			fcm: &FileChangeMapping{
				Version:           "1.0",
				Tasks:             nil,
				UnassignedChanges: make([]FileChange, 0),
			},
			wantErr: false, // Should be fixed during validation
		},
		{
			name: "nil unassigned changes",
			fcm: &FileChangeMapping{
				Version:           "1.0",
				Tasks:             make(map[string][]FileChange),
				UnassignedChanges: nil,
			},
			wantErr: false, // Should be fixed during validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fcm.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestNewFileChangeMapping(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")

	if fcm.Version != "1.0" {
		t.Errorf("Version = %q, want %q", fcm.Version, "1.0")
	}
	if fcm.Tasks == nil {
		t.Error("Tasks should not be nil")
	}
	if fcm.UnassignedChanges == nil {
		t.Error("UnassignedChanges should not be nil")
	}
	if fcm.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set")
	}
}

func TestAddFileChange(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}

	err := fcm.AddFileChange("task-1", fc)
	if err != nil {
		t.Fatalf("AddFileChange() error = %v", err)
	}

	changes := fcm.GetChangesForTask("task-1")
	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}
	if changes[0].Path != "src/main.go" {
		t.Errorf("Path = %q, want %q", changes[0].Path, "src/main.go")
	}
}

func TestAddFileChangeWithInvalidChange(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "", // Invalid: empty path
		ChangeType: "modified",
	}

	err := fcm.AddFileChange("task-1", fc)
	if err == nil {
		t.Error("AddFileChange() should return error for invalid FileChange")
	}
}

func TestAddFileChangeWithEmptyTaskID(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}

	err := fcm.AddFileChange("", fc)
	if err == nil {
		t.Error("AddFileChange() should return error for empty task ID")
	}
}

func TestAddUnassignedChange(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "added",
	}

	err := fcm.AddUnassignedChange(fc)
	if err != nil {
		t.Fatalf("AddUnassignedChange() error = %v", err)
	}

	if len(fcm.UnassignedChanges) != 1 {
		t.Errorf("Expected 1 unassigned change, got %d", len(fcm.UnassignedChanges))
	}
}

func TestGetPendingChanges(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")

	// Add pending changes
	fc1 := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
		IsPending:  true,
	}
	fcm.AddFileChange("task-1", fc1)

	// Add committed changes
	fc2 := FileChange{
		Path:       "src/helper.go",
		ChangeType: "added",
		IsPending:  false,
		CommitID:   "abc123",
	}
	fcm.AddFileChange("task-1", fc2)

	// Add unassigned pending change
	fc3 := FileChange{
		Path:       "docs/readme.md",
		ChangeType: "modified",
		IsPending:  true,
	}
	fcm.AddUnassignedChange(fc3)

	pending := fcm.GetPendingChanges()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending changes, got %d", len(pending))
	}

	// Verify all pending changes are included
	pendingPaths := make(map[string]bool)
	for _, p := range pending {
		pendingPaths[p.Path] = true
	}
	if !pendingPaths["src/main.go"] {
		t.Error("Missing pending change for src/main.go")
	}
	if !pendingPaths["docs/readme.md"] {
		t.Error("Missing pending change for docs/readme.md")
	}
	if pendingPaths["src/helper.go"] {
		t.Error("Should not include committed change src/helper.go")
	}
}

func TestTotalChangesCount(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")

	// Add changes to task-1
	fcm.AddFileChange("task-1", FileChange{Path: "file1.go", ChangeType: "modified"})
	fcm.AddFileChange("task-1", FileChange{Path: "file2.go", ChangeType: "added"})

	// Add changes to task-2
	fcm.AddFileChange("task-2", FileChange{Path: "file3.go", ChangeType: "deleted"})

	// Add unassigned changes
	fcm.AddUnassignedChange(FileChange{Path: "file4.go", ChangeType: "modified"})
	fcm.AddUnassignedChange(FileChange{Path: "file5.go", ChangeType: "added"})

	count := fcm.TotalChangesCount()
	if count != 5 {
		t.Errorf("TotalChangesCount() = %d, want 5", count)
	}
}

func TestFileChangeMappingJSONMarshaling(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// Add changes
	fc1 := FileChange{
		Path:        "src/main.go",
		ChangeType:  "modified",
		Description: "Updated main",
		LastChanged: now,
		CommitID:    "abc123",
		IsPending:   false,
	}
	fcm.AddFileChange("task-1", fc1)

	fc2 := FileChange{
		Path:       "docs/readme.md",
		ChangeType: "added",
	}
	fcm.AddUnassignedChange(fc2)

	// Marshal to JSON
	data, err := json.Marshal(fcm)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var fcm2 FileChangeMapping
	if err := json.Unmarshal(data, &fcm2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify structure is preserved
	if fcm2.Version != "1.0" {
		t.Errorf("Version = %q, want %q", fcm2.Version, "1.0")
	}
	if len(fcm2.Tasks) != 1 {
		t.Errorf("Tasks length = %d, want 1", len(fcm2.Tasks))
	}
	if len(fcm2.UnassignedChanges) != 1 {
		t.Errorf("UnassignedChanges length = %d, want 1", len(fcm2.UnassignedChanges))
	}

	// Verify task changes
	taskChanges := fcm2.Tasks["task-1"]
	if len(taskChanges) != 1 {
		t.Errorf("Task changes length = %d, want 1", len(taskChanges))
	}
	if taskChanges[0].Path != "src/main.go" {
		t.Errorf("Task change path = %q, want %q", taskChanges[0].Path, "src/main.go")
	}

	// Verify unassigned changes
	if fcm2.UnassignedChanges[0].Path != "docs/readme.md" {
		t.Errorf("Unassigned change path = %q, want %q", fcm2.UnassignedChanges[0].Path, "docs/readme.md")
	}
}

func TestTaskWithFileChanges(t *testing.T) {
	task := Task{
		ID:    "1.1",
		Title: "Test Task",
		FileChanges: []FileChange{
			{
				Path:       "src/main.go",
				ChangeType: "modified",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var task2 Task
	if err := json.Unmarshal(data, &task2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify FileChanges are preserved
	if len(task2.FileChanges) != 1 {
		t.Errorf("FileChanges length = %d, want 1", len(task2.FileChanges))
	}
	if task2.FileChanges[0].Path != "src/main.go" {
		t.Errorf("FileChange path = %q, want %q", task2.FileChanges[0].Path, "src/main.go")
	}
}

func TestLargeDatasetBenchmark(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")

	// Add a large number of changes
	for i := 0; i < 1000; i++ {
		taskID := "task-1"
		fc := FileChange{
			Path:       "src/file" + string(rune(i)) + ".go",
			ChangeType: "modified",
		}
		fcm.AddFileChange(taskID, fc)
	}

	// Test performance of operations
	changes := fcm.GetChangesForTask("task-1")
	if len(changes) != 1000 {
		t.Errorf("Expected 1000 changes, got %d", len(changes))
	}

	count := fcm.TotalChangesCount()
	if count != 1000 {
		t.Errorf("TotalChangesCount() = %d, want 1000", count)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && s[0:len(substr)] == substr || len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
