package dialog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SortOrder represents sorting options for the complexity report
type SortOrder int

const (
	SortByTaskID SortOrder = iota
	SortByScoreAsc
	SortByScoreDesc
	SortByLevelAsc
	SortByLevelDesc
)

// FilterSettings contains filter options for the complexity report
type FilterSettings struct {
	Levels map[taskmaster.ComplexityLevel]bool
	Tag    string
}

// NewFilterSettings creates default filter settings with all levels enabled
func NewFilterSettings() FilterSettings {
	return FilterSettings{
		Levels: map[taskmaster.ComplexityLevel]bool{
			taskmaster.ComplexityLow:      true,
			taskmaster.ComplexityMedium:   true,
			taskmaster.ComplexityHigh:     true,
			taskmaster.ComplexityVeryHigh: true,
		},
		Tag: "",
	}
}

// ComplexityReportDialog displays the analyzed task complexities in a tabular format
type ComplexityReportDialog struct {
	BaseDialog
	Report         *taskmaster.ComplexityReport
	Viewport       viewport.Model
	KeyMap         ComplexityReportKeyMap
	SelectedIndex  int
	SortOrder      SortOrder
	FilterSettings FilterSettings
	FilteredTasks  []taskmaster.TaskComplexity
	Help           help.Model
	ShowHelp       bool
	ShowLegend     bool
	width          int
	height         int
}

// Init satisfies Dialog interface
func (d *ComplexityReportDialog) Init() tea.Cmd {
	return nil
}

// ComplexityReportKeyMap defines keybindings for the complexity report dialog
type ComplexityReportKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	SortMode key.Binding
	Filter   key.Binding
	Export   key.Binding
	Help     key.Binding
	Close    key.Binding
}

func (k ComplexityReportKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.SortMode, k.Filter, k.Export, k.Help, k.Close}
}

func (k ComplexityReportKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.SortMode, k.Filter, k.Export},
		{k.Help, k.Close},
	}
}

// DefaultComplexityReportKeyMap returns default keybindings
func DefaultComplexityReportKeyMap() ComplexityReportKeyMap {
	return ComplexityReportKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select task"),
		),
		SortMode: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "change sort order"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter results"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export results"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		),
	}
}

// NewComplexityReportDialog creates a new dialog for displaying complexity analysis results
func NewComplexityReportDialog(
	report *taskmaster.ComplexityReport,
	style *DialogStyle,
) *ComplexityReportDialog {
	// Default viewport
	vp := viewport.New(80, 20)
	vp.KeyMap = viewport.KeyMap{} // Disable default viewport keybindings

	// Help model
	helpModel := help.New()

	// Create the dialog
	defaultWidth := 80
	defaultHeight := 22
	dialog := &ComplexityReportDialog{
		BaseDialog:     NewBaseDialog("Task Complexity Analysis Results", defaultWidth, defaultHeight, DialogKindCustom),
		Report:         report,
		Viewport:       vp,
		KeyMap:         DefaultComplexityReportKeyMap(),
		SelectedIndex:  0,
		SortOrder:      SortByTaskID,
		FilterSettings: NewFilterSettings(),
		ShowHelp:       false,
		Help:           helpModel,
		ShowLegend:     true,
		width:          defaultWidth,
		height:         defaultHeight,
	}
	dialog.Overlay = true
	dialog.SetZIndex(100)
	if style != nil {
		dialog.Style = style
	}
	dialog.SetRect(defaultWidth, defaultHeight, 0, 0)

	// Apply initial sorting and filtering
	dialog.applyFiltersAndSort()

	return dialog
}

// GetSelectedTask returns the currently selected task complexity info
func (d *ComplexityReportDialog) GetSelectedTask() *taskmaster.TaskComplexity {
	if d.SelectedIndex < 0 || d.SelectedIndex >= len(d.FilteredTasks) {
		return nil
	}
	return &d.FilteredTasks[d.SelectedIndex]
}

// applyFiltersAndSort applies the current filters and sorting to the report data
func (d *ComplexityReportDialog) applyFiltersAndSort() {
	if d.Report == nil || len(d.Report.Tasks) == 0 {
		d.FilteredTasks = []taskmaster.TaskComplexity{}
		return
	}

	// Apply filters
	filtered := make([]taskmaster.TaskComplexity, 0)
	for _, task := range d.Report.Tasks {
		// Filter by complexity level
		if !d.FilterSettings.Levels[task.Level] {
			continue
		}

		// Filter by tag (if specified)
		if d.FilterSettings.Tag != "" {
			// Here we'd need to have tag information in TaskComplexity
			// This is a placeholder for tag filtering logic
			// We might need to modify our data model or pass additional context
		}

		filtered = append(filtered, task)
	}

	// Sort the results
	switch d.SortOrder {
	case SortByTaskID:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].TaskID < filtered[j].TaskID
		})
	case SortByScoreAsc:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Score < filtered[j].Score
		})
	case SortByScoreDesc:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Score > filtered[j].Score
		})
	case SortByLevelAsc:
		sort.Slice(filtered, func(i, j int) bool {
			return string(filtered[i].Level) < string(filtered[j].Level)
		})
	case SortByLevelDesc:
		sort.Slice(filtered, func(i, j int) bool {
			return string(filtered[i].Level) > string(filtered[j].Level)
		})
	}

	d.FilteredTasks = filtered

	// Ensure selected index is valid
	if len(filtered) > 0 {
		if d.SelectedIndex >= len(filtered) {
			d.SelectedIndex = len(filtered) - 1
		} else if d.SelectedIndex < 0 {
			d.SelectedIndex = 0
		}
	} else {
		d.SelectedIndex = -1
	}
}

// ApplyFiltersAndSortForTest exposes filter recalculation for unit tests.
func (d *ComplexityReportDialog) ApplyFiltersAndSortForTest() {
	d.applyFiltersAndSort()
}

// cycleSortOrder changes to the next sort order in sequence
func (d *ComplexityReportDialog) cycleSortOrder() {
	switch d.SortOrder {
	case SortByTaskID:
		d.SortOrder = SortByScoreAsc
	case SortByScoreAsc:
		d.SortOrder = SortByScoreDesc
	case SortByScoreDesc:
		d.SortOrder = SortByLevelAsc
	case SortByLevelAsc:
		d.SortOrder = SortByLevelDesc
	case SortByLevelDesc:
		d.SortOrder = SortByTaskID
	}
	d.applyFiltersAndSort()
}

// Height implements Dialog.Height
func (d *ComplexityReportDialog) GetHeight() int {
	width, height, _, _ := d.GetRect()
	_ = width
	helpHeight := 0
	if d.ShowHelp {
		helpHeight = 5
	}
	legendHeight := 0
	if d.ShowLegend {
		legendHeight = 2
	}
	return height - 4 - helpHeight - legendHeight
}

// Width implements Dialog.Width
func (d *ComplexityReportDialog) GetWidth() int {
	width, _, _, _ := d.GetRect()
	return width - 4
}

// SetSize sets the dialog size

func (d *ComplexityReportDialog) SetSize(width, height int) {
	_, _, x, y := d.GetRect()
	d.width = width
	d.height = height
	d.SetRect(width, height, x, y)
	viewportHeight := d.GetHeight()
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	d.Viewport.Height = viewportHeight
	d.Viewport.Width = d.GetWidth()
}

// Update handles input and events
func (d *ComplexityReportDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		result, cmd := d.HandleKey(msg)
		if cmd != nil {
			return d, cmd
		}
		if result != DialogResultNone {
			return d, nil
		}
	case tea.WindowSizeMsg:
		// Adjust viewport when window resizes
		d.SetSize(msg.Width, msg.Height)
	case DialogSetFilterMsg:
		// Update filter settings
		d.FilterSettings = msg.Settings.(FilterSettings)
		d.applyFiltersAndSort()
	}

	// Update viewport
	var cmd tea.Cmd
	d.Viewport, cmd = d.Viewport.Update(msg)
	return d, cmd
}

// HandleKey processes key events
func (d *ComplexityReportDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch {
	case key.Matches(msg, d.KeyMap.Up):
		if d.SelectedIndex > 0 {
			d.SelectedIndex--
		}
		d.ensureSelectedVisible()
	case key.Matches(msg, d.KeyMap.Down):
		if d.SelectedIndex < len(d.FilteredTasks)-1 {
			d.SelectedIndex++
		}
		d.ensureSelectedVisible()
	case key.Matches(msg, d.KeyMap.Enter):
		if task := d.GetSelectedTask(); task != nil {
			return DialogResultConfirm, func() tea.Msg {
				return ComplexityReportResultMsg{Action: "select", TaskID: task.TaskID}
			}
		}
	case key.Matches(msg, d.KeyMap.SortMode):
		d.cycleSortOrder()
	case key.Matches(msg, d.KeyMap.Filter):
		return DialogResultNone, func() tea.Msg {
			return ComplexityReportResultMsg{Action: "filter"}
		}
	case key.Matches(msg, d.KeyMap.Export):
		return DialogResultNone, func() tea.Msg {
			return ComplexityReportResultMsg{Action: "export"}
		}
	case key.Matches(msg, d.KeyMap.Help):
		d.ShowHelp = !d.ShowHelp
		d.SetSize(d.width, d.height)
	case key.Matches(msg, d.KeyMap.Close):
		return DialogResultClose, func() tea.Msg {
			return DialogResultMsg{ID: d.ID, Button: "close"}
		}
	}
	return DialogResultNone, nil
}

// ensureSelectedVisible scrolls the viewport to make the selected item visible
func (d *ComplexityReportDialog) ensureSelectedVisible() {
	if d.SelectedIndex < 0 || len(d.FilteredTasks) == 0 {
		return
	}

	// Estimate the position of the selected item
	itemHeight := 1 // Each task takes up 1 line in the table
	selectedPos := d.SelectedIndex * itemHeight

	// Adjust viewport to ensure selection is visible
	if selectedPos < d.Viewport.YOffset {
		// Scroll up to show the selected item
		d.Viewport.SetYOffset(selectedPos)
	} else if selectedPos >= d.Viewport.YOffset+d.Viewport.Height {
		// Scroll down to show the selected item
		d.Viewport.SetYOffset(selectedPos - d.Viewport.Height + itemHeight)
	}
}

// View renders the dialog
func (d *ComplexityReportDialog) View() string {
	if d.Report == nil {
		return lipgloss.NewStyle().
			Width(d.GetWidth()).
			Height(d.GetHeight()).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No complexity analysis results available.")
	}

	// Render the tabular view
	var sb strings.Builder

	// Table header
	headerStyle := lipgloss.NewStyle().Bold(true)
	header := fmt.Sprintf("%-10s %-40s %-10s %-6s", "Task ID", "Title", "Complexity", "Score")
	sb.WriteString(headerStyle.Render(header))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", d.GetWidth()))
	sb.WriteString("\n")

	// Content only if we have filtered tasks
	if len(d.FilteredTasks) > 0 {
		// Table rows
		for i, task := range d.FilteredTasks {
			// Determine row style based on selection
			var rowStyle lipgloss.Style
			if i == d.SelectedIndex {
				rowStyle = lipgloss.NewStyle().Background(lipgloss.Color(d.Style.ButtonColor)).Foreground(lipgloss.Color("#FFFFFF"))
			} else {
				rowStyle = lipgloss.NewStyle()
			}

			// Determine complexity level style
			var levelStyle lipgloss.Style
			switch task.Level {
			case taskmaster.ComplexityLow:
				levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
			case taskmaster.ComplexityMedium:
				levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
			case taskmaster.ComplexityHigh:
				levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("orange"))
			case taskmaster.ComplexityVeryHigh:
				levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
			}

			// Format title for limited width
			title := task.Title
			if len(title) > 38 {
				title = title[:35] + "..."
			}

			// Format row
			levelStr := levelStyle.Render(string(task.Level))
			row := fmt.Sprintf("%-10s %-40s %-10s %-6d", task.TaskID, title, levelStr, task.Score)
			sb.WriteString(rowStyle.Render(row))
			sb.WriteString("\n")
		}
	} else {
		// No results after filtering
		sb.WriteString(lipgloss.NewStyle().
			Italic(true).
			Align(lipgloss.Center, lipgloss.Center).
			Render("\n\nNo tasks match the current filters.\n\n"))
	}

	// Append metadata
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Analyzed %d tasks, showing %d",
		len(d.Report.Tasks),
		len(d.FilteredTasks)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Analyzed at: %s",
		d.Report.AnalyzedAt.Format(time.RFC3339)))
	sb.WriteString("\n")

	// Prepare content for viewport
	d.Viewport.SetContent(sb.String())

	// Render legend if enabled
	legendView := ""
	if d.ShowLegend {
		lowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
		mediumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
		highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("orange"))
		veryHighStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))

		legendView = "\n" + lipgloss.NewStyle().Render(
			fmt.Sprintf("Legend: %s, %s, %s, %s | Sorting by: %s",
				lowStyle.Render("Low"),
				mediumStyle.Render("Medium"),
				highStyle.Render("High"),
				veryHighStyle.Render("Very High"),
				d.getSortDescription(),
			),
		)
	}

	// Render help if enabled
	helpView := ""
	if d.ShowHelp {
		helpView = "\n" + d.Help.View(d.KeyMap)
	}

	// Combine all parts
	views := []string{
		d.Viewport.View(),
		legendView,
		helpView,
	}

	// Wrap in dialog style
	return d.RenderBorder(strings.Join(views, ""))
}

// getSortDescription returns a user-friendly description of the current sort order
func (d *ComplexityReportDialog) getSortDescription() string {
	switch d.SortOrder {
	case SortByTaskID:
		return "Task ID"
	case SortByScoreAsc:
		return "Score (ascending)"
	case SortByScoreDesc:
		return "Score (descending)"
	case SortByLevelAsc:
		return "Complexity (low to high)"
	case SortByLevelDesc:
		return "Complexity (high to low)"
	default:
		return "Unknown"
	}
}

// ComplexityReportResultMsg is the message sent when the report dialog completes
type ComplexityReportResultMsg struct {
	Action string // "select", "filter", "export", or "close"
	TaskID string // Only set if action is "select"
}

// DialogSetFilterMsg is used to update the filter settings for the report
type DialogSetFilterMsg struct {
	Settings interface{}
}
