package dialog

import (
	"context"
	"fmt"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TagEditorFlowState represents the current state of the flow
type TagEditorFlowState int

const (
	// StateTagSelector - showing tag selector dialog
	StateTagSelector TagEditorFlowState = iota
	// StateAddTagDialog - showing add tag dialog
	StateAddTagDialog
	// StateComplete - flow completed with selection
	StateComplete
	// StateCancelled - flow cancelled by user
	StateCancelled
	// StateError - flow encountered an error
	StateError
)

// TagRefreshFunc is a callback function to refresh tag list after creation
// It should return the updated tag list or an error
type TagRefreshFunc func(ctx context.Context) (*taskmaster.TagList, error)

// FlowError wraps errors that occur during tag editor flow
type FlowError struct {
	Stage   string // e.g., "selector", "add_tag", "refresh"
	Message string
	Err     error
}

// Error implements the error interface
func (e *FlowError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Tag editor flow error at %s: %s (details: %v)", e.Stage, e.Message, e.Err)
	}
	return fmt.Sprintf("Tag editor flow error at %s: %s", e.Stage, e.Message)
}

// Unwrap implements error unwrapping
func (e *FlowError) Unwrap() error {
	return e.Err
}

// TagEditorFlow manages the flow between tag selector and tag creation
type TagEditorFlow struct {
	currentState  TagEditorFlowState
	config        TagSelectorConfig
	selector      *TagSelector
	tagService    *taskmaster.Service
	refreshFunc   TagRefreshFunc
	result        TagSelectorResult
	cancelled     bool
	lastTagList   *taskmaster.TagList
	newTagCreated string
	lastError     error
	errorCount    int
}

// NewTagEditorFlow creates a new tag editor flow
func NewTagEditorFlow(cfg TagSelectorConfig, tagService *taskmaster.Service) *TagEditorFlow {
	return &TagEditorFlow{
		currentState: StateTagSelector,
		config:       cfg,
		tagService:   tagService,
		cancelled:    false,
		lastTagList:  cfg.TagList,
		errorCount:   0,
	}
}

// SetRefreshFunc sets the callback function to refresh tag list
// This should be called before HandleNewTagCreated to enable automatic list refresh
func (f *TagEditorFlow) SetRefreshFunc(fn TagRefreshFunc) {
	f.refreshFunc = fn
}

// ShowTagEditorFlow orchestrates the complete tag selection and creation flow
// It displays the tag selector initially, and if the user selects "Add New Tag",
// it opens a dialog for creating a new tag. After creation, it refreshes the
// tag list and returns to the selector.
//
// The flow continues until the user either:
// 1. Selects existing tags and confirms (returns TagSelectorResult, true)
// 2. Cancels at any step (returns empty result, false)
func (f *TagEditorFlow) GetInitialSelector() *TagSelector {
	if f.selector == nil {
		f.selector = NewTagSelector(f.config)
	}
	return f.selector
}

// HandleSelectorResult processes the result from tag selector dialog
// If AddNewTag is selected, returns true to indicate flow should open add tag dialog
// Otherwise completes the flow with the selected tags
func (f *TagEditorFlow) HandleSelectorResult(result TagSelectorResult, userConfirmed bool) (shouldContinue bool, finalResult TagSelectorResult) {
	if !userConfirmed {
		// User cancelled selector
		f.currentState = StateCancelled
		f.cancelled = true
		return false, TagSelectorResult{}
	}

	if result.AddNewTag {
		// User wants to add a new tag
		f.currentState = StateAddTagDialog
		return true, TagSelectorResult{}
	}

	// User selected existing tags
	f.currentState = StateComplete
	f.result = result
	return false, result
}

// RefreshTagList updates the tag list with the latest data from the service
// This is called after a new tag is created
// If a refresh function is set, it will be called to update the list
func (f *TagEditorFlow) RefreshTagList(ctx context.Context) (*taskmaster.TagList, error) {
	// If a refresh function is provided, use it
	if f.refreshFunc != nil {
		list, err := f.refreshFunc(ctx)
		if err != nil {
			return f.lastTagList, err
		}
		f.lastTagList = list
		f.config.TagList = list
		return list, nil
	}

	// If no service provided, return current list
	if f.tagService == nil {
		return f.lastTagList, nil
	}

	// Get fresh tag list from service using CLI
	// This would typically be called by the main model with proper context
	return f.lastTagList, nil
}

// HandleNewTagCreated updates the flow after a new tag is created
// It refreshes the tag list and resets for selector display
func (f *TagEditorFlow) HandleNewTagCreated(newTagName string, updatedTagList *taskmaster.TagList) {
	f.newTagCreated = newTagName
	f.lastTagList = updatedTagList
	f.config.TagList = updatedTagList

	// Create fresh selector with updated tag list
	f.selector = nil
	f.currentState = StateTagSelector
}

// HandleNewTagCreatedWithRefresh handles new tag creation and automatically refreshes the tag list
// It calls the refresh function if available, then updates the flow state
// If refresh fails, it uses the provided updated list as fallback
func (f *TagEditorFlow) HandleNewTagCreatedWithRefresh(ctx context.Context, newTagName string, updatedTagList *taskmaster.TagList) error {
	f.newTagCreated = newTagName

	var refreshErr error

	// Try to refresh tag list
	if f.refreshFunc != nil {
		list, err := f.refreshFunc(ctx)
		if err != nil {
			// On error, fall back to provided list and save the error
			f.lastTagList = updatedTagList
			f.config.TagList = updatedTagList
			refreshErr = err
		} else {
			f.lastTagList = list
			f.config.TagList = list
		}
	} else {
		f.lastTagList = updatedTagList
		f.config.TagList = updatedTagList
	}

	// Reset selector and return to selector state (regardless of refresh error)
	f.selector = nil
	f.currentState = StateTagSelector
	return refreshErr
}

// HandleAddTagCancelled resets the flow to show selector again
// when user cancels the add tag dialog
func (f *TagEditorFlow) HandleAddTagCancelled() {
	f.currentState = StateTagSelector
	f.newTagCreated = ""
}

// IsCancelled returns true if user cancelled the flow
func (f *TagEditorFlow) IsCancelled() bool {
	return f.cancelled
}

// IsComplete returns true if flow has completed successfully
func (f *TagEditorFlow) IsComplete() bool {
	return f.currentState == StateComplete
}

// GetCurrentState returns the current flow state
func (f *TagEditorFlow) GetCurrentState() TagEditorFlowState {
	return f.currentState
}

// GetResult returns the final result if flow is complete
func (f *TagEditorFlow) GetResult() TagSelectorResult {
	return f.result
}

// Cancel cancels the entire flow
func (f *TagEditorFlow) Cancel() {
	f.currentState = StateCancelled
	f.cancelled = true
	f.lastError = nil
	f.errorCount = 0
}

// HandleError records an error in the flow
// Returns true if flow should continue, false if it should stop
func (f *TagEditorFlow) HandleError(stage string, message string, err error) error {
	f.errorCount++
	flowErr := &FlowError{
		Stage:   stage,
		Message: message,
		Err:     err,
	}
	f.lastError = flowErr

	// Set error state
	f.currentState = StateError
	return flowErr
}

// RecoverFromError clears the error state and returns to the appropriate dialog
// For selector errors, returns to selector state
// For add tag errors, returns to selector state (user can retry or cancel)
func (f *TagEditorFlow) RecoverFromError() {
	f.lastError = nil
	// Reset to selector state to allow user to retry or cancel
	f.currentState = StateTagSelector
}

// GetLastError returns the last error that occurred in the flow
func (f *TagEditorFlow) GetLastError() error {
	return f.lastError
}

// GetErrorCount returns the number of errors that have occurred
func (f *TagEditorFlow) GetErrorCount() int {
	return f.errorCount
}

// IsInErrorState returns true if flow is currently in error state
func (f *TagEditorFlow) IsInErrorState() bool {
	return f.currentState == StateError
}
