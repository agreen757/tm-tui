# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **BREAKING**: Task expansion now uses Task Master CLI commands instead of local functions
- Task expansion workflow redesigned with scope selection dialog
- Expansion progress is now shown in real-time from CLI output
- Improved error handling for expansion failures

### Added
- Support for expanding all tasks at once (`task-master expand --all`)
- Support for expanding task ranges (`--from` and `--to` flags)
- Support for tag-based expansion
- Real-time progress updates during expansion
- Cancellation support during expansion (Ctrl+C or ESC)
- `ExpansionScopeDialog` for configuring expansion options
- `ExecuteExpandWithProgress()` service method for CLI integration

### Removed
- Local task expansion preview dialog (CLI handles expansion directly)
- Local task expansion edit dialog (CLI handles expansion directly)
- Direct use of `ExpandTaskDrafts()` and `ApplySubtaskDrafts()` in UI layer

### Deprecated
- `ExpandTaskDrafts()` - Use CLI execution instead (kept for testing)
- `ApplySubtaskDrafts()` - Use CLI execution instead (kept for testing)
- Legacy expansion functions in command_handlers.go (kept for backward compatibility)

### Fixed
- Task expansion now properly persists changes via CLI
- Expansion with `--research` flag now works correctly
- Tasks are reliably reloaded after expansion
- Progress reporting is more accurate and informative

## [Previous versions]

See git history for older changes.
