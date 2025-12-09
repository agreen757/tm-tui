package taskmaster

import (
	"testing"
)

func TestValidateTasks(t *testing.T) {
	tests := []struct {
		name         string
		tasks        []Task
		wantWarnings int
	}{
		{
			name: "valid tasks",
			tasks: []Task{
				{
					ID:           "1",
					Status:       StatusPending,
					Priority:     PriorityHigh,
					Dependencies: []string{},
				},
			},
			wantWarnings: 0,
		},
		{
			name: "invalid status",
			tasks: []Task{
				{
					ID:       "1",
					Status:   "invalid-status",
					Priority: PriorityHigh,
				},
			},
			wantWarnings: 1,
		},
		{
			name: "invalid priority",
			tasks: []Task{
				{
					ID:       "1",
					Status:   StatusPending,
					Priority: "invalid-priority",
				},
			},
			wantWarnings: 1,
		},
		{
			name: "missing dependency",
			tasks: []Task{
				{
					ID:           "1",
					Status:       StatusPending,
					Priority:     PriorityHigh,
					Dependencies: []string{"999"},
				},
			},
			wantWarnings: 1,
		},
		{
			name: "circular dependency simple",
			tasks: []Task{
				{
					ID:           "1",
					Status:       StatusPending,
					Dependencies: []string{"2"},
				},
				{
					ID:           "2",
					Status:       StatusPending,
					Dependencies: []string{"1"},
				},
			},
			wantWarnings: 1,
		},
		{
			name: "circular dependency complex",
			tasks: []Task{
				{
					ID:           "1",
					Status:       StatusPending,
					Dependencies: []string{"2"},
				},
				{
					ID:           "2",
					Status:       StatusPending,
					Dependencies: []string{"3"},
				},
				{
					ID:           "3",
					Status:       StatusPending,
					Dependencies: []string{"1"},
				},
			},
			wantWarnings: 1,
		},
		{
			name: "multiple issues",
			tasks: []Task{
				{
					ID:           "1",
					Status:       "bad-status",
					Priority:     "bad-priority",
					Dependencies: []string{"999"},
				},
			},
			wantWarnings: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, _ := buildTaskIndex(tt.tasks)
			warnings := validateTasks(tt.tasks, index)
			
			if len(warnings) != tt.wantWarnings {
				t.Errorf("validateTasks() warnings = %d, want %d", len(warnings), tt.wantWarnings)
				for _, w := range warnings {
					t.Logf("  Warning: %s", w.String())
				}
			}
		})
	}
}

func TestValidateTask(t *testing.T) {
	tests := []struct {
		name         string
		task         Task
		index        map[string]*Task
		wantWarnings int
	}{
		{
			name: "valid task",
			task: Task{
				ID:           "1",
				Status:       StatusPending,
				Priority:     PriorityHigh,
				Dependencies: []string{},
			},
			index:        map[string]*Task{},
			wantWarnings: 0,
		},
		{
			name: "invalid status",
			task: Task{
				ID:       "1",
				Status:   "bad",
				Priority: PriorityHigh,
			},
			index:        map[string]*Task{},
			wantWarnings: 1,
		},
		{
			name: "invalid priority",
			task: Task{
				ID:       "1",
				Status:   StatusPending,
				Priority: "bad",
			},
			index:        map[string]*Task{},
			wantWarnings: 1,
		},
		{
			name: "missing dependency",
			task: Task{
				ID:           "1",
				Status:       StatusPending,
				Priority:     PriorityHigh,
				Dependencies: []string{"2"},
			},
			index:        map[string]*Task{},
			wantWarnings: 1,
		},
		{
			name: "valid dependency",
			task: Task{
				ID:           "1",
				Status:       StatusPending,
				Priority:     PriorityHigh,
				Dependencies: []string{"2"},
			},
			index: map[string]*Task{
				"2": {ID: "2", Status: StatusDone},
			},
			wantWarnings: 0,
		},
		{
			name: "recursive validation of subtasks",
			task: Task{
				ID:       "1",
				Status:   StatusPending,
				Priority: PriorityHigh,
				Subtasks: []Task{
					{
						ID:       "1.1",
						Status:   "bad-status",
						Priority: PriorityMedium,
					},
				},
			},
			index:        map[string]*Task{},
			wantWarnings: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := validateTask(&tt.task, tt.index)
			
			if len(warnings) != tt.wantWarnings {
				t.Errorf("validateTask() warnings = %d, want %d", len(warnings), tt.wantWarnings)
			}
		})
	}
}

func TestDetectCircularDependencies(t *testing.T) {
	tests := []struct {
		name         string
		tasks        []Task
		wantWarnings int
	}{
		{
			name: "no circular dependencies",
			tasks: []Task{
				{ID: "1", Dependencies: []string{}},
				{ID: "2", Dependencies: []string{"1"}},
				{ID: "3", Dependencies: []string{"2"}},
			},
			wantWarnings: 0,
		},
		{
			name: "simple cycle A->B->A",
			tasks: []Task{
				{ID: "A", Dependencies: []string{"B"}},
				{ID: "B", Dependencies: []string{"A"}},
			},
			wantWarnings: 1,
		},
		{
			name: "three node cycle A->B->C->A",
			tasks: []Task{
				{ID: "A", Dependencies: []string{"B"}},
				{ID: "B", Dependencies: []string{"C"}},
				{ID: "C", Dependencies: []string{"A"}},
			},
			wantWarnings: 1,
		},
		{
			name: "self-reference",
			tasks: []Task{
				{ID: "1", Dependencies: []string{"1"}},
			},
			wantWarnings: 1,
		},
		{
			name: "complex graph with cycle",
			tasks: []Task{
				{ID: "1", Dependencies: []string{}},
				{ID: "2", Dependencies: []string{"1"}},
				{ID: "3", Dependencies: []string{"2", "4"}},
				{ID: "4", Dependencies: []string{"3"}}, // Creates cycle 3->4->3
				{ID: "5", Dependencies: []string{"1"}},
			},
			wantWarnings: 1,
		},
		{
			name: "diamond dependency no cycle",
			tasks: []Task{
				{ID: "1", Dependencies: []string{}},
				{ID: "2", Dependencies: []string{"1"}},
				{ID: "3", Dependencies: []string{"1"}},
				{ID: "4", Dependencies: []string{"2", "3"}},
			},
			wantWarnings: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, _ := buildTaskIndex(tt.tasks)
			warnings := detectCircularDependencies(index)
			
			if len(warnings) != tt.wantWarnings {
				t.Errorf("detectCircularDependencies() warnings = %d, want %d", len(warnings), tt.wantWarnings)
				for _, w := range warnings {
					t.Logf("  Warning: %s", w.String())
				}
			}
		})
	}
}

func TestTask_IsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{StatusPending, true},
		{StatusInProgress, true},
		{StatusDone, true},
		{StatusDeferred, true},
		{StatusCancelled, true},
		{StatusBlocked, true},
		{"invalid", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			task := Task{Status: tt.status}
			if got := task.IsValidStatus(); got != tt.want {
				t.Errorf("Task.IsValidStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		want     bool
	}{
		{PriorityHigh, true},
		{PriorityMedium, true},
		{PriorityLow, true},
		{PriorityCritical, true},
		{"", true}, // Empty is valid
		{"invalid", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			task := Task{Priority: tt.priority}
			if got := task.IsValidPriority(); got != tt.want {
				t.Errorf("Task.IsValidPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_HasBlockedDependencies(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		index   map[string]*Task
		want    bool
	}{
		{
			name: "no dependencies",
			task: Task{ID: "1", Dependencies: []string{}},
			index: map[string]*Task{},
			want: false,
		},
		{
			name: "dependency completed",
			task: Task{ID: "2", Dependencies: []string{"1"}},
			index: map[string]*Task{
				"1": {ID: "1", Status: StatusDone},
			},
			want: false,
		},
		{
			name: "dependency pending",
			task: Task{ID: "2", Dependencies: []string{"1"}},
			index: map[string]*Task{
				"1": {ID: "1", Status: StatusPending},
			},
			want: true,
		},
		{
			name: "dependency in progress",
			task: Task{ID: "2", Dependencies: []string{"1"}},
			index: map[string]*Task{
				"1": {ID: "1", Status: StatusInProgress},
			},
			want: true,
		},
		{
			name: "multiple dependencies all done",
			task: Task{ID: "3", Dependencies: []string{"1", "2"}},
			index: map[string]*Task{
				"1": {ID: "1", Status: StatusDone},
				"2": {ID: "2", Status: StatusDone},
			},
			want: false,
		},
		{
			name: "multiple dependencies one pending",
			task: Task{ID: "3", Dependencies: []string{"1", "2"}},
			index: map[string]*Task{
				"1": {ID: "1", Status: StatusDone},
				"2": {ID: "2", Status: StatusPending},
			},
			want: true,
		},
		{
			name: "missing dependency",
			task: Task{ID: "2", Dependencies: []string{"1"}},
			index: map[string]*Task{},
			want: false, // Missing deps don't block
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.HasBlockedDependencies(tt.index); got != tt.want {
				t.Errorf("Task.HasBlockedDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}
