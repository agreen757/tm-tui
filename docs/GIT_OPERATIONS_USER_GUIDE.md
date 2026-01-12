# Git Operations User Guide

This guide explains how to use git operations in Task Master TUI, including understanding the Task Runner modal, interpreting results, and troubleshooting common issues.

## Quick Start: Using Git Operations

### Accessing Git Menu

To open the git menu:
1. In the TUI, press `g` (lowercase)
2. A menu appears with these options:
   - **Status** - See which files changed
   - **Switch Branch** - Change to a different branch
   - **Create Branch** - Create and switch to a new branch
   - **Recent Commits** - View recent commits

### Performing a Git Operation

1. **Open Git Menu** - Press `g`
2. **Select Operation** - Use arrow keys to highlight, press `Enter`
3. **Task Runner Opens** - See your command executing in real-time
4. **Monitor Progress** - Watch the output stream
5. **Wait for Completion** - Status shows "Completed" or "Failed"
6. **Close Modal** - Press `Esc` when done

## Understanding the Task Runner Modal

When you start a git operation, the Task Runner modal appears showing:

```
┌─────────────────────────────────────────────────────┐
│ Switch Branch: main                  [minimize]     │
├─────────────────────────────────────────────────────┤
│ Command: git checkout main                          │
│                                                     │
│ Output:                                             │
│ Updated 5 files                                     │
│ Switched to branch 'main'                           │
│                                                     │
│ ✓ Completed                                         │
├─────────────────────────────────────────────────────┤
│ ↑↓ scroll • M minimize • Esc close                  │
└─────────────────────────────────────────────────────┘
```

### Modal Components

**Header:**
- Shows the operation being performed
- Example: "Switch Branch: main"
- Minimize button to collapse modal while task runs

**Command:**
- Shows the exact git command being executed
- Example: "git checkout main"
- Helpful for understanding what's happening

**Output Area:**
- Streams git output in real-time
- Shows progress and results
- Scrollable if output is long

**Status Bar:**
- Shows current state: Running, Completed ✓, or Failed ✗
- Running: Operation is in progress
- Completed ✓: Operation succeeded
- Failed ✗: Operation encountered an error

**Control Hints:**
- `↑↓ scroll` - Use arrow keys to scroll output
- `M minimize` - Collapse modal but keep task running
- `Esc close` - Close modal (only available when not running)

## Keyboard Controls in Task Runner

| Control | Action |
|---------|--------|
| `↑` / `↓` | Scroll output up/down |
| `PageUp` / `PageDn` | Page scroll |
| `Home` | Jump to top of output |
| `End` | Jump to bottom of output |
| `M` | Minimize/maximize modal |
| `Ctrl+C` | Cancel operation (with confirmation) |
| `Esc` | Close modal (only when not running) |

## Common Git Operations

### Status: Check What Changed

**Purpose:** See which files you've modified, which are staged, and which are untracked.

**How to:**
1. Press `g` → Select "Status"
2. Task Runner shows:
   ```
   M  src/main.go
   M  internal/config.go
   ??  .env.local
   ```

**Understanding the Output:**
- `M` = Modified (you changed this file)
- `A` = Added (new file)
- `D` = Deleted (file removed)
- `??` = Untracked (git doesn't know about this file)
- Space before letter = unstaged (not ready to commit)
- Space replaced by letter = staged (ready to commit)

**When to Use:**
- Before creating a commit
- To see what you've worked on
- To understand repository state

### Switch Branch: Change Your Working Branch

**Purpose:** Move to a different branch to work on different features/fixes.

**How to:**
1. Press `g` → Select "Switch Branch"
2. A list appears showing all branches
3. Use arrow keys to highlight the branch you want
4. Press `Enter`
5. Task Runner shows the switch happening
6. Files update to match the selected branch

**Understanding the Output:**
```
Updated 5 files
Switched to branch 'feature/auth'
```
This means the switch was successful and 5 files were updated.

**Common Results:**
- **Success:** Branch is switched, files update
- **Error - "Your local changes to ... would be overwritten":** You have uncommitted changes. Commit or stash them first.
- **Error - "pathspec 'branch-name' did not match":** Branch doesn't exist

**When to Use:**
- Switch to work on a different feature
- Update to latest main branch
- Go back to a previous branch

### Create Branch: Make a New Feature Branch

**Purpose:** Create a new branch for working on a specific feature or fix.

**How to:**
1. Press `g` → Select "Create Branch"
2. A text input appears
3. Type your branch name (e.g., "feature/user-login")
4. Press `Enter`
5. Task Runner creates the branch and switches to it

**Branch Naming Tips:**
- Use lowercase letters, numbers, and hyphens
- Describe what the branch is for
- Examples: `feature/search`, `bugfix/login-error`, `docs/readme-update`

**Invalid Names:**
- ❌ Spaces: "my new branch"
- ❌ Special characters: "branch@1", "branch#fix"
- ❌ Empty: (no input)
- ✅ Valid: "my-new-branch", "feature/auth", "v1.0"

**Understanding the Output:**
```
Switched to a new branch 'feature/user-login'
```
This means your new branch was created successfully and you're now on it.

**When to Use:**
- Starting work on a new feature
- Creating a branch for a bug fix
- Organizing work into manageable units

### Recent Commits: View Recent Changes

**Purpose:** See what changes have been made recently, who made them, and when.

**How to:**
1. Press `g` → Select "Recent Commits"
2. A list appears showing recent commits
3. Navigate with arrow keys to see details
4. Press `Enter` to select

**Understanding the Output:**
```
abc1234 Add user authentication
def5678 Fix login page styling
ghi9012 Update documentation
```

Each line shows:
- Commit ID (first 7 characters)
- Commit message describing the change

**When to Use:**
- Review what others have done
- Find a specific change
- Understand repository history

## Reading Git Output and Errors

### Success Messages

These messages mean your operation worked:

```
✓ Completed
Updated X files
Switched to branch 'name'
Branch 'name' set up to track remote branch
```

**What to do:** Your operation succeeded! You can close the modal and continue working.

### Common Error Messages and Solutions

#### "Your local changes to ... would be overwritten by checkout"

**What it means:** You have unsaved changes that conflict with the branch you're switching to.

**Solution:**
1. Close the modal (press `Esc`)
2. Go to the Task Master TUI main view
3. Either:
   - Commit your changes: `git commit -m "Your message"`
   - Discard changes: `git checkout -- <file>` (careful!)
   - Stash changes: `git stash` (save temporarily)
4. Try switching branches again

#### "Branch already exists"

**What it means:** You tried to create a branch that already exists.

**Solution:**
1. Choose a different branch name
2. Or use "Switch Branch" to go to the existing branch

#### "fatal: Not a git repository"

**What it means:** This directory isn't a git repository, or you're in the wrong directory.

**Solution:**
1. Make sure you're in the right project directory
2. Or initialize git: `git init` (in terminal)
3. Or clone a repository first

#### "fatal: unable to read from remote repository"

**What it means:** Network connection problem or access denied.

**Solution:**
1. Check your internet connection
2. Verify you have access to the repository
3. Check SSH keys if using SSH (not needed for Task Master TUI git operations)
4. Try again later

#### "The following untracked working tree files would be overwritten"

**What it means:** You have files that git doesn't know about that would conflict.

**Solution:**
1. Move or delete those files
2. Or add them to `.gitignore` if they shouldn't be tracked
3. Try the operation again

### Warnings vs Errors

**Warnings** (operation still completes):
```
warning: LF will be replaced by CRLF in file.txt
```
These are usually about line endings and don't break anything.

**Errors** (operation fails):
```
fatal: unable to access 'https://...'
error: Your local changes to 'file.go' would be overwritten
```
These prevent the operation from completing and need fixing.

## Troubleshooting

### Problem: Task Runner Shows Empty Output

**Possible Causes:**
- Command completed very quickly
- Output is still loading
- Command succeeded with no output

**Solution:**
- Wait a moment for status to update
- Check the status bar at bottom
- If it shows "Completed ✓", the operation succeeded even with no output

### Problem: Minimize Button Doesn't Work

**Possible Causes:**
- Button might be labeled differently
- Press `M` instead of clicking

**Solution:**
- Use keyboard: Press `M` to minimize
- Look at the status bar for available commands

### Problem: Can't Close Modal

**Possible Causes:**
- Operation is still running
- Modal is minimized

**Solution:**
- Wait for status to show "Completed ✓" or "Failed ✗"
- If minimized, press `M` to maximize
- Then press `Esc` to close

### Problem: Operation Takes Too Long

**Possible Causes:**
- Large operation (many files)
- Network slowdown (if fetching from remote)
- Legitimate long-running operation

**Solution:**
- Wait longer (some operations naturally take time)
- Check output to see progress
- If truly stuck, try `Ctrl+C` to cancel

### Problem: Same Error Happens Again

**Solution:**
1. Read the error message carefully
2. Check the "Common Error Messages" section above
3. Follow the suggested solution
4. If it still fails, try in the terminal directly:
   ```bash
   git status  # Check repository state
   git log -1  # See latest commit
   ```

## Integration with Task Master Workflow

Git operations in Task Master TUI are designed to work seamlessly with task management:

### Starting a Task

1. You have a task to work on
2. Create a new branch: Press `g` → Create Branch
3. Name it after your task: `task-123-feature`
4. Work on your implementation
5. When done, switch back to main or create a pull request

### Switching Between Tasks

1. Your current task is on `task-123-feature`
2. Need to work on another task? `task-456-bugfix`
3. Commit your changes (if ready)
4. Press `g` → Switch Branch → `task-456-bugfix`
5. Task Runner switches your files
6. Continue working on new task

### Organizing Work with Branches

Each task should have its own branch:
```
main (production-ready code)
├── task-1-auth (authentication feature)
├── task-2-search (search feature)  
└── task-3-ui-fix (UI bug fix)
```

## Tips and Best Practices

### 1. Commit Before Switching Branches

Always commit your work before switching branches to avoid losing changes:

```
✓ Do this:
- Finish your work
- Commit: Your changes are saved
- Switch branches: Safe to move around

✗ Don't do this:
- Have uncommitted changes
- Switch branches: You might lose work
```

### 2. Use Descriptive Branch Names

Branch names should describe what you're working on:
- ✓ `feature/user-authentication`
- ✓ `bugfix/login-redirect`
- ✓ `docs/api-documentation`
- ✗ `branch1`
- ✗ `new-thing`

### 3. Keep Branches Current

Before starting work on a branch, switch to main and get latest:
1. Switch to main: `g` → Switch Branch → main
2. Do this regularly to avoid conflicts

### 4. Check Status Before Starting Work

Always check status before beginning work:
1. Press `g` → Status
2. Verify you're on the right branch
3. Confirm no unexpected changes exist

### 5. Understand Before Acting

Always read the full output before closing the modal:
- Check for warnings or unexpected messages
- Verify the operation did what you intended
- Look for errors that need fixing

## Getting Help

If you encounter issues:

1. **Read the error message** - Git messages are usually descriptive
2. **Check the "Common Error Messages" section** above
3. **Try in the terminal** - Run the same git command in a terminal for more details
   ```bash
   git status
   git log
   git branch
   ```
4. **Check project documentation** - Some projects have special git workflows
5. **Ask team members** - They may have encountered the same issue

## Summary

Git operations in Task Master TUI provide an easy way to manage branches and see repository status without leaving the TUI:

- Use `g` to open the git menu
- Watch Task Runner for real-time progress
- Read the output to understand what happened
- Refer to the error messages section if something fails
- Commit your work before switching branches
- Use descriptive branch names

With these tools, you can focus on your development work while the TUI handles the git management!
