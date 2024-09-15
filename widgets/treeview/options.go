// options.go
package treeview

import "github.com/mum4k/termdash/cell"

// Option represents a configuration option for the Treeview.
type Option func(*options)

// options holds the configuration for the Treeview.
type options struct {
	nodes         []*TreeNode
	labelColor    cell.Color
	expandedIcon  string
	collapsedIcon string
	leafIcon      string
	indentation   int
	waitingIcons  []string
	truncate      bool
	enableLogging bool
}

// newOptions initializes default options.
// Sample spinners:
// []string{'←','↖','↑','↗','→','↘','↓','↙'}
// []string{'◰','◳','◲','◱'}
// []string{'◴','◷','◶','◵'}
// []string{'◐','◓','◑','◒'}
// []string{'x','+'}
// []string{'⣾','⣽','⣻','⢿','⡿','⣟','⣯','⣷'}
// newOptions initializes default options.
func newOptions() *options {
	return &options{
		nodes:         []*TreeNode{},
		labelColor:    cell.ColorWhite,
		expandedIcon:  "▼",
		collapsedIcon: "▶",
		leafIcon:      "→",
		waitingIcons:  []string{"◐", "◓", "◑", "◒"},
		truncate:      false,
		indentation:   2, // Default indentation
	}
}

// Nodes sets the root nodes of the Treeview.
func Nodes(nodes ...*TreeNode) Option {
	return func(o *options) {
		o.nodes = nodes
	}
}

// Indentation sets the number of spaces for each indentation level.
func Indentation(spaces int) Option {
	return func(o *options) {
		o.indentation = spaces
	}
}

// Icons sets custom icons for expanded, collapsed, and leaf nodes.
func Icons(expanded, collapsed, leaf string) Option {
	return func(o *options) {
		o.expandedIcon = expanded
		o.collapsedIcon = collapsed
		o.leafIcon = leaf
	}
}

// LabelColor sets the color of the node labels.
func LabelColor(color cell.Color) Option {
	return func(o *options) {
		o.labelColor = color
	}
}

// WaitingIcons sets the icons for the spinner.
func WaitingIcons(icons []string) Option {
	return func(o *options) {
		o.waitingIcons = icons
	}
}

// Truncate enables or disables label truncation.
func Truncate(truncate bool) Option {
	return func(o *options) {
		o.truncate = truncate
	}
}

// EnableLogging enables or disables logging for debugging.
func EnableLogging(enable bool) Option {
	return func(o *options) {
		o.enableLogging = enable
	}
}

// Label sets the widget's label.
// Note: If the widget's label is managed by the container, this can be a no-op.
func Label(label string) Option {
	return func(o *options) {
		// No action needed, label is set in container's BorderTitle.
	}
}

// CollapsedIcon sets the icon for collapsed nodes.
func CollapsedIcon(icon string) Option {
	return func(o *options) {
		o.collapsedIcon = icon
	}
}

// ExpandedIcon sets the icon for expanded nodes.
func ExpandedIcon(icon string) Option {
	return func(o *options) {
		o.expandedIcon = icon
	}
}

// LeafIcon sets the icon for leaf nodes.
func LeafIcon(icon string) Option {
	return func(o *options) {
		o.leafIcon = icon
	}
}

// IndentationPerLevel sets the indentation per level.
// Alias to Indentation for compatibility with demo code.
func IndentationPerLevel(spaces int) Option {
	return Indentation(spaces)
}
