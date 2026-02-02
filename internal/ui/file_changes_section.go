package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileChangesSection represents the UI component for file changes in task details
type FileChangesSection struct {
	task        *taskmaster.Task
	selectedIdx int
	focused     bool
	styles      *Styles
}

// NewFileChangesSection creates a new file changes section
func NewFileChangesSection(task *taskmaster.Task, styles *Styles) *FileChangesSection {
	return &FileChangesSection{
		task:        task,
		selectedIdx: 0,
		focused:     false,
		styles:      styles,
	}
}

// Render draws the file changes section
func (s *FileChangesSection) Render(wrapWidth int) string {
	if s.task == nil || len(s.task.FileChanges) == 0 {
		return ""
	}

	var b strings.Builder

	// Header with count of changes
	count := len(s.task.FileChanges)
	headerText := fmt.Sprintf("File Changes (%d):", count)
	b.WriteString(s.styles.Subtitle.Render(headerText))
	b.WriteString("\n\n")

	// Render list of files with change indicators
	for i, fileChange := range s.task.FileChanges {
		// Visual indicator based on change type
		indicator := s.getChangeIndicator(fileChange.ChangeType)

		// Highlight selected file when focused
		isSelected := s.focused && i == s.selectedIdx

		// Build file entry
		fileLine := s.renderFileEntry(fileChange, indicator, isSelected, wrapWidth)
		b.WriteString(fileLine)
		b.WriteString("\n")
	}

	return b.String()
}

// getChangeIndicator returns a visual indicator for the file change type
func (s *FileChangesSection) getChangeIndicator(changeType string) string {
	switch changeType {
	case "added":
		return s.styles.Success.Render("+ ")
	case "modified":
		return s.styles.Warning.Render("~ ")
	case "deleted":
		return s.styles.Error.Render("- ")
	default:
		return s.styles.Subtle.Render("? ")
	}
}

// renderFileEntry renders a single file change entry
func (s *FileChangesSection) renderFileEntry(fc taskmaster.FileChange, indicator string, isSelected bool, wrapWidth int) string {
	var b strings.Builder

	// File path (show relative path, make it shorter if needed)
	displayPath := fc.Path
	if len(displayPath) > 60 {
		// Shorten long paths by showing filename and parent dir
		displayPath = "..." + filepath.Base(filepath.Dir(fc.Path)) + "/" + filepath.Base(fc.Path)
	}

	// Build the line
	b.WriteString("  ")
	b.WriteString(indicator)

	// Apply selection styling if focused
	pathStyle := lipgloss.NewStyle()
	if isSelected {
		pathStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true)
	}

	b.WriteString(pathStyle.Render(displayPath))

	// Show pending/committed status
	if fc.IsPending {
		b.WriteString(s.styles.Subtle.Render(" (uncommitted)"))
	} else if fc.CommitID != "" {
		commitShort := fc.CommitID
		if len(commitShort) > 7 {
			commitShort = commitShort[:7]
		}
		b.WriteString(s.styles.Subtle.Render(fmt.Sprintf(" [%s]", commitShort)))
	}

	// Description on next line if present
	if fc.Description != "" {
		b.WriteString("\n    ")
		// Wrap description text
		wrappedDesc := wrapText(fc.Description, wrapWidth-4)
		// Indent wrapped lines
		if strings.Contains(wrappedDesc, "\n") {
			lines := strings.Split(wrappedDesc, "\n")
			b.WriteString(s.styles.Subtle.Render(lines[0]))
			for _, line := range lines[1:] {
				b.WriteString("\n    ")
				b.WriteString(s.styles.Subtle.Render(line))
			}
		} else {
			b.WriteString(s.styles.Subtle.Render(wrappedDesc))
		}
	}

	return b.String()
}

// SetFocused sets the focus state of the section
func (s *FileChangesSection) SetFocused(focused bool) {
	s.focused = focused
}

// IsFocused returns whether the section is currently focused
func (s *FileChangesSection) IsFocused() bool {
	return s.focused
}

// SetSelected sets the selected index
func (s *FileChangesSection) SetSelected(idx int) {
	if idx >= 0 && idx < len(s.task.FileChanges) {
		s.selectedIdx = idx
	}
}

// GetSelected returns the currently selected index
func (s *FileChangesSection) GetSelected() int {
	return s.selectedIdx
}

// GetFileCount returns the number of file changes
func (s *FileChangesSection) GetFileCount() int {
	if s.task == nil {
		return 0
	}
	return len(s.task.FileChanges)
}

// HandleKeyEvent processes keyboard input for the section
func (s *FileChangesSection) HandleKeyEvent(msg tea.KeyMsg) tea.Cmd {
	if !s.focused || len(s.task.FileChanges) == 0 {
		return nil
	}

	switch msg.String() {
	case "up", "k":
		if s.selectedIdx > 0 {
			s.selectedIdx--
		}
		return nil

	case "down", "j":
		if s.selectedIdx < len(s.task.FileChanges)-1 {
			s.selectedIdx++
		}
		return nil

	case "enter":
		// Open file in system editor
		return s.OpenFile()

	case "alt+d":
		// View diff for selected file
		return s.ViewDiff()
	}

	return nil
}

// OpenFile opens the selected file in system editor
func (s *FileChangesSection) OpenFile() tea.Cmd {
	if s.selectedIdx < 0 || s.selectedIdx >= len(s.task.FileChanges) {
		return nil
	}

	filePath := s.task.FileChanges[s.selectedIdx].Path

	// Get absolute path
	absPath := filePath
	if !filepath.IsAbs(filePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return func() tea.Msg {
				return fileOperationMsg{
					success: false,
					message: fmt.Sprintf("Failed to get working directory: %v", err),
				}
			}
		}
		absPath = filepath.Join(cwd, filePath)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return func() tea.Msg {
			return fileOperationMsg{
				success: false,
				message: fmt.Sprintf("File not found: %s", filePath),
			}
		}
	}

	// Open with system editor (using $EDITOR or fallback to vim/vi/nano)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "vi", "nano"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
		if editor == "" {
			return func() tea.Msg {
				return fileOperationMsg{
					success: false,
					message: "No editor found. Set $EDITOR environment variable.",
				}
			}
		}
	}

	return tea.ExecProcess(exec.Command(editor, absPath), func(err error) tea.Msg {
		if err != nil {
			return fileOperationMsg{
				success: false,
				message: fmt.Sprintf("Failed to open file: %v", err),
			}
		}
		return fileOperationMsg{
			success: true,
			message: fmt.Sprintf("Opened %s", filePath),
		}
	})
}

// ViewDiff shows diff for the selected file
func (s *FileChangesSection) ViewDiff() tea.Cmd {
	if s.selectedIdx < 0 || s.selectedIdx >= len(s.task.FileChanges) {
		return nil
	}

	fileChange := s.task.FileChanges[s.selectedIdx]

	// Build git diff command based on file status
	var cmd *exec.Cmd
	if fileChange.IsPending {
		// Uncommitted changes - show diff against HEAD
		cmd = exec.Command("git", "diff", "HEAD", "--", fileChange.Path)
	} else if fileChange.CommitID != "" {
		// Show diff for specific commit
		cmd = exec.Command("git", "show", fileChange.CommitID, "--", fileChange.Path)
	} else {
		// Fallback - show current diff
		cmd = exec.Command("git", "diff", fileChange.Path)
	}

	// Execute command and capture output
	return func() tea.Msg {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return diffViewMsg{
				success: false,
				message: fmt.Sprintf("Failed to get diff: %v", err),
				content: "",
			}
		}

		return diffViewMsg{
			success: true,
			message: fmt.Sprintf("Diff for %s", fileChange.Path),
			content: string(output),
		}
	}
}

// fileOperationMsg represents the result of a file operation
type fileOperationMsg struct {
	success bool
	message string
}

// diffViewMsg represents the result of a diff view operation
type diffViewMsg struct {
	success bool
	message string
	content string
}
