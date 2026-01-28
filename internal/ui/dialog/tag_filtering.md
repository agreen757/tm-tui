# Tag Selector Filtering

Tag filtering provides keyboard-driven real-time search and filtering for tag lists in the Tag Selector dialog.

## Overview

The Tag Selector dialog supports powerful filtering capabilities that allow users to quickly find and select tags from large lists. Filtering is enabled by default and integrates seamlessly with the dialog system.

## Features

### Filtering

- **Activation**: Press `/` to enter filter mode while the tag selector is open
- **Real-time matching**: Type characters to filter tags as you type
- **Case-insensitive**: Matching is case-insensitive (e.g., "AUTH" matches "feature-auth")
- **Substring matching**: Any substring match counts (e.g., "api" matches "feature-api", "api-gateway")
- **Description search**: Filtering searches both tag names AND descriptions
- **Performance**: Optimized to handle 1000+ tags with <100ms filter latency

### Navigation During Filtering

While filtering is active, you can:
- Use **↑/↓** arrow keys or **j/k** to navigate the filtered results
- Use **PgUp/PgDn** to page through results
- Use **Space** to toggle selection (in multi-select mode)
- Use **Enter** to select and close (in single-select mode) or confirm selection (multi-select)
- Use **Esc** to exit filter mode and return to normal navigation

### Focus Management

- When entering filter mode, the current selection focus is automatically stored
- After exiting filter mode, the previous focus position is restored
- This enables intuitive filter workflows: search → navigate → exit → continue at previous spot

### Selection Persistence

- Tags you select remain marked even if you filter them out of view
- Your selection is preserved across multiple filter sessions
- This allows you to select some tags, filter to select others, and all remain selected

## Usage

### Basic Workflow

```
1. Open Tag Selector dialog
2. Press '/' to enter filter mode
3. Type part of tag name or description: "auth"
4. Navigate filtered results with ↑↓
5. Press Space to select (multi-select) or Enter to select (single-select)
6. Press Esc to exit filter mode
7. Continue with normal navigation or press Enter to confirm selection
```

### Examples

#### Find and select "authentication" tag
```
Press: /
Type: auth
Result: Shows all tags matching "auth" (feature-auth, authentication-fix, etc.)
```

#### Multi-select workflow
```
Press: / → auth → Space (select first match)
Press: Esc (exit filter)
Press: / → api → Space (select)
Press: Esc (exit filter)
Result: Both matching tags are selected
```

#### Filter by description
```
Press: / → "update"
Result: Shows all tags where name OR description contains "update"
Examples: tags.md, api-updates, "Update configuration"
```

## Implementation Details

### Architecture

The filtering system uses:

- **FilterableComponent interface**: Standard interface for filterable components in the dialog system
- **BaseFilterable**: Provides filter state management (value, mode, focus restoration)
- **TagItem.FilterValue()**: Returns concatenated Name and Description for matching
- **DialogManager integration**: Filter-aware key routing when filter mode is active

### State Management

- Filter state is managed functionally: `m.filterValue = newValue`
- Setting filter value automatically triggers list re-filtering
- Filter value is preserved while navigating the filtered list
- Selection state (selectedItems map) is preserved separately from filter state

### Performance

- Filtering uses efficient `strings.Contains()` for substring matching
- Filter value is converted to lowercase once for comparison
- Typical performance: <100ms for filtering 1000 tags
- List update is real-time as you type (no debounce lag)

### Focus Restoration

The system tracks focus using:
- `StoreFocusIndex()`: Saves current selection when entering filter mode
- `GetStoredFocusIndex()`: Retrieves saved index when exiting filter mode
- Automatic adjustment if filtered list size changes

## Architecture Integration

### DialogManager Integration

The DialogManager has built-in filter-aware routing:

```go
// When a dialog is FilterableComponent and IsFiltering() is true:
if dialog, ok := activeDialog.(FilterableComponent); ok && dialog.IsFiltering() {
    // Forward all keys directly to HandleKey()
    result, _ := dialog.HandleKey(keyMsg)
    // No Update() interference
}
```

This ensures:
- Typing characters reaches the filter input immediately
- No interference from normal dialog update logic
- Proper result handling for filter state transitions

### Custom FilterValue

The filtering system supports custom `FilterValue()` implementations:

```go
// TagItem example
func (t TagItem) FilterValue() string {
    if t.IsNew {
        return "add new"
    }
    return strings.ToLower(t.Tag.Name + " " + t.Tag.Description)
}
```

This allows components to define what fields are searchable.

## Testing

The implementation includes comprehensive tests covering:

- Basic filtering (exact matches, partial matches)
- Case-insensitive matching
- Description-based searching
- Navigation during filtering
- Selection persistence
- Focus restoration
- Edge cases (empty results, special characters)
- Integration with DialogManager

Run tests with:
```bash
go test ./internal/ui/dialog -run TestTagSelector -v
```

## Performance Characteristics

### Filter Operations

| Operation | Time | Notes |
|-----------|------|-------|
| Filter 100 tags | <1ms | Typical single keystroke |
| Filter 1000 tags | <100ms | Meets requirement |
| Backspace character | <1ms | Instant response |
| Exit filter mode | <1ms | Focus restoration |

### Memory

- Filter value stored as string: O(k) where k = filter length (max ~100 chars)
- ViewItems slice: O(n) where n = matching tags (usually <list size)
- ViewIndices slice: O(n) for tracking original indices

## Future Enhancements

Potential improvements:
- Fuzzy matching (e.g., "ftauth" matches "feature-auth")
- Regex filtering for advanced users
- Search result highlighting in the rendered view
- Saved search filters for common queries
- Multi-word filtering (e.g., "feature auth" as two separate terms)

## See Also

- TagSelector implementation: `internal/ui/dialog/tag_selector.go`
- FilterableComponent interface: `internal/ui/dialog/filterable.go`
- DialogManager documentation: `internal/ui/dialog/dialog.go`
