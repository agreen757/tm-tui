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
	filteringEnabled bool
	filterMode       bool
	filterValue      string
	onFilterChange   func(string)
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
func (b *BaseFilterable) EnterFilterMode() {
	if b.filteringEnabled {
		b.filterMode = true
		b.filterValue = ""
	}
}

// ExitFilterMode transitions the component out of filtering mode.
// This is typically triggered by the user pressing a key to exit filtering (like 'Esc').
// The filter value is preserved but filter mode is deactivated.
func (b *BaseFilterable) ExitFilterMode() {
	b.filterMode = false
}

// SetOnFilterChange sets a callback function that will be called whenever the filter value changes.
// This is useful for components that need to react to filter changes in real-time.
func (b *BaseFilterable) SetOnFilterChange(callback func(string)) {
	b.onFilterChange = callback
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
