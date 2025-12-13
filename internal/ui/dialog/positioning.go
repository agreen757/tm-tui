package dialog

import (
	"math"
)

// PositioningConfig holds configuration for dialog positioning
type PositioningConfig struct {
	MinDialogWidth  int
	MinDialogHeight int
	Padding         int
}

// DefaultPositioningConfig returns the default positioning configuration
func DefaultPositioningConfig() PositioningConfig {
	return PositioningConfig{
		MinDialogWidth:  20,
		MinDialogHeight: 6,
		Padding:         1,
	}
}

// DialogPosition represents a dialog's position and size
type DialogPosition struct {
	X      int
	Y      int
	Width  int
	Height int
}

// PositioningStrategy defines how dialogs should be positioned
type PositioningStrategy int

const (
	// StrategyCenter positions dialog in the center of the terminal
	StrategyCenter PositioningStrategy = iota
	// StrategyTopCenter positions dialog at top-center of the terminal
	StrategyTopCenter
	// StrategyTopLeft positions dialog at top-left corner
	StrategyTopLeft
)

// PositionDialogInBounds calculates optimal dialog position within terminal bounds
// with proper centering and bounds checking
func PositionDialogInBounds(termWidth, termHeight, dialogWidth, dialogHeight int, strategy PositioningStrategy) DialogPosition {
	config := DefaultPositioningConfig()
	return PositionDialogInBoundsWithConfig(termWidth, termHeight, dialogWidth, dialogHeight, strategy, config)
}

// PositionDialogInBoundsWithConfig calculates optimal dialog position with custom config
func PositionDialogInBoundsWithConfig(termWidth, termHeight, dialogWidth, dialogHeight int, strategy PositioningStrategy, config PositioningConfig) DialogPosition {
	// Ensure terminal dimensions are valid
	if termWidth < 0 {
		termWidth = 0
	}
	if termHeight < 0 {
		termHeight = 0
	}

	// Apply graceful degradation for small terminal sizes
	adjustedWidth, adjustedHeight := DegradeDialogSize(dialogWidth, dialogHeight, termWidth, termHeight, config)

	// Calculate position based on strategy
	var x, y int

	switch strategy {
	case StrategyTopCenter:
		x = (termWidth - adjustedWidth) / 2
		y = config.Padding
		if y < 0 {
			y = 0
		}

	case StrategyTopLeft:
		x = config.Padding
		y = config.Padding
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}

	case StrategyCenter:
		fallthrough
	default:
		x = (termWidth - adjustedWidth) / 2
		y = (termHeight - adjustedHeight) / 2
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
	}

	// Ensure dialog doesn't overflow terminal bounds
	x, y = ClampDialogPosition(x, y, adjustedWidth, adjustedHeight, termWidth, termHeight)

	return DialogPosition{
		X:      x,
		Y:      y,
		Width:  adjustedWidth,
		Height: adjustedHeight,
	}
}

// DegradeDialogSize reduces dialog size gracefully when terminal is too small
func DegradeDialogSize(desiredWidth, desiredHeight, termWidth, termHeight int, config PositioningConfig) (int, int) {
	width := desiredWidth
	height := desiredHeight

	// Minimum viable sizes
	minWidth := config.MinDialogWidth
	minHeight := config.MinDialogHeight

	// If terminal is very small, constrain to absolute minimum
	if termWidth < minWidth+2*config.Padding {
		width = termWidth - 2*config.Padding
		if width < 4 {
			width = termWidth
		}
	} else if width > termWidth-2*config.Padding {
		width = termWidth - 2*config.Padding
	}

	if termHeight < minHeight+2*config.Padding {
		height = termHeight - 2*config.Padding
		if height < 2 {
			height = termHeight
		}
	} else if height > termHeight-2*config.Padding {
		height = termHeight - 2*config.Padding
	}

	// Ensure we never go below minimum (unless terminal is impossibly small)
	if width < minWidth {
		width = minWidth
	}
	if height < minHeight {
		height = minHeight
	}

	return width, height
}

// ClampDialogPosition ensures the dialog position doesn't overflow terminal bounds
func ClampDialogPosition(x, y, width, height, termWidth, termHeight int) (int, int) {
	// Clamp x position
	if x+width > termWidth {
		x = termWidth - width
	}
	if x < 0 {
		x = 0
	}

	// Clamp y position
	if y+height > termHeight {
		y = termHeight - height
	}
	if y < 0 {
		y = 0
	}

	return x, y
}

// CalculateOptimalDialogSize calculates an optimal dialog size for the given terminal
// based on the desired content width/height
func CalculateOptimalDialogSize(contentWidth, contentHeight, termWidth, termHeight int) (int, int) {
	// Account for border (2 chars on each side) and padding (1 char on each side)
	dialogWidth := contentWidth + 4
	dialogHeight := contentHeight + 4

	// Additional space for title and footer
	dialogHeight += 2

	// Ensure we don't exceed terminal bounds
	if dialogWidth > termWidth-2 {
		dialogWidth = termWidth - 2
	}
	if dialogHeight > termHeight-2 {
		dialogHeight = termHeight - 2
	}

	// Enforce minimum dialog size
	minWidth := 20
	minHeight := 6
	if dialogWidth < minWidth {
		dialogWidth = minWidth
	}
	if dialogHeight < minHeight {
		dialogHeight = minHeight
	}

	return dialogWidth, dialogHeight
}

// IsTerminalTooSmall checks if the terminal is too small for normal dialog rendering
func IsTerminalTooSmall(termWidth, termHeight int) bool {
	minTermWidth := 40
	minTermHeight := 12
	return termWidth < minTermWidth || termHeight < minTermHeight
}

// GetFriendlyTerminalSizeWarning returns a message appropriate for the terminal size
func GetFriendlyTerminalSizeWarning(termWidth, termHeight int) string {
	if termWidth < 20 || termHeight < 4 {
		return "Terminal too small - some content may be truncated"
	}
	if termWidth < 40 || termHeight < 12 {
		return "Terminal size is small - dialog may be hard to read"
	}
	return ""
}

// AdjustDialogForHeight adjusts dialog content based on available height
func AdjustDialogForHeight(desiredHeight, availableHeight int) int {
	// Leave at least 2 lines of padding
	minHeight := 6
	if availableHeight < minHeight {
		return minHeight
	}
	if desiredHeight > availableHeight {
		return availableHeight - 1
	}
	return desiredHeight
}

// AdjustDialogForWidth adjusts dialog content based on available width
func AdjustDialogForWidth(desiredWidth, availableWidth int) int {
	// Leave at least 2 chars of padding on each side
	minWidth := 20
	if availableWidth < minWidth {
		return minWidth
	}
	if desiredWidth > availableWidth {
		return availableWidth - 2
	}
	return desiredWidth
}

// CalculateDialogPadding calculates appropriate padding based on terminal size
func CalculateDialogPadding(termWidth, termHeight int) int {
	if IsTerminalTooSmall(termWidth, termHeight) {
		return 0
	}
	return 1
}

// IsDialogFullyVisible checks if a dialog at the given position is fully visible
func IsDialogFullyVisible(x, y, width, height, termWidth, termHeight int) bool {
	if x < 0 || y < 0 {
		return false
	}
	if x+width > termWidth {
		return false
	}
	if y+height > termHeight {
		return false
	}
	return true
}

// RepositionDialogIfNeeded checks if a dialog needs repositioning after a resize
// and returns the new position if needed
func RepositionDialogIfNeeded(currentX, currentY, width, height, newTermWidth, newTermHeight int) (int, int, bool) {
	// Check if current position is still valid
	if IsDialogFullyVisible(currentX, currentY, width, height, newTermWidth, newTermHeight) {
		// Current position is still valid
		return currentX, currentY, false
	}

	// Need to recenter
	newX := (newTermWidth - width) / 2
	newY := (newTermHeight - height) / 2

	// Clamp to bounds
	newX, newY = ClampDialogPosition(newX, newY, width, height, newTermWidth, newTermHeight)

	return newX, newY, true
}

// ScaleDialogSize scales a dialog size proportionally to terminal size change
// useful for responsive dialogs
func ScaleDialogSize(width, height, oldTermWidth, oldTermHeight, newTermWidth, newTermHeight int) (int, int) {
	if oldTermWidth <= 0 || oldTermHeight <= 0 {
		return width, height
	}

	// Calculate scale factors (but cap at 1.0 to avoid growing dialogs)
	widthScale := math.Min(float64(newTermWidth)/float64(oldTermWidth), 1.0)
	heightScale := math.Min(float64(newTermHeight)/float64(oldTermHeight), 1.0)

	// Apply scaling
	newWidth := int(float64(width) * widthScale)
	newHeight := int(float64(height) * heightScale)

	// Ensure minimum sizes
	minWidth := 20
	minHeight := 6
	if newWidth < minWidth {
		newWidth = minWidth
	}
	if newHeight < minHeight {
		newHeight = minHeight
	}

	// Ensure we don't exceed new terminal bounds
	if newWidth > newTermWidth {
		newWidth = newTermWidth
	}
	if newHeight > newTermHeight {
		newHeight = newTermHeight
	}

	return newWidth, newHeight
}
