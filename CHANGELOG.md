# Changelog

All notable changes to Task Master TUI are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Memory Leak Analysis & Fixes (PRD)

**Memory Leak Fixes** - Comprehensive analysis and PRD for addressing 14 identified memory leaks across the codebase.

#### Added

- **Memory Leak Analysis Report**: Identified 14 potential memory leaks with severity classification
  - 5 High Severity: Goroutine leaks, uncancellable command contexts
  - 5 Medium Severity: Channel leaks, missing timeouts, timer allocation in hot paths
  - 4 Low Severity: Unbounded maps, missing cache eviction
- **Memory Leak Fixes PRD**: `.taskmaster/docs/memory-leak-fixes-prd.md` with:
  - Detailed analysis of each issue with exact file locations and line numbers
  - Code examples showing current patterns and required fixes
  - Acceptance criteria and success metrics for each fix
  - 4-phase implementation plan spanning 4 weeks
  - Comprehensive testing requirements and risk mitigations

#### Technical Details

- **High Priority Issues**: Progress forwarding goroutines, debounce timer callbacks, uncancellable external commands
- **Medium Priority Issues**: Channel cleanup, WaitGroup timeouts, timer allocation optimization
- **Low Priority Issues**: Unbounded command history, lastCheckTime map cleanup, task cache eviction

## [v0.1.28] - 2026-01-28

### Filter Status Indication

**Enhanced UI Visibility for Active Filters** - Added FilterStatusView component to display currently active filters and improve user awareness of filtering state.

#### Added

- **FilterStatusView Component**: New UI component displays active filter state
  - Visual indicator at the top of task list showing applied filters
  - Clear display of filter count and names/values
  - Automatic updates when filters are applied or cleared
  - Integrated seamlessly into main task list view

- **Filter Status Display**: Enhanced user awareness of filtering
  - Shows exactly which filters are currently active
  - Prevents confusion about hidden tasks
  - Makes it obvious when results are filtered vs. showing all tasks
  - Multiple filter indicators for complex filter combinations

#### Changed

- **Task List View**: Enhanced with filter status indicator
  - FilterStatusView renders above task list when filters active
  - Automatic hide/show based on filter state
  - Consistent styling with existing UI components

#### User Experience

- Better visibility into filtered task lists
- Reduced confusion about task visibility changes
- Clear feedback when applying or removing filters
- Improved understanding of task set size

## [v0.1.27] - 2026-01-20

### File Dialog and Tag Selector Improvements

**Enhanced UI Component Dimensions and Filtering** - Refactored file dialog and tag selector for better dimensions and improved filtering functionality.

#### Added

- Improved file dialog sizing and layout
- Enhanced tag selector functionality
- Better filtering implementation

#### Changed

- File dialog component dimensions optimized
- Tag selector visual improvements

## [v0.1.26] - 2026-01-19

### Test Enhancement and Caching Improvements

**Ready Tasks Dialog Testing and Cache Optimization** - Added comprehensive installation tests and enhanced caching mechanism for improved performance on slow file systems.

#### Added

- Comprehensive tests for slow file reads
- Enhanced caching mechanism in ReadyTasksDialog
- Support for slow storage systems

## [v0.1.25] - 2026-01-19

### Installation & Dependencies, Concurrent Task Execution, and Enterprise Features

**Installation Improvements, Multi-task Execution, and Production-Ready Features** - Comprehensive enhancements including streamlined installation, concurrent task execution support, advanced task management, and memory system integration.

#### Added

- **Multi-Task Execution Support**:
  - Run up to 9 tasks concurrently in separate tabs within the Task Runner modal
  - Use `Tab`/`Shift+Tab` to switch between active task tabs
  - Press `1-9` to jump directly to specific task tabs
  - Real-time streaming output for each concurrent task
  - Automatic logging of all task outputs to `.taskmaster/logs/crush-run-<task-id>-<timestamp>.log`
  - Independent cancellation of individual running tasks

- **Installation Targets**:
  - `make install-all`: Install all binaries (tm-tui, memory, task-master, crush, gemini) in one command
  - `make install-task-master`: Standalone Task Master AI CLI installation via npm
  - `make check-task-master`: Verify Task Master CLI is properly installed
  - `make check-deps`: Verify all runtime dependencies are available

- **Test Targets**:
  - `make test-coverage`: Run tests with coverage report generation
  - `make test-unit`: Run unit tests only
  - `make test-integration`: Run integration tests only
  - `make test-ci`: Run tests optimized for CI/CD pipelines
  - `make test-suite`: Run full test suite with coverage verification (target: >80%)

- **Comprehensive Troubleshooting Guide** (2000+ lines in README):
  - "Task Master CLI Not Found" - 10 solutions with verification steps
  - "Permission Errors During Installation" - 3 methods (nvm, npm fix, sudo)
  - "Node.js Not Installed or Wrong Version" - Platform-specific installation
  - "Multiple Node.js Installations" - Conflict resolution for nvm/Homebrew/system
  - "Platform-Specific Installation Notes" - macOS, Ubuntu/Debian, Alpine, Windows
  - "Edge Cases and Known Issues" - Network errors, disk space, npm version

- **Ready Tasks List Dialog**: Pre-fetches upcoming tasks to reduce waiting time during task selection
- **Execution Queue System**: Manages concurrent task execution with proper resource handling
- **Memory System Integration**: BadgerDB-backed persistent storage for cross-session context

#### Changed

- **Makefile**: Completely restructured with 500+ lines of documentation
  - Added 15+ new targets with clear descriptions
  - Enhanced error handling and user feedback
  - Consistent formatting and best practices

- **README.md**: Extensive expansion with new sections:
  - Multi-task execution workflows
  - Concurrent task management
  - Memory system documentation (BadgerDB integration)
  - Enhanced installation verification

- **Task Runner Modal**: Enhanced for multi-task support
  - Separate tabs for each running task
  - Per-task output streams and status
  - Individual task cancellation controls

#### Test Coverage

- Full test suite verifies installation targets work correctly
- Build verification confirms all changes compile without errors
- 3500+ new test lines for ready tasks list functionality
- 777+ new test lines for execution queue validation
- 706+ new test lines for task runner modal enhancements
- Help text validation ensures all targets appear in make help
- Code review confirms Makefile follows best practices

#### Docker & Platform Testing

- Comprehensive Docker-based testing for Linux (Ubuntu, Alpine)
- Windows testing infrastructure using Docker Git Bash simulation
- Installation verification for all major Node.js installation methods
- Edge case testing framework for network, permissions, and disk issues
- Platform-specific test reports and findings documentation

#### User Experience

- Faster, more reliable installation with `make install-all`
- Concurrent task execution enables parallel development workflows
- Clear error messages guide users to solutions
- Comprehensive documentation reduces support burden
- Platform-specific guidance eliminates guesswork
- Pre-fetching and execution queue improve overall responsiveness

## [v0.1.24] - 2026-01-16

### Task Score Highlighting

**Enhanced Complexity Visualization** - Improved visual representation of task complexity scores with consistent color-coding and enhanced accessibility support.

#### Added

- **Color-coded Complexity System**: Standardized color scheme for complexity levels
  - Low complexity: Royal Blue (#4169E1)
  - Medium complexity: Dark Turquoise (#00CED1)
  - High complexity: Orange (#FFA500)
  - Very High complexity: Crimson (#DC143C)
- **Accessibility Improvements**: Text-based alternatives for color-coded indicators
  - `GetComplexityLabel()`: Returns text representation (LOW, MEDIUM, HIGH, VERY HIGH)
  - `GetComplexityIndicator()`: Combined format with both text and numeric score
  - Consistent labeling across all UI components
- **Complexity Style Functions**:
  - `GetComplexityStyle()`: Style based on numeric complexity scores
  - `GetComplexityLevelStyle()`: Style based on complexity level enums
  - Consistent bold formatting and color application

#### Changed

- **internal/ui/styles.go**: Enhanced with complexity styling constants and functions
- **internal/ui/styles_test.go** (213 lines): Comprehensive test coverage for complexity styling
- **internal/ui/complexity.go**: Updated for consistent level representation
- **internal/ui/dialog/complexity_report.go**: Enhanced report dialog with new styling
- **internal/ui/app.go**: Updated complexity rendering

#### Test Coverage

- **200+ new test lines** focusing on complexity style consistency
- Tests for color constants and their correct hex values
- Tests for style application based on complexity thresholds
- Boundary testing for complexity level transitions

## [v0.1.24] - 2026-01-16

### Task Score Highlighting

**Enhanced Complexity Visualization** - Improved visual representation of task complexity scores with consistent color-coding and enhanced accessibility support.

#### Added

- **Color-coded Complexity System**: Standardized color scheme for complexity levels
  - Low complexity: Royal Blue (#4169E1)
  - Medium complexity: Dark Turquoise (#00CED1)
  - High complexity: Orange (#FFA500)
  - Very High complexity: Crimson (#DC143C)
- **Accessibility Improvements**: Text-based alternatives for color-coded indicators
  - `GetComplexityLabel()`: Returns text representation (LOW, MEDIUM, HIGH, VERY HIGH)
  - `GetComplexityIndicator()`: Combined format with both text and numeric score
  - Consistent labeling across all UI components
- **Complexity Style Functions**:
  - `GetComplexityStyle()`: Style based on numeric complexity scores
  - `GetComplexityLevelStyle()`: Style based on complexity level enums
  - Consistent bold formatting and color application

#### Changed

- **internal/ui/styles.go**: Enhanced with complexity styling constants and functions
- **internal/ui/styles_test.go** (213 lines): Comprehensive test coverage for complexity styling
- **internal/ui/complexity.go**: Updated for consistent level representation
- **internal/ui/dialog/complexity_report.go**: Enhanced report dialog with new styling
- **internal/ui/app.go**: Updated complexity rendering

#### Test Coverage

- **200+ new test lines** focusing on complexity style consistency
- Tests for color constants and their correct hex values
- Tests for style application based on complexity thresholds
- Boundary testing for complexity level transitions
- Comprehensive accessibility testing

#### User Experience

- More intuitive visual complexity representation
- Consistent color-coding throughout the application
- Improved readability with combined text and color indicators
- Better accessibility for users with color vision deficiencies
- Enhanced visual hierarchy in complexity reports

## [v0.1.23] - 2026-01-16

### Log Browser & Performance Optimization

**Interactive Log File Browser** - Browse, view, and search log files directly within the TUI with a comprehensive feature set and performance optimizations.

#### Added

- **Log File Browser Dialog** (`Ctrl+F`): Browse and view log files in the TUI
  - Navigate all log files in `.taskmaster/logs/` directory
  - View files with syntax highlighting and markdown support
  - Toggle line numbers, word wrap, and search within content
  - Optimized for large log files with virtualized rendering
  - Support for filtering logs by tag or task ID
- **LRU Cache Implementation**: Thread-safe Least Recently Used cache
  - Efficient key-value storage with configurable size limits
  - Automatic eviction of least recently used items
  - Thread-safe operations with mutex protection
  - Support for Get, Put, Delete, Clear, and Keys operations
  - Comprehensive benchmark suite for performance validation
- **State Preservation**: Save and restore UI state when opening dialogs
  - Remember selected task, scroll position, and panel focus
  - Seamless return to previous state after dialog closes
  - Handles nested dialog scenarios correctly

#### Changed

- **internal/ui/dialog/log_browser.go** (580 lines): Log file browser dialog implementation
- **internal/ui/dialog/log_file_browser.go** (797 lines): File navigation interface
- **internal/ui/dialog/log_viewer.go** (1265 lines): Log content viewer with advanced features
- **internal/ui/dialog/log_tag_selector.go** (451 lines): Tag-based log filtering
- **internal/ui/dialog/lru_cache.go** (118 lines): Thread-safe LRU cache implementation
- **internal/ui/keymap.go**: Added `Ctrl+F` keybinding for log browser
- **internal/ui/app.go**: State preservation for dialog integration

#### Test Coverage

- **10,000+ new test lines** across 14 new test files
- Comprehensive edge case handling for large files, performance scenarios
- Accessibility testing for keyboard navigation and screen readers
- Integration tests with mock file system for reliable testing
- Performance benchmarks for LRU cache operations

#### User Experience

- Quick access to log files with `Ctrl+F` from anywhere in the TUI
- File browser navigation with familiar keyboard controls
- Content viewer with line numbers, word wrap, and search
- Syntax highlighting for common file formats
- State preservation ensures seamless dialog integration
- Improved performance with caching for frequently accessed data

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

[v0.1.24]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.24
[v0.1.23]: https://github.com/agreen757/tm-tui/releases/tag/v0.1.23
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
