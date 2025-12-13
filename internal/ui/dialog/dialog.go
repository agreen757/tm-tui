package dialog

import (
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogType represents the type of dialog
type DialogType int

const (
	// DialogTypeModal is a basic modal dialog
	DialogTypeModal DialogType = iota
	// DialogTypeForm is a form dialog with input fields
	DialogTypeForm
	// DialogTypeList is a list dialog with selectable items
	DialogTypeList
	// DialogTypeConfirmation is a confirmation dialog with yes/no options
	DialogTypeConfirmation
	// DialogTypeProgress is a progress dialog with a progress bar
	DialogTypeProgress
	// DialogTypeCustom is a custom dialog type used for advanced layouts
	DialogTypeCustom
)

// DialogKind is preserved for backwards compatibility with older code.
type DialogKind = DialogType

const (
	DialogKindModal        DialogKind = DialogTypeModal
	DialogKindForm         DialogKind = DialogTypeForm
	DialogKindList         DialogKind = DialogTypeList
	DialogKindConfirmation DialogKind = DialogTypeConfirmation
	DialogKindProgress     DialogKind = DialogTypeProgress
	DialogKindCustom       DialogKind = DialogTypeCustom
)

// DialogResult represents the result of a dialog operation
type DialogResult int

const (
	// DialogResultNone indicates no result yet
	DialogResultNone DialogResult = iota
	// DialogResultClose indicates the dialog should be closed
	DialogResultClose
	// DialogResultCancel indicates the dialog was cancelled
	DialogResultCancel
	// DialogResultConfirm indicates the dialog was confirmed
	DialogResultConfirm
)

// ShortcutHint represents an instructional footer entry.
type ShortcutHint struct {
	Key   string
	Label string
}

// Dialog is the interface all dialog types must implement
type Dialog interface {
	// Init initializes the dialog
	Init() tea.Cmd
	// Update processes messages and updates dialog state
	Update(msg tea.Msg) (Dialog, tea.Cmd)
	// View renders the dialog
	View() string
	// HandleKey processes a key event
	HandleKey(msg tea.KeyMsg) (DialogResult, tea.Cmd)
	// SetRect sets the dialog's dimensions and position
	SetRect(width, height, x, y int)
	// GetRect returns the dialog's dimensions and position
	GetRect() (width, height, x, y int)
	// Title returns the dialog's title
	Title() string
	// Kind returns the dialog kind
	Kind() DialogKind
	// ZIndex returns the dialog's z-index
	ZIndex() int
	// SetZIndex sets the dialog's z-index
	SetZIndex(z int)
	// IsFocused returns whether the dialog has focus
	IsFocused() bool
	// SetFocused sets the dialog's focus state
	SetFocused(focused bool)
	// IsCancellable returns whether the dialog can be cancelled
	IsCancellable() bool
}

// BaseDialog implements common functionality for all dialog types
type BaseDialog struct {
	ID          string
	TitleText   string
	Description string
	width       int
	height      int
	x           int
	y           int
	zIndex      int
	focused     bool
	cancellable bool
	kind        DialogKind
	Overlay     bool
	Style       *DialogStyle
	footerHints []ShortcutHint
}

// DialogStyle contains styling information for dialogs
type DialogStyle struct {
	Border             lipgloss.Border
	BorderColor        lipgloss.Color
	FocusedBorderColor lipgloss.Color
	TitleColor         lipgloss.Color
	BackgroundColor    lipgloss.Color
	TextColor          lipgloss.Color
	ButtonColor        lipgloss.Color
	ErrorColor         lipgloss.Color
	SuccessColor       lipgloss.Color
	WarningColor       lipgloss.Color
}

// DefaultDialogStyleFunc is a function that returns the default dialog style
// This can be overridden to customize the default style
var DefaultDialogStyle = func() *DialogStyle {
	return &DialogStyle{
		Border:             lipgloss.RoundedBorder(),
		BorderColor:        lipgloss.Color("#444444"),
		FocusedBorderColor: lipgloss.Color("#6D98BA"),
		TitleColor:         lipgloss.Color("#EEEEEE"),
		BackgroundColor:    lipgloss.Color("#333333"),
		TextColor:          lipgloss.Color("#DDDDDD"),
		ButtonColor:        lipgloss.Color("#6D98BA"),
		ErrorColor:         lipgloss.Color("#F7768E"),
		SuccessColor:       lipgloss.Color("#9ECE6A"),
		WarningColor:       lipgloss.Color("#E0AF68"),
	}
}

// NewBaseDialog creates a new base dialog
func NewBaseDialog(title string, width, height int, kind DialogKind) BaseDialog {
	return BaseDialog{
		TitleText:   title,
		width:       width,
		height:      height,
		zIndex:      0,
		focused:     true,
		cancellable: true,
		kind:        kind,
		Style:       DefaultDialogStyle(),
	}
}

// SetFooterHints replaces the footer shortcuts shown beneath the dialog.
func (d *BaseDialog) SetFooterHints(hints ...ShortcutHint) {
	if hints == nil {
		d.footerHints = nil
		return
	}
	d.footerHints = append([]ShortcutHint{}, filterShortcutHints(hints)...)
}

// AddFooterHint appends a single shortcut hint to the footer.
func (d *BaseDialog) AddFooterHint(key, label string) {
	if key == "" || label == "" {
		return
	}
	d.footerHints = append(d.footerHints, ShortcutHint{Key: key, Label: label})
}

// FooterHints returns a copy of the currently configured hints.
func (d BaseDialog) FooterHints() []ShortcutHint {
	if len(d.footerHints) == 0 {
		return nil
	}
	hints := make([]ShortcutHint, len(d.footerHints))
	copy(hints, d.footerHints)
	return hints
}

// Title returns the dialog's title
func (d BaseDialog) Title() string {
	return d.TitleText
}

// Kind returns the dialog kind
func (d BaseDialog) Kind() DialogKind {
	return d.kind
}

// ZIndex returns the dialog's z-index
func (d BaseDialog) ZIndex() int {
	return d.zIndex
}

// SetZIndex sets the dialog's z-index
func (d *BaseDialog) SetZIndex(z int) {
	d.zIndex = z
}

// IsFocused returns whether the dialog has focus
func (d BaseDialog) IsFocused() bool {
	return d.focused
}

// SetFocused sets the dialog's focus state
func (d *BaseDialog) SetFocused(focused bool) {
	d.focused = focused
}

// IsCancellable returns whether the dialog can be cancelled
func (d BaseDialog) IsCancellable() bool {
	return d.cancellable
}

// SetCancellable sets whether the dialog can be cancelled
func (d *BaseDialog) SetCancellable(cancellable bool) {
	d.cancellable = cancellable
}

// SetRect sets the dialog's dimensions and position
func (d *BaseDialog) SetRect(width, height, x, y int) {
	d.width = width
	d.height = height
	d.x = x
	d.y = y
}

// GetRect returns the dialog's dimensions and position
func (d BaseDialog) GetRect() (width, height, x, y int) {
	return d.width, d.height, d.x, d.y
}

// Center centers the dialog within the given dimensions
func (d *BaseDialog) Center(containerWidth, containerHeight int) {
	d.x = (containerWidth - d.width) / 2
	d.y = (containerHeight - d.height) / 2
}

// RenderBorder adds a border and title to content
func (d BaseDialog) RenderBorder(content string) string {
	content = strings.TrimRight(content, "\n")
	if footer := d.renderFooter(); footer != "" {
		if content != "" {
			content += "\n\n" + footer
		} else {
			content = footer
		}
	}

	borderColor := d.Style.BorderColor
	if d.focused {
		borderColor = d.Style.FocusedBorderColor
	}

	style := lipgloss.NewStyle().
		Padding(0, 1).
		BorderStyle(d.Style.Border).
		BorderForeground(borderColor).
		Width(d.width - 2).  // Account for border
		Height(d.height - 2) // Account for border

	if d.TitleText != "" {
		style = style.BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)
	}

	renderer := style.Render(content)

	if d.TitleText != "" {
		// Add title in the center of the top border
		titleStyle := lipgloss.NewStyle().
			Foreground(d.Style.TitleColor).
			Background(d.Style.BackgroundColor).
			Bold(true)

		titleText := " " + d.TitleText + " "
		titleRendered := titleStyle.Render(titleText)

		titlePos := (d.width - lipgloss.Width(titleRendered)) / 2
		if titlePos < 0 {
			titlePos = 0
		}

		// Get the first line of the rendered content
		firstLine := renderer[:d.width]

		// Replace part of the first line with the title
		if titlePos+len(titleRendered) <= len(firstLine) {
			renderer = firstLine[:titlePos] + titleRendered + firstLine[titlePos+len(titleRendered):] + renderer[d.width:]
		}
	}

	return renderer
}

func (d BaseDialog) renderFooter() string {
	if len(d.footerHints) == 0 || d.Style == nil {
		return ""
	}
	keyStyle := lipgloss.NewStyle().
		Foreground(d.Style.ButtonColor).
		Bold(true)
	labelStyle := lipgloss.NewStyle().
		Foreground(d.Style.TextColor)

	parts := make([]string, 0, len(d.footerHints))
	for _, hint := range d.footerHints {
		if hint.Key == "" || hint.Label == "" {
			continue
		}
		parts = append(parts, keyStyle.Render(hint.Key)+": "+labelStyle.Render(hint.Label))
	}
	if len(parts) == 0 {
		return ""
	}

	width := d.width - 4
	if width < 1 {
		width = 1
	}

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Left).
		PaddingTop(1).
		Render(strings.Join(parts, "  "))
}

func filterShortcutHints(hints []ShortcutHint) []ShortcutHint {
	out := make([]ShortcutHint, 0, len(hints))
	for _, hint := range hints {
		if hint.Key == "" || hint.Label == "" {
			continue
		}
		out = append(out, hint)
	}
	return out
}

// HandleBaseKey handles common key events for all dialogs
func (d BaseDialog) HandleBaseKey(msg tea.KeyMsg) (DialogResult, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		if d.cancellable {
			return DialogResultCancel, nil
		}
	}
	return DialogResultNone, nil
}

// DialogResultProvider allows a dialog to expose a return value when it closes.
type DialogResultProvider interface {
	DialogResultValue() (interface{}, error)
}

// DialogCallback is invoked when a dialog finishes.
type DialogCallback func(value interface{}, err error) tea.Cmd

type dialogEntry struct {
	dialog           Dialog
	callback         DialogCallback
	previousFocusIdx int // Track previous focused element index for restoration
}

func (e dialogEntry) ZIndex() int {
	if e.dialog != nil {
		return e.dialog.ZIndex()
	}
	return 0
}

// DialogManager manages a stack of dialogs with proper focus management
// Features:
// - Focus trapping within active dialog only
// - Focus restoration to previous element when dialogs close
// - Visible focus indicators across all themes
// - Support for nested dialog stacking with proper focus recovery
type DialogManager struct {
	dialogs         []dialogEntry
	termWidth       int
	termHeight      int
	activeDialog    int
	Style           *DialogStyle
	positioningCfg  PositioningConfig
	lastWindowWidth int // Track last terminal width for detecting changes
	lastWindowHeight int // Track last terminal height for detecting changes
	focusHistory    []int // Track focus indices for nested dialogs
}

// NewDialogManager creates a new dialog manager
func NewDialogManager(termWidth, termHeight int) *DialogManager {
	return &DialogManager{
		dialogs:         []dialogEntry{},
		termWidth:       termWidth,
		termHeight:      termHeight,
		activeDialog:    -1,
		Style:           DefaultDialogStyle(),
		positioningCfg:  DefaultPositioningConfig(),
		lastWindowWidth: termWidth,
		lastWindowHeight: termHeight,
		focusHistory:    []int{},
	}
}

// PushDialog adds a dialog to the top of the stack without a callback.
func (m *DialogManager) PushDialog(dialog Dialog) {
	m.AddDialog(dialog, nil)
}

// AddDialog adds a dialog with a completion callback.
func (m *DialogManager) AddDialog(dialog Dialog, callback DialogCallback) {
	if dialog == nil {
		return
	}

	dialog.SetZIndex(len(m.dialogs))

	// Position the dialog properly within terminal bounds
	if m.termWidth > 0 && m.termHeight > 0 {
		width, height, _, _ := dialog.GetRect()
		if width > 0 && height > 0 {
			// Use positioning utility for proper centering and bounds checking
			pos := PositionDialogInBoundsWithConfig(
				m.termWidth, m.termHeight, width, height,
				StrategyCenter, m.positioningCfg,
			)
			dialog.SetRect(pos.Width, pos.Height, pos.X, pos.Y)
		}
	}

	// Save the focused element index of the current active dialog if it exists
	previousFocusIdx := -1
	if m.activeDialog >= 0 && m.activeDialog < len(m.dialogs) {
		if focusable, ok := m.dialogs[m.activeDialog].dialog.(FocusableDialog); ok {
			previousFocusIdx = focusable.FocusedIndex()
		}
		m.dialogs[m.activeDialog].dialog.SetFocused(false)
	}

	// Defocus other dialogs
	for i := range m.dialogs {
		m.dialogs[i].dialog.SetFocused(false)
	}

	m.dialogs = append(m.dialogs, dialogEntry{dialog: dialog, callback: callback, previousFocusIdx: previousFocusIdx})
	m.activeDialog = len(m.dialogs) - 1
	dialog.SetFocused(true)
}

// PopDialog removes the top dialog from the stack
func (m *DialogManager) PopDialog() Dialog {
	entry := m.popEntry()
	return entry.dialog
}

func (m *DialogManager) popEntry() dialogEntry {
	if len(m.dialogs) == 0 {
		return dialogEntry{}
	}

	entry := m.dialogs[len(m.dialogs)-1]
	m.dialogs = m.dialogs[:len(m.dialogs)-1]

	if len(m.dialogs) > 0 {
		m.activeDialog = len(m.dialogs) - 1
		m.dialogs[m.activeDialog].dialog.SetFocused(true)

		// Restore focus to the previous element if the dialog is focusable
		if focusable, ok := m.dialogs[m.activeDialog].dialog.(FocusableDialog); ok {
			if m.dialogs[m.activeDialog].previousFocusIdx >= 0 {
				focusable.SetFocusedIndex(m.dialogs[m.activeDialog].previousFocusIdx)
			}
		}
	} else {
		m.activeDialog = -1
	}

	return entry
}

// GetActiveDialog returns the active dialog
func (m *DialogManager) GetActiveDialog() Dialog {
	if m.activeDialog >= 0 && m.activeDialog < len(m.dialogs) {
		return m.dialogs[m.activeDialog].dialog
	}
	return nil
}

// HasDialogs returns true if there are any dialogs
func (m *DialogManager) HasDialogs() bool {
	return len(m.dialogs) > 0
}

// SetTerminalSize sets the terminal size for the dialog manager
// This should be called when the terminal is resized
func (m *DialogManager) SetTerminalSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	// Skip if size hasn't changed
	if m.lastWindowWidth == width && m.lastWindowHeight == height {
		return
	}

	m.termWidth = width
	m.termHeight = height
	m.lastWindowWidth = width
	m.lastWindowHeight = height

	// Reposition all dialogs for the new terminal size
	m.repositionDialogsForResize(width, height)
}

// repositionDialogsForResize repositions all dialogs after terminal resize
func (m *DialogManager) repositionDialogsForResize(newWidth, newHeight int) {
	for i := range m.dialogs {
		d := m.dialogs[i].dialog
		width, height, x, y := d.GetRect()

		// Check if current position is still valid
		if !IsDialogFullyVisible(x, y, width, height, newWidth, newHeight) {
			// Recalculate position using positioning strategy
			newPos := PositionDialogInBoundsWithConfig(
				newWidth, newHeight, width, height,
				StrategyCenter, m.positioningCfg,
			)
			d.SetRect(newPos.Width, newPos.Height, newPos.X, newPos.Y)
		}
	}
}

// HandleMsg processes tea.Msg for dialogs
func (m *DialogManager) HandleMsg(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}

	// If no dialogs, nothing to do
	if len(m.dialogs) == 0 {
		return nil
	}

	// Handle terminal resize for all dialogs
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		// Update terminal dimensions
		m.SetTerminalSize(msg.Width, msg.Height)

		// Notify dialogs of resize if they need to handle it
		for i := range m.dialogs {
			// Update dialog with window size message so internal content can adjust
			updatedDialog, cmd := m.dialogs[i].dialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			m.dialogs[i].dialog = updatedDialog
		}

		if len(cmds) > 0 {
			return tea.Batch(cmds...)
		}
		return nil
	}

	// Only the active dialog receives key events
	if m.activeDialog >= 0 {
		activeEntry := m.dialogs[m.activeDialog]
		activeDialog := activeEntry.dialog

		// Special handling for key messages - check if dialog handles or closes
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			result, cmd := activeDialog.HandleKey(keyMsg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			// Handle dialog result
			switch result {
			case DialogResultClose, DialogResultCancel, DialogResultConfirm:
				popped := m.popEntry()
				if popped.callback != nil {
					var value interface{}
					var err error
					if result == DialogResultConfirm || result == DialogResultClose {
						if provider, ok := popped.dialog.(DialogResultProvider); ok {
							value, err = provider.DialogResultValue()
						}
					}
					cbCmd := popped.callback(value, err)
					if cbCmd != nil {
						cmds = append(cmds, cbCmd)
					}
				}
				return tea.Batch(cmds...)
			}
		} else {
			// For non-key messages, update the dialog
			updatedDialog, cmd := activeDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			// Update the dialog in the stack
			m.dialogs[m.activeDialog].dialog = updatedDialog
		}
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

// RemoveDialogsByType removes all dialogs that match the provided type.
func (m *DialogManager) RemoveDialogsByType(dialogType DialogType) {
	filtered := make([]dialogEntry, 0, len(m.dialogs))
	for _, entry := range m.dialogs {
		if entry.dialog.Kind() == dialogType {
			continue
		}
		filtered = append(filtered, entry)
	}
	m.dialogs = filtered
	if len(m.dialogs) == 0 {
		m.activeDialog = -1
	} else {
		m.activeDialog = len(m.dialogs) - 1
		m.dialogs[m.activeDialog].dialog.SetFocused(true)
	}
}

// GetDialogByType returns the top-most dialog matching the provided type.
func (m *DialogManager) GetDialogByType(dialogType DialogType) (Dialog, bool) {
	for i := len(m.dialogs) - 1; i >= 0; i-- {
		if m.dialogs[i].dialog.Kind() == dialogType {
			return m.dialogs[i].dialog, true
		}
	}
	return nil, false
}

// GetDialogByModel attempts to find a dialog by its struct name.
func (m *DialogManager) GetDialogByModel(name string) (Dialog, bool) {
	for i := len(m.dialogs) - 1; i >= 0; i-- {
		d := m.dialogs[i].dialog
		typeName := reflect.TypeOf(d)
		if typeName == nil {
			continue
		}
		if typeName.Kind() == reflect.Pointer {
			typeName = typeName.Elem()
		}
		if typeName.Name() == name {
			return d, true
		}
	}
	return nil, false
}

// View renders all dialogs in the stack
func (m *DialogManager) View() string {
	if len(m.dialogs) == 0 {
		return ""
	}

	var renderedDialogs []string
	for _, entry := range m.dialogs {
		renderedDialogs = append(renderedDialogs, entry.dialog.View())
	}

	return renderedDialogs[len(renderedDialogs)-1]
}

// DialogResultMsg is emitted by dialogs that need to communicate actions back to the model.
type DialogResultMsg struct {
	ID     string
	Button string
	Value  interface{}
}
