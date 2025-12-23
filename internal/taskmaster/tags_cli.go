package taskmaster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TagContext describes a Task Master tag context returned from the CLI.
type TagContext struct {
	Name           string
	TaskCount      int
	CompletedCount int
	CreatedLabel   string
	Description    string
	Active         bool
}

// TagList aggregates the tag contexts returned from the CLI.
type TagList struct {
	Tags             []TagContext
	RawOutput        string
	GeneratedAt      time.Time
	MetadataIncluded bool
}

// TagOperationResult captures the stdout/stderr emitted by a tag command.
type TagOperationResult struct {
	Command     []string
	Output      string
	CompletedAt time.Time
}

// TagAddOptions configures the add-tag CLI invocation.
type TagAddOptions struct {
	Name            string
	CopyFromCurrent bool
	CopyFrom        string
	Description     string
}

// TagCopyOptions configures optional properties for copy-tag.
type TagCopyOptions struct {
	Description string
}

// ListTagContexts executes `task-master tags` and parses the resulting table.
func (s *Service) ListTagContexts(ctx context.Context, includeMetadata bool) (*TagList, error) {
	if !s.available {
		return nil, fmt.Errorf("taskmaster not available")
	}

	args := []string{"tags"}
	/* if includeMetadata {
		args = append(args, "--show-metadata")
	} */

	output, err := s.runSimpleTaskMasterCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	tagNames := s.availableTagNames()
	tags, err := parseTagTable(output, includeMetadata, tagNames)
	if err != nil {
		return nil, err
	}

	return &TagList{
		Tags:             tags,
		RawOutput:        output,
		GeneratedAt:      time.Now(),
		MetadataIncluded: includeMetadata,
	}, nil
}

// AddTagContext creates a new tag context via the CLI.
func (s *Service) AddTagContext(ctx context.Context, opts TagAddOptions) (*TagOperationResult, error) {
	if strings.TrimSpace(opts.Name) == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	args := []string{"tags", "add", strings.TrimSpace(opts.Name)}
	if opts.CopyFromCurrent {
		args = append(args, "--copy-from-current")
	}
	if opts.CopyFrom != "" {
		args = append(args, fmt.Sprintf("--copy-from=%s", strings.TrimSpace(opts.CopyFrom)))
	}
	if opts.Description != "" {
		args = append(args, fmt.Sprintf("-d=%s", opts.Description))
	}

	output, err := s.runSimpleTaskMasterCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	return &TagOperationResult{
		Command:     args,
		Output:      output,
		CompletedAt: time.Now(),
	}, nil
}

// DeleteTagContext removes a tag context and its tasks via CLI.
func (s *Service) DeleteTagContext(ctx context.Context, name string, skipConfirmation bool) (*TagOperationResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	args := []string{"delete-tag", strings.TrimSpace(name)}
	if skipConfirmation {
		args = append(args, "--yes")
	}

	output, err := s.runSimpleTaskMasterCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	return &TagOperationResult{
		Command:     args,
		Output:      output,
		CompletedAt: time.Now(),
	}, nil
}

// UseTagContext switches the active tag via CLI and updates in-memory config.
func (s *Service) UseTagContext(ctx context.Context, name string) (*TagOperationResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	args := []string{"use-tag", strings.TrimSpace(name)}
	output, err := s.runSimpleTaskMasterCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	if s.config != nil {
		s.config.ActiveTag = strings.TrimSpace(name)
	}

	return &TagOperationResult{
		Command:     args,
		Output:      output,
		CompletedAt: time.Now(),
	}, nil
}

// RenameTagContext renames an existing tag context using the CLI.
func (s *Service) RenameTagContext(ctx context.Context, oldName, newName string) (*TagOperationResult, error) {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == "" || newName == "" {
		return nil, fmt.Errorf("both old and new tag names are required")
	}

	args := []string{"rename-tag", oldName, newName}
	output, err := s.runSimpleTaskMasterCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	if s.config != nil && strings.EqualFold(s.config.ActiveTag, oldName) {
		s.config.ActiveTag = newName
	}

	return &TagOperationResult{
		Command:     args,
		Output:      output,
		CompletedAt: time.Now(),
	}, nil
}

// CopyTagContext copies an existing tag context via the CLI.
func (s *Service) CopyTagContext(ctx context.Context, sourceName, targetName string, opts TagCopyOptions) (*TagOperationResult, error) {
	sourceName = strings.TrimSpace(sourceName)
	targetName = strings.TrimSpace(targetName)
	if sourceName == "" || targetName == "" {
		return nil, fmt.Errorf("both source and target tag names are required")
	}

	args := []string{"copy-tag", sourceName, targetName}
	if opts.Description != "" {
		args = append(args, fmt.Sprintf("-d=%s", opts.Description))
	}

	output, err := s.runSimpleTaskMasterCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	return &TagOperationResult{
		Command:     args,
		Output:      output,
		CompletedAt: time.Now(),
	}, nil
}

// runSimpleTaskMasterCommand executes a CLI command and returns combined output.
func (s *Service) runSimpleTaskMasterCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "task-master", args...)
	if s.RootDir != "" {
		cmd.Dir = s.RootDir
	}

	// Ensure output is plain text to simplify parsing.
	env := append([]string{}, os.Environ()...)
	env = append(env, "NO_COLOR=1", "FORCE_COLOR=0", "CLICOLOR=0", "TERM=xterm-256color")
	cmd.Env = env

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("task-master %s failed: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(buf.String()))
	}

	return buf.String(), nil
}

// parseTagTable extracts tag contexts from the CLI table output.
func parseTagTable(output string, includeMetadata bool, candidates []string) ([]TagContext, error) {
	lines := strings.Split(output, "\n")
	var contexts []TagContext

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "│") {
			continue
		}

		cells := splitTableRow(line)
		if len(cells) == 0 {
			continue
		}
		if strings.Contains(strings.ToLower(cells[0]), "tag name") {
			continue
		}

		expectedCols := 3
		if includeMetadata {
			expectedCols = 5
		}
		if len(cells) < expectedCols {
			// Skip divider rows or malformed ones
			continue
		}

		context := TagContext{}
		nameCell := cells[0]
		if strings.HasPrefix(nameCell, "●") {
			context.Active = true
			nameCell = strings.TrimSpace(strings.TrimPrefix(nameCell, "●"))
		} else if strings.HasPrefix(nameCell, "○") {
			nameCell = strings.TrimSpace(strings.TrimPrefix(nameCell, "○"))
		}

		context.Name = resolveTagName(nameCell, candidates)
		context.TaskCount = parseTableInt(cells[1])
		context.CompletedCount = parseTableInt(cells[2])

		if includeMetadata && len(cells) >= 5 {
			context.CreatedLabel = cells[3]
			context.Description = cells[4]
		}

		if context.Name != "" {
			contexts = append(contexts, context)
		}
	}

	return contexts, nil
}

func splitTableRow(row string) []string {
	trimmed := strings.Trim(row, "│")
	parts := strings.Split(trimmed, "│")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cell := strings.TrimSpace(part)
		if cell != "" {
			cells = append(cells, cell)
		} else {
			cells = append(cells, "")
		}
	}
	return cells
}

func parseTableInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func (s *Service) availableTagNames() []string {
	if s.RootDir == "" {
		return nil
	}
	path := filepath.Join(s.RootDir, ".taskmaster", "tasks", "tasks.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	names := make([]string, 0, len(raw))
	for key := range raw {
		if key == "" {
			continue
		}
		switch key {
		case "tasks", "meta", "config", "version":
			continue
		}
		names = append(names, key)
	}
	sort.Strings(names)
	return names
}

func resolveTagName(partial string, candidates []string) string {
	clean := strings.TrimSpace(partial)
	if !strings.ContainsRune(clean, '…') || len(candidates) == 0 {
		return clean
	}
	prefix := strings.Split(clean, "…")[0]
	prefix = strings.TrimSpace(prefix)
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			return candidate
		}
	}
	return clean
}
