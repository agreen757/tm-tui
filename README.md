# Task Master TUI

An interactive terminal user interface for [Task Master AI](https://github.com/cyanheads/task-master-ai), forked from [Crush](https://github.com/charmbracelet/crush).

## Overview

Task Master TUI provides a beautiful, keyboard-driven interface for managing development tasks, viewing task hierarchies, and executing Task Master commands without leaving your terminal. This tool seamlessly integrates with Task Master AI to provide a rich terminal experience for project management. **Autonomously execute tasks with AI assistance and track detailed completion logs** for comprehensive project documentation and progress tracking.

## Features

- 🎯 **Task Management**: View and navigate task hierarchies with ease
- 🔄 **Real-time Sync**: Automatically updates when task files change using fsnotify
- ⌨️ **Keyboard-driven**: Full navigation and control via keyboard shortcuts
- 🎨 **Beautiful UI**: Built with Bubble Tea and Lipgloss for a polished experience
- 🔍 **Search & Filter**: Quickly find tasks by ID, title, status, or content
- 📊 **Complexity Analysis**: Analyze task complexity across your project with AI-powered scoring
- 🚀 **Task Expansion**: Generate subtasks using AI to break down complex tasks automatically
- 🏷️ **Task Tagging**: Organize tasks with custom tags and filter by tag groups
- 📦 **Project Management**: Manage multiple project-specific task views with project tags
- 🗑️ **Safe Deletion**: Delete tasks with confirmation dialogs to prevent accidents
- 📄 **PRD Creation & Parsing**: Create PRDs with AI assistance and load tasks from documents
- 🔴 **Live Output Streaming**: Real-time visual feedback for PRD generation and task execution
- 🤖 **Autonomous Task Execution**: AI-powered autonomous execution of tasks with intelligent decision-making and problem-solving
- 📋 **Detailed Completion Logging**: Comprehensive logging of task execution with implementation notes, challenges, solutions, and test results
- 💡 **Context-sensitive Help**: Dynamic help panels and status bar hints
- ⚙️ **Customizable**: Configure through simple JSON configuration
- 🎯 **Accessibility**: High-contrast themes, text labels for icons, keyboard-only navigation

## Recent Improvements (v0.1.21)

### Command Runner Modal Fix

**Fixed Double Modal Overlap** - Resolved issue where pressing CTRL+B to run commands caused two modals to appear simultaneously.

**What Was Fixed:**
- **Command Runner Dialog Overlap**: Pressing CTRL+B previously showed both the Command Runner form dialog and the Task Runner Modal at the same time, causing visual confusion
- **Root Cause**: Task Runner Modal was incorrectly being added to the dialog stack while the Command Runner Dialog was still active
- **Solution**: Task Runner Modal is now managed separately via `m.taskRunnerVisible` flag, not through the dialog stack

**Changes:**
- **internal/ui/command_handlers.go**: Removed `PushDialog()` calls for Task Runner Modal
  - `handleCommandRunnerSubmission()`: Task Runner created without adding to dialog stack
  - `ensureTaskRunnerModal()`: Updated for consistency
- Clean separation between modal types (Command Runner Dialog vs Task Runner Modal)
- Aligns with existing architecture where Task Runner is rendered independently

**User Experience:**
- **Before**: CTRL+B showed overlapping modals, unclear which was active
- **After**: Clean transition from Command Runner Dialog → Task Runner Modal
- Improved visual clarity and expected behavior

**Technical Details:**
- Task Runner Modal already managed via `m.taskRunnerVisible` flag in `app.go`
- No breaking changes to API or test expectations
- Minimal code change (2 PushDialog calls removed)

**Testing:**
- ✅ Independent test suite validates both approaches
- ✅ Build successful
- ✅ All existing tests pass
- ✅ Manual testing confirms clean modal transitions

## Previous Improvements (v0.1.20)

### Tag Management Enhancement & Keybinding Fixes

**Comprehensive Tag Management System** - Complete overhaul of tag management with new dialogs, improved workflows, and extensive test coverage.

**What Was Added:**
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

**Key Changes:**
- **internal/ui/dialog/tag_selector.go** (420 lines): New tag selection dialog with multi-select support
- **internal/ui/dialog/tag_selector_test.go** (1,401 lines): Comprehensive test suite for tag selector
- **internal/ui/dialog/tag_editor_flow.go** (262 lines): State machine for tag management workflows
- **internal/ui/dialog/tag_editor_flow_test.go** (1,463 lines): Complete flow testing with mock services
- **internal/ui/dialog/form_button_test.go** (304 lines): Button interaction and keyboard tests
- **internal/ui/dialog/expansion_scope_test.go** (148 lines): Expansion dialog tests
- **internal/ui/command_handlers.go**: Fixed Ctrl+T routing to tag management dialog
- **internal/ui/keymap.go**: Updated key bindings for tag management consistency
- **internal/ui/tag_helpers.go**: Enhanced helper functions for tag operations

**Bug Fixes:**
- Fixed Ctrl+T keybinding routing to open tag management dialog instead of tag switcher
- Improved tag context retrieval for complexity analysis
- Enhanced error handling in tag operations

**User Experience:**
- Intuitive tag selection with visual feedback
- Seamless tag creation workflow without leaving the dialog
- Consistent keyboard shortcuts across tag operations
- Better error messages and recovery hints
- Real-time tag list updates after creation

**Testing:**
- ✅ 2,800+ new test lines added
- ✅ All existing tests pass
- ✅ Build successful
- ✅ Manual testing confirms improved tag management

## Previous Improvements (v0.1.19)

### Git Operations Integration & Dialog Enhancements

**Comprehensive Git Management** - Full Git operations integration directly within the Task Master TUI with real-time streaming output and robust error handling.

**What Was Added:**
- **Git Menu** (`g` key): Access complete git operations without leaving the TUI
  - View repository status (staged/unstaged changes, untracked files)
  - Switch between branches with visual indicators
  - Create new branches with validation
  - View commit history with author and date information
- **Real-time Output Streaming**: Live streaming of git command output to Task Runner modal
- **Comprehensive Error Handling**: User-friendly error messages with helpful remediation suggestions
- **Log Rotation**: Automatic log file rotation with configurable retention policies
- **Dialog Improvements**: Enhanced confirmation dialog handling with comprehensive test coverage

**Key Changes:**
- **internal/git/operations.go** (338 lines): Git command execution and management
- **internal/ui/dialog/git_runner.go** (419 lines): Git operation execution in Task Runner
- **internal/ui/dialog/git_errors.go** (667 lines): Error handling and user-friendly messages
- **internal/ui/dialog/git_log_viewer.go** (387 lines): Commit history viewer with scrolling
- **internal/ui/dialog/log_rotator.go** (264 lines): Log file rotation and archiving
- **internal/ui/dialog/startup_error_logger.go** (58 lines): Startup error capture and logging
- **internal/ui/dialog/confirm_test.go** (246 lines): Confirmation dialog tests

**Documentation Added:**
- **docs/DEVELOPER_GIT_GUIDE.md**: Complete developer guide for git operations
- **docs/GIT_OPERATIONS_USER_GUIDE.md**: End-user documentation for git management
- **docs/GIT_RUNNER_API.md**: Technical API reference for git execution
- **internal/git/GIT_OPERATIONS.md**: Git operations module documentation
- **internal/ui/dialog/COMPREHENSIVE_IMPLEMENTATION_GUIDE.md**: Dialog implementation patterns and best practices

**User Experience:**
- Git operations execute immediately when selected
- Real-time progress feedback in Task Runner modal
- Clear status indicators (✓ for success, ✗ for errors)
- Helpful error messages guide users to solutions
- All operations logged automatically for audit trails

**Testing:**
- ✅ 1500+ new tests covering git operations, error handling, and dialogs
- ✅ All existing tests pass
- ✅ Build successful
- ✅ Manual testing confirms robust git integration

## Previous Improvements (v0.1.18)

### Enhanced UI Header & Layout Improvements

**Refined Visual Design** - Improved header styling and overall UI layout for better visual hierarchy and clarity.

**What Was Improved:**
- **Header Redesign**: Re-enabled and enhanced header component with modern styling
- **Visual Hierarchy**: Improved spacing and alignment for better content organization
- **Layout Consistency**: Refined layout.go with comprehensive test coverage (1366 new tests in layout_test.go)
- **Model Selection Dialog**: Enhanced presentation of model selection options with better formatting
- **Status Bar Integration**: Improved integration with status indicators

**Key Changes:**
- **internal/ui/layout.go**: Major refactoring with improved component rendering and spacing
- **internal/ui/layout_test.go**: Comprehensive test suite added (1366+ test lines)
- **internal/ui/app.go**: Enhanced app initialization and layout management
- **internal/ui/styles.go**: Updated color and style definitions for better visual consistency
- **internal/ui/dialog/model_selection.go**: Improved dialog presentation and formatting
- **internal/config/model_config.go**: Refined configuration handling

**User Experience:**
- Cleaner, more professional UI appearance
- Better visual feedback and navigation cues
- Improved readability with refined spacing and alignment
- More intuitive model selection interface
- Enhanced overall aesthetic consistency

**Testing:**
- ✅ 1366+ new layout tests added for comprehensive coverage
- ✅ All existing tests pass
- ✅ Build successful
- ✅ Manual testing confirms improved UI appearance

## Previous Improvements (v0.1.17)

### Command Runner Execution & Completion Detection (Deprecated Approach)

**Direct Crush Command Execution** - Fixed the Command Runner (Ctrl+B) to properly execute commands and detect completion.

**What Was Fixed:**
- **Issue #1**: After submitting a command via the Command Runner dialog, the Task Runner modal appeared but nothing happened. The command was never actually executed.
- **Issue #2**: When commands did execute (in previous versions), they would run to completion but the UI remained stuck showing "Running" status indefinitely.

**Root Causes:**
1. **Execution Not Triggered**: `handleCommandRunnerSubmission` only sent `TaskStartedMsg` to create the tab, but never called `continueAdHocCommand` to actually start the command execution.
2. **No Completion Detection**: `RunCommand` function returned `chan string` for output lines but never sent completion messages (`TaskCompletedMsg`/`TaskFailedMsg`) when the command finished.

**Changes:**
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

**User Experience:**
- **Before**: 
  - Modal showed "Running" but command never executed
  - If command did run, it would complete but stay in "Running" state forever
- **After**: 
  - Commands execute immediately when submitted
  - Real-time output streaming shows progress
  - Tab status updates to "Completed ✓" or "Failed ✗" when done
  - Clear visual feedback throughout the entire lifecycle

**Technical Details:**
- Proper message sequencing with `tea.Sequence` for guaranteed execution order
- Unified message types (`TaskOutputMsg`, `TaskCompletedMsg`, `TaskFailedMsg`) across all execution paths
- Command exit status tracked via `cmd.Wait()` error return
- Completion messages sent before channel close for proper UI updates

**Testing:**
- ✅ All unit tests pass (TestRunCommand*)
- ✅ Build successful
- ✅ Manual testing confirms proper execution and completion detection

## Previous Improvements (v0.1.16)

### PRD Generation Live Output Fix

**Real-time Streaming Output** - Fixed PRD creation feature to display live output during generation in the Task Runner modal.

**What Was Fixed:**
- **Issue**: After completing the PRD creation form (`Alt+Shift+P`) and selecting a model, nothing appeared to happen. The generation was running in the background but users had no visual feedback.
- **Solution**: Implemented proper channel-based message flow to stream Crush output in real-time to the Task Runner modal.

**Changes:**
- Task Runner modal now automatically appears when PRD generation starts
- Real-time streaming of Crush AI output during PRD creation
- Progress updates visible as the PRD is being generated
- Clear completion status before file save dialog
- Enhanced model selection workflow to properly trigger PRD generation
- Added `prdCreationPending` flag for correct routing between model selection and PRD execution
- Improved message ordering with `tea.Sequence` for guaranteed delivery
- Special handling for prd-creation task completion and failure states

**User Experience:**
- **Before**: Blank screen after form submission, unclear if anything was happening
- **After**: Immediate Task Runner modal with live streaming output showing generation progress

**Technical Details:**
- Channel-based messaging for streaming updates (1000 message buffer)
- Automatic TaskRunnerModal creation on TaskStartedMsg
- Separate handling for "prd-creation" task ID
- Integration with existing Crush execution infrastructure

## Previous Improvements (v0.1.15)

### Git Integration & PRD Creation Workflow

**Git Repository Management** - Full Git integration directly in the TUI with dedicated dialogs for common Git operations.

**Changes:**
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

**PRD Creation & Management** - Streamlined workflow for creating Product Requirements Documents directly in the TUI.

**Changes:**
- **Create PRD Dialog** (`Alt+Shift+P`): Interactive form for generating PRDs
  - Multi-field input form with title, summary, and scope
  - Textarea support for longer content
  - Configurable output filename and location
  - Real-time validation and feedback
- **PRD Prompt Selection**: Choose between user-provided and AI-generated prompts
- **State Preservation**: Form inputs preserved across dialog navigation
- Enhanced form dialog system with textarea support
- Path utilities for PRD file management

**Next Task Output Modal** - Dedicated dialog for viewing `task-master next` command output.

**Changes:**
- Modal displays real-time output from `task-master next` command
- Scrollable viewport for long output
- Clean separation from log panel for better focus
- Auto-shows when executing next task command
- Keyboard shortcuts for navigation and closing

**Benefits:**
- Complete Git workflow without context switching to terminal
- Faster branch management and status checking
- Streamlined PRD creation process integrated into task workflow
- Better organization of command output with dedicated modals
- Improved user experience with state preservation across dialogs
- Enhanced testing coverage with comprehensive test suites

## Previous Improvements (v0.1.13)

### Dialog Layering & Help Modal

**Centered Pop-up Help Dialog** - The help view now renders as a true modal overlay, centered on top of the TUI instead of replacing the screen.

**Changes:**
- Introduced a layered render pipeline to draw dialogs over the base UI
- Help dialog is now a modal with proper borders, positioning, and background fill
- Dialog stack honors focus and overlays without breaking the underlying layout

**Benefits:**
- Help stays readable without hiding the main UI context
- Consistent modal behavior aligned with other dialogs

## Previous Improvements (v0.1.12)

### Enhanced Command Integration

**Next Task Hotkey Output** - The `n` hotkey now fully integrates with the Log panel to display real-time command output.

**Changes:**
- Pressing `n` executes `task-master next` and streams output to the Log panel
- Log panel automatically shows when the command starts
- Output displays in real-time as the command executes
- Prevents concurrent command execution with guard checks
- Full command history logged to `.taskmaster/logs/tui-session.log`

**Benefits:**
- Seamless workflow for jumping to next available task
- Clear visibility of command results without context switching
- Integrated logging for debugging and audit trails

## Previous Improvements (v0.1.10)

### Critical Bug Fixes

**Task Selection & Display Stability** - Fixed a critical bug where selecting tasks in the tree view would display incorrect task details. For example, selecting Task 5 would show Task 10's information in the details panel.

**Root Cause:** 
- Subtasks in `tasks.json` had numeric IDs (1, 2, 3) instead of proper dotted notation ("1.1", "1.2", "1.3")
- This caused ID collisions in the internal task index where subtask ID "1" would overwrite parent task "1"
- Additionally, pointer instability in recursive task traversal created references to temporary memory

**Solution Implemented:**
- **Automatic ID Normalization**: Added `normalizeSubtaskIDs()` function that automatically fixes all subtask IDs during task loading
  - Parent task "1" now correctly has subtasks "1.1", "1.2", "1.3"
  - Nested subtasks properly cascade: "1.1" gets subtasks "1.1.1", "1.1.2", etc.
  - Eliminates all ID collisions in the task index
- **Pointer Stability**: Refactored all task indexing and flattening functions to use stable pointers
  - `buildTaskIndex()` now uses stable pointer recursion
  - `flattenTasks()` and `flattenAllTasks()` use task index lookups
  - All selection paths (`selectNext`, `selectPrevious`, `ensureTaskSelected`) consistently use stable pointers
  - Added defensive re-fetching in `renderTaskDetails()` to guarantee correctness

**Benefits:**
- Task details panel now always shows the correct task information
- Tasks with subtasks (Tasks 1-5) are now properly expandable in the tree view
- No more confusion between parent tasks and their subtasks
- Robust against future task file format variations
- Improved performance with O(1) task lookups

These fixes ensure a reliable and predictable user experience when navigating complex task hierarchies.

## Memory System

Task Master TUI includes a persistent memory system powered by **BadgerDB** for AI agents and LLMs. This enables cross-session learning, context preservation, and implementation artifact storage with high-performance key-value operations.

The memory system is a **core feature** that works seamlessly with your Task Master workflow, allowing agents to learn from previous implementations and maintain context across sessions.

### Quick Start

```bash
# Build the project (includes memory binary)
make build

# Store information
./bin/memory store -key "readme:main" -file README.md
./bin/memory store -key "log:2.1" -value "Completed auth implementation"

# Retrieve information
./bin/memory get -key "log:2.1"

# List all stored keys
./bin/memory list
./bin/memory list -prefix "log:"  # Filter by prefix

# Log task progress
./bin/memory log -task "2.1" -message "Started implementing JWT validation"
```

Memory data is stored in `.taskmaster/memory/` using BadgerDB for reliable persistence across sessions.

### How It Works

The memory system stores key-value pairs persistently with BadgerDB:

- **Keys**: Organized by prefix (task:, readme:, log:, context:)
- **Values**: Text or structured JSON data
- **Storage**: BadgerDB embedded key-value store at `.taskmaster/memory/`
- **Access**: Command-line tool or Go API
- **Performance**: O(1) lookups, ACID transactions, optimized for fast queries

### Memory Key Conventions

| Prefix | Purpose | Example |
|--------|---------|----------|
| `task:` | Task metadata and status | `task:2.1` → task info |
| `log:` | Task completion logs | `log:2.1` → implementation details |
| `readme:` | Cached documentation | `readme:main` → README content |
| `context:` | LLM context snapshots | `context:session-1` → session notes |

### Storage & Performance

**Default**: BadgerDB storage at `.taskmaster/memory/`
- **Fast**: O(1) key lookups
- **Reliable**: ACID transactions for data consistency
- **Embedded**: No external server required
- **Scalable**: Optimized for development and production
- **Typical disk usage**: <100MB for thousands of entries
- **Concurrent safe**: Handles multiple CLI invocations

**Implementation details**:
- Database path: `.taskmaster/memory/` (auto-created)
- Concurrent safe (handles multiple CLI invocations)
- Automatic garbage collection for obsolete values
- Supports key scanning with prefixes (efficient filtering)
- Data persists across Task Master TUI restarts

### Command Reference

Full command documentation available via:

```bash
./bin/memory help
```

Key commands:

**Store data**

```bash
./bin/memory store -key <key> -file <file>
./bin/memory store -key <key> -value <value> [-json]
```

**Retrieve data**

```bash
./bin/memory get -key <key>
```

**Delete data**

```bash
./bin/memory delete -key <key>
```

**List keys**

```bash
./bin/memory list [-prefix <prefix>] [-json]
```

**Log task activity**

```bash
./bin/memory log -task <id> -message "<activity>"
```

**List stored READMEs**

```bash
./bin/memory readmes
```

## Agent Workflow Integration

The memory system is designed to seamlessly integrate with AI agent workflows. Agents can use the memory system to maintain context, log progress, and store implementation artifacts across sessions.

### Integration with Task Master CLI

Agents working with Task Master can leverage memory for:

- **Context Preservation**: Store session context and previous implementation details
- **Task Logging**: Use `./bin/memory log` to track task completion and progress
- **Implementation Artifacts**: Store code snippets, design decisions, and test results
- **Cross-Session Learning**: Retrieve previous implementations to inform new work

### Typical Agent Workflow

1. **Initialize**: Agent creates a Task Master helper with `DefaultHelper()`
2. **Load Context**: Retrieve previous implementation notes from memory
3. **Implement**: Work on the task, storing progress in memory
4. **Persist**: Log completion details with `memory log` command
5. **Next Task**: Retrieve context for the next task from memory

### Example Integration

```bash
# Load previous implementation context
./bin/memory get -key "log:task-2.1"

# Store implementation notes during work
./bin/memory store -key "log:current-task" -value "Completed JWT validation middleware"

# Log task completion
./bin/memory log -task "3.1" -message "Implemented role-based access control"
```

This integration enables continuous learning and improved decision-making across development sessions.

## Prerequisites

- Go 1.23 or later
- [Task Master AI](https://github.com/cyanheads/task-master-ai) installed globally (`npm i -g task-master-ai`)
- A Task Master project (`.taskmaster` directory with tasks)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/agreen757/tm-tui.git
cd tm-tui

# Build and install (includes Crush CLI for task execution)
make install

# Or install everything (tm-tui, memory tool, crush)
make install-all

# Or just build without installing
make build

# Check if Crush is installed
make check-crush
```

### Using Go Install

```bash
# Install tm-tui
go install github.com/agreen757/tm-tui/cmd/tm-tui@latest

# Install Crush CLI (required for task execution feature)
go install github.com/charmbracelet/crush@latest
```

**Note:** After installation, ensure `~/go/bin` is in your PATH:

```bash
# For Zsh (macOS default)
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# For Bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bash_profile
source ~/.bash_profile
```

Alternatively, run directly with the full path:
```bash
~/go/bin/tm-tui
```

## Usage

Navigate to a directory with a Task Master project (containing `.taskmaster` directory) and run:

```bash
tm-tui

# or run directly without installing
go run ./cmd/tm-tui/main.go
```

### Keyboard Shortcuts

#### Navigation
- `↑/k` - Move up
- `↓/j` - Move down  
- `←/h` - Navigate left / collapse
- `→/l` - Navigate right / expand
- `Tab` - Switch between panels
- `PageUp/PageDn` - Scroll by page
- `Esc` - Back / close dialog

#### Task Management
- `n` - Jump to next available task
- `s` - Change task status
- `Enter` - Select item / toggle expand
- `Space` - Multi-select task for bulk operations
- `Ctrl+R` / `Alt+R` - Run task with Crush AI agent
- `Alt+X` - Expand tasks (opens scope selection dialog)
  - Supports single task, all tasks, task range, or by tag
  - AI-powered expansion with --research flag
  - Configurable depth (1-3 levels) and subtask count
- `Alt+D` - Delete selected task (with confirmation)
- `d` - Mark task as done (quick status change)
- `p` - Change task priority

#### PRD & Document Management
- `Alt+Shift+P` - Create new PRD (interactive form)
- `Alt+P` - Parse PRD file (load tasks from document)

#### Git Operations
- `g` - Open Git menu
  - View repository status
  - Switch branches
  - Create new branches
  - View commit history

#### Complexity & Analysis
- `Ctrl+Shift+C` - Analyze task complexity (AI-powered scoring)

#### Tag Management
- `Ctrl+T` - Manage tag contexts (switch active tag, visual selector)
- `Ctrl+Shift+T` - Add new tag context (create tag with options)
- `Ctrl+Shift+A` - Quick tag access (selector with add option)
- `Alt+Shift+T` - View project tags (filter by project context)

#### Project Management
- `Ctrl+P` - Command palette (general project access)
- `Alt+Shift+Q` - Quick project switch (find projects quickly)
- `Alt+Shift+S` - Project search (full-text project search)

#### Filtering & Search
- `/` - Search tasks by ID, title, or content
- `f` - Filter tasks by status or tag
- `F` - Clear all filters

#### View & Display
- `1` - Switch to tree view
- `2` - Switch to list view
- `Alt+T` - Cycle through view modes
- `Alt+L` - Toggle log panel
- `Alt+I` - Toggle details panel

#### Global Commands
- `?` - Show/hide help overlay
- `:` - Open command palette for additional commands
- `r` - Refresh tasks from disk
- `Ctrl+Z` - Undo (task modifications)
- `Ctrl+Shift+C` - Clear TUI state
- `q` - Quit TUI

## Common Workflows

### Creating a PRD with AI
1. Press `Alt+Shift+P` to open the "Create PRD" dialog
2. Fill in the PRD details:
   - **Title**: Short descriptive title for your project
   - **Summary**: Brief overview of what you're building
   - **Scope**: Detailed requirements, constraints, and technical details
   - **Output Filename**: Where to save the generated PRD
3. Select an AI model from the model selection dialog
4. **Watch real-time generation**: Task Runner modal appears showing live Crush output
5. Monitor as the AI generates your PRD with streaming feedback
6. Review and save when complete - file save dialog appears automatically
7. Use `Alt+P` to parse the generated PRD into tasks

**New in v0.1.16:** The Task Runner modal now displays live streaming output during PRD generation, providing immediate visual feedback and progress updates.

### Creating Tasks from a PRD
1. Press `Alt+P` to open the "Parse PRD" dialog
2. Select or enter the path to your PRD document
3. Review the generated tasks in the main view
4. Edit, organize, or prioritize as needed

### Analyzing Task Complexity
1. Navigate to a task or select multiple tasks with `Space`
2. Press `Ctrl+Shift+C` to open "Analyze Complexity" dialog
3. Choose analysis scope: all tasks, selected task, or by tag
   - Use visual tag selector to pick specific tags for analysis
4. View complexity scores (LOW, MEDIUM, HIGH, VERY HIGH)
5. Filter and sort results for focused planning

### Expanding Tasks into Subtasks
1. Select a task or prepare to expand all tasks
2. Press `Alt+X` to open the "Expand Tasks" dialog
3. Choose expansion scope:
   - **Selected task only** - Expand just the current task
   - **All tasks** - Expand all tasks in the project
   - **Task range** - Expand tasks from ID X to ID Y
   - **By tag** - Use visual tag selector to pick multiple tags for expansion
4. Configure options:
   - Expansion depth: 1-3 levels of nested subtasks
   - Number of subtasks: Leave blank for auto-detection
   - AI assistance: Enable `--research` for intelligent expansion
5. Monitor progress in real-time as CLI executes
6. Review newly created subtasks in the task tree
7. Tasks are automatically reloaded after expansion completes

**Note:** This feature executes `task-master expand` CLI commands. Ensure the Task Master CLI is properly installed and accessible.

### Managing Tags with Visual Selection

The new tag management system provides visual, interactive tag selection:

1. Press `Ctrl+T` to open Tag Management
   - Visual list shows all available tags with metadata
   - Active tag marked with indicator (●)
   - Each tag shows task count and completion status

2. **Switch to a tag**: Select it from the list
   - Immediate context switch
   - All subsequent operations use new tag context

3. **Create a new tag**: Select "Add New Tag..." option
   - Opens inline creation dialog
   - Configure name, copy-from options, description
   - New tag immediately available in list
   - Select it to switch to new context

4. **Quick tag access**: Press `Ctrl+Shift+A` for fast switching
   - Opens tag selector with minimal overhead
   - Add new tags on-the-fly
   - Close when done (Esc)

### Running Tasks with Crush AI

Task Master TUI integrates with [Crush](https://github.com/charmbracelet/crush), an AI-powered terminal assistant, to execute tasks **autonomously with minimal human intervention**. The tool handles complex implementation tasks, test execution, and debugging with detailed logging of all activities and outcomes.

#### Autonomous Execution Capabilities

Crush AI autonomously:
- **Analyzes task requirements** from the task description and implementation details
- **Explores the codebase** to understand project structure and patterns
- **Implements features** following existing code conventions and best practices
- **Writes and runs tests** to validate implementation
- **Debugs issues** by analyzing error messages and adjusting approach
- **Documents changes** through detailed implementation logs
- **Handles edge cases** and error scenarios systematically

#### Prerequisites
- Crush CLI installed and accessible in PATH (`crush` command available)
- API keys configured in Crush for your preferred AI provider

#### Running a Task
1. Navigate to the task you want to execute
2. Press `Ctrl+R` (or `Alt+R`) to start the task runner
3. Model selection inside the task runner is not currently enabled
4. To choose a model, open the Crush CLI (`crush`), press `Ctrl+L` to select a model, exit Crush, then restart the TUI
5. The Task Runner modal opens and shows real-time output from Crush
6. Monitor progress as Crush works through the task

#### Task Runner Modal Controls
- `Tab` / `Shift+Tab` - Switch between task tabs (when running multiple tasks)
- `1-9` - Jump directly to tab number
- `↑/↓` - Scroll output
- `PgUp/PgDn` - Page scroll
- `Home/End` - Scroll to top/bottom
- `M` - Minimize/maximize modal
- `Ctrl+C` - Cancel running task (with confirmation for long-running tasks)
- `Esc` - Close modal (only when no tasks are running)

#### Features
- **Real-time streaming**: See Crush's output as it works
- **Multi-task support**: Run up to 9 tasks concurrently in separate tabs
- **Automatic logging**: All output saved to `.taskmaster/logs/crush-run-<task-id>-<timestamp>.log`
- **Cancellation**: Stop tasks that are stuck or producing incorrect results
- **Minimizable**: Continue using the TUI while tasks run in the background

#### Detailed Completion Logging

When a task executes autonomously, all activities are logged to `.taskmaster/logs/` with comprehensive details:

**Log File Structure** (`.taskmaster/logs/crush-run-<task-id>-<timestamp>.log`):
- **Command Execution**: Full output from all commands run
- **Code Changes**: File modifications with before/after snippets
- **Test Results**: Complete test output and pass/fail status
- **Error Handling**: Error messages encountered and resolution strategies
- **Decision Points**: AI reasoning for implementation choices
- **Debugging Steps**: Investigation and troubleshooting process
- **Performance Notes**: Execution time and resource usage

**Task Completion Summary**:
1. Navigate to a completed task
2. Press `Ctrl+R` to view the task log in the Task Runner modal
3. Scroll through the output to see the complete execution history
4. Use `↑/↓` arrows to review specific sections
5. Task status automatically updates to "done" upon successful completion

**Manual Completion Logging**:
You can add detailed notes after task execution using the Task Master CLI:
```bash
# Log implementation details for a completed task
task-master update-subtask --id=<task-id> --prompt="Implementation completed:
- [List of features implemented]
- [Files modified and changes made]
- [Challenges encountered and solutions applied]
- [Test results and coverage achieved]
- [Integration with other system components]"
```

**Log Retention and Analysis**:
- All execution logs stored in `.taskmaster/logs/` for future reference
- Logs retained until manually deleted (automatic rotation available)
- Use logs to understand task execution patterns and identify bottlenecks
- Review logs during code review or documentation phases
- Reference logs when implementing related tasks for consistency

#### Task Prompt Generation
When you run a task, the TUI generates a prompt for Crush that includes:
- Task ID and title
- Full task description
- Implementation details
- Test strategy
- Priority level
- Dependencies (if any)

You can customize the prompt template by creating a `CRUSH_RUN_INSTRUCTIONS.md` file in your project root. The template supports Go text/template syntax with the following variables:
- `{{.TaskID}}` - Task identifier
- `{{.Title}}` - Task title
- `{{.Description}}` - Task description
- `{{.Details}}` - Implementation details
- `{{.TestStrategy}}` - Testing approach
- `{{.Priority}}` - Task priority
- `{{.Dependencies}}` - Comma-separated dependency IDs

### Managing Task Tags
1. Press `Alt+A` to add tags to the selected task
2. Create new tags or select from existing tags
3. Use `Ctrl+Shift+M` to open the tag context manager
4. Manage tag groups, rename, or organize tags
5. Filter tasks by tag using `f` and selecting tags

### Switching Projects
1. Press `Ctrl+T` to open the project switcher
2. Navigate to your desired project
3. View tasks specific to that project context
4. Use project tags to organize cross-project work

### Git Operations with Task Runner

Git operations in Task Master TUI are executed via the Task Runner, providing real-time streaming output and proper command lifecycle management.

#### Starting a Git Operation

1. Press `g` to open the Git Menu
2. Choose your operation:
   - **Status**: View repository status (files changed, untracked files, etc.)
   - **Switch Branch**: Switch to a different branch
   - **Create Branch**: Create a new branch
   - **Recent Commits**: View recent commit history
3. For branch operations, select from the list and press `Enter`
4. The **Task Runner modal** opens automatically showing:
   - Command being executed
   - Real-time streaming output from git
   - Status indicator (Running/Completed/Failed)

#### Understanding the Task Runner Output

When a git operation executes, you'll see:

- **Header**: Shows the git command being executed (e.g., "git checkout main")
- **Output**: Real-time streaming output from git, line by line
- **Status**: Updates at the bottom showing completion status
- **Logs**: All output is automatically saved to `.taskmaster/logs/git-command-<operation>-<timestamp>.log`

#### Task Runner Controls During Git Operations

- `↑/↓` - Scroll through output
- `Home/End` - Jump to top/bottom of output
- `PgUp/PgDn` - Page scroll through output
- `M` - Minimize/maximize the modal (continue using TUI in background)
- `Esc` - Close modal (when operation is complete)

#### Handling Git Operation Results

**Successful Operation:**
- Output shows git's normal success message
- Status bar displays "Completed ✓"
- You can press `Esc` to close the modal
- Changes are immediately reflected (branches refresh, status updates, etc.)

**Failed Operation:**
- Output shows git's error message
- Status bar displays "Failed ✗"
- Error details are logged for review
- You can close the modal and try again

#### Example Workflows

**Switching Branches**
```
1. Press g → Switch Branch
2. Highlight "feature/my-feature" in the list
3. Press Enter
4. Task Runner shows: git checkout feature/my-feature
5. Output streams in real-time
6. Branch switches, status updates
7. Press Esc to close
```

**Creating a Branch**
```
1. Press g → Create Branch
2. Enter branch name when prompted
3. Task Runner shows: git checkout -b new-feature
4. Output streams in real-time
5. New branch created and checked out
6. Press Esc to close
```

## Development

### Building

```bash
make build
```

### Running

```bash
make run
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## Project Structure

```
tm-tui/
├── cmd/
│   └── tm-tui/                  # Main executable entry point
├── configs/                     # Configuration files
│   └── default.json             # Default configuration
├── internal/                    # Internal packages
│   ├── cli/                     # CLI command definitions
│   ├── config/                  # Configuration loading and validation
│   ├── executor/                # Command execution service
│   ├── taskmaster/              # Task Master integration service
│   └── ui/                      # UI components and models
├── .taskmaster/                 # Task Master files (when used)
│   ├── tasks/                   # Task files directory
│   ├── docs/                    # Documentation
│   ├── reports/                 # Analysis reports
│   └── config.json              # Task Master config
├── go.mod                       # Go module definition
├── go.sum                       # Go dependency checksums
└── Makefile                     # Build and development targets
```

## Core Components

### Task Master Service

The Task Master service (`internal/taskmaster`) provides the following functionality:

- Detection of the nearest `.taskmaster` directory from the current working directory
- Loading and parsing of `tasks.json` files into an in-memory task tree
- Validation of tasks, including dependency checks and status validation
- Real-time file watching to detect changes to task files
- Fast indexing of tasks for O(1) lookups by ID

### UI Components

The UI layer (`internal/ui`) implements a rich terminal interface with:

- Multiple views (task list, task details, help panel)
- Keyboard navigation and shortcuts
- Status bar with contextual hints
- Search and filtering capabilities
- Styled rendering using Lipgloss
- Panel-based layout with dynamic resizing

### Executor Service

The executor service (`internal/executor`) handles:

- Running Task Master CLI commands
- Executing task-related operations
- Managing subprocesses
- Capturing command output for display

## Dependencies

The project relies on the following key Go modules:

### Primary Dependencies

- **github.com/charmbracelet/bubbletea** (v1.3.10): TUI framework for building interactive terminal applications
- **github.com/charmbracelet/bubbles** (v0.21.0): Common components for Bubble Tea applications (lists, viewports, text inputs)
- **github.com/charmbracelet/lipgloss** (v1.1.0): Style definitions for terminal UI applications
- **github.com/fsnotify/fsnotify** (v1.9.0): File system notifications for auto-refreshing when task files change
- **github.com/spf13/cobra** (v1.10.2): Command-line interface framework

### Notable Indirect Dependencies

- github.com/atotto/clipboard: Clipboard operations
- github.com/charmbracelet/x/ansi, cellbuf, term: Terminal utilities
- github.com/muesli/termenv: Terminal environment utilities
- github.com/rivo/uniseg: Unicode text segmentation

## Configuration

Configuration is loaded from `configs/default.json`. You can customize:

- Key bindings
- Theme colors
- UI behavior
- Refresh intervals

Example configuration:

```json
{
  "colors": {
    "accent": "#6D98BA",
    "background": "#1F2335",
    "foreground": "#C0CAF5",
    "success": "#9ECE6A",
    "warning": "#E0AF68",
    "error": "#F7768E"
  },
  "keymap": {
    "quit": "q",
    "help": "?",
    "search": "/"
  },
  "display": {
    "show_status_bar": true,
    "compact_mode": false,
    "theme": "dark"
  }
}
```

## Crush Configuration

Task Master TUI integrates with [Crush](https://github.com/charmbracelet/crush) for AI-powered task execution. The application automatically manages a `.crush.json` configuration file for you.

### Automatic Initialization

When you run tm-tui for the first time in a project, it automatically creates a `.crush.json` file with sensible defaults:

- **Location**: Project root (detected via `.taskmaster`, `.git`, or `go.mod` markers)
- **Default Content**: Schema reference, empty model field, skills paths
- **Non-destructive**: Existing `.crush.json` files are never modified or overwritten
- **Idempotent**: Multiple application starts won't modify your config

The initialization happens during application startup and logs warnings (not errors) if it encounters issues.

### Makefile Targets

The project includes Makefile targets for managing Crush configuration:

```bash
# Check if .crush.json exists (useful for CI/scripts)
make check-project-setup

# Create .crush.json if missing (part of installation)
make init-crush-config

# Full installation (includes init-crush-config)
make install
```

### Manual Configuration

You can manually edit `.crush.json` to configure:

- **model**: Your preferred AI model (e.g., `"gpt-4"`, `"claude-3-sonnet"`)
- **context_paths**: Additional files to include in Crush context
- **skills_paths**: Custom Crush skills directories

Example `.crush.json`:

```json
{
  "$schema": "https://charm.land/crush.json",
  "model": "claude-3-5-sonnet-20241022",
  "options": {
    "context_paths": [],
    "skills_paths": ["./.crush/skills"]
  },
  "version": "1.0"
}
```

### Environment Variables

Override default behavior with environment variables:

- `CRUSH_CONFIG_PATH`: Explicit path to .crush.json
- `CRUSH_PROJECT_ROOT`: Override project root detection

These are useful for multi-project setups or non-standard directory structures.

## Requirements

- Go 1.23+
- Task Master AI installed and accessible in PATH
- A `.taskmaster` directory in your working directory or parent directories

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT

## Credits

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Forked from [Crush](https://github.com/charmbracelet/crush)
- Integrates with [Task Master AI](https://github.com/cyanheads/task-master-ai)
- Uses [fsnotify](https://github.com/fsnotify/fsnotify) for file system monitoring
- UI components from [Charm](https://charm.sh/)'s libraries
