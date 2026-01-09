package dialog

import (
	"github.com/agreen757/tm-tui/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

// AgentSelectionMsg is sent when an agent is selected from the dialog
type AgentSelectionMsg struct {
	AgentType types.AgentType
}

// AgentItem implements ListItem for agent selection
type AgentItem struct {
	agentType   types.AgentType
	description string
}

// Title returns the display name of the agent
func (a AgentItem) Title() string {
	return a.agentType.String()
}

// Description returns the description of the agent
func (a AgentItem) Description() string {
	return a.description
}

// FilterValue returns the value to use for filtering
func (a AgentItem) FilterValue() string {
	return a.agentType.String() + " " + a.description
}

// GetAgentType returns the underlying AgentType enum value
func (a AgentItem) GetAgentType() types.AgentType {
	return a.agentType
}

// AgentSelectorDialog is a dialog for selecting the AI agent type
type AgentSelectorDialog struct {
	*ListDialog
	lastSelected types.AgentType
}

// NewAgentSelectorDialog creates a new agent selector dialog
func NewAgentSelectorDialog(width, height int) *AgentSelectorDialog {
	// Create list of available agents using AgentType enum
	agents := []AgentItem{
		{
			agentType:   types.AgentTypeCrush,
			description: "Terminal AI assistant with code execution",
		},
		{
			agentType:   types.AgentTypeGemini,
			description: "Google's Gemini AI model with advanced reasoning",
		},
	}

	// Convert to ListItem interface
	items := make([]ListItem, len(agents))
	for i, agent := range agents {
		items[i] = agent
	}

	// Create base list dialog
	listDialog := NewListDialog("Select AI Agent", width, height, items)
	listDialog.showDescription = true

	dialog := &AgentSelectorDialog{
		ListDialog:   listDialog,
		lastSelected: types.AgentTypeCrush, // Default to Crush
	}

	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	return dialog
}

// NewAgentSelectorDialogSimple creates an agent selector dialog with default settings
func NewAgentSelectorDialogSimple() *AgentSelectorDialog {
	return NewAgentSelectorDialog(60, 10)
}

// SetDefaultSelection sets the default selected agent
func (d *AgentSelectorDialog) SetDefaultSelection(agentType types.AgentType) {
	d.lastSelected = agentType

	// Find and select this agent in the list
	for i, item := range d.items {
		if agentItem, ok := item.(AgentItem); ok {
			if agentItem.GetAgentType() == agentType {
				d.selectedIndex = i
				break
			}
		}
	}
}

// GetSelectedAgent returns the selected agent type
func (d *AgentSelectorDialog) GetSelectedAgent() types.AgentType {
	if d.selectedIndex < 0 || d.selectedIndex >= len(d.items) {
		return types.AgentTypeCrush // Default
	}

	item, ok := d.items[d.selectedIndex].(AgentItem)
	if !ok {
		return types.AgentTypeCrush // Default
	}

	return item.GetAgentType()
}

// Update handles messages for the agent selector dialog
func (d *AgentSelectorDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	// Let the base ListDialog handle most updates
	updatedDialog, cmd := d.ListDialog.Update(msg)

	// Update our pointer to the ListDialog
	if listDialog, ok := updatedDialog.(*ListDialog); ok {
		d.ListDialog = listDialog
	}

	return d, cmd
}
