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

// GetTaskFromCLI retrieves task details directly from the task-master CLI
// This ensures we get the most up-to-date task information from the database
// It reloads tasks from disk (respecting the active tag) and returns a deep copy
func (s *Service) GetTaskFromCLI(taskID string) (*Task, error) {
	// Force reload tasks from disk to ensure we have the latest data
	// This respects the active tag from config
	ctx := context.WithValue(context.Background(), "force", true)
	if err := s.LoadTasks(ctx); err != nil {
		return nil, fmt.Errorf("failed to reload tasks: %w", err)
	}
	
	// Get the task from the freshly loaded index
	// This returns a pointer to the actual task in the tree structure
	s.mu.RLock()
	task, ok := s.TaskIndex[taskID]
	s.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	// Create a deep copy of the task to avoid mutations
	// and ensure we have a clean snapshot of dependencies
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
	
	return taskCopy, nil
}
