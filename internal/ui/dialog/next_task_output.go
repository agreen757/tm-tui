package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// maxOutputLines is the maximum number of lines to keep in the output buffer
const maxOutputLines = 1000

// NextTaskOutputContent displays task-master next output in a scrollable viewport
type NextTaskOutputContent struct {
	viewport viewport.Model
	output   []string
	loading  bool
	width    int
	height   int
}

// NewNextTaskOutputContent creates a new NextTaskOutputContent component
func NewNextTaskOutputContent() *NextTaskOutputContent {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.DefaultKeyMap()
	return &NextTaskOutputContent{
		viewport: vp,
		loading:  true,
		output:   []string{"Loading..."},
	}
}

// Init initializes the component
func (c *NextTaskOutputContent) Init() tea.Cmd {
	return nil
}

// SetSize updates the viewport dimensions
func (c *NextTaskOutputContent) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width
	c.viewport.Height = height
}

// Update processes messages and updates state
func (c *NextTaskOutputContent) Update(msg tea.Msg) (ModalContent, tea.Cmd) {
	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

// View renders the component with provided width and height
func (c *NextTaskOutputContent) View(width, height int) string {
	c.SetSize(width, height)
	return c.render()
}

// render is the internal render method
func (c *NextTaskOutputContent) render() string {
	if c.loading {
		return c.renderLoading()
	}

	if len(c.output) == 0 {
		return c.renderEmpty()
	}

	// Render viewport with output
	content := strings.Join(c.output, "\n")
	c.viewport.SetContent(content)
	return c.viewport.View()
}

// renderLoading renders the loading state
func (c *NextTaskOutputContent) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(c.width).
		Height(c.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return style.Render("Loading...")
}

// renderEmpty renders the empty state
func (c *NextTaskOutputContent) renderEmpty() string {
	style := lipgloss.NewStyle().
		Width(c.width).
		Height(c.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return style.Render("No output available")
}

// SetOutput updates the output lines and transitions from loading state
func (c *NextTaskOutputContent) SetOutput(lines []string) {
	c.output = lines
	c.loading = false

	// Handle empty output
	if len(lines) == 0 {
		c.output = []string{"No tasks available."}
	}

	c.viewport.SetContent(strings.Join(c.output, "\n"))
}

// SetLoading updates the loading state
func (c *NextTaskOutputContent) SetLoading(loading bool) {
	c.loading = loading
	if loading {
		c.output = []string{"Loading..."}
	}
}

// SetError updates output with formatted error message
func (c *NextTaskOutputContent) SetError(err error) {
	c.loading = false
	errorMsg := fmt.Sprintf("Error: %v\n\nCheck that task-master is properly installed and configured.", err)
	c.output = []string{errorMsg}
	c.viewport.SetContent(errorMsg)
}

// HandleKey processes key events and delegates to viewport if applicable
func (c *NextTaskOutputContent) HandleKey(msg tea.KeyMsg) tea.Cmd {
	// Let viewport handle scroll keys internally
	_, cmd := c.viewport.Update(msg)
	return cmd
}

// AppendOutput appends a line to the output buffer with batching and buffer management
func (c *NextTaskOutputContent) AppendOutput(line string) {
	c.output = append(c.output, line)

	// Trim if exceeding max lines
	if len(c.output) > maxOutputLines {
		c.output = c.output[len(c.output)-maxOutputLines:]
	}

	// Only update content every 10 lines to avoid excessive re-renders
	if len(c.output)%10 == 0 || strings.Contains(line, "Task:") {
		c.viewport.SetContent(strings.Join(c.output, "\n"))
	}
}

// FinalizeContent updates the viewport with final content and transitions from loading state
func (c *NextTaskOutputContent) FinalizeContent() {
	c.viewport.SetContent(strings.Join(c.output, "\n"))
	c.loading = false
}

// GetOutput returns a copy of the current output lines
func (c *NextTaskOutputContent) GetOutput() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(c.output))
	copy(result, c.output)
	return result
}
