package filechanges

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/agreen757/tm-tui/internal/git"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// GitServiceInterface defines the interface for git operations needed by the tracker
type GitServiceInterface interface {
	GetUncommittedChanges(ctx context.Context) ([]git.FileChange, error)
	GetCommitsWithTaskIDsFiltered(ctx context.Context, opts git.CommitFilterOptions) (*git.TaskCommitMapping, error)
}

// FileChangeTracker manages file change tracking across tasks
type FileChangeTracker struct {
	gitService              GitServiceInterface
	storage                 taskmaster.Storage
	mapping                 *taskmaster.FileChangeMapping
	activeTask              string
	refreshTicker           *time.Ticker
	stopChan                chan struct{}
	mutex                   sync.RWMutex
	repoPath                string
	commitParser            *CommitParser
	commitMessageTemplate   string // Template for commit message suggestions
}

// NewFileChangeTracker creates a new file change tracker
func NewFileChangeTracker(gitService GitServiceInterface, storage taskmaster.Storage, repoPath string) *FileChangeTracker {
	return &FileChangeTracker{
		gitService:            gitService,
		storage:               storage,
		repoPath:              repoPath,
		stopChan:              make(chan struct{}),
		commitParser:          NewCommitParser(),
		commitMessageTemplate: "Implement #{{.TaskID}}", // Default template
	}
}

// Initialize loads existing data and starts tracking
func (t *FileChangeTracker) Initialize(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Load mapping from storage
	mapping, err := t.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load file change mapping: %w", err)
	}

	// Initialize with loaded or empty mapping
	needsSave := false
	if mapping == nil {
		mapping = taskmaster.NewFileChangeMapping(taskmaster.SerializationVersion)
		needsSave = true
	}
	t.mapping = mapping

	// Process historical commits to populate initial task associations
	// Only do this if we started with an empty mapping
	if needsSave {
		if err := t.processHistoricalCommits(ctx); err != nil {
			// Log warning but don't fail initialization
			// Historical commit processing is best-effort
			fmt.Printf("Warning: failed to process historical commits: %v\n", err)
		}
		
		// Save initial state only if we created a new mapping
		if err := t.storage.Save(t.mapping); err != nil {
			return fmt.Errorf("failed to save initial mapping: %w", err)
		}
	}

	return nil
}

// SetActiveTask sets the currently active task for change tracking
func (t *FileChangeTracker) SetActiveTask(taskID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.activeTask = taskID
}

// GetActiveTask returns the currently active task ID
func (t *FileChangeTracker) GetActiveTask() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.activeTask
}

// GetFileChangesForTask returns file changes for a specific task
func (t *FileChangeTracker) GetFileChangesForTask(taskID string) []taskmaster.FileChange {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	if t.mapping == nil {
		return []taskmaster.FileChange{}
	}
	
	return t.mapping.GetChangesForTask(taskID)
}

// GetAllTaskFileChanges returns a copy of the entire task->changes mapping
func (t *FileChangeTracker) GetAllTaskFileChanges() map[string][]taskmaster.FileChange {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	if t.mapping == nil {
		return make(map[string][]taskmaster.FileChange)
	}
	
	// Return a copy to prevent external modification
	result := make(map[string][]taskmaster.FileChange, len(t.mapping.Tasks))
	for taskID, changes := range t.mapping.Tasks {
		result[taskID] = append([]taskmaster.FileChange{}, changes...)
	}
	
	return result
}

// GetUnassignedChanges returns changes not associated with any task
func (t *FileChangeTracker) GetUnassignedChanges() []taskmaster.FileChange {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	if t.mapping == nil {
		return []taskmaster.FileChange{}
	}
	
	// Return a copy to prevent external modification
	return append([]taskmaster.FileChange{}, t.mapping.UnassignedChanges...)
}

// StartPeriodicRefresh begins periodic refreshing of changes
func (t *FileChangeTracker) StartPeriodicRefresh(interval time.Duration) {
	t.mutex.Lock()
	if t.refreshTicker != nil {
		// Already running
		t.mutex.Unlock()
		return
	}
	
	t.refreshTicker = time.NewTicker(interval)
	tickerC := t.refreshTicker.C
	// Create a new stop channel for this refresh cycle
	stopChan := make(chan struct{})
	t.stopChan = stopChan
	t.mutex.Unlock()
	
	go func() {
		for {
			select {
			case <-tickerC:
				ctx := context.Background()
				if err := t.RefreshChanges(ctx); err != nil {
					fmt.Printf("Warning: periodic refresh failed: %v\n", err)
				}
			case <-stopChan:
				return
			}
		}
	}()
}

// Stop stops the periodic refresh and cleans up resources
func (t *FileChangeTracker) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	// Stop refresh ticker
	if t.refreshTicker != nil {
		t.refreshTicker.Stop()
		t.refreshTicker = nil
	}
	
	// Signal stop to goroutine if channel is open
	if t.stopChan != nil {
		select {
		case t.stopChan <- struct{}{}:
		default:
			// Channel might be full or closed, that's ok
		}
		t.stopChan = nil
	}
}

// RefreshChanges updates the file change mapping with latest changes
func (t *FileChangeTracker) RefreshChanges(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	// Detect uncommitted changes
	if err := t.detectUncommittedChanges(ctx); err != nil {
		return fmt.Errorf("failed to detect uncommitted changes: %w", err)
	}
	
	// Save updated mapping to storage
	if err := t.storage.Save(t.mapping); err != nil {
		return fmt.Errorf("failed to save file change mapping: %w", err)
	}
	
	return nil
}

// detectUncommittedChanges finds and processes uncommitted changes
func (t *FileChangeTracker) detectUncommittedChanges(ctx context.Context) error {
	// Skip if gitService is not available
	if t.gitService == nil {
		return nil
	}
	
	// Get uncommitted changes from git
	gitChanges, err := t.gitService.GetUncommittedChanges(ctx)
	if err != nil {
		return fmt.Errorf("failed to get uncommitted changes from git: %w", err)
	}
	
	// Clear existing pending changes before adding new ones
	// This ensures we don't accumulate stale pending changes
	t.clearPendingChanges()
	
	// Process each git change
	for _, gitChange := range gitChanges {
		// Convert git.FileChange to taskmaster.FileChange
		fc := convertGitChangeToFileChange(gitChange)
		
		// Associate change with task (or unassigned)
		t.associateChangeWithTask(fc)
	}
	
	return nil
}

// associateChangeWithTask associates a file change with the active task or adds to unassigned
func (t *FileChangeTracker) associateChangeWithTask(fc taskmaster.FileChange) {
	// Associate with active task if set
	if t.activeTask != "" {
		if err := t.mapping.AddFileChange(t.activeTask, fc); err != nil {
			// Log warning but continue processing other changes
			fmt.Printf("Warning: failed to add file change for task %s: %v\n", t.activeTask, err)
		}
	} else {
		// Add to unassigned changes
		if err := t.mapping.AddUnassignedChange(fc); err != nil {
			fmt.Printf("Warning: failed to add unassigned file change: %v\n", err)
		}
	}
}

// AssociateChangesWithTask manually associates files with a task
func (t *FileChangeTracker) AssociateChangesWithTask(ctx context.Context, files []string, taskID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.mapping == nil {
		return fmt.Errorf("mapping not initialized")
	}
	
	if taskID == "" {
		return fmt.Errorf("taskID cannot be empty")
	}
	
	if len(files) == 0 {
		return fmt.Errorf("files list cannot be empty")
	}
	
	// Process each file
	for _, filePath := range files {
		if filePath == "" {
			continue
		}
		
		// Check if file is in unassigned changes
		foundInUnassigned := false
		for i, change := range t.mapping.UnassignedChanges {
			if change.Path == filePath {
				// Move from unassigned to task
				if err := t.mapping.AddFileChange(taskID, change); err != nil {
					return fmt.Errorf("failed to add file change for %s: %w", filePath, err)
				}
				// Remove from unassigned
				t.mapping.UnassignedChanges = append(t.mapping.UnassignedChanges[:i], t.mapping.UnassignedChanges[i+1:]...)
				foundInUnassigned = true
				break
			}
		}
		
		if !foundInUnassigned {
			// Check if file exists in another task (conflict detection)
			conflictTaskID := t.findFileInTasks(filePath)
			if conflictTaskID != "" && conflictTaskID != taskID {
				// Move from old task to new task
				if err := t.moveChangeBetweenTasks(filePath, conflictTaskID, taskID); err != nil {
					return fmt.Errorf("failed to move file %s from task %s to %s: %w", filePath, conflictTaskID, taskID, err)
				}
			} else if conflictTaskID == "" {
				// File not found anywhere, create new manual association
				fc, err := taskmaster.NewFileChange(filePath, "modified", "Manually associated")
				if err != nil {
					return fmt.Errorf("failed to create file change for %s: %w", filePath, err)
				}
				fc.IsPending = false
				fc.LastChanged = time.Now()
				
				if err := t.mapping.AddFileChange(taskID, *fc); err != nil {
					return fmt.Errorf("failed to add manual file change for %s: %w", filePath, err)
				}
			}
		}
	}
	
	// Save updated mapping to storage
	if err := t.storage.Save(t.mapping); err != nil {
		return fmt.Errorf("failed to save file change mapping: %w", err)
	}
	
	return nil
}

// RemoveFileChangeFromTask removes a file change from a task
func (t *FileChangeTracker) RemoveFileChangeFromTask(ctx context.Context, filePath string, taskID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.mapping == nil {
		return fmt.Errorf("mapping not initialized")
	}
	
	if taskID == "" {
		return fmt.Errorf("taskID cannot be empty")
	}
	
	if filePath == "" {
		return fmt.Errorf("filePath cannot be empty")
	}
	
	// Get changes for the task
	changes := t.mapping.Tasks[taskID]
	if len(changes) == 0 {
		return fmt.Errorf("task %s has no file changes", taskID)
	}
	
	// Find and remove the file change
	found := false
	var updatedChanges []taskmaster.FileChange
	for _, change := range changes {
		if change.Path != filePath {
			updatedChanges = append(updatedChanges, change)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("file %s not found in task %s", filePath, taskID)
	}
	
	// Update task changes
	t.mapping.Tasks[taskID] = updatedChanges
	
	// Save updated mapping to storage
	if err := t.storage.Save(t.mapping); err != nil {
		return fmt.Errorf("failed to save file change mapping: %w", err)
	}
	
	return nil
}

// MoveChangeBetweenTasks moves a file change from one task to another
func (t *FileChangeTracker) MoveChangeBetweenTasks(ctx context.Context, filePath string, fromTaskID string, toTaskID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.mapping == nil {
		return fmt.Errorf("mapping not initialized")
	}
	
	if fromTaskID == "" || toTaskID == "" {
		return fmt.Errorf("taskIDs cannot be empty")
	}
	
	if filePath == "" {
		return fmt.Errorf("filePath cannot be empty")
	}
	
	if fromTaskID == toTaskID {
		return fmt.Errorf("source and destination task IDs are the same")
	}
	
	return t.moveChangeBetweenTasks(filePath, fromTaskID, toTaskID)
}

// moveChangeBetweenTasks is the internal helper for moving file changes
// Must be called with lock held
func (t *FileChangeTracker) moveChangeBetweenTasks(filePath string, fromTaskID string, toTaskID string) error {
	// Get changes from source task
	fromChanges := t.mapping.Tasks[fromTaskID]
	if len(fromChanges) == 0 {
		return fmt.Errorf("source task %s has no file changes", fromTaskID)
	}
	
	// Find the file change
	var changeToMove *taskmaster.FileChange
	var updatedFromChanges []taskmaster.FileChange
	for _, change := range fromChanges {
		if change.Path == filePath {
			changeToMove = &change
		} else {
			updatedFromChanges = append(updatedFromChanges, change)
		}
	}
	
	if changeToMove == nil {
		return fmt.Errorf("file %s not found in task %s", filePath, fromTaskID)
	}
	
	// Update source task
	t.mapping.Tasks[fromTaskID] = updatedFromChanges
	
	// Add to destination task
	if err := t.mapping.AddFileChange(toTaskID, *changeToMove); err != nil {
		// Rollback: restore to source task
		t.mapping.Tasks[fromTaskID] = fromChanges
		return fmt.Errorf("failed to add file change to task %s: %w", toTaskID, err)
	}
	
	// Save updated mapping to storage
	if err := t.storage.Save(t.mapping); err != nil {
		// Rollback both changes
		t.mapping.Tasks[fromTaskID] = fromChanges
		t.mapping.Tasks[toTaskID] = t.mapping.Tasks[toTaskID][:len(t.mapping.Tasks[toTaskID])-1]
		return fmt.Errorf("failed to save file change mapping: %w", err)
	}
	
	return nil
}

// findFileInTasks searches for a file path across all tasks
// Must be called with lock held
func (t *FileChangeTracker) findFileInTasks(filePath string) string {
	for taskID, changes := range t.mapping.Tasks {
		for _, change := range changes {
			if change.Path == filePath {
				return taskID
			}
		}
	}
	return ""
}


// clearPendingChanges removes all pending (uncommitted) changes from the mapping
func (t *FileChangeTracker) clearPendingChanges() {
	// Clear pending changes from all tasks while preserving committed changes
	for taskID, changes := range t.mapping.Tasks {
		var committedChanges []taskmaster.FileChange
		for _, change := range changes {
			// Keep only committed changes (IsPending == false)
			if !change.IsPending {
				committedChanges = append(committedChanges, change)
			}
		}
		
		// Update the task's changes list
		// If no committed changes remain, keep an empty slice (don't delete the task entirely)
		// This preserves task associations even if all current changes are pending
		t.mapping.Tasks[taskID] = committedChanges
	}
	
	// Clear pending unassigned changes
	var committedUnassigned []taskmaster.FileChange
	for _, change := range t.mapping.UnassignedChanges {
		if !change.IsPending {
			committedUnassigned = append(committedUnassigned, change)
		}
	}
	t.mapping.UnassignedChanges = committedUnassigned
}

// ProcessCommit processes a single commit for task associations
// Extracts task IDs from the commit message and associates the files with those tasks
func (t *FileChangeTracker) ProcessCommit(commitID, message string, files []taskmaster.FileChange) error {
	if t.commitParser == nil {
		return fmt.Errorf("commit parser not initialized")
	}
	
	// Extract task IDs from commit message
	taskIDs := t.commitParser.ExtractTaskIDs(message)
	if len(taskIDs) == 0 {
		// No task IDs found, add to unassigned changes
		for _, file := range files {
			fc := file
			fc.CommitID = commitID
			fc.IsPending = false
			if err := t.mapping.AddUnassignedChange(fc); err != nil {
				fmt.Printf("Warning: failed to add unassigned file change: %v\n", err)
			}
		}
		return nil
	}
	
	// Associate each file with each task
	for _, taskID := range taskIDs {
		for _, file := range files {
			fc := file
			fc.CommitID = commitID
			fc.IsPending = false
			
			// Validate file change
			if err := fc.Validate(); err != nil {
				fmt.Printf("Warning: invalid file change for %s: %v\n", file.Path, err)
				continue
			}
			
			// Add to task's file changes
			if err := t.mapping.AddFileChange(taskID, fc); err != nil {
				fmt.Printf("Warning: failed to add file change for task %s: %v\n", taskID, err)
			}
		}
	}
	
	return nil
}

// ProcessCommitHistory processes the git commit history for task associations
// Retrieves commits from git history, extracts task IDs, and associates file changes
func (t *FileChangeTracker) ProcessCommitHistory(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	// Skip if gitService is not available
	if t.gitService == nil {
		return nil
	}
	
	// Get recent commits with task IDs (limit to 100 for performance)
	opts := git.CommitFilterOptions{
		Limit:    100,
		NoMerges: true,
	}
	
	mapping, err := t.gitService.GetCommitsWithTaskIDsFiltered(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to get commits with task IDs: %w", err)
	}
	
	// Process each commit for task associations
	for _, commits := range mapping.TaskToCommits {
		for _, commit := range commits {
			// Get files changed in this commit
			files, err := t.getFilesInCommit(ctx, commit.Hash)
			if err != nil {
				// Log warning but continue processing other commits
				fmt.Printf("Warning: failed to get files for commit %s: %v\n", commit.Hash, err)
				continue
			}
			
			// Convert git.FileChange to taskmaster.FileChange
			var fcFiles []taskmaster.FileChange
			for _, file := range files {
				fc := taskmaster.FileChange{
					Path:        file.Path,
					ChangeType:  string(file.ChangeType),
					Description: fmt.Sprintf("From commit %s: %s", commit.Hash[:7], commit.Message),
					LastChanged: time.Now(),
					CommitID:    commit.Hash,
					IsPending:   false,
				}
				fcFiles = append(fcFiles, fc)
			}
			
			// Process this commit with the files
			if err := t.ProcessCommit(commit.Hash, commit.Message, fcFiles); err != nil {
				fmt.Printf("Warning: failed to process commit %s: %v\n", commit.Hash, err)
			}
		}
	}
	
	return nil
}

// GetCommitMessageSuggestion generates a commit message with the active task ID
// Returns a formatted suggestion with the active task ID, or empty string if no active task
// Supports template-based formatting using Go text/template syntax
func (t *FileChangeTracker) GetCommitMessageSuggestion() string {
	t.mutex.RLock()
	activeTask := t.activeTask
	template := t.commitMessageTemplate
	t.mutex.RUnlock()
	
	if activeTask == "" {
		return ""
	}
	
	return t.generateMessageFromTemplate(template, activeTask)
}

// SetCommitMessageTemplate sets a custom template for commit message suggestions
// Template should use Go text/template syntax with available variables:
// - TaskID: The active task ID
// - Example: "feat: implement #{{.TaskID}}"
func (t *FileChangeTracker) SetCommitMessageTemplate(template string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.commitMessageTemplate = template
}

// GetCommitMessageTemplate returns the current commit message template
func (t *FileChangeTracker) GetCommitMessageTemplate() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.commitMessageTemplate
}

// ResetCommitMessageTemplate resets the template to the default value
func (t *FileChangeTracker) ResetCommitMessageTemplate() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.commitMessageTemplate = "Implement #{{.TaskID}}"
}

// generateMessageFromTemplate generates a commit message using the provided template
// Supports simple variable substitution for TaskID
func (t *FileChangeTracker) generateMessageFromTemplate(tmpl string, taskID string) string {
	// Simple string replacement if template is basic
	if strings.Contains(tmpl, "{{.TaskID}}") {
		return strings.ReplaceAll(tmpl, "{{.TaskID}}", taskID)
	}
	
	// Try Go text/template if it's more complex
	templ, err := template.New("msg").Parse(tmpl)
	if err != nil {
		// Fall back to simple format if template is invalid
		return fmt.Sprintf("Implement #%s", taskID)
	}
	
	data := map[string]string{
		"TaskID": taskID,
	}
	
	var buf strings.Builder
	if err := templ.Execute(&buf, data); err != nil {
		// Fall back to simple format if execution fails
		return fmt.Sprintf("Implement #%s", taskID)
	}
	
	return buf.String()
}



// convertGitChangeToFileChange converts a git.FileChange to taskmaster.FileChange
func convertGitChangeToFileChange(gitChange git.FileChange) taskmaster.FileChange {
	changeType := gitChangeTypeToString(gitChange.ChangeType)
	
	// Build description based on change type and status
	description := buildChangeDescription(gitChange)
	
	return taskmaster.FileChange{
		Path:        gitChange.Path,
		ChangeType:  changeType,
		Description: description,
		LastChanged: time.Now(),
		CommitID:    "", // Empty for uncommitted changes
		IsPending:   true,
	}
}

// gitChangeTypeToString converts git.ChangeType to string for taskmaster.FileChange
func gitChangeTypeToString(ct git.ChangeType) string {
	switch ct {
	case git.ChangeTypeAdded:
		return "added"
	case git.ChangeTypeModified:
		return "modified"
	case git.ChangeTypeDeleted:
		return "deleted"
	case git.ChangeTypeRenamed:
		return "modified" // Treat rename as modified for now
	case git.ChangeTypeCopied:
		return "added" // Treat copy as added
	case git.ChangeTypeUntracked:
		return "added" // Treat untracked as added
	default:
		return "modified" // Default to modified
	}
}

// buildChangeDescription creates a human-readable description of the change
func buildChangeDescription(gitChange git.FileChange) string {
	switch gitChange.ChangeType {
	case git.ChangeTypeRenamed:
		if gitChange.OldPath != "" {
			return fmt.Sprintf("Renamed from %s", gitChange.OldPath)
		}
		return "Renamed file"
	case git.ChangeTypeCopied:
		if gitChange.OldPath != "" {
			return fmt.Sprintf("Copied from %s", gitChange.OldPath)
		}
		return "Copied file"
	case git.ChangeTypeUntracked:
		return "Untracked file"
	case git.ChangeTypeAdded:
		if gitChange.IsStaged {
			return "Staged for addition"
		}
		return "New file"
	case git.ChangeTypeModified:
		if gitChange.IsStaged && gitChange.IsModified {
			return "Staged and modified in working tree"
		} else if gitChange.IsStaged {
			return "Staged for commit"
		} else if gitChange.IsModified {
			return "Modified in working tree"
		}
		return "Modified"
	case git.ChangeTypeDeleted:
		if gitChange.IsStaged {
			return "Staged for deletion"
		}
		return "Deleted"
	default:
		return "Uncommitted change"
	}
}

// processHistoricalCommits analyzes git history for task associations
func (t *FileChangeTracker) processHistoricalCommits(ctx context.Context) error {
	// Skip if gitService is not available
	if t.gitService == nil {
		return nil
	}
	
	// Get recent commits with task IDs (limit to 100 for performance)
	opts := git.CommitFilterOptions{
		Limit:    100,
		NoMerges: true,
	}
	
	mapping, err := t.gitService.GetCommitsWithTaskIDsFiltered(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to get commits with task IDs: %w", err)
	}
	
	// For each task that has commits, we want to associate those file changes
	// We'll get the files changed in each commit that references a task
	for _, commits := range mapping.TaskToCommits {
		for _, commit := range commits {
			// Get files changed in this commit
			files, err := t.getFilesInCommit(ctx, commit.Hash)
			if err != nil {
				// Log warning but continue processing other commits
				fmt.Printf("Warning: failed to get files for commit %s: %v\n", commit.Hash, err)
				continue
			}
			
			// Convert git.FileChange to taskmaster.FileChange
			var fcFiles []taskmaster.FileChange
			for _, file := range files {
				fc := taskmaster.FileChange{
					Path:        file.Path,
					ChangeType:  string(file.ChangeType),
					Description: fmt.Sprintf("From commit %s: %s", commit.Hash[:7], commit.Message),
					LastChanged: time.Now(),
					CommitID:    commit.Hash,
					IsPending:   false,
				}
				fcFiles = append(fcFiles, fc)
			}
			
			// Process the commit using ProcessCommit
			if err := t.ProcessCommit(commit.Hash, commit.Message, fcFiles); err != nil {
				fmt.Printf("Warning: failed to process commit %s: %v\n", commit.Hash, err)
			}
		}
	}
	
	return nil
}

// getFilesInCommit retrieves the files changed in a specific commit
func (t *FileChangeTracker) getFilesInCommit(ctx context.Context, commitHash string) ([]git.FileChange, error) {
	// Use git diff-tree to get files in commit
	// This is more efficient than checking out the commit
	// The GitService doesn't have this method yet, so we'll need to add it
	// For now, return empty list (will be enhanced in future subtasks)
	return []git.FileChange{}, nil
}
