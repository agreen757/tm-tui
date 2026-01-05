# Path Utilities Package

The `pathutil` package provides unified, canonical path utilities for the Task Master TUI application, with a focus on PRD (Product Requirements Document) directory resolution.

## Overview

This package consolidates path resolution logic into a single source of truth, ensuring consistent behavior across the application. Previously, path resolution was duplicated across multiple modules (`create_prd.go`, `parse_prd.go`, `prd_form.go`), leading to inconsistencies and maintenance difficulties.

## Core Functions

### ResolvePrdDirectoryPath

```go
func ResolvePrdDirectoryPath(cfg *config.Config, lastUsedPath string) string
```

**Purpose**: Determines the appropriate PRD directory path using a priority-based fallback chain, without creating any directories.

**Parameters**:
- `cfg *config.Config` - Application configuration (may be nil)
- `lastUsedPath string` - Previously used path from user interaction (may be empty)

**Returns**: A valid path string (always succeeds, never returns empty string)

**Priority Chain**:
1. **Last Used Path** - If `lastUsedPath` is non-empty, it is returned directly
   - Provides better UX by remembering where users last accessed PRD files
   - Example: User previously selected `/projects/acme/prd/`, that path is suggested next time

2. **TaskMasterPath/.taskmaster/docs** - If `config.TaskMasterPath` is set and the directory exists
   - The canonical location within a Task Master project
   - Example: `/projects/acme/.taskmaster/docs`

3. **TaskMasterPath** - If `config.TaskMasterPath` is set and exists as a directory
   - Fallback when `.taskmaster/docs` doesn't exist yet
   - Example: `/projects/acme`

4. **Default: .taskmaster/docs** - Relative path in current directory
   - Used when no config is available or TaskMasterPath doesn't exist
   - Works in development or single-project contexts

5. **Current Working Directory** - Implicit fallback via relative path
   - The relative path `.taskmaster/docs` is always valid relative to CWD

**No Error Handling**: This function never fails - it always returns a valid path that can be used with filesystem operations.

### GetPrdDirectory

```go
func GetPrdDirectory(cfg *config.Config, lastUsedPath string) (string, error)
```

**Purpose**: Resolves the PRD directory path using `ResolvePrdDirectoryPath`, then ensures the directory exists by creating it if necessary.

**Parameters**: Same as `ResolvePrdDirectoryPath`

**Returns**:
- `string` - The canonical PRD directory path
- `error` - Error if directory creation fails (nil on success)

**Error Messages**: Errors are wrapped with context to help debugging:
```
failed to create PRD directory: permission denied
failed to create PRD directory: no such file or directory
```

**Directory Permissions**: Created directories use `0755` permissions (rwxr-xr-x), allowing the user to read/write and others to read/execute.

## Usage Examples

### PRD Creation Flow

When creating a PRD, use `GetPrdDirectory` to ensure the destination exists:

```go
import "github.com/agreen757/tm-tui/internal/pathutil"

// In PRD creation handler
docsDir, err := pathutil.GetPrdDirectory(config, "")
if err != nil {
    return fmt.Errorf("failed to prepare PRD directory: %w", err)
}

// Now safe to write to docsDir
filePath := filepath.Join(docsDir, filename)
if err := os.WriteFile(filePath, content, 0644); err != nil {
    return err
}
```

### PRD Parsing/Selection Flow

When selecting a PRD file to parse, use `ResolvePrdDirectoryPath` to suggest a starting directory for file browsing:

```go
import "github.com/agreen757/tm-tui/internal/pathutil"

// In PRD file selection handler
startDir := pathutil.ResolvePrdDirectoryPath(config, lastPrdPath)
fileDialog := dialog.NewFileSelectionDialog(
    "Select PRD File",
    startDir,
    78, 20,
    []string{".md", ".txt"},
)

// User can browse from startDir and select a file
```

### Path Display in Forms

When displaying the expected destination to the user:

```go
import "github.com/agreen757/tm-tui/internal/pathutil"

// In form field initialization
destPath := pathutil.ResolvePrdDirectoryPath(config, "")
field := dialog.FormField{
    ID:    "destination",
    Label: "Destination",
    Value: destPath,
    Help:  "Where the PRD file will be saved",
}
```

## Design Principles

### 1. Single Responsibility
- `ResolvePrdDirectoryPath` handles path resolution only
- `GetPrdDirectory` handles path resolution + directory creation
- Clear separation makes each function easy to test and reuse

### 2. No Coupling to Models
- Functions take only `config.Config` as parameter, not `Model`
- Decoupling allows use in any context (UI, CLI, tests)
- Makes testing simpler without mocking entire Model structures

### 3. Consistent Behavior
- All callers get same path resolution logic
- No duplication across modules
- Changes to logic only need to be made once

### 4. Backwards Compatibility
- Priority chain respects user preferences (lastUsedPath)
- Falls back to sensible defaults when preferences unavailable
- Works with projects that don't have `.taskmaster/docs` yet

### 5. User-Centric
- Remembers where users last worked (lastUsedPath)
- Suggests standard location (.taskmaster/docs) when available
- Gracefully handles cases where standard location doesn't exist yet

## Migration from Old Helpers

### From resolveDocsDir() in create_prd.go

Old:
```go
docsDir, err := m.resolveDocsDir()
```

New:
```go
docsDir, err := pathutil.GetPrdDirectory(m.config, "")
```

### From defaultPrdDirectory() in parse_prd.go

Old:
```go
startDir := m.defaultPrdDirectory()
```

New:
```go
startDir := pathutil.ResolvePrdDirectoryPath(m.config, m.lastPrdPath)
```

### From getDestinationPath() in prd_form.go

Old:
```go
destPath := getDestinationPath(cfg)
```

New:
```go
destPath := pathutil.ResolvePrdDirectoryPath(cfg, "")
```

## Testing

The package includes comprehensive unit tests covering:
- Priority ordering with `lastUsedPath`
- Nil and empty `TaskMasterPath` handling
- Default location fallback
- Existing directory detection
- Directory creation
- Error handling for permission issues
- Edge cases with relative/absolute paths

Run tests with:
```bash
go test ./internal/pathutil -v
```

## Future Extensions

The `pathutil` package is designed to be extensible for other path-related utilities:
- Document directory helpers
- Log file path resolution
- Configuration file path resolution
- Cache directory resolution

New functions should follow the same patterns:
1. Separation of resolution logic from creation logic
2. Clear priority/fallback chains
3. Comprehensive documentation and examples
4. Thorough test coverage

## Troubleshooting

### PRD Directory Not Created

**Symptom**: "failed to create PRD directory: permission denied"

**Solutions**:
- Check write permissions on TaskMasterPath parent directory
- Ensure `config.TaskMasterPath` points to a directory you own
- Try with default `.taskmaster/docs` path in a directory you control

### Wrong Directory Selected

**Symptom**: PRD files saved/loaded from unexpected location

**Solutions**:
- Clear `lastPrdPath` state if it's pointing to wrong location
- Check `config.TaskMasterPath` setting
- Verify `.taskmaster/docs` doesn't exist where not expected

### Path Inconsistencies

**Symptom**: Different modules using different paths for same project

**Solutions**:
- Ensure all modules import from `pathutil` package
- Don't duplicate path resolution logic
- Update all paths to use pathutil functions
