package dialog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TestTagEditorFlowCreation tests basic flow creation
func TestTagEditorFlowCreation(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	if flow == nil {
		t.Fatal("Expected flow to be created, got nil")
	}
	if flow.currentState != StateTagSelector {
		t.Errorf("Expected initial state to be StateTagSelector, got %d", flow.currentState)
	}
	if flow.cancelled {
		t.Error("Expected cancelled to be false initially")
	}
}

// TestInitialSelectorDisplay tests getting the initial selector
func TestInitialSelectorDisplay(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)
	selector := flow.GetInitialSelector()

	if selector == nil {
		t.Fatal("Expected selector to be created, got nil")
	}

	// Verify selector is cached
	selector2 := flow.GetInitialSelector()
	if selector != selector2 {
		t.Error("Expected same selector instance on second call")
	}
}

// TestHandleSelectorResultWithExistingTags tests selecting existing tags
func TestHandleSelectorResultWithExistingTags(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	result := TagSelectorResult{
		SelectedTags: []string{"feature"},
		AddNewTag:    false,
	}

	shouldContinue, finalResult := flow.HandleSelectorResult(result, true)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false when selecting existing tags")
	}
	if flow.currentState != StateComplete {
		t.Errorf("Expected state to be StateComplete, got %d", flow.currentState)
	}
	if !flow.IsComplete() {
		t.Error("Expected IsComplete() to return true")
	}
	if len(finalResult.SelectedTags) != 1 || finalResult.SelectedTags[0] != "feature" {
		t.Errorf("Expected 1 selected tag 'feature', got %v", finalResult.SelectedTags)
	}
}

// TestHandleSelectorResultWithAddNewTag tests selecting "Add New Tag"
func TestHandleSelectorResultWithAddNewTag(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}

	shouldContinue, _ := flow.HandleSelectorResult(result, true)

	if !shouldContinue {
		t.Error("Expected shouldContinue to be true when AddNewTag is selected")
	}
	if flow.currentState != StateAddTagDialog {
		t.Errorf("Expected state to be StateAddTagDialog, got %d", flow.currentState)
	}
}

// TestHandleSelectorResultWithCancel tests cancelling the selector
func TestHandleSelectorResultWithCancel(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    false,
	}

	shouldContinue, _ := flow.HandleSelectorResult(result, false)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false on cancel")
	}
	if flow.currentState != StateCancelled {
		t.Errorf("Expected state to be StateCancelled, got %d", flow.currentState)
	}
	if !flow.IsCancelled() {
		t.Error("Expected IsCancelled() to return true")
	}
}

// TestHandleNewTagCreated tests handling new tag creation
func TestHandleNewTagCreated(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Move to add tag state first
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	flow.HandleSelectorResult(result, true)

	// Create new tag list with new tag
	updatedTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "newfeature",
				TaskCount:      0,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-16",
				Active:         false,
			},
		},
	}

	flow.HandleNewTagCreated("newfeature", updatedTagList)

	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector after creating new tag, got %d", flow.currentState)
	}
	if flow.newTagCreated != "newfeature" {
		t.Errorf("Expected newTagCreated to be 'newfeature', got %s", flow.newTagCreated)
	}
	if len(flow.lastTagList.Tags) != 2 {
		t.Errorf("Expected 2 tags in updated list, got %d", len(flow.lastTagList.Tags))
	}

	// Verify selector is reset (recreated on next call)
	flow.HandleNewTagCreated("newfeature", updatedTagList)
	if flow.selector != nil {
		t.Error("Expected selector to be reset to nil after HandleNewTagCreated")
	}
}

// TestHandleAddTagCancelled tests cancelling add tag dialog
func TestHandleAddTagCancelled(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Move to add tag state
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	flow.HandleSelectorResult(result, true)

	if flow.currentState != StateAddTagDialog {
		t.Errorf("Expected state to be StateAddTagDialog, got %d", flow.currentState)
	}

	// Cancel add tag
	flow.HandleAddTagCancelled()

	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector after cancel, got %d", flow.currentState)
	}
	if flow.newTagCreated != "" {
		t.Errorf("Expected newTagCreated to be empty, got %s", flow.newTagCreated)
	}
}

// TestFlowMultiSelectMode tests multi-select mode flow
func TestFlowMultiSelectMode(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "bugfix",
				TaskCount:      3,
				CompletedCount: 1,
				CreatedLabel:   "2024-01-10",
				Active:         false,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: true,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	if !cfg.MultiSelect {
		t.Error("Expected MultiSelect to be true")
	}

	result := TagSelectorResult{
		SelectedTags: []string{"feature", "bugfix"},
		AddNewTag:    false,
	}

	shouldContinue, finalResult := flow.HandleSelectorResult(result, true)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false")
	}
	if len(finalResult.SelectedTags) != 2 {
		t.Errorf("Expected 2 selected tags, got %d", len(finalResult.SelectedTags))
	}
}

// TestFlowCancel tests the explicit Cancel method
func TestFlowCancel(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	flow.Cancel()

	if !flow.IsCancelled() {
		t.Error("Expected IsCancelled() to return true after Cancel()")
	}
	if flow.currentState != StateCancelled {
		t.Errorf("Expected state to be StateCancelled, got %d", flow.currentState)
	}
}

// TestFlowStateTransitions tests valid state transitions
func TestFlowStateTransitions(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Verify state at each step
	states := []TagEditorFlowState{
		StateTagSelector,
		StateTagSelector,
	}

	if flow.GetCurrentState() != states[0] {
		t.Errorf("Expected initial state %d, got %d", states[0], flow.GetCurrentState())
	}

	// Move through states
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	flow.HandleSelectorResult(result, true)

	if flow.GetCurrentState() != StateAddTagDialog {
		t.Errorf("Expected state %d, got %d", StateAddTagDialog, flow.GetCurrentState())
	}

	flow.HandleAddTagCancelled()

	if flow.GetCurrentState() != StateTagSelector {
		t.Errorf("Expected state %d, got %d", StateTagSelector, flow.GetCurrentState())
	}
}

// TestFlowWithEmptyTagList tests flow with no existing tags
func TestFlowWithEmptyTagList(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)
	selector := flow.GetInitialSelector()

	if selector == nil {
		t.Fatal("Expected selector to be created even with empty tag list")
	}

	// Should still allow adding new tag
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}

	shouldContinue, _ := flow.HandleSelectorResult(result, true)

	if !shouldContinue {
		t.Error("Expected shouldContinue to be true even with empty tag list")
	}
}

// TestFlowGetResult tests retrieving final result
func TestFlowGetResult(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	result := TagSelectorResult{
		SelectedTags: []string{"feature"},
		AddNewTag:    false,
	}

	flow.HandleSelectorResult(result, true)

	finalResult := flow.GetResult()

	if len(finalResult.SelectedTags) != 1 || finalResult.SelectedTags[0] != "feature" {
		t.Errorf("Expected result with 'feature' tag, got %v", finalResult.SelectedTags)
	}
	if finalResult.AddNewTag {
		t.Error("Expected AddNewTag to be false in final result")
	}
}

// TestSetRefreshFunc tests setting a refresh function
func TestSetRefreshFunc(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Mock refresh function
	flow.SetRefreshFunc(func(ctx context.Context) (*taskmaster.TagList, error) {
		return tagList, nil
	})

	if flow.refreshFunc == nil {
		t.Error("Expected refreshFunc to be set")
	}
}

// TestRefreshTagListWithCallback tests tag list refresh using callback
func TestRefreshTagListWithCallback(t *testing.T) {
	originalTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     originalTagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Mock refresh function that returns updated list
	updatedTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "newfeature",
				TaskCount:      0,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-16",
				Active:         false,
			},
		},
	}

	flow.SetRefreshFunc(func(ctx context.Context) (*taskmaster.TagList, error) {
		return updatedTagList, nil
	})

	refreshedList, err := flow.RefreshTagList(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(refreshedList.Tags) != 2 {
		t.Errorf("Expected 2 tags after refresh, got %d", len(refreshedList.Tags))
	}
	if flow.lastTagList != updatedTagList {
		t.Error("Expected lastTagList to be updated")
	}
	if flow.config.TagList != updatedTagList {
		t.Error("Expected config.TagList to be updated")
	}
}

// TestRefreshTagListWithoutCallback tests refresh without callback returns current list
func TestRefreshTagListWithoutCallback(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	refreshedList, err := flow.RefreshTagList(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if refreshedList != tagList {
		t.Error("Expected same tag list to be returned")
	}
}

// TestHandleNewTagCreatedWithRefresh tests new tag creation with refresh callback
func TestHandleNewTagCreatedWithRefresh(t *testing.T) {
	originalTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     originalTagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Move to add tag state
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	flow.HandleSelectorResult(result, true)

	// Create updated tag list
	updatedTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "bugfix",
				TaskCount:      0,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-16",
				Active:         false,
			},
		},
	}

	// Set refresh callback
	flow.SetRefreshFunc(func(ctx context.Context) (*taskmaster.TagList, error) {
		return updatedTagList, nil
	})

	// Handle new tag creation with refresh
	err := flow.HandleNewTagCreatedWithRefresh(context.Background(), "bugfix", updatedTagList)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector, got %d", flow.currentState)
	}
	if flow.newTagCreated != "bugfix" {
		t.Errorf("Expected newTagCreated to be 'bugfix', got %s", flow.newTagCreated)
	}
	if len(flow.lastTagList.Tags) != 2 {
		t.Errorf("Expected 2 tags after refresh, got %d", len(flow.lastTagList.Tags))
	}
	if flow.selector != nil {
		t.Error("Expected selector to be reset to nil")
	}
}

// TestHandleNewTagCreatedWithRefreshError tests refresh error handling
func TestHandleNewTagCreatedWithRefreshError(t *testing.T) {
	originalTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     originalTagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Move to add tag state
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	flow.HandleSelectorResult(result, true)

	// Create updated tag list
	updatedTagList := &taskmaster.TagList{
		Tags: append(originalTagList.Tags, taskmaster.TagContext{
			Name:           "bugfix",
			TaskCount:      0,
			CompletedCount: 0,
			CreatedLabel:   "2024-01-16",
			Active:         false,
		}),
	}

	// Set refresh callback that returns error
	flow.SetRefreshFunc(func(ctx context.Context) (*taskmaster.TagList, error) {
		return nil, ErrTestRefreshFailed
	})

	// Handle new tag creation with refresh error
	err := flow.HandleNewTagCreatedWithRefresh(context.Background(), "bugfix", updatedTagList)

	if err != ErrTestRefreshFailed {
		t.Errorf("Expected ErrTestRefreshFailed, got %v", err)
	}
	// Should still use provided list as fallback
	if len(flow.lastTagList.Tags) != 2 {
		t.Errorf("Expected 2 tags in fallback list, got %d", len(flow.lastTagList.Tags))
	}
	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to remain StateTagSelector on error, got %d", flow.currentState)
	}
}

// Custom error for testing
var ErrTestRefreshFailed = &testError{"refresh failed"}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestFlowError tests error wrapping
func TestFlowError(t *testing.T) {
	baseErr := fmt.Errorf("original error")
	flowErr := &FlowError{
		Stage:   "selector",
		Message: "selection failed",
		Err:     baseErr,
	}

	errStr := flowErr.Error()
	if !strings.Contains(errStr, "selector") {
		t.Errorf("Expected 'selector' in error message, got: %s", errStr)
	}
	if !strings.Contains(errStr, "selection failed") {
		t.Errorf("Expected 'selection failed' in error message, got: %s", errStr)
	}

	if flowErr.Unwrap() != baseErr {
		t.Error("Expected Unwrap to return original error")
	}
}

// TestHandleError tests error handling
func TestHandleError(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	testErr := fmt.Errorf("test error")
	err := flow.HandleError("selector", "Failed to load tags", testErr)

	if err == nil {
		t.Fatal("Expected error to be returned")
	}
	if !strings.Contains(err.Error(), "selector") {
		t.Errorf("Expected 'selector' in error, got: %s", err.Error())
	}
	if flow.currentState != StateError {
		t.Errorf("Expected state to be StateError, got %d", flow.currentState)
	}
	if flow.IsInErrorState() == false {
		t.Error("Expected IsInErrorState to return true")
	}
	if flow.GetErrorCount() != 1 {
		t.Errorf("Expected error count 1, got %d", flow.GetErrorCount())
	}
}

// TestRecoverFromError tests error recovery
func TestRecoverFromError(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Cause an error
	testErr := fmt.Errorf("test error")
	flow.HandleError("selector", "Failed", testErr)

	if !flow.IsInErrorState() {
		t.Error("Expected flow to be in error state")
	}

	// Recover from error
	flow.RecoverFromError()

	if flow.IsInErrorState() {
		t.Error("Expected flow to no longer be in error state")
	}
	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector after recovery, got %d", flow.currentState)
	}
	if flow.GetLastError() != nil {
		t.Error("Expected last error to be cleared")
	}
}

// TestComprehensiveCancellation tests cancellation at selector stage
func TestComprehensiveCancellation(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Cancel at selector
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    false,
	}

	shouldContinue, finalResult := flow.HandleSelectorResult(result, false)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false on cancel")
	}
	if !flow.IsCancelled() {
		t.Error("Expected IsCancelled to be true")
	}
	if len(finalResult.SelectedTags) > 0 {
		t.Error("Expected empty result on cancel")
	}
}

// TestCancellationWithErrorRecovery tests cancelling after error recovery
func TestCancellationWithErrorRecovery(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Cause an error
	testErr := fmt.Errorf("test error")
	flow.HandleError("selector", "Failed", testErr)

	// Recover
	flow.RecoverFromError()

	// Then cancel
	flow.Cancel()

	if !flow.IsCancelled() {
		t.Error("Expected flow to be cancelled after recovery and cancel")
	}
	if flow.IsInErrorState() {
		t.Error("Expected error state to be cleared after cancel")
	}
}

// TestErrorCountIncrement tests error counting
func TestErrorCountIncrement(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	if flow.GetErrorCount() != 0 {
		t.Errorf("Expected initial error count 0, got %d", flow.GetErrorCount())
	}

	// Add multiple errors
	for i := 1; i <= 3; i++ {
		testErr := fmt.Errorf("error %d", i)
		flow.HandleError("selector", fmt.Sprintf("Error %d", i), testErr)

		if flow.GetErrorCount() != i {
			t.Errorf("Expected error count %d, got %d", i, flow.GetErrorCount())
		}
	}

	// Cancel should clear error count
	flow.Cancel()
	if flow.GetErrorCount() != 0 {
		t.Errorf("Expected error count to be cleared on cancel, got %d", flow.GetErrorCount())
	}
}

// TestAddTagDialogCancelLoopback tests that cancelling add tag returns to selector
func TestAddTagDialogCancelLoopback(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Move to add tag dialog
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	shouldContinue, _ := flow.HandleSelectorResult(result, true)

	if !shouldContinue {
		t.Error("Expected shouldContinue to be true")
	}
	if flow.currentState != StateAddTagDialog {
		t.Errorf("Expected state to be StateAddTagDialog, got %d", flow.currentState)
	}

	// Cancel from add tag dialog
	flow.HandleAddTagCancelled()

	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector after add tag cancel, got %d", flow.currentState)
	}
	if flow.IsCancelled() {
		t.Error("Expected flow to not be marked as cancelled (still in loopback)")
	}

	// Now user cancels from selector
	result2 := TagSelectorResult{}
	shouldContinue2, _ := flow.HandleSelectorResult(result2, false)

	if shouldContinue2 {
		t.Error("Expected shouldContinue to be false on final cancel")
	}
	if !flow.IsCancelled() {
		t.Error("Expected flow to be cancelled after returning from add tag")
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

// TestCompleteFlowSelectExistingTags integration test for selecting existing tags
func TestCompleteFlowSelectExistingTags(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "bugfix",
				TaskCount:      3,
				CompletedCount: 1,
				CreatedLabel:   "2024-01-10",
				Active:         false,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: true,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Step 1: Get initial selector
	selector := flow.GetInitialSelector()
	if selector == nil {
		t.Fatal("Expected selector to be created")
	}

	// Step 2: Select tags and confirm
	result := TagSelectorResult{
		SelectedTags: []string{"feature", "bugfix"},
		AddNewTag:    false,
	}

	shouldContinue, finalResult := flow.HandleSelectorResult(result, true)

	// Verify final state
	if shouldContinue {
		t.Error("Expected shouldContinue to be false")
	}
	if !flow.IsComplete() {
		t.Error("Expected flow to be complete")
	}
	if len(finalResult.SelectedTags) != 2 {
		t.Errorf("Expected 2 selected tags, got %d", len(finalResult.SelectedTags))
	}
	if flow.IsCancelled() {
		t.Error("Expected flow to not be cancelled")
	}
}

// TestCompleteFlowCreateNewTag integration test for creating and selecting new tag
func TestCompleteFlowCreateNewTag(t *testing.T) {
	originalTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     originalTagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Step 1: User selects "Add New Tag"
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	shouldContinue, _ := flow.HandleSelectorResult(result, true)

	if !shouldContinue {
		t.Error("Expected shouldContinue to be true for add new tag")
	}
	if flow.currentState != StateAddTagDialog {
		t.Errorf("Expected state to be StateAddTagDialog, got %d", flow.currentState)
	}

	// Step 2: New tag is created
	updatedTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "newfeature",
				TaskCount:      0,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-16",
				Active:         false,
			},
		},
	}

	flow.HandleNewTagCreated("newfeature", updatedTagList)

	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector after creation, got %d", flow.currentState)
	}
	if len(flow.lastTagList.Tags) != 2 {
		t.Errorf("Expected 2 tags in list, got %d", len(flow.lastTagList.Tags))
	}

	// Step 3: User selects the newly created tag
	result2 := TagSelectorResult{
		SelectedTags: []string{"newfeature"},
		AddNewTag:    false,
	}

	shouldContinue2, finalResult := flow.HandleSelectorResult(result2, true)

	if shouldContinue2 {
		t.Error("Expected shouldContinue to be false")
	}
	if !flow.IsComplete() {
		t.Error("Expected flow to be complete")
	}
	if len(finalResult.SelectedTags) != 1 || finalResult.SelectedTags[0] != "newfeature" {
		t.Errorf("Expected newfeature in result, got %v", finalResult.SelectedTags)
	}
}

// TestCompleteFlowCancelFromSelector integration test for cancelling at selector
func TestCompleteFlowCancelFromSelector(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// User cancels at selector
	result := TagSelectorResult{}
	shouldContinue, finalResult := flow.HandleSelectorResult(result, false)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false")
	}
	if !flow.IsCancelled() {
		t.Error("Expected flow to be cancelled")
	}
	if flow.IsComplete() {
		t.Error("Expected flow to not be complete")
	}
	if len(finalResult.SelectedTags) > 0 {
		t.Error("Expected empty result on cancel")
	}
}

// TestCompleteFlowCancelFromAddTag integration test for cancelling add tag then selector
func TestCompleteFlowCancelFromAddTag(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Step 1: Select "Add New Tag"
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	shouldContinue, _ := flow.HandleSelectorResult(result, true)

	if !shouldContinue {
		t.Error("Expected shouldContinue to be true")
	}

	// Step 2: Cancel from add tag dialog
	flow.HandleAddTagCancelled()

	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector, got %d", flow.currentState)
	}
	if flow.IsCancelled() {
		t.Error("Expected flow to not be cancelled yet")
	}

	// Step 3: User selects a tag instead
	result2 := TagSelectorResult{
		SelectedTags: []string{"feature"},
		AddNewTag:    false,
	}

	shouldContinue2, finalResult := flow.HandleSelectorResult(result2, true)

	if shouldContinue2 {
		t.Error("Expected shouldContinue to be false")
	}
	if !flow.IsComplete() {
		t.Error("Expected flow to be complete after selecting tag")
	}
	if len(finalResult.SelectedTags) != 1 {
		t.Errorf("Expected 1 selected tag, got %d", len(finalResult.SelectedTags))
	}
}

// TestCompleteFlowWithRefreshCallback integration test with refresh callback
func TestCompleteFlowWithRefreshCallback(t *testing.T) {
	originalTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     originalTagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Set refresh callback that returns updated list
	updatedTagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
			{
				Name:           "newfeature",
				TaskCount:      0,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-16",
				Active:         false,
			},
		},
	}

	flow.SetRefreshFunc(func(ctx context.Context) (*taskmaster.TagList, error) {
		return updatedTagList, nil
	})

	// Step 1: Select "Add New Tag"
	result := TagSelectorResult{
		SelectedTags: []string{},
		AddNewTag:    true,
	}
	flow.HandleSelectorResult(result, true)

	// Step 2: Handle new tag creation with refresh
	err := flow.HandleNewTagCreatedWithRefresh(context.Background(), "newfeature", updatedTagList)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(flow.lastTagList.Tags) != 2 {
		t.Errorf("Expected 2 tags after refresh, got %d", len(flow.lastTagList.Tags))
	}

	// Step 3: User selects newly created tag
	result2 := TagSelectorResult{
		SelectedTags: []string{"newfeature"},
		AddNewTag:    false,
	}

	shouldContinue, finalResult := flow.HandleSelectorResult(result2, true)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false")
	}
	if !flow.IsComplete() {
		t.Error("Expected flow to be complete")
	}
	if finalResult.SelectedTags[0] != "newfeature" {
		t.Errorf("Expected newfeature selected, got %v", finalResult.SelectedTags)
	}
}

// TestCompleteFlowWithError integration test with error handling and recovery
func TestCompleteFlowWithError(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Active:         true,
			},
		},
	}

	cfg := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	flow := NewTagEditorFlow(cfg, nil)

	// Step 1: Error occurs
	testErr := fmt.Errorf("selector error")
	flowErr := flow.HandleError("selector", "Failed to load", testErr)

	if flowErr == nil {
		t.Fatal("Expected error to be returned")
	}
	if flow.IsInErrorState() == false {
		t.Error("Expected flow to be in error state")
	}

	// Step 2: User initiates recovery
	flow.RecoverFromError()

	if flow.IsInErrorState() {
		t.Error("Expected flow to no longer be in error state")
	}
	if flow.currentState != StateTagSelector {
		t.Errorf("Expected state to be StateTagSelector, got %d", flow.currentState)
	}

	// Step 3: User successfully selects tags
	result := TagSelectorResult{
		SelectedTags: []string{"feature"},
		AddNewTag:    false,
	}

	shouldContinue, finalResult := flow.HandleSelectorResult(result, true)

	if shouldContinue {
		t.Error("Expected shouldContinue to be false")
	}
	if !flow.IsComplete() {
		t.Error("Expected flow to be complete after recovery and selection")
	}
	if len(finalResult.SelectedTags) != 1 {
		t.Errorf("Expected 1 selected tag, got %d", len(finalResult.SelectedTags))
	}
}
