# Task Master TUI

An interactive terminal user interface for [Task Master AI](https://github.com/cyanheads/task-master-ai), forked from [Crush](https://github.com/charmbracelet/crush).

## Overview

Task Master TUI provides a beautiful, keyboard-driven interface for managing development tasks, viewing task hierarchies, and executing Task Master commands without leaving your terminal. This tool seamlessly integrates with Task Master AI to provide a rich terminal experience for project management.

## Features

- 🎯 **Task Management**: View and navigate task hierarchies with ease
- 🔄 **Real-time Sync**: Automatically updates when task files change using fsnotify
- ⌨️ **Keyboard-driven**: Full navigation and control via keyboard shortcuts
- 🎨 **Beautiful UI**: Built with Bubble Tea and Lipgloss for a polished experience
- 🔍 **Search & Filter**: Quickly find tasks by ID, title, status, or content
- 🚀 **Task Execution**: Run Task Master commands directly from the TUI
- 💡 **Context-sensitive Help**: Dynamic help panels and status bar hints
- ⚙️ **Customizable**: Configure through simple JSON configuration

## Prerequisites

- Go 1.23 or later
- [Task Master AI](https://github.com/cyanheads/task-master-ai) installed globally (`npm i -g task-master-ai`)
- A Task Master project (`.taskmaster` directory with tasks)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/adriangreen/tm-tui.git
cd tm-tui

# Build and install
make install

# Or just build
make build
```

### Using Go Install

```bash
go install github.com/adriangreen/tm-tui/cmd/tm-tui@latest
```

## Usage

Navigate to a directory with a Task Master project (containing `.taskmaster` directory) and run:

```bash
tm-tui
```

### Keyboard Shortcuts

- **Navigation:**
  - `↑/k` - Move up
  - `↓/j` - Move down
  - `←/h` - Navigate left 
  - `→/l` - Navigate right
  - `Tab` - Switch between panels
  - `Enter` - Select item / expand task
  - `Esc` - Back / close panel

- **Task Management:**
  - `n` - Jump to next available task
  - `e` - Edit selected task
  - `d` - Mark task as done
  - `p` - Change task priority
  - `s` - Change task status

- **Search & Filter:**
  - `/` - Search tasks
  - `f` - Filter by status
  - `F` - Clear filters

- **Misc:**
  - `r` - Refresh tasks
  - `?` - Show/hide help
  - `q` - Quit

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