package taskmaster

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/adriangreen/tm-tui/internal/config"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) *config.Config
		wantErr     bool
		wantAvail   bool
	}{
		{
			name: "successful initialization with auto-detect",
			setup: func(tmpDir string) *config.Config {
				// Create .taskmaster directory
				tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
				os.MkdirAll(tmDir, 0755)
				
				// Create tasks.json
				tasks := map[string]interface{}{
					"tasks": []Task{
						{ID: "1", Title: "Test Task", Status: StatusPending, Priority: PriorityHigh},
					},
				}
				data, _ := json.Marshal(tasks)
				os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
				
				// Change to temp directory for auto-detection
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				t.Cleanup(func() { os.Chdir(oldWd) })
				
				return &config.Config{}
			},
			wantErr:   false,
			wantAvail: true,
		},
		{
			name: "successful initialization with configured path",
			setup: func(tmpDir string) *config.Config {
				// Create .taskmaster directory
				tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
				os.MkdirAll(tmDir, 0755)
				
				// Create tasks.json
				tasks := map[string]interface{}{
					"tasks": []Task{
						{ID: "1", Title: "Test Task", Status: StatusPending, Priority: PriorityHigh},
					},
				}
				data, _ := json.Marshal(tasks)
				os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
				
				return &config.Config{
					TaskMasterPath: tmpDir,
				}
			},
			wantErr:   false,
			wantAvail: true,
		},
		{
			name: "no taskmaster directory - graceful degradation",
			setup: func(tmpDir string) *config.Config {
				// Change to temp directory with no .taskmaster
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				t.Cleanup(func() { os.Chdir(oldWd) })
				
				return &config.Config{}
			},
			wantErr:   false,
			wantAvail: false,
		},
		{
			name: "invalid tasks.json",
			setup: func(tmpDir string) *config.Config {
				// Create .taskmaster directory
				tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
				os.MkdirAll(tmDir, 0755)
				
				// Create invalid JSON
				os.WriteFile(filepath.Join(tmDir, "tasks.json"), []byte("{invalid}"), 0644)
				
				return &config.Config{
					TaskMasterPath: tmpDir,
				}
			},
			wantErr:   true,
			wantAvail: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := tt.setup(tmpDir)
			
			svc, err := NewService(cfg)
			
			if tt.wantErr {
				if err == nil {
					t.Error("NewService() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewService() unexpected error = %v", err)
				return
			}
			
			if svc.IsAvailable() != tt.wantAvail {
				t.Errorf("Service.IsAvailable() = %v, want %v", svc.IsAvailable(), tt.wantAvail)
			}
		})
	}
}

func TestService_LoadTasks(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	tasksFile := filepath.Join(tmDir, "tasks.json")
	
	// Create initial tasks
	tasks := map[string]interface{}{
		"tasks": []Task{
			{ID: "1", Title: "Task 1", Status: StatusPending, Priority: PriorityHigh},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(tasksFile, data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify initial load
	loadedTasks, _ := svc.GetTasks()
	if len(loadedTasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedTasks))
	}
	
	// Update tasks file
	tasks = map[string]interface{}{
		"tasks": []Task{
			{ID: "1", Title: "Task 1", Status: StatusPending, Priority: PriorityHigh},
			{ID: "2", Title: "Task 2", Status: StatusDone, Priority: PriorityMedium},
		},
	}
	data, _ = json.Marshal(tasks)
	time.Sleep(10 * time.Millisecond) // Ensure different mod time
	os.WriteFile(tasksFile, data, 0644)
	
	// Reload tasks
	ctx := context.Background()
	if err := svc.LoadTasks(ctx); err != nil {
		t.Errorf("LoadTasks() error = %v", err)
	}
	
	// Verify updated tasks
	loadedTasks, _ = svc.GetTasks()
	if len(loadedTasks) != 2 {
		t.Errorf("Expected 2 tasks after reload, got %d", len(loadedTasks))
	}
}

func TestService_LoadTasks_Caching(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	tasksFile := filepath.Join(tmDir, "tasks.json")
	
	tasks := map[string]interface{}{
		"tasks": []Task{
			{ID: "1", Title: "Task 1", Status: StatusPending, Priority: PriorityHigh},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(tasksFile, data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	// Get initial mod time
	info, _ := os.Stat(tasksFile)
	initialModTime := info.ModTime()
	
	// Reload without file change - should use cache
	ctx := context.Background()
	if err := svc.LoadTasks(ctx); err != nil {
		t.Errorf("LoadTasks() error = %v", err)
	}
	
	// Verify mod time is tracked
	if !svc.lastModTime.Equal(initialModTime) {
		t.Error("Service should track file mod time")
	}
	
	// Force reload
	forceCtx := context.WithValue(ctx, "force", true)
	if err := svc.LoadTasks(forceCtx); err != nil {
		t.Errorf("LoadTasks() with force error = %v", err)
	}
}

func TestService_GetTaskByID(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	
	tasks := map[string]interface{}{
		"tasks": []Task{
			{
				ID:       "1",
				Title:    "Parent",
				Status:   StatusPending,
				Priority: PriorityHigh,
				Subtasks: []Task{
					{ID: "1.1", Title: "Child", Status: StatusDone, Priority: PriorityMedium},
				},
			},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		id      string
		wantOK  bool
		wantID  string
	}{
		{"1", true, "1"},
		{"1.1", true, "1.1"},
		{"999", false, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			task, ok := svc.GetTaskByID(tt.id)
			
			if ok != tt.wantOK {
				t.Errorf("GetTaskByID(%s) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			
			if ok && task.ID != tt.wantID {
				t.Errorf("GetTaskByID(%s) ID = %s, want %s", tt.id, task.ID, tt.wantID)
			}
		})
	}
}

func TestService_GetNextTask(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	
	tasks := map[string]interface{}{
		"tasks": []Task{
			{
				ID:           "1",
				Title:        "Done Task",
				Status:       StatusDone,
				Priority:     PriorityHigh,
				Dependencies: []string{},
			},
			{
				ID:           "2",
				Title:        "Blocked Task",
				Status:       StatusPending,
				Priority:     PriorityHigh,
				Dependencies: []string{"3"},
			},
			{
				ID:           "3",
				Title:        "Available Task",
				Status:       StatusPending,
				Priority:     PriorityHigh,
				Dependencies: []string{},
			},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	next, err := svc.GetNextTask()
	if err != nil {
		t.Fatalf("GetNextTask() error = %v", err)
	}
	
	// Should get task 3 (not blocked)
	if next.ID != "3" {
		t.Errorf("GetNextTask() ID = %s, want '3'", next.ID)
	}
}

func TestService_GetTaskCount(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	
	tasks := map[string]interface{}{
		"tasks": []Task{
			{ID: "1", Status: StatusPending},
			{ID: "2", Status: StatusDone},
			{ID: "3", Status: StatusInProgress},
			{
				ID:     "4",
				Status: StatusPending,
				Subtasks: []Task{
					{ID: "4.1", Status: StatusDone},
					{ID: "4.2", Status: StatusPending},
				},
			},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	counts := svc.GetTaskCount()
	
	want := map[string]int{
		StatusPending:    3, // Tasks 1, 4, 4.2
		StatusDone:       2, // Tasks 2, 4.1
		StatusInProgress: 1, // Task 3
		StatusBlocked:    0,
		StatusDeferred:   0,
		StatusCancelled:  0,
	}
	
	for status, wantCount := range want {
		if counts[status] != wantCount {
			t.Errorf("GetTaskCount()[%s] = %d, want %d", status, counts[status], wantCount)
		}
	}
}

func TestService_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	
	tasks := map[string]interface{}{
		"tasks": []Task{
			{ID: "1", Title: "Task 1", Status: StatusPending, Priority: PriorityHigh},
			{ID: "2", Title: "Task 2", Status: StatusDone, Priority: PriorityMedium},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	// Hammer the service with concurrent reads and writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Concurrent reads
			svc.GetTasks()
			svc.GetTaskByID("1")
			svc.GetNextTask()
			svc.GetTaskCount()
			svc.GetValidationWarnings()
			svc.IsAvailable()
			
			// Concurrent reloads
			ctx := context.Background()
			svc.LoadTasks(ctx)
		}()
	}
	
	wg.Wait()
	
	// Verify service is still functional
	if !svc.IsAvailable() {
		t.Error("Service should still be available after concurrent access")
	}
}

func TestService_ValidationWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	
	// Create tasks with validation issues
	tasks := map[string]interface{}{
		"tasks": []Task{
			{
				ID:           "1",
				Title:        "Task with bad status",
				Status:       "invalid-status",
				Priority:     PriorityHigh,
				Dependencies: []string{"999"}, // Missing dependency
			},
		},
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	warnings := svc.GetValidationWarnings()
	if len(warnings) < 2 {
		t.Errorf("Expected at least 2 warnings, got %d", len(warnings))
	}
	
	// Check warnings through GetTasks
	_, warningStrs := svc.GetTasks()
	if len(warningStrs) < 2 {
		t.Errorf("GetTasks() expected at least 2 warning strings, got %d", len(warningStrs))
	}
}
