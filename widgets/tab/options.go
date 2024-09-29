// Package tab provides configuration options for the tabbed interface.
package tab

import "github.com/mum4k/termdash/cell"

// Option represents a configuration option for the Tab.
type Option interface {
	// set applies the option to the provided Options struct.
	set(*Options)
}

// Options holds the configuration for the Tab.
type Options struct {
	Tabs                []*Tab     // List of tabs.
	LabelColor          cell.Color // Color of the tab labels.
	ActiveTabColor      cell.Color // Background color of the active tab.
	InactiveTabColor    cell.Color // Background color of inactive tabs.
	ActiveIcon          string     // Icon for active tabs.
	InactiveIcon        string     // Icon for inactive tabs.
	NotificationIcon    string     // Icon for tabs with notifications.
	EnableLogging       bool       // Enables logging for debugging.
	FollowNotifications bool       // Whether to follow notifications automatically.
}

// NewOptions initializes default options or applies provided options.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		Tabs:                []*Tab{},
		LabelColor:          cell.ColorWhite,
		ActiveTabColor:      cell.ColorBlue,
		InactiveTabColor:    cell.ColorBlack,
		ActiveIcon:          "⦿",
		InactiveIcon:        "○",
		NotificationIcon:    "⚠",
		EnableLogging:       false,
		FollowNotifications: false,
	}
	for _, opt := range opts {
		opt.set(o)
	}
	return o
}

// option is a function that modifies Options.
type option func(*Options)

// set implements Option.set.
func (o option) set(opts *Options) {
	o(opts)
}

// Tabs sets the root tabs of the Tab widget.
func Tabs(tabs ...*Tab) Option {
	return option(func(o *Options) {
		o.Tabs = tabs
	})
}

// ActiveIcon sets custom icons for the active state.
func ActiveIcon(active string) Option {
	return option(func(o *Options) {
		o.ActiveIcon = active
	})
}

// InactiveIcon sets custom icons for the inactive state.
func InactiveIcon(inactive string) Option {
	return option(func(o *Options) {
		o.InactiveIcon = inactive
	})
}

// NotificationIcon sets custom icons for notifications.
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

// ActiveTabColor sets the background color of the active tab.
func ActiveTabColor(color cell.Color) Option {
	return option(func(o *Options) {
		o.ActiveTabColor = color
	})
}

// InactiveTabColor sets the background color of inactive tabs.
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
