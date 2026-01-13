package dialog

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// TestTagSelectorStructInit tests the TagSelectorConfig and TagSelectorResult struct initialization
func TestTagSelectorStructInit(t *testing.T) {
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

	config := TagSelectorConfig{
		Title:       "Select Tags",
		MultiSelect: false,
		TagList:     tagList,
	}

	if config.Title != "Select Tags" {
		t.Errorf("Expected title 'Select Tags', got %s", config.Title)
	}
	if config.MultiSelect {
		t.Error("Expected MultiSelect to be false")
	}
	if len(config.TagList.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(config.TagList.Tags))
	}
}

// TestTagSelectorCreation tests basic dialog creation
func TestTagSelectorCreation(t *testing.T) {
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

	config := TagSelectorConfig{
		Title:       "Test Dialog",
		MultiSelect: false,
		TagList:     tagList,
	}

	selector := NewTagSelector(config)

	if selector.config.Title != "Test Dialog" {
		t.Errorf("Expected title 'Test Dialog', got %s", selector.config.Title)
	}
	if selector.Title() != "Test Dialog" {
		t.Errorf("Expected title 'Test Dialog', got %s", selector.Title())
	}

	// Should have the tag plus the "Add New Tag" option
	if len(selector.items) != 2 {
		t.Errorf("Expected 2 items (1 tag + 1 add-new), got %d", len(selector.items))
	}

	// Last item should be "Add New Tag"
	lastItem := selector.items[len(selector.items)-1]
	if !lastItem.IsNew {
		t.Error("Expected last item to be 'Add New Tag'")
	}
	if lastItem.Title() != "➕ Add New Tag..." {
		t.Errorf("Expected 'Add New Tag' title, got %s", lastItem.Title())
	}
}

// TestTagListRendering tests that tags are rendered with their metadata
func TestTagListRendering(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Description:    "New features",
				Active:         false,
			},
		},
	}

	config := TagSelectorConfig{
		Title:       "Test Dialog",
		MultiSelect: false,
		TagList:     tagList,
	}

	selector := NewTagSelector(config)

	// Check tag item properties
	tagItem := selector.items[0]
	if tagItem.Tag.Name != "feature" {
		t.Errorf("Expected tag name 'feature', got %s", tagItem.Tag.Name)
	}
	if tagItem.Title() != "feature" {
		t.Errorf("Expected title 'feature', got %s", tagItem.Title())
	}

	// Description should include metadata
	desc := tagItem.Description()
	if !containsStr(desc, "5 tasks") {
		t.Errorf("Expected '5 tasks' in description, got %s", desc)
	}
	if !containsStr(desc, "2 completed") {
		t.Errorf("Expected '2 completed' in description, got %s", desc)
	}
	if !containsStr(desc, "2024-01-15") {
		t.Errorf("Expected created label in description, got %s", desc)
	}
	if !containsStr(desc, "New features") {
		t.Errorf("Expected description in metadata, got %s", desc)
	}
}

// TestTagItemDescription tests tag item description formatting
func TestTagItemDescription(t *testing.T) {
	tests := []struct {
		name     string
		tag      taskmaster.TagContext
		expected []string // strings that should be in description
	}{
		{
			name: "tag with all metadata",
			tag: taskmaster.TagContext{
				Name:           "feature",
				TaskCount:      5,
				CompletedCount: 2,
				CreatedLabel:   "2024-01-15",
				Description:    "New features",
			},
			expected: []string{"5 tasks", "2 completed", "2024-01-15", "New features"},
		},
		{
			name: "tag with no completed tasks",
			tag: taskmaster.TagContext{
				Name:           "feature",
				TaskCount:      3,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-15",
				Description:    "Features",
			},
			expected: []string{"3 tasks", "2024-01-15", "Features"},
		},
		{
			name: "tag with minimal metadata",
			tag: taskmaster.TagContext{
				Name:           "tag",
				TaskCount:      1,
				CompletedCount: 0,
			},
			expected: []string{"1 task"},
		},
		{
			name: "tag with single task",
			tag: taskmaster.TagContext{
				Name:           "single",
				TaskCount:      1,
				CompletedCount: 0,
				CreatedLabel:   "2024-01-01",
			},
			expected: []string{"1 task", "2024-01-01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := TagItem{Tag: tt.tag, Index: 0, IsNew: false}
			desc := item.Description()

			for _, expected := range tt.expected {
				if !containsStr(desc, expected) {
					t.Errorf("Expected '%s' in description, got '%s'", expected, desc)
				}
			}
		})
	}
}

// TestEmptyTagList tests behavior with no tags
func TestEmptyTagList(t *testing.T) {
	config := TagSelectorConfig{
		Title:       "Empty Test",
		MultiSelect: false,
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{},
		},
	}

	selector := NewTagSelector(config)

	// Should only have the "Add New Tag" option
	if len(selector.items) != 1 {
		t.Errorf("Expected 1 item (only add-new), got %d", len(selector.items))
	}

	if !selector.items[0].IsNew {
		t.Error("Expected only item to be 'Add New Tag'")
	}
}

// TestTagListWithNilTagList tests behavior when TagList is nil
func TestTagListWithNilTagList(t *testing.T) {
	config := TagSelectorConfig{
		Title:       "Nil Test",
		MultiSelect: false,
		TagList:     nil,
	}

	selector := NewTagSelector(config)

	// Should only have the "Add New Tag" option
	if len(selector.items) != 1 {
		t.Errorf("Expected 1 item (only add-new), got %d", len(selector.items))
	}

	if !selector.items[0].IsNew {
		t.Error("Expected only item to be 'Add New Tag'")
	}
}

// TestDefaultTitle tests that default title is set when not provided
func TestDefaultTitle(t *testing.T) {
	config := TagSelectorConfig{
		Title:       "",
		MultiSelect: false,
		TagList:     &taskmaster.TagList{Tags: []taskmaster.TagContext{}},
	}

	selector := NewTagSelector(config)

	if selector.Title() != "Select Tags" {
		t.Errorf("Expected default title 'Select Tags', got '%s'", selector.Title())
	}
}

// TestTagItemFilterValue tests filter value for tags
func TestTagItemFilterValue(t *testing.T) {
	tag := taskmaster.TagContext{
		Name:        "feature",
		Description: "New features",
	}

	item := TagItem{Tag: tag, Index: 0, IsNew: false}
	filterValue := item.FilterValue()

	if !containsStr(filterValue, "feature") {
		t.Errorf("Expected 'feature' in filter value, got %s", filterValue)
	}
	if !containsStr(filterValue, "new features") {
		t.Errorf("Expected 'new features' in filter value, got %s", filterValue)
	}
}

// TestMultiSelectConfig tests multi-select configuration
func TestMultiSelectConfig(t *testing.T) {
	config := TagSelectorConfig{
		Title:       "Multi Select Test",
		MultiSelect: true,
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
				{Name: "tag2", TaskCount: 2},
			},
		},
	}

	selector := NewTagSelector(config)

	if !selector.multiSelect {
		t.Error("Expected MultiSelect to be true")
	}
}

// TestDefaultMultiSelectFalse tests that default is single-select
func TestDefaultMultiSelectFalse(t *testing.T) {
	config := TagSelectorConfig{
		Title:       "Single Select Test",
		MultiSelect: false,
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
			},
		},
	}

	selector := NewTagSelector(config)

	if selector.multiSelect {
		t.Error("Expected MultiSelect to be false by default")
	}
}

// TestTagSelectorInit tests the Init command
func TestTagSelectorInit(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: &taskmaster.TagList{},
	})

	cmd := selector.Init()
	if cmd != nil {
		t.Error("Expected Init to return nil")
	}
}

// TestTagSelectorKind tests that Kind returns DialogKindList
func TestTagSelectorKind(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: &taskmaster.TagList{},
	})

	if selector.Kind() != DialogKindList {
		t.Errorf("Expected DialogKindList, got %v", selector.Kind())
	}
}

// TestActiveTagIndicator tests that active tags are marked
func TestActiveTagIndicator(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:   "feature",
				Active: true,
			},
			{
				Name:   "bugfix",
				Active: false,
			},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Check that the Tag.Active field is set correctly
	activeItem := selector.items[0]
	if !activeItem.Tag.Active {
		t.Errorf("Expected first tag to be active")
	}

	inactiveItem := selector.items[1]
	if inactiveItem.Tag.Active {
		t.Errorf("Expected second tag to be inactive")
	}
}

// TestTagSelectorViewRendering tests that View returns a string
func TestTagSelectorViewRendering(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title: "Test",
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
			},
		},
	})

	view := selector.View()
	if view == "" {
		t.Error("Expected View to return non-empty string")
	}
	if !containsStr(view, "tag1") {
		t.Errorf("Expected 'tag1' in view, got: %s", view)
	}
}

// TestTagSelectorSetRect tests SetRect updates dialog dimensions
func TestTagSelectorSetRect(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: &taskmaster.TagList{},
	})

	selector.SetRect(100, 50, 10, 10)

	width, height, x, y := selector.GetRect()
	if width != 100 || height != 50 || x != 10 || y != 10 {
		t.Errorf("Expected (100,50,10,10), got (%d,%d,%d,%d)", width, height, x, y)
	}
}

// TestSingleSelectModeNavigation tests up/down arrow navigation in single-select mode
func TestSingleSelectModeNavigation(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1", TaskCount: 1},
			{Name: "tag2", TaskCount: 2},
			{Name: "tag3", TaskCount: 3},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: false,
		TagList:     tagList,
	})

	// Initially selected index should be 0
	if selector.selectedIndex != 0 {
		t.Errorf("Expected initial index 0, got %d", selector.selectedIndex)
	}

	// Test down navigation
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for down key, got %v", result)
	}
	if selector.selectedIndex != 1 {
		t.Errorf("Expected index 1 after down, got %d", selector.selectedIndex)
	}

	// Test down again
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if selector.selectedIndex != 2 {
		t.Errorf("Expected index 2 after second down, got %d", selector.selectedIndex)
	}

	// Test up navigation
	result, _ = selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for up key, got %v", result)
	}
	if selector.selectedIndex != 1 {
		t.Errorf("Expected index 1 after up, got %d", selector.selectedIndex)
	}

	// Test boundary - can't go past end
	selector.selectedIndex = len(selector.viewItems) - 1
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if selector.selectedIndex != len(selector.viewItems)-1 {
		t.Errorf("Expected index to stay at %d at boundary", len(selector.viewItems)-1)
	}

	// Test boundary - can't go before start
	selector.selectedIndex = 0
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if selector.selectedIndex != 0 {
		t.Errorf("Expected index to stay at 0 at boundary")
	}
}

// TestSingleSelectEnterConfirms tests that Enter confirms single selection
func TestSingleSelectEnterConfirms(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "feature", TaskCount: 1},
			{Name: "bugfix", TaskCount: 2},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: false,
		TagList:     tagList,
	})

	// Navigate to second tag
	selector.selectedIndex = 1

	// Press Enter to select
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm, got %v", result)
	}

	// Check that the tag was selected
	selectedTags := selector.GetSelectedTags()
	if len(selectedTags) != 1 {
		t.Errorf("Expected 1 selected tag, got %d", len(selectedTags))
	}
	if selectedTags[0] != "bugfix" {
		t.Errorf("Expected 'bugfix' to be selected, got %s", selectedTags[0])
	}
}

// TestEscapeCancel tests that Escape cancels the dialog
func TestEscapeCancel(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: &taskmaster.TagList{Tags: []taskmaster.TagContext{{Name: "tag1"}}},
	})

	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel, got %v", result)
	}

	// Nothing should be selected
	selectedTags := selector.GetSelectedTags()
	if len(selectedTags) != 0 {
		t.Errorf("Expected no selected tags after cancel, got %d", len(selectedTags))
	}
}

// TestGetSelectedTags verifies correct tags are returned
func TestGetSelectedTags(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
			{Name: "tag3"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: false,
		TagList:     tagList,
	})

	// Select first tag
	selector.selectedItems[0] = true

	selected := selector.GetSelectedTags()
	if len(selected) != 1 || selected[0] != "tag1" {
		t.Errorf("Expected ['tag1'], got %v", selected)
	}
}

// TestAddNewTagOption tests detection of add-new-tag selection
func TestAddNewTagOption(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Select the "Add New Tag" option (last item)
	lastIdx := len(selector.items) - 1
	selector.selectedItems[lastIdx] = true

	if !selector.HasAddNewTag() {
		t.Error("Expected HasAddNewTag to return true")
	}

	result := selector.GetResult()
	if !result.AddNewTag {
		t.Error("Expected AddNewTag in result to be true")
	}
}

// TestMultiSelectToggle tests Space key toggles selection in multi-select mode
func TestMultiSelectToggle(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
			{Name: "tag3"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: true,
		TagList:     tagList,
	})

	selector.selectedIndex = 0

	// First press Space - should toggle on
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got %v", result)
	}
	if !selector.selectedItems[0] {
		t.Error("Expected item 0 to be selected after Space")
	}

	// Second press Space - should toggle off
	result, _ = selector.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone on toggle off, got %v", result)
	}
	if selector.selectedItems[0] {
		t.Error("Expected item 0 to be deselected after second Space")
	}
}

// TestMultiSelectEnterConfirms tests Enter confirms multi-selection
func TestMultiSelectEnterConfirms(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
			{Name: "tag3"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: true,
		TagList:     tagList,
	})

	// Select multiple tags
	selector.selectedItems[0] = true
	selector.selectedItems[2] = true

	// Press Enter to confirm
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm, got %v", result)
	}

	// Check selected tags
	selected := selector.GetSelectedTags()
	if len(selected) != 2 {
		t.Errorf("Expected 2 selected tags, got %d", len(selected))
	}
}

// TestCannotSelectAddNewTag verifies "Add New Tag" cannot be toggled in multi-select
func TestCannotSelectAddNewTag(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: true,
		TagList:     tagList,
	})

	// Move to "Add New Tag" option
	selector.selectedIndex = len(selector.viewItems) - 1

	// Try to toggle with Space
	selector.HandleKey(tea.KeyMsg{Type: tea.KeySpace})

	// "Add New Tag" should not be selectable
	lastIdx := len(selector.items) - 1
	if selector.selectedItems[lastIdx] {
		t.Error("Expected 'Add New Tag' not to be selectable with Space")
	}
}

// TestMetadataDisplayInRendering tests that tag metadata is rendered correctly
func TestMetadataDisplayInRendering(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{
				Name:           "feature",
				TaskCount:      10,
				CompletedCount: 3,
				CreatedLabel:   "2024-01-15",
				Description:    "New features work",
				Active:         true,
			},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Get the rendered view
	view := selector.View()

	// Check that metadata appears in the view
	if !containsStr(view, "10 tasks") {
		t.Errorf("Expected task count in view, got: %s", view)
	}
	if !containsStr(view, "3 completed") {
		t.Errorf("Expected completed count in view, got: %s", view)
	}
	if !containsStr(view, "2024-01-15") {
		t.Errorf("Expected created label in view, got: %s", view)
	}
	if !containsStr(view, "feature") {
		t.Errorf("Expected tag name in view, got: %s", view)
	}
}

// TestActiveStatusIndicator tests that active tags show active indicator
func TestActiveStatusIndicator(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "current", Active: true},
			{Name: "archived", Active: false},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	view := selector.View()

	// Active tag should show (active) indicator
	if !containsStr(view, "current") || !containsStr(view, "active") {
		t.Errorf("Expected 'current (active)' in view, got: %s", view)
	}

	// Inactive tag should not show (active)
	if containsStr(view, "archived (active)") {
		t.Errorf("Unexpected '(active)' for inactive tag in view: %s", view)
	}
}

// TestSelectionIndicators tests that selection checkboxes are rendered
func TestSelectionIndicators(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: true,
		TagList:     tagList,
	})

	// Select first tag
	selector.selectedItems[0] = true

	view := selector.View()

	// View should contain checkbox indicators
	if !containsStr(view, "☑") && !containsStr(view, "☐") {
		t.Errorf("Expected checkbox indicators in view, got: %s", view)
	}
}

// TestEmptyDescriptionHandling tests tags without descriptions
func TestEmptyDescriptionHandling(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "minimal", TaskCount: 1},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	item := selector.items[0]
	desc := item.Description()

	// Description should at least have task count
	if !containsStr(desc, "1 task") {
		t.Errorf("Expected '1 task' in description, got: %s", desc)
	}
}

// TestMultipleTagsDisplay tests rendering of multiple tags
func TestMultipleTagsDisplay(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1", TaskCount: 5},
			{Name: "tag2", TaskCount: 10},
			{Name: "tag3", TaskCount: 15},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	if len(selector.items) != 4 {
		t.Errorf("Expected 4 items (3 tags + 1 add-new), got %d", len(selector.items))
	}

	// Verify all tags are present
	view := selector.View()
	if !containsStr(view, "tag1") || !containsStr(view, "tag2") || !containsStr(view, "tag3") {
		t.Errorf("Expected all tags in view, got: %s", view)
	}
}

// TestBoundaryNavigationFirstItem tests navigation at first item boundary
func TestBoundaryNavigationFirstItem(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Start at index 0 and try to go up
	if selector.selectedIndex != 0 {
		t.Fatal("Expected initial index 0")
	}

	selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})

	// Should stay at 0
	if selector.selectedIndex != 0 {
		t.Errorf("Expected to stay at 0 at boundary, got %d", selector.selectedIndex)
	}
}

// TestBoundaryNavigationLastItem tests navigation at last item boundary
func TestBoundaryNavigationLastItem(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Move to last item (Add New Tag)
	selector.selectedIndex = len(selector.viewItems) - 1

	// Try to go down
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Should stay at last item
	if selector.selectedIndex != len(selector.viewItems)-1 {
		t.Errorf("Expected to stay at last item")
	}
}

// TestUpArrowKeyNavigation tests up arrow key for navigation
func TestUpArrowKeyNavigation(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	selector.selectedIndex = 1

	// Test up arrow key
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyUp})

	if selector.selectedIndex != 0 {
		t.Errorf("Expected index 0 after up arrow, got %d", selector.selectedIndex)
	}
}

// TestDownArrowKeyNavigation tests down arrow key for navigation
func TestDownArrowKeyNavigation(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	selector.selectedIndex = 0

	// Test down arrow key
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyDown})

	if selector.selectedIndex != 1 {
		t.Errorf("Expected index 1 after down arrow, got %d", selector.selectedIndex)
	}
}

// TestAddNewTagSelection tests selecting the "Add New Tag" option
func TestAddNewTagSelection(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Navigate to Add New Tag
	selector.selectedIndex = len(selector.viewItems) - 1

	// Press Enter
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm, got %v", result)
	}

	// Check result
	res := selector.GetResult()
	if !res.AddNewTag {
		t.Error("Expected AddNewTag to be true in result")
	}
}

// TestNoSelectionResult tests empty selection result
func TestNoSelectionResult(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Don't select anything, just get result
	result := selector.GetResult()

	if len(result.SelectedTags) != 0 {
		t.Errorf("Expected no selected tags, got %d", len(result.SelectedTags))
	}
	if result.AddNewTag {
		t.Error("Expected AddNewTag to be false")
	}
}

// TestMixedSelectionWithAddNewTag tests selecting regular tags and Add New Tag
func TestMixedSelectionWithAddNewTag(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: true,
		TagList:     tagList,
	})

	// Select first tag
	selector.selectedItems[0] = true

	// Try to select Add New Tag (should fail)
	lastIdx := len(selector.items) - 1
	selector.selectedIndex = lastIdx
	selector.HandleKey(tea.KeyMsg{Type: tea.KeySpace})

	// Check results
	selected := selector.GetSelectedTags()
	if len(selected) != 1 || selected[0] != "tag1" {
		t.Errorf("Expected ['tag1'], got %v", selected)
	}

	if selector.selectedItems[lastIdx] {
		t.Error("Expected Add New Tag not to be selectable")
	}
}

// TestLargeTagListNavigation tests navigation in a large list
func TestLargeTagListNavigation(t *testing.T) {
	tags := make([]taskmaster.TagContext, 50)
	for i := 0; i < 50; i++ {
		tags[i] = taskmaster.TagContext{Name: fmt.Sprintf("tag%d", i)}
	}

	tagList := &taskmaster.TagList{Tags: tags}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Navigate to middle
	for i := 0; i < 25; i++ {
		selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}

	if selector.selectedIndex != 25 {
		t.Errorf("Expected index 25 after 25 down presses, got %d", selector.selectedIndex)
	}

	// Navigate back up
	for i := 0; i < 10; i++ {
		selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	}

	if selector.selectedIndex != 15 {
		t.Errorf("Expected index 15 after 10 up presses, got %d", selector.selectedIndex)
	}
}

// TestDialogVisibility tests that dialog rendering produces output
func TestDialogVisibility(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1"},
			{Name: "tag2"},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Select Tags",
		TagList: tagList,
	})

	// Test View rendering
	view := selector.View()

	if view == "" {
		t.Error("Expected View to return non-empty string")
	}

	// Check title appears
	if !containsStr(view, "Select Tags") {
		t.Errorf("Expected title in view, got: %s", view)
	}

	// Check tags appear
	if !containsStr(view, "tag1") {
		t.Errorf("Expected tag1 in view, got: %s", view)
	}
}

// TestDialogUpdate processes window size changes
func TestDialogUpdate(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: &taskmaster.TagList{},
	})

	// Test update with window size
	dialog, _ := selector.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	s := dialog.(*TagSelector)
	if s.selectedIndex != 0 {
		t.Error("Expected selected index to be preserved")
	}
}

// TestUpdateMessage verifies Update returns same dialog
func TestUpdateMessage(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: &taskmaster.TagList{},
	})

	d, _ := selector.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if d != selector {
		t.Error("Expected Update to return same dialog pointer")
	}
}

// TestUpdateWithKeyMsgNavigation tests that Update does not process keyboard navigation
// (keyboard navigation is handled by HandleKey, called by DialogManager)
func TestUpdateWithKeyMsgNavigation(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title: "Test Tags",
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
				{Name: "tag2", TaskCount: 2},
				{Name: "tag3", TaskCount: 3},
			},
		},
		MultiSelect: false,
	})

	// Initial state
	initialIndex := selector.selectedIndex

	// Update with key message should not change selectedIndex
	// (DialogManager will call HandleKey separately)
	dialog, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Error("Expected no command from Update for key messages")
	}
	s := dialog.(*TagSelector)
	if s.selectedIndex != initialIndex {
		t.Errorf("Expected Update to not process key messages, selectedIndex changed from %d to %d", initialIndex, s.selectedIndex)
	}

	// Use HandleKey directly to test navigation
	result, cmd := selector.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got %v", result)
	}
	if selector.selectedIndex != 1 {
		t.Errorf("Expected selected index 1 after HandleKey down arrow, got %d", selector.selectedIndex)
	}
}

// TestUpdateWithKeyMsgConfirm tests that Update does not process confirmation
// (confirmation is handled by HandleKey, called by DialogManager)
func TestUpdateWithKeyMsgConfirm(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title: "Test Tags",
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
			},
		},
		MultiSelect: false,
	})

	// Update with Enter key should return the dialog (not nil)
	// DialogManager will call HandleKey which returns DialogResultConfirm
	dialog, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if dialog == nil {
		t.Error("Expected Update to return dialog, not nil")
	}
	if cmd != nil {
		t.Error("Expected no command from Update")
	}

	// Use HandleKey to test confirmation
	result, cmd := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm from HandleKey, got %v", result)
	}
}

// TestUpdateWithKeyMsgCancel tests that Update does not process cancellation
// (cancellation is handled by HandleKey, called by DialogManager)
func TestUpdateWithKeyMsgCancel(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title: "Test Tags",
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
			},
		},
		MultiSelect: false,
	})

	// Update with Esc key should return the dialog (not nil)
	// DialogManager will call HandleKey which returns DialogResultCancel
	dialog, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if dialog == nil {
		t.Error("Expected Update to return dialog, not nil")
	}
	if cmd != nil {
		t.Error("Expected no command from Update")
	}

	// Use HandleKey to test cancellation
	result, cmd := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel from HandleKey, got %v", result)
	}
}

// TestUpdateWithKeyMsgMultiSelectToggle tests that Update does not process space key
// (multi-select toggle is handled by HandleKey, called by DialogManager)
func TestUpdateWithKeyMsgMultiSelectToggle(t *testing.T) {
	selector := NewTagSelector(TagSelectorConfig{
		Title: "Test Tags",
		TagList: &taskmaster.TagList{
			Tags: []taskmaster.TagContext{
				{Name: "tag1", TaskCount: 1},
				{Name: "tag2", TaskCount: 2},
			},
		},
		MultiSelect: true,
	})

	// Update with space key should not change selectedItems
	// (DialogManager will call HandleKey separately)
	dialog, cmd := selector.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cmd != nil {
		t.Error("Expected no command from Update")
	}
	s := dialog.(*TagSelector)
	if s.selectedItems[0] {
		t.Error("Expected Update to not process space key, item 0 was selected")
	}

	// Use HandleKey to test toggle
	result, cmd := selector.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone, got %v", result)
	}
	if !selector.selectedItems[0] {
		t.Error("Expected first item to be selected after HandleKey space")
	}

	// Navigate down with HandleKey
	selector.HandleKey(tea.KeyMsg{Type: tea.KeyDown})

	// Toggle second item with HandleKey
	selector.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if !selector.selectedItems[1] {
		t.Error("Expected second item to be selected after HandleKey space")
	}
}

// TestDialogResultValueImplementation tests that TagSelector implements DialogResultProvider
func TestDialogResultValueImplementation(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1", TaskCount: 1},
			{Name: "tag2", TaskCount: 2},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: false,
		TagList:     tagList,
	})

	// Select first tag
	selector.selectedItems[0] = true

	// Get result via DialogResultValue
	value, err := selector.DialogResultValue()
	if err != nil {
		t.Errorf("Expected no error from DialogResultValue, got %v", err)
	}

	// Verify result type
	result, ok := value.(TagSelectorResult)
	if !ok {
		t.Errorf("Expected TagSelectorResult type, got %T", value)
	}

	// Verify result content
	if len(result.SelectedTags) != 1 {
		t.Errorf("Expected 1 selected tag, got %d", len(result.SelectedTags))
	}
	if result.SelectedTags[0] != "tag1" {
		t.Errorf("Expected 'tag1', got %s", result.SelectedTags[0])
	}
	if result.AddNewTag {
		t.Error("Expected AddNewTag to be false")
	}
}

// TestDialogResultValueWithAddNewTag tests DialogResultValue when Add New Tag is selected
func TestDialogResultValueWithAddNewTag(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1", TaskCount: 1},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Select "Add New Tag" option
	lastIdx := len(selector.items) - 1
	selector.selectedItems[lastIdx] = true

	// Get result via DialogResultValue
	value, err := selector.DialogResultValue()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result, ok := value.(TagSelectorResult)
	if !ok {
		t.Errorf("Expected TagSelectorResult, got %T", value)
	}

	// Verify AddNewTag is true
	if !result.AddNewTag {
		t.Error("Expected AddNewTag to be true")
	}
	if len(result.SelectedTags) != 0 {
		t.Errorf("Expected no selected tags, got %d", len(result.SelectedTags))
	}
}

// TestDialogResultValueMultiSelect tests DialogResultValue with multiple selections
func TestDialogResultValueMultiSelect(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1", TaskCount: 1},
			{Name: "tag2", TaskCount: 2},
			{Name: "tag3", TaskCount: 3},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:       "Test",
		MultiSelect: true,
		TagList:     tagList,
	})

	// Select multiple tags
	selector.selectedItems[0] = true
	selector.selectedItems[2] = true

	// Get result via DialogResultValue
	value, err := selector.DialogResultValue()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result, ok := value.(TagSelectorResult)
	if !ok {
		t.Errorf("Expected TagSelectorResult, got %T", value)
	}

	// Verify multiple selections
	if len(result.SelectedTags) != 2 {
		t.Errorf("Expected 2 selected tags, got %d", len(result.SelectedTags))
	}

	// Check tag names (order may vary)
	tagNames := make(map[string]bool)
	for _, name := range result.SelectedTags {
		tagNames[name] = true
	}
	if !tagNames["tag1"] || !tagNames["tag3"] {
		t.Errorf("Expected tag1 and tag3, got %v", result.SelectedTags)
	}
}

// TestDialogResultValueNoSelection tests DialogResultValue with no selection
func TestDialogResultValueNoSelection(t *testing.T) {
	tagList := &taskmaster.TagList{
		Tags: []taskmaster.TagContext{
			{Name: "tag1", TaskCount: 1},
		},
	}

	selector := NewTagSelector(TagSelectorConfig{
		Title:   "Test",
		TagList: tagList,
	})

	// Don't select anything
	value, err := selector.DialogResultValue()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result, ok := value.(TagSelectorResult)
	if !ok {
		t.Errorf("Expected TagSelectorResult, got %T", value)
	}

	// Verify empty result
	if len(result.SelectedTags) != 0 {
		t.Errorf("Expected no selected tags, got %d", len(result.SelectedTags))
	}
	if result.AddNewTag {
		t.Error("Expected AddNewTag to be false")
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsStr(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 && strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
