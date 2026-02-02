package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/agreen757/tm-tui/internal/filechanges"
	"github.com/agreen757/tm-tui/internal/taskmaster"
)

// FilePreview is a Bubble Tea component for displaying file content and diffs
type FilePreview struct {
	file          taskmaster.FileChange
	content       string
	viewport      viewport.Model
	loader        filechanges.ContentLoaderInterface
	diffGen       filechanges.DiffGeneratorInterface
	diffMode      bool
	width         int
	height        int
	syntaxEnabled bool
	lastError     string
}

// NewFilePreview creates a new file preview component
func NewFilePreview(
	loader filechanges.ContentLoaderInterface,
	diffGen filechanges.DiffGeneratorInterface,
) *FilePreview {
	return &FilePreview{
		loader:        loader,
		diffGen:       diffGen,
		diffMode:      false,
		syntaxEnabled: true,
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
	}
}

// SetFile sets the file to preview and loads its content
func (p *FilePreview) SetFile(ctx context.Context, file taskmaster.FileChange) error {
	p.file = file
	return p.Refresh(ctx)
}

// ToggleDiffMode switches between raw content and diff view
func (p *FilePreview) ToggleDiffMode(ctx context.Context) error {
	p.diffMode = !p.diffMode
	return p.Refresh(ctx)
}

// SetDimensions updates the viewport dimensions
func (p *FilePreview) SetDimensions(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width
	p.viewport.Height = height
}

// Refresh reloads the content and renders it
func (p *FilePreview) Refresh(ctx context.Context) error {
	p.lastError = ""

	if p.file.Path == "" {
		p.content = ""
		p.viewport.SetContent("")
		return nil
	}

	var content string
	var err error

	if p.diffMode {
		// Load diff content
		content, err = p.loadDiffContent(ctx)
		if err != nil {
			p.lastError = fmt.Sprintf("Failed to load diff: %v", err)
			p.content = p.lastError
			p.viewport.SetContent(p.lastError)
			return err
		}
	} else {
		// Load raw file content
		content, err = p.loader.LoadContent(ctx, p.file)
		if err != nil {
			p.lastError = fmt.Sprintf("Failed to load file: %v", err)
			p.content = p.lastError
			p.viewport.SetContent(p.lastError)
			return err
		}
	}

	p.content = content

	// Apply formatting and syntax highlighting
	formatted := p.formatContent()
	p.viewport.SetContent(formatted)

	return nil
}

// loadDiffContent loads the diff for the current file
func (p *FilePreview) loadDiffContent(ctx context.Context) (string, error) {
	if p.file.IsPending {
		// For pending changes, generate diff against HEAD
		return p.diffGen.GeneratePendingDiff(ctx, p.file.Path)
	} else if p.file.CommitID != "" {
		// For committed changes, we need a parent commit
		// For now, use HEAD as comparison
		return p.diffGen.GenerateDiff(ctx, p.file.Path, "HEAD", p.file.CommitID)
	}

	return "", fmt.Errorf("cannot generate diff for file %s: no pending changes or commit ID", p.file.Path)
}

// formatContent applies syntax highlighting and formatting
func (p *FilePreview) formatContent() string {
	if !p.syntaxEnabled {
		return p.content
	}

	lines := strings.Split(p.content, "\n")
	formatted := make([]string, len(lines))

	for i, line := range lines {
		formatted[i] = p.formatLine(line, i)
	}

	return strings.Join(formatted, "\n")
}

// formatLine applies formatting to a single line
func (p *FilePreview) formatLine(line string, lineNum int) string {
	if p.diffMode {
		return p.formatDiffLine(line)
	}

	return p.formatContentLine(line)
}

// formatDiffLine applies special formatting for diff lines
func (p *FilePreview) formatDiffLine(line string) string {
	if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
		// Added line - green background
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Render(line)
	}

	if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
		// Removed line - red background
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render(line)
	}

	if strings.HasPrefix(line, "@@") {
		// Hunk header - blue
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")).
			Render(line)
	}

	if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
		// File header - cyan
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Render(line)
	}

	return line
}

// formatContentLine applies formatting to content lines based on file type
func (p *FilePreview) formatContentLine(line string) string {
	// Simple syntax highlighting for common patterns
	fileExt := p.getFileExtension()

	switch fileExt {
	case "go":
		return p.highlightGoCode(line)
	case "json":
		return p.highlightJSON(line)
	case "yaml", "yml":
		return p.highlightYAML(line)
	default:
		return line
	}
}

// getFileExtension returns the file extension
func (p *FilePreview) getFileExtension() string {
	parts := strings.Split(p.file.Path, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// highlightGoCode applies basic Go syntax highlighting
func (p *FilePreview) highlightGoCode(line string) string {
	// Keywords to highlight
	keywords := []string{"func", "package", "import", "const", "var", "type", "if", "for", "return", "interface"}

	for _, keyword := range keywords {
		if strings.Contains(line, keyword) {
			line = strings.ReplaceAll(line, keyword,
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("5")).
					Render(keyword))
		}
	}

	return line
}

// highlightJSON applies basic JSON syntax highlighting
func (p *FilePreview) highlightJSON(line string) string {
	// Highlight keys in quotes
	if strings.Contains(line, ":") {
		parts := strings.Split(line, ":")
		if len(parts) > 0 && strings.Contains(parts[0], "\"") {
			parts[0] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Render(parts[0])
			return strings.Join(parts, ":")
		}
	}

	return line
}

// highlightYAML applies basic YAML syntax highlighting
func (p *FilePreview) highlightYAML(line string) string {
	// Highlight keys (before :)
	if strings.Contains(line, ":") {
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			key := parts[0]
			if !strings.HasPrefix(strings.TrimSpace(key), "#") {
				parts[0] = lipgloss.NewStyle().
					Foreground(lipgloss.Color("3")).
					Render(key)
			}
			return strings.Join(parts, ":")
		}
	}

	return line
}

// ToggleSyntaxHighlighting enables or disables syntax highlighting
func (p *FilePreview) ToggleSyntaxHighlighting() {
	p.syntaxEnabled = !p.syntaxEnabled
}

// GetContent returns the current content
func (p *FilePreview) GetContent() string {
	return p.content
}

// GetViewport returns the viewport model
func (p *FilePreview) GetViewport() *viewport.Model {
	return &p.viewport
}

// GetDiffMode returns whether diff mode is enabled
func (p *FilePreview) GetDiffMode() bool {
	return p.diffMode
}

// GetFile returns the current file
func (p *FilePreview) GetFile() taskmaster.FileChange {
	return p.file
}

// GetLastError returns the last error message
func (p *FilePreview) GetLastError() string {
	return p.lastError
}

// View renders the file preview
func (p *FilePreview) View() string {
	if p.file.Path == "" {
		return "No file selected"
	}

	header := p.renderHeader()
	content := p.viewport.View()

	return fmt.Sprintf("%s\n%s", header, content)
}

// renderHeader renders the header with file info
func (p *FilePreview) renderHeader() string {
	modeIndicator := "View"
	if p.diffMode {
		modeIndicator = "Diff"
	}

	status := "Committed"
	if p.file.IsPending {
		status = "Pending"
	}

	header := fmt.Sprintf("📄 %s [%s] (%s)", p.file.Path, modeIndicator, status)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Render(header)
}

// Update handles Bubble Tea messages
func (p *FilePreview) Update(msg tea.Msg) (*FilePreview, tea.Cmd) {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

// Model interface compliance
func (p *FilePreview) Init() tea.Cmd {
	return nil
}
