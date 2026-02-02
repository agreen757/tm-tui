package taskmaster

import (
	"encoding/json"
	"fmt"
	"time"
)

// FileChange represents a changed file associated with a task
type FileChange struct {
	Path        string    `json:"path"`        // File path relative to repository root
	ChangeType  string    `json:"changeType"`  // "added", "modified", "deleted"
	Description string    `json:"description"` // Optional description of the change
	LastChanged time.Time `json:"lastChanged"` // When the file was last changed
	CommitID    string    `json:"commitId"`    // Git commit ID (if committed)
	IsPending   bool      `json:"isPending"`   // True for uncommitted changes
}

// FileChangeMapping is the top-level structure for storage of file changes
type FileChangeMapping struct {
	Version           string                `json:"version"`           // Schema version
	LastUpdated       time.Time             `json:"lastUpdated"`       // When mapping was last updated
	Tasks             map[string][]FileChange `json:"tasks"`           // Task ID -> array of file changes
	UnassignedChanges []FileChange          `json:"unassignedChanges"` // Changes not assigned to any task
}

// Task represents a task from the Task Master system
type Task struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Status         string            `json:"status"`
	Priority       string            `json:"priority"`
	Dependencies   []string          `json:"dependencies"`
	Details        string            `json:"details"`
	TestStrategy   string            `json:"testStrategy"`
	Subtasks       []Task            `json:"subtasks"`
	Complexity     int               `json:"complexity"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	ParentID       string            `json:"parentId,omitempty"`
	EstimatedHours float64           `json:"estimatedHours,omitempty"`
	ActualHours    float64           `json:"actualHours,omitempty"`
	Notes          []string          `json:"notes,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	IsCategory     bool              `json:"isCategory,omitempty"`
	IsRoot         bool              `json:"isRoot,omitempty"`
	FileChanges    []FileChange      `json:"fileChanges,omitempty"` // Associated file changes

	// Navigation helpers (not serialized)
	Parent   *Task   `json:"-"`
	Children []*Task `json:"-"`
}

// TaskStatus constants
const (
	StatusPending    = "pending"
	StatusInProgress = "in-progress"
	StatusDone       = "done"
	StatusDeferred   = "deferred"
	StatusCancelled  = "cancelled"
	StatusBlocked    = "blocked"
)

// Priority constants
const (
	PriorityHigh     = "high"
	PriorityMedium   = "medium"
	PriorityLow      = "low"
	PriorityCritical = "critical"
)

// TasksFile represents the structure of tasks.json
type TasksFile struct {
	Tasks   []Task                 `json:"tasks"`
	Version string                 `json:"version,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// TaggedTasksFile represents tasks.json with tag-based structure (e.g., "master" tag)
type TaggedTasksFile struct {
	Tags map[string]struct {
		Tasks []Task `json:"tasks"`
	} `json:"-"` // We'll unmarshal this manually
}

// ValidationWarning represents a non-fatal issue found during validation
type ValidationWarning struct {
	TaskID  string
	Message string
}

func (w ValidationWarning) String() string {
	return fmt.Sprintf("Task %s: %s", w.TaskID, w.Message)
}

// IsComplete returns true if the task is marked as done
func (t *Task) IsComplete() bool {
	return t.Status == StatusDone
}

// IsValidStatus checks if the status is one of the defined constants
func (t *Task) IsValidStatus() bool {
	switch t.Status {
	case StatusPending, StatusInProgress, StatusDone, StatusDeferred, StatusCancelled, StatusBlocked:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if the priority is one of the defined constants
func (t *Task) IsValidPriority() bool {
	switch t.Priority {
	case PriorityHigh, PriorityMedium, PriorityLow, PriorityCritical, "":
		return true
	default:
		return false
	}
}

// HasBlockedDependencies checks if any dependencies are blocking this task
func (t *Task) HasBlockedDependencies(taskIndex map[string]*Task) bool {
	for _, depID := range t.Dependencies {
		if dep, ok := taskIndex[depID]; ok {
			if !dep.IsComplete() {
				return true
			}
		}
	}
	return false
}

// UnmarshalJSON implements custom JSON unmarshaling to handle int or string IDs
func (t *Task) UnmarshalJSON(data []byte) error {
	// Define an intermediate type to avoid recursion
	type Alias Task
	aux := &struct {
		ID           interface{}   `json:"id"`
		Dependencies []interface{} `json:"dependencies"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert ID to string
	switch v := aux.ID.(type) {
	case string:
		t.ID = v
	case float64:
		t.ID = fmt.Sprintf("%.0f", v)
	case int:
		t.ID = fmt.Sprintf("%d", v)
	default:
		if v != nil {
			t.ID = fmt.Sprintf("%v", v)
		}
	}

	// Convert dependencies to strings
	if aux.Dependencies != nil {
		t.Dependencies = make([]string, len(aux.Dependencies))
		for i, dep := range aux.Dependencies {
			switch v := dep.(type) {
			case string:
				t.Dependencies[i] = v
			case float64:
				t.Dependencies[i] = fmt.Sprintf("%.0f", v)
			case int:
				t.Dependencies[i] = fmt.Sprintf("%d", v)
			default:
				if v != nil {
					t.Dependencies[i] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	return nil
}

// ValidChangeTypes represents the valid values for FileChange.ChangeType
var ValidChangeTypes = map[string]bool{
	"added":    true,
	"modified": true,
	"deleted":  true,
}

// Validate checks if FileChange has valid data
func (fc *FileChange) Validate() error {
	if fc.Path == "" {
		return fmt.Errorf("file change path cannot be empty")
	}
	if fc.ChangeType == "" {
		return fmt.Errorf("file change type cannot be empty")
	}
	if !ValidChangeTypes[fc.ChangeType] {
		return fmt.Errorf("invalid change type: %q (must be 'added', 'modified', or 'deleted')", fc.ChangeType)
	}
	return nil
}

// NewFileChange creates a new FileChange with the given parameters and sets LastChanged to now
func NewFileChange(path string, changeType string, description string) (*FileChange, error) {
	fc := &FileChange{
		Path:        path,
		ChangeType:  changeType,
		Description: description,
		LastChanged: time.Now(),
		IsPending:   true,
	}
	if err := fc.Validate(); err != nil {
		return nil, err
	}
	return fc, nil
}

// Validate checks if FileChangeMapping has valid data
func (fcm *FileChangeMapping) Validate() error {
	if fcm.Version == "" {
		return fmt.Errorf("file change mapping version cannot be empty")
	}
	if fcm.Tasks == nil {
		fcm.Tasks = make(map[string][]FileChange)
	}
	if fcm.UnassignedChanges == nil {
		fcm.UnassignedChanges = make([]FileChange, 0)
	}
	// Validate all file changes in all tasks
	for taskID, changes := range fcm.Tasks {
		for i, change := range changes {
			if err := change.Validate(); err != nil {
				return fmt.Errorf("invalid change in task %q at index %d: %w", taskID, i, err)
			}
		}
	}
	// Validate unassigned changes
	for i, change := range fcm.UnassignedChanges {
		if err := change.Validate(); err != nil {
			return fmt.Errorf("invalid unassigned change at index %d: %w", i, err)
		}
	}
	return nil
}

// NewFileChangeMapping creates a new FileChangeMapping with the given version
func NewFileChangeMapping(version string) *FileChangeMapping {
	return &FileChangeMapping{
		Version:           version,
		LastUpdated:       time.Now(),
		Tasks:             make(map[string][]FileChange),
		UnassignedChanges: make([]FileChange, 0),
	}
}

// AddFileChange adds a FileChange to a specific task
func (fcm *FileChangeMapping) AddFileChange(taskID string, fc FileChange) error {
	if err := fc.Validate(); err != nil {
		return fmt.Errorf("cannot add invalid file change: %w", err)
	}
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	fcm.Tasks[taskID] = append(fcm.Tasks[taskID], fc)
	fcm.LastUpdated = time.Now()
	return nil
}

// AddUnassignedChange adds a FileChange to the unassigned list
func (fcm *FileChangeMapping) AddUnassignedChange(fc FileChange) error {
	if err := fc.Validate(); err != nil {
		return fmt.Errorf("cannot add invalid file change: %w", err)
	}
	fcm.UnassignedChanges = append(fcm.UnassignedChanges, fc)
	fcm.LastUpdated = time.Now()
	return nil
}

// GetChangesForTask returns all file changes associated with a task
func (fcm *FileChangeMapping) GetChangesForTask(taskID string) []FileChange {
	if changes, ok := fcm.Tasks[taskID]; ok {
		return changes
	}
	return []FileChange{}
}

// GetPendingChanges returns all pending (uncommitted) file changes
func (fcm *FileChangeMapping) GetPendingChanges() []FileChange {
	var pending []FileChange
	for _, changes := range fcm.Tasks {
		for _, change := range changes {
			if change.IsPending {
				pending = append(pending, change)
			}
		}
	}
	for _, change := range fcm.UnassignedChanges {
		if change.IsPending {
			pending = append(pending, change)
		}
	}
	return pending
}

// TotalChangesCount returns the total count of all file changes
func (fcm *FileChangeMapping) TotalChangesCount() int {
	count := len(fcm.UnassignedChanges)
	for _, changes := range fcm.Tasks {
		count += len(changes)
	}
	return count
}
