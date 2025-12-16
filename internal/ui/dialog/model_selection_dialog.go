package dialog

import (
	"fmt"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types for ModelSelectionDialog

// ModelSelectionMsg is sent when a model is selected
type ModelSelectionMsg struct {
	Provider  string
	ModelName string
	ModelID   string
}

// ModelSelectionDialogKeyMap defines keybindings
type ModelSelectionDialogKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Help     key.Binding
}

// DefaultModelSelectionKeyMap returns the default keybindings
func DefaultModelSelectionKeyMap() ModelSelectionDialogKeyMap {
	return ModelSelectionDialogKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdn"),
			key.WithHelp("pgdn", "page down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next section"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev section"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ModelSelectionDialog manages the model selection UI
type ModelSelectionDialog struct {
	keyMap        ModelSelectionDialogKeyMap
	providers     []string
	selectedIndex int
	focusedPane   string // "providers" or "models"
	modelIndex    int
	selectedModel config.AvailableModel
	loadError     string
	width         int
	height        int
	x             int
	y             int
	zIndex        int
	focused       bool
	cancellable   bool
}

// NewModelSelectionDialog creates a new model selection dialog
func NewModelSelectionDialog() *ModelSelectionDialog {
	// Load providers
	providers := config.GetAvailableProviders()
	if len(providers) == 0 {
		providers = []string{"anthropic"}
	}

	return &ModelSelectionDialog{
		keyMap:        DefaultModelSelectionKeyMap(),
		providers:     providers,
		selectedIndex: 0,
		focusedPane:   "providers",
		modelIndex:    0,
		cancellable:   true,
		focused:       true,
	}
}

// Init initializes the dialog
func (d *ModelSelectionDialog) Init() tea.Cmd {
	// Load current model config
	provider, _, err := config.LoadModelConfig()
	if err == nil {
		// Find the provider index
		for i, p := range d.providers {
			if p == provider {
				d.selectedIndex = i
				d.focusedPane = "providers"
				break
			}
		}
	}
	return nil
}

// Update processes events
func (d *ModelSelectionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keyMap.Cancel):
			return d, nil
		case key.Matches(msg, d.keyMap.Confirm):
			if d.focusedPane == "providers" {
				// Switch to models pane
				d.focusedPane = "models"
				d.modelIndex = 0
			} else {
				// Confirm model selection
				provider := d.providers[d.selectedIndex]
				models := config.GetModelsByProvider(provider)
				if d.modelIndex < len(models) {
					model := models[d.modelIndex]
					return d, func() tea.Msg {
						return ModelSelectionMsg{
							Provider:  model.Provider,
							ModelName: model.ModelID,
							ModelID:   model.ModelID,
						}
					}
				}
			}
		case key.Matches(msg, d.keyMap.Up):
			if d.focusedPane == "providers" && d.selectedIndex > 0 {
				d.selectedIndex--
			} else if d.focusedPane == "models" && d.modelIndex > 0 {
				d.modelIndex--
			}
		case key.Matches(msg, d.keyMap.Down):
			if d.focusedPane == "providers" && d.selectedIndex < len(d.providers)-1 {
				d.selectedIndex++
			} else if d.focusedPane == "models" {
				provider := d.providers[d.selectedIndex]
				models := config.GetModelsByProvider(provider)
				if d.modelIndex < len(models)-1 {
					d.modelIndex++
				}
			}
		case key.Matches(msg, d.keyMap.Tab):
			if d.focusedPane == "providers" {
				d.focusedPane = "models"
			} else {
				d.focusedPane = "providers"
			}
		case key.Matches(msg, d.keyMap.ShiftTab):
			if d.focusedPane == "models" {
				d.focusedPane = "providers"
			} else {
				d.focusedPane = "providers"
			}
		}
	}
	return d, nil
}

// View renders the dialog
func (d *ModelSelectionDialog) View() string {
	var s string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("Select AI Model Provider and Model")
	s += title + "\n\n"

	// Providers section
	s += d.renderProviders()
	s += "\n"

	// Models section
	if d.focusedPane == "models" {
		s += d.renderModels()
	} else {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			"(Press Enter or Tab to select a model)",
		)
	}

	// Help
	s += "\n\n" + d.renderHelp()

	if d.loadError != "" {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(
			"Error: "+d.loadError,
		)
	}

	return s
}

// renderProviders renders the providers list
func (d *ModelSelectionDialog) renderProviders() string {
	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	if d.focusedPane == "providers" {
		paneStyle = paneStyle.BorderForeground(lipgloss.Color("39"))
	}

	var items []string
	for i, provider := range d.providers {
		item := "  " + provider
		if i == d.selectedIndex {
			item = lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("16")).
				Render("▶ " + provider + " ")
		}
		items = append(items, item)
	}

	content := "Providers:\n"
	for _, item := range items {
		content += item + "\n"
	}

	return paneStyle.Render(content)
}

// renderModels renders the models list for selected provider
func (d *ModelSelectionDialog) renderModels() string {
	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	if d.focusedPane == "models" {
		paneStyle = paneStyle.BorderForeground(lipgloss.Color("39"))
	}

	provider := d.providers[d.selectedIndex]
	models := config.GetModelsByProvider(provider)

	var items []string
	for i, model := range models {
		item := "  " + model.ModelName + " (" + model.ModelID + ")"
		if i == d.modelIndex {
			item = lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("16")).
				Render("▶ " + model.ModelName + " (" + model.ModelID + ") ")
		}
		items = append(items, item)
	}

	content := fmt.Sprintf("Models (%s):\n", provider)
	for _, item := range items {
		content += item + "\n"
	}

	return paneStyle.Render(content)
}

// renderHelp renders the help text
func (d *ModelSelectionDialog) renderHelp() string {
	help := []string{
		"↑/↓/k/j: Navigate",
		"Tab: Switch panes",
		"Enter: Select",
		"Esc: Cancel",
		"?: Help",
	}

	helpText := ""
	for _, h := range help {
		helpText += lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(h) + "  "
	}
	return helpText
}

// HandleKey processes key messages (Dialog interface)
func (d *ModelSelectionDialog) HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	// Delegate to Update method
	_, cmd := d.Update(msg)
	return DialogResultNone, cmd
}

// SetRect sets the dialog dimensions and position (Dialog interface)
func (d *ModelSelectionDialog) SetRect(width, height, x, y int) {
	d.width = width
	d.height = height
	d.x = x
	d.y = y
}

// GetRect returns the dialog dimensions and position (Dialog interface)
func (d *ModelSelectionDialog) GetRect() (width, height, x, y int) {
	return d.width, d.height, d.x, d.y
}

// Title returns the dialog title (Dialog interface)
func (d *ModelSelectionDialog) Title() string {
	return "Select Model"
}

// Kind returns the dialog kind (Dialog interface)
func (d *ModelSelectionDialog) Kind() DialogKind {
	return DialogKindForm
}

// ZIndex returns the dialog's z-index (Dialog interface)
func (d *ModelSelectionDialog) ZIndex() int {
	return d.zIndex
}

// SetZIndex sets the dialog's z-index (Dialog interface)
func (d *ModelSelectionDialog) SetZIndex(z int) {
	d.zIndex = z
}

// IsFocused returns whether the dialog has focus (Dialog interface)
func (d *ModelSelectionDialog) IsFocused() bool {
	return d.focused
}

// SetFocused sets the dialog's focus state (Dialog interface)
func (d *ModelSelectionDialog) SetFocused(focused bool) {
	d.focused = focused
}

// IsCancellable returns whether the dialog can be cancelled (Dialog interface)
func (d *ModelSelectionDialog) IsCancellable() bool {
	return d.cancellable
}
