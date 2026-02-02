package ui

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/filechanges"
)

// TestActiveTaskTracking tests that active task is properly tracked during status changes
func TestActiveTaskTracking(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		newStatus      string
		shouldSetActive bool
		shouldClearActive bool
	}{
		{
			name:            "Set active when changing to in-progress",
			initialStatus:   "pending",
			newStatus:       "in-progress",
			shouldSetActive: true,
		},
		{
			name:             "Clear active when changing from in-progress to done",
			initialStatus:    "in-progress",
			newStatus:        "done",
			shouldClearActive: true,
		},
		{
			name:             "Clear active when changing from in-progress to pending",
			initialStatus:    "in-progress",
			newStatus:        "pending",
			shouldClearActive: true,
		},
		{
			name:            "No change when pending stays pending",
			initialStatus:   "pending",
			newStatus:       "pending",
			shouldSetActive: false,
		},
		{
			name:            "No change when changing to pending from done",
			initialStatus:   "done",
			newStatus:       "pending",
			shouldSetActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock file change tracker
			tracker := &filechanges.FileChangeTracker{}
			
			// Simulate task status change handling
			taskID := "1.1"
			
			if tt.newStatus == "in-progress" && tt.shouldSetActive {
				tracker.SetActiveTask(taskID)
			} else if tt.shouldClearActive && tracker.GetActiveTask() == taskID {
				tracker.SetActiveTask("")
			}

			// Verify active task state
			active := tracker.GetActiveTask()
			expectedActive := ""
			if tt.shouldSetActive {
				expectedActive = taskID
			}

			if active != expectedActive {
				t.Errorf("Expected active task %q, got %q", expectedActive, active)
			}
		})
	}
}

// TestConcurrentStatusChanges tests that active task tracking is thread-safe
func TestConcurrentStatusChanges(t *testing.T) {
	tracker := &filechanges.FileChangeTracker{}
	
	// Track which tasks were set as active
	var mu sync.Mutex
	activeTaskHistory := make([]string, 0)
	
	// Simulate concurrent status changes
	var wg sync.WaitGroup
	numGoroutines := 10
	numTasksPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numTasksPerGoroutine; j++ {
				taskID := ""
				if j%2 == 0 {
					// Simulate setting in-progress
					taskID = fmt.Sprintf("task-%d-%d", goroutineID, j)
					tracker.SetActiveTask(taskID)
				} else {
					// Simulate clearing
					if tracker.GetActiveTask() != "" {
						tracker.SetActiveTask("")
					}
				}

				if taskID != "" {
					mu.Lock()
					activeTaskHistory = append(activeTaskHistory, taskID)
					mu.Unlock()
				}

				// Small delay to simulate work
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify that we had concurrent access without panics
	// and the tracker is in a valid state
	activeTask := tracker.GetActiveTask()
	if activeTask == "" || activeTask != tracker.GetActiveTask() {
		t.Errorf("Tracker in invalid state after concurrent access")
	}
}

// TestActiveTaskClearanceOnStatusChange tests that active task is cleared when status changes
func TestActiveTaskClearanceOnStatusChange(t *testing.T) {
	tracker := &filechanges.FileChangeTracker{}
	taskID := "2.1"

	// Set task as active
	tracker.SetActiveTask(taskID)
	if tracker.GetActiveTask() != taskID {
		t.Errorf("Failed to set active task")
	}

	// Simulate status change away from in-progress
	if tracker.GetActiveTask() == taskID {
		tracker.SetActiveTask("")
	}

	if tracker.GetActiveTask() != "" {
		t.Errorf("Active task should be cleared, got %q", tracker.GetActiveTask())
	}
}

// TestMultipleTaskStatusChanges tests handling of multiple tasks changing status
func TestMultipleTaskStatusChanges(t *testing.T) {
	tracker := &filechanges.FileChangeTracker{}
	
	// Task 1 becomes in-progress
	tracker.SetActiveTask("1.1")
	if tracker.GetActiveTask() != "1.1" {
		t.Errorf("Task 1.1 should be active")
	}

	// Task 2 becomes in-progress (should replace task 1)
	tracker.SetActiveTask("1.2")
	if tracker.GetActiveTask() != "1.2" {
		t.Errorf("Task 1.2 should now be active")
	}

	// Task 2 is done, clear it
	if tracker.GetActiveTask() == "1.2" {
		tracker.SetActiveTask("")
	}

	if tracker.GetActiveTask() != "" {
		t.Errorf("Active task should be empty")
	}
}

// TestActiveTaskMaintenance tests that active task state is properly maintained
func TestActiveTaskMaintenance(t *testing.T) {
	tracker := &filechanges.FileChangeTracker{}

	// Test setting and getting active task multiple times
	for i := 1; i <= 5; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		tracker.SetActiveTask(taskID)
		
		retrieved := tracker.GetActiveTask()
		if retrieved != taskID {
			t.Errorf("Iteration %d: expected %q, got %q", i, taskID, retrieved)
		}
	}

	// Clear and verify
	tracker.SetActiveTask("")
	if tracker.GetActiveTask() != "" {
		t.Errorf("Active task should be empty after clearing")
	}
}

// TestTaskDeletionWhileActive tests behavior when active task is deleted
func TestTaskDeletionWhileActive(t *testing.T) {
	tracker := &filechanges.FileChangeTracker{}
	taskID := "3.1"

	// Set task as active
	tracker.SetActiveTask(taskID)

	// Simulate task deletion - should clear active if it's the deleted task
	// In a real scenario, this would be handled by the UI when a task is deleted
	currentActive := tracker.GetActiveTask()
	if currentActive == taskID {
		tracker.SetActiveTask("")
	}

	if tracker.GetActiveTask() != "" {
		t.Errorf("Active task should be cleared after deletion")
	}
}
