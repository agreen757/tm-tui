# Task Master TUI Architecture

This document describes the architecture of the Task Master TUI, focusing particularly on the dialog system and keyboard handling.

## Overview

Task Master TUI is built with the following key technologies:

- **Go** (1.25.0) - Core language
- **Bubble Tea** (1.3.10) - TUI framework using The Elm Architecture
- **Lipgloss** (1.1.0) - Terminal styling
- **Bubbles** (0.21.0) - UI components (lists, viewports, textinput)

The application follows the Model-View-Update (MVU) pattern as implemented in Bubble Tea:

1. **Model**: Central state container (`Model` struct)
2. **Update**: State transitions through message handling
3. **View**: Pure rendering functions

## Core Components

### Model

The `Model` structure in `internal/ui/app.go` serves as the central state container for the entire application. It contains:

- Service references (task service, executor service, etc.)
- UI state (selected task, panels, viewports, etc.)
- View modes and filters
- Dialog state

### State Management

State transitions happen through the `Update` method, which processes messages like:

- `TasksLoadedMsg`: Initial tasks loaded
- `TasksReloadedMsg`: Tasks reloaded from file changes
- `tea.KeyMsg`: Keyboard input
- `tea.WindowSizeMsg`: Terminal resizing

### Keyboard Handling

Keyboard shortcuts are defined in `internal/ui/keymap.go`:

- Default bindings in `DefaultKeyMap()`
- Configuration-based overrides in `NewKeyMap()`
- Mode-specific handling in the `Update` method

## Dialog System

The dialog system uses a stack-based approach to manage multiple dialog layers.

### DialogState

```go
// DialogType represents different types of dialogs
type DialogType int

const (
    DialogModal DialogType = iota
    DialogForm
    DialogList
    DialogConfirm
    DialogProgress
)

// DialogState represents the state of a dialog
type DialogState struct {
    Type           DialogType
    Title          string
    Content        string
    Width          int
    Height         int
    ShowBorder     bool
    ShowCloseButton bool
    
    // For form dialogs
    Fields         []FormField
    ActiveFieldIdx int
    
    // For list dialogs
    Items          []ListItem
    SelectedIndex  int
    
    // For confirm dialogs
    ConfirmText    string
    CancelText     string
    
    // For progress dialogs
    Progress       float64
    ProgressText   string
    
    // Dialog result
    Result         interface{}
    Completed      bool
}
```

The `Model` struct contains a dialog stack:

```go
type Model struct {
    // ... existing fields
    
    // Dialog stack (last item is top-most)
    DialogStack    []DialogState
}
```

### Dialog Rendering

Dialogs are rendered on top of the main UI in the View method:

1. Render the base UI
2. If dialog stack is not empty, render the topmost dialog
3. Support different dialog types with type-specific rendering functions
4. Center dialogs in the available space

### Dialog Keyboard Handling

Keyboard handling in dialogs takes precedence over global shortcuts:

1. Check if dialog stack is not empty
2. Handle dialog-specific keys (Tab, Enter, Escape, etc.)
3. If no dialog is open, handle global shortcuts

## Reserved Keyboard Shortcuts

The following keyboard shortcuts are reserved for dialog-driven features:

- **Alt+P**: Project dialog
- **Alt+C**: Complexity analysis dialog
- **Alt+E**: Execution dialog
- **Alt+D**: Dependencies management dialog
- **Ctrl+Shift+A**: Add tag context dialog
- **Ctrl+Shift+M**: Tag context management dialog
- **Ctrl+T**: Task template selection dialog

## Layout System

The UI layout is managed through the `calculateLayout` function in `internal/ui/layout.go`:

1. Calculate available space based on terminal size
2. Adjust panel dimensions based on visibility settings
3. Update viewport dimensions

Panels include:
- Task List panel
- Task Details panel
- Log panel
- Status bar
- Header

## Dialog Stack Example

Here's how dialog stacking works:

```go
// Example: Open a confirmation dialog
func (m *Model) openConfirmDialog(title, message string) {
    dialog := DialogState{
        Type: DialogConfirm,
        Title: title,
        Content: message,
        ConfirmText: "Yes",
        CancelText: "No",
        Width: 40,
        Height: 10,
        ShowBorder: true,
        ShowCloseButton: true,
    }
    
    m.DialogStack = append(m.DialogStack, dialog)
}

// Example: Open a form dialog from within a modal
func (m *Model) openNestedFormDialog() {
    dialog := DialogState{
        Type: DialogForm,
        Title: "Edit Task",
        Fields: []FormField{
            {Label: "Title", Value: m.selectedTask.Title},
            {Label: "Status", Value: m.selectedTask.Status},
        },
        Width: 60,
        Height: 15,
        ShowBorder: true,
        ShowCloseButton: true,
    }
    
    m.DialogStack = append(m.DialogStack, dialog)
}

// Example: Close the topmost dialog
func (m *Model) closeTopDialog() {
    if len(m.DialogStack) > 0 {
        m.DialogStack = m.DialogStack[:len(m.DialogStack)-1]
    }
}
```

## Best Practices

When implementing dialogs:

1. **State isolation**: Keep dialog-specific state in the DialogState
2. **Keyboard focus**: Ensure keyboard focus is managed correctly
3. **Z-ordering**: Handle stackable dialogs properly
4. **Accessibility**: Provide clear visual indicators for focused elements
5. **Consistency**: Use consistent styling and keyboard shortcuts

## Future Enhancements

Planned improvements to the dialog system:

1. Support for nested dialogs with proper focus management
2. Animation support for dialog transitions
3. Support for different dialog sizes (full-screen, centered, etc.)
4. More form field types (dropdowns, checkboxes, etc.)
5. Context-sensitive help in dialogs
