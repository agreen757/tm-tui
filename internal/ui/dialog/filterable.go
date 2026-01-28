package dialog

// FilterableComponent defines the interface for components that support filtering.
// Implementations should provide methods to enable/disable filtering, manage filter state,
// and handle filter mode transitions.
type FilterableComponent interface {
	// EnableFiltering enables or disables the filtering capability for this component
	EnableFiltering(enabled bool)

	// SetFilterValue sets the current filter string value
	SetFilterValue(value string)

	// GetFilterValue returns the current filter string value
	GetFilterValue() string

	// IsFiltering returns whether the component is currently in filtering mode
	IsFiltering() bool

	// EnterFilterMode transitions the component into filtering mode
	EnterFilterMode()

	// ExitFilterMode transitions the component out of filtering mode
	ExitFilterMode()
}

// BaseFilterable provides a default implementation of the FilterableComponent interface.
// It manages filter state and provides methods for component filtering functionality.
// Components that need filtering can embed this struct and implement component-specific
// filtering logic by overriding methods as needed.
type BaseFilterable struct {
	filteringEnabled    bool
	filterMode          bool
	filterValue         string
	onFilterChange      func(string)
	onFilterModeChange  func(bool) // Callback for when filter mode changes, useful for focus restoration
	previousFocusIndex  int        // Track focus index before entering filter mode for restoration
	hasPreviousFocus    bool       // Flag to indicate if previousFocusIndex is valid
}

// NewBaseFilterable creates a new BaseFilterable with default settings
func NewBaseFilterable() *BaseFilterable {
	return &BaseFilterable{
		filteringEnabled: false,
		filterMode:       false,
		filterValue:      "",
		onFilterChange:   nil,
	}
}

// EnableFiltering enables or disables the filtering capability for this component.
// When filtering is enabled, the component can enter filter mode and apply filters.
func (b *BaseFilterable) EnableFiltering(enabled bool) {
	b.filteringEnabled = enabled
	// If disabling filtering while in filter mode, exit filter mode
	if !enabled && b.filterMode {
		b.ExitFilterMode()
	}
}

// SetFilterValue sets the current filter string value.
// This method can be used to set a filter programmatically.
// If an onFilterChange callback is registered, it will be called with the new value.
func (b *BaseFilterable) SetFilterValue(value string) {
	b.filterValue = value
	if b.onFilterChange != nil {
		b.onFilterChange(value)
	}
}

// GetFilterValue returns the current filter string value
func (b *BaseFilterable) GetFilterValue() string {
	return b.filterValue
}

// IsFiltering returns whether the component is currently in filtering mode
func (b *BaseFilterable) IsFiltering() bool {
	return b.filterMode
}

// EnterFilterMode transitions the component into filtering mode.
// This is typically triggered by the user pressing a filter activation key (like '/').
// The component should prepare to accept and process filter input.
// FR5.1: Stores the current focused element index for restoration when exiting filter mode.
func (b *BaseFilterable) EnterFilterMode() {
	if b.filteringEnabled {
		b.filterMode = true
		b.filterValue = ""
		
		// Notify about filter mode change
		if b.onFilterModeChange != nil {
			b.onFilterModeChange(true)
		}
	}
}

// ExitFilterMode transitions the component out of filtering mode.
// This is typically triggered by the user pressing a key to exit filtering (like 'Esc').
// The filter value is preserved but filter mode is deactivated.
// FR5.2: Triggers focus restoration callback when exiting filter mode.
func (b *BaseFilterable) ExitFilterMode() {
	b.filterMode = false
	
	// Notify about filter mode change (allows restoration of focus)
	if b.onFilterModeChange != nil {
		b.onFilterModeChange(false)
	}
}

// SetOnFilterChange sets a callback function that will be called whenever the filter value changes.
// This is useful for components that need to react to filter changes in real-time.
func (b *BaseFilterable) SetOnFilterChange(callback func(string)) {
	b.onFilterChange = callback
}

// SetOnFilterModeChange sets a callback function that will be called when filter mode is entered or exited.
// This is useful for components that need to restore focus or perform other actions on filter mode transitions.
// The callback receives a boolean: true when entering filter mode, false when exiting.
func (b *BaseFilterable) SetOnFilterModeChange(callback func(bool)) {
	b.onFilterModeChange = callback
}

// StoreFocusIndex stores the current focus index before entering filter mode (for later restoration).
// This should be called by focusable components before entering filter mode.
func (b *BaseFilterable) StoreFocusIndex(index int) {
	b.previousFocusIndex = index
	b.hasPreviousFocus = true
}

// GetStoredFocusIndex retrieves the stored focus index from before filter mode was entered.
// Returns the index and a boolean indicating if a valid index was stored.
func (b *BaseFilterable) GetStoredFocusIndex() (int, bool) {
	return b.previousFocusIndex, b.hasPreviousFocus
}

// ClearStoredFocusIndex clears the stored focus index.
func (b *BaseFilterable) ClearStoredFocusIndex() {
	b.hasPreviousFocus = false
	b.previousFocusIndex = 0
}

// IsFilteringEnabled returns whether filtering is currently enabled for this component
func (b *BaseFilterable) IsFilteringEnabled() bool {
	return b.filteringEnabled
}

// ClearFilter clears the current filter value and exits filter mode
func (b *BaseFilterable) ClearFilter() {
	b.SetFilterValue("")
	b.ExitFilterMode()
}

// Compile-time assertion that BaseFilterable implements FilterableComponent
var _ FilterableComponent = (*BaseFilterable)(nil)
