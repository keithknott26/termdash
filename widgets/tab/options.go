// options.go
package tab

import "github.com/mum4k/termdash/cell"

// Option represents a configuration option for the Tab.
type Option interface {
	// set applies the option to the provided Options struct.
	set(*Options)
}

// Options holds the configuration for the Tab.
type Options struct {
	Tabs                []*Tab
	LabelColor          cell.Color
	ActiveTabColor      cell.Color
	InactiveTabColor    cell.Color
	ActiveIcon          string
	InactiveIcon        string
	NotificationIcon    string
	EnableLogging       bool
	FollowNotifications bool // Whether to follow notifications automatically
}

// newOptions initializes default options.
func newOptions() *Options {
	return &Options{
		Tabs:                []*Tab{},
		LabelColor:          cell.ColorWhite,
		ActiveTabColor:      cell.ColorBlue,
		InactiveTabColor:    cell.ColorBlack,
		ActiveIcon:          "⦿",
		InactiveIcon:        "○",
		NotificationIcon:    "⚠",
		EnableLogging:       false, // Default to false
		FollowNotifications: false, // Default to false
	}
}

// option is a function that modifies Options.
type option func(*Options)

// set implements Option.set.
func (o option) set(opts *Options) {
	o(opts)
}

// NewOptions creates a new Options struct with the given options applied.
func NewOptions(opts ...Option) *Options {
	o := newOptions()
	for _, opt := range opts {
		opt.set(o)
	}
	return o
}

// Tabs sets the root tabs of the Tab widget.
func Tabs(tabs ...*Tab) Option {
	return option(func(o *Options) {
		o.Tabs = tabs
	})
}

// ActiveIcon sets custom icons for the active state
func ActiveIcon(active string) Option {
	return option(func(o *Options) {
		o.ActiveIcon = active
	})
}

// ActiveIcon sets custom icons for the active state
func InactiveIcon(inactive string) Option {
	return option(func(o *Options) {
		o.InactiveIcon = inactive
	})
}

// ActiveIcon sets custom icons for the active state
func NotificationIcon(notification string) Option {
	return option(func(o *Options) {
		o.NotificationIcon = notification
	})
}

// LabelColor sets the color of the tab labels.
func LabelColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.LabelColor = color
	})
}

// ActiveTabColor sets the color of the active tab.
func ActiveTabColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.ActiveTabColor = color
	})
}

// ActiveTabColor sets the color of the inactive tabs.
func InactiveTabColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.InactiveTabColor = color
	})
}

// EnableLogging enables or disables logging for debugging.
func EnableLogging(enable bool) Option {
	return option(func(o *Options) {
		o.EnableLogging = enable
	})
}

// FollowNotifications sets whether the app should follow notifications automatically.
func FollowNotifications(enable bool) Option {
	return option(func(o *Options) {
		o.FollowNotifications = enable
	})
}
