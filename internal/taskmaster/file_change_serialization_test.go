package taskmaster

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMarshalFileChangeMapping(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:        "src/main.go",
		ChangeType:  "modified",
		Description: "Updated main logic",
		LastChanged: time.Now(),
		CommitID:    "abc123",
		IsPending:   false,
	}
	fcm.AddFileChange("task-1", fc)

	data, err := MarshalFileChangeMapping(fcm)
	if err != nil {
		t.Fatalf("MarshalFileChangeMapping() error = %v", err)
	}

	// Verify it's valid JSON
	var result FileChangeMapping
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	if result.Version != "1.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.0")
	}
}

func TestMarshalFileChangeMappingIndented(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}
	fcm.AddFileChange("task-1", fc)

	data, err := MarshalFileChangeMappingIndented(fcm)
	if err != nil {
		t.Fatalf("MarshalFileChangeMappingIndented() error = %v", err)
	}

	// Verify indentation
	if !strings.Contains(string(data), "  ") {
		t.Error("Indented output should contain spaces")
	}

	// Verify it's valid JSON
	var result FileChangeMapping
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}
}

func TestUnmarshalFileChangeMapping(t *testing.T) {
	original := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:        "src/main.go",
		ChangeType:  "modified",
		Description: "Test change",
		LastChanged: time.Now(),
		IsPending:   true,
	}
	original.AddFileChange("task-1", fc)

	// Marshal and unmarshal
	data, err := MarshalFileChangeMapping(original)
	if err != nil {
		t.Fatalf("MarshalFileChangeMapping() error = %v", err)
	}

	result, err := UnmarshalFileChangeMapping(data)
	if err != nil {
		t.Fatalf("UnmarshalFileChangeMapping() error = %v", err)
	}

	// Verify the structure
	if result.Version != "1.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.0")
	}
	if len(result.Tasks) != 1 {
		t.Errorf("Tasks count = %d, want 1", len(result.Tasks))
	}
	if len(result.Tasks["task-1"]) != 1 {
		t.Errorf("Task-1 changes count = %d, want 1", len(result.Tasks["task-1"]))
	}
}

func TestUnmarshalFileChangeMappingInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)

	_, err := UnmarshalFileChangeMapping(invalidJSON)
	if err == nil {
		t.Error("UnmarshalFileChangeMapping() should return error for invalid JSON")
	}
}

func TestUnmarshalFileChangeMappingInvalidData(t *testing.T) {
	// Valid JSON but invalid FileChangeMapping (empty version)
	invalidData := []byte(`{"version":"","tasks":{},"unassignedChanges":[]}`)

	_, err := UnmarshalFileChangeMapping(invalidData)
	if err == nil {
		t.Error("UnmarshalFileChangeMapping() should return error for invalid data")
	}
}

func TestUnmarshalFileChangeMappingFromReader(t *testing.T) {
	original := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}
	original.AddFileChange("task-1", fc)

	// Marshal to bytes
	data, err := MarshalFileChangeMapping(original)
	if err != nil {
		t.Fatalf("MarshalFileChangeMapping() error = %v", err)
	}

	// Create reader from bytes
	reader := bytes.NewReader(data)

	// Unmarshal from reader
	result, err := UnmarshalFileChangeMappingFromReader(reader)
	if err != nil {
		t.Fatalf("UnmarshalFileChangeMappingFromReader() error = %v", err)
	}

	if result.Version != "1.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.0")
	}
}

func TestUnmarshalFileChangeMappingFromReaderInvalid(t *testing.T) {
	invalidReader := strings.NewReader(`{invalid json}`)

	_, err := UnmarshalFileChangeMappingFromReader(invalidReader)
	if err == nil {
		t.Error("UnmarshalFileChangeMappingFromReader() should return error for invalid JSON")
	}
}

func TestMarshalFileChange(t *testing.T) {
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
		IsPending:  true,
	}

	data, err := MarshalFileChange(fc)
	if err != nil {
		t.Fatalf("MarshalFileChange() error = %v", err)
	}

	// Verify it's valid JSON
	var result FileChange
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	if result.Path != "src/main.go" {
		t.Errorf("Path = %q, want %q", result.Path, "src/main.go")
	}
}

func TestMarshalFileChangeInvalid(t *testing.T) {
	fc := FileChange{
		Path:       "", // Invalid: empty path
		ChangeType: "modified",
	}

	_, err := MarshalFileChange(fc)
	if err == nil {
		t.Error("MarshalFileChange() should return error for invalid FileChange")
	}
}

func TestUnmarshalFileChange(t *testing.T) {
	original := FileChange{
		Path:        "src/main.go",
		ChangeType:  "modified",
		Description: "Test change",
		LastChanged: time.Now(),
		CommitID:    "abc123",
		IsPending:   false,
	}

	// Marshal and unmarshal
	data, err := MarshalFileChange(original)
	if err != nil {
		t.Fatalf("MarshalFileChange() error = %v", err)
	}

	result, err := UnmarshalFileChange(data)
	if err != nil {
		t.Fatalf("UnmarshalFileChange() error = %v", err)
	}

	if result.Path != original.Path {
		t.Errorf("Path = %q, want %q", result.Path, original.Path)
	}
	if result.ChangeType != original.ChangeType {
		t.Errorf("ChangeType = %q, want %q", result.ChangeType, original.ChangeType)
	}
}

func TestMarshalTaskWithChanges(t *testing.T) {
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

	data, err := MarshalTaskWithChanges(&task)
	if err != nil {
		t.Fatalf("MarshalTaskWithChanges() error = %v", err)
	}

	// Verify it's valid JSON
	var result Task
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	if result.ID != "1.1" {
		t.Errorf("ID = %q, want %q", result.ID, "1.1")
	}
	if len(result.FileChanges) != 1 {
		t.Errorf("FileChanges count = %d, want 1", len(result.FileChanges))
	}
}

func TestMarshalTaskWithChangesIndented(t *testing.T) {
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

	data, err := MarshalTaskWithChangesIndented(&task)
	if err != nil {
		t.Fatalf("MarshalTaskWithChangesIndented() error = %v", err)
	}

	// Verify indentation
	if !strings.Contains(string(data), "  ") {
		t.Error("Indented output should contain spaces")
	}
}

func TestUnmarshalTaskWithChanges(t *testing.T) {
	original := Task{
		ID:    "1.1",
		Title: "Test Task",
		FileChanges: []FileChange{
			{
				Path:       "src/main.go",
				ChangeType: "modified",
				Description: "Updated main",
			},
		},
	}

	// Marshal and unmarshal
	data, err := MarshalTaskWithChanges(&original)
	if err != nil {
		t.Fatalf("MarshalTaskWithChanges() error = %v", err)
	}

	result, err := UnmarshalTaskWithChanges(data)
	if err != nil {
		t.Fatalf("UnmarshalTaskWithChanges() error = %v", err)
	}

	if result.ID != "1.1" {
		t.Errorf("ID = %q, want %q", result.ID, "1.1")
	}
	if result.FileChanges[0].Path != "src/main.go" {
		t.Errorf("FileChange path = %q, want %q", result.FileChanges[0].Path, "src/main.go")
	}
}

func TestLoadFileChangesFromJSON(t *testing.T) {
	jsonData := `{
		"version": "1.0",
		"lastUpdated": "2024-01-15T10:30:00Z",
		"tasks": {
			"task-1": [
				{
					"path": "src/main.go",
					"changeType": "modified"
				}
			]
		},
		"unassignedChanges": []
	}`

	result, err := LoadFileChangesFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFileChangesFromJSON() error = %v", err)
	}

	if result.Version != "1.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.0")
	}
	if len(result.Tasks) != 1 {
		t.Errorf("Tasks count = %d, want 1", len(result.Tasks))
	}
}

func TestDumpFileChangesToJSON(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}
	fcm.AddFileChange("task-1", fc)

	jsonStr, err := DumpFileChangesToJSON(fcm)
	if err != nil {
		t.Fatalf("DumpFileChangesToJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var result FileChangeMapping
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	// Verify it's indented
	if !strings.Contains(jsonStr, "\n") {
		t.Error("JSON should be indented with newlines")
	}
}

func TestSaveAndLoadFileChangeMapping(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test_changes.json")

	// Create a FileChangeMapping with some data
	original := NewFileChangeMapping("1.0")
	fc1 := FileChange{
		Path:        "src/main.go",
		ChangeType:  "modified",
		Description: "Updated main",
		CommitID:    "abc123",
		IsPending:   false,
	}
	fc2 := FileChange{
		Path:       "docs/readme.md",
		ChangeType: "added",
	}
	original.AddFileChange("task-1", fc1)
	original.AddUnassignedChange(fc2)

	// Save to file
	if err := SaveFileChangeMapping(filename, original); err != nil {
		t.Fatalf("SaveFileChangeMapping() error = %v", err)
	}

	// Load from file
	loaded, err := LoadFileChangeMapping(filename)
	if err != nil {
		t.Fatalf("LoadFileChangeMapping() error = %v", err)
	}

	// Verify the loaded data matches original
	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: got %q, want %q", loaded.Version, original.Version)
	}
	if len(loaded.Tasks) != len(original.Tasks) {
		t.Errorf("Tasks count mismatch: got %d, want %d", len(loaded.Tasks), len(original.Tasks))
	}
	if len(loaded.UnassignedChanges) != len(original.UnassignedChanges) {
		t.Errorf("Unassigned changes count mismatch: got %d, want %d", len(loaded.UnassignedChanges), len(original.UnassignedChanges))
	}
}

func TestSaveAndLoadFileChangeMappingIndented(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test_changes_indented.json")

	// Create a FileChangeMapping with some data
	original := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}
	original.AddFileChange("task-1", fc)

	// Save with indentation
	if err := SaveFileChangeMappingIndented(filename, original); err != nil {
		t.Fatalf("SaveFileChangeMappingIndented() error = %v", err)
	}

	// Read and verify indentation
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "  ") {
		t.Error("Saved file should contain indentation")
	}

	// Load and verify data
	loaded, err := LoadFileChangeMapping(filename)
	if err != nil {
		t.Fatalf("LoadFileChangeMapping() error = %v", err)
	}

	if loaded.Version != "1.0" {
		t.Errorf("Version = %q, want %q", loaded.Version, "1.0")
	}
}

func TestLoadFileChangeMappingNonExistent(t *testing.T) {
	_, err := LoadFileChangeMapping("/nonexistent/file/path.json")
	if err == nil {
		t.Error("LoadFileChangeMapping() should return error for nonexistent file")
	}
}

func TestCheckAndMigrateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "empty version",
			version: "",
			wantErr: false, // Should default to current version
		},
		{
			name:    "matching version",
			version: "1.0",
			wantErr: false,
		},
		{
			name:    "mismatched version",
			version: "2.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fcm := NewFileChangeMapping(tt.version)
			err := checkAndMigrateVersion(fcm)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkAndMigrateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigrationHelper(t *testing.T) {
	mh := NewMigrationHelper("1.0", "1.0")
	fcm := NewFileChangeMapping("1.0")
	fc := FileChange{
		Path:       "src/main.go",
		ChangeType: "modified",
	}
	fcm.AddFileChange("task-1", fc)

	// Migration should succeed
	err := mh.MigrateFileChangeMapping(fcm)
	if err != nil {
		t.Fatalf("MigrateFileChangeMapping() error = %v", err)
	}

	if fcm.Version != "1.0" {
		t.Errorf("Version = %q, want %q", fcm.Version, "1.0")
	}
}

func TestGetStorageStats(t *testing.T) {
	fcm := NewFileChangeMapping("1.0")

	// Add changes
	fc1 := FileChange{Path: "file1.go", ChangeType: "modified", IsPending: true}
	fc2 := FileChange{Path: "file2.go", ChangeType: "added", IsPending: false, CommitID: "abc"}
	fcm.AddFileChange("task-1", fc1)
	fcm.AddFileChange("task-1", fc2)

	fc3 := FileChange{Path: "file3.go", ChangeType: "modified", IsPending: true}
	fcm.AddUnassignedChange(fc3)

	// Get stats
	stats := fcm.GetStorageStats()

	if stats.TotalChanges != 3 {
		t.Errorf("TotalChanges = %d, want 3", stats.TotalChanges)
	}
	if stats.TotalTasks != 1 {
		t.Errorf("TotalTasks = %d, want 1", stats.TotalTasks)
	}
	if stats.PendingChanges != 2 {
		t.Errorf("PendingChanges = %d, want 2", stats.PendingChanges)
	}
	if stats.CommittedChanges != 1 {
		t.Errorf("CommittedChanges = %d, want 1", stats.CommittedChanges)
	}
	if stats.UnassignedChanges != 1 {
		t.Errorf("UnassignedChanges = %d, want 1", stats.UnassignedChanges)
	}
}

func TestSerializationRoundTrip(t *testing.T) {
	// Create a complex FileChangeMapping
	original := NewFileChangeMapping("1.0")

	// Add changes for multiple tasks
	for taskID := 1; taskID <= 3; taskID++ {
		for i := 0; i < 3; i++ {
			fc := FileChange{
				Path:        "file" + string(rune(i)) + ".go",
				ChangeType:  []string{"added", "modified", "deleted"}[i%3],
				Description: "Test change",
				LastChanged: time.Now(),
				CommitID:    "commit" + string(rune(taskID*10+i)),
				IsPending:   i%2 == 0,
			}
			original.AddFileChange("task-"+string(rune(48+taskID)), fc)
		}
	}

	// Add unassigned changes
	for i := 0; i < 2; i++ {
		fc := FileChange{
			Path:       "unassigned" + string(rune(48+i)) + ".go",
			ChangeType: "modified",
			IsPending:  true,
		}
		original.AddUnassignedChange(fc)
	}

	// Serialize and deserialize
	data, err := MarshalFileChangeMappingIndented(original)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}

	recovered, err := UnmarshalFileChangeMapping(data)
	if err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	// Verify all data is preserved
	if recovered.Version != original.Version {
		t.Errorf("Version mismatch")
	}
	if len(recovered.Tasks) != len(original.Tasks) {
		t.Errorf("Tasks count mismatch: got %d, want %d", len(recovered.Tasks), len(original.Tasks))
	}
	if recovered.TotalChangesCount() != original.TotalChangesCount() {
		t.Errorf("Total changes mismatch: got %d, want %d", recovered.TotalChangesCount(), original.TotalChangesCount())
	}
}

func TestSaveInvalidFileChangeMapping(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "invalid.json")

	// Create an invalid FileChangeMapping (empty version)
	invalid := &FileChangeMapping{
		Version: "", // Invalid
		Tasks:   make(map[string][]FileChange),
	}

	err := SaveFileChangeMapping(filename, invalid)
	if err == nil {
		t.Error("SaveFileChangeMapping() should return error for invalid mapping")
	}
}

// Benchmark tests
func BenchmarkMarshalFileChangeMapping(b *testing.B) {
	fcm := NewFileChangeMapping("1.0")
	for i := 0; i < 100; i++ {
		fc := FileChange{
			Path:       "file" + string(rune(i)) + ".go",
			ChangeType: "modified",
		}
		fcm.AddFileChange("task-1", fc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalFileChangeMapping(fcm)
	}
}

func BenchmarkUnmarshalFileChangeMapping(b *testing.B) {
	fcm := NewFileChangeMapping("1.0")
	for i := 0; i < 100; i++ {
		fc := FileChange{
			Path:       "file" + string(rune(i)) + ".go",
			ChangeType: "modified",
		}
		fcm.AddFileChange("task-1", fc)
	}

	data, _ := MarshalFileChangeMapping(fcm)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = UnmarshalFileChangeMapping(data)
	}
}
