package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adriangreen/tm-tui/internal/prd"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/adriangreen/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

type parsePrdMode = taskmaster.ParsePrdMode

const (
	parsePrdModeAppend  = taskmaster.ParsePrdModeAppend
	parsePrdModeReplace = taskmaster.ParsePrdModeReplace
)

type parsePrdOptions struct {
	Mode parsePrdMode
}

type parsePrdResultMsg struct {
	Summaries []prd.Summary
	TaskIDs   []string
	Mode      parsePrdMode
	Path      string
	Err       error
}

type parsePrdStreamClosedMsg struct{}

type prdResultListItem struct {
	title       string
	description string
	taskID      string
}

func (i *prdResultListItem) Title() string       { return i.title }
func (i *prdResultListItem) Description() string { return i.description }
func (i *prdResultListItem) FilterValue() string { return i.title }

func (m *Model) openParsePrdWorkflow() tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	startDir := m.defaultPrdDirectory()
	fileDialog := dialog.NewFileSelectionDialog("Select PRD File", startDir, 78, 20, []string{".md", ".txt"})
	if dm.Style != nil {
		dialog.ApplyStyleToDialog(fileDialog, dm.Style)
	}

	initCmd := fileDialog.Init()
	m.appState.AddDialog(fileDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			appErr := NewIOError("File Selection", "Failed to select file", err).
				WithRecoveryHints(
					"Check file permissions",
					"Try selecting a different file",
				)
			m.showAppError(appErr)
			return nil
		}
		path, _ := value.(string)
		if path == "" {
			return nil
		}
		m.lastPrdPath = filepath.Dir(path)
		if m.config != nil && m.config.StatePath != "" {
			if saveErr := m.SaveUIState(); saveErr != nil {
				m.addLogLine(fmt.Sprintf("Warning: failed to persist last PRD directory: %v", saveErr))
			}
		}
		return m.showParsePrdOptions(path)
	})

	return initCmd
}

func (m *Model) showParsePrdOptions(path string) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	fields := []dialog.FormField{
		{
			ID:    "mode",
			Label: "Import Mode",
			Type:  dialog.FormFieldTypeRadio,
			Options: []dialog.FormOption{
				{Value: string(parsePrdModeAppend), Label: "Append", Description: "Keep existing tasks and add new ones"},
				{Value: string(parsePrdModeReplace), Label: "Replace", Description: "Overwrite existing tasks with parsed output"},
			},
			Value: string(parsePrdModeAppend),
		},
	}

	title := fmt.Sprintf("File: %s", filepath.Base(path))
	optionsDialog := dialog.NewFormDialog(
		"Parse PRD Options",
		title,
		fields,
		[]string{"Continue", "Cancel"},
		dm.Style,
		func(form *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Continue" {
				return nil, nil
			}
			mode := parsePrdMode(stringValue(values, "mode"))
			if mode != parsePrdModeReplace {
				mode = parsePrdModeAppend
			}
			return parsePrdOptions{Mode: mode}, nil
		},
	)

	m.appState.AddDialog(optionsDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			appErr := NewOperationError("Parse PRD Options", "Failed to process options", err).
				WithRecoveryHints(
					"Try again",
				)
			m.showAppError(appErr)
			return nil
		}
		if value == nil {
			return nil
		}
		opts, ok := value.(parsePrdOptions)
		if !ok {
			return nil
		}
		return m.startParsePrdJob(path, opts.Mode)
	})

	return nil
}

func (m *Model) startParsePrdJob(path string, mode parsePrdMode) tea.Cmd {
	if m.config == nil || m.config.TaskMasterPath == "" {
		errMsg := "Task Master project not detected. Configure taskmasterPath in config or run inside a Task Master workspace."
		appErr := NewDependencyError("Parse PRD", errMsg, nil).
			WithRecoveryHints(
				"Check if you're running inside a Task Master workspace",
				"Set taskmasterPath in your config file",
				"Run 'task-master init' to set up a workspace",
			)
		m.showAppError(appErr)
		return nil
	}
	if m.taskService == nil {
		errMsg := "Task service unavailable. Restart the TUI inside a Task Master workspace."
		appErr := NewDependencyError("Parse PRD", errMsg, nil).
			WithRecoveryHints(
				"Restart the TUI application",
				"Ensure you're inside a Task Master workspace",
				"Check Task Master CLI installation",
			)
		m.showAppError(appErr)
		return nil
	}

	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	progress := dialog.NewProgressDialog("Parsing PRD", 70, 9)
	progress.SetCancellable(true)
	progress.SetLabel("Preparing…")
	if dm.Style != nil {
		dialog.ApplyStyleToDialog(progress, dm.Style)
	}
	m.appState.AddDialog(progress, nil)

	startID := 1
	if mode == parsePrdModeAppend {
		startID = m.nextRootTaskID()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.parsePrdCancel = cancel
	ch := make(chan tea.Msg)
	m.parsePrdChan = ch

	absPath := path
	if resolved, err := filepath.Abs(path); err == nil {
		absPath = resolved
	}
	go func() {
		defer close(ch)
		send := func(msg tea.Msg) bool {
			select {
			case <-ctx.Done():
				return false
			case ch <- msg:
				return true
			}
		}

		err := m.taskService.ParsePRDWithProgress(ctx, absPath, mode, func(state taskmaster.ParsePrdProgressState) {
			progress := state.Progress
			if progress < 0 {
				progress = 0
			}
			if progress > 1 {
				progress = 1
			}
			send(dialog.ProgressUpdateMsg{Progress: progress, Label: state.Label})
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			send(parsePrdResultMsg{Err: err, Path: absPath})
			return
		}

		data, readErr := os.ReadFile(absPath)
		if readErr != nil {
			send(parsePrdResultMsg{Err: readErr, Path: absPath})
			return
		}
		nodes, parseErr := prd.Parse(string(data))
		if parseErr != nil {
			send(parsePrdResultMsg{Err: parseErr, Path: absPath})
			return
		}

		ids := make([]string, len(nodes))
		for i := range nodes {
			ids[i] = strconv.Itoa(startID + i)
		}

		summaries := prd.Summaries(nodes)
		send(dialog.ProgressUpdateMsg{Progress: 1.0, Label: "Completed"})
		send(parsePrdResultMsg{Summaries: summaries, TaskIDs: ids, Mode: mode, Path: absPath})
	}()

	return m.waitForParsePrdMessages()
}

func (m *Model) waitForParsePrdMessages() tea.Cmd {
	ch := m.parsePrdChan
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		if msg, ok := <-ch; ok {
			return msg
		}
		return parsePrdStreamClosedMsg{}
	}
}

func (m *Model) handleParsePrdResult(msg parsePrdResultMsg) tea.Cmd {
	m.clearParsePrdRuntimeState()
	m.dismissActiveProgressDialog()

	if msg.Err != nil {
		appErr := NewParsingError("Parse PRD", fmt.Sprintf("Failed to parse %s", filepath.Base(msg.Path)), msg.Err).
			WithDetails("The file may be malformed or unsupported.").
			WithRecoveryHints(
				"Check the file format and encoding",
				"Verify the PRD document structure",
				"Try a different file or contact support",
			)
		m.showAppError(appErr)
		return nil
	}

	if len(msg.Summaries) == 0 {
		m.addLogLine("PRD parsing completed but no tasks were generated.")
		return LoadTasksCmd(m.taskService)
	}

	m.addLogLine(fmt.Sprintf("Generated %d tasks from %s (%s)", len(msg.Summaries), filepath.Base(msg.Path), msg.Mode))

	var cmds []tea.Cmd
	cmds = append(cmds, LoadTasksCmd(m.taskService))
	if dialogCmd := m.showParsePrdResults(msg); dialogCmd != nil {
		cmds = append(cmds, dialogCmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) showParsePrdResults(msg parsePrdResultMsg) tea.Cmd {
	dm := m.dialogManager()
	if dm == nil {
		return nil
	}

	items := make([]dialog.ListItem, 0, len(msg.Summaries))
	for i, summary := range msg.Summaries {
		descParts := []string{}
		if summary.Description != "" {
			descParts = append(descParts, summary.Description)
		}
		if summary.SubtaskCount > 0 {
			descParts = append(descParts, fmt.Sprintf("%d subtasks", summary.SubtaskCount))
		}
		description := strings.Join(descParts, " · ")
		items = append(items, &prdResultListItem{title: summary.Title, description: description, taskID: safeSliceValue(msg.TaskIDs, i)})
	}

	resultsDialog := dialog.NewListDialog("PRD Parse Results", 80, 20, items)
	resultsDialog.SetShowDescription(true)
	if dm.Style != nil {
		dialog.ApplyStyleToDialog(resultsDialog, dm.Style)
	}
	m.appState.AddDialog(resultsDialog, nil)
	return nil
}

func safeSliceValue(values []string, index int) string {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return ""
}

func (m *Model) defaultPrdDirectory() string {
	if m.lastPrdPath != "" {
		return m.lastPrdPath
	}
	if m.config != nil && m.config.TaskMasterPath != "" {
		docs := filepath.Join(m.config.TaskMasterPath, ".taskmaster", "docs")
		if info, err := os.Stat(docs); err == nil && info.IsDir() {
			return docs
		}
		return m.config.TaskMasterPath
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func (m *Model) clearParsePrdRuntimeState() {
	m.parsePrdChan = nil
	m.parsePrdCancel = nil
}

func (m *Model) cancelParsePrdJob() {
	if m.parsePrdCancel != nil {
		m.parsePrdCancel()
		m.parsePrdCancel = nil
	}
}

func (m *Model) dismissActiveProgressDialog() {
	if m.appState == nil || !m.appState.HasActiveDialog() {
		return
	}
	if _, ok := m.appState.ActiveDialog().(*dialog.ProgressDialog); ok {
		m.appState.PopDialog()
	}
}

func (m *Model) nextRootTaskID() int {
	maxID := 0
	for _, task := range m.tasks {
		if n, err := strconv.Atoi(task.ID); err == nil && n > maxID {
			maxID = n
		}
	}
	return maxID + 1
}
