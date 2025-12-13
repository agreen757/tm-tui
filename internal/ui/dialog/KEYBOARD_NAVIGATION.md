# Keyboard Navigation Standards

This document outlines the consistent keyboard navigation patterns used throughout the Task Master TUI dialogs.

## Universal Keyboard Shortcuts

All dialogs must support these fundamental shortcuts:

### Navigation
- **Tab** - Move focus to next focusable element (wraps around)
- **Shift+Tab** - Move focus to previous focusable element (wraps around)
- **Arrow Keys** - Context-dependent navigation (Up/Down for lists, Left/Right for toggles)
- **hjkl Keys** - Vi-style navigation (when arrow keys are used)

### Confirmation & Cancellation
- **Enter** - Confirm selection/submit form
- **Esc** - Cancel dialog, go back, or close (does not submit)
- **Space** - Toggle checkbox or select radio option

## Dialog-Specific Patterns

### Form Dialogs (dialog.FormDialog)
For dialogs with multiple input fields and buttons.

**Focusable Elements**: Form fields → Buttons

**Shortcuts**:
- Tab/Shift+Tab: Navigate between fields and buttons
- Enter: Submit the selected button (when focused on it)
- Space: Toggle checkbox when on checkbox field
- Left/Right/Up/Down: Navigate radio options when on radio field
- Esc: Cancel and close without submitting

**Footer Hints**:
```
Tab:Next Field  |  Shift+Tab:Previous  |  Enter:Submit  |  Esc:Cancel
```

### List Dialogs (dialog.ListDialog)
For dialogs displaying a list of selectable items.

**Focusable Elements**: List items

**Shortcuts**:
- Up/k: Move to previous item
- Down/j: Move to next item
- Enter: Select item and confirm
- Esc: Close without selection
- Tab: Optionally move to filter/search field if available
- Shift+Tab: Move back to list from filter field

**Footer Hints**:
```
↑/↓:Navigate  |  Enter:Select  |  Esc:Close
```

### File Selection Dialogs (dialog.FileSelectionDialog)
For selecting files from the filesystem.

**Focusable Elements**: File browser

**Shortcuts**:
- Up/k: Move to previous file/directory
- Down/j: Move to next file/directory
- Left/h: Go to parent directory (Backspace also works)
- Right/l or Enter: Open directory or select file
- Esc: Cancel and close

**Footer Hints**:
```
↑/↓:Navigate  |  Enter:Open/Select  |  Backspace:Parent  |  Esc:Cancel
```

### Confirmation Dialogs (dialog.ConfirmationDialog)
For yes/no/ok dialogs.

**Focusable Elements**: Buttons

**Shortcuts**:
- Left/h/Shift+Tab: Move to previous button
- Right/l/Tab: Move to next button
- Enter: Activate selected button
- Esc: Cancel (if cancellable)
- Y: Yes shortcut (when appropriate)
- N: No shortcut (when appropriate)

**Footer Hints**:
```
←/→:Change Selection  |  Enter:Confirm  |  Esc:Cancel
```

## Implementation Guidelines

### For Dialog Creators
1. **Always set footer hints** using `dialog.SetFooterHints()` 
2. **Use BaseFocusableDialog** for multi-element dialogs
3. **Implement HandleKey** to route keyboard input appropriately
4. **Ensure Tab/Shift+Tab work** for all focusable elements

### For Dialog Implementations
- Follow the pattern in `BaseFocusableDialog.HandleBaseFocusableKey()`
- Delegate Tab/Shift+Tab to the base implementation
- Handle field-specific keys (arrow keys, space) in field handlers
- Always handle Esc for cancellation via `HandleBaseKey()`

### Footer Hint Format
Footer hints should be clear and consistent:
- Use lowercase key names (tab, enter, esc)
- Use arrow symbols (↑↓←→) where appropriate
- Separate hints with `|` 
- Keep descriptions concise (1-2 words)
- Order: Primary actions first, cancellation last

## Testing Keyboard Navigation

To verify keyboard consistency:
1. Open dialog with keyboard only (no mouse)
2. Use Tab to navigate through all elements
3. Use Shift+Tab to navigate backward
4. Verify Enter confirms and Esc cancels
5. Check footer hints are visible and accurate
6. Verify all keyboard shortcuts are responsive

## Related Files

- `focusable.go` - Base class for focusable dialogs
- `form.go` - Form dialog implementation
- `list.go` - List dialog implementation
- `confirm.go` - Confirmation dialog implementation
- `file_selection.go` - File browser implementation
