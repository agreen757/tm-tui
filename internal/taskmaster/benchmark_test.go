package taskmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
)

// generateLargeTasks creates a large number of tasks for benchmarking
func generateLargeTasks(count int) []Task {
	tasks := make([]Task, 0, count)
	
	for i := 0; i < count; i++ {
		taskID := fmt.Sprintf("%d", i+1)
		task := Task{
			ID:          taskID,
			Title:       fmt.Sprintf("Task %d", i+1),
			Description: fmt.Sprintf("Description for task %d with some content", i+1),
			Status:      StatusPending,
			Priority:    PriorityMedium,
			Dependencies: []string{},
			Details:     fmt.Sprintf("Implementation details for task %d", i+1),
			TestStrategy: "Unit tests and integration tests",
			Subtasks:    []Task{},
		}
		
		// Add some dependencies
		if i > 0 && i%10 == 0 {
			task.Dependencies = []string{fmt.Sprintf("%d", i)}
		}
		
		// Add subtasks to every 10th task
		if i%10 == 0 {
			subtasks := make([]Task, 5)
			for j := 0; j < 5; j++ {
				subtasks[j] = Task{
					ID:          fmt.Sprintf("%s.%d", taskID, j+1),
					Title:       fmt.Sprintf("Subtask %d.%d", i+1, j+1),
					Description: "Subtask description",
					Status:      StatusPending,
					Priority:    PriorityLow,
					Dependencies: []string{},
				}
			}
			task.Subtasks = subtasks
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks
}

func setupBenchmarkService(b *testing.B, taskCount int) (*Service, string) {
	tmpDir := b.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	
	tasks := map[string]interface{}{
		"tasks": generateLargeTasks(taskCount),
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(filepath.Join(tmDir, "tasks.json"), data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	
	return svc, tmpDir
}

func BenchmarkLoadTasks_1K(b *testing.B) {
	tmpDir := b.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	tasksFile := filepath.Join(tmDir, "tasks.json")
	
	tasks := map[string]interface{}{
		"tasks": generateLargeTasks(1000),
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(tasksFile, data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	
	ctx := context.WithValue(context.Background(), "force", true)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.LoadTasks(ctx)
	}
}

func BenchmarkLoadTasks_10K(b *testing.B) {
	tmpDir := b.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster", "tasks")
	os.MkdirAll(tmDir, 0755)
	tasksFile := filepath.Join(tmDir, "tasks.json")
	
	tasks := map[string]interface{}{
		"tasks": generateLargeTasks(10000),
	}
	data, _ := json.Marshal(tasks)
	os.WriteFile(tasksFile, data, 0644)
	
	cfg := &config.Config{TaskMasterPath: tmpDir}
	svc, err := NewService(cfg)
	if err != nil {
		b.Fatal(err)
	}
	
	ctx := context.WithValue(context.Background(), "force", true)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.LoadTasks(ctx)
	}
}

func BenchmarkGetTaskByID_1K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Look up tasks at various positions
		svc.GetTaskByID("1")
		svc.GetTaskByID("500")
		svc.GetTaskByID("1000")
	}
}

func BenchmarkGetTaskByID_10K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Look up tasks at various positions
		svc.GetTaskByID("1")
		svc.GetTaskByID("5000")
		svc.GetTaskByID("10000")
	}
}

func BenchmarkGetTaskByID_WithSubtasks(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Look up subtasks
		svc.GetTaskByID("10.1")
		svc.GetTaskByID("50.3")
		svc.GetTaskByID("100.5")
	}
}

func BenchmarkGetTasks_1K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetTasks()
	}
}

func BenchmarkGetTasks_10K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetTasks()
	}
}

func BenchmarkGetTaskCount_1K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetTaskCount()
	}
}

func BenchmarkGetTaskCount_10K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetTaskCount()
	}
}

func BenchmarkGetNextTask_1K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetNextTask()
	}
}

func BenchmarkGetNextTask_10K(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetNextTask()
	}
}

func BenchmarkBuildTaskIndex_1K(b *testing.B) {
	tasks := generateLargeTasks(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildTaskIndex(tasks)
	}
}

func BenchmarkBuildTaskIndex_10K(b *testing.B) {
	tasks := generateLargeTasks(10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildTaskIndex(tasks)
	}
}

func BenchmarkValidateTasks_1K(b *testing.B) {
	tasks := generateLargeTasks(1000)
	index, _ := buildTaskIndex(tasks)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateTasks(tasks, index)
	}
}

func BenchmarkValidateTasks_10K(b *testing.B) {
	tasks := generateLargeTasks(10000)
	index, _ := buildTaskIndex(tasks)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateTasks(tasks, index)
	}
}

func BenchmarkFlattenTasks_1K(b *testing.B) {
	tasks := generateLargeTasks(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenTasks(tasks)
	}
}

func BenchmarkFlattenTasks_10K(b *testing.B) {
	tasks := generateLargeTasks(10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenTasks(tasks)
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	svc, _ := setupBenchmarkService(b, 1000)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			svc.GetTasks()
			svc.GetTaskByID("500")
			svc.GetNextTask()
			svc.GetTaskCount()
		}
	})
}
