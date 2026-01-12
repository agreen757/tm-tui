# Dialog Implementation Guide: Git Task Runner Integration

This guide explains how to implement dialogs that execute git commands via the Task Runner, with practical examples from the Task Master TUI codebase.

## Overview

When a dialog needs to execute a git command, it follows this pattern:

```
User interacts with dialog → Input validation → launchGit*() method → 
Task Runner starts → Git command executes → Output streams → Command completes → Dialog closes
```

## Core Pattern: launchGit* Method

Every dialog that executes a git command should have a `launchGit*()` method that:

1. Creates a TaskStartedMsg (shows Task Runner modal)
2. Calls ExecuteGitCommand (executes git with streaming)
3. Uses tea.Sequence to ensure proper ordering

### Basic Pattern Structure

```go
// launchGitOperation launches the git command via Task Runner
func (d *MyDialog) launchGitOperation(param string) tea.Cmd {
	return tea.Sequence(
		// First, send TaskStartedMsg to open the Task Runner modal
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-operation-id",
				TaskTitle: "Operation: " + param,
			}
		},
		// Then, execute the git command
		ExecuteGitCommand("git-operation-id", []string{"arg1", "arg2"}, d.tagName),
	)
}
```

## Real-World Example 1: Branch Creation Dialog

### Complete Implementation: BranchCreateDialog

The branch creation dialog demonstrates input validation and task launching.

**Step 1: Dialog Structure**

```go
type BranchCreateDialog struct {
	BaseDialog
	input    textinput.Model  // For user input
	repoPath string           // Git repository path
	tagName  string           // For log organization
	errorMsg string           // Validation errors
}

func NewBranchCreateDialog(repoPath string, tagName string) *BranchCreateDialog {
	input := textinput.New()
	input.Placeholder = "new-branch-name"
	input.Focus()

	dialog := &BranchCreateDialog{
		BaseDialog: NewBaseDialog("Create Branch", 60, 12, DialogKindForm),
		input:      input,
		repoPath:   repoPath,
		tagName:    tagName,
	}

	dialog.SetFooterHints(
		ShortcutHint{Key: "Enter", Label: "Create"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	return dialog
}
```

**Step 2: Input Validation**

Validate before launching the command:

```go
// isValidBranchName checks if the current input is a valid branch name
func (d *BranchCreateDialog) isValidBranchName() bool {
	branchName := strings.TrimSpace(d.input.Value())
	
	// Check not empty
	if branchName == "" {
		return false
	}
	
	// Check no spaces
	if strings.Contains(branchName, " ") {
		return false
	}
	
	// Check for invalid git characters
	invalidChars := []string{"\t", "\n", "\r", ":", "~", "^", "?", "*", "[", "\\"}
	for _, char := range invalidChars {
		if strings.Contains(branchName, char) {
			return false
		}
	}
	
	return true
}

// validateAndUpdateError provides real-time feedback
func (d *BranchCreateDialog) validateAndUpdateError() {
	branchName := d.input.Value()
	
	if strings.TrimSpace(branchName) == "" {
		d.errorMsg = ""
		return
	}
	
	if strings.Contains(branchName, " ") {
		d.errorMsg = "Branch names cannot contain spaces"
		return
	}
	
	d.errorMsg = ""
}
```

**Step 3: Launch Git Command**

The key method that launches the git operation:

```go
// launchGitCreateBranch launches the git branch creation via Task Runner
func (d *BranchCreateDialog) launchGitCreateBranch(branchName string) tea.Cmd {
	return tea.Sequence(
		// First, send TaskStartedMsg to open the Task Runner modal
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-create-branch",
				TaskTitle: "Create Branch: " + branchName,
			}
		},
		// Then, execute the git command
		ExecuteGitCommand("git-create-branch", []string{"checkout", "-b", branchName}, d.tagName),
	)
}
```

**Step 4: Handle Key Input**

In the HandleKey method, validate and launch:

```go
func (d *BranchCreateDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Handle base keys (ESC, etc.)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	switch msg.String() {
	case "enter":
		// Get and validate input
		branchName := strings.TrimSpace(d.input.Value())
		
		if !d.isValidBranchName() {
			// Don't proceed with invalid input
			return DialogResultNone, nil
		}
		
		// Close dialog and launch git operation
		return DialogResultClose, d.launchGitCreateBranch(branchName)
	}
	
	return DialogResultNone, nil
}
```

## Real-World Example 2: Branch Switch Dialog

### Complete Implementation: BranchSwitchDialog

The branch switch dialog shows how to handle a list selection dialog launching a git command.

**Step 1: Dialog Structure with State Management**

```go
type BranchSwitchDialog struct {
	BaseDialog
	list          bubbles.List
	repoPath      string
	currentBranch string
	tagName       string
	
	// State management with mutex for thread-safety
	mu               sync.RWMutex
	switching        bool              // Prevents duplicate selections
	currentTaskID    string            // Track active task
	selectedBranch   string            // Cache selected branch
}

func NewBranchSwitchDialog(repoPath string, callback func(string), tagName string) (*BranchSwitchDialog, error) {
	// Load branches from git
	branches, currentBranch, err := getBranches(repoPath)
	if err != nil {
		return nil, err
	}

	dialog := &BranchSwitchDialog{
		BaseDialog:     NewBaseDialog("Switch Branch", 60, 15, DialogKindList),
		repoPath:       repoPath,
		currentBranch:  currentBranch,
		tagName:        tagName,
		switching:      false,
	}

	// Populate list with branches
	var items []list.Item
	for _, branch := range branches {
		items = append(items, branchItem{title: branch})
	}
	dialog.list.SetItems(items)

	return dialog, nil
}
```

**Step 2: Launch Git Command with State Management**

```go
// launchGitSwitchBranch launches the git branch switch via Task Runner
func (d *BranchSwitchDialog) launchGitSwitchBranch(branch string) tea.Cmd {
	return tea.Sequence(
		// First, send TaskStartedMsg to open the Task Runner modal
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-switch-branch",
				TaskTitle: "Switch Branch: " + branch,
			}
		},
		// Then, execute the git command
		ExecuteGitCommand("git-switch-branch", []string{"checkout", branch}, d.tagName),
	)
}
```

**Step 3: Safe List Item Selection with Null Checks**

```go
func (d *BranchSwitchDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// First check base dialog keys (ESC for cancel)
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Check if already switching (prevent duplicate operations)
	d.mu.RLock()
	isSwitching := d.switching
	d.mu.RUnlock()

	if isSwitching {
		return DialogResultNone, nil
	}

	switch msg.String() {
	case "enter":
		// SAFE: Null check prevents panic
		item := d.list.SelectedItem()
		if item == nil {
			return DialogResultNone, nil
		}

		// SAFE: Type assertion with ok check
		branchItem, ok := item.(branchItem)
		if !ok {
			return DialogResultNone, nil
		}

		selected := branchItem.title

		// Prevent switching to current branch
		if selected == d.currentBranch {
			return DialogResultNone, nil
		}

		// Thread-safe state update
		d.mu.Lock()
		d.switching = true
		d.currentTaskID = "git-switch-branch"
		d.selectedBranch = selected
		d.mu.Unlock()

		// Close dialog and launch git switch
		return DialogResultClose, d.launchGitSwitchBranch(selected)
	}
	
	return DialogResultNone, nil
}
```

## Best Practices

### 1. Input Validation Before Launch

Always validate user input before calling launchGit*():

```go
// Good: Validate before launch
if !d.isValidBranchName() {
	return DialogResultNone, nil // Don't launch invalid command
}
return DialogResultClose, d.launchGitCreateBranch(branchName)

// Bad: No validation
return DialogResultClose, d.launchGitCreateBranch(d.input.Value())
```

### 2. Use Descriptive Task IDs

Task IDs appear in logs and UI:

```go
// Good: Descriptive IDs for easy tracking
ExecuteGitCommand("git-switch-branch", args, tagName)
ExecuteGitCommand("git-create-branch", args, tagName)

// Bad: Generic IDs are harder to debug
ExecuteGitCommand("cmd1", args, tagName)
ExecuteGitCommand("cmd2", args, tagName)
```

### 3. Prevent Duplicate Execution

Use state flags to prevent duplicate operations:

```go
// Good: Prevent duplicate selections during operation
d.mu.Lock()
d.switching = true
d.mu.Unlock()

// Then check before processing new inputs
d.mu.RLock()
isSwitching := d.switching
d.mu.RUnlock()

if isSwitching {
	return DialogResultNone, nil // Ignore input during operation
}
```

### 4. Always Include Null Checks

Defensive programming prevents panics:

```go
// Good: Null checks and type assertions with ok check
item := d.list.SelectedItem()
if item == nil {
	return DialogResultNone, nil
}

branchItem, ok := item.(branchItem)
if !ok {
	return DialogResultNone, nil
}

// Bad: Assumes item exists and is correct type
branchItem := d.list.SelectedItem().(branchItem)
```

### 5. Include Tag Name for Log Organization

Pass tagName to organize logs by project:

```go
// Good: Tag-organized logs for multi-project setups
ExecuteGitCommand("git-switch-branch", args, d.tagName)
// Logs to: .taskmaster/<tagName>/git-command-git-switch-branch-*.log

// Without tag (less organized)
ExecuteGitCommand("git-switch-branch", args, "")
// Logs to: .taskmaster/logs/git-command-git-switch-branch-*.log
```

### 6. Real-Time Feedback in Update Method

Show validation errors in real-time:

```go
func (d *BranchCreateDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
		return d, nil
	}

	// Handle text input
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	
	// Update validation feedback on every key press
	if _, ok := msg.(tea.KeyMsg); ok {
		d.validateAndUpdateError()
	}
	
	return d, cmd
}
```

## Error Handling Patterns

### Pattern: Handle Git Validation Error

```go
if err := dialog.ValidateGitBinary(); err != nil {
	if gitErr, ok := err.(*dialog.GitBinaryError); ok {
		// Show git setup error
		errDialog := dialog.NewErrorDialog(
			"Git Not Available",
			gitErr.Message,
		)
		m.appState.PushDialog(errDialog)
		return
	}
}
```

### Pattern: Handle Command Execution Error

```go
if err != nil {
	// Command failed to start
	m.addLogLine("Failed to start git command: " + err.Error())
	return err
}

// Command started, messages will come through channel
// Handle TaskFailedMsg in Update() method
```

### Pattern: Handle Command Failure Messages

In your model's Update method, handle TaskFailedMsg from git operations:

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dialog.TaskFailedMsg:
		if strings.HasPrefix(msg.TaskID, "git-") {
			// Git operation failed
			m.addLogLine("Git operation failed: " + msg.Error)
			
			// Show user-friendly error
			errDialog := dialog.NewErrorDialog(
				"Git Operation Failed",
				msg.Message,
			)
			m.appState.PushDialog(errDialog)
		}
	}
	return m, nil
}
```

## Testing Git Dialog Implementations

### Test Structure

```go
func TestBranchCreateDialog_ValidateBranchName(t *testing.T) {
	dialog := NewBranchCreateDialog("", "")
	
	tests := []struct {
		name     string
		input    string
		valid    bool
	}{
		{"valid simple", "feature/new", true},
		{"empty", "", false},
		{"has space", "my branch", false},
		{"has colon", "my:branch", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.input.SetValue(tt.input)
			if dialog.isValidBranchName() != tt.valid {
				t.Errorf("isValidBranchName(%q) = %v, want %v", 
					tt.input, dialog.isValidBranchName(), tt.valid)
			}
		})
	}
}
```

### Test Command Launching

```go
func TestBranchCreateDialog_LaunchCommand(t *testing.T) {
	dialog := NewBranchCreateDialog("", "test-tag")
	cmd := dialog.launchGitCreateBranch("test-branch")
	
	// Command should not be nil
	if cmd == nil {
		t.Fatal("launchGitCreateBranch returned nil command")
	}
	
	// Execute command and verify it produces messages
	msg := cmd()
	if msg == nil {
		t.Fatal("command execution returned nil")
	}
	
	// Should be TaskStartedMsg
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Errorf("expected TaskStartedMsg, got %T", msg)
	}
}
```

## Common Issues and Solutions

### Issue: Dialog opens but nothing happens on Enter

**Cause**: Missing HandleKey implementation or not returning DialogResultClose

**Solution**: 
```go
// Check that HandleKey is implemented
func (d *MyDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    // ... validation ...
    return DialogResultClose, d.launchGit*()  // Must return DialogResultClose
}
```

### Issue: Command executes but Task Runner doesn't appear

**Cause**: TaskStartedMsg not being sent before ExecuteGitCommand

**Solution**: Use tea.Sequence to ensure proper ordering
```go
return tea.Sequence(
    func() tea.Msg { return TaskStartedMsg{...} },
    ExecuteGitCommand(...),
)
```

### Issue: Duplicate commands executing

**Cause**: No state management to prevent double selection

**Solution**: Use mutex-protected state flag
```go
d.mu.Lock()
d.switching = true
d.mu.Unlock()

// Check before processing new input
d.mu.RLock()
if d.switching { return DialogResultNone, nil }
d.mu.RUnlock()
```

### Issue: Panic on nil selection

**Cause**: Not checking if selected item is nil before type assertion

**Solution**: Always add null checks
```go
item := d.list.SelectedItem()
if item == nil {
    return DialogResultNone, nil
}

branchItem, ok := item.(branchItem)
if !ok {
    return DialogResultNone, nil
}
```

## Summary

When implementing a dialog that launches git commands:

1. **Structure**: Create a dialog with BaseDialog
2. **Validation**: Implement validation before launch
3. **Launch Method**: Create launchGit*() with tea.Sequence
4. **Key Handling**: Return DialogResultClose and call launchGit*()
5. **State Management**: Use mutexes for thread-safety
6. **Safety**: Add null checks and type assertions
7. **Feedback**: Show real-time validation
8. **Testing**: Test validation and command launching
9. **Error Handling**: Handle both validation and execution errors
10. **Logging**: Include tagName for organized logs
