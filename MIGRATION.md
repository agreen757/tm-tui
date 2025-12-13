# Migration Guide: Task Expansion Refactoring

**Date:** 2025-12-13  
**Branch:** `refactor/task-expansion-cli-integration`  
**Status:** Complete

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
No action required. The UI flow has improved:
1. Press `Alt+X`
2. Select scope (single/all/range/tag)
3. Configure options
4. Monitor real-time progress
5. Expansion happens automatically

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

## New Features

### Scope Selection
The new expansion dialog supports multiple scopes:

- **Single task** - Expand a specific task by ID
- **All tasks** - Expand all tasks in the project
- **Task range** - Expand tasks from ID X to ID Y
- **By tag** - Expand tasks with specific tags

### CLI Integration
All expansion now goes through the Task Master CLI:

```bash
# Single task
task-master expand --id=1.2 --research

# All tasks
task-master expand --all

# Task range
task-master expand --from=1 --to=5

# With custom options
task-master expand --id=1.2 --num=5 --research
```

### Progress Reporting
Real-time progress updates show:

- Current stage (Analyzing, Generating, Applying)
- Current task being processed
- Tasks expanded so far
- Total subtasks created
- Progress percentage

### Cancellation Support
Users can cancel expansion during execution:
- Press `ESC` or `C` during progress dialog
- Context cancellation properly terminates CLI process

## Testing

All existing tests have been updated or replaced. Run:

```bash
make test
```

To verify:
```bash
go build ./...
go test ./...
```

## Rollback

If issues arise, revert to the previous commit before merge:

```bash
git checkout main
```

Or cherry-pick specific fixes if needed.

## Known Limitations

1. **Tag-based expansion** - CLI may not support `--tag` flag yet (future enhancement)
2. **Depth parameter** - CLI may not support `--depth` flag yet (uses default depth)
3. **Preview/Edit** - No longer available, CLI handles generation directly

## Benefits

### Consistency
- All task operations now use CLI as source of truth
- No sync issues between TUI and CLI state

### AI Support
- `--research` flag properly passed to CLI
- Better subtask generation with online research

### Reliability
- CLI handles all edge cases
- Proper error messages from authoritative source

### Maintainability
- Less code duplication
- Single implementation to maintain
- Follows established patterns (complexity analysis)

## Support

For issues or questions, see:
- GitHub Issues: https://github.com/adriangreen/tm-tui/issues
- Task Master AI: https://github.com/cyanheads/task-master-ai

## Summary

This refactoring transforms task expansion from a **local, in-memory operation** to a **CLI-driven, externally-executed workflow** that matches the complexity analysis pattern. It ensures consistency with Task Master's architecture, proper persistence, AI-powered expansion support, and maintainability.

### Key Architectural Shift
```
OLD: TUI → Go functions → Direct memory modification → Manual reload
NEW: TUI → CLI command → Task Master processes → Auto reload
```

This makes the TUI a **thin client** over the Task Master CLI, which is the correct architectural pattern per the project design.
