# Task Runner Auto-Close Feature

## Overview

The Task Runner modal now automatically closes when all tasks have completed, failed, or been cancelled. This prevents the modal from lingering after tasks finish and provides a smoother user experience.

## Features

### Auto-Close on Task Completion
- **Default Behavior**: Modal automatically closes 3 seconds after all tasks finish
- **Visual Countdown**: Footer displays remaining time (e.g., "Esc: close (auto in 3s)")
- **Cancellable**: User can press Esc to close immediately before timer expires
- **Smart Detection**: Only triggers when ALL tasks are no longer running

### Configuration Options

The auto-close behavior can be customized:

```go
// In task_runner_modal.go
modal := NewTaskRunnerModal(width, height, style)

// Disable auto-close (tasks remain visible until manually closed)
modal.SetAutoCloseOnFailure(false)

// Change auto-close delay (default: 3 seconds)
modal.SetAutoCloseDelay(5 * time.Second)
```

## User Experience

### Before Auto-Close
```
┌─ Task Runner ─────────────────────────────────┐
│  Task 10  Task 5                              │
│                                                │
│  [OUT] ✓ Task completed successfully          │
│  [ERR] ERROR: remote error: tls: bad record MAC│
│                                                │
│  Tab/Shift+Tab: switch  ↑/↓: scroll           │
│  M: minimize  Ctrl+C: cancel  Esc: (running)  │
└────────────────────────────────────────────────┘
```

### After All Tasks Complete
```
┌─ Task Runner ─────────────────────────────────┐
│  ✓ Task 10  ✗ Task 5                          │
│                                                │
│  [OUT] ✓ Task completed successfully          │
│  [ERR] ERROR: Agent processing failed         │
│                                                │
│  Tab/Shift+Tab: switch  ↑/↓: scroll           │
│  M: minimize  Esc: close (auto in 3s)         │
└────────────────────────────────────────────────┘
```

After 3 seconds: Modal closes automatically, user returns to main TUI.

## Implementation Details

### Architecture

1. **Close Timer**: When all tasks finish, a `closeTimer` is set to `time.Now() + autoCloseDelay`
2. **Tick Messages**: UI continues receiving tick messages to update countdown
3. **Auto-Close Message**: After delay, `TaskRunnerAutoCloseMsg` is sent
4. **Modal Closure**: Modal processes the message and closes itself if no tasks are running

### Message Flow

```
Task Completes/Fails
       ↓
checkAutoClose()
       ↓
Set closeTimer = now + 3s
       ↓
Schedule TaskRunnerAutoCloseMsg
       ↓
Continue TickMsg updates (for countdown display)
       ↓
[3 seconds later]
       ↓
TaskRunnerAutoCloseMsg received
       ↓
Modal checks: no running tasks? close timer expired?
       ↓
Modal returns nil (closes itself)
       ↓
App hides modal, returns to main UI
```

### Cancellation of Auto-Close

Auto-close is cancelled when:
- User starts a new task
- User presses Esc before timer expires
- User calls `SetAutoCloseOnFailure(false)`

## Benefits

### For Users
- **No Manual Cleanup**: Don't need to remember to close the modal
- **Stay Informed**: Can still see final output for 3 seconds
- **Quick Override**: Press Esc to close immediately if desired
- **Visual Feedback**: Countdown shows exactly when modal will close

### For Development
- **Prevents Orphaned Modals**: Ensures UI state stays clean
- **Reduces Cognitive Load**: One less thing for users to manage
- **Improved Flow**: Seamless transition back to task list

## Edge Cases Handled

1. **New Task During Countdown**: Timer is cancelled, modal stays open
2. **User Closes Manually**: Timer is ignored, modal closes immediately
3. **Multiple Failed Tasks**: Timer starts only after ALL tasks finish
4. **Minimized Modal**: Timer still runs, countdown shows when maximized
5. **Slow UI Updates**: Non-blocking timer messages don't hang UI

## Testing

### Manual Testing
1. Run a task that completes quickly (success)
2. Observe 3-second countdown in footer
3. Modal closes automatically

### Failed Task Testing
1. Run a task that fails (e.g., TLS error)
2. Observe countdown: "Esc: close (auto in 3s)"
3. Modal closes after 3 seconds

### Cancellation Testing
1. Run a task that completes
2. During countdown, press Esc
3. Modal closes immediately (countdown bypassed)

### Multiple Tasks Testing
1. Run 3 tasks simultaneously
2. Wait for 2 to complete
3. Countdown should NOT start (1 still running)
4. Wait for last task to complete
5. NOW countdown starts

## Configuration Examples

### Disable Auto-Close
```go
// Keep modal open until user manually closes
modal.SetAutoCloseOnFailure(false)
```

### Longer Delay for Review
```go
// Give user 10 seconds to review output
modal.SetAutoCloseDelay(10 * time.Second)
```

### Instant Close (No Delay)
```go
// Close immediately when tasks finish
modal.SetAutoCloseDelay(0)
```

## Future Enhancements

Potential improvements:
1. **Per-Task Configuration**: Different delays for success vs failure
2. **User Preference**: Store delay in config file
3. **Smart Delay**: Longer delay for failed tasks (to review errors)
4. **Audio/Visual Cue**: Optional notification when auto-closing
5. **Keep-Alive Key**: Hold a key to prevent auto-close

## Related Documentation

- [CRUSH_RUNNER_ARCHITECTURE.md](./CRUSH_RUNNER_ARCHITECTURE.md) - Overall architecture
- Task Runner Modal implementation: `internal/ui/dialog/task_runner_modal.go`
- App-level integration: `internal/ui/app.go` (TaskRunnerAutoCloseMsg handler)
