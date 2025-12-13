# Dialog Positioning and Resize Handling

## Overview

The dialog system now includes comprehensive positioning and resize handling utilities to ensure dialogs remain properly positioned and visible when the terminal is resized. This implements graceful degradation for extremely small terminal sizes while maintaining optimal dialog placement.

## Key Features

### 1. **Dialog Positioning Utility** (`positioning.go`)

Provides functions for calculating optimal dialog positions within terminal bounds:

- **`PositionDialogInBounds()`** - Centers dialogs and applies bounds checking
- **`PositionDialogInBoundsWithConfig()`** - Advanced positioning with custom configuration
- **`DegradeDialogSize()`** - Gracefully reduces dialog size when terminal is too small
- **`ClampDialogPosition()`** - Ensures dialog position stays within terminal bounds
- **`IsDialogFullyVisible()`** - Checks if a dialog is fully visible in the terminal
- **`RepositionDialogIfNeeded()`** - Determines if a dialog needs repositioning after resize
- **`CalculateOptimalDialogSize()`** - Calculates appropriate dialog dimensions for content
- **`IsTerminalTooSmall()`** - Detects when terminal is too small for normal rendering

### 2. **DialogManager Enhancements** (`dialog.go`)

Enhanced `DialogManager` with:

- **`SetTerminalSize()`** - Updates terminal dimensions and repositions dialogs as needed
- **`repositionDialogsForResize()`** - Internal method for repositioning all dialogs
- **Improved `HandleMsg()`** - Better terminal resize event handling
- **Improved `AddDialog()`** - Uses positioning utilities when adding dialogs

### 3. **Positioning Strategies**

Dialogs can be positioned using different strategies:

```go
const (
    StrategyCenter    // Center dialog in terminal (default)
    StrategyTopCenter // Position at top-center
    StrategyTopLeft   // Position at top-left
)
```

### 4. **Configuration**

Customize positioning behavior with `PositioningConfig`:

```go
config := PositioningConfig{
    MinDialogWidth:  20,  // Minimum dialog width
    MinDialogHeight: 6,   // Minimum dialog height
    Padding:         1,   // Padding around dialogs
}
```

## How It Works

### On Dialog Creation

When a dialog is added to the manager:

1. Dialog receives initial dimensions
2. `AddDialog()` calls positioning utility
3. Dialog is centered using `StrategyCenter`
4. Final position is clamped to terminal bounds

### On Terminal Resize

When a `WindowSizeMsg` is received:

1. `SetTerminalSize()` updates manager dimensions
2. For each dialog in the stack:
   - Check if current position is still valid
   - If not, recalculate position using positioning utilities
   - Apply bounds checking
3. Dialogs are degraded if they exceed terminal bounds

### Graceful Degradation

When terminal is too small:

1. Dialog size is reduced to fit within bounds
2. Minimum size constraints are respected
3. Padding is adjusted if needed
4. Dialog remains centered or repositioned as needed

## Examples

### Basic Dialog Positioning

```go
dm := dialog.NewDialogManager(100, 30)

// Add a dialog - it will be automatically centered
dialog := NewSomeDialog("Title", 50, 20)
dm.AddDialog(dialog, nil)

// Dialog is now at position (25, 5) - centered in 100x30 terminal
```

### Handling Terminal Resize

```go
// When terminal resizes (automatically called by Bubble Tea)
msg := tea.WindowSizeMsg{Width: 80, Height: 24}
dm.HandleMsg(msg)

// All dialogs are automatically repositioned and resized as needed
```

### Custom Positioning

```go
// Position a dialog with custom strategy
pos := dialog.PositionDialogInBounds(
    termWidth, termHeight,
    dialogWidth, dialogHeight,
    dialog.StrategyTopCenter,
)

// Apply the position
myDialog.SetRect(pos.Width, pos.Height, pos.X, pos.Y)
```

## Testing

Comprehensive test coverage in:

- `positioning_test.go` - Tests for all positioning utilities
- `dialog_manager_test.go` - Tests for DialogManager resize handling

Test scenarios include:

- Normal terminal sizes
- Dialogs exceeding terminal bounds
- Extreme terminal shrinkage
- Multiple dialogs with simultaneous resize
- Dialog growth after shrinkage
- Positioning accuracy and bounds checking

### Running Tests

```bash
# Run all positioning tests
go test -v ./internal/ui/dialog -run "Positioning"

# Run dialog manager resize tests
go test -v ./internal/ui/dialog -run "DialogManager"

# Run all dialog tests
go test -v ./internal/ui/dialog
```

## Bounds Checking

The system ensures dialogs never overflow terminal bounds:

1. **Horizontal bounds**: `x >= 0 && x + width <= termWidth`
2. **Vertical bounds**: `y >= 0 && y + height <= termHeight`
3. **Clamping**: Positions are automatically adjusted if they would overflow

## Minimum Sizes

Dialogs have minimum size constraints:

- **Minimum width**: 20 characters
- **Minimum height**: 6 lines

These are enforced even on extremely small terminals, but dialogs will degrade gracefully.

## Performance Considerations

- Positioning calculations are O(1)
- Resize handling updates only affected dialogs
- Duplicate resize messages are skipped (checked by `SetTerminalSize()`)
- No allocations in hot paths

## Edge Cases Handled

1. **Negative terminal dimensions** - Treated as zero
2. **Dialog larger than terminal** - Dialog is reduced to fit
3. **Position with negative coordinates** - Clamped to zero
4. **Very small terminals** - Dialog degraded with minimum size enforcement
5. **Rapid resizes** - Duplicate calls are skipped for efficiency

## Integration with Existing Code

The positioning system integrates seamlessly with:

- All existing dialog types (Modal, Form, List, etc.)
- Bubble Tea's `WindowSizeMsg`
- Existing dialog callbacks and result handling
- Theme and style system

No breaking changes to the dialog API.

## Future Enhancements

Possible improvements:

1. **Draggable dialogs** - Use positioning for manual dialog movement
2. **Dialog animations** - Smooth transitions when repositioning
3. **Responsive dialogs** - Adaptive layouts for different terminal sizes
4. **Dialog persistence** - Remember user-preferred positions
5. **Accessibility** - Ensure adequate whitespace for screen readers

## Files Modified/Created

- **Created**: `internal/ui/dialog/positioning.go` - Core positioning utilities
- **Created**: `internal/ui/dialog/positioning_test.go` - Comprehensive positioning tests
- **Created**: `internal/ui/dialog/dialog_manager_test.go` - DialogManager resize tests
- **Modified**: `internal/ui/dialog/dialog.go` - Enhanced DialogManager and AddDialog methods
- **Removed**: `internal/ui/dialog/dialog_export.go` - Replaced with full implementation

## References

Related code locations:

- Dialog interface: `dialog.go:62-90`
- DialogManager struct: `dialog.go:376-385`
- SetTerminalSize method: `dialog.go:477-500`
- AddDialog method: `dialog.go:408-434`
- PositionDialogInBounds: `positioning.go:93-147`
- DegradeDialogSize: `positioning.go:149-192`
