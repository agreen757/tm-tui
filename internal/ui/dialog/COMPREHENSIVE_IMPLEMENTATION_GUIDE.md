# Comprehensive Dialog Implementation Guide
## Task Master TUI - Dialog Interface Best Practices

**Version**: 1.0  
**Date**: 2026-01-12  
**Status**: Final  
**Related Tasks**: Task 3 (git-dialog-fix)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Review Findings](#review-findings)
3. [The Dialog Interface Contract](#the-dialog-interface-contract)
4. [Recommended Patterns](#recommended-patterns)
5. [Anti-Patterns & Issues Found](#anti-patterns--issues-found)
6. [Pattern Selection Decision Tree](#pattern-selection-decision-tree)
7. [Implementation Checklist](#implementation-checklist)
8. [Code Review Guidelines](#code-review-guidelines)
9. [Testing Strategy](#testing-strategy)
10. [Migration Guide](#migration-guide)
11. [Reference Materials](#reference-materials)

---

## Executive Summary

### Project Context

Task Master TUI implements a Dialog interface with two key methods: `Update()` and `HandleKey()`. A comprehensive code review identified duplicate key handling issues and established best practices for future implementations.

### Key Findings

- **Total Dialogs Reviewed**: 11 implementations
- **Correct Implementations**: 6 dialogs (55%)
- **Minor Issues**: 3 dialogs (27%) - non-standard but functional
- **Critical Issues**: 2 dialogs (18%) - duplicate key handling

### Critical Issues Identified

1. **branch_create.go** - Duplicate "enter" key handling in both methods
2. **expand_preview.go** - Duplicate handleKeyMsg() calls from both methods

### Resolution

- **Documented**: 2 recommended patterns for correct implementation
- **Identified**: Good examples to follow (commits.go, git_status.go, form.go)
- **Analyzed**: Root causes of duplicate handling issues
- **Provided**: Clear migration path for fixing problematic implementations

---

## Review Findings

### Pattern Distribution

| Pattern | Count | Status | Files |
|---------|-------|--------|-------|
| **State-Only Update** (Recommended) | 4 | ✅ Correct | commits.go, git_status.go, form.go, next_task_output.go |
| **Delegation** (Recommended) | 2 | ✅ Correct | expand_edit.go, file_selection.go |
| **Update→HandleKey Delegation** | 3 | ⚠️ Minor | branch_switch.go, complexity_report.go, task_runner_modal.go |
| **Duplicate Handling** | 2 | ❌ Fix Needed | **branch_create.go**, **expand_preview.go** |

### Files Requiring Action

| Priority | File | Issue | Effort |
|----------|------|-------|--------|
| 🔴 High | branch_create.go | Duplicate "enter" key handling | ~20 lines |
| 🔴 High | expand_preview.go | Duplicate handleKeyMsg() calls | ~10 lines |
| 🟡 Low | complexity_report.go | Redundant viewport update | ~5 lines |
| 🟡 Low | task_runner_modal.go | Document intentional cascade | Comments only |

---

## The Dialog Interface Contract

### Interface Definition

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

### Method Call Sequence

The `DialogManager` calls methods in this specific order:

```
User Presses Key
    ↓
1. Update(tea.KeyMsg) called
    ↓
2. HandleKey(tea.KeyMsg) called
    ↓
3. Commands executed
```

**This sequence is guaranteed and intentional.**

### Method Responsibilities

#### Update(msg tea.Msg)

**Primary Responsibilities**:
- Process window resize events (`tea.WindowSizeMsg`)
- Handle custom result messages (command completion)
- Update component states (textinput, textarea, list, viewport)
- Forward messages to child components

**Should NOT**:
- Contain dialog control flow logic (submit, cancel, navigation)
- Handle special keys like Enter, Esc, Tab
- Make dialog-level decisions

#### HandleKey(msg tea.KeyMsg)

**Primary Responsibilities**:
- Handle ALL keyboard-driven dialog control flow
- Process special keys (Enter, Esc, Tab, Arrow keys)
- Route input to appropriate handlers
- Return appropriate `DialogResult` (Confirm, Cancel, None)

**Should**:
- Be the single source of truth for keyboard actions
- Call base dialog methods (HandleBaseKey, HandleBaseFocusableKey)
- Contain all validation and command creation logic

---

## Recommended Patterns

### Pattern 1: State-Only Update ⭐ RECOMMENDED

**Best For**: Simple dialogs, forms, basic input collection

**Principle**: `Update()` handles only state updates, `HandleKey()` handles ALL keyboard routing

#### Example: git_status.go (Exemplar Implementation)

```go
// Update only handles state messages
func (d *GitStatusDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Handle window resize
        d.Center(msg.Width, msg.Height)
        return d, nil

    case GitStatusRefreshMsg:
        // Handle command result
        d.refreshing = false
        if msg.err != nil {
            d.errorMsg = msg.err.Error()
        } else {
            d.status = msg.status
        }
        return d, nil
    }

    return d, nil
}

// HandleKey contains ALL keyboard logic
func (d *GitStatusDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    // First check base dialog keys (ESC for cancel)
    result, cmd := d.HandleBaseKey(msg)
    if result != DialogResultNone {
        return result, cmd
    }

    // Guard against processing while loading
    if d.refreshing {
        return DialogResultNone, nil
    }

    // Handle dialog-specific keys
    switch msg.String() {
    case "r", "R":
        d.refreshing = true
        return DialogResultNone, d.refreshStatusCmd()
    }

    return DialogResultNone, nil
}
```

**Characteristics**:
- ✅ `Update()` handles NO keyboard keys
- ✅ `HandleKey()` contains ALL keyboard logic
- ✅ Clear separation of concerns
- ✅ Easy to understand and maintain
- ✅ No potential for duplicate logic

**When to Use**:
- Simple forms with textinput/textarea
- Dialogs with straightforward validation
- Single-mode dialogs
- Basic input collection

**Good Examples**: `commits.go`, `git_status.go`, `form.go`

---

### Pattern 2: Delegation ⭐ RECOMMENDED

**Best For**: Complex dialogs with multiple modes, state machines, or sophisticated input handling

**Principle**: Single function handles all logic, both methods delegate to it

#### Example: expand_edit.go (Delegation Pattern)

```go
// Update delegates to shared handler
func (d *SubtaskEditDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    if !d.focused {
        return d, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        return d.handleKeyMsg(msg)  // ← Delegates to shared function
    }

    return d, nil
}

// HandleKey also delegates to shared handler
func (d *SubtaskEditDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    _, cmd := d.handleKeyMsg(msg)  // ← Same shared function
    return DialogResultNone, cmd
}

// handleKeyMsg is the single source of truth
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
    
    case "enter":
        if d.confirmCallback != nil {
            d.confirmCallback(d.drafts)
        }
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

**Characteristics**:
- ✅ Both methods delegate to shared function
- ✅ Single source of truth for keyboard logic
- ✅ No code duplication
- ✅ Supports complex state machines
- ✅ Easy to test (test shared function once)

**When to Use**:
- Multi-mode dialogs (edit mode vs list mode)
- Complex navigation logic
- State machines with transitions
- Sophisticated input handling

**Good Examples**: `expand_edit.go`, `file_selection.go`

---

### Pattern 3: Update→HandleKey Delegation ⚠️ NON-STANDARD

**Status**: Functionally correct but non-standard

**Principle**: `Update()` delegates keyboard messages to `HandleKey()`

#### Example: branch_switch.go

```go
func (d *BranchSwitchDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        d.Center(msg.Width, msg.Height)
        return d, nil

    case tea.KeyMsg:
        // ⚠️ Calls HandleKey from Update
        result, cmd := d.HandleKey(msg)
        if result != DialogResultNone {
            return nil, cmd
        }

        // Forward remaining keys to list
        d.list, cmd = d.list.Update(msg)
        return d, cmd
    }

    return d, nil
}

func (d *BranchSwitchDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    // All keyboard logic here
}
```

**Characteristics**:
- ⚠️ Unusual pattern (non-standard)
- ✅ No duplicate logic
- ⚠️ HandleKey called twice (once by Update, once by DialogManager)
- ✅ Works due to idempotent behavior

**Verdict**: Acceptable but not recommended for new code

**Files Using This**: `branch_switch.go`, `complexity_report.go`, `task_runner_modal.go`

---

## Anti-Patterns & Issues Found

### ❌ Anti-Pattern: Duplicate Key Handling

**Problem**: Same keys processed in both `Update()` and `HandleKey()` with identical or similar logic

#### Bad Example 1: branch_create.go (Critical Issue)

```go
// Update() - Lines 67-77
func (d *BranchCreateDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":  // ❌ Handles enter
            if !d.isValidBranchName() {
                d.errorMsg = "Invalid branch name"
                return d, nil
            }
            branchName := strings.TrimSpace(d.input.Value())
            d.loading = true
            return d, d.createBranchCmd(branchName)  // ❌ Creates command
        }
    }
    // ...
}

// HandleKey() - Lines 223-237
func (d *BranchCreateDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    switch msg.String() {
    case "enter":  // ❌ Also handles enter
        if d.loading {
            return DialogResultNone, nil  // ← Fragile mitigation
        }
        if !d.isValidBranchName() {  // ❌ Duplicate validation
            d.errorMsg = "Invalid branch name"
            return DialogResultNone, nil
        }
        branchName := strings.TrimSpace(d.input.Value())  // ❌ Duplicate logic
        d.loading = true
        return DialogResultNone, d.createBranchCmd(branchName)  // ❌ Duplicate command
    }
    return DialogResultNone, nil
}
```

**Issues**:
- ❌ Duplicate validation (`isValidBranchName()`)
- ❌ Duplicate command creation (`createBranchCmd()`)
- ❌ Relies on `d.loading` flag to prevent double execution (fragile)
- ❌ Changes must be made in two places
- ❌ Violates DRY principle

#### Bad Example 2: expand_preview.go (Critical Issue)

```go
// Update() calls handleKeyMsg
func (d *ExpandTaskPreviewDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return d.handleKeyMsg(msg)  // ❌ First call
    }
    return d, nil
}

// HandleKey() also calls handleKeyMsg
func (d *ExpandTaskPreviewDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    _, cmd := d.handleKeyMsg(msg)  // ❌ Second call
    return DialogResultNone, cmd
}
```

**Issues**:
- ❌ Keys processed twice per keystroke
- ❌ `handleKeyMsg()` executes twice
- ❌ Callbacks called twice
- ❌ Potential for race conditions

### Why Duplicate Handling is Dangerous

1. **Fragile Mitigation**: Relies on flag state (d.loading) to prevent issues
2. **Race Conditions**: If call sequence changes, behavior breaks
3. **Double Execution**: Commands or callbacks may run twice
4. **Maintenance Burden**: Changes need updates in multiple places
5. **Testing Complexity**: Requires understanding implicit contracts

---

## Pattern Selection Decision Tree

```
Start: Need to implement Dialog interface
    ↓
Does your dialog have complex state management?
    │
    ├─ YES: Multiple modes (edit/list), state machines?
    │   │
    │   └─ Use Pattern 2: Delegation
    │       • Create shared handleKeyMsg() function
    │       • Both Update() and HandleKey() delegate to it
    │       • Example: expand_edit.go
    │
    └─ NO: Simple form or single-mode dialog?
        │
        └─ Use Pattern 1: State-Only Update
            • Update() handles only state messages
            • HandleKey() handles ALL keyboard logic
            • Example: git_status.go, form.go

❌ NEVER use duplicate key handling
❌ AVOID Pattern 3 (Update→HandleKey delegation) for new code
```

### Quick Reference

| Dialog Type | Recommended Pattern | Example File |
|-------------|---------------------|--------------|
| Simple form with textinput | Pattern 1 (State-Only) | form.go |
| Dialog with validation | Pattern 1 (State-Only) | git_status.go |
| Multi-mode dialog (edit/list) | Pattern 2 (Delegation) | expand_edit.go |
| Complex state machine | Pattern 2 (Delegation) | expand_edit.go |
| List-based navigation | Pattern 1 or 2 | commits.go |
| Modal content component | Pattern 1 (State-Only) | next_task_output.go |

---

## Implementation Checklist

### ✅ Before Writing Code

- [ ] Read this guide completely
- [ ] Review good examples (commits.go, git_status.go, expand_edit.go)
- [ ] Decide which pattern to use (Pattern 1 or 2)
- [ ] Sketch out keyboard handling logic

### ✅ Update() Method

- [ ] Only processes non-keyboard messages (WindowSize, custom messages)
- [ ] Updates component states (textinput, viewport, list)
- [ ] Forwards messages to child components
- [ ] Does NOT contain dialog control flow logic
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

## Code Review Guidelines

### For Reviewers

Use this checklist when reviewing dialog PRs:

#### 1. Pattern Identification

- [ ] Identify which pattern the implementation uses
- [ ] Verify pattern is appropriate for dialog complexity
- [ ] Check consistency with similar dialogs in codebase

#### 2. Update() Method Review

- [ ] Confirm no keyboard key handling
- [ ] Verify WindowSizeMsg handled (if needed)
- [ ] Check custom message handling (result messages)
- [ ] Ensure component state updates only

#### 3. HandleKey() Method Review

- [ ] Verify ALL keyboard logic is here
- [ ] Check for base dialog method calls (HandleBaseKey)
- [ ] Confirm appropriate DialogResult returns
- [ ] Validate loading/state guards are present

#### 4. Duplication Check

- [ ] Search for same key strings in both methods
- [ ] Look for duplicate validation logic
- [ ] Check for duplicate command creation
- [ ] Verify no fragile flag-based mitigation

#### 5. Testing Requirements

- [ ] Unit tests for Update() (state changes)
- [ ] Unit tests for HandleKey() (keyboard actions)
- [ ] Integration test for no double execution
- [ ] Manual testing for user experience

### Common Review Comments

**✅ Approved Patterns**:
- "✅ Great use of State-Only Update pattern"
- "✅ Delegation pattern well-implemented"
- "✅ Clear separation between Update() and HandleKey()"

**⚠️ Minor Issues**:
- "⚠️ Consider moving key handling from Update() to HandleKey()"
- "⚠️ Viewport update may be redundant (KeyMap disabled)"
- "⚠️ Pattern is non-standard but functionally correct"

**❌ Must Fix**:
- "❌ Duplicate key handling detected in both methods"
- "❌ Enter key processed twice - use Pattern 1 or 2"
- "❌ Fragile d.loading mitigation - refactor to single source of truth"

---

## Testing Strategy

### Unit Tests

#### Test Update() Method

```go
func TestDialog_Update_StateMessagesOnly(t *testing.T) {
    tests := []struct {
        name     string
        msg      tea.Msg
        wantCmd  bool
        validate func(d Dialog)
    }{
        {
            name: "WindowSize updates dialog size",
            msg:  tea.WindowSizeMsg{Width: 100, Height: 50},
            wantCmd: false,
            validate: func(d Dialog) {
                // Verify size updated correctly
            },
        },
        {
            name: "Custom result message processed",
            msg:  customResultMsg{data: "test"},
            wantCmd: true,
            validate: func(d Dialog) {
                // Verify state changed
            },
        },
        {
            name: "KeyMsg NOT handled by Update",
            msg:  tea.KeyMsg{Type: tea.KeyEnter},
            wantCmd: false,
            validate: func(d Dialog) {
                // Verify no state changes from keys
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
```

#### Test HandleKey() Method

```go
func TestDialog_HandleKey_AllKeyboardLogic(t *testing.T) {
    tests := []struct {
        name       string
        key        tea.KeyMsg
        wantResult DialogResult
        wantCmd    bool
        validate   func(d Dialog)
    }{
        {
            name:       "Enter submits dialog",
            key:        tea.KeyMsg{Type: tea.KeyEnter},
            wantResult: DialogResultConfirm,
            wantCmd:    true,
            validate: func(d Dialog) {
                // Verify submit action taken
            },
        },
        {
            name:       "Esc cancels dialog",
            key:        tea.KeyMsg{Type: tea.KeyEsc},
            wantResult: DialogResultCancel,
            wantCmd:    false,
            validate: func(d Dialog) {
                // Verify cancel action taken
            },
        },
        {
            name:       "Loading state prevents action",
            key:        tea.KeyMsg{Type: tea.KeyEnter},
            wantResult: DialogResultNone,
            wantCmd:    false,
            validate: func(d Dialog) {
                // Set d.loading = true before test
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            d := NewDialog()
            if tt.validate != nil {
                tt.validate(d)  // Setup state
            }
            
            result, cmd := d.HandleKey(tt.key)
            
            if result != tt.wantResult {
                t.Errorf("HandleKey() result = %v, want %v", result, tt.wantResult)
            }
            
            if (cmd != nil) != tt.wantCmd {
                t.Errorf("HandleKey() cmd = %v, want %v", cmd != nil, tt.wantCmd)
            }
        })
    }
}
```

### Integration Tests

#### Test DialogManager Flow

```go
func TestDialogManager_CallSequence(t *testing.T) {
    // Verify Update() called before HandleKey()
    manager := NewDialogManager()
    dialog := NewTestDialog()
    
    // Simulate user pressing Enter
    keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
    
    // Track call order
    var callOrder []string
    dialog.onUpdate = func() { callOrder = append(callOrder, "Update") }
    dialog.onHandleKey = func() { callOrder = append(callOrder, "HandleKey") }
    
    // Process key through manager
    manager.ProcessKey(keyMsg)
    
    // Verify call sequence
    assert.Equal(t, []string{"Update", "HandleKey"}, callOrder)
}
```

#### Test No Double Execution

```go
func TestDialog_NoDoubleExecution(t *testing.T) {
    dialog := NewDialog()
    
    // Track command executions
    executionCount := 0
    dialog.onCommand = func() { executionCount++ }
    
    // Simulate DialogManager flow
    keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
    dialog.Update(keyMsg)     // Called first by DialogManager
    dialog.HandleKey(keyMsg)  // Called second by DialogManager
    
    // Command should only execute once
    assert.Equal(t, 1, executionCount, "Command executed more than once")
}
```

### Manual Testing Checklist

- [ ] Press Enter → Dialog submits correctly
- [ ] Press Esc → Dialog cancels correctly
- [ ] Type text → Input updates in real-time
- [ ] Press Enter while loading → No double submission
- [ ] Navigate with arrows → Selection moves correctly
- [ ] Tab between fields → Focus changes correctly
- [ ] Resize window → Dialog recenters correctly

---

## Migration Guide

### For branch_create.go

**Current Issue**: Duplicate "enter" key handling

**Steps to Fix**:

1. **Remove key handling from Update()**

```diff
func (d *BranchCreateDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
-   case tea.KeyMsg:
-       if d.loading {
-           return d, nil
-       }
-
-       switch msg.String() {
-       case "enter":
-           if !d.isValidBranchName() {
-               d.errorMsg = "Invalid branch name"
-               return d, nil
-           }
-
-           branchName := strings.TrimSpace(d.input.Value())
-           d.loading = true
-           d.errorMsg = ""
-           return d, d.createBranchCmd(branchName)
-
-       case "esc":
-           return nil, nil
-       }
-
+   case tea.WindowSizeMsg:
+       d.Center(msg.Width, msg.Height)
+       return d, nil
+
    case branchCreateMsg:
        if msg.err != nil {
            d.loading = false
            d.errorMsg = "Error: " + msg.err.Error()
            return d, nil
        }
        
        if d.onComplete != nil {
            d.onComplete(msg.branchName, msg.output, nil)
        }
        return nil, nil
    }
    
    // Handle text input updates
    var cmd tea.Cmd
    d.input, cmd = d.input.Update(msg)
    
    if _, ok := msg.(tea.KeyMsg); ok {
        d.validateAndUpdateError()
    }
    
    return d, cmd
}
```

2. **Keep HandleKey() as-is** (already correct)

3. **Add tests**

```go
func TestBranchCreateDialog_UpdateDoesNotHandleEnter(t *testing.T) {
    dialog := NewBranchCreateDialog("/repo", nil)
    dialog.input.SetValue("test-branch")
    
    enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
    updatedDialog, cmd := dialog.Update(enterMsg)
    
    // Should NOT create branch command
    assert.False(t, updatedDialog.(*BranchCreateDialog).loading)
}
```

### For expand_preview.go

**Current Issue**: Both methods call handleKeyMsg()

**Steps to Fix**:

1. **Remove key handling from Update()**

```diff
func (d *ExpandTaskPreviewDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    if !d.focused {
        return d, nil
    }

-   switch msg := msg.(type) {
-   case tea.KeyMsg:
-       return d.handleKeyMsg(msg)
-   }

    return d, nil
}
```

2. **Keep HandleKey() and handleKeyMsg() as-is** (already correct)

3. **Add tests** (similar to branch_create.go)

### For complexity_report.go (Optional)

**Current Issue**: Redundant viewport update

**Steps to Fix**:

```diff
func (d *ComplexityReportDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        d.updateSize(msg.Width, msg.Height)
        return d, nil

    case tea.KeyMsg:
        result, cmd := d.HandleKey(msg)
        if result != DialogResultNone {
            return nil, cmd
        }

-       // Redundant: viewport KeyMap is disabled
-       var cmd2 tea.Cmd
-       d.viewport, cmd2 = d.viewport.Update(msg)
-       return d, tea.Batch(cmd, cmd2)
+       return d, cmd
    }

    return d, nil
}
```

---

## Reference Materials

### Good Examples to Study

#### Simple Dialogs (Pattern 1)

1. **git_status.go** - Perfect State-Only Update example
   - Lines 52-66: Update() handles only state
   - Lines 186-201: HandleKey() handles all keys
   - Clean, clear, maintainable

2. **commits.go** - List-based dialog
   - Lines 70-86: Update() handles state
   - Lines 224-266: HandleKey() handles navigation
   - Good list integration pattern

3. **form.go** - Form with textinput/textarea
   - Lines 338-348: Update() handles events only
   - Lines 350-380: HandleKey() handles all keyboard logic
   - Excellent textarea handling

#### Complex Dialogs (Pattern 2)

1. **expand_edit.go** - Multi-mode dialog
   - Lines 94-104: Both methods delegate to handleKeyMsg()
   - Lines 107-170: Shared keyboard logic
   - Good state machine example

2. **file_selection.go** - File browser
   - Delegation pattern
   - Complex navigation
   - Good example of shared logic

### Anti-Pattern Examples

1. **branch_create.go** - Duplicate key handling
   - Lines 67-77: Update() handles "enter"
   - Lines 223-237: HandleKey() also handles "enter"
   - ❌ Fragile d.loading mitigation

2. **expand_preview.go** - Duplicate delegation
   - Lines 79: Update() calls handleKeyMsg()
   - Lines 314: HandleKey() also calls handleKeyMsg()
   - ❌ Keys processed twice

### Related Documentation

- **DIALOG_IMPLEMENTATION_GUIDE.md** - Detailed patterns with examples
- **ISSUE_REPORT_branch_create.md** - Detailed analysis of branch_create.go issue
- **VERIFICATION_REPORT_branch_switch.md** - Analysis of branch_switch.go (acceptable pattern)
- **DIALOG_REVIEW_REPORT.md** - Comprehensive review of all 6 remaining dialogs

### Base Dialog Classes

- **BaseDialog** - Basic dialog with common functionality
  - `HandleBaseKey()` - Handles Esc for cancel
  - `Center()` - Centers dialog on resize
  - `RenderBorder()` - Draws dialog border

- **BaseFocusableDialog** - Dialog with focus management
  - `HandleBaseFocusableKey()` - Handles Tab/Shift+Tab for focus
  - `FocusNext()` / `FocusPrev()` - Move focus between elements
  - `FocusedIndex()` - Get currently focused element

---

## Appendix: Quick Reference Cards

### Pattern 1 Template

```go
func (d *MyDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        d.Center(msg.Width, msg.Height)
        return d, nil
    case MyResultMsg:
        // Handle command result
        return d, nil
    }
    return d, nil
}

func (d *MyDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    result, cmd := d.HandleBaseKey(msg)
    if result != DialogResultNone {
        return result, cmd
    }
    
    switch msg.String() {
    case "enter":
        // Handle enter
        return DialogResultConfirm, cmd
    }
    
    return DialogResultNone, nil
}
```

### Pattern 2 Template

```go
func (d *MyDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return d.handleKeyMsg(msg)
    }
    return d, nil
}

func (d *MyDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
    _, cmd := d.handleKeyMsg(msg)
    return DialogResultNone, cmd
}

func (d *MyDialog) handleKeyMsg(msg tea.KeyMsg) (Dialog, tea.Cmd) {
    switch msg.String() {
    case "enter":
        // Handle enter
        return d, cmd
    }
    return d, nil
}
```

---

## Summary

### Key Takeaways

1. **Two Recommended Patterns**: State-Only Update and Delegation
2. **Clear Separation**: Update() for state, HandleKey() for control flow
3. **No Duplication**: Single source of truth for each action
4. **Good Examples Exist**: Study commits.go, git_status.go, expand_edit.go
5. **Issues Identified**: 2 files need fixing (branch_create.go, expand_preview.go)

### Action Items

**Immediate (High Priority)**:
- Fix branch_create.go
- Fix expand_preview.go

**Optional (Low Priority)**:
- Refactor complexity_report.go (remove redundant viewport update)
- Document task_runner_modal.go pattern (add comments)

**For New Development**:
- Always use Pattern 1 or Pattern 2
- Follow this guide
- Review good examples before implementing
- Use implementation checklist
- Write tests to prevent regression

---

**Document Version**: 1.0 Final  
**Last Updated**: 2026-01-12  
**Status**: Complete  
**Related Tasks**: Task 3.5 (git-dialog-fix)  
**Authors**: Task Master TUI Code Review Team
