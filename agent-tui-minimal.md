# Product Requirements Document: Task Master Agent TUI

## Overview

This project is a fork of the [Crush TUI](https://github.com/charmbracelet/crush/tree/main) application that provides an interactive terminal user interface for managing and executing development tasks. It directly integrates with the local [Task Master AI](https://github.com/cyanheads/task-master-ai) project when detected in the filesystem.

## Project Goals

1. Create a TUI application that seamlessly integrates Task Master AI functionality
2. Provide real-time task monitoring and execution through an intuitive interface
3. Maintain the elegant design principles of Crush while extending for task management
4. Enable developers to manage complex task hierarchies without leaving the terminal

## Core Features

### 1. Task Master Integration

**Must Have:**
- Auto-detect `.taskmaster` directory in current working directory or parent directories
- Load and display task hierarchy from `.taskmaster/tasks/tasks.json`
- Real-time sync with task file changes using file watchers
- Display task metadata: ID, title, status, priority, dependencies
- Support all task statuses: pending, in-progress, done, deferred, cancelled, blocked
- Execute task-master CLI commands directly from the TUI

**Should Have:**
- Gracefully degrade when `.taskmaster` is not found (show helpful setup instructions)
- Cache task data to reduce file I/O
- Validate task structure on load and show warnings for inconsistencies

**Nice to Have:**
- Support multiple Task Master projects with quick switching
- Task search and filtering across large task sets

### 2. Task Display & Navigation

**Must Have:**
- Tree view showing task hierarchy with visual indentation
- Color-coded status indicators (pending=yellow, in-progress=blue, done=green, etc.)
- Display task ID, title, and status in list view
- Keyboard navigation (arrows, vim keys, tab/shift-tab)
- Expand/collapse task subtasks
- Quick jump to task by ID

**Should Have:**
- Multiple view modes: tree, flat list, kanban-style
- Show task priority with visual indicators (high=!, medium=-, low=·)
- Display dependency relationships with arrows/lines
- Show task counts by status in header
- Breadcrumb navigation for deep hierarchies

**Nice to Have:**
- Minimap for large task hierarchies
- Timeline/gantt view for task dependencies
- Tag-based filtering and grouping

### 3. Task Operations

**Must Have:**
- Mark task as in-progress, done, blocked, cancelled, deferred
- View task details in side panel (description, test strategy, dependencies)
- Execute "next task" command to find next available task
- Show task subtasks and navigate into them
- Update task status via keyboard shortcuts

**Should Have:**
- Add new tasks inline
- Update task description and details
- Add/remove dependencies
- Reorder tasks via drag-and-drop (keyboard-based)
- Bulk status updates for multiple selected tasks

**Nice to Have:**
- Task templates for common patterns
- Duplicate tasks with modifications
- Archive completed tasks
- Undo/redo for task operations

### 4. Task Execution & Monitoring

**Must Have:**
- Execute task-master CLI commands from TUI
- Display command output in dedicated log panel
- Show real-time execution status
- Support for long-running operations (expand, analyze-complexity, research)
- Cancel running operations with Ctrl+C
- Persist execution logs to `.taskmaster/logs/tui-session.log`

**Should Have:**
- Queue multiple commands for sequential execution
- Show progress indicators for AI operations
- Automatically update task view when commands complete
- Command history with recall (up/down arrows)
- Show estimated time remaining for known operations

**Nice to Have:**
- Parallel execution of independent commands
- Integrated terminal for custom commands
- Script/macro recording and playback

### 5. Configuration & Preferences

**Must Have:**
- Load configuration from `.taskmaster/config.json` for AI model settings
- Support environment variables for API keys
- Configurable key bindings (load from `configs/default.json`)
- Persist TUI state between sessions (`.taskmaster/tui-state.json`)

**Should Have:**
- Theme customization (colors, fonts, borders)
- Configurable views and layouts
- Keyboard shortcut customization
- Export/import TUI preferences

**Nice to Have:**
- Multiple config profiles (personal, work, project-specific)
- Live config reload without restart
- Integration with git for config versioning

### 6. Help & Documentation

**Must Have:**
- In-app help panel showing keyboard shortcuts
- Status bar with context-sensitive hints
- Error messages with actionable suggestions
- Link to Task Master documentation

**Should Have:**
- Tutorial/onboarding flow for first-time users
- Command palette (Ctrl+P) for quick access to all features
- Searchable help documentation
- Tips of the day or contextual suggestions

**Nice to Have:**
- Interactive tutorial mode
- Video tutorials embedded in help
- AI-assisted help (query GPT for Task Master questions)

## Technical Architecture

### Technology Stack

- **Language:** Go 1.25+
- **TUI Framework:** Bubble Tea (charmbracelet/bubbletea)
- **Component Library:** Bubbles (charmbracelet/bubbles)
- **Styling:** Lipgloss (charmbracelet/lipgloss)
- **File Watching:** fsnotify
- **CLI Framework:** Cobra
- **Task Master Integration:** Direct CLI invocation + JSON parsing

### Project Structure

```
tm-tui/
├── cmd/
│   └── tm-tui/
│       └── main.go                 # Application entry point
├── internal/
│   ├── cli/
│   │   └── root.go                 # Cobra root command
│   ├── config/
│   │   ├── config.go              # Config loading and management
│   │   └── watcher.go             # File watcher for live reload
│   ├── taskmaster/
│   │   ├── service.go             # Task Master integration layer
│   │   ├── types.go               # Task data structures
│   │   └── cli.go                 # CLI command execution
│   ├── ui/
│   │   ├── app.go                 # Main Bubble Tea application
│   │   ├── layout.go              # View layout and rendering
│   │   ├── keymap.go              # Keyboard bindings
│   │   ├── menu.go                # Menu/task list component
│   │   ├── details_panel.go       # Task details view
│   │   ├── log_panel.go           # Command output log
│   │   └── help_panel.go          # Help and shortcuts
│   └── executor/
│       ├── service.go             # Command execution service
│       └── state.go               # Execution state management
├── configs/
│   └── default.json               # Default TUI configuration
├── go.mod
├── go.sum
└── README.md
```

### Key Components

#### 1. Task Master Service (`internal/taskmaster/`)

**Responsibilities:**
- Detect and load `.taskmaster` directory
- Parse `tasks.json` and build in-memory task tree
- Watch for file changes and reload tasks
- Execute task-master CLI commands
- Parse CLI output and update task state
- Validate task structure and dependencies

**Key Types:**
```go
type Task struct {
    ID           string      `json:"id"`
    Title        string      `json:"title"`
    Description  string      `json:"description"`
    Status       TaskStatus  `json:"status"`
    Priority     Priority    `json:"priority"`
    Dependencies []string    `json:"dependencies"`
    Details      string      `json:"details"`
    TestStrategy string      `json:"testStrategy"`
    Subtasks     []Task      `json:"subtasks"`
}

type TaskStatus string
const (
    StatusPending     TaskStatus = "pending"
    StatusInProgress  TaskStatus = "in-progress"
    StatusDone        TaskStatus = "done"
    StatusDeferred    TaskStatus = "deferred"
    StatusCancelled   TaskStatus = "cancelled"
    StatusBlocked     TaskStatus = "blocked"
)

type Priority string
const (
    PriorityHigh   Priority = "high"
    PriorityMedium Priority = "medium"
    PriorityLow    Priority = "low"
)
```

#### 2. UI Application (`internal/ui/`)

**Responsibilities:**
- Implement Bubble Tea Model interface
- Handle user input and keyboard navigation
- Render task tree and panels
- Coordinate between components
- Manage view state and layout

**Key Model Structure:**
```go
type Model struct {
    tasks          []taskmaster.Task
    selectedIndex  int
    expandedNodes  map[string]bool
    viewMode       ViewMode
    focusedPanel   Panel
    detailsPanel   DetailsPanel
    logPanel       LogPanel
    helpPanel      HelpPanel
    executor       *executor.Service
    taskMaster     *taskmaster.Service
    config         *config.Config
    width, height  int
    err            error
}
```

#### 3. Executor Service (`internal/executor/`)

**Responsibilities:**
- Execute task-master CLI commands in background
- Stream command output to log panel
- Handle command cancellation
- Maintain execution history
- Provide execution status

**Key Interface:**
```go
type Service interface {
    Execute(cmd string, args ...string) error
    Cancel() error
    GetOutput() <-chan string
    IsRunning() bool
    GetHistory() []Command
}
```

#### 4. Config Management (`internal/config/`)

**Responsibilities:**
- Load TUI configuration from JSON
- Watch for config changes
- Validate config values
- Provide default values
- Persist state between sessions

### Data Flow

1. **Initialization:**
   - Load config from `configs/default.json`
   - Detect `.taskmaster` directory
   - Load tasks from `tasks.json`
   - Initialize file watchers
   - Start Bubble Tea program

2. **Task Updates:**
   - User selects task → Update selected index
   - User changes status → Execute CLI command → Wait for completion → Reload tasks
   - File watcher detects change → Reload tasks → Update view

3. **Command Execution:**
   - User triggers command → Executor.Execute()
   - Stream output to log panel → Update UI
   - Command completes → Reload tasks if needed
   - Log to session file

4. **View Updates:**
   - Task data changes → Rebuild view tree
   - User navigates → Update selection and scroll
   - Layout changes → Recalculate panel sizes
   - Redraw UI with Bubble Tea

## User Experience

### Primary User Flows

#### Flow 1: Start Working on Next Task

1. User opens TUI: `tm-tui`
2. TUI detects `.taskmaster` directory and loads tasks
3. User presses `n` (next task)
4. TUI executes `task-master next` and displays result
5. User views task details in side panel
6. User presses `i` (in-progress) to mark task as started
7. TUI executes `task-master set-status --id=X --status=in-progress`
8. Task status updates in view
9. User exits to terminal to implement task
10. User returns to TUI and presses `d` (done) when complete

#### Flow 2: Explore Task Hierarchy

1. User opens TUI with large task list
2. User navigates with arrow keys or vim keys (j/k)
3. User presses Enter to expand/collapse subtasks
4. User presses `/` to search for specific task
5. TUI filters task list based on search query
6. User selects task and views details with `?`
7. User sees task dependencies and related tasks
8. User presses Esc to return to main view

#### Flow 3: Manage Task Status

1. User views task list filtered by status (pending)
2. User selects multiple tasks with Space
3. User presses `b` to mark all as blocked
4. TUI prompts for confirmation
5. User confirms with Enter
6. TUI executes bulk status update commands
7. Task list updates to reflect new statuses
8. User views log panel (`l`) to see command history

### Keyboard Shortcuts

| Key | Action | Context |
|-----|--------|---------|
| `↑/k` | Move up | Task list |
| `↓/j` | Move down | Task list |
| `←/h` | Collapse task | Expanded task |
| `→/l` | Expand task | Task with subtasks |
| `Enter` | Toggle expand | Task with subtasks |
| `Space` | Select/deselect | Task list |
| `n` | Next task | Anywhere |
| `i` | Mark in-progress | Selected task |
| `d` | Mark done | Selected task |
| `b` | Mark blocked | Selected task |
| `c` | Mark cancelled | Selected task |
| `f` | Mark deferred | Selected task |
| `p` | Mark pending | Selected task |
| `?` | Show help | Anywhere |
| `/` | Search | Task list |
| `:` | Command mode | Anywhere |
| `Tab` | Switch panel | Multi-panel view |
| `Ctrl+C` | Cancel/quit | Anywhere |
| `L` | Toggle log panel | Anywhere |
| `D` | Toggle details panel | Anywhere |

### Visual Design

**Layout:**
```
┌─────────────────────────────────────────────────────────────────────────┐
│ Task Master TUI - Project: MyProject        [10 pending] [2 in-progress] │
├─────────────────────────────────────────────────────────────────────────┤
│                    │                                                      │
│  Task List         │  Task Details                                       │
│                    │                                                      │
│  1.0 Setup         │  ID: 1.2                                            │
│  1.1 Init Project  │  Title: Implement user authentication               │
│  1.2 Configure ›   │  Status: pending                                    │
│                    │  Priority: high                                     │
│  2.0 Features      │  Dependencies: 1.1                                  │
│  2.1 Auth System ● │                                                      │
│                    │  Description:                                       │
│  3.0 Testing       │  Set up JWT-based authentication system with       │
│  3.1 Unit Tests    │  bcrypt password hashing...                        │
│                    │                                                      │
│                    │  Test Strategy:                                     │
│                    │  - Unit tests for auth functions                    │
│                    │  - Integration tests for login flow                 │
│                    │                                                      │
├─────────────────────────────────────────────────────────────────────────┤
│ Log Output                                                                │
│ $ task-master next                                                        │
│ Next available task: 2.1 (Auth System)                                    │
│                                                                           │
├─────────────────────────────────────────────────────────────────────────┤
│ ↑/↓: Navigate │ Enter: Expand │ n: Next │ i: In Progress │ ?: Help │ q: Quit │
└─────────────────────────────────────────────────────────────────────────┘
```

**Color Scheme:**
- Pending: Yellow (#FFD700)
- In Progress: Blue (#4169E1)
- Done: Green (#32CD32)
- Blocked: Red (#DC143C)
- Deferred: Gray (#808080)
- Cancelled: Dark Gray (#404040)
- High Priority: Bold + Red exclamation
- Medium Priority: Normal
- Low Priority: Dimmed

**Status Indicators:**
- `●` In Progress (blue circle)
- `✓` Done (green checkmark)
- `✗` Cancelled (red X)
- `⊘` Blocked (red no entry)
- `⊃` Deferred (gray bracket)
- `○` Pending (hollow circle)

## Success Metrics

### MVP Success Criteria

1. **Functional Completeness:**
   - All "Must Have" features implemented
   - Can load and display tasks from `.taskmaster`
   - Can execute basic task-master commands
   - Stable and crash-free for 1-hour session

2. **User Experience:**
   - Users can find next task in <5 seconds
   - Task status updates appear within 2 seconds
   - Keyboard navigation feels responsive (<100ms)
   - No confusing UI states or unclear feedback

3. **Integration:**
   - Works with existing Task Master projects without modification
   - Respects Task Master's task structure and conventions
   - Logs don't interfere with Task Master CLI logs

### Long-term Success Metrics

- 80%+ of Task Master users prefer TUI over CLI for daily workflow
- Average task completion time reduced by 20%
- 90%+ uptime with no data corruption issues
- Community contributions for new features and themes

## Development Phases

### Phase 1: MVP (Weeks 1-2)

**Goals:**
- Basic TUI with task list display
- Task Master integration (load tasks, detect directory)
- Core navigation (arrows, expand/collapse)
- Essential status updates (pending → in-progress → done)
- Simple command execution and log output

**Deliverables:**
- Working TUI that can display tasks
- Ability to mark tasks as in-progress and done
- Help panel with keyboard shortcuts
- Basic error handling

### Phase 2: Enhanced Features (Weeks 3-4)

**Goals:**
- Task details panel with full metadata
- Search and filtering
- All status transitions (blocked, deferred, cancelled)
- Command history and cancellation
- File watching for live updates

**Deliverables:**
- Multi-panel layout with details and logs
- Comprehensive keyboard shortcuts
- Robust command execution with error handling
- Real-time task sync

### Phase 3: Polish & Optimization (Weeks 5-6)

**Goals:**
- Performance optimization for large task sets
- Theme and color customization
- Configuration management
- Documentation and tutorials
- Testing and bug fixes

**Deliverables:**
- Stable 1.0 release
- User documentation
- Example configurations
- CI/CD pipeline

### Phase 4: Advanced Features (Post-1.0)

**Goals:**
- Multiple view modes (kanban, timeline)
- Advanced filtering and grouping
- Task templates and bulk operations
- Integrated terminal
- Plugin system

## Risk Assessment

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Task Master breaking changes | High | Medium | Pin to stable Task Master version, add version check |
| Performance issues with 1000+ tasks | Medium | Medium | Implement virtual scrolling, lazy loading |
| File watching conflicts | Low | Low | Use debouncing, detect external changes |
| Cross-platform compatibility | Medium | Low | Test on Linux, macOS, Windows regularly |

### User Experience Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Learning curve too steep | High | Medium | Excellent onboarding, in-app help, tutorials |
| Feature overload | Medium | Medium | Phased rollout, progressive disclosure |
| Keyboard shortcut conflicts | Low | High | Allow customization, use standard conventions |

## Open Questions

1. **Task Editing:** Should users be able to edit task descriptions directly in TUI, or always use CLI/file edits?
   - **Recommendation:** Start with view-only, add inline editing in Phase 3

2. **Multiple Projects:** How should TUI handle multiple Task Master projects in parent directories?
   - **Recommendation:** Use closest `.taskmaster` directory, add project switcher in Phase 4

3. **AI Integration:** Should TUI invoke Task Master AI commands (parse-prd, analyze-complexity) directly?
   - **Recommendation:** Yes, but show clear warnings about long operations and API usage

4. **Concurrency:** Should TUI allow multiple commands to run in parallel?
   - **Recommendation:** Sequential only for MVP, add parallel execution in Phase 4 with mutex guards

5. **State Persistence:** What TUI state should be saved between sessions?
   - **Recommendation:** Expanded nodes, last selected task, view mode, panel sizes

## Appendix

### Related Links

- [Crush TUI GitHub](https://github.com/charmbracelet/crush)
- [Task Master AI GitHub](https://github.com/cyanheads/task-master-ai)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Styling Guide](https://github.com/charmbracelet/lipgloss)

### Glossary

- **TUI:** Terminal User Interface
- **Task Master:** Task Master AI project management system
- **Bubble Tea:** Go TUI framework by Charm
- **Subtask:** Child task in task hierarchy (e.g., 1.1, 1.2 are subtasks of 1)
- **Task ID:** Hierarchical identifier (e.g., "1.2.3")
- **PRD:** Product Requirements Document

---

**Document Version:** 1.0  
**Last Updated:** 2025-12-09  
**Author:** Task Master AI + Claude  
**Status:** Draft - Ready for Implementation
