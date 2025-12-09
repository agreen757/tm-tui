package ui

import (
	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the keybindings for the TUI
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	
	// Task operations
	ToggleExpand key.Binding
	Collapse     key.Binding
	Expand       key.Binding
	Select       key.Binding
	NextTask     key.Binding
	Refresh      key.Binding
	JumpToID     key.Binding
	Search       key.Binding
	Filter       key.Binding
	
	// Status changes
	SetInProgress key.Binding
	SetDone       key.Binding
	SetBlocked    key.Binding
	SetCancelled  key.Binding
	SetDeferred   key.Binding
	SetPending    key.Binding
	
	// Panel focus
	FocusTaskList key.Binding
	FocusDetails  key.Binding
	FocusLog      key.Binding
	CyclePanel    key.Binding
	Back          key.Binding
	
	// Panel visibility
	ToggleDetails key.Binding
	ToggleLog     key.Binding
	
	// View modes
	ViewTree   key.Binding
	ViewList   key.Binding
	CycleView  key.Binding
	
	// Help and quit
	Help   key.Binding
	Quit   key.Binding
	Cancel key.Binding
	
	// State management
	ClearState key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "collapse"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "expand"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		
		// Task operations
		ToggleExpand: key.NewBinding(
			key.WithKeys("enter", "e"),
			key.WithHelp("enter/e", "toggle expand"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "collapse"),
		),
		Expand: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "expand"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select/deselect"),
		),
		NextTask: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next task"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		JumpToID: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "jump to task ID"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search tasks"),
		),
		Filter: key.NewBinding(
			key.WithKeys("F"),
			key.WithHelp("F", "filter by status"),
		),
		
		// Status changes
		SetInProgress: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "set in-progress"),
		),
		SetDone: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "set done"),
		),
		SetBlocked: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "set blocked"),
		),
		SetCancelled: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "set cancelled"),
		),
		SetDeferred: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "set deferred"),
		),
		SetPending: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "set pending"),
		),
		
		// Panel focus
		FocusTaskList: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "focus tasks"),
		),
		FocusDetails: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "focus details"),
		),
		FocusLog: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "focus log"),
		),
		CyclePanel: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle panels"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/clear"),
		),
		
		// Panel visibility
		ToggleDetails: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "toggle details"),
		),
		ToggleLog: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "toggle log"),
		),
		
		// View modes
		ViewTree: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tree view"),
		),
		ViewList: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "list view"),  
		),
		CycleView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "cycle view"),
		),
		
		// Help and quit
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "cancel/quit"),
		),
		
		// State management
		ClearState: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "clear TUI state"),
		),
	}
}

// NewKeyMap creates a KeyMap from configuration, falling back to defaults for missing keys
func NewKeyMap(cfg *config.Config) KeyMap {
	// Start with defaults
	km := DefaultKeyMap()
	
	// If no config or no keybindings, return defaults
	if cfg == nil || len(cfg.KeyBindings) == 0 {
		return km
	}
	
	// Helper function to get key binding or use default
	getKey := func(name, defaultKey string) string {
		if key, ok := cfg.KeyBindings[name]; ok && key != "" {
			return key
		}
		return defaultKey
	}
	
	// Override with configured keybindings
	if quitKey := getKey("quit", "q"); quitKey != "" {
		km.Quit = key.NewBinding(
			key.WithKeys(quitKey),
			key.WithHelp(quitKey, "quit"),
		)
	}
	
	if helpKey := getKey("help", "?"); helpKey != "" {
		km.Help = key.NewBinding(
			key.WithKeys(helpKey),
			key.WithHelp(helpKey, "toggle help"),
		)
	}
	
	if nextKey := getKey("next", "n"); nextKey != "" {
		km.NextTask = key.NewBinding(
			key.WithKeys(nextKey),
			key.WithHelp(nextKey, "next task"),
		)
	}
	
	if refreshKey := getKey("refresh", "r"); refreshKey != "" {
		km.Refresh = key.NewBinding(
			key.WithKeys(refreshKey),
			key.WithHelp(refreshKey, "refresh"),
		)
	}
	
	if expandKey := getKey("expand", "enter"); expandKey != "" {
		km.ToggleExpand = key.NewBinding(
			key.WithKeys(expandKey),
			key.WithHelp(expandKey, "toggle expand"),
		)
	}
	
	if detailsKey := getKey("details", "d"); detailsKey != "" {
		km.ToggleDetails = key.NewBinding(
			key.WithKeys(detailsKey),
			key.WithHelp(detailsKey, "toggle details"),
		)
	}
	
	if inProgressKey := getKey("inProgress", "i"); inProgressKey != "" {
		km.SetInProgress = key.NewBinding(
			key.WithKeys(inProgressKey),
			key.WithHelp(inProgressKey, "set in-progress"),
		)
	}
	
	if doneKey := getKey("done", "D"); doneKey != "" {
		km.SetDone = key.NewBinding(
			key.WithKeys(doneKey),
			key.WithHelp(doneKey, "set done"),
		)
	}
	
	if blockedKey := getKey("blocked", "b"); blockedKey != "" {
		km.SetBlocked = key.NewBinding(
			key.WithKeys(blockedKey),
			key.WithHelp(blockedKey, "set blocked"),
		)
	}
	
	if cancelledKey := getKey("cancelled", "c"); cancelledKey != "" {
		km.SetCancelled = key.NewBinding(
			key.WithKeys(cancelledKey),
			key.WithHelp(cancelledKey, "set cancelled"),
		)
	}
	
	return km
}

// ShortHelp returns a short help text for the status bar
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.ToggleExpand, k.Select, k.NextTask, k.Help, k.Quit}
}

// FullHelp returns the full help text
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.PageUp, k.PageDown, k.ToggleExpand, k.Select},
		{k.NextTask, k.Refresh, k.JumpToID},
		{k.SetInProgress, k.SetDone, k.SetBlocked, k.SetCancelled},
		{k.SetDeferred, k.SetPending},
		{k.FocusTaskList, k.FocusDetails, k.FocusLog, k.CyclePanel},
		{k.ToggleDetails, k.ToggleLog},
		{k.Help, k.Quit, k.Cancel, k.ClearState},
	}
}
