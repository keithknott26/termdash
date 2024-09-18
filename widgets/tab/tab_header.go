// tab/tab_header.go
package tab

import (
	"fmt"
	"image"
	"strings"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

// TabHeader displays the tab names and highlights the active tab.
type TabHeader struct {
	tm            *TabManager
	widget        *text.Text
	tabRectangles []image.Rectangle // New field to track tab positions
	height        int               // Stores the height of the TabHeader
	opts          *Options
}

// NewTabHeader creates a new TabHeader.
func NewTabHeader(tm *TabManager, opts *Options) (*TabHeader, error) {
	w, err := text.New(
		text.WrapAtRunes(),      // Use WrapAtRunes for accurate width calculations
		text.DisableScrolling(), // Disable scrolling to keep the header static
	)
	if err != nil {
		return nil, err
	}

	// If opts is nil, create default options
	if opts == nil {
		opts = NewOptions(
			LabelColor(cell.ColorWhite),
			ActiveTabColor(cell.ColorBlue),
			InactiveTabColor(cell.ColorBlack),
			ActiveIcon("⦿"),
			InactiveIcon("○"),
			NotificationIcon("⚠"),
			EnableLogging(false),
		)
	}

	th := &TabHeader{
		tm:     tm,
		widget: w,
		height: 2, // Assuming it occupies 2 lines (tabs and underline)
		opts:   opts,
	}

	if err := th.Update(); err != nil {
		return nil, err
	}

	return th, nil
}

// update refreshes the tab header display.
func (th *TabHeader) update() error {
	tabNames := th.tm.GetTabNames()
	activeIndex := th.tm.GetActiveIndex()

	th.widget.Reset()
	th.tabRectangles = make([]image.Rectangle, 0, len(tabNames)) // Initialize

	currentX := 1

	// Build the tabs with styling
	var totalWidth int

	for i := range tabNames {
		// Determine icons
		var notificationIcon string
		if th.tm.tabs[i].Notification {
			notificationIcon = th.opts.NotificationIcon
		} else {
			notificationIcon = " " // Empty space if no notification
		}

		var stateIcon string
		if i == th.tm.GetActiveIndex() {
			stateIcon = th.opts.ActiveIcon
		} else {
			stateIcon = th.opts.InactiveIcon
		}
		// Construct tab label with icon and name
		tabText := fmt.Sprintf("│%s%s %s ", notificationIcon, stateIcon, th.tm.GetTabName(i))
		tabWidth := len([]rune(tabText))
		totalWidth += tabWidth

		// Assuming the TabHeader spans 2 lines (tabs and underline)
		const tabHeaderHeight = 2

		// In your update function
		rect := image.Rectangle{
			Min: image.Point{X: currentX, Y: 0},
			Max: image.Point{X: currentX + tabWidth, Y: tabHeaderHeight},
		}

		th.tabRectangles = append(th.tabRectangles, rect)

		currentX += tabWidth

		if i == activeIndex {
			// Active tab styling
			if err := th.widget.Write(tabText,
				text.WriteCellOpts(cell.Bold(), cell.FgColor(th.opts.LabelColor), cell.BgColor(th.opts.ActiveTabColor)),
			); err != nil {
				return err
			}
		} else {
			// Inactive tab styling
			if err := th.widget.Write(tabText,
				text.WriteCellOpts(cell.FgColor(th.opts.LabelColor), cell.BgColor(th.opts.InactiveTabColor)),
			); err != nil {
				return err
			}
		}
	}

	// Add the closing border
	closingBorder := "│\n"
	totalWidth += len([]rune(closingBorder))
	if err := th.widget.Write(closingBorder, text.WriteCellOpts(cell.FgColor(th.opts.LabelColor))); err != nil {
		return err
	}

	// Add an underline to separate the tabs from the content
	underline := strings.Repeat("─", totalWidth)
	if err := th.widget.Write(underline, text.WriteCellOpts(cell.FgColor(th.opts.LabelColor))); err != nil {
		return err
	}

	return nil
}

// Update refreshes the tab header to reflect changes.
func (th *TabHeader) Update() error {
	th.widget.Reset()
	for idx, tab := range th.tm.tabs {
		var icon string
		var color cell.Color
		if idx == th.tm.GetActiveIndex() {
			icon = th.opts.ActiveIcon
			color = th.opts.ActiveTabColor
		} else {
			icon = th.opts.InactiveIcon
			color = th.opts.InactiveTabColor
		}
		label := fmt.Sprintf("%s %s", icon, tab.Name)
		if err := th.widget.Write(label, text.WriteCellOpts(
			cell.FgColor(th.opts.LabelColor),
			cell.BgColor(color),
		)); err != nil {
			return err
		}
		if idx < len(th.tm.tabs)-1 {
			if err := th.widget.Write(" | "); err != nil {
				return err
			}
		}
	}
	th.update()
	return nil
}

// Widget returns the underlying text widget.
func (th *TabHeader) Widget() *text.Text {
	return th.widget
}

// Height returns the height of the TabHeader.
func (th *TabHeader) Height() int {
	return th.height
}

// GetClickedTab returns the index of the tab that was clicked based on the mouse position.
// Returns -1 if no tab was clicked.
func (th *TabHeader) GetClickedTab(p image.Point) int {
	for i, rect := range th.tabRectangles {
		if rect.Min.X <= p.X && p.X < rect.Max.X &&
			rect.Min.Y <= p.Y && p.Y < rect.Max.Y {
			return i
		}
	}
	return -1
}
