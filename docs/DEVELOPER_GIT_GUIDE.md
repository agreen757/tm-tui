# Developer Documentation: Adding New Git Operations

This guide explains how to add new git operations to Task Master TUI, including creating dialogs, integrating with the message system, testing, and following established error handling patterns.

## Overview: Anatomy of a Git Operation

Every git operation follows this flow:

1. **Dialog Creation** - User-facing dialog for input
2. **Validation** - Validate user input before execution
3. **Launch Method** - `launchGit*()` method using tea.Sequence
4. **Message Routing** - Handle TaskStartedMsg and ExecuteGitCommand
5. **Output Streaming** - Task Runner displays real-time output
6. **Completion** - TaskCompletedMsg or TaskFailedMsg handling

## Step 1: Create the Dialog Class

### Dialog Structure Template

```go
package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
)

// NewGitOperationDialog creates a new git operation dialog
type MyGitOperationDialog struct {
	BaseDialog
	// Add fields specific to your dialog
	repoPath string
	tagName  string
	// Example field: input for text input operations
	// input    textinput.Model
}

// NewMyGitOperationDialog creates a new instance
func NewMyGitOperationDialog(repoPath string, tagName string) *MyGitOperationDialog {
	dialog := &MyGitOperationDialog{
		BaseDialog: NewBaseDialog("Operation Title", width, height, DialogKindForm),
		repoPath:   repoPath,
		tagName:    tagName,
	}

	dialog.SetFooterHints(
		ShortcutHint{Key: "Enter", Label: "Execute"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	return dialog
}

// Init initializes the dialog
func (d *MyGitOperationDialog) Init() tea.Cmd {
	// Return any initialization command
	return nil
}

// Update processes messages
func (d *MyGitOperationDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
		return d, nil
	}
	return d, nil
}

// HandleKey processes keyboard input
func (d *MyGitOperationDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Handle base keys
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}

	// Handle your specific keys
	switch msg.String() {
	case "enter":
		// Validate input
		// Launch git operation
		return DialogResultClose, d.launchMyGitOperation()
	}

	return DialogResultNone, nil
}

// View renders the dialog
func (d *MyGitOperationDialog) View() string {
	// Return rendered content
	return d.RenderBorder("")
}
```

## Step 2: Implement Input Validation

### Validation Methods

Create validation methods specific to your dialog:

```go
// Validate your operation parameters
func (d *MyGitOperationDialog) isValid() bool {
	// Add your validation logic
	return true
}

// Update validation messages in real-time
func (d *MyGitOperationDialog) updateValidationMessage() {
	// Provide real-time feedback to user
}
```

### Validation Best Practices

1. **Validate Before Launch** - Never launch a command with invalid input
2. **Real-Time Feedback** - Show validation errors as user types
3. **Clear Messages** - Explain what's invalid and why
4. **Safe Type Checking** - Always use type assertions with ok check

```go
// Example: Validate branch name
func (d *MyGitOperationDialog) isValidBranchName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	if strings.ContainsAny(name, " \t:~^?*[\\") {
		return false
	}
	return true
}
```

## Step 3: Implement the Launch Method

### Launch Method Template

```go
// launchMyGitOperation launches the git operation via Task Runner
func (d *MyGitOperationDialog) launchMyGitOperation() tea.Cmd {
	return tea.Sequence(
		// First, send TaskStartedMsg to open the Task Runner modal
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-my-operation",
				TaskTitle: "My Operation Description",
			}
		},
		// Then, execute the git command
		ExecuteGitCommand(
			"git-my-operation",
			[]string{"git", "subcommand", "args"},
			d.tagName,
		),
	)
}
```

### Launch Method Guidelines

1. **Use tea.Sequence** - Ensures TaskStartedMsg is sent first
2. **Descriptive Task ID** - Use `git-*` prefix for easy identification
3. **Clear Task Title** - Shows in Task Runner header
4. **Include Tag Name** - For log organization
5. **Proper Arguments** - Pass all necessary git arguments

### Example: Different Operation Types

**Simple Command (No Additional Parameters):**
```go
func (d *GitStatusDialog) launchStatus() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-status",
				TaskTitle: "Check Repository Status",
			}
		},
		ExecuteGitCommand("git-status", []string{"status", "--porcelain"}, d.tagName),
	)
}
```

**Command with Parameters from User Input:**
```go
func (d *MyGitOperationDialog) launchWithParam(param string) tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-operation",
				TaskTitle: "Operation: " + param,
			}
		},
		ExecuteGitCommand("git-operation", []string{"subcommand", param}, d.tagName),
	)
}
```

## Step 4: Integrate with Message Handlers

### Adding Dialog Creation Handlers in app.go

Add the dialog to the git menu in `app.go`:

```go
// In app.go, add a method to open your dialog
func (m *Model) openMyGitOperationDialog() {
	if !m.gitAvailable || !m.gitRepoInfo.IsRepo {
		m.addLogLine("Error: Not in a Git repository")
		return
	}

	m.addLogLine("Starting my git operation...")

	// Get tag name for logging
	tagName := ""
	if m.activeProject != nil {
		tagName = m.activeProject.PrimaryTag()
	}
	if tagName == "" && m.taskService != nil {
		tagName = m.taskService.GetActiveTag()
	}

	// Create and show dialog
	dialog := NewMyGitOperationDialog(m.gitRepoInfo.RootPath, tagName)
	m.appState.PushDialog(dialog)
}
```

### Adding Dialog to Git Menu

In `git_menu.go`, add your operation to the git menu list:

```go
// In the git menu, add your operation
type gitMenuItem struct {
	name        string
	description string
	handler     func(*Model) tea.Cmd
}

var gitMenuItems = []gitMenuItem{
	{
		name:        "Status",
		description: "View repository status",
		handler: func(m *Model) tea.Cmd {
			m.openGitStatusDialog()
			return nil
		},
	},
	// Add your operation here:
	{
		name:        "My Operation",
		description: "Description of what it does",
		handler: func(m *Model) tea.Cmd {
			m.openMyGitOperationDialog()
			return nil
		},
	},
}
```

### Routing Git Menu Selection

Handle menu selection in `app.go`:

```go
func (m *Model) handleGitMenuSelection(selection string) {
	switch selection {
	case "status":
		m.openGitStatusDialog()
	case "switch-branch":
		m.openBranchSwitchDialog()
	case "create-branch":
		m.openBranchCreateDialog()
	case "my-operation":  // Add your operation
		m.openMyGitOperationDialog()
	}
}
```

## Step 5: Error Handling Patterns

### Error Handling in Launch Method

Always follow this pattern for robustness:

```go
func (d *MyGitOperationDialog) launchMyGitOperation() tea.Cmd {
	// Validate git is available (if not already checked)
	if err := ValidateGitBinary(); err != nil {
		// Return error message
		return func() tea.Msg {
			return TaskFailedMsg{
				TaskID:  "git-my-operation",
				Error:   err.Error(),
				Message: "Git is not available",
			}
		}
	}

	// Validate input
	if !d.isValid() {
		return func() tea.Msg {
			return TaskFailedMsg{
				TaskID:  "git-my-operation",
				Error:   "Invalid input",
				Message: "Please check your input and try again",
			}
		}
	}

	// Safe to launch
	return tea.Sequence(
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-my-operation",
				TaskTitle: "My Operation",
			}
		},
		ExecuteGitCommand("git-my-operation", args, d.tagName),
	)
}
```

### Handling TaskFailedMsg in app.go

In your model's Update method, handle failures:

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dialog.TaskFailedMsg:
		if strings.HasPrefix(msg.TaskID, "git-") {
			// Log the error
			m.addLogLine("Git operation failed: " + msg.Error)
			
			// Show error dialog if needed
			if msg.Message != "" {
				errDialog := dialog.NewErrorDialog(
					"Git Operation Failed",
					msg.Message,
				)
				m.appState.PushDialog(errDialog)
			}
			
			// Optionally refresh repository info
			if err := m.refreshGitInfo(); err != nil {
				m.addLogLine("Failed to refresh git info: " + err.Error())
			}
		}
	}
	return m, nil
}
```

### Error Categories and Handling

1. **Setup Errors** (git not found, not in repo)
   ```go
   return TaskFailedMsg{
       TaskID:  "git-my-operation",
       Error:   "Git binary not found",
       Message: "Please install git before continuing",
   }
   ```

2. **Input Validation Errors** (invalid parameters)
   ```go
   return TaskFailedMsg{
       TaskID:  "git-my-operation",
       Error:   "Invalid branch name",
       Message: "Branch names cannot contain spaces",
   }
   ```

3. **Execution Errors** (git command fails)
   - Handled by ExecuteGitCommand automatically
   - Output displayed in Task Runner
   - User can see the actual git error message

## Step 6: Testing Git Operations

### Test Setup with Temporary Repository

Use temporary git repositories for testing:

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) string {
	tmpDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}
	
	// Configure minimal git config for testing
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git: %v", err)
	}
	
	return tmpDir
}

// cleanupTestRepo removes the temporary repository
func cleanupTestRepo(t *testing.T, path string) {
	if err := os.RemoveAll(path); err != nil {
		t.Logf("Failed to cleanup test repo: %v", err)
	}
}
```

### Unit Test Template

```go
func TestMyGitOperation_ValidInput(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer cleanupTestRepo(t, repoPath)
	
	dialog := NewMyGitOperationDialog(repoPath, "test-tag")
	
	// Set up test conditions
	// Test that validation passes
	if !dialog.isValid() {
		t.Error("isValid() should return true for valid input")
	}
	
	// Test that launch method returns a command
	cmd := dialog.launchMyGitOperation()
	if cmd == nil {
		t.Fatal("launchMyGitOperation should return a command")
	}
}

func TestMyGitOperation_InvalidInput(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer cleanupTestRepo(t, repoPath)
	
	dialog := NewMyGitOperationDialog(repoPath, "test-tag")
	
	// Set invalid input
	// Test that validation fails
	if dialog.isValid() {
		t.Error("isValid() should return false for invalid input")
	}
}
```

### Integration Test Template

```go
func TestMyGitOperation_EndToEnd(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer cleanupTestRepo(t, repoPath)
	
	dialog := NewMyGitOperationDialog(repoPath, "test-tag")
	cmd := dialog.launchMyGitOperation()
	
	// Execute the command
	msg := cmd()
	
	// Verify TaskStartedMsg
	if started, ok := msg.(TaskStartedMsg); !ok {
		t.Errorf("Expected TaskStartedMsg, got %T", msg)
	} else if started.TaskID != "git-my-operation" {
		t.Errorf("Expected TaskID 'git-my-operation', got %s", started.TaskID)
	}
}
```

### Testing Error Scenarios

```go
func TestMyGitOperation_InvalidGitRepo(t *testing.T) {
	// Use non-repo directory
	dialog := NewMyGitOperationDialog("/tmp", "test-tag")
	
	cmd := dialog.launchMyGitOperation()
	msg := cmd()
	
	// Should fail with proper error
	if _, ok := msg.(TaskFailedMsg); !ok {
		t.Errorf("Expected TaskFailedMsg for non-repo directory")
	}
}
```

## Step 7: Consistency Requirements

### Naming Conventions

1. **Dialog Class Names**
   ```go
   MyGitOperationDialog  // PascalCase with Dialog suffix
   ```

2. **Method Names**
   ```go
   launchMyGitOperation()  // camelCase with launch prefix
   isValid()              // validation methods use is/has prefix
   ```

3. **Task IDs**
   ```go
   "git-my-operation"     // lowercase with hyphens
   ```

4. **Menu Items**
   ```go
   "my-operation"         // lowercase with hyphens in git menu
   ```

### Code Style Standards

1. **Comments**
   ```go
   // MyGitOperationDialog allows performing my git operation
   type MyGitOperationDialog struct {
   ```

2. **Error Messages**
   - User-friendly in TaskFailedMsg.Message
   - Detailed in TaskFailedMsg.Error
   - Include actionable suggestions

3. **Logging**
   - Use `m.addLogLine()` for user-visible messages
   - Include tag name in all git operations
   - Log start and end of operations

### Dialog Standards

All git dialogs should:

1. ✓ Extend BaseDialog
2. ✓ Include footer hints for keyboard shortcuts
3. ✓ Implement Init(), Update(), HandleKey(), View()
4. ✓ Use tea.WindowSizeMsg to center on screen
5. ✓ Validate input before launch
6. ✓ Use tea.Sequence in launch methods
7. ✓ Include repository and tag name in constructor
8. ✓ Handle nil and type assertion edge cases

## Complete Example: Adding a Log Viewer Operation

Here's a complete working example of adding a new git operation:

### Step 1: Create Dialog

```go
// MyGitLogDialog shows recent commits
type MyGitLogDialog struct {
	BaseDialog
	repoPath string
	tagName  string
	count    int
}

func NewMyGitLogDialog(repoPath string, tagName string) *MyGitLogDialog {
	dialog := &MyGitLogDialog{
		BaseDialog: NewBaseDialog("View Git Log", 70, 15, DialogKindForm),
		repoPath:   repoPath,
		tagName:    tagName,
		count:      10, // Default: show last 10 commits
	}
	dialog.SetFooterHints(
		ShortcutHint{Key: "Enter", Label: "View"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)
	return dialog
}

func (d *MyGitLogDialog) Init() tea.Cmd {
	return nil
}

func (d *MyGitLogDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.Center(msg.Width, msg.Height)
	}
	return d, nil
}

func (d *MyGitLogDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	result, cmd := d.HandleBaseKey(msg)
	if result != DialogResultNone {
		return result, cmd
	}
	
	if msg.String() == "enter" {
		return DialogResultClose, d.launchViewLog()
	}
	return DialogResultNone, nil
}

func (d *MyGitLogDialog) View() string {
	return d.RenderBorder("View recent commits (last " + string(rune(d.count)) + ")")
}

func (d *MyGitLogDialog) launchViewLog() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return TaskStartedMsg{
				TaskID:    "git-log-view",
				TaskTitle: "View Git Log",
			}
		},
		ExecuteGitCommand(
			"git-log-view",
			[]string{"log", "--oneline", "-n", "10"},
			d.tagName,
		),
	)
}
```

### Step 2: Add to app.go

```go
func (m *Model) openMyGitLogDialog() {
	if !m.gitAvailable || !m.gitRepoInfo.IsRepo {
		m.addLogLine("Error: Not in a Git repository")
		return
	}
	
	dialog := NewMyGitLogDialog(m.gitRepoInfo.RootPath, "")
	m.appState.PushDialog(dialog)
}
```

### Step 3: Add to Git Menu

Update git menu to include the new operation:
```go
{
	name:        "View Log",
	description: "View recent commits",
	handler: func(m *Model) tea.Cmd {
		m.openMyGitLogDialog()
		return nil
	},
}
```

## Checklist for Adding a New Git Operation

- [ ] Create dialog class extending BaseDialog
- [ ] Implement Init(), Update(), HandleKey(), View() methods
- [ ] Add input validation methods
- [ ] Create launchGit*() method using tea.Sequence
- [ ] Add openMyGitOperationDialog() in app.go
- [ ] Add dialog creation to git menu
- [ ] Handle TaskStartedMsg and TaskFailedMsg
- [ ] Update git command routing in app.go
- [ ] Add unit tests with temporary git repo
- [ ] Add integration test for end-to-end flow
- [ ] Add error scenario tests
- [ ] Follow naming conventions
- [ ] Add footer hints for keyboard shortcuts
- [ ] Include tag name for log organization
- [ ] Document public methods with comments
- [ ] Test with real git repository

## Reference: Git Commands Available

Common git commands to use in your operations:

```
status              - Check repository status
branch              - List or manage branches
checkout            - Switch branches or restore files
log                 - View commit history
commit              - Record changes
add                 - Stage changes
diff                - Show changes
clone               - Clone repository
pull                - Fetch and merge
push                - Upload changes
merge               - Merge branches
rebase              - Rebase branches
stash               - Save temporary changes
tag                 - Create version tags
```

## Troubleshooting Development Issues

### Issue: Dialog doesn't appear

**Cause**: Dialog not added to git menu or handler not called

**Solution**:
1. Verify handler in git menu
2. Check `handleGitMenuSelection` in app.go
3. Ensure `openMyGitOperationDialog()` calls `PushDialog`

### Issue: Command doesn't execute

**Cause**: Missing TaskStartedMsg or command not properly returned

**Solution**:
1. Use tea.Sequence to ensure order
2. Verify ExecuteGitCommand is called second
3. Check command arguments are correct

### Issue: Test repo doesn't work

**Cause**: Git not initialized or configured

**Solution**:
```go
cmd = exec.Command("git", "init")
cmd = exec.Command("git", "config", "user.email", "test@test.com")
cmd = exec.Command("git", "config", "user.name", "Test User")
```

## Summary

Adding a new git operation involves:

1. Creating a dialog class with proper structure
2. Implementing input validation
3. Creating a `launchGit*()` method with tea.Sequence
4. Integrating with app.go message handlers
5. Adding to git menu
6. Implementing proper error handling
7. Writing comprehensive tests
8. Following established patterns and naming conventions
9. Testing end-to-end functionality
10. Ensuring consistency with existing operations

Follow these guidelines and patterns, and you can quickly add new git capabilities to Task Master TUI!
