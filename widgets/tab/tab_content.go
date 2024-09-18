// tab/tab_content.go
package tab

import (
	"github.com/mum4k/termdash/container"
)

// TabContent displays the content of the active tab.
type TabContent struct {
	tm *TabManager
}

// NewTabContent creates a new TabContent.
func NewTabContent(tm *TabManager) *TabContent {
	return &TabContent{
		tm: tm,
	}
}

// Update updates the tab content based on the active tab.
func (tc *TabContent) Update(cont *container.Container) error {
	activeTab := tc.tm.GetActiveTab()
	if activeTab == nil {
		return nil
	}
	return cont.Update("tabContent", activeTab.Content)
}
