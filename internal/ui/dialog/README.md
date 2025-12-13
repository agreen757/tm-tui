# Dialog Component Framework

This package provides a comprehensive dialog framework for building rich terminal user interfaces with the Bubble Tea TUI library. It includes modal, form, list, confirmation, and progress dialog types with full keyboard navigation and focus management.

## Features

- Multiple dialog types: Modal, Form, List, Confirmation, Progress
- Z-index stacking for nested dialogs
- Focus trapping to ensure keyboard events go to the active dialog
- Automatic centering and resize handling
- Consistent keyboard navigation across dialog types
- Accessible design with clear focus indicators
- Themeable styling through DialogStyle

## Dialog Types

### Modal Dialog

A simple modal dialog with customizable content.

```go
content := dialog.NewSimpleModalContent("This is a simple modal dialog.")
modal := dialog.NewModalDialog("Modal Title", 60, 10, content)
```

### Form Dialog

A dialog with various form fields (text inputs, checkboxes, radio buttons).

```go
fields := []dialog.FormField{
    dialog.NewTextField("Name", "Enter your name", true),  // Required field
    dialog.NewTextField("Email", "Enter your email", false),
    dialog.NewCheckboxField("Subscribe", false),
    dialog.NewRadioGroupField("Plan", []string{"Free", "Basic", "Premium"}, 0),
}
form := dialog.NewFormDialog("Form Dialog", 70, 20, fields)
```

### List Dialog

A dialog with selectable items, supporting single or multi-select.

```go
items := []dialog.ListItem{
    dialog.NewSimpleListItem("Option 1", "Description for option 1"),
    dialog.NewSimpleListItem("Option 2", "Description for option 2"),
}
list := dialog.NewListDialog("List Dialog", 50, 15, items)
list.SetMultiSelect(true)  // Enable multi-select
```

### Confirmation Dialog

A simple Yes/No dialog for user confirmation.

```go
// Simple Yes/No dialog
confirm := dialog.YesNo("Confirmation", "Are you sure you want to proceed?", false)

// Danger mode (for destructive actions)
dangerConfirm := dialog.YesNo("Warning", "This action cannot be undone!", true)
```

### Progress Dialog

A progress dialog with a progress bar for long-running operations.

```go
progress := dialog.NewProgressDialog("Progress", 60, 10)
progress.SetProgress(0.5)  // 50% complete
progress.SetLabel("Processing files...")
```

## Dialog Manager

The DialogManager handles stacking, focus management, and input routing for multiple dialogs.

```go
// Create a dialog manager with terminal dimensions
manager := dialog.NewDialogManager(termWidth, termHeight)

// Add dialogs to the stack
manager.PushDialog(myDialog)

// Handle input messages
cmd := manager.HandleMsg(msg)

// Render all dialogs
dialogContent := manager.View()
```

## Positioning and Resize Handling

Dialogs automatically handle terminal resizing with intelligent positioning and graceful degradation:

### Features

- **Automatic Centering**: Dialogs are centered when first displayed
- **Bounds Checking**: Dialogs never overflow terminal bounds
- **Graceful Degradation**: Dialog sizes are reduced when terminal is too small
- **Smart Repositioning**: Dialogs are automatically repositioned when terminal resizes
- **Minimum Sizes**: Enforced minimum dialog dimensions for usability

### Usage

The dialog system automatically handles positioning and resizing. No additional configuration is needed:

```go
// Dialogs are automatically centered and positioned
manager := dialog.NewDialogManager(termWidth, termHeight)
dialog := NewMyDialog("Title", 50, 20)
manager.AddDialog(dialog, nil)

// Terminal resize is automatically handled
msg := tea.WindowSizeMsg{Width: newWidth, Height: newHeight}
manager.HandleMsg(msg)  // Dialogs are repositioned automatically
```

### Custom Positioning

For advanced use cases, dialogs can be positioned with custom strategies:

```go
// Manual positioning with bounds checking
pos := dialog.PositionDialogInBounds(
    termWidth, termHeight,
    dialogWidth, dialogHeight,
    dialog.StrategyCenter,  // Or StrategyTopCenter, StrategyTopLeft
)

dialog.SetRect(pos.Width, pos.Height, pos.X, pos.Y)
```

See [POSITIONING_README.md](POSITIONING_README.md) for detailed documentation on positioning strategies and configuration.

## Keyboard Navigation

- **Tab/Shift+Tab**: Navigate between fields in forms and focusable elements
- **Arrow keys**: Navigate lists and radio options
- **Enter**: Confirm/submit/select
- **Space**: Toggle checkboxes and select list items
- **Escape**: Cancel/close dialog

## Message Types

The dialog framework uses Bubble Tea's message system to communicate dialog actions:

- **FormSubmitMsg**: Sent when a form is submitted
- **ListSelectionMsg**: Sent when a list item is selected
- **ConfirmationMsg**: Sent with the result of a confirmation dialog
- **ProgressUpdateMsg**: Used to update progress dialog state
- **ProgressCompleteMsg**: Sent when progress reaches 100%
- **ProgressCancelMsg**: Sent when a progress dialog is canceled

## Running the Demo

A demo application showcases all dialog types and their interactions:

```bash
go run internal/ui/dialog/demo/main/main.go
```

## Running Tests

Comprehensive tests verify dialog behavior, focus management, and keyboard navigation:

```bash
./run_tests.sh
```

## Implementation Details

- **BaseDialog**: Core dialog functionality (position, dimensions, borders)
- **BaseFocusableDialog**: Adds focus management for dialogs with multiple elements
- **DialogManager**: Manages dialog stack, z-index, and event routing
- **DialogStyle**: Centralizes styling for consistent appearance

## Accessibility Features

- Clear focus indicators (highlighting, bold, underlined)
- Keyboard shortcuts for common actions
- Consistent navigation patterns across dialog types
- High-contrast styling options

## Usage in Application

To integrate dialogs in your Bubble Tea application:

1. Create a DialogManager in your Model
2. Push dialogs as needed based on user actions
3. Route tea.Msg events to the dialog manager in your Update function
4. Render dialog content in your View function when dialogs are active