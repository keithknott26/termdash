// Package tab provides functionality for managing tabbed interfaces.
package tab

import (
	"fmt"
	"image"
	"strings"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/text"
)

// Header displays the tab names and highlights the active tab.
type Header struct {
	tm            *Manager          // Reference to the Tab Manager.
	widget        *text.Text        // Text widget for displaying the header.
	tabRectangles []image.Rectangle // Tracks the positions of tabs for mouse interactions.
	height        int               // Stores the height of the Header.
	opts          *Options          // Configuration options for the Header.
}

// NewHeader creates a new Header.
func NewHeader(tm *Manager, opts *Options) (*Header, error) {
	w, err := text.New(
		text.WrapAtRunes(),      // Use WrapAtRunes for accurate width calculations.
		text.DisableScrolling(), // Disable scrolling to keep the header static.
	)
	if err != nil {
		return nil, err
	}

	// If opts is nil, create default options.
	if opts == nil {
		opts = NewOptions()
	}

	header := &Header{
		tm:     tm,
		widget: w,
		height: 2, // Assuming it occupies 2 lines (tabs and underline).
		opts:   opts,
	}

	if err := header.Update(); err != nil {
		return nil, err
	}

	return header, nil
}

// Update refreshes the tab header to reflect changes.
func (h *Header) Update() error {
	tabNames := h.tm.GetTabNames()
	activeIndex := h.tm.GetActiveIndex()

	h.widget.Reset()
	h.tabRectangles = make([]image.Rectangle, 0, len(tabNames)) // Initialize.

	currentX := 1
	var totalWidth int

	for i, tabName := range tabNames {
		// Determine icons.
		var notificationIcon string
		if h.tm.tabs[i].HasNotification() {
			notificationIcon = h.opts.NotificationIcon
		} else {
			notificationIcon = " " // Empty space if no notification.
		}

		var stateIcon string
		if i == activeIndex {
			stateIcon = h.opts.ActiveIcon
		} else {
			stateIcon = h.opts.InactiveIcon
		}

		// Construct tab label with icons and name.
		tabText := fmt.Sprintf("│%s%s %s ", notificationIcon, stateIcon, tabName)
		tabWidth := len([]rune(tabText))
		totalWidth += tabWidth

		// Assuming the Header spans 2 lines (tabs and underline).
		const headerHeight = 2

		// Compute the rectangle for the tab.
		rect := image.Rectangle{
			Min: image.Point{X: currentX, Y: 0},
			Max: image.Point{X: currentX + tabWidth, Y: headerHeight},
		}

		h.tabRectangles = append(h.tabRectangles, rect)

		currentX += tabWidth

		if i == activeIndex {
			// Active tab styling.
			if err := h.widget.Write(tabText,
				text.WriteCellOpts(cell.Bold(), cell.FgColor(h.opts.LabelColor), cell.BgColor(h.opts.ActiveTabColor)),
			); err != nil {
				return err
			}
		} else {
			// Inactive tab styling.
			if err := h.widget.Write(tabText,
				text.WriteCellOpts(cell.FgColor(h.opts.LabelColor), cell.BgColor(h.opts.InactiveTabColor)),
			); err != nil {
				return err
			}
		}
	}

	// Add the closing border.
	closingBorder := "│\n"
	totalWidth += len([]rune(closingBorder))
	if err := h.widget.Write(closingBorder, text.WriteCellOpts(cell.FgColor(h.opts.LabelColor))); err != nil {
		return err
	}

	// Add an underline to separate the tabs from the content.
	underline := strings.Repeat("─", totalWidth)
	if err := h.widget.Write(underline, text.WriteCellOpts(cell.FgColor(h.opts.LabelColor))); err != nil {
		return err
	}

	return nil
}

// Widget returns the underlying text widget.
func (h *Header) Widget() *text.Text {
	return h.widget
}

// Height returns the height of the Header.
func (h *Header) Height() int {
	return h.height
}

// GetClickedTab returns the index of the tab that was clicked based on the mouse position.
// Returns -1 if no tab was clicked.
func (h *Header) GetClickedTab(p image.Point) int {
	for i, rect := range h.tabRectangles {
		if rect.Min.X <= p.X && p.X < rect.Max.X &&
			rect.Min.Y <= p.Y && p.Y < rect.Max.Y {
			return i
		}
	}
	return -1
}
