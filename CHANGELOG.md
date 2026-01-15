# Changelog

All notable changes to Task Master TUI are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.22] - 2025-01-15

### Task Update Functionality

**In-Place Task Updates** - Add notes, progress updates, and implementation details directly to tasks without leaving the TUI.

#### Added

- **Update Task Dialog** (`Ctrl+U`): Interactive form for adding updates to tasks
  - Multi-line textarea for detailed notes
  - Support for implementation notes, progress logs, and findings
  - Integration with Task Master CLI `update-task` and `update-subtask` commands
  - Real-time validation and feedback
  - Works with both parent tasks and subtasks
- **Task Master Executor Integration**: New unified interface for executing Task Master commands
  - `UpdateTask()` and `UpdateSubtask()` methods for command execution
  - Proper error handling and output capture
  - Stream output to Task Runner modal for visual feedback
- **Command Handler Enhancement**: Improved command routing for task updates
  - Automatic detection of task vs. subtask based on ID format
  - Proper prompt formatting with newlines and special characters
  - Integration with existing dialog and modal systems
- **Comprehensive Test Coverage**: 1,700+ new test lines across components
  - Update task dialog tests with edge cases
  - Task Master executor tests for command execution
  - Command handler tests for routing and formatting
  - Integration tests for full workflow

#### Changed

- **internal/ui/dialog/update_task.go** (79 lines): New update task dialog with textarea support
- **internal/ui/dialog/update_task_test.go** (371 lines): Complete test suite for update dialog
- **internal/executor/taskmaster.go** (73 lines): Unified Task Master command executor interface
- **internal/executor/taskmaster_test.go** (511 lines): Comprehensive executor tests
- **internal/ui/command_handlers.go** (232 new lines): Enhanced command routing with update support
- **internal/ui/command_handlers_test.go** (721 lines): Command handler tests
- **internal/ui/keymap.go**: Added `Ctrl+U` keybinding for update task
- **internal/ui/helpview.go**: Updated help text to include update task shortcut
- **internal/ui/app.go** (85 new lines): App integration for update task flow

#### User Experience

- Quick access to update any task with `Ctrl+U`
- Large textarea for detailed notes and implementation details
- Real-time execution feedback in Task Runner modal
- Clear status indicators (✓ for success, ✗ for errors)
- Seamless integration with existing task management workflow
- Output logged automatically for audit trails

## [v0.1.21] - 2025-01-14

### Command Runner Modal Fix

**Fixed Double Modal Overlap** - Resolved issue where pressing CTRL+B to run commands caused two modals to appear simultaneously.

#### Fixed

- **Command Runner Dialog Overlap**: Pressing CTRL+B previously showed both the Command Runner form dialog and the Task Runner Modal at the same time, causing visual confusion
- **Root Cause**: Task Runner Modal was incorrectly being added to the dialog stack while the Command Runner Dialog was still active
- **Solution**: Task Runner Modal is now managed separately via `m.taskRunnerVisible` flag, not through the dialog stack

#### Changed

- **internal/ui/command_handlers.go**: Removed `PushDialog()` calls for Task Runner Modal
  - `handleCommandRunnerSubmission()`: Task Runner created without adding to dialog stack
  - `ensureTaskRunnerModal()`: Updated for consistency
- Clean separation between modal types (Command Runner Dialog vs Task Runner Modal)
- Aligns with existing architecture where Task Runner is rendered independently

#### Technical Details

- Task Runner Modal already managed via `m.taskRunnerVisible` flag in `app.go`
- No breaking changes to API or test expectations
- Minimal code change (2 PushDialog calls removed)

## [v0.1.20] - 2025-01-13

### Tag Management Enhancement & Keybinding Fixes

**Comprehensive Tag Management System** - Complete overhaul of tag management with new dialogs, improved workflows, and extensive test coverage.

#### Added

- **Tag Selector Dialog**: New multi-select and single-select tag picker with search and filtering
  - Clean, intuitive interface for browsing and selecting tags
  - Support for creating new tags directly from the selector
  - Real-time filtering and keyboard navigation
- **Tag Editor Flow**: State machine-based workflow for managing tag creation and selection
  - Handles complex multi-step tag operations
  - Proper state transitions between selection and creation
  - Integrated refresh after tag creation
- **Enhanced Tag Commands**: Improved command routing and dialog integration
  - `Ctrl+T`: Opens tag context manager (fixed routing issue)
  - `Ctrl+Shift+A`: Add or manage tags (quick access)
  - `Ctrl+Shift+M`: Manage tag contexts (backward compatibility)
- **Comprehensive Test Coverage**: 2,800+ new test lines across tag management components
  - Tag selector tests with edge cases and error handling
  - Tag editor flow tests for state transitions
  - Form button interaction tests
  - Expansion scope dialog tests

#### Changed

- **internal/ui/dialog/tag_selector.go** (420 lines): New tag selection dialog with multi-select support
- **internal/ui/dialog/tag_selector_test.go** (1,401 lines): Comprehensive test suite for tag selector
- **internal/ui/dialog/tag_editor_flow.go** (262 lines): State machine for tag management workflows
- **internal/ui/dialog/tag_editor_flow_test.go** (1,463 lines): Complete flow testing with mock services
- **internal/ui/dialog/form_button_test.go** (304 lines): Button interaction and keyboard tests
- **internal/ui/dialog/expansion_scope_test.go** (148 lines): Expansion dialog tests
- **internal/ui/command_handlers.go**: Fixed Ctrl+T routing to tag management dialog
- **internal/ui/keymap.go**: Updated key bindings for tag management consistency
- **internal/ui/tag_helpers.go**: Enhanced helper functions for tag operations

#### Fixed

- Fixed Ctrl+T keybinding routing to open tag management dialog instead of tag switcher
- Improved tag context retrieval for complexity analysis
- Enhanced error handling in tag operations

## [v0.1.19] - 2025-01-12

### Git Operations Integration & Dialog Enhancements

**Comprehensive Git Management** - Full Git operations integration directly within the Task Master TUI with real-time streaming output and robust error handling.

#### Added

- **Git Menu** (`g` key): Access complete git operations without leaving the TUI
  - View repository status (staged/unstaged changes, untracked files)
  - Switch between branches with visual indicators
  - Create new branches with validation
  - View commit history with author and date information
- **Real-time Output Streaming**: Live streaming of git command output to Task Runner modal
- **Comprehensive Error Handling**: User-friendly error messages with helpful remediation suggestions
- **Log Rotation**: Automatic log file rotation with configurable retention policies
- **Dialog Improvements**: Enhanced confirmation dialog handling with comprehensive test coverage

#### Changed

- **internal/git/operations.go** (338 lines): Git command execution and management
- **internal/ui/dialog/git_runner.go** (419 lines): Git operation execution in Task Runner
- **internal/ui/dialog/git_errors.go** (667 lines): Error handling and user-friendly messages
- **internal/ui/dialog/git_log_viewer.go** (387 lines): Commit history viewer with scrolling
- **internal/ui/dialog/log_rotator.go** (264 lines): Log file rotation and archiving
- **internal/ui/dialog/startup_error_logger.go** (58 lines): Startup error capture and logging
- **internal/ui/dialog/confirm_test.go** (246 lines): Confirmation dialog tests

#### Documentation

- **docs/DEVELOPER_GIT_GUIDE.md**: Complete developer guide for git operations
- **docs/GIT_OPERATIONS_USER_GUIDE.md**: End-user documentation for git management
- **docs/GIT_RUNNER_API.md**: Technical API reference for git execution
- **internal/git/GIT_OPERATIONS.md**: Git operations module documentation
- **internal/ui/dialog/COMPREHENSIVE_IMPLEMENTATION_GUIDE.md**: Dialog implementation patterns and best practices

## [v0.1.18] - 2025-01-11

### Enhanced UI Header & Layout Improvements

**Refined Visual Design** - Improved header styling and overall UI layout for better visual hierarchy and clarity.

#### Changed

- **Header Redesign**: Re-enabled and enhanced header component with modern styling
- **Visual Hierarchy**: Improved spacing and alignment for better content organization
- **Layout Consistency**: Refined layout.go with comprehensive test coverage (1366 new tests in layout_test.go)
- **Model Selection Dialog**: Enhanced presentation of model selection options with better formatting
- **Status Bar Integration**: Improved integration with status indicators

#### Files Modified

- **internal/ui/layout.go**: Major refactoring with improved component rendering and spacing
- **internal/ui/layout_test.go**: Comprehensive test suite added (1366+ test lines)
- **internal/ui/app.go**: Enhanced app initialization and layout management
- **internal/ui/styles.go**: Updated color and style definitions for better visual consistency
- **internal/ui/dialog/model_selection.go**: Improved dialog presentation and formatting
- **internal/config/model_config.go**: Refined configuration handling

## [v0.1.17] - 2025-01-10

### Command Runner Execution & Completion Detection

**Direct Crush Command Execution** - Fixed the Command Runner (Ctrl+B) to properly execute commands and detect completion.

#### Fixed

- **Issue #1**: After submitting a command via the Command Runner dialog, the Task Runner modal appeared but nothing happened. The command was never actually executed.
- **Issue #2**: When commands did execute (in previous versions), they would run to completion but the UI remained stuck showing "Running" status indefinitely.

#### Root Causes

1. **Execution Not Triggered**: `handleCommandRunnerSubmission` only sent `TaskStartedMsg` to create the tab, but never called `continueAdHocCommand` to actually start the command execution.
2. **No Completion Detection**: `RunCommand` function returned `chan string` for output lines but never sent completion messages (`TaskCompletedMsg`/`TaskFailedMsg`) when the command finished.

#### Changed

- **Command Execution Flow** (internal/ui/command_handlers.go):
  - Changed to use `tea.Sequence` to chain both tab creation and command execution
  - Ensures `executeAdHocCommand` (creates tab) runs before `continueAdHocCommand` (starts execution)
  - Matches the pattern used by regular task execution in `startCrushRun`
- **Completion Message Support** (internal/ui/dialog/crush_runner.go):
  - Changed `RunCommand` signature from `(<-chan string, error)` to `(chan tea.Msg, error)`
  - Output lines now wrapped in `TaskOutputMsg` for proper routing
  - Added completion detection after `cmd.Wait()`:
    - Sends `TaskCompletedMsg` on successful completion (exit code 0)
    - Sends `TaskFailedMsg` on error (non-zero exit code)
  - Removed need for `convertOutputChannelToMsgChannel` helper

## [v0.1.16] - 2025-01-09

### PRD Generation Live Output Fix

**Real-time Streaming Output** - Fixed PRD creation feature to display live output during generation in the Task Runner modal.

#### Fixed

- **Issue**: After completing the PRD creation form (`Alt+Shift+P`) and selecting a model, nothing appeared to happen. The generation was running in the background but users had no visual feedback.
- **Solution**: Implemented proper channel-based message flow to stream Crush output in real-time to the Task Runner modal.

#### Changed

- Task Runner modal now automatically appears when PRD generation starts
- Real-time streaming of Crush AI output during PRD creation
- Progress updates visible as the PRD is being generated
- Clear completion status before file save dialog
- Enhanced model selection workflow to properly trigger PRD generation
- Added `prdCreationPending` flag for correct routing between model selection and PRD execution
- Improved message ordering with `tea.Sequence` for guaranteed delivery
- Special handling for prd-creation task completion and failure states

#### Technical Details

- Channel-based messaging for streaming updates (1000 message buffer)
- Automatic TaskRunnerModal creation on TaskStartedMsg
- Separate handling for "prd-creation" task ID
- Integration with existing Crush execution infrastructure

## [v0.1.15] - 2025-01-08

### Git Integration & PRD Creation Workflow

**Git Repository Management** - Full Git integration directly in the TUI with dedicated dialogs for common Git operations.

#### Added

- **Git Menu Dialog** (`g` key): Access Git operations without leaving the TUI
  - View repository status with file changes
  - Switch between branches with visual indicators
  - Create new branches with validation
  - View commit history with details
- **Git Status Dialog**: Real-time view of staged/unstaged changes, untracked files
- **Branch Management**: List, switch, and create branches with confirmation prompts
- **Commit History Viewer**: Browse commits with author, date, and message details
- Improved Git repository detection and validation
- Enhanced executor service with better binary existence checking

#### PRD Creation & Management

- **Create PRD Dialog** (`Alt+Shift+P`): Interactive form for generating PRDs
  - Multi-field input form with title, summary, and scope
  - Textarea support for longer content
  - Configurable output filename and location
  - Real-time validation and feedback
- **PRD Prompt Selection**: Choose between user-provided and AI-generated prompts
- **State Preservation**: Form inputs preserved across dialog navigation
- Enhanced form dialog system with textarea support
- Path utilities for PRD file management

#### Next Task Output Modal

- Modal displays real-time output from `task-master next` command
- Scrollable viewport for long output
- Clean separation from log panel for better focus
- Auto-shows when executing next task command
- Keyboard shortcuts for navigation and closing

## [v0.1.13] - 2025-01-07

### Dialog Layering & Help Modal

**Centered Pop-up Help Dialog** - The help view now renders as a true modal overlay, centered on top of the TUI instead of replacing the screen.

#### Changed

- Introduced a layered render pipeline to draw dialogs over the base UI
- Help dialog is now a modal with proper borders, positioning, and background fill
- Dialog stack honors focus and overlays without breaking the underlying layout

## [v0.1.12] - 2025-01-06

### Enhanced Command Integration

**Next Task Hotkey Output** - The `n` hotkey now fully integrates with the Log panel to display real-time command output.

#### Changed

- Pressing `n` executes `task-master next` and streams output to the Log panel
- Log panel automatically shows when the command starts
- Output displays in real-time as the command executes
- Prevents concurrent command execution with guard checks
- Full command history logged to `.taskmaster/logs/tui-session.log`

## [v0.1.10] - 2025-01-05

### Critical Bug Fixes

**Task Selection & Display Stability** - Fixed a critical bug where selecting tasks in the tree view would display incorrect task details.

#### Fixed

- **Root Cause**: Subtasks in `tasks.json` had numeric IDs (1, 2, 3) instead of proper dotted notation ("1.1", "1.2", "1.3")
- This caused ID collisions in the internal task index where subtask ID "1" would overwrite parent task "1"
- Additionally, pointer instability in recursive task traversal created references to temporary memory

#### Changed

- **Automatic ID Normalization**: Added `normalizeSubtaskIDs()` function that automatically fixes all subtask IDs during task loading
  - Parent task "1" now correctly has subtasks "1.1", "1.2", "1.3"
  - Nested subtasks properly cascade: "1.1" gets subtasks "1.1.1", "1.1.2", etc.
  - Eliminates all ID collisions in the task index
- **Pointer Stability**: Refactored all task indexing and flattening functions to use stable pointers
  - `buildTaskIndex()` now uses stable pointer recursion
  - `flattenTasks()` and `flattenAllTasks()` use task index lookups
  - All selection paths (`selectNext`, `selectPrevious`, `ensureTaskSelected`) consistently use stable pointers
  - Added defensive re-fetching in `renderTaskDetails()` to guarantee correctness

[v0.1.22]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.22
[v0.1.21]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.21
[v0.1.20]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.20
[v0.1.19]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.19
[v0.1.18]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.18
[v0.1.17]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.17
[v0.1.16]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.16
[v0.1.15]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.15
[v0.1.13]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.13
[v0.1.12]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.12
[v0.1.10]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.10
