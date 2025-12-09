package taskmaster

import (
	"encoding/json"
	"fmt"
	"time"
)

// Task represents a task from the Task Master system
type Task struct {
	ID              string            `json:"id"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	Status          string            `json:"status"`
	Priority        string            `json:"priority"`
	Dependencies    []string          `json:"dependencies"`
	Details         string            `json:"details"`
	TestStrategy    string            `json:"testStrategy"`
	Subtasks        []Task            `json:"subtasks"`
	Complexity      int               `json:"complexity"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	ParentID        string            `json:"parentId,omitempty"`
	EstimatedHours  float64           `json:"estimatedHours,omitempty"`
	ActualHours     float64           `json:"actualHours,omitempty"`
	Notes           []string          `json:"notes,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	
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
