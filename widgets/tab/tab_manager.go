// tab/tab_manager.go
package tab

import (
	"sync"
	"time"

	"github.com/mum4k/termdash/container"
)

// Tab represents a single tab with a name and content.
type Tab struct {
	Name                 string
	Content              container.Option
	Notification         bool          // Field to track notification state
	notifyTime           time.Time     // Timestamp when the notification was set
	notificationDuration time.Duration // Duration of the notification
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

// TabManager manages multiple tabs and the active tab index.
type TabManager struct {
	tabs        []*Tab
	activeIndex int
	mu          sync.Mutex
}

// NewTabManager creates a new TabManager.
func NewTabManager() *TabManager {
	return &TabManager{
		tabs:        []*Tab{},
		activeIndex: 0,
	}
}

// AddTab adds a new tab to the manager.
func (tm *TabManager) AddTab(tab *Tab) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tabs = append(tm.tabs, tab)
}

// NextTab switches to the next tab.
func (tm *TabManager) NextTab() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if len(tm.tabs) == 0 {
		return
	}
	tm.activeIndex = (tm.activeIndex + 1) % len(tm.tabs)
}

// PreviousTab switches to the previous tab.
func (tm *TabManager) PreviousTab() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if len(tm.tabs) == 0 {
		return
	}
	tm.activeIndex = (tm.activeIndex - 1 + len(tm.tabs)) % len(tm.tabs)
}

// GetActiveTab returns the currently active tab.
func (tm *TabManager) GetActiveTab() *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if len(tm.tabs) == 0 {
		return nil
	}
	return tm.tabs[tm.activeIndex]
}

// GetTabNum returns the total number of tabs.
func (tm *TabManager) GetTabNum() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return len(tm.tabs)
}

// GetTabNames returns the names of all tabs.
func (tm *TabManager) GetTabNames() []string {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	names := make([]string, len(tm.tabs))
	for i, tab := range tm.tabs {
		names[i] = tab.Name
	}
	return names
}

// GetTabNames returns the name of the tab with index x.
func (tm *TabManager) GetTabName(index int) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	var name string
	for i, tab := range tm.tabs {
		if i == index {
			name = tab.Name
		}
	}
	return name
}

// GetTabIndex returns the index of a given tab.
func (tm *TabManager) GetTabIndex(targetTab *Tab) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for index, tab := range tm.tabs {
		if tab == targetTab {
			return index
		}
	}
	return -1
}

// GetActiveIndex returns the index of the active tab.
func (tm *TabManager) GetActiveIndex() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.activeIndex
}

// SetActiveTab sets the active tab to the specified index.
func (tm *TabManager) SetActiveTab(index int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if index >= 0 && index < len(tm.tabs) {
		tm.activeIndex = index
		// Do not clear the notification here; let it expire naturally.
	}
}

// SetNotification sets the notification state for the specified tab with a duration.
func (tm *TabManager) SetNotification(index int, hasNotification bool, duration time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if index >= 0 && index < len(tm.tabs) {
		tm.tabs[index].SetNotification(hasNotification, duration)
	}
}

// GetNotifiedTabs returns a list of tabs with active notifications.
func (tm *TabManager) GetNotifiedTabs() []*Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	var notifiedTabs []*Tab
	for _, tab := range tm.tabs {
		if tab.HasNotification() {
			notifiedTabs = append(notifiedTabs, tab)
		}
	}
	return notifiedTabs
}
