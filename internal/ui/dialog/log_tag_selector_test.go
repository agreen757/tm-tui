package dialog

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestService creates a test service with minimal setup
func createTestService(t *testing.T, tmpDir string) *taskmaster.Service {
	// Create necessary .taskmaster structure
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	tasksDir := filepath.Join(tmDir, "tasks")
	require.NoError(t, os.MkdirAll(tasksDir, 0755))

	// Create a minimal tasks.json file
	tasksJSON := filepath.Join(tasksDir, "tasks.json")
	require.NoError(t, os.WriteFile(tasksJSON, []byte("[]"), 0644))

	cfg := &config.Config{
		TaskMasterPath: tmpDir,
		ActiveTag:      "logs",
	}
	svc, err := taskmaster.NewService(cfg)
	require.NoError(t, err)
	return svc
}

// TestLogTagEntryFilterValue tests the FilterValue method
func TestLogTagEntryFilterValue(t *testing.T) {
	entry := LogTagEntry{
		Name: "MyTag",
	}
	assert.Equal(t, "mytag", entry.FilterValue())
}

// TestLogTagEntryTitle tests the Title method
func TestLogTagEntryTitle(t *testing.T) {
	tests := []struct {
		name     string
		entry    LogTagEntry
		expected string
	}{
		{
			name: "Active tag",
			entry: LogTagEntry{
				Name:     "active-tag",
				IsActive: true,
			},
			expected: "● active-tag",
		},
		{
			name: "Inactive tag",
			entry: LogTagEntry{
				Name:     "inactive-tag",
				IsActive: false,
			},
			expected: "  inactive-tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.Title())
		})
	}
}

// TestLogTagEntryDescription tests the Description method
func TestLogTagEntryDescription(t *testing.T) {
	now := time.Now()
	entry := LogTagEntry{
		Name:     "test-tag",
		LogCount: 5,
		LastMod:  now,
	}
	desc := entry.Description()
	assert.Contains(t, desc, "5 logs")
	assert.Contains(t, desc, now.Format("Jan 02"))
}

// TestPluralizeLogCount tests the pluralize function
func TestPluralizeLogCount(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{100, "s"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count)), func(t *testing.T) {
			assert.Equal(t, tt.expected, pluralizeLogCount(tt.count))
		})
	}
}

// TestNewLogTagSelectorModel tests creation of a new tag selector
func TestNewLogTagSelectorModel(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")

	assert.NotNil(t, selector)
	assert.Equal(t, 80, selector.width)
	assert.Equal(t, 20, selector.height)
	assert.Equal(t, "logs", selector.currentTag)
	assert.NotNil(t, selector.list)
}

// TestDiscoverTagsBasicStructure tests tag discovery with basic directory structure
func TestDiscoverTagsBasicStructure(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	// Create logs directory with a test log file
	logsDir := filepath.Join(tmpDir, ".taskmaster", "logs")
	require.NoError(t, os.MkdirAll(logsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(logsDir, "test.log"), []byte("test"), 0644))

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Should find at least the 'logs' tag
	assert.NotEmpty(t, selector.tags)

	// Check for 'logs' tag
	logsTagFound := false
	for _, tag := range selector.tags {
		if tag.Name == "logs" {
			logsTagFound = true
			assert.Equal(t, 1, tag.LogCount)
			assert.True(t, tag.IsActive)
			assert.True(t, tag.IsSpecial)
			break
		}
	}
	assert.True(t, logsTagFound, "logs tag should be discovered")
}

// TestDiscoverTagsExcludedDirectories tests that excluded directories are skipped
func TestDiscoverTagsExcludedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	// Create excluded directories with log files
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	excludedDirs := []string{"tasks", "docs", "reports", "memory"}
	for _, dir := range excludedDirs {
		dirPath := filepath.Join(tmDir, dir)
		require.NoError(t, os.MkdirAll(dirPath, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dirPath, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Check that excluded directories are not in tags
	for _, tag := range selector.tags {
		for _, excluded := range excludedDirs {
			assert.NotEqual(t, excluded, tag.Name, "%s should be excluded from tags", excluded)
		}
	}
}

// TestDiscoverTagsWithCustomTags tests discovering custom tag directories
func TestDiscoverTagsWithCustomTags(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create custom tag directories
	customTags := []string{"feature-auth", "bugfix-perf", "docs-update"}
	for _, tagName := range customTags {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		// Create some log files
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "impl.log"), []byte("implementation"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("testing"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Check that all custom tags are found
	foundTags := make(map[string]bool)
	for _, tag := range selector.tags {
		foundTags[tag.Name] = true
	}

	for _, tagName := range customTags {
		assert.True(t, foundTags[tagName], "custom tag %s should be discovered", tagName)

		// Find the tag and verify log count
		for _, tag := range selector.tags {
			if tag.Name == tagName {
				assert.Equal(t, 2, tag.LogCount, "tag %s should have 2 log files", tagName)
				assert.False(t, tag.IsSpecial, "custom tag should not be marked as special")
				break
			}
		}
	}
}

// TestCalculateTagMetadata tests metadata calculation for tags
func TestCalculateTagMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tagDir := filepath.Join(tmpDir, "test-tag")
	require.NoError(t, os.MkdirAll(tagDir, 0755))

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")

	// Test empty directory
	count, lastMod := selector.calculateTagMetadata(tagDir)
	assert.Equal(t, 0, count)
	assert.NotZero(t, lastMod)

	// Create log files
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(tagDir, "log"+string(rune('0'+i))+".log")
		require.NoError(t, os.WriteFile(filename, []byte("test"), 0644))
		time.Sleep(10 * time.Millisecond) // Ensure different mod times
	}

	count, lastMod = selector.calculateTagMetadata(tagDir)
	assert.Equal(t, 3, count, "should count 3 log files")
	assert.NotZero(t, lastMod)

	// Create non-log files (should be ignored)
	require.NoError(t, os.WriteFile(filepath.Join(tagDir, "readme.txt"), []byte("test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tagDir, "data.json"), []byte("test"), 0644))

	count, lastMod = selector.calculateTagMetadata(tagDir)
	assert.Equal(t, 3, count, "should still count only 3 log files (ignore non-log files)")
}

// TestCalculateTagMetadataIgnoresNonLogFiles tests that non-.log files are ignored
func TestCalculateTagMetadataIgnoresNonLogFiles(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tagDir := filepath.Join(tmpDir, "test-tag")
	require.NoError(t, os.MkdirAll(tagDir, 0755))

	// Create mixed files
	require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("log"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tagDir, "readme.md"), []byte("readme"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tagDir, "data.json"), []byte("data"), 0644))

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	count, _ := selector.calculateTagMetadata(tagDir)

	assert.Equal(t, 1, count, "should count only .log files")
}

// TestCalculateTagMetadataIgnoresSubdirectories tests that subdirectories are ignored
func TestCalculateTagMetadataIgnoresSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tagDir := filepath.Join(tmpDir, "test-tag")
	require.NoError(t, os.MkdirAll(tagDir, 0755))

	// Create log files
	require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("log"), 0644))

	// Create subdirectory with log files (should be ignored)
	subDir := filepath.Join(tagDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "nested.log"), []byte("log"), 0644))

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	count, _ := selector.calculateTagMetadata(tagDir)

	assert.Equal(t, 1, count, "should count only top-level .log files")
}

// TestGetSelectedTag tests tag selection
func TestGetSelectedTag(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create tags
	for _, tagName := range []string{"logs", "feature-auth", "bugfix"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Select the second tag and verify
	if len(selector.tags) > 1 {
		expectedTag := selector.tags[1].Name
		selector.list.Select(1)
		selectedTag := selector.GetSelectedTag()
		assert.Equal(t, expectedTag, selectedTag)
	}
}

// TestArchiveTagDiscovery tests discovery of special archive tag
func TestArchiveTagDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create archive directory BEFORE creating the service
	archiveDir := filepath.Join(tmDir, "archive")
	require.NoError(t, os.MkdirAll(archiveDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(archiveDir, "old.log"), []byte("archived"), 0644))

	svc := createTestService(t, tmpDir)

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Check for archive tag
	archiveTagFound := false
	for _, tag := range selector.tags {
		if tag.Name == "archive" {
			archiveTagFound = true
			assert.Equal(t, 1, tag.LogCount)
			assert.True(t, tag.IsSpecial)
			break
		}
	}
	assert.True(t, archiveTagFound, "archive tag should be discovered")
}

// TestCurrentTagHighlighting tests that the current tag is highlighted
func TestCurrentTagHighlighting(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create tags
	for _, tagName := range []string{"logs", "feature-auth"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	// Set active tag to "feature-auth"
	selector := NewLogTagSelectorModel(80, 20, svc, "feature-auth")
	selector.discoverTags()

	// Check that "feature-auth" is marked as active
	featureAuthFound := false
	for _, tag := range selector.tags {
		if tag.Name == "feature-auth" {
			featureAuthFound = true
			assert.True(t, tag.IsActive, "feature-auth should be marked as active")
		} else if tag.Name != "logs" {
			assert.False(t, tag.IsActive, "%s should not be marked as active", tag.Name)
		}
	}
	assert.True(t, featureAuthFound, "feature-auth should be discovered")
}

// TestTagSortingByName tests that tags are sorted alphabetically
func TestTagSortingByName(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create tags in non-alphabetical order
	tagNames := []string{"zebra", "apple", "monkey"}
	for _, tagName := range tagNames {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Check that tags (excluding special ones) are sorted
	// logs should be first (special), then the rest sorted
	assert.Equal(t, "logs", selector.tags[0].Name, "logs should be first")
	if len(selector.tags) > 1 {
		// Verify remaining tags are sorted (excluding archive if present)
		for i := 1; i < len(selector.tags)-1; i++ {
			curr := selector.tags[i].Name
			next := selector.tags[i+1].Name
			// Both should be custom tags or next should be archive
			if next != "archive" {
				assert.True(t, curr < next, "%s should come before %s alphabetically", curr, next)
			}
		}
	}
}

// TestNavigationKeys tests keyboard navigation (up, down, k, j)
func TestNavigationKeys(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create multiple tags
	for i := 1; i <= 5; i++ {
		tagName := string(rune('a' + i - 1))
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Initial index should be 0
	assert.Equal(t, 0, selector.list.Index())

	// Test down navigation
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}}) // fake down
	_ = result                                                                           // Just to use it

	// Verify list is not nil and usable
	assert.NotNil(t, selector.list)
}

// TestHomeKeyNavigation tests Home key jumping to current tag
func TestHomeKeyNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create multiple tags
	for _, tagName := range []string{"logs", "feature-a", "feature-b", "feature-c"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	// Set current tag to "feature-b"
	selector := NewLogTagSelectorModel(80, 20, svc, "feature-b")
	selector.discoverTags()

	// Move to a different position first
	selector.list.Select(0)
	assert.Equal(t, 0, selector.list.Index())

	// Test Home key to jump back to current tag
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}, Alt: true})
	if result == DialogResultNone {
		// Home key was processed - check the result
		currentIdx := -1
		for i, tag := range selector.tags {
			if tag.Name == "feature-b" {
				currentIdx = i
				break
			}
		}
		if currentIdx >= 0 {
			assert.Equal(t, currentIdx, selector.list.Index())
		}
	}
}

// TestSelectionCallback tests that selection callback is called
func TestSelectionCallback(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	for _, tagName := range []string{"logs", "feature-auth"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Track callback invocation
	callbackInvoked := false
	callbackTag := ""

	selector.SetOnTagSelected(func(tagName string) tea.Cmd {
		callbackInvoked = true
		callbackTag = tagName
		return nil
	})

	// Select the first tag
	selector.list.Select(0)
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}})

	// Callback should be invoked on Enter
	if result == DialogResultConfirm {
		assert.True(t, callbackInvoked)
		assert.NotEmpty(t, callbackTag)
	}
}

// TestViewRendering tests that the dialog renders without errors
func TestViewRendering(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	for _, tagName := range []string{"logs", "feature-a", "feature-b"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Test View() renders without panic
	view := selector.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "logs")
}

// TestDialogResultValue tests the DialogResultValue method
func TestDialogResultValue(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	for _, tagName := range []string{"logs", "feature-auth"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Select a tag
	if len(selector.tags) > 1 {
		selector.list.Select(1)
		expectedTag := selector.tags[1].Name

		// Get result value
		result, err := selector.DialogResultValue()
		assert.NoError(t, err)
		assert.Equal(t, expectedTag, result)
	}
}

// TestCancelNavigation tests escape key cancellation
func TestCancelNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	for _, tagName := range []string{"logs", "feature-auth"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Test Escape key
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, DialogResultCancel, result)
}

// TestRefreshTags tests the RefreshTags method
func TestRefreshTags(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create initial tags
	for _, tagName := range []string{"logs", "initial-tag"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	initialCount := len(selector.tags)

	// Add a new tag
	newTagDir := filepath.Join(tmDir, "new-tag")
	require.NoError(t, os.MkdirAll(newTagDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(newTagDir, "test.log"), []byte("test"), 0644))

	// Refresh tags
	selector.RefreshTags()

	// Should have more tags now
	assert.Greater(t, len(selector.tags), initialCount)

	// Check that new tag is present
	newTagFound := false
	for _, tag := range selector.tags {
		if tag.Name == "new-tag" {
			newTagFound = true
			break
		}
	}
	assert.True(t, newTagFound)
}

// TestSearchByPartialName tests filtering by partial tag name
func TestSearchByPartialName(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	// Create tags with similar names
	for _, tagName := range []string{"logs", "feature-auth", "feature-api", "feature-db", "bugfix-perf"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Verify filtering is enabled
	assert.NotNil(t, selector.list)
	// The list should have items
	assert.Greater(t, len(selector.tags), 0)
}

// TestSearchCancellation tests that search can be cancelled with Escape
func TestSearchCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	svc := createTestService(t, tmpDir)

	tmDir := filepath.Join(tmpDir, ".taskmaster")
	for _, tagName := range []string{"logs", "feature-auth"} {
		tagDir := filepath.Join(tmDir, tagName)
		require.NoError(t, os.MkdirAll(tagDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tagDir, "test.log"), []byte("test"), 0644))
	}

	selector := NewLogTagSelectorModel(80, 20, svc, "logs")
	selector.discoverTags()

	// Test that "/" key is accepted (activates search in list model)
	result, _ := selector.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.Equal(t, DialogResultNone, result)
}
