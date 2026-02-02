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

func TestNewJSONStorage(t *testing.T) {
	storage := NewJSONStorage("/path/to/file.json")
	if storage == nil {
		t.Fatal("NewJSONStorage returned nil")
	}
	if storage.FilePath() != "/path/to/file.json" {
		t.Errorf("Expected file path /path/to/file.json, got %s", storage.FilePath())
	}
}

func TestJSONStorage_LoadNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(filepath.Join(tempDir, "nonexistent.json"))

	mapping, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() should not error on non-existent file: %v", err)
	}

	// Should return empty mapping
	if mapping == nil {
		t.Fatal("Load() returned nil mapping")
	}
	if mapping.Version != SerializationVersion {
		t.Errorf("Expected version %s, got %s", SerializationVersion, mapping.Version)
	}
	if len(mapping.Tasks) != 0 {
		t.Errorf("Expected empty tasks map, got %d entries", len(mapping.Tasks))
	}
	if len(mapping.UnassignedChanges) != 0 {
		t.Errorf("Expected empty unassigned changes, got %d entries", len(mapping.UnassignedChanges))
	}
}

func TestJSONStorage_LoadEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.json")

	// Create empty file
	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	storage := NewJSONStorage(filePath)
	mapping, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() should not error on empty file: %v", err)
	}

	// Should return empty mapping
	if mapping == nil {
		t.Fatal("Load() returned nil mapping")
	}
	if mapping.Version != SerializationVersion {
		t.Errorf("Expected version %s, got %s", SerializationVersion, mapping.Version)
	}
}

func TestJSONStorage_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage := NewJSONStorage(filePath)

	// Create a test mapping
	fc1, err := NewFileChange("path/to/file1.go", "modified", "Updated function")
	if err != nil {
		t.Fatalf("Failed to create file change: %v", err)
	}
	fc2, err := NewFileChange("path/to/file2.go", "added", "New file")
	if err != nil {
		t.Fatalf("Failed to create file change: %v", err)
	}

	originalMapping := &FileChangeMapping{
		Version:     SerializationVersion,
		LastUpdated: time.Now().Truncate(time.Second), // Truncate for comparison
		Tasks: map[string][]FileChange{
			"1": {*fc1},
			"2": {*fc2},
		},
		UnassignedChanges: []FileChange{},
	}

	// Save
	if err := storage.Save(originalMapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if !storage.Exists() {
		t.Error("File should exist after Save()")
	}

	// Load
	loadedMapping, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded data
	if loadedMapping.Version != originalMapping.Version {
		t.Errorf("Version mismatch: expected %s, got %s", originalMapping.Version, loadedMapping.Version)
	}
	if len(loadedMapping.Tasks) != len(originalMapping.Tasks) {
		t.Errorf("Tasks count mismatch: expected %d, got %d", len(originalMapping.Tasks), len(loadedMapping.Tasks))
	}
	if len(loadedMapping.Tasks["1"]) != 1 {
		t.Errorf("Expected 1 change in task 1, got %d", len(loadedMapping.Tasks["1"]))
	}
	if loadedMapping.Tasks["1"][0].Path != "path/to/file1.go" {
		t.Errorf("Path mismatch: expected path/to/file1.go, got %s", loadedMapping.Tasks["1"][0].Path)
	}
}

func TestJSONStorage_SaveAtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "atomic.json")
	storage := NewJSONStorage(filePath)

	fc, err := NewFileChange("test.go", "added", "Test file")
	if err != nil {
		t.Fatalf("Failed to create file change: %v", err)
	}

	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// Save
	if err := storage.Save(mapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify temp file is removed
	tempFile := filePath + ".tmp"
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temp file should be removed after successful save")
	}

	// Verify target file exists
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("Target file should exist: %v", err)
	}
}

func TestJSONStorage_SaveCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "dir", "file.json")
	storage := NewJSONStorage(nestedPath)

	fc, err := NewFileChange("test.go", "added", "Test file")
	if err != nil {
		t.Fatalf("Failed to create file change: %v", err)
	}

	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// Save should create nested directories
	if err := storage.Save(mapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(nestedPath)); err != nil {
		t.Errorf("Directory should be created: %v", err)
	}

	// Verify file exists
	if !storage.Exists() {
		t.Error("File should exist after save")
	}
}

func TestJSONStorage_SaveInvalidMapping(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(filepath.Join(tempDir, "invalid.json"))

	// Create invalid mapping (missing version)
	invalidMapping := &FileChangeMapping{
		Version:           "", // Invalid: empty version
		Tasks:             make(map[string][]FileChange),
		UnassignedChanges: []FileChange{},
	}

	// Save should fail validation
	err := storage.Save(invalidMapping)
	if err == nil {
		t.Fatal("Save() should fail for invalid mapping")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestJSONStorage_LoadCorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "corrupted.json")

	// Write corrupted JSON
	if err := os.WriteFile(filePath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	storage := NewJSONStorage(filePath)
	_, err := storage.Load()
	if err == nil {
		t.Fatal("Load() should fail for corrupted JSON")
	}
}

func TestJSONStorage_LoadFromReader(t *testing.T) {
	storage := NewJSONStorage("dummy.json")

	fc, err := NewFileChange("test.go", "added", "Test")
	if err != nil {
		t.Fatalf("Failed to create file change: %v", err)
	}

	originalMapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(originalMapping, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Load from reader
	reader := bytes.NewReader(data)
	loadedMapping, err := storage.LoadFromReader(reader)
	if err != nil {
		t.Fatalf("LoadFromReader() failed: %v", err)
	}

	// Verify
	if loadedMapping.Version != originalMapping.Version {
		t.Errorf("Version mismatch: expected %s, got %s", originalMapping.Version, loadedMapping.Version)
	}
	if len(loadedMapping.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedMapping.Tasks))
	}
}

func TestJSONStorage_LoadFromReaderEmpty(t *testing.T) {
	storage := NewJSONStorage("dummy.json")
	reader := bytes.NewReader([]byte{})

	mapping, err := storage.LoadFromReader(reader)
	if err != nil {
		t.Fatalf("LoadFromReader() should not error on empty data: %v", err)
	}

	// Should return empty mapping
	if mapping == nil {
		t.Fatal("LoadFromReader() returned nil")
	}
	if len(mapping.Tasks) != 0 {
		t.Errorf("Expected empty tasks, got %d", len(mapping.Tasks))
	}
}

func TestJSONStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "delete.json")
	storage := NewJSONStorage(filePath)

	// Create a file
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	if err := storage.Save(mapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if !storage.Exists() {
		t.Fatal("File should exist before delete")
	}

	// Delete
	if err := storage.Delete(); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify file is gone
	if storage.Exists() {
		t.Error("File should not exist after delete")
	}
}

func TestJSONStorage_DeleteNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(filepath.Join(tempDir, "nonexistent.json"))

	// Delete non-existent file should not error
	if err := storage.Delete(); err != nil {
		t.Errorf("Delete() should not error on non-existent file: %v", err)
	}
}

func TestJSONStorage_Exists(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "exists.json")
	storage := NewJSONStorage(filePath)

	// File doesn't exist yet
	if storage.Exists() {
		t.Error("Exists() should return false for non-existent file")
	}

	// Create file
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	if err := storage.Save(mapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// File should exist now
	if !storage.Exists() {
		t.Error("Exists() should return true for existing file")
	}
}

func TestJSONStorage_SaveUpdatesLastUpdated(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(filepath.Join(tempDir, "timestamp.json"))

	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Time{}, // Zero time
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	beforeSave := time.Now()

	// Save
	if err := storage.Save(mapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	afterSave := time.Now()

	// Verify LastUpdated was set during save
	if mapping.LastUpdated.Before(beforeSave) || mapping.LastUpdated.After(afterSave) {
		t.Errorf("LastUpdated not updated correctly: %v (should be between %v and %v)",
			mapping.LastUpdated, beforeSave, afterSave)
	}
}

func TestJSONStorage_SaveWithInvalidFileChanges(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(filepath.Join(tempDir, "invalid_changes.json"))

	// Create mapping with invalid file change
	mapping := &FileChangeMapping{
		Version:     SerializationVersion,
		LastUpdated: time.Now(),
		Tasks: map[string][]FileChange{
			"1": {{
				Path:        "", // Invalid: empty path
				ChangeType:  "modified",
				Description: "Test",
			}},
		},
		UnassignedChanges: []FileChange{},
	}

	// Save should fail validation
	err := storage.Save(mapping)
	if err == nil {
		t.Fatal("Save() should fail for invalid file changes")
	}
}

func TestJSONStorage_LoadWithSchemaVersion(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "versioned.json")

	// Write file with specific version
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           "1.0",
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Load
	storage := NewJSONStorage(filePath)
	loadedMapping, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify version is preserved
	if loadedMapping.Version != "1.0" {
		t.Errorf("Version mismatch: expected 1.0, got %s", loadedMapping.Version)
	}
}

func TestJSONStorage_IntegrationRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(filepath.Join(tempDir, "roundtrip.json"))

	// Create complex mapping
	fc1, _ := NewFileChange("file1.go", "added", "New file")
	fc2, _ := NewFileChange("file2.go", "modified", "Updated")
	fc3, _ := NewFileChange("file3.go", "deleted", "Removed")
	fc4, _ := NewFileChange("file4.go", "added", "Unassigned")

	fc1.CommitID = "abc123"
	fc1.IsPending = false
	fc2.IsPending = true

	originalMapping := &FileChangeMapping{
		Version:     SerializationVersion,
		LastUpdated: time.Now().Truncate(time.Second),
		Tasks: map[string][]FileChange{
			"1":     {*fc1, *fc2},
			"2":     {*fc3},
			"3.1.2": {},
		},
		UnassignedChanges: []FileChange{*fc4},
	}

	// Save
	if err := storage.Save(originalMapping); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load
	loadedMapping, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Comprehensive verification
	if loadedMapping.Version != originalMapping.Version {
		t.Errorf("Version mismatch")
	}
	if len(loadedMapping.Tasks) != len(originalMapping.Tasks) {
		t.Errorf("Tasks count mismatch: expected %d, got %d", len(originalMapping.Tasks), len(loadedMapping.Tasks))
	}
	if len(loadedMapping.UnassignedChanges) != len(originalMapping.UnassignedChanges) {
		t.Errorf("Unassigned changes count mismatch: expected %d, got %d", len(originalMapping.UnassignedChanges), len(loadedMapping.UnassignedChanges))
	}

	// Verify specific file change details
	if loadedMapping.Tasks["1"][0].CommitID != "abc123" {
		t.Errorf("CommitID not preserved: expected abc123, got %s", loadedMapping.Tasks["1"][0].CommitID)
	}
	if loadedMapping.Tasks["1"][0].IsPending != false {
		t.Error("IsPending not preserved")
	}
	if loadedMapping.Tasks["1"][1].IsPending != true {
		t.Error("IsPending not preserved for second change")
	}
}
