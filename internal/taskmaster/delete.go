package taskmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agreen757/tm-tui/internal/config"
)

// DeleteOptions captures options surfaced to the delete workflow.
type DeleteOptions struct {
	Recursive bool `json:"recursive"`
	Force     bool `json:"force"`
}

// TaskSummary is a light-weight projection used for dialogs and warnings.
type TaskSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// DeleteImpact describes what would happen if the delete runs with the
// provided options.
type DeleteImpact struct {
	Selected         []TaskSummary `json:"selected"`
	Descendants      []TaskSummary `json:"descendants"`
	Dependents       []TaskSummary `json:"dependents"`
	TotalDeleteCount int           `json:"totalDeleteCount"`
	BlockingReason   string        `json:"blockingReason,omitempty"`
	WarningMessages  []string      `json:"warningMessages,omitempty"`
}

// DeleteResult captures the outcome of a delete operation.
type DeleteResult struct {
	DeletedCount int
	Warnings     []string
	Undo         *UndoToken
}

// AnalyzeDeleteImpact reports which tasks would be affected by a deletion.
func (s *Service) AnalyzeDeleteImpact(taskIDs []string, opts DeleteOptions) (*DeleteImpact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.available {
		return nil, fmt.Errorf("taskmaster not available")
	}
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("no tasks provided")
	}

	impact := &DeleteImpact{}
	selected := make(map[string]*Task)
	for _, id := range taskIDs {
		task, ok := s.TaskIndex[id]
		if !ok {
			return nil, fmt.Errorf("task %s not found", id)
		}
		selected[id] = task
		impact.Selected = append(impact.Selected, summarizeTask(task))
	}
	sortSummaries(impact.Selected)

	depMap := s.buildDependencyMapLocked()
	descendants := make(map[string]*Task)
	for _, task := range selected {
		collectDescendants(task, descendants)
	}
	impact.Descendants = summariesFromMap(descendants)

	dependents := make(map[string]*Task)
	queue := make([]*Task, 0)
	for _, task := range selected {
		for _, dep := range depMap[task.ID] {
			if _, skip := selected[dep.ID]; skip {
				continue
			}
			dependents[dep.ID] = dep
			queue = append(queue, dep)
		}
	}
	if opts.Recursive {
		seen := make(map[string]bool)
		for len(queue) > 0 {
			task := queue[0]
			queue = queue[1:]
			if seen[task.ID] {
				continue
			}
			seen[task.ID] = true
			for _, dep := range depMap[task.ID] {
				if _, skip := selected[dep.ID]; skip {
					continue
				}
				dependents[dep.ID] = dep
				queue = append(queue, dep)
			}
		}
	}
	removeFromMap(dependents, selected)
	removeFromMap(dependents, descendants)
	impact.Dependents = summariesFromMap(dependents)

	impact.TotalDeleteCount = len(selected)
	if opts.Recursive {
		impact.TotalDeleteCount += len(descendants) + len(dependents)
	}

	if !opts.Recursive {
		switch {
		case len(impact.Descendants) > 0:
			impact.BlockingReason = fmt.Sprintf("%d subtasks detected", len(impact.Descendants))
		case len(impact.Dependents) > 0:
			impact.BlockingReason = fmt.Sprintf("%d dependent tasks detected", len(impact.Dependents))
		}
	}

	impact.WarningMessages = buildDeleteWarnings(impact)
	return impact, nil
}

// DeleteTasks removes the requested tasks and persists the change.
func (s *Service) DeleteTasks(ctx context.Context, taskIDs []string, opts DeleteOptions) (*DeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.available {
		return nil, fmt.Errorf("taskmaster not available")
	}
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("no tasks provided")
	}

	depMap := s.buildDependencyMapLocked()
	deleteSet := make(map[string]struct{})
	warnings := []string{}
	queue := append([]string(nil), taskIDs...)
	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		id := queue[0]
		queue = queue[1:]
		if _, exists := deleteSet[id]; exists {
			continue
		}
		task, ok := s.TaskIndex[id]
		if !ok {
			if opts.Force {
				warnings = append(warnings, fmt.Sprintf("Task %s not found and was skipped", id))
				continue
			}
			return nil, fmt.Errorf("task %s not found", id)
		}

		if !opts.Recursive {
			if len(task.Subtasks) > 0 {
				return nil, fmt.Errorf("task %s has %d subtasks; enable recursive deletion", id, len(task.Subtasks))
			}
			if deps := depMap[id]; len(deps) > 0 {
				return nil, fmt.Errorf("task %s has %d dependent tasks; enable recursive deletion", id, len(deps))
			}
		}

		deleteSet[id] = struct{}{}
		if opts.Recursive {
			for i := range task.Subtasks {
				queue = append(queue, task.Subtasks[i].ID)
			}
			for _, dep := range depMap[id] {
				queue = append(queue, dep.ID)
			}
		}
	}

	if len(deleteSet) == 0 {
		return nil, fmt.Errorf("no tasks deleted")
	}

	tasksPath := s.tasksFilePath()
	_, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to backup tasks file: %w", err)
	}

	s.Tasks = filterTasksBySet(s.Tasks, deleteSet)
	s.rebuildIndexAndValidate()
	if err := s.persistTasksLocked(s.Tasks); err != nil {
		return nil, err
	}
	if info, err := os.Stat(tasksPath); err == nil {
		s.lastModTime = info.ModTime()
	}

	result := &DeleteResult{
		DeletedCount: len(deleteSet),
		Warnings:     warnings,
		Undo:         nil, // Undo is not currently supported via this interface
	}

	return result, nil
}

// UndoAction is a placeholder for undo support. Currently not implemented.
func (s *Service) UndoAction(ctx context.Context, actionID string) error {
	return fmt.Errorf("undo is not currently supported")
}

func (s *Service) buildDependencyMapLocked() map[string][]*Task {
	depMap := make(map[string][]*Task)
	for _, task := range s.TaskIndex {
		for _, depID := range task.Dependencies {
			depMap[depID] = append(depMap[depID], task)
		}
	}
	return depMap
}

func summarizeTask(task *Task) TaskSummary {
	if task == nil {
		return TaskSummary{}
	}
	return TaskSummary{ID: task.ID, Title: task.Title, Status: task.Status}
}

func summariesFromMap(tasks map[string]*Task) []TaskSummary {
	result := make([]TaskSummary, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, summarizeTask(task))
	}
	sortSummaries(result)
	return result
}

func sortSummaries(items []TaskSummary) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].ID == items[j].ID {
			return items[i].Title < items[j].Title
		}
		return items[i].ID < items[j].ID
	})
}

func removeFromMap(target map[string]*Task, refs map[string]*Task) {
	for id := range refs {
		delete(target, id)
	}
}

func collectDescendants(task *Task, acc map[string]*Task) {
	if task == nil {
		return
	}
	for i := range task.Subtasks {
		child := &task.Subtasks[i]
		if _, exists := acc[child.ID]; !exists {
			acc[child.ID] = child
		}
		collectDescendants(child, acc)
	}
}

func filterTasksBySet(tasks []Task, deleteSet map[string]struct{}) []Task {
	result := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if _, remove := deleteSet[task.ID]; remove {
			continue
		}
		task.Subtasks = filterTasksBySet(task.Subtasks, deleteSet)
		result = append(result, task)
	}
	return result
}

func (s *Service) persistTasksLocked(tasks []Task) error {
	data, err := os.ReadFile(s.tasksFilePath())
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil && len(raw) > 0 {
		tag := detectTaskContainerKey(raw, s.config)
		if tag != "" {
			entry := map[string]interface{}{}
			if existing, ok := raw[tag]; ok && len(existing) > 0 {
				_ = json.Unmarshal(existing, &entry)
			}
			entry["tasks"] = tasks
			updated, err := json.Marshal(entry)
			if err != nil {
				return fmt.Errorf("failed to marshal tag entry: %w", err)
			}
			raw[tag] = updated
			final, err := json.MarshalIndent(raw, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal tasks file: %w", err)
			}
			return os.WriteFile(s.tasksFilePath(), final, 0644)
		}

		if _, ok := raw["tasks"]; ok {
			payload, err := json.Marshal(tasks)
			if err != nil {
				return fmt.Errorf("failed to marshal tasks: %w", err)
			}
			raw["tasks"] = payload
			final, err := json.MarshalIndent(raw, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal tasks file: %w", err)
			}
			return os.WriteFile(s.tasksFilePath(), final, 0644)
		}
	}

	final, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}
	return os.WriteFile(s.tasksFilePath(), final, 0644)
}

func detectTaskContainerKey(raw map[string]json.RawMessage, cfg *config.Config) string {
	if cfg != nil && cfg.ActiveTag != "" {
		if _, ok := raw[cfg.ActiveTag]; ok {
			return cfg.ActiveTag
		}
	}
	if _, ok := raw["master"]; ok {
		if containsTasks(raw["master"]) {
			return "master"
		}
	}
	for key, value := range raw {
		if containsTasks(value) {
			return key
		}
	}
	return ""
}

func containsTasks(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	entry := map[string]interface{}{}
	if err := json.Unmarshal(raw, &entry); err != nil {
		return false
	}
	_, ok := entry["tasks"]
	return ok
}

func (s *Service) tasksFilePath() string {
	return filepath.Join(s.RootDir, ".taskmaster", "tasks", "tasks.json")
}

func buildDeleteWarnings(impact *DeleteImpact) []string {
	warnings := []string{}
	if len(impact.Descendants) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d subtasks will be removed: %s", len(impact.Descendants), summarizeList(impact.Descendants)))
	}
	if len(impact.Dependents) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d dependent tasks will be removed: %s", len(impact.Dependents), summarizeList(impact.Dependents)))
	}
	return warnings
}

func summarizeList(items []TaskSummary) string {
	const maxItems = 4
	snippets := make([]string, 0, maxItems)
	for i, item := range items {
		if i >= maxItems {
			break
		}
		snippets = append(snippets, fmt.Sprintf("%s (%s)", item.ID, item.Title))
	}
	if len(items) > maxItems {
		snippets = append(snippets, fmt.Sprintf("…+%d", len(items)-maxItems))
	}
	return strings.Join(snippets, ", ")
}
