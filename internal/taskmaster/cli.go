package taskmaster

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ExecuteCommand executes a task-master CLI command
func (s *Service) ExecuteCommand(args ...string) (string, error) {
	cmd := exec.Command("task-master", args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("command failed: %w", err)
	}

	return stdout.String(), nil
}

// SetTaskStatus sets the status of a task
func (s *Service) SetTaskStatus(taskID, status string) error {
	_, err := s.ExecuteCommand("set-status", fmt.Sprintf("--id=%s", taskID), fmt.Sprintf("--status=%s", status))
	if err != nil {
		return err
	}

	// Reload tasks after status change
	ctx := context.WithValue(context.Background(), "force", true)
	return s.LoadTasks(ctx)
}

// ExpandTask expands a task into subtasks
func (s *Service) ExpandTask(taskID string, research bool) error {
	args := []string{"expand", fmt.Sprintf("--id=%s", taskID)}
	if research {
		args = append(args, "--research")
	}

	_, err := s.ExecuteCommand(args...)
	if err != nil {
		return err
	}

	// Reload tasks after expansion
	ctx := context.WithValue(context.Background(), "force", true)
	return s.LoadTasks(ctx)
}

// AddTask adds a new task
func (s *Service) AddTask(prompt string, research bool) error {
	args := []string{"add-task", fmt.Sprintf("--prompt=%s", prompt)}
	if research {
		args = append(args, "--research")
	}

	_, err := s.ExecuteCommand(args...)
	if err != nil {
		return err
	}

	// Reload tasks after adding
	ctx := context.WithValue(context.Background(), "force", true)
	return s.LoadTasks(ctx)
}

// UpdateTask updates an existing task
func (s *Service) UpdateTask(taskID, prompt string) error {
	_, err := s.ExecuteCommand("update-task", fmt.Sprintf("--id=%s", taskID), fmt.Sprintf("--prompt=%s", prompt))
	if err != nil {
		return err
	}

	// Reload tasks after update
	ctx := context.WithValue(context.Background(), "force", true)
	return s.LoadTasks(ctx)
}

// GetTaskDetails returns formatted task details
func (s *Service) GetTaskDetails(taskID string) (string, error) {
	task, err := s.GetTask(taskID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Task %s: %s\n", task.ID, task.Title))
	b.WriteString(fmt.Sprintf("Status: %s\n", task.Status))
	b.WriteString(fmt.Sprintf("Priority: %s\n", task.Priority))
	
	if len(task.Dependencies) > 0 {
		b.WriteString(fmt.Sprintf("Dependencies: %s\n", strings.Join(task.Dependencies, ", ")))
	}
	
	if task.Description != "" {
		b.WriteString(fmt.Sprintf("\nDescription:\n%s\n", task.Description))
	}
	
	if task.Details != "" {
		b.WriteString(fmt.Sprintf("\nDetails:\n%s\n", task.Details))
	}
	
	if task.TestStrategy != "" {
		b.WriteString(fmt.Sprintf("\nTest Strategy:\n%s\n", task.TestStrategy))
	}

	return b.String(), nil
}

// GetTaskFromCLI retrieves task details directly by reading tasks.json
// This ensures we get the correct task for the active tag context
func (s *Service) GetTaskFromCLI(taskID string) (*Task, error) {
	// Get the active tag from config
	tag := s.config.ActiveTag
	if tag == "" {
		tag = "master"
	}
	
	// Read tasks from file for the specific tag
	tasks, err := LoadTasksFromFile(s.RootDir, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks from file: %w", err)
	}
	
	// Search for the task recursively
	var findTask func(tasks []Task) *Task
	findTask = func(tasks []Task) *Task {
		for i := range tasks {
			if tasks[i].ID == taskID {
				// Found it! Create a deep copy to return
				return copyTask(&tasks[i])
			}
			// Check subtasks
			if result := findTask(tasks[i].Subtasks); result != nil {
				return result
			}
		}
		return nil
	}
	
	task := findTask(tasks)
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	return task, nil
}

// copyTask creates a deep copy of a task (without subtasks to avoid bloat)
func copyTask(task *Task) *Task {
	taskCopy := &Task{
		ID:             task.ID,
		Title:          task.Title,
		Description:    task.Description,
		Status:         task.Status,
		Priority:       task.Priority,
		Dependencies:   make([]string, len(task.Dependencies)),
		Details:        task.Details,
		TestStrategy:   task.TestStrategy,
		Complexity:     task.Complexity,
		EstimatedHours: task.EstimatedHours,
		ActualHours:    task.ActualHours,
		ParentID:       task.ParentID,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
	
	// Deep copy dependencies
	copy(taskCopy.Dependencies, task.Dependencies)
	
	// Deep copy metadata if present
	if task.Metadata != nil {
		taskCopy.Metadata = make(map[string]string)
		for k, v := range task.Metadata {
			taskCopy.Metadata[k] = v
		}
	}
	
	// Deep copy notes if present
	if task.Notes != nil {
		taskCopy.Notes = make([]string, len(task.Notes))
		copy(taskCopy.Notes, task.Notes)
	}
	
	// Deep copy tags if present
	if task.Tags != nil {
		taskCopy.Tags = make([]string, len(task.Tags))
		copy(taskCopy.Tags, task.Tags)
	}
	
	return taskCopy
}
