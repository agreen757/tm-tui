package taskmaster

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// SerializationVersion defines the current schema version for file change tracking
const SerializationVersion = "1.0"

// LoadFileChangeMapping loads a FileChangeMapping from a JSON file
func LoadFileChangeMapping(filename string) (*FileChangeMapping, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filename, err)
	}

	fcm, err := UnmarshalFileChangeMapping(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal FileChangeMapping from %q: %w", filename, err)
	}

	return fcm, nil
}

// SaveFileChangeMapping saves a FileChangeMapping to a JSON file
func SaveFileChangeMapping(filename string, fcm *FileChangeMapping) error {
	if err := fcm.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid FileChangeMapping: %w", err)
	}

	data, err := MarshalFileChangeMapping(fcm)
	if err != nil {
		return fmt.Errorf("failed to marshal FileChangeMapping: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", filename, err)
	}

	return nil
}

// SaveFileChangeMappingIndented saves a FileChangeMapping to a JSON file with indentation
func SaveFileChangeMappingIndented(filename string, fcm *FileChangeMapping) error {
	if err := fcm.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid FileChangeMapping: %w", err)
	}

	data, err := MarshalFileChangeMappingIndented(fcm)
	if err != nil {
		return fmt.Errorf("failed to marshal FileChangeMapping: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", filename, err)
	}

	return nil
}

// MarshalFileChangeMapping marshals a FileChangeMapping to JSON bytes
func MarshalFileChangeMapping(fcm *FileChangeMapping) ([]byte, error) {
	if err := fcm.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid FileChangeMapping: %w", err)
	}

	data, err := json.Marshal(fcm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileChangeMapping: %w", err)
	}

	return data, nil
}

// MarshalFileChangeMappingIndented marshals a FileChangeMapping to indented JSON bytes
func MarshalFileChangeMappingIndented(fcm *FileChangeMapping) ([]byte, error) {
	if err := fcm.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid FileChangeMapping: %w", err)
	}

	data, err := json.MarshalIndent(fcm, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileChangeMapping: %w", err)
	}

	return data, nil
}

// UnmarshalFileChangeMapping unmarshals JSON bytes into a FileChangeMapping
func UnmarshalFileChangeMapping(data []byte) (*FileChangeMapping, error) {
	fcm := &FileChangeMapping{
		Tasks:             make(map[string][]FileChange),
		UnassignedChanges: make([]FileChange, 0),
	}

	if err := json.Unmarshal(data, fcm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate the deserialized data
	if err := fcm.Validate(); err != nil {
		return nil, fmt.Errorf("deserialized data is invalid: %w", err)
	}

	// Check schema version
	if err := checkAndMigrateVersion(fcm); err != nil {
		return nil, fmt.Errorf("version check failed: %w", err)
	}

	return fcm, nil
}

// UnmarshalFileChangeMappingFromReader unmarshals JSON from an io.Reader into a FileChangeMapping
func UnmarshalFileChangeMappingFromReader(reader io.Reader) (*FileChangeMapping, error) {
	fcm := &FileChangeMapping{
		Tasks:             make(map[string][]FileChange),
		UnassignedChanges: make([]FileChange, 0),
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(fcm); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate the deserialized data
	if err := fcm.Validate(); err != nil {
		return nil, fmt.Errorf("deserialized data is invalid: %w", err)
	}

	// Check schema version
	if err := checkAndMigrateVersion(fcm); err != nil {
		return nil, fmt.Errorf("version check failed: %w", err)
	}

	return fcm, nil
}

// checkAndMigrateVersion checks the schema version and performs migrations if needed
func checkAndMigrateVersion(fcm *FileChangeMapping) error {
	if fcm.Version == "" {
		// Default to current version if not specified
		fcm.Version = SerializationVersion
		return nil
	}

	// Check if version is compatible
	if fcm.Version == SerializationVersion {
		return nil
	}

	// Version 1.0 to 1.0 migration (no-op for now)
	// Future versions can add migration logic here
	// For now, we accept the current version and warn about mismatches

	if fcm.Version != SerializationVersion {
		// Log a warning but don't fail - allows reading of potentially future versions
		return fmt.Errorf("version mismatch: file version is %q, expected %q (may cause compatibility issues)", fcm.Version, SerializationVersion)
	}

	return nil
}

// LoadFileChangesFromJSON loads file changes from raw JSON data
func LoadFileChangesFromJSON(jsonData string) (*FileChangeMapping, error) {
	return UnmarshalFileChangeMapping([]byte(jsonData))
}

// DumpFileChangesToJSON dumps a FileChangeMapping to a formatted JSON string
func DumpFileChangesToJSON(fcm *FileChangeMapping) (string, error) {
	data, err := MarshalFileChangeMappingIndented(fcm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MarshalFileChange marshals a single FileChange to JSON
func MarshalFileChange(fc FileChange) ([]byte, error) {
	if err := fc.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid FileChange: %w", err)
	}

	data, err := json.Marshal(fc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileChange: %w", err)
	}

	return data, nil
}

// UnmarshalFileChange unmarshals JSON bytes into a FileChange
func UnmarshalFileChange(data []byte) (*FileChange, error) {
	fc := &FileChange{}

	if err := json.Unmarshal(data, fc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate the deserialized data
	if err := fc.Validate(); err != nil {
		return nil, fmt.Errorf("deserialized FileChange is invalid: %w", err)
	}

	return fc, nil
}

// MarshalTaskWithChanges marshals a Task with its FileChanges to JSON
func MarshalTaskWithChanges(task *Task) ([]byte, error) {
	data, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Task: %w", err)
	}

	return data, nil
}

// MarshalTaskWithChangesIndented marshals a Task with its FileChanges to indented JSON
func MarshalTaskWithChangesIndented(task *Task) ([]byte, error) {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Task: %w", err)
	}

	return data, nil
}

// UnmarshalTaskWithChanges unmarshals JSON bytes into a Task with FileChanges
func UnmarshalTaskWithChanges(data []byte) (*Task, error) {
	task := &Task{}

	if err := json.Unmarshal(data, task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate FileChanges if present
	for i, fc := range task.FileChanges {
		if err := fc.Validate(); err != nil {
			return nil, fmt.Errorf("invalid FileChange at index %d: %w", i, err)
		}
	}

	return task, nil
}

// MigrationHelper provides methods for handling data format migrations
type MigrationHelper struct {
	sourceVersion string
	targetVersion string
}

// NewMigrationHelper creates a new MigrationHelper for a specific version upgrade
func NewMigrationHelper(sourceVersion string, targetVersion string) *MigrationHelper {
	return &MigrationHelper{
		sourceVersion: sourceVersion,
		targetVersion: targetVersion,
	}
}

// MigrateFileChangeMapping performs version migration on a FileChangeMapping
func (mh *MigrationHelper) MigrateFileChangeMapping(fcm *FileChangeMapping) error {
	if fcm.Version == mh.targetVersion {
		return nil // Already at target version
	}

	// Implement migration logic based on source and target versions
	// For now, just update the version and timestamp
	fcm.Version = mh.targetVersion
	fcm.LastUpdated = time.Now()

	return fcm.Validate()
}

// GetMigrationPath returns the migration path for a version upgrade
func GetMigrationPath(from string, to string) []string {
	if from == to {
		return []string{from}
	}

	// Define migration paths (linear for now, can be extended to support multiple paths)
	migrationPaths := map[string][]string{
		"1.0": {"1.0"},
	}

	if path, ok := migrationPaths[from]; ok {
		return path
	}

	return []string{from} // No migration path found, return current version
}

// FileChangeStorageStats contains statistics about a FileChangeMapping
type FileChangeStorageStats struct {
	TotalChanges       int
	TotalTasks         int
	PendingChanges     int
	CommittedChanges   int
	UnassignedChanges  int
	LastModified       time.Time
	StorageVersion     string
}

// GetStorageStats returns statistics about the FileChangeMapping
func (fcm *FileChangeMapping) GetStorageStats() FileChangeStorageStats {
	stats := FileChangeStorageStats{
		TotalChanges:      fcm.TotalChangesCount(),
		TotalTasks:        len(fcm.Tasks),
		LastModified:      fcm.LastUpdated,
		StorageVersion:    fcm.Version,
		UnassignedChanges: len(fcm.UnassignedChanges),
	}

	// Count pending and committed changes
	for _, changes := range fcm.Tasks {
		for _, change := range changes {
			if change.IsPending {
				stats.PendingChanges++
			} else {
				stats.CommittedChanges++
			}
		}
	}

	for _, change := range fcm.UnassignedChanges {
		if change.IsPending {
			stats.PendingChanges++
		} else {
			stats.CommittedChanges++
		}
	}

	return stats
}
