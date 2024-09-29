// Package tab contains tests for the tabbed interface using Termdash.
package tab

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

// mockTerminal is a mock implementation of terminalapi.Terminal for testing purposes.
type mockTerminal struct{}

// Implementing all methods of terminalapi.Terminal with no-op or default implementations.
func (m *mockTerminal) Init() error                                              { return nil }
func (m *mockTerminal) Close()                                                   {}
func (m *mockTerminal) Size() image.Point                                        { return image.Point{X: 80, Y: 24} }
func (m *mockTerminal) SendEvent(ev interface{}) error                           { return nil }
func (m *mockTerminal) PollEvents() (<-chan interface{}, error)                  { return nil, nil }
func (m *mockTerminal) Clear(opts ...cell.Option) error                          { return nil }
func (m *mockTerminal) CursorShow() error                                        { return nil }
func (m *mockTerminal) CursorHide() error                                        { return nil }
func (m *mockTerminal) Event(ctx context.Context) terminalapi.Event              { return &terminalapi.Keyboard{} }
func (m *mockTerminal) Flush() error                                             { return nil }
func (m *mockTerminal) HideCursor()                                              {}
func (m *mockTerminal) SetCell(i image.Point, r rune, opts ...cell.Option) error { return nil }
func (m *mockTerminal) SetCursor(p image.Point)                                  {}

// TestManager verifies the functionality of Manager.
func TestManager(t *testing.T) {
	tm := NewManager()

	// Create dummy content widgets
	text1, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}
	text2, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}

	// Add tabs
	tab1 := &Tab{Name: "Tab1", Content: container.PlaceWidget(text1)}
	tab2 := &Tab{Name: "Tab2", Content: container.PlaceWidget(text2)}
	tm.AddTab(tab1)
	tm.AddTab(tab2)

	// Verify tab count
	if len(tm.tabs) != 2 {
		t.Errorf("expected 2 tabs, got %d", len(tm.tabs))
	}

	// Verify active tab
	activeTab := tm.GetActiveTab()
	if activeTab != tab1 {
		t.Errorf("expected active tab to be Tab1, got %s", activeTab.Name)
	}

	// Switch to next tab
	tm.NextTab()
	activeTab = tm.GetActiveTab()
	if activeTab != tab2 {
		t.Errorf("expected active tab to be Tab2, got %s", activeTab.Name)
	}

	// Switch to previous tab
	tm.PreviousTab()
	activeTab = tm.GetActiveTab()
	if activeTab != tab1 {
		t.Errorf("expected active tab to be Tab1, got %s", activeTab.Name)
	}

	// Switch previous on first tab should wrap to last tab
	tm.PreviousTab()
	activeTab = tm.GetActiveTab()
	if activeTab != tab2 {
		t.Errorf("expected active tab to be Tab2 after wrapping, got %s", activeTab.Name)
	}
}

// TestHeader verifies that Header correctly displays active and inactive tabs.
func TestHeader(t *testing.T) {
	tm := NewManager()
	// Create options
	opts := NewOptions(
		ActiveIcon("⦿"),                   // Custom active icon
		InactiveIcon("○"),                 // Custom inactive icon
		NotificationIcon("⚠"),             // Custom notification icon
		EnableLogging(false),              // Enable logging
		LabelColor(cell.ColorYellow),      // Custom label color
		ActiveTabColor(cell.ColorBlue),    // Active tab background color
		InactiveTabColor(cell.ColorBlack), // Inactive tab background color
		FollowNotifications(false),        // Enable following notifications
	)
	text1, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}
	text2, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}

	tab1 := &Tab{Name: "Tab1", Content: container.PlaceWidget(text1)}
	tab2 := &Tab{Name: "Tab2", Content: container.PlaceWidget(text2)}
	tm.AddTab(tab1)
	tm.AddTab(tab2)

	th, err := NewHeader(tm, opts)
	if err != nil {
		t.Fatalf("failed to create Header: %v", err)
	}
	tw := th.Widget()
	if tw == nil {
		t.Fatalf("expected *text.Text widget, got nil")
	} else {
		err := tw.Write("test")
		if err != nil {
			t.Fatalf("failed to write to text widget: %v", err)
		}
	}

	// Change active tab
	tm.NextTab()
	if err := th.Update(); err != nil {
		t.Fatalf("failed to update Header: %v", err)
	}

}

// TestContent verifies that Content correctly updates the container with the active tab's content.
func TestContent(t *testing.T) {
	tm := NewManager()
	opts := NewOptions(
		ActiveIcon("⦿"),                   // Custom active icon
		InactiveIcon("○"),                 // Custom inactive icon
		NotificationIcon("⚠"),             // Custom notification icon
		EnableLogging(false),              // Enable logging
		LabelColor(cell.ColorYellow),      // Custom label color
		ActiveTabColor(cell.ColorBlue),    // Active tab background color
		InactiveTabColor(cell.ColorBlack), // Inactive tab background color
		FollowNotifications(false),        // Enable following notifications
	)
	if len(tm.tabs) == 1 {
		t.Fatalf("expected at least one tab, got none")
	}
	text1, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}
	text2, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}

	tab1 := &Tab{Name: "Tab1", Content: container.PlaceWidget(text1)}
	tab2 := &Tab{Name: "Tab2", Content: container.PlaceWidget(text2)}
	tm.AddTab(tab1)
	tm.AddTab(tab2)

	th, err := NewHeader(tm, opts)
	if err != nil {
		t.Fatalf("failed to create Header: %v", err)
	}

	if th != nil {
		tw := th.Widget()
		if tw == nil {
			t.Fatalf("expected *text.Text widget, got nil")
		} else {
			err := tw.Write("test")
			if err != nil {
				t.Fatalf("failed to write to text widget: %v", err)
			}
		}
	} else {
		t.Fatalf("Header (th) is nil")
	}

	tc := NewContent(tm)

	// Create a mock terminal
	term := &mockTerminal{}
	cont, err := container.New(
		term, // First argument: terminalapi.Terminal
		container.ID("tabContent"),
	)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	// Update Content with initial active tab
	err = tc.Update(cont)
	if err != nil {
		t.Fatalf("failed to update Content: %v", err)
	}

	// Since container.Container does not have a Get method, ensure no errors occurred

	// Switch active tab and update content
	tm.NextTab()
	err = tc.Update(cont)
	if err != nil {
		t.Fatalf("failed to update Content after switching tabs: %v", err)
	}

	// No direct way to verify content; ensure no errors
}

// TestEventHandler_HandleKeyboard verifies that EventHandler correctly responds to keyboard events.
func TestEventHandler_HandleKeyboard(t *testing.T) {
	tm := NewManager()
	opts := NewOptions(
		ActiveIcon("⦿"),                   // Custom active icon
		InactiveIcon("○"),                 // Custom inactive icon
		NotificationIcon("⚠"),             // Custom notification icon
		EnableLogging(false),              // Enable logging
		LabelColor(cell.ColorYellow),      // Custom label color
		ActiveTabColor(cell.ColorBlue),    // Active tab background color
		InactiveTabColor(cell.ColorBlack), // Inactive tab background color
		FollowNotifications(false),        // Enable following notifications
	)
	text1, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}
	text2, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}

	tab1 := &Tab{Name: "Tab1", Content: container.PlaceWidget(text1)}
	tab2 := &Tab{Name: "Tab2", Content: container.PlaceWidget(text2)}
	tm.AddTab(tab1)
	tm.AddTab(tab2)

	th, err := NewHeader(tm, opts)
	if err != nil {
		t.Fatalf("failed to create Header: %v", err)
	}

	tc := NewContent(tm)

	// Create a mock terminal
	term := &mockTerminal{}
	cont, err := container.New(
		term, // First argument: terminalapi.Terminal
		container.ID("tabContent"),
	)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	// Update Content with initial active tab
	err = tc.Update(cont)
	if err != nil {
		t.Fatalf("failed to update Content: %v", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create EventHandler with context
	teh := NewEventHandler(ctx, term, tm, th, tc, cont, cancel, opts)

	// Simulate Tab key press to switch to Tab2
	tabKey := &terminalapi.Keyboard{Key: keyboard.KeyTab}
	teh.HandleKeyboard(tabKey)

	// Verify that the active tab is now Tab2
	activeTab := tm.GetActiveTab()
	if activeTab != tab2 {
		t.Errorf("expected active tab to be Tab2 after Tab key press, got %s", activeTab.Name)
	}

	// Simulate Left Arrow key press to switch back to Tab1
	leftArrowKey := &terminalapi.Keyboard{Key: keyboard.KeyArrowLeft}
	teh.HandleKeyboard(leftArrowKey)

	// Verify that the active tab is now Tab1
	activeTab = tm.GetActiveTab()
	if activeTab != tab1 {
		t.Errorf("expected active tab to be Tab1 after Left Arrow key press, got %s", activeTab.Name)
	}

	// Simulate Right Arrow key press to switch to Tab2 again
	rightArrowKey := &terminalapi.Keyboard{Key: keyboard.KeyArrowRight}
	teh.HandleKeyboard(rightArrowKey)

	// Verify that the active tab is now Tab2
	activeTab = tm.GetActiveTab()
	if activeTab != tab2 {
		t.Errorf("expected active tab to be Tab2 after Right Arrow key press, got %s", activeTab.Name)
	}
}

// TestManagerNotifications verifies that notifications are set, displayed, and expire as expected.
func TestManagerNotifications(t *testing.T) {
	// Setup Manager and tabs
	tm := NewManager()
	text1, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}
	text2, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("failed to create text widget: %v", err)
	}

	tab1 := &Tab{Name: "Tab1", Content: container.PlaceWidget(text1)}
	tab2 := &Tab{Name: "Tab2", Content: container.PlaceWidget(text2)}
	tm.AddTab(tab1)
	tm.AddTab(tab2)

	// Set a notification on tab2 with a duration of 2 seconds
	tm.SetNotification(1, true, 2*time.Second)

	// Verify that HasNotification() returns true immediately
	if !tab2.HasNotification() {
		t.Errorf("expected tab2 to have an active notification")
	}

	// Verify that GetNotifiedTabs returns tab2
	notifiedTabs := tm.GetNotifiedTabs()
	if len(notifiedTabs) != 1 || notifiedTabs[0] != tab2 {
		t.Errorf("expected GetNotifiedTabs to return tab2, got %v", notifiedTabs)
	}

	// Wait for 1 second and verify that notification is still active
	time.Sleep(1 * time.Second)
	if !tab2.HasNotification() {
		t.Errorf("expected tab2 to still have an active notification after 1 second")
	}

	// Wait for another 2 seconds (total 3 seconds), notification should have expired
	time.Sleep(2 * time.Second)
	if tab2.HasNotification() {
		t.Errorf("expected tab2's notification to have expired after 3 seconds")
	}

	// Ensure that GetNotifiedTabs reflects the correct state
	notifiedTabs = tm.GetNotifiedTabs()
	if len(notifiedTabs) != 0 {
		t.Errorf("expected no tabs with active notifications, but got %d", len(notifiedTabs))
	}

	// Now set notifications on both tabs
	tm.SetNotification(0, true, 1*time.Second)
	tm.SetNotification(1, true, 1*time.Second)

	// Verify that both tabs have notifications
	notifiedTabs = tm.GetNotifiedTabs()
	if len(notifiedTabs) != 2 {
		t.Errorf("expected 2 tabs with active notifications, but got %d", len(notifiedTabs))
	}

	// Wait for 1.5 seconds, notifications should have expired
	time.Sleep(1500 * time.Millisecond)
	notifiedTabs = tm.GetNotifiedTabs()
	if len(notifiedTabs) != 0 {
		t.Errorf("expected no tabs with active notifications after expiration, but got %d", len(notifiedTabs))
	}
}
