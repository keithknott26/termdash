// Package tab provides functionality for managing tabbed interfaces.
package tab

import (
	"sync"
	"time"

	"github.com/mum4k/termdash/container"
)

// Tab represents a single tab with a name and content.
type Tab struct {
	Name                 string           // Name of the tab.
	Content              container.Option // Content to display when the tab is active.
	Notification         bool             // Indicates if the tab has a notification.
	notifyTime           time.Time        // Timestamp when the notification was set.
	notificationDuration time.Duration    // Duration of the notification.
}

// SetNotification sets the notification state of the tab with a duration.
func (t *Tab) SetNotification(hasNotification bool, duration time.Duration) {
	t.Notification = hasNotification
	if hasNotification {
		t.notifyTime = time.Now()
		t.notificationDuration = duration
	} else {
		t.notifyTime = time.Time{}
		t.notificationDuration = 0
	}
}

// HasNotification returns whether the tab has an active notification.
func (t *Tab) HasNotification() bool {
	if t.Notification && time.Since(t.notifyTime) < t.notificationDuration {
		return true
	}
	// If the notification duration has passed, clear the notification.
	t.Notification = false
	t.notificationDuration = 0
	return false
}

// Manager handles multiple tabs and the active tab index.
type Manager struct {
	tabs        []*Tab     // Slice of tabs managed by the Manager.
	activeIndex int        // Index of the currently active tab.
	mu          sync.Mutex // Mutex to protect concurrent access.
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		tabs:        []*Tab{},
		activeIndex: 0,
	}
}

// AddTab adds a new tab to the manager.
func (m *Manager) AddTab(tab *Tab) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tabs = append(m.tabs, tab)
}

// NextTab switches to the next tab.
func (m *Manager) NextTab() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.tabs) == 0 {
		return
	}
	m.activeIndex = (m.activeIndex + 1) % len(m.tabs)
}

// PreviousTab switches to the previous tab.
func (m *Manager) PreviousTab() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.tabs) == 0 {
		return
	}
	m.activeIndex = (m.activeIndex - 1 + len(m.tabs)) % len(m.tabs)
}

// GetActiveTab returns the currently active tab.
func (m *Manager) GetActiveTab() *Tab {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.tabs) == 0 {
		return nil
	}
	return m.tabs[m.activeIndex]
}

// GetTabNum returns the total number of tabs.
func (m *Manager) GetTabNum() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tabs)
}

// GetTabNames returns the names of all tabs.
func (m *Manager) GetTabNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, len(m.tabs))
	for i, tab := range m.tabs {
		names[i] = tab.Name
	}
	return names
}

// GetTabName returns the name of the tab at the specified index.
func (m *Manager) GetTabName(index int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index >= 0 && index < len(m.tabs) {
		return m.tabs[index].Name
	}
	return ""
}

// GetTabIndex returns the index of a given tab.
func (m *Manager) GetTabIndex(targetTab *Tab) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	for index, tab := range m.tabs {
		if tab == targetTab {
			return index
		}
	}
	return -1
}

// GetActiveIndex returns the index of the active tab.
func (m *Manager) GetActiveIndex() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeIndex
}

// SetActiveTab sets the active tab to the specified index.
func (m *Manager) SetActiveTab(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index >= 0 && index < len(m.tabs) {
		m.activeIndex = index
		// Do not clear the notification here; let it expire naturally.
	}
}

// SetNotification sets the notification state for the specified tab with a duration.
func (m *Manager) SetNotification(index int, hasNotification bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index >= 0 && index < len(m.tabs) {
		m.tabs[index].SetNotification(hasNotification, duration)
	}
}

// GetNotifiedTabs returns a list of tabs with active notifications.
func (m *Manager) GetNotifiedTabs() []*Tab {
	m.mu.Lock()
	defer m.mu.Unlock()
	var notifiedTabs []*Tab
	for _, tab := range m.tabs {
		if tab.HasNotification() {
			notifiedTabs = append(notifiedTabs, tab)
		}
	}
	return notifiedTabs
}
