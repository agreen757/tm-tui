package main

import (
	"fmt"
	"testing"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

// TestApproach1_PopDialogBeforeTaskRunner tests popping the command runner dialog
// before creating the task runner modal
func TestApproach1_PopDialogBeforeTaskRunner(t *testing.T) {
	fmt.Println("\n=== Testing Approach 1: Pop Command Runner Dialog First ===")

	// Simulate dialog manager state
	manager := dialog.NewDialogManager(100, 50)

	// Step 1: Create and add Command Runner Dialog (simulating openCommandRunner)
	commandRunnerDialog := dialog.NewFormDialog(
		"Command Runner",
		"Enter a command prompt",
		[]dialog.FormField{
			{
				ID:          "prompt",
				Label:       "Command Prompt",
				Type:        dialog.FormFieldTypeText,
				Placeholder: "Enter command...",
			},
		},
		[]string{"Execute", "Cancel"},
		nil, // style
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			return values, nil
		},
	)

	manager.AddDialog(commandRunnerDialog, nil)
	dialogCount1 := countDialogs(manager)
	fmt.Printf("After adding Command Runner Dialog: %d dialogs in stack\n", dialogCount1)
	fmt.Printf("  - Has active dialog: %v\n", manager.GetActiveDialog() != nil)

	// Step 2: Pop Command Runner Dialog before creating Task Runner
	poppedDialog := manager.PopDialog()
	dialogCount2 := countDialogs(manager)
	fmt.Printf("After popping Command Runner Dialog: %d dialogs in stack\n", dialogCount2)
	fmt.Printf("  - Popped dialog: %v\n", poppedDialog != nil)

	// Step 3: Create and add Task Runner Modal
	taskRunnerModal := dialog.NewTaskRunnerModal(80, 30, nil)
	manager.PushDialog(taskRunnerModal)
	dialogCount3 := countDialogs(manager)
	fmt.Printf("After adding Task Runner Modal: %d dialogs in stack\n", dialogCount3)
	fmt.Printf("  - Has active dialog: %v\n", manager.GetActiveDialog() != nil)

	// Verify: Should only have 1 dialog (Task Runner Modal)
	if dialogCount3 != 1 {
		t.Errorf("Expected 1 dialog in stack, got %d", dialogCount3)
	} else {
		fmt.Println("✓ Approach 1: Only 1 modal in stack (correct)")
	}
}

// TestApproach2_NoDialogStackForTaskRunner tests NOT pushing the Task Runner Modal
// to the dialog stack at all
func TestApproach2_NoDialogStackForTaskRunner(t *testing.T) {
	fmt.Println("\n=== Testing Approach 2: Don't Push Task Runner to Dialog Stack ===")

	// Simulate dialog manager state
	manager := dialog.NewDialogManager(100, 50)

	// Step 1: Create and add Command Runner Dialog
	commandRunnerDialog := dialog.NewFormDialog(
		"Command Runner",
		"Enter a command prompt",
		[]dialog.FormField{
			{
				ID:          "prompt",
				Label:       "Command Prompt",
				Type:        dialog.FormFieldTypeText,
				Placeholder: "Enter command...",
			},
		},
		[]string{"Execute", "Cancel"},
		nil,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			return values, nil
		},
	)

	manager.AddDialog(commandRunnerDialog, nil)
	dialogCount1 := countDialogs(manager)
	fmt.Printf("After adding Command Runner Dialog: %d dialogs in stack\n", dialogCount1)

	// Step 2: Command Runner Dialog automatically pops when callback returns
	// (simulating dialog completion)
	manager.PopDialog()
	dialogCount2 := countDialogs(manager)
	fmt.Printf("After Command Runner Dialog callback completes: %d dialogs in stack\n", dialogCount2)

	// Step 3: Create Task Runner Modal but DON'T push to dialog stack
	// Instead, manage it separately via m.taskRunnerVisible flag
	taskRunnerModal := dialog.NewTaskRunnerModal(80, 30, nil)
	taskRunnerVisible := true // Simulating m.taskRunnerVisible = true
	
	fmt.Printf("Task Runner Modal created (managed separately): %v\n", taskRunnerModal != nil)
	fmt.Printf("Task Runner visible flag: %v\n", taskRunnerVisible)
	dialogCount3 := countDialogs(manager)
	fmt.Printf("Dialog stack count: %d\n", dialogCount3)

	// Verify: Dialog stack should be empty (0 dialogs)
	if dialogCount3 != 0 {
		t.Errorf("Expected 0 dialogs in stack, got %d", dialogCount3)
	} else {
		fmt.Println("✓ Approach 2: Dialog stack empty, Task Runner managed separately (correct)")
	}
}

// TestCurrentBehavior_BothInStack tests the CURRENT buggy behavior
// where both dialogs end up in the stack
func TestCurrentBehavior_BothInStack(t *testing.T) {
	fmt.Println("\n=== Testing Current Buggy Behavior ===")

	// Simulate dialog manager state
	manager := dialog.NewDialogManager(100, 50)

	// Step 1: Create and add Command Runner Dialog
	commandRunnerDialog := dialog.NewFormDialog(
		"Command Runner",
		"Enter a command prompt",
		[]dialog.FormField{
			{
				ID:          "prompt",
				Label:       "Command Prompt",
				Type:        dialog.FormFieldTypeText,
				Placeholder: "Enter command...",
			},
		},
		[]string{"Execute", "Cancel"},
		nil,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			return values, nil
		},
	)

	manager.AddDialog(commandRunnerDialog, nil)
	dialogCount1 := countDialogs(manager)
	fmt.Printf("After adding Command Runner Dialog: %d dialogs in stack\n", dialogCount1)

	// Step 2: In handleCommandRunnerSubmission, Task Runner Modal is created
	// and pushed BEFORE the Command Runner Dialog callback completes
	taskRunnerModal := dialog.NewTaskRunnerModal(80, 30, nil)
	manager.PushDialog(taskRunnerModal)
	dialogCount2 := countDialogs(manager)
	fmt.Printf("After adding Task Runner Modal (BEFORE popping Command Runner): %d dialogs in stack\n", dialogCount2)

	// Now both dialogs are in the stack!
	if dialogCount2 != 2 {
		t.Errorf("Expected 2 dialogs in stack (buggy behavior), got %d", dialogCount2)
	} else {
		fmt.Println("✗ Current Behavior: 2 modals in stack simultaneously (BUG CONFIRMED)")
	}

	// Step 3: Eventually Command Runner Dialog pops, but there's a moment where both exist
	manager.PopDialog() // This would pop Task Runner (top of stack)
	dialogCount3 := countDialogs(manager)
	fmt.Printf("After first pop: %d dialogs in stack\n", dialogCount3)
	manager.PopDialog() // This pops Command Runner
	dialogCount4 := countDialogs(manager)
	fmt.Printf("After second pop: %d dialogs in stack\n", dialogCount4)
}

// TestDialogStackOrdering verifies dialog stack LIFO behavior
func TestDialogStackOrdering(t *testing.T) {
	fmt.Println("\n=== Testing Dialog Stack Ordering (LIFO) ===")

	manager := dialog.NewDialogManager(100, 50)

	// Add first dialog
	dialog1 := dialog.NewFormDialog("Dialog 1", "", []dialog.FormField{}, []string{"OK"}, nil, nil)
	manager.AddDialog(dialog1, nil)
	dialogCount1 := countDialogs(manager)
	fmt.Printf("Added Dialog 1, stack count: %d\n", dialogCount1)

	// Add second dialog
	dialog2 := dialog.NewFormDialog("Dialog 2", "", []dialog.FormField{}, []string{"OK"}, nil, nil)
	manager.AddDialog(dialog2, nil)
	dialogCount2 := countDialogs(manager)
	fmt.Printf("Added Dialog 2, stack count: %d\n", dialogCount2)

	// Pop should return Dialog 2 first (Last In, First Out)
	popped := manager.PopDialog()
	if popped == nil {
		t.Fatal("Expected to pop Dialog 2, got nil")
	}
	dialogCount3 := countDialogs(manager)
	fmt.Printf("Popped dialog (should be Dialog 2), stack count: %d\n", dialogCount3)

	// Pop should return Dialog 1 next
	popped = manager.PopDialog()
	if popped == nil {
		t.Fatal("Expected to pop Dialog 1, got nil")
	}
	dialogCount4 := countDialogs(manager)
	fmt.Printf("Popped dialog (should be Dialog 1), stack count: %d\n", dialogCount4)

	if dialogCount4 != 0 {
		t.Errorf("Expected empty stack, got %d dialogs", dialogCount4)
	} else {
		fmt.Println("✓ Dialog stack follows LIFO ordering correctly")
	}
}

// countDialogs returns the number of dialogs in the manager by repeatedly
// checking GetActiveDialog and counting
func countDialogs(manager *dialog.DialogManager) int {
	count := 0
	// Save current state
	var dialogs []dialog.Dialog
	
	// Pop all dialogs and count them
	for manager.GetActiveDialog() != nil {
		d := manager.PopDialog()
		if d != nil {
			dialogs = append(dialogs, d)
			count++
		} else {
			break
		}
	}
	
	// Restore dialogs in reverse order (to maintain original order)
	for i := len(dialogs) - 1; i >= 0; i-- {
		manager.PushDialog(dialogs[i])
	}
	
	return count
}

func main() {
	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{"TestApproach1_PopDialogBeforeTaskRunner", TestApproach1_PopDialogBeforeTaskRunner},
			{"TestApproach2_NoDialogStackForTaskRunner", TestApproach2_NoDialogStackForTaskRunner},
			{"TestCurrentBehavior_BothInStack", TestCurrentBehavior_BothInStack},
			{"TestDialogStackOrdering", TestDialogStackOrdering},
		},
		nil,
		nil,
	)
}
