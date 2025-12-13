# Product Requirements Document: Task Expansion CLI Integration Refactoring

**Version:** 1.0  
**Date:** 2025-12-12  
**Author:** Development Team  
**Status:** Planning  

---

## Executive Summary

The current task expansion implementation (task 5) uses in-memory/local Go functions (`ExpandTaskDrafts`, `ApplySubtaskDrafts`) instead of calling the Task Master CLI commands (`task-master expand --id=<id>` or `task-master expand --all`). This refactoring will align it with the complexity analysis pattern (task 3), which properly executes CLI commands with progress reporting.

**Goal:** Transform task expansion from a local, in-memory operation to a CLI-driven, externally-executed workflow with real-time progress reporting.

---

## Current State Analysis

### Current Implementation (Task Expansion)
**Files:** `internal/ui/command_handlers.go:615-721`, `internal/taskmaster/service.go:468-514`

**Flow:**
1. User triggers expansion → Options dialog
2. `runExpandTask()` calls local Go function `taskmaster.ExpandTaskDrafts()`
3. Drafts generated in-memory
4. Preview/Edit dialogs shown
5. `ApplySubtaskDrafts()` modifies task tree locally
6. Manual `LoadTasks()` reload

**Issues:**
- ❌ Bypasses Task Master CLI entirely
- ❌ No actual `task-master expand` command execution
- ❌ Changes not persisted via CLI (may cause sync issues)
- ❌ No real progress reporting during CLI execution
- ❌ Cannot support `--research` flag properly (AI-powered expansion)
- ❌ Doesn't support `--all` option for batch expansion

### Reference Implementation (Complexity Analysis)
**Files:** `internal/ui/complexity.go`, `internal/taskmaster/service.go:402-455`

**Flow:**
1. User triggers analysis → Scope dialog → `ComplexityScopeSelectedMsg`
2. `handleComplexityScopeSelected()` creates progress dialog
3. Goroutine calls `AnalyzeComplexityWithProgress()` (pure Go analysis, no CLI)
4. Progress channel streams updates → `ComplexityAnalysisProgressMsg`
5. Completion → `ComplexityAnalysisCompletedMsg` → Report dialog

**Pattern to adopt:**
- ✅ Dialog-based workflow with scope selection
- ✅ Progress dialog with goroutine + channel
- ✅ Typed messages for progress and completion
- ✅ Proper cancellation support
- ✅ Clean state management

**Note:** Complexity analysis doesn't use CLI either, but it's purely analytical. Expansion **must** use CLI because it modifies tasks.

---

## Target Architecture

### CLI Command Support
```bash
# Single task expansion
task-master expand --id=1.2 [--research] [--depth=2] [--num=5]

# Batch expansion
task-master expand --all [--research] [--from=1] [--to=5]
```

### New Flow Pattern

```
User Action (Alt+X)
    ↓
ExpansionScopeDialog (new)
    - Options: "Selected Task", "All Tasks", "Task Range", "By Tag"
    - Expansion depth (1-3)
    - Number of subtasks (optional)
    - Research flag (--research)
    ↓
ExpansionScopeSelectedMsg
    ↓
handleExpansionScopeSelected()
    - Creates ExpandProgressDialog
    - Starts goroutine with CLI execution
    ↓
Goroutine: ExecuteExpandWithProgress()
    - Calls `task-master expand` with args
    - Streams stdout/stderr
    - Sends progress updates
    ↓
ExpansionProgressMsg (periodic updates)
    ↓
ExpansionCompletedMsg
    ↓
handleExpansionCompleted()
    - Closes progress dialog
    - Reloads tasks
    - Shows notification/report
```

---

## Git Workflow

### Branch Management

**Branch Name:** `refactor/task-expansion-cli-integration`

**Branch Strategy:**
1. Create feature branch from `main`
2. Commit changes incrementally by phase
3. Final PR review before merge

**Commit Structure:**
```
Phase 0: Setup
- [setup] Create feature branch and PRD

Phase 1: Service Layer
- [service] Add ExecuteExpandWithProgress method
- [service] Update ExpandProgressState type
- [service] Add CLI command building logic
- [service] Add progress parsing from CLI output
- [service] Add streaming output handling

Phase 2: UI Layer
- [ui] Create ExpansionScopeDialog component
- [ui] Create expansion workflow file
- [ui] Create ExpansionProgressDialog
- [ui] Update command handlers for new flow
- [ui] Add expansion message types
- [ui] Update app state for expansion tracking
- [ui] Remove deprecated expansion code

Phase 3: Testing
- [test] Add service layer unit tests
- [test] Add UI workflow tests
- [test] Add integration tests
- [test] Update existing test mocks

Phase 4: CLI Integration
- [cli] Verify CLI command support
- [cli] Update CLI wrapper methods
- [cli] Add error handling for CLI failures

Phase 5: Documentation
- [docs] Update README.md
- [docs] Update AGENTS.md
- [docs] Add migration notes
- [docs] Update inline code documentation

Final: PR Preparation
- [final] Run all tests and linting
- [final] Update CHANGELOG
- [final] PR description and review checklist
```

### Git Commands Sequence

```bash
# Phase 0: Setup
git checkout main
git pull origin main
git checkout -b refactor/task-expansion-cli-integration
git add TASK_EXPANSION_FIX.md
git commit -m "docs: Add PRD for task expansion CLI integration refactor"

# Work progresses through phases...
# Each phase has multiple commits (see commit structure above)

# Final preparation
git log --oneline main..HEAD  # Review all commits
git diff main --stat           # Review all changes
# Create PR when ready
```

---

## Detailed Refactoring Steps

### Phase 0: Setup and Planning

#### 0.1 Create Git Branch
```bash
git checkout main
git pull origin main
git checkout -b refactor/task-expansion-cli-integration
```

#### 0.2 Create PRD Document
- [x] Create `TASK_EXPANSION_FIX.md` in project root
- [ ] Commit PRD to feature branch
```bash
git add TASK_EXPANSION_FIX.md
git commit -m "docs: Add PRD for task expansion CLI integration refactor"
```

#### 0.3 Verify Task Master CLI Capabilities
- [ ] Test `task-master expand --id=<id>` locally
- [ ] Test `task-master expand --all` locally
- [ ] Document CLI output format for progress parsing
- [ ] Verify `--research`, `--depth`, `--num` flag support
- [ ] Document any missing features needed

---

### Phase 1: Service Layer Updates

#### 1.1 Create New Service Method
**File:** `internal/taskmaster/service.go`

**Task:** Add `ExecuteExpandWithProgress()` method

```go
// ExecuteExpandWithProgress executes task-master expand CLI command with progress reporting
func (s *Service) ExecuteExpandWithProgress(
    ctx context.Context, 
    scope string,           // "single", "all", "range", "tag"
    taskID string,          // for single task
    fromID string,          // for range
    toID string,            // for range
    tags []string,          // for tag-based
    opts ExpandTaskOptions,
    onProgress func(ExpandProgressState),
) error {
    // Build CLI command args
    args := []string{"expand"}
    
    switch scope {
    case "single":
        args = append(args, fmt.Sprintf("--id=%s", taskID))
    case "all":
        args = append(args, "--all")
    case "range":
        if fromID != "" {
            args = append(args, fmt.Sprintf("--from=%s", fromID))
        }
        if toID != "" {
            args = append(args, fmt.Sprintf("--to=%s", toID))
        }
    case "tag":
        // May need CLI support for tag-based expansion
        for _, tag := range tags {
            args = append(args, fmt.Sprintf("--tag=%s", tag))
        }
    }
    
    if opts.UseAI {
        args = append(args, "--research")
    }
    if opts.Depth > 0 {
        args = append(args, fmt.Sprintf("--depth=%d", opts.Depth))
    }
    if opts.NumSubtasks > 0 {
        args = append(args, fmt.Sprintf("--num=%d", opts.NumSubtasks))
    }
    
    // Execute command with streaming output
    cmd := exec.CommandContext(ctx, "task-master", args...)
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to create stdout pipe: %w", err)
    }
    
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("failed to create stderr pipe: %w", err)
    }
    
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start command: %w", err)
    }
    
    // Stream output and parse progress
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            line := scanner.Text()
            // Parse progress from output
            if onProgress != nil {
                state := parseExpandProgress(line)
                onProgress(state)
            }
        }
    }()
    
    // Capture errors
    errOutput, _ := io.ReadAll(stderr)
    
    // Wait for completion
    if err := cmd.Wait(); err != nil {
        return fmt.Errorf("command failed: %w\n%s", err, errOutput)
    }
    
    // Reload tasks after expansion
    return s.LoadTasks(ctx)
}

// parseExpandProgress parses progress information from CLI output
func parseExpandProgress(line string) ExpandProgressState {
    // TODO: Implement based on actual CLI output format
    // Example parsing:
    // "Expanding task 1.2..." → Stage: "Expanding", CurrentTask: "1.2"
    // "Generated 5 subtasks" → TasksExpanded: 5
    // "Progress: 3/10" → Progress: 0.3
    
    return ExpandProgressState{
        Stage:    "Processing",
        Message:  line,
        Progress: 0.5, // Placeholder
    }
}
```

**Git commit:**
```bash
git add internal/taskmaster/service.go
git commit -m "service: Add ExecuteExpandWithProgress method for CLI integration"
```

#### 1.2 Update Progress State Type
**File:** `internal/taskmaster/types.go`

```go
// ExpandProgressState represents the current state of task expansion
type ExpandProgressState struct {
    Stage         string  // "Analyzing", "Generating", "Applying", "Complete"
    Progress      float64 // 0.0 to 1.0
    CurrentTask   string  // Task ID being expanded
    TasksExpanded int     // Number of tasks expanded so far
    TotalTasks    int     // Total tasks to expand
    Message       string  // Status message from CLI
    SubtasksCreated int   // Total subtasks created
}
```

**Git commit:**
```bash
git add internal/taskmaster/types.go
git commit -m "service: Update ExpandProgressState for CLI progress tracking"
```

#### 1.3 Mark Deprecated Functions
**File:** `internal/taskmaster/expand.go`

Add deprecation notices:
```go
// Deprecated: Use ExecuteExpandWithProgress instead.
// This function is kept for testing purposes only.
// ExpandTaskDrafts generates subtask drafts locally without using CLI.
func ExpandTaskDrafts(task *Task, opts ExpandTaskOptions) []SubtaskDraft {
    // ... existing implementation
}

// Deprecated: CLI handles task application automatically.
// This function is kept for testing purposes only.
// ApplySubtaskDrafts applies subtask drafts to a parent task.
func ApplySubtaskDrafts(parent *Task, drafts []SubtaskDraft) ([]string, error) {
    // ... existing implementation
}
```

**Git commit:**
```bash
git add internal/taskmaster/expand.go
git commit -m "service: Deprecate local expansion functions in favor of CLI"
```

---

### Phase 2: UI Layer Updates

#### 2.1 Create Expansion Scope Dialog
**File:** `internal/ui/dialog/expansion_scope.go` (new file)

```go
package dialog

import (
    "fmt"
    
    tea "github.com/charmbracelet/bubbletea"
)

// ExpansionScopeDialog allows user to select expansion scope and options
type ExpansionScopeDialog struct {
    BaseDialog
    form          *FormDialog
    selectedTask  string
}

// ExpansionScopeResult contains the user's expansion configuration
type ExpansionScopeResult struct {
    Scope       string   // "single", "all", "range", "tag"
    TaskID      string   // for single task expansion
    FromID      string   // for range expansion
    ToID        string   // for range expansion
    Tags        []string // for tag-based expansion
    Depth       int      // 1-3 levels
    NumSubtasks int      // optional, 0 = auto
    UseAI       bool     // --research flag
}

func NewExpansionScopeDialog(selectedTaskID string, style *DialogStyle) (*ExpansionScopeDialog, error) {
    fields := []FormField{
        {
            ID:    "scope",
            Label: "Expansion scope",
            Type:  FormFieldTypeRadio,
            Options: []FormOption{
                {Value: "single", Label: "Selected task only"},
                {Value: "all", Label: "All tasks"},
                {Value: "range", Label: "Task range (from/to)"},
                {Value: "tag", Label: "Tasks by tag"},
            },
            Value: func() string {
                if selectedTaskID != "" {
                    return "single"
                }
                return "all"
            }(),
            Required: true,
        },
        {
            ID:          "taskID",
            Label:       "Task ID (for single scope)",
            Type:        FormFieldTypeText,
            Value:       selectedTaskID,
            Placeholder: "e.g., 1.2",
            Condition: func(values map[string]interface{}) bool {
                return stringValue(values, "scope") == "single"
            },
        },
        {
            ID:          "fromID",
            Label:       "From task ID (for range)",
            Type:        FormFieldTypeText,
            Placeholder: "e.g., 1",
            Condition: func(values map[string]interface{}) bool {
                return stringValue(values, "scope") == "range"
            },
        },
        {
            ID:          "toID",
            Label:       "To task ID (for range)",
            Type:        FormFieldTypeText,
            Placeholder: "e.g., 5",
            Condition: func(values map[string]interface{}) bool {
                return stringValue(values, "scope") == "range"
            },
        },
        {
            ID:          "tags",
            Label:       "Tags (comma-separated)",
            Type:        FormFieldTypeText,
            Placeholder: "e.g., backend,api",
            Condition: func(values map[string]interface{}) bool {
                return stringValue(values, "scope") == "tag"
            },
        },
        {
            ID:    "depth",
            Label: "Expansion depth",
            Type:  FormFieldTypeRadio,
            Options: []FormOption{
                {Value: "1", Label: "1 level"},
                {Value: "2", Label: "2 levels"},
                {Value: "3", Label: "3 levels"},
            },
            Value: "2",
        },
        {
            ID:          "num",
            Label:       "Number of subtasks per task",
            Type:        FormFieldTypeText,
            Placeholder: "Leave blank for auto-detection",
        },
        {
            ID:      "research",
            Label:   "Enable AI-powered expansion (--research)",
            Type:    FormFieldTypeCheckbox,
            Checked: true,
        },
    }
    
    form := NewFormDialog(
        "Expand Tasks",
        "Configure task expansion options. This will execute 'task-master expand' command.",
        fields,
        []string{"Expand", "Cancel"},
        style,
        func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
            if button != "Expand" {
                return nil, nil
            }
            
            scope := stringValue(values, "scope")
            if scope == "" {
                return nil, fmt.Errorf("scope is required")
            }
            
            result := ExpansionScopeResult{
                Scope:       scope,
                Depth:       parseIntValue(stringValue(values, "depth"), 2),
                NumSubtasks: parseIntValue(stringValue(values, "num"), 0),
                UseAI:       boolValue(values, "research"),
            }
            
            switch scope {
            case "single":
                result.TaskID = stringValue(values, "taskID")
                if result.TaskID == "" {
                    return nil, fmt.Errorf("task ID is required for single scope")
                }
            case "range":
                result.FromID = stringValue(values, "fromID")
                result.ToID = stringValue(values, "toID")
                if result.FromID == "" && result.ToID == "" {
                    return nil, fmt.Errorf("at least one of from/to ID is required for range")
                }
            case "tag":
                tagsStr := stringValue(values, "tags")
                if tagsStr == "" {
                    return nil, fmt.Errorf("tags are required for tag scope")
                }
                result.Tags = parseTagList(tagsStr)
            }
            
            return result, nil
        },
    )
    
    return &ExpansionScopeDialog{
        form:         form,
        selectedTask: selectedTaskID,
    }, nil
}

func (d *ExpansionScopeDialog) Init() tea.Cmd {
    return d.form.Init()
}

func (d *ExpansionScopeDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    d.form, cmd = d.form.Update(msg)
    return d, cmd
}

func (d *ExpansionScopeDialog) View() string {
    return d.form.View()
}

// Helper functions
func parseIntValue(s string, fallback int) int {
    if s == "" {
        return fallback
    }
    var result int
    fmt.Sscanf(s, "%d", &result)
    if result <= 0 {
        return fallback
    }
    return result
}

func parseTagList(s string) []string {
    // Split by comma and trim whitespace
    parts := strings.Split(s, ",")
    result := make([]string, 0, len(parts))
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            result = append(result, trimmed)
        }
    }
    return result
}
```

**Git commit:**
```bash
git add internal/ui/dialog/expansion_scope.go
git commit -m "ui: Create ExpansionScopeDialog for configuration selection"
```

#### 2.2 Create Expansion Workflow Handler
**File:** `internal/ui/expansion.go` (new file)

```go
package ui

import (
    "context"
    "fmt"
    "time"
    
    "github.com/adriangreen/tm-tui/internal/taskmaster"
    "github.com/adriangreen/tm-tui/internal/ui/dialog"
    tea "github.com/charmbracelet/bubbletea"
)

// Message types for expansion workflow

type expansionStreamClosedMsg struct{}

type ExpansionScopeSelectedMsg struct {
    Scope       string
    TaskID      string
    FromID      string
    ToID        string
    Tags        []string
    Depth       int
    NumSubtasks int
    UseAI       bool
}

type ExpansionProgressMsg struct {
    Progress        float64
    Stage           string
    CurrentTask     string
    TasksExpanded   int
    TotalTasks      int
    SubtasksCreated int
    Message         string
    Error           error
}

type ExpansionCompletedMsg struct {
    TasksExpanded   int
    SubtasksCreated int
    Error           error
}

// Workflow methods

func (m *Model) showExpansionScopeDialog() {
    dm := m.dialogManager()
    if dm == nil {
        return
    }
    
    // Get the ID of the selected task
    selectedTaskID := ""
    if m.selectedTask != nil {
        selectedTaskID = m.selectedTask.ID
    }
    
    // Create the dialog
    scopeDialog, err := dialog.NewExpansionScopeDialog(selectedTaskID, dm.Style)
    if err != nil {
        m.logLines = append(m.logLines, fmt.Sprintf("Error creating expansion scope dialog: %s", err))
        return
    }
    
    // Show the dialog and handle the result
    m.appState.AddDialog(scopeDialog, func(value interface{}, err error) tea.Cmd {
        if err != nil {
            return func() tea.Msg {
                return ErrorMsg{Err: err}
            }
        }
        
        if value == nil {
            return nil
        }
        
        result, ok := value.(dialog.ExpansionScopeResult)
        if !ok {
            return func() tea.Msg {
                return ErrorMsg{Err: fmt.Errorf("invalid result type from expansion scope dialog")}
            }
        }
        
        // Create a message with the selected scope
        return func() tea.Msg {
            return ExpansionScopeSelectedMsg{
                Scope:       result.Scope,
                TaskID:      result.TaskID,
                FromID:      result.FromID,
                ToID:        result.ToID,
                Tags:        result.Tags,
                Depth:       result.Depth,
                NumSubtasks: result.NumSubtasks,
                UseAI:       result.UseAI,
            }
        }
    })
}

func (m *Model) handleExpansionScopeSelected(msg ExpansionScopeSelectedMsg) tea.Cmd {
    dm := m.dialogManager()
    if dm == nil {
        return nil
    }
    
    // Determine total tasks to expand
    var totalTasks int
    switch msg.Scope {
    case "single":
        totalTasks = 1
    case "all":
        totalTasks = len(m.taskIndex)
    case "range":
        // Count tasks in range
        for id := range m.taskIndex {
            if (msg.FromID == "" || id >= msg.FromID) && (msg.ToID == "" || id <= msg.ToID) {
                totalTasks++
            }
        }
    case "tag":
        // Count tasks with matching tags
        for _, task := range m.taskIndex {
            for _, taskTag := range task.Tags {
                for _, selectedTag := range msg.Tags {
                    if taskTag == selectedTag {
                        totalTasks++
                        break
                    }
                }
            }
        }
    }
    
    m.currentExpansionScope = msg.Scope
    m.currentExpansionTags = append([]string(nil), msg.Tags...)
    
    // Create and show progress dialog
    progressDialog := dialog.NewProgressDialog(
        "Expanding Tasks",
        fmt.Sprintf("Running task-master expand with %s scope...", msg.Scope),
        dm.Style,
    )
    progressDialog.SetCancelable(true)
    
    m.appState.AddDialog(progressDialog, func(value interface{}, err error) tea.Cmd {
        if err != nil {
            return func() tea.Msg {
                return ErrorMsg{Err: err}
            }
        }
        
        // Handle cancellation
        if value == nil {
            m.cancelExpansion()
            return nil
        }
        
        return nil
    })
    
    m.expansionStartedAt = time.Now()
    
    // Start expansion work
    return m.startExpansion(msg.Scope, msg.TaskID, msg.FromID, msg.ToID, msg.Tags, taskmaster.ExpandTaskOptions{
        Depth:       msg.Depth,
        NumSubtasks: msg.NumSubtasks,
        UseAI:       msg.UseAI,
    })
}

func (m *Model) startExpansion(scope, taskID, fromID, toID string, tags []string, opts taskmaster.ExpandTaskOptions) tea.Cmd {
    if m.expansionCancel != nil {
        m.expansionCancel()
        m.expansionCancel = nil
    }
    
    progressCh := make(chan tea.Msg, 32)
    m.expansionMsgCh = progressCh
    
    ctx, cancel := context.WithCancel(context.Background())
    m.expansionCancel = cancel
    
    go func() {
        defer close(progressCh)
        
        err := m.taskService.ExecuteExpandWithProgress(
            ctx,
            scope,
            taskID,
            fromID,
            toID,
            tags,
            opts,
            func(state taskmaster.ExpandProgressState) {
                msg := ExpansionProgressMsg{
                    Progress:        state.Progress,
                    Stage:           state.Stage,
                    CurrentTask:     state.CurrentTask,
                    TasksExpanded:   state.TasksExpanded,
                    TotalTasks:      state.TotalTasks,
                    SubtasksCreated: state.SubtasksCreated,
                    Message:         state.Message,
                }
                select {
                case progressCh <- msg:
                case <-ctx.Done():
                }
            },
        )
        
        // Send completion message
        completionMsg := ExpansionCompletedMsg{Error: err}
        if err == nil {
            // Extract stats from final state if available
            // For now, we'll get this from the reload
            completionMsg.TasksExpanded = 1 // Placeholder
        }
        
        select {
        case progressCh <- completionMsg:
        case <-ctx.Done():
        }
    }()
    
    return m.waitForExpansionMessages()
}

func (m *Model) handleExpansionProgress(msg ExpansionProgressMsg) tea.Cmd {
    // Find the progress dialog
    if dm := m.dialogManager(); dm != nil {
        if progressDialog, ok := dm.GetDialogByType(dialog.DialogTypeProgress); ok {
            if pd, ok := progressDialog.(*dialog.ProgressDialog); ok {
                // Update progress
                pd.SetProgress(msg.Progress)
                pd.SetMessage(fmt.Sprintf("%s: %s", msg.Stage, msg.Message))
                
                if msg.CurrentTask != "" {
                    pd.SetSubtext(fmt.Sprintf("Current task: %s", msg.CurrentTask))
                }
            }
        }
    }
    
    return m.waitForExpansionMessages()
}

func (m *Model) handleExpansionCompleted(msg ExpansionCompletedMsg) tea.Cmd {
    dm := m.dialogManager()
    
    // Close the progress dialog
    if dm != nil {
        dm.RemoveDialogsByType(dialog.DialogTypeProgress)
    }
    
    m.clearExpansionRuntimeState()
    
    // Handle errors
    if msg.Error != nil {
        if errors.Is(msg.Error, context.Canceled) {
            m.ShowNotificationDialog("Expansion Cancelled", "Task expansion was cancelled.", "warning", 3*time.Second)
            return nil
        }
        m.ShowNotificationDialog("Expansion Failed", fmt.Sprintf("Error expanding tasks: %s", msg.Error), "error", 5*time.Second)
        return nil
    }
    
    // Success notification
    duration := time.Since(m.expansionStartedAt)
    message := fmt.Sprintf("Successfully expanded tasks in %s", duration.Round(time.Millisecond))
    m.ShowNotificationDialog("Expansion Complete", message, "success", 3*time.Second)
    
    // Reload tasks to show new subtasks
    return LoadTasksCmd(m.taskService)
}

func (m *Model) waitForExpansionMessages() tea.Cmd {
    ch := m.expansionMsgCh
    if ch == nil {
        return nil
    }
    
    return func() tea.Msg {
        if msg, ok := <-ch; ok {
            return msg
        }
        return expansionStreamClosedMsg{}
    }
}

func (m *Model) cancelExpansion() {
    if m.expansionCancel != nil {
        m.expansionCancel()
        m.expansionCancel = nil
    }
}

func (m *Model) clearExpansionRuntimeState() {
    m.cancelExpansion()
    m.expansionMsgCh = nil
    m.currentExpansionScope = ""
    m.currentExpansionTags = nil
    m.expansionStartedAt = time.Time{}
    m.waitingForExpansionHold = false
}
```

**Git commit:**
```bash
git add internal/ui/expansion.go
git commit -m "ui: Create expansion workflow handler with CLI integration"
```

#### 2.3 Update Message Types
**File:** `internal/ui/messages.go`

Add new message types:
```go
// Expansion workflow messages
type ExpansionScopeSelectedMsg struct {
    Scope       string
    TaskID      string
    FromID      string
    ToID        string
    Tags        []string
    Depth       int
    NumSubtasks int
    UseAI       bool
}

type ExpansionProgressMsg struct {
    Progress        float64
    Stage           string
    CurrentTask     string
    TasksExpanded   int
    TotalTasks      int
    SubtasksCreated int
    Message         string
    Error           error
}

type ExpansionCompletedMsg struct {
    TasksExpanded   int
    SubtasksCreated int
    Error           error
}
```

Remove deprecated:
```go
// DEPRECATED: Remove this line
// type expandTaskStreamClosedMsg struct{}
```

**Git commit:**
```bash
git add internal/ui/messages.go
git commit -m "ui: Add expansion workflow message types"
```

#### 2.4 Update App State
**File:** `internal/ui/app.go`

Remove old state:
```go
// REMOVE these fields from Model struct:
// expandTaskDrafts    []taskmaster.SubtaskDraft
// expandTaskParentID  string
// expandTaskMsgCh     chan tea.Msg
// expandTaskCancel    context.CancelFunc
```

Add new state:
```go
// Add to Model struct:
// Expansion workflow state
expansionMsgCh          chan tea.Msg
expansionCancel         context.CancelFunc
currentExpansionScope   string
currentExpansionTags    []string
expansionStartedAt      time.Time
waitingForExpansionHold bool
```

Update `Update()` method to handle new messages:
```go
case ExpansionScopeSelectedMsg:
    return m, m.handleExpansionScopeSelected(msg)
case ExpansionProgressMsg:
    return m, m.handleExpansionProgress(msg)
case ExpansionCompletedMsg:
    return m, m.handleExpansionCompleted(msg)
case expansionStreamClosedMsg:
    return m, m.waitForExpansionMessages()
```

**Git commit:**
```bash
git add internal/ui/app.go
git commit -m "ui: Update app state for CLI-based expansion workflow"
```

#### 2.5 Update Command Handlers
**File:** `internal/ui/command_handlers.go`

**Refactor `handleExpandTaskCommand()`:**
```go
func (m *Model) handleExpandTaskCommand() tea.Cmd {
    if m.taskService == nil || !m.taskService.IsAvailable() {
        appErr := NewDependencyError("Expand Task", "Task Master CLI is not available in this workspace.", nil).
            WithRecoveryHints(
                "Restart the TUI inside a Task Master workspace",
                "Check Task Master CLI installation",
            )
        m.showAppError(appErr)
        return nil
    }
    
    // Show scope dialog (new approach)
    m.showExpansionScopeDialog()
    return nil
}
```

**Remove deprecated functions:**
```go
// REMOVE these functions:
// - openTaskSelectionDialog() (lines 79-127)
// - startExpandWorkflow() (lines 172-246)
// - runExpandTask() (lines 615-647)
// - showExpandPreviewDialog() (lines 649-667)
// - showExpandEditDialog() (lines 669-694)
// - applyExpandTaskDrafts() (lines 696-721)
// - waitForExpandTaskMessages() (lines 723-735)
// - cancelExpandTask() (lines 737-742)
// - clearExpandTaskRuntimeState() (lines 744-747)
```

**Git commit:**
```bash
git add internal/ui/command_handlers.go
git commit -m "ui: Refactor command handlers to use CLI-based expansion"
```

#### 2.6 Update Progress Dialog (if needed)
**File:** `internal/ui/dialog/expansion_progress.go` (optional, reuse existing)

If `ProgressDialog` is sufficient, no new file needed. Otherwise, create specialized dialog:

```go
package dialog

// ExpansionProgressDialog wraps ProgressDialog with expansion-specific formatting
type ExpansionProgressDialog struct {
    *ProgressDialog
    scope           string
    tasksExpanded   int
    totalTasks      int
    subtasksCreated int
}

func NewExpansionProgressDialog(scope string, totalTasks int, style *DialogStyle) *ExpansionProgressDialog {
    pd := NewProgressDialog(
        "Expanding Tasks",
        fmt.Sprintf("Running task-master expand (scope: %s)", scope),
        style,
    )
    pd.SetCancelable(true)
    
    return &ExpansionProgressDialog{
        ProgressDialog: pd,
        scope:          scope,
        totalTasks:     totalTasks,
    }
}

func (d *ExpansionProgressDialog) UpdateExpansion(progress float64, stage, message string, tasksExpanded, subtasksCreated int) {
    d.SetProgress(progress)
    d.SetMessage(fmt.Sprintf("%s: %s", stage, message))
    d.SetSubtext(fmt.Sprintf("Expanded %d/%d tasks | Created %d subtasks", tasksExpanded, d.totalTasks, subtasksCreated))
    d.tasksExpanded = tasksExpanded
    d.subtasksCreated = subtasksCreated
}
```

**Git commit:**
```bash
git add internal/ui/dialog/expansion_progress.go
git commit -m "ui: Add specialized progress dialog for expansion workflow"
```

---

### Phase 3: Testing Updates

#### 3.1 Service Layer Tests
**File:** `internal/taskmaster/service_test.go`

```go
func TestExecuteExpandWithProgress_SingleTask(t *testing.T) {
    // Test single task expansion
    // Mock CLI execution
    // Verify progress callbacks
    // Verify task reload
}

func TestExecuteExpandWithProgress_AllTasks(t *testing.T) {
    // Test --all flag
    // Verify batch execution
}

func TestExecuteExpandWithProgress_WithResearch(t *testing.T) {
    // Test --research flag
    // Verify AI assistance enabled
}

func TestExecuteExpandWithProgress_Cancellation(t *testing.T) {
    // Test context cancellation
    // Verify CLI process killed
}

func TestExecuteExpandWithProgress_CLIError(t *testing.T) {
    // Test CLI command failure
    // Verify error handling
}

func TestParseExpandProgress(t *testing.T) {
    // Test progress parsing from various CLI outputs
    tests := []struct {
        input    string
        expected ExpandProgressState
    }{
        {
            input: "Expanding task 1.2...",
            expected: ExpandProgressState{
                Stage:       "Expanding",
                CurrentTask: "1.2",
            },
        },
        {
            input: "Generated 5 subtasks for task 1.2",
            expected: ExpandProgressState{
                SubtasksCreated: 5,
                CurrentTask:     "1.2",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := parseExpandProgress(tt.input)
            // Assert result matches expected
        })
    }
}
```

**Git commit:**
```bash
git add internal/taskmaster/service_test.go
git commit -m "test: Add service layer tests for CLI expansion"
```

#### 3.2 UI Workflow Tests
**File:** `internal/ui/expansion_test.go` (new file)

```go
package ui

import (
    "testing"
    "time"
    
    "github.com/adriangreen/tm-tui/internal/taskmaster"
)

func TestShowExpansionScopeDialog(t *testing.T) {
    // Test dialog creation
    // Test with/without selected task
}

func TestHandleExpansionScopeSelected(t *testing.T) {
    // Test progress dialog creation
    // Test goroutine spawning
}

func TestExpansionWorkflow_SingleTask(t *testing.T) {
    // Integration test: scope → progress → completion
}

func TestExpansionWorkflow_Cancellation(t *testing.T) {
    // Test canceling during progress
}

func TestExpansionWorkflow_Error(t *testing.T) {
    // Test error handling and display
}
```

**Git commit:**
```bash
git add internal/ui/expansion_test.go
git commit -m "test: Add UI workflow tests for expansion"
```

#### 3.3 Update Existing Tests
**File:** `internal/ui/command_handlers_test.go`

Remove tests for deprecated functions:
```go
// REMOVE these test functions:
// - TestHandleExpandTaskCommand_NoService
// - TestStartExpandWorkflow
// - TestRunExpandTask
// - TestApplyExpandTaskDrafts
```

Add new tests:
```go
func TestHandleExpandTaskCommand_ShowsScopeDialog(t *testing.T) {
    // Test that command shows scope dialog
}

func TestHandleExpandTaskCommand_NoService(t *testing.T) {
    // Test error when service unavailable
}
```

**Git commit:**
```bash
git add internal/ui/command_handlers_test.go
git commit -m "test: Update command handler tests for new workflow"
```

---

### Phase 4: CLI Integration

#### 4.1 Update CLI Wrapper
**File:** `internal/taskmaster/cli.go`

Update existing `ExpandTask()` method:
```go
// ExpandTask expands tasks via CLI (simplified wrapper)
func (s *Service) ExpandTask(scope string, taskID string, fromID, toID string, opts ExpandTaskOptions) error {
    args := []string{"expand"}
    
    switch scope {
    case "single":
        args = append(args, fmt.Sprintf("--id=%s", taskID))
    case "all":
        args = append(args, "--all")
    case "range":
        if fromID != "" {
            args = append(args, fmt.Sprintf("--from=%s", fromID))
        }
        if toID != "" {
            args = append(args, fmt.Sprintf("--to=%s", toID))
        }
    }
    
    if opts.UseAI {
        args = append(args, "--research")
    }
    if opts.Depth > 0 {
        args = append(args, fmt.Sprintf("--depth=%d", opts.Depth))
    }
    if opts.NumSubtasks > 0 {
        args = append(args, fmt.Sprintf("--num=%d", opts.NumSubtasks))
    }
    
    _, err := s.ExecuteCommand(args...)
    if err != nil {
        return err
    }
    
    // Reload tasks
    ctx := context.WithValue(context.Background(), "force", true)
    return s.LoadTasks(ctx)
}
```

**Git commit:**
```bash
git add internal/taskmaster/cli.go
git commit -m "cli: Update ExpandTask wrapper for CLI execution"
```

#### 4.2 Add CLI Error Handling
**File:** `internal/taskmaster/errors.go` (or in `service.go`)

```go
// ExpandError represents an error during task expansion
type ExpandError struct {
    Scope   string
    TaskID  string
    Stage   string
    Message string
    Err     error
}

func (e *ExpandError) Error() string {
    if e.TaskID != "" {
        return fmt.Sprintf("expansion failed for task %s at stage %s: %s", e.TaskID, e.Stage, e.Message)
    }
    return fmt.Sprintf("expansion failed (scope: %s) at stage %s: %s", e.Scope, e.Stage, e.Message)
}

func (e *ExpandError) Unwrap() error {
    return e.Err
}
```

**Git commit:**
```bash
git add internal/taskmaster/errors.go
git commit -m "cli: Add specialized error types for expansion failures"
```

---

### Phase 5: Documentation

#### 5.1 Update README.md
**File:** `README.md`

Update keyboard shortcuts section (around line 84):
```markdown
#### Task Management
- `n` - Jump to next available task
- `s` - Change task status
- `Enter` - Select item / toggle expand
- `Space` - Multi-select task for bulk operations
- `Alt+X` - Expand tasks (opens scope selection dialog)
  - Supports single task, all tasks, task range, or by tag
  - AI-powered expansion with --research flag
  - Configurable depth (1-3 levels) and subtask count
- `Alt+D` - Delete selected task (with confirmation)
- `d` - Mark task as done (quick status change)
- `p` - Change task priority
```

Update workflows section (around line 131):
```markdown
### Expanding Tasks into Subtasks
1. Select a task or prepare to expand all tasks
2. Press `Alt+X` to open the "Expand Tasks" dialog
3. Choose expansion scope:
   - **Selected task only** - Expand just the current task
   - **All tasks** - Expand all tasks in the project
   - **Task range** - Expand tasks from ID X to ID Y
   - **By tag** - Expand all tasks with specific tags
4. Configure options:
   - Expansion depth: 1-3 levels of nested subtasks
   - Number of subtasks: Leave blank for auto-detection
   - AI assistance: Enable `--research` for intelligent expansion
5. Monitor progress in real-time as CLI executes
6. Review newly created subtasks in the task tree
7. Tasks are automatically reloaded after expansion completes

**Note:** This feature executes `task-master expand` CLI commands. Ensure the Task Master CLI is properly installed and accessible.
```

**Git commit:**
```bash
git add README.md
git commit -m "docs: Update README with new expansion workflow"
```

#### 5.2 Update AGENTS.md
**File:** `AGENTS.md` (if exists) or `.taskmaster/CLAUDE.md`

Update command reference:
```markdown
## Task Expansion

### CLI Commands (Primary)
```bash
# Expand single task
task-master expand --id=1.2 [--research] [--depth=2] [--num=5]

# Expand all tasks
task-master expand --all [--research]

# Expand task range
task-master expand --from=1 --to=5 [--research]
```

### TUI Commands
- `Alt+X` - Open expansion scope dialog
  - Select scope (single/all/range/tag)
  - Configure depth and AI assistance
  - Monitor progress in real-time

### Integration Notes
- TUI now executes CLI commands instead of using local functions
- Progress is streamed from CLI stdout
- Tasks are automatically reloaded after expansion
- Supports cancellation (Ctrl+C during progress)
```

**Git commit:**
```bash
git add .taskmaster/CLAUDE.md
git commit -m "docs: Update agent documentation with CLI expansion"
```

#### 5.3 Add Migration Notes
**File:** `MIGRATION.md` (new file)

```markdown
# Migration Guide: Task Expansion Refactoring

**Date:** 2025-12-12  
**Version:** v2.0.0  
**Branch:** `refactor/task-expansion-cli-integration`

## Overview

Task expansion has been refactored to use Task Master CLI commands instead of local Go functions. This ensures consistency, proper persistence, and support for AI-powered expansion.

## Breaking Changes

### Removed Functions
The following internal functions have been removed or deprecated:

- `openTaskSelectionDialog()` - Replaced by `ExpansionScopeDialog`
- `startExpandWorkflow()` - Replaced by `handleExpansionScopeSelected()`
- `runExpandTask()` - Replaced by `ExecuteExpandWithProgress()`
- `showExpandPreviewDialog()` - Removed (CLI handles expansion directly)
- `showExpandEditDialog()` - Removed (CLI handles expansion directly)
- `applyExpandTaskDrafts()` - Replaced by CLI execution

### Deprecated Functions (Testing Only)
These functions remain but are marked deprecated:

- `ExpandTaskDrafts()` - Use CLI instead
- `ApplySubtaskDrafts()` - Use CLI instead

## New Architecture

### Before (Local)
```
User → Options Dialog → ExpandTaskDrafts() → Preview → Edit → ApplySubtaskDrafts() → Reload
```

### After (CLI)
```
User → Scope Dialog → ExecuteExpandWithProgress() → CLI Execution → Progress Updates → Auto Reload
```

## Migration Path

### For Users
No action required. The UI flow remains similar:
1. Press `Alt+X`
2. Configure options
3. Expansion happens automatically

### For Developers
If you were using the internal expansion functions:

**Old:**
```go
drafts := taskmaster.ExpandTaskDrafts(task, opts)
newIDs, err := taskmaster.ApplySubtaskDrafts(parentTask, drafts)
```

**New:**
```go
err := taskService.ExecuteExpandWithProgress(
    ctx,
    "single",  // scope
    task.ID,   // taskID
    "",        // fromID (for range)
    "",        // toID (for range)
    nil,       // tags
    opts,
    func(state taskmaster.ExpandProgressState) {
        // Handle progress updates
    },
)
```

## Testing

All existing tests have been updated or replaced. Run:

```bash
make test
```

## Rollback

If issues arise, revert to the `main` branch before merge:

```bash
git checkout main
```

## Support

For issues or questions, see:
- GitHub Issues: https://github.com/adriangreen/tm-tui/issues
- Task Master AI: https://github.com/cyanheads/task-master-ai
```

**Git commit:**
```bash
git add MIGRATION.md
git commit -m "docs: Add migration guide for expansion refactoring"
```

#### 5.4 Add Inline Documentation
**Files:** Various

Add godoc comments to new functions:
```go
// ExecuteExpandWithProgress executes the task-master expand CLI command with
// real-time progress reporting. It supports multiple expansion scopes:
//   - "single": Expand a single task by ID
//   - "all": Expand all tasks in the project
//   - "range": Expand tasks within a specified ID range
//   - "tag": Expand tasks matching specified tags
//
// The onProgress callback receives updates during execution, including stage,
// progress percentage, and current task being processed. The context can be
// used for cancellation.
//
// After successful expansion, tasks are automatically reloaded.
func (s *Service) ExecuteExpandWithProgress(
    ctx context.Context,
    scope string,
    taskID string,
    fromID string,
    toID string,
    tags []string,
    opts ExpandTaskOptions,
    onProgress func(ExpandProgressState),
) error {
    // ... implementation
}
```

**Git commit:**
```bash
git add internal/taskmaster/service.go internal/ui/expansion.go
git commit -m "docs: Add godoc comments for expansion functions"
```

---

### Phase 6: Final Preparation & PR

#### 6.1 Run All Tests
```bash
# Run unit tests
make test

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

**Git commit if fixes needed:**
```bash
git add <fixed_files>
git commit -m "test: Fix failing tests after refactoring"
```

#### 6.2 Run Linting
```bash
make lint

# Or manually
golangci-lint run ./...
```

**Git commit if fixes needed:**
```bash
git add <fixed_files>
git commit -m "style: Fix linting issues"
```

#### 6.3 Update CHANGELOG
**File:** `CHANGELOG.md`

Add entry:
```markdown
## [Unreleased]

### Changed
- **BREAKING**: Task expansion now uses Task Master CLI commands instead of local functions
- Task expansion workflow redesigned with scope selection dialog
- Expansion progress is now shown in real-time from CLI output
- Improved error handling for expansion failures

### Added
- Support for expanding all tasks at once (`task-master expand --all`)
- Support for expanding task ranges (`--from` and `--to` flags)
- Support for tag-based expansion
- Real-time progress updates during expansion
- Cancellation support during expansion
- `ExpansionScopeDialog` for configuring expansion options
- `ExecuteExpandWithProgress()` service method for CLI integration

### Removed
- Local task expansion preview dialog (CLI handles expansion directly)
- Local task expansion edit dialog (CLI handles expansion directly)
- Direct use of `ExpandTaskDrafts()` and `ApplySubtaskDrafts()` in UI layer

### Deprecated
- `ExpandTaskDrafts()` - Use CLI execution instead
- `ApplySubtaskDrafts()` - Use CLI execution instead

### Fixed
- Task expansion now properly persists changes via CLI
- Expansion with `--research` flag now works correctly
- Tasks are reliably reloaded after expansion
```

**Git commit:**
```bash
git add CHANGELOG.md
git commit -m "docs: Update CHANGELOG for expansion refactoring"
```

#### 6.4 Create PR Description
**File:** `PR_DESCRIPTION.md` (temporary, for GitHub PR)

```markdown
# Refactor: Task Expansion CLI Integration

## Summary

Refactors task expansion to use Task Master CLI commands (`task-master expand`) instead of local Go functions. This ensures proper persistence, AI-powered expansion support, and architectural consistency with other features.

## Motivation

The previous implementation:
- Bypassed the Task Master CLI entirely
- Made direct in-memory modifications to task tree
- Did not support AI-powered expansion (`--research` flag)
- Caused potential sync issues with CLI state
- Limited to single-task expansion only

## Changes

### Architecture
- Task expansion now executes `task-master expand` CLI commands
- Progress is streamed from CLI stdout in real-time
- Follows the same pattern as complexity analysis workflow
- Supports single task, all tasks, task ranges, and tag-based expansion

### User-Facing
- New scope selection dialog for choosing expansion target
- Real-time progress reporting during CLI execution
- Support for batch expansion (all tasks, ranges)
- Improved error messages from CLI
- Cancellation support (Ctrl+C during progress)

### Developer-Facing
- New `ExecuteExpandWithProgress()` service method
- New `ExpansionScopeDialog` component
- Deprecated local expansion functions
- Updated message types for expansion workflow
- Comprehensive test coverage

## Testing

- [x] Unit tests for service layer
- [x] UI workflow tests
- [x] Integration tests for CLI execution
- [x] Manual testing of all expansion scopes
- [x] Cancellation testing
- [x] Error handling testing

## Checklist

- [x] Code compiles without errors
- [x] All tests pass
- [x] Linting passes
- [x] Documentation updated (README, AGENTS.md)
- [x] CHANGELOG updated
- [x] Migration guide created
- [x] No breaking changes to public API
- [x] Backward compatibility considered

## Reviewers

- [ ] Architecture review
- [ ] Code review
- [ ] Documentation review
- [ ] Manual testing

## Related Issues

Closes #X (replace with actual issue number if exists)
Relates to task 5 in `.taskmaster/tasks/`

## Screenshots/Demo

(Add screenshots or GIF of new expansion workflow)

## Migration Notes

See `MIGRATION.md` for detailed migration guide.

Summary:
- Users: No action required, UI flow is similar
- Developers: Use `ExecuteExpandWithProgress()` instead of `ExpandTaskDrafts()`
- Local expansion functions deprecated but available for testing

## Post-Merge Tasks

- [ ] Update task-master-ai documentation (if needed)
- [ ] Announce breaking changes in release notes
- [ ] Monitor for user feedback
- [ ] Consider adding expansion report/summary feature (future enhancement)
```

#### 6.5 Final Review
```bash
# Review all commits
git log --oneline main..HEAD

# Review all changes
git diff main --stat
git diff main

# Ensure no debugging code or TODO comments
grep -r "TODO" internal/ --exclude-dir=vendor
grep -r "FIXME" internal/ --exclude-dir=vendor
grep -r "fmt.Println" internal/ --exclude-dir=vendor

# Check for formatting
gofmt -d internal/
```

#### 6.6 Push Branch and Create PR
```bash
# Push feature branch
git push -u origin refactor/task-expansion-cli-integration

# Create PR via GitHub CLI (if installed)
gh pr create \
  --title "Refactor: Task Expansion CLI Integration" \
  --body-file PR_DESCRIPTION.md \
  --base main \
  --head refactor/task-expansion-cli-integration

# Or create PR manually via GitHub web interface
```

---

## Implementation Checklist Summary

### Phase 0: Setup ✓
- [x] Create feature branch
- [x] Create PRD document
- [x] Commit PRD

### Phase 1: Service Layer
- [ ] Add `ExecuteExpandWithProgress()` method
- [ ] Update `ExpandProgressState` type
- [ ] Add progress parsing logic
- [ ] Mark deprecated functions
- [ ] Commit service layer changes

### Phase 2: UI Layer
- [ ] Create `ExpansionScopeDialog`
- [ ] Create `expansion.go` workflow file
- [ ] Update message types
- [ ] Update app state
- [ ] Refactor command handlers
- [ ] Remove deprecated UI code
- [ ] Commit UI layer changes

### Phase 3: Testing
- [ ] Add service layer tests
- [ ] Add UI workflow tests
- [ ] Update existing tests
- [ ] Verify all tests pass
- [ ] Commit test changes

### Phase 4: CLI Integration
- [ ] Update CLI wrapper methods
- [ ] Add error handling
- [ ] Verify CLI command support
- [ ] Commit CLI changes

### Phase 5: Documentation
- [ ] Update README.md
- [ ] Update AGENTS.md / CLAUDE.md
- [ ] Create MIGRATION.md
- [ ] Add inline godoc comments
- [ ] Commit documentation changes

### Phase 6: Final Preparation
- [ ] Run all tests
- [ ] Run linting
- [ ] Update CHANGELOG
- [ ] Create PR description
- [ ] Final review
- [ ] Push branch
- [ ] Create PR

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| CLI doesn't support required flags | High | Medium | Verify CLI capabilities first; add feature requests to task-master-ai repo if needed |
| Progress parsing unreliable | Medium | Medium | Add fallback to spinner without specific progress percentage; robust parsing with error handling |
| Breaking existing workflows | Medium | Low | Thorough testing; maintain backward compatibility where possible; clear migration guide |
| Performance regression | Low | Low | CLI execution is async and should be similar to current performance; monitor in testing |
| Loss of preview/edit features | Medium | Low | Document as intentional simplification; CLI is source of truth; consider post-expansion review dialog |
| Merge conflicts | Low | Low | Regular rebasing on main; communicate with team about refactoring |

---

## Timeline Estimate

| Phase | Estimated Time | Dependencies |
|-------|----------------|--------------|
| Phase 0: Setup | 0.5 hours | None |
| Phase 1: Service Layer | 3-4 hours | CLI verification |
| Phase 2: UI Layer | 5-6 hours | Phase 1 complete |
| Phase 3: Testing | 2-3 hours | Phase 1, 2 complete |
| Phase 4: CLI Integration | 2 hours | Phase 1 complete |
| Phase 5: Documentation | 1 hour | All phases |
| Phase 6: Final Prep & PR | 1-2 hours | All phases |
| **Total** | **14.5-18.5 hours** | |

---

## Success Criteria

1. ✅ `task-master expand --id=X` executes via CLI, not local Go functions
2. ✅ `task-master expand --all` supported
3. ✅ Progress dialog shows real-time CLI output
4. ✅ `--research` flag properly passed to CLI
5. ✅ Cancellation works (Ctrl+C to CLI process)
6. ✅ Tasks reload automatically after expansion
7. ✅ Error messages from CLI displayed to user
8. ✅ Pattern matches complexity analysis workflow
9. ✅ All existing tests pass (or updated appropriately)
10. ✅ README and docs updated with new workflow
11. ✅ Migration guide created for developers
12. ✅ PR created and ready for review

---

## Open Questions

1. **Does Task Master CLI support `--depth` and `--num` flags?**  
   → Need to verify in task-master-ai source code
   → **Action:** Test CLI locally before implementation

2. **What is the exact format of CLI progress output?**  
   → Need to capture stdout during expansion to design parser
   → **Action:** Run `task-master expand --id=X` and log output

3. **Should we keep preview/edit dialogs?**  
   → **Recommendation:** Remove for consistency with complexity analysis
   → CLI is authoritative and handles generation
   → **Decision:** Remove preview/edit, show completion notification

4. **How to handle partial failures in batch expansion?**  
   → If `task-master expand --all` fails on task 3 of 10, what happens?
   → **Action:** Test error scenarios and document behavior

5. **Should we add a post-expansion summary dialog?**  
   → Show stats: "Expanded 5 tasks, created 23 subtasks"
   → **Decision:** Add simple notification, can enhance later

---

## Post-Implementation Enhancements (Future)

### V2 Features
- [ ] Post-expansion summary dialog with statistics
- [ ] Dry-run mode / preview before execution
- [ ] Expansion history/audit log
- [ ] Undo/rollback support for expansion
- [ ] Custom expansion templates
- [ ] Batch operations with retry on failure
- [ ] Export expansion report

### V3 Features
- [ ] Parallel expansion for multiple tasks
- [ ] Interactive CLI mode (if CLI supports it)
- [ ] AI-powered expansion recommendations
- [ ] Expansion presets (saved configurations)
- [ ] Integration with project planning tools

---

## Summary

This refactoring transforms task expansion from a **local, in-memory operation** to a **CLI-driven, externally-executed workflow** that matches the complexity analysis pattern. It ensures consistency with Task Master's architecture, proper persistence, AI-powered expansion support, and maintainability.

### Key Architectural Shift
```
OLD: TUI → Go functions → Direct memory modification → Manual reload
NEW: TUI → CLI command → Task Master processes → Auto reload
```

This makes the TUI a **thin client** over the Task Master CLI, which is the correct architectural pattern per the project design.

---

**End of PRD**
