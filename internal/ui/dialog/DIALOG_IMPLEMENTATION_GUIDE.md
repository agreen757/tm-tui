# Dialog Implementation Guide

## Overview

This guide documents the recommended patterns for implementing the `Dialog` interface in Task Master TUI. The Dialog interface requires both `Update()` and `HandleKey()` methods, and proper implementation of both is critical to avoid duplicate logic, race conditions, and fragile state management.

## Table of Contents

1. [Dialog Interface Contract](#dialog-interface-contract)
2. [Recommended Patterns](#recommended-patterns)
3. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
4. [Pattern Selection Guide](#pattern-selection-guide)
5. [Code Review Checklist](#code-review-checklist)
6. [Testing Recommendations](#testing-recommendations)

---

## Dialog Interface Contract

The `Dialog` interface defines two key methods for handling user input:

```go
type Dialog interface {
    // Update processes Bubble Tea messages (events, window resize, etc.)
    // Returns the updated dialog instance and any commands to execute
    Update(msg tea.Msg) (Dialog, tea.Cmd)
    
    // HandleKey processes keyboard input specifically
    // Returns a DialogResult indicating action taken and any commands
    HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd)
    
    // ... other interface methods
}
```

### Method Responsibilities

**`Update(msg tea.Msg)`**
- Process non-keyboard messages (WindowSize, custom messages, etc.)
- Update component states (textinput, textarea, list, viewport)
- Forward messages to child components
- **Should NOT** contain dialog control flow logic (submit, cancel, navigation)

**`HandleKey(msg tea.KeyMsg)`**
- Handle all keyboard-driven dialog control flow
- Process special keys (Enter, Esc, Tab, Arrow keys)
- Route input to appropriate handlers
- Return appropriate `DialogResult` (Confirm, Cancel, None)

### Call Sequence

The `DialogManager` calls these methods in this order:

1. **First**: `Update(msg)` - State updates and message propagation
2. **Then**: `HandleKey(msg)` - Control flow and user action handling

This sequence is intentional - state updates happen before control decisions.

---

## Recommended Patterns

### Pattern 1: State-Only Update (Form.go Pattern)

**When to use**: Simple dialogs with textinput/textarea fields, forms, or basic input collection

**Principle**: `Update()` handles only state updates, `HandleKey()` handles ALL keyboard routing

#### Example: FormDialog

```go
// Update propagates messages to specialized handlers.
func (d *FormDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch m := msg.(type) {
    case tea.WindowSizeMsg:
        d.Center(m.Width, m.Height)
    }
    for _, handler := range d.events {
        handler(d, msg)
    }
    return d, nil
}

// HandleKey routes keyboard input to the focused element.
func (d *FormDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    focused := d.FocusedIndex()
    
    // Special handling for textarea fields
    if focused < len(d.fields) && d.fields[focused].Type == FormFieldTypeTextArea {
        switch msg.String() {
        case "tab":
            return DialogResultNone, d.FocusNext()
        case "shift+tab":
            return DialogResultNone, d.FocusPrev()
        case "esc":
            return d.HandleBaseKey(msg)
        default:
            // All other keys go to textarea
            return d.handleFieldKey(focused, msg)
        }
    }
    
    // For non-textarea fields, use normal base handling
    if result, cmd := d.HandleBaseFocusableKey(msg); result != DialogResultNone {
        return result, cmd
    }
    
    // Handle field-specific input
    return d.handleFieldKey(focused, msg)
}
```

**Key Characteristics**:
- ✅ `Update()` only handles WindowSize and event propagation
- ✅ `HandleKey()` contains ALL keyboard logic
- ✅ No duplicate code between methods
- ✅ Clear separation of concerns

---

### Pattern 2: Delegation Pattern (SubtaskEditDialog Pattern)

**When to use**: Complex dialogs with multiple modes, navigation states, or sophisticated input handling

**Principle**: Single function handles all logic, both methods delegate to it

#### Example: SubtaskEditDialog

```go
// Update implements Dialog interface
func (d *SubtaskEditDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    if !d.focused {
        return d, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        return d.handleKeyMsg(msg)  // Delegate to shared handler
    }

    return d, nil
}

// HandleKey implements Dialog interface
func (d *SubtaskEditDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    _, cmd := d.handleKeyMsg(msg)  // Delegate to shared handler
    return DialogResultNone, cmd
}

// handleKeyMsg is the single source of truth for keyboard logic
func (d *SubtaskEditDialog) handleKeyMsg(msg tea.KeyMsg) (Dialog, tea.Cmd) {
    // In edit mode, handle text input
    if d.editingMode {
        return d.handleEditMode(msg)
    }
    
    // In list mode, handle navigation
    return d.handleListMode(msg)
}

func (d *SubtaskEditDialog) handleListMode(msg tea.KeyMsg) (Dialog, tea.Cmd) {
    switch msg.String() {
    case "up", "k":
        if d.selectedIndex > 0 {
            d.selectedIndex--
            d.ensureVisible()
        }
        return d, nil
    
    case "down", "j":
        if d.selectedIndex < len(d.drafts)-1 {
            d.selectedIndex++
            d.ensureVisible()
        }
        return d, nil
    
    case "enter":
        if d.confirmCallback != nil {
            d.confirmCallback(d.drafts)
        }
        return d, nil
    
    case "a", "A":
        d.drafts = append(d.drafts, taskmaster.SubtaskDraft{
            Title: "New subtask",
        })
        d.selectedIndex = len(d.drafts) - 1
        d.startEditMode()
        return d, nil
    
    case "esc", "ctrl+c":
        if d.cancelCallback != nil {
            d.cancelCallback()
        }
        return d, nil
    }
    
    return d, nil
}
```

**Key Characteristics**:
- ✅ Both `Update()` and `HandleKey()` delegate to `handleKeyMsg()`
- ✅ Single source of truth for keyboard logic
- ✅ No code duplication
- ✅ Easy to maintain and test
- ✅ Supports complex state machines (edit mode vs list mode)

---

## Anti-Patterns to Avoid

### ❌ Anti-Pattern: Duplicate Key Handling

**Problem**: Same keys processed in both `Update()` and `HandleKey()` with identical logic

#### Bad Example: branch_create.go (Current Issue)

```go
// Update processes messages and updates dialog state
func (d *BranchCreateDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if d.loading {
            return d, nil
        }

        switch msg.String() {
        case "enter":  // ❌ Duplicate handling
            if !d.isValidBranchName() {
                d.errorMsg = "Invalid branch name"
                return d, nil
            }
            branchName := strings.TrimSpace(d.input.Value())
            d.loading = true
            d.errorMsg = ""
            return d, d.createBranchCmd(branchName)  // ❌ Duplicate logic

        case "esc":  // ❌ Duplicate handling
            return nil, nil
        }
    // ...
}

// HandleKey processes a key event
func (d *BranchCreateDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    result, cmd := d.HandleBaseKey(msg)
    if result != DialogResultNone {
        return result, cmd
    }

    switch msg.String() {
    case "enter":  // ❌ Same key handled here too
        if d.loading {
            return DialogResultNone, nil
        }
        if !d.isValidBranchName() {
            d.errorMsg = "Invalid branch name"
            return DialogResultNone, nil
        }
        branchName := strings.TrimSpace(d.input.Value())
        d.loading = true
        d.errorMsg = ""
        return DialogResultNone, d.createBranchCmd(branchName)  // ❌ Duplicate logic
    }
    
    return DialogResultNone, nil
}
```

**Issues**:
- ❌ Duplicate validation logic (`isValidBranchName()`)
- ❌ Duplicate command creation (`createBranchCmd()`)
- ❌ Relies on `d.loading` flag to prevent double execution (fragile)
- ❌ Changes must be made in two places
- ❌ Difficult to test and maintain
- ❌ Violates DRY (Don't Repeat Yourself) principle

### Why It's Fragile

While this pattern "works" due to the `d.loading` flag mitigation:

1. **Call sequence**:
   - User presses "enter"
   - `Update()` called first → sets `d.loading=true` → returns `createBranchCmd()`
   - `HandleKey()` called next → sees `d.loading=true` → returns `DialogResultNone, nil`
   - Command executes only once

2. **Fragility points**:
   - If `d.loading` check missing in `HandleKey()` → command executes twice
   - If `Update()` ordering changes → race condition
   - If flag state management has bugs → unpredictable behavior
   - Testing requires understanding this implicit contract

### ✅ Corrected Version (Recommended)

```go
// Update only handles component state updates
func (d *BranchCreateDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        d.Center(msg.Width, msg.Height)
    
    case branchCreateMsg:
        if msg.err != nil {
            d.loading = false
            d.errorMsg = "Error: " + msg.err.Error()
            return d, nil
        }
        
        // Success - call callback and close dialog
        if d.onComplete != nil {
            d.onComplete(msg.branchName, msg.output, nil)
        }
        return nil, nil
    }
    
    // Forward to textinput for state updates
    var cmd tea.Cmd
    d.input, cmd = d.input.Update(msg)
    
    // Real-time validation on input changes
    if _, ok := msg.(tea.KeyMsg); ok {
        d.validateAndUpdateError()
    }
    
    return d, cmd
}

// HandleKey contains ALL keyboard control flow logic
func (d *BranchCreateDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    // First check base dialog keys (ESC for cancel)
    result, cmd := d.HandleBaseKey(msg)
    if result != DialogResultNone {
        return result, cmd
    }
    
    // Don't process input while loading
    if d.loading {
        return DialogResultNone, nil
    }
    
    // Handle dialog-specific keys
    switch msg.String() {
    case "enter":
        if !d.isValidBranchName() {
            d.errorMsg = "Invalid branch name"
            return DialogResultNone, nil
        }
        
        branchName := strings.TrimSpace(d.input.Value())
        d.loading = true
        d.errorMsg = ""
        return DialogResultNone, d.createBranchCmd(branchName)
    }
    
    return DialogResultNone, nil
}
```

**Improvements**:
- ✅ `Update()` handles only state updates and message processing
- ✅ `HandleKey()` contains all keyboard control flow
- ✅ No duplicate logic
- ✅ Single source of truth for "enter" key action
- ✅ `d.loading` flag only needed in one place
- ✅ Easy to maintain and test

---

## Pattern Selection Guide

### Decision Tree

```
Does your dialog have textinput/textarea components?
│
├─ YES: Is there complex state management (modes, navigation)?
│   │
│   ├─ YES: Use **Delegation Pattern** (Pattern 2)
│   │   - Multiple modes (edit vs list)
│   │   - Complex navigation logic
│   │   - State transitions
│   │   Example: SubtaskEditDialog, ComplexityReportDialog
│   │
│   └─ NO: Use **State-Only Update Pattern** (Pattern 1)
│       - Simple forms
│       - Basic input collection
│       - Straightforward validation
│       Example: FormDialog, BranchCreateDialog (corrected)
│
└─ NO: Consider custom pattern but follow these principles:
    - Update() handles state only
    - HandleKey() handles control flow
    - No duplicate logic between methods
```

### Pattern Comparison

| Aspect | State-Only Update | Delegation |
|--------|-------------------|------------|
| **Complexity** | Low | Medium to High |
| **Best For** | Simple forms, single-mode dialogs | Complex state machines, multi-mode dialogs |
| **Maintainability** | High (clear separation) | High (single source of truth) |
| **Testability** | Easy | Moderate (more test cases) |
| **Code Duplication** | None | None |
| **Examples** | FormDialog, FileSelectionDialog | SubtaskEditDialog, ExpandEditDialog |

---

## Code Review Checklist

Use this checklist when reviewing dialog implementations:

### ✅ Update() Method

- [ ] Only processes non-keyboard messages (WindowSize, custom messages)
- [ ] Updates component states (textinput, viewport, list)
- [ ] Forwards messages to child components
- [ ] Does NOT contain dialog control flow (submit, cancel, navigation)
- [ ] Does NOT handle "enter", "esc", "tab", or other control keys
- [ ] Returns updated dialog and optional command

### ✅ HandleKey() Method

- [ ] Contains ALL keyboard control flow logic
- [ ] Handles special keys (Enter, Esc, Tab, Arrow keys)
- [ ] Calls base dialog methods (HandleBaseKey, HandleBaseFocusableKey)
- [ ] Returns appropriate DialogResult (Confirm, Cancel, None)
- [ ] No duplicate logic with Update() method
- [ ] Guards against invalid states (loading, uninitialized)

### ✅ General Principles

- [ ] No duplicate code between Update() and HandleKey()
- [ ] Single source of truth for each keyboard action
- [ ] State flags (loading, focused) used consistently
- [ ] Clear separation of concerns
- [ ] Follows one of the recommended patterns
- [ ] Well-commented for complex logic

### ❌ Red Flags

- [ ] Same key processed in both methods
- [ ] Duplicate validation logic
- [ ] Duplicate command creation
- [ ] Relies on flag state to prevent double execution
- [ ] Changes require edits in multiple places
- [ ] Unclear method responsibilities

---

## Testing Recommendations

### Unit Test Structure

```go
func TestDialogUpdate(t *testing.T) {
    tests := []struct {
        name     string
        msg      tea.Msg
        wantCmd  bool
        validate func(d Dialog)
    }{
        {
            name: "WindowSize message updates size",
            msg:  tea.WindowSizeMsg{Width: 100, Height: 50},
            wantCmd: false,
            validate: func(d Dialog) {
                // Verify size updated
            },
        },
        {
            name: "Custom message processed",
            msg:  customResultMsg{data: "test"},
            wantCmd: true,
            validate: func(d Dialog) {
                // Verify state changed
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            d := NewDialog()
            updatedDialog, cmd := d.Update(tt.msg)
            
            if (cmd != nil) != tt.wantCmd {
                t.Errorf("Update() cmd = %v, want %v", cmd != nil, tt.wantCmd)
            }
            
            tt.validate(updatedDialog)
        })
    }
}

func TestDialogHandleKey(t *testing.T) {
    tests := []struct {
        name       string
        key        tea.KeyMsg
        wantResult DialogResult
        wantCmd    bool
        validate   func(d Dialog)
    }{
        {
            name:       "Enter submits form",
            key:        tea.KeyMsg{Type: tea.KeyEnter},
            wantResult: DialogResultConfirm,
            wantCmd:    true,
            validate: func(d Dialog) {
                // Verify form submitted
            },
        },
        {
            name:       "Esc cancels dialog",
            key:        tea.KeyMsg{Type: tea.KeyEsc},
            wantResult: DialogResultCancel,
            wantCmd:    false,
            validate: func(d Dialog) {
                // Verify no changes made
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            d := NewDialog()
            result, cmd := d.HandleKey(tt.key)
            
            if result != tt.wantResult {
                t.Errorf("HandleKey() result = %v, want %v", result, tt.wantResult)
            }
            
            if (cmd != nil) != tt.wantCmd {
                t.Errorf("HandleKey() cmd = %v, want %v", cmd != nil, tt.wantCmd)
            }
            
            tt.validate(d)
        })
    }
}
```

### Integration Testing

Test the complete DialogManager flow:

```go
func TestDialogManagerFlow(t *testing.T) {
    // Verify Update() called before HandleKey()
    // Verify state updates properly sequence
    // Verify no double-execution of commands
}
```

### Anti-Pattern Detection Tests

```go
func TestNoDoubleExecution(t *testing.T) {
    // Ensure pressing Enter doesn't execute command twice
    // Verify loading flag prevents race conditions
}

func TestKeyHandlingConsistency(t *testing.T) {
    // Ensure same key produces same result regardless of call order
}
```

---

## Summary

### Key Principles

1. **Separation of Concerns**: Update() for state, HandleKey() for control flow
2. **Single Source of Truth**: Each action handled in one place only
3. **No Duplication**: Avoid duplicate logic between methods
4. **Clear Patterns**: Follow established patterns for consistency
5. **Testability**: Design for easy unit and integration testing

### Quick Reference

| You Want To... | Use This Pattern |
|----------------|------------------|
| Simple form with textinput | State-Only Update (Pattern 1) |
| Complex multi-mode dialog | Delegation (Pattern 2) |
| Dialog with list navigation | Delegation (Pattern 2) |
| Basic input collection | State-Only Update (Pattern 1) |

### Resources

- **Good Examples**: `form.go`, `expand_edit.go`, `file_selection.go`
- **Issue Examples**: `branch_create.go` (current), `branch_switch.go` (suspected)
- **Base Implementations**: `BaseDialog`, `BaseFocusableDialog`

---

## Appendix: DialogResult Types

```go
type DialogResult int

const (
    DialogResultNone     DialogResult = iota  // No action, continue dialog
    DialogResultConfirm                       // User confirmed (Enter)
    DialogResultCancel                        // User cancelled (Esc)
)
```

### Usage Guidelines

- **DialogResultNone**: Keep dialog open, continue processing
- **DialogResultConfirm**: Close dialog, user accepted action
- **DialogResultCancel**: Close dialog, user rejected action

Return these from `HandleKey()` based on user intent, not from `Update()`.

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-12  
**Status**: Initial Release  
**Related Tasks**: Task 3.1 (git-dialog-fix)
