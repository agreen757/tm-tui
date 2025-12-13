package dialog

import (
	"testing"
)

func TestPositionDialogInBounds_CenterPositioning(t *testing.T) {
	tests := []struct {
		name             string
		termWidth        int
		termHeight       int
		dialogWidth      int
		dialogHeight     int
		expectX          int
		expectY          int
		expectWidth      int
		expectHeight     int
	}{
		{
			name:         "Normal size terminal, normal dialog",
			termWidth:    100,
			termHeight:   30,
			dialogWidth:  50,
			dialogHeight: 20,
			expectX:      25,
			expectY:      5,
			expectWidth:  50,
			expectHeight: 20,
		},
		{
			name:         "Dialog exceeds terminal width",
			termWidth:    40,
			termHeight:   30,
			dialogWidth:  80,
			dialogHeight: 20,
			expectX:      1,
			expectY:      5,
			expectWidth:  38, // 40 - 2*padding(1)
			expectHeight: 20,
		},
		{
			name:         "Dialog exceeds terminal height",
			termWidth:    100,
			termHeight:   10,
			dialogWidth:  50,
			dialogHeight: 30,
			expectX:      25,
			expectY:      1,
			expectWidth:  50,
			expectHeight: 8, // 10 - 2*padding(1)
		},
		{
			name:         "Very small terminal",
			termWidth:    20,
			termHeight:   8,
			dialogWidth:  50,
			dialogHeight: 20,
			expectX:      0,
			expectY:      1,
			expectWidth:  20, // Clamped to terminal width
			expectHeight: 6,  // Minimum height
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := PositionDialogInBounds(tt.termWidth, tt.termHeight, tt.dialogWidth, tt.dialogHeight, StrategyCenter)

			if pos.X != tt.expectX {
				t.Errorf("X: got %d, expected %d", pos.X, tt.expectX)
			}
			if pos.Y != tt.expectY {
				t.Errorf("Y: got %d, expected %d", pos.Y, tt.expectY)
			}
			if pos.Width != tt.expectWidth {
				t.Errorf("Width: got %d, expected %d", pos.Width, tt.expectWidth)
			}
			if pos.Height != tt.expectHeight {
				t.Errorf("Height: got %d, expected %d", pos.Height, tt.expectHeight)
			}
		})
	}
}

func TestDegradeDialogSize(t *testing.T) {
	config := DefaultPositioningConfig()

	tests := []struct {
		name              string
		desiredWidth      int
		desiredHeight     int
		termWidth         int
		termHeight        int
		expectWidth       int
		expectHeight      int
		expectDegraded    bool
	}{
		{
			name:           "Normal size - no degradation",
			desiredWidth:   50,
			desiredHeight:  20,
			termWidth:      100,
			termHeight:     30,
			expectWidth:    50,
			expectHeight:   20,
			expectDegraded: false,
		},
		{
			name:           "Terminal narrower - degrade width",
			desiredWidth:   80,
			desiredHeight:  20,
			termWidth:      40,
			termHeight:     30,
			expectWidth:    38, // 40 - 2*padding(1)
			expectHeight:   20,
			expectDegraded: true,
		},
		{
			name:           "Terminal shorter - degrade height",
			desiredWidth:   50,
			desiredHeight:  40,
			termWidth:      100,
			termHeight:     10,
			expectWidth:    50,
			expectHeight:   8, // 10 - 2*padding(1)
			expectDegraded: true,
		},
		{
			name:           "Minimum size enforcement",
			desiredWidth:   50,
			desiredHeight:  20,
			termWidth:      10,
			termHeight:     4,
			expectWidth:    20, // Minimum width
			expectHeight:   6,  // Minimum height
			expectDegraded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := DegradeDialogSize(
				tt.desiredWidth, tt.desiredHeight,
				tt.termWidth, tt.termHeight,
				config,
			)

			if width != tt.expectWidth {
				t.Errorf("Width: got %d, expected %d", width, tt.expectWidth)
			}
			if height != tt.expectHeight {
				t.Errorf("Height: got %d, expected %d", height, tt.expectHeight)
			}
		})
	}
}

func TestClampDialogPosition(t *testing.T) {
	tests := []struct {
		name      string
		x         int
		y         int
		width     int
		height    int
		termWidth int
		termHeight int
		expectX   int
		expectY   int
	}{
		{
			name:       "Normal position - no clamping",
			x:          25,
			y:          5,
			width:      50,
			height:     20,
			termWidth:  100,
			termHeight: 30,
			expectX:    25,
			expectY:    5,
		},
		{
			name:       "Overflow right - clamp X",
			x:          75,
			y:          5,
			width:      50,
			height:     20,
			termWidth:  100,
			termHeight: 30,
			expectX:    50,
			expectY:    5,
		},
		{
			name:       "Overflow bottom - clamp Y",
			x:          25,
			y:          20,
			width:      50,
			height:     20,
			termWidth:  100,
			termHeight: 30,
			expectX:    25,
			expectY:    10,
		},
		{
			name:       "Negative position - clamp to zero",
			x:          -10,
			y:          -5,
			width:      50,
			height:     20,
			termWidth:  100,
			termHeight: 30,
			expectX:    0,
			expectY:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := ClampDialogPosition(tt.x, tt.y, tt.width, tt.height, tt.termWidth, tt.termHeight)

			if x != tt.expectX {
				t.Errorf("X: got %d, expected %d", x, tt.expectX)
			}
			if y != tt.expectY {
				t.Errorf("Y: got %d, expected %d", y, tt.expectY)
			}
		})
	}
}

func TestIsTerminalTooSmall(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expectSmall bool
	}{
		{
			name:     "Normal terminal",
			width:    100,
			height:   30,
			expectSmall: false,
		},
		{
			name:     "Too narrow",
			width:    30,
			height:   30,
			expectSmall: true,
		},
		{
			name:     "Too short",
			width:    100,
			height:   10,
			expectSmall: true,
		},
		{
			name:     "Both too small",
			width:    20,
			height:   5,
			expectSmall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTerminalTooSmall(tt.width, tt.height)
			if result != tt.expectSmall {
				t.Errorf("IsTerminalTooSmall: got %v, expected %v", result, tt.expectSmall)
			}
		})
	}
}

func TestIsDialogFullyVisible(t *testing.T) {
	tests := []struct {
		name      string
		x         int
		y         int
		width     int
		height    int
		termWidth int
		termHeight int
		expectVisible bool
	}{
		{
			name:          "Dialog fully visible",
			x:             10,
			y:             5,
			width:         50,
			height:        20,
			termWidth:     100,
			termHeight:    30,
			expectVisible: true,
		},
		{
			name:          "Dialog overflows right",
			x:             75,
			y:             5,
			width:         50,
			height:        20,
			termWidth:     100,
			termHeight:    30,
			expectVisible: false,
		},
		{
			name:          "Dialog overflows bottom",
			x:             25,
			y:             20,
			width:         50,
			height:        20,
			termWidth:     100,
			termHeight:    30,
			expectVisible: false,
		},
		{
			name:          "Dialog has negative position",
			x:             -5,
			y:             5,
			width:         50,
			height:        20,
			termWidth:     100,
			termHeight:    30,
			expectVisible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDialogFullyVisible(tt.x, tt.y, tt.width, tt.height, tt.termWidth, tt.termHeight)
			if result != tt.expectVisible {
				t.Errorf("IsDialogFullyVisible: got %v, expected %v", result, tt.expectVisible)
			}
		})
	}
}

func TestRepositionDialogIfNeeded(t *testing.T) {
	tests := []struct {
		name           string
		currentX       int
		currentY       int
		width          int
		height         int
		newTermWidth   int
		newTermHeight  int
		expectReposition bool
		expectX        int
		expectY        int
	}{
		{
			name:           "Position still valid",
			currentX:       25,
			currentY:       5,
			width:          50,
			height:         20,
			newTermWidth:   100,
			newTermHeight:  30,
			expectReposition: false,
			expectX:        25,
			expectY:        5,
		},
		{
			name:           "Terminal shrunk - reposition needed",
			currentX:       50,
			currentY:       15,
			width:          50,
			height:         20,
			newTermWidth:   60,
			newTermHeight:  25,
			expectReposition: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newX, newY, repositioned := RepositionDialogIfNeeded(
				tt.currentX, tt.currentY, tt.width, tt.height,
				tt.newTermWidth, tt.newTermHeight,
			)

			if repositioned != tt.expectReposition {
				t.Errorf("Repositioned: got %v, expected %v", repositioned, tt.expectReposition)
			}

			if !repositioned && tt.expectReposition == false {
				if newX != tt.currentX || newY != tt.currentY {
					t.Errorf("Position changed unexpectedly: got (%d, %d), expected (%d, %d)",
						newX, newY, tt.currentX, tt.currentY)
				}
			}
		})
	}
}

func TestCalculateOptimalDialogSize(t *testing.T) {
	tests := []struct {
		name         string
		contentWidth int
		contentHeight int
		termWidth    int
		termHeight   int
		expectWidth  int
		expectHeight int
	}{
		{
			name:          "Normal content",
			contentWidth:  40,
			contentHeight: 15,
			termWidth:     100,
			termHeight:    30,
			expectWidth:   44,
			expectHeight:  21,
		},
		{
			name:          "Content exceeds terminal",
			contentWidth:  90,
			contentHeight: 50,
			termWidth:     100,
			termHeight:    30,
			expectWidth:   94,
			expectHeight:  28,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := CalculateOptimalDialogSize(
				tt.contentWidth, tt.contentHeight,
				tt.termWidth, tt.termHeight,
			)

			if width != tt.expectWidth {
				t.Errorf("Width: got %d, expected %d", width, tt.expectWidth)
			}
			if height != tt.expectHeight {
				t.Errorf("Height: got %d, expected %d", height, tt.expectHeight)
			}
		})
	}
}
